package aiagent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	MaxSteps     int                             // max LLM ↔ tool round-trips
	Validator    func(result *AgentResult) error // post-execution validation
}

type Agent struct {
	cfg    AgentConfig
	client *chat.Client
}

type AgentResult struct {
	Success  bool        `json:"success"`
	Output   string      `json:"output"`
	Steps    []AgentStep `json:"steps"`
	Error    string      `json:"error,omitempty"`
	Duration string      `json:"duration"`
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

func (a *Agent) Run(ctx context.Context, input string, contextData map[string]interface{}) *AgentResult {
	start := time.Now()
	result := &AgentResult{Steps: make([]AgentStep, 0)}

	defer func() {
		if rec := recover(); rec != nil {
			result.Error = fmt.Sprintf("agent panic: %v", rec)
		}
		result.Duration = time.Since(start).String()
	}()

	systemPrompt := a.buildSystemPrompt(contextData)

	return a.runLoop(ctx, systemPrompt, input, contextData, result)
}

func (a *Agent) buildSystemPrompt(contextData map[string]interface{}) string {
	prompt := a.cfg.SystemPrompt

	if len(a.cfg.Tools) > 0 {
		prompt += "\n\n" + chat.ToolSystemPrompt(a.cfg.Tools)
	}

	if len(contextData) > 0 {
		dataStr, _ := json.MarshalIndent(contextData, "", "  ")
		prompt += fmt.Sprintf("\n\n## Context Data\n```json\n%s\n```\n", string(dataStr))
	}

	prompt += fmt.Sprintf(`
## Instructions
You are an AI agent named "%s".
%s

Think step by step. If you need to use tools, call them using XML format.

When you have a final answer, start your response with "FINAL:".
`, a.cfg.Name, a.cfg.Description)

	return prompt
}

func (a *Agent) runLoop(ctx context.Context, systemPrompt, input string, ctxData map[string]interface{}, result *AgentResult) *AgentResult {
	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(input),
	}

	for step := 0; step < a.cfg.MaxSteps; step++ {
		stepStart := time.Now()
		log.Printf("[aiagent:%s] step %d/%d", a.cfg.Name, step+1, a.cfg.MaxSteps)

		// PLAN + EXECUTE: LLM call with tools
		agentStep := AgentStep{Type: "execute"}
		resp, err := a.client.Generate(ctx, messages)
		if err != nil {
			result.Error = fmt.Sprintf("LLM error: %v", err)
			return result
		}

		agentStep.Output = resp.Content

		// Check for tool calls (native or XML)
		var toolCalls []chat.ToolCall

		if chat.HasToolCalls(resp) {
			for _, tc := range chat.ParseToolCalls(resp) {
				toolCalls = append(toolCalls, chat.ToolCall{Name: tc.Function.Name, Params: mapFromJSON(tc.Function.Arguments)})
			}
		} else if chat.HasToolCallsStr(resp.Content) {
			toolCalls = chat.ParseToolCallsStr(resp.Content)
		}

		if len(toolCalls) > 0 {
			agentStep.Type = "execute"
			agentStep.Tools = toolCalls

			// Add assistant message with tool calls
			assistantMsg := resp
			messages = append(messages, assistantMsg)

			// Execute each tool
			var toolResults []string
			for _, tc := range toolCalls {
				result := a.executeTool(ctx, tc, ctxData)
				toolResults = append(toolResults, fmt.Sprintf("[%s]: %s", tc.Name, result))
			}

			// Add tool results as messages
			toolResultStr := strings.Join(toolResults, "\n")
			messages = append(messages, schema.UserMessage("Tool results:\n"+toolResultStr+"\n\nContinue with next step or provide final answer."))

			agentStep.Duration = time.Since(stepStart).String()
			result.Steps = append(result.Steps, agentStep)
			continue
		}

		// No tool calls - check for final answer
		agentStep.Duration = time.Since(stepStart).String()
		result.Steps = append(result.Steps, agentStep)

		if strings.HasPrefix(strings.TrimSpace(resp.Content), "FINAL:") {
			result.Output = strings.TrimPrefix(resp.Content, "FINAL:")
			result.Output = strings.TrimSpace(result.Output)
		} else {
			result.Output = resp.Content
		}
		result.Success = true

		// VALIDATE
		if a.cfg.Validator != nil {
			validateStep := AgentStep{Type: "validate", Input: result.Output}
			vStart := time.Now()
			if err := a.cfg.Validator(result); err != nil {
				result.Success = false
				result.Error = err.Error()
				validateStep.Output = "FAILED: " + err.Error()

				// Ask LLM to fix
				messages = append(messages, schema.UserMessage(fmt.Sprintf(
					"Validation failed: %s\nPlease fix and try again.", err.Error())))
				validateStep.Duration = time.Since(vStart).String()
				result.Steps = append(result.Steps, validateStep)
				continue
			}
			validateStep.Output = "OK"
			validateStep.Duration = time.Since(vStart).String()
			result.Steps = append(result.Steps, validateStep)
		}

		return result
	}

	result.Error = fmt.Sprintf("max steps (%d) exceeded", a.cfg.MaxSteps)
	return result
}

func (a *Agent) executeTool(ctx context.Context, tc chat.ToolCall, ctxData map[string]interface{}) string {
	for _, def := range a.cfg.Tools {
		if def.Name == tc.Name {
			params := tc.Params
			if params == nil {
				params = make(map[string]string)
			}
			for k, v := range ctxData {
				if _, ok := params[k]; !ok {
					params[k] = fmt.Sprintf("%v", v)
				}
			}
			return def.Handler(ctx, params)
		}
	}
	return fmt.Sprintf("unknown tool: %s", tc.Name)
}

func mapFromJSON(jsonStr string) map[string]string {
	if jsonStr == "" {
		return nil
	}
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return map[string]string{"_raw": jsonStr}
	}
	result := make(map[string]string)
	for k, v := range raw {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}
