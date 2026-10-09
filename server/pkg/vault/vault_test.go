package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestParseRef(t *testing.T) {
	cases := []struct {
		raw     string
		want    Ref
		wantErr bool
	}{
		{raw: "compose/connectors/512#password", want: Ref{Path: "compose/connectors/512", Key: "password"}},
		{raw: "secret:compose/connectors/512#password", want: Ref{Mount: "secret", Path: "compose/connectors/512", Key: "password"}},
		{raw: "vault:secret:compose/connectors/512#password", want: Ref{Mount: "secret", Path: "compose/connectors/512", Key: "password"}},
		{raw: "", wantErr: true},
		{raw: "no-hash-here", wantErr: true},
		{raw: "path#", wantErr: true},
	}
	for _, c := range cases {
		got, err := ParseRef(c.raw)
		if c.wantErr {
			if err == nil {
				t.Errorf("ParseRef(%q): expected error, got %#v", c.raw, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseRef(%q): unexpected error: %v", c.raw, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseRef(%q) = %#v, want %#v", c.raw, got, c.want)
		}
	}
}

// newFakeVaultServer serves a single KV v2 secret at /v1/<mount>/data/<path>,
// counting how many times it was actually hit (to assert on caching).
func newFakeVaultServer(t *testing.T, mount, path string, data map[string]interface{}) (*httptest.Server, *int32) {
	t.Helper()
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantPath := "/v1/" + mount + "/data/" + path
		if r.URL.Path != wantPath {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Header.Get("X-Vault-Token") == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		atomic.AddInt32(&hits, 1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"data": data,
			},
		})
	}))
	t.Cleanup(srv.Close)
	return srv, &hits
}

func TestClient_Resolve_Success(t *testing.T) {
	srv, hits := newFakeVaultServer(t, "secret", "compose/connectors/512", map[string]interface{}{
		"password": "s3cr3t",
	})

	c := &Client{Addr: srv.URL, Token: "test-token", Mount: "secret"}
	got, err := c.Resolve(context.Background(), "compose/connectors/512#password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "s3cr3t" {
		t.Fatalf("got %q, want %q", got, "s3cr3t")
	}
	if atomic.LoadInt32(hits) != 1 {
		t.Fatalf("expected 1 request, got %d", *hits)
	}
}

func TestClient_Resolve_UsesCache(t *testing.T) {
	srv, hits := newFakeVaultServer(t, "secret", "compose/connectors/512", map[string]interface{}{
		"password": "s3cr3t",
	})

	c := &Client{Addr: srv.URL, Token: "test-token", Mount: "secret", cacheTTL: time.Minute}
	for i := 0; i < 5; i++ {
		if _, err := c.Resolve(context.Background(), "compose/connectors/512#password"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if got := atomic.LoadInt32(hits); got != 1 {
		t.Fatalf("expected exactly 1 request thanks to caching, got %d", got)
	}
}

func TestClient_Resolve_MissingKey(t *testing.T) {
	srv, _ := newFakeVaultServer(t, "secret", "compose/connectors/512", map[string]interface{}{
		"other": "value",
	})
	c := &Client{Addr: srv.URL, Token: "test-token", Mount: "secret"}
	if _, err := c.Resolve(context.Background(), "compose/connectors/512#password"); err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestClient_Resolve_NotFound(t *testing.T) {
	srv, _ := newFakeVaultServer(t, "secret", "compose/connectors/512", map[string]interface{}{"password": "x"})
	c := &Client{Addr: srv.URL, Token: "test-token", Mount: "secret"}
	if _, err := c.Resolve(context.Background(), "compose/connectors/999#password"); err == nil {
		t.Fatal("expected error for a path the fake server doesn't serve")
	}
}

func TestClient_Resolve_NotConfigured(t *testing.T) {
	c := &Client{} // no Addr
	if c.Configured() {
		t.Fatal("expected an empty-Addr client to be unconfigured")
	}
	if _, err := c.Resolve(context.Background(), "x#y"); err == nil {
		t.Fatal("expected error when Vault is not configured")
	}
}

func TestResolveField_EmptyRefReturnsPlaintextWithoutCallingVault(t *testing.T) {
	// no server at all - if this tried to call Vault it would fail/hang
	got, err := ResolveField(context.Background(), "", "plain-value")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "plain-value" {
		t.Fatalf("got %q, want %q", got, "plain-value")
	}
}

func TestFromEnv_MountDefaultsToSecret(t *testing.T) {
	t.Setenv("VAULT_ADDR", "")
	t.Setenv("VAULT_KV_MOUNT", "")
	c := FromEnv()
	if c.Mount != "secret" {
		t.Fatalf("expected default mount %q, got %q", "secret", c.Mount)
	}
	if c.Configured() {
		t.Fatal("expected client with empty VAULT_ADDR to be unconfigured")
	}
}
