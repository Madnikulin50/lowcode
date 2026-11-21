package cmp

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// Сквозная согласованность между томами одной стороны: один и тот же параметр (класс бетона,
// водонепроницаемость, морозостойкость, класс арматуры и значения словаря value_params.go — толщина,
// отметки, защитный слой, нагрузки…) для одного и того же конструктивного элемента
// не должен называться по-разному в разных томах. Параметр привязывается к элементу по словарю
// конструктивных слов в той же строке или в двух предыдущих; параметры без элемента в находки не
// попадают (только в таблицу параметров), чтобы не плодить шум. Все находки — «к проверке».

var elementStems = []struct{ stem, name string }{
	{"фундамент", "фундамент"}, {"ростверк", "ростверк"}, {"свай", "свая"}, {"свайн", "свая"},
	{"стен", "стена"}, {"плит", "плита"}, {"колонн", "колонна"}, {"балк", "балка"}, {"ригел", "ригель"},
	{"перекрыт", "перекрытие"}, {"диафрагм", "диафрагма"}, {"опор", "опора"}, {"лестниц", "лестница"},
	{"подготовк", "подготовка"}, {"стяжк", "стяжка"}, {"приямк", "приямок"}, {"пилон", "пилон"}, {"перемыч", "перемычка"},
	{"кровл", "кровля"}, {"отмостк", "отмостка"}, {"перегородк", "перегородка"}, {"кладк", "кладка"}, {"покрыт", "покрытие"},
}

func elementOf(text string) string {
	for _, w := range strings.FieldsFunc(strings.ToLower(text), func(r rune) bool { return !unicode.IsLetter(r) }) {
		for _, e := range elementStems {
			if strings.HasPrefix(w, e.stem) {
				return e.name
			}
		}
	}
	return ""
}

// ParamRow is a parameter value met for an element in a volume (the table of parameters).
type ParamRow struct {
	Side    string   `json:"side"`
	Volume  string   `json:"volume"`
	Kind    string   `json:"kind"`
	Element string   `json:"element"`
	Values  []string `json:"values"`
	Page    int      `json:"page"`
}

type paramSeen struct {
	vals map[string]ParamHit
}

// volumeParams collects, per (parameter kind, element), the values a volume names (value_params.go).
func volumeParams(v *Volume) map[[2]string]*paramSeen { return paramsOfLines(v.Lines) }

func valueList(ps *paramSeen) []string {
	var vs []string
	for v := range ps.vals {
		vs = append(vs, v)
	}
	sort.Strings(vs)
	return vs
}

func disjoint(a, b *paramSeen) bool {
	for v := range a.vals {
		if _, ok := b.vals[v]; ok {
			return false
		}
	}
	return true
}

const maxCrossPerSide = 20

// crossVolume checks the volumes of one side against each other and returns the findings and the
// table of parameters. Needs at least two volumes with parameters for the same element.
func crossVolume(docs []*Document, vols []*Volume, side string) ([]Finding, []ParamRow) {
	if len(vols) < 2 {
		return nil, nil
	}
	byID := map[string]*Document{}
	for _, d := range docs {
		byID[d.ID] = d
	}
	type vp struct {
		v    *Volume
		seen map[[2]string]*paramSeen
	}
	var all []vp
	var rows []ParamRow
	for _, v := range vols {
		seen := volumeParams(v)
		all = append(all, vp{v, seen})
		var keys [][2]string
		for k := range seen {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			if keys[i][0] != keys[j][0] {
				return keys[i][0] < keys[j][0]
			}
			return keys[i][1] < keys[j][1]
		})
		for _, k := range keys {
			row := ParamRow{Side: side, Volume: volLabel(v), Kind: k[0], Element: k[1], Values: valueList(seen[k])}
			for _, h := range seen[k].vals {
				row.Page = h.Line.Page
				break
			}
			rows = append(rows, row)
		}
	}
	// conflicts of one parameter/element against the same baseline volume are one finding
	type conflict struct {
		cur  *Volume
		vals []string
		hit  ParamHit
	}
	type group struct {
		key       [2]string
		base      *Volume
		baseVals  []string
		baseHit   ParamHit
		conflicts []conflict
	}
	var groups []*group
	index := map[string]*group{}
	for i := 1; i < len(all); i++ {
		cur := all[i]
		var keys [][2]string
		for k := range cur.seen {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(a, b int) bool {
			if keys[a][0] != keys[b][0] {
				return keys[a][0] < keys[b][0]
			}
			return keys[a][1] < keys[b][1]
		})
		for _, k := range keys {
			// the first earlier volume that names this parameter for this element is the baseline
			var base *vp
			for j := 0; j < i; j++ {
				if _, ok := all[j].seen[k]; ok {
					base = &all[j]
					break
				}
			}
			if base == nil || !disjoint(base.seen[k], cur.seen[k]) {
				continue
			}
			sig := fmt.Sprintf("%s|%s|%s", k[0], k[1], base.v.Code)
			g := index[sig]
			if g == nil {
				g = &group{key: k, base: base.v, baseVals: valueList(base.seen[k])}
				for _, h := range base.seen[k].vals {
					g.baseHit = h
					break
				}
				index[sig] = g
				groups = append(groups, g)
			}
			c := conflict{cur: cur.v, vals: valueList(cur.seen[k])}
			for _, h := range cur.seen[k].vals {
				c.hit = h
				break
			}
			g.conflicts = append(g.conflicts, c)
		}
	}
	var out []Finding
	for _, g := range groups {
		if len(out) >= maxCrossPerSide {
			break
		}
		first := g.conflicts[0]
		var rest []string
		for _, c := range g.conflicts[1:] {
			if len(rest) == 3 {
				rest = append(rest, fmt.Sprintf("и ещё в %d", len(g.conflicts)-4))
				break
			}
			rest = append(rest, fmt.Sprintf("«%s» (стр. %d): %s", shortVol(c.cur), c.hit.Line.Page, strings.Join(c.vals, ", ")))
		}
		text := fmt.Sprintf("Согласованность между томами: «%s» для «%s» — в томе «%s» (стр. %d): %s; в томе «%s» (стр. %d): %s",
			g.key[0], g.key[1], shortVol(g.base), g.baseHit.Line.Page, strings.Join(g.baseVals, ", "), shortVol(first.cur), first.hit.Line.Page, strings.Join(first.vals, ", "))
		if len(rest) > 0 {
			text += "; также в томах " + strings.Join(rest, "; ")
		}
		text += ". Проверьте, что это разные элементы, а не расхождение."
		f := Finding{
			Kind: KindCrossVolume, Type: "content", Severity: "low", Description: text,
			Anchor:      trunc(first.hit.Line.Text, 200),
			Evidence:    Evidence{Rule: "cross-volume:" + g.key[0] + ":" + g.key[1]},
			Fingerprint: fingerprint(KindCrossVolume, side, g.key[0], g.key[1], g.base.Code, strings.Join(g.baseVals, ","), first.cur.Code, strings.Join(first.vals, ",")),
		}
		box := boxPtr(first.hit.Line.Box)
		if side == SideRef {
			f.RefDoc, f.PageRef, f.BoxRef, f.Evidence.VolumeRef = first.cur.DocID, first.hit.Line.Page, box, volLabel(first.cur)
		} else {
			f.TargetDoc, f.PageTarget, f.BoxTarget, f.Evidence.VolumeTgt = first.cur.DocID, first.hit.Line.Page, box, volLabel(first.cur)
		}
		out = append(out, f)
	}
	return out, rows
}

// shortVol names a volume in a sentence: its шифр, or the start of its title.
func shortVol(v *Volume) string {
	if v.RawCode != "" {
		return v.RawCode
	}
	return trunc(v.Title, 50)
}
