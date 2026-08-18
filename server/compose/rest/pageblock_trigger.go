package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/compose/mcp/handlers"
	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/pkg/api"
)

const (
	stockReorderBatchChainID = "stock_reorder_batch"
	storeRiskChainID         = "demo_store_risk"
)

type PageBlockTrigger struct{}

// jsonID accepts Corteza string IDs ("123") and numeric JSON (123).
type jsonID uint64

func (id *jsonID) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "" || s == "null" || s == `""` {
		*id = 0
		return nil
	}
	if len(s) > 0 && s[0] == '"' {
		var str string
		if err := json.Unmarshal(b, &str); err != nil {
			return err
		}
		str = strings.TrimSpace(str)
		if str == "" {
			*id = 0
			return nil
		}
		n, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return err
		}
		*id = jsonID(n)
		return nil
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid id %s", s)
	}
	*id = jsonID(n)
	return nil
}

type triggerRequest struct {
	ChainID     string                 `json:"chainID"`
	PageID      jsonID                 `json:"pageID"`
	BlockID     string                 `json:"blockID"`
	RecordID    string                 `json:"recordID"`
	Record      map[string]interface{} `json:"record"`
	ModuleID    jsonID                 `json:"moduleID"`
	NamespaceID jsonID                 `json:"namespaceID"`
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
		summary, err := service.RunStockReorder(r.Context(), uint64(req.NamespaceID))
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
	if engine.Chain(req.ChainID) == nil && handlers.OnChainMissing != nil {
		handlers.OnChainMissing(r.Context(), req.ChainID)
	}

	bag := make(map[string]interface{})
	if req.Context != nil {
		for k, v := range req.Context {
			bag[k] = v
		}
	}

	flattenTriggerContext(bag, &req)

	if tok := bearerToken(r); tok != "" {
		bag["authToken"] = tok
	}

	var riskIn *service.StoreRiskInput
	if req.ChainID == storeRiskChainID {
		riskIn, err = enrichStoreRiskContext(r.Context(), bag)
		if err != nil {
			api.Send(w, r, map[string]interface{}{
				"success": false,
				"chainID": req.ChainID,
				"blockID": req.BlockID,
				"error":   err.Error(),
			})
			return
		}
	}

	result, err := engine.Run(r.Context(), req.ChainID, bag)
	if err != nil {
		api.Send(w, r, fmt.Errorf("chain execution failed: %w", err))
		return
	}

	if riskIn != nil && result != nil && result.Success {
		_ = service.PersistStoreRiskSlice(r.Context(), riskIn, residualScore(result.Output))
		if result.Output != nil {
			for k, v := range service.StoreRiskInputMap(riskIn) {
				if _, exists := result.Output[k]; !exists {
					result.Output[k] = v
				}
			}
		}
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
			summary, err := service.RunStockReorder(r.Context(), uint64(req.NamespaceID))
			if err != nil {
				resultMap["error"] = err.Error()
			} else {
				resultMap["success"] = true
				resultMap["output"] = summary
			}
			results = append(results, resultMap)
			continue
		}
		if engine.Chain(req.ChainID) == nil && handlers.OnChainMissing != nil {
			handlers.OnChainMissing(r.Context(), req.ChainID)
		}
		bag := buildContext(&req)
		if tok := bearerToken(r); tok != "" {
			bag["authToken"] = tok
		}
		var riskIn *service.StoreRiskInput
		if req.ChainID == storeRiskChainID {
			riskIn, err = enrichStoreRiskContext(r.Context(), bag)
			if err != nil {
				resultMap["error"] = err.Error()
				results = append(results, resultMap)
				continue
			}
		}
		result, err := engine.Run(r.Context(), req.ChainID, bag)
		if err != nil {
			resultMap["error"] = err.Error()
		} else {
			resultMap["success"] = result.Success
			resultMap["output"] = result.Output
			if riskIn != nil && result.Success {
				_ = service.PersistStoreRiskSlice(r.Context(), riskIn, residualScore(result.Output))
			}
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
	flattenTriggerContext(ctx, req)
	return ctx
}

func bearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if len(auth) > 7 && strings.EqualFold(auth[:7], "bearer ") {
		return strings.TrimSpace(auth[7:])
	}
	return ""
}

func flattenTriggerContext(ctx map[string]interface{}, req *triggerRequest) {
	if req.RecordID != "" {
		ctx["recordID"] = req.RecordID
	}
	if req.Record != nil {
		for k, v := range req.Record {
			if k == "values" {
				flattenValues(ctx, v)
				continue
			}
			if _, exists := ctx[k]; !exists {
				ctx[k] = v
			}
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
}

func flattenValues(ctx map[string]interface{}, raw interface{}) {
	switch vals := raw.(type) {
	case map[string]interface{}:
		for k, v := range vals {
			if k == "toJSON" {
				continue
			}
			ctx[k] = unwrapRecordValue(v)
		}
	case []interface{}:
		for _, item := range vals {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			name, _ := m["name"].(string)
			if name == "" {
				continue
			}
			ctx[name] = unwrapRecordValue(m["value"])
		}
	}
}

func unwrapRecordValue(v interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		if val, has := t["value"]; has && len(t) <= 4 {
			return val
		}
		return t
	case []interface{}:
		if len(t) == 1 {
			return unwrapRecordValue(t[0])
		}
		return t
	default:
		return v
	}
}

func enrichStoreRiskContext(ctx context.Context, bag map[string]interface{}) (*service.StoreRiskInput, error) {
	storeID := service.ContextStoreID(bag)
	if storeID == "" {
		storeID = service.ContextStoreID(map[string]interface{}{
			"store_id": bag["recordID"],
		})
	}
	if storeID == "" {
		return nil, fmt.Errorf("store_id is required")
	}
	in, err := service.GatherStoreRiskInput(ctx, storeID)
	if err != nil {
		return nil, err
	}
	for k, v := range service.StoreRiskInputMap(in) {
		bag[k] = v
	}
	return in, nil
}

func residualScore(output map[string]interface{}) float64 {
	if output == nil {
		return 0
	}
	switch v := output["residualScore"].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case json.Number:
		f, _ := v.Float64()
		return f
	default:
		return 0
	}
}

func MountPageBlockTriggerRoutes(r chi.Router) {
	t := PageBlockTrigger{}
	r.Post("/pageblock/trigger", t.Run)
	r.Post("/pageblock/trigger/batch", t.Batch)
}
