package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/logger"
	"github.com/madnikulin50/lowcode/server/system/types"
)

// newTestApp builds a minimal CortezaApp for handler tests, bypassing
// New()'s InitCLI() — that registers flags on the global pflag
// CommandLine, which panics ("flag redefined") if called more than once
// per process, as multiple tests in this package would do.
func newTestApp() *CortezaApp {
	return &CortezaApp{Log: logger.Default()}
}

func setupServiceIdentity(t *testing.T) {
	t.Helper()

	role := &types.Role{ID: 1, Handle: auth.BypassRoleHandle}
	user := &types.User{ID: 42, Handle: auth.ServiceUserHandle}
	auth.SetSystemUsers(types.UserSet{user}, types.RoleSet{role})

	issuer, err := auth.NewTokenIssuer(
		auth.WithSecretSigner("test-signing-secret"),
		auth.WithStore(func(context.Context, auth.TokenRequest) error { return nil }),
	)
	if err != nil {
		t.Fatalf("NewTokenIssuer: %v", err)
	}
	auth.TokenIssuer = issuer
}

func TestMountAgentEnrollNotMountedWithoutSecret(t *testing.T) {
	t.Setenv("AGENT_SHARED_SECRET", "")

	a := newTestApp()
	r := chi.NewRouter()
	a.mountAgentEnroll(r, "/")

	req := httptest.NewRequest(http.MethodPost, "/agents/enroll", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (route should not exist)", rec.Code)
	}
}

func TestMountAgentEnrollRejectsWrongSecret(t *testing.T) {
	t.Setenv("AGENT_SHARED_SECRET", "correct-horse")
	setupServiceIdentity(t)

	a := newTestApp()
	r := chi.NewRouter()
	a.mountAgentEnroll(r, "/")

	req := httptest.NewRequest(http.MethodPost, "/agents/enroll", nil)
	req.Header.Set("X-Agent-Secret", "wrong")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestMountAgentEnrollIssuesToken(t *testing.T) {
	t.Setenv("AGENT_SHARED_SECRET", "correct-horse")
	setupServiceIdentity(t)

	a := newTestApp()
	r := chi.NewRouter()
	a.mountAgentEnroll(r, "/")

	req := httptest.NewRequest(http.MethodPost, "/agents/enroll", nil)
	req.Header.Set("X-Agent-Secret", "correct-horse")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rec.Code, rec.Body.String())
	}
	if rec.Body.Len() == 0 {
		t.Fatal("empty response body")
	}
}
