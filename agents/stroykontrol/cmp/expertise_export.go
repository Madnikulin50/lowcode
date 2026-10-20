package cmp

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// Workflow statuses of an expertise remark. The labels are the expertise
// system's own wording, so an export and the registry read the same.
const (
	RemarkAwaiting  = "awaiting_answer" // Ответ не предоставлен
	RemarkBeingFix  = "being_fixed"     // Устраняется заявителем
	RemarkAnswerRev = "answer_review"   // Требуется рассмотрение ответа
	RemarkResolved  = "resolved"        // Устранено
)

// RemarkStatusLabels maps a status code to its label.
var RemarkStatusLabels = map[string]string{
	RemarkAwaiting:  "Ответ не предоставлен",
	RemarkBeingFix:  "Устраняется заявителем",
	RemarkAnswerRev: "Требуется рассмотрение ответа",
	RemarkResolved:  "Устранено",
}

// StatusCode turns the status text of an export into a code ("" if unknown).
func StatusCode(src string) string {
	s := norm(src)
	for code, label := range RemarkStatusLabels {
		if s == norm(label) {
			return code
		}
	}
	return ""
}

// ExportRow is one remark of an expertise export («Выгрузка данных»).
type ExportRow struct {
	Key              string // identity inside the object: the number, made unique
	Number           string // as written, "30.1"
	Batch            string // the part before the dot: the letter the remark came with
	Text             string
	Refs             string // "Ссылка на материалы", raw
	Section          string
	Basis            string
	Answer           string
	RejectReason     string
	PrevAnswer       string
	PrevRejectReason string
	Discipline       string // "НД": МТ, ПОС, СМ, КР…
	Expert           string
	Criticality      string // critical | noncritical | ""
	Proposals        string
	Expected         string
	Iterations       int
	ChangedAt        time.Time // zero if absent
	SourceStatus     string    // the text of the status column
	Status           string    // its code
	Source           string    // file name
}

// ErrNotExport says the table is not in the export's format (the caller may
// fall back to the generic remark parser).
var ErrNotExport = errors.New("не формат выгрузки замечаний экспертизы")

// export column headings (normalized) → field
var exportColumns = map[string]string{
	"вывод о несоответствии":        "text",
	"ссылка на материалы":           "refs",
	"раздел документации":           "section",
	"основание":                     "basis",
	"ответ на замечание":            "answer",
	"причина отклонения":            "reject",
	"предыдущий ответ на замечание": "prevAnswer",
	"предыдущая причина отклонения": "prevReject",
	"нд": "discipline",
	"эксперт-соисполнитель":              "expert",
	"критичность":                        "criticality",
	"предложения по отработке замечаний": "proposals",
	"описание ожидаемого результата":     "expected",
	"итераций": "iterations",
	"дата последнего изменения": "changedAt",
	"статус":      "status",
	"№ замечания": "number",
}

var changedLayouts = []string{"02.01.2006 15:04:05", "02.01.2006 15:04", "02.01.2006", "2006-01-02 15:04:05", "2006-01-02"}

func parseChanged(s string) time.Time {
	s = strings.TrimSpace(s)
	for _, l := range changedLayouts {
		if t, err := time.Parse(l, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func shortHash(s string) string {
	h := sha1.Sum([]byte(norm(s)))
	return hex.EncodeToString(h[:])[:8]
}

// ParseExport reads an expertise export: a table whose headings are the ones
// of the expertise system (text, references, answer, previous answer, expert,
// criticality, iterations, status, number…). It returns ErrNotExport when the
// headings are not recognized.
func ParseExport(data []byte, source string) ([]ExportRow, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть Excel: %w", err)
	}
	defer f.Close()
	for _, sheet := range f.GetSheetList() {
		rows, err := f.GetRows(sheet)
		if err != nil {
			continue
		}
		if out, ok := parseExportRows(rows, source); ok {
			return out, nil
		}
	}
	return nil, ErrNotExport
}

func parseExportRows(rows [][]string, source string) ([]ExportRow, bool) {
	hdr := -1
	col := map[string]int{}
	for r := 0; r < len(rows) && r < 15; r++ {
		m := map[string]int{}
		for c, cell := range rows[r] {
			if f, ok := exportColumns[norm(cell)]; ok {
				if _, dup := m[f]; !dup {
					m[f] = c
				}
			}
		}
		_, hasText := m["text"]
		_, hasNum := m["number"]
		_, hasStatus := m["status"]
		if hasText && hasNum && hasStatus {
			hdr, col = r, m
			break
		}
	}
	if hdr < 0 {
		return nil, false
	}
	get := func(row []string, f string) string {
		c, ok := col[f]
		if !ok || c >= len(row) {
			return ""
		}
		return strings.TrimSpace(strings.ReplaceAll(row[c], " ", " "))
	}
	var out []ExportRow
	seen := map[string]int{}
	for _, row := range rows[hdr+1:] {
		text := get(row, "text")
		number := get(row, "number")
		if text == "" && number == "" {
			continue
		}
		e := ExportRow{
			Number: number, Text: text, Refs: get(row, "refs"), Section: get(row, "section"), Basis: get(row, "basis"),
			Answer: get(row, "answer"), RejectReason: get(row, "reject"), PrevAnswer: get(row, "prevAnswer"), PrevRejectReason: get(row, "prevReject"),
			Discipline: get(row, "discipline"), Expert: get(row, "expert"), Proposals: get(row, "proposals"), Expected: get(row, "expected"),
			ChangedAt: parseChanged(get(row, "changedAt")), SourceStatus: get(row, "status"), Source: source,
		}
		e.Status = StatusCode(e.SourceStatus)
		e.Iterations, _ = strconv.Atoi(get(row, "iterations"))
		switch norm(get(row, "criticality")) {
		case "критичное":
			e.Criticality = "critical"
		case "некритичное":
			e.Criticality = "noncritical"
		}
		key := number
		if key == "" {
			key = "без номера: " + shortHash(text)
		}
		seen[key]++
		if seen[key] > 1 {
			key = fmt.Sprintf("%s#%d", key, seen[key])
		}
		e.Key = key
		if i := strings.Index(number, "."); i > 0 {
			e.Batch = number[:i]
		} else {
			e.Batch = number
		}
		out = append(out, e)
	}
	return out, len(out) > 0
}

// VolumeRef is a volume (and place in it) a remark points to.
type VolumeRef struct {
	Tom   string `json:"tom,omitempty"`   // "4.5.1"
	Code  string `json:"code"`            // "ПЗ5.1"
	Sheet int    `json:"sheet,omitempty"` // лист
	Page  int    `json:"page,omitempty"`  // стр.
}

var (
	refSplitRe = regexp.MustCompile(`[;\n\r]+`)
	tomRe      = regexp.MustCompile(`(?i)том\s+(\d+(?:\.\d+)*)`)
	// the discipline mark after the project code: "… – ПЗ5.1", "…-ТКР5.4.1", "– ИОС"
	codeTailRe = regexp.MustCompile(`[–—-]\s*([А-ЯЁA-Z]{2,6}[\dА-ЯЁA-Z]*(?:\.\d+)*)\.?\s*$`)
	codeAnyRe  = regexp.MustCompile(`[–—-]\s*([А-ЯЁA-Z]{2,6}\d[\dА-ЯЁA-Z]*(?:\.\d+)*)`)
	sheetRe    = regexp.MustCompile(`(?i)лист\s*(\d+)`)
	pageRe2    = regexp.MustCompile(`(?i)стр\.?\s*(\d+)`)
)

// ParseVolumeRefs extracts the volumes named in a "Ссылка на материалы" cell:
// "Том 4.5.1, 28/12-22-ГП-599 – ПЗ5.1; Том 4.5.6, 28/12-22-ГП-599 – ТХ5.6" or
// "том 3.5.4.1, 28/12-22-ГП-599-ТКР5.4.1, раздел 2.3, лист 19 (стр. 25)".
func ParseVolumeRefs(s string) []VolumeRef {
	var out []VolumeRef
	seen := map[string]bool{}
	for _, part := range refSplitRe.Split(s, -1) {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		var v VolumeRef
		if m := tomRe.FindStringSubmatch(part); m != nil {
			v.Tom = m[1]
		}
		// the last discipline mark of the entry wins (the project code comes first)
		head := part
		if i := strings.IndexAny(head, ",;("); i > 0 && strings.Count(head, "–")+strings.Count(head, "-") > 0 {
			// keep everything up to the first place reference ("раздел", "лист", "стр.")
			if j := regexp.MustCompile(`(?i),\s*(раздел|лист|стр|приложение)`).FindStringIndex(head); j != nil {
				head = head[:j[0]]
			}
		}
		if m := codeTailRe.FindStringSubmatch(strings.TrimSpace(head)); m != nil {
			v.Code = m[1]
		} else if ms := codeAnyRe.FindAllStringSubmatch(head, -1); len(ms) > 0 {
			v.Code = ms[len(ms)-1][1]
		}
		if m := sheetRe.FindStringSubmatch(part); m != nil {
			v.Sheet, _ = strconv.Atoi(m[1])
		}
		if m := pageRe2.FindStringSubmatch(part); m != nil {
			v.Page, _ = strconv.Atoi(m[1])
		}
		v.Code = strings.TrimRight(v.Code, ".")
		if v.Code == "" && v.Tom == "" {
			continue
		}
		k := v.Tom + "|" + v.Code
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, v)
	}
	return out
}

// sameVolume reports whether a document's шифр ends with the code a remark
// names ("100720-1/П-3-ТКР5-1" ~ "ТКР5.1"): dots and dashes are one separator.
func sameVolume(docCode, refCode string) bool {
	a, b := NormCode(docCode), NormCode(refCode)
	if a == "" || b == "" {
		return false
	}
	// as written (a dot is part of the code: КЖ1 is not КЖ1.13): the code at the end of
	// the шифр, or inside it ("КЖ1.13-Расчеты", "…-ОП-КЖ1.13-С")
	if a == b || strings.HasSuffix(a, "-"+b) || strings.HasPrefix(a, b+"-") || strings.Contains(a, "-"+b+"-") {
		return true
	}
	// "ТКР5.1" is written "ТКР5-1" in a шифр: dots and dashes are one separator, at the end only
	a, b = strings.ReplaceAll(a, ".", "-"), strings.ReplaceAll(b, ".", "-")
	return a == b || strings.HasSuffix(a, "-"+b)
}

// MatchVolume tells whether a ref points into the volume (by its code or title).
func MatchVolume(ref VolumeRef, docCode, docTitle string) bool {
	if ref.Code != "" && sameVolume(docCode, ref.Code) {
		return true
	}
	return ref.Code != "" && strings.Contains(strings.ToLower(docTitle), strings.ToLower(ref.Code))
}
