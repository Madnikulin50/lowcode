package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/compose/mcp/handlers"
	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/pkg/api"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
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

	if tok := persistedAuthToken(r); tok != "" {
		bag["authToken"] = tok
	}
	injectAgentCallback(r, bag)

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
		"output":  mergeNodeBodies(result.Output, result.Nodes),
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
		if tok := persistedAuthToken(r); tok != "" {
			bag["authToken"] = tok
		}
		injectAgentCallback(r, bag)
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
	h := r.Header.Get("Authorization")
	if len(h) > 7 && strings.EqualFold(h[:7], "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}

// persistedAuthToken mints a stored access token for background HTTP nodes.
// The browser JWT is rotated on refresh and then fails TokenIssuer.Validate
// (jti missing from auth_oa2tokens) when the CMDB agent calls the API back.
func persistedAuthToken(r *http.Request) string {
	fallback := bearerToken(r)
	ident := auth.GetIdentityFromContext(r.Context())
	if ident == nil || !ident.Valid() || auth.TokenIssuer == nil {
		return fallback
	}
	tok, err := auth.TokenIssuer.Issue(r.Context(),
		auth.WithIdentity(ident),
		auth.WithScope("api", "profile"),
		auth.WithAudience("rule-chain"),
		auth.WithExpiration(4*time.Hour),
	)
	if err != nil || len(tok) == 0 {
		return fallback
	}
	return string(tok)
}

func injectAgentCallback(_ *http.Request, bag map[string]interface{}) {
	if bag == nil {
		return
	}
	agentURL := strings.TrimSpace(fmt.Sprintf("%v", bag["agentUrl"]))
	if agentURL == "" || agentURL == "<nil>" {
		agentURL = strings.TrimRight(os.Getenv("CMDB_AGENT_URL"), "/")
	}
	if agentURL == "" {
		agentURL = "http://localhost:8085/api"
	}
	bag["agentUrl"] = strings.TrimRight(agentURL, "/")

	ingestID := strings.TrimSpace(fmt.Sprintf("%v", bag["ingestChainID"]))
	if ingestID == "" || ingestID == "<nil>" {
		ingestID = "cmdb-ingest-scan"
	}
	bag["ingestChainID"] = ingestID

	if cb := strings.TrimSpace(fmt.Sprintf("%v", bag["callbackUrl"])); cb != "" && cb != "<nil>" {
		return
	}
	bag["callbackUrl"] = composeAPIRoot() + "/compose/rulechain/" + ingestID + "/run"
}

// composeAPIRoot is the origin (+ HTTP_API_BASE_URL) the Go server actually mounts on.
// Vite uses window.CortezaAPI='/api' and proxies it; the server default is HTTP_API_BASE_URL=/
// so http://localhost:3333/api/compose/... is chi's HTML 404, not the rule-chain runner.
func composeAPIRoot() string {
	apiBase := strings.TrimSpace(os.Getenv("HTTP_API_BASE_URL"))
	if apiBase == "" {
		apiBase = "/"
	}
	if explicit := strings.TrimRight(strings.TrimSpace(os.Getenv("CORTEZA_API")), "/"); explicit != "" && explicit != "<nil>" {
		if apiBase == "/" && strings.HasSuffix(explicit, "/api") {
			explicit = strings.TrimSuffix(explicit, "/api")
		}
		return explicit
	}
	addr := strings.TrimSpace(os.Getenv("HTTP_ADDR"))
	if addr == "" || addr == ":80" {
		addr = ":3333"
	}
	origin := addr
	if strings.HasPrefix(addr, ":") {
		origin = "http://localhost" + addr
	} else if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		origin = "http://" + addr
	}
	origin = strings.TrimRight(origin, "/")
	prefix := "/" + strings.Trim(path.Join(strings.TrimSpace(os.Getenv("HTTP_BASE_URL")), apiBase), "/")
	if prefix == "/" {
		return origin
	}
	return origin + prefix
}

func mergeNodeBodies(out map[string]interface{}, nodes interface{}) map[string]interface{} {
	if out == nil {
		out = map[string]interface{}{}
	}
	raw, err := json.Marshal(nodes)
	if err != nil {
		return out
	}
	var list []map[string]interface{}
	if json.Unmarshal(raw, &list) != nil {
		return out
	}
	for i := len(list) - 1; i >= 0; i-- {
		output, _ := list[i]["output"].(map[string]interface{})
		if output == nil {
			continue
		}
		body := output["body"]
		if body == nil {
			continue
		}
		out["result"] = body
		m, ok := body.(map[string]interface{})
		if !ok {
			break
		}
		for k, v := range m {
			if _, exists := out[k]; !exists {
				out[k] = v
			}
		}
		break
	}
	return out
}

func flattenTriggerContext(ctx map[string]interface{}, req *triggerRequest) {
	if req.RecordID == "" && req.Record != nil {
		if id := bagNonPlaceholder(map[string]interface{}{"recordID": req.Record["recordID"]}, "recordID"); id != "" {
			req.RecordID = id
		}
	}
	if req.RecordID != "" {
		ctx["recordID"] = req.RecordID
		if bagNonPlaceholder(ctx, "documentID") == "" {
			ctx["documentID"] = req.RecordID
		}
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
	aliasTriggerRecordIDs(ctx)
}

// aliasTriggerRecordIDs fills recordID from page-block context (sourceID/policyID/projectID/…)
// and the other way around, so chain templates {{recordID}} and {{projectID}} both work.
// Uninterpolated "${recordID}" / "{{recordID}}" leftovers are treated as empty.
func aliasTriggerRecordIDs(ctx map[string]interface{}) {
	if ctx == nil {
		return
	}
	recID := bagNonPlaceholder(ctx, "recordID")
	if recID == "" {
		if v := bagNonPlaceholder(ctx, "documentID"); v != "" {
			recID = v
		}
	}
	if recID == "" {
		// Do not steal projectID when this record *points at* a project
		// (document, WBS, RFC). That sent a project id to submit-approval
		// → Compose "API error 200: not found" in the documents module.
		keys := []string{"sourceID", "policyID", "snapshotID"}
		if bagNonPlaceholder(ctx, "project") == "" {
			keys = append(keys, "projectID")
		}
		for _, k := range keys {
			if v := bagNonPlaceholder(ctx, k); v != "" {
				recID = v
				break
			}
		}
	}
	if recID == "" {
		return
	}
	ctx["recordID"] = recID
	if bagNonPlaceholder(ctx, "projectID") == "" {
		ctx["projectID"] = recID
	}
	if bagNonPlaceholder(ctx, "sourceID") == "" && bagNonPlaceholder(ctx, "policyID") == "" && bagNonPlaceholder(ctx, "snapshotID") == "" {
		ctx["sourceID"] = recID
	}
	if bagNonPlaceholder(ctx, "policyID") == "" {
		if p := bagNonPlaceholder(ctx, "policy"); p != "" {
			ctx["policyID"] = p
		}
	}
	if bagNonPlaceholder(ctx, "snapshotID") == "" {
		if s := bagNonPlaceholder(ctx, "snapshot"); s != "" {
			ctx["snapshotID"] = s
		}
	}
}

func bagNonPlaceholder(ctx map[string]interface{}, key string) string {
	if ctx == nil {
		return ""
	}
	v, ok := ctx[key]
	if !ok || v == nil {
		return ""
	}
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	if s == "" || s == "<nil>" {
		return ""
	}
	if strings.Contains(s, "${") || strings.Contains(s, "{{") {
		return ""
	}
	if s == "0" {
		return ""
	}
	return s
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
