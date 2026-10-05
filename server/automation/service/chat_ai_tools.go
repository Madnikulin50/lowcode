package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

// aiAuthoringSystemPrompt tells the workflow assistant how to build AI
// workflows with the tools below. The order matters: a draft that has not
// been validated is a guess.
const aiAuthoringSystemPrompt = `

To build a workflow that uses AI steps:
1. Call list_ai_functions to see the functions, their arguments and results. Do not invent function or argument names.
2. Write the flow description (JSON) and call validate_ai_workflow with it. If it reports problems, fix exactly those and validate again.
3. Show the user what the workflow does in plain words, and only after they agree call create_ai_workflow. It is created switched off; tell the user to review and enable it in the editor.
For a common case first check list_ai_templates: create_workflow_from_template installs a ready-made one.
In a flow description an argument value that starts with "=" is an expression over workflow variables; anything else is a constant. Reuse a library prompt by giving the argument the value "@prompt:<handle>".`

func (c *workflowChat) aiAuthoringTools() []chat.ToolDef {
	return []chat.ToolDef{
		{
			Name:        "list_ai_functions",
			Description: "List the functions an AI workflow can call (AI steps and logging), with their arguments and results, plus the user prompts and gateways. Call this before writing a flow.",
			Handler: func(context.Context, map[string]string) string {
				return AIFunctionCatalog()
			},
		},
		{
			Name:        "validate_ai_workflow",
			Description: "Check a flow description against the platform without saving anything. Returns OK or the list of problems to fix.",
			Params: []chat.ParamDef{
				{Name: "flow", Type: "object", Required: true, Description: flowParamHelp},
			},
			Handler: validateAIWorkflowHandler,
		},
		{
			Name:        "create_ai_workflow",
			Description: "Create a workflow (and its trigger) from a flow description that passed validate_ai_workflow. It is created disabled and runs as the user.",
			Params: []chat.ParamDef{
				{Name: "flow", Type: "object", Required: true, Description: flowParamHelp},
			},
			Handler: createAIWorkflowHandler,
		},
		{
			Name:        "list_ai_templates",
			Description: "List ready-made AI workflow templates.",
			Handler: func(context.Context, map[string]string) string {
				out := make([]string, 0, 2)
				for _, t := range AITemplates() {
					out = append(out, fmt.Sprintf("%s - %s (trigger: %s)", t.Key, t.Description, t.Trigger))
				}
				return strings.Join(out, "\n")
			},
		},
		{
			Name:        "create_workflow_from_template",
			Description: "Install a ready-made AI workflow template. It is created disabled and runs as the user.",
			Params: []chat.ParamDef{
				{Name: "template", Type: "string", Required: true, Description: "Template key from list_ai_templates"},
			},
			Handler: func(ctx context.Context, p map[string]string) string {
				wf, err := InstallAITemplate(ctx, strings.TrimSpace(p["template"]))
				if err != nil {
					return err.Error()
				}
				return fmt.Sprintf("Created workflow %q (ID %d), disabled. Review and enable it in the workflow editor.", wf.Handle, wf.ID)
			},
		},
	}
}

const flowParamHelp = `Flow description: {"name","description","trigger":{"resourceType","eventType","constraints":[{"name","op","values"}]},"steps":[{"id","function"|"prompt"|"gateway","args":{...},"results":{"variable":"functionResult"}}],"flow":[{"from","to","when"}]}`

func validateAIWorkflowHandler(ctx context.Context, p map[string]string) string {
	compiled, errText := compileFromParams(ctx, p)
	if errText != "" {
		return errText
	}
	if len(compiled.Issues) > 0 {
		return "Problems to fix:\n- " + strings.Join(compiled.Issues, "\n- ")
	}
	return fmt.Sprintf("OK: %d steps, %d links, trigger: %s. Nothing was saved.", len(compiled.Workflow.Steps), len(compiled.Workflow.Paths), describeTrigger(compiled))
}

func createAIWorkflowHandler(ctx context.Context, p map[string]string) string {
	compiled, errText := compileFromParams(ctx, p)
	if errText != "" {
		return errText
	}
	if len(compiled.Issues) > 0 {
		return "Not created - problems to fix:\n- " + strings.Join(compiled.Issues, "\n- ")
	}

	wf, err := compiled.Install(ctx)
	if err != nil {
		return "Not created: " + err.Error()
	}
	return fmt.Sprintf("Created workflow %q (ID %d), disabled, trigger: %s. Review and enable it in the workflow editor.", wf.Handle, wf.ID, describeTrigger(compiled))
}

func compileFromParams(ctx context.Context, p map[string]string) (*AIFlowCompiled, string) {
	ident := auth.GetIdentityFromContext(ctx)
	if ident == nil || ident.Identity() == 0 {
		return nil, "A signed-in user is required."
	}
	if DefaultWorkflow == nil {
		return nil, "Workflow service not available."
	}

	spec, err := ParseAIFlow(p["flow"])
	if err != nil {
		return nil, err.Error()
	}
	return CompileAIFlow(DefaultWorkflow, spec, ident.Identity()), ""
}

func describeTrigger(c *AIFlowCompiled) string {
	if c.Trigger == nil {
		return "none (start it manually)"
	}
	return c.Trigger.ResourceType + " / " + c.Trigger.EventType
}
