package mcp

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/server"
)

var globalMcp *server.MCPServer

func InitMcp(ctx context.Context) {
	s := server.NewMCPServer(
		"lowcode-server", // имя сервера
		"1.0.0",          // версия
	)
	globalMcp = s

	go func(s *server.MCPServer) {

		// Запускаем сервер через STDIO (стандартный канал для MCP)
		if err := server.ServeStdio(s); err != nil {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}(s)

}
