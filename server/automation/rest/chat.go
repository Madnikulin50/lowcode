package rest

import (
	"context"

	"github.com/madnikulin50/lowcode/server/automation/rest/request"
	"github.com/madnikulin50/lowcode/server/automation/service"
)

type (
	Chat struct {
		chat interface {
			Ask(ctx context.Context, params *service.WorkflowChatPromptArguments) (interface{}, error)
			AskStream(ctx context.Context, params *service.WorkflowChatPromptArguments, stream service.ChatStreamFunc) error
		}
	}
)

func (Chat) New() *Chat {
	return &Chat{
		chat: service.DefaultWorkflowChat,
	}
}

func (ctrl *Chat) Ask(ctx context.Context, r *request.ChatAsk) (interface{}, error) {
	return ctrl.chat.Ask(ctx, &service.WorkflowChatPromptArguments{
		Prompt:   r.Prompt,
		Messages: toWorkflowChatMessages(r.Messages),
		Workflow: r.WorkflowID,
		Facts:    r.Facts,
	})
}

func (ctrl *Chat) AskStream(ctx context.Context, r *request.ChatAsk, stream service.ChatStreamFunc) error {
	return ctrl.chat.AskStream(ctx, &service.WorkflowChatPromptArguments{
		Prompt:   r.Prompt,
		Messages: toWorkflowChatMessages(r.Messages),
		Workflow: r.WorkflowID,
		Facts:    r.Facts,
	}, stream)
}

func toWorkflowChatMessages(in []request.ChatMessage) []service.ChatMessage {
	out := make([]service.ChatMessage, len(in))
	for i, m := range in {
		out[i] = service.ChatMessage{Role: m.Role, Content: m.Content}
	}
	return out
}
