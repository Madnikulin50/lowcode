package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

const (
	// DefaultModel is the preferred Ollama model for chat / MCP tool orchestration.
	// Override with CHAT_MODEL.
	DefaultModel = "qwen3:8b"
)

type (
	StreamFunc func(token string, reason string, done bool) error

	Client struct {
		cm    *ollama.ChatModel
		model string
	}
)

var (
	toolsCapMu    sync.RWMutex
	toolsCapCache = map[string]bool{}
)

// Known families that support Ollama native tool calling when /api/show is
// unavailable or does not report capabilities (older Ollama).
var toolModelPrefixes = []string{
	"llama3-groq-tool-use",
	"command-r-plus",
	"command-r",
	"qwen2.5",
	"qwen3",
	"qwen2",
	"llama3.1",
	"llama3.2",
	"llama3.3",
	"deepseek-r1",
	"mistral",
	"mixtral",
}

// Models that must not receive native tools (broken / XML-only path).
var toolModelDenied = []string{
	"deepseek-v2",
}

func ollamaURL() string {
	return EffectiveOllamaURL()
}

// EffectiveOllamaURL resolves the Ollama base URL in order:
// ai.ollama-url setting → OLLAMA_URL → OLLAMA_HOST → http://127.0.0.1:11434
func EffectiveOllamaURL() string {
	if cfg := currentConfig(); strings.TrimSpace(cfg.OllamaURL) != "" {
		return normalizeOllamaBaseURL(cfg.OllamaURL)
	}
	if u := strings.TrimSpace(os.Getenv("OLLAMA_URL")); u != "" {
		return normalizeOllamaBaseURL(u)
	}
	if u := strings.TrimSpace(os.Getenv("OLLAMA_HOST")); u != "" {
		return normalizeOllamaBaseURL(u)
	}
	return "http://127.0.0.1:11434"
}

// normalizeOllamaBaseURL accepts full URLs or Ollama-style hosts (host:port)
// and returns a scheme://host:port base without a trailing slash.
func normalizeOllamaBaseURL(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "http://127.0.0.1:11434"
	}
	s = strings.TrimRight(s, "/")

	scheme, hostport, ok := strings.Cut(s, "://")
	defaultPort := "11434"
	switch {
	case !ok:
		scheme, hostport = "http", s
	case scheme == "http":
		defaultPort = "80"
	case scheme == "https":
		defaultPort = "443"
	default:
		// keep scheme; fall back to 11434 for unknown schemes' empty ports
	}

	hostport, _, _ = strings.Cut(hostport, "/")
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		host, port = "127.0.0.1", defaultPort
		if ip := net.ParseIP(strings.Trim(hostport, "[]")); ip != nil {
			host = ip.String()
		} else if hostport != "" {
			host = hostport
		}
	}
	if port == "" {
		port = defaultPort
	}
	if n, err := strconv.ParseInt(port, 10, 32); err != nil || n > 65535 || n < 0 {
		port = defaultPort
	}

	return (&url.URL{
		Scheme: scheme,
		Host:   net.JoinHostPort(host, port),
	}).String()
}

// DefaultModelName returns the model for compose chat:
// ai.roles.compose-chat → CHAT_MODEL → DefaultModel.
func DefaultModelName() string {
	return ModelForRole(RoleComposeChat)
}

func NewClient(model string) (*Client, error) {
	if model == "" {
		model = DefaultModelName()
	}
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
	if model == "" {
		model = DefaultModelName()
	}
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
	cfg := currentConfig()
	if !cfg.Enabled {
		return []string{}, nil
	}

	all, err := DiscoverModels()
	if err != nil {
		return nil, err
	}

	allowed := cfg.EnabledModelNames()
	if allowed == nil {
		return all, nil
	}
	allow := make(map[string]struct{}, len(allowed))
	for _, n := range allowed {
		allow[n] = struct{}{}
	}
	out := make([]string, 0, len(all))
	for _, name := range all {
		if _, ok := allow[name]; ok {
			out = append(out, name)
		}
	}
	return out, nil
}

// DiscoverModels lists chat-capable models from Ollama /api/tags.
// Uses tag metadata only (no per-model /api/show) so the admin UI stays fast.
func DiscoverModels() ([]string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(ollamaURL() + "/api/tags")
	if err != nil {
		return nil, fmt.Errorf("failed to list ollama models: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list ollama models: status %d", resp.StatusCode)
	}

	var tags struct {
		Models []struct {
			Name    string `json:"name"`
			Details struct {
				Family   string   `json:"family"`
				Families []string `json:"families"`
			} `json:"details"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, fmt.Errorf("failed to decode ollama models: %w", err)
	}

	names := make([]string, 0, len(tags.Models))
	for _, m := range tags.Models {
		family := m.Details.Family
		if family == "" && len(m.Details.Families) > 0 {
			family = m.Details.Families[0]
		}
		if !isLikelyChatModel(m.Name, family) {
			continue
		}
		names = append(names, m.Name)
	}
	return names, nil
}

func isLikelyChatModel(name, family string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	f := strings.ToLower(strings.TrimSpace(family))
	for _, bad := range []string{"embed", "whisper", "clip", "coder-vision"} {
		if strings.Contains(n, bad) {
			return false
		}
	}
	switch f {
	case "bert", "nomic-bert", "clip", "whisper":
		return false
	}
	return n != ""
}

func supportsChat(name string) (bool, error) {
	caps, err := modelCapabilities(name)
	if err != nil {
		return false, err
	}
	for _, c := range caps {
		if c == "completion" {
			return true, nil
		}
	}
	return false, nil
}

func modelCapabilities(name string) ([]string, error) {
	payload, _ := json.Marshal(map[string]string{"model": name})
	resp, err := http.Post(ollamaURL()+"/api/show", "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to inspect model %s: %w", name, err)
	}
	defer resp.Body.Close()

	var show struct {
		Capabilities []string `json:"capabilities"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&show); err != nil {
		return nil, fmt.Errorf("failed to decode model %s: %w", name, err)
	}
	return show.Capabilities, nil
}

// modelBase strips registry path and tag: "library/qwen3:8b" → "qwen3".
func modelBase(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	if i := strings.Index(name, ":"); i >= 0 {
		name = name[:i]
	}
	return name
}

func toolsDenied(name string) bool {
	base := modelBase(name)
	for _, d := range toolModelDenied {
		if base == d || strings.HasPrefix(base, d+"-") {
			return true
		}
	}
	return false
}

func toolsAllowlisted(name string) bool {
	base := modelBase(name)
	for _, p := range toolModelPrefixes {
		if base == p || strings.HasPrefix(base, p+"-") || strings.HasPrefix(base, p+".") {
			return true
		}
	}
	return false
}

// ModelSupportsTools reports whether native Ollama tool calling should be
// enabled for the model. Prefer /api/show "tools" capability; fall back to a
// known-family allowlist when Ollama is unreachable or silent.
func ModelSupportsTools(name string) bool {
	if name == "" {
		name = DefaultModelName()
	}
	if toolsDenied(name) {
		return false
	}

	toolsCapMu.RLock()
	if v, ok := toolsCapCache[name]; ok {
		toolsCapMu.RUnlock()
		return v
	}
	toolsCapMu.RUnlock()

	supported := false
	if caps, err := modelCapabilities(name); err == nil {
		for _, c := range caps {
			if c == "tools" {
				supported = true
				break
			}
		}
	}
	if !supported {
		supported = toolsAllowlisted(name)
	}

	toolsCapMu.Lock()
	toolsCapCache[name] = supported
	toolsCapMu.Unlock()
	return supported
}

func (c *Client) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	return c.cm.Generate(ctx, messages, opts...)
}

func (c *Client) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return c.cm.Stream(ctx, messages, opts...)
}

func (c *Client) IsToolsSupported() bool {
	return ModelSupportsTools(c.model)
}

// ModelLikelySupportsTools is a fast allowlist/denylist check without calling Ollama.
func ModelLikelySupportsTools(name string) bool {
	if toolsDenied(name) {
		return false
	}
	return toolsAllowlisted(name)
}

func (c *Client) Model() string {
	return c.model
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
