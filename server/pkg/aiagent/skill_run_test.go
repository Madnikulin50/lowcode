package aiagent

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
	"github.com/stretchr/testify/require"
)

// a skill that needs a toolkit the agent does not start with
func setupDemoSkill(t *testing.T) {
	t.Helper()
	DefaultCatalog().Register(ToolKit{Name: "demo.kit", Tools: []chat.ToolDef{
		echoTool("demo_read"),
		{Name: "demo_write", Description: "changes data", Mutating: true, Handler: func(context.Context, map[string]string) string { return "written" }},
	}})
	SetSkillSource(SkillSourceDB, fakeSkillSource{[]Skill{{
		Handle: "estimate", Description: "When estimating work", Body: "Count the walls first.", Version: 3,
		Requires:  []string{"demo.kit", "gone.kit"},
		Resources: []SkillFile{{Path: "prices.csv", Content: "wall,10"}},
	}}})
	t.Cleanup(func() {
		DefaultCatalog().Unregister("demo.kit")
		SetSkillSource(SkillSourceDB, nil)
	})
}

func skillOptions(t *testing.T, m *fakeModel, confirmed bool) (Options, *SkillRun) {
	t.Helper()
	ctx := context.Background()
	run := NewSkillRun(SkillsFor(ctx, []string{"*"}))
	var tools []chat.ToolDef
	tools = SkillTools(run, func() []chat.ToolDef { return tools })
	return Options{
		Client: m, Messages: []*schema.Message{schema.UserMessage("estimate the job")},
		Tools: tools, NeedsConfirm: NeedsConfirmFromToolDefs(tools), Confirmed: confirmed,
		MaxSteps: 6, SkipWarm: true,
	}, run
}

func TestSkillRun_LoadBringsTools(t *testing.T) {
	setupDemoSkill(t)
	req := require.New(t)

	m := &fakeModel{tools: true, gen: []genTurn{
		// before the skill is loaded the tool does not exist
		{content: `<tool name="demo_read"><param name="q">early</param></tool>`},
		{content: `<tool name="load_skill"><param name="handle">estimate</param></tool>`},
		{content: `<tool name="demo_read"><param name="q">walls</param></tool>`},
		{content: "FINAL: done"},
	}}
	opt, run := skillOptions(t, m, false)

	var fed []string
	opt.Continue = func(_ string, toolResult string, _ []Call) ContinueHint {
		fed = append(fed, toolResult)
		return AgentContinue("", toolResult, nil)
	}

	out := Run(ContextWithSkillRun(context.Background(), run), opt)
	req.True(out.Success, out.Error)
	req.Len(fed, 3)
	req.Contains(fed[0], "unknown tool: demo_read")
	req.Contains(fed[1], "Count the walls first.")
	req.Contains(fed[1], "prices.csv")
	req.Contains(fed[1], "demo_read", "the tools the skill brought are announced")
	req.Contains(fed[1], "gone.kit", "a missing toolkit is reported, not fatal")
	req.Equal("echo:walls", fed[2])
	req.Equal([]string{"estimate@3"}, run.Loaded())
}

func TestSkillRun_BroughtMutatingToolNeedsConfirm(t *testing.T) {
	setupDemoSkill(t)
	req := require.New(t)

	script := []genTurn{
		{content: `<tool name="load_skill"><param name="handle">estimate</param></tool>`},
		{content: `<tool name="demo_write"></tool>`},
		{content: "FINAL: done"},
	}

	// not confirmed: the run stops at the write, which did not execute
	opt, run := skillOptions(t, &fakeModel{tools: true, gen: script}, false)
	out := Run(ContextWithSkillRun(context.Background(), run), opt)
	req.True(out.ConfirmNeeded)
	req.Equal("demo_write", out.ConfirmCalls[0].Name)

	// confirmed: it goes through
	opt, run = skillOptions(t, &fakeModel{tools: true, gen: script}, true)
	var fed []string
	opt.Continue = func(_ string, toolResult string, _ []Call) ContinueHint {
		fed = append(fed, toolResult)
		return AgentContinue("", toolResult, nil)
	}
	out = Run(ContextWithSkillRun(context.Background(), run), opt)
	req.True(out.Success, out.Error)
	req.Equal("written", fed[1])
}

func TestSkillRun_ReadResourceAndAllowList(t *testing.T) {
	setupDemoSkill(t)
	req := require.New(t)
	ctx := context.Background()

	run := NewSkillRun(SkillsFor(ctx, []string{"estimate"}))
	req.Contains(run.load(ctx, "other", "", nil), "not available")
	req.Contains(run.readResource(ctx, "estimate", "prices.csv"), "load skill", "must load before reading its files")

	run.load(ctx, "estimate", "", nil)
	req.Equal("wall,10", run.readResource(ctx, "estimate", "prices.csv"))
	req.Contains(run.readResource(ctx, "estimate", "nope.txt"), "no file")
	req.Contains(run.readResource(ctx, "estimate", "../x"), "no file")

	// an agent without skills has none of this
	a := &Agent{cfg: AgentConfig{Name: "plain"}}
	r, _, tools := a.startSkills(ctx, nil)
	req.Nil(r)
	req.Empty(tools)
}

func TestAgent_StartSkills(t *testing.T) {
	setupDemoSkill(t)
	req := require.New(t)
	ctx := context.Background()

	// catalog agent: meta tools and a catalog in the prompt, no extra tools yet
	a := &Agent{cfg: AgentConfig{Name: "universal", Skills: []string{"*"}}}
	run, pre, tools := a.startSkills(ctx, []chat.ToolDef{echoTool("base")})
	req.NotNil(run)
	req.Empty(pre)
	names := map[string]bool{}
	for _, t := range tools {
		names[t.Name] = true
	}
	req.True(names["base"] && names[ToolLoadSkill] && names[ToolListSkills] && names[ToolReadSkillResource])
	req.False(names["demo_read"])

	prompt := a.buildSystemPrompt(nil, nil, tools, run, pre)
	req.Contains(prompt, "estimate: When estimating work")
	req.NotContains(prompt, "Count the walls first.", "the body is loaded on demand")

	// a skill the caller chose: text in the prompt, tools from turn one
	a = &Agent{cfg: AgentConfig{Name: "step"}}
	run, pre, tools = a.startSkills(ContextWithSkill(ctx, "estimate@3"), nil)
	req.Contains(pre, "Count the walls first.")
	names = map[string]bool{}
	for _, t := range tools {
		names[t.Name] = true
	}
	req.True(names["demo_read"])
	req.Equal([]string{"estimate@3"}, run.Loaded())

	// a skill that does not exist is told to the model, not hidden
	_, pre, _ = a.startSkills(ContextWithSkill(ctx, "ghost"), nil)
	req.True(strings.Contains(pre, "ghost") && strings.Contains(pre, "could not be loaded"))
}

func TestAgent_ApprovableToolsIncludeSkillKits(t *testing.T) {
	setupDemoSkill(t)
	req := require.New(t)

	a := &Agent{cfg: AgentConfig{Name: "universal", Skills: []string{"*"}}}
	got, err := a.ExecApproved(context.Background(), []Call{{Name: "demo_write"}})
	req.NoError(err)
	req.Equal("written", got)

	a = &Agent{cfg: AgentConfig{Name: "noskills"}}
	_, err = a.ExecApproved(context.Background(), []Call{{Name: "demo_write"}})
	req.Error(err)
}

func TestAttachSkills_CatalogAndRestore(t *testing.T) {
	setupDemoSkill(t)
	req := require.New(t)
	ctx := context.Background()

	base := Options{
		Messages: []*schema.Message{schema.SystemMessage("You are helpful."), schema.UserMessage("hi")},
		Tools:    []chat.ToolDef{echoTool("base")},
	}

	// nothing to load: untouched
	req.Equal(len(base.Tools), len(AttachSkills(ctx, base, nil, nil).Tools))

	opt := AttachSkills(ctx, base, []string{"*"}, nil)
	req.NotNil(opt.Skills)
	req.Contains(opt.Messages[0].Content, "estimate: When estimating work")
	req.Equal("You are helpful.", base.Messages[0].Content, "the caller's messages are not changed")
	req.Len(opt.Tools, 4)

	// a later request confirming a call to a tool the skill brought: restored
	opt = AttachSkills(ctx, base, []string{"*"}, []string{"estimate"})
	names := map[string]bool{}
	for _, tl := range opt.Tools {
		names[tl.Name] = true
	}
	req.True(names["demo_write"])
	req.True(opt.NeedsConfirm([]Call{{Name: "demo_write"}}), "and it still needs the yes")
	req.Equal([]string{"estimate@3"}, opt.Skills.Loaded())

	// a skill the agent may not load is not restored
	opt = AttachSkills(ctx, base, []string{"other"}, []string{"estimate"})
	req.Nil(opt.Skills)
}

func TestRun_ReportsLoadedSkills(t *testing.T) {
	setupDemoSkill(t)
	m := &fakeModel{tools: true, gen: []genTurn{
		{content: `<tool name="load_skill"><param name="handle">estimate</param></tool>`},
		{content: "FINAL: done"},
	}}
	opt, run := skillOptions(t, m, false)
	opt.Skills = run
	out := Run(context.Background(), opt)
	require.Equal(t, []string{"estimate@3"}, out.Skills)
}
