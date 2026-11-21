package cmp

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// Positions are matched by the words of their names and the unit. A position
// that has no counterpart is reported as information only: bills are often
// aggregated in one stage and split in the other, and a specification lists
// accompanying materials a bill does not.

const (
	itemMatchMin = 0.55  // name similarity to call two positions the same
	qtyTolerance = 0.005 // relative difference below which quantities are equal
	qtyMedium    = 0.03
	qtyHigh      = 0.10
	maxItemInfo  = 25 // information findings per table kind and volume
)

var (
	dashDigitRe = regexp.MustCompile(`([а-яa-z])-(\d)`)
	itemStop    = map[string]bool{"для": true, "при": true, "или": true, "без": true, "под": true, "над": true, "том": true, "числе": true, "гост": true}
)

func itemTokens(name string) map[string]bool {
	s := dashDigitRe.ReplaceAllString(NormKey(name), "$1$2")
	out := map[string]bool{}
	for _, w := range strings.FieldsFunc(s, func(r rune) bool { return !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == ',') }) {
		w = strings.Trim(w, ".,")
		hasDigit, letters := false, 0
		for _, r := range w {
			if unicode.IsDigit(r) {
				hasDigit = true
			} else if unicode.IsLetter(r) {
				letters++
			}
		}
		rs := []rune(w)
		switch {
		case itemStop[w]:
		case hasDigit && len(rs) >= 2:
			out[w] = true
		case letters >= 4:
			if len(rs) > 6 {
				w = string(rs[:6]) // crude stem: "траншеи"/"траншее"
			}
			out[w] = true
		}
	}
	return out
}

func dice(a, b map[string]bool) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := 0
	for k := range a {
		if b[k] {
			n++
		}
	}
	return 2 * float64(n) / float64(len(a)+len(b))
}

type itemPair struct{ a, b int }

// matchItems pairs positions one-to-one: same unit, similar names, best
// similarity first; among equally good candidates (a bill repeats "Выравнивающая
// стяжка…" with different quantities) the one in the same relative place wins.
func matchItems(a, b []Item) (pairs []itemPair, aOnly, bOnly []int) {
	ta := make([]map[string]bool, len(a))
	tb := make([]map[string]bool, len(b))
	for i := range a {
		ta[i] = itemTokens(a[i].Name)
	}
	for j := range b {
		tb[j] = itemTokens(b[j].Name)
	}
	type cand struct {
		i, j int
		sim  float64
		off  float64
	}
	var cs []cand
	for i := range a {
		for j := range b {
			if a[i].Unit != b[j].Unit {
				continue
			}
			s := dice(ta[i], tb[j])
			if s < itemMatchMin {
				continue
			}
			off := math.Abs(float64(i)/float64(len(a)) - float64(j)/float64(len(b)))
			cs = append(cs, cand{i, j, s, off})
		}
	}
	sort.SliceStable(cs, func(x, y int) bool {
		if math.Abs(cs[x].sim-cs[y].sim) > 1e-9 {
			return cs[x].sim > cs[y].sim
		}
		return cs[x].off < cs[y].off
	})
	usedA, usedB := make([]bool, len(a)), make([]bool, len(b))
	for _, c := range cs {
		if usedA[c.i] || usedB[c.j] {
			continue
		}
		usedA[c.i], usedB[c.j] = true, true
		pairs = append(pairs, itemPair{c.i, c.j})
	}
	sort.Slice(pairs, func(x, y int) bool { return pairs[x].a < pairs[y].a })
	for i := range a {
		if !usedA[i] {
			aOnly = append(aOnly, i)
		}
	}
	for j := range b {
		if !usedB[j] {
			bOnly = append(bOnly, j)
		}
	}
	return
}

func tableLabel(kind string) string {
	if kind == TableSpec {
		return "Спецификация"
	}
	return "ВОР"
}

func fmtQty(q float64, unit string) string {
	s := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.3f", q), "0"), ".")
	return strings.ReplaceAll(s, ".", ",") + " " + unit
}

func itemsOf(items []Item, kind string) []Item {
	var out []Item
	for _, it := range items {
		if it.Table == kind {
			out = append(out, it)
		}
	}
	return out
}

func box(it Item) *Box { b := it.Box; return &b }

// tableFindings compares the ВОР / specification positions of the reference
// volume with those of its target volume(s), and checks each side's ВОР
// against its own specification.
func tableFindings(p *Pair, st *Stats) []Finding {
	var tgtItems []Item
	for _, t := range p.Targets {
		tgtItems = append(tgtItems, t.Items...)
	}
	st.TableItems += len(p.Ref.Items) + len(tgtItems)
	var out []Finding
	refLbl, tgtLbl := volLabel(p.Ref), ""
	for i, t := range p.Targets {
		if i > 0 {
			tgtLbl += "; "
		}
		tgtLbl += volLabel(t)
	}
	for _, kind := range []string{TableVOR, TableSpec} {
		ra, tb := itemsOf(p.Ref.Items, kind), itemsOf(tgtItems, kind)
		if len(ra) == 0 || len(tb) == 0 {
			continue // nothing to compare with (one stage has no such table)
		}
		pairs, aOnly, bOnly := matchItems(ra, tb)
		// positions count toward the similarity: equal ones twice (once per side)
		st.simDen += float64(len(ra) + len(tb))
		for _, pr := range pairs {
			x, y := ra[pr.a], tb[pr.b]
			rel := math.Abs(y.Qty-x.Qty) / math.Max(math.Abs(x.Qty), 1e-9)
			if x.Qty == 0 && y.Qty == 0 || rel <= qtyTolerance {
				st.simNum += 2
				continue
			}
			sev := "low"
			if rel > qtyHigh {
				sev = "high"
			} else if rel > qtyMedium {
				sev = "medium"
			}
			note := ""
			// a fivefold jump is more likely two different positions with alike
			// names, or a misread cell, than a real edit: never "high"
			if ratio := y.Qty / math.Max(math.Abs(x.Qty), 1e-9); ratio > 5 || ratio < 0.2 {
				sev, note = "low", " (значения несопоставимы — проверьте, те ли это позиции и верно ли прочитана таблица)"
			}
			out = append(out, Finding{
				Kind: KindVOR, Type: "content", Severity: sev,
				Description: fmt.Sprintf("%s: «%s» — в ПД %s (стр. %d), в РД %s (стр. %d), изменение %+.1f%%%s", tableLabel(kind), trunc(x.Name, 110), fmtQty(x.Qty, x.Unit), x.Page, fmtQty(y.Qty, y.Unit), y.Page, 100*(y.Qty-x.Qty)/math.Max(math.Abs(x.Qty), 1e-9), note),
				RefDoc:      x.DocID, TargetDoc: y.DocID, PageRef: x.Page, PageTarget: y.Page, BoxRef: box(x), BoxTarget: box(y),
				Anchor:      trunc(x.Name, 200),
				Evidence:    Evidence{Ref: fmtQty(x.Qty, x.Unit), Target: fmtQty(y.Qty, y.Unit), VolumeRef: refLbl, VolumeTgt: tgtLbl, Rule: "vor-compare:" + kind},
				Fingerprint: fingerprint(KindVOR, p.Ref.Code, kind, x.Key, fmtQty(x.Qty, x.Unit), fmtQty(y.Qty, y.Unit)),
			})
		}
		for n, i := range aOnly {
			if n >= maxItemInfo {
				break
			}
			x := ra[i]
			out = append(out, Finding{
				Kind: KindVOR, Type: "content", Severity: "low",
				Description: fmt.Sprintf("%s: позиция «%s» (%s, ПД стр. %d) не найдена в РД — возможно, укрупнена, разбита или переименована", tableLabel(kind), trunc(x.Name, 110), fmtQty(x.Qty, x.Unit), x.Page),
				RefDoc:      x.DocID, PageRef: x.Page, BoxRef: box(x), Anchor: trunc(x.Name, 200),
				Evidence:    Evidence{Ref: fmtQty(x.Qty, x.Unit), VolumeRef: refLbl, VolumeTgt: tgtLbl, Rule: "vor-unmatched:" + kind},
				Fingerprint: fingerprint(KindVOR, p.Ref.Code, kind, x.Key, fmtQty(x.Qty, x.Unit), "-"),
			})
		}
		for n, j := range bOnly {
			if n >= maxItemInfo {
				break
			}
			y := tb[j]
			out = append(out, Finding{
				Kind: KindVOR, Type: "content", Severity: "low",
				Description: fmt.Sprintf("%s: позиция «%s» (%s, РД стр. %d) отсутствует в ПД — детализация или новая позиция", tableLabel(kind), trunc(y.Name, 110), fmtQty(y.Qty, y.Unit), y.Page),
				TargetDoc:   y.DocID, PageTarget: y.Page, BoxTarget: box(y), Anchor: trunc(y.Name, 200),
				Evidence:    Evidence{Target: fmtQty(y.Qty, y.Unit), VolumeRef: refLbl, VolumeTgt: tgtLbl, Rule: "vor-unmatched:" + kind},
				Fingerprint: fingerprint(KindVOR, p.Ref.Code, kind, "-", "-", y.Key, fmtQty(y.Qty, y.Unit)),
			})
		}
	}
	out = append(out, internalConsistency(p.Ref, false)...)
	for _, t := range p.Targets {
		out = append(out, internalConsistency(t, true)...)
	}
	return out
}

// internalConsistency checks a volume's bill of quantities against its own
// specification: a position found in both must agree. A bill position that
// the specification spreads over several lines (the position is aggregated)
// is compared with their sum; one with no specification counterpart is not
// a mismatch (a bill lists works, a specification materials).
func internalConsistency(v *Volume, target bool) []Finding {
	vor, spec := itemsOf(v.Items, TableVOR), itemsOf(v.Items, TableSpec)
	if len(vor) == 0 || len(spec) == 0 {
		return nil
	}
	ts := make([]map[string]bool, len(spec))
	for j := range spec {
		ts[j] = itemTokens(spec[j].Name)
	}
	var out []Finding
	for _, x := range vor {
		tx := itemTokens(x.Name)
		var cand []int
		for j, y := range spec {
			if y.Unit == x.Unit && dice(tx, ts[j]) >= itemMatchMin+0.05 {
				cand = append(cand, j)
			}
		}
		if len(cand) == 0 {
			continue
		}
		sum, best := 0.0, cand[0]
		ok := false
		for _, j := range cand {
			sum += spec[j].Qty
			if relDiff(x.Qty, spec[j].Qty) <= qtyTolerance {
				ok = true
			}
			if math.Abs(spec[j].Qty-x.Qty) < math.Abs(spec[best].Qty-x.Qty) {
				best = j
			}
		}
		if ok || relDiff(x.Qty, sum) <= qtyTolerance {
			continue // equal to a line, or to the sum of the lines it was split into
		}
		y := spec[best]
		rel := relDiff(x.Qty, y.Qty)
		sev := "low"
		if rel > qtyMedium {
			sev = "medium"
		}
		side := "ПД"
		if target {
			side = "РД"
		}
		f := Finding{
			Kind: KindVOR, Type: "content", Severity: sev,
			Description: fmt.Sprintf("ВОР и спецификация (%s) расходятся: «%s» — ВОР %s (стр. %d), спецификация %s (стр. %d)", side, trunc(x.Name, 100), fmtQty(x.Qty, x.Unit), x.Page, fmtQty(y.Qty, y.Unit), y.Page),
			Anchor:      trunc(x.Name, 200),
			Evidence:    Evidence{Ref: "ВОР " + fmtQty(x.Qty, x.Unit), Target: "спецификация " + fmtQty(y.Qty, y.Unit), Rule: "vor-vs-spec"},
			Fingerprint: fingerprint(KindVOR, v.Code, side, "internal", x.Key, fmtQty(x.Qty, x.Unit), fmtQty(y.Qty, y.Unit)),
		}
		if target {
			f.TargetDoc, f.PageTarget, f.BoxTarget = x.DocID, x.Page, box(x)
			f.Evidence.VolumeTgt = volLabel(v)
		} else {
			f.RefDoc, f.PageRef, f.BoxRef = x.DocID, x.Page, box(x)
			f.Evidence.VolumeRef = volLabel(v)
		}
		out = append(out, f)
		if len(out) >= 30 {
			break
		}
	}
	return out
}

func relDiff(a, b float64) float64 {
	return math.Abs(a-b) / math.Max(math.Abs(a), 1e-9)
}
