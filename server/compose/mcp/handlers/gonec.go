package handlers

import (
	"context"
	"encoding/json"
	"os"

	"github.com/madnikulin50/lowcode/server/pkg/gonec"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var GonecEngine *gonec.Engine

func SetGonecEngine(e *gonec.Engine) {
	GonecEngine = e
}

func initGonec(ctx context.Context, s *server.MCPServer) {
	if mcpScriptsEnabled() {
		s.AddTool(mcp.NewTool("gonec_run",
			mcp.WithDescription("Compile and run a Go program in a sandbox. Use for AI-generated Go code execution. The code runs as a standalone Go program with access to standard library and lowcode SDK."),
			mcp.WithString("code", mcp.Description("Go source code to run. Must be a valid Go program. Can use standard library packages and the lowcode SDK."), mcp.Required()),
			mcp.WithString("timeout", mcp.Description("Execution timeout in seconds (default: 30)")),
		), handleGonecRun)
	}

	s.AddTool(mcp.NewTool("gonec_validate",
		mcp.WithDescription("Check if Go code is syntactically valid and compiles. Returns error details if compilation fails."),
		mcp.WithString("code", mcp.Description("Go source code to validate"), mcp.Required()),
	), handleGonecValidate)

	s.AddTool(mcp.NewTool("gonec_format",
		mcp.WithDescription("Format Go code using gofmt. Returns the formatted version."),
		mcp.WithString("code", mcp.Description("Go source code to format"), mcp.Required()),
	), handleGonecFormat)
}

func handleGonecRun(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	code := getString(args, "code")

	if code == "" {
		return textResult("Error: 'code' is required"), nil
	}

	engine := GonecEngine
	if engine == nil {
		engine = gonec.New(os.TempDir())
	}

	result := engine.Run(ctx, gonec.Sanitize(code))
	return jsonResult(result), nil
}

func handleGonecValidate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	code := getString(args, "code")

	if code == "" {
		return textResult("Error: 'code' is required"), nil
	}

	engine := GonecEngine
	if engine == nil {
		engine = gonec.New(os.TempDir())
	}

	err := engine.Validate(ctx, gonec.Sanitize(code))
	if err != nil {
		return jsonResult(map[string]interface{}{
			"valid": false,
			"error": err.Error(),
		}), nil
	}

	formatted := engine.Format(ctx, code)
	return jsonResult(map[string]interface{}{
		"valid":     true,
		"formatted": formatted,
	}), nil
}

func handleGonecFormat(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	code := getString(args, "code")

	if code == "" {
		return textResult("Error: 'code' is required"), nil
	}

	engine := GonecEngine
	if engine == nil {
		engine = gonec.New(os.TempDir())
	}

	formatted := engine.Format(ctx, code)
	return jsonResult(map[string]interface{}{
		"code": formatted,
	}), nil
}

func handleGonecRunWithContext(ctx context.Context, code string) (string, error) {
	engine := GonecEngine
	if engine == nil {
		engine = gonec.New(os.TempDir())
	}
	result := engine.Run(ctx, gonec.Sanitize(code))
	data, _ := json.Marshal(result)
	return string(data), nil
}
