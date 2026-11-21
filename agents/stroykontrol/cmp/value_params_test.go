package cmp

import (
	"reflect"
	"strings"
	"testing"
)

func TestValueHits(t *testing.T) {
	cases := map[string]map[string][]string{
		"Толщина фундаментной плиты 0,6 м.":                  {"Толщина": {"600 мм"}},
		"Плита толщиной 600 мм, защитный слой 40":            {"Толщина": {"600 мм"}, "Защитный слой": {"40 мм"}},
		"Отметка подошвы фундамента -2,400":                  {"Отметка": {"-2,400"}},
		"За отметку 0,000 принята абсолютная отметка 152,30": {"Абсолютная отметка нуля": {"152,3"}},
		"Сейсмичность площадки строительства 7 баллов":       {"Сейсмичность": {"7 баллов"}},
		"Снеговая нагрузка 180 кгс/м2":                       {"Снеговая нагрузка": {"1,765 кПа"}},
		"Предел огнестойкости стен REI 150, колонн R 90":     {"Предел огнестойкости": {"REI 150", "R 90"}},
		"Арматура Ø12 А500С шаг 200":                         {"Диаметр": {"12 мм"}, "Шаг": {"200 мм"}},
		"Уровень ответственности здания — нормальный":        {"Уровень ответственности": {"нормальный"}},
		"Глубина заложения фундамента: 2,4 м.":               {"Глубина заложения": {"2,4 м"}},
		"Высота этажа 3 300":                                 {}, // no unit: not read
		"Уклон кровли 2 %":                                   {"Уклон": {"2 %"}},
	}
	for line, want := range cases {
		if got := valueHits(line); !reflect.DeepEqual(got, want) && !(len(got) == 0 && len(want) == 0) {
			t.Errorf("%q: %v, want %v", line, got, want)
		}
	}
}

// docWithDrawing is a volume with a text page and a drawing sheet.
func docWithDrawing(side string, text, drawing []string) (*Document, *Volume) {
	d := &Document{ID: side + "d", Side: side}
	p1, p2 := fragPage(1, text...), fragPage(2, drawing...)
	p2.Drawing = true
	d.Pages = []Page{p1, p2}
	v := &Volume{DocID: d.ID, Side: side, Code: "2024-15-КЖ", RawCode: "2024-15-КЖ", FirstPage: 1, LastPage: 2, Lines: p1.Lines, Drawings: []int{2}}
	return d, v
}

func TestDrawingTextFindings(t *testing.T) {
	d, v := docWithDrawing(SideTarget,
		[]string{"Плита фундаментная толщиной 600 мм из бетона В30", "Защитный слой бетона плиты 40 мм", "Сейсмичность площадки 6 баллов"},
		[]string{"Плита Пм1, толщина 500 мм, бетон В30", "Защитный слой плиты 40", "Стена подвала толщиной 300 мм"})
	fs := drawingTextFindings([]*Document{d}, []*Volume{v}, SideTarget)
	if len(fs) != 1 {
		t.Fatalf("want one finding (the thickness of the slab), got %+v", fs)
	}
	f := fs[0]
	if !strings.Contains(f.Description, "«Толщина» для «плита»") || !strings.Contains(f.Description, "600 мм") || !strings.Contains(f.Description, "500 мм") ||
		f.PageTarget != 2 || f.BoxTarget == nil || f.Kind != KindCrossVolume {
		t.Errorf("finding: %+v", f)
	}
	// the drawing names the value of the text among others: no finding
	d, v = docWithDrawing(SideRef, []string{"Толщина плиты 600 мм"}, []string{"Плита Пм1 толщиной 600 мм", "Плита Пм2 толщиной 400 мм"})
	if fs := drawingTextFindings([]*Document{d}, []*Volume{v}, SideRef); len(fs) != 0 {
		t.Errorf("no conflict expected: %+v", fs)
	}
}

func TestCrossVolumeDictionaryValues(t *testing.T) {
	d := twoVolumeDoc(SideRef, []string{"Отметка подошвы фундамента -2,400"}, []string{"Отметка подошвы фундамента -2,600"})
	fs, rows := crossOf(t, d, SideRef)
	if len(fs) != 1 || !strings.Contains(fs[0].Description, "«Отметка» для «фундамент»") {
		t.Errorf("findings: %+v", fs)
	}
	if len(rows) != 2 {
		t.Errorf("rows: %+v", rows)
	}
}
