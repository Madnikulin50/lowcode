package main

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/cmp"
)

func atoiSafe(s string) int { n, _ := strconv.Atoi(strings.TrimSpace(s)); return n }

func dialogFinding(d discrepancyRecord, names map[string]string) cmp.DialogFinding {
	pr, pt := atoiSafe(d.PageRef), atoiSafe(d.PageTarget)
	if pr == 0 && pt == 0 {
		pr, pt = atoiSafe(d.PageNumber), atoiSafe(d.PageNumber)
	}
	return cmp.DialogFinding{
		ID: d.ID, Severity: d.Severity, Kind: d.Kind, Description: strings.TrimPrefix(strings.TrimPrefix(d.Description, "[детерм.] "), "[ИИ] "),
		RefDoc: d.DocRef, TargetDoc: d.DocTarget, RefDocName: names[d.DocRef], TgtDocName: names[d.DocTarget], PageRef: pr, PageTarget: pt,
	}
}

// sheetExcerpt is the text of a page around the anchor (the rows of the sheet nearest to it).
func sheetExcerpt(ns, recordID, side, docID, name string, page int, anchor string) (cmp.DialogExcerpt, bool) {
	if docID == "" || page <= 0 {
		return cmp.DialogExcerpt{}, false
	}
	data, _, err := compose.FetchDocument(ns, recordID, docID)
	if err != nil || !cmp.IsPDF(data) {
		return cmp.DialogExcerpt{}, false
	}
	pages, err := cmp.ExtractPages(docID, data, cmpCacheDir)
	if err != nil || page > len(pages) {
		return cmp.DialogExcerpt{}, false
	}
	var lines []string
	hit := -1
	key := cmp.NormKey(anchor)
	for _, l := range pages[page-1].Lines {
		if l.Stamp {
			continue
		}
		if hit < 0 && key != "" && len(key) > 10 && strings.Contains(l.Key, key[:min(len(key), 40)]) {
			hit = len(lines)
		}
		lines = append(lines, l.Text)
	}
	if len(lines) == 0 {
		return cmp.DialogExcerpt{}, false
	}
	from, to := 0, len(lines)
	if hit >= 0 {
		from, to = max(0, hit-5), min(len(lines), hit+7)
	} else if to > 14 {
		to = 14
	}
	return cmp.DialogExcerpt{Side: side, Doc: docID, Name: name, Page: page, Lines: lines[from:to]}, true
}

// dialogFactsFor gathers what the model may rely on for a comparison and the selected finding.
func dialogFactsFor(ns, recordID, discrepancyID string) (cmp.DialogFacts, error) {
	rec, err := compose.Comparison(ns, recordID)
	if err != nil {
		return cmp.DialogFacts{}, err
	}
	discs, err := compose.Discrepancies(ns, recordID)
	if err != nil {
		return cmp.DialogFacts{}, err
	}
	docs, _ := compose.Documents(ns, recordID)
	names := map[string]string{}
	for _, d := range docs {
		names[d.ID] = d.Name
	}
	ref, tgt := cmp.SideNames(modeOf(rec.Kind))
	f := cmp.DialogFacts{Title: rec.Title, SetsLabel: rec.SetsLabel, Similarity: rec.SimilarityPercent, RefName: ref, TargetName: tgt}
	var sel *discrepancyRecord
	for i := range discs {
		if discs[i].ID == discrepancyID {
			sel = &discs[i]
		}
	}
	if sel != nil {
		df := dialogFinding(*sel, names)
		f.Selected = &df
		// neighbours: the same documents, nearby sheets, most serious first
		rank := map[string]int{"high": 0, "medium": 1, "low": 2}
		var rel []discrepancyRecord
		for _, d := range discs {
			if d.ID == sel.ID || d.Kind == cmp.KindNotRelease || d.Kind == cmp.KindScopeDiff {
				continue
			}
			near := func(a, b string) bool {
				x, y := atoiSafe(a), atoiSafe(b)
				return x > 0 && y > 0 && x-y <= 2 && y-x <= 2
			}
			if (d.DocRef == sel.DocRef && near(d.PageRef, sel.PageRef)) || (d.DocTarget == sel.DocTarget && near(d.PageTarget, sel.PageTarget)) {
				rel = append(rel, d)
			}
		}
		sort.SliceStable(rel, func(i, j int) bool { return rank[rel[i].Severity] < rank[rel[j].Severity] })
		for i, d := range rel {
			if i >= 8 {
				break
			}
			f.Related = append(f.Related, dialogFinding(d, names))
		}
		if ex, ok := sheetExcerpt(ns, recordID, cmp.SideRef, df.RefDoc, df.RefDocName, df.PageRef, sel.Anchor); ok {
			f.Excerpts = append(f.Excerpts, ex)
		}
		if ex, ok := sheetExcerpt(ns, recordID, cmp.SideTarget, df.TargetDoc, df.TgtDocName, df.PageTarget, ""); ok {
			f.Excerpts = append(f.Excerpts, ex)
		}
	}
	if exp, err := compose.Expertise(ns, recordID); err == nil {
		for i, r := range exp {
			if i >= 5 {
				break
			}
			f.Remarks = append(f.Remarks, fmt.Sprintf("№%s: %s", firstNonEmpty(r.Number, r.Key), trimTo(r.Text, 240)))
		}
	}
	return f, nil
}

func trimTo(s string, n int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= n {
		return string(r)
	}
	return string(r[:n]) + "…"
}

// handleDialog answers a message of the specialist about the comparison (and the selected finding).
func handleDialog(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NamespaceID   string    `json:"namespaceID"`
		RecordID      string    `json:"recordID"`
		DiscrepancyID string    `json:"discrepancyID"`
		Messages      []cmp.Msg `json:"messages"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	ns := firstNonEmpty(body.NamespaceID, defaultNamespaceID)
	if ns == "" || body.RecordID == "" || len(body.Messages) == 0 {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("recordID and messages are required"))
		return
	}
	if !suggester.Enabled() {
		writeErr(w, http.StatusServiceUnavailable, fmt.Errorf("модель не настроена (--remark-model / REMARK_MODEL): диалог недоступен, остальное работает"))
		return
	}
	facts, err := dialogFactsFor(ns, body.RecordID, body.DiscrepancyID)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()
	text, jumps, err := suggester.Converse(ctx, facts, body.Messages)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	if jumps == nil {
		jumps = []cmp.Jump{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"reply": text, "jumps": jumps})
}
