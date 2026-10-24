package cmp

import (
	"strings"
	"testing"
)

// gridPage lays out rows of cells at the given column starts (a text page with words).
func gridPage(n int, starts []float64, rows ...[]string) Page {
	p := Page{Number: n, HasText: true}
	for i, r := range rows {
		y := 0.1 + 0.03*float64(i)
		var text []string
		for k, c := range r {
			x := starts[k]
			for _, w := range strings.Fields(c) {
				wd := 0.008 * float64(len([]rune(w)))
				p.Words = append(p.Words, Word{T: w, B: Box{x, y, x + wd, y + 0.012}})
				x += wd + 0.004
			}
			text = append(text, c)
		}
		t := strings.Join(text, " ")
		p.Lines = append(p.Lines, Line{Page: n, Text: t, Key: NormKey(t), Box: Box{starts[0], y, 0.9, y + 0.012}})
	}
	return p
}

var tblCols = []float64{0.1, 0.17, 0.3, 0.55, 0.7}

func TestExtractTables(t *testing.T) {
	p := gridPage(3, tblCols,
		[]string{"Общие данные по зданию приведены ниже в таблице"}, // prose: one cell, not a table
		[]string{"№", "Помещение", "Отделка пола", "Площадь", "Класс"},
		[]string{"1", "Тамбур", "керамогранит", "6,2", "Ф3.1"},
		[]string{"2", "Холл", "керамогранит", "24,0", "Ф3.1"},
		[]string{"3", "Кабинет", "линолеум", "18,5", "Ф4.3"},
	)
	ts := ExtractTables("d", p)
	if len(ts) != 1 {
		t.Fatalf("tables: %+v", ts)
	}
	tb := ts[0]
	if len(tb.Header) != 5 || tb.Header[1] != "Помещение" || len(tb.Rows) != 3 {
		t.Fatalf("table: %+v", tb)
	}
	if tb.keyCol != 1 { // the running number is not a key
		t.Errorf("key column %d", tb.keyCol)
	}
	if tb.Rows[1].Cells[2] != "керамогранит" || tb.Rows[2].Cells[3] != "18,5" {
		t.Errorf("cells: %+v", tb.Rows)
	}
	if ExtractTables("d", Page{Number: 1, Drawing: true, Words: p.Words}) != nil {
		t.Error("drawings are not read as tables")
	}
}

func TestCellFindings(t *testing.T) {
	ref := gridPage(3, tblCols,
		[]string{"№", "Помещение", "Отделка пола", "Площадь", "Класс"},
		[]string{"1", "Тамбур", "керамогранит", "6,2", "Ф3.1"},
		[]string{"2", "Холл", "керамогранит", "24,0", "Ф3.1"},
		[]string{"3", "Кабинет", "линолеум", "18,5", "Ф4.3"},
		[]string{"4", "Склад", "бетон", "12,0", "Ф5.2"},
	)
	tgt := gridPage(4, tblCols,
		[]string{"№", "Помещение", "Отделка пола", "Площадь", "Класс"},
		[]string{"1", "Тамбур", "керамогранит", "6,2", "Ф3.1"},
		[]string{"2", "Холл", "керамогранит", "30,0", "Ф3.1"}, // +25 %: high
		[]string{"3", "Кабинет", "ламинат", "18,5", "Ф4.3"},   // a word: low
		[]string{"4", "Санузел", "керамогранит", "3,1", "Ф3.1"},
	)
	p := &Pair{Ref: &Volume{Code: "X", Tables: ExtractTables("pd", ref)}, Targets: []*Volume{{Tables: ExtractTables("rd", tgt)}}}
	fs := cellFindings(p)
	byText := func(s string) *Finding {
		for i := range fs {
			if strings.Contains(fs[i].Description, s) {
				return &fs[i]
			}
		}
		return nil
	}
	if f := byText("строка «Холл», столбец «Площадь»"); f == nil || f.Severity != "high" || f.PageRef != 3 || f.PageTarget != 4 || f.BoxRef == nil || f.BoxTarget == nil {
		t.Errorf("area: %+v", f)
	}
	if f := byText("строка «Кабинет», столбец «Отделка пола»"); f == nil || f.Severity != "low" || f.Evidence.Ref != "линолеум" {
		t.Errorf("floor: %+v", f)
	}
	if byText("строки «Склад»") == nil || byText("строка «Санузел»") == nil {
		t.Errorf("removed / added rows: %+v", fs)
	}
	if len(fs) != 4 {
		t.Errorf("want 4 findings, got %d: %+v", len(fs), fs)
	}

	// the line findings of the same rows are dropped, others stay
	line := func(page int, y float64) Finding {
		b := Box{0.1, y, 0.9, y + 0.012}
		return Finding{Kind: KindTableRow, RefDoc: "pd", PageRef: page, BoxRef: &b}
	}
	all := dropCovered(append([]Finding{line(3, 0.13+0.03), line(3, 0.5)}, fs...))
	if len(all) != 5 {
		t.Errorf("dropCovered: %d", len(all))
	}
}
