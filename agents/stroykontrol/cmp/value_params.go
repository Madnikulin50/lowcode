package cmp

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Словарь значений сверх ключевых параметров (ТЗ 4.2.7, 3.2.5): толщина, глубина заложения, защитный
// слой, шаг, диаметр, высота, пролёт, отметки, нагрузки, уклон, сейсмичность, предел огнестойкости,
// уровень ответственности. Значение читается после названия параметра в той же строке и приводится к
// единице словаря («0,2 м» и «200 мм» — одно значение). Параметр относится к конструктивному элементу
// строки (или двух предыдущих строк), как ключевые параметры в crossvol.go; параметры объекта в целом
// (сейсмичность, снеговая нагрузка, абсолютная отметка нуля, уровень ответственности) элемента не
// требуют. Значения участвуют в сквозной сверке томов одной стороны и в сверке «чертёж ↔ текст»:
// значение, которое текст тома (пояснительная записка, общие указания) называет для элемента, должно
// встречаться и на чертежах той же стороны, если чертежи называют этот параметр для того же элемента.
// Сверка идёт по текстовому слою чертежей; несколько разных значений на чертежах (разные плиты одной
// марки) расхождением не считаются, пока среди них есть значение текста.

// ObjectWide is the element of the parameters of the object as a whole.
const ObjectWide = "объект в целом"

type valueKind struct {
	Name   string
	Re     *regexp.Regexp // group 1: the number (or the word), group 2: the unit, when there is one
	Unit   string         // the unit of the dictionary; "" — compared as read
	Scale  map[string]float64
	Bare   bool // a number without a unit is in the unit of the dictionary
	Global bool // a parameter of the object, not of an element
	Canon  func(m []string) string
}

const lb = `(?:^|[^\p{L}])` // the start of a word (\b does not work for Cyrillic)
const le = `(?:[^\p{L}]|$)`

const num = `([+\-−±]?\d+(?:[.,]\d+)?)`

var lengthMM = map[string]float64{"мм": 1, "см": 10, "м": 1000}
var lengthM = map[string]float64{"мм": 0.001, "см": 0.01, "м": 1}

var valueKinds = []valueKind{
	{Name: "Абсолютная отметка нуля", Global: true,
		Re:    regexp.MustCompile(`(?i)` + lb + `(?:0[.,]000|нул\p{L}*)\D{0,40}?абсолютн\p{L}*\s+отметк\p{L}*\s*` + num),
		Canon: func(m []string) string { return fmtNum(parseDec(m[1]), 2) }},
	{Name: "Отметка",
		Re:    regexp.MustCompile(`(?i)` + lb + `(?:отметк\p{L}*|отм\.)[^\d+\-−±]{0,30}?([+\-−±]?\d+[.,]\d{3})` + `(?:\D|$)`),
		Canon: func(m []string) string { return fmtSigned(parseDec(m[1]), 3) }},
	{Name: "Глубина заложения", Unit: "м", Scale: lengthM,
		Re: regexp.MustCompile(`(?i)` + lb + `глубин\p{L}*\s+заложени\p{L}*\D{0,30}?` + num + `\s*(мм|см|м)` + le)},
	{Name: "Защитный слой", Unit: "мм", Scale: lengthMM, Bare: true,
		Re: regexp.MustCompile(`(?i)` + lb + `защитн\p{L}*\s+сло\p{L}*\D{0,30}?` + num + `\s*(мм|см|м)?` + le)},
	{Name: "Толщина", Unit: "мм", Scale: lengthMM,
		Re: regexp.MustCompile(`(?i)` + lb + `толщин\p{L}*\D{0,30}?` + num + `\s*(мм|см|м)` + le)},
	{Name: "Шаг", Unit: "мм", Scale: lengthMM, Bare: true,
		Re: regexp.MustCompile(`(?i)` + lb + `шаг\p{L}*\D{0,20}?` + num + `\s*(мм|см|м)?` + le)},
	{Name: "Диаметр", Unit: "мм", Scale: lengthMM, Bare: true,
		Re: regexp.MustCompile(`(?i)(?:` + lb + `диаметр\p{L}*\D{0,20}?|[Ø⌀]\s?)` + num + `\s*(мм|см|м)?` + le)},
	{Name: "Высота", Unit: "мм", Scale: lengthMM,
		Re: regexp.MustCompile(`(?i)` + lb + `высот\p{L}*\D{0,30}?` + num + `\s*(мм|см|м)` + le)},
	{Name: "Пролёт", Unit: "м", Scale: lengthM,
		Re: regexp.MustCompile(`(?i)` + lb + `прол[её]т\p{L}*\D{0,30}?` + num + `\s*(мм|см|м)` + le)},
	{Name: "Снеговая нагрузка", Unit: "кПа", Scale: pressure, Global: true,
		Re: regexp.MustCompile(`(?i)` + lb + `снегов\p{L}*\s+(?:нагрузк\p{L}*|район\p{L}*)\D{0,30}?` + num + `\s*(кПа|кН/м2|кН/м²|кгс/м2|кгс/м²|т/м2|т/м²)`)},
	{Name: "Ветровое давление", Unit: "кПа", Scale: pressure, Global: true,
		Re: regexp.MustCompile(`(?i)` + lb + `ветров\p{L}*\s+(?:нагрузк\p{L}*|давлени\p{L}*)\D{0,30}?` + num + `\s*(кПа|кН/м2|кН/м²|кгс/м2|кгс/м²|т/м2|т/м²)`)},
	{Name: "Нагрузка", Unit: "кПа", Scale: pressure,
		Re: regexp.MustCompile(`(?i)` + lb + `(?:полезн\p{L}*|временн\p{L}*|нормативн\p{L}*|равномерно распредел\p{L}*)?\s*нагрузк\p{L}*\D{0,30}?` + num + `\s*(кПа|кН/м2|кН/м²|кгс/м2|кгс/м²|т/м2|т/м²)`)},
	{Name: "Уклон",
		Re:    regexp.MustCompile(`(?i)` + lb + `уклон\p{L}*\D{0,20}?` + num + `\s*(%|‰)`),
		Canon: func(m []string) string { return fmtNum(parseDec(m[1]), 3) + " " + m[2] }},
	{Name: "Сейсмичность", Global: true,
		Re:    regexp.MustCompile(`(?i)` + lb + `сейсмичност\p{L}*\D{0,40}?(\d{1,2})\s*балл`),
		Canon: func(m []string) string { return m[1] + " баллов" }},
	{Name: "Предел огнестойкости",
		Re: regexp.MustCompile(`(?:^|[^\p{L}])(REI|EI|RE|R)\s?(\d{2,3})` + le),
		Canon: func(m []string) string {
			if n, _ := strconv.Atoi(m[2]); n < 15 || n > 360 {
				return ""
			}
			return m[1] + " " + m[2]
		}},
	{Name: "Уровень ответственности", Global: true,
		Re:    regexp.MustCompile(`(?i)уровень\s+ответственности\D{0,20}?(повышенн|нормальн|пониженн)`),
		Canon: func(m []string) string { return strings.ToLower(m[1]) + "ый" }},
}

var pressure = map[string]float64{"кпа": 1, "кн/м2": 1, "кн/м²": 1, "кгс/м2": 0.00980665, "кгс/м²": 0.00980665, "т/м2": 9.80665, "т/м²": 9.80665}

func parseDec(s string) float64 {
	s = strings.NewReplacer(",", ".", "−", "-", "±", "", "+", "").Replace(strings.TrimSpace(s))
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// fmtNum: 0.2 → "0,2", 200 → "200" (at most prec decimals, no trailing zeros).
func fmtNum(v float64, prec int) string {
	s := strconv.FormatFloat(v, 'f', prec, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	return strings.Replace(s, ".", ",", 1)
}

// fmtSigned is an elevation: "-2,400", "+3,600", "0,000".
func fmtSigned(v float64, prec int) string {
	s := strings.Replace(strconv.FormatFloat(math.Abs(v), 'f', prec, 64), ".", ",", 1)
	switch {
	case v < 0:
		return "-" + s
	case v > 0:
		return "+" + s
	}
	return s
}

// valueHits reads the values of the dictionary in a line: parameter → canonical values.
func valueHits(text string) map[string][]string {
	out := map[string][]string{}
	masked := text
	for _, k := range valueKinds {
		for _, loc := range k.Re.FindAllStringSubmatchIndex(masked, -1) {
			m := make([]string, len(loc)/2)
			for g := range m {
				if loc[2*g] >= 0 {
					m[g] = masked[loc[2*g]:loc[2*g+1]]
				}
			}
			v := ""
			switch {
			case k.Canon != nil:
				v = k.Canon(m)
			case k.Scale != nil:
				u := strings.ToLower(m[2])
				if u == "" && !k.Bare {
					continue
				}
				sc := 1.0
				if u != "" {
					sc = k.Scale[u]
				}
				x := math.Abs(parseDec(m[1])) * sc
				if x == 0 {
					continue
				}
				v = fmtNum(x, 3) + " " + k.Unit
			}
			if v != "" {
				out[k.Name] = append(out[k.Name], v)
			}
		}
		// a span read by one parameter is not read again by the next ones («абсолютной отметке» ≠ «отметка»)
		masked = k.Re.ReplaceAllStringFunc(masked, func(s string) string { return strings.Repeat("#", len(s)) })
	}
	return out
}

// paramsOfLines collects, per (parameter, element), the values that lines name: the key parameters
// (params.go) and the values of the dictionary.
func paramsOfLines(lines []Line) map[[2]string]*paramSeen {
	out := map[[2]string]*paramSeen{}
	add := func(kind, el, val string, l Line) {
		key := [2]string{kind, el}
		ps := out[key]
		if ps == nil {
			ps = &paramSeen{vals: map[string]ParamHit{}}
			out[key] = ps
		}
		if _, ok := ps.vals[val]; !ok {
			ps.vals[val] = ParamHit{Value: val, Line: l}
		}
	}
	var ctx []string // the last lines of the current page, for the element of a line without one
	curPage := -1
	for _, l := range lines {
		if l.Page != curPage {
			curPage, ctx = l.Page, nil
		}
		el := elementOf(l.Text)
		if el == "" {
			for i := len(ctx) - 1; i >= 0 && el == ""; i-- {
				el = elementOf(ctx[i])
			}
		}
		ctx = append(ctx, l.Text)
		if len(ctx) > 2 {
			ctx = ctx[1:]
		}
		for name, vals := range valueHits(l.Text) {
			e := el
			if isGlobalValue(name) {
				e = ObjectWide
			}
			if e == "" {
				continue
			}
			for _, v := range vals {
				add(name, e, v, l)
			}
		}
		if el == "" {
			continue
		}
		for ki, kind := range ExtractParams([]Line{l}) {
			for val, hit := range kind {
				add(paramKinds[ki].Name, el, val, hit.Line)
			}
		}
	}
	return out
}

func isGlobalValue(name string) bool {
	for _, k := range valueKinds {
		if k.Name == name {
			return k.Global
		}
	}
	return false
}

// drawingLines are the lines of the text layer of the drawing sheets of a volume (title block excluded).
func drawingLines(d *Document, v *Volume) []Line {
	var out []Line
	for _, pn := range v.Drawings {
		if pn < 1 || pn > len(d.Pages) {
			continue
		}
		for _, l := range d.Pages[pn-1].Lines {
			if !l.Stamp {
				out = append(out, l)
			}
		}
	}
	return out
}

// drawingTextFindings compares the values the texts of one side name (explanatory notes, general
// instructions) with the values its drawings name for the same parameter and element.
func drawingTextFindings(docs []*Document, vols []*Volume, side string) []Finding {
	byID := map[string]*Document{}
	for _, d := range docs {
		byID[d.ID] = d
	}
	type at struct {
		v   *Volume
		hit ParamHit
	}
	text, draw := map[[2]string]*paramSeen{}, map[[2]string]*paramSeen{}
	textAt, drawAt := map[[2]string]at{}, map[[2]string]at{}
	merge := func(dst map[[2]string]*paramSeen, where map[[2]string]at, src map[[2]string]*paramSeen, v *Volume) {
		for k, ps := range src {
			if dst[k] == nil {
				dst[k] = &paramSeen{vals: map[string]ParamHit{}}
				first := ""
				for val := range ps.vals {
					if first == "" || val < first {
						first = val
					}
				}
				where[k] = at{v, ps.vals[first]}
			}
			for val, h := range ps.vals {
				if _, ok := dst[k].vals[val]; !ok {
					dst[k].vals[val] = h
				}
			}
		}
	}
	for _, v := range vols {
		d := byID[v.DocID]
		if d == nil {
			continue
		}
		merge(text, textAt, paramsOfLines(v.Lines), v)
		merge(draw, drawAt, paramsOfLines(drawingLines(d, v)), v)
	}
	var keys [][2]string
	for k := range text {
		if draw[k] != nil && disjoint(text[k], draw[k]) {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i][0] != keys[j][0] {
			return keys[i][0] < keys[j][0]
		}
		return keys[i][1] < keys[j][1]
	})
	var out []Finding
	for _, k := range keys {
		if len(out) >= maxCrossPerSide {
			break
		}
		ta, da := textAt[k], drawAt[k]
		tv, dv := valueList(text[k]), valueList(draw[k])
		desc := fmt.Sprintf("Чертёж ↔ текст: «%s» для «%s» — в тексте «%s» (стр. %d): %s; на чертежах «%s» (стр. %d): %s. Проверьте, что это разные элементы, а не расхождение.",
			k[0], k[1], shortVol(ta.v), ta.hit.Line.Page, strings.Join(tv, ", "), shortVol(da.v), da.hit.Line.Page, strings.Join(dv, ", "))
		f := Finding{
			Kind: KindCrossVolume, Type: "content", Severity: "low", Description: desc,
			Anchor:      trunc(da.hit.Line.Text, 200),
			Evidence:    Evidence{Ref: ta.hit.Line.Text, Target: da.hit.Line.Text, Rule: "drawing-text:" + k[0] + ":" + k[1]},
			Fingerprint: fingerprint(KindCrossVolume, side, "drawing-text", k[0], k[1], strings.Join(tv, ","), strings.Join(dv, ",")),
		}
		box := boxPtr(da.hit.Line.Box)
		if side == SideRef {
			f.RefDoc, f.PageRef, f.BoxRef, f.Evidence.VolumeRef = da.v.DocID, da.hit.Line.Page, box, volLabel(da.v)
		} else {
			f.TargetDoc, f.PageTarget, f.BoxTarget, f.Evidence.VolumeTgt = da.v.DocID, da.hit.Line.Page, box, volLabel(da.v)
		}
		out = append(out, f)
	}
	return out
}
