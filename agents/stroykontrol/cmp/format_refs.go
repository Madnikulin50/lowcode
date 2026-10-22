package cmp

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// Проверки комплектности внутри файла (ТЗ 3.2.2–3.2.4, 3.2.8): ссылки на таблицы, приложения и
// разделы ведут к существующему заголовку; приложения, заявленные в содержании, есть; раздел не
// пуст; листы, перечисленные в ведомости рабочих чертежей, есть в томе.
//
// Ссылка на другой том, на нормативный или договорный документ не проверяется: её цель
// находится вне файла. Отсутствие ссылочного документа несоответствием не считается (ТЗ 3.2.8),
// поэтому такие ссылки здесь не разбираются вовсе. Как и прочие формальные проверки, находки
// «низкие» и всегда «к проверке».

const maxRefFindingsPerVolume = 4

var (
	// "Таблица 3", "Таблица А.1 — Нагрузки": a table heading at the start of a line
	tableHeadRe = regexp.MustCompile(`^таблица\s+((?:[а-яa-z]\.)?\d+(?:\.\d+)*)(?:[^\p{L}\p{N}]|$)`)
	// "по таблице 3", "см. табл. 2.1"
	tableRefRe = regexp.MustCompile(`табл(?:ица|ице|ицу|ицы|ицей|ицах|ицам|ицами|\.)?\s*((?:[а-яa-z]\.)?\d+(?:\.\d+)*)(?:[^\p{L}\p{N}]|$)`)
	// "Приложение А", "ПРИЛОЖЕНИЕ 2": the letter must be a capital one (not the preposition in "приложении в виде")
	appendixHeadRe = regexp.MustCompile(`^(?i:приложение)\s+([А-ЯA-Z]|\d{1,2})(?:[^\p{L}\p{N}]|$)`)
	appendixRefRe  = regexp.MustCompile(`(?i:приложени[а-яё]{1,2})\s+([А-ЯA-Z]|\d{1,2})(?:[^\p{L}\p{N}]|$)`)
	// "см. раздел 3", "в разделе 4.2"
	sectionRefRe = regexp.MustCompile(`раздел[а-яё]{0,3}\s+(\d{1,2}(?:\.\d{1,2})?)(?:[^\p{L}\p{N}]|$)`)
	// "3 Конструктивные решения", "3.2 Фундаменты"
	sectionHeadRe = regexp.MustCompile(`^(\d{1,2}(?:\.\d{1,2}){0,2})\.?\s+[А-ЯЁA-Z]`)
	tailNumRe     = regexp.MustCompile(`\S\s+\d{1,3}$`)
	// "Ведомость рабочих чертежей основного комплекта"
	drawingListRe = regexp.MustCompile(`^ведомость рабочих чертежей`)
	drawingRowRe  = regexp.MustCompile(`^(\d{1,3})(?:\s*-\s*(\d{1,3}))?\s+\S`)
	// the number of a set of rules or a standard: "35.13330.2011", "341.1325800.2017"
	normNumRe = regexp.MustCompile(`\d{1,3}\.1[0-9]{4,6}`)
	// the letters of the Russian appendix lettering and their Latin twins
	latinTwin = strings.NewReplacer("A", "А", "B", "В", "C", "С", "E", "Е", "H", "Н", "K", "К", "M", "М", "O", "О", "P", "Р", "T", "Т", "X", "Х")
)

// external words: a reference that has one of them leads out of the file
var externalWords = map[string]bool{
	"шифр": true, "гост": true, "снип": true, "пуэ": true, "сто": true, "договор": true, "договора": true, "контракт": true,
	"контракта": true, "задание": true, "задания": true, "постановление": true, "приказ": true, "регламент": true,
	"регламента": true, "документации": true, "документация": true, "нормативный": true, "нормативного": true,
}

// pointsOutside reports whether a line speaks of another volume, a norm or a contract, so that a
// "таблица N" / "приложение X" in it is not a reference into this file.
func pointsOutside(key string) bool {
	if normNumRe.MatchString(key) {
		return true // "35.13330", "341.1325800": the number of a set of rules
	}
	for _, w := range strings.FieldsFunc(key, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }) {
		if externalWords[w] || w == "сп" || len(w) > 2 && strings.HasPrefix(w, "сп") && unicode.IsDigit([]rune(w)[2]) {
			return true // "СП", "СП341"
		}
		if r := []rune(w); strings.HasPrefix(w, "том") && len(r) <= 6 { // том, тома, томе, тому, томов
			return true
		}
	}
	return false
}

func foldID(s string) string { return latinTwin.Replace(strings.ToUpper(s)) }

// refTargets are the headings of a file a reference may lead to.
type refTargets struct {
	tables, appendices, sections map[string]bool
}

// headRest is the text of a line after its heading prefix; a rest that ends with a page number is a
// line of a contents table, not a heading.
func contentsLike(text string, prefixLen int) bool {
	rest := strings.TrimSpace(text[prefixLen:])
	return rest != "" && tailNumRe.MatchString(rest)
}

func headingTargets(d *Document) refTargets {
	t := refTargets{tables: map[string]bool{}, appendices: map[string]bool{}, sections: map[string]bool{}}
	for _, p := range d.Pages {
		for _, l := range p.Lines {
			if l.Stamp {
				continue
			}
			if m := tableHeadRe.FindStringSubmatchIndex(l.Key); m != nil {
				if !contentsLike(l.Key, m[3]) {
					t.tables[l.Key[m[2]:m[3]]] = true
				}
				continue
			}
			text := strings.TrimSpace(l.Text)
			if m := appendixHeadRe.FindStringSubmatchIndex(text); m != nil {
				if !contentsLike(text, m[3]) {
					t.appendices[foldID(text[m[2]:m[3]])] = true
				}
				continue
			}
			if num, ok := sectionHeading(text); ok && len(p.Words) == 0 {
				t.sections[num] = true
			}
		}
	}
	return t
}

// checkReferences: "таблица N", "приложение X" and "раздел N" mentioned in the text of the volume must
// have a heading in the file. A reference is reported once per volume.
func checkReferences(d *Document, v *Volume, side string) []Finding {
	targets := headingTargets(d)
	// no numbered headings read at all: the file does not tell what a section number would lead to
	sections := len(targets.sections) >= 4
	type seen struct{ kind, id string }
	done := map[seen]bool{}
	var out []Finding
	for pn := v.FirstPage; pn <= v.LastPage && pn <= len(d.Pages) && len(out) < maxRefFindingsPerVolume; pn++ {
		p := d.Pages[pn-1]
		if p.Cover || p.Drawing {
			continue
		}
		prev := ""
		for _, l := range p.Lines {
			before := prev
			prev = l.Key
			if l.Stamp || len(out) >= maxRefFindingsPerVolume {
				continue
			}
			if tableHeadRe.MatchString(l.Key) || appendixHeadRe.MatchString(strings.TrimSpace(l.Text)) || pageRefRe.MatchString(strings.TrimSpace(l.Text)) {
				continue // a heading or a contents line is not a reference
			}
			// the cue may stand on the previous row of the same sentence ("…по СП 35.13330" / "(таблица 6.10)")
			if pointsOutside(l.Key) || pointsOutside(before) {
				continue
			}
			add := func(kind, id, label string) {
				k := seen{kind, id}
				if done[k] {
					return
				}
				done[k] = true
				out = append(out, formatFinding(d, v, side, pn, "format-ref-"+kind,
					fmt.Sprintf("Ссылка на «%s» (стр. %d): такого заголовка в файле не найдено.", label, pn)))
			}
			for _, m := range tableRefRe.FindAllStringSubmatch(l.Key, -1) {
				if !targets.tables[m[1]] {
					add("table", m[1], "Таблица "+m[1])
				}
			}
			for _, m := range appendixRefRe.FindAllStringSubmatch(strings.TrimSpace(l.Text), -1) {
				if !targets.appendices[foldID(m[1])] {
					add("appendix", foldID(m[1]), "Приложение "+foldID(m[1]))
				}
			}
			if sections {
				for _, m := range sectionRefRe.FindAllStringSubmatch(l.Key, -1) {
					if !targets.sections[m[1]] {
						add("section", m[1], "раздел "+m[1])
					}
				}
			}
		}
	}
	return out
}

// checkDeclaredAppendices: an appendix the contents of the volume lists must be in the file.
func checkDeclaredAppendices(d *Document, v *Volume, side string) []Finding {
	head, entries := contentsOf(d, v)
	if head == 0 {
		return nil
	}
	targets := headingTargets(d)
	var out []Finding
	for _, e := range entries {
		if len(out) >= maxRefFindingsPerVolume {
			break
		}
		m := appendixHeadRe.FindStringSubmatch(strings.TrimSpace(e.text))
		if m == nil {
			continue
		}
		id := foldID(m[1])
		if !targets.appendices[id] {
			out = append(out, formatFinding(d, v, side, head, "format-appendix",
				fmt.Sprintf("Содержание тома называет «Приложение %s» (стр. %d), но приложения с таким заголовком в файле нет.", id, e.page)))
		}
	}
	return out
}

// sectionHeading reads a numbered heading ("3 Конструктивные решения"). A line with digits after the
// number is a table row or a list item with a quantity, not a heading; a line that ends like a
// sentence is a clause.
func sectionHeading(text string) (string, bool) {
	text = strings.TrimSpace(text)
	m := sectionHeadRe.FindStringSubmatch(text)
	if m == nil || len([]rune(text)) >= 100 || tailNumRe.MatchString(text) {
		return "", false
	}
	rest := text[len(m[1]):]
	if strings.IndexFunc(rest, unicode.IsDigit) >= 0 || strings.HasSuffix(text, ".") && strings.Count(text, " ") > 6 || strings.HasSuffix(text, ";") {
		return "", false
	}
	return m[1], true
}

// sectionDepth is the depth of a numbered heading: "3" is 1, "3.2" is 2.
func sectionDepth(num string) int { return strings.Count(num, ".") + 1 }

// checkEmptySections: a numbered heading directly followed by a heading of the same or a higher
// level, or by the end of the volume, has no text. Only files with a numbering scheme of at least
// five headings are looked at, so that a numbered list is not taken for headings.
func checkEmptySections(d *Document, v *Volume, side string) []Finding {
	type head struct {
		num   string
		text  string
		page  int
		empty bool // no text lines after it
	}
	var all []head
	contents, _ := contentsOf(d, v)
	// a line of a volume's text pages in reading order (title block, contents page and drawings skipped)
	var last *head
	for pn := v.FirstPage; pn <= v.LastPage && pn <= len(d.Pages); pn++ {
		p := d.Pages[pn-1]
		if p.Cover || p.Drawing || pn == contents || len(p.Words) > 0 { // a table page: its numbers are rows
			continue
		}
		for _, l := range p.Lines {
			if l.Stamp || skipLine(l.Key) {
				continue
			}
			text := strings.TrimSpace(l.Text)
			if num, ok := sectionHeading(text); ok {
				all = append(all, head{num: num, text: text, page: pn, empty: true})
				last = &all[len(all)-1]
				continue
			}
			if last != nil {
				last.empty = false
			}
		}
	}
	if len(all) < 5 {
		return nil
	}
	var out []Finding
	for i, h := range all {
		if len(out) >= maxRefFindingsPerVolume {
			break
		}
		if !h.empty {
			continue
		}
		// the end of the volume, or the next heading is not deeper: nothing stands under this one
		if i+1 < len(all) && sectionDepth(all[i+1].num) > sectionDepth(h.num) {
			continue
		}
		out = append(out, formatFinding(d, v, side, h.page, "format-empty-section",
			fmt.Sprintf("Раздел «%s» (стр. %d) не содержит текста: сразу за заголовком идёт следующий.", trunc(h.text, 80), h.page)))
	}
	return out
}

// checkDrawingList: the sheets the «Ведомость рабочих чертежей» lists must exist in the volume,
// by the sheet numbers of the title blocks. It is judged only when most sheets have a readable
// number — otherwise a missing sheet is just an unread title block.
func checkDrawingList(d *Document, v *Volume, side string) []Finding {
	var listed []int
	listPages := map[int]bool{}
	starts := map[int]bool{} // pages with the heading; the page after one is its continuation
	for pn := v.FirstPage; pn <= v.LastPage && pn <= len(d.Pages); pn++ {
		p := d.Pages[pn-1]
		at := -1
		for i, l := range p.Lines {
			if !l.Stamp && drawingListRe.MatchString(l.Key) && len([]rune(l.Text)) < 90 {
				at = i
				break
			}
		}
		if at < 0 && !starts[pn-1] {
			continue
		}
		starts[pn] = at >= 0
		listPages[pn] = true
		for i, l := range p.Lines {
			if l.Stamp || (at >= 0 && i <= at) {
				continue
			}
			if m := drawingRowRe.FindStringSubmatch(strings.TrimSpace(l.Text)); m != nil {
				from, _ := strconv.Atoi(m[1])
				to := from
				if m[2] != "" {
					to, _ = strconv.Atoi(m[2])
				}
				if to < from || to-from > 100 {
					continue
				}
				for n := from; n <= to; n++ {
					listed = append(listed, n)
				}
			}
		}
	}
	if len(listed) < 3 {
		return nil
	}
	have := map[int]bool{}
	read, pages := 0, 0
	for pn := v.FirstPage; pn <= v.LastPage && pn <= len(d.Pages); pn++ {
		p := d.Pages[pn-1]
		if p.Cover {
			continue
		}
		pages++
		if no, ok := sheetNumber(p); ok {
			have[no] = true
			read++
		}
	}
	if pages == 0 || float64(read) < 0.6*float64(pages) {
		return nil
	}
	sort.Ints(listed)
	var out []Finding
	page := firstPage(listPages, v.FirstPage)
	for i, n := range listed {
		if len(out) >= maxRefFindingsPerVolume {
			break
		}
		if i > 0 && listed[i-1] == n || have[n] {
			continue
		}
		out = append(out, formatFinding(d, v, side, page, "format-sheet-list",
			fmt.Sprintf("Ведомость рабочих чертежей указывает лист %d, а листа с таким номером в основной надписи тома нет.", n)))
	}
	return out
}

func firstPage(pages map[int]bool, def int) int {
	min := 0
	for p := range pages {
		if min == 0 || p < min {
			min = p
		}
	}
	if min == 0 {
		return def
	}
	return min
}
