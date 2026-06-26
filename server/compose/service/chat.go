package service

import (
	"context"
	"fmt"
	"log"

	"github.com/madnikulin50/lowcode/server/pkg/actionlog"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/ollama/ollama/api"
)

type (
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
		Prompt    string
		Namespace uint64
		Module    uint64
		Page      uint64
		Record    uint64
	}

	chatSettings struct {
	}
)

func (c *chat) getClient() (*api.Client, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}

	messages := []api.Message{
		{
			Role:    "system",
			Content: "Provide very brief, concise responses",
		},
		{
			Role:    "user",
			Content: "Name some unusual animals",
		},
		{
			Role:    "assistant",
			Content: "Monotreme, platypus, echidna",
		},
		{
			Role:    "user",
			Content: "which of these is the most dangerous?",
		},
	}

	ctx := context.Background()
	req := &api.ChatRequest{
		Model:    "llama3.2",
		Messages: messages,
	}

	respFunc := func(resp api.ChatResponse) error {
		fmt.Print(resp.Message.Content)
		return nil
	}

	err = client.Chat(ctx, req, respFunc)
	if err != nil {
		log.Fatal(err)
	}
	return client, nil
}

func (c *chat) Ask(ctx context.Context, ask *ChatPromptArguments) (interface{}, error) {
	client, err := c.getClient()

	messages := []api.Message{
		{
			Role:    "user",
			Content: ask.Prompt,
		},
	}
	req := &api.ChatRequest{
		Model:    "llama3.2",
		Messages: messages,
	}
	out := ""
	respFunc := func(resp api.ChatResponse) error {
		out += resp.Message.Content
		return nil
	}

	err = client.Chat(ctx, req, respFunc)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"response": out,
	}, nil
}
