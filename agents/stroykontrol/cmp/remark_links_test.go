package cmp

import (
	"encoding/json"
	"testing"
)

func TestSheetMapRoundTrip(t *testing.T) {
	d := fmtDoc("d", SideRef, fmtPage{1, "", body(1, "а")}, fmtPage{2, "", body(2, "б")}, fmtPage{0, "", body(3, "в")}, fmtPage{7, "", body(4, "г")})
	m := SheetMap(d.Pages)
	if m[1] != 1 || m[2] != 2 || m[4] != 7 || len(m) != 3 {
		t.Fatalf("pages without a readable number are left out: %v", m)
	}
	s := FormatSheetMap(m)
	if s != "1:1,2:2,4:7" {
		t.Fatalf("format: %q", s)
	}
	back := ParseSheetMap(s)
	if len(back) != 3 || back[4] != 7 || len(ParseSheetMap("")) != 0 || len(ParseSheetMap("мусор,1:x")) != 0 {
		t.Errorf("parse: %v", back)
	}
}

func TestLocateRemarkSheetPageAndFindings(t *testing.T) {
	docs := []LocDoc{
		{ID: "pd", Name: "ТКР5.1", Side: SideRef, Codes: "100720-1/П-3-ТКР5-1", Pages: 30, Sheets: map[int]int{1: 1, 2: 2, 8: 7}},
		{ID: "rd", Name: "ТКР5.1 РД", Side: SideTarget, Codes: "100720-1/Р-3-ТКР5-1", Pages: 5}, // sheets not read
		{ID: "other", Name: "КЖ1", Side: SideRef, Codes: "КЖ1", Pages: 5},
	}
	refs := []VolumeRef{{Code: "ТКР5.1", Sheet: 7}, {Code: "ТКР5.1", Page: 3}, {Code: "ТКР5.1", Page: 99}}
	finds := []FoundAt{{ID: "f1", Comparison: "c1", Text: "Глубина 2,4 → 2,6", RefDoc: "pd", RefPage: 8, RefBox: "0.1,0.2,0.5,0.3", TargetDoc: "rd", TargetPage: 2, TargetBox: "0.1,0.2,0.5,0.3"}}
	got := LocateRemark(refs, docs, finds)
	by := map[string]RemarkLocation{}
	for _, l := range got {
		by[l.Kind+"|"+l.DocID+"|"+itoa(l.Page)+"|"+itoa(l.Sheet)] = l
	}
	// the sheet in a document whose sheets were read: the page of that sheet
	if l, ok := by["place|pd|008|007"]; !ok || l.Side != SideRef || l.Note != "" {
		t.Errorf("sheet 7 is on page 8 of the ПД file: %+v", got)
	}
	// the sheet in a document whose sheets were not read: the start, and it says why
	if l, ok := by["place|rd|000|007"]; !ok || l.Note == "" {
		t.Errorf("sheet 7 of the РД file cannot be located: %+v", got)
	}
	// a page is a page; a page past the end of the file is not trusted
	if _, ok := by["place|pd|003|000"]; !ok {
		t.Errorf("page 3: %+v", got)
	}
	if l, ok := by["place|pd|000|000"]; !ok || l.Note == "" {
		t.Errorf("page 99 of 30: %+v", got)
	}
	for k := range by {
		if k[:11] == "place|other" {
			t.Errorf("a volume the remark does not name must not be linked: %s", k)
		}
	}
	f := by["finding||000|000"]
	if f.FindingID != "f1" || f.RefPage != 8 || f.RefBox == "" || f.TargetDoc != "rd" || f.Comparison != "c1" {
		t.Errorf("finding location: %+v", f)
	}
	// JSON round trip (the field links_json)
	var back []RemarkLocation
	if err := json.Unmarshal([]byte(LocationsJSON(got)), &back); err != nil || len(back) != len(got) {
		t.Errorf("json: %v %d/%d", err, len(back), len(got))
	}
	if LocationsJSON(nil) != "" {
		t.Error("no locations, no value")
	}
	// the same place named twice is one place
	if again := LocateRemark([]VolumeRef{{Code: "ТКР5.1", Page: 3}, {Code: "ТКР5.1", Page: 3}}, docs, nil); len(again) != 2 {
		t.Errorf("the ПД and РД files, each once: %+v", again)
	}
}
