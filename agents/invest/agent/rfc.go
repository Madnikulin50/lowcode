package agent

import (
	"context"
	"fmt"
	"time"
)

type SimulateResult struct {
	RFCID     string  `json:"rfcID"`
	EACBefore float64 `json:"eac_before"`
	EACAfter  float64 `json:"eac_after"`
	EndAfter  string  `json:"end_after,omitempty"`
	Simulated bool    `json:"simulated"`
	Message   string  `json:"message"`
	SPI       float64 `json:"spi,omitempty"`
	CPI       float64 `json:"cpi,omitempty"`
	EAC       float64 `json:"eac,omitempty"`
	AC        float64 `json:"ac,omitempty"`
}

// SimulateImpact computes EAC before/after and new planned finish (FR-030).
func SimulateImpact(projectEAC, budgetPlanned, deltaBudget float64, end time.Time, deltaDays float64) (eacBefore, eacAfter float64, endAfter time.Time) {
	eacBefore = projectEAC
	if eacBefore == 0 {
		eacBefore = budgetPlanned
	}
	eacAfter = eacBefore + deltaBudget
	endAfter = end
	if !end.IsZero() && deltaDays != 0 {
		endAfter = end.Add(time.Duration(deltaDays) * 24 * time.Hour)
	}
	return
}

func ApplyBudgetDelta(oldBudget, delta float64) float64 {
	n := oldBudget + delta
	if n < 0 {
		return 0
	}
	return n
}

func (e *Engine) SimulateRFC(ctx context.Context, req JobRequest) (*SimulateResult, error) {
	req.Normalize()
	cz := e.client(req)
	id := ParseID(req.RecordID)
	if id == 0 {
		return nil, fmt.Errorf("recordID (RFC) required")
	}
	rfc, err := cz.LoadRFC(ctx, id)
	if err != nil {
		return nil, err
	}
	proj, err := cz.LoadProject(ctx, ParseID(rfc.Project))
	if err != nil {
		return nil, err
	}
	before, after, endAfter := SimulateImpact(proj.EAC, proj.BudgetPlanned, rfc.DeltaBudget, proj.EndPlanned, rfc.DeltaDays)
	vals := compactValues(map[string]string{
		"eac_before": fmtNum(before),
		"eac_after":  fmtNum(after),
		"simulated":  "1",
	})
	if !endAfter.IsZero() {
		vals["end_after"] = endAfter.Format("2006-01-02")
	}
	if err := cz.UpdateValues(ctx, "change_requests", rfc.ID, vals); err != nil {
		return nil, err
	}
	return &SimulateResult{
		RFCID:     formatUint(rfc.ID),
		EACBefore: before,
		EACAfter:  after,
		EndAfter:  vals["end_after"],
		Simulated: true,
		Message:   fmt.Sprintf("Если утвердить RFC, EAC %.0f → %.0f", before, after),
		EAC:       after,
	}, nil
}

func (e *Engine) ApproveRFC(ctx context.Context, req JobRequest) (*SimulateResult, error) {
	req.Normalize()
	cz := e.client(req)
	id := ParseID(req.RecordID)
	if id == 0 {
		return nil, fmt.Errorf("recordID (RFC) required")
	}
	rfc, err := cz.LoadRFC(ctx, id)
	if err != nil {
		return nil, err
	}
	if !rfc.Simulated && rfc.EACAfter == 0 && rfc.EACBefore == 0 {
		return nil, fmt.Errorf("сначала выполните симуляцию EAC (кнопка «Симулировать»)")
	}
	proj, err := cz.LoadProject(ctx, ParseID(rfc.Project))
	if err != nil {
		return nil, err
	}
	oldBudget := proj.BudgetPlanned
	newBudget := ApplyBudgetDelta(oldBudget, rfc.DeltaBudget)
	oldEnd := proj.EndPlanned
	_, _, newEnd := SimulateImpact(proj.EAC, proj.BudgetPlanned, rfc.DeltaBudget, proj.EndPlanned, rfc.DeltaDays)
	if rfc.EACBefore == 0 {
		rfc.EACBefore, rfc.EACAfter, newEnd = SimulateImpact(proj.EAC, proj.BudgetPlanned, rfc.DeltaBudget, proj.EndPlanned, rfc.DeltaDays)
	}

	projPatch := compactValues(map[string]string{
		"budget_planned": fmtNum(newBudget),
	})
	if !newEnd.IsZero() {
		projPatch["end_planned"] = newEnd.Format("2006-01-02")
	}
	if err := cz.UpdateValues(ctx, "projects", proj.ID, projPatch); err != nil {
		return nil, fmt.Errorf("update project: %w", err)
	}
	if err := cz.UpdateValues(ctx, "change_requests", rfc.ID, compactValues(map[string]string{
		"status":     "approved",
		"eac_before": fmtNum(rfc.EACBefore),
		"eac_after":  fmtNum(rfc.EACAfter),
		"simulated":  "1",
	})); err != nil {
		return nil, err
	}
	_, _ = cz.CreateValues(ctx, "change_log", compactValues(map[string]string{
		"rfc":        formatUint(rfc.ID),
		"project":    rfc.Project,
		"summary":    "Утверждён RFC: " + rfc.Title,
		"old_budget": fmtNum(oldBudget),
		"new_budget": fmtNum(newBudget),
		"old_end":    dateOnly(oldEnd),
		"new_end":    dateOnly(newEnd),
		"actor":      req.UserID,
		"changed_at": fmtDate(time.Now()),
	}))
	evm, err := e.RecalculateEVM(ctx, JobRequest{NamespaceID: req.NamespaceID, ProjectID: rfc.Project, Token: req.Token})
	res := &SimulateResult{
		RFCID:     formatUint(rfc.ID),
		EACBefore: rfc.EACBefore,
		EACAfter:  rfc.EACAfter,
		Simulated: true,
		Message:   "RFC утверждён, baseline обновлён, EVM пересчитан",
	}
	if evm != nil {
		res.SPI, res.CPI, res.EAC, res.AC = evm.SPI, evm.CPI, evm.EAC, evm.AC
	}
	if err != nil {
		res.Message += " (EVM: " + err.Error() + ")"
	}
	return res, nil
}

func (e *Engine) RejectRFC(ctx context.Context, req JobRequest) (*SimulateResult, error) {
	req.Normalize()
	cz := e.client(req)
	id := ParseID(req.RecordID)
	if id == 0 {
		return nil, fmt.Errorf("recordID (RFC) required")
	}
	if err := cz.UpdateValues(ctx, "change_requests", id, map[string]string{"status": "rejected"}); err != nil {
		return nil, err
	}
	return &SimulateResult{RFCID: formatUint(id), Message: "RFC отклонён"}, nil
}

func dateOnly(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func boolish(s string) bool {
	switch s {
	case "1", "true", "TRUE", "yes", "on":
		return true
	}
	return false
}
