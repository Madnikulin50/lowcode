package agent

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestComposeOriginCandidates(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"http://localhost:3333/api", []string{"http://localhost:3333", "http://localhost:3333/api"}},
		{"http://127.0.0.1:3333", []string{"http://127.0.0.1:3333", "http://127.0.0.1:3333/api"}},
		{"http://127.0.0.1:3333/compose", []string{"http://127.0.0.1:3333", "http://127.0.0.1:3333/api", "http://127.0.0.1:3333/compose"}},
		{"http://127.0.0.1:3333/api/compose", []string{"http://127.0.0.1:3333", "http://127.0.0.1:3333/api", "http://127.0.0.1:3333/api/compose"}},
	}
	for _, tc := range cases {
		got := composeOriginCandidates(tc.in)
		if len(got) != len(tc.want) {
			t.Fatalf("%s: got %v want %v", tc.in, got, tc.want)
		}
		for i := range tc.want {
			if got[i] != tc.want[i] {
				t.Fatalf("%s: got %v want %v", tc.in, got, tc.want)
			}
		}
	}
}

func TestAPIErrorBodyTruncatesHTML(t *testing.T) {
	raw := []byte("<!doctype html>\n<html lang=\"en\">\n<head><link href=\"https://cdn.jsdelivr.net/npm/bootstrap@4.6.0/dist/css/bootstrap.min.css\"")
	msg := apiErrorBody(raw)
	if strings.Contains(msg, "bootstrap") || strings.Contains(msg, "<html") {
		t.Fatalf("html leaked: %s", msg)
	}
	if !strings.Contains(msg, "--api=") {
		t.Fatalf("hint missing: %s", msg)
	}
}

func TestDiscoverPicksComposeNotAPIPrefix(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/compose/namespace/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":{"set":[{"namespaceID":"1","slug":"invest"}]}}`))
	})
	mux.HandleFunc("/api/compose/namespace/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!doctype html><html lang="en"><head><link href="https://cdn.jsdelivr.net/npm/bootstrap@4.6.0/dist/css/bootstrap.min.css"`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cz := NewCorteza(srv.URL+"/api", "", 0, time.Second)
	if err := cz.Discover(context.Background()); err != nil {
		t.Fatal(err)
	}
	if cz.BaseURL() != srv.URL {
		t.Fatalf("origin %s want %s", cz.BaseURL(), srv.URL)
	}
	raw, err := cz.request(context.Background(), "GET", "/compose/namespace/?limit=1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(raw, []byte("invest")) {
		t.Fatalf("body %s", raw)
	}
}

func TestLoadWBSUsesDiscoveredOrigin(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/compose/namespace/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path
		switch {
		case strings.Contains(path, "/module/") && strings.Contains(path, "/record/"):
			_, _ = w.Write([]byte(`{"response":{"set":[{"recordID":"9","values":[{"name":"code","value":"1.1"},{"name":"name","value":"Stage"},{"name":"project","value":"42"}]}]}}`))
		case strings.Contains(path, "/module/"):
			_, _ = w.Write([]byte(`{"response":{"set":[{"moduleID":"7","handle":"wbs_items"}]}}`))
		default:
			_, _ = w.Write([]byte(`{"response":{"set":[{"namespaceID":"1","slug":"invest"}]}}`))
		}
	})
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`<!doctype html><html lang="en"><head></head><body>nope</body></html>`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cz := NewCorteza(srv.URL+"/api", "", 0, time.Second)
	items, err := cz.LoadWBS(context.Background(), "42")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Code != "1.1" || items[0].ProjectID != "42" {
		t.Fatalf("items %+v", items)
	}
}

func TestCortezaJSONErrorOnHTTP200(t *testing.T) {
	raw := []byte(`{"error":{"message":"not allowed to read this module","meta":{"type":"notAllowedToRead"}}}`)
	got := cortezaJSONError(raw)
	if got != "not allowed to read this module" {
		t.Fatalf("got %q", got)
	}
	if cortezaJSONError([]byte(`{"response":{"set":[]}}`)) != "" {
		t.Fatal("success body must not look like an error")
	}
	detailed := []byte(`{"error":{"message":"3 issue(s) found","details":[{"kind":"empty","meta":{"field":"name"}},{"kind":"empty","meta":{"field":"code"}},{"kind":"empty","meta":{"field":"project"}}]}}`)
	got = cortezaJSONError(detailed)
	if !strings.Contains(got, "name") || !strings.Contains(got, "code") || !strings.Contains(got, "project") {
		t.Fatalf("details missing: %q", got)
	}
}

func TestOverlayRecordValuesKeepsRequiredFields(t *testing.T) {
	existing := []composeRecordValue{
		{Name: "project", Value: "42"},
		{Name: "code", Value: "1"},
		{Name: "name", Value: "Stage"},
		{Name: "spi", Value: ""},
	}
	got := overlayRecordValues(existing, map[string]string{"spi": "1.0000", "cpi": "0.9500"})
	m := map[string]string{}
	for _, v := range got {
		m[v.Name] = v.Value
	}
	if m["project"] != "42" || m["code"] != "1" || m["name"] != "Stage" {
		t.Fatalf("lost required fields: %+v", got)
	}
	if m["spi"] != "1.0000" || m["cpi"] != "0.9500" {
		t.Fatalf("metrics not applied: %+v", got)
	}
}

func TestSaveProjectEVMRejectsEmptyID(t *testing.T) {
	cz := NewCorteza("http://localhost", "", 0, time.Second)
	if err := cz.SaveProjectEVM(context.Background(), "", EVMResult{SPI: 1}); err == nil {
		t.Fatal("expected error for empty projectID")
	}
	if err := cz.SaveProjectEVM(context.Background(), "{{projectID}}", EVMResult{}); err == nil {
		t.Fatal("expected error for placeholder projectID")
	}
}
