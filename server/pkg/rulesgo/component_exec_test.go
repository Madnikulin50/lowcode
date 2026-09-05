package rulesgo

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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

func TestComponentExecutorPostsJobsContract(t *testing.T) {
	var gotPath, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		gotBody = string(raw)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"j1","status":"running"}`))
	}))
	t.Cleanup(srv.Close)

	ex := &componentExecutor{spec: RemoteSpec{Service: "backup", Operation: "backup", Async: true, Ingest: "backup-ingest-job"}}
	ec := &ExecutionContext{
		Variables: map[string]interface{}{"createdRecordID": "77"},
		Results:   map[string]interface{}{},
		Input: map[string]interface{}{
			"sourceID":    "src-1",
			"namespaceID": "ns-1",
			"authToken":   "tok",
			"agentUrl":    srv.URL,
		},
	}
	out, err := ex.Execute(context.Background(), ChainNode{
		Type:   "backup/run",
		Config: json.RawMessage(`{"sourceID":"{{sourceID}}","jobID":"{{createdRecordID}}"}`),
	}, ec)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/jobs" {
		t.Fatalf("path %s", gotPath)
	}
	if out["jobID"] != "j1" {
		t.Fatalf("jobID %v", out["jobID"])
	}
	var payload map[string]string
	if err := json.Unmarshal([]byte(gotBody), &payload); err != nil {
		t.Fatalf("%s: %v", gotBody, err)
	}
	if payload["operation"] != "backup" || payload["sourceID"] != "src-1" || payload["jobID"] != "77" || payload["recordID"] != "77" {
		t.Fatalf("payload %#v", payload)
	}
	if payload["token"] != "tok" || payload["namespaceID"] != "ns-1" {
		t.Fatalf("auth/ns %#v", payload)
	}
}

func TestComponentExecutorNodeURLWinsOverContext(t *testing.T) {
	var gotHost string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHost = r.Host
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"1"}`))
	}))
	t.Cleanup(srv.Close)

	ex := &componentExecutor{spec: RemoteSpec{Service: "backup", Operation: "due"}}
	_, err := ex.Execute(context.Background(), ChainNode{
		Type:   "backup/due",
		Config: json.RawMessage(`{"url":"` + srv.URL + `"}`),
	}, &ExecutionContext{
		Variables: map[string]interface{}{},
		Input:     map[string]interface{}{"agentUrl": "http://127.0.0.1:1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotHost == "" {
		t.Fatal("node url was not used")
	}
}
