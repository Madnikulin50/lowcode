package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"

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
	fx.analyze = func(in []cmp.Input) (*cmp.Result, error) {
		return cmp.AnalyzeWith(in, t.TempDir(), cmp.AnalyzeOptions{Rasterize: rasterPage, OCR: cmp.NewOCR("tesseract", "rus", t.TempDir())})
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
	byNum := func(url string) map[string]expertiseRecord {
		var out comparisonOut
		w := get(handleComparison, url)
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
			t.Fatal(err)
		}
		m := map[string]expertiseRecord{}
		for _, e := range out.Expertise {
			m[e.Number] = e
		}
		return m
	}
	v1 := byNum("/api/comparison?recordID=pack-1&namespaceID=x")
	if len(v1) != 4 {
		t.Fatalf("4 remarks expected, got %d", len(v1))
	}
	// iteration 1: the differences the remarks complain about are still there; the answers say they were fixed
	for _, n := range []string{"1", "2", "3"} {
		e := v1[n]
		if e.Verdict != "still_differs" || len(e.Findings) == 0 || !e.ClaimMismatch || e.Status != "open" {
			t.Errorf("v1 remark %s: %+v", n, e)
		}
	}
	if v1["4"].Verdict != "no_match" {
		t.Errorf("remark 4 is not about a ПД/РД difference: %+v", v1["4"])
	}
	// iteration 2: depth and concrete grade were really fixed → resolved; the quantity was not
	v2 := byNum("/api/comparison?recordID=pack-1-v2&namespaceID=x")
	if v2["1"].Verdict != "resolved" || v2["3"].Verdict != "resolved" {
		t.Errorf("v2 fixed remarks: 1=%+v 3=%+v", v2["1"], v2["3"])
	}
	if v2["2"].Verdict != "still_differs" || !v2["2"].ClaimMismatch {
		t.Errorf("v2 remark 2 (quantity) is not fixed: %+v", v2["2"])
	}
	// the specialist's own status
	if w := post(handleExpertiseStatus, "/api/expertise/status", map[string]string{"recordID": "pack-1", "id": v1["2"].ID, "status": "disputed"}); w.Code != 200 {
		t.Fatalf("status: %d %s", w.Code, w.Body)
	}
	if got := byNum("/api/comparison?recordID=pack-1&namespaceID=x")["2"].Status; got != "disputed" {
		t.Errorf("status not kept: %s", got)
	}
	if w := post(handleExpertiseStatus, "/api/expertise/status", map[string]string{"recordID": "pack-1", "id": "ex-2", "status": "bogus"}); w.Code != 400 {
		t.Errorf("bogus status accepted: %d", w.Code)
	}
}
