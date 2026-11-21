package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	automationService "github.com/madnikulin50/lowcode/server/automation/service"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

// chatWorkflowToolDefs lets the assistant and agents list, run and follow
// automation workflows - the AI, approval and integration processes built in
// the workflow editor. Running one changes data, so it is Mutating and goes
// through the same "да" confirmation as other writes.
func chatWorkflowToolDefs() []chat.ToolDef {
	return []chat.ToolDef{
		{
			Name:        "list_workflows",
			Description: "List the enabled automation workflows the user may run.",
			Params: []chat.ParamDef{
				{Name: "query", Type: "string", Description: "Optional text to narrow the list"},
			},
			Handler: chatListWorkflows,
		},
		{
			Name:        "run_workflow",
			Description: "Run an automation workflow by ID or handle (see list_workflows). Use async=true for long processes (approvals, delays) and follow them with workflow_status.",
			Mutating:    true,
			Params: []chat.ParamDef{
				{Name: "workflow", Type: "string", Required: true, Description: "Workflow ID or handle"},
				{Name: "input", Type: "object", Description: "Object passed to the workflow as its input"},
				{Name: "async", Type: "boolean", Description: "true to return the session ID without waiting"},
			},
			Handler: chatRunWorkflow,
		},
		{
			Name:        "list_workflow_templates",
			Description: "List ready-made AI workflow templates that can be installed.",
			Handler: func(context.Context, map[string]string) string {
				return chatJSON(automationService.AITemplates())
			},
		},
		{
			Name:        "install_workflow_template",
			Description: "Install an AI workflow template (see list_workflow_templates). The workflow and its trigger are created disabled, to be reviewed and enabled in the workflow editor.",
			Mutating:    true,
			Params: []chat.ParamDef{
				{Name: "template", Type: "string", Required: true, Description: "Template key"},
			},
			Handler: chatInstallWorkflowTemplate,
		},
		{
			Name:        "workflow_status",
			Description: "Show the status of a workflow run started with run_workflow.",
			Params: []chat.ParamDef{
				{Name: "sessionID", Type: "string", Required: true, Description: "Session ID returned by run_workflow"},
			},
			Handler: chatWorkflowStatus,
		},
	}
}

func chatListWorkflows(ctx context.Context, params map[string]string) string {
	list, err := automationService.ListWorkflowBriefs(ctx, params["query"])
	if err != nil {
		return err.Error()
	}
	if len(list) == 0 {
		return "No workflows found"
	}
	return chatJSON(list)
}

func chatRunWorkflow(ctx context.Context, params map[string]string) string {
	ref := strings.TrimSpace(params["workflow"])
	if ref == "" {
		return "workflow is required"
	}

	input := map[string]interface{}{}
	if raw := strings.TrimSpace(params["input"]); raw != "" {
		if err := json.Unmarshal([]byte(raw), &input); err != nil {
			return fmt.Sprintf("input must be a JSON object: %v", err)
		}
	}

	async, _ := strconv.ParseBool(strings.TrimSpace(params["async"]))

	res, err := automationService.ExecWorkflowByRef(ctx, ref, input, async)
	if err != nil {
		return err.Error()
	}
	return chatJSON(res)
}

func chatWorkflowStatus(ctx context.Context, params map[string]string) string {
	id, err := strconv.ParseUint(strings.TrimSpace(params["sessionID"]), 10, 64)
	if err != nil || id == 0 {
		return "sessionID must be a numeric session ID"
	}

	brief, err := automationService.WorkflowSessionBrief(ctx, id)
	if err != nil {
		return err.Error()
	}
	return chatJSON(brief)
}

func chatJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf(`{"error":%q}`, err.Error())
	}
	return string(b)
}

func chatInstallWorkflowTemplate(ctx context.Context, params map[string]string) string {
	key := strings.TrimSpace(params["template"])
	if key == "" {
		return "template is required"
	}

	wf, err := automationService.InstallAITemplate(ctx, key)
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("Installed workflow %q (ID %d), disabled - review and enable it in the workflow editor.", wf.Handle, wf.ID)
}
