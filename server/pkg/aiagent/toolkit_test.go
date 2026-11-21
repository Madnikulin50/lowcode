package aiagent

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

func TestCatalogResolve(t *testing.T) {
	c := NewCatalog()
	c.Register(ToolKit{Name: "a", Tools: []chat.ToolDef{{Name: "one"}, {Name: "two"}}})
	c.Register(ToolKit{Name: "b", Tools: []chat.ToolDef{{Name: "two"}, {Name: "three"}}})
	got := namesOf(c.Resolve("a", "b"))
	if strings.Join(got, ",") != "one,two,three" {
		t.Fatalf("got %v", got)
	}
	star := namesOf(c.Resolve("*"))
	if len(star) != 3 {
		t.Fatalf("star=%v", star)
	}
	if len(c.Resolve("missing")) != 0 {
		t.Fatal("unknown kit should be empty")
	}
}

func TestRemoteKitFromMeta(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/meta" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"handle": "cmdb",
				"name":   "CMDB",
				"components": []map[string]any{
					{
						"type":        "cmdb/scan",
						"label":       "Scan",
						"description": "Scan CIDR",
						"service":     "cmdb",
						"operation":   "scan",
						"async":       true,
						"configFields": []map[string]any{
							{"key": "cidr", "label": "CIDR", "required": true},
						},
					},
				},
			})
		case r.URL.Path == "/jobs" && r.Method == http.MethodPost:
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"cidr":"10.0.0.0/24"`) {
				t.Errorf("body=%s", body)
			}
			_, _ = w.Write([]byte(`{"id":"job-1","status":"running"}`))
		case r.URL.Path == "/jobs/job-1":
			_, _ = w.Write([]byte(`{"id":"job-1","status":"completed"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	svc, err := FetchRemoteMeta(ctx, srv.URL, "")
	if err != nil {
		t.Fatal(err)
	}
	kit := RemoteKit(svc)
	if kit.Name != "cmdb" {
		t.Fatalf("name=%s", kit.Name)
	}
	var scan, status chat.ToolDef
	for _, tool := range kit.Tools {
		switch tool.Name {
		case "cmdb_scan":
			scan = tool
		case "cmdb_job_status":
			status = tool
		}
	}
	if scan.Handler == nil {
		t.Fatal("missing cmdb_scan")
	}
	out := scan.Handler(ctx, map[string]string{"cidr": "10.0.0.0/24"})
	if !strings.Contains(out, "job-1") {
		t.Fatalf("start=%s", out)
	}
	if status.Handler == nil {
		t.Fatal("missing job status")
	}
	st := status.Handler(ctx, map[string]string{"jobID": "job-1"})
	if !strings.Contains(st, "completed") {
		t.Fatalf("status=%s", st)
	}
}

func TestFallbackKitNames(t *testing.T) {
	c := NewCatalog()
	c.RegisterFallbacks()
	want := []string{"cmdb_scan", "backup_run", "backup_restore", "invest_evm"}
	got := namesOf(c.Resolve("cmdb", "backup", "invest"))
	set := map[string]bool{}
	for _, n := range got {
		set[n] = true
	}
	for _, n := range want {
		if !set[n] {
			t.Fatalf("missing %s in %v", n, got)
		}
	}
}

func TestDefaultNeedsConfirmRestore(t *testing.T) {
	if !DefaultNeedsConfirm([]Call{{Name: "backup_restore"}}) {
		t.Fatal("restore should confirm")
	}
	if DefaultNeedsConfirm([]Call{{Name: "cmdb_scan"}}) {
		t.Fatal("scan should not confirm")
	}
}

func TestInvestUsesCustomPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"spi":1}`))
	}))
	defer srv.Close()

	svc := RemoteService{
		Handle:  "invest",
		BaseURL: srv.URL,
		Components: []RemoteComponent{
			{Type: "invest/evm", Operation: "evm", Path: "/recalculate-evm", Fields: []RemoteField{{Key: "projectID"}}},
		},
	}
	kit := RemoteKit(svc)
	if kit.Tools[0].Handler == nil {
		t.Fatal("no handler")
	}
	out := kit.Tools[0].Handler(context.Background(), map[string]string{"projectID": "1"})
	if gotPath != "/recalculate-evm" {
		t.Fatalf("path=%s", gotPath)
	}
	if !strings.Contains(out, "spi") {
		t.Fatalf("out=%s", out)
	}
}

func namesOf(tools []chat.ToolDef) []string {
	out := make([]string, 0, len(tools))
	for _, t := range tools {
		out = append(out, t.Name)
	}
	return out
}
