package chat

import (
	"context"
	"io"
	"testing"
)

func TestModelBase(t *testing.T) {
	cases := map[string]string{
		"qwen3:8b":                               "qwen3",
		"Qwen3:8b-q4_K_M":                        "qwen3",
		"library/qwen3:8b":                       "qwen3",
		"llama3-groq-tool-use:8b":                "llama3-groq-tool-use",
		"registry.ollama.ai/library/qwen2.5:14b": "qwen2.5",
		"deepseek-r1":                            "deepseek-r1",
	}
	for in, want := range cases {
		if got := modelBase(in); got != want {
			t.Errorf("modelBase(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestToolsAllowlisted(t *testing.T) {
	yes := []string{
		"qwen3:8b",
		"qwen2.5:7b",
		"qwen2.5-coder:7b",
		"llama3.1:8b",
		"llama3-groq-tool-use:8b",
		"deepseek-r1:latest",
		"mistral:7b",
	}
	for _, name := range yes {
		if !toolsAllowlisted(name) {
			t.Errorf("toolsAllowlisted(%q) = false, want true", name)
		}
	}
	no := []string{"phi3:mini", "nomic-embed-text", "llava:7b"}
	for _, name := range no {
		if toolsAllowlisted(name) {
			t.Errorf("toolsAllowlisted(%q) = true, want false", name)
		}
	}
}

func TestToolsDenied(t *testing.T) {
	if !toolsDenied("deepseek-v2") {
		t.Fatal("deepseek-v2 should be denied")
	}
	if !toolsDenied("deepseek-v2:latest") {
		t.Fatal("deepseek-v2:latest should be denied")
	}
	if toolsDenied("deepseek-r1") {
		t.Fatal("deepseek-r1 should not be denied")
	}
	if toolsDenied("qwen3:8b") {
		t.Fatal("qwen3:8b should not be denied")
	}
}

func TestModelSupportsToolsDeniedOverridesCache(t *testing.T) {
	// denylist must win even without talking to Ollama
	if ModelSupportsTools("deepseek-v2") {
		t.Fatal("deepseek-v2 must not support tools")
	}
}

func TestPickInstalledModel(t *testing.T) {
	installed := []string{
		"qwen3:latest",
		"qwen3:32b",
		"deepseek-r1:latest",
		"llama3.2:latest",
	}
	cases := map[string]string{
		"qwen3:latest":    "qwen3:latest",
		"qwen3:8b":        "qwen3:latest",
		"qwen3":           "qwen3:latest",
		"qwen3:32b":       "qwen3:32b",
		"deepseek-r1:7b":  "deepseek-r1:latest",
		"missing:latest":  "",
		"llama3.2:latest": "llama3.2:latest",
	}
	for in, want := range cases {
		if got := pickInstalledModel(in, installed); got != want {
			t.Errorf("pickInstalledModel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestModelNamesMatch(t *testing.T) {
	if !modelNamesMatch("deepseek-r1:latest", "deepseek-r1:latest") {
		t.Fatal("exact should match")
	}
	if !modelNamesMatch("qwen3", "qwen3:latest") {
		t.Fatal("tagless vs :latest should match")
	}
	if modelNamesMatch("qwen3:8b", "qwen3:latest") {
		t.Fatal("size tag vs :latest must not match without resolve")
	}
	if modelNamesMatch("qwen3:8b", "qwen3:32b") {
		t.Fatal("different sizes must not match")
	}
}

func TestIsLikelyChatModel(t *testing.T) {
	if isLikelyChatModel("nomic-embed-text:latest", "nomic-bert") {
		t.Fatal("embed models must be excluded")
	}
	if isLikelyChatModel("x/z-image-turbo:latest", "") {
		t.Fatal("image models must be excluded")
	}
	if !isLikelyChatModel("deepseek-r1:latest", "qwen3") {
		t.Fatal("chat models must be included")
	}
}

func TestPreferDefaultModel(t *testing.T) {
	t.Setenv("CHAT_MODEL", "deepseek-r1:latest")
	got := preferDefaultModel([]string{"qwen3.6:latest", "deepseek-r1:latest", "llama3.2:latest"})
	if got[0] != "deepseek-r1:latest" {
		t.Fatalf("default should be first, got %v", got)
	}
}

func TestDefaultModelName(t *testing.T) {
	t.Setenv("CHAT_MODEL", "")
	if DefaultModelName() != DefaultModel {
		t.Fatalf("DefaultModelName() = %q, want %q", DefaultModelName(), DefaultModel)
	}
	t.Setenv("CHAT_MODEL", "llama3.1:8b")
	if DefaultModelName() != "llama3.1:8b" {
		t.Fatalf("CHAT_MODEL override failed: %q", DefaultModelName())
	}
}

func TestModelForRoleUsesConfig(t *testing.T) {
	t.Setenv("CHAT_MODEL", "")
	SetConfigProvider(func() Config {
		return Config{
			Enabled: true,
			Catalog: []CatalogEntry{
				{Name: "qwen3:8b", Enabled: true},
				{Name: "deepseek-r1", Enabled: false},
			},
			Roles: RoleModels{ComposeChat: "qwen3:8b", MCPAgent: "deepseek-r1"},
		}
	})
	t.Cleanup(func() { SetConfigProvider(nil) })

	if got := ModelForRole(RoleComposeChat); got != "qwen3:8b" {
		t.Fatalf("compose.chat = %q", got)
	}
	// disabled role assignment falls back to DefaultModel
	if got := ModelForRole(RoleMCPAgent); got != DefaultModel {
		t.Fatalf("mcp.agent disabled fallback = %q, want %q", got, DefaultModel)
	}
}

func TestNormalizeOllamaBaseURL(t *testing.T) {
	cases := map[string]string{
		"":                       "http://127.0.0.1:11434",
		"localhost":              "http://localhost:11434",
		"127.0.0.1:11434":        "http://127.0.0.1:11434",
		"http://ollama:11434":    "http://ollama:11434",
		"http://ollama:11434/":   "http://ollama:11434",
		"https://ollama.example": "https://ollama.example:443",
		"host.docker.internal":   "http://host.docker.internal:11434",
	}
	for in, want := range cases {
		if got := normalizeOllamaBaseURL(in); got != want {
			t.Errorf("normalizeOllamaBaseURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEffectiveOllamaURLFromHost(t *testing.T) {
	SetConfigProvider(func() Config { return Config{Enabled: true} })
	t.Cleanup(func() { SetConfigProvider(nil) })
	t.Setenv("OLLAMA_URL", "")
	t.Setenv("OLLAMA_HOST", "10.0.0.5:11434")
	if got := EffectiveOllamaURL(); got != "http://10.0.0.5:11434" {
		t.Fatalf("EffectiveOllamaURL() = %q", got)
	}
	t.Setenv("OLLAMA_URL", "http://from-url:11434")
	if got := EffectiveOllamaURL(); got != "http://from-url:11434" {
		t.Fatalf("OLLAMA_URL should win: %q", got)
	}
}

func TestIsTimeout(t *testing.T) {
	if IsTimeout(nil) {
		t.Fatal("nil should not be timeout")
	}
	if !IsTimeout(context.DeadlineExceeded) {
		t.Fatal("context.DeadlineExceeded should be timeout")
	}
	if IsTimeout(io.EOF) {
		t.Fatal("EOF should not be timeout")
	}
}
