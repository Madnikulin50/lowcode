package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/mark3labs/mcp-go/mcp"
)

const defaultMCPRecordLimit uint = 50

func mcpTransportEnabled() bool {
	return os.Getenv("MCP_STDIO") == "true" || os.Getenv("MCP_SSE_ADDR") != ""
}

func mcpScriptsEnabled() bool {
	return os.Getenv("MCP_ENABLE_SCRIPTS") == "true"
}

// withAuth keeps a real caller identity. It never impersonates user 1.
// STDIO (local process, no SSE) may fall back to the service user.
func withAuth(ctx context.Context) context.Context {
	if ident := auth.GetIdentityFromContext(ctx); ident != nil && ident.Valid() {
		return ctx
	}
	if os.Getenv("MCP_STDIO") == "true" && os.Getenv("MCP_SSE_ADDR") == "" {
		if su := auth.ServiceUserOrNil(); su != nil {
			return auth.SetIdentityToContext(ctx, su)
		}
	}
	return ctx
}

// WithAuth is the exported form used by the MCP bridge CRUD helper.
func WithAuth(ctx context.Context) context.Context {
	return withAuth(ctx)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getFieldString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case string:
			return val
		case bool:
			if val {
				return "true"
			}
			return "false"
		}
	}
	return ""
}

func getFieldBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func parseUint64(m map[string]interface{}, key string) uint64 {
	s := getString(m, key)
	if s == "" {
		return 0
	}
	var v uint64
	fmt.Sscanf(s, "%d", &v)
	return v
}

func toJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf(`{"error":"%v"}`, err)
	}
	return string(b)
}

func errorResult(err error) *mcp.CallToolResult {
	return mcp.NewToolResultText("Error: " + err.Error())
}

func textResult(text string) *mcp.CallToolResult {
	return mcp.NewToolResultText(text)
}

func jsonResult(v interface{}) *mcp.CallToolResult {
	return mcp.NewToolResultText(toJSON(v))
}
