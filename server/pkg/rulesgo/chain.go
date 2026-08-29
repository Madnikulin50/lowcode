package rulesgo

import (
	"encoding/json"
	"fmt"
	"strings"
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
	v := ec.Get(key)
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	s := fmt.Sprintf("%v", v)
	if s == "<nil>" {
		return ""
	}
	return s
}

func (ec *ExecutionContext) Get(key string) interface{} {
	key = strings.TrimSpace(key)
	if ec == nil || key == "" {
		return nil
	}
	if v := lookupPath(ec.Variables, key); !isEmptyGet(v) {
		return v
	}
	if v := lookupPath(ec.Results, key); !isEmptyGet(v) {
		return v
	}
	if v := lookupPath(ec.Input, key); !isEmptyGet(v) {
		return v
	}
	return nil
}

func isEmptyGet(v interface{}) bool {
	if v == nil {
		return true
	}
	if s, ok := v.(string); ok && strings.TrimSpace(s) == "" {
		return true
	}
	return false
}

func lookupPath(bag map[string]interface{}, key string) interface{} {
	if bag == nil || key == "" {
		return nil
	}
	if v, ok := bag[key]; ok {
		return v
	}
	if !strings.Contains(key, ".") {
		return mapLookupCI(bag, key)
	}
	var cur interface{} = bag
	for _, part := range strings.Split(key, ".") {
		part = strings.TrimSpace(part)
		m := asStringMap(cur)
		if m == nil {
			return nil
		}
		v, ok := m[part]
		if !ok {
			v = mapLookupCI(m, part)
		}
		if v == nil {
			return nil
		}
		cur = v
	}
	return cur
}

func mapLookupCI(m map[string]interface{}, key string) interface{} {
	if m == nil || key == "" {
		return nil
	}
	for k, v := range m {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return nil
}

func (ec *ExecutionContext) Set(key string, value interface{}) {
	ec.Variables[key] = value
}

func (ec *ExecutionContext) SetResult(nodeID string, value interface{}) {
	ec.Results[nodeID] = value
}
