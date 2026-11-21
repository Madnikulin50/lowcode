package cmp

import (
	"strings"
	"testing"
)

func fragPage(n int, lines ...string) Page {
	p := Page{Number: n, HasText: true}
	for i, t := range lines {
		y := 0.1 + 0.05*float64(i)
		p.Lines = append(p.Lines, Line{Page: n, Text: t, Key: NormKey(t), Box: Box{0.1, y, 0.6, y + 0.03}})
	}
	return p
}

func TestLocateFragments(t *testing.T) {
	pd := FragDoc{LocDoc{ID: "pd", Name: "КЖ", Side: SideRef, Sheets: map[int]int{2: 5}}, []Page{
		fragPage(1, "Общие указания по проекту", "Нагрузки приняты по СП 20.13330"),
		fragPage(2, "1.3 Глубина заложения фундамента: 2,4 м.", "Отметка подошвы минус 2,400"),
	}}
	rd := FragDoc{LocDoc{ID: "rd", Name: "КЖ изм. 1", Side: SideTarget}, []Page{
		fragPage(1, "1.3 Глубина заложения фундамента: 2,6 м.", "Отметка подошвы минус 2,600"),
	}}
	got := LocateFragments("Глубина заложения фундамента 2,6 м не соответствует стадии П (2,4 м). Обосновать.", nil, []FragDoc{pd, rd})
	if len(got) != 2 {
		t.Fatalf("want a place on each side, got %+v", got)
	}
	if got[0].Side != SideRef || got[0].Page != 2 || got[0].Sheet != 5 || !strings.Contains(got[0].Text, "2,4") || got[0].Box == "" || got[0].Kind != LocFragment {
		t.Errorf("source side: %+v", got[0])
	}
	if got[1].Side != SideTarget || got[1].Page != 1 || !strings.Contains(got[1].Text, "2,6") {
		t.Errorf("corrected side: %+v", got[1])
	}
	// a quotation wins over the words and gives the lines it takes
	got = LocateFragments("Текст «отметка подошвы минус 2,600» не согласован с разрезом", nil, []FragDoc{rd})
	if len(got) != 1 || got[0].Note != "по цитате из замечания" || !strings.HasPrefix(got[0].Text, "Отметка подошвы") {
		t.Errorf("quotation: %+v", got)
	}
	// nothing in common — no place
	if got := LocateFragments("Не представлены расчёты теплопередачи ограждающих конструкций.", nil, []FragDoc{pd, rd}); len(got) != 0 {
		t.Errorf("unrelated remark found %+v", got)
	}
}

func TestLocateFragmentsNamedPage(t *testing.T) {
	// the same words on two pages: the page the remark refers to wins
	d := FragDoc{LocDoc{ID: "d", Name: "КЖ", Side: SideTarget, Codes: "2024-15-КЖ"}, []Page{
		fragPage(1, "Арматура фундаментной плиты класса А500С"),
		fragPage(2, "Арматура фундаментной плиты класса А500С"),
	}}
	refs := ParseVolumeRefs("2024-15-КЖ, стр. 2")
	got := LocateFragments("Класс арматуры фундаментной плиты А500С не обоснован", refs, []FragDoc{d})
	if len(got) != 1 || got[0].Page != 2 {
		t.Errorf("want page 2, got %+v (refs %+v)", got, refs)
	}
}

func TestQuoteBox(t *testing.T) {
	p := Passage{Text: "первая строка фрагмента\nвторая строка с цитатой здесь\nтретья", boxes: []Box{{0, 0.1, 1, 0.2}, {0, 0.3, 1, 0.4}, {0, 0.5, 1, 0.6}}}
	if b := p.quoteBox(NormKey("строка с цитатой")); b != "[0.0000,0.3000,1.0000,0.4000]" {
		t.Errorf("box %s", b)
	}
	if b := p.quoteBox(NormKey("чего нет")); b != "[0.0000,0.1000,1.0000,0.6000]" {
		t.Errorf("fallback box %s", b)
	}
}
