package chat

import (
	"os"
	"strings"
	"sync"
)

// Role identifiers for AI model assignment (admin settings ai.roles.*).
const (
	RoleComposeChat    = "compose.chat"
	RoleMCPAgent       = "mcp.agent"
	RoleAutomationChat = "automation.chat"
	RoleRulesgoAI      = "rulesgo.ai"
)

// Config is the runtime view of admin AI settings used by chat/MCP/automation.
type Config struct {
	Enabled   bool
	OllamaURL string
	Catalog   []CatalogEntry
	Roles     RoleModels
}

type CatalogEntry struct {
	Name    string
	Enabled bool
	Label   string
	Note    string
}

type RoleModels struct {
	ComposeChat    string
	MCPAgent       string
	AutomationChat string
	RulesgoAI      string
}

var (
	configMu       sync.RWMutex
	configProvider func() Config
)

// SetConfigProvider registers a callback that returns current AI settings
// (typically wired to system.CurrentSettings.AI).
func SetConfigProvider(fn func() Config) {
	configMu.Lock()
	defer configMu.Unlock()
	configProvider = fn
}

func currentConfig() Config {
	configMu.RLock()
	fn := configProvider
	configMu.RUnlock()
	if fn == nil {
		return Config{Enabled: true}
	}
	return fn()
}

func (c Config) RoleModel(role string) string {
	switch role {
	case RoleComposeChat:
		return strings.TrimSpace(c.Roles.ComposeChat)
	case RoleMCPAgent:
		return strings.TrimSpace(c.Roles.MCPAgent)
	case RoleAutomationChat:
		return strings.TrimSpace(c.Roles.AutomationChat)
	case RoleRulesgoAI:
		return strings.TrimSpace(c.Roles.RulesgoAI)
	default:
		return ""
	}
}

func (c Config) catalogIndex() map[string]CatalogEntry {
	out := make(map[string]CatalogEntry, len(c.Catalog))
	for _, e := range c.Catalog {
		name := strings.TrimSpace(e.Name)
		if name == "" {
			continue
		}
		out[name] = e
	}
	return out
}

// IsModelEnabled reports whether name may be offered in pickers / used as role default.
// Empty catalog ⇒ all models enabled (backward compatible).
func (c Config) IsModelEnabled(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	if len(c.Catalog) == 0 {
		return true
	}
	e, ok := c.catalogIndex()[name]
	return ok && e.Enabled
}

// EnabledModelNames returns enabled catalog names. Empty catalog ⇒ nil (no filter).
func (c Config) EnabledModelNames() []string {
	if len(c.Catalog) == 0 {
		return nil
	}
	out := make([]string, 0, len(c.Catalog))
	for _, e := range c.Catalog {
		if e.Enabled && strings.TrimSpace(e.Name) != "" {
			out = append(out, strings.TrimSpace(e.Name))
		}
	}
	return out
}

// RolesUsing returns role ids that currently point at the given model.
func (c Config) RolesUsing(name string) []string {
	name = strings.TrimSpace(name)
	var roles []string
	if c.Roles.ComposeChat == name {
		roles = append(roles, RoleComposeChat)
	}
	if c.Roles.MCPAgent == name {
		roles = append(roles, RoleMCPAgent)
	}
	if c.Roles.AutomationChat == name {
		roles = append(roles, RoleAutomationChat)
	}
	if c.Roles.RulesgoAI == name {
		roles = append(roles, RoleRulesgoAI)
	}
	return roles
}

// ModelForRole resolves the model for a usage role:
// role assignment (if enabled) → CHAT_MODEL env → DefaultModel constant.
func ModelForRole(role string) string {
	cfg := currentConfig()
	if m := cfg.RoleModel(role); m != "" && cfg.IsModelEnabled(m) {
		return m
	}
	if m := strings.TrimSpace(os.Getenv("CHAT_MODEL")); m != "" {
		return m
	}
	return DefaultModel
}

// CurrentConfig exposes the active AI config (for admin/API helpers).
func CurrentConfig() Config {
	return currentConfig()
}
