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

func initPageBlocks(ctx context.Context, s *server.MCPServer) {
	s.AddTool(mcp.NewTool("read_pageblock",
		mcp.WithDescription("Read a specific page block from a page"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("pageID", mcp.Description("Page ID"), mcp.Required()),
		mcp.WithString("blockID", mcp.Description("Block index (0-based) or block title"), mcp.Required()),
	), handleReadPageBlock)

	s.AddTool(mcp.NewTool("list_pageblocks",
		mcp.WithDescription("List all blocks on a page"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("pageID", mcp.Description("Page ID"), mcp.Required()),
	), handleListPageBlocks)

	s.AddTool(mcp.NewTool("create_pageblock",
		mcp.WithDescription("Add a new block to a page"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("pageID", mcp.Description("Page ID"), mcp.Required()),
		mcp.WithString("kind", mcp.Description("Block kind: RecordList, Record, Chart, Metric, Content, Navigation, SocialFeed, File, IFrame, Geometry, Comment, Automation, Tabs, Report, Progress"), mcp.Required()),
		mcp.WithString("title", mcp.Description("Block title")),
		mcp.WithString("description", mcp.Description("Block description")),
		mcp.WithString("options", mcp.Description("JSON object with block-specific options. Examples:\n- RecordList: {\"moduleID\":\"123\",\"fields\":[\"name\",\"email\"]}\n- Metric: {\"moduleID\":\"123\",\"field\":\"amount\",\"aggregate\":\"sum\"}\n- Chart: {\"chartID\":\"456\"}\n- Content: {\"body\":\"<p>Hello</p>\"}")),
	), handleCreatePageBlock)

	s.AddTool(mcp.NewTool("update_pageblock",
		mcp.WithDescription("Update a page block's title, description, or options"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("pageID", mcp.Description("Page ID"), mcp.Required()),
		mcp.WithString("blockID", mcp.Description("Block index (0-based) or block title"), mcp.Required()),
		mcp.WithString("title", mcp.Description("New block title")),
		mcp.WithString("description", mcp.Description("New block description")),
		mcp.WithString("options", mcp.Description("JSON object with updated block options")),
	), handleUpdatePageBlock)

	s.AddTool(mcp.NewTool("delete_pageblock",
		mcp.WithDescription("Remove a block from a page"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("pageID", mcp.Description("Page ID"), mcp.Required()),
		mcp.WithString("blockID", mcp.Description("Block index (0-based) or block title"), mcp.Required()),
	), handleDeletePageBlock)

	s.AddResourceTemplate(mcp.NewResourceTemplate("mcp://pageblock/{pageID}/{idx}", "Page block by index",
		mcp.WithTemplateDescription("Read a specific page block by page ID and index"),
	), handlePageBlockResource)

	s.AddResourceTemplate(mcp.NewResourceTemplate("mcp://pageblock/{pageID}/list", "Page blocks list",
		mcp.WithTemplateDescription("List all blocks on a page"),
	), handlePageBlockListResource)
}

func findBlockIndex(blocks types.PageBlocks, blockID string) int {
	for i, b := range blocks {
		if b.Title == blockID {
			return i
		}
	}
	var idx int
	fmt.Sscanf(blockID, "%d", &idx)
	if idx >= 0 && idx < len(blocks) {
		return idx
	}
	return -1
}

func handleReadPageBlock(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	pID := parseUint64(args, "pageID")
	blockID := getString(args, "blockID")

	p, err := service.DefaultPage.FindByID(ctx, nsID, pID)
	if err != nil {
		return errorResult(fmt.Errorf("page not found: %w", err)), nil
	}

	idx := findBlockIndex(p.Blocks, blockID)
	if idx < 0 {
		return textResult("Block not found"), nil
	}

	return jsonResult(map[string]interface{}{
		"index": idx,
		"block": p.Blocks[idx],
	}), nil
}

func handleListPageBlocks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	pID := parseUint64(args, "pageID")

	p, err := service.DefaultPage.FindByID(ctx, nsID, pID)
	if err != nil {
		return errorResult(fmt.Errorf("page not found: %w", err)), nil
	}

	return jsonResult(map[string]interface{}{
		"pageID": fmt.Sprintf("%d", p.ID),
		"title":  p.Title,
		"blocks": p.Blocks,
		"count":  len(p.Blocks),
	}), nil
}

func handleCreatePageBlock(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	pID := parseUint64(args, "pageID")
	kind := getString(args, "kind")
	title := getString(args, "title")
	description := getString(args, "description")
	optionsJSON := getString(args, "options")

	p, err := service.DefaultPage.FindByID(ctx, nsID, pID)
	if err != nil {
		return errorResult(fmt.Errorf("page not found: %w", err)), nil
	}

	var options map[string]interface{}
	if optionsJSON != "" {
		if err := json.Unmarshal([]byte(optionsJSON), &options); err != nil {
			return textResult(fmt.Sprintf("Invalid options JSON: %v", err)), nil
		}
	}

	block := types.PageBlock{
		Title:       title,
		Description: description,
		Kind:        kind,
		Options:     options,
	}

	p.Blocks = append(p.Blocks, block)

	updated, err := service.DefaultPage.Update(ctx, p)
	if err != nil {
		return errorResult(err), nil
	}

	newIdx := len(updated.Blocks) - 1
	return jsonResult(map[string]interface{}{
		"pageID": fmt.Sprintf("%d", updated.ID),
		"index":  newIdx,
		"block":  updated.Blocks[newIdx],
	}), nil
}

func handleUpdatePageBlock(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	pID := parseUint64(args, "pageID")
	blockID := getString(args, "blockID")

	p, err := service.DefaultPage.FindByID(ctx, nsID, pID)
	if err != nil {
		return errorResult(fmt.Errorf("page not found: %w", err)), nil
	}

	idx := findBlockIndex(p.Blocks, blockID)
	if idx < 0 {
		return textResult("Block not found"), nil
	}

	block := &p.Blocks[idx]

	if v := getString(args, "title"); v != "" {
		block.Title = v
	}
	if v := getString(args, "description"); v != "" {
		block.Description = v
	}
	if v := getString(args, "options"); v != "" {
		var options map[string]interface{}
		if err := json.Unmarshal([]byte(v), &options); err != nil {
			return textResult(fmt.Sprintf("Invalid options JSON: %v", err)), nil
		}
		block.Options = options
	}

	updated, err := service.DefaultPage.Update(ctx, p)
	if err != nil {
		return errorResult(err), nil
	}

	return jsonResult(updated.Blocks[idx]), nil
}

func handleDeletePageBlock(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	pID := parseUint64(args, "pageID")
	blockID := getString(args, "blockID")

	p, err := service.DefaultPage.FindByID(ctx, nsID, pID)
	if err != nil {
		return errorResult(fmt.Errorf("page not found: %w", err)), nil
	}

	idx := findBlockIndex(p.Blocks, blockID)
	if idx < 0 {
		return textResult("Block not found"), nil
	}

	deleted := p.Blocks[idx]
	p.Blocks = append(p.Blocks[:idx], p.Blocks[idx+1:]...)

	_, err = service.DefaultPage.Update(ctx, p)
	if err != nil {
		return errorResult(err), nil
	}

	return jsonResult(map[string]interface{}{
		"deleted": deleted,
		"message": "Block removed",
	}), nil
}

func handlePageBlockResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	ctx = withAuth(ctx)
	pageIDStr, _ := request.Params.Arguments["pageID"].(string)
	idxStr, _ := request.Params.Arguments["idx"].(string)

	var pageID uint64
	var blockIdx int
	fmt.Sscanf(pageIDStr, "%d", &pageID)
	fmt.Sscanf(idxStr, "%d", &blockIdx)

	_, err := service.DefaultPage.FindByID(ctx, 0, pageID)
	if err != nil {
		return nil, err
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:  "mcp://pageblock/" + pageIDStr + "/" + idxStr,
			Text: fmt.Sprintf(`{"pageID":"%s","index":%d}`, pageIDStr, blockIdx),
		},
	}, nil
}

func handlePageBlockListResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	ctx = withAuth(ctx)
	pageIDStr, _ := request.Params.Arguments["pageID"].(string)
	var pageID uint64
	fmt.Sscanf(pageIDStr, "%d", &pageID)

	p, err := service.DefaultPage.FindByID(ctx, 0, pageID)
	if err != nil {
		return nil, err
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:  "mcp://pageblock/" + pageIDStr + "/list",
			Text: toJSON(p.Blocks),
		},
	}, nil
}
