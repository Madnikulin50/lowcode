package cmp

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// Проверки оформления внутри одной стороны (ПД или РД): сквозная нумерация листов в основной
// надписи, ссылки «стр. N» в содержании тома, единый шифр, листы-дубли. Это не сравнение с эталоном
// и не оценка проектных решений: находки низкой критичности, всегда «к проверке» специалистом.

const maxFormatPerVolume = 6

var (
	// the stage letter, the sheet number and (on the first sheet) the total: "П 1 6", "Р 1", "Р 3"
	sheetRowRe = regexp.MustCompile(`(?:^|\s)(?:П|Р|РП)\s+(\d{1,3})(?:\s+(\d{1,3}))?\s*$`)
	// the sheet number printed after the column heads of the next sheets: "Изм. Кол.уч Лист № док Подп. Дата 9"
	sheetDateRe = regexp.MustCompile(`(?i)(?:^|[^\p{L}])дат[а]?\s+(\d{1,3})\s*$`)
	pageRefRe   = regexp.MustCompile(`(?i)^(.*?)\s*стр\.?\s*(\d{1,3})\s*$`)
)

// sheetNumber reads the sheet number from the title block of a page.
func sheetNumber(p Page) (int, bool) {
	for _, l := range p.Lines {
		if !l.Stamp {
			continue
		}
		if roleRe.MatchString(l.Key) {
			if m := sheetRowRe.FindStringSubmatch(l.Text); m != nil {
				n, _ := strconv.Atoi(m[1])
				return n, true
			}
		}
		if strings.Contains(l.Key, "изм") {
			if m := sheetDateRe.FindStringSubmatch(l.Text); m != nil {
				n, _ := strconv.Atoi(m[1])
				return n, true
			}
		}
	}
	return 0, false
}

func formatFinding(d *Document, v *Volume, side string, page int, rule, text string) Finding {
	f := Finding{
		Kind: KindFormat, Type: "format", Severity: "low",
		Description: text,
		Evidence:    Evidence{Rule: rule},
		Fingerprint: fingerprint(KindFormat, side, d.ID, v.Code, rule, strconv.Itoa(page), text),
	}
	label := volLabel(v)
	if side == SideRef {
		f.RefDoc, f.PageRef, f.Evidence.VolumeRef = d.ID, page, label
	} else {
		f.TargetDoc, f.PageTarget, f.Evidence.VolumeTgt = d.ID, page, label
	}
	return f
}

// FormatChecks runs the formatting checks over one volume.
func FormatChecks(d *Document, v *Volume, side string) []Finding {
	var out []Finding
	out = append(out, checkSheetNumbers(d, v, side)...)
	out = append(out, checkVolumeCode(d, v, side)...)
	out = append(out, checkDuplicates(d, v, side)...)
	out = append(out, checkContents(d, v, side)...)
	return out
}

// checkSheetNumbers: between two consecutive pages with a readable sheet number the number must
// grow by the number of pages between them; a repeat, a step back or a jump is reported once, then
// the count goes on from the new number.
func checkSheetNumbers(d *Document, v *Volume, side string) []Finding {
	type sn struct{ page, no int }
	var seq []sn
	for pn := v.FirstPage; pn <= v.LastPage && pn <= len(d.Pages); pn++ {
		if no, ok := sheetNumber(d.Pages[pn-1]); ok {
			seq = append(seq, sn{pn, no})
		}
	}
	if len(seq) < 4 {
		return nil // too few readable title blocks to say anything about the numbering
	}
	var out []Finding
	for i := 1; i < len(seq) && len(out) < maxFormatPerVolume; i++ {
		prev, cur := seq[i-1], seq[i]
		dp, ds := cur.page-prev.page, cur.no-prev.no
		// a volume file often holds several documents, each numbered from sheet 1 on
		if ds == dp || cur.no == 1 {
			continue
		}
		var text string
		switch {
		case cur.no == prev.no:
			text = fmt.Sprintf("Номер листа %d в основной надписи повторяется: стр. %d и стр. %d.", cur.no, prev.page, cur.page)
		case ds < 0:
			text = fmt.Sprintf("Нарушен порядок листов: на стр. %d — лист %d, на следующей отмеченной стр. %d — лист %d.", prev.page, prev.no, cur.page, cur.no)
		default:
			text = fmt.Sprintf("Нумерация листов не сходится со страницами: стр. %d — лист %d, стр. %d — лист %d (между ними %d стр., а номер вырос на %d).", prev.page, prev.no, cur.page, cur.no, dp, ds)
		}
		out = append(out, formatFinding(d, v, side, cur.page, "format-sheet-number", text))
	}
	return out
}

func commonPrefix(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	n := 0
	for n < len(ra) && n < len(rb) && ra[n] == rb[n] {
		n++
	}
	return n
}

func codePrefixLess(a, b string) bool { return strings.HasPrefix(a, b) || strings.HasPrefix(b, a) }

// checkVolumeCode: a sheet whose title block names another шифр than the volume has.
func checkVolumeCode(d *Document, v *Volume, side string) []Finding {
	if v.Code == "" {
		return nil
	}
	type miss struct {
		page int
		code string
	}
	var bad []miss
	read := 0
	for pn := v.FirstPage; pn <= v.LastPage && pn <= len(d.Pages); pn++ {
		p := d.Pages[pn-1]
		if p.Cover {
			continue
		}
		raw := pageCode(p)
		if raw == "" {
			continue
		}
		read++
		n := NormCode(raw)
		if n == v.Code || codePrefixLess(n, v.Code) || volumesMatch(raw, v.RawCode) {
			continue
		}
		// only a шифр of the same project (the same start) is a mistake in the title block; any other
		// code-like text in the stamp area (a mark of a line, a drawing number) is not
		if commonPrefix(n, v.Code) < 8 {
			continue
		}
		bad = append(bad, miss{pn, raw})
	}
	// when most sheets differ it is the volume code that was misread, not the sheets that are wrong
	if len(bad) == 0 || read == 0 || float64(len(bad)) > 0.3*float64(read) {
		return nil
	}
	var out []Finding
	for _, b := range bad {
		if len(out) >= maxFormatPerVolume {
			break
		}
		out = append(out, formatFinding(d, v, side, b.page, "format-code",
			fmt.Sprintf("В основной надписи стр. %d указан шифр «%s», у тома — «%s».", b.page, b.code, v.RawCode)))
	}
	return out
}

// checkDuplicates: two sheets of a volume with the same text.
func checkDuplicates(d *Document, v *Volume, side string) []Finding {
	seen := map[string]int{}
	var out []Finding
	for pn := v.FirstPage; pn <= v.LastPage && pn <= len(d.Pages) && len(out) < maxFormatPerVolume; pn++ {
		p := d.Pages[pn-1]
		if p.Cover || p.Drawing {
			continue
		}
		var keys []string
		for _, l := range p.Lines {
			if !l.Stamp && !skipLine(l.Key) {
				keys = append(keys, l.Key)
			}
		}
		if len(keys) < 8 {
			continue
		}
		h := strings.Join(keys, "\n")
		if first, ok := seen[h]; ok {
			out = append(out, formatFinding(d, v, side, pn, "format-duplicate", fmt.Sprintf("Лист на стр. %d дублирует лист на стр. %d (одинаковый текст).", pn, first)))
			continue
		}
		seen[h] = pn
	}
	return out
}

var contentsHeadRe = regexp.MustCompile(`(?i)^(содержание тома|содержание|состав тома)(?:$|[^\p{L}])`)

func contentWords(s string) []string {
	var out []string
	for _, w := range strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) }) {
		if len([]rune(w)) >= 5 {
			out = append(out, w)
		}
	}
	return out
}

// checkContents: the «Содержание тома» refers to sheets as "стр. N" — every reference must lead to
// a page of the file, and the section named must be on a text page it points to.
func checkContents(d *Document, v *Volume, side string) []Finding {
	// the contents page: within the first pages of the volume, with the heading "Содержание тома"
	head := 0
	for pn := v.FirstPage; pn <= v.LastPage && pn <= len(d.Pages) && pn < v.FirstPage+6; pn++ {
		for _, l := range d.Pages[pn-1].Lines {
			if !l.Stamp && contentsHeadRe.MatchString(l.Key) && len([]rune(l.Text)) < 40 {
				head = pn
			}
		}
		if head != 0 {
			break
		}
	}
	if head == 0 {
		return nil
	}
	type entry struct {
		text string
		page int
	}
	var entries []entry
	var pending []string
	for _, l := range d.Pages[head-1].Lines {
		if l.Stamp || skipLine(l.Key) {
			continue
		}
		if m := pageRefRe.FindStringSubmatch(strings.TrimSpace(l.Text)); m != nil {
			n, _ := strconv.Atoi(m[2])
			text := strings.TrimSpace(strings.Join(append(pending, m[1]), " "))
			pending = nil
			if n > 0 {
				entries = append(entries, entry{text, n})
			}
			continue
		}
		if contentsHeadRe.MatchString(l.Key) || strings.Contains(l.Key, "обозначение наименование") {
			pending = nil
			continue
		}
		pending = append(pending, l.Text)
		if len(pending) > 6 { // a stretch with no page reference is not an entry
			pending = pending[len(pending)-3:]
		}
	}
	if len(entries) < 3 {
		return nil
	}
	var out []Finding
	for _, e := range entries {
		if len(out) >= maxFormatPerVolume {
			break
		}
		if e.page > len(d.Pages) {
			out = append(out, formatFinding(d, v, side, head, "format-contents",
				fmt.Sprintf("Содержание тома ссылается на стр. %d («%s»), а в файле всего %d стр.", e.page, trunc(e.text, 80), len(d.Pages))))
			continue
		}
		target := d.Pages[e.page-1]
		if target.Drawing || !target.HasText {
			continue // nothing to read on that page
		}
		words := contentWords(e.text)
		if len(words) == 0 {
			continue
		}
		var body strings.Builder
		for _, l := range target.Lines {
			body.WriteString(l.Key + " ")
		}
		text := body.String()
		hit := false
		for _, w := range words {
			stem := w
			if r := []rune(w); len(r) > 6 {
				stem = string(r[:len(r)-2]) // endings differ: «свайного» / «свайное»
			}
			if strings.Contains(text, stem) {
				hit = true
				break
			}
		}
		if !hit {
			out = append(out, formatFinding(d, v, side, head, "format-contents",
				fmt.Sprintf("Содержание тома: «%s» — стр. %d, но на этой странице такого раздела не найдено.", trunc(e.text, 80), e.page)))
		}
	}
	return out
}

// formatFindings runs the checks over all the volumes of the documents of one side.
func formatFindings(docs []*Document, vols []*Volume, side string) []Finding {
	byID := map[string]*Document{}
	for _, d := range docs {
		byID[d.ID] = d
	}
	var out []Finding
	for _, v := range vols {
		if d := byID[v.DocID]; d != nil {
			out = append(out, FormatChecks(d, v, side)...)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return false })
	return out
}
