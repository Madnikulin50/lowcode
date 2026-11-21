package cmp

import (
	"strings"
	"testing"
)

func TestMarksAreReadAndNormsAreNot(t *testing.T) {
	for in, want := range map[string]string{
		"Ф-2 Ленточный фундамент":                  "Ф-2",
		"Плиты Пл-1, Пл-2а и БС-3":                 "ПЛ-1,ПЛ-2А,БС-3",
		"Блок Ф-12 (см. лист 4)":                   "Ф-12",
		"марка Ф-3 на оси А":                       "Ф-3",
		"F-3 латинской раскладкой":                 "", // Latin F alone is a class of frost resistance
		"Бетон В-25, W6, F150, арматура А-500":     "",
		"Узел КЖ1-3 тома":                          "",
		"Нагрузки по таблице Т-4 ГОСТ 27751":       "",
		"требования СП 63.13330, марка Ф-5":        "", // a line that speaks of a norm is not read
		"Акт по форме КС-2, раствор М-100, МSК-64": "",
		"Позиция 12-3 в таблице":                   "",
		"Ф-2.1":                                    "", // a dotted number is a clause, not a mark
	} {
		if got := strings.Join(Marks(in), ","); got != want {
			t.Errorf("Marks(%q) = %q, want %q", in, got, want)
		}
	}
}

func marksDoc(id, side string, pages ...[]string) *Document {
	var ps []fmtPage
	for i, b := range pages {
		ps = append(ps, fmtPage{i + 1, "", b})
	}
	for len(ps) < 4 { // a volume needs a few sheets with a title block to be detected
		ps = append(ps, fmtPage{len(ps) + 1, "", body(len(ps), "ю")})
	}
	return fmtDoc(id, side, ps...)
}

func TestMarksMissingInTheTargetAndAddedByIt(t *testing.T) {
	pd := marksDoc("pd", SideRef, []string{"Ф-1 Фундамент под колонну", "Ф-2 Ленточный фундамент", "Ф-3 Столбчатый фундамент", "Пл-1 Плита перекрытия"})
	rd := marksDoc("rd", SideTarget, []string{"Ф-1 Фундамент под колонну", "Ф-2 Ленточный фундамент", "Пл-1 Плита перекрытия", "Пл-2 Плита покрытия"})
	docs := map[string]*Document{"pd": pd, "rd": rd}
	pair := &Pair{Ref: volOf(t, pd), Targets: []*Volume{volOf(t, rd)}}
	fs := elementFindings(pair, docs)
	if len(fs) != 2 {
		t.Fatalf("a missing and an added mark:\n%s", formatMessages(fs))
	}
	byRule := map[string]Finding{}
	for _, f := range fs {
		byRule[f.Evidence.Rule] = f
		if f.Kind != KindElement {
			t.Errorf("kind: %s", f.Kind)
		}
	}
	miss, add := byRule["marks-missing"], byRule["marks-added"]
	if miss.Severity != "medium" || !strings.Contains(miss.Description, "Ф-3") || strings.Contains(miss.Description, "Ф-2") || miss.RefDoc != "pd" || miss.PageRef == 0 || miss.BoxRef == nil {
		t.Errorf("missing: %+v", miss)
	}
	if add.Severity != "low" || !strings.Contains(add.Description, "ПЛ-2") || add.TargetDoc != "rd" || add.BoxTarget == nil {
		t.Errorf("added: %+v", add)
	}
	// a fragment (the РД covers part of the ПД) is not blamed for the marks it does not have
	pair.Fragment = true
	if fs := elementFindings(pair, docs); len(fs) != 1 || fs[0].Evidence.Rule != "marks-added" {
		t.Errorf("fragment: %s", formatMessages(fs))
	}
	// too few marks say nothing
	few := marksDoc("few", SideTarget, []string{"Ф-1 Фундамент"})
	if fs := elementFindings(&Pair{Ref: volOf(t, pd), Targets: []*Volume{volOf(t, few)}}, map[string]*Document{"pd": pd, "few": few}); len(fs) != 0 {
		t.Errorf("one mark is not a comparison: %s", formatMessages(fs))
	}
}

func TestSpecificationAgainstTheDrawingsOfOneSide(t *testing.T) {
	d := marksDoc("d", SideRef,
		[]string{"Ф-1 Фундамент", "Ф-2 Фундамент", "Пл-1 Плита", "Б-7 Балка на чертеже"},
		[]string{"Спецификация элементов"})
	v := volOf(t, d)
	v.Items = []Item{
		{Table: "spec", Name: "Ф-1 Фундамент монолитный", Page: 2}, {Table: "spec", Name: "Ф-2 Фундамент монолитный", Page: 2},
		{Table: "spec", Name: "Пл-1 Плита перекрытия", Page: 2}, {Table: "spec", Name: "Пл-9 Плита покрытия", Page: 2},
	}
	fs := specDrawingFindings([]*Document{d}, []*Volume{v}, SideRef)
	msg := formatMessages(fs)
	if len(fs) != 2 || !strings.Contains(msg, "ПЛ-9") || !strings.Contains(msg, "Б-7") {
		t.Fatalf("Пл-9 is only in the specification, Б-7 only on a sheet:\n%s", msg)
	}
	for _, f := range fs {
		if f.Severity != "low" || f.Kind != KindCrossVolume || f.RefDoc != "d" {
			t.Errorf("%+v", f)
		}
	}
	// a volume without a specification is not judged
	v.Items = nil
	if fs := specDrawingFindings([]*Document{d}, []*Volume{v}, SideRef); len(fs) != 0 {
		t.Errorf("no specification, nothing to compare: %s", formatMessages(fs))
	}
}

func TestCompareReportsMarksUnlessSwitchedOff(t *testing.T) {
	pd := marksDoc("pd", SideRef, []string{"Ф-1 Фундамент", "Ф-2 Фундамент", "Ф-3 Фундамент", "Пл-1 Плита"})
	rd := marksDoc("rd", SideTarget, []string{"Ф-1 Фундамент", "Ф-2 Фундамент", "Пл-1 Плита", "Пл-2 Плита"})
	count := func(o Options) int {
		n := 0
		for _, f := range Compare([]*Document{pd}, []*Document{rd}, o).Findings {
			if f.Kind == KindElement {
				n++
			}
		}
		return n
	}
	if count(Options{NoFormat: true, NoCross: true}) != 2 || count(Options{NoFormat: true, NoCross: true, NoElements: true}) != 0 {
		t.Error("the marks are compared unless NoElements")
	}
}
