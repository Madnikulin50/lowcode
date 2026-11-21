package handlers

import (
	"context"
	"fmt"

	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Module CRUD tools (read_module, list_modules, search_modules,
// create_module, update_module, delete_module) come from the shared
// registry in service.ModuleToolDefs — see tooldefs.go. Resource templates
// stay hand-written here since they are MCP-only.
func initModules(ctx context.Context, s *server.MCPServer) {
	registerToolDefs(s, service.ModuleToolDefs()...)

	s.AddResourceTemplate(mcp.NewResourceTemplate("mcp://module/{nsID}/{id}", "Module by ID",
		mcp.WithTemplateDescription("Read a module by namespace and module ID"),
	), handleModuleResource)

	s.AddResourceTemplate(mcp.NewResourceTemplate("mcp://module/{nsID}/list", "Modules in namespace",
		mcp.WithTemplateDescription("List all modules in a namespace"),
	), handleModuleListResource)
}

func handleModuleResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	ctx = withAuth(ctx)
	nsIDStr, _ := request.Params.Arguments["nsID"].(string)
	idStr, _ := request.Params.Arguments["id"].(string)
	var nsID, mID uint64
	fmt.Sscanf(nsIDStr, "%d", &nsID)
	fmt.Sscanf(idStr, "%d", &mID)
	mod, err := service.DefaultModule.FindByID(ctx, nsID, mID)
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{URI: "mcp://module/" + nsIDStr + "/" + idStr, Text: toJSON(mod)},
	}, nil
}

func handleModuleListResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	ctx = withAuth(ctx)
	nsIDStr, _ := request.Params.Arguments["nsID"].(string)
	var nsID uint64
	fmt.Sscanf(nsIDStr, "%d", &nsID)
	set, _, err := service.DefaultModule.Find(ctx, types.ModuleFilter{NamespaceID: nsID})
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{URI: "mcp://module/" + nsIDStr + "/list", Text: toJSON(set)},
	}, nil
}
