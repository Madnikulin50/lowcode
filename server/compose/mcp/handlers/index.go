package handlers

import (
	"context"

	"github.com/mark3labs/mcp-go/server"
)

func Init(ctx context.Context, s *server.MCPServer) {
	initNamespace(ctx, s)
	initModules(ctx, s)
	initModuleRecords(ctx, s)
	initPages(ctx, s)
	initCharts(ctx, s)
}
