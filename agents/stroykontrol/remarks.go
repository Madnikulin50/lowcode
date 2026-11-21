package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/cmp"
	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// --- registry of expertise remarks in Compose --------------------------------

func remarkFromIDRecord(r idcheck.Record) remarkRecord {
	it, _ := strconv.Atoi(r.Get("iterations"))
	rec := remarkRecord{
		ID: r.ID, Key: r.Get("remark_key"), Number: r.Get("number"), Batch: r.Get("batch"), Text: r.Get("text"), Refs: r.Get("refs_text"),
		Section: r.Get("section"), Basis: r.Get("basis"), Answer: r.Get("answer"), RejectReason: r.Get("reject_reason"),
		PrevAnswer: r.Get("prev_answer"), PrevRejectReason: r.Get("prev_reject_reason"), Discipline: r.Get("discipline"),
		Expert: r.Get("expert"), Criticality: r.Get("criticality"), Proposals: r.Get("proposals"), Expected: r.Get("expected_result"),
		Iterations: it, ChangedAt: r.Get("source_changed_at"), SourceStatus: r.Get("source_status"), Status: r.Get("status"),
		Assignee: r.Get("assignee"), DueDate: r.Get("due_date"), Note: r.Get("pto_note"), EstimateH: r.Get("estimate_h"), Verdict: r.Get("verdict"),
		Score: r.Get("match_score"), MatchNote: r.Get("match_note"), ClaimMismatch: r.Get("claim_mismatch") != "",
		Archived: r.Get("archived") != "", Origin: r.Get("origin"),
		AIVerdict: r.Get("ai_verdict"), AINote: r.Get("ai_note"), AIAt: r.Get("ai_at"),
		OutcomeDecision: r.Get("outcome_decision"), OutcomeFinal: r.Get("outcome_final"), OutcomeComment: r.Get("outcome_comment"),
		OutcomeBy: r.Get("outcome_by"), OutcomeAt: r.Get("outcome_at"),
		Findings: nonEmpty(r.Values["matched_discrepancies"]), Documents: nonEmpty(r.Values["linked_documents"]),
	}
	set, _ := cmp.ParseRequirementSet(r.Get("requirements_json"))
	rec.Requirements, rec.RequirementsStale = set.Items, set.Stale
	if rec.Requirements == nil {
		rec.Requirements = []cmp.Requirement{}
	}
	rec.Materials, rec.MaterialsState = cmp.ParseMaterialsJSON(r.Get("materials_json")), r.Get("materials_state")
	if rec.Materials == nil {
		rec.Materials = []cmp.Material{}
	}
	rec.fillOutcome()
	_ = json.Unmarshal([]byte(r.Get("links_json")), &rec.Links)
	_ = json.Unmarshal([]byte(r.Get("ai_cites_json")), &rec.AICites)
	if rec.AICites == nil {
		rec.AICites = []cmp.AICite{}
	}
	if rec.Links == nil {
		rec.Links = []cmp.RemarkLocation{}
	}
	_ = json.Unmarshal([]byte(r.Get("volume_refs")), &rec.VolumeRefs)
	if rec.VolumeRefs == nil {
		rec.VolumeRefs = []cmp.VolumeRef{}
	}
	return rec
}

func nonEmpty(in []string) []string {
	out := []string{}
	for _, s := range in {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// sortRemarks orders by the letter number and the index inside it, numerically.
func sortRemarks(l []remarkRecord) {
	part := func(s string) (int, int, string) {
		a, b, _ := strings.Cut(s, ".")
		x, ea := strconv.Atoi(a)
		y, eb := strconv.Atoi(b)
		if ea != nil {
			x = 1 << 30
		}
		if eb != nil {
			y = 1 << 30
		}
		return x, y, s
	}
	sort.SliceStable(l, func(i, j int) bool {
		a1, a2, as := part(l[i].Number)
		b1, b2, bs := part(l[j].Number)
		if a1 != b1 {
			return a1 < b1
		}
		if a2 != b2 {
			return a2 < b2
		}
		return as < bs
	})
}

func (c *ComposeClient) Board(ns, objectID string) (*boardOut, error) {
	b := composeBackend{c}
	obj, err := c.getRecord(ns, cmp.ModObjects, objectID)
	if err != nil {
		return nil, err
	}
	recs, err := b.Search(ns, cmp.ModExpertise, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return nil, err
	}
	out := &boardOut{ObjectID: objectID, ObjectName: obj.val("name"), Remarks: make([]remarkRecord, 0, len(recs)), Imports: []importRecord{}, Comparisons: []boardComparison{}, Documents: map[string]string{}}
	need := map[string]bool{}
	for _, r := range recs {
		rr := remarkFromIDRecord(r)
		out.Remarks = append(out.Remarks, rr)
		for _, d := range rr.Documents {
			need[d] = true
		}
	}
	sortRemarks(out.Remarks)
	for id := range need {
		if d, err := c.getRecord(ns, cmp.ModDocuments, id); err == nil {
			out.Documents[id] = strings.TrimSpace(d.val("volume_code") + " " + d.val("volume_title"))
		}
	}
	imps, err := b.Search(ns, cmp.ModExpertImports, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return nil, err
	}
	for _, i := range imps {
		n := func(k string) int { v, _ := strconv.Atoi(i.Get(k)); return v }
		out.Imports = append(out.Imports, importRecord{ID: i.ID, File: i.Get("file"), At: i.Get("imported_at"), Total: n("total"), New: n("new_count"),
			Changed: n("changed_count"), StatusChanges: n("status_changes"), Reopened: n("reopened"), Archived: n("archived_count"), Details: i.Get("details")})
	}
	sort.Slice(out.Imports, func(i, j int) bool { return out.Imports[i].At > out.Imports[j].At })
	comps, err := b.Search(ns, comparisonModule, fmt.Sprintf("object = '%s'", objectID))
	if err == nil {
		for _, cp := range comps {
			out.Comparisons = append(out.Comparisons, boardComparison{ID: cp.ID, Title: cp.Get("title")})
		}
	}
	out.Dynamics = []cmp.StatusSnapshot{}
	if dyn, err := cmp.Dynamics(b, ns, objectID); err == nil { // a namespace without the module has no dynamics yet
		out.Dynamics = dyn
	}
	return out, nil
}

func (c *ComposeClient) RemarkHistory(ns, objectID, remarkID string) ([]remarkHistoryRecord, error) {
	r, err := c.getRecord(ns, cmp.ModExpertise, remarkID)
	if err != nil {
		return nil, err
	}
	if r.val("object") != objectID {
		return nil, fmt.Errorf("remark %s does not belong to object %s", remarkID, objectID)
	}
	recs, err := composeBackend{c}.Search(ns, cmp.ModRemarkHistory, fmt.Sprintf("remark = '%s'", remarkID))
	if err != nil {
		return nil, err
	}
	out := make([]remarkHistoryRecord, 0, len(recs))
	for _, h := range recs {
		it, _ := strconv.Atoi(h.Get("iteration"))
		out = append(out, remarkHistoryRecord{Iteration: it, Answer: h.Get("answer"), RejectReason: h.Get("reject_reason"), Status: h.Get("status"),
			SourceStatus: h.Get("source_status"), At: h.Get("at"), Author: h.Get("author"), Source: h.Get("source"), Note: h.Get("note"), Kind: h.Get("kind")})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].At != out[j].At {
			return out[i].At < out[j].At
		}
		return out[i].Iteration < out[j].Iteration
	})
	return out, nil
}

func valOr(p *string, def string) string {
	if p != nil {
		return *p
	}
	return def
}

func validDate(s string) bool {
	if s == "" {
		return true
	}
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

func (c *ComposeClient) UpdateRemark(ns, objectID, remarkID string, p remarkPatch) (*remarkRecord, error) {
	b := composeBackend{c}
	cur, err := c.getRecord(ns, cmp.ModExpertise, remarkID)
	if err != nil {
		return nil, err
	}
	if cur.val("object") != objectID {
		return nil, fmt.Errorf("remark %s does not belong to object %s", remarkID, objectID)
	}
	if err := p.check(cur.val("status"), cur.val("assignee")); err != nil {
		return nil, err
	}
	by := p.Actor.Label()
	v := idcheck.Values{}
	statusComment := ""
	if p.StatusComment != nil {
		statusComment = strings.TrimSpace(*p.StatusComment)
	}
	if p.Status != nil {
		if err := cmp.CheckStatusChange(cur.val("status"), *p.Status, statusComment); err != nil {
			return nil, err
		}
		v["status"] = []string{*p.Status}
	}
	if p.EstimateH != nil {
		est, err := cmp.ParseEstimate(*p.EstimateH)
		if err != nil {
			return nil, err
		}
		v["estimate_h"] = []string{est}
	}
	if p.Assignee != nil {
		v["assignee"] = []string{strings.TrimSpace(*p.Assignee)}
	}
	if p.DueDate != nil {
		if !validDate(*p.DueDate) {
			return nil, fmt.Errorf("dueDate must be YYYY-MM-DD")
		}
		v["due_date"] = []string{*p.DueDate}
	}
	if p.Note != nil {
		v["pto_note"] = []string{*p.Note}
	}
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	var decisionHistory *idcheck.Values
	if p.OutcomeFinal != nil || p.OutcomeComment != nil {
		raw, err := b.GetRecord(ns, cmp.ModExpertise, remarkID)
		if err != nil {
			return nil, err
		}
		cur := remarkFromIDRecord(*raw)
		final, comment := cur.OutcomeFinal, cur.OutcomeComment
		if p.OutcomeFinal != nil {
			final = *p.OutcomeFinal
		}
		if p.OutcomeComment != nil {
			comment = strings.TrimSpace(*p.OutcomeComment)
		}
		kind, err := cmp.ResolveDecision(cur.Outcome, final, comment)
		if err != nil {
			return nil, err
		}
		v["outcome_final"] = []string{final}
		v["outcome_decision"] = []string{kind}
		v["outcome_comment"] = []string{comment}
		if final == "" {
			v["outcome_by"], v["outcome_at"] = []string{""}, []string{""}
		} else {
			v["outcome_by"], v["outcome_at"] = []string{by}, []string{now}
			hv := idcheck.V("remark", remarkID, "source", "решение по проверке: "+cmp.OutcomeLabels[final], "author", by,
				"note", comment, "status", cur.Status, "source_status", cur.SourceStatus, "kind", cmp.HistDecision)
			it := max(cur.Iterations, 1)
			hv["iteration"] = []string{strconv.Itoa(it)}
			hv["at"] = []string{now}
			decisionHistory = &hv
		}
	}
	if err := b.Update(ns, cmp.ModExpertise, remarkID, v); err != nil {
		return nil, err
	}
	if decisionHistory != nil {
		if _, err := b.Create(ns, cmp.ModRemarkHistory, *decisionHistory); err != nil {
			return nil, err
		}
	}
	// assignments, due dates and notes are history too
	for _, wc := range cmp.WorkChanges(cur.val("assignee"), cur.val("due_date"), cur.val("pto_note"), p.Assignee, p.DueDate, p.Note) {
		hv := idcheck.V("remark", remarkID, "status", firstNonEmpty(valOr(p.Status, cur.val("status"))), "source", wc.Source, "author", by,
			"note", wc.Note, "kind", wc.Kind, "source_status", cur.val("source_status"))
		it, _ := strconv.Atoi(cur.val("iterations"))
		hv["iteration"] = []string{strconv.Itoa(max(it, 1))}
		hv["at"] = []string{now}
		if _, err := b.Create(ns, cmp.ModRemarkHistory, hv); err != nil {
			return nil, err
		}
	}
	if p.Status != nil && *p.Status != cur.val("status") {
		it, _ := strconv.Atoi(cur.val("iterations"))
		hv := idcheck.V("remark", remarkID, "status", *p.Status, "source", "вручную", "author", by, "note", statusComment, "kind", cmp.HistStatus,
			"source_status", cur.val("source_status"), "answer", cur.val("answer"), "reject_reason", cur.val("reject_reason"))
		hv["iteration"] = []string{strconv.Itoa(max(it, 1))}
		hv["at"] = []string{time.Now().UTC().Format("2006-01-02T15:04:05Z")}
		if _, err := b.Create(ns, cmp.ModRemarkHistory, hv); err != nil {
			return nil, err
		}
	}
	rec, err := b.GetRecord(ns, cmp.ModExpertise, remarkID)
	if err != nil {
		return nil, err
	}
	out := remarkFromIDRecord(*rec)
	return &out, nil
}

// SaveRequirements implements store.
func (c *ComposeClient) SaveRequirements(ns, objectID, remarkID string, in []cmp.RequirementInput, by cmp.Actor) (*remarkRecord, error) {
	b := composeBackend{c}
	cur, err := c.getRecord(ns, cmp.ModExpertise, remarkID)
	if err != nil {
		return nil, err
	}
	if cur.val("object") != objectID {
		return nil, fmt.Errorf("remark %s does not belong to object %s", remarkID, objectID)
	}
	old, _ := cmp.ParseRequirementSet(cur.val("requirements_json"))
	if err := checkRequirementEdit(by, old, in, cur.val("status"), cur.val("assignee")); err != nil {
		return nil, err
	}
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	set, err := cmp.ApplyRequirementEdit(old, cur.val("text"), in, by.Label(), now)
	if err != nil {
		return nil, err
	}
	if err := b.Update(ns, cmp.ModExpertise, remarkID, idcheck.V("requirements_json", set.JSON())); err != nil {
		return nil, err
	}
	it, _ := strconv.Atoi(cur.val("iterations"))
	for _, wc := range cmp.RequirementChanges(old, set) {
		hv := idcheck.V("remark", remarkID, "status", cur.val("status"), "source", wc.Source, "author", by.Label(), "note", wc.Note, "kind", wc.Kind,
			"source_status", cur.val("source_status"))
		hv["iteration"] = []string{strconv.Itoa(max(it, 1))}
		hv["at"] = []string{now}
		if _, err := b.Create(ns, cmp.ModRemarkHistory, hv); err != nil {
			return nil, err
		}
	}
	rec, err := b.GetRecord(ns, cmp.ModExpertise, remarkID)
	if err != nil {
		return nil, err
	}
	out := remarkFromIDRecord(*rec)
	return &out, nil
}

// Expertise: the remarks of the comparison's object that concern this
// comparison — linked to one of its documents or matched to one of its findings.
func (c *ComposeClient) Expertise(ns, comparisonID string) ([]remarkRecord, error) {
	cmpRec, err := c.getRecord(ns, comparisonModule, comparisonID)
	if err != nil {
		return nil, err
	}
	objectID := cmpRec.val("object")
	if objectID == "" {
		return []remarkRecord{}, nil
	}
	recs, err := composeBackend{c}.Search(ns, cmp.ModExpertise, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return nil, err
	}
	discs, err := c.Discrepancies(ns, comparisonID)
	if err != nil {
		return nil, err
	}
	docs, err := c.Documents(ns, comparisonID)
	if err != nil {
		return nil, err
	}
	mine := map[string]bool{}
	for _, d := range discs {
		mine[d.ID] = true
	}
	for _, d := range docs {
		mine[d.ID] = true
	}
	out := []remarkRecord{}
	for _, r := range recs {
		rr := remarkFromIDRecord(r)
		if rr.Archived {
			continue
		}
		for _, id := range append(append([]string{}, rr.Findings...), rr.Documents...) {
			if mine[id] {
				out = append(out, rr)
				break
			}
		}
	}
	sortRemarks(out)
	return out, nil
}

// --- HTTP handlers -----------------------------------------------------------

func objectFromQuery(r *http.Request) (ns, objectID string, err error) {
	q := r.URL.Query()
	ns = firstNonEmpty(q.Get("namespaceID"), defaultNamespaceID)
	objectID = strings.TrimSpace(q.Get("objectID"))
	if ns == "" || objectID == "" {
		return "", "", fmt.Errorf("objectID and namespaceID are required")
	}
	return
}

func handleRemarksBoard(w http.ResponseWriter, r *http.Request) {
	ns, obj, err := objectFromQuery(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	out, err := compose.Board(ns, obj)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	out.TaskCounts = map[string]cmp.TaskCount{}
	if taskB != nil { // a namespace without the module of tasks has none yet
		if ts, err := cmp.ListTasks(taskB, ns, cmp.TaskFilter{Object: obj}, time.Now()); err == nil {
			out.TaskCounts = cmp.CountTasks(ts)
		}
	}
	if cmpRunner != nil {
		if st := cmpRunner.Status("object:" + obj); st != "" {
			w.Header().Set("X-Import-Status", st)
		}
		if st := cmpRunner.Status("ai:" + obj); st != "" {
			w.Header().Set("X-AI-Status", st)
		}
		if e := aiResultOf(obj); e != "" {
			w.Header().Set("X-AI-Error", url.PathEscape(e))
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func handleRemarkHistory(w http.ResponseWriter, r *http.Request) {
	ns, obj, err := objectFromQuery(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	id := r.URL.Query().Get("remarkID")
	if id == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("remarkID is required"))
		return
	}
	out, err := compose.RemarkHistory(ns, obj, id)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func handleRemarkUpdate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NamespaceID string `json:"namespaceID"`
		ObjectID    string `json:"objectID"`
		ID          string `json:"id"`
		remarkPatch
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	ns := firstNonEmpty(body.NamespaceID, defaultNamespaceID)
	if ns == "" || body.ObjectID == "" || body.ID == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("objectID and id are required"))
		return
	}
	by, err := actorOf(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err)
		return
	}
	body.remarkPatch.Actor = by
	out, err := compose.UpdateRemark(ns, body.ObjectID, body.ID, body.remarkPatch)
	if err == nil && cmpRunner != nil && (body.Status != nil || body.DueDate != nil) {
		_, _ = cmp.RecordSnapshot(cmpRunner.Backend, ns, body.ObjectID, time.Now()) // best effort: the dynamics
	}
	if err == nil && (body.Status != nil || body.Assignee != nil || body.DueDate != nil) {
		notif.soon(ns, body.ObjectID) // the assignee hears of it now, not at the next pass
	}
	if err != nil {
		code := http.StatusBadGateway
		var fb forbiddenError
		if errors.As(err, &fb) {
			code = http.StatusForbidden
		} else if strings.Contains(err.Error(), "unknown status") || strings.Contains(err.Error(), "dueDate") ||
			strings.Contains(err.Error(), "outcome") || strings.Contains(err.Error(), "statusComment") || strings.Contains(err.Error(), "estimateH") {
			code = http.StatusBadRequest
		}
		writeErr(w, code, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleRemarkRequirements — POST /api/remarks/requirements: the specialist's list of the requirements of a remark
// (with the decisions on them). The check then runs again, so that the new and edited requirements get their results.
func handleRemarkRequirements(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NamespaceID  string                 `json:"namespaceID"`
		ObjectID     string                 `json:"objectID"`
		ID           string                 `json:"id"`
		Requirements []cmp.RequirementInput `json:"requirements"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	ns := firstNonEmpty(body.NamespaceID, defaultNamespaceID)
	if ns == "" || body.ObjectID == "" || body.ID == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("objectID and id are required"))
		return
	}
	by, err := actorOf(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err)
		return
	}
	out, err := compose.SaveRequirements(ns, body.ObjectID, body.ID, body.Requirements, by)
	if err != nil {
		code := http.StatusBadGateway
		var fb forbiddenError
		if errors.As(err, &fb) {
			code = http.StatusForbidden
		} else if strings.Contains(err.Error(), "requirements") || strings.Contains(err.Error(), "outcome") {
			code = http.StatusBadRequest
		}
		writeErr(w, code, err)
		return
	}
	if cmpRunner != nil { // the results of the new and edited requirements
		if _, err := cmpRunner.CheckObjectRemarks(r.Context(), ns, body.ObjectID); err == nil {
			if again, err := compose.UpdateRemark(ns, body.ObjectID, body.ID, remarkPatch{}); err == nil {
				out = again
			}
		}
	}
	writeJSON(w, http.StatusOK, out)
}

// handleExpertiseImport imports the exports attached to an object
// (?objectID=&force=1). Called by the workbench and by the rule chain
// "stroykontrol-import-remarks".
func handleExpertiseImport(w http.ResponseWriter, r *http.Request) {
	if cmpRunner == nil {
		writeJSON(w, http.StatusOK, map[string]any{"note": "in --fixtures mode remarks are read from remarks*.xlsx of the case on every request"})
		return
	}
	ns, obj, err := objectFromQuery(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	sum, err := cmpRunner.ImportObjectRemarks(r.Context(), ns, obj, r.URL.Query().Get("force") != "")
	if err != nil {
		writeErr(w, http.StatusUnprocessableEntity, err)
		return
	}
	writeJSON(w, http.StatusOK, sum)
}

// handleExpertiseCheck re-links the remarks of an object with the documents and
// re-checks the answers against the findings.
func handleExpertiseCheck(w http.ResponseWriter, r *http.Request) {
	if cmpRunner == nil {
		writeJSON(w, http.StatusOK, map[string]any{"note": "in --fixtures mode the check runs on every request"})
		return
	}
	ns, obj, err := objectFromQuery(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	n, err := cmpRunner.CheckObjectRemarks(r.Context(), ns, obj)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"updated": n})
}

var (
	aiResultMu sync.Mutex
	aiResults  = map[string]string{} // object ID → the last error of a background AI reading ("" = fine)
)

func setAIResult(obj, msg string) {
	aiResultMu.Lock()
	aiResults[obj] = msg
	aiResultMu.Unlock()
}

func aiResultOf(obj string) string {
	aiResultMu.Lock()
	defer aiResultMu.Unlock()
	return aiResults[obj]
}

// handleRemarkAI lets the model read ambiguous remarks against the fragments of their volumes. They go to the background — a local model needs minutes, longer than the gateway waits for
// one request; progress is in X-AI-Status of the board, a failure in X-AI-Error.
func handleRemarkAI(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NamespaceID string   `json:"namespaceID"`
		ObjectID    string   `json:"objectID"`
		ID          string   `json:"id"`
		IDs         []string `json:"ids"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	ns := firstNonEmpty(body.NamespaceID, defaultNamespaceID)
	ids := nonEmpty(append([]string{body.ID}, body.IDs...))
	if ns == "" || body.ObjectID == "" || len(ids) == 0 {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("objectID and id (or ids) are required"))
		return
	}
	if cmpRunner == nil {
		writeErr(w, http.StatusNotImplemented, fmt.Errorf("ИИ-разбор замечаний работает только с Compose (не в режиме --fixtures)"))
		return
	}
	if !cmpRunner.LLM.Enabled() {
		writeErr(w, http.StatusUnprocessableEntity, fmt.Errorf("модель не настроена: задайте REMARK_MODEL или --remark-model"))
		return
	}
	for _, id := range ids {
		rec, err := cmpRunner.Backend.GetRecord(ns, cmp.ModExpertise, id)
		if err != nil || rec.Get("object") != body.ObjectID {
			writeErr(w, http.StatusBadRequest, fmt.Errorf("замечание %s не относится к объекту %s", id, body.ObjectID))
			return
		}
	}
	const maxBatch = 25
	if len(ids) > maxBatch {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("за один раз можно разобрать не более %d замечаний", maxBatch))
		return
	}
	key := "ai:" + body.ObjectID
	if cmpRunner.Status(key) != "" {
		writeErr(w, http.StatusConflict, fmt.Errorf("разбор уже идёт: %s", cmpRunner.Status(key)))
		return
	}
	setAIResult(body.ObjectID, "")
	go func() {
		done, failed, err := cmpRunner.AIBatch(context.Background(), ns, key, ids)
		log.Printf("remarks ai: object %s: done %d, failed %d, first error: %v", body.ObjectID, done, failed, err)
		if err != nil {
			setAIResult(body.ObjectID, fmt.Sprintf("ИИ-разбор: удалось %d из %d; ошибка: %s", done, len(ids), trimTo(err.Error(), 200)))
		}
	}()
	writeJSON(w, http.StatusAccepted, map[string]any{"queued": len(ids)})
}

// checkRequirementEdit: a decision on a requirement is a confirmation (ТЗ 3.8.4); the rest of the list is work.
func checkRequirementEdit(by cmp.Actor, old cmp.RequirementSet, in []cmp.RequirementInput, status, assignee string) error {
	prev := map[string]string{}
	for _, q := range old.Items {
		prev[strings.TrimSpace(q.Text)] = q.OutcomeFinal
	}
	c := cmp.RemarkChange{FromStatus: status, Assignee: assignee, WorkFields: true}
	for _, q := range in {
		if q.OutcomeFinal != prev[strings.TrimSpace(q.Text)] {
			c.OutcomeDecision = true
		}
	}
	if err := by.CheckRemarkChange(c); err != nil {
		return forbiddenError{err}
	}
	return nil
}
