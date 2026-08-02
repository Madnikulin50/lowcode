package chat

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type (
	StreamFunc func(token string, reason string, done bool) error

	Client struct {
		cm    *ollama.ChatModel
		model string
	}
)

func ollamaURL() string {
	u := os.Getenv("OLLAMA_URL")
	if u == "" {
		u = "http://localhost:11434"
	}
	return u
}

func NewClient(model string) (*Client, error) {
	cm, err := ollama.NewChatModel(context.Background(), &ollama.ChatModelConfig{
		BaseURL: ollamaURL(),
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: 3 * time.Minute,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ollama chat model: %w", err)
	}
	return &Client{cm: cm, model: model}, nil
}

func (c *Client) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	return c.cm.Generate(ctx, messages, opts...)
}

func (c *Client) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return c.cm.Stream(ctx, messages, opts...)
}

func (c *Client) IsToolsSupported() bool {
	switch c.model {
	case "deepseek-v2":
		return false
	}
	return true
}

func (c *Client) Ask(ctx context.Context, system, user string) (string, error) {
	msgs := make([]*schema.Message, 0, 2)
	if system != "" {
		msgs = append(msgs, schema.SystemMessage(system))
	}
	msgs = append(msgs, schema.UserMessage(user))
	resp, err := c.cm.Generate(ctx, msgs)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

func (c *Client) AskWithMessages(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	return c.cm.Generate(ctx, messages)
}

func (c *Client) AskStream(ctx context.Context, system, user string, fn StreamFunc) error {
	msgs := make([]*schema.Message, 0, 2)
	if system != "" {
		msgs = append(msgs, schema.SystemMessage(system))
	}
	msgs = append(msgs, schema.UserMessage(user))
	return c.AskStreamWithMessages(ctx, msgs, fn)
}

func (c *Client) AskStreamWithMessages(ctx context.Context, messages []*schema.Message, fn StreamFunc) error {
	stream, err := c.cm.Stream(ctx, messages)
	if err != nil {
		return err
	}
	defer stream.Close()

	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return fn("", "", true)
			}
			return err
		}
		if err := fn(chunk.Content, chunk.ReasoningContent, chunk.ResponseMeta != nil && chunk.ResponseMeta.FinishReason != ""); err != nil {
			return err
		}
	}
}
