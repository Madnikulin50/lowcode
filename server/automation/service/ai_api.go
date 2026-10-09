package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
)

// A small, serialisable view of workflows and their sessions for callers that
// are not the workflow editor: the chat and MCP tools, through which an
// agent or a user lists workflows, starts one and follows it.

type (
	WorkflowBrief struct {
		ID          string `json:"id"`
		Handle      string `json:"handle"`
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
	}

	SessionBrief struct {
		ID         string     `json:"sessionID"`
		WorkflowID string     `json:"workflowID"`
		Status     string     `json:"status"`
		Error      string     `json:"error,omitempty"`
		CreatedAt  time.Time  `json:"createdAt"`
		Suspended  *time.Time `json:"suspendedAt,omitempty"`
		Completed  *time.Time `json:"completedAt,omitempty"`
	}
)

// ListWorkflowBriefs returns the enabled workflows the caller may read,
// optionally narrowed by a free-text query.
func ListWorkflowBriefs(ctx context.Context, query string) ([]WorkflowBrief, error) {
	if DefaultWorkflow == nil {
		return nil, fmt.Errorf("workflow service not available")
	}

	set, _, err := DefaultWorkflow.Search(ctx, types.WorkflowFilter{
		Query:       query,
		Disabled:    filter.StateExcluded,
		SubWorkflow: filter.StateExcluded,
	})
	if err != nil {
		return nil, err
	}

	out := make([]WorkflowBrief, 0, len(set))
	for _, wf := range set {
		b := WorkflowBrief{ID: strconv.FormatUint(wf.ID, 10), Handle: wf.Handle}
		if wf.Meta != nil {
			b.Name, b.Description = wf.Meta.Name, wf.Meta.Description
		}
		out = append(out, b)
	}
	return out, nil
}

// WorkflowSessionBrief reports where a workflow run stands, for following up
// on a run started with async.
func WorkflowSessionBrief(ctx context.Context, sessionID uint64) (*SessionBrief, error) {
	if DefaultSession == nil {
		return nil, fmt.Errorf("session service not available")
	}

	ses, err := DefaultSession.LookupByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	return &SessionBrief{
		ID:         strconv.FormatUint(ses.ID, 10),
		WorkflowID: strconv.FormatUint(ses.WorkflowID, 10),
		Status:     ses.Status.String(),
		Error:      ses.Error,
		CreatedAt:  ses.CreatedAt,
		Suspended:  ses.SuspendedAt,
		Completed:  ses.CompletedAt,
	}, nil
}

// CanBuildWorkflows says whether the caller builds workflows - may create
// them. It is the bar for the test buttons of the AI editors: a test costs a
// model call, so it is for the people who author flows, not for anyone who
// can reach the endpoint (the token middleware lets anonymous requests
// through and leaves permissions to the services).
func CanBuildWorkflows(ctx context.Context) bool {
	return DefaultAccessControl != nil && DefaultAccessControl.CanCreateWorkflow(ctx)
}
