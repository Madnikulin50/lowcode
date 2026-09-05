package aiagent

import (
	"embed"
	"io/fs"
	"os"
	"strings"
	"sync"
)

//go:embed defs/*.yaml
var builtinFS embed.FS

var (
	extrasMu       sync.RWMutex
	extrasProvider func() []AgentSpec
)

func SetExtrasProvider(fn func() []AgentSpec) {
	extrasMu.Lock()
	extrasProvider = fn
	extrasMu.Unlock()
}

func BuiltinSpecs() []AgentSpec {
	entries, err := fs.ReadDir(builtinFS, "defs")
	if err != nil {
		return nil
	}
	var out []AgentSpec
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		raw, err := builtinFS.ReadFile("defs/" + e.Name())
		if err != nil {
			continue
		}
		s, err := ParseSpecYAML(raw)
		if err != nil {
			continue
		}
		s.Source = "builtin"
		out = append(out, s)
	}
	return out
}

func extrasSpecs() []AgentSpec {
	extrasMu.RLock()
	fn := extrasProvider
	extrasMu.RUnlock()
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

func fileSpecs() []AgentSpec {
	dir := strings.TrimSpace(os.Getenv("AI_AGENTS_DIR"))
	if dir == "" {
		return nil
	}
	files, err := LoadSpecDir(dir)
	if err != nil {
		return nil
	}
	return files
}

func EffectiveSpecs() []AgentSpec {
	specs := overlaySpecs(BuiltinSpecs(), extrasSpecs())
	if files := fileSpecs(); len(files) > 0 {
		specs = overlaySpecs(specs, files)
	}
	var enabled []AgentSpec
	for _, s := range specs {
		if s.IsEnabled() && s.Handle != "" {
			enabled = append(enabled, s)
		}
	}
	return enabled
}

// CatalogPayload is the admin editor snapshot: builtins + file overlay + toolkit names.
// Settings overlay is edited separately via ai.agents.
func CatalogPayload() map[string]interface{} {
	conns := SeedConnectors()
	safe := make([]Connector, 0, len(conns))
	for _, c := range conns {
		c.Token = ""
		safe = append(safe, c)
	}
	return map[string]interface{}{
		"builtins":   BuiltinSpecs(),
		"files":      fileSpecs(),
		"toolkits":   DefaultCatalog().Names(),
		"kits":       DefaultCatalog().Summaries(),
		"connectors": safe,
	}
}
