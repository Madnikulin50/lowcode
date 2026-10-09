package rulesgo

import (
	"context"
	"encoding/json"
	"testing"
)

// A missing field must behave like an empty one for "empty"/"notEmpty" —
// ec.Get returns Go nil for a key that isn't set anywhere, and
// fmt.Sprintf("%v", nil) is the literal string "<nil>", which used to make
// notEmpty see a missing field as non-empty (e.g. invest-threshold-alert's
// "Есть email" gate letting the mail node run with an empty "to" and
// failing with "at least one recipient required").
func TestConditionNodeMissingFieldIsEmpty(t *testing.T) {
	ec := &ExecutionContext{Variables: map[string]interface{}{}}

	node := ChainNode{
		ID:     "check",
		Type:   "condition",
		Config: json.RawMessage(`{"field":"notifyEmail","operator":"notEmpty"}`),
	}

	exec := &conditionExecutor{}
	out, err := exec.Execute(context.Background(), node, ec)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if out["passed"] != "false" {
		t.Fatalf("passed = %v, want false for a missing field", out["passed"])
	}
	if got := ec.Variables["check_result"]; got != "" {
		t.Fatalf("check_result = %q, want empty string (so conditional edges skip the branch)", got)
	}
}

func TestConditionNodeEmptyFieldIsEmpty(t *testing.T) {
	ec := &ExecutionContext{Variables: map[string]interface{}{"notifyEmail": ""}}

	node := ChainNode{
		ID:     "check",
		Type:   "condition",
		Config: json.RawMessage(`{"field":"notifyEmail","operator":"notEmpty"}`),
	}

	exec := &conditionExecutor{}
	out, err := exec.Execute(context.Background(), node, ec)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if out["passed"] != "false" {
		t.Fatalf("passed = %v, want false for an explicitly empty field", out["passed"])
	}
}

func TestConditionNodeNotEmptyFieldPasses(t *testing.T) {
	ec := &ExecutionContext{Variables: map[string]interface{}{"notifyEmail": "ops@example.com"}}

	node := ChainNode{
		ID:     "check",
		Type:   "condition",
		Config: json.RawMessage(`{"field":"notifyEmail","operator":"notEmpty"}`),
	}

	exec := &conditionExecutor{}
	out, err := exec.Execute(context.Background(), node, ec)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if out["passed"] != "true" {
		t.Fatalf("passed = %v, want true when the field is set", out["passed"])
	}
	if got := ec.Get("check_result"); got != "true" {
		t.Fatalf("check_result = %q, want \"true\"", got)
	}
}
