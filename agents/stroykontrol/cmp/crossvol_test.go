package cmp

import (
	"strings"
	"testing"
)

func twoVolumeDoc(side string, a, b []string) *Document {
	return fmtDoc("d", side,
		fmtPage{1, "2024-15-КЖ1", a}, fmtPage{2, "2024-15-КЖ1", body(2, "а")},
		fmtPage{3, "2024-15-КЖ2", b}, fmtPage{4, "2024-15-КЖ2", body(4, "б")},
	)
}

func crossOf(t *testing.T, d *Document, side string) ([]Finding, []ParamRow) {
	t.Helper()
	vs := DetectVolumes(d)
	if len(vs) != 2 {
		t.Fatalf("want two volumes, got %d", len(vs))
	}
	return crossVolume([]*Document{d}, vs, side)
}

func TestCrossVolumeSameElementDifferentGrade(t *testing.T) {
	d := twoVolumeDoc(SideRef,
		[]string{"Общие данные", "Фундаменты выполнить из бетона класса В25, W6, F150"},
		[]string{"Пояснительная записка", "Фундаменты выполнить из бетона класса В30, W6, F150"},
	)
	fs, rows := crossOf(t, d, SideRef)
	if len(fs) != 1 {
		t.Fatalf("want one finding (the grade; W and F agree), got %d: %+v", len(fs), fs)
	}
	f := fs[0]
	if f.Kind != KindCrossVolume || f.Severity != "low" || f.RefDoc != "d" || f.PageRef != 3 || f.TargetDoc != "" {
		t.Fatalf("%+v", f)
	}
	for _, want := range []string{"Класс бетона", "«фундамент»", "B25", "B30", "Проверьте"} {
		if !strings.Contains(f.Description, want) {
			t.Fatalf("description lacks %q: %s", want, f.Description)
		}
	}
	if len(rows) < 6 { // 3 kinds × 2 volumes
		t.Fatalf("the table of parameters: %+v", rows)
	}
}

func TestCrossVolumeDifferentElementsAreNotAConflict(t *testing.T) {
	d := twoVolumeDoc(SideTarget,
		[]string{"Общие данные", "Фундаменты выполнить из бетона класса В25"},
		[]string{"Пояснительная записка", "Стены подвала выполнить из бетона класса В30"},
	)
	if fs, _ := crossOf(t, d, SideTarget); len(fs) != 0 {
		t.Fatalf("a foundation and a wall may have different grades: %+v", fs)
	}
}

func TestCrossVolumeNeedsAnElementAndAgreementIsFine(t *testing.T) {
	d := twoVolumeDoc(SideRef,
		[]string{"Общие данные", "Бетон класса В25 для всех конструкций"},
		[]string{"Пояснительная записка", "Бетон класса В30 для прочих конструкций"},
	)
	if fs, _ := crossOf(t, d, SideRef); len(fs) != 0 {
		t.Fatalf("no element named → no finding: %+v", fs)
	}
	d = twoVolumeDoc(SideRef,
		[]string{"Общие данные", "Ростверки выполнить из бетона класса В30"},
		[]string{"Пояснительная записка", "Ростверки выполнить из бетона класса В30"},
	)
	if fs, _ := crossOf(t, d, SideRef); len(fs) != 0 {
		t.Fatalf("equal values: %+v", fs)
	}
}

func TestCrossVolumeElementFromThePreviousLine(t *testing.T) {
	d := twoVolumeDoc(SideRef,
		[]string{"Свайные фундаменты", "выполнить из бетона класса В30"},
		[]string{"Свайные фундаменты", "выполнить из бетона класса В25"},
	)
	fs, _ := crossOf(t, d, SideRef)
	if len(fs) == 0 {
		t.Fatal("the element named one line above must be taken")
	}
}

func TestCrossVolumeRunsInsideCompareAndCanBeSwitchedOff(t *testing.T) {
	ref := twoVolumeDoc(SideRef,
		[]string{"Общие данные", "Фундаменты выполнить из бетона класса В25"},
		[]string{"Пояснительная записка", "Фундаменты выполнить из бетона класса В30"},
	)
	tgt := textDoc("t", SideTarget, []string{"Общие данные", "Фундаменты выполнить из бетона класса В25"})
	res := Compare([]*Document{ref}, []*Document{tgt}, Options{NoFormat: true})
	n := 0
	for _, f := range res.Findings {
		if f.Kind == KindCrossVolume {
			n++
		}
	}
	if n != 1 || len(res.Params) == 0 {
		t.Fatalf("cross findings %d, params %d", n, len(res.Params))
	}
	res = Compare([]*Document{ref}, []*Document{tgt}, Options{NoFormat: true, NoCross: true})
	for _, f := range res.Findings {
		if f.Kind == KindCrossVolume {
			t.Fatal("switched off")
		}
	}
}
