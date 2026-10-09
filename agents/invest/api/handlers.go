package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/agents/invest/agent"
)

type Handler struct {
	eng *agent.Engine
}

func New(eng *agent.Engine) *Handler {
	return &Handler{eng: eng}
}

func (h *Handler) Mount(r chi.Router) {
	r.Get("/health", h.health)
	r.Post("/recalculate-evm", h.recalculate)
	r.Post("/critical-path", h.criticalPath)
	r.Post("/alerts", h.alerts)
	r.Post("/submit-approval", h.submitApproval)
	r.Post("/decide-approval", h.decideApproval)
	r.Post("/escalate-approval", h.escalateApproval)
	r.Post("/simulate-rfc", h.simulateRFC)
	r.Post("/approve-rfc", h.approveRFC)
	r.Post("/reject-rfc", h.rejectRFC)
	r.Post("/clone-wbs", h.cloneWBS)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	jsonResp(w, map[string]string{"status": "ok", "service": "invest-engine"})
}

func (h *Handler) recalculate(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJob(w, r)
	if !ok {
		return
	}
	res, err := h.eng.RecalculateEVM(r.Context(), req)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResp(w, res)
}

func (h *Handler) criticalPath(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJob(w, r)
	if !ok {
		return
	}
	res, err := h.eng.CriticalPath(r.Context(), req)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResp(w, res)
}

func (h *Handler) alerts(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJob(w, r)
	if !ok {
		return
	}
	res, err := h.eng.Alerts(r.Context(), req)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResp(w, res)
}

func (h *Handler) submitApproval(w http.ResponseWriter, r *http.Request) {
	h.wrap(w, r, func(ctx context.Context, req agent.JobRequest) (interface{}, error) {
		return h.eng.SubmitApproval(ctx, req)
	})
}

func (h *Handler) decideApproval(w http.ResponseWriter, r *http.Request) {
	h.wrap(w, r, func(ctx context.Context, req agent.JobRequest) (interface{}, error) {
		return h.eng.DecideApproval(ctx, req)
	})
}

func (h *Handler) escalateApproval(w http.ResponseWriter, r *http.Request) {
	h.wrap(w, r, func(ctx context.Context, req agent.JobRequest) (interface{}, error) {
		return h.eng.EscalateApproval(ctx, req)
	})
}

func (h *Handler) simulateRFC(w http.ResponseWriter, r *http.Request) {
	h.wrap(w, r, func(ctx context.Context, req agent.JobRequest) (interface{}, error) {
		return h.eng.SimulateRFC(ctx, req)
	})
}

func (h *Handler) approveRFC(w http.ResponseWriter, r *http.Request) {
	h.wrap(w, r, func(ctx context.Context, req agent.JobRequest) (interface{}, error) {
		return h.eng.ApproveRFC(ctx, req)
	})
}

func (h *Handler) rejectRFC(w http.ResponseWriter, r *http.Request) {
	h.wrap(w, r, func(ctx context.Context, req agent.JobRequest) (interface{}, error) {
		return h.eng.RejectRFC(ctx, req)
	})
}

func (h *Handler) cloneWBS(w http.ResponseWriter, r *http.Request) {
	h.wrap(w, r, func(ctx context.Context, req agent.JobRequest) (interface{}, error) {
		return h.eng.CloneWBSTemplate(ctx, req)
	})
}

func (h *Handler) wrap(w http.ResponseWriter, r *http.Request, fn func(context.Context, agent.JobRequest) (interface{}, error)) {
	req, ok := decodeJob(w, r)
	if !ok {
		return
	}
	res, err := fn(r.Context(), req)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResp(w, res)
}

func decodeJob(w http.ResponseWriter, r *http.Request) (agent.JobRequest, bool) {
	var req agent.JobRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
			jsonError(w, "invalid JSON", http.StatusBadRequest)
			return req, false
		}
	}
	return req, true
}

func jsonResp(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	// Top-level payload so rule-chain HTTP promote copies spi/cpi/eac into {{spi}}.
	_ = json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
