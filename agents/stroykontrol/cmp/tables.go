package cmp

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// Таблицы общего вида (ТЗ 3.3.3): кроме ВОР и спецификаций (vor.go), таблицы текстовых листов
// разбираются на ячейки по словам страницы. Строка таблицы — слова на одной высоте; ячейки — группы
// слов, разделённые промежутком; таблица — не меньше трёх строк подряд, ячейки которых стоят под
// столбцами первой строки (шапки). Таблицы двух редакций сопоставляются по шапке, строки — по
// ключевому столбцу (первый столбец с неповторяющимися значениями, не номер по порядку), ячейки — по
// названию столбца. Изменённая ячейка — находка «таблица» с местом строки на обеих страницах; строчные
// находки текстового сравнения для этой же строки снимаются, чтобы расхождение не считалось дважды.
// Таблицы, нарисованные без текстового слоя, и таблицы на чертежах не разбираются.

const (
	rowTol      = 0.006 // words of one row: the centres differ by less (page fractions)
	cellGap     = 0.012 // a gap wider than this between two words splits cells
	colTol      = 0.02
	minCols     = 3
	minRows     = 3
	maxRowGap   = 0.06
	wrapGap     = 0.025 // a row closer than this to the previous one, with an empty first column, continues it
	maxCellInfo = 12    // cell findings of one table pair, the rest are counted in one finding
)

// TableRow is a row of a table.
type TableRow struct {
	Cells []string
	Box   Box
}

// Table is a table read from the words of a page.
type Table struct {
	DocID  string
	Page   int
	Header []string
	Rows   []TableRow
	Box    Box
	Starts []float64 // the left edges of the columns
	keyCol int
}

type cell struct {
	text string
	box  Box
}

func rowsOfWords(ws []Word) [][]Word {
	sort.SliceStable(ws, func(i, j int) bool { return ws[i].B.cy() < ws[j].B.cy() })
	var rows [][]Word
	for _, w := range ws {
		if n := len(rows); n > 0 && abs(rows[n-1][0].B.cy()-w.B.cy()) < rowTol {
			rows[n-1] = append(rows[n-1], w)
			continue
		}
		rows = append(rows, []Word{w})
	}
	for _, r := range rows {
		sort.SliceStable(r, func(i, j int) bool { return r[i].B.X0 < r[j].B.X0 })
	}
	return rows
}

func cellsOf(r []Word) []cell {
	var out []cell
	for _, w := range r {
		if n := len(out); n > 0 && w.B.X0-out[n-1].box.X1 < cellGap {
			out[n-1].text += " " + w.T
			out[n-1].box = out[n-1].box.union(w.B)
			continue
		}
		out = append(out, cell{w.T, w.B})
	}
	return out
}

// columnOf is the column whose span (from its start to the next start) holds the cell's centre.
func columnOf(starts []float64, c cell) int {
	x := c.box.cx()
	col := -1
	for i, s := range starts {
		if x >= s-colTol {
			col = i
		}
	}
	return col
}

// ExtractTables reads the tables of a text page.
func ExtractTables(docID string, p Page) []Table {
	if p.Drawing || p.Cover {
		return nil
	}
	return extractTables(docID, p)
}

func extractTables(docID string, p Page) []Table { return extractTablesMin(docID, p, minRows) }

// extractTablesMin reads the tables with at least min rows (a continuation of a table may have one).
func extractTablesMin(docID string, p Page, min int) []Table {
	if len(p.Words) == 0 {
		return nil
	}
	var stamps []Box
	for _, l := range p.Lines {
		if l.Stamp {
			stamps = append(stamps, l.Box)
		}
	}
	var ws []Word
	for _, w := range p.Words {
		inStamp := false
		for _, b := range stamps {
			if w.B.cx() >= b.X0 && w.B.cx() <= b.X1 && w.B.cy() >= b.Y0-0.002 && w.B.cy() <= b.Y1+0.002 {
				inStamp = true
			}
		}
		if !inStamp {
			ws = append(ws, w)
		}
	}
	rows := rowsOfWords(ws)
	var out []Table
	for i := 0; i < len(rows); {
		head := cellsOf(rows[i])
		if len(head) < minCols {
			i++
			continue
		}
		starts := make([]float64, len(head))
		for k, c := range head {
			starts[k] = c.box.X0
		}
		t := Table{DocID: docID, Page: p.Number, Box: rowBox(head), Starts: starts}
		for _, c := range head {
			t.Header = append(t.Header, c.text)
		}
		last := rows[i][0].B.cy()
		j := i + 1
		for ; j < len(rows); j++ {
			if rows[j][0].B.cy()-last > maxRowGap {
				break
			}
			cs := cellsOf(rows[j])
			if len(cs) < 2 && (len(t.Rows) == 0 || columnOf(starts, cs[0]) < 1) {
				break
			}
			cells := make([]string, len(head))
			aligned := 0
			for _, c := range cs {
				k := columnOf(starts, c)
				if k < 0 {
					continue
				}
				if abs(c.box.X0-starts[k]) < colTol || (k+1 < len(starts) && c.box.X1 < starts[k+1]) || k == len(starts)-1 {
					aligned++
				}
				cells[k] = strings.TrimSpace(cells[k] + " " + c.text)
			}
			if aligned*10 < len(cs)*7 {
				break
			}
			if n := len(t.Rows); n > 0 && cells[0] == "" && rows[j][0].B.cy()-last < wrapGap {
				// the first column is empty right under a row: the wrapped text of its cells
				for k, v := range cells {
					t.Rows[n-1].Cells[k] = joinWrapped(t.Rows[n-1].Cells[k], v)
				}
				t.Rows[n-1].Box = t.Rows[n-1].Box.union(rowBox(cs))
				t.Box = t.Box.union(rowBox(cs))
				last = rows[j][0].B.cy()
				continue
			}
			t.Rows = append(t.Rows, TableRow{Cells: cells, Box: rowBox(cs)})
			t.Box = t.Box.union(rowBox(cs))
			last = rows[j][0].B.cy()
		}
		if len(t.Rows) >= min {
			t.keyCol = keyColumn(t)
			out = append(out, t)
			i = j
			continue
		}
		i++
	}
	return out
}

// joinWrapped joins the lines of a wrapped cell («2024-15-» + «КЖ2» → «2024-15-КЖ2»).
func joinWrapped(a, b string) string {
	if b == "" {
		return a
	}
	if strings.HasSuffix(a, "-") && !strings.HasSuffix(a, " -") {
		return a + b
	}
	return strings.TrimSpace(a + " " + b)
}

func rowBox(cs []cell) Box {
	b := cs[0].box
	for _, c := range cs[1:] {
		b = b.union(c.box)
	}
	return b
}

// keyColumn is the first column whose values are filled, unique and not a running number.
func keyColumn(t Table) int {
	for k := range t.Header {
		seen := map[string]bool{}
		ok, running := true, true
		for i, r := range t.Rows {
			v := NormKey(r.Cells[k])
			if v == "" || seen[v] {
				ok = false
				break
			}
			seen[v] = true
			if v != fmt.Sprint(i+1) && v != fmt.Sprint(i+1)+"." {
				running = false
			}
		}
		if ok && !running {
			return k
		}
	}
	return 0
}

func (t Table) key(r TableRow) string { return NormKey(r.Cells[t.keyCol]) }

func (t Table) label() string {
	return trunc(strings.Join(t.Header, " | "), 80)
}

// headerTokens are the words of the header, for matching tables of two editions.
func headerTokens(h []string) map[string]bool {
	out := map[string]bool{}
	for _, c := range h {
		for _, w := range strings.Fields(NormKey(c)) {
			out[w] = true
		}
	}
	return out
}

// volumeTables reads the tables of the text pages of a volume (not the ВОР and specification pages).
func volumeTables(d *Document, from, to int, skip map[int]bool) []Table {
	var out []Table
	for pn := from; pn <= to; pn++ {
		if skip[pn] {
			continue
		}
		out = append(out, ExtractTables(d.ID, d.Pages[pn-1])...)
	}
	return out
}

// isNumber: the cell holds a value (numbers compare stronger than words).
func isNumber(s string) bool {
	s = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(s), " ", ""), ",", ".")
	if s == "" {
		return false
	}
	for _, r := range s {
		if !(r >= '0' && r <= '9' || r == '.' || r == '-' || r == '+') {
			return false
		}
	}
	return true
}

// cellFindings compares the tables of the reference volume with those of its target volumes.
func cellFindings(p *Pair) []Finding {
	var tgt []Table
	for _, t := range p.Targets {
		tgt = append(tgt, t.Tables...)
	}
	var out []Finding
	used := map[int]bool{}
	for _, a := range p.Ref.Tables {
		best, bs := -1, 0.0
		for j, b := range tgt {
			if used[j] {
				continue
			}
			if s := dice(headerTokens(a.Header), headerTokens(b.Header)); s > bs {
				best, bs = j, s
			}
		}
		if best < 0 || bs < 0.6 {
			continue
		}
		used[best] = true
		out = append(out, compareTables(p, a, tgt[best])...)
	}
	return out
}

func compareTables(p *Pair, a, b Table) []Finding {
	// the columns of b by the name of a's
	colB := map[int]int{}
	for i, h := range a.Header {
		for j, g := range b.Header {
			if NormKey(h) == NormKey(g) {
				colB[i] = j
				break
			}
		}
	}
	rowsB := map[string]TableRow{}
	for _, r := range b.Rows {
		rowsB[b.key(r)] = r
	}
	label := a.label()
	var out []Finding
	extra := 0
	add := func(f Finding) {
		if len(out) >= maxCellInfo {
			extra++
			return
		}
		out = append(out, f)
	}
	seenA := map[string]bool{}
	for _, ra := range a.Rows {
		k := a.key(ra)
		seenA[k] = true
		rb, ok := rowsB[k]
		bxA, bxB := ra.Box, rb.Box
		if !ok {
			add(Finding{Kind: KindTableCell, Type: "content", Severity: "low",
				Description: fmt.Sprintf("Таблица «%s»: строки «%s» (ПД стр. %d) нет в РД", label, trunc(ra.Cells[a.keyCol], 80), a.Page),
				RefDoc:      a.DocID, PageRef: a.Page, BoxRef: &bxA, Anchor: trunc(strings.Join(ra.Cells, " "), 200),
				Evidence:    Evidence{Ref: strings.Join(ra.Cells, " | "), Rule: "table-row-removed"},
				Fingerprint: fingerprint(KindTableCell, p.Ref.Code, label, k, "-")})
			continue
		}
		for i, va := range ra.Cells {
			j, ok := colB[i]
			if !ok || i == a.keyCol {
				continue
			}
			vb := rb.Cells[j]
			if NormKey(va) == NormKey(vb) {
				continue
			}
			sev := "low"
			if isNumber(va) && isNumber(vb) {
				sev = "medium"
				if x, y := parseNum(va), parseNum(vb); x != 0 && math.Abs(y-x)/math.Abs(x) > 0.1 {
					sev = "high"
				}
			}
			add(Finding{Kind: KindTableCell, Type: "content", Severity: sev,
				Description: fmt.Sprintf("Таблица «%s», строка «%s», столбец «%s»: в ПД «%s» (стр. %d), в РД «%s» (стр. %d)", label, trunc(ra.Cells[a.keyCol], 60), a.Header[i], trunc(va, 60), a.Page, trunc(vb, 60), b.Page),
				RefDoc:      a.DocID, TargetDoc: b.DocID, PageRef: a.Page, PageTarget: b.Page, BoxRef: &bxA, BoxTarget: &bxB,
				Anchor:      trunc(strings.Join(ra.Cells, " "), 200),
				Evidence:    Evidence{Ref: va, Target: vb, Rule: "table-cell"},
				Fingerprint: fingerprint(KindTableCell, p.Ref.Code, label, k, a.Header[i], va, vb)})
		}
	}
	for _, rb := range b.Rows {
		k := b.key(rb)
		if seenA[k] {
			continue
		}
		bxB := rb.Box
		add(Finding{Kind: KindTableCell, Type: "content", Severity: "low",
			Description: fmt.Sprintf("Таблица «%s»: строка «%s» (РД стр. %d) добавлена", label, trunc(rb.Cells[b.keyCol], 80), b.Page),
			TargetDoc:   b.DocID, PageTarget: b.Page, BoxTarget: &bxB, Anchor: trunc(strings.Join(rb.Cells, " "), 200),
			Evidence:    Evidence{Target: strings.Join(rb.Cells, " | "), Rule: "table-row-added"},
			Fingerprint: fingerprint(KindTableCell, p.Ref.Code, label, "-", k)})
	}
	if extra > 0 {
		out = append(out, Finding{Kind: KindTableCell, Type: "content", Severity: "low",
			Description: fmt.Sprintf("Таблица «%s»: ещё %d различий по ячейкам — откройте таблицу на стр. %d / %d", label, extra, a.Page, b.Page),
			RefDoc:      a.DocID, TargetDoc: b.DocID, PageRef: a.Page, PageTarget: b.Page, BoxRef: &a.Box, BoxTarget: &b.Box,
			Evidence: Evidence{Rule: "table-cell-more"}, Fingerprint: fingerprint(KindTableCell, p.Ref.Code, label, "more")})
	}
	return out
}

func parseNum(s string) float64 {
	v, _ := parseQty(strings.ReplaceAll(strings.TrimSpace(s), " ", ""))
	return v
}

// dropCovered removes the line findings of the text comparison that a table finding already describes:
// the same page and a line box inside the row of the table on that side.
func dropCovered(fs []Finding) []Finding {
	type place struct {
		doc  string
		page int
		box  Box
	}
	var rows []place
	for _, f := range fs {
		if f.Kind != KindTableCell {
			continue
		}
		if f.BoxRef != nil {
			rows = append(rows, place{f.RefDoc, f.PageRef, *f.BoxRef})
		}
		if f.BoxTarget != nil {
			rows = append(rows, place{f.TargetDoc, f.PageTarget, *f.BoxTarget})
		}
	}
	if len(rows) == 0 {
		return fs
	}
	inside := func(doc string, page int, b *Box) bool {
		if b == nil {
			return false
		}
		for _, r := range rows {
			if r.doc == doc && r.page == page && b.cy() >= r.box.Y0-0.004 && b.cy() <= r.box.Y1+0.004 {
				return true
			}
		}
		return false
	}
	out := fs[:0]
	for _, f := range fs {
		if (f.Kind == KindText || f.Kind == KindTableRow) && (inside(f.RefDoc, f.PageRef, f.BoxRef) || inside(f.TargetDoc, f.PageTarget, f.BoxTarget)) {
			continue
		}
		out = append(out, f)
	}
	return out
}
