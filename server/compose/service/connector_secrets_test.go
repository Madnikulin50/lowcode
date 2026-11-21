package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/vault"
)

func newFakeVaultForSecrets(t *testing.T, secrets map[string]string) *vault.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// expects /v1/secret/data/compose/connectors/1
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"data": secrets,
			},
		})
	}))
	t.Cleanup(srv.Close)
	return &vault.Client{Addr: srv.URL, Token: "t", Mount: "secret"}
}

func TestResolveConnectorSecrets_NoRefsIsNoOp(t *testing.T) {
	cfg := types.ModuleConfigConnector{Type: "db", DBConnectionString: "postgres://plain"}
	out, err := resolveConnectorSecrets(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.DBConnectionString != "postgres://plain" {
		t.Fatalf("expected plaintext untouched, got %q", out.DBConnectionString)
	}
}

func TestResolveConnectorSecrets_DBConnectionStringFromVault(t *testing.T) {
	restore := vault.SetDefault(newFakeVaultForSecrets(t, map[string]string{
		"connectionString": "postgres://vault-resolved",
	}))
	defer restore()

	cfg := types.ModuleConfigConnector{
		Type:               "db",
		DBConnectionString: "postgres://plain-should-be-overridden",
		SecretRefs:         map[string]string{"dbConnectionString": "compose/connectors/1#connectionString"},
	}
	out, err := resolveConnectorSecrets(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.DBConnectionString != "postgres://vault-resolved" {
		t.Fatalf("expected vault value to win, got %q", out.DBConnectionString)
	}
}

func TestResolveConnectorSecrets_RestHeaderFromVault_DoesNotMutateCallerMap(t *testing.T) {
	restore := vault.SetDefault(newFakeVaultForSecrets(t, map[string]string{
		"token": "Bearer abc123",
	}))
	defer restore()

	originalHeaders := map[string]string{"Accept": "application/json"}
	cfg := types.ModuleConfigConnector{
		Type:        "rest",
		RestHeaders: originalHeaders,
		SecretRefs:  map[string]string{"restHeader:Authorization": "compose/connectors/1#token"},
	}
	out, err := resolveConnectorSecrets(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.RestHeaders["Authorization"] != "Bearer abc123" {
		t.Fatalf("expected resolved Authorization header, got %#v", out.RestHeaders)
	}
	if out.RestHeaders["Accept"] != "application/json" {
		t.Fatalf("expected existing headers preserved, got %#v", out.RestHeaders)
	}
	if _, ok := originalHeaders["Authorization"]; ok {
		t.Fatalf("caller's original RestHeaders map must not be mutated: %#v", originalHeaders)
	}
}

func TestResolveConnectorSecrets_UnknownFieldErrors(t *testing.T) {
	restore := vault.SetDefault(newFakeVaultForSecrets(t, map[string]string{"x": "y"}))
	defer restore()

	cfg := types.ModuleConfigConnector{
		SecretRefs: map[string]string{"somethingWeird": "compose/connectors/1#x"},
	}
	if _, err := resolveConnectorSecrets(context.Background(), cfg); err == nil {
		t.Fatal("expected error for unknown SecretRefs field")
	}
}

func TestResolveConnectorSecrets_VaultUnconfiguredErrors(t *testing.T) {
	restore := vault.SetDefault(&vault.Client{}) // no Addr
	defer restore()

	cfg := types.ModuleConfigConnector{
		SecretRefs: map[string]string{"dbConnectionString": "compose/connectors/1#connectionString"},
	}
	if _, err := resolveConnectorSecrets(context.Background(), cfg); err == nil {
		t.Fatal("expected error when Vault is not configured")
	}
}

func TestWithResolvedConnectorSecrets_NoRefsReturnsSamePointer(t *testing.T) {
	mod := &types.Module{ID: 1}
	out, err := withResolvedConnectorSecrets(context.Background(), mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != mod {
		t.Fatal("expected the same *Module pointer when there are no SecretRefs (no-op fast path)")
	}
}

func TestWithResolvedConnectorSecrets_DoesNotMutateOriginalModule(t *testing.T) {
	restore := vault.SetDefault(newFakeVaultForSecrets(t, map[string]string{
		"connectionString": "postgres://vault-resolved",
	}))
	defer restore()

	mod := &types.Module{ID: 1}
	mod.Config.Connector = types.ModuleConfigConnector{
		Type:               "db",
		DBConnectionString: "postgres://plain",
		SecretRefs:         map[string]string{"dbConnectionString": "compose/connectors/1#connectionString"},
	}

	out, err := withResolvedConnectorSecrets(context.Background(), mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == mod {
		t.Fatal("expected a distinct *Module copy when secrets are resolved")
	}
	if out.Config.Connector.DBConnectionString != "postgres://vault-resolved" {
		t.Fatalf("resolved copy not updated: %q", out.Config.Connector.DBConnectionString)
	}
	if mod.Config.Connector.DBConnectionString != "postgres://plain" {
		t.Fatalf("original module must be untouched, got %q", mod.Config.Connector.DBConnectionString)
	}
}
