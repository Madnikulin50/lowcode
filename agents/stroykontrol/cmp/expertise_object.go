package cmp

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// Modules of the expertise registry.
const (
	ModObjects        = "objects"
	ModRemarkHistory  = "expertise_remark_history"
	ModExpertImports  = "expertise_imports"
	VerdictNoDocument = "no_documents" // the volumes the remark names are not loaded: nothing to check against
)

// ObjectImportSummary is what an import of an export reports.
type ObjectImportSummary struct {
	Files         int `json:"files"`
	Remarks       int `json:"remarks"`
	New           int `json:"new"`
	Changed       int `json:"changed"`
	StatusChanges int `json:"statusChanges"`
	Reopened      int `json:"reopened"`
	Archived      int `json:"archived"`
	Checked       int `json:"checked"`
	Skipped       int `json:"skipped"` // exports that had been imported already
}

func dateValue(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02T15:04:05Z")
}

func remarkValues(object string, e ExportRow, refs []VolumeRef, st string) idcheck.Values {
	rj, _ := json.Marshal(refs)
	v := idcheck.V(
		"object", object, "remark_key", e.Key, "number", e.Number, "batch", e.Batch,
		"text", e.Text, "refs_text", e.Refs, "section", e.Section, "basis", e.Basis,
		"answer", e.Answer, "reject_reason", e.RejectReason,
		"prev_answer", e.PrevAnswer, "prev_reject_reason", e.PrevRejectReason,
		"discipline", e.Discipline, "expert", e.Expert, "criticality", e.Criticality,
		"proposals", e.Proposals, "expected_result", e.Expected,
		"source_status", e.SourceStatus, "source_file", e.Source, "status", st,
	)
	if len(refs) > 0 {
		v["volume_refs"] = []string{string(rj)}
	}
	// Update overlays values: an explicit empty value clears a stale one
	for _, k := range []string{"answer", "reject_reason", "prev_answer", "prev_reject_reason", "proposals", "expected_result", "archived", "criticality"} {
		if _, ok := v[k]; !ok {
			v[k] = []string{""}
		}
	}
	v["iterations"] = []string{strconv.Itoa(e.Iterations)}
	v["source_changed_at"] = []string{dateValue(e.ChangedAt)}
	return v
}

func snapshotValues(remarkID string, s Snapshot) idcheck.Values {
	v := idcheck.V("remark", remarkID, "answer", s.Answer, "reject_reason", s.RejectReason, "status", s.Status,
		"source_status", s.SourceStatus, "source", s.Source, "author", "агент", "kind", firstNonEmpty(s.Kind, HistImport))
	v["iteration"] = []string{strconv.Itoa(s.Iteration)}
	v["at"] = []string{dateValue(s.At)}
	return v
}

func storedFrom(r idcheck.Record) *StoredRemark {
	it, _ := strconv.Atoi(r.Get("iterations"))
	return &StoredRemark{
		Status: r.Get("status"), SourceStatus: r.Get("source_status"), Answer: r.Get("answer"),
		RejectReason: r.Get("reject_reason"), Iterations: it, Archived: r.Get("archived") != "",
	}
}

// ExpertToExport lets the generic parser (any Excel with a remark column)
// feed the same registry.
func ExpertToExport(l []Expert) []ExportRow {
	out := make([]ExportRow, 0, len(l))
	for i, e := range l {
		key := e.Number
		if key == "" {
			key = "без номера: " + shortHash(e.Text)
		}
		_ = i
		out = append(out, ExportRow{Key: key, Number: e.Number, Text: e.Text, Answer: e.Answer, Refs: e.Where, Section: e.Section, Source: e.Source, Iterations: 1})
	}
	return out
}

func attachmentOrder(atts []string) []string {
	out := append([]string(nil), atts...)
	sort.Slice(out, func(i, j int) bool {
		a, _ := strconv.ParseUint(out[i], 10, 64)
		b, _ := strconv.ParseUint(out[j], 10, 64)
		return a < b
	})
	return out
}

// ImportObjectRemarks reads the expertise exports attached to an object
// (field remarks_file, imported in the order they were uploaded; one that had
// been imported before is skipped) and updates the object's registry of
// remarks: new remarks, answers, iterations and statuses, with a history of
// each remark and a journal of the import. The remarks are then checked
// against the documents and comparisons of the object.
func (r *Runner) ImportObjectRemarks(ctx context.Context, ns, objectID string, force bool) (*ObjectImportSummary, error) {
	b := r.Backend
	obj, err := b.GetRecord(ns, ModObjects, objectID)
	if err != nil {
		return nil, err
	}
	atts := attachmentOrder(obj.Values["remarks_file"])
	if len(atts) == 0 {
		return nil, fmt.Errorf("к объекту не приложена выгрузка замечаний (поле «Выгрузки замечаний экспертизы»)")
	}
	journal, err := b.Search(ns, ModExpertImports, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return nil, err
	}
	done := map[string]bool{}
	for _, j := range journal {
		done[j.Get("attachment")] = true
	}
	var pending []string
	for _, a := range atts {
		if !done[a] {
			pending = append(pending, a)
		}
	}
	if len(pending) == 0 && force {
		pending = atts[len(atts)-1:]
	}
	sum := &ObjectImportSummary{Skipped: len(atts) - len(pending)}
	if len(pending) == 0 {
		sum.Checked, err = r.CheckObjectRemarks(ctx, ns, objectID)
		return sum, err
	}

	recs, err := b.Search(ns, ModExpertise, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return nil, err
	}
	byKey := map[string]idcheck.Record{}
	for _, x := range recs {
		byKey[x.Get("remark_key")] = x
	}

	for _, att := range pending {
		data, name, err := b.Download(ns, att)
		if err != nil {
			return nil, fmt.Errorf("загрузка выгрузки: %w", err)
		}
		rows, err := ParseExport(data, name)
		if err == ErrNotExport {
			var generic []Expert
			if generic, err = ParseExpertise(data, name); err == nil {
				rows = ExpertToExport(generic)
			}
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		now := time.Now().UTC()
		var changes []Change
		inFile := map[string]bool{}
		for _, row := range rows {
			inFile[row.Key] = true
			refs := ParseVolumeRefs(row.Refs)
			var st *StoredRemark
			ex, known := byKey[row.Key]
			if known {
				st = storedFrom(ex)
			}
			rec := Reconcile(st, row, now)
			v := remarkValues(objectID, row, refs, rec.Status)
			id := ex.ID
			if !known {
				v["archived"] = []string{""}
				if k := obj.Get("data_kind"); k != "" {
					v["data_kind"] = []string{k} // demo objects keep their remarks out of the working lists
				}
				if id, err = b.Create(ns, ModExpertise, v); err != nil {
					return nil, err
				}
				byKey[row.Key] = idcheck.Record{ID: id, Values: v}
			} else if differs(ex, v) {
				if err := b.Update(ns, ModExpertise, id, v); err != nil {
					return nil, err
				}
				nv := idcheck.Record{ID: id, Values: map[string][]string{}}
				for k, x := range ex.Values {
					nv.Values[k] = x
				}
				for k, x := range v {
					nv.Values[k] = x
				}
				byKey[row.Key] = nv
			}
			for _, s := range rec.History {
				if _, err := b.Create(ns, ModRemarkHistory, snapshotValues(id, s)); err != nil {
					return nil, err
				}
			}
			changes = append(changes, rec.Changes...)
		}
		// remarks that are no longer in the export
		for key, ex := range byKey {
			if inFile[key] || ex.Get("archived") != "" || ex.Get("source_file") == "" || ex.Get("origin") == OriginComparison {
				continue
			}
			if err := b.Update(ns, ModExpertise, ex.ID, idcheck.V("archived", "да")); err != nil {
				return nil, err
			}
			changes = append(changes, Change{key, ChangeArchived, "нет в выгрузке " + name})
		}

		var nNew, nStatus, nReopen, nArch, nChanged int
		for _, c := range changes {
			switch c.Kind {
			case ChangeNew:
				nNew++
			case ChangeStatus:
				nStatus++
			case ChangeReopened:
				nReopen++
			case ChangeArchived:
				nArch++
			default:
				nChanged++
			}
		}
		jv := idcheck.V("object", objectID, "file", name, "attachment", att, "details", JournalText(changes, 200))
		jv["imported_at"] = []string{dateValue(now)}
		for k, n := range map[string]int{"total": len(rows), "new_count": nNew, "changed_count": nChanged, "status_changes": nStatus, "reopened": nReopen, "archived_count": nArch} {
			jv[k] = []string{strconv.Itoa(n)}
		}
		if _, err := b.Create(ns, ModExpertImports, jv); err != nil {
			return nil, err
		}
		sum.Files++
		sum.Remarks = len(rows)
		sum.New += nNew
		sum.Changed += nChanged
		sum.StatusChanges += nStatus
		sum.Reopened += nReopen
		sum.Archived += nArch
	}
	sum.Checked, err = r.CheckObjectRemarks(ctx, ns, objectID)
	return sum, err
}

// differs reports whether writing v would change the stored record.
func differs(stored idcheck.Record, v idcheck.Values) bool {
	for k, x := range v {
		if strings.Join(stored.Values[k], "\n") != strings.Join(x, "\n") {
			return true
		}
	}
	return false
}

// docCodes splits a cmp_documents volume_code ("a, b") into its volumes.
func docCodes(csv string) []string {
	var out []string
	for _, p := range strings.Split(csv, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// DocLabel is a loaded document as the check needs it: its volumes (шифры,
// comma separated) and title.
type DocLabel struct {
	ID, Codes, Title string
}

// CheckRemark links a remark to the loaded documents it names and checks it
// against the findings of those documents (all findings when it names no
// volume). docOf gives the documents (reference, target) of a finding. It is
// pure: the same function serves the registry and the --fixtures mode.
func CheckRemark(e Expert, refs []VolumeRef, docs []DocLabel, cands []Candidate, docOf map[string][2]string) (RemarkMatch, []string) {
	linked := map[string]bool{}
	for _, ref := range refs {
		for _, d := range docs {
			codes := docCodes(d.Codes)
			for _, code := range codes {
				if MatchVolume(ref, code, d.Title) {
					linked[d.ID] = true
				}
			}
			if len(codes) == 0 && MatchVolume(ref, "", d.Title) {
				linked[d.ID] = true
			}
		}
	}
	var linkedIDs []string
	for id := range linked {
		linkedIDs = append(linkedIDs, id)
	}
	sort.Strings(linkedIDs)
	if len(refs) > 0 && len(linked) == 0 {
		return RemarkMatch{Number: e.Number, Verdict: VerdictNoDocument, Note: "Тома, на которые ссылается замечание, не загружены в сравнения объекта — проверка по документам невозможна."}, nil
	}
	pool := cands
	if len(refs) > 0 {
		pool = nil
		for _, c := range cands {
			if d := docOf[c.ID]; linked[d[0]] || linked[d[1]] {
				pool = append(pool, c)
			}
		}
	}
	return MatchRemarks([]Expert{e}, pool)[0], linkedIDs
}

// CheckObjectRemarks links every remark of the object to the loaded documents
// it names and checks it against the findings of the object's comparisons.
// It returns the number of remarks whose check result was written.
func (r *Runner) CheckObjectRemarks(ctx context.Context, ns, objectID string) (int, error) {
	b := r.Backend
	remarks, err := b.Search(ns, ModExpertise, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return 0, err
	}
	comps, err := b.Search(ns, ModComparisons, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return 0, err
	}
	// a comparison that a later iteration continues is superseded: its findings are the earlier state of
	// the same pair, and the later iteration says what became of them (fixed, still there)
	superseded := map[string]bool{}
	for _, c := range comps {
		if p := c.Get("parent_comparison"); p != "" {
			superseded[p] = true
		}
	}
	var docs []DocLabel
	var docRecs, findings []idcheck.Record
	for _, c := range comps {
		ds, err := b.Search(ns, ModDocuments, fmt.Sprintf("comparison = '%s'", c.ID))
		if err != nil {
			return 0, err
		}
		for _, d := range ds {
			docs = append(docs, DocLabel{ID: d.ID, Codes: d.Get("volume_code"), Title: d.Get("volume_title")})
			docRecs = append(docRecs, d)
		}
		if superseded[c.ID] {
			continue
		}
		fs, err := b.Search(ns, ModDiscrepancies, fmt.Sprintf("comparison = '%s'", c.ID))
		if err != nil {
			return 0, err
		}
		findings = append(findings, fs...)
	}
	cands := CandidatesFrom(findings)
	locDocs := locDocsFrom(docRecs, func(string) string { return "" })
	foundAt := map[string]FoundAt{} // finding ID → its places
	for _, f := range findings {
		foundAt[f.ID] = FoundAt{
			ID: f.ID, Comparison: f.Get("comparison"), Text: strings.TrimPrefix(strings.TrimPrefix(f.Get("description"), "[детерм.] "), "[ИИ] "),
			RefDoc: f.Get("doc_ref"), TargetDoc: f.Get("doc_target"), RefPage: atoi(f.Get("page_pd")), TargetPage: atoi(f.Get("page_rd")),
			RefBox: f.Get("bbox_pd"), TargetBox: f.Get("bbox_rd"),
		}
	}
	docOf := map[string][2]string{} // finding ID → its documents
	for _, f := range findings {
		docOf[f.ID] = [2]string{f.Get("doc_ref"), f.Get("doc_target")}
	}

	written := 0
	for _, rem := range remarks {
		if rem.Get("archived") != "" || rem.Get("origin") == OriginComparison {
			continue // remarks made from a comparison already know their findings and documents
		}
		var refs []VolumeRef
		_ = json.Unmarshal([]byte(rem.Get("volume_refs")), &refs)
		e := Expert{Number: rem.Get("number"), Text: rem.Get("text"), Where: rem.Get("section"), Answer: rem.Get("answer")}
		m, linkedIDs := CheckRemark(e, refs, docs, cands, docOf)

		v := idcheck.V("verdict", m.Verdict, "match_note", m.Note)
		v["match_score"] = []string{strconv.FormatFloat(100*m.Score, 'f', 0, 64)}
		v["matched_discrepancies"] = m.FindingIDs
		if len(m.FindingIDs) == 0 {
			v["matched_discrepancies"] = []string{""}
		}
		v["linked_documents"] = linkedIDs
		if len(linkedIDs) == 0 {
			v["linked_documents"] = []string{""}
		}
		// where the remark leads: the places its references name and the regions of the findings it was matched to
		var matched []FoundAt
		for _, id := range m.FindingIDs {
			if f, ok := foundAt[id]; ok {
				matched = append(matched, f)
			}
		}
		links := LocationsJSON(LocateRemark(refs, locDocs, matched))
		v["links_json"] = []string{links}
		v["claim_mismatch"] = []string{""}
		if m.ClaimMissing {
			v["claim_mismatch"] = []string{"да"}
		}
		// a new edition fixed the difference and the designer has answered: to the quality control (ТЗ 3.7.3)
		handedTo, handover := AutoHandover(rem.Get("status"), rem.Get("answer"), rem.Get("verdict"), m.Verdict)
		if handover {
			v["status"] = []string{handedTo}
		}
		if rem.Get("verdict") == m.Verdict && rem.Get("match_note") == m.Note && rem.Get("links_json") == links && strings.Join(rem.Values["matched_discrepancies"], ",") == strings.Join(m.FindingIDs, ",") &&
			strings.Join(rem.Values["linked_documents"], ",") == strings.Join(linkedIDs, ",") {
			continue
		}
		if err := b.Update(ns, ModExpertise, rem.ID, v); err != nil {
			return written, err
		}
		if handover {
			it, _ := strconv.Atoi(rem.Get("iterations"))
			if _, err := b.Create(ns, ModRemarkHistory, snapshotValues(rem.ID, Snapshot{
				Iteration: max(it, 1), Answer: rem.Get("answer"), RejectReason: rem.Get("reject_reason"), Status: handedTo,
				SourceStatus: rem.Get("source_status"), At: time.Now().UTC(), Source: "автопередача на контроль: новая редакция загружена, расхождение устранено", Kind: HistAuto,
			})); err != nil {
				return written, err
			}
		}
		written++
	}
	_, _ = RecordSnapshot(b, ns, objectID, time.Now()) // best effort: the dynamics must not fail the check
	return written, nil
}
