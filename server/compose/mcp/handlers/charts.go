package handlers

import (
	"context"
	"fmt"

	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Chart CRUD tools (read_chart, list_charts, search_charts, create_chart,
// update_chart, delete_chart) come from the shared registry in
// service.ChartToolDefs — see tooldefs.go. Resource templates stay
// hand-written here since they are MCP-only.
func initCharts(ctx context.Context, s *server.MCPServer) {
	registerToolDefs(s, service.ChartToolDefs()...)

	s.AddResourceTemplate(mcp.NewResourceTemplate("mcp://chart/{nsID}/{id}", "Chart by ID",
		mcp.WithTemplateDescription("Read a chart by namespace and chart ID"),
	), handleChartResource)

	s.AddResourceTemplate(mcp.NewResourceTemplate("mcp://chart/{nsID}/list", "Charts in namespace",
		mcp.WithTemplateDescription("List all charts in a namespace"),
	), handleChartListResource)
}

func handleChartResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	ctx = withAuth(ctx)
	nsIDStr, _ := request.Params.Arguments["nsID"].(string)
	idStr, _ := request.Params.Arguments["id"].(string)
	var nsID, cID uint64
	fmt.Sscanf(nsIDStr, "%d", &nsID)
	fmt.Sscanf(idStr, "%d", &cID)
	ch, err := service.DefaultChart.FindByID(ctx, nsID, cID)
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{URI: "mcp://chart/" + nsIDStr + "/" + idStr, Text: toJSON(ch)},
	}, nil
}

func handleChartListResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	ctx = withAuth(ctx)
	nsIDStr, _ := request.Params.Arguments["nsID"].(string)
	var nsID uint64
	fmt.Sscanf(nsIDStr, "%d", &nsID)
	set, _, err := service.DefaultChart.Find(ctx, types.ChartFilter{NamespaceID: nsID})
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{URI: "mcp://chart/" + nsIDStr + "/list", Text: toJSON(set)},
	}, nil
}
