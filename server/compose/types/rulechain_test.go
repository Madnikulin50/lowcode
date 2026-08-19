package types

import (
	"encoding/json"
	"testing"

	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
)

func TestRuleChainEngineRoundtrip(t *testing.T) {
	src := &rulesgo.Chain{
		ID:          "cmdb-trigger-scan",
		NamespaceID: 42,
		Name:        "CMDB scan",
		Description: "desc",
		EntryNode:   "http_scan",
		Nodes: []rulesgo.ChainNode{
			{ID: "http_scan", Type: "http", Label: "Scan", Config: json.RawMessage(`{"url":"http://x"}`)},
		},
		Edges:  []rulesgo.ChainEdge{{From: "a", To: "b", Label: "ok"}},
		Config: json.RawMessage(`{"retries":1}`),
	}

	row, err := RuleChainFromEngine(src)
	if err != nil {
		t.Fatal(err)
	}
	if row.Handle != src.ID || row.NamespaceID != 42 || row.EntryNode != "http_scan" {
		t.Fatalf("row mismatch: %+v", row)
	}
	if !json.Valid(row.Nodes) || !json.Valid(row.Edges) || !json.Valid(row.Config) {
		t.Fatalf("invalid json columns: nodes=%s edges=%s config=%s", row.Nodes, row.Edges, row.Config)
	}

	back, err := row.ToEngine()
	if err != nil {
		t.Fatal(err)
	}
	if back.ID != src.ID || back.NamespaceID != src.NamespaceID || back.Name != src.Name {
		t.Fatalf("engine mismatch: %+v", back)
	}
	if len(back.Nodes) != 1 || back.Nodes[0].Type != "http" {
		t.Fatalf("nodes: %+v", back.Nodes)
	}
	if len(back.Edges) != 1 || back.Edges[0].From != "a" {
		t.Fatalf("edges: %+v", back.Edges)
	}
}
