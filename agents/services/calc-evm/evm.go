package calcevm

import "time"

// Recalc computes PV, EV, SPI, CPI, EAC for a work package.
//
// BAC = budget_planned
// PV  = BAC * plannedPercent(now, start, end)   — linear time-phased
// EV  = BAC * (percent_complete / 100)
// AC  = actual_cost
// SPI = EV / PV  (1 if PV==0)
// CPI = EV / AC  (1 if AC==0)
// EAC = AC + (BAC-EV)/CPI  (BAC if CPI==0)
func Recalc(bac, percentComplete, actualCost float64, start, end, now time.Time) Result {
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
	return Result{PV: pv, EV: ev, AC: ac, SPI: spi, CPI: cpi, EAC: eac, BAC: bac}
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
func MergeFacts(items []Item, facts []Fact) []Item {
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
	out := make([]Item, len(items))
	copy(out, items)
	for i := range out {
		id := out[i].ID
		if p, ok := pct[id]; ok && p > out[i].PercentComplete {
			out[i].PercentComplete = p
		}
		if c, ok := cost[id]; ok && c > 0 {
			out[i].ActualCost = c
		}
	}
	return out
}

func Apply(items []Item, now time.Time) []Item {
	out := make([]Item, len(items))
	copy(out, items)
	for i := range out {
		r := Recalc(out[i].BudgetPlanned, out[i].PercentComplete, out[i].ActualCost, out[i].StartPlanned, out[i].EndPlanned, now)
		out[i].PV = r.PV
		out[i].EV = r.EV
		out[i].SPI = r.SPI
		out[i].CPI = r.CPI
		out[i].EAC = r.EAC
	}
	return out
}

func Aggregate(items []Item, projectID string) Result {
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
	return Result{PV: pv, EV: ev, AC: ac, SPI: spi, CPI: cpi, EAC: eac, BAC: bac}
}

func Run(in Input) Output {
	now := in.Now
	if now.IsZero() {
		now = time.Now()
	}
	items := in.Items
	facts := in.Facts
	if in.ProjectID != "" {
		items = filterItems(items, in.ProjectID)
		facts = filterFacts(facts, in.ProjectID, items)
	}
	items = MergeFacts(items, facts)
	items = Apply(items, now)
	agg := Aggregate(items, in.ProjectID)
	return Output{
		WBS: len(items), Updated: len(items), ProjectID: in.ProjectID,
		Project: agg, Projects: rollupByProject(items),
		SPI: agg.SPI, CPI: agg.CPI, EAC: agg.EAC, AC: agg.AC,
		Items: items,
	}
}

func filterItems(items []Item, projectID string) []Item {
	out := make([]Item, 0, len(items))
	for _, it := range items {
		if it.ProjectID == projectID {
			out = append(out, it)
		}
	}
	return out
}

func filterFacts(facts []Fact, projectID string, items []Item) []Fact {
	ids := map[string]struct{}{}
	for _, it := range items {
		ids[it.ID] = struct{}{}
	}
	out := make([]Fact, 0, len(facts))
	for _, f := range facts {
		if f.ProjectID == projectID {
			out = append(out, f)
			continue
		}
		if _, ok := ids[f.WBSID]; ok {
			out = append(out, f)
		}
	}
	return out
}

func rollupByProject(items []Item) []ProjectRollup {
	order := make([]string, 0)
	groups := map[string][]Item{}
	for _, it := range items {
		if it.ProjectID == "" {
			continue
		}
		if _, ok := groups[it.ProjectID]; !ok {
			order = append(order, it.ProjectID)
		}
		groups[it.ProjectID] = append(groups[it.ProjectID], it)
	}
	out := make([]ProjectRollup, 0, len(order))
	for _, pid := range order {
		out = append(out, ProjectRollup{ProjectID: pid, Result: Aggregate(groups[pid], pid)})
	}
	return out
}
