package sdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUseSyncCallHTTP(t *testing.T) {
	svc := New(Config{Handle: "calc-evm", Name: "EVM", Listen: ":0"})
	svc.Register(Desc{D: Descriptor{Type: "calc/evm", Label: "EVM", Operation: "evm"}})
	svc.Sync("evm")
	svc.UseSync(func(_ context.Context, op string, req StartRequest) (any, error) {
		if op != "evm" {
			t.Fatalf("op %s", op)
		}
		if req.Param("projectID") != "7" {
			t.Fatalf("project %q", req.Param("projectID"))
		}
		return map[string]any{"spi": 1.0, "cpi": 0.9, "eac": 100, "ac": 50, "wbs": 2}, nil
	})

	ts := httptest.NewServer(svc.Router())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/call/evm", "application/json", strings.NewReader(`{"projectID":"7"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status %d", resp.StatusCode)
	}
	var m map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		t.Fatal(err)
	}
	if m["spi"] != 1.0 || m["wbs"] != float64(2) {
		t.Fatalf("%+v", m)
	}
}

func TestUseAsyncStartJob(t *testing.T) {
	svc := New(Config{Handle: "scan-cidr", Name: "Scan", Listen: ":0"})
	svc.Register(Desc{D: Descriptor{Type: "scan/cidr", Label: "Scan", Operation: "scan", Async: true}})
	done := make(chan struct{})
	svc.UseAsync(func(_ context.Context, j *Job) error {
		defer close(done)
		if j.Param("cidr") != "10.0.0.0/30" {
			t.Errorf("cidr %q", j.Param("cidr"))
		}
		j.SetItems([]map[string]any{{"ip": "10.0.0.1"}})
		j.SetResult(map[string]any{"found": 1})
		return nil
	})

	ts := httptest.NewServer(svc.Router())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/jobs", "application/json", strings.NewReader(`{"cidr":"10.0.0.0/30"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var started map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&started); err != nil {
		t.Fatal(err)
	}
	if started["status"] != "running" || started["operation"] != "scan" {
		t.Fatalf("start %+v", started)
	}
	<-done
	id, _ := started["id"].(string)
	resp, err = http.Get(ts.URL + "/api/jobs/" + id)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var doneEnv map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&doneEnv); err != nil {
		t.Fatal(err)
	}
	if doneEnv["status"] != "completed" {
		t.Fatalf("done %+v", doneEnv)
	}
}
