package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/chat"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type (
	ChatMessage struct {
		Role    string
		Content string
	}

	ChatStreamFunc = chat.StreamFunc

	workflowChat struct {
		client *chat.Client
		tools  []chat.ToolDef
	}

	WorkflowChatPromptArguments struct {
		Session  string
		Prompt   string
		Facts    []string
		Messages []ChatMessage
		Workflow uint64
		Trigger  uint64
	}
)

func WorkflowChat() *workflowChat {
	c, err := chat.NewClient(chat.ModelForRole(chat.RoleAutomationChat))
	if err != nil {
		return nil
	}

	tools := []chat.ToolDef{
		{
			Name:        "create_workflow",
			Description: "Create a new workflow with steps and paths",
			Params: []chat.ParamDef{
				{Name: "name", Type: "string", Required: true, Description: "Workflow name (goes in meta.name, handle is auto-generated)"},
				{Name: "handle", Type: "string", Required: false, Description: "URL-safe handle for the workflow"},
				{Name: "description", Type: "string", Required: false, Description: "Workflow description (goes in meta.description)"},
				{Name: "enabled", Type: "string", Required: false, Description: "Whether the workflow is enabled (true/false, default false)"},
				{Name: "steps", Type: "json", Required: false, Description: `JSON array of workflow step objects. Each step: {"stepID":"string","kind":"string","ref":"string","arguments":[{"name":"...","expr":"..."}],"results":[{"name":"...","expr":"..."}],"meta":{"name":"...","description":"..."}}`},
				{Name: "paths", Type: "json", Required: false, Description: `JSON array of workflow path objects (connections between steps). Each path: {"parentID":"string","childID":"string","expr":"...","meta":{"name":"..."}}`},
			},
			Handler: createWorkflowHandler,
		},
	}

	return &workflowChat{
		client: c,
		tools:  tools,
	}
}

func (c *workflowChat) buildMessages(ask *WorkflowChatPromptArguments) []*schema.Message {
	msgs := make([]*schema.Message, 0, len(ask.Messages)+2)

	hasSystem := false
	for _, m := range ask.Messages {
		role := schema.User
		if m.Role == "assistant" || m.Role == "system" {
			role = schema.RoleType(m.Role)
		}
		if m.Role == "system" {
			hasSystem = true
		}
		msgs = append(msgs, &schema.Message{Role: role, Content: m.Content})
	}
	if !hasSystem {
		msgs = append([]*schema.Message{
			schema.SystemMessage("You are an AI assistant. You can call tools by outputting XML like this:\n<tool name=\"tool_name\">\n<param name=\"param1\">value1</param>\n</tool>\n\nAlways ask the user to confirm before creating anything. For listing/viewing tools, call them immediately without asking for confirmation."),
		}, msgs...)
	}
	if ask.Prompt != "" {
		msgs = append(msgs, schema.UserMessage(ask.Prompt))
	}
	return msgs
}

func (c *workflowChat) Ask(ctx context.Context, ask *WorkflowChatPromptArguments) (interface{}, error) {
	if err := chat.EnsureWarm(ctx, c.client.Model()); err != nil {
		return nil, err
	}
	toolInfos, err := chat.ToToolInfos(c.tools)
	if err != nil {
		return nil, fmt.Errorf("failed to build tool infos: %w", err)
	}

	msgs := c.buildMessages(ask)
	var opts []model.Option
	if c.client.IsToolsSupported() {
		opts = append(opts, model.WithTools(toolInfos))
	}
	out, err := c.client.Generate(ctx, msgs, opts...)
	if err != nil {
		return nil, err
	}

	content := out.Content

	toolCalls := out.ToolCalls
	if len(toolCalls) == 0 && chat.HasToolCallsStr(content) {
		xmlCalls := chat.ParseToolCallsStr(content)
		toolCalls = make([]schema.ToolCall, len(xmlCalls))
		for i, xc := range xmlCalls {
			paramJSON, _ := json.Marshal(xc.Params)
			toolCalls[i] = schema.ToolCall{
				Function: schema.FunctionCall{
					Name:      xc.Name,
					Arguments: string(paramJSON),
				},
			}
		}
	}

	if len(toolCalls) > 0 {
		if !userConfirmedWorkflow(ask.Prompt) {
			return map[string]any{
				"response": content + "\n\n⚠️ **Обнаружено предлагаемое действие.** Напишите **«да»** чтобы подтвердить.",
			}, nil
		}
		parsed := make([]CallParamWorkflow, 0, len(toolCalls))
		for _, tc := range toolCalls {
			parsed = append(parsed, CallParamWorkflow{
				Name:   tc.Function.Name,
				Params: tc.Function.Arguments,
			})
		}
		result := execWorkflowToolCalls(ctx, parsed, c.tools)
		if result != "" {
			return map[string]any{"response": result}, nil
		}
	}

	return map[string]any{
		"response": content,
	}, nil
}

func (c *workflowChat) AskStream(ctx context.Context, ask *WorkflowChatPromptArguments, stream chat.StreamFunc) error {
	if err := chat.EnsureWarm(ctx, c.client.Model()); err != nil {
		return err
	}
	toolInfos, err := chat.ToToolInfos(c.tools)
	if err != nil {
		return fmt.Errorf("failed to build tool infos: %w", err)
	}

	msgs := c.buildMessages(ask)
	var opts []model.Option
	if c.client.IsToolsSupported() {
		opts = append(opts, model.WithTools(toolInfos))
	}
	streamReader, err := c.client.Stream(ctx, msgs, opts...)
	if err != nil {
		return err
	}
	defer streamReader.Close()

	var fullContent string
	var toolCalls []schema.ToolCall
	for {
		chunk, err := streamReader.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if chunk.Content != "" {
			fullContent += chunk.Content
			if err := stream(chunk.Content, "", false); err != nil {
				return err
			}
		}
		if chunk.ToolCalls != nil && len(chunk.ToolCalls) > 0 {
			toolCalls = append(toolCalls, chunk.ToolCalls...)
		}
		if len(chunk.ReasoningContent) > 0 {
			if err := stream("", chunk.ReasoningContent, false); err != nil {
				return err
			}
		}
	}

	if len(toolCalls) == 0 && chat.HasToolCallsStr(fullContent) {
		xmlCalls := chat.ParseToolCallsStr(fullContent)
		toolCalls = make([]schema.ToolCall, len(xmlCalls))
		for i, xc := range xmlCalls {
			paramJSON, _ := json.Marshal(xc.Params)
			toolCalls[i] = schema.ToolCall{
				Function: schema.FunctionCall{
					Name:      xc.Name,
					Arguments: string(paramJSON),
				},
			}
		}
	}

	if len(toolCalls) > 0 {
		if !userConfirmedWorkflow(ask.Prompt) {
			stream("\n\n⚠️ **Обнаружено предлагаемое действие.** Напишите **«да»** чтобы подтвердить.", "", false)

			return stream("", "", true)
		}
		parsed := make([]CallParamWorkflow, 0, len(toolCalls))
		for _, tc := range toolCalls {
			parsed = append(parsed, CallParamWorkflow{
				Name:   tc.Function.Name,
				Params: tc.Function.Arguments,
			})
		}
		result := execWorkflowToolCalls(ctx, parsed, c.tools)
		if result != "" {
			stream("\n\n"+result, "", false)
		}
		return stream("", "", true)
	}

	return stream("", "", true)
}

type CallParamWorkflow struct {
	Name   string
	Params string
}

func execWorkflowToolCalls(ctx context.Context, calls []CallParamWorkflow, tools []chat.ToolDef) string {
	var results []string
	for _, call := range calls {
		params := make(map[string]string)
		if call.Params != "" {
			raw := make(map[string]any)
			if err := json.Unmarshal([]byte(call.Params), &raw); err == nil {
				for k, v := range raw {
					params[k] = fmt.Sprintf("%v", v)
				}
			}
		}
		for _, t := range tools {
			if t.Name == call.Name {
				results = append(results, t.Handler(ctx, params))
				break
			}
		}
	}
	if len(results) > 0 {
		return "**Tool Results:**\n" + strings.Join(results, "\n")
	}
	return ""
}

func userConfirmedWorkflow(prompt string) bool {
	lower := strings.ToLower(strings.TrimSpace(prompt))
	switch lower {
	case "да", "yes", "y", "ok", "ок", "гоу", "do it", "создай", "подтверждаю", "confirm", "выполнить":
		return true
	}
	return false
}

func createWorkflowHandler(ctx context.Context, params map[string]string) string {
	name := params["name"]
	handle := params["handle"]
	description := params["description"]
	enabled := params["enabled"]

	if name == "" {
		return "Missing required parameter: name"
	}

	if handle == "" {
		handle = name
	}

	meta := &types.WorkflowMeta{
		Name:        name,
		Description: description,
	}

	var steps types.WorkflowStepSet
	if s := params["steps"]; s != "" {
		if err := json.Unmarshal([]byte(s), &steps); err != nil {
			return fmt.Sprintf("Invalid steps JSON: %v", err)
		}
	}

	var paths types.WorkflowPathSet
	if p := params["paths"]; p != "" {
		if err := json.Unmarshal([]byte(p), &paths); err != nil {
			return fmt.Sprintf("Invalid paths JSON: %v", err)
		}
	}

	wf := &types.Workflow{
		Handle: handle,
		Meta:   meta,
		Steps:  steps,
		Paths:  paths,
	}

	if enabled == "true" {
		wf.Enabled = true
	}

	created, err := DefaultWorkflow.Create(ctx, wf)
	if err != nil {
		return fmt.Sprintf("Failed to create workflow '%s': %v", name, err)
	}

	return fmt.Sprintf("✅ Workflow '%s' created! ID: %d, Handle: %s, Steps: %d", created.Meta.Name, created.ID, created.Handle, len(created.Steps))
}
