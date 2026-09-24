package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/sql"

	"github.com/madnikulin50/lowcode/server/pkg/expr"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
	labelTypes "github.com/madnikulin50/lowcode/server/pkg/label/types"
)

type (
	// Workflow represents entire workflow definition
	Workflow struct {
		ID      uint64                           `json:"workflowID,string" schema:"col=id,dal=id,unique"`
		Handle  string                           `json:"handle" schema:"col=handle,dal=text:64,unique,ignoreCase"`
		Labels  map[string]labelTypes.LabelValue `json:"labels,omitempty"`
		Meta    *WorkflowMeta                    `json:"meta,omitempty" schema:"col=meta,dal=json:empty,omit"`
		Enabled bool                             `json:"enabled" schema:"col=enabled,dal=bool:true,sortable"`

		Trace bool `json:"trace" schema:"col=trace,dal=bool:false"`

		// how much time do we keep completed sessions (in sec)
		KeepSessions int `json:"keepSessions" schema:"col=keep_sessions,dal=number:default0"`

		// Initial input scope
		Scope *expr.Vars `json:"scope" schema:"col=scope,dal=json:empty,omit"`

		Steps WorkflowStepSet `json:"steps" schema:"col=steps,dal=json:empty,omit"`
		Paths WorkflowPathSet `json:"paths" schema:"col=paths,dal=json:empty,omit"`

		// Collection of issues from the last parse
		Issues WorkflowIssueSet `json:"issues,omitempty" schema:"col=issues,dal=json:empty,omit"`

		RunAs uint64 `json:"runAs,string" schema:"col=run_as,dal=userref"`

		OwnedBy   uint64     `json:"ownedBy,string" schema:"col=owned_by,dal=userref"`
		CreatedAt time.Time  `json:"createdAt,omitempty" schema:"col=created_at,dal=timestamp:now,sortable"`
		CreatedBy uint64     `json:"createdBy,string" schema:"col=created_by,dal=userref"`
		UpdatedAt *time.Time `json:"updatedAt,omitempty" schema:"col=updated_at,dal=timestamp:nil,sortable"`
		UpdatedBy uint64     `json:"updatedBy,string,omitempty" schema:"col=updated_by,dal=userref"`
		DeletedAt *time.Time `json:"deletedAt,omitempty" schema:"col=deleted_at,dal=timestamp:nil,sortable"`
		DeletedBy uint64     `json:"deletedBy,string,omitempty" schema:"col=deleted_by,dal=userref"`
	}

	WorkflowFilter struct {
		WorkflowID []string `json:"workflowID"`

		Handle string `json:"handle"`

		Query string `json:"query"`

		Deleted  filter.State `json:"deleted"`
		Disabled filter.State `json:"disabled"`

		// include sub-workflows
		SubWorkflow filter.State `json:"subWorkflow"`

		LabeledIDs []uint64                         `json:"-"`
		Labels     map[string]labelTypes.LabelValue `json:"labels,omitempty"`

		// Check fn is called by store backend for each resource found function can
		// modify the resource and return false if store should not return it
		//
		// Store then loads additional resources to satisfy the paging parameters
		Check func(*Workflow) (bool, error) `json:"-"`

		// Standard helpers for paging and sorting
		filter.Sorting
		filter.Paging
	}

	WorkflowMeta struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		Visual      map[string]interface{} `json:"visual"`

		// list as one of the sub-workflows, when set to true
		// there should be no enabled triggers on this workflow
		SubWorkflow bool `json:"subWorkflow,omitempty"`
	}

	WorkflowIssue struct {
		// url encoded location of the error:
		Culprit     map[string]int `json:"culprit"`
		Description string         `json:"description"`
	}

	WorkflowExecParams struct {
		// When executed as a sub-workflow
		CallerWorkflowID uint64

		// When executed as a sub-workflow
		CallerSessionID uint64

		// When executed as a sub-workflow
		CallerStepID uint64

		// Start with this specific step
		StepID uint64

		EventType    string
		ResourceType string

		// Enable execution tracing
		Trace bool

		// Do not wait for workflow to be finished
		Async bool

		// Wait for workflow to be executed even if it's deferred
		Wait bool

		Input *expr.Vars
	}
)

// CheckDeferred returns true if any of the steps is deferred.
//
// Workflow is considered deferred when delay or prompt step types are used.
// Deferred workflows cannot short-circuit triggers or prevent creation/update on before triggers
//
// @todo add flag on workflow to explicitly mark workflow as deferred even when there are no delay or prompt steps
func (r Workflow) CheckDeferred() bool {
	return r.Steps.HasDeferred()
}

// Executable returns true if workflow is valid and enabled
func (r Workflow) Executable() bool {
	return r.DeletedAt == nil && r.Enabled
}

func (r Workflow) Dict() map[string]interface{} {
	return map[string]interface{}{
		"ID":         r.ID,
		"workflowID": r.ID,
		"labels":     r.Labels,
		"ownedBy":    r.OwnedBy,
		"createdAt":  r.CreatedAt,
		"createdBy":  r.CreatedBy,
		"updatedAt":  r.UpdatedAt,
		"updatedBy":  r.UpdatedBy,
		"deletedAt":  r.DeletedAt,
		"deletedBy":  r.DeletedBy,
	}
}

func (vv *WorkflowMeta) Scan(src any) error           { return sql.ParseJSON(src, vv) }
func (vv *WorkflowMeta) Value() (driver.Value, error) { return json.Marshal(vv) }

func (issue *WorkflowIssue) String() string {
	return fmt.Sprintf("%s [%v]", issue.Description, issue.Culprit)
}

func (set *WorkflowIssueSet) Scan(src any) error          { return sql.ParseJSON(src, set) }
func (set WorkflowIssueSet) Value() (driver.Value, error) { return json.Marshal(set) }

func (set WorkflowIssueSet) Error() string {
	switch len(set) {
	case 0:
		return fmt.Sprintf("no workflow issue found")
	case 1:
		return fmt.Sprintf("1 workflow issue found")
	default:
		return fmt.Sprintf("%d workflow issues found", len(set))
	}
}

func (set WorkflowIssueSet) Append(err error, culprit map[string]int) WorkflowIssueSet {
	if culprit == nil {
		culprit = make(map[string]int)
	}

	return append(set, &WorkflowIssue{
		Culprit:     culprit,
		Description: err.Error(),
	})
}

// Distinct returns set of issues without duplicates
func (set WorkflowIssueSet) Distinct() (out WorkflowIssueSet) {
	idx := make(map[string]bool)
	out = make([]*WorkflowIssue, 0, len(set))

	for i := range set {
		if idx[set[i].String()] {
			continue
		}

		out = append(out, set[i])
		idx[set[i].String()] = true
	}

	return
}

func (set WorkflowIssueSet) SetCulprit(name string, pos int) WorkflowIssueSet {
	for i := range set {
		set[i].Culprit[name] = pos
	}

	return set
}
