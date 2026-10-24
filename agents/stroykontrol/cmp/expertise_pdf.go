package cmp

import (
	"fmt"
	"regexp"
	"strings"
)

// Замечания экспертизы и ответы из PDF (ТЗ 2.2.3, 3.4.1). Образца выгрузки в PDF нет, поэтому разбор
// общий. Если в PDF есть таблица со столбцом «Замечание» (по словам страницы, tables.go), её строки
// читаются тем же разбором, что и Excel: сначала как выгрузка экспертизы (статус, эксперт, итерации…),
// иначе как таблица любого вида (номер, замечание, ответ, ссылка на раздел). Таблица, продолженная на
// следующих страницах, собирается в одну: шапка, повторённая на странице, пропускается, столбцы
// продолжения сопоставляются со столбцами первой страницы по положению. Если таблицы нет, а текст
// говорит о замечаниях, замечания читаются как нумерованные пункты («1.», «2.3)», «Замечание № 4»);
// абзац «Ответ: …» после пункта — ответ проектировщика. PDF без текстового слоя не разбирается.

// ParseRemarksFile reads an export of remarks: Excel (the expertise system's export or any table with
// a remark column) or PDF.
func ParseRemarksFile(data []byte, name, cacheDir string) ([]ExportRow, error) {
	if !IsPDF(data) && !strings.HasSuffix(strings.ToLower(name), ".pdf") {
		rows, err := ParseExport(data, name)
		if err == ErrNotExport {
			var generic []Expert
			if generic, err = ParseExpertise(data, name); err == nil {
				rows = ExpertToExport(generic)
			}
		}
		return rows, err
	}
	pages, err := ExtractPages("remarks", data, cacheDir)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать PDF: %w", err)
	}
	return ParseRemarksPDF(pages, name)
}

// ParseRemarksPDF reads the remarks from the pages of a PDF.
func ParseRemarksPDF(pages []Page, source string) ([]ExportRow, error) {
	words := 0
	for _, p := range pages {
		words += len(p.Lines)
	}
	if words == 0 {
		return nil, fmt.Errorf("в PDF нет текстового слоя: приложите выгрузку в Excel или PDF с текстом")
	}
	if rows := remarkTableRows(pages); len(rows) > 1 {
		if out, ok := parseExportRows(rows, source); ok {
			return out, nil
		}
		if out, ok := parseExpertRows(rows, source); ok {
			return ExpertToExport(out), nil
		}
	}
	if out := numberedRemarks(pages, source); len(out) > 0 {
		return ExpertToExport(out), nil
	}
	return nil, fmt.Errorf("в PDF не найдены замечания: нет таблицы со столбцом «Замечание» и нумерованных пунктов замечаний")
}

// remarkTableRows are the rows of the table of remarks (its header first), across pages.
func remarkTableRows(pages []Page) [][]string {
	var main *Table
	var rows [][]string
	for _, p := range pages {
		q := p
		q.Drawing, q.Cover = false, false // a landscape table may look like a sheet
		min := minRows
		if main != nil {
			min = 1
		}
		for _, t := range extractTablesMin("remarks", q, min) {
			t := t
			if main == nil {
				if !isRemarkHeader(t.Header) {
					continue
				}
				main = &t
				rows = append(rows, t.Header)
				for _, r := range t.Rows {
					rows = append(rows, r.Cells)
				}
				continue
			}
			// a continuation: the header repeated, or the first row of the page read as a header
			cont := t.Rows
			if !sameHeader(t.Header, main.Header) {
				cont = append([]TableRow{{Cells: t.Header, Box: t.Box}}, t.Rows...)
			}
			for _, r := range cont {
				rows = append(rows, remap(r.Cells, t.Starts, main.Starts))
			}
		}
	}
	return rows
}

func isRemarkHeader(h []string) bool {
	for _, c := range h {
		if n := norm(c); hdrText.MatchString(n) && len([]rune(n)) <= 60 {
			return true
		}
	}
	return false
}

func sameHeader(a, b []string) bool {
	return dice(headerTokens(a), headerTokens(b)) >= 0.8
}

// remap puts the cells of a continuation into the columns of the main table (nearest left edge).
func remap(cells []string, from, to []float64) []string {
	if len(from) == len(to) {
		return cells
	}
	out := make([]string, len(to))
	for i, c := range cells {
		if i >= len(from) || c == "" {
			continue
		}
		best, bd := 0, 2.0
		for k, x := range to {
			if d := abs(x - from[i]); d < bd {
				best, bd = k, d
			}
		}
		out[best] = strings.TrimSpace(out[best] + " " + c)
	}
	return out
}

var (
	remarkItemRe   = regexp.MustCompile(`^(?:замечание\s*№?\s*)?(\d{1,3}(?:\.\d{1,3}){0,2})[.)]?\s+(\S.*)$`)
	remarkAnswerRe = regexp.MustCompile(`^(?:ответ(?:\s+проектировщика|\s+проектной\s+организации)?|пояснение|исполнено|устранено)\s*[:.\-—]\s*(.*)$`)
	remarkWordRe   = regexp.MustCompile(`замечани`)
	remarkSecRe    = regexp.MustCompile(`^(?:раздел|подраздел|том|книга)\s+\S+`)
)

// numberedRemarks reads the remarks of a text without a table: numbered items with optional answers.
func numberedRemarks(pages []Page, source string) []Expert {
	var lines []Line
	mentions := false
	for _, p := range pages {
		for _, l := range p.Lines {
			if l.Stamp {
				continue
			}
			lines = append(lines, l)
			if remarkWordRe.MatchString(strings.ToLower(l.Text)) {
				mentions = true
			}
		}
	}
	if !mentions {
		return nil
	}
	var out []Expert
	section := ""
	inAnswer := false
	for _, l := range lines {
		t := strings.TrimSpace(l.Text)
		low := strings.ToLower(t)
		if m := remarkAnswerRe.FindStringSubmatch(low); m != nil && len(out) > 0 {
			inAnswer = true
			out[len(out)-1].Answer = strings.TrimSpace(out[len(out)-1].Answer + " " + strings.TrimSpace(t[len(t)-len(m[1]):]))
			continue
		}
		if m := remarkItemRe.FindStringSubmatch(t); m != nil && !looksLikeValue(m[2]) {
			inAnswer = false
			out = append(out, Expert{Number: m[1], Text: m[2], Source: source, Section: section})
			continue
		}
		if remarkSecRe.MatchString(low) && len([]rune(t)) < 120 {
			section, inAnswer = t, false
			continue
		}
		if len(out) == 0 {
			continue // the heading of the document
		}
		e := &out[len(out)-1]
		if inAnswer {
			e.Answer = joinWrapped(e.Answer, t)
		} else {
			e.Text = joinWrapped(e.Text, t)
		}
	}
	for i := range out {
		out[i].Where = volumeRefIn(out[i].Text)
	}
	return out
}

// looksLikeValue: «2,4 м.» after a number is a value, not the text of a remark.
func looksLikeValue(s string) bool {
	r := []rune(strings.TrimSpace(s))
	return len(r) < 12
}

var volRefInRe = regexp.MustCompile(`(?i)(?:том|шифр|раздел)\s*[№:]?\s*([0-9А-ЯЁA-Z][0-9А-ЯЁA-Z.\-/]*\d[0-9А-ЯЁA-Z.\-/]*)`)

// volumeRefIn is the volume a remark names in its text («том 2024-15-КЖ2»), for the links to the documents.
func volumeRefIn(s string) string {
	if m := volRefInRe.FindStringSubmatch(s); m != nil {
		return strings.TrimRight(m[1], ".-/")
	}
	return ""
}
