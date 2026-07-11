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

func initCreate(ctx context.Context, s *server.MCPServer) {
	s.AddTool(mcp.NewTool("create_module",
		mcp.WithDescription("Create a new module with fields"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("name", mcp.Description("Module name"), mcp.Required()),
		mcp.WithString("handle", mcp.Description("URL-safe handle")),
		mcp.WithString("fields",
			mcp.Description(`JSON array of fields. Each field: {"name":"...","kind":"String|Number|DateTime|Select|Bool|User|Record|File|URL|Email","required":true/false,"label":"..."}`),
			mcp.Required(),
		),
	), handleCreateModule)

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

	s.AddTool(mcp.NewTool("create_page",
		mcp.WithDescription("Create a new page"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("title", mcp.Description("Page title"), mcp.Required()),
		mcp.WithString("handle", mcp.Description("URL-safe handle")),
		mcp.WithString("description", mcp.Description("Page description")),
		mcp.WithString("moduleID", mcp.Description("Module ID that this page is for")),
		mcp.WithString("selfID", mcp.Description("Parent page ID (0 for root)")),
		mcp.WithString("blocks",
			mcp.Description(`JSON array of page blocks. Example: [{"kind":"RecordList","options":{"fields":["field1","field2"]}}]`),
		),
	), handleCreatePage)
}

func handleCreateModule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid arguments")
	}

	namespaceID := parseUint64(args, "namespaceID")
	name := getString(args, "name")
	handle := getString(args, "handle")
	fieldsJSON := getString(args, "fields")

	var fields []map[string]interface{}
	if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("Invalid fields JSON: %v", err)), nil
	}

	fieldSet := make(types.ModuleFieldSet, 0, len(fields))
	for _, f := range fields {
		fieldSet = append(fieldSet, &types.ModuleField{
			Name:     getFieldString(f, "name"),
			Kind:     getFieldString(f, "kind"),
			Label:    getFieldString(f, "label"),
			Required: getFieldBool(f, "required"),
		})
	}

	mod := &types.Module{
		NamespaceID: namespaceID,
		Name:        name,
		Handle:      handle,
		Fields:      fieldSet,
		Config:      types.ModuleConfig{},
	}

	created, err := service.DefaultModule.Create(ctx, mod)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("Failed to create module: %v", err)), nil
	}

	result, _ := json.Marshal(map[string]interface{}{
		"moduleID": fmt.Sprintf("%d", created.ID),
		"name":     created.Name,
		"handle":   created.Handle,
		"fields":   len(created.Fields),
	})
	return mcp.NewToolResultText(string(result)), nil
}

func handleCreateChart(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid arguments")
	}

	namespaceID := parseUint64(args, "namespaceID")
	name := getString(args, "name")
	handle := getString(args, "handle")
	configJSON := getString(args, "config")

	var config types.ChartConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("Invalid config JSON: %v", err)), nil
	}

	chart := &types.Chart{
		NamespaceID: namespaceID,
		Name:        name,
		Handle:      handle,
		Config:      config,
	}

	created, err := service.DefaultChart.Create(ctx, chart)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("Failed to create chart: %v", err)), nil
	}

	result, _ := json.Marshal(map[string]interface{}{
		"chartID": fmt.Sprintf("%d", created.ID),
		"name":    created.Name,
		"handle":  created.Handle,
	})
	return mcp.NewToolResultText(string(result)), nil
}

func handleCreatePage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid arguments")
	}

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
			return mcp.NewToolResultText(fmt.Sprintf("Invalid blocks JSON: %v", err)), nil
		}
	}

	page := &types.Page{
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

	created, err := service.DefaultPage.Create(ctx, page)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("Failed to create page: %v", err)), nil
	}

	result, _ := json.Marshal(map[string]interface{}{
		"pageID": fmt.Sprintf("%d", created.ID),
		"title":  created.Title,
		"handle": created.Handle,
	})
	return mcp.NewToolResultText(string(result)), nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func parseUint64(m map[string]interface{}, key string) uint64 {
	s := getString(m, key)
	if s == "" {
		return 0
	}
	var v uint64
	fmt.Sscanf(s, "%d", &v)
	return v
}

func getFieldString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case string:
			return val
		case bool:
			if val {
				return "true"
			}
			return "false"
		}
	}
	return ""
}

func getFieldBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}
