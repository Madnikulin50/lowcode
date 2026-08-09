package chat

import (
	"bytes"
	"context"
	"encoding/json"
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
		KeepAlive: func() *time.Duration {
			d := 30 * time.Minute
			return &d
		}(),
		HTTPClient: &http.Client{
			Timeout: 3 * time.Minute,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ollama chat model: %w", err)
	}
	return &Client{cm: cm, model: model}, nil
}

// WarmUp forces Ollama to load the model into memory so that the first
// real chat message is not delayed by model loading. num_predict is limited
// to 1 token so the request completes in seconds and does not occupy
// Ollama's serial inference queue.
func WarmUp(model string) error {
	payload, _ := json.Marshal(map[string]any{
		"model":      model,
		"messages":   []map[string]string{{"role": "user", "content": "hi"}},
		"stream":     false,
		"keep_alive": "30m",
		"options": map[string]any{
			"num_predict": 1,
		},
	})
	resp, err := http.Post(ollamaURL()+"/api/chat", "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to warm up model %s: %w", model, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to warm up model %s: status %d", model, resp.StatusCode)
	}
	return nil
}

func AvailableModels() ([]string, error) {
	resp, err := http.Get(ollamaURL() + "/api/tags")
	if err != nil {
		return nil, fmt.Errorf("failed to list ollama models: %w", err)
	}
	defer resp.Body.Close()

	var tags struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, fmt.Errorf("failed to decode ollama models: %w", err)
	}

	names := make([]string, 0, len(tags.Models))
	for _, m := range tags.Models {
		ok, err := supportsChat(m.Name)
		if err != nil {
			continue
		}
		if ok {
			names = append(names, m.Name)
		}
	}
	return names, nil
}

func supportsChat(name string) (bool, error) {
	payload, _ := json.Marshal(map[string]string{"model": name})
	resp, err := http.Post(ollamaURL()+"/api/show", "application/json", bytes.NewReader(payload))
	if err != nil {
		return false, fmt.Errorf("failed to inspect model %s: %w", name, err)
	}
	defer resp.Body.Close()

	var show struct {
		Capabilities []string `json:"capabilities"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&show); err != nil {
		return false, fmt.Errorf("failed to decode model %s: %w", name, err)
	}

	for _, c := range show.Capabilities {
		if c == "completion" {
			return true, nil
		}
	}
	return false, nil
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
	case "deepseek-r1":
		return true
	}
	return false
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
