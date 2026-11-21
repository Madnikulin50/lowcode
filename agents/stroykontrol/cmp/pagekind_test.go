package cmp

import (
	"strings"
	"testing"
)

func TestSheetFormat(t *testing.T) {
	for _, c := range []struct {
		w, h float64
		want string
	}{
		{595, 842, "A4 книжный"}, {842, 595, "A4 альбомный"}, {1191, 842, "A3 альбомный"},
		{1684, 2384, "A1 книжный"}, {2384, 1684, "A1 альбомный"},
		{1704, 3002, "1704×3002 пт книжный"}, {1786, 842, "A4×3 альбомный"}, {630 / 25.4 * 72, 297 / 25.4 * 72, "A4×3 альбомный"}, {500, 500, "500×500 пт книжный"},
	} {
		if got := SheetFormat(c.w, c.h); got != c.want {
			t.Errorf("SheetFormat(%v,%v) = %q, want %q", c.w, c.h, got, c.want)
		}
	}
	if SheetFormat(0, 10) != "" {
		t.Error("no size, no format")
	}
}

func scanned(d *Document, conf float64) *Document {
	for i := range d.Pages {
		d.Pages[i].OCR, d.Pages[i].OCRConf, d.Pages[i].Raster = true, conf, true
		for j := range d.Pages[i].Lines {
			d.Pages[i].Lines[j].OCR = true
		}
	}
	return d
}

func TestScanNotesSayWhereTheTextIsNotTrustworthy(t *testing.T) {
	d := fmtDoc("scan", SideTarget, fmtPage{1, "", body(1, "а")}, fmtPage{2, "", body(2, "б")}, fmtPage{3, "", body(3, "в")})
	scanned(d, 90)
	d.Pages[1].OCRConf = 70 // weak
	d.Pages[2].OCR, d.Pages[2].OCRReject = false, true
	plain := fmtDoc("plain", SideRef, fmtPage{1, "", body(1, "а")}, fmtPage{2, "", body(2, "б")})
	notes := strings.Join(scanNotes([]*Document{plain, d}), "\n")
	if !strings.Contains(notes, "стр. 2 распознаны со сниженной достоверностью") || !strings.Contains(notes, "стр. 3 — скан не распознан") || !strings.Contains(notes, "все страницы растровые") {
		t.Fatalf("notes:\n%s", notes)
	}
	if strings.Contains(notes, "plain") {
		t.Errorf("a plain vector PDF needs no note:\n%s", notes)
	}
}

func TestWeaklyRecognizedFindingIsLowAndSaysHowSure(t *testing.T) {
	mk := func(id, side string, depth string, conf float64) *Document {
		pages := five("а")
		pages[1].body = append([]string{"Глубина заложения фундамента " + depth + " м."}, pages[1].body...)
		return scanned(fmtDoc(id, side, pages...), conf)
	}
	for _, c := range []struct {
		conf float64
		sev  string
		text string
	}{{70, "low", "достоверность 70 %"}, {92, "medium", "достоверность 92 %"}} {
		res := Compare([]*Document{mk("pd", SideRef, "2,4", c.conf)}, []*Document{mk("rd", SideTarget, "2,6", c.conf)}, Options{NoFormat: true, NoCross: true})
		var found bool
		for _, f := range res.Findings {
			if strings.Contains(f.Description, "Глубина") && f.OCR {
				found = true
				if f.Severity != c.sev || !strings.Contains(f.Description, c.text) || !strings.HasSuffix(f.Evidence.Rule, "+ocr") {
					t.Errorf("conf %v: %s | %s | %s", c.conf, f.Severity, f.Description, f.Evidence.Rule)
				}
			}
		}
		if !found {
			t.Fatalf("conf %v: the changed number must be found:\n%+v", c.conf, res.Findings)
		}
	}
}
