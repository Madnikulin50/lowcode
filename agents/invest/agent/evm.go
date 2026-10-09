package agent

import (
	"time"

	calcevm "github.com/madnikulin50/lowcode/agents/services/calc-evm"
)

func RecalcEVM(bac, percentComplete, actualCost float64, start, end, now time.Time) EVMResult {
	return evmFrom(calcevm.Recalc(bac, percentComplete, actualCost, start, end, now))
}

func MergeFacts(items []WBSItem, facts []ProgressFact) []WBSItem {
	out := calcevm.MergeFacts(toCalcItems(items), toCalcFacts(facts))
	return applyCalc(items, out)
}

func ApplyEVM(items []WBSItem, now time.Time) []WBSItem {
	return applyCalc(items, calcevm.Apply(toCalcItems(items), now))
}

func AggregateProject(items []WBSItem, projectID string) EVMResult {
	return evmFrom(calcevm.Aggregate(toCalcItems(items), projectID))
}

func runEVM(items []WBSItem, facts []ProgressFact, projectID string, now time.Time) ([]WBSItem, EVMResult) {
	out := calcevm.Run(calcevm.Input{
		Now: now, ProjectID: projectID,
		Items: toCalcItems(items), Facts: toCalcFacts(facts),
	})
	return applyCalc(items, out.Items), evmFrom(out.Project)
}

func toCalcItems(items []WBSItem) []calcevm.Item {
	out := make([]calcevm.Item, len(items))
	for i, it := range items {
		out[i] = calcevm.Item{
			ID: formatUint(it.ID), ProjectID: it.ProjectID,
			BudgetPlanned: it.BudgetPlanned, PercentComplete: it.PercentComplete,
			ActualCost: it.ActualCost, StartPlanned: it.StartPlanned, EndPlanned: it.EndPlanned,
			PV: it.PV, EV: it.EV, SPI: it.SPI, CPI: it.CPI, EAC: it.EAC,
		}
	}
	return out
}

func toCalcFacts(facts []ProgressFact) []calcevm.Fact {
	out := make([]calcevm.Fact, len(facts))
	for i, f := range facts {
		out[i] = calcevm.Fact{WBSID: f.WBSID, Percent: f.Percent, Cost: f.Cost}
	}
	return out
}

func applyCalc(orig []WBSItem, computed []calcevm.Item) []WBSItem {
	byID := map[string]calcevm.Item{}
	for _, it := range computed {
		byID[it.ID] = it
	}
	out := make([]WBSItem, len(orig))
	copy(out, orig)
	for i := range out {
		c, ok := byID[formatUint(out[i].ID)]
		if !ok {
			continue
		}
		out[i].PercentComplete = c.PercentComplete
		out[i].ActualCost = c.ActualCost
		out[i].PV = c.PV
		out[i].EV = c.EV
		out[i].SPI = c.SPI
		out[i].CPI = c.CPI
		out[i].EAC = c.EAC
	}
	return out
}

func evmFrom(r calcevm.Result) EVMResult {
	return EVMResult{PV: r.PV, EV: r.EV, AC: r.AC, SPI: r.SPI, CPI: r.CPI, EAC: r.EAC, BAC: r.BAC}
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
