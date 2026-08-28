package sdk

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Backend is implemented by an agent that still owns domain job state
// (backup JobStatus, cmdb ScanStatus). SDK HTTP maps it onto Envelope.
type Backend interface {
	StartJob(ctx context.Context, req StartRequest) (*Envelope, error)
	GetJob(id string) *Envelope
	ListJobs() []*Envelope
}

// SyncBackend handles POST /api/call/{op} and sync operations like backup/due.
type SyncBackend interface {
	Call(ctx context.Context, operation string, req StartRequest) (any, error)
}

// ItemsBackend is GET /api/jobs/{id}/items (cmdb devices on a scan).
type ItemsBackend interface {
	JobItems(id string) any
}

// HealthBackend overrides the default {status:ok} payload.
type HealthBackend interface {
	Health(ctx context.Context) map[string]any
}

func jsonWrite(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	jsonWrite(w, code, map[string]string{"error": msg})
}

func readBody(r *http.Request) []byte {
	if r.Body == nil {
		return nil
	}
	b, _ := io.ReadAll(io.LimitReader(r.Body, 8<<20))
	return b
}

func (s *Service) startJobHTTP(w http.ResponseWriter, r *http.Request) {
	s.startJobHTTPOp(w, r, "")
}

func (s *Service) startJobHTTPOp(w http.ResponseWriter, r *http.Request, forceOp string) {
	req, err := DecodeStartRequest(readBody(r))
	if err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if forceOp != "" {
		req.Operation = forceOp
	}
	if req.Operation == "" {
		req.Operation = s.defaultOp(req)
	}
	if s.backend == nil {
		jsonError(w, "backend not configured", http.StatusServiceUnavailable)
		return
	}
	if s.syncOps[req.Operation] {
		sb, ok := s.backend.(SyncBackend)
		if !ok {
			jsonError(w, "sync calls not supported", http.StatusNotFound)
			return
		}
		out, err := sb.Call(r.Context(), req.Operation, req)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonWrite(w, http.StatusOK, out)
		return
	}
	env, err := s.backend.StartJob(r.Context(), req)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonWrite(w, http.StatusOK, env)
}

func (s *Service) defaultOp(req StartRequest) string {
	if k := req.Param("kind"); k != "" && k != "full" && k != "incremental" {
		return k
	}
	if s.cfg.Handle != "" {
		switch s.cfg.Handle {
		case "backup":
			return "backup"
		case "cmdb":
			return "scan"
		}
	}
	if len(s.comps) == 1 {
		d := s.comps[0].Descriptor()
		return firstNonEmpty(d.Operation, d.Type)
	}
	return "backup"
}

func (s *Service) getJobHTTP(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "jobID")
	if id == "" {
		id = chi.URLParam(r, "scanID")
	}
	if s.backend == nil {
		jsonError(w, "backend not configured", http.StatusServiceUnavailable)
		return
	}
	env := s.backend.GetJob(id)
	if env == nil {
		jsonError(w, "job not found", http.StatusNotFound)
		return
	}
	jsonWrite(w, http.StatusOK, env)
}

func (s *Service) listJobsHTTP(w http.ResponseWriter, r *http.Request) {
	if s.backend == nil {
		jsonWrite(w, http.StatusOK, []*Envelope{})
		return
	}
	jsonWrite(w, http.StatusOK, s.backend.ListJobs())
}

func (s *Service) itemsHTTP(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "jobID")
	if id == "" {
		id = chi.URLParam(r, "scanID")
	}
	ib, ok := s.backend.(ItemsBackend)
	if !ok {
		jsonError(w, "items not supported", http.StatusNotFound)
		return
	}
	items := ib.JobItems(id)
	if items == nil {
		jsonError(w, "job not found", http.StatusNotFound)
		return
	}
	jsonWrite(w, http.StatusOK, items)
}

func (s *Service) callHTTP(w http.ResponseWriter, r *http.Request) {
	op := chi.URLParam(r, "op")
	sb, ok := s.backend.(SyncBackend)
	if !ok {
		jsonError(w, "sync calls not supported", http.StatusNotFound)
		return
	}
	req, err := DecodeStartRequest(readBody(r))
	if err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	out, err := sb.Call(r.Context(), op, req)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonWrite(w, http.StatusOK, out)
}

func (s *Service) healthHTTP(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{
		"status":  "ok",
		"service": s.cfg.Handle,
	}
	if hb, ok := s.backend.(HealthBackend); ok {
		for k, v := range hb.Health(r.Context()) {
			payload[k] = v
		}
	}
	jsonWrite(w, http.StatusOK, payload)
}

func (s *Service) metaHTTP(w http.ResponseWriter, _ *http.Request) {
	jsonWrite(w, http.StatusOK, s.Meta())
}

func (s *Service) registerHTTP(w http.ResponseWriter, r *http.Request) {
	if s.compose != nil && s.cfg.Heartbeat.Module != "" {
		ident := s.cfg.Identity(s.Meta().Capabilities)
		if err := s.compose.Heartbeat(r.Context(), s.cfg.Heartbeat.Module, ident); err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	jsonWrite(w, http.StatusOK, map[string]string{"status": "ok"})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func pathHasParam(p string) bool {
	return strings.Contains(p, "{")
}
