package agent

import (
	"context"
	"fmt"
	"time"
)

type CloneResult struct {
	ProjectID string `json:"projectID"`
	Created   int    `json:"created"`
	Skipped   int    `json:"skipped"`
	Message   string `json:"message"`
}

type WBSTemplate struct {
	ID              uint64
	Type            string
	Code            string
	Name            string
	Level           string
	ParentCode      string
	PredecessorCode string
	BudgetPlanned   float64
	DurationDays    float64
}

func (e *Engine) CloneWBSTemplate(ctx context.Context, req JobRequest) (*CloneResult, error) {
	req.Normalize()
	cz := e.client(req)
	projectID := req.ProjectID
	if projectID == "" {
		projectID = req.RecordID
	}
	pid := ParseID(projectID)
	if pid == 0 {
		return nil, fmt.Errorf("projectID required")
	}
	proj, err := cz.LoadProject(ctx, pid)
	if err != nil {
		return nil, err
	}
	typeID := proj.ConstructionType
	if typeID == "" || typeID == "0" {
		return nil, fmt.Errorf("у проекта не задан тип конструкции — выберите его в карточке")
	}
	tpls, err := cz.LoadWBSTemplates(ctx, typeID)
	if err != nil {
		return nil, err
	}
	if len(tpls) == 0 {
		return &CloneResult{ProjectID: formatUint(pid), Message: "Нет шаблонов WBS для этого типа конструкции"}, nil
	}
	existing, _ := cz.LoadWBS(ctx, formatUint(pid))
	have := map[string]uint64{}
	for _, it := range existing {
		have[it.Code] = it.ID
	}
	created, skipped := 0, 0
	start := time.Now()
	if !proj.EndPlanned.IsZero() {
		start = time.Now()
	}
	cursor := start
	ids := map[string]uint64{}
	for k, v := range have {
		ids[k] = v
	}
	for _, t := range tpls {
		if _, ok := have[t.Code]; ok {
			skipped++
			continue
		}
		dur := t.DurationDays
		if dur <= 0 {
			dur = 30
		}
		end := cursor.Add(time.Duration(dur) * 24 * time.Hour)
		vals := compactValues(map[string]string{
			"project":          formatUint(pid),
			"code":             t.Code,
			"name":             t.Name,
			"level":            firstNonEmpty(t.Level, "work"),
			"budget_planned":   fmtNum(t.BudgetPlanned),
			"start_planned":    cursor.Format("2006-01-02"),
			"end_planned":      end.Format("2006-01-02"),
			"percent_complete": "0",
		})
		nid, err := cz.CreateValues(ctx, "wbs_items", vals)
		if err != nil {
			return nil, fmt.Errorf("create wbs %s: %w", t.Code, err)
		}
		ids[t.Code] = nid
		have[t.Code] = nid
		created++
		cursor = end
	}
	for _, t := range tpls {
		id := ids[t.Code]
		if id == 0 {
			continue
		}
		patch := map[string]string{}
		if t.ParentCode != "" {
			if pid := ids[t.ParentCode]; pid != 0 {
				patch["parent"] = formatUint(pid)
			}
		}
		if t.PredecessorCode != "" {
			if pid := ids[t.PredecessorCode]; pid != 0 {
				patch["predecessor"] = formatUint(pid)
			}
		}
		if len(patch) > 0 {
			_ = cz.UpdateValues(ctx, "wbs_items", id, patch)
		}
	}
	return &CloneResult{
		ProjectID: formatUint(pid),
		Created:   created,
		Skipped:   skipped,
		Message:   fmt.Sprintf("Создано %d элементов WBS, пропущено %d (уже были)", created, skipped),
	}, nil
}
