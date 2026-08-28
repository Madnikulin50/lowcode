package rulesgo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// RemoteSpec describes a remote agent component (cmdb/scan, backup/run, …).
type RemoteSpec struct {
	Type       string // palette type, e.g. backup/run
	Service    string
	Operation  string
	Async      bool
	Ingest     string
	DefaultURL string
}

func defaultAgentURL(service string) string {
	switch strings.ToLower(strings.TrimSpace(service)) {
	case "cmdb":
		return envOr("CMDB_AGENT_URL", "http://localhost:8085/api")
	case "backup":
		return envOr("BACKUP_AGENT_URL", "http://localhost:8087/api")
	case "invest":
		return envOr("INVEST_AGENT_URL", "http://localhost:8086/api")
	default:
		return ""
	}
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

type componentExecutor struct {
	spec   RemoteSpec
	detach DetachStartFunc
}

func (n *componentExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	spec := n.spec
	if spec.Service == "" || spec.Operation == "" {
		var cfg struct {
			Service   string `json:"service"`
			Operation string `json:"operation"`
			URL       string `json:"url"`
			Async     bool   `json:"async"`
			Ingest    string `json:"ingestChainID"`
		}
		_ = json.Unmarshal(node.Config, &cfg)
		if spec.Service == "" {
			spec.Service = resolveTemplateValue(cfg.Service, ec)
		}
		if spec.Operation == "" {
			spec.Operation = resolveTemplateValue(cfg.Operation, ec)
		}
		if spec.Ingest == "" {
			spec.Ingest = resolveTemplateValue(cfg.Ingest, ec)
		}
		if cfg.URL != "" {
			spec.DefaultURL = resolveTemplateValue(cfg.URL, ec)
		}
		if cfg.Async {
			spec.Async = true
		}
	}
	if spec.Service == "" || spec.Operation == "" {
		return nil, fmt.Errorf("service and operation are required")
	}

	base := strings.TrimRight(spec.DefaultURL, "/")
	if base == "" {
		base = strings.TrimRight(defaultAgentURL(spec.Service), "/")
	}
	if emptyAny(ec.Get("agentUrl")) && base != "" {
		ec.Set("agentUrl", base)
	}
	if v := strings.TrimSpace(fmt.Sprintf("%v", ec.Get("agentUrl"))); v != "" && v != "<nil>" {
		base = strings.TrimRight(v, "/")
	}
	if base == "" {
		return map[string]interface{}{"status": "agent_not_configured", "service": spec.Service, "operation": spec.Operation}, nil
	}

	body := map[string]interface{}{"operation": spec.Operation}
	if len(node.Config) > 0 {
		var extra map[string]interface{}
		if json.Unmarshal(node.Config, &extra) == nil {
			for k, v := range extra {
				switch k {
				case "service", "operation", "url", "async", "ingestChainID":
					continue
				}
				if s, ok := v.(string); ok {
					body[k] = resolveTemplateValue(s, ec)
				} else {
					body[k] = v
				}
			}
		}
	}
	copyIfEmpty(body, "namespaceID", ec.Get("namespaceID"))
	copyIfEmpty(body, "token", firstEC(ec, "authToken", "token"))
	copyIfEmpty(body, "callbackUrl", ec.Get("callbackUrl"))
	copyIfEmpty(body, "recordID", firstEC(ec, "createdRecordID", "scanRecordID", "jobID"))

	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	url := base + "/jobs"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if tok, ok := body["token"].(string); ok && tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", spec.Service, spec.Operation, err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	result := map[string]interface{}{"statusCode": resp.StatusCode, "service": spec.Service, "operation": spec.Operation}
	var jsonBody interface{}
	if json.Unmarshal(respBody, &jsonBody) == nil {
		result["body"] = jsonBody
		if m, ok := jsonBody.(map[string]interface{}); ok {
			if id, ok := m["id"]; ok {
				s := fmt.Sprintf("%v", id)
				ec.Set("scanID", s)
				ec.Set("jobID", s)
				result["jobID"] = s
			}
		}
	} else {
		result["body"] = string(respBody)
	}
	if resp.StatusCode >= 400 {
		return result, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncateHTTPBody(respBody))
	}

	if spec.Async && n.detach != nil && spec.Ingest != "" {
		jobID := fmt.Sprintf("%v", result["jobID"])
		vars := map[string]interface{}{}
		for k, v := range ec.Input {
			vars[k] = v
		}
		for k, v := range ec.Variables {
			vars[k] = v
		}
		vars["jobID"] = jobID
		vars["scanID"] = jobID
		vars["agentUrl"] = base
		_ = n.detach(ctx, DetachConfig{
			Kind:          "poll",
			IngestChainID: spec.Ingest,
			StatusURL:     base + "/jobs/" + jobID,
			Interval:      2,
			Timeout:       3600,
			Until:         "completed,failed,done,error",
		}, vars)
		result["detached"] = true
		result["ingestChainID"] = spec.Ingest
	}
	return result, nil
}

func copyIfEmpty(body map[string]interface{}, key string, v interface{}) {
	if _, ok := body[key]; ok {
		return
	}
	if emptyAny(v) {
		return
	}
	body[key] = v
}

func firstEC(ec *ExecutionContext, keys ...string) interface{} {
	for _, k := range keys {
		if v := ec.Get(k); !emptyAny(v) {
			return v
		}
	}
	return nil
}
