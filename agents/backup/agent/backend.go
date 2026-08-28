package agent

import (
	"context"
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
	return &sdk.Envelope{
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
			"bytesRead":    st.BytesRead,
			"bytesWritten": st.BytesWritten,
			"files":        st.Files,
			"sourceID":     st.SourceID,
			"policyID":     st.PolicyID,
			"s3Bucket":     st.S3Bucket,
			"s3Key":        st.S3Key,
			"checksum":     st.Checksum,
			"engine":       st.Engine,
			"resticID":     st.ResticID,
			"snapshotID":   st.SnapshotID,
			"kind":         st.Kind,
		},
	}
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
		DestPath:      req.Param("destPath"),
		DestType:      req.Param("destType"),
		Kind:          firstNonEmpty(req.Operation, req.Param("kind")),
		RetentionDays: int(req.ParamID("retentionDays")),
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
	case "due":
		list, err := a.RunDue(ctx, jr)
		if err != nil {
			return nil, err
		}
		if len(list) == 0 {
			return &sdk.Envelope{Service: "backup", Operation: "due", Status: sdk.StatusCompleted, Kind: sdk.KindComplete}, nil
		}
		return list[0].Envelope(sdk.KindProgress), nil
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
		return a.RunDue(ctx, jr)
	default:
		env, err := a.StartJob(ctx, req)
		return env, err
	}
}
