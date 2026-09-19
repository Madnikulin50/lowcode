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

	cl := a.clientForRun()
	systemPrompt := a.buildSystemPrompt(cl, contextData)
	tools := a.resolveTools()
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
	}
	result = Run(ctx, opt)
	return result
}

func (a *Agent) clientForRun() ChatModel {
	// a.cfg.Model may be a literal Ollama model name (agent config a user
	// typed directly) or a role alias like "mcp.agent" (what every built-in
	// spec under defs/*.yaml uses) - ResolveModel tells the two apart and
	// only resolves the latter, so a call never hits a model literally
	// named after the role (see ResolveModel's doc comment).
	want := chat.ResolveModel(a.cfg.Model)
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

func (a *Agent) buildSystemPrompt(cl ChatModel, contextData map[string]interface{}) string {
	prompt := a.cfg.SystemPrompt
	tools := a.resolveTools()
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
