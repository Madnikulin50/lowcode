package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/compose/mcp/handlers"
	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/pkg/api"
)

const stockReorderBatchChainID = "stock_reorder_batch"

type PageBlockTrigger struct{}

type triggerRequest struct {
	ChainID     string                 `json:"chainID"`
	PageID      uint64                 `json:"pageID"`
	BlockID     string                 `json:"blockID"`
	RecordID    string                 `json:"recordID"`
	Record      map[string]interface{} `json:"record"`
	ModuleID    uint64                 `json:"moduleID"`
	NamespaceID uint64                 `json:"namespaceID"`
	UserID      string                 `json:"userID"`
	Context     map[string]interface{} `json:"context"`
}

func (t PageBlockTrigger) Run(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		api.Send(w, r, err)
		return
	}

	var req triggerRequest
	if err := json.Unmarshal(body, &req); err != nil {
		api.Send(w, r, fmt.Errorf("invalid JSON: %w", err))
		return
	}

	if req.ChainID == "" {
		api.Send(w, r, fmt.Errorf("chainID is required"))
		return
	}

	if req.ChainID == stockReorderBatchChainID {
		summary, err := service.RunStockReorder(r.Context(), req.NamespaceID)
		if err != nil {
			api.Send(w, r, map[string]interface{}{
				"success": false,
				"chainID": req.ChainID,
				"blockID": req.BlockID,
				"error":   err.Error(),
			})
			return
		}
		api.Send(w, r, map[string]interface{}{
			"success": true,
			"chainID": req.ChainID,
			"blockID": req.BlockID,
			"output":  summary,
		})
		return
	}

	engine := handlers.RuleEngine
	if engine == nil {
		api.Send(w, r, fmt.Errorf("rule engine not initialized"))
		return
	}

	context := make(map[string]interface{})
	if req.Context != nil {
		for k, v := range req.Context {
			context[k] = v
		}
	}

	if req.RecordID != "" {
		context["recordID"] = req.RecordID
	}
	if req.Record != nil {
		for k, v := range req.Record {
			context[k] = v
		}
	}
	if req.PageID > 0 {
		context["pageID"] = fmt.Sprintf("%d", req.PageID)
	}
	if req.ModuleID > 0 {
		context["moduleID"] = fmt.Sprintf("%d", req.ModuleID)
	}
	if req.NamespaceID > 0 {
		context["namespaceID"] = fmt.Sprintf("%d", req.NamespaceID)
	}
	if req.UserID != "" {
		context["userID"] = req.UserID
	}

	result, err := engine.Run(r.Context(), req.ChainID, context)
	if err != nil {
		api.Send(w, r, fmt.Errorf("chain execution failed: %w", err))
		return
	}

	api.Send(w, r, map[string]interface{}{
		"success": result.Success,
		"chainID": req.ChainID,
		"blockID": req.BlockID,
		"output":  result.Output,
		"nodes":   result.Nodes,
		"error":   result.Error,
	})
}

func (t PageBlockTrigger) Batch(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		api.Send(w, r, err)
		return
	}

	var triggers []triggerRequest
	if err := json.Unmarshal(body, &triggers); err != nil {
		api.Send(w, r, fmt.Errorf("invalid JSON: %w", err))
		return
	}

	engine := handlers.RuleEngine
	if engine == nil {
		api.Send(w, r, fmt.Errorf("rule engine not initialized"))
		return
	}

	results := make([]map[string]interface{}, 0, len(triggers))
	for _, req := range triggers {
		resultMap := map[string]interface{}{
			"chainID": req.ChainID,
			"blockID": req.BlockID,
		}
		if req.ChainID == stockReorderBatchChainID {
			summary, err := service.RunStockReorder(r.Context(), req.NamespaceID)
			if err != nil {
				resultMap["error"] = err.Error()
			} else {
				resultMap["success"] = true
				resultMap["output"] = summary
			}
			results = append(results, resultMap)
			continue
		}
		context := buildContext(&req)
		result, err := engine.Run(r.Context(), req.ChainID, context)
		if err != nil {
			resultMap["error"] = err.Error()
		} else {
			resultMap["success"] = result.Success
			resultMap["output"] = result.Output
		}
		results = append(results, resultMap)
	}

	api.Send(w, r, map[string]interface{}{
		"results": results,
		"total":   len(results),
	})
}

func buildContext(req *triggerRequest) map[string]interface{} {
	ctx := make(map[string]interface{})
	if req.Context != nil {
		for k, v := range req.Context {
			ctx[k] = v
		}
	}
	if req.RecordID != "" {
		ctx["recordID"] = req.RecordID
	}
	if req.Record != nil {
		for k, v := range req.Record {
			ctx[k] = v
		}
	}
	if req.PageID > 0 {
		ctx["pageID"] = fmt.Sprintf("%d", req.PageID)
	}
	if req.ModuleID > 0 {
		ctx["moduleID"] = fmt.Sprintf("%d", req.ModuleID)
	}
	if req.NamespaceID > 0 {
		ctx["namespaceID"] = fmt.Sprintf("%d", req.NamespaceID)
	}
	if req.UserID != "" {
		ctx["userID"] = req.UserID
	}
	return ctx
}

func MountPageBlockTriggerRoutes(r chi.Router) {
	t := PageBlockTrigger{}
	r.Post("/pageblock/trigger", t.Run)
	r.Post("/pageblock/trigger/batch", t.Batch)
}
