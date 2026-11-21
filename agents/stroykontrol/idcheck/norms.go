package idcheck

import (
	"regexp"
	"sort"
	"strings"
	"time"
)

// Norm is one regulatory document with its validity window.
type Norm struct {
	Code       string    `json:"code"`
	Title      string    `json:"title"`
	ValidFrom  time.Time `json:"valid_from"`
	ValidTo    time.Time `json:"valid_to"` // zero = in force
	ReplacedBy string    `json:"replaced_by"`
	WorkTags   []string  `json:"work_tags"` // work keywords that require this norm in п.6
}

// Requirement says a work type needs a norm cited in п.6.
type Requirement struct {
	Tag    string
	Prefix string
	Title  string
}

// NormCatalog is the lookup table for check 5. It starts from DefaultNorms
// and is overlaid with the namespace's editable `norm_docs` records.
type NormCatalog struct {
	byKey map[string]*Norm
}

func d(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

// DefaultNorms is a seed list of documents typical for bridge / road /
// concrete works. It is a starting point, not a legal reference: the
// namespace's «Нормативные документы» module is the editable source of
// truth and overrides these entries by code.
var DefaultNorms = []Norm{
	{Code: "СП 48.13330.2019", Title: "Организация строительства", ValidFrom: d("2020-06-25"), WorkTags: []string{"*"}},
	{Code: "СП 48.13330.2011", Title: "Организация строительства", ValidFrom: d("2011-05-20"), ValidTo: d("2020-06-25"), ReplacedBy: "СП 48.13330.2019"},
	{Code: "СНиП 12-01-2004", Title: "Организация строительства", ValidTo: d("2011-05-20"), ReplacedBy: "СП 48.13330.2011"},
	{Code: "СП 46.13330.2012", Title: "Мосты и трубы (актуализированная редакция СНиП 3.06.04-91)", ValidFrom: d("2013-07-01"), WorkTags: []string{"мост", "путепровод", "иссо", "опор", "пролетн", "насад", "ригел", "ростверк"}},
	{Code: "СНиП 3.06.04-91", Title: "Мосты и трубы", ValidTo: d("2013-07-01"), ReplacedBy: "СП 46.13330.2012"},
	{Code: "СП 70.13330.2012", Title: "Несущие и ограждающие конструкции (актуализированная редакция СНиП 3.03.01-87)", ValidFrom: d("2013-07-01"), WorkTags: []string{"бетон", "арматур", "армирован", "опалуб", "металлоконструк", "монтаж"}},
	{Code: "СНиП 3.03.01-87", Title: "Несущие и ограждающие конструкции", ValidTo: d("2013-07-01"), ReplacedBy: "СП 70.13330.2012"},
	{Code: "СП 45.13330.2017", Title: "Земляные сооружения, основания и фундаменты", ValidFrom: d("2017-08-28"), WorkTags: []string{"свай", "котлован", "обратн", "грунт", "земляны"}},
	{Code: "СП 45.13330.2012", Title: "Земляные сооружения, основания и фундаменты", ValidFrom: d("2013-01-01"), ValidTo: d("2017-08-28"), ReplacedBy: "СП 45.13330.2017"},
	{Code: "СП 78.13330.2012", Title: "Автомобильные дороги (актуализированная редакция СНиП 3.06.03-85)", ValidFrom: d("2013-07-01"), WorkTags: []string{"дорожн", "покрыт", "асфальт", "земляное полотно", "щебен"}},
	{Code: "СНиП 3.06.03-85", Title: "Автомобильные дороги", ValidTo: d("2013-07-01"), ReplacedBy: "СП 78.13330.2012"},
	{Code: "СП 35.13330.2011", Title: "Мосты и трубы (проектирование)", ValidFrom: d("2011-05-20")},
	{Code: "СП 34.13330.2021", Title: "Автомобильные дороги (проектирование)"},
	{Code: "СП 435.1325800.2018", Title: "Конструкции бетонные и железобетонные монолитные. Правила производства и приемки работ", ValidFrom: d("2019-06-14")},
	{Code: "СП 371.1325800.2017", Title: "Опалубка. Правила проектирования"},
	{Code: "СП 63.13330.2018", Title: "Бетонные и железобетонные конструкции", ValidFrom: d("2019-06-20")},
	{Code: "СП 16.13330.2017", Title: "Стальные конструкции"},
	{Code: "СП 28.13330.2017", Title: "Защита строительных конструкций от коррозии"},
	{Code: "СП 72.13330.2016", Title: "Защита строительных конструкций и сооружений от коррозии"},
	{Code: "СП 53-101-98", Title: "Изготовление и контроль качества стальных строительных конструкций"},
	{Code: "ГОСТ 18105-2018", Title: "Бетоны. Правила контроля и оценки прочности", ValidFrom: d("2019-01-01")},
	{Code: "ГОСТ 18105-2010", Title: "Бетоны. Правила контроля и оценки прочности", ValidTo: d("2019-01-01"), ReplacedBy: "ГОСТ 18105-2018"},
	{Code: "ГОСТ 10180-2012", Title: "Бетоны. Методы определения прочности по контрольным образцам", ValidFrom: d("2013-07-01")},
	{Code: "ГОСТ 22690-2015", Title: "Бетоны. Определение прочности механическими методами неразрушающего контроля", ValidFrom: d("2016-04-01")},
	{Code: "ГОСТ 22690-88", Title: "Бетоны. Определение прочности механическими методами неразрушающего контроля", ValidTo: d("2016-04-01"), ReplacedBy: "ГОСТ 22690-2015"},
	{Code: "ГОСТ 7473-2010", Title: "Смеси бетонные. Технические условия", ValidFrom: d("2012-01-01")},
	{Code: "ГОСТ 26633-2015", Title: "Бетоны тяжелые и мелкозернистые", ValidFrom: d("2016-09-01")},
	{Code: "ГОСТ 26633-2012", Title: "Бетоны тяжелые и мелкозернистые", ValidTo: d("2016-09-01"), ReplacedBy: "ГОСТ 26633-2015"},
	{Code: "ГОСТ 34028-2016", Title: "Прокат арматурный для железобетонных конструкций", ValidFrom: d("2018-01-01")},
	{Code: "ГОСТ 10922-2012", Title: "Арматурные и закладные изделия, их сварные, вязаные и механические соединения", ValidFrom: d("2013-07-01")},
	{Code: "ГОСТ 14098-2014", Title: "Соединения сварные арматуры и закладных изделий", ValidFrom: d("2015-07-01")},
	{Code: "РД-11-02-2006", Title: "Требования к составу и порядку ведения исполнительной документации", ValidTo: d("2023-09-01"), ReplacedBy: "Приказ Минстроя России от 16.05.2023 № 344/пр"},
	{Code: "РД 11-05-2007", Title: "Порядок ведения общего и (или) специального журнала", ValidTo: d("2023-09-01"), ReplacedBy: "Приказ Минстроя России от 02.12.2022 № 1026/пр"},
}

// NewNormCatalog builds a catalog from base entries plus overrides.
func NewNormCatalog(base []Norm, overrides ...Norm) *NormCatalog {
	c := &NormCatalog{byKey: map[string]*Norm{}}
	for _, list := range [][]Norm{base, overrides} {
		for i := range list {
			n := list[i]
			c.byKey[normKey(n.Code)] = &n
		}
	}
	return c
}

// Lookup finds a norm by code (spacing/look-alike insensitive).
func (c *NormCatalog) Lookup(code string) *Norm {
	return c.byKey[normKey(code)]
}

// All lists catalog entries sorted by code.
func (c *NormCatalog) All() []Norm {
	var out []Norm
	for _, n := range c.byKey {
		out = append(out, *n)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out
}

// Required returns norms that the given work description needs in п.6.
// Only norms currently in force (no ValidTo) can be required.
func (c *NormCatalog) Required(work string) []Requirement {
	work = strings.ToLower(work)
	var out []Requirement
	for _, n := range c.All() {
		if !n.ValidTo.IsZero() {
			continue
		}
		for _, t := range n.WorkTags {
			if t == "*" || (t != "" && strings.Contains(work, strings.ToLower(t))) {
				out = append(out, Requirement{Tag: tagLabel(t, work), Prefix: normPrefix(n.Code), Title: n.Title})
				break
			}
		}
	}
	return out
}

func tagLabel(tag, work string) string {
	if tag == "*" {
		return "любых"
	}
	return tag + "…"
}

// normPrefix drops the edition year: "СП 70.13330.2012" → "СП 70.13330",
// so a newer edition of the same set of rules also satisfies the
// requirement.
func normPrefix(code string) string {
	if m := regexp.MustCompile(`^(СП\s*\d+\.\d+)\.\d{4}$`).FindStringSubmatch(code); m != nil {
		return m[1]
	}
	if m := regexp.MustCompile(`^(ГОСТ(?:\s*Р)?\s*[\d.]+)-\d{2,4}$`).FindStringSubmatch(code); m != nil {
		return m[1]
	}
	return code
}

func normKey(code string) string {
	k := compact(code)
	k = strings.ReplaceAll(k, "СНИП", "СНИП")
	return strings.TrimRight(k, ".,;")
}

var normCodeRe = regexp.MustCompile(`(?i)(СП\s*\d{1,3}\.\d{5,7}(?:\.\d{4})?|СП\s*\d{2,3}-\d{3}-\d{2,4}|ГОСТ\s*(?:Р\s*)?(?:ИСО\s*)?[\d.]+-\d{2,4}|СНиП\s*[\d.]+-[\d.]+-?\d{0,4}|СТО\s*[\w.\-]+|РД[-\s]*\d+-\d+-\d{4}|ВСН\s*[\d\-]+)`)
var editionOfRe = regexp.MustCompile(`(?i)(редакци[яи]|взамен)\s*$`)

// ExtractNormCodes pulls regulatory document codes out of п.6 text. A
// СНиП code inside a СП title ("СП 46.13330.2012 Мосты и трубы.
// Актуализированная редакция СНиП 3.06.04-91") is part of the title, not
// a separate reference, and is skipped.
func ExtractNormCodes(text string) []string {
	var out []string
	seen := map[string]bool{}
	prevEnd, prevSP := 0, false
	for _, loc := range normCodeRe.FindAllStringIndex(text, -1) {
		code := strings.Join(strings.Fields(text[loc[0]:loc[1]]), " ")
		code = strings.TrimRight(code, ".,;")
		isSNiP := strings.HasPrefix(strings.ToUpper(code), "СНИП")
		between := text[prevEnd:loc[0]]
		inTitle := prevSP && isSNiP && !strings.ContainsAny(between, ";\n") && !strings.Contains(between, "\",") && !strings.Contains(between, "»,")
		prefix := strings.TrimSpace(text[max(0, loc[0]-40):loc[0]])
		if editionOfRe.MatchString(prefix) {
			inTitle = true
		}
		if !inTitle {
			if k := normKey(code); !seen[k] {
				seen[k] = true
				out = append(out, code)
			}
		}
		prevEnd = loc[1]
		if !isSNiP {
			prevSP = strings.HasPrefix(strings.ToUpper(code), "СП")
		}
	}
	return out
}

// isNormCode reports whether a п.6 line is a regulatory document rather
// than the project/working documentation reference.
func isNormCode(s string) bool {
	u := strings.ToUpper(strings.TrimSpace(s))
	for _, p := range []string{"СП", "ГОСТ", "СНИП", "СТО", "ВСН"} {
		if strings.HasPrefix(u, p) {
			return true
		}
	}
	return false
}
