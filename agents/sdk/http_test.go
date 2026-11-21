package sdk

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestCallbackThrottleAndEnvelope(t *testing.T) {
	var n atomic.Int32
	var last []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n.Add(1)
		last, _ = io.ReadAll(r.Body)
		if r.Header.Get("Authorization") != "Bearer tok" {
			t.Errorf("auth %q", r.Header.Get("Authorization"))
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	cb := NewCallback()
	j := &Job{
		ID:          "scan-1",
		Operation:   "scan",
		Status:      StatusRunning,
		CallbackURL: srv.URL,
		Token:       "tok",
		RecordID:    "99",
		NamespaceID: 42,
		svc:         &Service{cfg: Config{Handle: "cmdb"}},
	}
	cb.Notify(j, KindProgress)
	cb.Notify(j, KindProgress)
	time.Sleep(50 * time.Millisecond)
	if n.Load() != 1 {
		t.Fatalf("progress posts=%d", n.Load())
	}
	j.Status = StatusCompleted
	j.Progress = 100
	cb.Notify(j, KindComplete)
	time.Sleep(50 * time.Millisecond)
	if n.Load() != 2 {
		t.Fatalf("complete posts=%d", n.Load())
	}
	var raw map[string]any
	if err := json.Unmarshal(last, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["scanRecordID"] != "99" || raw["status"] != "completed" || raw["kind"] != "complete" {
		t.Fatalf("callback body %+v", raw)
	}
}

type testBackend struct {
	jobs map[string]*Envelope
}

func (t *testBackend) StartJob(_ context.Context, req StartRequest) (*Envelope, error) {
	id := "job-1"
	env := &Envelope{
		Service: "cmdb", Operation: firstNonEmpty(req.Operation, "scan"),
		ID: id, JobID: id, Status: StatusRunning, Kind: KindProgress,
		NamespaceID: req.NamespaceID.String(), RecordID: req.RecordID,
		Result: map[string]any{"cidr": req.Param("cidr")},
	}
	if t.jobs == nil {
		t.jobs = map[string]*Envelope{}
	}
	t.jobs[id] = env
	return env, nil
}

func (t *testBackend) GetJob(id string) *Envelope { return t.jobs[id] }
func (t *testBackend) ListJobs() []*Envelope {
	var out []*Envelope
	for _, e := range t.jobs {
		out = append(out, e)
	}
	return out
}

func TestMetaAndJobsHTTP(t *testing.T) {
	be := &testBackend{jobs: map[string]*Envelope{}}
	svc := New(Config{Handle: "cmdb", Name: "CMDB", Listen: ":0"})
	svc.SetBackend(be)
	svc.Register(Desc{D: Descriptor{
		Type: "cmdb/scan", Label: "Scan", Category: CategoryAction,
		Execution: ExecRemote, Async: true, Operation: "scan",
		ConfigFields: []Field{{Key: "cidr", Widget: "string", Label: "CIDR", Required: true, Template: true}},
	}})
	ts := httptest.NewServer(svc.Router())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/meta")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var meta Meta
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		t.Fatal(err)
	}
	if meta.Handle != "cmdb" || len(meta.Components) != 1 || meta.Components[0].Type != "cmdb/scan" {
		t.Fatalf("meta %+v", meta)
	}

	resp, err = http.Post(ts.URL+"/api/jobs", "application/json", strings.NewReader(`{"cidr":"10.0.0.0/24","namespaceID":"42"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["status"] != "running" || m["operation"] != "scan" || m["id"] != "job-1" {
		t.Fatalf("start %s", raw)
	}
}
