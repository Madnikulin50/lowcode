package types

import (
	"encoding/json"
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
)

type (
	RuleChain struct {
		ID          uint64          `json:"ruleChainID,string"`
		Handle      string          `json:"id"`
		NamespaceID uint64          `json:"namespaceID,string,omitempty"`
		Name        string          `json:"name"`
		Description string          `json:"description,omitempty"`
		EntryNode   string          `json:"entryNode"`
		Nodes       json.RawMessage `json:"nodes"`
		Edges       json.RawMessage `json:"edges"`
		Config      json.RawMessage `json:"config,omitempty"`
		CreatedAt   time.Time       `json:"createdAt,omitempty"`
		UpdatedAt   *time.Time      `json:"updatedAt,omitempty"`
		DeletedAt   *time.Time      `json:"deletedAt,omitempty"`
	}

	RuleChainFilter struct {
		NamespaceID uint64       `json:"namespaceID,string"`
		Handle      string       `json:"handle"`
		Query       string       `json:"query"`
		Deleted     filter.State `json:"deleted"`

		filter.Sorting
		filter.Paging
	}

	RuleChainSet []*RuleChain
)

func (rc *RuleChain) ToEngine() (*rulesgo.Chain, error) {
	if rc == nil {
		return nil, nil
	}
	out := &rulesgo.Chain{
		ID:          rc.Handle,
		NamespaceID: rc.NamespaceID,
		Name:        rc.Name,
		Description: rc.Description,
		EntryNode:   rc.EntryNode,
		Config:      rc.Config,
	}
	if len(rc.Nodes) > 0 {
		if err := json.Unmarshal(rc.Nodes, &out.Nodes); err != nil {
			return nil, err
		}
	}
	if len(rc.Edges) > 0 {
		if err := json.Unmarshal(rc.Edges, &out.Edges); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func RuleChainFromEngine(c *rulesgo.Chain) (*RuleChain, error) {
	if c == nil {
		return nil, nil
	}
	rc := &RuleChain{
		Handle:      c.ID,
		NamespaceID: c.NamespaceID,
		Name:        c.Name,
		Description: c.Description,
		EntryNode:   c.EntryNode,
		Config:      c.Config,
	}
	nodes, err := json.Marshal(c.Nodes)
	if err != nil {
		return nil, err
	}
	if c.Nodes == nil {
		nodes = []byte("[]")
	}
	edges, err := json.Marshal(c.Edges)
	if err != nil {
		return nil, err
	}
	if c.Edges == nil {
		edges = []byte("[]")
	}
	rc.Nodes = nodes
	rc.Edges = edges
	if len(rc.Config) == 0 {
		rc.Config = []byte("{}")
	}
	return rc, nil
}
