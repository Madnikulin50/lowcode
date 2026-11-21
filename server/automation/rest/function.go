package rest

import (
	"context"

	"github.com/madnikulin50/lowcode/server/automation/rest/request"
	"github.com/madnikulin50/lowcode/server/automation/service"
	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

type (
	Function struct {
		reg interface {
			Functions() []*types.Function
		}
	}

	functionSetPayload struct {
		Set []*types.Function `json:"set"`
	}
)

func (Function) New() *Function {
	ctrl := &Function{reg: service.Registry()}
	return ctrl
}

func (ctrl Function) List(ctx context.Context, _ *request.FunctionList) (interface{}, error) {
	return functionSetPayload{Set: withLiveAIChoices(ctx, ctrl.reg.Functions())}, nil
}

// withLiveAIChoices turns the agent, model and skill parameters of the ai* functions
// into pick-lists of what exists right now (agents defined on the platform,
// models Ollama has, skills in the library), so the editor needs no list of its own and a new agent
// shows up without a code change.
//
// The registry's own definitions are shared, so affected functions are copied
// rather than edited. The pick-list is a convenience, not a limit: a step can
// still be given any value by switching the argument to an expression.
func withLiveAIChoices(ctx context.Context, ff []*types.Function) []*types.Function {
	var agents, models, skills []map[string]interface{}
	loaded := false
	load := func() {
		if loaded {
			return
		}
		loaded = true

		agents = []map[string]interface{}{{"value": "", "text": "(none - plain assistant)"}}
		for _, a := range aiagent.DefaultRegistry().ListInfo() {
			text := a.Handle
			if a.Description != "" {
				text += " - " + a.Description
			}
			agents = append(agents, map[string]interface{}{"value": a.Handle, "text": text})
		}

		models = []map[string]interface{}{{"value": "", "text": "(agent's own model)"}}
		for _, m := range chat.ModelChoices() {
			models = append(models, map[string]interface{}{"value": m, "text": m})
		}

		skills = []map[string]interface{}{{"value": "", "text": "(none)"}}
		for _, sk := range aiagent.ListSkills(ctx) {
			text := sk.Handle
			if sk.Description != "" {
				text += " - " + sk.Description
			}
			skills = append(skills, map[string]interface{}{"value": sk.Handle, "text": text})
		}
	}

	out := make([]*types.Function, len(ff))
	for i, f := range ff {
		out[i] = f
		if f.Labels["ai"] == "" {
			continue
		}

		cp := *f
		cp.Parameters = make([]*types.Param, len(f.Parameters))
		for j, p := range f.Parameters {
			cp.Parameters[j] = p
			if p.Name != "agent" && p.Name != "model" && p.Name != "skill" {
				continue
			}

			load()
			options := agents
			switch p.Name {
			case "model":
				options = models
			case "skill":
				options = skills
			}
			if p.Required {
				// there is no "none" for a required choice
				options = options[1:]
			}

			pc := *p
			meta := types.ParamMeta{}
			if p.Meta != nil {
				meta = *p.Meta
			}
			meta.Visual = map[string]interface{}{"input": map[string]interface{}{
				"type":       "select",
				"default":    "",
				"properties": map[string]interface{}{"options": options},
			}}
			pc.Meta = &meta
			cp.Parameters[j] = &pc
		}
		out[i] = &cp
	}
	return out
}
