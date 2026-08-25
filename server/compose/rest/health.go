package rest

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/compose/mcp/handlers"
	"github.com/madnikulin50/lowcode/server/pkg/api"
)

type HealthService struct {
	startTime time.Time
}

var healthService = &HealthService{startTime: time.Now()}

func (h *HealthService) Status(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":     "ok",
		"version":    "1.0.0",
		"uptime":     time.Since(h.startTime).String(),
		"startedAt":  h.startTime.Format(time.RFC3339),
		"goVersion":  runtime.Version(),
		"goroutines": runtime.NumGoroutine(),
	}

	components := map[string]interface{}{}

	if handlers.RuleEngine != nil {
		chains := handlers.RuleEngine.Chains()
		totalNodes := 0
		for _, c := range chains {
			totalNodes += len(c.Nodes)
		}
		components["rulesgo"] = map[string]interface{}{
			"status": "ok",
			"chains": len(chains),
			"nodes":  totalNodes,
		}
	} else {
		components["rulesgo"] = map[string]interface{}{"status": "not_initialized"}
	}

	if handlers.JSRuntime != nil {
		components["js_runtime"] = map[string]interface{}{"status": "ok"}
	} else {
		components["js_runtime"] = map[string]interface{}{"status": "not_initialized"}
	}

	if handlers.AgentRegistry != nil {
		components["ai_agents"] = map[string]interface{}{
			"status": "ok",
			"agents": handlers.AgentRegistry.List(),
		}
	} else {
		components["ai_agents"] = map[string]interface{}{"status": "not_initialized"}
	}

	if handlers.GonecEngine != nil {
		components["gonec"] = map[string]interface{}{"status": "ok"}
	} else {
		components["gonec"] = map[string]interface{}{"status": "not_initialized"}
	}

	status["components"] = components

	api.Send(w, r, status)
}

func (h *HealthService) Ping(w http.ResponseWriter, r *http.Request) {
	api.Send(w, r, map[string]interface{}{
		"pong": true,
		"time": time.Now().Format(time.RFC3339Nano),
	})
}

func (h *HealthService) Info(w http.ResponseWriter, r *http.Request) {
	chains := map[string]interface{}{}
	if handlers.RuleEngine != nil {
		all := handlers.RuleEngine.Chains()
		summary := make([]map[string]interface{}, 0, len(all))
		for _, c := range all {
			summary = append(summary, map[string]interface{}{
				"id":          c.ID,
				"name":        c.Name,
				"description": c.Description,
				"nodeCount":   len(c.Nodes),
			})
		}
		chains["list"] = summary
		chains["total"] = len(all)
	}

	agents := map[string]interface{}{}
	if handlers.AgentRegistry != nil {
		agents["list"] = handlers.AgentRegistry.List()
	}

	info := map[string]interface{}{
		"app":     "lowcode-server",
		"version": "1.0.0",
		"uptime":  time.Since(healthService.startTime).String(),
		"chains":  chains,
		"agents":  agents,
		"endpoints": []string{
			"GET  /health",
			"GET  /health/ping",
			"GET  /health/info",
			"GET  /admin/rulechain/",
			"POST /admin/rulechain/",
			"GET  /admin/rulechain/nodes",
			"GET  /admin/rulechain/stats",
			"GET  /admin/rulechain/{id}",
			"PUT  /admin/rulechain/{id}",
			"DELETE /admin/rulechain/{id}",
			"POST /admin/rulechain/{id}/test",
			"POST /rulechain/{id}/run",
			"GET  /rulechain/",
			"POST /rulechain/import",
			"GET  /rulechain/{id}/export",
			"POST /pageblock/trigger",
			"POST /pageblock/trigger/batch",
			"POST /ai/script/run",
			"POST /mcp/call",
		},
	}

	data, _ := json.MarshalIndent(info, "", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func MountHealthRoutes(r chi.Router) {
	r.Get("/health", healthService.Status)
	r.Get("/health/ping", healthService.Ping)
	r.Get("/health/info", healthService.Info)
}
