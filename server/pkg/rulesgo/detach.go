package rulesgo

import (
	"context"
	"fmt"
	"os"
	"strings"
)

type DetachConfig struct {
	Kind          string `json:"kind"`
	IngestChainID string `json:"ingestChainID"`
	StatusURL     string `json:"statusUrl"`
	ItemsURL      string `json:"itemsUrl"`
	Interval      int    `json:"interval"`
	Timeout       int    `json:"timeout"`
	Until         string `json:"until"`
}

type DetachStartFunc func(ctx context.Context, cfg DetachConfig, vars map[string]interface{}) error

type detachExecutor struct {
	start DetachStartFunc
}

func (n *detachExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[DetachConfig](node.Config)
	if err != nil {
		return nil, err
	}
	if emptyAny(ec.Get("agentUrl")) {
		agentURL := strings.TrimRight(os.Getenv("CMDB_AGENT_URL"), "/")
		if agentURL == "" {
			agentURL = "http://localhost:8085/api"
		}
		if ec.Variables == nil {
			ec.Variables = make(map[string]interface{})
		}
		ec.Set("agentUrl", agentURL)
	}
	cfg.Kind = resolveTemplateValue(cfg.Kind, ec)
	if cfg.Kind == "" {
		cfg.Kind = "poll"
	}
	cfg.IngestChainID = resolveTemplateValue(cfg.IngestChainID, ec)
	cfg.StatusURL = resolveTemplateValue(cfg.StatusURL, ec)
	cfg.ItemsURL = resolveTemplateValue(cfg.ItemsURL, ec)
	cfg.Until = resolveTemplateValue(cfg.Until, ec)
	if cfg.IngestChainID == "" {
		return nil, fmt.Errorf("detach ingestChainID is required")
	}
	if cfg.Kind == "poll" && cfg.StatusURL == "" {
		return nil, fmt.Errorf("detach statusUrl is required for poll")
	}

	vars := map[string]interface{}{}
	for k, v := range ec.Input {
		vars[k] = v
	}
	for k, v := range ec.Variables {
		vars[k] = v
	}
	if emptyAny(vars["scanRecordID"]) {
		if v := vars["createdRecordID"]; !emptyAny(v) {
			vars["scanRecordID"] = v
		}
	}

	if n.start == nil {
		return map[string]interface{}{"status": "detach_not_configured", "ingestChainID": cfg.IngestChainID}, nil
	}
	if err := n.start(ctx, cfg, vars); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"detached":      true,
		"kind":          cfg.Kind,
		"ingestChainID": cfg.IngestChainID,
		"jobID":         fmt.Sprintf("%v", vars["jobID"]),
	}, nil
}
