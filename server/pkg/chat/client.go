package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"golang.org/x/sync/singleflight"
)

const (
	// DefaultModel is the preferred Ollama model for chat / MCP tool orchestration.
	// Override with CHAT_MODEL.
	DefaultModel = "qwen3:8b"
)

type (
	StreamFunc func(token string, reason string, done bool) error

	// StatusFunc reports model lifecycle events over the chat stream
	// (e.g. "warming", "ready") so the UI can show a warmup indicator.
	StatusFunc func(status string) error

	Client struct {
		cm    *ollama.ChatModel
		model string
	}
)

const (
	StatusWarming = "warming"
	StatusReady   = "ready"

	// WarmUpTimeout covers cold model load into RAM (no generation).
	// Generation timeout (NewClient HTTPClient) starts only after WarmUp returns.
	WarmUpTimeout = 10 * time.Minute
)

type statusCtxKey struct{}

var (
	toolsCapMu    sync.RWMutex
	toolsCapCache = map[string]bool{}
	warmGroup     singleflight.Group
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
		HTTPClient: ollamaHTTPClient(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ollama chat model: %w", err)
	}
	return &Client{cm: cm, model: model}, nil
}

// ollamaHTTPClient has no overall Timeout: thinking models (qwen3, deepseek-r1)
// can spend several minutes on reasoning after warmup. The SSE request
// context still cancels when the browser disconnects.
// ResponseHeaderTimeout covers a hung Ollama before the first byte.
func ollamaHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = 10 * time.Minute
	return &http.Client{
		Timeout:   0,
		Transport: transport,
	}
}

// IsTimeout reports whether err is an HTTP/context deadline.
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, os.ErrDeadlineExceeded) {
		return true
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "Client.Timeout exceeded") ||
		strings.Contains(msg, "context deadline exceeded")
}

// ContextWithStatus attaches a StatusFunc used by EnsureWarm to push
// warmup progress to SSE clients.
func ContextWithStatus(ctx context.Context, fn StatusFunc) context.Context {
	if fn == nil {
		return ctx
	}
	return context.WithValue(ctx, statusCtxKey{}, fn)
}

func EmitStatus(ctx context.Context, status string) {
	if status == "" {
		return
	}
	fn, _ := ctx.Value(statusCtxKey{}).(StatusFunc)
	if fn == nil {
		return
	}
	_ = fn(status)
}

func (c *Client) Model() string {
	return c.model
}

// IsModelLoaded reports whether Ollama currently has the model in memory.
func IsModelLoaded(model string) bool {
	if model == "" {
		model = DefaultModelName()
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(ollamaURL() + "/api/ps")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	var ps struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ps); err != nil {
		return false
	}
	want := strings.ToLower(strings.TrimSpace(model))
	for _, m := range ps.Models {
		if modelNamesMatch(want, m.Name) {
			return true
		}
	}
	return false
}

// WarmUp forces Ollama to load the model into memory so that the first
// real chat message is not delayed by model loading. It only loads weights
// (empty /api/generate) and does not run a thinking/chat completion —
// num_predict does not cap reasoning tokens on qwen3 / deepseek-r1.
// Concurrent calls for the same model coalesce (singleflight).
func WarmUp(model string) error {
	if model == "" {
		model = DefaultModelName()
	}
	if IsModelLoaded(model) {
		return nil
	}
	model = ResolveInstalledModel(model)
	// Skip even if another warmup is still in flight: once /api/ps shows
	// the model, generation is no longer needed for "warm".
	if IsModelLoaded(model) {
		return nil
	}
	_, err, _ := warmGroup.Do(model, func() (any, error) {
		if IsModelLoaded(model) {
			return nil, nil
		}
		return nil, warmUpOnce(model)
	})
	return err
}

// EnsureWarm emits status events and warms the model when needed so the
// generation HTTP timeout does not include cold-load time.
func EnsureWarm(ctx context.Context, model string) error {
	if model == "" {
		model = DefaultModelName()
	}
	if IsModelLoaded(model) {
		return nil
	}
	model = ResolveInstalledModel(model)
	if IsModelLoaded(model) {
		return nil
	}
	EmitStatus(ctx, StatusWarming)
	err := WarmUp(model)
	if err != nil {
		return err
	}
	EmitStatus(ctx, StatusReady)
	return nil
}

func warmUpOnce(model string) error {
	// Empty prompt = load weights only (Ollama skips generation). think:false
	// is a safety net if a given Ollama build still evaluates the template.
	payload, _ := json.Marshal(map[string]any{
		"model":      model,
		"prompt":     "",
		"stream":     false,
		"keep_alive": "30m",
		"think":      false,
	})
	client := &http.Client{Timeout: WarmUpTimeout}
	resp, err := client.Post(ollamaURL()+"/api/generate", "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to warm up model %s: %w", model, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("failed to warm up model %s: status %d: %s", model, resp.StatusCode, strings.TrimSpace(string(body)))
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
	return preferDefaultModel(names), nil
}

func isLikelyChatModel(name, family string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	f := strings.ToLower(strings.TrimSpace(family))
	for _, bad := range []string{"embed", "whisper", "clip", "coder-vision", "image"} {
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

// ResolveInstalledModel maps a configured name (e.g. qwen3:8b) to a tag
// that actually exists in Ollama. Returns want unchanged if none match.
func ResolveInstalledModel(want string) string {
	want = strings.TrimSpace(want)
	if want == "" {
		return want
	}
	names, err := DiscoverModels()
	if err != nil || len(names) == 0 {
		return want
	}
	if picked := pickInstalledModel(want, names); picked != "" {
		return picked
	}
	return want
}

func preferDefaultModel(names []string) []string {
	if len(names) < 2 {
		return names
	}
	out := append([]string(nil), names...)
	def := pickInstalledModel(DefaultModelName(), out)
	sort.SliceStable(out, func(i, j int) bool {
		if def != "" {
			if out[i] == def {
				return true
			}
			if out[j] == def {
				return false
			}
		}
		return out[i] < out[j]
	})
	return out
}

// pickInstalledModel chooses an installed tag for want.
// Exact match wins; otherwise same base with :latest (so qwen3:8b → qwen3:latest).
func pickInstalledModel(want string, installed []string) string {
	want = strings.ToLower(strings.TrimSpace(want))
	if want == "" {
		return ""
	}
	for _, n := range installed {
		if strings.ToLower(strings.TrimSpace(n)) == want {
			return n
		}
	}
	wantBase := modelBase(want)
	var same []string
	for _, n := range installed {
		if modelBase(n) == wantBase {
			same = append(same, n)
		}
	}
	if len(same) == 0 {
		return ""
	}
	wantTag := modelTag(want)
	if wantTag == "" || wantTag == "latest" || isParamSizeTag(wantTag) {
		for _, n := range same {
			if modelTag(n) == "latest" || modelTag(n) == "" {
				return n
			}
		}
	}
	for _, n := range same {
		if modelTag(n) == wantTag {
			return n
		}
	}
	return same[0]
}

func modelTag(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	if i := strings.Index(name, ":"); i >= 0 {
		return name[i+1:]
	}
	return ""
}

func isParamSizeTag(tag string) bool {
	if tag == "" || strings.ContainsAny(tag, "/-") {
		return false
	}
	return strings.HasSuffix(tag, "b") && len(tag) <= 6
}

// modelNamesMatch is true for the same installed tag, or name vs name:latest.
// Same base with different size tags (qwen3:8b vs qwen3:32b) do not match.
func modelNamesMatch(want, got string) bool {
	want = strings.ToLower(strings.TrimSpace(want))
	got = strings.ToLower(strings.TrimSpace(got))
	if want == "" || got == "" {
		return false
	}
	if want == got {
		return true
	}
	if modelBase(want) != modelBase(got) {
		return false
	}
	wt, gt := modelTag(want), modelTag(got)
	latest := func(t string) bool { return t == "" || t == "latest" }
	return latest(wt) && latest(gt)
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
