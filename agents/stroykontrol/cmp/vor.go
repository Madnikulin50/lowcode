package cmp

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Bill of quantities (ВОР) and specification tables are read from the words of
// a page: the header names the columns (name, unit, quantity), every number in
// the quantity column is an item, its name is the text of the name column
// nearest to it (quantities are usually centred in a cell whose name wraps
// over several rows).

// Table kinds.
const (
	TableVOR  = "vor"  // ведомость объёмов работ
	TableSpec = "spec" // спецификация оборудования, изделий и материалов
)

// Item is one position of a ВОР or a specification.
type Item struct {
	Table string // vor | spec
	Name  string
	Key   string // normalized name
	Unit  string
	Qty   float64
	DocID string
	Page  int
	Box   Box
}

var (
	vorTitleRe  = regexp.MustCompile(`ведомост[ьи] объ[её]мов работ`)
	specTitleRe = regexp.MustCompile(`спецификаци[яи] оборудования|спецификаци[яи] материалов`)
	qtyRe       = regexp.MustCompile(`^\d+(?:[.,]\d+)?$`)
	unitRe      = regexp.MustCompile(`^(м3|м³|м2|м²|м|пм|п\.м\.?|м\.п\.?|шт\.?|кг|т|компл\.?|л|км)$`)
)

func normUnit(u string) string {
	u = strings.ToLower(strings.TrimSpace(u))
	u = strings.NewReplacer("³", "3", "²", "2", ".", "").Replace(u)
	switch u {
	case "мп":
		return "пм"
	}
	return u
}

func parseQty(s string) (float64, bool) {
	if !qtyRe.MatchString(s) {
		return 0, false
	}
	v, err := strconv.ParseFloat(strings.ReplaceAll(s, ",", "."), 64)
	return v, err == nil
}

func (b Box) cx() float64 { return (b.X0 + b.X1) / 2 }
func (b Box) cy() float64 { return (b.Y0 + b.Y1) / 2 }

// pageTableKind says what kind of table a page holds ("" if none), from the
// titles printed on it (the heading and the title block).
func pageTableKind(p Page) string {
	vor, spec := false, false
	for _, l := range p.Lines {
		if vorTitleRe.MatchString(l.Key) {
			vor = true
		}
		if specTitleRe.MatchString(l.Key) {
			spec = true
		}
	}
	switch {
	case vor && !spec:
		return TableVOR
	case spec && !vor:
		return TableSpec
	}
	return ""
}

type columns struct {
	posRight float64 // right edge of the position-number column
	cu, cq   float64 // centres of the unit and quantity columns
	hdrBot   float64 // bottom of the header
}

// findColumns locates the header words «Наименование», «Ед. изм.», «Кол.».
func findColumns(p Page) (columns, bool) {
	var c columns
	var haveName, haveUnit, haveQty bool
	var posRight float64
	var nameY float64
	for _, w := range p.Words {
		lw := strings.ToLower(strings.Trim(w.T, ".,"))
		switch {
		case strings.HasPrefix(lw, "наименование") && !haveName && w.B.Y0 < 0.4:
			haveName, nameY = true, w.B.cy()
		case (lw == "ед" || lw == "единица") && !haveUnit && w.B.Y0 < 0.4:
			haveUnit, c.cu = true, w.B.cx()
		case (strings.HasPrefix(lw, "кол") || lw == "количество") && !haveQty && w.B.Y0 < 0.4 && w.B.X0 > 0.3:
			haveQty, c.cq = true, w.B.cx()
		}
	}
	if !haveName || !haveUnit || !haveQty {
		return c, false
	}
	// the position-number column header: "№", "п/п", "Поз." at the left on the header band
	for _, w := range p.Words {
		lw := strings.ToLower(strings.Trim(w.T, ".,"))
		if (lw == "№" || lw == "п/п" || lw == "поз" || lw == "n") && w.B.X0 < 0.25 && w.B.cy() < nameY+0.06 && w.B.cy() > nameY-0.06 {
			if w.B.X1 > posRight {
				posRight = w.B.X1
			}
		}
	}
	if posRight == 0 {
		posRight = 0.14
	}
	c.posRight = posRight
	// the header ends at the numbered row "1 2 3 4 5 6" or a few rows below the title
	c.hdrBot = nameY + 0.05
	for _, l := range p.Lines {
		if strings.HasPrefix(l.Key, "1 2 3") && l.Box.Y0 > nameY-0.02 && l.Box.Y0 < nameY+0.12 {
			c.hdrBot = l.Box.Y1
			break
		}
	}
	return c, true
}

// ExtractItems reads the positions of the ВОР / specification on one page;
// kind is the table kind the page belongs to (see volumeItems).
func ExtractItems(docID string, p Page, kind string) []Item {
	if kind == "" || len(p.Words) == 0 {
		return nil
	}
	col, ok := findColumns(p)
	if !ok {
		return nil
	}
	const bottom = 0.85 // above the title block
	type qrow struct {
		qty  float64
		unit string
		box  Box
		name []Word
	}
	var rows []*qrow
	for _, w := range qtyWords(p, col) {
		if w.B.cy() <= col.hdrBot || w.B.cy() >= bottom {
			continue
		}
		q, ok := parseQty(w.T)
		if !ok {
			continue
		}
		// the unit sits in its column on the same band
		unit := ""
		for _, u := range p.Words {
			if abs(u.B.cy()-w.B.cy()) < 0.008 && abs(u.B.cx()-col.cu) < 0.04 && unitRe.MatchString(strings.ToLower(u.T)) {
				unit = normUnit(u.T)
				break
			}
		}
		if unit == "" {
			continue
		}
		rows = append(rows, &qrow{qty: q, unit: unit, box: w.B})
	}
	if len(rows) == 0 {
		return nil
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].box.cy() < rows[j].box.cy() })
	// name words: left of the unit column, right of the position column
	for _, w := range p.Words {
		if w.B.cy() <= col.hdrBot || w.B.cy() >= bottom || w.B.X0 <= col.posRight+0.005 || w.B.X1 >= col.cu-0.03 {
			continue
		}
		best, bd := -1, 1.0
		for i, r := range rows {
			if d := abs(w.B.cy() - r.box.cy()); d < bd {
				best, bd = i, d
			}
		}
		if best >= 0 && bd < 0.07 {
			rows[best].name = append(rows[best].name, w)
		}
	}
	var out []Item
	for _, r := range rows {
		sort.SliceStable(r.name, func(i, j int) bool {
			if abs(r.name[i].B.cy()-r.name[j].B.cy()) > 0.006 {
				return r.name[i].B.cy() < r.name[j].B.cy()
			}
			return r.name[i].B.X0 < r.name[j].B.X0
		})
		var parts []string
		box := r.box
		for _, w := range r.name {
			parts = append(parts, w.T)
			box = box.union(w.B)
		}
		name := strings.Join(parts, " ")
		key := NormKey(name)
		if len([]rune(key)) < 3 {
			continue
		}
		out = append(out, Item{Table: kind, Name: name, Key: key, Unit: r.unit, Qty: r.qty, DocID: docID, Page: p.Number, Box: box})
	}
	return out
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// minTableItems: a page counts as a parsed table page only with this many
// positions; fewer is more likely a misread than a table.
const minTableItems = 3

// volumeItems reads the ВОР and specification positions of the pages
// from..to of a document. Only the first sheet of a table carries its title;
// the following sheets (which repeat just the column header) inherit the kind.
// tablePages are the pages that were read as tables: their lines are left out
// of the text diff, the positions are compared instead.
func volumeItems(d *Document, from, to int) (items []Item, tablePages map[int]bool) {
	tablePages = map[int]bool{}
	last := ""
	for pn := from; pn <= to; pn++ {
		p := d.Pages[pn-1]
		if p.Drawing || p.Cover {
			continue
		}
		kind := pageTableKind(p)
		if kind == "" {
			if _, ok := findColumns(p); ok {
				kind = last
			}
		}
		if kind == "" {
			continue
		}
		last = kind
		got := ExtractItems(d.ID, p, kind)
		if len(got) >= minTableItems {
			items = append(items, got...)
			tablePages[pn] = true
		}
	}
	return items, tablePages
}

var (
	thousandsHead = regexp.MustCompile(`^\d{1,3}$`)
	thousandsTail = regexp.MustCompile(`^\d{3}(?:[.,]\d+)?$`)
)

// qtyWords returns the numbers of the quantity column. A number written with
// a space as the thousands separator ("6 322,83") arrives as two words; they
// are joined, otherwise "6" would be read as the quantity.
func qtyWords(p Page, col columns) []Word {
	var in []Word
	for _, w := range p.Words {
		if abs(w.B.cx()-col.cq) <= 0.05 {
			in = append(in, w)
		}
	}
	sort.SliceStable(in, func(i, j int) bool {
		if abs(in[i].B.cy()-in[j].B.cy()) > 0.006 {
			return in[i].B.cy() < in[j].B.cy()
		}
		return in[i].B.X0 < in[j].B.X0
	})
	var out []Word
	for i := 0; i < len(in); i++ {
		w := in[i]
		if i+1 < len(in) && thousandsHead.MatchString(w.T) && thousandsTail.MatchString(in[i+1].T) &&
			abs(in[i+1].B.cy()-w.B.cy()) < 0.006 && in[i+1].B.X0-w.B.X1 < 0.02 {
			out = append(out, Word{T: w.T + in[i+1].T, B: w.B.union(in[i+1].B)})
			i++
			continue
		}
		if qtyRe.MatchString(w.T) && abs(w.B.cx()-col.cq) <= 0.045 {
			out = append(out, w)
		}
	}
	return out
}
