package cmp

import (
	"testing"
)

func TestChangeRegistration(t *testing.T) {
	ref := loadDoc(t, "r", SideRef, "changes-1/pd-kzh2.pdf")
	tgt := loadDoc(t, "t", SideTarget, "changes-1/rd-kzh2 изм. 1.pdf")
	res := Compare([]*Document{ref}, []*Document{tgt}, Options{Mode: ModeVersions, NoElements: true})

	got := map[int]SheetChange{}
	for _, sc := range res.SheetChanges {
		got[sc.Sheet] = sc
	}
	want := map[int][3]string{ // fact, declared, ok
		1: {SheetSame, "", "ok"},
		2: {SheetChanged, SheetChanged, "ok"},
		3: {SheetChanged, "", "issue"},
		4: {SheetAnnulled, SheetAnnulled, "ok"},
		5: {SheetSame, SheetReplaced, "issue"},
		6: {SheetNew, SheetNew, "ok"},
		7: {SheetNew, SheetNew, "ok"},
	}
	if len(got) != len(want) {
		t.Fatalf("sheets: %+v", res.SheetChanges)
	}
	for no, w := range want {
		sc := got[no]
		ok := "issue"
		if sc.OK {
			ok = "ok"
		}
		if sc.Fact != w[0] || sc.Declared != w[1] || ok != w[2] || sc.Change != 1 {
			t.Errorf("sheet %d: %+v, want %v", no, sc, w)
		}
	}
	if got[2].Stamp != SheetChanged || got[5].Stamp != SheetReplaced || got[6].Stamp != SheetNew {
		t.Errorf("stamp marks: %+v %+v %+v", got[2], got[5], got[6])
	}

	rules := map[string]string{}
	for _, f := range res.Findings {
		if f.Kind == KindChangeReg {
			rules[f.Evidence.Rule] = f.Severity
		}
	}
	if rules["change-undeclared"] != "high" || rules["change-no-diff"] != "low" || len(rules) != 2 {
		t.Errorf("findings: %v", rules)
	}
	if st := res.State(); len(st.SheetChanges) != 7 {
		t.Errorf("state: %d sheets", len(st.SheetChanges))
	}
}

func TestChangeRegistrationStagesSkipped(t *testing.T) {
	ref := loadDoc(t, "r", SideRef, "changes-1/pd-kzh2.pdf")
	tgt := loadDoc(t, "t", SideTarget, "changes-1/rd-kzh2 изм. 1.pdf")
	res := Compare([]*Document{ref}, []*Document{tgt}, Options{NoElements: true})
	if len(res.SheetChanges) != 0 {
		t.Errorf("ПД → РД is not checked for the change registration: %+v", res.SheetChanges)
	}
}

func TestStampChangeRows(t *testing.T) {
	p := Page{Lines: []Line{
		{Text: "1 - Зам. 15-24 01.24", Key: NormKey("1 - Зам. 15-24 01.24"), Stamp: true},
		{Text: "2 3 Нов. 27-24", Key: NormKey("2 3 Нов. 27-24"), Stamp: true},
		{Text: "1 Изменение толщины плиты", Key: NormKey("1 Изменение толщины плиты")},
	}}
	sc := stampChanges(p)
	if sc[1] != SheetReplaced || sc[2] != SheetNew || len(sc) != 2 {
		t.Errorf("got %v", sc)
	}
	if !changeRowRe.MatchString(NormKey("1 - Аннул. 15-24")) || changeRowRe.MatchString(NormKey("1 Изменение толщины плиты")) {
		t.Error("change row pattern")
	}
}

func TestPermittedSheets(t *testing.T) {
	mk := func(s string) Line { return Line{Text: s, Key: NormKey(s)} }
	d := &Document{Pages: []Page{
		{Lines: []Line{mk("Пояснительная записка")}},
		{Lines: []Line{mk("Разрешение на внесение изменений № 15-24"), mk("Изменить листы 2, 5–6 и лист 9"), mk("Изм. Кол.уч. Лист № док. Подп. Дата")}},
	}}
	d.Pages[1].Lines[2].Stamp = true
	got, page := permittedSheets(d, 1, 2)
	if page != 2 || len(got) != 4 || !got[2] || !got[5] || !got[6] || !got[9] {
		t.Errorf("permission: %v on %d", got, page)
	}
	if got, _ := permittedSheets(&Document{Pages: d.Pages[:1]}, 1, 1); got != nil {
		t.Errorf("no permission: %v", got)
	}
}
