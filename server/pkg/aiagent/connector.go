package aiagent

import (
	"net/url"
	"os"
	"strings"
	"sync"
)

// Connector is an external HTTP toolkit (GET {url}/meta, POST /jobs).
type Connector struct {
	Handle  string `json:"handle"`
	URL     string `json:"url"`
	Enabled *bool  `json:"enabled,omitempty"`
	Token   string `json:"token,omitempty"`
	Source  string `json:"source,omitempty"`
}

func (c Connector) IsEnabled() bool {
	if c.Enabled == nil {
		return true
	}
	return *c.Enabled
}

func ReservedKitName(name string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	return strings.HasPrefix(n, "compose.")
}

func ValidConnectorURL(raw string) (string, error) {
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	if raw == "" {
		return "", errConnectorURLEmpty
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
	default:
		return "", errConnectorURLScheme
	}
	if u.Host == "" {
		return "", errConnectorURLEmpty
	}
	return raw, nil
}

var (
	errConnectorURLEmpty  = errString("toolkit url is required")
	errConnectorURLScheme = errString("toolkit url must be http or https")
)

type errString string

func (e errString) Error() string { return string(e) }

var (
	connMu       sync.RWMutex
	connProvider func() []Connector
)

func SetConnectorsProvider(fn func() []Connector) {
	connMu.Lock()
	connProvider = fn
	connMu.Unlock()
}

func ParseToolKitEnv(raw string) []Connector {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	raw = strings.ReplaceAll(raw, ",", ";")
	var out []Connector
	for _, part := range strings.Split(raw, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		handle, u, ok := strings.Cut(part, "=")
		handle = strings.TrimSpace(handle)
		u = strings.TrimSpace(u)
		if !ok || handle == "" || u == "" || ReservedKitName(handle) {
			continue
		}
		if cleaned, err := ValidConnectorURL(u); err == nil {
			u = cleaned
		}
		out = append(out, Connector{Handle: handle, URL: u, Source: "env"})
	}
	return out
}

func SeedConnectors() []Connector {
	var out []Connector
	for _, handle := range []string{"cmdb", "backup", "invest"} {
		u := strings.TrimRight(strings.TrimSpace(DefaultAgentURL(handle)), "/")
		if u == "" {
			continue
		}
		out = append(out, Connector{Handle: handle, URL: u, Source: "env"})
	}
	return overlayConnectors(out, ParseToolKitEnv(os.Getenv("AI_TOOLKITS")))
}

func extrasConnectors() []Connector {
	connMu.RLock()
	fn := connProvider
	connMu.RUnlock()
	if fn == nil {
		return nil
	}
	extra := fn()
	for i := range extra {
		if extra[i].Source == "" {
			extra[i].Source = "settings"
		}
	}
	return extra
}

func overlayConnectors(base, over []Connector) []Connector {
	idx := map[string]int{}
	out := make([]Connector, 0, len(base)+len(over))
	for _, c := range base {
		h := strings.TrimSpace(c.Handle)
		if h == "" || ReservedKitName(h) {
			continue
		}
		c.Handle = h
		idx[h] = len(out)
		out = append(out, c)
	}
	for _, c := range over {
		h := strings.TrimSpace(c.Handle)
		if h == "" || ReservedKitName(h) {
			continue
		}
		c.Handle = h
		if i, ok := idx[h]; ok {
			out[i] = c
			continue
		}
		idx[h] = len(out)
		out = append(out, c)
	}
	return out
}

func EffectiveConnectors() []Connector {
	return overlayConnectors(SeedConnectors(), extrasConnectors())
}

func AssistantKitNames() []string {
	for _, s := range EffectiveSpecs() {
		if s.Handle == "assistant" && len(s.Toolkits) > 0 {
			return append([]string(nil), s.Toolkits...)
		}
	}
	return []string{"cmdb", "backup", "invest"}
}
