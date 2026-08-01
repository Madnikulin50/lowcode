package rulesgo

import (
	"context"
	"encoding/json"
	"fmt"
)

type NodeExecutor interface {
	Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error)
}

type Registry struct {
	nodes map[string]NodeExecutor
}

func NewRegistry() *Registry {
	return &Registry{
		nodes: make(map[string]NodeExecutor),
	}
}

func (r *Registry) Register(nodeType string, executor NodeExecutor) {
	r.nodes[nodeType] = executor
}

func (r *Registry) Execute(ctx context.Context, nodeType string, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	executor, ok := r.nodes[nodeType]
	if !ok {
		return nil, fmt.Errorf("unknown node type: %s", nodeType)
	}
	return executor.Execute(ctx, node, ec)
}

func (r *Registry) Get(nodeType string) (NodeExecutor, bool) {
	executor, ok := r.nodes[nodeType]
	return executor, ok
}

func ParseNodeConfig[T any](raw json.RawMessage) (T, error) {
	var cfg T
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return cfg, fmt.Errorf("failed to parse node config: %w", err)
		}
	}
	return cfg, nil
}
