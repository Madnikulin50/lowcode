package rulesgo

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestExtractJSONObject(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
		ok    bool
	}{
		{name: "plain object", input: `{"a":1}`, want: `{"a":1}`, ok: true},
		{name: "wrapped in prose and fences", input: "Sure, here you go:\n```json\n{\"a\":1}\n```\nHope that helps!", want: `{"a":1}`, ok: true},
		{name: "nested braces", input: `noise {"a":{"b":2}} trailing`, want: `{"a":{"b":2}}`, ok: true},
		{name: "brace inside string literal", input: `{"a":"x}y"}`, want: `{"a":"x}y"}`, ok: true},
		{name: "no object at all", input: "no json here", want: "", ok: false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := extractJSONObject(c.input)
			if ok != c.ok {
				t.Fatalf("ok=%v, want %v (got %q)", ok, c.ok, got)
			}
			if ok && got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestValidateOutputSchema(t *testing.T) {
	schema := map[string]string{"risk": "number", "reason": "string", "flags": "array"}

	ok := map[string]interface{}{"risk": 3.0, "reason": "x", "flags": []interface{}{"a"}}
	if err := validateOutputSchema(ok, schema); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	missing := map[string]interface{}{"risk": 3.0, "reason": "x"}
	if err := validateOutputSchema(missing, schema); err == nil {
		t.Fatal("expected error for missing field")
	}

	wrongType := map[string]interface{}{"risk": "not-a-number", "reason": "x", "flags": []interface{}{}}
	if err := validateOutputSchema(wrongType, schema); err == nil {
		t.Fatal("expected error for wrong type")
	}
}

func TestAIOperation_HappyPath(t *testing.T) {
	var gotPrompt string
	n := &aiOperationExecutor{
		call: func(ctx context.Context, agent, prompt, model string, allowMutating bool) (*AIOperationResult, error) {
			gotPrompt = prompt
			return &AIOperationResult{Output: `{"risk": 7, "reason": "late payments"}`}, nil
		},
	}

	node := ChainNode{Config: json.RawMessage(`{
		"agent": "assistant",
		"prompt": "Assess churn risk",
		"inputs": {"customerName": "{{name}}"},
		"outputSchema": {"risk": "number", "reason": "string"}
	}`)}
	ec := &ExecutionContext{Variables: map[string]interface{}{"name": "Acme"}}

	out, err := n.Execute(context.Background(), node, ec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotPrompt, `"customerName": "Acme"`) {
		t.Fatalf("expected resolved inputs in prompt, got: %s", gotPrompt)
	}
	if !strings.Contains(gotPrompt, `"risk": number`) {
		t.Fatalf("expected output schema rendered into prompt, got: %s", gotPrompt)
	}
	result, _ := out["result"].(map[string]interface{})
	if result["risk"] != 7.0 || result["reason"] != "late payments" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if out["raw"] != `{"risk": 7, "reason": "late payments"}` {
		t.Fatalf("expected raw output preserved, got: %#v", out["raw"])
	}
}

func TestAIOperation_RetriesOnInvalidJSON(t *testing.T) {
	calls := 0
	n := &aiOperationExecutor{
		call: func(ctx context.Context, agent, prompt, model string, allowMutating bool) (*AIOperationResult, error) {
			calls++
			if calls == 1 {
				return &AIOperationResult{Output: "sorry, I can't do that"}, nil
			}
			return &AIOperationResult{Output: `{"ok": true}`}, nil
		},
	}
	node := ChainNode{Config: json.RawMessage(`{"agent":"assistant","prompt":"do it","outputSchema":{"ok":"boolean"},"maxRetries":2}`)}

	out, err := n.Execute(context.Background(), node, &ExecutionContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls (1 retry), got %d", calls)
	}
	result, _ := out["result"].(map[string]interface{})
	if result["ok"] != true {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestAIOperation_FailsAfterExhaustingRetries(t *testing.T) {
	calls := 0
	n := &aiOperationExecutor{
		call: func(ctx context.Context, agent, prompt, model string, allowMutating bool) (*AIOperationResult, error) {
			calls++
			return &AIOperationResult{Output: "never valid json"}, nil
		},
	}
	node := ChainNode{Config: json.RawMessage(`{"agent":"assistant","prompt":"do it","maxRetries":1}`)}
	if _, err := n.Execute(context.Background(), node, &ExecutionContext{}); err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if calls != 2 { // 1 initial + 1 retry
		t.Fatalf("expected 2 attempts with maxRetries=1, got %d", calls)
	}
}

func TestAIOperation_ValidatesOutputSchemaTypeMismatch(t *testing.T) {
	calls := 0
	n := &aiOperationExecutor{
		call: func(ctx context.Context, agent, prompt, model string, allowMutating bool) (*AIOperationResult, error) {
			calls++
			return &AIOperationResult{Output: `{"risk": "high"}`}, nil // should be a number
		},
	}
	node := ChainNode{Config: json.RawMessage(`{"agent":"assistant","prompt":"assess","outputSchema":{"risk":"number"},"maxRetries":0}`)}
	if _, err := n.Execute(context.Background(), node, &ExecutionContext{}); err == nil {
		t.Fatal("expected error for type mismatch")
	}
	if calls != 1 {
		t.Fatalf("expected exactly 1 call with maxRetries=0, got %d", calls)
	}
}

func TestAIOperation_BlockedOnConfirmNeeded(t *testing.T) {
	n := &aiOperationExecutor{
		call: func(ctx context.Context, agent, prompt, model string, allowMutating bool) (*AIOperationResult, error) {
			return &AIOperationResult{ConfirmNeeded: true, ConfirmCalls: []string{"delete_customer"}}, nil
		},
	}
	node := ChainNode{Config: json.RawMessage(`{"agent":"assistant","prompt":"clean up"}`)}
	_, err := n.Execute(context.Background(), node, &ExecutionContext{})
	if err == nil {
		t.Fatal("expected error when ConfirmNeeded")
	}
	if !strings.Contains(err.Error(), "delete_customer") {
		t.Fatalf("expected error to mention the blocked tool, got: %v", err)
	}
}

func TestAIOperation_PassesAllowMutatingThrough(t *testing.T) {
	var gotAllow bool
	n := &aiOperationExecutor{
		call: func(ctx context.Context, agent, prompt, model string, allowMutating bool) (*AIOperationResult, error) {
			gotAllow = allowMutating
			return &AIOperationResult{Output: `{}`}, nil
		},
	}
	node := ChainNode{Config: json.RawMessage(`{"agent":"assistant","prompt":"go","allowMutating":true}`)}
	if _, err := n.Execute(context.Background(), node, &ExecutionContext{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gotAllow {
		t.Fatal("expected allowMutating=true to be passed through to the call func")
	}
}

func TestAIOperation_RequiresAgentAndPrompt(t *testing.T) {
	n := &aiOperationExecutor{call: func(context.Context, string, string, string, bool) (*AIOperationResult, error) {
		return &AIOperationResult{Output: "{}"}, nil
	}}
	if _, err := n.Execute(context.Background(), ChainNode{Config: json.RawMessage(`{"prompt":"x"}`)}, &ExecutionContext{}); err == nil {
		t.Fatal("expected error for missing agent")
	}
	if _, err := n.Execute(context.Background(), ChainNode{Config: json.RawMessage(`{"agent":"x"}`)}, &ExecutionContext{}); err == nil {
		t.Fatal("expected error for missing prompt")
	}
}

func TestAIOperation_NotConfiguredWhenNoCallFunc(t *testing.T) {
	n := &aiOperationExecutor{}
	out, err := n.Execute(context.Background(), ChainNode{Config: json.RawMessage(`{"agent":"x","prompt":"y"}`)}, &ExecutionContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["status"] != "not_configured" {
		t.Fatalf("unexpected output: %#v", out)
	}
}
