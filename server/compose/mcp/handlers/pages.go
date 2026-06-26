package handlers

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func initPages(ctx context.Context, s *server.MCPServer) {
	s.AddTool(mcp.NewTool("query_page",
		mcp.WithDescription("Получить страницу из по части имени"),
		mcp.WithString("name",
			mcp.Description("Имя для поиска (частичное совпадение)"),
			mcp.Required(),
		),
	), handleQueryPages)
}

func handleQueryPages(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Параметр 'name' обязателен")
	}

	nameRaw, ok := arguments["name"]
	if !ok {
		return nil, fmt.Errorf("Параметр 'name' обязателен")
	}
	name, ok := nameRaw.(string)
	if !ok {
		return nil, fmt.Errorf("Параметр 'name' должен быть строкой")
	}

	// Запрос через GORM

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(name),
			},
		},
	}, nil
}
