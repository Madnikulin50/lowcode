package service

import (
	"context"
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/drivers/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakePromptAC struct{ read, write bool }

func (a fakePromptAC) CanSearchWorkflows(context.Context) bool { return a.read }
func (a fakePromptAC) CanCreateWorkflow(context.Context) bool  { return a.write }

func newPromptLib(t *testing.T, ac promptAccessController) (*promptLibrary, context.Context) {
	t.Helper()
	ctx := auth.SetIdentityToContext(context.Background(), auth.Authenticated(5))

	st, err := sqlite.ConnectInMemory(ctx)
	require.NoError(t, err)
	require.NoError(t, store.Upgrade(ctx, zap.NewNop(), st))
	// the in-memory database is shared by every test in the process
	require.NoError(t, st.TruncateAutomationPromptVersions(ctx))

	return PromptLibrary(st, ac), ctx
}

func TestPromptLibrary_VersionsAndRollback(t *testing.T) {
	lib, ctx := newPromptLib(t, fakePromptAC{true, true})
	req := require.New(t)

	// the first version is live even if not asked to be
	v1, err := lib.Save(ctx, SavePrompt{Handle: "triage", Text: "Rate urgency.", Description: "Ticket triage"})
	req.NoError(err)
	req.Equal(1, v1.Version)
	req.True(v1.Active)
	req.Equal(uint64(5), v1.CreatedBy)

	// a new version is saved but not live until activated: it can be tested first
	v2, err := lib.Save(ctx, SavePrompt{Handle: "triage", Text: "Rate urgency from 1-5.", Note: "numeric scale"})
	req.NoError(err)
	req.Equal(2, v2.Version)
	req.False(v2.Active)
	req.Equal("Ticket triage", v2.Description, "description carries over")

	active, err := lib.Get(ctx, "triage", 0)
	req.NoError(err)
	req.Equal(1, active.Version)

	// roll forward
	_, err = lib.Activate(ctx, "triage", 2)
	req.NoError(err)
	active, _ = lib.Get(ctx, "triage", 0)
	req.Equal(2, active.Version)

	// ...and back: the old version is untouched
	_, err = lib.Activate(ctx, "triage", 1)
	req.NoError(err)
	active, _ = lib.Get(ctx, "triage", 0)
	req.Equal("Rate urgency.", active.Text)

	// activate-on-save
	v3, err := lib.Save(ctx, SavePrompt{Handle: "triage", Text: "v3", Activate: true})
	req.NoError(err)
	req.True(v3.Active)

	hist, err := lib.History(ctx, "triage")
	req.NoError(err)
	req.Len(hist, 3)
	activeCount := 0
	for i, v := range hist {
		req.Equal(i+1, v.Version, "oldest first")
		if v.Active {
			activeCount++
		}
	}
	req.Equal(1, activeCount, "exactly one version is active")

	_, err = lib.Activate(ctx, "triage", 9)
	req.ErrorContains(err, "no version 9")
}

func TestPromptLibrary_ListAndDelete(t *testing.T) {
	lib, ctx := newPromptLib(t, fakePromptAC{true, true})
	req := require.New(t)

	_, err := lib.Save(ctx, SavePrompt{Handle: "b_prompt", Text: "B"})
	req.NoError(err)
	_, err = lib.Save(ctx, SavePrompt{Handle: "a_prompt", Text: "A1", Description: "first"})
	req.NoError(err)
	cases := types.PromptCases{{Name: "x", Expect: map[string]interface{}{"k": "v"}}}
	_, err = lib.Save(ctx, SavePrompt{Handle: "a_prompt", Text: "A2", Activate: true, Cases: &cases})
	req.NoError(err)

	list, err := lib.List(ctx)
	req.NoError(err)
	req.Len(list, 2)
	req.Equal("a_prompt", list[0].Handle, "sorted by handle")
	req.Equal(2, list[0].Versions)
	req.Equal(2, list[0].Active)
	req.Equal(1, list[0].Cases)
	req.Equal("first", list[0].Description)

	req.NoError(lib.Delete(ctx, "a_prompt"))
	list, _ = lib.List(ctx)
	req.Len(list, 1)

	_, err = lib.Get(ctx, "a_prompt", 0)
	req.ErrorContains(err, "does not exist")
	req.ErrorContains(lib.Delete(ctx, "a_prompt"), "does not exist")
}

func TestPromptLibrary_CasesCarryForward(t *testing.T) {
	lib, ctx := newPromptLib(t, fakePromptAC{true, true})
	req := require.New(t)

	cases := types.PromptCases{{Name: "one", Contains: []string{"x"}}}
	_, err := lib.Save(ctx, SavePrompt{Handle: "p", Text: "v1", Cases: &cases})
	req.NoError(err)

	v2, err := lib.Save(ctx, SavePrompt{Handle: "p", Text: "v2"}) // nil: keep
	req.NoError(err)
	req.Len(v2.Cases, 1, "cases carry forward so versions are judged on the same examples")

	none := types.PromptCases{}
	v3, err := lib.Save(ctx, SavePrompt{Handle: "p", Text: "v3", Cases: &none}) // explicit empty: clear
	req.NoError(err)
	req.Empty(v3.Cases)
}

func TestPromptLibrary_Validation(t *testing.T) {
	lib, ctx := newPromptLib(t, fakePromptAC{true, true})
	req := require.New(t)

	for _, handle := range []string{"", "Upper", "1abc", "has space", strings.Repeat("a", 65), "a/b"} {
		_, err := lib.Save(ctx, SavePrompt{Handle: handle, Text: "x"})
		req.Error(err, "handle %q", handle)
	}

	_, err := lib.Save(ctx, SavePrompt{Handle: "ok", Text: "  \n"})
	req.ErrorContains(err, "empty")

	_, err = lib.Save(ctx, SavePrompt{Handle: "ok", Text: strings.Repeat("я", maxPromptLen+1)})
	req.ErrorContains(err, "longer")

	tooMany := make(types.PromptCases, maxPromptCases+1)
	_, err = lib.Save(ctx, SavePrompt{Handle: "ok", Text: "x", Cases: &tooMany})
	req.ErrorContains(err, "at most")

	_, err = lib.Save(context.Background(), SavePrompt{Handle: "ok", Text: "x"}) // no user
	req.Error(err)
}

func TestPromptLibrary_Permissions(t *testing.T) {
	ctx := auth.SetIdentityToContext(context.Background(), auth.Authenticated(5))
	reader, _ := newPromptLib(t, fakePromptAC{read: true})
	nobody, _ := newPromptLib(t, fakePromptAC{})

	_, err := reader.Save(ctx, SavePrompt{Handle: "p", Text: "x"})
	require.Error(t, err, "reading is not writing")
	require.Error(t, reader.Delete(ctx, "p"))
	_, err = reader.Activate(ctx, "p", 1)
	require.Error(t, err)

	_, err = reader.List(ctx)
	require.NoError(t, err)

	_, err = nobody.List(ctx)
	require.Error(t, err)
	_, err = nobody.Get(ctx, "p", 0)
	require.Error(t, err)
	_, err = nobody.History(ctx, "p")
	require.Error(t, err)
}

func TestPromptLibrary_ResolvesReferences(t *testing.T) {
	lib, ctx := newPromptLib(t, fakePromptAC{true, true})
	req := require.New(t)

	_, err := lib.Save(ctx, SavePrompt{Handle: "greet", Text: "Say hello."})
	req.NoError(err)
	_, err = lib.Save(ctx, SavePrompt{Handle: "greet", Text: "Say hello warmly."}) // saved, not active
	req.NoError(err)

	aiagent.SetPromptResolver(lib.resolve)
	defer aiagent.SetPromptResolver(nil)

	// a running workflow resolves without a user in the context
	bg := context.Background()

	text, ref, err := aiagent.ResolvePrompt(bg, "@prompt:greet")
	req.NoError(err)
	req.Equal("Say hello.", text)
	req.Equal("greet@1", ref)

	text, ref, err = aiagent.ResolvePrompt(bg, "@prompt:greet@2")
	req.NoError(err)
	req.Equal("Say hello warmly.", text, "pinned versions resolve even when not active")
	req.Equal("greet@2", ref)

	_, _, err = aiagent.ResolvePrompt(bg, "@prompt:greet@5")
	req.ErrorContains(err, "no version 5")
	_, _, err = aiagent.ResolvePrompt(bg, "@prompt:nope")
	req.ErrorContains(err, "does not exist")
}

func TestPromptLibrary_Eval_ComparesVersions(t *testing.T) {
	lib, ctx := newPromptLib(t, fakePromptAC{true, true})
	req := require.New(t)

	cases := types.PromptCases{
		{Name: "outage", Inputs: map[string]interface{}{"ticket": "server down"}, Expect: map[string]interface{}{"priority": "high"}},
		{Name: "jam", Inputs: map[string]interface{}{"ticket": "printer jam"}, Expect: map[string]interface{}{"priority": "low"}},
	}
	_, err := lib.Save(ctx, SavePrompt{Handle: "triage", Text: "VERSION-ONE: rate the ticket.", Cases: &cases})
	req.NoError(err)
	_, err = lib.Save(ctx, SavePrompt{Handle: "triage", Text: "VERSION-TWO: rate the ticket carefully."})
	req.NoError(err)

	// the "model" gets the second version right and the first only half right
	run := func(_ context.Context, _, prompt string, _ bool) (*aiagent.AgentResult, error) {
		switch {
		case strings.Contains(prompt, "VERSION-TWO") && strings.Contains(prompt, "server down"):
			return &aiagent.AgentResult{Success: true, Output: `{"priority":"high"}`}, nil
		case strings.Contains(prompt, "VERSION-TWO"):
			return &aiagent.AgentResult{Success: true, Output: `{"priority":"low"}`}, nil
		case strings.Contains(prompt, "server down"):
			return &aiagent.AgentResult{Success: true, Output: `{"priority":"low"}`}, nil // wrong
		default:
			return &aiagent.AgentResult{Success: true, Output: `{"priority":"low"}`}, nil
		}
	}

	reports, err := lib.Eval(ctx, run, "triage", []int{1, 2}, "", "", nil, nil)
	req.NoError(err)
	req.Len(reports, 2)
	req.Equal("triage@1", reports[1].PromptRef)
	req.Equal(1, reports[1].Passed)
	req.Equal(2, reports[2].Passed)
	req.Greater(reports[2].PassRate, reports[1].PassRate)

	// no versions named: the active one (version 1)
	reports, err = lib.Eval(ctx, run, "triage", nil, "", "", nil, nil)
	req.NoError(err)
	req.Contains(reports, 1)

	// ad-hoc cases replace the stored ones for the run
	reports, err = lib.Eval(ctx, run, "triage", []int{2}, "", "", types.PromptCases{
		{Name: "only", Inputs: map[string]interface{}{"ticket": "server down"}, Expect: map[string]interface{}{"priority": "high"}},
	}, nil)
	req.NoError(err)
	req.Equal(1, reports[2].Total)

	// nothing to evaluate against
	_, err = lib.Save(ctx, SavePrompt{Handle: "bare", Text: "x"})
	req.NoError(err)
	_, err = lib.Eval(ctx, run, "bare", nil, "", "", nil, nil)
	req.ErrorContains(err, "no test cases")
}

func TestPromptAPI_Parsing(t *testing.T) {
	req := require.New(t)

	v, err := ParsePromptVersions("")
	req.NoError(err)
	req.Empty(v)
	v, err = ParsePromptVersions("1, 3;5")
	req.NoError(err)
	req.Equal([]int{1, 3, 5}, v)
	for _, bad := range []string{"x", "0", "-1", "1,a"} {
		_, err = ParsePromptVersions(bad)
		req.Error(err, bad)
	}

	c, err := ParsePromptCases("")
	req.NoError(err)
	req.Nil(c, "nothing given is not the same as clear")
	c, err = ParsePromptCases("[]")
	req.NoError(err)
	req.NotNil(c)
	req.Empty(*c)
	c, err = ParsePromptCases(`[{"name":"a","inputs":{"k":1},"expect":{"p":"high"},"contains":["x"]}]`)
	req.NoError(err)
	req.Equal("a", (*c)[0].Name)
	_, err = ParsePromptCases(`{"name":"a"}`)
	req.Error(err)
}

func TestPromptAPI_EndToEnd(t *testing.T) {
	lib, ctx := newPromptLib(t, fakePromptAC{true, true})
	prev := DefaultPrompts
	DefaultPrompts = lib
	defer func() { DefaultPrompts = prev }()
	req := require.New(t)

	saved, err := SavePromptVersion(ctx, SavePrompt{Handle: "api_demo", Text: "Be brief.", Description: "demo"})
	req.NoError(err)
	req.Equal(1, saved.Version)
	req.Empty(saved.Text, "a save acknowledges, it does not echo the text")

	got, err := GetPrompt(ctx, "api_demo", 0)
	req.NoError(err)
	req.Equal("Be brief.", got.Text)

	_, err = SavePromptVersion(ctx, SavePrompt{Handle: "api_demo", Text: "Be very brief."})
	req.NoError(err)
	h, err := PromptHistory(ctx, "api_demo")
	req.NoError(err)
	req.Len(h, 2)
	req.True(h[0].Active)
	req.False(h[1].Active)

	act, err := ActivatePrompt(ctx, "api_demo", 2)
	req.NoError(err)
	req.True(act.Active)

	list, err := ListPrompts(ctx)
	req.NoError(err)
	req.Len(list, 1)

	req.NoError(DeletePrompt(ctx, "api_demo"))

	DefaultPrompts = nil
	_, err = ListPrompts(ctx)
	req.ErrorContains(err, "not available")
}

func TestPromptLibrary_Eval_ReportsEachCaseAsItIsJudged(t *testing.T) {
	lib, ctx := newPromptLib(t, fakePromptAC{true, true})
	req := require.New(t)

	cases := types.PromptCases{
		{Name: "one", Contains: []string{"ok"}},
		{Name: "two", Contains: []string{"ok"}},
	}
	_, err := lib.Save(ctx, SavePrompt{Handle: "obs", Text: "Say ok.", Cases: &cases})
	req.NoError(err)

	type seen struct {
		version int
		name    string
	}
	var order []seen
	run := func(context.Context, string, string, bool) (*aiagent.AgentResult, error) {
		// at the time the model is asked about case two, case one is already reported
		return &aiagent.AgentResult{Success: true, Output: "ok"}, nil
	}

	_, err = lib.Eval(ctx, run, "obs", nil, "", "", nil, func(v int, r aiagent.EvalResult) {
		order = append(order, seen{v, r.Case})
	})
	req.NoError(err)
	req.Equal([]seen{{1, "one"}, {1, "two"}}, order)
}
