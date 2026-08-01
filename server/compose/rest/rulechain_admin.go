package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

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

	chains := engine.Chains()
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

	var chain struct {
		ID          string            `json:"id"`
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Nodes       []json.RawMessage `json:"nodes"`
		Edges       []json.RawMessage `json:"edges"`
		EntryNode   string            `json:"entryNode"`
	}

	if err := json.Unmarshal(body, &chain); err != nil {
		api.Send(w, r, fmt.Errorf("invalid JSON: %w", err))
		return
	}

	if chain.Name == "" {
		api.Send(w, r, fmt.Errorf("name is required"))
		return
	}

	engine := handlers.RuleEngine
	if engine == nil {
		api.Send(w, r, fmt.Errorf("engine not initialized"))
		return
	}

	// Generate ID if empty
	if chain.ID == "" {
		chain.ID = fmt.Sprintf("rc_%s_%s", chain.Name, safeID(4))
	}

	// Import the chain
	chainJSON, _ := json.Marshal(chain)
	imported, err := engine.ImportChain(chainJSON)
	if err != nil {
		api.Send(w, r, fmt.Errorf("import failed: %w", err))
		return
	}

	api.Send(w, r, map[string]interface{}{
		"created": true,
		"chainID": imported.ID,
		"chain":   imported,
	})
}

func (a RuleChainAdmin) Update(w http.ResponseWriter, r *http.Request) {
	chainID := chi.URLParam(r, "chainID")
	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)

	var update struct {
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Nodes       []json.RawMessage `json:"nodes"`
		Edges       []json.RawMessage `json:"edges"`
		EntryNode   string            `json:"entryNode"`
	}

	if err := json.Unmarshal(body, &update); err != nil {
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

	if update.Name != "" {
		existing.Name = update.Name
	}
	if update.Description != "" {
		existing.Description = update.Description
	}
	if update.EntryNode != "" {
		existing.EntryNode = update.EntryNode
	}
	if update.Nodes != nil {
		nodesJSON, _ := json.Marshal(update.Nodes)
		var nodes []rulesgo.ChainNode
		if err := json.Unmarshal(nodesJSON, &nodes); err != nil {
			api.Send(w, r, fmt.Errorf("invalid nodes: %w", err))
			return
		}
		existing.Nodes = nodes
	}
	if update.Edges != nil {
		edgesJSON, _ := json.Marshal(update.Edges)
		var edges []rulesgo.ChainEdge
		if err := json.Unmarshal(edgesJSON, &edges); err != nil {
			api.Send(w, r, fmt.Errorf("invalid edges: %w", err))
			return
		}
		existing.Edges = edges
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
				"model":  "string — model name (default: deepseek-v2)",
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
