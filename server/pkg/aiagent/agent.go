package aiagent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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

func (a *Agent) Run(ctx context.Context, input string, contextData map[string]interface{}) (result *AgentResult) {
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
	opt := Options{
		Client:      cl,
		Messages:    []*schema.Message{schema.SystemMessage(systemPrompt), schema.UserMessage(input)},
		Tools:       a.resolveTools(),
		MaxSteps:    a.cfg.MaxSteps,
		ExtraParams: extraFromContext(contextData),
		Continue:    AgentContinue,
		Validator:   a.cfg.Validator,
	}
	if a.cfg.Confirm {
		opt.NeedsConfirm = DefaultNeedsConfirm
	}
	result = Run(ctx, opt)
	return result
}

func (a *Agent) clientForRun() ChatModel {
	want := strings.TrimSpace(a.cfg.Model)
	if want == "" {
		want = chat.ModelForRole(chat.RoleMCPAgent)
	}
	if a.client != nil && a.client.Model() == want {
		return a.client
	}
	cl, err := chat.NewClient(want)
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
