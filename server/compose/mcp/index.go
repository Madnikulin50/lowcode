package mcp

import (
	"context"
	"log"
	"os"

	"github.com/madnikulin50/lowcode/server/compose/mcp/handlers"
	"github.com/mark3labs/mcp-go/server"
)

var globalMcp *server.MCPServer

func InitMcp(ctx context.Context) {
	s := server.NewMCPServer(
		"lowcode-server",
		"1.0.0",
	)
	handlers.Init(ctx, s)
	globalMcp = s

	if os.Getenv("MCP_STDIO") == "true" {
		go func(s *server.MCPServer) {
			log.Println("MCP stdio server starting")
			if err := server.ServeStdio(s); err != nil {
				log.Fatalf("MCP stdio server error: %v", err)
			}
		}(s)
	}

	if addr := os.Getenv("MCP_SSE_ADDR"); addr != "" {
		sse := server.NewSSEServer(s)
		go func() {
			log.Printf("MCP SSE server starting on %s", addr)
			if err := sse.Start(addr); err != nil {
				log.Fatalf("MCP SSE server error: %v", err)
			}
		}()
	}
}
