package cmp

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// Комплектность сверх томов (ТЗ 3.2.1, 3.2.4, 3.2.8). В ведомости состава кроме томов можно задать
// обязательные разделы (марка: «КЖ»), приложения (код «Приложение Б», при необходимости — в каком
// томе) и ссылочные документы. Раздел есть, если в комплекте есть том этой марки; приложение — если
// в томе (или в любом томе комплекта) есть заголовок этого приложения; ссылочный документ — если в
// комплекте есть том с этим шифром или наименованием. Отсутствие ссылочного документа учитывается
// в комплектности только когда строка ведомости помечена обязательной (п. 3.2.8).
// Ссылки на другие тома проекта (п. 3.2.4) агент читает из текста томов — шифры с тем же префиксом
// проекта, что и у самого тома («см. 2024-15-КЖ3»), — и показывает те, которых нет в комплектах
// объекта. Это сведения, а не несоответствие, пока ведомость не требует такого тома.

// Kinds of a row of the statement of volumes.
const (
	UnitVolume    = "volume"
	UnitSection   = "section"
	UnitAppendix  = "appendix"
	UnitReference = "reference"
)

// UnitLabels name the kinds.
var UnitLabels = map[string]string{UnitVolume: "том", UnitSection: "раздел", UnitAppendix: "приложение", UnitReference: "ссылочный документ"}

// VolRefAt is a reference of a volume to another volume of the project.
type VolRefAt struct {
	Code string `json:"code"`
	Page int    `json:"page"`
}

var refCodeTailRe = regexp.MustCompile(`^[А-ЯЁA-Z]{1,6}[0-9]*(?:[.\-][0-9]{1,3}){0,3}`)

// projectPrefix is the part of a volume code before its discipline mark: "2024-15-КЖ2" → "2024-15-".
func projectPrefix(code string) string {
	code = strings.TrimSpace(code)
	for i := len(code) - 1; i > 0; i-- {
		c := code[i-1]
		if c == '-' || c == '/' || c == '.' || c == ' ' {
			r := []rune(code[i:])
			if len(r) > 0 && unicode.IsLetter(r[0]) {
				if p := code[:i]; len([]rune(p)) >= 4 && strings.IndexFunc(p, unicode.IsDigit) >= 0 {
					return p
				}
			}
		}
	}
	return ""
}

// ExtractVolumeRefs finds in the text of a volume the codes of other volumes of the same project.
func ExtractVolumeRefs(pages []Page, ownCodes []string) []VolRefAt {
	own := map[string]bool{}
	prefixes := map[string]bool{}
	for _, c := range ownCodes {
		own[NormCode(c)] = true
		if p := projectPrefix(c); p != "" {
			prefixes[p] = true
		}
	}
	var out []VolRefAt
	seen := map[string]bool{}
	for pi, p := range pages {
		for _, l := range p.Lines {
			if l.Stamp {
				continue // the title block carries the volume's own code
			}
			for pre := range prefixes {
				text := l.Text
				for {
					i := strings.Index(text, pre)
					if i < 0 {
						break
					}
					tail := refCodeTailRe.FindString(text[i+len(pre):])
					text = text[i+len(pre):]
					if tail == "" {
						continue
					}
					code := pre + strings.TrimRight(tail, ".-")
					n := NormCode(code)
					if own[n] || seen[n] {
						continue
					}
					seen[n] = true
					out = append(out, VolRefAt{Code: code, Page: pi + 1})
				}
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out
}

// VolRefsJSON is the stored form (field volume_refs_json of a document); "[]" when there are none,
// so that a read document is not read again.
func VolRefsJSON(r []VolRefAt) string {
	if len(r) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(r)
	return string(b)
}

func parseVolRefs(s string) []VolRefAt {
	var r []VolRefAt
	_ = json.Unmarshal([]byte(s), &r)
	return r
}

// appendixID reads the letter or number of an appendix from a row of the statement: "Приложение Б" / "Б".
func appendixID(code string) string {
	c := strings.TrimSpace(code)
	if m := appendixHeadRe.FindStringSubmatch(c); m != nil {
		return foldID(m[1])
	}
	if r := []rune(c); len(r) <= 2 {
		return foldID(c)
	}
	return ""
}

// unitPresent says whether the set has a section, an appendix or a reference of the statement, and where.
func unitPresent(s *SetInfo, e expectedVolume) (bool, string) {
	switch e.kind {
	case UnitSection:
		want := Series(e.code)
		if want == "" {
			want = strings.ToUpper(strings.TrimSpace(e.code))
		}
		for _, d := range s.Docs {
			for _, sec := range strings.Split(firstNonEmpty(d.Section, SectionOf(strings.Join(d.Codes, ","))), ",") {
				if strings.EqualFold(strings.TrimSpace(sec), want) {
					return true, firstNonEmpty(strings.Join(d.Codes, ", "), d.ID)
				}
			}
		}
	case UnitAppendix:
		id := appendixID(e.code)
		for _, d := range s.Docs {
			if e.parent != "" && !hasVolume(d.Codes, e.parent) {
				continue
			}
			for _, h := range d.Headings {
				if h.Kind != MatAppendix {
					continue
				}
				if (id != "" && h.ID == id) || (e.title != "" && sharesWords(e.title, h.Text)) {
					return true, fmt.Sprintf("%s, стр. %d", firstNonEmpty(strings.Join(d.Codes, ", "), d.ID), h.Page)
				}
			}
		}
	case UnitReference:
		for _, d := range s.Docs {
			for _, c := range d.Codes {
				if volumesMatch(c, e.code) {
					return true, c
				}
			}
			if e.title != "" && d.Title != "" && sharesWords(e.title, d.Title) {
				return true, firstNonEmpty(strings.Join(d.Codes, ", "), d.ID)
			}
		}
	}
	return false, ""
}

// UnitState is a section, an appendix or a reference of the statement in one set.
type UnitState struct {
	Kind     string `json:"kind"`
	Code     string `json:"code"`
	Title    string `json:"title,omitempty"`
	Parent   string `json:"parent,omitempty"`
	Required bool   `json:"required"`
	Present  bool   `json:"present"`
	Where    string `json:"where,omitempty"`
}

// unitLabel is "приложение Б (КЖ2)".
func unitLabel(e expectedVolume) string {
	l := UnitLabels[e.kind] + " " + e.code
	if e.parent != "" {
		l += " (в " + e.parent + ")"
	}
	return l
}

// checkUnits fills the units of a set; a missing required unit counts against completeness.
func checkUnits(s *SetInfo, units []expectedVolume) {
	for _, e := range units {
		if e.stage != "" && e.stage != s.Stage {
			continue
		}
		ok, where := unitPresent(s, e)
		s.Units = append(s.Units, UnitState{Kind: e.kind, Code: e.code, Title: e.title, Parent: e.parent, Required: e.required, Present: ok, Where: where})
		if !e.required {
			continue
		}
		s.Total++
		if ok {
			s.Present++
		} else {
			s.Missing = append(s.Missing, unitLabel(e))
		}
	}
}

// missingRefs lists the volumes the documents of a set refer to that the object does not have.
func missingRefs(s *SetInfo, all []SetInfo) []string {
	var out []string
	for _, d := range s.Docs {
		for _, r := range d.Refs {
			have := false
			for _, o := range all {
				for _, od := range o.Docs {
					if hasVolume(od.Codes, r.Code) {
						have = true
					}
				}
			}
			if !have {
				out = append(out, fmt.Sprintf("%s (ссылка в %s, стр. %d)", r.Code, firstNonEmpty(strings.Join(d.Codes, ", "), d.ID), r.Page))
			}
		}
	}
	return out
}
