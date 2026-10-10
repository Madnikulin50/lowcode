package types

import (
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/filter"
)

type (
	// WorkflowVersion is one saved snapshot of a workflow (its definition and
	// settings, not its triggers or run state). Versions are immutable and
	// numbered per workflow from 1; restoring one applies its snapshot through
	// the regular update, which saves a new version.
	WorkflowVersion struct {
		ID uint64 `json:"versionID,string" schema:"col=id,dal=id,unique"`

		WorkflowID uint64 `json:"workflowID,string" schema:"col=workflow_id,dal=id"`
		Version    int    `json:"version" schema:"col=version,dal=number:default0,sortable"`

		// Snapshot is the workflow serialized as JSON
		Snapshot string `json:"snapshot,omitempty" schema:"col=snapshot,dal"`
		Note     string `json:"note,omitempty" schema:"col=note,dal"`

		CreatedAt time.Time `json:"createdAt,omitempty" schema:"col=created_at,dal=timestamp:now,sortable"`
		CreatedBy uint64    `json:"createdBy,string" schema:"col=created_by,dal=userref"`
	}

	WorkflowVersionFilter struct {
		VersionID  []uint64 `json:"versionID"`
		WorkflowID uint64   `json:"workflowID,string"`

		// Check fn is called by store backend for each resource found function can
		// modify the resource and return false if store should not return it
		//
		// Store then loads additional resources to satisfy the paging parameters
		Check func(*WorkflowVersion) (bool, error) `json:"-"`

		// Standard helpers for paging and sorting
		filter.Sorting
		filter.Paging
	}
)
