package agent

import (
	"context"
	"fmt"

	"github.com/madnikulin50/lowcode/agents/sdk"
)

func (s *ScanStatus) Envelope(kind sdk.EventKind, recordID, namespaceID string, items []Device) *sdk.Envelope {
	if s == nil {
		return nil
	}
	st := sdk.NormalizeStatus(s.Status)
	if kind == "" {
		kind = sdk.KindFromStatus(st)
	}
	found := s.Found
	if found == 0 && len(items) > 0 {
		found = len(items)
	}
	var itemAny any
	if items != nil {
		itemAny = items
	} else if s.Items != nil {
		itemAny = s.Items
	}
	return &sdk.Envelope{
		Schema:      sdk.SchemaV1,
		Service:     "cmdb",
		Operation:   "scan",
		ID:          s.ID,
		JobID:       s.ID,
		Kind:        kind,
		Status:      st,
		Progress:    s.Progress,
		Message:     s.Message,
		Error:       firstNonEmpty(s.Error, s.Message),
		NamespaceID: namespaceID,
		RecordID:    recordID,
		StartedAt:   s.StartedAt,
		FinishedAt:  s.FinishedAt,
		Result: map[string]any{
			"found":      found,
			"scanningIP": s.ScanningIP,
			"target":     s.Target,
			"totalIPs":   s.TotalIPs,
			"scannedIPs": s.ScannedIPs,
			"cidrs":      s.CIDRs,
			"moduleID":   s.ModuleID,
		},
		Items: itemAny,
	}
}

func (a *Agent) StartJob(ctx context.Context, req sdk.StartRequest) (*sdk.Envelope, error) {
	target := ScanTarget{
		CIDR:         req.Param("cidr"),
		NamespaceID:  FlexUint64(req.NamespaceID.Uint64()),
		ModuleID:     FlexUint64(req.ParamID("moduleID")),
		Token:        req.Token,
		API:          req.Param("api"),
		ScanRecordID: firstNonEmpty(req.RecordID, req.JobID, req.Param("scanRecordID")),
		CallbackURL:  req.CallbackURL,
	}
	if target.CIDR == "" {
		return nil, fmt.Errorf("cidr is required")
	}
	st, err := a.StartScan(ctx, target)
	if err != nil {
		return nil, err
	}
	return st.Envelope(sdk.KindProgress, target.ScanRecordID, req.NamespaceID.String(), nil), nil
}

func (a *Agent) GetJob(id string) *sdk.Envelope {
	s := a.GetStatus(id)
	if s == nil {
		return nil
	}
	return s.Envelope("", "", "", s.Items)
}

func (a *Agent) ListJobs() []*sdk.Envelope {
	scans := a.ListScans()
	out := make([]*sdk.Envelope, 0, len(scans))
	for i := range scans {
		out = append(out, scans[i].Envelope("", "", "", scans[i].Items))
	}
	return out
}

func (a *Agent) JobItems(id string) any {
	s := a.GetStatus(id)
	if s == nil {
		return nil
	}
	if s.Items == nil {
		return []Device{}
	}
	return s.Items
}

func (a *Agent) Health(_ context.Context) map[string]any {
	return map[string]any{"status": "ok", "service": "cmdb"}
}
