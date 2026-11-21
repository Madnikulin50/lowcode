package cmp

import (
	"reflect"
	"testing"
)

func TestParseMaterials(t *testing.T) {
	ms := ParseMaterials("Внесены изменения в том 5.4.1 (ТКР5.4.1), листы 12, 14-15. Расчёт фундаментов представлен в Приложении Б тома КР5.2. Письмо № 12-34/5 от 01.02.2024 прилагается.")
	type got struct{ Kind, Code, ID, Doc string }
	var g []got
	for _, m := range ms {
		g = append(g, got{m.Kind, m.Code, m.ID, m.Doc})
	}
	want := []got{
		{MatVolume, "ТКР5.4.1", "", ""},
		{MatVolume, "КР5.2", "", ""},
		{MatAppendix, "КР5.2", "Б", ""},
		{MatDocument, "КР5.2", "", "расчёт"},
		{MatDocument, "", "", "письмо"},
	}
	if !reflect.DeepEqual(g, want) {
		t.Fatalf("materials:\n got %+v\nwant %+v", g, want)
	}
	if !reflect.DeepEqual(ms[0].Sheets, []int{12, 14, 15}) || ms[0].Tom != "5.4.1" {
		t.Errorf("sheets/tom of the first volume: %+v", ms[0])
	}
}

func TestParseMaterialsSkips(t *testing.T) {
	for _, s := range []string{
		"Расчёт не требуется, так как нагрузки не изменились.",
		"Принято. Откорректировано согласно СП 20.13330.2016, приложение Б.",
		"Бетон класса В25, арматура А500С, W6, F150.",
		"",
	} {
		if ms := ParseMaterials(s); len(ms) != 0 {
			t.Errorf("%q: nothing declared, got %+v", s, ms)
		}
	}
}

func TestCheckMaterials(t *testing.T) {
	docs := []MatDoc{
		{ID: "d1", Name: "ТКР5.4.1", Codes: "100720-1/П-3-ТКР5-4-1", Pages: 30, Sheets: map[int]int{3: 1, 14: 12, 16: 14}},
		{ID: "d2", Name: "КР5.2", Codes: "100720-1/П-3-КР5-2", Pages: 40,
			Headings: []DocHeading{{Kind: MatAppendix, ID: "А", Text: "Приложение А", Page: 30}, {Kind: MatDocument, Doc: "расчёт", Text: "Расчёт фундаментов на продавливание", Page: 12}}},
		{ID: "f1", Name: "Письмо ГИП 12-34.pdf", Supporting: true},
	}
	ms := CheckMaterials(ParseMaterials("Изменены листы 12, 14-15 тома ТКР5.4.1. Расчёт фундаментов см. в Приложении Б тома КР5.2. Письмо № 12-34 прилагается. Справка в томе ИОС4.1."), docs)
	states := map[string]string{}
	for _, m := range ms {
		states[m.Kind+":"+m.Label] = m.State
	}
	want := map[string]string{
		"volume:ТКР5.4.1":       MatMissing, // sheet 15 is not there
		"volume:КР5.2":          MatFound,
		"appendix:Приложение Б": MatMissing, // the volume has only Приложение А
		"document:Расчёт фундаментов см. в Приложении Б тома КР5.2.": MatFound,
		"document:Письмо № 12-34 прилагается.":                       MatFound, // a supporting file
		"volume:ИОС4.1": MatNoDocument,
		"document:Справка в томе ИОС4.1.": MatNoDocument,
	}
	if !reflect.DeepEqual(states, want) {
		t.Fatalf("states:\n got %v\nwant %v", states, want)
	}
	if MaterialsSummary(ms) != MaterialsMissing {
		t.Errorf("summary: %s", MaterialsSummary(ms))
	}
	if ms[0].Note == "" || ms[0].At != 14 {
		t.Errorf("first volume: %+v", ms[0])
	}
}

func TestMaterialsSummaryAndOutcome(t *testing.T) {
	if MaterialsSummary(nil) != "" {
		t.Error("nothing declared → empty")
	}
	if s := MaterialsSummary([]Material{{State: MatFound}, {State: MatNoDocument}}); s != MaterialsUnloaded {
		t.Errorf("unloaded: %s", s)
	}
	if s := MaterialsSummary([]Material{{State: MatFound}}); s != MaterialsOK {
		t.Errorf("ok: %s", s)
	}
	if WithMaterials(OutcomeFixed, MaterialsMissing) != OutcomeCheck || WithMaterials(OutcomeNotFixed, MaterialsMissing) != OutcomeNotFixed || WithMaterials(OutcomeFixed, MaterialsOK) != OutcomeFixed {
		t.Error("outcome with materials")
	}
}

func TestExtractHeadings(t *testing.T) {
	d := &Document{Pages: []Page{
		{Cover: true, Lines: []Line{{Text: "Приложение А"}}},
		{Lines: []Line{{Text: "Приложение Б . . . . . 12"}, {Text: "Расчёт плиты перекрытия"}, {Text: "Приложение В (обязательное)"}, {Text: "Изм. Кол.уч. Лист", Stamp: true}}},
	}}
	h := ExtractHeadings(d)
	want := []DocHeading{
		{Kind: MatDocument, Doc: "расчёт", Text: "Расчёт плиты перекрытия", Page: 2},
		{Kind: MatAppendix, ID: "В", Text: "Приложение В (обязательное)", Page: 2},
	}
	if !reflect.DeepEqual(h, want) {
		t.Fatalf("headings:\n got %+v\nwant %+v", h, want)
	}
	if got := ParseHeadings(HeadingsJSON(h)); !reflect.DeepEqual(got, h) {
		t.Errorf("round trip: %+v", got)
	}
}

func TestParseMaterialsVolumeMark(t *testing.T) {
	ms := ParseMaterials("Расчёты добавлены в том АР")
	if len(ms) != 2 || ms[0].Kind != MatVolume || ms[0].Code != "АР" || ms[1].Kind != MatDocument || ms[1].Code != "АР" {
		t.Fatalf("got %+v", ms)
	}
}
