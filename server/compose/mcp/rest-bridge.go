package mcp

import (
	"context"

	"github.com/madnikulin50/lowcode/server/compose/mcp/handlers"
	"github.com/mark3labs/mcp-go/server"
)

func ApplyToServer(ctx context.Context,
	namespaceID uint64) {
	InitNamespace(ctx, globalMcp, namespaceID)
}

func InitNamespace(ctx context.Context,
	s *server.MCPServer,
	namespaceID uint64) {

	handlers.Init(ctx, s)
}
