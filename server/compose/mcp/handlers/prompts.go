package handlers

import (
	"context"
	"strconv"
	"strings"

	automationService "github.com/madnikulin50/lowcode/server/automation/service"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Prompt library tools. A prompt is referred to from AI workflow steps and
// rule chain AI nodes as "@prompt:<handle>" (the active version) or
// "@prompt:<handle>@<n>" (pinned to version n).
func initPrompts(ctx context.Context, s *server.MCPServer) {
	s.AddTool(mcp.NewTool("prompt_list",
		mcp.WithDescription("List the prompts in the prompt library (name, active version, number of versions and test cases)"),
	), handlePromptList)

	s.AddTool(mcp.NewTool("prompt_get",
		mcp.WithDescription("Show a prompt's text. Without a version: the active one."),
		mcp.WithString("handle", mcp.Description("Prompt handle"), mcp.Required()),
		mcp.WithNumber("version", mcp.Description("Version number")),
	), handlePromptGet)

	s.AddTool(mcp.NewTool("prompt_history",
		mcp.WithDescription("List all versions of a prompt, oldest first, with the note saved with each"),
		mcp.WithString("handle", mcp.Description("Prompt handle"), mcp.Required()),
	), handlePromptHistory)

	s.AddTool(mcp.NewTool("prompt_save",
		mcp.WithDescription("Save a new version of a prompt (creating the prompt if the handle is new). Old versions are kept. The new version becomes the one in use only if activate=true (the first version always is), so you can test it with prompt_eval before switching."),
		mcp.WithString("handle", mcp.Description("Lowercase letters, digits, _ - . ; starts with a letter"), mcp.Required()),
		mcp.WithString("text", mcp.Description("The instruction. For rule chain nodes it may contain {{variables}}."), mcp.Required()),
		mcp.WithString("description", mcp.Description("What the prompt is for (kept from the previous version if omitted)")),
		mcp.WithString("note", mcp.Description("Why this version exists")),
		mcp.WithString("cases", mcp.Description(`Test cases as a JSON array, e.g. [{"name":"outage","inputs":{"ticket":"server down"},"expect":{"priority":"high"}}]. Omit to keep the previous version's cases; "[]" clears them.`)),
		mcp.WithBoolean("activate", mcp.Description("Make this the version in use")),
	), handlePromptSave)

	s.AddTool(mcp.NewTool("prompt_activate",
		mcp.WithDescription("Make a version the one in use - to roll back to an earlier version, or forward to one saved without activation"),
		mcp.WithString("handle", mcp.Description("Prompt handle"), mcp.Required()),
		mcp.WithNumber("version", mcp.Description("Version number"), mcp.Required()),
	), handlePromptActivate)

	s.AddTool(mcp.NewTool("prompt_delete",
		mcp.WithDescription("Delete a prompt with all its versions. Steps that refer to it will fail until it is saved again."),
		mcp.WithString("handle", mcp.Description("Prompt handle"), mcp.Required()),
	), handlePromptDelete)

	s.AddTool(mcp.NewTool("prompt_eval",
		mcp.WithDescription("Run a prompt's test cases on the real model and report how many pass. Name several versions (for example \"1,2\") to compare them on the same cases. Uses the model, so it takes a while."),
		mcp.WithString("handle", mcp.Description("Prompt handle"), mcp.Required()),
		mcp.WithString("versions", mcp.Description("Versions to evaluate, e.g. \"1,2\". Default: the active one")),
		mcp.WithString("cases", mcp.Description("Test cases as a JSON array, used instead of the stored ones for this run")),
		mcp.WithString("agent", mcp.Description("Agent to run the prompt with (default: a plain assistant with no tools)")),
		mcp.WithString("model", mcp.Description("Model to use for this run")),
	), handlePromptEval)
}

func handlePromptList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	list, err := automationService.ListPrompts(ctx)
	if err != nil {
		return errorResult(err), nil
	}
	if len(list) == 0 {
		return textResult("No prompts found"), nil
	}
	return jsonResult(map[string]interface{}{"prompts": list}), nil
}

func handlePromptGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	p, err := automationService.GetPrompt(ctx, strings.TrimSpace(getString(args, "handle")), getInt(args, "version"))
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(p), nil
}

func handlePromptHistory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	h, err := automationService.PromptHistory(ctx, strings.TrimSpace(getString(argsMap(request), "handle")))
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(map[string]interface{}{"versions": h}), nil
}

func handlePromptSave(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)

	cases, err := automationService.ParsePromptCases(getString(args, "cases"))
	if err != nil {
		return errorResult(err), nil
	}
	activate, _ := args["activate"].(bool)

	p, err := automationService.SavePromptVersion(ctx, automationService.SavePrompt{
		Handle:      strings.TrimSpace(getString(args, "handle")),
		Text:        getString(args, "text"),
		Description: getString(args, "description"),
		Note:        getString(args, "note"),
		Cases:       cases,
		Activate:    activate,
	})
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(p), nil
}

func handlePromptActivate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)

	version := getInt(args, "version")
	if version < 1 {
		return textResult("version is required"), nil
	}
	p, err := automationService.ActivatePrompt(ctx, strings.TrimSpace(getString(args, "handle")), version)
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(p), nil
}

func handlePromptDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	handle := strings.TrimSpace(getString(argsMap(request), "handle"))
	if err := automationService.DeletePrompt(ctx, handle); err != nil {
		return errorResult(err), nil
	}
	return textResult("Deleted prompt " + handle), nil
}

func handlePromptEval(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)

	reports, err := automationService.EvaluatePrompt(ctx,
		strings.TrimSpace(getString(args, "handle")),
		getString(args, "versions"), getString(args, "agent"), getString(args, "model"), getString(args, "cases"))
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(map[string]interface{}{"reports": reports}), nil
}

// getInt reads a whole number argument. MCP clients send JSON numbers (which
// arrive as float64) but some send them as strings; getString/parseUint64
// only see the latter and would quietly read a numeric version as missing.
func getInt(m map[string]interface{}, key string) int {
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(v))
		return n
	}
	return 0
}
