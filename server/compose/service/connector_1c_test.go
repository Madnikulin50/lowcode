package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/madnikulin50/lowcode/server/pkg/vault"
)

func testModule1C() *types.Module {
	return &types.Module{
		ID: 1,
		Fields: types.ModuleFieldSet{
			{Name: "Код"},
			{Name: "Наименование"},
		},
	}
}

func newFake1CODataServer(t *testing.T, wantUser, wantPass string, items []map[string]interface{}) (*httptest.Server, *http.Request) {
	t.Helper()
	var lastReq *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastReq = r
		if wantUser != "" {
			if u, p, ok := r.BasicAuth(); !ok || u != wantUser || p != wantPass {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"value": items})
	}))
	t.Cleanup(srv.Close)
	return srv, lastReq
}

func TestFetch1C_ParsesValueEnvelope(t *testing.T) {
	srv, _ := newFake1CODataServer(t, "svc", "s3cr3t", []map[string]interface{}{
		{"Код": "1", "Наименование": "Товар А"},
		{"Код": "2", "Наименование": "Товар Б"},
	})

	mod := testModule1C()
	mod.Config.Connector = types.ModuleConfigConnector{
		Type:         "1c-odata",
		RestURL:      srv.URL,
		OneCEntity:   "Catalog_Номенклатура",
		OneCUsername: "svc",
		OneCPassword: "s3cr3t",
	}

	svc := &connectorSvc{}
	set, outFilter, err := svc.Fetch(context.Background(), mod, types.RecordFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(set) != 2 {
		t.Fatalf("expected 2 records, got %d", len(set))
	}
	if outFilter.Total != 2 {
		t.Fatalf("expected total=2, got %d", outFilter.Total)
	}
	if v := set[0].Values.FilterByName("Наименование")[0].Value; v != "Товар А" {
		t.Fatalf("unexpected value: %q", v)
	}
}

func TestFetch1C_WrongCredentialsError(t *testing.T) {
	srv, _ := newFake1CODataServer(t, "svc", "s3cr3t", nil)

	mod := testModule1C()
	mod.Config.Connector = types.ModuleConfigConnector{
		Type:         "1c-odata",
		RestURL:      srv.URL,
		OneCEntity:   "Catalog_Номенклатура",
		OneCUsername: "svc",
		OneCPassword: "wrong",
	}

	svc := &connectorSvc{}
	if _, _, err := svc.Fetch(context.Background(), mod, types.RecordFilter{}); err == nil {
		t.Fatal("expected error for wrong credentials")
	}
}

func TestFetch1C_RequiresURLAndEntity(t *testing.T) {
	svc := &connectorSvc{}

	mod := testModule1C()
	mod.Config.Connector = types.ModuleConfigConnector{Type: "1c-odata", OneCEntity: "x"}
	if _, _, err := svc.Fetch(context.Background(), mod, types.RecordFilter{}); err == nil {
		t.Fatal("expected error for missing RestURL")
	}

	mod.Config.Connector = types.ModuleConfigConnector{Type: "1c-odata", RestURL: "http://x"}
	if _, _, err := svc.Fetch(context.Background(), mod, types.RecordFilter{}); err == nil {
		t.Fatal("expected error for missing entity")
	}
}

func TestFetch1C_PasswordFromVaultOverridesPlaintext(t *testing.T) {
	srv, _ := newFake1CODataServer(t, "svc", "vault-password", []map[string]interface{}{
		{"Код": "1", "Наименование": "X"},
	})

	vaultSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": map[string]interface{}{"password": "vault-password"}},
		})
	}))
	defer vaultSrv.Close()
	restore := vault.SetDefault(&vault.Client{Addr: vaultSrv.URL, Token: "t", Mount: "secret"})
	defer restore()

	mod := testModule1C()
	mod.Config.Connector = types.ModuleConfigConnector{
		Type:         "1c-odata",
		RestURL:      srv.URL,
		OneCEntity:   "Catalog_Номенклатура",
		OneCUsername: "svc",
		OneCPassword: "placeholder-should-be-overridden",
		SecretRefs:   map[string]string{"oneCPassword": "compose/connectors/1#password"},
	}

	svc := &connectorSvc{}
	set, _, err := svc.Fetch(context.Background(), mod, types.RecordFilter{})
	if err != nil {
		t.Fatalf("unexpected error (vault password should have authenticated): %v", err)
	}
	if len(set) != 1 {
		t.Fatalf("expected 1 record, got %d", len(set))
	}
	// original module config must be untouched (plaintext password preserved)
	if mod.Config.Connector.OneCPassword != "placeholder-should-be-overridden" {
		t.Fatalf("original module must not be mutated, got %q", mod.Config.Connector.OneCPassword)
	}
}

func TestTest1C_SuccessAndFailure(t *testing.T) {
	srv, _ := newFake1CODataServer(t, "svc", "s3cr3t", nil)
	svc := &connectorSvc{}

	ok := types.ModuleConfigConnector{Type: "1c-odata", RestURL: srv.URL, OneCEntity: "Catalog_Номенклатура", OneCUsername: "svc", OneCPassword: "s3cr3t"}
	if err := svc.Test(context.Background(), ok); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bad := ok
	bad.OneCPassword = "wrong"
	if err := svc.Test(context.Background(), bad); err == nil {
		t.Fatal("expected error for wrong credentials")
	}

	empty := types.ModuleConfigConnector{Type: "1c-odata"}
	if err := svc.Test(context.Background(), empty); err == nil {
		t.Fatal("expected error for missing RestURL")
	}
}

func TestOneCQuery_CapsTopByFilterLimit(t *testing.T) {
	cfg := types.ModuleConfigConnector{OneCTop: 100}
	q := oneCQuery(cfg, types.RecordFilter{})
	if q.Get("$top") != "100" {
		t.Fatalf("expected default $top=100, got %s", q.Get("$top"))
	}

	q = oneCQuery(cfg, types.RecordFilter{Paging: filter.Paging{Limit: 10}})
	if q.Get("$top") != "10" {
		t.Fatalf("expected filter.Limit to cap $top to 10, got %s", q.Get("$top"))
	}

	cfg.OneCSelect = "Код,Наименование"
	cfg.OneCFilter = "Код eq '1'"
	q = oneCQuery(cfg, types.RecordFilter{})
	if q.Get("$select") != "Код,Наименование" || q.Get("$filter") != "Код eq '1'" {
		t.Fatalf("unexpected query: %#v", q)
	}
}
