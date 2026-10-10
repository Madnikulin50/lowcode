package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/madnikulin50/lowcode/server/pkg/id"
	"github.com/madnikulin50/lowcode/server/store"
	"go.uber.org/zap"
)

// Version history of workflows.
//
// Every successful create and change of a workflow saves its definition as the
// next numbered version (identical consecutive snapshots are not repeated).
// Restoring a version applies the definition through the regular update, so
// permissions, validation and the stale check work as for any edit, and the
// restore itself becomes the newest version.
//
// A snapshot holds what defines the workflow: handle, name and description,
// steps, paths, scope, trace and session settings. Restore leaves alone what
// is about who and whether: Enabled, RunAs and the owner (an old version must
// not silently switch a workflow on or change whose rights it runs with),
// and the triggers, which are separate records.

const workflowVersionsKept = 200

type (
	workflowVersion struct {
		store store.Storer
		ac    workflowVersionAccessController
	}

	workflowVersionAccessController interface {
		CanReadWorkflow(context.Context, *types.Workflow) bool
		CanUpdateWorkflow(context.Context, *types.Workflow) bool
	}

	workflowSnapshot struct {
		Handle       string                `json:"handle"`
		Meta         *types.WorkflowMeta   `json:"meta,omitempty"`
		Enabled      bool                  `json:"enabled"`
		Trace        bool                  `json:"trace"`
		KeepSessions int                   `json:"keepSessions"`
		Scope        *expr.Vars            `json:"scope,omitempty"`
		Steps        types.WorkflowStepSet `json:"steps"`
		Paths        types.WorkflowPathSet `json:"paths"`
	}
)

func WorkflowVersion(s store.Storer, ac workflowVersionAccessController) *workflowVersion {
	return &workflowVersion{store: s, ac: ac}
}

func snapshotWorkflow(wf *types.Workflow) ([]byte, error) {
	return json.Marshal(workflowSnapshot{
		Handle:       wf.Handle,
		Meta:         wf.Meta,
		Enabled:      wf.Enabled,
		Trace:        wf.Trace,
		KeepSessions: wf.KeepSessions,
		Scope:        wf.Scope,
		Steps:        wf.Steps,
		Paths:        wf.Paths,
	})
}

// recordWorkflowVersion saves the workflow as a new version. History is an
// aid: a failure is logged and never fails the change that triggered it.
func recordWorkflowVersion(ctx context.Context, s store.Storer, wf *types.Workflow) {
	if s == nil || wf == nil {
		return
	}

	if err := saveWorkflowVersion(ctx, s, wf, ""); err != nil && DefaultLogger != nil {
		DefaultLogger.Warn("could not save workflow version", zap.Uint64("workflowID", wf.ID), zap.Error(err))
	}
}

func saveWorkflowVersion(ctx context.Context, s store.Storer, wf *types.Workflow, note string) (err error) {
	snapshot, err := snapshotWorkflow(wf)
	if err != nil {
		return
	}

	existing, _, err := store.SearchAutomationWorkflowVersions(ctx, s, types.WorkflowVersionFilter{WorkflowID: wf.ID})
	if err != nil {
		return
	}

	latest, latestSnapshot := 0, ""
	for _, v := range existing {
		if v.Version > latest {
			latest, latestSnapshot = v.Version, v.Snapshot
		}
	}

	if latest > 0 && latestSnapshot == string(snapshot) {
		return nil
	}

	if err = store.CreateAutomationWorkflowVersion(ctx, s, &types.WorkflowVersion{
		ID:         id.Next(),
		WorkflowID: wf.ID,
		Version:    latest + 1,
		Snapshot:   string(snapshot),
		Note:       note,
		CreatedAt:  *now(),
		CreatedBy:  auth.GetIdentityFromContext(ctx).Identity(),
	}); err != nil {
		return
	}

	if over := len(existing) + 1 - workflowVersionsKept; over > 0 {
		old := append(types.WorkflowVersionSet(nil), existing...)
		sort.Slice(old, func(i, j int) bool { return old[i].Version < old[j].Version })
		for _, v := range old[:over] {
			if err = store.DeleteAutomationWorkflowVersionByID(ctx, s, v.ID); err != nil {
				return
			}
		}
	}

	return nil
}

func (svc workflowVersion) workflow(ctx context.Context, workflowID uint64, write bool) (*types.Workflow, error) {
	wf, err := loadWorkflow(ctx, svc.store, workflowID)
	if err != nil {
		return nil, err
	}

	if write && !svc.ac.CanUpdateWorkflow(ctx, wf) {
		return nil, WorkflowErrNotAllowedToUpdate()
	}
	if !write && !svc.ac.CanReadWorkflow(ctx, wf) {
		return nil, WorkflowErrNotAllowedToRead()
	}

	return wf, nil
}

// List returns the versions of a workflow, newest first, without snapshots
func (svc workflowVersion) List(ctx context.Context, workflowID uint64) (types.WorkflowVersionSet, error) {
	if _, err := svc.workflow(ctx, workflowID, false); err != nil {
		return nil, err
	}

	set, _, err := store.SearchAutomationWorkflowVersions(ctx, svc.store, types.WorkflowVersionFilter{
		WorkflowID: workflowID,
		Sorting:    filter.Sorting{Sort: filter.SortExprSet{{Column: "version", Descending: true}}},
	})
	if err != nil {
		return nil, err
	}

	for _, v := range set {
		v.Snapshot = ""
	}

	return set, nil
}

// Get returns one version with its snapshot
func (svc workflowVersion) Get(ctx context.Context, workflowID uint64, version int) (*types.WorkflowVersion, error) {
	if _, err := svc.workflow(ctx, workflowID, false); err != nil {
		return nil, err
	}

	set, _, err := store.SearchAutomationWorkflowVersions(ctx, svc.store, types.WorkflowVersionFilter{WorkflowID: workflowID})
	if err != nil {
		return nil, err
	}

	for _, v := range set {
		if v.Version == version {
			return v, nil
		}
	}

	return nil, errors.NotFound("version %d of the workflow was not found", version)
}

// Restore applies the definition of a version to the workflow
func (svc workflowVersion) Restore(ctx context.Context, workflowID uint64, version int) (*types.Workflow, error) {
	cur, err := svc.workflow(ctx, workflowID, true)
	if err != nil {
		return nil, err
	}

	v, err := svc.Get(ctx, workflowID, version)
	if err != nil {
		return nil, err
	}

	var snap workflowSnapshot
	if err = json.Unmarshal([]byte(v.Snapshot), &snap); err != nil {
		return nil, fmt.Errorf("version %d has a damaged snapshot: %v", version, err)
	}

	return DefaultWorkflow.Update(ctx, restoredWorkflow(cur, snap))
}

// restoredWorkflow is the update that puts a snapshot back: the definition
// comes from the snapshot, everything about who and whether stays as it is
func restoredWorkflow(cur *types.Workflow, snap workflowSnapshot) *types.Workflow {
	upd := &types.Workflow{
		ID:           cur.ID,
		Handle:       snap.Handle,
		Meta:         snap.Meta,
		Trace:        snap.Trace,
		KeepSessions: snap.KeepSessions,
		Scope:        snap.Scope,
		Steps:        snap.Steps,
		Paths:        snap.Paths,

		Enabled: cur.Enabled,
		RunAs:   cur.RunAs,
		OwnedBy: cur.OwnedBy,
	}

	if upd.Meta == nil {
		upd.Meta = cur.Meta
	}

	return upd
}
