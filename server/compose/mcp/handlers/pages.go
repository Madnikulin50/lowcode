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

func initPages(ctx context.Context, s *server.MCPServer) {
	s.AddTool(mcp.NewTool("read_page",
		mcp.WithDescription("Read a page by ID"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("pageID", mcp.Description("Page ID"), mcp.Required()),
	), handleReadPage)

	s.AddTool(mcp.NewTool("list_pages",
		mcp.WithDescription("List all pages in a namespace"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
	), handleListPages)

	s.AddTool(mcp.NewTool("search_pages",
		mcp.WithDescription("Search pages by title"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("query", mcp.Description("Search query"), mcp.Required()),
	), handleSearchPages)

	s.AddTool(mcp.NewTool("create_page",
		mcp.WithDescription("Create a new page"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("title", mcp.Description("Page title"), mcp.Required()),
		mcp.WithString("handle", mcp.Description("URL-safe handle")),
		mcp.WithString("description", mcp.Description("Page description")),
		mcp.WithString("moduleID", mcp.Description("Module ID that this page is for")),
		mcp.WithString("selfID", mcp.Description("Parent page ID (0 for root)")),
		mcp.WithString("blocks", mcp.Description(`JSON array of page blocks. Example: [{"kind":"RecordList","options":{"fields":["field1","field2"]}}]`)),
	), handleCreatePage)

	s.AddTool(mcp.NewTool("update_page",
		mcp.WithDescription("Update a page's title, handle, description, or blocks"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("pageID", mcp.Description("Page ID"), mcp.Required()),
		mcp.WithString("title", mcp.Description("Page title")),
		mcp.WithString("handle", mcp.Description("URL-safe handle")),
		mcp.WithString("description", mcp.Description("Page description")),
		mcp.WithString("blocks", mcp.Description("JSON array of page blocks")),
	), handleUpdatePage)

	s.AddTool(mcp.NewTool("delete_page",
		mcp.WithDescription("Delete a page by ID"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("pageID", mcp.Description("Page ID"), mcp.Required()),
	), handleDeletePage)

	s.AddResourceTemplate(mcp.NewResourceTemplate("mcp://page/{nsID}/{id}", "Page by ID",
		mcp.WithTemplateDescription("Read a page by namespace and page ID"),
	), handlePageResource)

	s.AddResourceTemplate(mcp.NewResourceTemplate("mcp://page/{nsID}/list", "Pages in namespace",
		mcp.WithTemplateDescription("List all pages in a namespace"),
	), handlePageListResource)
}

func handleReadPage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	pID := parseUint64(args, "pageID")
	p, err := service.DefaultPage.FindByID(ctx, nsID, pID)
	if err != nil {
		return errorResult(fmt.Errorf("page not found: %w", err)), nil
	}
	return jsonResult(p), nil
}

func handleListPages(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	set, _, err := service.DefaultPage.Find(ctx, types.PageFilter{NamespaceID: nsID})
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(set), nil
}

func handleSearchPages(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	query := getString(args, "query")
	set, _, err := service.DefaultPage.Find(ctx, types.PageFilter{
		NamespaceID: nsID,
		Query:       query,
	})
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(set), nil
}

func handleCreatePage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	namespaceID := parseUint64(args, "namespaceID")
	title := getString(args, "title")
	handle := getString(args, "handle")
	description := getString(args, "description")
	moduleID := parseUint64(args, "moduleID")
	selfID := parseUint64(args, "selfID")
	blocksJSON := getString(args, "blocks")

	var blocks types.PageBlocks
	if blocksJSON != "" {
		if err := json.Unmarshal([]byte(blocksJSON), &blocks); err != nil {
			return textResult(fmt.Sprintf("Invalid blocks JSON: %v", err)), nil
		}
	}

	p := &types.Page{
		NamespaceID: namespaceID,
		SelfID:      selfID,
		ModuleID:    moduleID,
		Title:       title,
		Handle:      handle,
		Description: description,
		Visible:     true,
		Weight:      0,
		Blocks:      blocks,
	}

	created, err := service.DefaultPage.Create(ctx, p)
	if err != nil {
		return errorResult(err), nil
	}

	return jsonResult(map[string]interface{}{
		"pageID": fmt.Sprintf("%d", created.ID),
		"title":  created.Title,
		"handle": created.Handle,
	}), nil
}

func handleUpdatePage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	pID := parseUint64(args, "pageID")

	p, err := service.DefaultPage.FindByID(ctx, nsID, pID)
	if err != nil {
		return errorResult(fmt.Errorf("page not found: %w", err)), nil
	}

	if v := getString(args, "title"); v != "" {
		p.Title = v
	}
	if v := getString(args, "handle"); v != "" {
		p.Handle = v
	}
	if v := getString(args, "description"); v != "" {
		p.Description = v
	}
	if v := getString(args, "blocks"); v != "" {
		var blocks types.PageBlocks
		if err := json.Unmarshal([]byte(v), &blocks); err != nil {
			return textResult(fmt.Sprintf("Invalid blocks JSON: %v", err)), nil
		}
		p.Blocks = blocks
	}

	updated, err := service.DefaultPage.Update(ctx, p)
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(updated), nil
}

func handleDeletePage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	pID := parseUint64(args, "pageID")
	if err := service.DefaultPage.DeleteByID(ctx, nsID, pID, types.PageChildrenOnDeleteForce); err != nil {
		return errorResult(err), nil
	}
	return textResult("Page deleted"), nil
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
