package rulesgo

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestAutomationCorrelate_ResolvesWithTemplatedKeyAndInput(t *testing.T) {
	var gotKey string
	var gotInput map[string]interface{}
	n := &automationCorrelateExecutor{
		resolve: func(ctx context.Context, key string, input map[string]interface{}) error {
			gotKey = key
			gotInput = input
			return nil
		},
	}

	node := ChainNode{Config: json.RawMessage(`{
		"key": "{{orderNumber}}",
		"input": {"status": "{{status}}"}
	}`)}
	ec := &ExecutionContext{Variables: map[string]interface{}{"orderNumber": "ORDER-42", "status": "confirmed"}}

	out, err := n.Execute(context.Background(), node, ec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotKey != "ORDER-42" {
		t.Fatalf("expected resolved key ORDER-42, got %q", gotKey)
	}
	if gotInput["status"] != "confirmed" {
		t.Fatalf("unexpected input: %#v", gotInput)
	}
	if out["success"] != true {
		t.Fatalf("unexpected output: %#v", out)
	}
}

func TestAutomationCorrelate_DefaultsInputToIngestEnvelope(t *testing.T) {
	var gotInput map[string]interface{}
	n := &automationCorrelateExecutor{
		resolve: func(ctx context.Context, key string, input map[string]interface{}) error {
			gotInput = input
			return nil
		},
	}

	node := ChainNode{Config: json.RawMessage(`{"key": "x"}`)}
	ec := &ExecutionContext{Input: map[string]interface{}{"value": "raw message body"}}

	if _, err := n.Execute(context.Background(), node, ec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotInput["value"] != "raw message body" {
		t.Fatalf("expected ingest envelope forwarded as input, got %#v", gotInput)
	}
}

func TestAutomationCorrelate_RequiresKey(t *testing.T) {
	n := &automationCorrelateExecutor{resolve: func(context.Context, string, map[string]interface{}) error { return nil }}
	if _, err := n.Execute(context.Background(), ChainNode{Config: json.RawMessage(`{}`)}, &ExecutionContext{}); err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestAutomationCorrelate_PropagatesResolveError(t *testing.T) {
	n := &automationCorrelateExecutor{resolve: func(context.Context, string, map[string]interface{}) error {
		return errNoMatch
	}}
	node := ChainNode{Config: json.RawMessage(`{"key":"x"}`)}
	if _, err := n.Execute(context.Background(), node, &ExecutionContext{}); err == nil {
		t.Fatal("expected the resolve error to propagate")
	}
}

func TestAutomationCorrelate_NotConfigured(t *testing.T) {
	n := &automationCorrelateExecutor{}
	out, err := n.Execute(context.Background(), ChainNode{Config: json.RawMessage(`{"key":"x"}`)}, &ExecutionContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["status"] != "not_configured" {
		t.Fatalf("unexpected output: %#v", out)
	}
}

var errNoMatch = errors.New("no pending prompt is waiting on key")
