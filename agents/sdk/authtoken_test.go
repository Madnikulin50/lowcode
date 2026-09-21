package sdk

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSelfTokenPrefersExplicitTOKEN(t *testing.T) {
	t.Setenv("TOKEN", "explicit-token")
	t.Setenv("AGENT_SHARED_SECRET", "should-not-be-used")

	if got := SelfToken("http://example.invalid"); got != "explicit-token" {
		t.Fatalf("got %q, want explicit-token", got)
	}
}

func TestSelfTokenEmptyWithoutAnyAuth(t *testing.T) {
	t.Setenv("TOKEN", "")
	t.Setenv("AGENT_SHARED_SECRET", "")

	if got := SelfToken("http://example.invalid"); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func TestSelfTokenEnrollExchange(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/agents/enroll" {
			t.Errorf("path %q", r.URL.Path)
		}
		if r.Header.Get("X-Agent-Secret") != "shhh" {
			t.Errorf("secret %q", r.Header.Get("X-Agent-Secret"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"minted-token"}`))
	}))
	defer srv.Close()

	t.Setenv("TOKEN", "")
	t.Setenv("AGENT_SHARED_SECRET", "shhh")

	if got := SelfToken(srv.URL); got != "minted-token" {
		t.Fatalf("got %q, want minted-token", got)
	}
}

func TestSelfTokenEnrollRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	t.Setenv("TOKEN", "")
	t.Setenv("AGENT_SHARED_SECRET", "wrong")

	if got := SelfToken(srv.URL); got != "" {
		t.Fatalf("got %q, want empty on rejected secret", got)
	}
}
