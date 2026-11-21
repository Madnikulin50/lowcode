package handlers

import (
	"context"

	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/mark3labs/mcp-go/server"
)

func Init(ctx context.Context, s *server.MCPServer) {
	initNamespace(ctx, s)
	initModules(ctx, s)
	initModuleRecords(ctx, s)
	initPages(ctx, s)
	initCharts(ctx, s)
	registerToolDefs(s, service.VersionToolDefs()...)
	initPageBlocks(ctx, s)
	initRecordResources(ctx, s)
	initMail(ctx, s)
	initRules(ctx, s)
	if mcpScriptsEnabled() {
		initAIScripts(ctx, s)
	}
	initAgents(ctx, s)
	initWorkflows(ctx, s)
	initPrompts(ctx, s)
	initSkills(ctx, s)
	initGonec(ctx, s)
}
