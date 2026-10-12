package cmp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// ModExpertise is the module of imported expertise remarks.
const ModExpertise = "expertise_remarks"

// ImportSummary is what an import reports back.
type ImportSummary struct {
	Remarks       int `json:"remarks"`
	StillDiffers  int `json:"stillDiffers"`
	Resolved      int `json:"resolved"`
	NoMatch       int `json:"noMatch"`
	ClaimMismatch int `json:"claimMismatch"`
}

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

// ImportRemarks reads the expertise tables attached to the comparison
// (field remarks_file: the expert's remarks and, possibly in another file,
// the designer's answers), matches each remark with the comparison's
// findings and writes the result into expertise_remarks. A re-import
// updates the remarks by number and keeps the specialist's own status.
func (r *Runner) ImportRemarks(ctx context.Context, ns, comparisonID string) (*ImportSummary, error) {
	b := r.Backend
	rec, err := b.GetRecord(ns, ModComparisons, comparisonID)
	if err != nil {
		return nil, err
	}
	atts := rec.Values["remarks_file"]
	if len(atts) == 0 {
		return nil, fmt.Errorf("к сравнению не приложен Excel с замечаниями (поле «Замечания экспертизы»)")
	}
	var lists [][]Expert
	for _, att := range atts {
		data, name, err := b.Download(ns, att)
		if err != nil {
			return nil, fmt.Errorf("загрузка файла замечаний: %w", err)
		}
		list, err := ParseExpertise(data, name)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		lists = append(lists, list)
	}
	experts := MergeExperts(lists...)

	discs, err := b.Search(ns, ModDiscrepancies, fmt.Sprintf("comparison = '%s'", comparisonID))
	if err != nil {
		return nil, err
	}
	matches := MatchRemarks(experts, CandidatesFrom(discs))

	old, err := b.Search(ns, ModExpertise, fmt.Sprintf("comparison = '%s'", comparisonID))
	if err != nil {
		return nil, err
	}
	byNumber := map[string]idcheck.Record{}
	for _, o := range old {
		byNumber[norm(o.Get("number"))] = o
	}
	sum := &ImportSummary{Remarks: len(experts)}
	seen := map[string]bool{}
	for i, e := range experts {
		m := matches[i]
		switch m.Verdict {
		case VerdictStillDiffers:
			sum.StillDiffers++
		case VerdictResolved:
			sum.Resolved++
		default:
			sum.NoMatch++
		}
		if m.ClaimMissing {
			sum.ClaimMismatch++
		}
		v := idcheck.V(
			"comparison", comparisonID,
			"number", e.Number, "text", e.Text, "answer", e.Answer,
			"location", firstNonEmpty(e.Where, e.Section), "source_file", e.Source,
			"verdict", m.Verdict, "match_note", m.Note,
		)
		v["match_score"] = []string{strconv.FormatFloat(100*m.Score, 'f', 0, 64)}
		v["matched_discrepancies"] = m.FindingIDs
		if len(m.FindingIDs) == 0 {
			v["matched_discrepancies"] = []string{""} // clear stale links
		}
		if m.ClaimMissing {
			v["claim_mismatch"] = []string{"да"}
		} else {
			v["claim_mismatch"] = []string{""}
		}
		key := norm(e.Number)
		seen[key] = true
		if o, ok := byNumber[key]; ok {
			if err := b.Update(ns, ModExpertise, o.ID, v); err != nil {
				return nil, err
			}
			continue
		}
		v["status"] = []string{"open"}
		if _, err := b.Create(ns, ModExpertise, v); err != nil {
			return nil, err
		}
	}
	// imported remarks that are no longer in the file go; hand-made ones stay
	for key, o := range byNumber {
		if !seen[key] && o.Get("source_file") != "" {
			if err := b.Delete(ns, ModExpertise, o.ID); err != nil {
				return nil, err
			}
		}
	}
	return sum, nil
}

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if s != "" {
			return s
		}
	}
	return ""
}
