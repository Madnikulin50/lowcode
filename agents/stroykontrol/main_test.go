package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/cmp"
)

func setupFixtures(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skip("poppler not installed")
	}
	fx, err := NewFixtureStore("fixtures")
	if err != nil {
		t.Skip(err)
	}
	raster = NewRasterizer(t.TempDir())
	fx.analyze = func(in []cmp.Input, mode string, pairing []cmp.ManualPair) (*cmp.Result, error) {
		return cmp.AnalyzeWith(in, t.TempDir(), cmp.AnalyzeOptions{Rasterize: rasterPage, OCR: cmp.NewOCR("tesseract", "rus", t.TempDir()), Mode: mode, Pairing: pairing})
	}
	compose = fx
	defaultNamespaceID = "x"
	suggester = cmp.NewSuggester("http://127.0.0.1:1", "")
}

func get(h http.HandlerFunc, url string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h(w, httptest.NewRequest(http.MethodGet, url, nil))
	return w
}

func post(h http.HandlerFunc, url string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	h(w, httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b)))
	return w
}

func TestPackComparisonEndToEnd(t *testing.T) {
	setupFixtures(t)
	w := get(handleComparison, "/api/comparison?recordID=pack-1&namespaceID=x")
	if w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	var out comparisonOut
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Status != "done" || len(out.Documents) != 3 {
		t.Fatalf("status %q docs %+v", out.Status, out.Documents)
	}
	sides := map[string]int{}
	for _, d := range out.Documents {
		sides[d.Side]++
	}
	if sides[cmp.SideRef] != 1 || sides[cmp.SideTarget] != 2 {
		t.Errorf("sides: %v", sides)
	}
	var depth *discrepancyRecord
	for i, d := range out.Discrepancies {
		if strings.Contains(d.Description, "Глубина заложения") {
			depth = &out.Discrepancies[i]
		}
		if d.Status != cmp.StatusNew || d.ID == "" {
			t.Errorf("discrepancy needs id and default status: %+v", d)
		}
	}
	if depth == nil || depth.DocRef != "pd-package.pdf" || depth.DocTarget != "rd-kzh.pdf" ||
		depth.PageRef != "1" || depth.PageTarget != "1" || cmp.ParseBox(depth.BboxRef) == nil || cmp.ParseBox(depth.BboxTarget) == nil {
		t.Fatalf("depth discrepancy lacks positions: %+v", depth)
	}
	// volume info appears on the documents once compared
	for _, d := range out.Documents {
		if d.Name == "pd-package.pdf" && !strings.Contains(d.VolumeCode, "КЖ") {
			t.Errorf("volume codes not reported: %+v", d)
		}
	}

	// page image and text of a package document
	if w := get(handlePage, "/api/page?recordID=pack-1&namespaceID=x&doc=rd-kzh.pdf&n=2"); w.Code != 200 || w.Header().Get("Content-Type") != "image/jpeg" {
		t.Errorf("page: %d %s", w.Code, w.Header().Get("Content-Type"))
	}
	if w := get(handleText, "/api/text?recordID=pack-1&namespaceID=x&doc=pd-package.pdf"); w.Code != 200 || !strings.Contains(w.Body.String(), "Глубина") {
		t.Errorf("text: %d", w.Code)
	}
	// a document id can never leave the case directory
	if w := get(handleFile, "/api/file?recordID=pack-1&namespaceID=x&doc=../pack-2/pd-is.pdf"); w.Code == 200 {
		t.Error("path traversal through doc id")
	}

	// review: confirm one finding, status sticks
	if w := post(handleDiscrepancyStatus, "/api/discrepancy/status", map[string]string{"recordID": "pack-1", "id": depth.ID, "status": "confirmed"}); w.Code != 200 {
		t.Fatalf("status: %d %s", w.Code, w.Body)
	}
	if w := post(handleDiscrepancyStatus, "/api/discrepancy/status", map[string]string{"recordID": "pack-1", "id": depth.ID, "status": "bogus"}); w.Code != 400 {
		t.Errorf("invalid status accepted: %d", w.Code)
	}
	w = get(handleComparison, "/api/comparison?recordID=pack-1&namespaceID=x")
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	for _, d := range out.Discrepancies {
		if d.ID == depth.ID && d.Status != cmp.StatusConfirmed {
			t.Errorf("status not kept: %+v", d)
		}
	}

	// remark card from the confirmed finding
	w = post(handleRemarkCreate, "/api/remark", map[string]any{"recordID": "pack-1", "discrepancyIDs": []string{depth.ID}})
	if w.Code != 200 {
		t.Fatalf("remark: %d %s", w.Code, w.Body)
	}
	var card cmp.Remark
	_ = json.Unmarshal(w.Body.Bytes(), &card)
	if card.ID == "" || card.Status != "draft" || !strings.Contains(card.WhatDiverges, "Глубина") ||
		!strings.Contains(card.DocumentRef, "pd-package.pdf") || !strings.Contains(card.DocumentRef, "rd-kzh.pdf") {
		t.Fatalf("card: %+v", card)
	}
	// edit + issue
	card.FinalText = "Привести глубину заложения в соответствие с ПД."
	card.Status = "issued"
	if w := post(handleRemarkSave, "/api/remark/save", map[string]any{"recordID": "pack-1", "remark": card}); w.Code != 200 {
		t.Fatalf("save: %d %s", w.Code, w.Body)
	}
	rs, _ := compose.Remarks("x", "pack-1")
	if len(rs) != 1 || rs[0].FinalText == "" || rs[0].Status != "issued" {
		t.Errorf("remarks: %+v", rs)
	}
	// the model is not configured: the button reports it, nothing is invented
	if w := post(handleRemarkSuggest, "/api/remark/suggest", map[string]string{"recordID": "pack-1", "id": card.ID}); w.Code != http.StatusServiceUnavailable {
		t.Errorf("suggest without a model: %d", w.Code)
	}
}

func TestFragmentFixtureHasNoErrors(t *testing.T) {
	setupFixtures(t)
	w := get(handleComparison, "/api/comparison?recordID=pack-2&namespaceID=x")
	var out comparisonOut
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	if out.Status != "done" {
		t.Fatalf("%s", w.Body)
	}
	for _, d := range out.Discrepancies {
		if d.Severity != "low" {
			t.Errorf("fragment fixture raised %s: %s", d.Severity, d.Description)
		}
	}
}

// Single-file fixtures written before packs existed keep working: legacy
// side=pd/rd addressing and their hand-written discrepancies.json.
func TestLegacyFixtureStillWorks(t *testing.T) {
	setupFixtures(t)
	w := get(handleComparison, "/api/comparison?recordID=demo-1&namespaceID=x")
	var out comparisonOut
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	if len(out.Documents) != 2 || len(out.Discrepancies) != 2 || out.Discrepancies[0].PageNumber != "1" {
		t.Fatalf("%s", w.Body)
	}
	if w := get(handlePages, "/api/pages?recordID=demo-1&namespaceID=x&side=pd"); w.Code != 200 || !strings.Contains(w.Body.String(), `"kind":"pdf"`) {
		t.Errorf("legacy side addressing: %d %s", w.Code, w.Body)
	}
}

func TestIterationInFixtures(t *testing.T) {
	setupFixtures(t)
	if _, err := os.Stat("fixtures/pack-1-v2"); err != nil {
		t.Skip("fixture pack-1-v2 missing")
	}
	// the reviewer confirms a finding of iteration 1
	w := get(handleComparison, "/api/comparison?recordID=pack-1&namespaceID=x")
	var v1 comparisonOut
	_ = json.Unmarshal(w.Body.Bytes(), &v1)
	var table *discrepancyRecord
	for i, d := range v1.Discrepancies {
		if strings.Contains(d.Description, "Ф-2") {
			table = &v1.Discrepancies[i]
		}
	}
	if table == nil {
		t.Fatal("no table finding in pack-1")
	}
	post(handleDiscrepancyStatus, "/api/discrepancy/status", map[string]string{"recordID": "pack-1", "id": table.ID, "status": "confirmed"})

	w = get(handleComparison, "/api/comparison?recordID=pack-1-v2&namespaceID=x")
	var v2 comparisonOut
	if err := json.Unmarshal(w.Body.Bytes(), &v2); err != nil {
		t.Fatal(err)
	}
	it := v2.Iteration
	if it == nil || it.ParentID != "pack-1" || it.No != "2" || it.Resolved < 2 || it.Persisting < 1 || it.New != 1 {
		t.Fatalf("iteration summary: %+v", it)
	}
	states := map[string]discrepancyRecord{}
	for _, d := range v2.Discrepancies {
		for _, k := range []string{"Глубина заложения", "Ф-2", "350"} {
			if strings.Contains(d.Description, k) {
				states[k] = d
			}
		}
	}
	if states["Глубина заложения"].IterState != "resolved" || states["350"].IterState != "new" {
		t.Errorf("states: %+v", states)
	}
	if d := states["Ф-2"]; d.IterState != "persisting" || d.Status != "confirmed" {
		t.Errorf("the unfixed finding must persist with the earlier decision: %+v", d)
	}
	if v1.Iteration != nil {
		t.Error("the first iteration has no parent")
	}
}

func TestExpertiseRemarksInFixtures(t *testing.T) {
	setupFixtures(t)
	if _, err := os.Stat("fixtures/pack-1/remarks.xlsx"); err != nil {
		t.Skip("run: go run fixtures/gen_remarks.go")
	}
	board := func(caseID string) (map[string]remarkRecord, *boardOut) {
		var out boardOut
		w := get(handleRemarksBoard, "/api/remarks/board?objectID="+caseID+"&namespaceID=x")
		if w.Code != 200 {
			t.Fatalf("%s: %d %s", caseID, w.Code, w.Body)
		}
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
			t.Fatal(err)
		}
		m := map[string]remarkRecord{}
		for _, r := range out.Remarks {
			m[r.Number] = r
		}
		return m, &out
	}
	v1, b1 := board("pack-1")
	if len(v1) != 5 || len(b1.Imports) != 1 || b1.Imports[0].Total != 5 {
		t.Fatalf("5 remarks and one import expected: %d / %+v", len(v1), b1.Imports)
	}
	// iteration 1: the differences the remarks complain about are still there, the answers say they were fixed
	for _, n := range []string{"30.1", "30.2", "30.3"} {
		e := v1[n]
		if e.Verdict != "still_differs" || len(e.Findings) == 0 || !e.ClaimMismatch || len(e.Documents) == 0 {
			t.Errorf("v1 remark %s: %+v", n, e)
		}
	}
	if v1["30.1"].Status != "answer_review" || v1["30.2"].Status != "resolved" || v1["30.3"].Status != "being_fixed" || v1["30.4"].Status != "awaiting_answer" {
		t.Errorf("statuses come from the export: %v %v %v %v", v1["30.1"].Status, v1["30.2"].Status, v1["30.3"].Status, v1["30.4"].Status)
	}
	if v1["30.1"].Criticality != "critical" || v1["30.3"].Iterations != 2 || v1["30.3"].PrevAnswer != "Принято" {
		t.Errorf("fields: %+v", v1["30.3"])
	}
	if v1["30.4"].Verdict != "no_match" || len(v1["30.4"].Documents) == 0 {
		t.Errorf("remark 4 names a loaded volume with no related difference: %+v", v1["30.4"])
	}
	if v1["30.5"].Verdict != "no_documents" {
		t.Errorf("remark 5 names a volume that is not loaded: %+v", v1["30.5"])
	}
	// iteration 2: depth and concrete grade were really fixed; the quantity was not, although "resolved"
	v2, _ := board("pack-1-v2")
	if v2["30.1"].Verdict != "resolved" || v2["30.3"].Verdict != "resolved" || v2["30.2"].Verdict != "still_differs" || !v2["30.2"].ClaimMismatch {
		t.Errorf("v2: 1=%s 3=%s 2=%s/%v", v2["30.1"].Verdict, v2["30.3"].Verdict, v2["30.2"].Verdict, v2["30.2"].ClaimMismatch)
	}

	// history: remark 3 is on its second iteration → the previous one and the current
	var hist []remarkHistoryRecord
	w := get(handleRemarkHistory, "/api/remarks/history?objectID=pack-1&namespaceID=x&remarkID="+v1["30.3"].ID)
	_ = json.Unmarshal(w.Body.Bytes(), &hist)
	if len(hist) != 2 || hist[0].Iteration != 1 || hist[0].Answer != "Принято" || hist[0].RejectReason == "" || hist[1].Iteration != 2 {
		t.Fatalf("history: %+v", hist)
	}

	// the specialist works the remark: status, assignee, due date, note
	upd := map[string]any{"objectID": "pack-1", "id": v1["30.1"].ID, "status": "resolved", "assignee": "Петров", "dueDate": "2026-11-01", "note": "проверено по ПД"}
	if w := post(handleRemarkUpdate, "/api/remarks/update", upd); w.Code != 200 {
		t.Fatalf("update: %d %s", w.Code, w.Body)
	}
	v1, _ = board("pack-1")
	if e := v1["30.1"]; e.Status != "resolved" || e.Assignee != "Петров" || e.DueDate != "2026-11-01" || e.Note == "" {
		t.Errorf("update not kept: %+v", e)
	}
	w = get(handleRemarkHistory, "/api/remarks/history?objectID=pack-1&namespaceID=x&remarkID="+v1["30.1"].ID)
	_ = json.Unmarshal(w.Body.Bytes(), &hist)
	if len(hist) < 2 || hist[len(hist)-1].Source != "вручную" || hist[len(hist)-1].Status != "resolved" {
		t.Errorf("a manual status change is a history entry: %+v", hist)
	}
	for _, bad := range []map[string]any{
		{"objectID": "pack-1", "id": v1["30.1"].ID, "status": "bogus"},
		{"objectID": "pack-1", "id": v1["30.1"].ID, "dueDate": "01.11.2026"},
	} {
		if w := post(handleRemarkUpdate, "/api/remarks/update", bad); w.Code != 400 {
			t.Errorf("invalid update accepted: %v → %d", bad, w.Code)
		}
	}
	// the comparison viewer shows the remarks that concern it
	var cmpOut comparisonOut
	w = get(handleComparison, "/api/comparison?recordID=pack-1&namespaceID=x")
	_ = json.Unmarshal(w.Body.Bytes(), &cmpOut)
	if len(cmpOut.Expertise) != 5 {
		t.Errorf("comparison panel: %d remarks", len(cmpOut.Expertise))
	}
}

func TestIssueListFromComparison(t *testing.T) {
	setupFixtures(t)
	var cmpOut comparisonOut
	_ = json.Unmarshal(get(handleComparison, "/api/comparison?recordID=pack-1&namespaceID=x").Body.Bytes(), &cmpOut)

	// nothing is confirmed yet: the list is empty and registering is refused with a reason
	var pre issueOut
	_ = json.Unmarshal(get(handleIssuePreview, "/api/issue/preview?recordID=pack-1&namespaceID=x").Body.Bytes(), &pre)
	if len(pre.Active) != 0 || pre.Confirmed != 0 {
		t.Fatalf("no confirmed findings, no remarks: %+v", pre)
	}
	if w := post(handleIssueRegister, "/api/issue/register?recordID=pack-1&namespaceID=x", map[string]string{}); w.Code != http.StatusUnprocessableEntity {
		t.Errorf("empty list registered: %d %s", w.Code, w.Body)
	}
	if w := get(handleIssueExport, "/api/issue/export?recordID=pack-1&namespaceID=x&format=xlsx"); w.Code != http.StatusUnprocessableEntity {
		t.Errorf("empty export: %d", w.Code)
	}

	// the specialist confirms the depth and the table quantity (both on the first two sheets)
	for _, sub := range []string{"Глубина заложения", "Ф-2", "Класс бетона"} {
		for _, d := range cmpOut.Discrepancies {
			if strings.Contains(d.Description, sub) {
				post(handleDiscrepancyStatus, "/api/discrepancy/status", map[string]string{"recordID": "pack-1", "id": d.ID, "status": "confirmed"})
			}
		}
	}
	_ = json.Unmarshal(get(handleIssuePreview, "/api/issue/preview?recordID=pack-1&namespaceID=x").Body.Bytes(), &pre)
	if len(pre.Active) < 2 || pre.Confirmed < 2 {
		t.Fatalf("confirmed findings make remarks: %+v", pre)
	}
	for i, is := range pre.Active {
		if is.Number != fmt.Sprint(i+1) || is.Text == "" || is.Refs == "" || is.Count == 0 {
			t.Errorf("remark %d: %+v", i, is)
		}
	}

	// into the registry: numbered ПТО-1.n, with the findings linked
	w := post(handleIssueRegister, "/api/issue/register?recordID=pack-1&namespaceID=x", map[string]string{})
	var reg cmp.IssueRegistered
	_ = json.Unmarshal(w.Body.Bytes(), &reg)
	if w.Code != 200 || reg.Created != len(pre.Active) || len(reg.Numbers) == 0 || reg.Numbers[0] != "ПТО-1.1" {
		t.Fatalf("register: %d %s", w.Code, w.Body)
	}
	var board boardOut
	_ = json.Unmarshal(get(handleRemarksBoard, "/api/remarks/board?objectID=pack-1&namespaceID=x").Body.Bytes(), &board)
	made := 0
	for _, r := range board.Remarks {
		if r.Origin == cmp.OriginComparison {
			made++
			if r.Status != "awaiting_answer" || len(r.Findings) == 0 || r.Verdict != "still_differs" {
				t.Errorf("generated remark: %+v", r)
			}
		}
	}
	if made != reg.Created {
		t.Errorf("registry holds %d generated remarks, created %d", made, reg.Created)
	}
	// the same list again updates, never duplicates; the export now carries the registry numbers
	w = post(handleIssueRegister, "/api/issue/register?recordID=pack-1&namespaceID=x", map[string]string{})
	_ = json.Unmarshal(w.Body.Bytes(), &reg)
	if reg.Created != 0 || reg.Updated != len(pre.Active) {
		t.Errorf("regeneration: %+v", reg)
	}

	// Excel in the format of an expertise export
	x := get(handleIssueExport, "/api/issue/export?recordID=pack-1&namespaceID=x&format=xlsx")
	if x.Code != 200 || !strings.Contains(x.Header().Get("Content-Disposition"), "filename*=UTF-8''") || !strings.Contains(x.Header().Get("Content-Type"), "spreadsheetml") {
		t.Fatalf("xlsx: %d %v", x.Code, x.Header())
	}
	rows, err := cmp.ParseExport(x.Body.Bytes(), "out.xlsx")
	if err != nil || len(rows) != len(pre.Active) || rows[0].Number != "ПТО-1.1" || rows[0].Status != cmp.RemarkAwaiting {
		t.Fatalf("the Excel must read back as an expertise export: %v %+v", err, rows)
	}
	// Word letter
	d := get(handleIssueExport, "/api/issue/export?recordID=pack-1&namespaceID=x&format=docx")
	if d.Code != 200 || !strings.Contains(d.Header().Get("Content-Type"), "wordprocessingml") || d.Body.Len() < 500 || !bytes.HasPrefix(d.Body.Bytes(), []byte("PK")) {
		t.Errorf("docx: %d %d bytes", d.Code, d.Body.Len())
	}
	if w := get(handleIssueExport, "/api/issue/export?recordID=pack-1&namespaceID=x&format=pdf"); w.Code != 400 {
		t.Errorf("unknown format: %d", w.Code)
	}
}

func TestVersionsComparisonUsesBeforeAfterNames(t *testing.T) {
	setupFixtures(t)
	w := get(handleComparison, "/api/comparison?recordID=versions-1&namespaceID=x")
	if w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	var out comparisonOut
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Kind != cmp.ModeVersions || out.RefName != "До" || out.TargetName != "После" {
		t.Fatalf("kind=%q names=%q/%q", out.Kind, out.RefName, out.TargetName)
	}
	if len(out.Discrepancies) == 0 {
		t.Fatal("two versions of the volume must differ")
	}
	for _, d := range out.Discrepancies {
		for _, bad := range []string{"(ПД", "(РД", " ПД,", " РД,"} {
			if strings.Contains(d.Description, bad) {
				t.Fatalf("stage names leaked into a versions comparison: %q", d.Description)
			}
		}
	}
	// the report speaks of «До»/«После» too
	rep, _, err := reportFor("x", "versions-1")
	if err != nil {
		t.Fatal(err)
	}
	if rep.RefName != "До" || rep.TargetName != "После" {
		t.Fatalf("report names %q/%q", rep.RefName, rep.TargetName)
	}
}

func TestPairingRoundTripInFixtures(t *testing.T) {
	setupFixtures(t)
	w := get(handlePairingGet, "/api/pairing?recordID=pack-1&namespaceID=x")
	if w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	var out pairingOut
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.State == nil || len(out.State.Pairs) != 2 || len(out.Documents) != 3 {
		t.Fatalf("state %+v docs %d", out.State, len(out.Documents))
	}
	var pdID string
	for _, d := range out.Documents {
		if d.Side == cmp.SideRef {
			pdID = d.ID
		}
	}
	// pair pages 1–2 of the ПД package with the whole РД КЖ file
	var rdID string
	for _, d := range out.Documents {
		if d.Name == "rd-kzh.pdf" {
			rdID = d.ID
		}
	}
	body := map[string]any{"namespaceID": "x", "recordID": "pack-1", "pairing": []cmp.ManualPair{
		{Ref: cmp.Slice{Doc: pdID, From: 1, To: 3}, Targets: []cmp.Slice{{Doc: rdID}}},
	}}
	if w := post(handlePairingSet, "/api/pairing", body); w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	w = get(handlePairingGet, "/api/pairing?recordID=pack-1&namespaceID=x")
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Pairing) != 1 || out.State == nil {
		t.Fatalf("%+v", out)
	}
	manual := 0
	for _, p := range out.State.Pairs {
		if p.Manual {
			manual++
			if p.Ref.From != 1 || p.Ref.To != 3 {
				t.Fatalf("manual pair: %+v", p)
			}
		}
	}
	if manual != 1 {
		t.Fatalf("the saved pair must be used by the next comparison: %+v", out.State.Pairs)
	}
	// a broken range is refused
	bad := map[string]any{"namespaceID": "x", "recordID": "pack-1", "pairing": []cmp.ManualPair{{Ref: cmp.Slice{Doc: pdID, From: 5, To: 2}, Targets: []cmp.Slice{{Doc: rdID}}}}}
	if w := post(handlePairingSet, "/api/pairing", bad); w.Code != 400 {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestDialogGivesTheModelFactsAndReturnsSheetJumps(t *testing.T) {
	setupFixtures(t)
	var system string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var body struct {
			Messages []cmp.Msg `json:"messages"`
		}
		_ = json.NewDecoder(req.Body).Decode(&body)
		system = body.Messages[0].Content
		out, _ := json.Marshal(map[string]any{"message": map[string]string{"content": "Глубина изменилась: [[ПД:1]] → [[РД:1]]; [[РД:77]] такого листа нет."}})
		_, _ = w.Write(out)
	}))
	defer srv.Close()
	suggester = cmp.NewSuggester(srv.URL, "m")

	var cmpOut comparisonOut
	w := get(handleComparison, "/api/comparison?recordID=pack-1&namespaceID=x")
	if err := json.Unmarshal(w.Body.Bytes(), &cmpOut); err != nil {
		t.Fatal(err)
	}
	var id string
	for _, d := range cmpOut.Discrepancies {
		if strings.Contains(d.Description, "Глубина заложения") {
			id = d.ID
		}
	}
	if id == "" {
		t.Fatal("no depth discrepancy")
	}
	w = post(handleDialog, "/api/dialog", map[string]any{"namespaceID": "x", "recordID": "pack-1", "discrepancyID": id,
		"messages": []cmp.Msg{{Role: "user", Content: "Объясни различие"}}})
	if w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	var out struct {
		Reply string     `json:"reply"`
		Jumps []cmp.Jump `json:"jumps"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Jumps) != 2 || out.Jumps[0].Doc != "pd-package.pdf" || out.Jumps[1].Doc != "rd-kzh.pdf" || strings.Contains(out.Reply, "77") {
		t.Fatalf("reply %q jumps %+v", out.Reply, out.Jumps)
	}
	for _, want := range []string{"Глубина заложения фундамента", "Текст листа: ПД стр. 1", "Текст листа: РД стр. 1"} {
		if !strings.Contains(system, want) {
			t.Fatalf("the model must be shown %q:\n%s", want, system)
		}
	}
	// without a model the panel gets a clear refusal, not a crash
	suggester = cmp.NewSuggester("http://127.0.0.1:1", "")
	w = post(handleDialog, "/api/dialog", map[string]any{"recordID": "pack-1", "messages": []cmp.Msg{{Role: "user", Content: "?"}}})
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", w.Code)
	}
}

func TestTokenRefreshFallsBackToTheTokenFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "dev-token")
	current := "expired"
	cur := func() string { return current }
	failing := func() (string, error) { return "", fmt.Errorf("node не найден") }

	rm := withTokenFileFallback(failing, file, cur)
	if _, err := rm(); err == nil || !strings.Contains(err.Error(), "node не найден") {
		t.Fatalf("no file → the minting error: %v", err)
	}
	_ = os.WriteFile(file, []byte("expired\n"), 0o600)
	if _, err := rm(); err == nil {
		t.Fatal("the same expired token in the file is not a refresh")
	}
	_ = os.WriteFile(file, []byte("fresh\n"), 0o600)
	if tok, err := rm(); err != nil || tok != "fresh" {
		t.Fatalf("a fresher token from the file must be taken: %q %v", tok, err)
	}
	ok := withTokenFileFallback(func() (string, error) { return "minted", nil }, file, cur)
	if tok, _ := ok(); tok != "minted" {
		t.Fatalf("minting wins over the file: %q", tok)
	}
}

func TestFindNodeWorksWithoutNodeOnPath(t *testing.T) {
	t.Setenv("PATH", "")
	if p, err := findNode(); err != nil {
		t.Skipf("no node anywhere on this machine: %v", err)
	} else if !strings.HasSuffix(p, "node") {
		t.Fatalf("%s", p)
	}
}

func TestAssistAnswersAboutTheSheetOnScreen(t *testing.T) {
	setupFixtures(t)
	skills := t.TempDir()
	_ = os.MkdirAll(filepath.Join(skills, "read-drawings"), 0o755)
	_ = os.WriteFile(filepath.Join(skills, "read-drawings", "SKILL.md"), []byte("---\nname: read-drawings\ndescription: x\n---\n# Навык\n\nШаг 1. Прочитай основную надпись.\n\nШаг 2. Найди оси."), 0o644)
	oldSkills := assistSkillsDir
	assistSkillsDir = skills
	defer func() { assistSkillsDir = oldSkills }()

	var system, model string
	var gotImages int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var body struct {
			Model    string `json:"model"`
			Messages []struct {
				Role, Content string
				Images        []string `json:"images"`
			} `json:"messages"`
		}
		_ = json.NewDecoder(req.Body).Decode(&body)
		model = body.Model
		system = body.Messages[0].Content
		gotImages = 0
		for _, m := range body.Messages {
			gotImages += len(m.Images)
		}
		if body.Model == "broken-vision" {
			http.Error(w, "model not found", 404)
			return
		}
		out, _ := json.Marshal(map[string]any{"message": map[string]string{"content": "На листе — конструктивные решения, см. [[ПД:1]]."}})
		_, _ = w.Write(out)
	}))
	defer srv.Close()
	suggester = cmp.NewSuggester(srv.URL, "text-model")
	assistVision = cmp.NewSuggester(srv.URL, "vision-model")

	ask := func(withImage bool) (int, map[string]any) {
		w := post(handleAssist, "/api/assist", map[string]any{
			"namespaceID": "x", "messages": []cmp.Msg{{Role: "user", Content: "Что на этом листе?"}},
			"context": map[string]any{"source": "comparison", "recordID": "pack-1", "scope": "both", "withImage": withImage,
				"ref": map[string]any{"doc": "pd-package.pdf", "page": 1}, "target": map[string]any{"doc": "rd-kzh.pdf", "page": 1}},
		})
		var out map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return w.Code, out
	}

	code, out := ask(false)
	if code != 200 {
		t.Fatalf("%d %v", code, out)
	}
	for _, want := range []string{"Открытый лист: ПД: pd-package.pdf, стр. 1 из 7", "Открытый лист: РД: rd-kzh.pdf, стр. 1 из 3", "Глубина заложения фундамента"} {
		if !strings.Contains(system, want) {
			t.Fatalf("facts lack %q:\n%s", want, system)
		}
	}
	if gotImages != 0 || model != "text-model" || strings.Contains(system, "Методика чтения чертежей") {
		t.Fatalf("a text page needs no picture and no drawing method: images=%d model=%s", gotImages, model)
	}
	if jumps, _ := out["jumps"].([]any); len(jumps) != 1 {
		t.Fatalf("%v", out)
	}

	// on request the picture goes to the multimodal model together with the drawing method
	code, out = ask(true)
	if code != 200 || gotImages != 2 || model != "vision-model" || !strings.Contains(system, "Шаг 1. Прочитай основную надпись") || strings.Contains(system, "description: x") {
		t.Fatalf("code=%d images=%d model=%s\n%s", code, gotImages, model, system)
	}
	if out["image"] != true || out["skill"] != true {
		t.Fatalf("%v", out)
	}

	// the picture model is missing: the answer comes from the text, and says so
	assistVision = cmp.NewSuggester(srv.URL, "broken-vision")
	code, out = ask(true)
	if code != 200 || !strings.Contains(out["reply"].(string), "ответ только по тексту листа") || model != "text-model" {
		t.Fatalf("%d %v model=%s", code, out, model)
	}

	// a picture model that does not answer in time: the answer comes from the text after the limit
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) { time.Sleep(2 * time.Second) }))
	defer slow.Close()
	assistVision = cmp.NewSuggester(slow.URL, "slow")
	oldTimeout := assistVisionTimeout
	assistVisionTimeout = 300 * time.Millisecond
	defer func() { assistVisionTimeout = oldTimeout }()
	started := time.Now()
	code, out = ask(true)
	if code != 200 || !strings.Contains(out["reply"].(string), "ответ только по тексту листа") || time.Since(started) > 1500*time.Millisecond {
		t.Fatalf("%d %v after %v", code, out, time.Since(started))
	}

	// without any picture model the facts say there is no picture
	assistVision = cmp.NewSuggester(srv.URL, "")
	ask(true)
	if !strings.Contains(system, "Изображение листа не передано") {
		t.Fatalf("%s", system)
	}
	suggester = cmp.NewSuggester("http://127.0.0.1:1", "")
	if w := post(handleAssist, "/api/assist", map[string]any{"messages": []cmp.Msg{{Role: "user", Content: "?"}}}); w.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", w.Code)
	}
}

func TestComparisonFileOpensInTheDocumentViewer(t *testing.T) {
	setupFixtures(t)
	w := get(handleAssistDoc, "/api/assist/doc?source=comparison_doc&recordID=pack-1&id=pd-package.pdf&namespaceID=x")
	if w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	var info struct {
		Name      string
		PageCount int
		Title     string
	}
	if err := json.Unmarshal(w.Body.Bytes(), &info); err != nil {
		t.Fatal(err)
	}
	if info.Name != "pd-package.pdf" || info.PageCount != 7 {
		t.Fatalf("%+v", info)
	}
	w = get(handleAssistPage, "/api/assist/page?source=comparison_doc&recordID=pack-1&id=pd-package.pdf&namespaceID=x&n=2")
	if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), "image/") {
		t.Fatalf("page image: %d %s", w.Code, w.Header().Get("Content-Type"))
	}
	if w := get(handleAssistDoc, "/api/assist/doc?source=comparison_doc&id=pd-package.pdf&namespaceID=x"); w.Code == 200 {
		t.Fatal("a comparison file needs its recordID")
	}
}

func TestAssistAsyncAnswerIsCollectedByPolling(t *testing.T) {
	setupFixtures(t)
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		<-release // a slow model: longer than the gateway lets a request wait
		out, _ := json.Marshal(map[string]any{"message": map[string]string{"content": "Готово."}})
		_, _ = w.Write(out)
	}))
	defer srv.Close()
	suggester = cmp.NewSuggester(srv.URL, "m")
	w := post(handleAssist, "/api/assist", map[string]any{"namespaceID": "x", "async": true, "messages": []cmp.Msg{{Role: "user", Content: "?"}},
		"context": map[string]any{"source": "comparison", "recordID": "pack-1", "scope": "ref", "ref": map[string]any{"doc": "pd-package.pdf", "page": 1}}})
	if w.Code != http.StatusAccepted {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	var job struct{ JobID string }
	_ = json.Unmarshal(w.Body.Bytes(), &job)
	state := func() map[string]any {
		var out map[string]any
		_ = json.Unmarshal(get(handleAssistResult, "/api/assist/result?job="+job.JobID).Body.Bytes(), &out)
		return out
	}
	if st := state(); st["status"] != "running" {
		t.Fatalf("the model is still thinking: %v", st)
	}
	close(release)
	var st map[string]any
	for i := 0; i < 100; i++ {
		if st = state(); st["status"] != "running" {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if st["status"] != "done" || st["reply"] != "Готово." {
		t.Fatalf("%v", st)
	}
	if st := func() map[string]any {
		var o map[string]any
		_ = json.Unmarshal(get(handleAssistResult, "/api/assist/result?job=nope").Body.Bytes(), &o)
		return o
	}(); st["status"] != "error" {
		t.Fatalf("an unknown job is an error the client can show: %v", st)
	}
}
