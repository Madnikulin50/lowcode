package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var (
	AgentRegistry *aiagent.Registry
	AgentClient   *chat.Client
)

func SetAgentRegistry(reg *aiagent.Registry) {
	AgentRegistry = reg
}

func SetAgentClient(c *chat.Client) {
	AgentClient = c
}

func initAgents(ctx context.Context, s *server.MCPServer) {
	s.AddTool(mcp.NewTool("agent_call",
		mcp.WithDescription("Call an AI agent by name to perform a task. Agents have access to all data and tools."),
		mcp.WithString("agent", mcp.Description("Agent handle from agent_list (crud-agent, assistant, cmdb-operator, or a custom catalog agent)"), mcp.Required()),
		mcp.WithString("task", mcp.Description("Task description for the agent"), mcp.Required()),
		mcp.WithString("context", mcp.Description("JSON object with additional context data")),
	), handleAgentCall)

	s.AddTool(mcp.NewTool("agent_list",
		mcp.WithDescription("List all available AI agents"),
	), handleAgentList)
}

func handleAgentCall(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	agentName := getString(args, "agent")
	task := getString(args, "task")
	contextJSON := getString(args, "context")

	var ctxData map[string]interface{}
	if contextJSON != "" {
		if err := json.Unmarshal([]byte(contextJSON), &ctxData); err != nil {
			return textResult(fmt.Sprintf("Invalid context JSON: %v", err)), nil
		}
	}

	if AgentRegistry == nil {
		return textResult("Agent registry not initialized"), nil
	}

	result, err := AgentRegistry.RunAgent(ctx, agentName, task, ctxData)
	if err != nil {
		return errorResult(err), nil
	}

	return jsonResult(result), nil
}

func handleAgentList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	if AgentRegistry == nil {
		return textResult("Agent registry not initialized"), nil
	}

	return jsonResult(map[string]interface{}{
		"agents": AgentRegistry.ListInfo(),
	}), nil
}
