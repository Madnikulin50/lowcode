package aiagent

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/madnikulin50/lowcode/server/pkg/chat"
	"gopkg.in/yaml.v3"
)

// AgentSpec is a named agent definition (YAML / admin settings).
type AgentSpec struct {
	Handle      string   `yaml:"handle" json:"handle"`
	Enabled     *bool    `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Prompt      string   `yaml:"prompt,omitempty" json:"prompt,omitempty"`
	Model       string   `yaml:"model,omitempty" json:"model,omitempty"`
	Toolkits    []string `yaml:"toolkits,omitempty" json:"toolkits,omitempty"`
	MaxSteps    int      `yaml:"maxSteps,omitempty" json:"maxSteps,omitempty"`
	Confirm     bool     `yaml:"confirm,omitempty" json:"confirm,omitempty"`
	Source      string   `yaml:"-" json:"source,omitempty"`
}

func (s AgentSpec) IsEnabled() bool {
	if s.Enabled == nil {
		return true
	}
	return *s.Enabled
}

func (s AgentSpec) Info() AgentInfo {
	return AgentInfo{
		Handle:      s.Handle,
		Description: s.Description,
		Toolkits:    append([]string(nil), s.Toolkits...),
		Model:       s.Model,
		MaxSteps:    s.MaxSteps,
		Confirm:     s.Confirm,
		Source:      s.Source,
	}
}

type AgentInfo struct {
	Handle      string   `json:"handle"`
	Description string   `json:"description,omitempty"`
	Toolkits    []string `json:"toolkits,omitempty"`
	Model       string   `json:"model,omitempty"`
	MaxSteps    int      `json:"maxSteps,omitempty"`
	Confirm     bool     `json:"confirm,omitempty"`
	Source      string   `json:"source,omitempty"`
}

func ParseSpecYAML(raw []byte) (AgentSpec, error) {
	var s AgentSpec
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(false)
	if err := dec.Decode(&s); err != nil {
		return s, err
	}
	s.Handle = strings.TrimSpace(s.Handle)
	if s.Handle == "" {
		return s, fmt.Errorf("agent spec missing handle")
	}
	return s, nil
}

func LoadSpecDir(dir string) ([]AgentSpec, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil, nil
	}
	matches, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	yml, _ := filepath.Glob(filepath.Join(dir, "*.yml"))
	matches = append(matches, yml...)
	var out []AgentSpec
	for _, p := range matches {
		raw, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		s, err := ParseSpecYAML(raw)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		s.Source = "file"
		out = append(out, s)
	}
	return out, nil
}

func overlaySpecs(base, over []AgentSpec) []AgentSpec {
	return mergeSpecs(base, over, false)
}

// mergeSpecs overlays `over` onto `base` by handle. When keepDisabled is false
// (runtime), enabled:false removes the agent. When true (admin catalog), the
// disabled spec stays so the UI can show and re-enable it.
func mergeSpecs(base, over []AgentSpec, keepDisabled bool) []AgentSpec {
	idx := map[string]int{}
	out := make([]AgentSpec, 0, len(base)+len(over))
	for _, s := range base {
		if s.Handle == "" {
			continue
		}
		if !keepDisabled && !s.IsEnabled() {
			continue
		}
		idx[s.Handle] = len(out)
		out = append(out, s)
	}
	for _, s := range over {
		if s.Handle == "" {
			continue
		}
		if !s.IsEnabled() && !keepDisabled {
			if i, ok := idx[s.Handle]; ok {
				out = append(out[:i], out[i+1:]...)
				idx = reindex(out)
			}
			continue
		}
		if i, ok := idx[s.Handle]; ok {
			out[i] = s
			continue
		}
		idx[s.Handle] = len(out)
		out = append(out, s)
	}
	return out
}

func reindex(specs []AgentSpec) map[string]int {
	idx := make(map[string]int, len(specs))
	for i, s := range specs {
		idx[s.Handle] = i
	}
	return idx
}

func resolveSpecModel(model string) string {
	model = strings.TrimSpace(model)
	switch model {
	case "", chat.RoleMCPAgent, "mcp-agent":
		return chat.ModelForRole(chat.RoleMCPAgent)
	case chat.RoleComposeChat, "compose-chat":
		return chat.ModelForRole(chat.RoleComposeChat)
	case chat.RoleAutomationChat, "automation-chat":
		return chat.ModelForRole(chat.RoleAutomationChat)
	case chat.RoleRulesgoAI, "rulesgo-ai":
		return chat.ModelForRole(chat.RoleRulesgoAI)
	default:
		return model
	}
}

func NewFromSpec(client *chat.Client, spec AgentSpec) *Agent {
	return New(client, AgentConfig{
		Name:         spec.Handle,
		Description:  spec.Description,
		SystemPrompt: strings.TrimSpace(spec.Prompt),
		Model:        resolveSpecModel(spec.Model),
		Toolkits:     spec.Toolkits,
		MaxSteps:     spec.MaxSteps,
		Confirm:      spec.Confirm,
	})
}
