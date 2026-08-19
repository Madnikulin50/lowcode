package rulesgo

import (
	"encoding/json"
)

type Chain struct {
	ID          string          `json:"id"`
	NamespaceID uint64          `json:"namespaceID,omitempty"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Nodes       []ChainNode     `json:"nodes"`
	Edges       []ChainEdge     `json:"edges"`
	EntryNode   string          `json:"entryNode"`
	Config      json.RawMessage `json:"config,omitempty"`
}

type ChainNode struct {
	ID     string          `json:"id"`
	Type   string          `json:"type"`
	Label  string          `json:"label"`
	Config json.RawMessage `json:"config"`
}

type ChainEdge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label,omitempty"`
	// Condition for conditional branching; if empty, always follows
	Condition string `json:"condition,omitempty"`
}

type ChainResult struct {
	ChainID string                 `json:"chainID"`
	Nodes   []NodeResult           `json:"nodes"`
	Output  map[string]interface{} `json:"output"`
	Success bool                   `json:"success"`
	Error   string                 `json:"error,omitempty"`
}

type NodeResult struct {
	NodeID string                 `json:"nodeID"`
	Type   string                 `json:"type"`
	Output map[string]interface{} `json:"output"`
	Next   []string               `json:"next"`
	Error  string                 `json:"error,omitempty"`
}

type ExecutionContext struct {
	Variables map[string]interface{} `json:"variables"`
	Results   map[string]interface{} `json:"results"`
	Input     map[string]interface{} `json:"input"`
}

func (ec *ExecutionContext) GetString(key string) string {
	if v, ok := ec.Variables[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	if v, ok := ec.Input[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (ec *ExecutionContext) Get(key string) interface{} {
	if v, ok := ec.Variables[key]; ok {
		return v
	}
	if v, ok := ec.Results[key]; ok {
		return v
	}
	if v, ok := ec.Input[key]; ok {
		return v
	}
	return nil
}

func (ec *ExecutionContext) Set(key string, value interface{}) {
	ec.Variables[key] = value
}

func (ec *ExecutionContext) SetResult(nodeID string, value interface{}) {
	ec.Results[nodeID] = value
}
