package agent

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMatchCron(t *testing.T) {
	tm := time.Date(2026, 8, 21, 2, 0, 0, 0, time.Local)
	if !MatchCron("0 2 * * *", tm) {
		t.Fatal("daily 02:00 should match")
	}
	if MatchCron("0 3 * * *", tm) {
		t.Fatal("03:00 should not match 02:00")
	}
	if !MatchCron("@daily", time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)) {
		t.Fatal("@daily midnight")
	}
	if !MatchCron("*/15 * * * *", time.Date(2026, 1, 1, 10, 30, 0, 0, time.Local)) {
		t.Fatal("*/15 at :30")
	}
	if CronValid("0 2 * * *") != nil || CronValid("@hourly") != nil {
		t.Fatal("valid cron")
	}
	if CronValid("bad") == nil {
		t.Fatal("invalid cron")
	}
}

func TestParseUNC(t *testing.T) {
	h, s, p, err := ParseUNC(`\\fileserver\share\docs`, "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if h != "fileserver" || s != "share" || p != "docs" {
		t.Fatalf("got %s %s %s", h, s, p)
	}
	h, s, p, err = ParseUNC("smb://nas/backup/data", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if h != "nas" || s != "backup" || p != "data" {
		t.Fatalf("smb url got %s %s %s", h, s, p)
	}
	h, s, p, err = ParseUNC("", "host", "shr", "a/b")
	if err != nil || h != "host" || s != "shr" || p != "a/b" {
		t.Fatalf("fields: %s %s %s %v", h, s, p, err)
	}
}

func TestResolveSecret(t *testing.T) {
	t.Setenv("BACKUP_SECRET_lab", "s3cret")
	if ResolveSecret("lab") != "s3cret" {
		t.Fatal("handle")
	}
	if ResolveSecret("") != "" {
		t.Fatal("empty")
	}
}

func TestObjectKey(t *testing.T) {
	k := ObjectKey("demo fs", "job1", "archive.tar.gz")
	if !bytes.Contains([]byte(k), []byte("demo-fs")) {
		t.Fatalf("key %s", k)
	}
	if filepath.Base(k) != "archive.tar.gz" {
		t.Fatalf("base %s", k)
	}
}

func TestPackUnpackTarGz(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "hello.txt"), []byte("hello backup"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(src, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "n.txt"), []byte("nested"), 0o644); err != nil {
		t.Fatal(err)
	}
	walker, err := NewLocalFS(src)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	files, n, err := PackTarGz(context.Background(), &buf, walker, nil)
	if err != nil {
		t.Fatal(err)
	}
	if files != 2 || n == 0 {
		t.Fatalf("files=%d n=%d", files, n)
	}
	got, err := UnpackTarGz(context.Background(), bytes.NewReader(buf.Bytes()), dst)
	if err != nil {
		t.Fatal(err)
	}
	if got != 2 {
		t.Fatalf("unpacked %d", got)
	}
	body, err := os.ReadFile(filepath.Join(dst, "hello.txt"))
	if err != nil || string(body) != "hello backup" {
		t.Fatalf("hello: %s %v", body, err)
	}
	body, err = os.ReadFile(filepath.Join(dst, "sub", "n.txt"))
	if err != nil || string(body) != "nested" {
		t.Fatalf("nested: %s %v", body, err)
	}
}

func TestSafeJoin(t *testing.T) {
	root := t.TempDir()
	if _, err := safeJoin(root, "../etc/passwd"); err == nil {
		t.Fatal("expected traversal error")
	}
}

func TestCleanRecordID(t *testing.T) {
	if CleanRecordID("{{recordID}}") != "" || CleanRecordID("${recordID}") != "" || CleanRecordID("0") != "" {
		t.Fatal("placeholders")
	}
	if CleanRecordID("510291663494250497") != "510291663494250497" {
		t.Fatal("id")
	}
	req := JobRequest{SourceID: "{{recordID}}", Source: "123"}
	if req.ResolvedSourceID() != "123" {
		t.Fatalf("alias %s", req.ResolvedSourceID())
	}
}

func TestParseBool(t *testing.T) {
	if !ParseBool("1") || !ParseBool("true") || ParseBool("0") {
		t.Fatal("bool")
	}
}

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
			if i >= len(got) || got[i] != tc.want[i] {
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
		_, _ = w.Write([]byte(`{"response":{"set":[]}}`))
	})
	mux.HandleFunc("/api/compose/namespace/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!doctype html><html lang="en"><head><link href="https://cdn.jsdelivr.net/npm/bootstrap@4.6.0/dist/css/bootstrap.min.css"`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cz := NewCorteza(srv.URL+"/api", "", 0)
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
	if !bytes.Contains(raw, []byte(`"set"`)) && !bytes.Contains(raw, []byte("set")) {
		t.Fatalf("body %s", raw)
	}
}

func TestComposeAPIErrorOnHTTP200(t *testing.T) {
	raw := []byte(`{"error":{"message":"not allowed to search or list namespaces","meta":{"type":"notAllowedToSearch"}}}`)
	if got := composeAPIError(raw); got != "not allowed to search or list namespaces" {
		t.Fatalf("got %q", got)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/compose/namespace/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(raw)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	cz := NewCorteza(srv.URL, "", 0)
	_, err := cz.request(context.Background(), "GET", "/compose/namespace/?limit=1", nil)
	if err == nil || !strings.Contains(err.Error(), "not allowed") {
		t.Fatalf("want API error, got %v", err)
	}
}

func TestParseResticSnapshotID(t *testing.T) {
	raw := []byte(`{"message_type":"status"}
{"message_type":"summary","snapshot_id":"abc123"}
`)
	if parseResticSnapshotID(raw) != "abc123" {
		t.Fatal(parseResticSnapshotID(raw))
	}
}
