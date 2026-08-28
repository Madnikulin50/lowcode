package rulesgo

import (
	"context"
	"encoding/json"
	"testing"
)

func TestRemoteCatalogRegistered(t *testing.T) {
	r := DefaultRegistry(&DefaultConfig{})
	for _, spec := range RemoteCatalog() {
		if _, ok := r.Get(spec.Type); !ok {
			t.Errorf("missing %s", spec.Type)
		}
	}
	if _, ok := r.Get("service.call"); !ok {
		t.Fatal("service.call")
	}
}

func TestComponentExecutorNotConfigured(t *testing.T) {
	ex := &componentExecutor{spec: RemoteSpec{Service: "nosuch", Operation: "x"}}
	out, err := ex.Execute(context.Background(), ChainNode{Type: "nosuch/x", Config: json.RawMessage(`{}`)}, &ExecutionContext{
		Variables: map[string]interface{}{},
		Results:   map[string]interface{}{},
		Input:     map[string]interface{}{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out["status"] != "agent_not_configured" {
		t.Fatalf("%v", out)
	}
}
