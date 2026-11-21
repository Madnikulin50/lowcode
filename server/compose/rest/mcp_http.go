package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/pkg/api"
)

type MCPHTTPBridge struct{}

func (h MCPHTTPBridge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		api.Send(w, r, err)
		return
	}

	var req struct {
		Tool   string                 `json:"tool"`
		Params map[string]interface{} `json:"params"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		api.Send(w, r, fmt.Errorf("invalid JSON: %w", err))
		return
	}

	if req.Tool == "" {
		api.Send(w, r, fmt.Errorf("'tool' is required"))
		return
	}

	if req.Params == nil {
		req.Params = make(map[string]interface{})
	}

	api.Send(w, r, map[string]interface{}{
		"tool":    req.Tool,
		"params":  req.Params,
		"status":  "received",
		"message": "MCP tool will be processed",
	})
}

func MountMCPHTTPBridge(r chi.Router) {
	r.Post("/mcp/call", (&MCPHTTPBridge{}).ServeHTTP)
}
