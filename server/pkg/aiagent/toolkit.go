package aiagent

import (
	"sync"

	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

// ToolKit is a named group of tools, Hermes-style (compose.records, cmdb, …).
type ToolKit struct {
	Name        string
	Description string
	Tools       []chat.ToolDef
	Remote      bool
	URL         string
}

type Catalog struct {
	mu    sync.RWMutex
	kits  map[string]ToolKit
	order []string
}

func NewCatalog() *Catalog {
	return &Catalog{kits: make(map[string]ToolKit)}
}

var (
	defaultCatalog     *Catalog
	defaultCatalogOnce sync.Once
)

func DefaultCatalog() *Catalog {
	defaultCatalogOnce.Do(func() {
		defaultCatalog = NewCatalog()
		defaultCatalog.RegisterFallbacks()
	})
	return defaultCatalog
}

func (c *Catalog) Register(k ToolKit) {
	if c == nil || k.Name == "" {
		return
	}
	if ReservedKitName(k.Name) && k.Remote {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.kits[k.Name]; !exists {
		c.order = append(c.order, k.Name)
	}
	c.kits[k.Name] = k
}

func (c *Catalog) Unregister(name string) {
	if c == nil || name == "" || ReservedKitName(name) {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.kits[name]; !ok {
		return
	}
	delete(c.kits, name)
	out := c.order[:0]
	for _, n := range c.order {
		if n != name {
			out = append(out, n)
		}
	}
	c.order = out
}

func (c *Catalog) Get(name string) (ToolKit, bool) {
	if c == nil {
		return ToolKit{}, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	k, ok := c.kits[name]
	return k, ok
}

func (c *Catalog) Names() []string {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]string, len(c.order))
	copy(out, c.order)
	return out
}

func (c *Catalog) List() []ToolKit {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]ToolKit, 0, len(c.order))
	for _, name := range c.order {
		out = append(out, c.kits[name])
	}
	return out
}

func (c *Catalog) Summaries() []map[string]interface{} {
	kits := c.List()
	out := make([]map[string]interface{}, 0, len(kits))
	for _, k := range kits {
		names := make([]string, 0, len(k.Tools))
		for _, t := range k.Tools {
			if t.Name != "" {
				names = append(names, t.Name)
			}
		}
		out = append(out, map[string]interface{}{
			"name":        k.Name,
			"description": k.Description,
			"remote":      k.Remote,
			"url":         k.URL,
			"tools":       names,
		})
	}
	return out
}

// Resolve flattens named kits, skipping unknown names and duplicate tool names.
// "*" means every registered kit in registration order.
func (c *Catalog) Resolve(names ...string) []chat.ToolDef {
	if c == nil || len(names) == 0 {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	seen := map[string]struct{}{}
	var out []chat.ToolDef
	appendKit := func(k ToolKit) {
		for _, t := range k.Tools {
			if t.Name == "" {
				continue
			}
			if _, ok := seen[t.Name]; ok {
				continue
			}
			seen[t.Name] = struct{}{}
			out = append(out, t)
		}
	}
	for _, name := range names {
		if name == "*" {
			for _, n := range c.order {
				appendKit(c.kits[n])
			}
			continue
		}
		if k, ok := c.kits[name]; ok {
			appendKit(k)
		}
	}
	return out
}

func Flatten(kits ...ToolKit) []chat.ToolDef {
	seen := map[string]struct{}{}
	var out []chat.ToolDef
	for _, k := range kits {
		for _, t := range k.Tools {
			if t.Name == "" {
				continue
			}
			if _, ok := seen[t.Name]; ok {
				continue
			}
			seen[t.Name] = struct{}{}
			out = append(out, t)
		}
	}
	return out
}
