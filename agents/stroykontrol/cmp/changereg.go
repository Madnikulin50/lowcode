package cmp

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Регистрация изменений (ТЗ 3.3.7, 3.3.8): при сравнении двух редакций тома («До» → «После»)
// агент читает, что заявлено в новой редакции, и сверяет с тем, что на самом деле изменилось.
// Заявлено:
//   - в таблице регистрации изменений (ТРИ, ГОСТ Р 21.101) — номера изменённых, заменённых, новых и
//     аннулированных листов по каждому изменению;
//   - в графах изменений основной надписи листа — «1 - Зам. 15-24 01.24»;
//   - в разрешении на внесение изменений — листы, которые оно разрешает менять (если оно в томе).
// Фактически: лист с расхождениями сравнения — изменён; лист, которого не было, — новый; лист, которого
// не стало, — аннулирован. Расхождение между заявленным и фактическим — находка вида change_reg.

const KindChangeReg = "change_reg"

// What happened to a sheet (fact) and what the edition says about it (declared).
const (
	SheetSame      = "same"
	SheetChanged   = "changed"
	SheetReplaced  = "replaced" // declared only: «заменённых», «Зам.»
	SheetNew       = "new"
	SheetAnnulled  = "annulled"
	SheetAbsent    = "absent" // fact only: not in either edition (declared for nothing)
	maxChangeSheet = 999
)

// SheetStateLabels are the Russian names of the sheet states.
var SheetStateLabels = map[string]string{
	SheetSame: "без изменений", SheetChanged: "изменён", SheetReplaced: "заменён", SheetNew: "новый",
	SheetAnnulled: "аннулирован", SheetAbsent: "нет в томе",
}

// SheetChange is the status of one sheet of a volume in the new edition.
type SheetChange struct {
	Volume     string `json:"volume"`           // the volume label
	Sheet      int    `json:"sheet"`            // the sheet number from the title block
	Change     int    `json:"change,omitempty"` // the change number the check is made for
	Fact       string `json:"fact"`
	Declared   string `json:"declared,omitempty"` // from the table of the change registration
	Stamp      string `json:"stamp,omitempty"`    // from the change rows of the sheet's title block
	Permitted  *bool  `json:"permitted,omitempty"`
	RefDoc     string `json:"refDoc,omitempty"`
	PageRef    int    `json:"pageRef,omitempty"`
	TargetDoc  string `json:"targetDoc,omitempty"`
	PageTarget int    `json:"pageTarget,omitempty"`
	Diffs      int    `json:"diffs,omitempty"` // findings of the comparison on the sheet
	OK         bool   `json:"ok"`
	Note       string `json:"note,omitempty"`
}

// ChangeTable is one row group of a table of the change registration: what one change declares.
type ChangeTable struct {
	Change int
	Sheets map[int]string // sheet → SheetChanged|SheetReplaced|SheetNew|SheetAnnulled
	Doc    string         // № док.
	Page   int
	Box    Box
}

var (
	triHeadRe   = regexp.MustCompile(`таблица регистрации изменений`)
	permitRe    = regexp.MustCompile(`разрешение.{0,40}внесени[ея] изменени`)
	permSheetRe = regexp.MustCompile(`(?i)(?:^|[^\p{L}])лист(?:ы|ов)?\s*(?:№\s*)?(\d{1,3}(?:\s*(?:[,;–—-]|и)\s*\d{1,3})*)`)
	triCols     = []struct{ stem, state string }{
		{"измененн", SheetChanged}, {"замененн", SheetReplaced}, {"новых", SheetNew}, {"аннулированн", SheetAnnulled},
	}
	stampKinds = map[string]string{"изм": SheetChanged, "зам": SheetReplaced, "нов": SheetNew, "аннул": SheetAnnulled}
)

// ParseChangeTables reads the tables of the change registration on the pages from..to of a document.
// The columns are found by their heads («изменённых», «заменённых», «новых», «аннулированных»); a row
// whose first cell is a number starts a change, the rows under it without one continue it.
func ParseChangeTables(d *Document, from, to int) []ChangeTable {
	var out []ChangeTable
	for pn := from; pn <= to && pn <= len(d.Pages); pn++ {
		p := d.Pages[pn-1]
		if len(p.Words) == 0 || !pageHas(p, triHeadRe, "изменен") {
			continue
		}
		out = append(out, changeTableOfPage(p, d.ID)...)
	}
	return out
}

func pageHas(p Page, re *regexp.Regexp, also string) bool {
	head, word := false, also == ""
	for _, l := range p.Lines {
		if l.Stamp {
			continue
		}
		if re.MatchString(l.Key) {
			head = true
		}
		if also != "" && strings.Contains(l.Key, also) {
			word = true
		}
	}
	return head && word
}

type triCol struct {
	state  string
	x0, x1 float64 // the head word
	lo, hi float64 // the column span
	y      float64
}

func changeTableOfPage(p Page, docID string) []ChangeTable {
	var cols []triCol
	for _, c := range triCols {
		for _, w := range p.Words {
			if k := NormKey(w.T); strings.HasPrefix(k, c.stem) && w.B.Y0 < 0.85 {
				cols = append(cols, triCol{state: c.state, x0: w.B.X0, x1: w.B.X1, y: w.B.Y1})
				break
			}
		}
	}
	if len(cols) < 2 {
		return nil
	}
	sort.Slice(cols, func(i, j int) bool { return cols[i].x0 < cols[j].x0 })
	headY := 0.0
	for i := range cols {
		if cols[i].y > headY {
			headY = cols[i].y
		}
		if i == 0 {
			continue
		}
		mid := (cols[i-1].x1 + cols[i].x0) / 2
		cols[i-1].hi, cols[i].lo = mid, mid
	}
	gap := (cols[1].x0 - cols[0].x1) / 2
	if gap < 0.01 {
		gap = 0.01
	}
	cols[0].lo = cols[0].x0 - gap
	cols[len(cols)-1].hi = cols[len(cols)-1].x1 + gap

	// the rows of the table body: the words under the heads, above the title block, grouped by height
	var body []Word
	for _, w := range p.Words {
		if w.B.Y0 > headY && w.B.Y0 < 0.86 {
			body = append(body, w)
		}
	}
	sort.SliceStable(body, func(i, j int) bool {
		if d := body[i].B.Y0 - body[j].B.Y0; d < -0.004 || d > 0.004 {
			return d < 0
		}
		return body[i].B.X0 < body[j].B.X0
	})
	var rows [][]Word
	for _, w := range body {
		if n := len(rows); n > 0 && w.B.Y0-rows[n-1][0].B.Y0 < 0.006 {
			rows[n-1] = append(rows[n-1], w)
			continue
		}
		rows = append(rows, []Word{w})
	}

	var out []ChangeTable
	for _, row := range rows {
		cells := map[string][]string{}
		num, doc := 0, ""
		box := Box{X0: 1, Y0: 1}
		for _, w := range row {
			cx := (w.B.X0 + w.B.X1) / 2
			box = box.union(w.B)
			placed := false
			for _, c := range cols {
				if cx >= c.lo && cx < c.hi {
					cells[c.state] = append(cells[c.state], w.T)
					placed = true
					break
				}
			}
			if placed {
				continue
			}
			if cx < cols[0].lo && num == 0 {
				if n, err := strconv.Atoi(strings.Trim(w.T, ".")); err == nil && n > 0 && n < 100 {
					num = n
				}
			} else if cx > cols[len(cols)-1].hi && doc == "" && strings.ContainsAny(w.T, "0123456789") && strings.ContainsAny(w.T, "-/") {
				doc = w.T
			}
		}
		if len(cells) == 0 && num == 0 {
			continue
		}
		if num == 0 && len(out) == 0 {
			continue // the rest of the head
		}
		if num != 0 {
			out = append(out, ChangeTable{Change: num, Sheets: map[int]string{}, Doc: doc, Page: p.Number, Box: box})
		}
		t := &out[len(out)-1]
		t.Box = t.Box.union(box)
		for _, c := range cols {
			for _, n := range expandSheets(strings.Join(cells[c.state], " ")) {
				if n > 0 && n <= maxChangeSheet {
					t.Sheets[n] = c.state
				}
			}
		}
	}
	return out
}

// stampChanges reads the change rows of a page's title block: change number → the declared state.
func stampChanges(p Page) map[int]string {
	out := map[int]string{}
	for _, l := range p.Lines {
		if !l.Stamp {
			continue
		}
		if m := changeRowRe.FindStringSubmatch(l.Key); m != nil {
			n, _ := strconv.Atoi(m[1])
			out[n] = stampKinds[m[2]]
		}
	}
	return out
}

// permittedSheets reads the sheets named in a permission to make changes, when it is in the volume;
// nil when there is none.
func permittedSheets(d *Document, from, to int) (map[int]bool, int) {
	for pn := from; pn <= to && pn <= len(d.Pages); pn++ {
		p := d.Pages[pn-1]
		if !pageHas(p, permitRe, "") {
			continue
		}
		out := map[int]bool{}
		for _, l := range p.Lines {
			if l.Stamp {
				continue
			}
			for _, m := range permSheetRe.FindAllStringSubmatch(l.Text, -1) {
				for _, n := range expandSheets(strings.ReplaceAll(m[1], ";", ",")) {
					out[n] = true
				}
			}
		}
		return out, pn
	}
	return nil, 0
}

// diffKinds are the findings that mean the content of a sheet changed.
var diffKinds = map[string]bool{KindText: true, KindTableRow: true, KindParam: true, KindDrawing: true, KindElement: true, KindVOR: true}

type volSheets struct {
	doc   string
	pages map[int]int // sheet → page
}

func sheetsOf(vs []*Volume, sheets map[string]map[int]int) []volSheets {
	var out []volSheets
	for _, v := range vs {
		s := volSheets{doc: v.DocID, pages: map[int]int{}}
		for page, no := range sheets[v.DocID] {
			if page >= v.FirstPage && page <= v.LastPage {
				if old, ok := s.pages[no]; !ok || page < old {
					s.pages[no] = page
				}
			}
		}
		out = append(out, s)
	}
	return out
}

func findSheet(vs []volSheets, no int) (string, int) {
	for _, v := range vs {
		if p, ok := v.pages[no]; ok {
			return v.doc, p
		}
	}
	return "", 0
}

func sheetAt(vs []volSheets, doc string, page int) (int, bool) {
	for _, v := range vs {
		if v.doc != doc {
			continue
		}
		for no, p := range v.pages {
			if p == page {
				return no, true
			}
		}
	}
	return 0, false
}

// changeRegistration checks the change registration of the new edition of a pair against what the
// comparison found (fs — the findings of the pair). It returns the statuses of the sheets, the findings
// and the notes; nothing when the new edition declares no change and is not newer than the old one.
func changeRegistration(p *Pair, docs map[string]*Document, sheets map[string]map[int]int, fs []Finding) ([]SheetChange, []Finding, []string) {
	refS := sheetsOf([]*Volume{p.Ref}, sheets)
	tgtS := sheetsOf(p.Targets, sheets)

	var tables []ChangeTable
	var triDoc string
	stamps := map[int]map[int]string{} // sheet → change → state
	var perm map[int]bool
	permPage, permDoc := 0, ""
	newest := 0
	for _, v := range p.Targets {
		d := docs[v.DocID]
		if d == nil {
			continue
		}
		if v.Change > newest {
			newest = v.Change
		}
		if ts := ParseChangeTables(d, v.FirstPage, v.LastPage); len(ts) > 0 {
			tables = append(tables, ts...)
			triDoc = d.ID
		}
		for pn := v.FirstPage; pn <= v.LastPage && pn <= len(d.Pages); pn++ {
			if no, ok := sheets[d.ID][pn]; ok {
				if sc := stampChanges(d.Pages[pn-1]); len(sc) > 0 {
					stamps[no] = sc
				}
			}
		}
		if perm == nil {
			if ps, pg := permittedSheets(d, v.FirstPage, v.LastPage); ps != nil {
				perm, permPage, permDoc = ps, pg, d.ID
			}
		}
	}
	// the change the check is made for: the edition of the volume, else the last one the table lists
	n := newest
	byNo := map[int]*ChangeTable{}
	for i := range tables {
		t := &tables[i]
		if old := byNo[t.Change]; old != nil {
			for k, s := range t.Sheets {
				old.Sheets[k] = s
			}
			continue
		}
		byNo[t.Change] = t
		if newest == 0 && t.Change > n {
			n = t.Change
		}
	}
	for _, sc := range stamps {
		for c := range sc {
			if newest == 0 && len(tables) == 0 && c > n {
				n = c
			}
		}
	}
	if n == 0 || p.Ref.Change >= n {
		return nil, nil, nil
	}
	label := volLabel(p.Targets[0])
	var notes []string
	var out []Finding
	add := func(sev, rule, text string, sc *SheetChange) {
		f := Finding{Kind: KindChangeReg, Type: "content", Severity: sev, Description: text,
			Evidence:    Evidence{VolumeRef: volLabel(p.Ref), VolumeTgt: label, Rule: rule},
			Fingerprint: fingerprint(KindChangeReg, p.Ref.Code, strconv.Itoa(n), rule, text)}
		if sc != nil {
			f.RefDoc, f.PageRef, f.TargetDoc, f.PageTarget = sc.RefDoc, sc.PageRef, sc.TargetDoc, sc.PageTarget
		}
		if f.TargetDoc == "" && f.RefDoc == "" {
			f.TargetDoc, f.PageTarget = p.Targets[0].DocID, p.Targets[0].FirstPage
		}
		out = append(out, f)
	}

	cur := byNo[n]
	triPages := map[int]bool{}
	for _, t := range tables {
		if no, ok := sheets[triDoc][t.Page]; ok {
			triPages[no] = true
		}
	}
	if permDoc != "" {
		if no, ok := sheets[permDoc][permPage]; ok {
			triPages[no] = true
		}
	}
	switch {
	case len(tables) == 0:
		add("low", "change-no-table", fmt.Sprintf("В томе «%s» (изм. %d) не найдена таблица регистрации изменений: заявленные изменения сверены только по графам изменений основной надписи.", label, n), nil)
	case cur == nil:
		nums := make([]string, 0, len(byNo))
		for c := range byNo {
			nums = append(nums, strconv.Itoa(c))
		}
		sort.Strings(nums)
		add("medium", "change-number", fmt.Sprintf("Том «%s» выпущен как изм. %d, а в таблице регистрации изменений нет строки изменения %d (есть: %s).", label, n, n, strings.Join(nums, ", ")), nil)
	}
	if perm == nil {
		notes = append(notes, fmt.Sprintf("Разрешение на внесение изменений в томе «%s» не найдено — проверьте его отдельно (ТЗ 3.3.9).", label))
	}

	// the facts: the sheets with differences on either side
	diffs := map[int]int{}
	for _, f := range fs {
		if !diffKinds[f.Kind] {
			continue
		}
		if no, ok := sheetAt(tgtS, f.TargetDoc, f.PageTarget); ok && f.PageTarget > 0 {
			diffs[no]++
		} else if no, ok := sheetAt(refS, f.RefDoc, f.PageRef); ok && f.PageRef > 0 {
			diffs[no]++
		}
	}
	all := map[int]bool{}
	for _, v := range append(append([]volSheets{}, refS...), tgtS...) {
		for no := range v.pages {
			all[no] = true
		}
	}
	if cur != nil {
		for no := range cur.Sheets {
			all[no] = true
		}
	}
	for no, sc := range stamps {
		if _, ok := sc[n]; ok {
			all[no] = true
		}
	}
	nos := make([]int, 0, len(all))
	for no := range all {
		nos = append(nos, no)
	}
	sort.Ints(nos)

	var res []SheetChange
	for _, no := range nos {
		sc := SheetChange{Volume: label, Sheet: no, Change: n, Diffs: diffs[no]}
		sc.RefDoc, sc.PageRef = findSheet(refS, no)
		sc.TargetDoc, sc.PageTarget = findSheet(tgtS, no)
		inRef, inTgt := sc.PageRef > 0, sc.PageTarget > 0
		switch {
		case inRef && inTgt && sc.Diffs > 0:
			sc.Fact = SheetChanged
		case inRef && inTgt:
			sc.Fact = SheetSame
		case inTgt:
			sc.Fact = SheetNew
		case inRef:
			sc.Fact = SheetAnnulled
		default:
			sc.Fact = SheetAbsent
		}
		if cur != nil {
			sc.Declared = cur.Sheets[no]
		}
		if st, ok := stamps[no][n]; ok {
			sc.Stamp = st
		}
		if perm != nil && sc.Declared != "" {
			ok := perm[no]
			sc.Permitted = &ok
		}
		declared := firstNonEmpty(sc.Declared, sc.Stamp)
		service := triPages[no] && sc.Fact == SheetNew // the table or the permission itself
		sheet := fmt.Sprintf("Лист %d тома «%s»", no, label)
		var issues []string
		problem := func(sev, rule, text string) {
			issues = append(issues, text)
			add(sev, rule, sheet+": "+text, &sc)
		}
		switch {
		case declared == "" && sc.Fact == SheetChanged:
			problem("high", "change-undeclared", fmt.Sprintf("изменён (расхождений: %d), но изменение %d его не называет — нет ни в таблице регистрации изменений, ни в графах изменений основной надписи.", sc.Diffs, n))
		case declared == "" && sc.Fact == SheetNew && !service:
			problem("medium", "change-undeclared-new", fmt.Sprintf("появился в изм. %d, но не заявлен как новый.", n))
		case declared == "" && sc.Fact == SheetAnnulled:
			problem("medium", "change-undeclared-annul", fmt.Sprintf("отсутствует в изм. %d, но не заявлен как аннулированный.", n))
		case declared == SheetChanged || declared == SheetReplaced:
			switch sc.Fact {
			case SheetSame:
				problem("low", "change-no-diff", fmt.Sprintf("заявлен как %s, но сравнение различий на нём не нашло.", SheetStateLabels[declared]))
			case SheetNew:
				problem("low", "change-kind", fmt.Sprintf("заявлен как %s, но в прежней редакции его не было — это новый лист.", SheetStateLabels[declared]))
			case SheetAnnulled, SheetAbsent:
				problem("medium", "change-kind", fmt.Sprintf("заявлен как %s, но в новой редакции его нет.", SheetStateLabels[declared]))
			}
		case declared == SheetNew && sc.Fact != SheetNew && !service:
			if sc.Fact == SheetAnnulled || sc.Fact == SheetAbsent {
				problem("medium", "change-kind", "заявлен как новый, но в новой редакции его нет.")
			} else {
				problem("low", "change-kind", "заявлен как новый, но был и в прежней редакции.")
			}
		case declared == SheetAnnulled && sc.Fact != SheetAnnulled && sc.Fact != SheetAbsent && !annulMarked(docs[sc.TargetDoc], sc.PageTarget):
			problem("medium", "change-annul-present", "заявлен как аннулированный, но остаётся в томе без отметки «Аннулирован».")
		}
		if sc.Declared != "" && sc.Stamp != "" && sc.Declared != sc.Stamp {
			problem("medium", "change-stamp", fmt.Sprintf("в таблице регистрации изменений — «%s», в графах изменений основной надписи — «%s».", SheetStateLabels[sc.Declared], SheetStateLabels[sc.Stamp]))
		}
		if cur != nil && sc.Stamp != "" && sc.Declared == "" {
			problem("medium", "change-stamp", fmt.Sprintf("в основной надписи отмечено изменение %d («%s»), а в таблице регистрации изменений лист не указан.", n, SheetStateLabels[sc.Stamp]))
		}
		if sc.Permitted != nil && !*sc.Permitted {
			problem("medium", "change-not-permitted", fmt.Sprintf("указан в изменении %d, но не назван в разрешении на внесение изменений.", n))
		}
		sc.OK = len(issues) == 0
		sc.Note = strings.Join(issues, " ")
		res = append(res, sc)
	}
	return res, out, notes
}

// annulMarked: the sheet stays in the volume crossed out — «Аннулирован» is written on it.
func annulMarked(d *Document, page int) bool {
	if d == nil || page < 1 || page > len(d.Pages) {
		return false
	}
	for _, l := range d.Pages[page-1].Lines {
		if strings.Contains(l.Key, "аннулирован") && !l.Stamp {
			return true
		}
	}
	return false
}
