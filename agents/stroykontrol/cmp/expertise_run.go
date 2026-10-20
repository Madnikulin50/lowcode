package cmp

import (
	"encoding/json"
	"strings"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// ModExpertise is the module of imported expertise remarks.
const ModExpertise = "expertise_remarks"

// CandidatesFrom turns stored discrepancy records into match candidates.
func CandidatesFrom(recs []idcheck.Record) []Candidate {
	out := make([]Candidate, 0, len(recs))
	for _, d := range recs {
		var ev Evidence
		_ = json.Unmarshal([]byte(d.Get("evidence")), &ev)
		text := strings.TrimPrefix(strings.TrimPrefix(d.Get("description"), "[детерм.] "), "[ИИ] ")
		out = append(out, Candidate{
			ID:        d.ID,
			Text:      text + " " + d.Get("anchor_text") + " " + ev.Ref + " " + ev.Target,
			IterState: d.Get("iter_state"),
			Severity:  d.Get("severity"),
		})
	}
	return out
}

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if s != "" {
			return s
		}
	}
	return ""
}
