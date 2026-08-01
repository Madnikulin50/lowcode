package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/madnikulin50/lowcode/server/compose/mcp/handlers"
	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
)

func TestBridge_RuleEngine(t *testing.T) {
	// Simulate bridge initialization
	engine := rulesgo.NewEngine(rulesgo.DefaultRegistry(&rulesgo.DefaultConfig{}))
	handlers.SetRuleEngine(engine)

	// Create a simple chain
	chain := &rulesgo.Chain{
		ID:        "test_bridge",
		Name:      "Bridge Test",
		EntryNode: "c1",
		Nodes: []rulesgo.ChainNode{
			{ID: "c1", Type: "condition", Label: "Check", Config: json.RawMessage(`{"field":"x","operator":"eq","value":"1"}`)},
		},
	}
	engine.RegisterChain(chain)

	result, err := engine.Run(context.Background(), "test_bridge", map[string]interface{}{"x": 1})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}

	t.Logf("Bridge test passed: %d nodes, success=%v", len(result.Nodes), result.Success)
}

func TestBridge_DemoChains(t *testing.T) {
	persist := rulesgo.NewMemoryPersistence()
	engineWithPersist := rulesgo.NewEngineWithPersistence(
		rulesgo.DefaultRegistry(&rulesgo.DefaultConfig{}),
		persist,
	)
	handlers.SetRuleEngine(engineWithPersist.Engine)

	registerDemoChains(engineWithPersist)

	engine := engineWithPersist.Engine

	// Test welcome email chain
	result, err := engine.Run(context.Background(), "demo_welcome_email", map[string]interface{}{
		"name":  "Test User",
		"email": "test@example.com",
	})
	if err != nil {
		t.Fatalf("welcome email chain failed: %v", err)
	}
	if !result.Success {
		t.Errorf("welcome email chain should succeed, got: %s", result.Error)
	}
	t.Logf("Welcome email chain: %d nodes", len(result.Nodes))

	// Test lead scoring chain (hot lead)
	result, err = engine.Run(context.Background(), "demo_lead_scoring", map[string]interface{}{
		"name":    "Hot Corp",
		"company": "Hot Inc",
		"budget":  15000,
		"email":   "hot@corp.com",
	})
	if err != nil {
		t.Fatalf("lead scoring chain failed: %v", err)
	}
	t.Logf("Lead scoring chain: %d nodes, success=%v", len(result.Nodes), result.Success)

	// Test data cleanup chain
	result, err = engine.Run(context.Background(), "demo_data_cleanup", map[string]interface{}{
		"name":  "  Dirty User  ",
		"email": "DIRTY@example.com",
	})
	if err != nil {
		t.Fatalf("data cleanup chain failed: %v", err)
	}
	t.Logf("Data cleanup chain: %d nodes, success=%v", len(result.Nodes), result.Success)
}

func TestBridge_AllNodeTypes(t *testing.T) {
	registry := rulesgo.DefaultRegistry(&rulesgo.DefaultConfig{})
	engine := rulesgo.NewEngine(registry)
	handlers.SetRuleEngine(engine)

	nodeTests := []struct {
		nodeType string
		config   string
		input    map[string]interface{}
	}{
		{"condition", `{"field":"test","operator":"eq","value":"ok"}`, map[string]interface{}{"test": "ok"}},
		{"condition", `{"field":"count","operator":"gte","value":"5"}`, map[string]interface{}{"count": 10}},
		{"condition", `{"field":"name","operator":"notEmpty"}`, map[string]interface{}{"name": "Alice"}},
	}

	for _, tt := range nodeTests {
		t.Run(tt.nodeType+"_"+tt.config[:20], func(t *testing.T) {
			chain := &rulesgo.Chain{
				ID:        "test_" + tt.nodeType,
				Name:      "Test " + tt.nodeType,
				EntryNode: "n1",
				Nodes:     []rulesgo.ChainNode{{ID: "n1", Type: tt.nodeType, Label: "N1", Config: json.RawMessage(tt.config)}},
			}
			engine.RegisterChain(chain)

			result, err := engine.Run(context.Background(), "test_"+tt.nodeType, tt.input)
			if err != nil {
				t.Fatalf("run failed: %v", err)
			}
			if !result.Success {
				t.Errorf("node %s failed: %s", tt.nodeType, result.Error)
			}
		})
	}
}

func TestBridge_ChainImportExport(t *testing.T) {
	engine := rulesgo.NewEngine(rulesgo.DefaultRegistry(&rulesgo.DefaultConfig{}))
	handlers.SetRuleEngine(engine)

	chain := &rulesgo.Chain{
		ID:        "test_import",
		Name:      "Import Test",
		EntryNode: "start",
		Nodes: []rulesgo.ChainNode{
			{ID: "start", Type: "condition", Label: "Check", Config: json.RawMessage(`{"field":"ok","operator":"eq","value":"yes"}`)},
		},
	}
	engine.RegisterChain(chain)

	data, err := engine.ExportChain("test_import")
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	var exported rulesgo.Chain
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("invalid export: %v", err)
	}

	if exported.Name != "Import Test" {
		t.Fatalf("name mismatch: %s", exported.Name)
	}

	imported, err := engine.ImportChain(data)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if imported.ID != "test_import" {
		t.Fatalf("import ID mismatch: %s", imported.ID)
	}

	engine.DeleteChain("test_import")
	if engine.Chain("test_import") != nil {
		t.Fatal("chain should be deleted")
	}

	t.Logf("Export/Import/Delete OK")
}
