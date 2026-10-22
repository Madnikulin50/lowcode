package cmp

import (
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

func TestDetectChangeFromTheCoverAndTheName(t *testing.T) {
	d := fmtDoc("d", SideRef, fmtPage{1, "", append([]string{"03102022/РД-3-ПИК1-НК2 (Изм. 2)", "Книга 3. Часть 2"}, body(1, "а")...)},
		fmtPage{2, "", body(2, "б")}, fmtPage{3, "", body(3, "в")}, fmtPage{4, "", body(4, "г")}, fmtPage{5, "", body(5, "д")})
	d.Name = "Комплект РД НК2.pdf" // the name says nothing: the cover does
	v := volOf(t, d)
	if v.Change != 2 || v.ChangeSource != ChangeFromCover || v.Book != "3" || v.Part != "2" {
		t.Fatalf("cover: change=%d(%s) book=%q part=%q", v.Change, v.ChangeSource, v.Book, v.Part)
	}
	// no cover mark: a single-volume file takes the number from its name
	d2 := fmtDoc("d2", SideTarget, five("а")...)
	d2.Name = "03102022_РД-ИССО10.1-ОП-КЖ6-1 изм. 3.pdf"
	if v := volOf(t, d2); v.Change != 3 || v.ChangeSource != ChangeFromName {
		t.Fatalf("name: change=%d(%s)", v.Change, v.ChangeSource)
	}
	// the title-block table header «Изм. Кол.уч. Лист…» and the file name of a multi-volume file are not marks
	d3 := fmtDoc("d3", SideTarget, five("б")...)
	d3.Name = "КЖ1.pdf"
	d3.Pages[0].Lines = append(d3.Pages[0].Lines, mkLine("Изм. Лист Содержание изменения Код Примечание", 0.1, 0.5, 0.9, 0.52))
	if v := volOf(t, d3); v.Change != 0 {
		t.Errorf("an original edition has no change number: %d", v.Change)
	}
}

func TestSupersedeKeepsTheNewestEditionOfAVolume(t *testing.T) {
	mk := func(code string, change int) *Volume {
		return &Volume{Code: code, RawCode: code, Title: "Том " + code, Change: change}
	}
	keep, dropped := supersede([]*Volume{mk("КЖ6-1", 0), mk("КЖ6-1", 2), mk("НК2", 1), mk("НК2", 1), mk("ЭС21", 0)})
	if len(dropped) != 1 || dropped[0].Change != 0 || dropped[0].Code != "КЖ6-1" || len(keep) != 4 {
		t.Fatalf("keep=%d dropped=%+v", len(keep), dropped)
	}
	if n := supersededNote(SideTarget, dropped); !strings.Contains(n, "исходная редакция") || !strings.Contains(n, "проверяемой стороны") {
		t.Errorf("note: %q", n)
	}
	if supersededNote(SideRef, nil) != "" {
		t.Error("nothing dropped, nothing said")
	}
}

func TestEditionNotesOnlyForVersionsOfOneStage(t *testing.T) {
	ref := &Volume{Code: "КЖ1", RawCode: "КЖ1", Title: "КЖ", Change: 3}
	older := &Volume{Code: "КЖ1", RawCode: "КЖ1", Change: 2}
	same := &Volume{Code: "КЖ1", RawCode: "КЖ1", Change: 3}
	newer := &Volume{Code: "КЖ1", RawCode: "КЖ1", Change: 4}
	if got := editionNotes([]*Pair{{Ref: ref, Targets: []*Volume{older}}}, ModeVersions); len(got) != 1 || !strings.Contains(got[0], "обратном порядке") {
		t.Errorf("reversed: %v", got)
	}
	if got := editionNotes([]*Pair{{Ref: ref, Targets: []*Volume{same}}}, ModeVersions); len(got) != 1 || !strings.Contains(got[0], "один номер изменения") {
		t.Errorf("same: %v", got)
	}
	if got := editionNotes([]*Pair{{Ref: ref, Targets: []*Volume{newer}}}, ModeVersions); len(got) != 0 {
		t.Errorf("a newer edition is the normal case: %v", got)
	}
	if got := editionNotes([]*Pair{{Ref: ref, Targets: []*Volume{older}}}, ModeStages); len(got) != 0 {
		t.Errorf("ПД and РД numbers are not ordered against each other: %v", got)
	}
	if got := editionNotes([]*Pair{{Ref: &Volume{Change: 0}, Targets: []*Volume{older}}}, ModeVersions); len(got) != 0 {
		t.Errorf("an original edition carries no number to compare: %v", got)
	}
}

func TestEditionValuesKeepWhatWasTypedForTheSameFile(t *testing.T) {
	vs := []VolumeInfo{{Change: 2, Book: "3"}}
	// a new file: everything the agent read is written
	v := EditionValues(nil, "h1", vs)
	if v["change_no"][0] != "2" || v["book"][0] != "3" || v["file_hash"][0] != "h1" {
		t.Fatalf("%v", v)
	}
	// the same file, the specialist typed a change number: it stays; an empty field is filled
	rec := &idcheck.Record{Values: map[string][]string{"file_hash": {"h1"}, "change_no": {"5"}}}
	v = EditionValues(rec, "h1", vs)
	if _, ok := v["change_no"]; ok || v["book"][0] != "3" {
		t.Errorf("typed values survive: %v", v)
	}
	// another file under the same record: what was typed belonged to the old one
	v = EditionValues(rec, "h2", []VolumeInfo{{}})
	if v["change_no"][0] != "" || v["file_hash"][0] != "h2" {
		t.Errorf("a new file replaces the old values: %v", v)
	}
	// volumes of different editions in one file: no single number
	if v := EditionValues(nil, "h", []VolumeInfo{{Change: 1}, {Change: 2}}); v["change_no"][0] != "" {
		t.Errorf("mixed: %v", v)
	}
}

func TestCompareSkipsAnOlderEditionOfTheSameVolume(t *testing.T) {
	lines := func(n int) []string { return body(n, "а") }
	pd := fmtDoc("pd", SideRef, five("а")...)
	old := fmtDoc("old", SideTarget, fmtPage{1, "", lines(1)}, fmtPage{2, "", lines(2)}, fmtPage{3, "", lines(3)}, fmtPage{4, "", lines(4)}, fmtPage{5, "", lines(5)})
	old.Name = "КЖ1.pdf"
	cur := fmtDoc("cur", SideTarget, fmtPage{1, "", lines(1)}, fmtPage{2, "", lines(2)}, fmtPage{3, "", lines(3)}, fmtPage{4, "", lines(4)}, fmtPage{5, "", lines(5)})
	cur.Name = "КЖ1 изм. 2.pdf"
	res := Compare([]*Document{pd}, []*Document{old, cur}, Options{NoFormat: true, NoCross: true})
	joined := strings.Join(res.Notes, "\n")
	if !strings.Contains(joined, "старые редакции") {
		t.Fatalf("the older edition must be left out and said so:\n%s", joined)
	}
	if len(res.Pairs) != 1 || len(res.Pairs[0].Targets) != 1 || res.Pairs[0].Targets[0].DocID != "cur" || res.Pairs[0].Targets[0].Change != 2 {
		t.Errorf("only the newest edition is paired: %+v", res.Pairs)
	}
}
