package handlers

import (
	"context"
	"fmt"

	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func initNamespace(ctx context.Context, s *server.MCPServer) {
	s.AddTool(mcp.NewTool("read_namespace",
		mcp.WithDescription("Read a namespace by ID"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
	), handleReadNamespace)

	s.AddTool(mcp.NewTool("list_namespaces",
		mcp.WithDescription("List all namespaces"),
	), handleListNamespaces)

	s.AddTool(mcp.NewTool("search_namespaces",
		mcp.WithDescription("Search namespaces by name or slug"),
		mcp.WithString("query", mcp.Description("Search query"), mcp.Required()),
	), handleSearchNamespaces)

	s.AddTool(mcp.NewTool("create_namespace",
		mcp.WithDescription("Create a new namespace"),
		mcp.WithString("name", mcp.Description("Namespace name"), mcp.Required()),
		mcp.WithString("slug", mcp.Description("URL-safe slug")),
	), handleCreateNamespace)

	s.AddTool(mcp.NewTool("update_namespace",
		mcp.WithDescription("Update a namespace"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("name", mcp.Description("Namespace name")),
		mcp.WithString("slug", mcp.Description("URL-safe slug")),
	), handleUpdateNamespace)

	s.AddTool(mcp.NewTool("delete_namespace",
		mcp.WithDescription("Delete a namespace by ID"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
	), handleDeleteNamespace)

	s.AddResourceTemplate(mcp.NewResourceTemplate("mcp://namespace/{id}", "Namespace by ID",
		mcp.WithTemplateDescription("Read a namespace by ID"),
	), handleNamespaceResource)

	s.AddResourceTemplate(mcp.NewResourceTemplate("mcp://namespace/list", "All namespaces",
		mcp.WithTemplateDescription("List all namespaces"),
	), handleNamespaceListResource)
}

func argsMap(request mcp.CallToolRequest) map[string]interface{} {
	if m, ok := request.Params.Arguments.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}

func handleReadNamespace(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	ns, err := service.DefaultNamespace.FindByID(ctx, nsID)
	if err != nil {
		return errorResult(fmt.Errorf("namespace not found: %w", err)), nil
	}
	return jsonResult(ns), nil
}

func handleListNamespaces(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	set, _, err := service.DefaultNamespace.Find(ctx, types.NamespaceFilter{})
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(set), nil
}

func handleSearchNamespaces(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	query := getString(args, "query")
	set, _, err := service.DefaultNamespace.Find(ctx, types.NamespaceFilter{
		Query: query,
	})
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(set), nil
}

func handleCreateNamespace(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	ns := &types.Namespace{
		Name: getString(args, "name"),
		Slug: getString(args, "slug"),
	}
	created, err := service.DefaultNamespace.Create(ctx, ns)
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(map[string]interface{}{
		"namespaceID": fmt.Sprintf("%d", created.ID),
		"name":        created.Name,
		"slug":        created.Slug,
	}), nil
}

func handleUpdateNamespace(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	ns, err := service.DefaultNamespace.FindByID(ctx, nsID)
	if err != nil {
		return errorResult(fmt.Errorf("namespace not found: %w", err)), nil
	}
	if v := getString(args, "name"); v != "" {
		ns.Name = v
	}
	if v := getString(args, "slug"); v != "" {
		ns.Slug = v
	}
	updated, err := service.DefaultNamespace.Update(ctx, ns)
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(updated), nil
}

func handleDeleteNamespace(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	if err := service.DefaultNamespace.DeleteByID(ctx, nsID); err != nil {
		return errorResult(err), nil
	}
	return textResult("Namespace deleted"), nil
}

func handleNamespaceResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	ctx = withAuth(ctx)
	id, _ := request.Params.Arguments["id"].(string)
	var nsID uint64
	fmt.Sscanf(id, "%d", &nsID)
	ns, err := service.DefaultNamespace.FindByID(ctx, nsID)
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{URI: "mcp://namespace/" + id, Text: toJSON(ns)},
	}, nil
}

func handleNamespaceListResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	ctx = withAuth(ctx)
	set, _, err := service.DefaultNamespace.Find(ctx, types.NamespaceFilter{})
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{URI: "mcp://namespace/list", Text: toJSON(set)},
	}, nil
}
