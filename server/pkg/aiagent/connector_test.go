package aiagent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

func TestReservedKitName(t *testing.T) {
	if !ReservedKitName("compose.records") || !ReservedKitName("Compose.Mail") {
		t.Fatal("compose.* must be reserved")
	}
	if ReservedKitName("cmdb") || ReservedKitName("compose") {
		t.Fatal("cmdb / compose without dot must not be reserved")
	}
}

func TestValidConnectorURL(t *testing.T) {
	if _, err := ValidConnectorURL(""); err == nil {
		t.Fatal("empty")
	}
	if _, err := ValidConnectorURL("ftp://x"); err == nil {
		t.Fatal("scheme")
	}
	got, err := ValidConnectorURL("http://localhost:8085/api/")
	if err != nil {
		t.Fatal(err)
	}
	if got != "http://localhost:8085/api" {
		t.Fatalf("got %s", got)
	}
	if _, err := ValidConnectorURL("http:///nohost"); err == nil {
		t.Fatal("host required")
	}
}

func TestParseToolKitEnv(t *testing.T) {
	got := ParseToolKitEnv(" cmdb = http://127.0.0.1:1/api ; backup=http://127.0.0.1:2, compose.records=http://evil ")
	if len(got) != 2 {
		t.Fatalf("got %#v", got)
	}
	if got[0].Handle != "cmdb" || got[0].URL != "http://127.0.0.1:1/api" || got[0].Source != "env" {
		t.Fatalf("cmdb=%#v", got[0])
	}
	if got[1].Handle != "backup" {
		t.Fatalf("backup=%#v", got[1])
	}
	if ParseToolKitEnv("") != nil {
		t.Fatal("empty env")
	}
}

func TestOverlayConnectors(t *testing.T) {
	en := false
	base := []Connector{
		{Handle: "cmdb", URL: "http://seed", Source: "env"},
		{Handle: "compose.records", URL: "http://evil", Source: "env"},
	}
	over := []Connector{
		{Handle: "cmdb", URL: "http://db", Source: "settings", Enabled: &en},
		{Handle: "custom", URL: "http://c:1"},
	}
	got := overlayConnectors(base, over)
	if len(got) != 2 {
		t.Fatalf("got %#v", got)
	}
	if got[0].Handle != "cmdb" || got[0].URL != "http://db" || got[0].IsEnabled() {
		t.Fatalf("overlay %#v", got[0])
	}
	if got[1].Handle != "custom" {
		t.Fatalf("custom %#v", got[1])
	}
}

func TestEffectiveConnectorsSettingsWin(t *testing.T) {
	t.Setenv("CMDB_AGENT_URL", "http://seed-cmdb/api")
	t.Setenv("BACKUP_AGENT_URL", "http://seed-backup/api")
	t.Setenv("INVEST_AGENT_URL", "http://seed-invest/api")
	t.Setenv("AI_TOOLKITS", "")
	SetConnectorsProvider(func() []Connector {
		return []Connector{{Handle: "cmdb", URL: "http://from-db/api", Source: "settings"}}
	})
	defer SetConnectorsProvider(nil)

	got := EffectiveConnectors()
	var cmdb Connector
	for _, c := range got {
		if c.Handle == "cmdb" {
			cmdb = c
		}
	}
	if cmdb.URL != "http://from-db/api" {
		t.Fatalf("cmdb=%#v connectors=%#v", cmdb, got)
	}
}

func TestRegisterRejectsRemoteCompose(t *testing.T) {
	c := NewCatalog()
	c.Register(ToolKit{Name: "compose.records", Tools: []chat.ToolDef{{Name: "ok"}}})
	c.Register(ToolKit{Name: "compose.records", Remote: true, Tools: []chat.ToolDef{{Name: "evil"}}})
	k, ok := c.Get("compose.records")
	if !ok || k.Remote || k.Tools[0].Name != "ok" {
		t.Fatalf("%#v", k)
	}
	c.Unregister("compose.records")
	if _, ok := c.Get("compose.records"); !ok {
		t.Fatal("in-process compose kit must stay")
	}
}

func TestRefreshRemotesFromConnector(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/meta" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"handle": "cmdb",
			"name":   "Live CMDB",
			"components": []map[string]any{
				{"type": "cmdb/scan", "operation": "scan", "label": "Scan"},
			},
		})
	}))
	defer srv.Close()

	c := NewCatalog()
	c.RegisterFallbacks()
	SetConnectorsProvider(func() []Connector {
		return []Connector{{Handle: "cmdb", URL: srv.URL, Source: "settings"}}
	})
	defer SetConnectorsProvider(nil)

	c.RefreshRemotesNow(context.Background())
	kit, ok := c.Get("cmdb")
	if !ok {
		t.Fatal("missing cmdb")
	}
	if kit.URL != srv.URL {
		t.Fatalf("url=%s", kit.URL)
	}
	if !kit.Remote {
		t.Fatal("expected remote kit")
	}
}

func TestRefreshRemotesUnregistersDisabledAndStale(t *testing.T) {
	c := NewCatalog()
	c.RegisterFallbacks()
	c.Register(ToolKit{Name: "gone", Remote: true, Tools: []chat.ToolDef{{Name: "g"}}})

	off := false
	SetConnectorsProvider(func() []Connector {
		return []Connector{{Handle: "cmdb", URL: "http://127.0.0.1:9/api", Enabled: &off, Source: "settings"}}
	})
	defer SetConnectorsProvider(nil)

	c.RefreshRemotesNow(context.Background())
	if _, ok := c.Get("cmdb"); ok {
		t.Fatal("disabled cmdb should be unregistered")
	}
	if _, ok := c.Get("gone"); ok {
		t.Fatal("stale remote should be unregistered")
	}
	if _, ok := c.Get("backup"); !ok {
		t.Fatal("seed backup fallback should remain")
	}
}
