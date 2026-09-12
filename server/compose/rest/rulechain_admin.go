package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/compose/mcp/handlers"
	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/api"
	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
)

type RuleChainAdmin struct{}

func (a RuleChainAdmin) List(w http.ResponseWriter, r *http.Request) {
	engine := handlers.RuleEngine
	if engine == nil {
		api.Send(w, r, fmt.Errorf("engine not initialized"))
		return
	}

	nsFilter := parseUint64String(r.URL.Query().Get("namespaceID"))

	all := engine.Chains()
	chains := make([]*rulesgo.Chain, 0, len(all))
	for _, c := range all {
		if nsFilter > 0 && c.NamespaceID != 0 && c.NamespaceID != nsFilter {
			continue
		}
		chains = append(chains, c)
	}

	limit := queryInt(r, "limit", 50)
	offset := queryInt(r, "offset", 0)

	total := len(chains)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}

	result := make([]map[string]interface{}, 0, end-offset)
	for _, c := range chains[offset:end] {
		result = append(result, map[string]interface{}{
			"id":          c.ID,
			"namespaceID": strconv.FormatUint(c.NamespaceID, 10),
			"name":        c.Name,
			"description": c.Description,
			"nodeCount":   len(c.Nodes),
			"edgeCount":   len(c.Edges),
			"entryNode":   c.EntryNode,
		})
	}

	api.Send(w, r, map[string]interface{}{
		"chains": result,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (a RuleChainAdmin) Get(w http.ResponseWriter, r *http.Request) {
	chainID := chi.URLParam(r, "chainID")
	engine := handlers.RuleEngine
	if engine == nil {
		api.Send(w, r, fmt.Errorf("engine not initialized"))
		return
	}

	chain := engine.Chain(chainID)
	if chain == nil {
		api.Send(w, r, fmt.Errorf("chain not found"))
		return
	}

	api.Send(w, r, map[string]interface{}{
		"chain": chain,
	})
}

func (a RuleChainAdmin) Create(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)

	payload, err := decodeChainPayload(body)
	if err != nil {
		api.Send(w, r, fmt.Errorf("invalid JSON: %w", err))
		return
	}

	if payload.Name == "" {
		api.Send(w, r, fmt.Errorf("name is required"))
		return
	}

	engine := handlers.RuleEngine
	if engine == nil {
		api.Send(w, r, fmt.Errorf("engine not initialized"))
		return
	}

	if payload.ID == "" {
		payload.ID = fmt.Sprintf("rc_%s_%s", payload.Name, safeID(4))
	}

	if engine.Chain(payload.ID) != nil {
		api.Send(w, r, fmt.Errorf("chain %s already exists", payload.ID))
		return
	}

	// Register even if PostgreSQL persist fails so the chain is usable in this process.
	engine.RegisterChain(payload)
	if err := engine.PersistChain(payload); err != nil {
		log.Printf("[rulechain] persist %s: %v", payload.ID, err)
	}

	api.Send(w, r, map[string]interface{}{
		"created": true,
		"chainID": payload.ID,
		"chain":   payload,
	})
}

func (a RuleChainAdmin) Update(w http.ResponseWriter, r *http.Request) {
	chainID := chi.URLParam(r, "chainID")
	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)

	payload, err := decodeChainPayload(body)
	if err != nil {
		api.Send(w, r, fmt.Errorf("invalid JSON: %w", err))
		return
	}

	engine := handlers.RuleEngine
	if engine == nil {
		api.Send(w, r, fmt.Errorf("engine not initialized"))
		return
	}

	existing := engine.Chain(chainID)
	if existing == nil {
		api.Send(w, r, fmt.Errorf("chain not found"))
		return
	}

	existing.ID = chainID
	if payload.Name != "" {
		existing.Name = payload.Name
	}
	existing.Description = payload.Description
	if payload.EntryNode != "" {
		existing.EntryNode = payload.EntryNode
	}
	if payload.NamespaceID > 0 {
		existing.NamespaceID = payload.NamespaceID
	}
	if payload.Nodes != nil {
		existing.Nodes = payload.Nodes
	}
	if payload.Edges != nil {
		existing.Edges = payload.Edges
	}
	if len(payload.Config) > 0 {
		existing.Config = payload.Config
	}

	if err := engine.PersistChain(existing); err != nil {
		api.Send(w, r, fmt.Errorf("persist failed: %w", err))
		return
	}

	api.Send(w, r, map[string]interface{}{
		"updated": true,
		"chainID": chainID,
		"chain":   existing,
	})
}

func (a RuleChainAdmin) Delete(w http.ResponseWriter, r *http.Request) {
	chainID := chi.URLParam(r, "chainID")
	engine := handlers.RuleEngine
	if engine == nil {
		api.Send(w, r, fmt.Errorf("engine not initialized"))
		return
	}

	chain := engine.Chain(chainID)
	if chain == nil {
		api.Send(w, r, fmt.Errorf("chain not found"))
		return
	}

	if err := engine.DeleteChain(r.Context(), chainID); err != nil {
		api.Send(w, r, fmt.Errorf("delete failed: %w", err))
		return
	}

	api.Send(w, r, map[string]interface{}{
		"deleted": true,
		"chainID": chainID,
	})
}

func (a RuleChainAdmin) Test(w http.ResponseWriter, r *http.Request) {
	chainID := chi.URLParam(r, "chainID")
	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)

	var input map[string]interface{}
	if len(body) > 0 {
		json.Unmarshal(body, &input)
	}

	engine := handlers.RuleEngine
	if engine == nil {
		api.Send(w, r, fmt.Errorf("engine not initialized"))
		return
	}

	result, err := engine.RunWithLog(r.Context(), chainID, input, "manual-test")
	if err != nil {
		api.Send(w, r, err)
		return
	}

	api.Send(w, r, result)
}

func (a RuleChainAdmin) Stats(w http.ResponseWriter, r *http.Request) {
	engine := handlers.RuleEngine
	if engine == nil {
		api.Send(w, r, fmt.Errorf("engine not initialized"))
		return
	}

	chains := engine.Chains()
	nodeStats := make(map[string]int)
	totalNodes := 0

	for _, c := range chains {
		totalNodes += len(c.Nodes)
		for _, n := range c.Nodes {
			nodeStats[n.Type]++
		}
	}

	api.Send(w, r, map[string]interface{}{
		"totalChains": len(chains),
		"totalNodes":  totalNodes,
		"nodeTypes":   nodeStats,
		"registry":    nodeTypes(),
	})
}

// ListRuns returns persisted run-log entries (compose_rule_chain_run) for one
// chain, newest first. Only summary fields - the full per-node trace is
// fetched per-run via GetRun to keep the list payload light.
func (a RuleChainAdmin) ListRuns(w http.ResponseWriter, r *http.Request) {
	chainID := chi.URLParam(r, "chainID")

	f := types.RuleChainRunFilter{
		ChainID:     chainID,
		NamespaceID: parseUint64String(r.URL.Query().Get("namespaceID")),
		TriggerType: strings.TrimSpace(r.URL.Query().Get("triggerType")),
		Limit:       queryInt(r, "limit", 50),
		Offset:      queryInt(r, "offset", 0),
	}
	if v := strings.TrimSpace(r.URL.Query().Get("success")); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			f.Success = &b
		}
	}

	runs, _, err := service.SearchRuleChainRuns(r.Context(), f)
	if err != nil {
		api.Send(w, r, err)
		return
	}

	out := make([]map[string]interface{}, 0, len(runs))
	for _, run := range runs {
		out = append(out, map[string]interface{}{
			"runID":       strconv.FormatUint(run.ID, 10),
			"namespaceID": strconv.FormatUint(run.NamespaceID, 10),
			"chainID":     run.ChainID,
			"triggerType": run.TriggerType,
			"success":     run.Success,
			"error":       run.Error,
			"startedAt":   run.StartedAt,
			"finishedAt":  run.FinishedAt,
			"durationMs":  run.DurationMs,
		})
	}

	api.Send(w, r, map[string]interface{}{
		"runs":   out,
		"total":  len(out),
		"limit":  f.Limit,
		"offset": f.Offset,
	})
}

// GetRun returns one persisted run-log entry, including its full per-node
// trace, input and output.
func (a RuleChainAdmin) GetRun(w http.ResponseWriter, r *http.Request) {
	chainID := chi.URLParam(r, "chainID")
	runID := parseUint64String(chi.URLParam(r, "runID"))
	if runID == 0 {
		api.Send(w, r, fmt.Errorf("invalid runID"))
		return
	}

	run, err := service.LookupRuleChainRun(r.Context(), runID)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	if chainID != "" && run.ChainID != chainID {
		api.Send(w, r, fmt.Errorf("run not found for chain %s", chainID))
		return
	}

	api.Send(w, r, map[string]interface{}{
		"run": run,
	})
}

func (a RuleChainAdmin) NodeTypes(w http.ResponseWriter, r *http.Request) {
	api.Send(w, r, map[string]interface{}{
		"nodes": nodeTypes(),
	})
}

func queryInt(r *http.Request, key string, defaultVal int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}

func safeID(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[i*7%len(chars)]
	}
	return string(b)
}

type chainPayloadRaw struct {
	ID          string          `json:"id"`
	NamespaceID json.RawMessage `json:"namespaceID"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	EntryNode   string          `json:"entryNode"`
	Nodes       json.RawMessage `json:"nodes"`
	Edges       json.RawMessage `json:"edges"`
	Config      json.RawMessage `json:"config"`
}

func decodeChainPayload(body []byte) (*rulesgo.Chain, error) {
	var raw chainPayloadRaw
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	c := &rulesgo.Chain{
		ID:          strings.TrimSpace(raw.ID),
		NamespaceID: parseFlexibleUint64(raw.NamespaceID),
		Name:        raw.Name,
		Description: raw.Description,
		EntryNode:   raw.EntryNode,
		Config:      raw.Config,
	}
	if len(raw.Nodes) > 0 && string(raw.Nodes) != "null" {
		if err := json.Unmarshal(raw.Nodes, &c.Nodes); err != nil {
			return nil, fmt.Errorf("invalid nodes: %w", err)
		}
	}
	if len(raw.Edges) > 0 && string(raw.Edges) != "null" {
		if err := json.Unmarshal(raw.Edges, &c.Edges); err != nil {
			return nil, fmt.Errorf("invalid edges: %w", err)
		}
	}
	return c, nil
}

func parseUint64String(s string) uint64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "null" {
		return 0
	}
	n, _ := strconv.ParseUint(s, 10, 64)
	return n
}

func parseFlexibleUint64(raw json.RawMessage) uint64 {
	if len(raw) == 0 || string(raw) == "null" {
		return 0
	}
	var n uint64
	if err := json.Unmarshal(raw, &n); err == nil {
		return n
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return parseUint64String(s)
	}
	return 0
}

func MountRuleChainAdminRoutes(r chi.Router) {
	admin := RuleChainAdmin{}
	r.Route("/admin/rulechain", func(r chi.Router) {
		r.Get("/", admin.List)
		r.Post("/", admin.Create)
		r.Get("/nodes", admin.NodeTypes)
		r.Get("/stats", admin.Stats)
		r.Get("/{chainID}", admin.Get)
		r.Put("/{chainID}", admin.Update)
		r.Delete("/{chainID}", admin.Delete)
		r.Post("/{chainID}/test", admin.Test)
		r.Get("/{chainID}/runs", admin.ListRuns)
		r.Get("/{chainID}/runs/{runID}", admin.GetRun)
	})
}
