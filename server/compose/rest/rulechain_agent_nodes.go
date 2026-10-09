package rest

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func agentNodeTypes() []nodeTypeDef {
	return []nodeTypeDef{
		{
			Type:        "service.call",
			Label:       "Service call",
			Description: "Invoke a lowcode agent component by service + operation (SDK contract v1)",
			ConfigFields: []nodeTypeField{
				nf("service", "enum", "Service", req, opts("cmdb", "backup", "invest")),
				nf("operation", "string", "Operation", req, tmpl),
				nf("url", "string", "Agent URL override", tmpl, help("Default from CMDB_AGENT_URL / BACKUP_AGENT_URL")),
				nf("async", "bool", "Async job", def(true)),
				nf("ingestChainID", "string", "Ingest chain", tmpl),
			},
		},
		{
			Type:        "cmdb/scan",
			Label:       "CMDB: scan network",
			Description: "Scan a CIDR range via the CMDB agent",
			ConfigFields: []nodeTypeField{
				nf("cidr", "string", "CIDR", req, tmpl),
				nf("namespaceID", "string", "Namespace ID", tmpl),
			},
		},
		{
			Type:        "backup/run",
			Label:       "Backup: run",
			Description: "Start a backup from a source or policy",
			ConfigFields: []nodeTypeField{
				nf("sourceID", "string", "Source ID", tmpl),
				nf("policyID", "string", "Policy ID", tmpl),
			},
		},
		{
			Type:        "backup/restore",
			Label:       "Backup: restore",
			Description: "Restore a snapshot",
			ConfigFields: []nodeTypeField{
				nf("snapshotID", "string", "Snapshot ID", req, tmpl),
				nf("destType", "string", "Destination type", tmpl),
				nf("destPath", "string", "Destination path", tmpl),
			},
		},
		{
			Type:        "backup/prune",
			Label:       "Backup: prune",
			Description: "Apply retention",
			ConfigFields: []nodeTypeField{
				nf("policyID", "string", "Policy ID", tmpl),
				nf("sourceID", "string", "Source ID", tmpl),
				nf("retentionDays", "number", "Retention days"),
			},
		},
		{
			Type:        "backup/due",
			Label:       "Backup: run due",
			Description: "Run policies whose cron matches now",
		},
	}
}

func fetchLiveAgentNodeTypes() []nodeTypeDef {
	urls := []string{
		strings.TrimRight(envOr("CMDB_AGENT_URL", ""), "/"),
		strings.TrimRight(envOr("BACKUP_AGENT_URL", ""), "/"),
	}
	var out []nodeTypeDef
	client := &http.Client{Timeout: 200 * time.Millisecond}
	for _, base := range urls {
		if base == "" {
			continue
		}
		resp, err := client.Get(base + "/meta")
		if err != nil {
			continue
		}
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()
		if resp.StatusCode >= 400 {
			continue
		}
		var meta struct {
			Components []nodeTypeDef `json:"components"`
		}
		if json.Unmarshal(raw, &meta) != nil {
			continue
		}
		out = append(out, meta.Components...)
	}
	return out
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
