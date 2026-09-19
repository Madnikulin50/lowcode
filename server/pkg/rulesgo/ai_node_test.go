package rulesgo

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestAINode_BlocksOnConfirmNeeded(t *testing.T) {
	n := &aiExecutor{
		call: func(ctx context.Context, agent, prompt, model string, allowMutating bool) (*AIOperationResult, error) {
			return &AIOperationResult{ConfirmNeeded: true, ConfirmCalls: []string{"delete_customer"}}, nil
		},
	}
	node := ChainNode{Config: json.RawMessage(`{"agent":"assistant","prompt":"clean up the CRM"}`)}
	_, err := n.Execute(context.Background(), node, &ExecutionContext{Variables: map[string]interface{}{}})
	if err == nil {
		t.Fatal("expected error when the agent's answer requires confirmation")
	}
	if !strings.Contains(err.Error(), "delete_customer") {
		t.Fatalf("expected error to name the blocked tool, got: %v", err)
	}
}

func TestAINode_PassesAllowMutatingThrough(t *testing.T) {
	var gotAllow bool
	n := &aiExecutor{
		call: func(ctx context.Context, agent, prompt, model string, allowMutating bool) (*AIOperationResult, error) {
			gotAllow = allowMutating
			return &AIOperationResult{Output: "done"}, nil
		},
	}
	node := ChainNode{Config: json.RawMessage(`{"agent":"assistant","prompt":"go ahead","allowMutating":true}`)}
	if _, err := n.Execute(context.Background(), node, &ExecutionContext{Variables: map[string]interface{}{}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gotAllow {
		t.Fatal("expected allowMutating=true to reach the call func")
	}
}

func TestAINode_DefaultsAllowMutatingFalse(t *testing.T) {
	var gotAllow = true // start true so we can tell if it was actually set to false
	n := &aiExecutor{
		call: func(ctx context.Context, agent, prompt, model string, allowMutating bool) (*AIOperationResult, error) {
			gotAllow = allowMutating
			return &AIOperationResult{Output: "done"}, nil
		},
	}
	node := ChainNode{Config: json.RawMessage(`{"agent":"assistant","prompt":"go ahead"}`)}
	if _, err := n.Execute(context.Background(), node, &ExecutionContext{Variables: map[string]interface{}{}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAllow {
		t.Fatal("expected allowMutating to default to false")
	}
}

func TestAINode_RequiresPrompt(t *testing.T) {
	n := &aiExecutor{call: func(context.Context, string, string, string, bool) (*AIOperationResult, error) {
		return &AIOperationResult{Output: "x"}, nil
	}}
	if _, err := n.Execute(context.Background(), ChainNode{Config: json.RawMessage(`{"agent":"assistant"}`)}, &ExecutionContext{Variables: map[string]interface{}{}}); err == nil {
		t.Fatal("expected error for missing prompt")
	}
}

func TestAINode_NotConfiguredWhenNoCallFunc(t *testing.T) {
	n := &aiExecutor{}
	out, err := n.Execute(context.Background(), ChainNode{Config: json.RawMessage(`{"agent":"assistant","prompt":"hi"}`)}, &ExecutionContext{Variables: map[string]interface{}{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["status"] != "not_configured" {
		t.Fatalf("unexpected output: %#v", out)
	}
}
