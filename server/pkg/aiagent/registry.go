package aiagent

import (
	"context"
	"fmt"
	"sync"

	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

type Registry struct {
	mu         sync.RWMutex
	agents     map[string]*Agent
	client     *chat.Client
	extraTools []chat.ToolDef
}

var (
	defaultRegMu sync.RWMutex
	defaultReg   *Registry
)

func SetDefaultRegistry(r *Registry) {
	defaultRegMu.Lock()
	defaultReg = r
	defaultRegMu.Unlock()
}

func DefaultRegistry() *Registry {
	defaultRegMu.RLock()
	defer defaultRegMu.RUnlock()
	return defaultReg
}

func NewRegistry(client *chat.Client) *Registry {
	r := &Registry{
		agents: make(map[string]*Agent),
		client: client,
	}
	SetDefaultRegistry(r)
	return r
}

func (r *Registry) Register(agent *Agent) {
	r.mu.Lock()
	r.agents[agent.Name()] = agent
	r.mu.Unlock()
}

func (r *Registry) RegisterDefault(tools []chat.ToolDef) {
	r.Reload(tools)
}

func (r *Registry) Reload(extraTools []chat.ToolDef) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if extraTools != nil {
		r.extraTools = extraTools
	}
	r.agents = make(map[string]*Agent)
	for _, spec := range EffectiveSpecs() {
		a := NewFromSpec(r.client, spec)
		if len(r.extraTools) > 0 {
			a.cfg.Tools = r.extraTools
		}
		r.agents[a.Name()] = a
	}
}

func (r *Registry) Get(name string) *Agent {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	a := r.agents[name]
	extra := r.extraTools
	r.mu.RUnlock()
	if a != nil {
		return a
	}
	for _, spec := range EffectiveSpecs() {
		if spec.Handle == name {
			a := NewFromSpec(r.client, spec)
			if len(extra) > 0 {
				a.cfg.Tools = extra
			}
			r.Register(a)
			return a
		}
	}
	return nil
}

func (r *Registry) List() []string {
	infos := r.ListInfo()
	names := make([]string, 0, len(infos))
	for _, i := range infos {
		names = append(names, i.Handle)
	}
	return names
}

func (r *Registry) ListInfo() []AgentInfo {
	if r == nil {
		specs := EffectiveSpecs()
		out := make([]AgentInfo, 0, len(specs))
		for _, s := range specs {
			out = append(out, s.Info())
		}
		return out
	}
	specs := EffectiveSpecs()
	out := make([]AgentInfo, 0, len(specs))
	seen := map[string]struct{}{}
	for _, s := range specs {
		out = append(out, s.Info())
		seen[s.Handle] = struct{}{}
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for name, a := range r.agents {
		if _, ok := seen[name]; ok {
			continue
		}
		out = append(out, AgentInfo{
			Handle:      name,
			Description: a.Description(),
			Toolkits:    append([]string(nil), a.cfg.Toolkits...),
			Model:       a.cfg.Model,
			MaxSteps:    a.cfg.MaxSteps,
			Confirm:     a.cfg.Confirm,
			Source:      "runtime",
		})
	}
	return out
}

func (r *Registry) RunAgent(ctx context.Context, name, input string, contextData map[string]interface{}) (*AgentResult, error) {
	agent := r.Get(name)
	if agent == nil {
		return nil, fmt.Errorf("agent not found: %s (available: %v)", name, r.List())
	}
	return agent.Run(ctx, input, contextData), nil
}
