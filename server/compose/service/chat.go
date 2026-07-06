package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/actionlog"
	"github.com/madnikulin50/lowcode/server/pkg/xml_to_map"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/ollama/ollama/api"
	"github.com/ollama/ollama/envconfig"
)

type (
	ChatStreamFunc func(token string, done bool) error

	chat struct {
		actionlog actionlog.Recorder
		ac        chatAccessController
		eventbus  eventDispatcher
		store     store.Storer
		locale    ResourceTranslationsManagerService

		chatSettings *chatSettings
	}

	chatAccessController interface {
		CanAsk(ctx context.Context) bool
	}

	ChatPromptArguments struct {
		Session   string
		Prompt    string
		Facts     []string
		Namespace uint64
		Module    uint64
		Page      uint64
		Record    uint64
	}

	chatSettings struct {
	}
)

func (c *chat) getClient(needStart bool) (*api.Client, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}

	messages := []api.Message{}
	if needStart {
		messages = append(messages, api.Message{
			Role:    "system",
			Content: "Provide very brief, concise responses",
		}, api.Message{
			Role:    "user",
			Content: "Name some unusual animals",
		}, api.Message{
			Role:    "assistant",
			Content: "Monotreme, platypus, echidna",
		}, api.Message{
			Role:    "user",
			Content: "which of these is the most dangerous?",
		})
		ctx := context.Background()
		req := &api.ChatRequest{
			Model:    "deepseek-v2",
			Messages: messages,
		}

		respFunc := func(resp api.ChatResponse) error {
			fmt.Print(resp.Message.Content)
			return nil
		}

		err = client.Chat(ctx, req, respFunc)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

func (c *chat) Ask(ctx context.Context, ask *ChatPromptArguments) (interface{}, error) {

	needStartInstructions := true
	if strings.Contains(ask.Prompt, "<system_prompt>") {
		needStartInstructions = false
	}

	client, err := c.getClient(needStartInstructions)
	if err != nil {
		return nil, err
	}
	messages := []api.Message{
		{
			Role:    "user",
			Content: ask.Prompt,
		},
	}
	model := "deepseek-v2"
	data, err := xml_to_map.ParseXMLToMap(strings.NewReader(ask.Prompt))
	if err == nil {
		m, ok := data["model"]
		if ok {
			model = m
		}
	}

	req := &api.ChatRequest{
		//Model: "llama3.2",
		//Model:    "Qwen3.5",
		Model:    model,
		Messages: messages,
		Options: map[string]any{
			"num_predict": 16384,
		},
	}
	out := ""
	respFunc := func(resp api.ChatResponse) error {
		out += resp.Message.Content
		return nil
	}

	err = client.Chat(ctx, req, respFunc)
	if err != nil && len(out) == 0 {
		return nil, err
	}

	return map[string]any{
		"response": out,
	}, nil
}

func (c *chat) getModel(ask *ChatPromptArguments) string {
	model := "deepseek-v2"
	data, err := xml_to_map.ParseXMLToMap(strings.NewReader(ask.Prompt))
	if err == nil {
		m, ok := data["model"]
		if ok {
			model = m
		}
	}
	return model
}

func (c *chat) AskStream(ctx context.Context, ask *ChatPromptArguments, stream ChatStreamFunc) error {
	client := api.NewClient(envconfig.Host(), &http.Client{
		Timeout: 3 * time.Minute,
	})

	if client == nil {
		return fmt.Errorf("unable to create client")
	}
	streaming := true
	messages := []api.Message{
		{Role: "user", Content: ask.Prompt},
	}

	req := &api.ChatRequest{
		Model:    c.getModel(ask),
		Messages: messages,
		Stream:   &streaming,
		Options: map[string]any{
			"num_predict": 16384,
		},
	}
	chatCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	return client.Chat(chatCtx, req, func(resp api.ChatResponse) error {
		if resp.Done {
			cancel()
			fmt.Print("done: %s\r\n", resp.DoneReason)
		}
		return stream(resp.Message.Content, resp.Done)
	})
}
