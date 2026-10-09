package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	automationService "github.com/madnikulin50/lowcode/server/automation/service"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func initWorkflows(ctx context.Context, s *server.MCPServer) {
	s.AddTool(mcp.NewTool("workflow_list",
		mcp.WithDescription("List the enabled automation workflows you may run (AI, approval and integration processes)"),
		mcp.WithString("query", mcp.Description("Optional text to narrow the list")),
	), handleWorkflowList)

	s.AddTool(mcp.NewTool("workflow_run",
		mcp.WithDescription("Run an automation workflow by ID or handle and return its results. With async=true it returns the session ID immediately; long workflows (approvals, delays) should be run async and followed with workflow_session."),
		mcp.WithString("workflow", mcp.Description("Workflow ID or handle from workflow_list"), mcp.Required()),
		mcp.WithString("input", mcp.Description("JSON object passed to the workflow as its input")),
		mcp.WithBoolean("async", mcp.Description("Do not wait for the workflow to finish")),
	), handleWorkflowRun)

	s.AddTool(mcp.NewTool("workflow_templates",
		mcp.WithDescription("List ready-made AI workflow templates (for example explaining anomaly findings or reviewing risk escalations)"),
	), handleWorkflowTemplates)

	s.AddTool(mcp.NewTool("workflow_install_template",
		mcp.WithDescription("Install an AI workflow template from workflow_templates. The workflow and its trigger are created DISABLED and run as you; review and enable them in the workflow editor."),
		mcp.WithString("template", mcp.Description("Template key from workflow_templates"), mcp.Required()),
	), handleWorkflowInstallTemplate)

	s.AddTool(mcp.NewTool("workflow_session",
		mcp.WithDescription("Show the status of a workflow run started with workflow_run"),
		mcp.WithString("sessionID", mcp.Description("Session ID returned by workflow_run"), mcp.Required()),
	), handleWorkflowSession)
}

func handleWorkflowList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)

	list, err := automationService.ListWorkflowBriefs(ctx, getString(args, "query"))
	if err != nil {
		return errorResult(err), nil
	}
	if len(list) == 0 {
		return textResult("No workflows found"), nil
	}
	return jsonResult(map[string]interface{}{"workflows": list}), nil
}

func handleWorkflowRun(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)

	ref := getString(args, "workflow")
	if ref == "" {
		return textResult("workflow is required"), nil
	}

	input := map[string]interface{}{}
	if raw := getString(args, "input"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &input); err != nil {
			return textResult(fmt.Sprintf("input must be a JSON object: %v", err)), nil
		}
	}

	async, _ := args["async"].(bool)

	res, err := automationService.ExecWorkflowByRef(ctx, ref, input, async)
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(res), nil
}

func handleWorkflowSession(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)

	id := parseUint64(args, "sessionID")
	if id == 0 {
		return textResult("sessionID must be a numeric session ID"), nil
	}

	brief, err := automationService.WorkflowSessionBrief(ctx, id)
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(brief), nil
}

func handleWorkflowTemplates(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return jsonResult(map[string]interface{}{"templates": automationService.AITemplates()}), nil
}

func handleWorkflowInstallTemplate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	key := getString(argsMap(request), "template")
	if key == "" {
		return textResult("template is required"), nil
	}

	wf, err := automationService.InstallAITemplate(ctx, key)
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(map[string]interface{}{
		"installed":  true,
		"workflowID": wf.ID,
		"handle":     wf.Handle,
		"enabled":    wf.Enabled,
		"note":       "created disabled - review and enable it in the workflow editor",
	}), nil
}
