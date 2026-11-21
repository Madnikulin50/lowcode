package service

import (
	"context"
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

func workflowTool(t *testing.T, name string) chat.ToolDef {
	t.Helper()
	for _, d := range chatWorkflowToolDefs() {
		if d.Name == name {
			return d
		}
	}
	t.Fatalf("tool %q not defined", name)
	return chat.ToolDef{}
}

func TestChatWorkflowTools_MutatingFlags(t *testing.T) {
	// running a workflow or installing a template changes data, so chat must
	// ask for the user's "да" first; reading must not
	mutating := map[string]bool{
		"list_workflows":            false,
		"list_workflow_templates":   false,
		"workflow_status":           false,
		"run_workflow":              true,
		"install_workflow_template": true,
	}
	for name, want := range mutating {
		if got := workflowTool(t, name).Mutating; got != want {
			t.Errorf("%s: Mutating = %v, want %v", name, got, want)
		}
	}
	if len(chatWorkflowToolDefs()) != len(mutating) {
		t.Errorf("unexpected tool set: %d tools", len(chatWorkflowToolDefs()))
	}
}

func TestChatWorkflowTools_Validation(t *testing.T) {
	ctx := context.Background()

	for _, c := range []struct {
		tool   string
		params map[string]string
		want   string
	}{
		{"run_workflow", map[string]string{}, "workflow is required"},
		{"run_workflow", map[string]string{"workflow": "w", "input": "[1]"}, "input must be a JSON object"},
		{"workflow_status", map[string]string{"sessionID": "x"}, "numeric session ID"},
		{"install_workflow_template", map[string]string{}, "template is required"},
	} {
		if got := workflowTool(t, c.tool).Handler(ctx, c.params); !strings.Contains(got, c.want) {
			t.Errorf("%s(%v) = %q, want it to contain %q", c.tool, c.params, got, c.want)
		}
	}

	if got := workflowTool(t, "list_workflow_templates").Handler(ctx, nil); !strings.Contains(got, "anomaly-explain") {
		t.Errorf("template list = %q", got)
	}
}

func TestWorkflowsToolkitReachesTheAssistant(t *testing.T) {
	cat := aiagent.NewCatalog()
	RegisterComposeToolKits(cat)

	kit, ok := cat.Get("workflows")
	if !ok || len(kit.Tools) == 0 {
		t.Fatal("workflows toolkit is not registered")
	}

	found := false
	for _, name := range aiagent.AssistantKitNames() {
		if name == "workflows" {
			found = true
		}
	}
	if !found {
		t.Error("the assistant agent does not list the workflows toolkit, so chat would never offer it")
	}
}

func TestChatPromptTools(t *testing.T) {
	mutating := map[string]bool{
		"list_prompts":    false,
		"get_prompt":      false,
		"evaluate_prompt": false,
		"save_prompt":     true, // changes what live workflows tell the model
		"activate_prompt": true,
	}
	defs := chatPromptToolDefs()
	if len(defs) != len(mutating) {
		t.Fatalf("unexpected tool set: %d tools", len(defs))
	}
	for _, d := range defs {
		want, ok := mutating[d.Name]
		if !ok || d.Mutating != want {
			t.Errorf("%s: Mutating = %v, want %v (known: %v)", d.Name, d.Mutating, want, ok)
		}
	}

	cat := aiagent.NewCatalog()
	RegisterComposeToolKits(cat)
	if _, ok := cat.Get("prompts"); !ok {
		t.Fatal("prompts toolkit is not registered")
	}
	found := false
	for _, name := range aiagent.AssistantKitNames() {
		found = found || name == "prompts"
	}
	if !found {
		t.Error("the assistant does not list the prompts toolkit")
	}

	for _, d := range defs {
		if d.Name == "save_prompt" {
			if got := d.Handler(context.Background(), map[string]string{"handle": "x", "text": "t", "cases": "nope"}); !strings.Contains(got, "cases must be a JSON array") {
				t.Errorf("bad cases: %q", got)
			}
		}
		if d.Name == "activate_prompt" {
			if got := d.Handler(context.Background(), map[string]string{"handle": "x", "version": "abc"}); !strings.Contains(got, "version number") {
				t.Errorf("bad version: %q", got)
			}
		}
	}
}
