package handlers

import (
	"context"
	"testing"

	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/mark3labs/mcp-go/server"
)

// TestRegisterToolDefsBuildsSchema exercises the registry -> MCP adapter
// (registerToolDefs, mcpTool) for every Def in the shared module/chart/page
// registry (server/compose/service/tool_{modules,charts,pages}.go), without
// a live DB — registration only builds tool schemas, it never runs a
// Handler. This is the regression test for the tool-def unification: it
// would fail if a Def's Params ever produced a schema mcp-go rejects.
func TestRegisterToolDefsBuildsSchema(t *testing.T) {
	s := server.NewMCPServer("test", "0")

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("registerToolDefs panicked: %v", r)
		}
	}()

	registerToolDefs(s, service.ModuleToolDefs()...)
	registerToolDefs(s, service.ChartToolDefs()...)
	registerToolDefs(s, service.PageToolDefs()...)

	wantAtLeast := []string{
		"read_module", "list_modules", "search_modules", "create_module", "update_module", "delete_module",
		"read_chart", "list_charts", "search_charts", "create_chart", "update_chart", "delete_chart",
		"read_page", "list_pages", "search_pages", "create_page", "update_page", "delete_page",
	}
	tools := s.ListTools()
	for _, name := range wantAtLeast {
		if _, ok := tools[name]; !ok {
			t.Errorf("tool %q was not registered", name)
		}
	}
}

// TestInitDoesNotPanic exercises the full MCP handler registration
// (mcp/handlers/index.go), including the shared tool registry alongside
// every other handler group — schema registration must not touch the DB.
func TestInitDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Init panicked: %v", r)
		}
	}()
	Init(context.Background(), server.NewMCPServer("test", "0"))
}
