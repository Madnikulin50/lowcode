// The workflow node runs an automation (BPMN-style) workflow from a rule
// chain - the mirror image of the compose.runRuleChain workflow function
// (compose/service/rulechain_bridge.go), which runs a chain from a workflow.
// Together they let either engine reuse the other: chains for integration
// plumbing (Kafka, 1C, HTTP, CRUD), workflows for long-running, human-in-the-
// loop and ai* step orchestration.
package rulesgo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type (
	WorkflowRunResult struct {
		SessionID string
		// Results are the workflow's output variables; empty for an async
		// start, which returns before the workflow finishes.
		Results map[string]interface{}
	}

	// WorkflowExecFunc runs the workflow identified by ref (numeric ID or
	// handle). With async the call returns once the session is started.
	WorkflowExecFunc func(ctx context.Context, ref string, input map[string]interface{}, async bool) (*WorkflowRunResult, error)
)

type wfConfig struct {
	// WorkflowID is the workflow's numeric ID or its handle.
	WorkflowID string `json:"workflowID"`

	// Payload is a JSON object (templates allowed) used as the workflow input.
	Payload string `json:"payload"`

	// Input: named value -> template, merged over Payload.
	Input map[string]string `json:"input,omitempty"`

	// Async starts the workflow and moves on without waiting for the result.
	Async bool `json:"async,omitempty"`
}

type wfExecutor struct {
	exec WorkflowExecFunc
}

func (n *wfExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[wfConfig](node.Config)
	if err != nil {
		return nil, err
	}

	ref := strings.TrimSpace(resolveTemplateValue(cfg.WorkflowID, ec))
	if ref == "" {
		return nil, fmt.Errorf("workflow: workflowID is required")
	}

	input := map[string]interface{}{}
	if payload := strings.TrimSpace(resolveTemplateValue(cfg.Payload, ec)); payload != "" {
		if err := json.Unmarshal([]byte(payload), &input); err != nil {
			return nil, fmt.Errorf("workflow: payload must be a JSON object: %w", err)
		}
	}
	for k, v := range cfg.Input {
		input[k] = resolveTemplateValue(v, ec)
	}

	if n.exec == nil {
		return map[string]interface{}{"workflowID": ref, "status": "not_configured"}, nil
	}

	res, err := n.exec(ctx, ref, input, cfg.Async)
	if err != nil {
		return nil, fmt.Errorf("workflow: %w", err)
	}

	status := "completed"
	if cfg.Async {
		status = "started"
	}
	out := map[string]interface{}{
		"success":    true,
		"workflowID": ref,
		"status":     status,
	}
	if res != nil {
		out["sessionID"] = res.SessionID
		if res.Results != nil {
			out["results"] = res.Results
		}
	}
	return out, nil
}
