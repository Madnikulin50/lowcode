// This file is the engine side of the architecture note's Layer C: it turns
// an already-computed AttributionResult into text. Nothing here reads raw
// Evidence — only Attribution and the Assessment's own score/level — so the
// wording always explains a computed number, never reasons about risk from
// scratch.
//
// Both functions below are deterministic templates/heuristics, not live LLM
// calls. That is a deliberate v1 choice, not a placeholder: it makes the
// "Explain reads only breakdown" and "suggest a draft a human edits"
// contracts exercisable and correct without needing this deployment's model
// credentials wired up, and it pins down exactly what a real
// server/pkg/aiagent call receives as context on the day that swap is made
// (see each function's doc comment for where it goes).
package riskengine

import (
	"fmt"
	"sort"
	"strings"
)

// NarrateAttribution renders a short Russian summary of which factors moved
// the score. To upgrade to a live explanation: call the risk-analyst agent
// (server/pkg/aiagent) with `attr` (and a.Level/a.ResidualScore/a.InherentScore)
// as its prompt context in place of building the sentence here — the agent
// should still never receive raw Evidence, only this already-computed
// breakdown.
func NarrateAttribution(a *RiskAssessment, attr *AttributionResult, factors map[uint64]*RiskFactorDef) string {
	if attr == nil {
		return ""
	}
	label := factorLabelByHandle(factors)

	var up, down []FactorContribution
	for _, c := range attr.Contributions {
		if c.Value >= 0 {
			up = append(up, c)
		} else {
			down = append(down, c)
		}
	}
	sort.Slice(up, func(i, j int) bool { return up[i].Value > up[j].Value })
	sort.Slice(down, func(i, j int) bool { return down[i].Value < down[j].Value })

	var b strings.Builder
	fmt.Fprintf(&b, "Уровень риска: %s (residual %.1f, inherent %.1f).", a.Level, a.ResidualScore, a.InherentScore)

	if n := min(len(up), 3); n > 0 {
		parts := make([]string, n)
		for i := 0; i < n; i++ {
			parts[i] = fmt.Sprintf("%s (+%.1f)", label(up[i].Handle), up[i].Value)
		}
		fmt.Fprintf(&b, " Основной вклад: %s.", strings.Join(parts, ", "))
	}
	if n := min(len(down), 2); n > 0 {
		parts := make([]string, n)
		for i := 0; i < n; i++ {
			parts[i] = fmt.Sprintf("%s (%.1f)", label(down[i].Handle), down[i].Value)
		}
		fmt.Fprintf(&b, " Смягчают: %s.", strings.Join(parts, ", "))
	}
	return b.String()
}

func factorLabelByHandle(factors map[uint64]*RiskFactorDef) func(handle string) string {
	byHandle := make(map[string]*RiskFactorDef, len(factors))
	for _, fd := range factors {
		byHandle[fd.Handle] = fd
	}
	return func(handle string) string {
		if fd, ok := byHandle[handle]; ok && fd.Label != "" {
			return fd.Label
		}
		return handle
	}
}
