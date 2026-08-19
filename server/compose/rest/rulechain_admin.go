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

	engine.DeleteChain(chainID)

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

	result, err := engine.Run(r.Context(), chainID, input)
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

func (a RuleChainAdmin) NodeTypes(w http.ResponseWriter, r *http.Request) {
	api.Send(w, r, map[string]interface{}{
		"nodes": nodeTypes(),
	})
}

func nodeTypes() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"type":        "condition",
			"label":       "Condition",
			"description": "Evaluate a condition (eq, neq, gt, lt, contains, empty, notEmpty)",
			"configSchema": map[string]interface{}{
				"field":    "string (required) — variable or field name",
				"operator": "string (required) — eq | neq | gt | lt | gte | lte | contains | empty | notEmpty",
				"value":    "string — value to compare against",
			},
		},
		{
			"type":        "crud",
			"label":       "CRUD Record",
			"description": "Create, update, delete, or search records in a module",
			"configSchema": map[string]interface{}{
				"operation":   "string (required) — create | update | delete | search",
				"moduleID":    "uint64 — module ID",
				"namespaceID": "uint64 — namespace ID",
				"recordID":    "string — record ID (for update/delete)",
				"fields":      "object — field name: value pairs (for create/update)",
				"query":       "string — search query (for search)",
			},
		},
		{
			"type":        "crud.upsert",
			"label":       "Upsert Record",
			"description": "Find a record by matchBy fields and update it, or create it",
			"configSchema": map[string]interface{}{
				"moduleID":     "uint64 — module ID",
				"moduleHandle": "string — module handle (preferred over moduleID)",
				"namespaceID":  "uint64 — namespace ID",
				"matchBy":      "string[] — ordered fields to match (e.g. mac_address, ip_address)",
				"fields":       "object — field templates, {{item.ip}} inside foreach",
			},
		},
		{
			"type":        "foreach",
			"label":       "For Each Item",
			"description": "Loop body nodes once per item in an array (items / devices)",
			"configSchema": map[string]interface{}{
				"items":    "string — variable name holding the array (default: items)",
				"itemVar":  "string — prefix for item fields (default: item)",
				"maxItems": "int — optional cap",
				"failFast": "bool — stop on first body error",
			},
		},
		{
			"type":        "detach",
			"label":       "Detach (poll)",
			"description": "Start a background poller that feeds the ingest chain; does not block Run",
			"configSchema": map[string]interface{}{
				"kind":          "string — poll",
				"ingestChainID": "string (required) — chain to run with the envelope",
				"statusUrl":     "string — GET job status",
				"itemsUrl":      "string — GET items on complete",
				"interval":      "int — seconds (default 2)",
				"timeout":       "int — seconds (default 900)",
				"until":         "string — comma statuses that stop polling",
			},
		},
		{
			"type":        "mail",
			"label":       "Send Email",
			"description": "Send an email notification",
			"configSchema": map[string]interface{}{
				"to":          "string (required) — recipient email",
				"subject":     "string (required) — email subject",
				"body":        "string (required) — email body (HTML supported)",
				"cc":          "string — CC emails",
				"contentType": "string — html | plain",
			},
		},
		{
			"type":        "http",
			"label":       "HTTP Request",
			"description": "Make an HTTP request to an external API",
			"configSchema": map[string]interface{}{
				"url":     "string (required) — URL",
				"method":  "string — GET | POST | PUT | DELETE",
				"headers": "object — header name: value pairs",
				"body":    "string — request body",
				"timeout": "int — timeout in seconds",
			},
		},
		{
			"type":        "ai",
			"label":       "AI Agent",
			"description": "Call an AI agent (crud-agent, assistant) with a prompt",
			"configSchema": map[string]interface{}{
				"agent":  "string (required) — agent name: crud-agent | assistant",
				"prompt": "string (required) — prompt with {{variable}} support",
				"model":  "string — model name (default: qwen3:8b / CHAT_MODEL)",
			},
		},
		{
			"type":        "script",
			"label":       "JavaScript",
			"description": "Execute JavaScript code with lowcode runtime API",
			"configSchema": map[string]interface{}{
				"code": "string (required) — JS code with access to runtime.mcp, runtime.mail, runtime.http, runtime.log",
			},
		},
		{
			"type":        "gonec",
			"label":       "Go Code",
			"description": "Compile and execute Go code in a sandbox",
			"configSchema": map[string]interface{}{
				"code":    "string (required) — Go source code",
				"timeout": "int — execution timeout in seconds",
			},
		},
		{
			"type":        "workflow",
			"label":       "Trigger Workflow",
			"description": "Trigger a Corteza workflow",
			"configSchema": map[string]interface{}{
				"workflowID": "string (required) — workflow ID to trigger",
				"payload":    "string — JSON payload to pass to workflow",
			},
		},
		{
			"type":        "fork",
			"label":       "Fork",
			"description": "Split execution into multiple parallel branches",
			"configSchema": map[string]interface{}{
				"branches": "int — number of branches (min: 2)",
			},
		},
		{
			"type":        "score.matrix",
			"label":       "Risk Matrix",
			"description": "5×5 (or NxN) likelihood × impact → score",
			"configSchema": map[string]interface{}{
				"likelihoodField": "string — input field (default: likelihood)",
				"impactField":     "string — input field (default: impact)",
				"scale":           "int — clamp 1..scale (default: 5)",
				"formula":         "string — product | sum (ignored if matrix set)",
				"matrix":          "number[][] — optional custom cell lookup [L-1][I-1]",
				"outScore":        "string — output variable (default: score)",
			},
		},
		{
			"type":        "score.weighted",
			"label":       "Weighted Score",
			"description": "Σ weightᵢ · normalize(fieldᵢ / maxᵢ) → score 0..100",
			"configSchema": map[string]interface{}{
				"factors":   "array — [{field, weight, max, invert?}]",
				"normalize": "bool — default true (scale to 0..scaleMax)",
				"scaleMax":  "number — default 100",
				"outScore":  "string — output variable (default: score)",
			},
		},
		{
			"type":        "risk.band",
			"label":       "Risk Band",
			"description": "Map score → level; optional residual = score × (1 − control)",
			"configSchema": map[string]interface{}{
				"scoreField":      "string — default score",
				"controlField":    "string — 0..1 control effectiveness",
				"bands":           "array — [{name, max}] ascending max",
				"criticalLevels":  "string[] — levels that set is_critical (default: [critical])",
				"outLevel":        "string — default level",
				"outResidual":     "string — default residualScore",
				"outCriticalFlag": "string — default is_critical (empty unless critical)",
			},
		},
	}
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
	})
}
