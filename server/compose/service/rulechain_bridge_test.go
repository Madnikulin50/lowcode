package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/madnikulin50/lowcode/server/pkg/expr"
	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
)

func TestRunRuleChainHandler_RunsChainAndReturnsOutput(t *testing.T) {
	engine := rulesgo.NewEngineWithPersistence(rulesgo.DefaultRegistry(&rulesgo.DefaultConfig{}), rulesgo.NewMemoryPersistence())
	engine.RegisterChain(&rulesgo.Chain{
		ID:        "bpmn_bridge_test",
		Name:      "Bridge Test",
		EntryNode: "n1",
		Nodes: []rulesgo.ChainNode{
			{ID: "n1", Type: "condition", Config: json.RawMessage(`{"field":"x","operator":"eq","value":"1"}`)},
		},
	})

	restore := DefaultRuleEngine
	SetRuleEngine(engine)
	defer func() { DefaultRuleEngine = restore }()

	in := &expr.Vars{}
	if err := in.Set("chainID", "bpmn_bridge_test"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := in.Set("input", map[string]interface{}{"x": "1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := runRuleChainHandler(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !out.Has("success") || !out.Has("output") || !out.Has("error") {
		t.Fatalf("expected success/output/error in result, got: %#v", out.Dict())
	}
	success, _ := expr.Select(out, "success")
	if success.Get() != true {
		t.Fatalf("expected success=true, got %#v", success.Get())
	}
}

func TestRunRuleChainHandler_RequiresChainID(t *testing.T) {
	engine := rulesgo.NewEngineWithPersistence(rulesgo.DefaultRegistry(&rulesgo.DefaultConfig{}), rulesgo.NewMemoryPersistence())
	restore := DefaultRuleEngine
	SetRuleEngine(engine)
	defer func() { DefaultRuleEngine = restore }()

	if _, err := runRuleChainHandler(context.Background(), &expr.Vars{}); err == nil {
		t.Fatal("expected error for missing chainID")
	}
}

func TestRunRuleChainHandler_ErrorsWhenEngineNotConfigured(t *testing.T) {
	restore := DefaultRuleEngine
	DefaultRuleEngine = nil
	defer func() { DefaultRuleEngine = restore }()

	in := &expr.Vars{}
	in.Set("chainID", "x")
	if _, err := runRuleChainHandler(context.Background(), in); err == nil {
		t.Fatal("expected error when rule engine is not configured")
	}
}

func TestRunRuleChainHandler_ErrorsForUnknownChain(t *testing.T) {
	engine := rulesgo.NewEngineWithPersistence(rulesgo.DefaultRegistry(&rulesgo.DefaultConfig{}), rulesgo.NewMemoryPersistence())
	restore := DefaultRuleEngine
	SetRuleEngine(engine)
	defer func() { DefaultRuleEngine = restore }()

	in := &expr.Vars{}
	in.Set("chainID", "does_not_exist")
	if _, err := runRuleChainHandler(context.Background(), in); err == nil {
		t.Fatal("expected error for unknown chain")
	}
}

func TestRunRuleChainFunction_Definition(t *testing.T) {
	fn := RunRuleChainFunction()
	if fn.Ref != "compose.runRuleChain" {
		t.Fatalf("unexpected ref: %s", fn.Ref)
	}
	if fn.Kind != "function" {
		t.Fatalf("unexpected kind: %s", fn.Kind)
	}
	if fn.Handler == nil {
		t.Fatal("expected a handler")
	}
	if len(fn.Parameters) != 2 || len(fn.Results) != 3 {
		t.Fatalf("unexpected signature: %d params, %d results", len(fn.Parameters), len(fn.Results))
	}
}
