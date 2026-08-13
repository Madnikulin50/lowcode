package rest

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/madnikulin50/lowcode/server/compose/rest/request"
	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

type (
	chatPayload struct {
		Response string `json:"response"`
	}

	Chat struct {
		chat interface {
			Ask(ctx context.Context, params *service.ChatPromptArguments) (interface{}, error)
			AskStream(ctx context.Context, params *service.ChatPromptArguments, stream service.ChatStreamFunc) error
			Models(ctx context.Context) ([]string, error)
			ModelsInfo(ctx context.Context) (map[string]any, error)
			DiscoverModels(ctx context.Context) ([]string, error)
			WarmUp(ctx context.Context, model string) error
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
		&service.ChatPromptArguments{
			Chat:      r.ChatID,
			Prompt:    r.Prompt,
			Messages:  toChatMessages(r.Messages),
			Files:     toChatFiles(r.Files),
			Namespace: r.NamespaceID,
			Page:      r.PageID,
			Facts:     r.Facts,
			Model:     r.Model,
		})

	return ctrl.makePayload(ctx, res, err)
}

func (ctrl *Chat) AskStream(ctx context.Context, r *request.ChatAsk, stream service.ChatStreamFunc) error {
	return ctrl.chat.AskStream(ctx, &service.ChatPromptArguments{
		Prompt:    r.Prompt,
		Messages:  toChatMessages(r.Messages),
		Files:     toChatFiles(r.Files),
		Namespace: r.NamespaceID,
		Page:      r.PageID,
		Facts:     r.Facts,
		Model:     r.Model,
	}, stream)
}

func (ctrl *Chat) Models(ctx context.Context) (interface{}, error) {
	return ctrl.chat.ModelsInfo(ctx)
}

func (ctrl *Chat) DiscoverModels(ctx context.Context) (interface{}, error) {
	models, err := ctrl.chat.DiscoverModels(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(models))
	cfg := chat.CurrentConfig()
	for _, name := range models {
		out = append(out, map[string]any{
			"name": name,
			// Avoid per-model /api/show here — use allowlist so admin sync stays fast.
			"tools":   chat.ModelLikelySupportsTools(name),
			"enabled": cfg.IsModelEnabled(name),
			"roles":   cfg.RolesUsing(name),
		})
	}
	return map[string]any{
		"models":     out,
		"ollamaURL":  chat.EffectiveOllamaURL(),
		"ollamaFrom": ollamaURLSourceLabel(),
	}, nil
}

func ollamaURLSourceLabel() string {
	cfg := chat.CurrentConfig()
	if strings.TrimSpace(cfg.OllamaURL) != "" {
		return "settings"
	}
	if strings.TrimSpace(os.Getenv("OLLAMA_URL")) != "" {
		return "OLLAMA_URL"
	}
	if strings.TrimSpace(os.Getenv("OLLAMA_HOST")) != "" {
		return "OLLAMA_HOST"
	}
	return "default"
}

func (ctrl *Chat) WarmUp(ctx context.Context, r *request.ChatAsk) error {
	return ctrl.chat.WarmUp(ctx, r.Model)
}

func toChatMessages(in []request.ChatMessage) []service.ChatMessage {
	out := make([]service.ChatMessage, len(in))
	for i, m := range in {

		out[i] = service.ChatMessage{Role: m.Role, Content: m.Content}
	}
	return out
}

func toChatFiles(in []request.ChatFile) []service.ChatFile {
	out := make([]service.ChatFile, len(in))
	for i, f := range in {
		out[i] = service.ChatFile{Name: f.Name, Content: f.Content}
	}
	return out
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
