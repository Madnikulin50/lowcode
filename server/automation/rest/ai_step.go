package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/madnikulin50/lowcode/server/automation/service"
	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/api"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
)

// AIStep lets the workflow editor try an AI step against the real model with
// sample arguments, without building a whole test workflow around it.
type AIStep struct{}

func (AIStep) New() *AIStep { return &AIStep{} }

// testableAIFunctions are the AI functions that are safe to run from a "test
// step" button: they read and think but change nothing. aiTool and aiRunCalls
// act on the platform (they are how an approved change is carried out), so
// trying them out is a job for a real test run of the workflow.
var testableAIFunctions = map[string]bool{
	"aiAsk":       true,
	"aiExtract":   true,
	"aiClassify":  true,
	"aiRagSearch": true,
}

// canTestAIStep says whether the caller may use the test button: a test costs
// a model call, so it is for people who build workflows. A variable so a test
// can stand in for the access control, which needs a whole RBAC setup.
var canTestAIStep = service.CanBuildWorkflows

// aiStepTimeout bounds a test, whatever timeoutSec the step is given.
const aiStepTimeout = 3 * time.Minute

type aiStepRequest struct {
	Args map[string]interface{} `json:"args"`
}

// Test runs one AI function with the given arguments and returns its results
// (for the LLM steps that includes the trace: prompt, response, model,
// tokens, time). Nothing is saved.
//
// Data changes are off whatever the arguments say - allowMutating is forced
// false - and the caller must be allowed to create workflows, i.e. be someone
// who builds them: a test costs a model call.
func (AIStep) Test(w http.ResponseWriter, r *http.Request) {
	ref := chi.URLParam(r, "ref")
	if !testableAIFunctions[ref] {
		api.Send(w, r, fmt.Errorf("%q cannot be tested on its own (only %s)", ref, "aiAsk, aiExtract, aiClassify, aiRagSearch"))
		return
	}

	if !canTestAIStep(r.Context()) {
		api.Send(w, r, service.WorkflowErrNotAllowedToCreate())
		return
	}

	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)

	var req aiStepRequest
	if err := json.Unmarshal(body, &req); err != nil {
		api.Send(w, r, fmt.Errorf("invalid JSON: %w", err))
		return
	}

	// ?stream=1: send what the model produces as it happens, then the result
	if r.URL.Query().Get("stream") == "1" {
		if sse, ok := api.NewSSE(w); ok {
			ctx := aiagent.ContextWithProgress(r.Context(), func(p aiagent.Progress) { _ = sse.Send(p.Event()) })

			out, err := runAIStep(ctx, ref, req.Args)
			if err != nil {
				_ = sse.Send(map[string]interface{}{"error": err.Error(), "done": true})
				return
			}
			_ = sse.Send(map[string]interface{}{"result": out, "done": true})
			return
		}
	}

	out, err := runAIStep(r.Context(), ref, req.Args)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, out)
}

// runAIStep casts args to the function's parameter types, runs it, and
// returns its results as plain data.
func runAIStep(ctx context.Context, ref string, args map[string]interface{}) (map[string]interface{}, error) {
	f := service.Registry().Function(ref)
	if f == nil {
		return nil, fmt.Errorf("function %q is not registered", ref)
	}

	in, err := castArgs(f, args)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, aiStepTimeout)
	defer cancel()

	res, err := f.Handler(ctx, in)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return map[string]interface{}{}, nil
	}
	return res.Dict(), nil
}

// castArgs builds the typed input a function handler expects from plain JSON
// values, checking required parameters and ignoring unknown ones.
func castArgs(f *types.Function, args map[string]interface{}) (*expr.Vars, error) {
	in := &expr.Vars{}

	for _, p := range f.Parameters {
		// a test must never change data
		raw, has := args[p.Name]
		switch p.Name {
		case "allowMutating":
			raw, has = false, true
		case "deferConfirm":
			raw, has = true, true
		}

		if !has || raw == nil || raw == "" {
			if p.Required {
				return nil, fmt.Errorf("%s is required", p.Name)
			}
			continue
		}

		var typ expr.Type
		for _, name := range p.Types {
			if typ = service.Registry().Type(name); typ != nil {
				break
			}
		}
		if typ == nil {
			return nil, fmt.Errorf("%s: unknown type %v", p.Name, p.Types)
		}

		val, err := typ.Cast(raw)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p.Name, err)
		}
		if err = expr.Assign(in, p.Name, val); err != nil {
			return nil, fmt.Errorf("%s: %w", p.Name, err)
		}
	}

	return in, nil
}

func MountAIStepRoutes(r chi.Router) {
	s := AIStep{}
	r.Route("/ai", func(r chi.Router) {
		r.Post("/functions/{ref}/test", s.Test)
	})
}
