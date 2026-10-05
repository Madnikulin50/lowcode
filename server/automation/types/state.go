package types

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/madnikulin50/lowcode/server/pkg/sql"
	"github.com/madnikulin50/lowcode/server/pkg/wfexec"
)

type (
	// State is a persisted point a session is suspended on (a delay or a
	// prompt), kept so the session can be put back after a server restart.
	//
	// A session can have more than one state. Rows exist only while the
	// session is suspended: they are replaced on every suspension and
	// removed when the state is resumed or the session ends.
	State struct {
		ID         uint64 `json:"stateID,string" schema:"col=id,dal=id,unique"`
		SessionID  uint64 `json:"sessionID,string" schema:"col=session_id,dal=ref:corteza::automation:session,sortable"`
		WorkflowID uint64 `json:"workflowID,string" schema:"col=workflow_id,dal=ref:corteza::automation:workflow,sortable"`

		// wfexec.SnapshotDelayed | wfexec.SnapshotPrompted
		Kind string `json:"kind" schema:"col=kind,dal=text:16,sortable"`

		// when a delayed state is due; nil for prompts
		ResumeAt *time.Time `json:"resumeAt,omitempty" schema:"col=resume_at,dal=timestamp:nil,sortable"`

		CreatedAt time.Time `json:"createdAt,omitempty" schema:"col=created_at,dal=timestamp:now,sortable"`

		Snapshot *StateSnapshot `json:"snapshot" schema:"col=snapshot,dal=json:empty,omit"`
	}

	// StateSnapshot is what wfexec needs to put the state back into a session
	StateSnapshot wfexec.SuspendedSnapshot

	StateFilter struct {
		StateID    []string `json:"stateID"`
		SessionID  []string `json:"sessionID"`
		WorkflowID []string `json:"workflowID"`

		// Check fn is called by store backend for each resource found function can
		// modify the resource and return false if store should not return it
		//
		// Store then loads additional resources to satisfy the paging parameters
		Check func(*State) (bool, error) `json:"-"`

		// Standard helpers for paging and sorting
		filter.Sorting
		filter.Paging
	}
)

func (vv *StateSnapshot) Scan(src any) error           { return sql.ParseJSON(src, vv) }
func (vv *StateSnapshot) Value() (driver.Value, error) { return json.Marshal(vv) }
