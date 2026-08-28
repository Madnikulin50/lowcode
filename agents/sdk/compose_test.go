package sdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDiscoverPicksComposeOrigin(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/compose/namespace/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":{"set":[]}}`))
	})
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(404)
		_, _ = w.Write([]byte("<!doctype html><html>nope</html>"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewClient(srv.URL+"/api", "", 0)
	if err := c.Discover(context.Background()); err != nil {
		t.Fatal(err)
	}
	if c.BaseURL() != srv.URL {
		t.Fatalf("origin %s want %s", c.BaseURL(), srv.URL)
	}
}
