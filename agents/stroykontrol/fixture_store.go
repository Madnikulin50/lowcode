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

// FixtureStore serves the comparison viewer's data from a local directory
// instead of a live Compose instance — no Compose server, network, or API
// token needed at all. Meant for UI development and demos: point --fixtures
// at a directory laid out like
//
//	<dir>/<caseID>/meta.json           {"title","status","similarityPercent","comment"}  (optional)
//	<dir>/<caseID>/discrepancies.json  [{"type","severity","description","pageNumber", ...}]  (optional)
//	<dir>/<caseID>/pd*.<ext>           reference stage (П) files — one or several
//	<dir>/<caseID>/rd*.<ext>           target stage (Р/РД) files — one or several
//
// Without discrepancies.json the PDF files are compared on first request
// (the same engine the Compose runner uses). The recordID a request carries
// is just the caseID (namespaceID is accepted but ignored). See
// fixtures/README.md.
type FixtureStore struct {
	dir     string
	analyze func([]cmp.Input, string, []cmp.ManualPair) (*cmp.Result, error) // inputs, mode, manual pairing

	mu       sync.Mutex
	computed map[string]*cmp.Result
	pairing  map[string][]cmp.ManualPair // caseID → manual pairing (kept in memory)
	status   map[string]string           // discrepancy id → status
	remarks  map[string][]cmp.Remark
	boards   map[string]*fxBoard
	nextID   int
}

func NewFixtureStore(dir string) (*FixtureStore, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("fixtures dir: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("fixtures dir %q is not a directory", dir)
	}
	return &FixtureStore{dir: dir, computed: map[string]*cmp.Result{}, pairing: map[string][]cmp.ManualPair{}, status: map[string]string{}, remarks: map[string][]cmp.Remark{}, boards: map[string]*fxBoard{}}, nil
}

// caseDir resolves a caseID to its directory, filepath.Base'd so a
// recordID can never escape the fixtures dir via "..".
func (f *FixtureStore) caseDir(caseID string) string {
	return filepath.Join(f.dir, filepath.Base(caseID))
}

var docExts = map[string]bool{".pdf": true, ".docx": true, ".xlsx": true, ".dxf": true}

// caseFiles lists the document files of a case with their side.
func (f *FixtureStore) caseFiles(dir string) []documentRecord {
	entries, _ := os.ReadDir(dir)
	var out []documentRecord
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !docExts[strings.ToLower(filepath.Ext(name))] {
			continue
		}
		low := strings.ToLower(name)
		switch {
		case strings.HasPrefix(low, "pd"):
			out = append(out, documentRecord{ID: name, Side: cmp.SideRef, Name: name})
		case strings.HasPrefix(low, "rd"):
			out = append(out, documentRecord{ID: name, Side: cmp.SideTarget, Name: name})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func mimetypeFor(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".pdf":
		return "application/pdf"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "application/octet-stream"
	}
}

type fixtureMeta struct {
	Title             string `json:"title"`
	Status            string `json:"status"`
	SimilarityPercent string `json:"similarityPercent"`
	Comment           string `json:"comment"`
	Parent            string `json:"parent"` // caseID of the previous iteration
	Kind              string `json:"kind"`   // "versions": two versions of one stage (files named pd-* are the earlier one)
}

func (f *FixtureStore) checkCase(recordID string) (string, error) {
	dir := f.caseDir(recordID)
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return "", fmt.Errorf("no such fixture case %q under %s", recordID, f.dir)
	}
	return dir, nil
}

// analysis compares the case's files once and remembers the result.
func (f *FixtureStore) analysis(recordID, dir string) *cmp.Result {
	if f.analyze == nil {
		return nil
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if r, ok := f.computed[recordID]; ok {
		return r
	}
	var inputs []cmp.Input
	for _, d := range f.caseFiles(dir) {
		data, err := os.ReadFile(filepath.Join(dir, d.ID))
		if err != nil {
			continue
		}
		inputs = append(inputs, cmp.Input{ID: d.ID, Name: d.Name, Side: d.Side, Data: data})
	}
	mode := cmp.ModeStages
	if b, err := os.ReadFile(filepath.Join(dir, "meta.json")); err == nil {
		var m fixtureMeta
		if json.Unmarshal(b, &m) == nil {
			mode = modeOf(m.Kind)
		}
	}
	res, err := f.analyze(inputs, mode, f.pairing[recordID])
	if err != nil {
		res = nil
	}
	f.computed[recordID] = res
	return res
}

func (f *FixtureStore) Comparison(_, recordID string) (*comparisonRecord, error) {
	dir, err := f.checkCase(recordID)
	if err != nil {
		return nil, err
	}
	var meta fixtureMeta
	b, err := os.ReadFile(filepath.Join(dir, "meta.json"))
	switch {
	case err == nil:
		if err := json.Unmarshal(b, &meta); err != nil {
			return nil, fmt.Errorf("fixture %s: meta.json: %w", recordID, err)
		}
	case os.IsNotExist(err):
		meta.Title = recordID // meta.json is optional; still usable with just files
	default:
		return nil, err
	}
	rec := &comparisonRecord{Title: meta.Title, Status: meta.Status, SimilarityPercent: meta.SimilarityPercent, Comment: meta.Comment, ObjectID: recordID, Kind: modeOf(meta.Kind)}
	if meta.Parent != "" {
		rec.ParentID = meta.Parent
		rec.IterationNo = strconv.Itoa(f.iterationNo(meta.Parent) + 1)
		if pm, err := f.Comparison("", meta.Parent); err == nil {
			rec.ParentTitle = pm.Title
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "discrepancies.json")); err != nil && rec.Status == "" {
		if res := f.analysis(recordID, dir); res != nil {
			rec.Status = "done"
			rec.SimilarityPercent = strconv.FormatFloat(res.Stats.Similarity, 'f', 1, 64)
			rec.Comment = strings.Join(res.Notes, " ")
		}
	}
	return rec, nil
}

func (f *FixtureStore) Discrepancies(_, recordID string) ([]discrepancyRecord, error) {
	dir, err := f.checkCase(recordID)
	if err != nil {
		return nil, err
	}
	var out []discrepancyRecord
	b, err := os.ReadFile(filepath.Join(dir, "discrepancies.json"))
	switch {
	case err == nil:
		if err := json.Unmarshal(b, &out); err != nil {
			return nil, fmt.Errorf("fixture %s: discrepancies.json: %w", recordID, err)
		}
		for i := range out {
			if out[i].ID == "" {
				out[i].ID = fmt.Sprintf("fx-%d", i+1)
			}
		}
	case os.IsNotExist(err):
		if res := f.analysis(recordID, dir); res != nil {
			for _, fd := range res.Findings {
				out = append(out, findingRecord(fd, "fx-"+fd.Fingerprint))
			}
			out = f.iterate(recordID, res, out)
		}
	default:
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := range out {
		if st, ok := f.status[recordID+"/"+out[i].ID]; ok {
			out[i].Status = st
		}
		if out[i].Status == "" {
			out[i].Status = cmp.StatusNew
		}
	}
	return out, nil
}

func (f *FixtureStore) Documents(_, recordID string) ([]documentRecord, error) {
	dir, err := f.checkCase(recordID)
	if err != nil {
		return nil, err
	}
	docs := f.caseFiles(dir)
	if res := f.computedIfAny(recordID); res != nil {
		for i := range docs {
			if vs := res.Volumes[docs[i].ID]; len(vs) > 0 {
				var codes []string
				for _, v := range vs {
					codes = append(codes, v.Code)
				}
				docs[i].VolumeCode = strings.Join(codes, ", ")
				docs[i].VolumeTitle = vs[0].Title
				docs[i].Pages = strconv.Itoa(vs[len(vs)-1].To)
			}
		}
	}
	return docs, nil
}

func (f *FixtureStore) computedIfAny(recordID string) *cmp.Result {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.computed[recordID]
}

func (f *FixtureStore) FetchDocument(_, recordID, docID string) ([]byte, string, error) {
	dir, err := f.checkCase(recordID)
	if err != nil {
		return nil, "", err
	}
	if docID == legacyPD || docID == legacyRD {
		want := cmp.SideRef
		if docID == legacyRD {
			want = cmp.SideTarget
		}
		docID = ""
		for _, d := range f.caseFiles(dir) {
			if d.Side == want {
				docID = d.ID
				break
			}
		}
	}
	path := filepath.Join(dir, filepath.Base(docID))
	data, err := os.ReadFile(path)
	if err != nil || docID == "" {
		return nil, "", fmt.Errorf("fixture case %q has no document %q", recordID, docID)
	}
	return data, mimetypeFor(path), nil
}

func (f *FixtureStore) SetDiscrepancyStatus(_, recordID, discrepancyID, status string) error {
	if _, err := f.checkCase(recordID); err != nil {
		return err
	}
	f.mu.Lock()
	f.status[recordID+"/"+discrepancyID] = status
	f.mu.Unlock()
	return nil
}

func (f *FixtureStore) Remarks(_, recordID string) ([]cmp.Remark, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]cmp.Remark{}, f.remarks[recordID]...), nil
}

func (f *FixtureStore) SaveRemark(_, recordID string, r cmp.Remark) (cmp.Remark, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r.ID == "" {
		f.nextID++
		r.ID = fmt.Sprintf("rm-%d", f.nextID)
		f.remarks[recordID] = append(f.remarks[recordID], r)
		return r, nil
	}
	for i, x := range f.remarks[recordID] {
		if x.ID == r.ID {
			f.remarks[recordID][i] = r
			return r, nil
		}
	}
	return r, fmt.Errorf("remark %s not found", r.ID)
}

// parentOf reads meta.json's parent of a case ("" if none).
func (f *FixtureStore) parentOf(caseID string) string {
	b, err := os.ReadFile(filepath.Join(f.caseDir(caseID), "meta.json"))
	if err != nil {
		return ""
	}
	var m fixtureMeta
	_ = json.Unmarshal(b, &m)
	return m.Parent
}

func (f *FixtureStore) iterationNo(caseID string) int {
	n := 1
	for p := f.parentOf(caseID); p != "" && n < 50; p = f.parentOf(p) {
		n++
	}
	return n
}

// iterate marks the findings of a case that has a parent: still there /
// new, inheriting the reviewer's decisions, plus the ones the corrections
// fixed (the parent's findings that are gone).
func (f *FixtureStore) iterate(recordID string, res *cmp.Result, cur []discrepancyRecord) []discrepancyRecord {
	parent := f.parentOf(recordID)
	if parent == "" {
		return cur
	}
	pd, err := f.Discrepancies("", parent)
	if err != nil {
		return cur
	}
	var prior []cmp.Prior
	byFP := map[string]discrepancyRecord{}
	for _, d := range pd {
		if d.IterState == cmp.IterResolved {
			continue
		}
		prior = append(prior, cmp.Prior{Fingerprint: d.Fingerprint, Status: d.Status})
		byFP[d.Fingerprint] = d
	}
	state, status, resolved := cmp.Carry(prior, res.Findings)
	for i := range cur {
		cur[i].IterState = state[cur[i].Fingerprint]
		if s := status[cur[i].Fingerprint]; s != "" {
			cur[i].Status = s
		}
	}
	for _, p := range resolved {
		d := byFP[p.Fingerprint]
		d.ID = "fx-" + p.Fingerprint + "-resolved"
		d.IterState = cmp.IterResolved
		cur = append(cur, d)
	}
	return cur
}

// RegisterIssued: in --fixtures mode the registry is the case's in-memory board.
func (f *FixtureStore) RegisterIssued(_, caseID string, in cmp.IssueInput, res cmp.IssueResult) (*cmp.IssueRegistered, error) {
	b, err := f.loadBoard(caseID)
	if err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	out := &cmp.IssueRegistered{}
	byGroup := map[string]*fxRemark{}
	maxIdx := 0
	for _, r := range b.remarks {
		if r.rec.Origin == cmp.OriginComparison {
			byGroup[r.rec.Key] = r
			var bt, idx int
			if _, err := fmt.Sscanf(strings.TrimPrefix(r.rec.Number, "ПТО-"), "%d.%d", &bt, &idx); err == nil && idx > maxIdx {
				maxIdx = idx
			}
		}
	}
	now := fmtTime(time.Now().UTC())
	note := "сформировано по сравнению «" + in.ComparisonTitle + "»"
	for _, is := range res.Active {
		docs := []string{}
		for _, d := range is.DocIDs {
			docs = append(docs, d)
		}
		if ex, ok := byGroup[is.GroupKey]; ok {
			if ex.rec.Status == cmp.RemarkResolved {
				ex.rec.Status = cmp.RemarkAwaiting
				ex.history = append(ex.history, remarkHistoryRecord{Iteration: 1, Status: cmp.RemarkAwaiting, At: now, Author: "агент", Source: "возвращено: расхождение снова есть"})
				out.Reopened++
			}
			ex.rec.Text, ex.rec.Refs, ex.rec.Section, ex.rec.Criticality = is.Text, is.Refs, is.Section, is.Criticality
			ex.rec.Findings, ex.rec.Documents = nonNil(is.FindingIDs), nonNil(docs)
			out.Updated++
			continue
		}
		maxIdx++
		number := fmt.Sprintf("ПТО-1.%d", maxIdx)
		rr := remarkRecord{ID: "ex-" + is.GroupKey, Key: is.GroupKey, Number: number, Batch: "ПТО-1", Text: is.Text, Refs: is.Refs, Section: is.Section, Basis: is.Basis,
			Criticality: is.Criticality, Iterations: 1, ChangedAt: now, SourceStatus: cmp.RemarkStatusLabels[cmp.RemarkAwaiting], Status: cmp.RemarkAwaiting,
			Verdict: cmp.VerdictStillDiffers, MatchNote: note, Findings: nonNil(is.FindingIDs), Documents: nonNil(docs), VolumeRefs: []cmp.VolumeRef{}, Origin: cmp.OriginComparison}
		b.remarks[rr.ID] = &fxRemark{rec: rr, history: []remarkHistoryRecord{{Iteration: 1, Status: cmp.RemarkAwaiting, At: now, Author: "агент", Source: note}}}
		out.Created++
		out.Numbers = append(out.Numbers, number)
	}
	for _, is := range res.Fixed {
		if ex, ok := byGroup[is.GroupKey]; ok && ex.rec.Status != cmp.RemarkResolved {
			ex.rec.Status = cmp.RemarkResolved
			ex.history = append(ex.history, remarkHistoryRecord{Iteration: 1, Status: cmp.RemarkResolved, At: now, Author: "агент", Source: "устранено: расхождений в сравнении больше нет"})
			out.Closed++
		}
	}
	return out, nil
}

func (f *FixtureStore) IssuedNumbers(_, caseID string) (map[string]string, error) {
	b, err := f.loadBoard(caseID)
	if err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	out := map[string]string{}
	for _, r := range b.remarks {
		if r.rec.Origin == cmp.OriginComparison {
			out[r.rec.Key] = r.rec.Number
		}
	}
	return out, nil
}

func (f *FixtureStore) ObjectName(_, objectID string) string {
	if c, err := f.Comparison("", objectID); err == nil {
		return c.Title
	}
	return ""
}

// Pairing implements store: the documents, the manual pairing and what the last comparison paired.
func (f *FixtureStore) Pairing(ns, recordID string) (*pairingOut, error) {
	dir, err := f.checkCase(recordID)
	if err != nil {
		return nil, err
	}
	docs, err := f.Documents(ns, recordID)
	if err != nil {
		return nil, err
	}
	out := &pairingOut{Documents: docs, Pairing: []cmp.ManualPair{}}
	f.mu.Lock()
	out.Pairing = append(out.Pairing, f.pairing[recordID]...)
	f.mu.Unlock()
	if res := f.analysis(recordID, dir); res != nil {
		st := res.State()
		out.State = &st
	}
	return out, nil
}

// SetPairing implements store: the comparison is recomputed on the next request.
func (f *FixtureStore) SetPairing(_, recordID string, p []cmp.ManualPair) error {
	if _, err := f.checkCase(recordID); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pairing[recordID] = p
	delete(f.computed, recordID)
	return nil
}
