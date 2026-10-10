package rest

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/madnikulin50/lowcode/server/automation/service"
	"github.com/madnikulin50/lowcode/server/pkg/api"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
	"github.com/madnikulin50/lowcode/server/pkg/payload"
)

// WorkflowVersions is the REST face of the workflow version history
// (automation/service/workflow_version.go). Permissions are checked by the
// service against the workflow: read to look, update to restore.
//
//	GET  /workflows/{workflowID}/versions
//	GET  /workflows/{workflowID}/versions/{version}
//	POST /workflows/{workflowID}/versions/{version}/restore
//
// Hand-written (not in rest.yaml); the client methods are hand-written too.
type WorkflowVersions struct{}

func (WorkflowVersions) List(w http.ResponseWriter, r *http.Request) {
	set, err := service.DefaultWorkflowVersions.List(r.Context(), payload.ParseUint64(chi.URLParam(r, "workflowID")))
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, map[string]interface{}{"set": set})
}

func (WorkflowVersions) Get(w http.ResponseWriter, r *http.Request) {
	version, err := strconv.Atoi(chi.URLParam(r, "version"))
	if err != nil {
		api.Send(w, r, errors.InvalidData("invalid version"))
		return
	}

	v, err := service.DefaultWorkflowVersions.Get(r.Context(), payload.ParseUint64(chi.URLParam(r, "workflowID")), version)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, v)
}

func (WorkflowVersions) Restore(w http.ResponseWriter, r *http.Request) {
	version, err := strconv.Atoi(chi.URLParam(r, "version"))
	if err != nil {
		api.Send(w, r, errors.InvalidData("invalid version"))
		return
	}

	wf, err := service.DefaultWorkflowVersions.Restore(r.Context(), payload.ParseUint64(chi.URLParam(r, "workflowID")), version)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, wf)
}

func MountWorkflowVersionRoutes(r chi.Router) {
	v := WorkflowVersions{}
	r.Get("/workflows/{workflowID}/versions", v.List)
	r.Get("/workflows/{workflowID}/versions/{version}", v.Get)
	r.Post("/workflows/{workflowID}/versions/{version}/restore", v.Restore)
}
