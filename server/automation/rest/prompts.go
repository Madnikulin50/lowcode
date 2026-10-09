package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/madnikulin50/lowcode/server/automation/service"
	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/api"
)

// Prompts is the REST face of the prompt library (automation/service/prompt.go),
// for the admin screen. Who may read or change prompts is decided by the
// library, not here.
type Prompts struct{}

func (Prompts) New() *Prompts { return &Prompts{} }

type (
	savePromptRequest struct {
		Text        string `json:"text"`
		Description string `json:"description"`
		Note        string `json:"note"`
		Activate    bool   `json:"activate"`
		// Cases is a pointer so "not sent" (keep the previous cases) differs
		// from "sent empty" (clear them)
		Cases *types.PromptCases `json:"cases"`
	}

	evalPromptRequest struct {
		Versions []int             `json:"versions"`
		Agent    string            `json:"agent"`
		Model    string            `json:"model"`
		Cases    types.PromptCases `json:"cases"`
	}
)

func (Prompts) List(w http.ResponseWriter, r *http.Request) {
	list, err := service.ListPrompts(r.Context())
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, map[string]interface{}{"set": list})
}

// History lists every version of a prompt, oldest first.
func (Prompts) History(w http.ResponseWriter, r *http.Request) {
	h, err := service.PromptHistory(r.Context(), chi.URLParam(r, "handle"))
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, map[string]interface{}{"set": h})
}

// Version returns one version with its text and test cases; 0 is the active one.
func (Prompts) Version(w http.ResponseWriter, r *http.Request) {
	n, err := strconv.Atoi(chi.URLParam(r, "version"))
	if err != nil || n < 0 {
		api.Send(w, r, fmt.Errorf("version must be a number"))
		return
	}

	p, cases, err := service.GetPromptWithCases(r.Context(), chi.URLParam(r, "handle"), n)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, map[string]interface{}{"prompt": p, "cases": cases})
}

func (Prompts) Save(w http.ResponseWriter, r *http.Request) {
	var req savePromptRequest
	if !readJSON(w, r, &req) {
		return
	}

	p, err := service.SavePromptVersion(r.Context(), service.SavePrompt{
		Handle:      chi.URLParam(r, "handle"),
		Text:        req.Text,
		Description: req.Description,
		Note:        req.Note,
		Cases:       req.Cases,
		Activate:    req.Activate,
	})
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, p)
}

func (Prompts) Activate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Version int `json:"version"`
	}
	if !readJSON(w, r, &req) {
		return
	}

	p, err := service.ActivatePrompt(r.Context(), chi.URLParam(r, "handle"), req.Version)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, p)
}

func (Prompts) Delete(w http.ResponseWriter, r *http.Request) {
	if err := service.DeletePrompt(r.Context(), chi.URLParam(r, "handle")); err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, map[string]interface{}{"deleted": chi.URLParam(r, "handle")})
}

// Eval runs the prompt's test cases on the model. With ?stream=1 each case is
// sent as it is judged, then the reports; a run over many cases on a local
// model takes minutes and a screen should not sit blank for them.
func (Prompts) Eval(w http.ResponseWriter, r *http.Request) {
	var req evalPromptRequest
	if !readJSON(w, r, &req) {
		return
	}

	handle := chi.URLParam(r, "handle")
	versions := make([]string, 0, len(req.Versions))
	for _, v := range req.Versions {
		versions = append(versions, strconv.Itoa(v))
	}
	cases := ""
	if len(req.Cases) > 0 {
		b, _ := json.Marshal(req.Cases)
		cases = string(b)
	}
	joined := strings.Join(versions, ",")

	if r.URL.Query().Get("stream") == "1" {
		if sse, ok := api.NewSSE(w); ok {
			reports, err := service.EvaluatePromptWithProgress(r.Context(), handle, joined, req.Agent, req.Model, cases,
				func(version int, res aiagent.EvalResult) {
					_ = sse.Send(map[string]interface{}{"version": version, "case": res, "done": false})
				})
			if err != nil {
				_ = sse.Send(map[string]interface{}{"error": err.Error(), "done": true})
				return
			}
			_ = sse.Send(map[string]interface{}{"reports": reports, "done": true})
			return
		}
	}

	reports, err := service.EvaluatePrompt(r.Context(), handle, joined, req.Agent, req.Model, cases)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, map[string]interface{}{"reports": reports})
}

func readJSON(w http.ResponseWriter, r *http.Request, into interface{}) bool {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err == nil && len(strings.TrimSpace(string(body))) > 0 {
		err = json.Unmarshal(body, into)
	}
	if err != nil {
		api.Send(w, r, fmt.Errorf("invalid JSON: %w", err))
		return false
	}
	return true
}

func MountPromptRoutes(r chi.Router) {
	p := Prompts{}
	r.Route("/prompts", func(r chi.Router) {
		r.Get("/", p.List)
		r.Get("/{handle}", p.History)
		r.Get("/{handle}/versions/{version}", p.Version)
		r.Post("/{handle}", p.Save)
		r.Post("/{handle}/activate", p.Activate)
		r.Post("/{handle}/eval", p.Eval)
		r.Delete("/{handle}", p.Delete)
	})
}
