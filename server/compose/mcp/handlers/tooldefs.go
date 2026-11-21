package handlers

import (
	"context"
	"fmt"

	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerToolDefs registers each service.Def as an MCP tool on s. It is the
// MCP-side half of the shared tool registry in server/compose/service
// (tooldef.go, tool_modules.go, tool_charts.go, tool_pages.go) — the other
// half, ToChatToolDefs, feeds the same Defs to chat. Every registered tool
// gets the same withAuth(ctx) + JSON/text result wrapping the hand-written
// handlers used before this registry existed.
func registerToolDefs(s *server.MCPServer, defs ...service.Def) {
	for _, d := range defs {
		s.AddTool(mcpTool(d), mcpToolHandler(d))
	}
}

func mcpTool(d service.Def) mcp.Tool {
	opts := []mcp.ToolOption{mcp.WithDescription(d.Description)}
	for _, p := range d.Params {
		popts := []mcp.PropertyOption{mcp.Description(p.Description)}
		if p.Required {
			popts = append(popts, mcp.Required())
		}
		opts = append(opts, mcp.WithString(p.Name, popts...))
	}
	return mcp.NewTool(d.Name, opts...)
}

func mcpToolHandler(d service.Def) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ctx = withAuth(ctx)
		args := make(map[string]string)
		for k, v := range argsMap(request) {
			if s, ok := v.(string); ok {
				args[k] = s
			} else {
				args[k] = fmt.Sprintf("%v", v)
			}
		}
		data, err := d.Handler(ctx, args)
		if err != nil {
			return errorResult(err), nil
		}
		if s, ok := data.(string); ok {
			return textResult(s), nil
		}
		return jsonResult(data), nil
	}
}
