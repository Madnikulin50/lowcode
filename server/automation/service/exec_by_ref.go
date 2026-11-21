package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
)

// ExecWorkflowByRef runs a workflow identified by numeric ID or handle with
// the given input - the entry point behind the rule chain "workflow" node
// (see pkg/rulesgo/workflow_node.go).
//
// It runs as the identity carried by ctx, so the workflow's execute
// permission is checked against the chain's caller. With async the call
// returns as soon as the session is started, but only when the workflow is
// deferred (delay/prompt steps); a plain workflow has nothing to wait for
// after it ends, so it simply finishes inline.
func ExecWorkflowByRef(ctx context.Context, ref string, input map[string]interface{}, async bool) (*rulesgo.WorkflowRunResult, error) {
	if DefaultWorkflow == nil {
		return nil, fmt.Errorf("workflow service not available")
	}

	id, err := resolveWorkflowRef(DefaultWorkflow, ref)
	if err != nil {
		return nil, err
	}

	in, err := expr.NewVars(input)
	if err != nil {
		return nil, fmt.Errorf("invalid workflow input: %w", err)
	}

	results, sessionID, _, err := DefaultWorkflow.Exec(ctx, id, types.WorkflowExecParams{
		Async: async,
		Input: in,
	})
	if err != nil {
		return nil, err
	}

	res := &rulesgo.WorkflowRunResult{SessionID: strconv.FormatUint(sessionID, 10)}
	if results != nil {
		res.Results = results.Dict()
	}
	return res, nil
}

func resolveWorkflowRef(svc *workflow, ref string) (uint64, error) {
	ref = strings.TrimSpace(ref)
	if id := svc.handleToID(ref); id != 0 {
		return id, nil
	}
	if id, err := strconv.ParseUint(ref, 10, 64); err == nil && id != 0 {
		return id, nil
	}
	return 0, fmt.Errorf("workflow %q not found", ref)
}
