package agent

import (
	"context"
	"fmt"
	"strings"
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

func (e *Engine) Discover(ctx context.Context) error {
	return e.cz.Discover(ctx)
}

func (e *Engine) APIOrigin() string {
	return e.cz.BaseURL()
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
	Saved     int       `json:"saved,omitempty"`
	SaveError string    `json:"saveError,omitempty"`
	// Flat copies so rule-chain templates {{spi}}/{{cpi}}/{{eac}}/{{ac}}
	// resolve after an HTTP node (promote copies top-level JSON keys).
	SPI float64 `json:"spi"`
	CPI float64 `json:"cpi"`
	EAC float64 `json:"eac"`
	AC  float64 `json:"ac"`
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
	resp := &RecalcResponse{
		WBS: len(items), Project: agg, Updated: len(items), ProjectID: req.ProjectID,
		SPI: agg.SPI, CPI: agg.CPI, EAC: agg.EAC, AC: agg.AC,
	}
	if req.ProjectID != "" && req.ProjectID != "0" {
		// Card / API with a project: engine writes the aggregate as a fallback.
		// The rule chain also PATCHes the same fields after HTTP succeeds.
		if err := cz.SaveProjectEVM(ctx, req.ProjectID, agg); err != nil {
			resp.SaveError = err.Error()
		} else {
			resp.Saved = 1
		}
	} else {
		seen := map[string][]WBSItem{}
		for _, it := range items {
			if it.ProjectID == "" {
				continue
			}
			seen[it.ProjectID] = append(seen[it.ProjectID], it)
		}
		var errs []string
		for pid, group := range seen {
			if err := cz.SaveProjectEVM(ctx, pid, AggregateProject(group, pid)); err != nil {
				errs = append(errs, fmt.Sprintf("%s: %v", pid, err))
				continue
			}
			resp.Saved++
		}
		if len(errs) > 0 {
			resp.SaveError = strings.Join(errs, "; ")
		}
	}
	return resp, nil
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
