package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/mark3labs/mcp-go/mcp"
)

func withAuth(ctx context.Context) context.Context {
	return auth.SetIdentityToContext(ctx, auth.Authenticated(1))
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
