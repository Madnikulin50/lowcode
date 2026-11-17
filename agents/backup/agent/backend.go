package agent

import (
	"context"
	"fmt"
	"strconv"

	"github.com/madnikulin50/lowcode/agents/sdk"
)

func (st *JobStatus) Envelope(kind sdk.EventKind) *sdk.Envelope {
	if st == nil {
		return nil
	}
	op := st.Kind
	if op == "" || op == string(KindFull) || op == string(KindIncremental) {
		op = "backup"
	}
	ns := ""
	if st.NamespaceID > 0 {
		ns = strconv.FormatUint(st.NamespaceID, 10)
	}
	env := &sdk.Envelope{
		Schema:      sdk.SchemaV1,
		Service:     "backup",
		Operation:   op,
		ID:          st.ID,
		JobID:       st.ID,
		Kind:        kind,
		Status:      sdk.NormalizeStatus(string(st.Status)),
		Progress:    st.Progress,
		Message:     st.Message,
		Error:       st.Error,
		NamespaceID: ns,
		RecordID:    st.JobRecordID,
		StartedAt:   st.StartedAt,
		FinishedAt:  st.FinishedAt,
		Result: map[string]any{
			"bytesRead":     st.BytesRead,
			"bytesWritten":  st.BytesWritten,
			"files":         st.Files,
			"sourceID":      st.SourceID,
			"policyID":      st.PolicyID,
			"s3Endpoint":    st.S3Endpoint,
			"s3Bucket":      st.S3Bucket,
			"s3Key":         st.S3Key,
			"checksum":      st.Checksum,
			"engine":        st.Engine,
			"resticID":      st.ResticID,
			"snapshotID":    st.SnapshotID,
			"kind":          st.Kind,
			"extractInline": st.ExtractInline,
		},
	}
	for k, v := range st.Summary {
		if _, ok := env.Result[k]; !ok {
			env.Result[k] = v
		}
	}
	return env
}

func jobRequestFrom(req sdk.StartRequest) JobRequest {
	return JobRequest{
		SourceID:      req.Param("sourceID"),
		Source:        req.Param("source"),
		PolicyID:      req.Param("policyID"),
		JobRecordID:   firstNonEmpty(req.RecordID, req.JobID, req.Param("jobID")),
		RestoreID:     firstNonEmpty(req.Param("restoreID"), req.RecordID),
		SnapshotID:    req.Param("snapshotID"),
		NamespaceID:   req.NamespaceID.String(),
		Token:         req.Token,
		CallbackURL:   req.CallbackURL,
		DestPath:      firstNonEmpty(req.Param("destPath"), req.Param("path")),
		DestType:      req.Param("destType"),
		Kind:          firstNonEmpty(req.Operation, req.Param("kind")),
		RetentionDays: int(req.ParamID("retentionDays")),
		VerifyLevel:   firstNonEmpty(req.Param("verifyLevel"), req.Param("level")),
		Namespaces:    req.Param("namespaces"),
		ExtractTo:     req.Param("extractTo"),
		Basis:         req.Param("basis"),
		DestinationID: firstNonEmpty(req.Param("destinationID"), req.Param("destination")),
		ScopeID:       req.Param("scopeID"),
	}
}

func (a *Agent) StartJob(ctx context.Context, req sdk.StartRequest) (*sdk.Envelope, error) {
	jr := jobRequestFrom(req)
	op := jr.Kind
	var (
		st  *JobStatus
		err error
	)
	switch op {
	case "restore":
		st, err = a.StartRestore(ctx, jr)
	case "prune":
		st, err = a.StartPrune(ctx, jr)
	case "vault":
		st, err = a.StartVault(ctx, jr)
	case "scan":
		st, err = a.StartScan(ctx, jr)
	case "verify":
		st, err = a.StartVerify(ctx, jr)
	case "extract":
		st, err = a.StartExtract(ctx, jr)
	case "manifest":
		st, err = a.StartArchiveManifest(ctx, jr)
	case "destroy":
		st, err = a.StartDestroy(ctx, jr)
	case "due":
		// "due" only lists due policies now — see Agent.DuePolicies — so it
		// doesn't fit the single-job Envelope shape StartJob returns.
		// svc.Sync("due") routes real traffic through Call below; this path
		// only exists for callers that hit StartJob directly.
		list, err := a.DuePolicies(ctx, jr)
		if err != nil {
			return nil, err
		}
		return &sdk.Envelope{
			Service:   "backup",
			Operation: "due",
			Status:    sdk.StatusCompleted,
			Kind:      sdk.KindComplete,
			Result:    map[string]any{"due": list},
		}, nil
	case "audit":
		// Фоновый прогон: обход областей пишет находки в audit_findings,
		// сводка — в JobStatus.Summary (видна по GET /api/jobs/{id}).
		st, err := a.StartAudit(ctx, jr)
		if err != nil {
			return nil, err
		}
		return st.Envelope(sdk.KindProgress), nil
	default:
		st, err = a.StartBackup(ctx, jr)
	}
	if err != nil {
		return nil, err
	}
	return st.Envelope(sdk.KindProgress), nil
}

func (a *Agent) GetJob(id string) *sdk.Envelope {
	st := a.GetStatus(id)
	if st == nil {
		return nil
	}
	return st.Envelope(sdk.KindFromStatus(sdk.NormalizeStatus(string(st.Status))))
}

func (a *Agent) ListJobs() []*sdk.Envelope {
	list := a.ListJobsStatus()
	out := make([]*sdk.Envelope, 0, len(list))
	for _, st := range list {
		out = append(out, st.Envelope(sdk.KindFromStatus(sdk.NormalizeStatus(string(st.Status)))))
	}
	return out
}

func (a *Agent) Call(ctx context.Context, operation string, req sdk.StartRequest) (any, error) {
	jr := jobRequestFrom(req)
	switch operation {
	case "due":
		return a.DuePolicies(ctx, jr)
	case "clean-point":
		return a.CleanPoint(ctx, jr)
	case "browse":
		return a.Browse(ctx, jr)
	case "audit":
		return a.Audit(ctx, jr)
	case "compliance":
		return a.Compliance(ctx, jr)
	case "verify-due":
		return a.VerifyDue(ctx, jr)
	case "dest-test":
		return a.DestTest(ctx, jr)
	case "reconcile":
		// On-demand cleanup for "jobs"/"restores" records stuck "running"
		// from an earlier process — see ReconcileStaleJobs. Uses the
		// caller's own token (like restore/prune), so it works even when
		// this agent has no static token of its own configured.
		cz := a.Corteza(jr)
		if cz == nil {
			return nil, fmt.Errorf("corteza is not configured")
		}
		n, err := a.ReconcileStaleJobs(ctx, cz)
		if err != nil {
			return nil, err
		}
		return map[string]any{"reconciled": n}, nil
	default:
		env, err := a.StartJob(ctx, req)
		return env, err
	}
}
