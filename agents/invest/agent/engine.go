package agent

import (
	"context"
	"fmt"
	"time"
)

type Engine struct {
	cfg Config
	cz  *Corteza
}

func NewEngine(cfg Config) *Engine {
	return &Engine{
		cfg: cfg,
		cz:  NewCorteza(cfg.CortezaAPI, cfg.Token, cfg.NamespaceID, cfg.HTTPTimeout),
	}
}

func (e *Engine) client(req JobRequest) *Corteza {
	cz := e.cz.WithToken(req.Token).WithNamespace(uint64(req.NamespaceID))
	return cz
}

type RecalcResponse struct {
	WBS       int       `json:"wbs"`
	Project   EVMResult `json:"project"`
	Updated   int       `json:"updated"`
	ProjectID string    `json:"projectID,omitempty"`
}

func (e *Engine) RecalculateEVM(ctx context.Context, req JobRequest) (*RecalcResponse, error) {
	req.Normalize()
	cz := e.client(req)
	items, err := cz.LoadWBS(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("load wbs: %w", err)
	}
	facts, err := cz.LoadFacts(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("load facts: %w", err)
	}
	now := time.Now()
	items = MergeFacts(items, facts)
	items = ApplyEVM(items, now)
	if err := cz.SaveWBSMetrics(ctx, items); err != nil {
		return nil, err
	}
	agg := AggregateProject(items, req.ProjectID)
	if req.ProjectID != "" && req.ProjectID != "0" {
		_ = cz.SaveProjectEVM(ctx, req.ProjectID, agg)
	} else {
		seen := map[string][]WBSItem{}
		for _, it := range items {
			if it.ProjectID == "" {
				continue
			}
			seen[it.ProjectID] = append(seen[it.ProjectID], it)
		}
		for pid, group := range seen {
			_ = cz.SaveProjectEVM(ctx, pid, AggregateProject(group, pid))
		}
	}
	return &RecalcResponse{WBS: len(items), Project: agg, Updated: len(items), ProjectID: req.ProjectID}, nil
}

type PathResponse struct {
	WBS      int `json:"wbs"`
	Critical int `json:"critical"`
}

func (e *Engine) CriticalPath(ctx context.Context, req JobRequest) (*PathResponse, error) {
	req.Normalize()
	cz := e.client(req)
	items, err := cz.LoadWBS(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("load wbs: %w", err)
	}
	acts := ComputeCriticalPath(WBSToActivities(items))
	items = ApplyCritical(items, acts)
	if err := cz.SaveWBSCritical(ctx, items); err != nil {
		return nil, err
	}
	n := 0
	for _, it := range items {
		if it.IsCritical {
			n++
		}
	}
	return &PathResponse{WBS: len(items), Critical: n}, nil
}

type AlertsResponse struct {
	Alerts  int     `json:"alerts"`
	Created int     `json:"created"`
	Items   []Alert `json:"items"`
}

func (e *Engine) Alerts(ctx context.Context, req JobRequest) (*AlertsResponse, error) {
	req.Normalize()
	cz := e.client(req)
	docs, err := cz.LoadDocuments(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("load documents: %w", err)
	}
	items, err := cz.LoadWBS(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("load wbs: %w", err)
	}
	now := time.Now()
	items = ApplyEVM(items, now)
	alerts := CollectAlerts(docs, items, now, req.CPIThreshold)
	created := 0
	for _, a := range alerts {
		if err := cz.EnsureRisk(ctx, a); err != nil {
			continue
		}
		created++
	}
	return &AlertsResponse{Alerts: len(alerts), Created: created, Items: alerts}, nil
}
