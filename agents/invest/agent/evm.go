package agent

import "time"

// RecalcEVM computes PV, EV, SPI, CPI, EAC for a work package.
//
// BAC = budget_planned
// PV  = BAC * plannedPercent(now, start, end)   — linear time-phased
// EV  = BAC * (percent_complete / 100)
// AC  = actual_cost
// SPI = EV / PV  (1 if PV==0)
// CPI = EV / AC  (1 if AC==0)
// EAC = AC + (BAC-EV)/CPI  (BAC if CPI==0)
func RecalcEVM(bac, percentComplete, actualCost float64, start, end, now time.Time) EVMResult {
	if percentComplete < 0 {
		percentComplete = 0
	}
	if percentComplete > 100 {
		percentComplete = 100
	}
	pv := bac * plannedFraction(start, end, now)
	ev := bac * (percentComplete / 100)
	ac := actualCost
	spi := 1.0
	if pv != 0 {
		spi = ev / pv
	}
	cpi := 1.0
	if ac != 0 {
		cpi = ev / ac
	}
	eac := bac
	if cpi != 0 {
		eac = ac + (bac-ev)/cpi
	}
	return EVMResult{PV: pv, EV: ev, AC: ac, SPI: spi, CPI: cpi, EAC: eac, BAC: bac}
}

func plannedFraction(start, end, now time.Time) float64 {
	if start.IsZero() || end.IsZero() || !end.After(start) {
		if now.Before(start) {
			return 0
		}
		return 1
	}
	if now.Before(start) {
		return 0
	}
	if !now.Before(end) {
		return 1
	}
	total := end.Sub(start).Seconds()
	if total <= 0 {
		return 1
	}
	return now.Sub(start).Seconds() / total
}

// MergeFacts overlays progress facts onto WBS percent/AC when facts exist.
func MergeFacts(items []WBSItem, facts []ProgressFact) []WBSItem {
	pct := map[string]float64{}
	cost := map[string]float64{}
	for _, f := range facts {
		if f.WBSID == "" {
			continue
		}
		if f.Percent > pct[f.WBSID] {
			pct[f.WBSID] = f.Percent
		}
		cost[f.WBSID] += f.Cost
	}
	out := make([]WBSItem, len(items))
	copy(out, items)
	for i := range out {
		id := formatUint(out[i].ID)
		if p, ok := pct[id]; ok && p > out[i].PercentComplete {
			out[i].PercentComplete = p
		}
		if c, ok := cost[id]; ok && c > 0 {
			out[i].ActualCost = c
		}
	}
	return out
}

func ApplyEVM(items []WBSItem, now time.Time) []WBSItem {
	out := make([]WBSItem, len(items))
	copy(out, items)
	for i := range out {
		r := RecalcEVM(out[i].BudgetPlanned, out[i].PercentComplete, out[i].ActualCost, out[i].StartPlanned, out[i].EndPlanned, now)
		out[i].PV = r.PV
		out[i].EV = r.EV
		out[i].SPI = r.SPI
		out[i].CPI = r.CPI
		out[i].EAC = r.EAC
	}
	return out
}

func AggregateProject(items []WBSItem, projectID string) EVMResult {
	var bac, pv, ev, ac float64
	for _, it := range items {
		if projectID != "" && it.ProjectID != projectID {
			continue
		}
		bac += it.BudgetPlanned
		pv += it.PV
		ev += it.EV
		ac += it.ActualCost
	}
	spi := 1.0
	if pv != 0 {
		spi = ev / pv
	}
	cpi := 1.0
	if ac != 0 {
		cpi = ev / ac
	}
	eac := bac
	if cpi != 0 {
		eac = ac + (bac-ev)/cpi
	}
	return EVMResult{PV: pv, EV: ev, AC: ac, SPI: spi, CPI: cpi, EAC: eac, BAC: bac}
}

func formatUint(id uint64) string {
	if id == 0 {
		return ""
	}
	const digits = "0123456789"
	var buf [20]byte
	i := len(buf)
	for id > 0 {
		i--
		buf[i] = digits[id%10]
		id /= 10
	}
	return string(buf[i:])
}
