package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/cmp"
)

// --- remarks in --fixtures mode ------------------------------------------------
//
// A fixture case is both the comparison and its "object": the remarks are read
// from <case>/remarks*.xlsx or remarks*.pdf (the expertise export, or any table with a remark
// column, or numbered remarks of a PDF) on first use, linked to the case's documents and checked against its
// findings exactly like the registry does. Changes the specialist makes in the
// workbench live in memory.

type fxRemark struct {
	rec     remarkRecord
	history []remarkHistoryRecord
	set     cmp.RequirementSet                // the requirements as stored (with the hash of the text)
	check   func(text string) cmp.RemarkMatch // the check of one requirement against the findings of the case
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
	pdfs, _ := filepath.Glob(filepath.Join(dir, "remarks*.pdf"))
	files = append(files, pdfs...)
	sort.Strings(files)
	if len(files) > 0 {
		discs, err := f.Discrepancies("", caseID)
		if err != nil {
			return nil, err
		}
		docs, _ := f.Documents("", caseID)
		var labels []cmp.DocLabel
		var locDocs []cmp.LocDoc
		var matDocs []cmp.MatDoc
		for _, d := range docs {
			labels = append(labels, cmp.DocLabel{ID: d.ID, Codes: d.VolumeCode, Title: d.VolumeTitle})
			pg, _ := strconv.Atoi(d.Pages)
			matDocs = append(matDocs, cmp.MatDoc{ID: d.ID, Name: firstNonEmpty(strings.TrimSpace(d.VolumeCode+" "+d.VolumeTitle), d.Name), Codes: d.VolumeCode, Title: d.VolumeTitle,
				Pages: pg, Sheets: d.Sheets, Headings: d.Headings})
			locDocs = append(locDocs, cmp.LocDoc{ID: d.ID, Name: firstNonEmpty(strings.TrimSpace(d.VolumeCode+" "+d.VolumeTitle), d.Name), Side: d.Side, Codes: d.VolumeCode, Title: d.VolumeTitle, Pages: pg, Sheets: d.Sheets, Comparison: caseID})
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
			rows, err := cmp.ParseRemarksFile(data, name, "")
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
				// the requirements, each checked on its own
				checkText := func(text string) cmp.RemarkMatch {
					mq, _ := cmp.CheckRemark(cmp.Expert{Number: row.Number, Text: text, Where: row.Section, Answer: row.Answer}, refs, labels, cands, docOf)
					return mq
				}
				reqSet, _ := cmp.EnsureRequirements(cmp.RequirementSet{}, row.Text)
				fillRequirements(&reqSet, checkText)
				rr.Requirements = append([]cmp.Requirement{}, reqSet.Items...)
				rr.Materials = cmp.CheckMaterials(cmp.ParseMaterials(row.Answer), matDocs)
				rr.MaterialsState = cmp.MaterialsSummary(rr.Materials)
				rr.fillOutcome()
				var matched []cmp.FoundAt
				for _, id := range m.FindingIDs {
					if f, ok := foundAt[id]; ok {
						matched = append(matched, f)
					}
				}
				rr.Links = append(cmp.LocateRemark(refs, locDocs, matched), cmp.LocateFragments(row.Text, refs, f.fragmentDocs(caseID, locDocs, linked, m.FindingIDs, docOf))...)
				if rr.Links == nil {
					rr.Links = []cmp.RemarkLocation{}
				}
				fr := &fxRemark{rec: rr, set: reqSet, check: checkText}
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
	if err := p.check(r.rec.Status, r.rec.Assignee); err != nil {
		return nil, err
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
			r.rec.OutcomeBy, r.rec.OutcomeAt = p.Actor.Label(), now
			r.history = append(r.history, remarkHistoryRecord{Iteration: max(r.rec.Iterations, 1), Status: r.rec.Status, SourceStatus: r.rec.SourceStatus,
				At: now, Author: p.Actor.Label(), Source: "решение по проверке: " + cmp.OutcomeLabels[final], Note: comment, Kind: cmp.HistDecision})
		}
	}
	if p.Status != nil && *p.Status != r.rec.Status {
		r.history = append(r.history, remarkHistoryRecord{Iteration: max(r.rec.Iterations, 1), Answer: r.rec.Answer, RejectReason: r.rec.RejectReason,
			Status: *p.Status, SourceStatus: r.rec.SourceStatus, At: fmtTime(time.Now().UTC()), Author: p.Actor.Label(), Source: "вручную", Note: statusComment, Kind: cmp.HistStatus})
		r.rec.Status = *p.Status
	}
	for _, wc := range cmp.WorkChanges(r.rec.Assignee, r.rec.DueDate, r.rec.Note, p.Assignee, p.DueDate, p.Note) {
		r.history = append(r.history, remarkHistoryRecord{Iteration: max(r.rec.Iterations, 1), Status: valOr(p.Status, r.rec.Status), SourceStatus: r.rec.SourceStatus,
			At: fmtTime(time.Now().UTC()), Author: p.Actor.Label(), Source: wc.Source, Note: wc.Note, Kind: wc.Kind})
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

// fillRequirements checks the requirements that have no result yet (new or edited ones).
func fillRequirements(set *cmp.RequirementSet, check func(text string) cmp.RemarkMatch) {
	for i := range set.Items {
		q := &set.Items[i]
		if q.Verdict != "" || q.Outcome != "" {
			continue
		}
		mq := check(q.Text)
		q.Verdict, q.Score, q.Note, q.FindingIDs = mq.Verdict, mq.Score, mq.Note, mq.FindingIDs
		q.Outcome = cmp.CheckOutcome(q.Verdict, q.AIVerdict)
	}
}

// SaveRequirements implements store: the list is replaced, the new and edited requirements are checked.
func (f *FixtureStore) SaveRequirements(_, caseID, remarkID string, in []cmp.RequirementInput, by cmp.Actor) (*remarkRecord, error) {
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
	if err := checkRequirementEdit(by, r.set, in, r.rec.Status, r.rec.Assignee); err != nil {
		return nil, err
	}
	now := fmtTime(time.Now().UTC())
	set, err := cmp.ApplyRequirementEdit(r.set, r.rec.Text, in, by.Label(), now)
	if err != nil {
		return nil, err
	}
	if r.check != nil {
		fillRequirements(&set, r.check)
	}
	for _, wc := range cmp.RequirementChanges(r.set, set) {
		r.history = append(r.history, remarkHistoryRecord{Iteration: max(r.rec.Iterations, 1), Status: r.rec.Status, SourceStatus: r.rec.SourceStatus,
			At: now, Author: by.Label(), Source: wc.Source, Note: wc.Note, Kind: wc.Kind})
	}
	r.set = set
	r.rec.Requirements, r.rec.RequirementsStale = append([]cmp.Requirement{}, set.Items...), set.Stale
	r.rec.fillOutcome()
	out := r.rec
	return &out, nil
}

// fxNotifyRead: the notifications read in --fixtures mode (by id: key|recipient).
var fxNotifyRead = struct {
	sync.Mutex
	m map[string]bool
}{m: map[string]bool{}}

// Notifications computes the notifications of a case from its registry and history in memory (ТЗ 3.7.5):
// nothing is stored or sent; the recipients are the assignees by name, and the managers of
// <case>/recipients.json if there is one (a list of cmp.Recipient).
func (f *FixtureStore) Notifications(caseID, recipient string) ([]cmp.Notification, error) {
	b, err := f.loadBoard(caseID)
	if err != nil {
		return nil, err
	}
	var dir []cmp.Recipient
	if data, err := os.ReadFile(filepath.Join(f.caseDir(caseID), "recipients.json")); err == nil {
		if err := json.Unmarshal(data, &dir); err != nil {
			return nil, fmt.Errorf("recipients.json: %w", err)
		}
	}
	f.mu.Lock()
	var remarks []cmp.NotifyRemark
	var hist []cmp.NotifyHist
	for id, r := range b.remarks {
		remarks = append(remarks, cmp.NotifyRemark{ID: id, Number: r.rec.Number, Object: caseID, Status: r.rec.Status, Assignee: r.rec.Assignee,
			DueDate: r.rec.DueDate, Text: r.rec.Text, Archived: r.rec.Archived})
		for i, h := range r.history {
			hist = append(hist, cmp.NotifyHist{ID: id + "#" + strconv.Itoa(i), Remark: id, Kind: h.Kind, Status: h.Status, Source: h.Source, Note: h.Note, Author: h.Author, At: h.At})
		}
	}
	f.mu.Unlock()
	all := cmp.NotifyEvents(caseID, remarks, hist, dir, map[string]bool{}, time.Now(), cmp.NotifyOptions{})
	who := cmp.NormName(recipient)
	out := []cmp.Notification{}
	fxNotifyRead.Lock()
	defer fxNotifyRead.Unlock()
	for _, n := range all {
		if who != "" && cmp.NormName(n.Recipient) != who {
			continue
		}
		n.ID = n.Key + "|" + n.Recipient
		n.Read = fxNotifyRead.m[n.ID]
		out = append(out, n)
	}
	return out, nil
}

// MarkNotificationsRead marks notifications of the --fixtures mode as read.
func (f *FixtureStore) MarkNotificationsRead(ids []string) {
	fxNotifyRead.Lock()
	defer fxNotifyRead.Unlock()
	for _, id := range ids {
		fxNotifyRead.m[id] = true
	}
}

// fragmentDocs reads the pages of the documents the fragment of a remark is looked for in.
func (f *FixtureStore) fragmentDocs(caseID string, docs []cmp.LocDoc, linked, findings []string, docOf map[string][2]string) []cmp.FragDoc {
	want := map[string]bool{}
	for _, id := range linked {
		want[id] = true
	}
	for _, id := range findings {
		want[docOf[id][0]], want[docOf[id][1]] = true, true
	}
	var out []cmp.FragDoc
	for _, d := range docs {
		if !want[d.ID] {
			continue
		}
		data, _, err := f.FetchDocument("", caseID, d.ID)
		if err != nil || !cmp.IsPDF(data) {
			continue
		}
		if pages, err := cmp.ExtractPages(d.ID, data, ""); err == nil {
			out = append(out, cmp.FragDoc{LocDoc: d, Pages: pages})
		}
	}
	return out
}
