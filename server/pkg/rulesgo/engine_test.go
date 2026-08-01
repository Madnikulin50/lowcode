package rulesgo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestEngine_Run_Linear(t *testing.T) {
	registry := NewRegistry()
	registry.Register("condition", &conditionExecutor{})

	engine := NewEngine(registry)

	chain := &Chain{
		ID:        "test_linear",
		Name:      "Test Linear",
		EntryNode: "check_name",
		Nodes: []ChainNode{
			{ID: "check_name", Type: "condition", Label: "Has Name?", Config: makeCondCfg("name", "notEmpty", "")},
		},
		Edges: []ChainEdge{},
	}

	engine.RegisterChain(chain)

	result, err := engine.Run(context.Background(), "test_linear", map[string]interface{}{
		"name": "Alice",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got: %+v", result)
	}
	if len(result.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result.Nodes))
	}

	t.Logf("Result: %+v", result)
}

func TestEngine_Run_MultiNode(t *testing.T) {
	registry := NewRegistry()
	registry.Register("condition", &conditionExecutor{})

	engine := NewEngine(registry)

	chain := &Chain{
		ID:        "test_multi",
		Name:      "Test Multi-Node",
		EntryNode: "step1",
		Nodes: []ChainNode{
			{ID: "step1", Type: "condition", Label: "Step 1", Config: makeCondCfg("age", "gte", "18")},
			{ID: "step2", Type: "condition", Label: "Step 2", Config: makeCondCfg("name", "notEmpty", "")},
			{ID: "step3", Type: "condition", Label: "Step 3", Config: makeCondCfg("email", "contains", "@")},
		},
		Edges: []ChainEdge{
			{From: "step1", To: "step2"},
			{From: "step2", To: "step3"},
		},
	}

	engine.RegisterChain(chain)

	result, err := engine.Run(context.Background(), "test_multi", map[string]interface{}{
		"age":   25,
		"name":  "Bob",
		"email": "bob@example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success")
	}
	if len(result.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(result.Nodes))
	}

	t.Logf("Result: %+v", result)
}

func TestEngine_Run_ConditionalBranch(t *testing.T) {
	registry := NewRegistry()
	registry.Register("condition", &conditionExecutor{})

	engine := NewEngine(registry)

	chain := &Chain{
		ID:        "test_branch",
		Name:      "Test Branch",
		EntryNode: "check",
		Nodes: []ChainNode{
			{ID: "check", Type: "condition", Label: "Check budget", Config: makeCondCfg("budget", "gte", "10000")},
			{ID: "hot", Type: "condition", Label: "Hot lead", Config: makeCondCfg("type", "eq", "hot")},
			{ID: "cold", Type: "condition", Label: "Cold lead", Config: makeCondCfg("type", "eq", "cold")},
		},
		Edges: []ChainEdge{
			{From: "check", To: "hot", Condition: "check_result"},
		},
	}

	engine.RegisterChain(chain)

	result, err := engine.Run(context.Background(), "test_branch", map[string]interface{}{
		"budget": 15000,
		"type":   "hot",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success")
	}

	// Should visit check → hot (because budget >= 10000 → check_result = "true")
	if len(result.Nodes) < 2 {
		t.Fatalf("expected at least 2 nodes (check + hot), got %d", len(result.Nodes))
	}

	nodeIDs := make([]string, len(result.Nodes))
	for i, n := range result.Nodes {
		nodeIDs[i] = n.NodeID
	}
	t.Logf("Visited: %v", nodeIDs)
}

func TestEngine_Run_TemplateVariables(t *testing.T) {
	registry := NewRegistry()
	registry.Register("condition", &conditionExecutor{})

	engine := NewEngine(registry)

	chain := &Chain{
		ID:        "test_template",
		Name:      "Test Templates",
		EntryNode: "greet",
		Nodes: []ChainNode{
			{ID: "greet", Type: "condition", Label: "Greet", Config: makeCondCfg("name", "notEmpty", "")},
		},
		Edges: []ChainEdge{},
	}

	engine.RegisterChain(chain)

	result, err := engine.Run(context.Background(), "test_template", map[string]interface{}{
		"name":     "{{ctx.user}}",
		"ctx.user": "Charlie",
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	t.Logf("Result: %+v", result)
}

func TestEngine_ImportExport(t *testing.T) {
	registry := NewRegistry()
	registry.Register("condition", &conditionExecutor{})
	engine := NewEngine(registry)

	original := &Chain{
		ID:        "test_export",
		Name:      "Export Test",
		EntryNode: "start",
		Nodes: []ChainNode{
			{ID: "start", Type: "condition", Label: "Start", Config: makeCondCfg("ok", "eq", "true")},
		},
	}

	data, err := engine.ExportChain("test_export")
	if err == nil {
		t.Error("expected error for non-existent chain")
	}

	engine.RegisterChain(original)

	data, err = engine.ExportChain("test_export")
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	var exported Chain
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("invalid export JSON: %v", err)
	}
	if exported.ID != "test_export" {
		t.Fatalf("ID mismatch: %s", exported.ID)
	}

	imported, err := engine.ImportChain(data)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if imported.ID != "test_export" {
		t.Fatalf("import ID mismatch: %s", imported.ID)
	}

	t.Logf("Export/Import OK")
}

func TestEngine_DeleteChain(t *testing.T) {
	registry := NewRegistry()
	registry.Register("condition", &conditionExecutor{})
	engine := NewEngine(registry)

	chain := &Chain{
		ID:        "test_delete",
		Name:      "Delete Test",
		EntryNode: "n1",
		Nodes:     []ChainNode{{ID: "n1", Type: "condition", Label: "N1", Config: makeCondCfg("x", "eq", "1")}},
	}
	engine.RegisterChain(chain)

	if engine.Chain("test_delete") == nil {
		t.Fatal("chain should exist")
	}

	engine.DeleteChain("test_delete")

	if engine.Chain("test_delete") != nil {
		t.Fatal("chain should be deleted")
	}

	if len(engine.Chains()) != 0 {
		t.Fatalf("expected 0 chains, got %d", len(engine.Chains()))
	}
}

func TestEngine_NodeTypeNotFound(t *testing.T) {
	registry := NewRegistry()
	engine := NewEngine(registry)

	chain := &Chain{
		ID:        "test_unknown",
		Name:      "Unknown Node",
		EntryNode: "n1",
		Nodes:     []ChainNode{{ID: "n1", Type: "unknown_type", Label: "N1"}},
	}
	engine.RegisterChain(chain)

	result, err := engine.Run(context.Background(), "test_unknown", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Fatal("expected failure for unknown node type")
	}
	if !strings.Contains(result.Error, "unknown node type") {
		t.Fatalf("expected 'unknown node type' in error, got: %s", result.Error)
	}
}

func TestEngine_Persist(t *testing.T) {
	persist := NewMemoryPersistence()

	registry := NewRegistry()
	registry.Register("condition", &conditionExecutor{})

	engine := NewEngineWithPersistence(registry, persist)

	chain := &Chain{
		ID:        "test_persist",
		Name:      "Persist Test",
		EntryNode: "n1",
		Nodes:     []ChainNode{{ID: "n1", Type: "condition", Label: "N1", Config: makeCondCfg("x", "eq", "1")}},
	}
	engine.CreateChain(context.Background(), chain)

	saved, err := persist.LoadChains(context.Background())
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(saved) != 1 || saved[0].ID != "test_persist" {
		t.Fatalf("persistence mismatch: %+v", saved)
	}

	engine.DeleteChain(context.Background(), "test_persist")
	saved, _ = persist.LoadChains(context.Background())
	if len(saved) != 0 {
		t.Fatalf("expected 0 after delete, got %d", len(saved))
	}
}

func TestEngine_ExecutionLog(t *testing.T) {
	persist := NewMemoryPersistence()
	registry := NewRegistry()
	registry.Register("condition", &conditionExecutor{})
	engine := NewEngineWithPersistence(registry, persist)

	chain := &Chain{
		ID:        "test_log",
		Name:      "Log Test",
		EntryNode: "n1",
		Nodes:     []ChainNode{{ID: "n1", Type: "condition", Label: "N1", Config: makeCondCfg("x", "eq", "1")}},
	}
	engine.CreateChain(context.Background(), chain)

	_, err := engine.RunWithLog(context.Background(), "test_log", map[string]interface{}{"x": 1}, "manual")
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	logs := engine.ExecutionLogs(10)
	if len(logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(logs))
	}
	if logs[0].TriggerType != "manual" {
		t.Fatalf("expected manual trigger, got %s", logs[0].TriggerType)
	}

	t.Logf("Log entry: duration=%s, trigger=%s", logs[0].Duration, logs[0].TriggerType)
}

func makeCondCfg(field, operator, value string) []byte {
	cfg := conditionConfig{
		Field:    field,
		Operator: operator,
		Value:    value,
	}
	data, _ := json.Marshal(cfg)
	return data
}

func ExampleEngine() {
	registry := NewRegistry()
	registry.Register("condition", &conditionExecutor{})

	engine := NewEngine(registry)

	chain := &Chain{
		ID:        "example_chain",
		Name:      "Example",
		EntryNode: "greeting_check",
		Nodes: []ChainNode{
			{ID: "greeting_check", Type: "condition", Label: "Has Name?", Config: makeCondCfg("name", "notEmpty", "")},
		},
		Edges: []ChainEdge{},
	}
	engine.RegisterChain(chain)

	result, _ := engine.Run(context.Background(), "example_chain", map[string]interface{}{
		"name": "World",
	})

	fmt.Printf("Success: %v\n", result.Success)
	fmt.Printf("Nodes: %d\n", len(result.Nodes))
	// Output:
	// Success: true
	// Nodes: 1
}
