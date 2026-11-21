package handlers

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestInitWorkflowsRegistersTools(t *testing.T) {
	s := server.NewMCPServer("test", "0")
	initWorkflows(context.Background(), s)

	tools := s.ListTools()
	for _, name := range []string{"workflow_list", "workflow_run", "workflow_session", "workflow_templates", "workflow_install_template"} {
		if _, ok := tools[name]; !ok {
			t.Errorf("tool %q was not registered", name)
		}
	}
}

func callWith(args map[string]interface{}) mcp.CallToolRequest {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	return req
}

func resultText(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if res == nil || len(res.Content) == 0 {
		t.Fatal("empty result")
	}
	tc, ok := res.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("unexpected content %T", res.Content[0])
	}
	return tc.Text
}

func TestWorkflowToolsRejectBadInput(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name string
		run  func() (*mcp.CallToolResult, error)
		want string
	}{
		{"run without workflow", func() (*mcp.CallToolResult, error) {
			return handleWorkflowRun(ctx, callWith(map[string]interface{}{}))
		}, "workflow is required"},
		{"run with non-object input", func() (*mcp.CallToolResult, error) {
			return handleWorkflowRun(ctx, callWith(map[string]interface{}{"workflow": "w", "input": "[1]"}))
		}, "input must be a JSON object"},
		{"session with non-numeric id", func() (*mcp.CallToolResult, error) {
			return handleWorkflowSession(ctx, callWith(map[string]interface{}{"sessionID": "abc"}))
		}, "numeric session ID"},
		{"install without template", func() (*mcp.CallToolResult, error) {
			return handleWorkflowInstallTemplate(ctx, callWith(map[string]interface{}{}))
		}, "template is required"},
	}

	for _, c := range cases {
		res, err := c.run()
		if err != nil {
			t.Fatalf("%s: handlers must report problems in the result, got Go error %v", c.name, err)
		}
		if got := resultText(t, res); !contains(got, c.want) {
			t.Errorf("%s: got %q, want it to contain %q", c.name, got, c.want)
		}
	}
}

func TestWorkflowTemplatesListed(t *testing.T) {
	res, err := handleWorkflowTemplates(context.Background(), callWith(nil))
	if err != nil {
		t.Fatal(err)
	}
	text := resultText(t, res)
	for _, key := range []string{"anomaly-explain", "risk-escalation-review"} {
		if !contains(text, key) {
			t.Errorf("template %q missing from %s", key, text)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestInitPromptsRegistersTools(t *testing.T) {
	s := server.NewMCPServer("test", "0")
	initPrompts(context.Background(), s)

	tools := s.ListTools()
	for _, name := range []string{"prompt_list", "prompt_get", "prompt_history", "prompt_save", "prompt_activate", "prompt_delete", "prompt_eval"} {
		if _, ok := tools[name]; !ok {
			t.Errorf("tool %q was not registered", name)
		}
	}
}

func TestPromptToolsReportProblemsInTheResult(t *testing.T) {
	ctx := context.Background()

	// the library is not wired in a unit test: every call says so, none panics
	for name, call := range map[string]func() (*mcp.CallToolResult, error){
		"list": func() (*mcp.CallToolResult, error) { return handlePromptList(ctx, callWith(nil)) },
		"get": func() (*mcp.CallToolResult, error) {
			return handlePromptGet(ctx, callWith(map[string]interface{}{"handle": "x"}))
		},
		"history": func() (*mcp.CallToolResult, error) {
			return handlePromptHistory(ctx, callWith(map[string]interface{}{"handle": "x"}))
		},
		"save": func() (*mcp.CallToolResult, error) {
			return handlePromptSave(ctx, callWith(map[string]interface{}{"handle": "x", "text": "t"}))
		},
		"delete": func() (*mcp.CallToolResult, error) {
			return handlePromptDelete(ctx, callWith(map[string]interface{}{"handle": "x"}))
		},
		"eval": func() (*mcp.CallToolResult, error) {
			return handlePromptEval(ctx, callWith(map[string]interface{}{"handle": "x"}))
		},
		"activate": func() (*mcp.CallToolResult, error) {
			return handlePromptActivate(ctx, callWith(map[string]interface{}{"handle": "x", "version": 2}))
		},
	} {
		res, err := call()
		if err != nil {
			t.Fatalf("%s: handlers must not return Go errors, got %v", name, err)
		}
		if got := resultText(t, res); !contains(got, "not available") {
			t.Errorf("%s: got %q", name, got)
		}
	}

	res, _ := handlePromptActivate(ctx, callWith(map[string]interface{}{"handle": "x"}))
	if got := resultText(t, res); !contains(got, "version is required") {
		t.Errorf("activate without a version: %q", got)
	}

	res, _ = handlePromptSave(ctx, callWith(map[string]interface{}{"handle": "x", "text": "t", "cases": "{not json"}))
	if got := resultText(t, res); !contains(got, "cases must be a JSON array") {
		t.Errorf("bad cases: %q", got)
	}
}

func TestGetInt(t *testing.T) {
	for _, c := range []struct {
		in   interface{}
		want int
	}{
		{float64(3), 3}, // what a JSON number becomes
		{3, 3},
		{int64(3), 3},
		{" 3 ", 3},
		{"x", 0},
		{nil, 0},
		{true, 0},
	} {
		if got := getInt(map[string]interface{}{"v": c.in}, "v"); got != c.want {
			t.Errorf("getInt(%#v) = %d, want %d", c.in, got, c.want)
		}
	}
	if getInt(map[string]interface{}{}, "missing") != 0 {
		t.Error("missing key should be 0")
	}
}
