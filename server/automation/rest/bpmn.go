package rest

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/madnikulin50/lowcode/server/automation/service"
	"github.com/madnikulin50/lowcode/server/pkg/api"
)

// BPMN exposes the BPMN 2.0 compiler (automation/service/bpmn_compile.go) as
// a REST endpoint: a BPMN-notation editor (see the client3 bpmn-js
// integration) sends raw BPMN XML here to preview the Steps/Paths it will
// compile to, or to get a workflow payload ready to save through the
// existing workflow create/update endpoints - no separate storage, no
// changes to the execution engine.
type BPMN struct{}

func (BPMN) New() *BPMN { return &BPMN{} }

// Compile compiles a BPMN 2.0 XML document and returns the resulting
// workflow (Steps/Paths only - Handle/Meta/Enabled are left for the caller
// to fill in before POSTing to the normal workflow create endpoint).
func (BPMN) Compile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	if len(raw) == 0 {
		api.Send(w, r, io.ErrUnexpectedEOF)
		return
	}

	wf, err := service.CompileBPMN(raw)
	if err != nil {
		api.Send(w, r, err)
		return
	}

	api.Send(w, r, map[string]interface{}{
		"steps": wf.Steps,
		"paths": wf.Paths,
	})
}

func MountBPMNRoutes(r chi.Router) {
	b := BPMN{}
	r.Route("/bpmn", func(r chi.Router) {
		r.Post("/compile", b.Compile)
	})
}
