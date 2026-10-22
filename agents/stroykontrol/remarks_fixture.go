package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/cmp"
)

// --- remarks in --fixtures mode ------------------------------------------------
//
// A fixture case is both the comparison and its "object": the remarks are read
// from <case>/remarks*.xlsx (the expertise export, or any table with a remark
// column) on first use, linked to the case's documents and checked against its
// findings exactly like the registry does. Changes the specialist makes in the
// workbench live in memory.

type fxRemark struct {
	rec     remarkRecord
	history []remarkHistoryRecord
}

type fxBoard struct {
	remarks map[string]*fxRemark
	imports []importRecord
	docs    map[string]string
}

func fmtTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02T15:04:05Z")
}

func (f *FixtureStore) loadBoard(caseID string) (*fxBoard, error) {
	dir, err := f.checkCase(caseID)
	if err != nil {
		return nil, err
	}
	f.mu.Lock()
	if b, ok := f.boards[caseID]; ok {
		f.mu.Unlock()
		return b, nil
	}
	f.mu.Unlock()

	board := &fxBoard{remarks: map[string]*fxRemark{}, docs: map[string]string{}}
	files, _ := filepath.Glob(filepath.Join(dir, "remarks*.xlsx"))
	sort.Strings(files)
	if len(files) > 0 {
		discs, err := f.Discrepancies("", caseID)
		if err != nil {
			return nil, err
		}
		docs, _ := f.Documents("", caseID)
		var labels []cmp.DocLabel
		var locDocs []cmp.LocDoc
		for _, d := range docs {
			labels = append(labels, cmp.DocLabel{ID: d.ID, Codes: d.VolumeCode, Title: d.VolumeTitle})
			pg, _ := strconv.Atoi(d.Pages)
			locDocs = append(locDocs, cmp.LocDoc{ID: d.ID, Name: strings.TrimSpace(d.VolumeCode + " " + d.VolumeTitle), Side: d.Side, Codes: d.VolumeCode, Title: d.VolumeTitle, Pages: pg, Comparison: caseID})
			board.docs[d.ID] = strings.TrimSpace(d.VolumeCode + " " + d.VolumeTitle)
		}
		var cands []cmp.Candidate
		docOf := map[string][2]string{}
		foundAt := map[string]cmp.FoundAt{}
		for _, d := range discs {
			pp, _ := strconv.Atoi(d.PageRef)
			tp, _ := strconv.Atoi(d.PageTarget)
			foundAt[d.ID] = cmp.FoundAt{ID: d.ID, Comparison: caseID, Text: strings.TrimPrefix(strings.TrimPrefix(d.Description, "[детерм.] "), "[ИИ] "),
				RefDoc: d.DocRef, TargetDoc: d.DocTarget, RefPage: pp, TargetPage: tp, RefBox: d.BboxRef, TargetBox: d.BboxTarget}
			var ev cmp.Evidence
			_ = jsonUnmarshal(d.Evidence, &ev)
			text := strings.TrimPrefix(strings.TrimPrefix(d.Description, "[детерм.] "), "[ИИ] ")
			cands = append(cands, cmp.Candidate{ID: d.ID, Text: text + " " + d.Anchor + " " + ev.Ref + " " + ev.Target, IterState: d.IterState, Severity: d.Severity})
			docOf[d.ID] = [2]string{d.DocRef, d.DocTarget}
		}
		now := time.Now().UTC()
		for _, p := range files {
			data, err := os.ReadFile(p)
			if err != nil {
				return nil, err
			}
			name := filepath.Base(p)
			rows, err := cmp.ParseExport(data, name)
			if err == cmp.ErrNotExport {
				var generic []cmp.Expert
				if generic, err = cmp.ParseExpertise(data, name); err == nil {
					rows = cmp.ExpertToExport(generic)
				}
			}
			if err != nil {
				return nil, fmt.Errorf("%s: %w", name, err)
			}
			imp := importRecord{ID: "imp-" + name, File: name, At: fmtTime(now), Total: len(rows), New: len(rows)}
			for _, row := range rows {
				rec := cmp.Reconcile(nil, row, now)
				refs := cmp.ParseVolumeRefs(row.Refs)
				m, linked := cmp.CheckRemark(cmp.Expert{Number: row.Number, Text: row.Text, Where: row.Section, Answer: row.Answer}, refs, labels, cands, docOf)
				rr := remarkRecord{
					ID: "ex-" + row.Key, Key: row.Key, Number: row.Number, Batch: row.Batch, Text: row.Text, Refs: row.Refs, VolumeRefs: refs,
					Section: row.Section, Basis: row.Basis, Answer: row.Answer, RejectReason: row.RejectReason, PrevAnswer: row.PrevAnswer,
					PrevRejectReason: row.PrevRejectReason, Discipline: row.Discipline, Expert: row.Expert, Criticality: row.Criticality,
					Proposals: row.Proposals, Expected: row.Expected, Iterations: row.Iterations, ChangedAt: fmtTime(row.ChangedAt),
					SourceStatus: row.SourceStatus, Status: rec.Status, Verdict: m.Verdict, MatchNote: m.Note, ClaimMismatch: m.ClaimMissing,
					Score: strconv.FormatFloat(100*m.Score, 'f', 0, 64), Findings: nonNil(m.FindingIDs), Documents: nonNil(linked),
				}
				if rr.VolumeRefs == nil {
					rr.VolumeRefs = []cmp.VolumeRef{}
				}
				rr.fillOutcome()
				var matched []cmp.FoundAt
				for _, id := range m.FindingIDs {
					if f, ok := foundAt[id]; ok {
						matched = append(matched, f)
					}
				}
				rr.Links = cmp.LocateRemark(refs, locDocs, matched)
				if rr.Links == nil {
					rr.Links = []cmp.RemarkLocation{}
				}
				fr := &fxRemark{rec: rr}
				for _, s := range rec.History {
					fr.history = append(fr.history, remarkHistoryRecord{Iteration: s.Iteration, Answer: s.Answer, RejectReason: s.RejectReason,
						Status: s.Status, SourceStatus: s.SourceStatus, At: fmtTime(s.At), Author: "агент", Source: s.Source})
				}
				board.remarks[rr.ID] = fr
			}
			board.imports = append(board.imports, imp)
		}
	}
	f.mu.Lock()
	f.boards[caseID] = board
	f.mu.Unlock()
	return board, nil
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func (f *FixtureStore) Board(_, caseID string) (*boardOut, error) {
	b, err := f.loadBoard(caseID)
	if err != nil {
		return nil, err
	}
	title := caseID
	if c, err := f.Comparison("", caseID); err == nil && c.Title != "" {
		title = c.Title
	}
	out := &boardOut{ObjectID: caseID, ObjectName: title, Remarks: []remarkRecord{}, Imports: b.imports, Documents: b.docs,
		Comparisons: []boardComparison{{ID: caseID, Title: title}}}
	if out.Imports == nil {
		out.Imports = []importRecord{}
	}
	f.mu.Lock()
	for _, r := range b.remarks {
		out.Remarks = append(out.Remarks, r.rec)
	}
	f.mu.Unlock()
	sortRemarks(out.Remarks)
	// no store behind the fixtures: the dynamics is the one point of today
	rows := make([]cmp.StatusRow, 0, len(out.Remarks))
	for _, r := range out.Remarks {
		if !r.Archived {
			rows = append(rows, cmp.StatusRow{Status: r.Status, DueDate: r.DueDate})
		}
	}
	out.Dynamics = []cmp.StatusSnapshot{cmp.CountStatuses(rows, time.Now().UTC().Format("2006-01-02"))}
	return out, nil
}

func (f *FixtureStore) RemarkHistory(_, caseID, remarkID string) ([]remarkHistoryRecord, error) {
	b, err := f.loadBoard(caseID)
	if err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := b.remarks[remarkID]
	if !ok {
		return nil, fmt.Errorf("remark %s not found", remarkID)
	}
	return append([]remarkHistoryRecord{}, r.history...), nil
}

func (f *FixtureStore) UpdateRemark(_, caseID, remarkID string, p remarkPatch) (*remarkRecord, error) {
	b, err := f.loadBoard(caseID)
	if err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := b.remarks[remarkID]
	if !ok {
		return nil, fmt.Errorf("remark %s not found", remarkID)
	}
	statusComment := ""
	if p.StatusComment != nil {
		statusComment = strings.TrimSpace(*p.StatusComment)
	}
	if p.Status != nil {
		if err := cmp.CheckStatusChange(r.rec.Status, *p.Status, statusComment); err != nil {
			return nil, err
		}
	}
	if p.EstimateH != nil {
		est, err := cmp.ParseEstimate(*p.EstimateH)
		if err != nil {
			return nil, err
		}
		r.rec.EstimateH = est
	}
	if p.DueDate != nil && !validDate(*p.DueDate) {
		return nil, fmt.Errorf("dueDate must be YYYY-MM-DD")
	}
	if p.OutcomeFinal != nil || p.OutcomeComment != nil {
		final, comment := r.rec.OutcomeFinal, r.rec.OutcomeComment
		if p.OutcomeFinal != nil {
			final = *p.OutcomeFinal
		}
		if p.OutcomeComment != nil {
			comment = strings.TrimSpace(*p.OutcomeComment)
		}
		kind, err := cmp.ResolveDecision(r.rec.Outcome, final, comment)
		if err != nil {
			return nil, err
		}
		now := fmtTime(time.Now().UTC())
		r.rec.OutcomeFinal, r.rec.OutcomeDecision, r.rec.OutcomeComment = final, kind, comment
		if final == "" {
			r.rec.OutcomeBy, r.rec.OutcomeAt = "", ""
		} else {
			r.rec.OutcomeBy, r.rec.OutcomeAt = "специалист ПТО", now
			r.history = append(r.history, remarkHistoryRecord{Iteration: max(r.rec.Iterations, 1), Status: r.rec.Status, SourceStatus: r.rec.SourceStatus,
				At: now, Author: "специалист ПТО", Source: "решение по проверке: " + cmp.OutcomeLabels[final], Note: comment, Kind: cmp.HistDecision})
		}
	}
	if p.Status != nil && *p.Status != r.rec.Status {
		r.history = append(r.history, remarkHistoryRecord{Iteration: max(r.rec.Iterations, 1), Answer: r.rec.Answer, RejectReason: r.rec.RejectReason,
			Status: *p.Status, SourceStatus: r.rec.SourceStatus, At: fmtTime(time.Now().UTC()), Author: "специалист ПТО", Source: "вручную", Note: statusComment, Kind: cmp.HistStatus})
		r.rec.Status = *p.Status
	}
	for _, wc := range cmp.WorkChanges(r.rec.Assignee, r.rec.DueDate, r.rec.Note, p.Assignee, p.DueDate, p.Note) {
		r.history = append(r.history, remarkHistoryRecord{Iteration: max(r.rec.Iterations, 1), Status: valOr(p.Status, r.rec.Status), SourceStatus: r.rec.SourceStatus,
			At: fmtTime(time.Now().UTC()), Author: "специалист ПТО", Source: wc.Source, Note: wc.Note, Kind: wc.Kind})
	}
	if p.Assignee != nil {
		r.rec.Assignee = strings.TrimSpace(*p.Assignee)
	}
	if p.DueDate != nil {
		r.rec.DueDate = *p.DueDate
	}
	if p.Note != nil {
		r.rec.Note = *p.Note
	}
	out := r.rec
	return &out, nil
}

func (f *FixtureStore) Expertise(_, caseID string) ([]remarkRecord, error) {
	b, err := f.loadBoard(caseID)
	if err != nil {
		return nil, err
	}
	out := []remarkRecord{}
	f.mu.Lock()
	for _, r := range b.remarks {
		out = append(out, r.rec)
	}
	f.mu.Unlock()
	sortRemarks(out)
	return out, nil
}
