package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func initCharts(ctx context.Context, s *server.MCPServer) {
	s.AddTool(mcp.NewTool("read_chart",
		mcp.WithDescription("Read a chart by ID"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("chartID", mcp.Description("Chart ID"), mcp.Required()),
	), handleReadChart)

	s.AddTool(mcp.NewTool("list_charts",
		mcp.WithDescription("List all charts in a namespace"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
	), handleListCharts)

	s.AddTool(mcp.NewTool("search_charts",
		mcp.WithDescription("Search charts by name"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("query", mcp.Description("Search query"), mcp.Required()),
	), handleSearchCharts)

	s.AddTool(mcp.NewTool("create_chart",
		mcp.WithDescription("Create a new chart"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("name", mcp.Description("Chart name"), mcp.Required()),
		mcp.WithString("handle", mcp.Description("URL-safe handle")),
		mcp.WithString("config",
			mcp.Description(`JSON chart config. Example: {"reports":[{"moduleID":"...","dimensions":[{"field":"createdAt","modifier":"MONTH"}],"metrics":[{"field":"count","type":"bar","label":"Count"}]}]}`),
			mcp.Required(),
		),
	), handleCreateChart)

	s.AddTool(mcp.NewTool("update_chart",
		mcp.WithDescription("Update a chart's name, handle, or config"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("chartID", mcp.Description("Chart ID"), mcp.Required()),
		mcp.WithString("name", mcp.Description("Chart name")),
		mcp.WithString("handle", mcp.Description("URL-safe handle")),
		mcp.WithString("config", mcp.Description("JSON chart config")),
	), handleUpdateChart)

	s.AddTool(mcp.NewTool("delete_chart",
		mcp.WithDescription("Delete a chart by ID"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("chartID", mcp.Description("Chart ID"), mcp.Required()),
	), handleDeleteChart)

	s.AddResourceTemplate(mcp.NewResourceTemplate("mcp://chart/{nsID}/{id}", "Chart by ID",
		mcp.WithTemplateDescription("Read a chart by namespace and chart ID"),
	), handleChartResource)

	s.AddResourceTemplate(mcp.NewResourceTemplate("mcp://chart/{nsID}/list", "Charts in namespace",
		mcp.WithTemplateDescription("List all charts in a namespace"),
	), handleChartListResource)
}

func handleReadChart(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	cID := parseUint64(args, "chartID")
	ch, err := service.DefaultChart.FindByID(ctx, nsID, cID)
	if err != nil {
		return errorResult(fmt.Errorf("chart not found: %w", err)), nil
	}
	return jsonResult(ch), nil
}

func handleListCharts(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	set, _, err := service.DefaultChart.Find(ctx, types.ChartFilter{NamespaceID: nsID})
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(set), nil
}

func handleSearchCharts(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	query := getString(args, "query")
	set, _, err := service.DefaultChart.Find(ctx, types.ChartFilter{
		NamespaceID: nsID,
		Query:       query,
	})
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(set), nil
}

func handleCreateChart(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)

	namespaceID := parseUint64(args, "namespaceID")
	name := getString(args, "name")
	handle := getString(args, "handle")
	configJSON := getString(args, "config")

	var config types.ChartConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return textResult(fmt.Sprintf("Invalid config JSON: %v", err)), nil
	}

	ch := &types.Chart{
		NamespaceID: namespaceID,
		Name:        name,
		Handle:      handle,
		Config:      config,
	}

	created, err := service.DefaultChart.Create(ctx, ch)
	if err != nil {
		return errorResult(err), nil
	}

	return jsonResult(map[string]interface{}{
		"chartID": fmt.Sprintf("%d", created.ID),
		"name":    created.Name,
		"handle":  created.Handle,
	}), nil
}

func handleUpdateChart(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	cID := parseUint64(args, "chartID")

	ch, err := service.DefaultChart.FindByID(ctx, nsID, cID)
	if err != nil {
		return errorResult(fmt.Errorf("chart not found: %w", err)), nil
	}

	if v := getString(args, "name"); v != "" {
		ch.Name = v
	}
	if v := getString(args, "handle"); v != "" {
		ch.Handle = v
	}
	if v := getString(args, "config"); v != "" {
		var cfg types.ChartConfig
		if err := json.Unmarshal([]byte(v), &cfg); err != nil {
			return textResult(fmt.Sprintf("Invalid config JSON: %v", err)), nil
		}
		ch.Config = cfg
	}

	updated, err := service.DefaultChart.Update(ctx, ch)
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(updated), nil
}

func handleDeleteChart(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	cID := parseUint64(args, "chartID")
	if err := service.DefaultChart.DeleteByID(ctx, nsID, cID); err != nil {
		return errorResult(err), nil
	}
	return textResult("Chart deleted"), nil
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
