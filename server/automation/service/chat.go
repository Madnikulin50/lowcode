package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/chat"

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

func (c *workflowChat) buildMessages(ask *WorkflowChatPromptArguments, useTools bool) []*schema.Message {
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
		sys := "You are an AI assistant."
		if useTools {
			sys = "You are an AI assistant. You can call tools by outputting XML like this:\n<tool name=\"tool_name\">\n<param name=\"param1\">value1</param>\n</tool>\n\nAlways ask the user to confirm before creating anything. For listing/viewing tools, call them immediately without asking for confirmation."
		}
		msgs = append([]*schema.Message{
			schema.SystemMessage(sys),
		}, msgs...)
	}
	if ask.Prompt != "" {
		msgs = append(msgs, schema.UserMessage(ask.Prompt))
	}
	return msgs
}

func (c *workflowChat) runtimeOpts(ask *WorkflowChatPromptArguments, stream chat.StreamFunc) aiagent.Options {
	useTools := c.client.IsToolsSupported()
	tools := []chat.ToolDef(nil)
	if useTools {
		tools = c.tools
	}
	return aiagent.Options{
		Client:       c.client,
		Messages:     c.buildMessages(ask, useTools),
		Tools:        tools,
		MaxSteps:     4,
		Confirmed:    aiagent.UserConfirmed(ask.Prompt),
		NeedsConfirm: aiagent.DefaultNeedsConfirm,
		Continue: func(_ string, toolResult string, _ []aiagent.Call) aiagent.ContinueHint {
			return aiagent.ContinueHint{
				UserMessage:  "Tool results:\n" + toolResult + "\n\nWrite a short confirmation in the user's language. Do not call tools.",
				DisableTools: true,
			}
		},
		Stream:      stream,
		HideToolXML: stream != nil,
		EmptyAnswer: "Модель не сгенерировала ответ.",
	}
}

func (c *workflowChat) Ask(ctx context.Context, ask *WorkflowChatPromptArguments) (interface{}, error) {
	out := aiagent.Run(ctx, c.runtimeOpts(ask, nil))
	if out.Err != nil {
		return nil, out.Err
	}
	if out.ConfirmNeeded {
		return map[string]any{
			"response": out.Output + "\n\n⚠️ **Обнаружено предлагаемое действие.** Напишите **«да»** чтобы подтвердить.",
		}, nil
	}
	return map[string]any{"response": out.Output}, nil
}

func (c *workflowChat) AskStream(ctx context.Context, ask *WorkflowChatPromptArguments, stream chat.StreamFunc) error {
	out := aiagent.Run(ctx, c.runtimeOpts(ask, stream))
	if out.Err != nil {
		return out.Err
	}
	if out.ConfirmNeeded {
		if err := stream("\n\n⚠️ **Обнаружено предлагаемое действие.** Напишите **«да»** чтобы подтвердить.", "", false); err != nil {
			return err
		}
	}
	return stream("", "", true)
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
