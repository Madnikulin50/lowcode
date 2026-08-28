package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/agents/backup/agent"
)

type Handler struct {
	ag *agent.Agent
}

func New(ag *agent.Agent) *Handler {
	return &Handler{ag: ag}
}

func (h *Handler) Mount(r chi.Router) {
	r.Post("/jobs", h.startJob)
	r.Post("/jobs/due", h.runDue)
	r.Get("/jobs", h.listJobs)
	r.Get("/jobs/{jobID}", h.getJob)
	r.Post("/restore", h.restore)
	r.Post("/prune", h.prune)
	r.Post("/register", h.register)
	r.Get("/health", h.health)
}

func (h *Handler) startJob(w http.ResponseWriter, r *http.Request) {
	var req agent.JobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	kind := req.Kind
	switch kind {
	case "restore":
		st, err := h.ag.StartRestore(r.Context(), req)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResp(w, st)
		return
	case "prune":
		st, err := h.ag.StartPrune(r.Context(), req)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResp(w, st)
		return
	case "due":
		list, err := h.ag.RunDue(r.Context(), req)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResp(w, list)
		return
	}
	if req.ResolvedSourceID() == "" && req.ResolvedPolicyID() == "" {
		jsonError(w, "sourceID or policyID is required", http.StatusBadRequest)
		return
	}
	st, err := h.ag.StartBackup(r.Context(), req)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResp(w, st)
}

func (h *Handler) runDue(w http.ResponseWriter, r *http.Request) {
	var req agent.JobRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	list, err := h.ag.RunDue(r.Context(), req)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResp(w, list)
}

func (h *Handler) listJobs(w http.ResponseWriter, r *http.Request) {
	jsonResp(w, h.ag.ListJobsStatus())
}

func (h *Handler) getJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "jobID")
	st := h.ag.GetStatus(id)
	if st == nil {
		jsonError(w, "job not found", http.StatusNotFound)
		return
	}
	jsonResp(w, st)
}

func (h *Handler) restore(w http.ResponseWriter, r *http.Request) {
	var req agent.JobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	st, err := h.ag.StartRestore(r.Context(), req)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResp(w, st)
}

func (h *Handler) prune(w http.ResponseWriter, r *http.Request) {
	var req agent.JobRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	st, err := h.ag.StartPrune(r.Context(), req)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResp(w, st)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	h.ag.Heartbeat(r.Context())
	jsonResp(w, map[string]string{"status": "ok"})
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	jsonResp(w, h.ag.Health(r.Context()))
}

func jsonResp(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
