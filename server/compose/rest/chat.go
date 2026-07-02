package rest

import (
	"context"
	"fmt"
	"strings"

	"github.com/madnikulin50/lowcode/server/compose/rest/request"
	"github.com/madnikulin50/lowcode/server/compose/service"
)

type (
	chatPayload struct {
		Response string `json:"response"`
	}

	Chat struct {
		chat interface {
			Ask(ctx context.Context, params *service.ChatPromptArguments) (interface{}, error)
		}
		namespace service.NamespaceService
	}
)

func (Chat) New() *Chat {
	return &Chat{
		chat:      service.DefaultChat,
		namespace: service.DefaultNamespace,
	}
}

func (ctrl *Chat) Ask(ctx context.Context, r *request.ChatAsk) (interface{}, error) {
	res, err := ctrl.chat.Ask(ctx,
		&service.ChatPromptArguments{Prompt: r.Prompt, Namespace: r.NamespaceID, Page: r.PageID, Facts: r.Facts})

	return ctrl.makePayload(ctx, res, err)
}

func (ctrl Chat) makePayload(ctx context.Context, itf interface{}, err error) (*chatPayload, error) {
	if err != nil {
		return nil, err
	}
	m, ok := itf.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid payload type: %T", itf)
	}
	response := m["response"].(string)
	response = strings.Replace(response, "\n", "\r\n", -1)
	p := &chatPayload{Response: response}

	return p, nil
}
