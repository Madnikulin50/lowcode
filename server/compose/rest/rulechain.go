package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/compose/mcp/handlers"
	"github.com/madnikulin50/lowcode/server/pkg/api"
	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
)

type RuleChain struct {
	handler func() http.Handler
}

func (rc RuleChain) New() *RuleChain {
	return &rc
}

func (rc *RuleChain) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		rc.Run(w, r)
	case http.MethodGet:
		rc.List(w, r)
	default:
		api.Send(w, r, fmt.Errorf("method not allowed: %s", r.Method))
	}
}

func (rc *RuleChain) Run(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	chainID := chi.URLParam(r, "chainID")
	if chainID == "" {
		api.Send(w, r, fmt.Errorf("chainID is required"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		api.Send(w, r, err)
		return
	}

	var input map[string]interface{}
	if len(body) > 0 {
		dec := json.NewDecoder(bytes.NewReader(body))
		dec.UseNumber()
		if err := dec.Decode(&input); err != nil {
			api.Send(w, r, fmt.Errorf("invalid JSON body: %w", err))
			return
		}
	}
	if input == nil {
		input = map[string]interface{}{}
	}
	input = rulesgo.NormalizeIngestEnvelope(input)

	engine := handlers.RuleEngine
	if engine == nil {
		api.Send(w, r, fmt.Errorf("rule engine not initialized"))
		return
	}
	if engine.Chain(chainID) == nil && handlers.OnChainMissing != nil {
		handlers.OnChainMissing(r.Context(), chainID)
	}

	runCtx := r.Context()
	kind, _ := input["kind"].(string)
	if kind == "complete" || kind == "failed" {
		var cancel context.CancelFunc
		runCtx, cancel = context.WithTimeout(context.WithoutCancel(r.Context()), 2*time.Minute)
		defer cancel()
	}

	result, err := engine.Run(runCtx, chainID, input)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	if result != nil && result.Success {
		rulesgo.CancelPollIfTerminal(input)
	}

	api.Send(w, r, result)
}

func (rc *RuleChain) Executions(w http.ResponseWriter, r *http.Request) {
	api.Send(w, r, map[string]interface{}{
		"executions": []interface{}{},
		"message":    "execution log available via persistence engine",
	})
}

func (rc *RuleChain) Export(w http.ResponseWriter, r *http.Request) {
	chainID := chi.URLParam(r, "chainID")
	engine := handlers.RuleEngine
	if engine == nil {
		api.Send(w, r, fmt.Errorf("rule engine not initialized"))
		return
	}
	chain := engine.Chain(chainID)
	if chain == nil {
		api.Send(w, r, fmt.Errorf("chain not found: %s", chainID))
		return
	}
	data, _ := json.MarshalIndent(chain, "", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (rc *RuleChain) Import(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	engine := handlers.RuleEngine
	if engine == nil {
		api.Send(w, r, fmt.Errorf("rule engine not initialized"))
		return
	}
	chain, err := engine.ImportChain(body)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, map[string]interface{}{
		"imported": chain.ID,
		"name":     chain.Name,
	})
}

func (rc *RuleChain) List(w http.ResponseWriter, r *http.Request) {
	engine := handlers.RuleEngine
	if engine == nil {
		api.Send(w, r, fmt.Errorf("rule engine not initialized"))
		return
	}

	chains := engine.Chains()
	result := make([]map[string]interface{}, 0, len(chains))
	for _, c := range chains {
		result = append(result, map[string]interface{}{
			"id":          c.ID,
			"name":        c.Name,
			"description": c.Description,
			"nodeCount":   len(c.Nodes),
		})
	}

	api.Send(w, r, map[string]interface{}{
		"chains": result,
		"total":  len(result),
	})
}

type aiScriptRunRequest struct {
	Script string                 `json:"script"`
	Input  map[string]interface{} `json:"input,omitempty"`
}

type aiScriptEndpoint struct{}

func (a aiScriptEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)

	var req aiScriptRunRequest
	if len(body) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			api.Send(w, r, fmt.Errorf("invalid JSON: %w", err))
			return
		}
	}

	if req.Script == "" {
		api.Send(w, r, fmt.Errorf("script is required"))
		return
	}

	if handlers.JSRuntime == nil {
		api.Send(w, r, fmt.Errorf("JS runtime not initialized"))
		return
	}

	result := handlers.JSRuntime.Run(r.Context(), req.Script, req.Input)
	api.Send(w, r, result)
}

func MountRuleChainRoutes(r chi.Router) {
	r.Route("/rulechain", func(r chi.Router) {
		rc := (&RuleChain{}).New()
		r.Get("/", rc.List)
		r.Get("/executions", rc.Executions)
		r.Get("/{chainID}", rc.ServeHTTP)
		r.Get("/{chainID}/export", rc.Export)
		r.Post("/{chainID}/run", rc.Run)
		r.Post("/import", rc.Import)
	})

	r.Post("/ai/script/run", (&aiScriptEndpoint{}).ServeHTTP)
}
