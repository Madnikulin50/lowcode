package rulesgo

import (
	"context"
	"testing"
)

func wfNode(cfg string) ChainNode {
	return ChainNode{ID: "n", Type: "workflow", Config: []byte(cfg)}
}

func TestWorkflowNode(t *testing.T) {
	var gotRef string
	var gotIn map[string]interface{}
	var gotAsync bool
	n := &wfExecutor{exec: func(_ context.Context, ref string, in map[string]interface{}, async bool) (*WorkflowRunResult, error) {
		gotRef, gotIn, gotAsync = ref, in, async
		return &WorkflowRunResult{SessionID: "42", Results: map[string]interface{}{"ok": true}}, nil
	}}

	out, err := n.Execute(context.Background(),
		wfNode(`{"workflowID":"my_wf","payload":"{\"a\":1}","input":{"b":"x"},"async":true}`),
		&ExecutionContext{})
	if err != nil {
		t.Fatal(err)
	}
	if gotRef != "my_wf" || gotIn["b"] != "x" || gotIn["a"] != float64(1) || !gotAsync {
		t.Fatalf("ref=%q in=%v async=%v", gotRef, gotIn, gotAsync)
	}
	if out["status"] != "started" || out["sessionID"] != "42" {
		t.Fatalf("out=%v", out)
	}
}

func TestWorkflowNodeErrors(t *testing.T) {
	n := &wfExecutor{}
	if _, err := n.Execute(context.Background(), wfNode(`{}`), &ExecutionContext{}); err == nil {
		t.Fatal("expected workflowID required")
	}
	if _, err := n.Execute(context.Background(), wfNode(`{"workflowID":"1","payload":"[1]"}`), &ExecutionContext{}); err == nil {
		t.Fatal("expected bad payload error")
	}
	out, err := n.Execute(context.Background(), wfNode(`{"workflowID":"1"}`), &ExecutionContext{})
	if err != nil || out["status"] != "not_configured" {
		t.Fatalf("out=%v err=%v", out, err)
	}
}
