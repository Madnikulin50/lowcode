package handlers

import (
	"context"
	"fmt"

	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Page CRUD tools (read_page, list_pages, search_pages, create_page,
// update_page, delete_page) come from the shared registry in
// service.PageToolDefs — see tooldefs.go. Resource templates stay
// hand-written here since they are MCP-only.
func initPages(ctx context.Context, s *server.MCPServer) {
	registerToolDefs(s, service.PageToolDefs()...)

	s.AddResourceTemplate(mcp.NewResourceTemplate("mcp://page/{nsID}/{id}", "Page by ID",
		mcp.WithTemplateDescription("Read a page by namespace and page ID"),
	), handlePageResource)

	s.AddResourceTemplate(mcp.NewResourceTemplate("mcp://page/{nsID}/list", "Pages in namespace",
		mcp.WithTemplateDescription("List all pages in a namespace"),
	), handlePageListResource)
}

func handlePageResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	ctx = withAuth(ctx)
	nsIDStr, _ := request.Params.Arguments["nsID"].(string)
	idStr, _ := request.Params.Arguments["id"].(string)
	var nsID, pID uint64
	fmt.Sscanf(nsIDStr, "%d", &nsID)
	fmt.Sscanf(idStr, "%d", &pID)
	p, err := service.DefaultPage.FindByID(ctx, nsID, pID)
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{URI: "mcp://page/" + nsIDStr + "/" + idStr, Text: toJSON(p)},
	}, nil
}

func handlePageListResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	ctx = withAuth(ctx)
	nsIDStr, _ := request.Params.Arguments["nsID"].(string)
	var nsID uint64
	fmt.Sscanf(nsIDStr, "%d", &nsID)
	set, _, err := service.DefaultPage.Find(ctx, types.PageFilter{NamespaceID: nsID})
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{URI: "mcp://page/" + nsIDStr + "/list", Text: toJSON(set)},
	}, nil
}
