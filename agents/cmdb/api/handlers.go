package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/agents/cmdb/agent"
)

type Handler struct {
	ag *agent.Agent
}

func New(ag *agent.Agent) *Handler {
	return &Handler{ag: ag}
}

func (h *Handler) Mount(r chi.Router) {
	r.Post("/scan", h.startScan)
	r.Get("/scans", h.listScans)
	r.Get("/scans/{scanID}", h.getScan)
	r.Get("/devices", h.listDevices)
	r.Get("/devices/{recordID}", h.getDevice)
	r.Delete("/devices/{recordID}", h.deleteDevice)
	r.Post("/modules/ensure", h.ensureModule)
	r.Get("/debug/db", h.debugDB)
	r.Get("/debug/network", h.debugNetwork)
}

func (h *Handler) startScan(w http.ResponseWriter, r *http.Request) {
	var target agent.ScanTarget
	if err := json.NewDecoder(r.Body).Decode(&target); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if target.CIDR == "" {
		jsonError(w, "cidr is required", http.StatusBadRequest)
		return
	}

	status, err := h.ag.StartScan(r.Context(), target)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResp(w, status)
}

func (h *Handler) listScans(w http.ResponseWriter, r *http.Request) {
	jsonResp(w, h.ag.ListScans())
}

func (h *Handler) getScan(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "scanID")
	s := h.ag.GetStatus(id)
	if s == nil {
		jsonError(w, "scan not found", http.StatusNotFound)
		return
	}
	jsonResp(w, s)
}

func (h *Handler) listDevices(w http.ResponseWriter, r *http.Request) {
	modID := queryUint64(r, "moduleID")
	devices, err := h.ag.ListDevices(r.Context(), modID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResp(w, devices)
}

func (h *Handler) getDevice(w http.ResponseWriter, r *http.Request) {
	recordID := parseUint64(chi.URLParam(r, "recordID"))
	modID := queryUint64(r, "moduleID")
	device, err := h.ag.GetDevice(r.Context(), modID, recordID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResp(w, device)
}

func (h *Handler) deleteDevice(w http.ResponseWriter, r *http.Request) {
	recordID := parseUint64(chi.URLParam(r, "recordID"))
	modID := queryUint64(r, "moduleID")
	if err := h.ag.DeleteDevice(r.Context(), modID, recordID); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResp(w, map[string]string{"status": "deleted"})
}

func (h *Handler) ensureModule(w http.ResponseWriter, r *http.Request) {
	modID, err := h.ag.EnsureModule(r.Context())
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResp(w, map[string]uint64{"moduleID": modID})
}

func (h *Handler) debugNetwork(w http.ResponseWriter, r *http.Request) {
	requested := r.URL.Query().Get("cidr")
	if requested == "" {
		requested = "auto"
	}
	jsonResp(w, map[string]interface{}{
		"requested":     requested,
		"resolved":      agent.ResolveScanCIDRs(requested),
		"localCIDRs":    agent.LocalIPv4CIDRs(),
		"officeDefault": agent.ResolveScanCIDRs("192.168.1.0/24"),
	})
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

func (h *Handler) debugDB(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	devices, err := h.ag.ListDevices(ctx, 0)
	if err != nil {
		jsonResp(w, map[string]interface{}{
			"error":   err.Error(),
			"devices": []agent.Device{},
			"count":   0,
		})
		return
	}
	jsonResp(w, map[string]interface{}{
		"devices": devices,
		"count":   len(devices),
	})
}

func queryUint64(r *http.Request, key string) uint64 {
	return parseUint64(r.URL.Query().Get(key))
}

func parseUint64(s string) uint64 {
	if s == "" {
		return 0
	}
	var v uint64
	fmt.Sscanf(s, "%d", &v)
	return v
}
