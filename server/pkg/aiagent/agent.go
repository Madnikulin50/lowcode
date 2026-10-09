package aiagent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

type AgentConfig struct {
	Name         string
	Description  string
	SystemPrompt string
	Model        string
	Tools        []chat.ToolDef
	Toolkits     []string
	Skills       []string // handles the agent may load; "*" for all
	MaxSteps     int
	Confirm      bool
	Validator    func(result *AgentResult) error
}

type Agent struct {
	cfg    AgentConfig
	client *chat.Client
}

type AgentResult struct {
	Success       bool        `json:"success"`
	Output        string      `json:"output"`
	Steps         []AgentStep `json:"steps"`
	Error         string      `json:"error,omitempty"`
	Duration      string      `json:"duration"`
	ConfirmNeeded bool        `json:"confirmNeeded,omitempty"`
	ConfirmCalls  []Call      `json:"confirmCalls,omitempty"`
	Err           error       `json:"-"`

	// What the run cost, for observability: the model that answered, how many
	// LLM turns it took and, when the model reports usage, the tokens spent.
	Model            string `json:"model,omitempty"`
	LLMCalls         int    `json:"llmCalls,omitempty"`
	PromptTokens     int    `json:"promptTokens,omitempty"`
	CompletionTokens int    `json:"completionTokens,omitempty"`

	// Skills loaded during the run, as "handle@version"
	Skills []string `json:"skills,omitempty"`
}

func (r *AgentResult) addTurn(usage *schema.TokenUsage) {
	r.LLMCalls++
	if usage != nil {
		r.PromptTokens += usage.PromptTokens
		r.CompletionTokens += usage.CompletionTokens
	}
}

type AgentStep struct {
	Type     string          `json:"type"` // plan, execute, validate, respond
	Input    string          `json:"input"`
	Output   string          `json:"output"`
	Tools    []chat.ToolCall `json:"tools,omitempty"`
	Duration string          `json:"duration"`
}

func New(client *chat.Client, cfg AgentConfig) *Agent {
	if cfg.MaxSteps <= 0 {
		cfg.MaxSteps = 5
	}
	if cfg.Model == "" {
		cfg.Model = chat.ModelForRole(chat.RoleMCPAgent)
	}
	return &Agent{
		cfg:    cfg,
		client: client,
	}
}

func (a *Agent) Name() string        { return a.cfg.Name }
func (a *Agent) Description() string { return a.cfg.Description }

// Run executes the agent without allowing any mutating tool call to go
// through unconfirmed - equivalent to RunConfirmed(ctx, input, contextData,
// false). Use this from any non-interactive caller (a scheduled job, a
// rulechain node, ...): there's no human to type "да", so a mutating call
// the agent attempts is reported via AgentResult.ConfirmNeeded/ConfirmCalls
// instead of executing. Callers that intend to allow it (e.g. a rulechain
// node with an explicit allowMutating:true) should call RunConfirmed with
// confirmed=true instead.
func (a *Agent) Run(ctx context.Context, input string, contextData map[string]interface{}) (result *AgentResult) {
	return a.RunConfirmed(ctx, input, contextData, false)
}

// RunConfirmed is like Run, but confirmed stands in for the interactive "да"
// a chat user would otherwise have to type before a mutating tool call
// executes (see aiagent.NeedsConfirmFromToolDefs and Options.Confirmed).
// Every tool call is checked against its declared chat.ToolDef.Mutating flag
// regardless of a.cfg.Confirm - so an agent has exactly as much power as its
// caller explicitly grants, never silently more, no matter which surface
// (chat, MCP, or a rulechain ai/ai.operation node) is driving it.
func (a *Agent) RunConfirmed(ctx context.Context, input string, contextData map[string]interface{}, confirmed bool) (result *AgentResult) {
	start := time.Now()
	result = &AgentResult{Steps: make([]AgentStep, 0)}

	defer func() {
		if rec := recover(); rec != nil {
			if result == nil {
				result = &AgentResult{}
			}
			result.Error = fmt.Sprintf("agent panic: %v", rec)
		}
		if result != nil && result.Duration == "" {
			result.Duration = time.Since(start).String()
		}
	}()

	cl := a.clientForRun(ctx)
	tools := a.resolveTools()

	// skills: the agent gets a catalog and the means to load from it; a skill
	// the caller chose is loaded up front
	skillRun, preloaded, tools := a.startSkills(ctx, tools)
	if skillRun != nil {
		ctx = ContextWithSkillRun(ctx, skillRun)
	}
	systemPrompt := a.buildSystemPrompt(cl, contextData, tools, skillRun, preloaded)
	opt := Options{
		Client:       cl,
		Messages:     []*schema.Message{schema.SystemMessage(systemPrompt), schema.UserMessage(input)},
		Tools:        tools,
		MaxSteps:     a.cfg.MaxSteps,
		ExtraParams:  extraFromContext(contextData),
		Continue:     AgentContinue,
		Validator:    a.cfg.Validator,
		NeedsConfirm: NeedsConfirmFromToolDefs(tools),
		Confirmed:    confirmed,
		Skills:       skillRun,
	}
	if stream := streamFromContext(ctx); stream != nil {
		// someone is watching: generate as a stream, without the tool-call XML
		opt.Stream = stream
		opt.HideToolXML = true
	}
	result = Run(ctx, opt)
	if result != nil {
		result.Skills = skillRun.Loaded()
	}
	return result
}

// startSkills sets up skills for a run: the catalog the agent may use, the
// tools to load from it, and the skill the caller picked, if any. Returns nil
// when the run has no skills at all.
func (a *Agent) startSkills(ctx context.Context, tools []chat.ToolDef) (*SkillRun, string, []chat.ToolDef) {
	chosen := SkillFromContext(ctx)
	if len(a.cfg.Skills) == 0 && chosen == "" {
		return nil, "", tools
	}

	run := NewSkillRun(SkillsFor(ctx, a.cfg.Skills))
	var preloaded string
	if chosen != "" {
		text, extra, err := run.Preload(ctx, chosen, tools)
		if err != nil {
			preloaded = fmt.Sprintf("(the skill %q could not be loaded: %v)", chosen, err)
		} else {
			preloaded = text
			tools = Flatten(ToolKit{Name: "_base", Tools: tools}, ToolKit{Name: "_skill", Tools: extra})
		}
	}

	if len(run.catalog) > 0 {
		var current []chat.ToolDef
		current = tools
		meta := SkillTools(run, func() []chat.ToolDef { return current })
		tools = Flatten(ToolKit{Name: "_base", Tools: tools}, ToolKit{Name: "_skills", Tools: meta})
		current = tools
	}
	return run, preloaded, tools
}

func (a *Agent) clientForRun(ctx context.Context) ChatModel {
	// a.cfg.Model may be a literal Ollama model name (agent config a user
	// typed directly) or a role alias like "mcp.agent" (what every built-in
	// spec under defs/*.yaml uses) - ResolveModel tells the two apart and
	// only resolves the latter, so a call never hits a model literally
	// named after the role (see ResolveModel's doc comment).
	want := chat.ResolveModel(a.cfg.Model)
	if override := modelFromContext(ctx); override != "" {
		// a step or node asked for a specific model for this call only
		want = chat.ResolveModel(override)
	}
	if a.client != nil && a.client.Model() == want {
		return a.client
	}
	cl, err := chat.NewClientNoThink(want)
	if err != nil {
		if a.client != nil {
			return a.client
		}
		return nil
	}
	return cl
}

func (a *Agent) resolveTools() []chat.ToolDef {
	if len(a.cfg.Toolkits) == 0 {
		return a.cfg.Tools
	}
	resolved := DefaultCatalog().Resolve(a.cfg.Toolkits...)
	if len(a.cfg.Tools) == 0 {
		return resolved
	}
	return Flatten(ToolKit{Name: "_cfg", Tools: a.cfg.Tools}, ToolKit{Name: "_kit", Tools: resolved})
}

// ExecApproved runs tool calls a human has approved - typically the
// ConfirmCalls an earlier RunConfirmed(.., false) turn proposed and then
// stopped on. Only tools this agent may use are executable, so an approval
// payload that was tampered with in transit cannot reach arbitrary tools.
func (a *Agent) ExecApproved(ctx context.Context, calls []Call) (string, error) {
	tools := a.approvableTools(ctx, a.resolveTools())
	allowed := make(map[string]struct{}, len(tools))
	for _, t := range tools {
		allowed[t.Name] = struct{}{}
	}
	for _, c := range calls {
		if _, ok := allowed[c.Name]; !ok {
			return "", fmt.Errorf("tool %q is not available to agent %q", c.Name, a.Name())
		}
	}
	return ExecCalls(ctx, calls, tools, nil), nil
}

func (a *Agent) buildSystemPrompt(cl ChatModel, contextData map[string]interface{}, tools []chat.ToolDef, skills *SkillRun, preloaded string) string {
	prompt := a.cfg.SystemPrompt
	if cat := skills.CatalogPrompt(); cat != "" {
		prompt += "\n\n" + cat
	}
	if preloaded != "" {
		prompt += "\n\n## Skill to follow\n" + preloaded
	}
	useTools := cl != nil && cl.IsToolsSupported() && len(tools) > 0

	if useTools {
		prompt += "\n\n" + chat.ToolSystemPrompt(tools)
	}

	if len(contextData) > 0 {
		dataStr, _ := json.MarshalIndent(contextData, "", "  ")
		prompt += fmt.Sprintf("\n\n## Context Data\n```json\n%s\n```\n", string(dataStr))
	}

	if useTools {
		prompt += fmt.Sprintf(`
## Instructions
You are an AI agent named "%s".
%s

Think step by step. If you need to use tools, call them using XML format.

When you have a final answer, start your response with "FINAL:".
`, a.cfg.Name, a.cfg.Description)
	} else {
		prompt += fmt.Sprintf(`
## Instructions
You are an AI agent named "%s".
%s

When you have a final answer, start your response with "FINAL:".
`, a.cfg.Name, a.cfg.Description)
	}

	return prompt
}
