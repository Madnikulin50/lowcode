package aiagent

import (
	"context"
	"fmt"
	"sync"

	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

type Registry struct {
	mu     sync.RWMutex
	agents map[string]*Agent
	client *chat.Client
}

func NewRegistry(client *chat.Client) *Registry {
	return &Registry{
		agents: make(map[string]*Agent),
		client: client,
	}
}

func (r *Registry) Register(agent *Agent) {
	r.mu.Lock()
	r.agents[agent.Name()] = agent
	r.mu.Unlock()
}

func (r *Registry) RegisterDefault(tools []chat.ToolDef) {
	r.Register(NewCRUDAgent(r.client, tools))
	r.Register(NewAssistantAgent(r.client, tools))
}

func (r *Registry) Get(name string) *Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.agents[name]
}

func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.agents))
	for name := range r.agents {
		names = append(names, name)
	}
	return names
}

func (r *Registry) RunAgent(ctx context.Context, name, input string, contextData map[string]interface{}) (*AgentResult, error) {
	agent := r.Get(name)
	if agent == nil {
		return nil, fmt.Errorf("agent not found: %s (available: %v)", name, r.List())
	}
	return agent.Run(ctx, input, contextData), nil
}
