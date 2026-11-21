package rulesgo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"context"

	"github.com/madnikulin50/lowcode/server/pkg/vault"
)

func TestBuildODataFilter(t *testing.T) {
	cases := []struct {
		name    string
		matchBy map[string]string
		want    string
	}{
		{name: "empty", matchBy: nil, want: ""},
		{name: "single", matchBy: map[string]string{"Код": "42"}, want: "Код eq '42'"},
		{
			name:    "multiple sorted deterministically",
			matchBy: map[string]string{"Артикул": "SKU-1", "Код": "42"},
			want:    "Артикул eq 'SKU-1' and Код eq '42'",
		},
		{
			name:    "escapes single quotes",
			matchBy: map[string]string{"Наименование": "O'Brien"},
			want:    "Наименование eq 'O''Brien'",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := buildODataFilter(c.matchBy); got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestExtractODataItems(t *testing.T) {
	t.Run("modern value shape", func(t *testing.T) {
		items, err := extractODataItems([]byte(`{"odata.metadata":"x","value":[{"Ref_Key":"abc"}]}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 1 || items[0]["Ref_Key"] != "abc" {
			t.Fatalf("unexpected items: %#v", items)
		}
	})
	t.Run("legacy d.results shape", func(t *testing.T) {
		items, err := extractODataItems([]byte(`{"d":{"results":[{"Ref_Key":"legacy"}]}}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 1 || items[0]["Ref_Key"] != "legacy" {
			t.Fatalf("unexpected items: %#v", items)
		}
	})
	t.Run("empty result set", func(t *testing.T) {
		items, err := extractODataItems([]byte(`{"value":[]}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 0 {
			t.Fatalf("expected no items, got %#v", items)
		}
	})
}

// fake1CServer simulates just enough of a 1C OData service for the
// find -> patch-or-create flow: a GET with $filter looks the record up in
// a fixed in-memory catalog, PATCH updates it, POST creates a new one.
type fake1CServer struct {
	mu       []map[string]interface{}
	requests []string // "METHOD path?query" for assertions
}

func newFake1CServer(t *testing.T, seed []map[string]interface{}) (*httptest.Server, *fake1CServer) {
	t.Helper()
	f := &fake1CServer{mu: seed}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, p, ok := r.BasicAuth(); !ok || u != "svc" || p != "s3cr3t" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		f.requests = append(f.requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)

		switch r.Method {
		case http.MethodGet:
			filter := r.URL.Query().Get("$filter")
			var matches []map[string]interface{}
			for _, item := range f.mu {
				if filter == "" || filter == "Код eq '42'" && item["Код"] == "42" {
					matches = append(matches, item)
				}
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"value": matches})

		case http.MethodPost:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			body["Ref_Key"] = "new-guid-123"
			f.mu = append(f.mu, body)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(body)

		case http.MethodPatch:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, f
}

func TestOneCSync_UpdatesExistingRecordByBusinessKey(t *testing.T) {
	srv, fake := newFake1CServer(t, []map[string]interface{}{
		{"Ref_Key": "existing-guid", "Код": "42", "Наименование": "Old Name"},
	})

	n := &oneCSyncExecutor{}
	node := ChainNode{Config: json.RawMessage(`{
		"url": "` + srv.URL + `",
		"entity": "Catalog_Номенклатура",
		"username": "svc",
		"password": "s3cr3t",
		"matchBy": {"Код": "{{code}}"},
		"fields": {"Наименование": "{{name}}"}
	}`)}
	ec := &ExecutionContext{Variables: map[string]interface{}{"code": "42", "name": "New Name"}}

	out, err := n.Execute(context.Background(), node, ec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["action"] != "updated" {
		t.Fatalf("expected action=updated, got %#v", out)
	}
	if out["ref"] != "existing-guid" {
		t.Fatalf("expected ref=existing-guid, got %#v", out["ref"])
	}

	found := false
	for _, r := range fake.requests {
		if r == "PATCH /Catalog_Номенклатура(guid'existing-guid')?$format=json" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a PATCH to the matched Ref_Key, got requests: %#v", fake.requests)
	}
}

func TestOneCSync_CreatesRecordWhenNotFound(t *testing.T) {
	srv, fake := newFake1CServer(t, nil) // empty catalog - nothing will match

	n := &oneCSyncExecutor{}
	node := ChainNode{Config: json.RawMessage(`{
		"url": "` + srv.URL + `",
		"entity": "Catalog_Номенклатура",
		"username": "svc",
		"password": "s3cr3t",
		"matchBy": {"Код": "{{code}}"},
		"fields": {"Наименование": "{{name}}"}
	}`)}
	ec := &ExecutionContext{Variables: map[string]interface{}{"code": "42", "name": "Brand New"}}

	out, err := n.Execute(context.Background(), node, ec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["action"] != "created" {
		t.Fatalf("expected action=created, got %#v", out)
	}
	if out["ref"] != "new-guid-123" {
		t.Fatalf("expected the fake server's generated ref, got %#v", out["ref"])
	}

	var posted map[string]interface{}
	for _, item := range fake.mu {
		if item["Ref_Key"] == "new-guid-123" {
			posted = item
		}
	}
	if posted == nil || posted["Код"] != "42" || posted["Наименование"] != "Brand New" {
		t.Fatalf("expected created record to carry both matchBy and fields, got %#v", posted)
	}
}

func TestOneCSync_RequiresURLEntityAndMatchBy(t *testing.T) {
	n := &oneCSyncExecutor{}

	if _, err := n.Execute(context.Background(), ChainNode{Config: json.RawMessage(`{"entity":"x","matchBy":{"a":"b"}}`)}, &ExecutionContext{}); err == nil {
		t.Fatal("expected error for missing url")
	}
	if _, err := n.Execute(context.Background(), ChainNode{Config: json.RawMessage(`{"url":"http://x","matchBy":{"a":"b"}}`)}, &ExecutionContext{}); err == nil {
		t.Fatal("expected error for missing entity")
	}
	if _, err := n.Execute(context.Background(), ChainNode{Config: json.RawMessage(`{"url":"http://x","entity":"y"}`)}, &ExecutionContext{}); err == nil {
		t.Fatal("expected error for missing matchBy")
	}
}

func TestOneCSync_WrongCredentialsFail(t *testing.T) {
	srv, _ := newFake1CServer(t, nil)

	n := &oneCSyncExecutor{}
	node := ChainNode{Config: json.RawMessage(`{
		"url": "` + srv.URL + `",
		"entity": "Catalog_Номенклатура",
		"username": "svc",
		"password": "wrong-password",
		"matchBy": {"Код": "42"}
	}`)}

	if _, err := n.Execute(context.Background(), node, &ExecutionContext{}); err == nil {
		t.Fatal("expected error for wrong credentials (401)")
	}
}

func TestOneCSync_PasswordSecretRefOverridesPlaintext(t *testing.T) {
	vaultSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": map[string]interface{}{"password": "s3cr3t"}},
		})
	}))
	defer vaultSrv.Close()
	restore := vault.SetDefault(&vault.Client{Addr: vaultSrv.URL, Token: "t", Mount: "secret"})
	defer restore()

	srv, _ := newFake1CServer(t, nil)

	n := &oneCSyncExecutor{}
	node := ChainNode{Config: json.RawMessage(`{
		"url": "` + srv.URL + `",
		"entity": "Catalog_Номенклатура",
		"username": "svc",
		"password": "placeholder-should-be-overridden",
		"passwordSecretRef": "compose/connectors/1#password",
		"matchBy": {"Код": "42"},
		"fields": {}
	}`)}

	out, err := n.Execute(context.Background(), node, &ExecutionContext{})
	if err != nil {
		t.Fatalf("unexpected error (vault-resolved password should have authenticated): %v", err)
	}
	if out["action"] != "created" {
		t.Fatalf("unexpected result: %#v", out)
	}
}
