package cmp

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// Text on paired drawing sheets (notes, dimensions, marks, schedules drawn as text) is compared as
// a set of labels: a drawing has no reading order, and a shifted view moves everything. A label that
// is only on one side is a candidate for an edit when a similar one sits at about the same place
// on the other.

const (
	capDrawChanged = 40
	capDrawMissing = 20
	capDrawAdded   = 20
	// pairing of a removed label with an added one: how alike, how far apart (page fractions)
	drawPairSim  = 0.5
	drawPairDist = 0.2
)

type drawLabel struct {
	text, key string
	box       Box
}

// drawingLabels joins the words of a sheet into labels: words on one line with a small gap.
func drawingLabels(pg Page) []drawLabel {
	ws := append([]Word(nil), pg.Labels...)
	if len(ws) == 0 {
		return nil
	}
	sort.SliceStable(ws, func(i, j int) bool {
		ci, cj := (ws[i].B.Y0+ws[i].B.Y1)/2, (ws[j].B.Y0+ws[j].B.Y1)/2
		if ci != cj {
			return ci < cj
		}
		return ws[i].B.X0 < ws[j].B.X0
	})
	// a word of height h may be followed by a gap of about a character and a half
	asp := 1.0
	if pg.W > 0 && pg.H > 0 {
		asp = pg.H / pg.W // x fractions are of the width, y of the height
	}
	var out []drawLabel
	var cur []Word
	flush := func() {
		if len(cur) == 0 {
			return
		}
		parts := make([]string, len(cur))
		box := cur[0].B
		for i, w := range cur {
			parts[i] = w.T
			box = box.union(w.B)
		}
		text := strings.Join(parts, " ")
		out = append(out, drawLabel{text: text, key: NormKey(text), box: box})
		cur = nil
	}
	for _, w := range ws {
		if len(cur) > 0 {
			prev := cur[len(cur)-1]
			h := prev.B.Y1 - prev.B.Y0
			dy := (w.B.Y0+w.B.Y1)/2 - (prev.B.Y0+prev.B.Y1)/2
			if dy < 0 {
				dy = -dy
			}
			gap := w.B.X0 - prev.B.X1
			if dy > h*0.5 || gap > h*asp*1.5 || gap < -h*asp {
				flush()
			}
		}
		cur = append(cur, w)
	}
	flush()
	// noise: marks without letters or digits, single characters
	res := out[:0]
	for _, l := range out {
		if len([]rune(l.key)) < 2 {
			continue
		}
		has := false
		for _, r := range l.key {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				has = true
				break
			}
		}
		if has && !stampRe.MatchString(l.key) {
			res = append(res, l)
		}
	}
	return res
}

func labelDist(a, b Box) float64 {
	dx := (a.X0+a.X1)/2 - (b.X0+b.X1)/2
	dy := (a.Y0+a.Y1)/2 - (b.Y0+b.Y1)/2
	return dx*dx + dy*dy
}

// sheetTextFindings compares the text of one pair of drawing sheets.
func sheetTextFindings(rs, ts sheet, volume, by string, skip []Region, st *Stats) []Finding {
	if rs.page < 1 || rs.page > len(rs.doc.Pages) || ts.page < 1 || ts.page > len(ts.doc.Pages) {
		return nil
	}
	rp, tp := rs.doc.Pages[rs.page-1], ts.doc.Pages[ts.page-1]
	keep := func(side string, sh sheet, ls []drawLabel) []drawLabel {
		if len(skip) == 0 {
			return ls
		}
		out := ls[:0]
		for _, l := range ls {
			if !excluded(skip, side, sh.doc.ID, sh.page, l.box) {
				out = append(out, l)
			}
		}
		return out
	}
	ra, ta := keep(SideRef, rs, drawingLabels(rp)), keep(SideTarget, ts, drawingLabels(tp))
	if len(ra) == 0 || len(ta) == 0 {
		return nil // no text layer on one of the sheets: only the picture says something
	}
	// multiset difference
	byKey := map[string][]int{}
	for j, l := range ta {
		byKey[l.key] = append(byKey[l.key], j)
	}
	usedT := make([]bool, len(ta))
	var onlyR []int
	equal := 0
	for i, l := range ra {
		if js := byKey[l.key]; len(js) > 0 {
			// the nearest of the equal labels
			best := 0
			for k, j := range js {
				if labelDist(l.box, ta[j].box) < labelDist(l.box, ta[js[best]].box) {
					best = k
				}
			}
			usedT[js[best]] = true
			byKey[l.key] = append(js[:best], js[best+1:]...)
			equal++
			continue
		}
		onlyR = append(onlyR, i)
	}
	var onlyT []int
	for j := range ta {
		if !usedT[j] {
			onlyT = append(onlyT, j)
		}
	}
	st.simNum += 2 * float64(equal)
	st.simDen += float64(len(ra) + len(ta))
	if len(onlyR)+len(onlyT) == 0 {
		return nil
	}
	// a removed label next to a similar added one is an edit
	type cand struct {
		i, j int
		sc   float64
	}
	var cands []cand
	for _, i := range onlyR {
		for _, j := range onlyT {
			if labelDist(ra[i].box, ta[j].box) > drawPairDist*drawPairDist {
				continue
			}
			if sim := similarity(ra[i].key, ta[j].key); sim >= drawPairSim {
				cands = append(cands, cand{i, j, sim})
			}
		}
	}
	sort.SliceStable(cands, func(a, b int) bool { return cands[a].sc > cands[b].sc })
	usedR := map[int]bool{}
	usedAdd := map[int]bool{}
	type pair struct{ i, j int }
	var changed []pair
	for _, c := range cands {
		if usedR[c.i] || usedAdd[c.j] {
			continue
		}
		usedR[c.i], usedAdd[c.j] = true, true
		changed = append(changed, pair{c.i, c.j})
	}
	sort.Slice(changed, func(a, b int) bool { return ra[changed[a].i].box.Y0 < ra[changed[b].i].box.Y0 })

	ev := func(r, t string) Evidence {
		return Evidence{Ref: trunc(r, 400), Target: trunc(t, 400), VolumeRef: volume, Rule: "drawing-text"}
	}
	line := func(l drawLabel, sh sheet) Line {
		return Line{DocID: sh.doc.ID, Page: sh.page, Text: l.text, Key: l.key, Box: l.box}
	}
	var out []Finding
	nCh, nMiss, nAdd := 0, 0, 0
	for _, c := range changed {
		nCh++
		if nCh > capDrawChanged {
			continue
		}
		rl, tl := line(ra[c.i], rs), line(ta[c.j], ts)
		f := changedFinding(rl, tl, ev(rl.Text, tl.Text))
		f.Kind = KindDrawing
		f.Description += by
		f.Fingerprint = fingerprint(KindDrawing, "text", rl.Key, tl.Key)
		out = append(out, f)
	}
	for _, i := range onlyR {
		if usedR[i] {
			continue
		}
		nMiss++
		if nMiss > capDrawMissing {
			continue
		}
		l := ra[i]
		out = append(out, Finding{
			Kind: KindDrawing, Type: "content", Severity: "medium",
			Description: fmt.Sprintf("На чертеже есть в ПД, не найдено в РД: «%s» (ПД, стр. %d; РД, стр. %d)%s", trunc(l.text, 160), rs.page, ts.page, by),
			RefDoc:      rs.doc.ID, TargetDoc: ts.doc.ID, PageRef: rs.page, PageTarget: ts.page, BoxRef: boxPtr(l.box), Anchor: trunc(l.text, 200),
			Evidence:    ev(l.text, ""),
			Fingerprint: fingerprint(KindDrawing, "text-missing", l.key),
		})
	}
	for _, j := range onlyT {
		if usedAdd[j] {
			continue
		}
		nAdd++
		if nAdd > capDrawAdded {
			continue
		}
		l := ta[j]
		out = append(out, Finding{
			Kind: KindDrawing, Type: "content", Severity: "low",
			Description: fmt.Sprintf("На чертеже есть в РД, нет в ПД: «%s» (РД, стр. %d; ПД, стр. %d)%s", trunc(l.text, 160), ts.page, rs.page, by),
			RefDoc:      rs.doc.ID, TargetDoc: ts.doc.ID, PageRef: rs.page, PageTarget: ts.page, BoxTarget: boxPtr(l.box), Anchor: trunc(l.text, 200),
			Evidence:    ev("", l.text),
			Fingerprint: fingerprint(KindDrawing, "text-added", l.key),
		})
	}
	if over := nCh - capDrawChanged + nMiss - capDrawMissing + nAdd - capDrawAdded; over > 0 && (nCh > capDrawChanged || nMiss > capDrawMissing || nAdd > capDrawAdded) {
		out = append(out, Finding{
			Kind: KindDrawing, Type: "content", Severity: "low",
			Description: fmt.Sprintf("Чертёж ПД стр. %d / РД стр. %d: показаны не все различия надписей (изменённых %d, только в ПД %d, только в РД %d).", rs.page, ts.page, nCh, nMiss, nAdd),
			RefDoc:      rs.doc.ID, TargetDoc: ts.doc.ID, PageRef: rs.page, PageTarget: ts.page,
			Evidence:    Evidence{VolumeRef: volume, Rule: "cap"},
			Fingerprint: fingerprint(KindDrawing, "text-cap", fmt.Sprint(rs.page), fmt.Sprint(ts.page)),
		})
	}
	return out
}
