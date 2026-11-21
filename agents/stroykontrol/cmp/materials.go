package cmp

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// Материалы, заявленные в ответе проектировщика (ТЗ 3.4.9, 3.4.10): тома, листы, приложения,
// расчёты, пояснительные записки, письма и справки, на которые ссылается ответ, ищутся среди
// загруженных документов объекта и обосновывающих материалов (файлы замечания и пакетов исправления).
// Проверяется только наличие: есть ли такой том, лист, приложение или документ; правильность
// содержания оценивает специалист. Разбор детерминированный, по шаблонам, без языковой модели.

// Kinds of a declared material.
const (
	MatVolume   = "volume"   // a volume (and its sheets or pages) the answer names
	MatAppendix = "appendix" // «Приложение Б»
	MatDocument = "document" // a calculation, an explanatory note, a letter, a certificate…
)

// States of a declared material.
const (
	MatFound      = "found"       // the material is there (in the document or among the supporting files)
	MatMissing    = "missing"     // the named volume is loaded, but the sheet, page or appendix is not in it
	MatNotFound   = "not_found"   // no loaded document or supporting file holds it
	MatNoDocument = "no_document" // the named volume is not loaded: nothing to look in
)

// MatStateLabels are the names of the states.
var MatStateLabels = map[string]string{
	MatFound: "есть", MatMissing: "нет в документе", MatNotFound: "не найдено", MatNoDocument: "том не загружен",
}

// Summary states of the materials of a remark (field materials_state).
const (
	MaterialsOK       = "ok"       // every declared material was found
	MaterialsMissing  = "missing"  // something declared is not there
	MaterialsUnloaded = "unloaded" // nothing is missing, but some named volumes are not loaded
)

// MaterialsStateLabels are the names of the summary states.
var MaterialsStateLabels = map[string]string{
	MaterialsOK: "все найдены", MaterialsMissing: "не найдены", MaterialsUnloaded: "тома не загружены",
}

// Material is one thing the answer says it provides, and what the check found.
type Material struct {
	Kind   string `json:"kind"`
	Label  string `json:"label"`            // as written in the answer
	Code   string `json:"code,omitempty"`   // the volume it is in («ТКР5.1»), when the answer names one
	Tom    string `json:"tom,omitempty"`    // «5.4.1»
	Sheets []int  `json:"sheets,omitempty"` // листы
	Page   int    `json:"page,omitempty"`   // стр.
	ID     string `json:"id,omitempty"`     // the appendix letter or number
	Doc    string `json:"doc,omitempty"`    // document kind: «расчёт», «пояснительная записка», …

	State string `json:"state"`
	DocID string `json:"docID,omitempty"` // where it was found
	In    string `json:"in,omitempty"`    // the name of that document or file
	At    int    `json:"at,omitempty"`    // the page it was found on
	Note  string `json:"note,omitempty"`
}

// DocHeading is a heading of a document a declared material may be: an appendix, a calculation, a note.
type DocHeading struct {
	Kind string `json:"k"`            // MatAppendix | MatDocument
	ID   string `json:"id,omitempty"` // appendix letter / number
	Doc  string `json:"d,omitempty"`  // document kind
	Text string `json:"t"`
	Page int    `json:"p"`
}

// MatDoc is a document the materials are looked for in: a loaded volume or a supporting file.
type MatDoc struct {
	ID, Name, Codes, Title string
	Pages                  int
	Sheets                 map[int]int // PDF page → sheet number ("" — not read)
	Headings               []DocHeading
	Supporting             bool // a file attached to the remark or a correction package, not a volume of a comparison
}

const (
	maxHeadings         = 300
	maxMaterials        = 20
	maxHeadingLen       = 140
	maxMaterialLabel    = 120
	maxSheetsPerMention = 40
)

// the kinds of documents an answer refers to, by the stem they start with
var docKinds = []struct{ stem, name string }{
	{"расч", "расчёт"},
	{"пояснительн", "пояснительная записка"},
	{"справк", "справка"},
	{"письм", "письмо"},
	{"протокол", "протокол"},
	{"технические услови", "технические условия"},
	{"заключени", "заключение"},
	{"обосновани", "обоснование"},
	{"согласовани", "согласование"},
	{"сертификат", "сертификат"},
	{"паспорт", "паспорт"},
}

var (
	matHeadRe = regexp.MustCompile(`^(?i)(расч[её]т|пояснительная записка|справка|письмо|протокол|технические условия|заключение|обоснование|согласование|сертификат|паспорт)(?:[^\p{L}]|$)`)
	matDocRe  = regexp.MustCompile(`(?i)(расч[её]т[а-яё]*|пояснительн[а-яё]*\s+записк[а-яё]*|справк[а-яё]*|письм[а-яё]*|протокол[а-яё]*|технически[а-яё]*\s+услови[а-яё]*|заключени[а-яё]*|обосновани[а-яё]*|согласовани[а-яё]*|сертификат[а-яё]*|паспорт[а-яё]*)`)
	// a discipline code of a volume: «ТКР5.4.1», «КЖ1», «ПЗ5-1» (two to five capitals and a number)
	matCodeRe = regexp.MustCompile(`(?:^|[^\p{L}\p{N}])([А-ЯЁ]{2,5}\d{1,2}(?:[.\-]\d{1,2}){0,3})(?:[^\p{L}\p{N}]|$)`)
	// "в том АР", "в томе ИОС": a volume named by its discipline mark only
	matTomCodeRe = regexp.MustCompile(`(?i:(?:^|[^\p{L}])том[а-яё]{0,2})\s+([А-ЯЁ]{2,5})(?:[^\p{L}\p{N}]|$)`)
	matTomRe     = regexp.MustCompile(`(?i)(?:^|[^\p{L}])том[а-яё]{0,2}\s+(\d{1,2}(?:\.\d{1,2}){0,3})`)
	matSheetRe   = regexp.MustCompile(`(?i)(?:^|[^\p{L}])(?:лист[а-яё]{0,3}|л\.)\s*((?:\d{1,3}\s*(?:[-–—,]|и)\s*)*\d{1,3})`)
	matPageRe    = regexp.MustCompile(`(?i)(?:^|[^\p{L}])стр[а-яё.]*\s*(\d{1,4})`)
	// "Приложение Б", "приложении 3"; the letter must be a capital one
	matAppRe = regexp.MustCompile(`(?i:приложени[а-яё]{0,2})\s+([А-ЯA-Z]|\d{1,2})(?:[^\p{L}\p{N}]|$)`)
	// the clause says the material is not there or not needed: nothing is declared by it
	matNegRe = regexp.MustCompile(`(?i)не\s+требу|не\s+предусм|не\s+представля|не\s+разрабатыва|отсутству|не\s+нужн|не\s+выполня`)
	// words of norms and contracts: an appendix in such a clause belongs to them, not to the documentation
	matOutsideRe = regexp.MustCompile(`(?i)гост|снип|(?:^|[^\p{L}])сп(?:[^\p{L}]|$)|договор|контракт|задани|постановлени|приказ|регламент|\d{1,3}\.1\d{4,6}`)
	matSplitRe   = regexp.MustCompile(`[;\n]+`)
	// codes that are not volumes: norms, classes
	notVolumeCodes = map[string]bool{"СП": true, "ГОСТ": true, "СНИП": true, "ТУ": true, "СТО": true, "ПУЭ": true, "РД": true, "ВСН": true, "МДС": true, "СН": true}
)

// docKindOf names the kind of document a word starts ("расчётом" → "расчёт").
func docKindOf(s string) string {
	s = strings.ToLower(strings.ReplaceAll(s, "ё", "е"))
	for _, k := range docKinds {
		if strings.HasPrefix(s, strings.ReplaceAll(k.stem, "ё", "е")) {
			return k.name
		}
	}
	return ""
}

// ExtractHeadings reads the headings of a document a declared material may be: appendices («Приложение Б»)
// and documents set in it (a calculation, a note, a letter…) at the start of a line. Contents lines (ending
// with a page number), title blocks and the cover are skipped.
func ExtractHeadings(d *Document) []DocHeading {
	var out []DocHeading
	seen := map[string]bool{}
	for pi, p := range d.Pages {
		if p.Cover {
			continue
		}
		for _, l := range p.Lines {
			if l.Stamp || len(out) >= maxHeadings {
				continue
			}
			text := strings.TrimSpace(l.Text)
			if text == "" || len([]rune(text)) > maxHeadingLen {
				continue
			}
			h := DocHeading{Page: pi + 1, Text: text}
			if m := appendixHeadRe.FindStringSubmatchIndex(text); m != nil {
				if contentsLike(text, m[3]) {
					continue
				}
				h.Kind, h.ID = MatAppendix, foldID(text[m[2]:m[3]])
			} else if m := matHeadRe.FindStringSubmatchIndex(text); m != nil {
				if contentsLike(text, m[3]) {
					continue
				}
				h.Kind, h.Doc = MatDocument, docKindOf(text[m[2]:m[3]])
			} else {
				continue
			}
			k := h.Kind + "|" + h.ID + "|" + strings.ToLower(h.Text)
			if seen[k] {
				continue
			}
			seen[k] = true
			out = append(out, h)
		}
	}
	return out
}

// HeadingsJSON is the stored form of the headings (field headings_json of a document).
func HeadingsJSON(h []DocHeading) string {
	if len(h) == 0 {
		return ""
	}
	b, _ := json.Marshal(h)
	return string(b)
}

// ParseHeadings reads the stored form.
func ParseHeadings(s string) []DocHeading {
	var h []DocHeading
	_ = json.Unmarshal([]byte(s), &h)
	return h
}

// expandSheets reads "12, 14-16 и 20" into the sheet numbers.
func expandSheets(s string) []int {
	var out []int
	s = strings.NewReplacer("–", "-", "—", "-", " и ", ",").Replace(s)
	for _, part := range strings.Split(s, ",") {
		a, b, rng := strings.Cut(strings.TrimSpace(part), "-")
		x, err := strconv.Atoi(strings.TrimSpace(a))
		if err != nil {
			continue
		}
		y := x
		if rng {
			if v, err := strconv.Atoi(strings.TrimSpace(b)); err == nil && v >= x && v-x < maxSheetsPerMention {
				y = v
			}
		}
		for n := x; n <= y && len(out) < maxSheetsPerMention; n++ {
			out = append(out, n)
		}
	}
	return out
}

func matClauses(answer string) []string {
	var out []string
	for _, part := range matSplitRe.Split(answer, -1) {
		for _, sn := range splitSentences(part) {
			if sn = strings.TrimSpace(sn); sn != "" {
				out = append(out, sn)
			}
		}
	}
	return out
}

func volumeCodes(clause string) []string {
	var out []string
	for _, m := range matCodeRe.FindAllStringSubmatch(clause, -1) {
		code := strings.TrimRight(m[1], ".-")
		letters := code
		if i := strings.IndexFunc(code, unicode.IsDigit); i > 0 {
			letters = code[:i]
		}
		if notVolumeCodes[letters] {
			continue
		}
		out = append(out, code)
	}
	return out
}

func clip(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > maxMaterialLabel {
		return string(r[:maxMaterialLabel]) + "…"
	}
	return s
}

// ParseMaterials reads the materials an answer of the designer says it provides: the volumes (with their
// sheets and pages), the appendices and the documents (calculations, notes, letters…). A clause that says a
// material is not needed or not there declares nothing.
func ParseMaterials(answer string) []Material {
	var out []Material
	seen := map[string]bool{}
	add := func(m Material) {
		k := strings.Join([]string{m.Kind, m.Code, m.Tom, fmt.Sprint(m.Sheets), strconv.Itoa(m.Page), m.ID, m.Doc}, "|")
		if seen[k] || len(out) >= maxMaterials {
			return
		}
		seen[k] = true
		out = append(out, m)
	}
	for _, cl := range matClauses(answer) {
		if matNegRe.MatchString(cl) {
			continue
		}
		codes := volumeCodes(cl)
		if len(codes) == 0 {
			if m := matTomCodeRe.FindStringSubmatch(cl); m != nil && !notVolumeCodes[m[1]] {
				codes = []string{m[1]}
			}
		}
		tom := ""
		if m := matTomRe.FindStringSubmatch(cl); m != nil {
			tom = m[1]
		}
		code := ""
		if len(codes) > 0 {
			code = codes[0]
		}
		var sheets []int
		for _, m := range matSheetRe.FindAllStringSubmatch(cl, -1) {
			sheets = append(sheets, expandSheets(m[1])...)
		}
		page := 0
		if m := matPageRe.FindStringSubmatch(cl); m != nil {
			page, _ = strconv.Atoi(m[1])
		}
		// documents and appendices of the clause belong to its volume, if it names one
		var inner []Material
		if !matOutsideRe.MatchString(cl) {
			for _, m := range matAppRe.FindAllStringSubmatch(cl, -1) {
				id := foldID(m[1])
				if r := []rune(m[1]); unicode.IsLetter(r[0]) && !unicode.IsUpper(r[0]) {
					continue
				}
				inner = append(inner, Material{Kind: MatAppendix, Label: "Приложение " + id, ID: id, Code: code, Tom: tom})
			}
		}
		for _, loc := range matDocRe.FindAllStringIndex(cl, -1) {
			kind := docKindOf(cl[loc[0]:loc[1]])
			if kind == "" || kind == "согласование" && len(inner) > 0 {
				continue
			}
			// the label: the document word and what follows it up to the next punctuation
			rest := cl[loc[0]:]
			if i := strings.IndexAny(rest, ",;:()«»\""); i > 0 {
				rest = rest[:i]
			}
			inner = append(inner, Material{Kind: MatDocument, Label: clip(rest), Doc: kind, Code: code, Tom: tom})
		}
		for _, c := range codes {
			v := Material{Kind: MatVolume, Label: c, Code: c, Tom: tom}
			if c == code {
				v.Sheets, v.Page = sheets, page
			}
			if tom != "" && c == code {
				v.Label = "том " + tom + " (" + c + ")"
			}
			add(v)
		}
		if len(codes) == 0 && tom != "" {
			add(Material{Kind: MatVolume, Label: "том " + tom, Tom: tom, Sheets: sheets, Page: page})
		}
		for _, m := range inner {
			add(m)
		}
	}
	return out
}

// docsOfVolume lists the documents a material that names a volume can be in.
func docsOfVolume(m Material, docs []MatDoc) []MatDoc {
	var out []MatDoc
	for _, d := range docs {
		if d.Supporting {
			continue
		}
		hit := false
		if m.Code != "" {
			ref := VolumeRef{Code: m.Code}
			codes := docCodes(d.Codes)
			hit = len(codes) == 0 && MatchVolume(ref, "", d.Title)
			for _, c := range codes {
				hit = hit || MatchVolume(ref, c, d.Title)
			}
		} else if m.Tom != "" {
			low := strings.ToLower(d.Title + " " + d.Name)
			hit = strings.Contains(low, "том "+m.Tom+" ") || strings.HasSuffix(low, "том "+m.Tom)
		}
		if hit {
			out = append(out, d)
		}
	}
	return out
}

func sharesWords(label, text string) bool {
	lt := remarkTokens(label)
	tt := remarkTokens(text)
	n := 0
	for k := range lt {
		if _, ok := tt[k]; ok {
			n++
		}
	}
	return n >= 2 || n >= 1 && len(lt) <= 2
}

// CheckMaterials looks for every declared material in the documents: a volume must be loaded and have the
// named sheets and pages; an appendix or a document must have its heading in the named volume (any loaded
// document or supporting file when the answer names none), or be a supporting file named so.
func CheckMaterials(ms []Material, docs []MatDoc) []Material {
	out := make([]Material, 0, len(ms))
	for _, m := range ms {
		m.State, m.DocID, m.In, m.At, m.Note = "", "", "", 0, ""
		scoped := m.Code != "" || m.Tom != ""
		var pool []MatDoc
		if scoped {
			pool = docsOfVolume(m, docs)
		}
		switch m.Kind {
		case MatVolume:
			checkVolume(&m, pool)
		default:
			if scoped && len(pool) == 0 {
				// the volume is not loaded, but the material may have come as a separate file
				if d, h, ok := findHeading(m, supportingOnly(docs)); ok {
					m.State, m.DocID, m.In, m.At = MatFound, d.ID, d.Name, h
					break
				}
				m.State, m.Note = MatNoDocument, "том "+firstNonEmpty(m.Code, m.Tom)+" не загружен"
				break
			}
			if !scoped {
				pool = docs
			} else {
				pool = append(pool, supportingOnly(docs)...)
			}
			if d, h, ok := findHeading(m, pool); ok {
				m.State, m.DocID, m.In, m.At = MatFound, d.ID, d.Name, h
				break
			}
			if scoped {
				what := "приложения " + m.ID
				if m.Kind == MatDocument {
					what = "документа «" + m.Doc + "»"
				}
				m.State, m.Note = MatMissing, fmt.Sprintf("в томе %s нет заголовка %s", firstNonEmpty(m.Code, m.Tom), what)
			} else if len(docs) == 0 {
				m.State, m.Note = MatNoDocument, "документы не загружены"
			} else {
				m.State, m.Note = MatNotFound, "нет ни в загруженных томах, ни среди обосновывающих материалов"
			}
		}
		out = append(out, m)
	}
	return out
}

func supportingOnly(docs []MatDoc) []MatDoc {
	var out []MatDoc
	for _, d := range docs {
		if d.Supporting {
			out = append(out, d)
		}
	}
	return out
}

func checkVolume(m *Material, pool []MatDoc) {
	if len(pool) == 0 {
		m.State, m.Note = MatNoDocument, "том не загружен"
		return
	}
	d := pool[0]
	m.State, m.DocID, m.In = MatFound, d.ID, d.Name
	if m.Page > 0 {
		if d.Pages > 0 && m.Page > d.Pages {
			m.State, m.Note = MatMissing, fmt.Sprintf("в томе %d стр., а ответ называет стр. %d", d.Pages, m.Page)
			return
		}
		m.At = m.Page
	}
	if len(m.Sheets) == 0 {
		return
	}
	var sheets map[int]int
	for _, x := range pool { // the sheets may be in any edition of the volume
		if len(x.Sheets) > 0 {
			sheets, d = x.Sheets, x
			break
		}
	}
	if sheets == nil {
		m.Note = "номера листов тома не прочитаны: листы не проверены"
		return
	}
	var absent []string
	for _, s := range m.Sheets {
		pg := pageOfSheet(sheets, s)
		if pg == 0 {
			absent = append(absent, strconv.Itoa(s))
		} else if m.At == 0 {
			m.At = pg
		}
	}
	m.DocID, m.In = d.ID, d.Name
	if len(absent) > 0 {
		m.State, m.Note = MatMissing, "в основных надписях тома нет листов "+strings.Join(absent, ", ")
	}
}

// findHeading finds an appendix or a document among the headings of the pool, or a supporting file whose
// name says it is that material.
func findHeading(m Material, pool []MatDoc) (MatDoc, int, bool) {
	for _, d := range pool {
		for _, h := range d.Headings {
			if h.Kind != m.Kind {
				continue
			}
			if m.Kind == MatAppendix && h.ID == m.ID {
				return d, h.Page, true
			}
			if m.Kind == MatDocument && h.Doc == m.Doc && sharesWords(m.Label, h.Text) {
				return d, h.Page, true
			}
		}
	}
	for _, d := range pool {
		if !d.Supporting {
			continue
		}
		name := strings.ToLower(d.Name)
		switch m.Kind {
		case MatAppendix:
			if strings.Contains(name, "прил") && strings.Contains(" "+strings.ToUpper(name)+" ", " "+m.ID+" ") || strings.Contains(strings.ToUpper(name), "ПРИЛОЖЕНИЕ "+m.ID) || strings.Contains(strings.ToUpper(name), "ПРИЛОЖЕНИЕ_"+m.ID) {
				return d, 0, true
			}
		case MatDocument:
			if docKindIn(name) == m.Doc {
				return d, 0, true
			}
		}
	}
	return MatDoc{}, 0, false
}

func docKindIn(s string) string {
	for _, w := range strings.FieldsFunc(s, func(r rune) bool { return !unicode.IsLetter(r) }) {
		if k := docKindOf(w); k != "" {
			return k
		}
	}
	return ""
}

// MaterialsSummary is the summary state of the materials of a remark: "" when the answer declares none.
func MaterialsSummary(ms []Material) string {
	if len(ms) == 0 {
		return ""
	}
	unloaded := false
	for _, m := range ms {
		switch m.State {
		case MatMissing, MatNotFound:
			return MaterialsMissing
		case MatNoDocument:
			unloaded = true
		}
	}
	if unloaded {
		return MaterialsUnloaded
	}
	return MaterialsOK
}

// MaterialsJSON is the stored form of the materials (field materials_json).
func MaterialsJSON(ms []Material) string {
	if len(ms) == 0 {
		return ""
	}
	b, _ := json.Marshal(ms)
	return string(b)
}

// ParseMaterialsJSON reads the stored form.
func ParseMaterialsJSON(s string) []Material {
	var ms []Material
	_ = json.Unmarshal([]byte(s), &ms)
	return ms
}

// MaterialsNote is the one line the registry shows: what was not found.
func MaterialsNote(ms []Material) string {
	var miss []string
	found := 0
	for _, m := range ms {
		switch m.State {
		case MatFound:
			found++
		default:
			miss = append(miss, m.Label+" — "+MatStateLabels[m.State])
		}
	}
	sort.Strings(miss)
	if len(miss) == 0 {
		return fmt.Sprintf("заявлено %d, все найдены", len(ms))
	}
	return fmt.Sprintf("заявлено %d, найдено %d; %s", len(ms), found, strings.Join(miss, "; "))
}

// WithMaterials applies the check of the materials to the outcome of a remark (ТЗ 3.4.10): a remark whose
// answer declares a material that is not there is not called fixed by the system — the specialist looks.
func WithMaterials(outcome, state string) string {
	if state == MaterialsMissing && outcome == OutcomeFixed {
		return OutcomeCheck
	}
	return outcome
}
