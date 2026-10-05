package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	automationService "github.com/madnikulin50/lowcode/server/automation/service"
	"github.com/madnikulin50/lowcode/server/compose/mcp/handlers"
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/api"
	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
)

// The AI nodes (ai, ai.operation) in the chain editor: live choices for the
// agent and model fields, and a way to try a node against the real model
// without saving the chain or running the rest of it.

var aiNodeTypes = map[string]bool{"ai": true, "ai.operation": true}

// agentChoice is one agent the editor can offer in the Agent field.
type agentChoice struct {
	Handle      string
	Description string
}

// liveAgentChoices lists the agents that exist right now (built-in specs,
// files and admin-defined ones), so a new agent shows up in the editor
// without a code change.
func liveAgentChoices() []agentChoice {
	infos := aiagent.DefaultRegistry().ListInfo()
	out := make([]agentChoice, 0, len(infos))
	for _, i := range infos {
		out = append(out, agentChoice{Handle: i.Handle, Description: i.Description})
	}
	return out
}

// enrichAINodes fills the AI nodes' fields with live data and adds the
// settings the runtime supports but the static catalog never exposed.
func enrichAINodes(defs []nodeTypeDef, agents []agentChoice, models []string) []nodeTypeDef {
	handles := make([]string, 0, len(agents))
	labels := make(map[string]string, len(agents))
	for _, a := range agents {
		handles = append(handles, a.Handle)
		if a.Description != "" {
			labels[a.Handle] = a.Description
		}
	}
	sort.Strings(handles)

	for i := range defs {
		if !aiNodeTypes[defs[i].Type] {
			continue
		}

		// work on a copy: the static catalog is shared
		fields := append([]nodeTypeField(nil), defs[i].ConfigFields...)
		for j := range fields {
			switch fields[j].Key {
			case "agent":
				if len(handles) > 0 {
					fields[j].Options = handles
					fields[j].OptionLabels = labels
				}
			case "model":
				fields[j].Suggestions = models
				fields[j].Help = "Optional. A model name, or a role (rulesgo.ai, mcp.agent). Empty: the agent's own model"
			case "outputSchema":
				fields[j].ValueOptions = []string{"string", "number", "boolean", "array", "object"}
			}
		}

		if defs[i].Type == "ai" {
			fields = append(fields,
				nf("timeout", "number", "Timeout (seconds)", help("Stop waiting for the model after this long; empty: no limit of its own")),
				nf("optional", "bool", "Continue if the AI fails", help("The chain goes on without an answer (ai_response is empty) instead of stopping")),
			)
		}
		defs[i].ConfigFields = fields
	}
	return defs
}

// canTestChainNode says whether the caller may use the node test button. The
// rule chain admin routes do no permission checks of their own, but this one
// spends the model, so it asks the workflow builders' permission. A variable
// so a test can stand in for the access control.
var canTestChainNode = automationService.CanBuildWorkflows

type nodeTestRequest struct {
	Type   string                 `json:"type"`
	Config json.RawMessage        `json:"config"`
	Input  map[string]interface{} `json:"input"`
}

// nodeTestTimeout bounds a test run, whatever the node's own timeout says.
const nodeTestTimeout = 3 * time.Minute

// NodeTest runs one AI node against the real model with a sample input, for
// the "Test node" button. Nothing is saved and the rest of the chain does not
// run.
//
// Only the AI nodes can be tried this way: others (crud, http, mail, ...)
// have side effects that a test button must not have. And even for AI nodes
// data-changing tool calls are forced off, whatever the saved config says.
func (a RuleChainAdmin) NodeTest(w http.ResponseWriter, r *http.Request) {
	if !canTestChainNode(r.Context()) {
		api.Send(w, r, automationService.WorkflowErrNotAllowedToCreate())
		return
	}

	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)

	var req nodeTestRequest
	if err := json.Unmarshal(body, &req); err != nil {
		api.Send(w, r, fmt.Errorf("invalid JSON: %w", err))
		return
	}
	req.Type = strings.TrimSpace(req.Type)
	if !aiNodeTypes[req.Type] {
		api.Send(w, r, fmt.Errorf("only AI nodes (ai, ai.operation) can be tested on their own"))
		return
	}

	cfg, err := withoutMutation(req.Config)
	if err != nil {
		api.Send(w, r, err)
		return
	}

	engine := handlers.RuleEngine
	if engine == nil {
		api.Send(w, r, fmt.Errorf("engine not initialized"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), nodeTestTimeout)
	defer cancel()

	node := rulesgo.ChainNode{ID: "test", Type: req.Type, Config: cfg}

	// ?stream=1: send what the model produces as it happens, then the result
	if r.URL.Query().Get("stream") == "1" {
		if sse, ok := api.NewSSE(w); ok {
			ctx = aiagent.ContextWithProgress(ctx, func(p aiagent.Progress) { _ = sse.Send(p.Event()) })

			res, err := engine.ExecuteNode(ctx, node, req.Input)
			if err != nil {
				_ = sse.Send(map[string]interface{}{"error": err.Error(), "done": true})
				return
			}
			_ = sse.Send(map[string]interface{}{"result": res, "done": true})
			return
		}
	}

	res, err := engine.ExecuteNode(ctx, node, req.Input)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, res)
}

// withoutMutation returns config with allowMutating forced off.
func withoutMutation(config json.RawMessage) (json.RawMessage, error) {
	m := map[string]interface{}{}
	if len(config) > 0 {
		if err := json.Unmarshal(config, &m); err != nil {
			return nil, fmt.Errorf("config must be a JSON object: %w", err)
		}
	}
	m["allowMutating"] = false
	return json.Marshal(m)
}

// ruleChainSchemas is the node catalog in the shape the chain validator
// wants (pkg/rulesgo/validate.go), live agent and model choices included.
func ruleChainSchemas() []rulesgo.NodeSchema {
	defs := nodeTypes()
	out := make([]rulesgo.NodeSchema, 0, len(defs))
	for _, d := range defs {
		s := rulesgo.NodeSchema{Type: d.Type, Label: d.Label, Description: d.Description}
		for _, f := range d.ConfigFields {
			s.Fields = append(s.Fields, rulesgo.FieldSchema{
				Key: f.Key, Widget: f.Widget, Label: f.Label, Required: f.Required,
				Options: f.Options, Default: f.Default, Help: f.Help, VisibleIf: f.VisibleIf,
			})
		}
		out = append(out, s)
	}
	return out
}
