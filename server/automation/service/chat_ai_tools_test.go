package service

import (
	"context"
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/stretchr/testify/require"
)

func authoringTool(t *testing.T, name string) func(context.Context, map[string]string) string {
	t.Helper()
	for _, d := range (&workflowChat{}).aiAuthoringTools() {
		if d.Name == name {
			return d.Handler
		}
	}
	t.Fatalf("no tool %q", name)
	return nil
}

func TestAuthoringTools_ConfirmationRules(t *testing.T) {
	names := map[string]bool{ // tool -> asks the user first
		"list_ai_functions":             false,
		"validate_ai_workflow":          false,
		"list_ai_templates":             false,
		"create_ai_workflow":            true,
		"create_workflow_from_template": true,
	}

	defs := (&workflowChat{}).aiAuthoringTools()
	require.Len(t, defs, len(names))
	for _, d := range defs {
		want, known := names[d.Name]
		require.True(t, known, d.Name)
		got := aiagent.DefaultNeedsConfirm([]aiagent.Call{{Name: d.Name}})
		require.Equal(t, want, got, "%s: the assistant must %sask before running it", d.Name, map[bool]string{true: "", false: "not "}[want])
	}
}

func TestAuthoringTools_ValidateAndCreate(t *testing.T) {
	svc := templateWorkflowService(t, func(context.Context, string, string, bool) (*aiagent.AgentResult, error) {
		return &aiagent.AgentResult{Success: true}, nil
	})
	_ = svc // installs itself as DefaultWorkflow for the converter

	ctx := auth.SetIdentityToContext(context.Background(), auth.Authenticated(9))
	validate := authoringTool(t, "validate_ai_workflow")

	out := validate(ctx, map[string]string{"flow": anomalyFlow})
	require.Contains(t, out, "OK: 2 steps, 1 links")
	require.Contains(t, out, "anomaly:finding / onCreate")
	require.Contains(t, out, "Nothing was saved")

	// problems come back as a list the model can act on
	out = validate(ctx, map[string]string{"flow": `{"name":"x","steps":[{"id":"a","function":"aiThink"}],"flow":[]}`})
	require.Contains(t, out, "Problems to fix:")
	require.Contains(t, out, `unknown function "aiThink"`)

	// a description that is not even JSON
	out = validate(ctx, map[string]string{"flow": "make it so"})
	require.Contains(t, out, "not valid")

	// nobody signed in
	out = validate(context.Background(), map[string]string{"flow": anomalyFlow})
	require.Contains(t, out, "signed-in user")

	// create refuses an invalid flow without touching the platform
	out = authoringTool(t, "create_ai_workflow")(ctx, map[string]string{"flow": `{"name":"x","steps":[],"flow":[]}`})
	require.Contains(t, out, "Not created")
	require.Contains(t, out, "steps is empty")
}

func TestAuthoringTools_Templates(t *testing.T) {
	out := authoringTool(t, "list_ai_templates")(context.Background(), nil)
	require.Contains(t, out, "anomaly-explain")
	require.Contains(t, out, "risk-escalation-review")

	out = authoringTool(t, "create_workflow_from_template")(auth.SetIdentityToContext(context.Background(), auth.Authenticated(9)), map[string]string{"template": "nope"})
	require.Contains(t, out, "unknown template")
}

func TestAuthoringSystemPrompt(t *testing.T) {
	for _, phrase := range []string{"list_ai_functions", "validate_ai_workflow", "create_ai_workflow", "list_ai_templates", "@prompt:", `starts with "="`} {
		require.True(t, strings.Contains(aiAuthoringSystemPrompt, phrase), phrase)
	}
}
