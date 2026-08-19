package agent

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestComposeJobStatus(t *testing.T) {
	if composeJobStatus("done") != "completed" {
		t.Fatalf("done")
	}
	if composeJobStatus("error") != "failed" {
		t.Fatalf("error")
	}
	if composeJobStatus("running") != "running" {
		t.Fatalf("running")
	}
}

func TestPostCallbackEnvelope(t *testing.T) {
	var got ingestEnvelope
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			t.Errorf("auth header %q", r.Header.Get("Authorization"))
		}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &got); err != nil {
			t.Errorf("json: %v", err)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer srv.Close()

	a := New(Config{}, nil)
	id := "scan-callback-1"
	now := time.Now()
	fin := now.Add(time.Second)
	a.mu.Lock()
	a.scans[id] = &ScanStatus{
		ID: id, Status: "done", Progress: 100, Found: 1,
		Target: "10.0.0.0/24", StartedAt: now, FinishedAt: &fin,
	}
	a.mu.Unlock()

	a.postCallback(ScanTarget{
		CallbackURL:  srv.URL,
		Token:        "tok",
		NamespaceID:  42,
		ScanRecordID: "99",
	}, id, "complete", []Device{{IP: "10.0.0.1", MAC: "aa:bb:cc:dd:ee:ff"}})

	if got.JobID != id || got.Kind != "complete" || got.Status != "completed" {
		t.Fatalf("envelope %+v", got)
	}
	if got.ScanRecordID != "99" || got.NamespaceID != 42 {
		t.Fatalf("ids %+v", got)
	}
	if len(got.Items) != 1 || got.Items[0].IP != "10.0.0.1" {
		t.Fatalf("items %+v", got.Items)
	}
}

func TestNotifyCallbackThrottlesProgress(t *testing.T) {
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n.Add(1)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	a := New(Config{}, nil)
	id := "scan-throttle"
	a.mu.Lock()
	a.scans[id] = &ScanStatus{ID: id, Status: "running"}
	a.mu.Unlock()
	tgt := ScanTarget{CallbackURL: srv.URL}

	a.notifyCallback(tgt, id, "progress", nil)
	a.notifyCallback(tgt, id, "progress", nil)
	time.Sleep(50 * time.Millisecond)
	if n.Load() != 1 {
		t.Fatalf("progress posts=%d want 1", n.Load())
	}
}
