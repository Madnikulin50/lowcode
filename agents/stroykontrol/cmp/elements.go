package cmp

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// Проектные элементы (ТЗ 3.3.2, 4.2.7): марки конструкций и оборудования («Ф-2», «Пл-1», «БС-3»).
// Марки читаются по тексту листов — и текстовых, и чертежей, у которых есть текстовый слой, — и
// сопоставляются двумя способами:
//   - между сторонами: марка, которая есть в ПД и пропала в РД (или появилась в РД) в паре томов;
//   - внутри стороны: марка есть в спецификации, но не на чертежах, и наоборот.
// У чертежей, где текст «обведён» кривыми (так выгружают часть САПР), текстового слоя нет, и
// марок на них не прочитать: такие листы в этих проверках не участвуют. Находки низкой критичности и
// всегда «к проверке»: марку могли переименовать, а не убрать.

const (
	maxMarksPerFinding = 10
	// a pair of volumes, or a side, with fewer marks than this says nothing about marks
	minMarksToCompare = 3
)

// markRe: a capital letter and up to two more ("Пл-1", "БС-3"), an optional digit, a hyphen and one to three digits ("Ф-2", "ПЛ-12", "БС1-3", "Ф-2а").
var markRe = regexp.MustCompile(`(?:^|[^\p{L}\p{N}.\-])([А-ЯA-Z][А-Яа-яA-Za-z]{0,2}\d?-\d{1,3}[а-я]?)(?:[^\p{L}\p{N}\-]|$)`)

// markDeny are the letters that are not a mark's: the classes of concrete and rebar, the grades of mortar («М-100»), the
// forms of acts («КС-2»), the seismic scale («МSК-64»), units, the cipher letters.
var markDeny = map[string]bool{"М": true, "M": true, "КС": true, "KC": true, "МСК": true, "MSK": true, "МSК": true, "В": true, "B": true, "А": true, "A": true, "W": true, "F": true, "Т": true, "T": true, "О": true, "O": true, "С": true, "C": true, "Р": true, "P": true}

// markSeries are the discipline letters of a шифр: "КЖ1-3" names a sheet or a volume, not an element.
var markSeries = map[string]bool{"КЖ": true, "КМ": true, "АР": true, "ОВ": true, "ВК": true, "ЭС": true, "НК": true, "ЭОМ": true, "СС": true,
	"ПЗ": true, "ПД": true, "РД": true, "ТКР": true, "ГП": true, "ПОС": true, "АС": true, "КР": true, "ТХ": true, "ОП": true, "ИОС": true, "ОД": true}

func letters(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) {
			b.WriteRune(unicode.ToUpper(r))
		}
	}
	// Latin look-alikes are folded onto Cyrillic (S is a С)
	return strings.NewReplacer("S", "С").Replace(latinFold.Replace(b.String()))
}

// Marks reads the marks of a text line: canonical (upper-case, Latin look-alikes folded).
func Marks(text string) []string {
	if pointsOutside(NormKey(text)) {
		return nil // a norm, a contract, another volume: "таблица Т-4 ГОСТ …", "по шифру КЖ1-3"
	}
	var out []string
	for _, loc := range markRe.FindAllStringSubmatchIndex(text, -1) {
		raw := text[loc[2]:loc[3]]
		if rest := text[loc[3]:]; len(rest) >= 2 && rest[0] == '.' && rest[1] >= '0' && rest[1] <= '9' {
			continue // "Ф-2.1" is a clause number, not a mark
		}
		head := strings.TrimRight(raw[:strings.Index(raw, "-")], "0123456789")
		l := letters(head)
		if l == "" || markDeny[l] || markSeries[l] {
			continue
		}
		out = append(out, latinFold.Replace(strings.ToUpper(raw)))
	}
	return out
}

// markSet is the marks of a volume with the first place each was met: its text lines, and the
// lines of its drawing sheets.
type markSet map[string]markAt

type markAt struct {
	Page int
	Box  Box
	Line string
}

func volumeMarks(d *Document, v *Volume) markSet {
	out := markSet{}
	add := func(l Line) {
		if l.Stamp {
			return
		}
		for _, m := range Marks(l.Text) {
			if _, ok := out[m]; !ok {
				out[m] = markAt{Page: l.Page, Box: l.Box, Line: l.Text}
			}
		}
	}
	for _, l := range v.Lines {
		add(l)
	}
	for _, pn := range v.Drawings {
		if pn >= 1 && pn <= len(d.Pages) {
			for _, l := range d.Pages[pn-1].Lines {
				l.Page = pn
				add(l)
			}
		}
	}
	return out
}

func sortedMarks(m markSet, except markSet) []string {
	var out []string
	for k := range m {
		if _, ok := except[k]; !ok {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

func listMarks(ms []string) string {
	if len(ms) > maxMarksPerFinding {
		return strings.Join(ms[:maxMarksPerFinding], ", ") + fmt.Sprintf(" и ещё %d", len(ms)-maxMarksPerFinding)
	}
	return strings.Join(ms, ", ")
}

// elementFindings compares the marks of the two sides of a pair of volumes. A mark of the ПД that the РД does
// not have is «средняя» (a design element that disappeared or was renamed); a mark that is only in the РД is
// «низкая» (a more detailed РД is normal) and is not reported for fragments.
func elementFindings(p *Pair, docs map[string]*Document) []Finding {
	rd, td := docs[p.Ref.DocID], map[string]markSet{}
	if rd == nil {
		return nil
	}
	rm := volumeMarks(rd, p.Ref)
	tm := markSet{}
	for _, t := range p.Targets {
		d := docs[t.DocID]
		if d == nil {
			continue
		}
		for k, v := range volumeMarks(d, t) {
			if _, ok := tm[k]; !ok {
				tm[k] = v
			}
		}
		td[t.DocID] = nil
	}
	if len(rm) < minMarksToCompare || len(tm) < minMarksToCompare {
		return nil
	}
	label := volLabel(p.Ref)
	var out []Finding
	if gone := sortedMarks(rm, tm); len(gone) > 0 && !p.Fragment {
		first := rm[gone[0]]
		box := first.Box
		out = append(out, Finding{
			Kind: KindElement, Type: "content", Severity: "medium",
			Description: fmt.Sprintf("Марки, которые есть в ПД и не найдены в РД (том «%s»): %s. Проверьте, не убраны ли элементы или не переименованы.", trunc(label, 70), listMarks(gone)),
			RefDoc:      p.Ref.DocID, PageRef: first.Page, BoxRef: &box, Anchor: trunc(first.Line, 200),
			Evidence:    Evidence{VolumeRef: label, Rule: "marks-missing"},
			Fingerprint: fingerprint(KindElement, "missing", p.Ref.Code, p.Ref.Title, strings.Join(gone, ",")),
		})
	}
	if added := sortedMarks(tm, rm); len(added) > 0 {
		first := tm[added[0]]
		box := first.Box
		f := Finding{
			Kind: KindElement, Type: "content", Severity: "low",
			Description: fmt.Sprintf("Марки, которые есть в РД и не найдены в ПД (том «%s»): %s. Детализация РД сама по себе не нарушение; проверьте, что это новые элементы, а не переименованные.", trunc(label, 70), listMarks(added)),
			PageTarget:  first.Page, BoxTarget: &box, Anchor: trunc(first.Line, 200),
			Evidence:    Evidence{VolumeRef: label, Rule: "marks-added"},
			Fingerprint: fingerprint(KindElement, "added", p.Ref.Code, p.Ref.Title, strings.Join(added, ",")),
		}
		// the target document the mark was found in
		for _, t := range p.Targets {
			if d := docs[t.DocID]; d != nil && volumeMarks(d, t)[added[0]].Page == first.Page {
				f.TargetDoc = t.DocID
				break
			}
		}
		if f.TargetDoc == "" && len(p.Targets) > 0 {
			f.TargetDoc = p.Targets[0].DocID
		}
		out = append(out, f)
	}
	return out
}

// specMarks reads the marks of the names of the specification positions of a volume.
func specMarks(v *Volume) markSet {
	out := markSet{}
	for _, it := range v.Items {
		if it.Table != "spec" {
			continue
		}
		for _, m := range Marks(it.Name) {
			if _, ok := out[m]; !ok {
				out[m] = markAt{Page: it.Page, Box: it.Box, Line: it.Name}
			}
		}
	}
	return out
}

// specDrawingFindings checks, within one side, the specification against the drawings: a mark the
// specification lists that no sheet of the side mentions, and a mark the sheets carry that the specification
// does not know. Judged only where there is something to compare: a specification with marks and sheets with marks.
func specDrawingFindings(docs []*Document, vols []*Volume, side string) []Finding {
	byID := map[string]*Document{}
	for _, d := range docs {
		byID[d.ID] = d
	}
	spec, draw := markSet{}, markSet{}
	var specVol, drawVol *Volume
	for _, v := range vols {
		d := byID[v.DocID]
		if d == nil {
			continue
		}
		for k, a := range specMarks(v) {
			if _, ok := spec[k]; !ok {
				spec[k], specVol = a, firstVol(specVol, v, k, spec)
			}
		}
		// the marks of the drawings: every line of the volume except the specification's own rows
		inSpec := map[string]bool{}
		for _, it := range v.Items {
			inSpec[it.Name] = true
		}
		for k, a := range volumeMarks(d, v) {
			if inSpec[a.Line] {
				continue
			}
			if _, ok := draw[k]; !ok {
				draw[k], drawVol = a, firstVol(drawVol, v, k, draw)
			}
		}
	}
	if len(spec) < minMarksToCompare || len(draw) < minMarksToCompare {
		return nil
	}
	mk := func(rule, text string, ms []string, at markAt, v *Volume) Finding {
		box := at.Box
		f := Finding{
			Kind: KindCrossVolume, Type: "content", Severity: "low", Description: text,
			Anchor:      trunc(at.Line, 200),
			Evidence:    Evidence{Rule: rule, VolumeRef: "", VolumeTgt: ""},
			Fingerprint: fingerprint(KindCrossVolume, side, rule, strings.Join(ms, ",")),
		}
		if side == SideRef {
			f.RefDoc, f.PageRef, f.BoxRef, f.Evidence.VolumeRef = v.DocID, at.Page, &box, volLabel(v)
		} else {
			f.TargetDoc, f.PageTarget, f.BoxTarget, f.Evidence.VolumeTgt = v.DocID, at.Page, &box, volLabel(v)
		}
		return f
	}
	var out []Finding
	if only := sortedMarks(spec, draw); len(only) > 0 {
		out = append(out, mk("spec-not-on-drawings",
			fmt.Sprintf("Спецификация называет марки, которых нет ни в тексте, ни на чертежах: %s. Проверьте, не пропущены ли элементы на листах.", listMarks(only)), only, spec[only[0]], specVol))
	}
	if only := sortedMarks(draw, spec); len(only) > 0 {
		out = append(out, mk("drawings-not-in-spec",
			fmt.Sprintf("В тексте и на чертежах есть марки, которых нет в спецификации: %s. Проверьте, не пропущены ли позиции в спецификации.", listMarks(only)), only, draw[only[0]], drawVol))
	}
	return out
}

// firstVol keeps the volume a mark was first met in.
func firstVol(cur, v *Volume, mark string, set markSet) *Volume {
	if cur == nil {
		return v
	}
	return cur
}
