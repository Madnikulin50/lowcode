package cmp

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
)

// Options tunes Compare.
type Options struct {
	Imager      Imager       // optional: enables raster comparison of pages with no text layer
	MaxFindings int          // 0 → 400
	Mode        string       // ModeStages (default) | ModeVersions: how the sides are named in the texts
	Pairing     []ManualPair // pairs made by hand: they are compared as told, the rest is matched automatically
	NoFormat    bool         // skip the formatting checks of each side (numbering, contents, шифр, duplicates)
	NoCross     bool         // skip the consistency checks between the volumes of one side
	NoElements  bool         // skip the comparison of the marks of design elements (and the check of the specification against the drawings)
	Exclude     []Region     // areas of pages left out of the comparison (text inside them is not compared, drawing differences there are not reported)
}

const (
	capChanged = 150
	capMissing = 100
	capAdded   = 40
)

var leadNumRe = regexp.MustCompile(`^\d+(?:\.\d+)*[.)]?\s+`)

func trunc(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func fingerprint(parts ...string) string {
	h := sha1.Sum([]byte(strings.Join(parts, "\x1f")))
	return hex.EncodeToString(h[:])[:16]
}

func volInfo(v *Volume) VolumeInfo {
	return VolumeInfo{DocID: v.DocID, Index: v.Index, Code: v.RawCode, Title: v.Title, From: v.FirstPage, To: v.LastPage,
		Change: v.Change, Part: v.Part, Book: v.Book, Subsection: v.Subsection}
}

func volLabel(v *Volume) string {
	s := v.Title
	if v.RawCode != "" {
		s += " [" + v.RawCode + "]"
	}
	if s == "" {
		s = v.DocName
	}
	return s
}

// Compare checks the target documents against the reference documents.
// The reference stage (П) is the standard: the target (Р/РД) must agree
// with it but may be more detailed, may cover only part of a reference
// volume, and may be incomplete — volumes that have no counterpart are
// reported as informational "not released"/"out of scope", never as errors.
func Compare(refDocs, tgtDocs []*Document, opt Options) *Result {
	if opt.MaxFindings <= 0 {
		opt.MaxFindings = 400
	}
	res := &Result{Volumes: map[string][]VolumeInfo{}, Sheets: map[string]map[int]int{}, Headings: map[string][]DocHeading{}}
	if len(opt.Exclude) > 0 {
		var nr, nt int
		refDocs, nr = withoutRegions(refDocs, SideRef, opt.Exclude)
		tgtDocs, nt = withoutRegions(tgtDocs, SideTarget, opt.Exclude)
		res.Notes = append(res.Notes, fmt.Sprintf("Исключаемых областей: %d; из сравнения убрано строк текста: %d (ПД %d, РД %d).", len(opt.Exclude), nr+nt, nr, nt))
	}

	var refVols, tgtVols []*Volume
	collect := func(docs []*Document, side string, dst *[]*Volume) int {
		pages := 0
		for _, d := range docs {
			pages += len(d.Pages)
			res.Sheets[d.ID] = SheetMap(d.Pages)
			res.Headings[d.ID] = ExtractHeadings(d)
			for _, v := range DetectVolumes(d) {
				*dst = append(*dst, v)
				res.Volumes[d.ID] = append(res.Volumes[d.ID], volInfo(v))
			}
		}
		return pages
	}
	res.Notes = append(res.Notes, scanNotes(append(append([]*Document{}, refDocs...), tgtDocs...))...)
	res.Stats.RefPages = collect(refDocs, SideRef, &refVols)
	res.Stats.TargetPages = collect(tgtDocs, SideTarget, &tgtVols)
	// formatting checks run on the volumes as found, before manual slicing changes them
	var formatFs []Finding
	if !opt.NoFormat {
		formatFs = append(formatFindings(refDocs, refVols, SideRef), formatFindings(tgtDocs, tgtVols, SideTarget)...)
	}
	if !opt.NoCross {
		fr, rr := crossVolume(refDocs, refVols, SideRef)
		ft, rt := crossVolume(tgtDocs, tgtVols, SideTarget)
		formatFs = append(formatFs, append(fr, ft...)...)
		formatFs = append(formatFs, drawingTextFindings(refDocs, refVols, SideRef)...)
		formatFs = append(formatFs, drawingTextFindings(tgtDocs, tgtVols, SideTarget)...)
		if !opt.NoElements {
			formatFs = append(formatFs, specDrawingFindings(refDocs, refVols, SideRef)...)
			formatFs = append(formatFs, specDrawingFindings(tgtDocs, tgtVols, SideTarget)...)
		}
		res.Params = append(rr, rt...)
	}

	docByID := map[string]*Document{}
	for _, d := range append(append([]*Document{}, refDocs...), tgtDocs...) {
		docByID[d.ID] = d
	}

	side := map[string]string{}
	for _, d := range refDocs {
		side[d.ID] = SideRef
	}
	for _, d := range tgtDocs {
		side[d.ID] = SideTarget
	}
	var forced []*Pair
	if len(opt.Pairing) > 0 {
		var notes []string
		forced, refVols, tgtVols, notes = applyPairing(opt.Pairing, docByID, side, refVols, tgtVols)
		res.Notes = append(res.Notes, notes...)
	}
	// an older edition of a volume that is there again with a higher change number is not compared
	var dropRef, dropTgt []*Volume
	refVols, dropRef = supersede(refVols)
	tgtVols, dropTgt = supersede(tgtVols)
	for _, n := range []string{supersededNote(SideRef, dropRef), supersededNote(SideTarget, dropTgt)} {
		if n != "" {
			res.Notes = append(res.Notes, n)
		}
	}
	m := Match(refVols, tgtVols)
	m.Pairs = append(forced, m.Pairs...)
	res.Notes = append(res.Notes, editionNotes(m.Pairs, opt.Mode)...)
	for _, v := range m.RefOnly {
		res.RefOnly = append(res.RefOnly, volInfo(v))
	}
	for _, v := range m.TargetOnly {
		res.TargetOnly = append(res.TargetOnly, volInfo(v))
	}
	differing := map[string]bool{} // "docID:page" of reference pages with findings
	findings := formatFs

	for _, p := range m.Pairs {
		pi := PairInfo{Ref: volInfo(p.Ref), Fragment: p.Fragment, Manual: p.Manual, Score: p.Score}
		for _, t := range p.Targets {
			pi.Targets = append(pi.Targets, volInfo(t))
		}
		res.Pairs = append(res.Pairs, pi)
		res.Stats.ComparedPages += p.Ref.LastPage - p.Ref.FirstPage + 1

		fs := comparePair(p, &res.Stats)
		fs = append(fs, tableFindings(p, &res.Stats)...)
		fs = dropCovered(append(fs, cellFindings(p)...))
		fs = append(fs, compareDrawings(p, docByID, opt.Imager, &res.Stats, opt.Exclude)...)
		if !opt.NoElements {
			fs = append(fs, elementFindings(p, docByID)...)
		}
		if opt.Mode == ModeVersions {
			sc, cf, notes := changeRegistration(p, docByID, res.Sheets, fs)
			res.SheetChanges = append(res.SheetChanges, sc...)
			res.Notes = append(res.Notes, notes...)
			fs = append(fs, cf...)
		}
		for _, f := range fs {
			if f.PageRef > 0 {
				differing[fmt.Sprintf("%s:%d", f.RefDoc, f.PageRef)] = true
			}
		}
		findings = append(findings, fs...)
	}

	for _, v := range m.RefOnly {
		findings = append(findings, Finding{
			Kind: KindNotRelease, Type: "missing_page", Severity: "low",
			Description: fmt.Sprintf("В загруженной РД нет тома, соответствующего «%s» (ПД, стр. %d–%d). Возможно, он ещё не выпущен — это не считается ошибкой.", volLabel(v), v.FirstPage, v.LastPage),
			RefDoc:      v.DocID, PageRef: v.FirstPage,
			Evidence:    Evidence{VolumeRef: volLabel(v), Rule: "volume-unmatched"},
			Fingerprint: fingerprint(KindNotRelease, v.Code, v.Title),
		})
	}
	for _, v := range m.TargetOnly {
		findings = append(findings, Finding{
			Kind: KindScopeDiff, Type: "extra_page", Severity: "low",
			Description: fmt.Sprintf("Том РД «%s» (%s, стр. %d–%d) не сопоставлен ни с одним томом ПД — вне объёма загруженной ПД.", volLabel(v), v.DocName, v.FirstPage, v.LastPage),
			TargetDoc:   v.DocID, PageTarget: v.FirstPage,
			Evidence:    Evidence{VolumeTgt: volLabel(v), Rule: "volume-unmatched"},
			Fingerprint: fingerprint(KindScopeDiff, v.Code, v.Title),
		})
	}

	for i := range findings {
		if f := &findings[i]; f.OCR {
			// recognized text can carry misreadings: never "high", and say why; the weaker the recognition
			// of the page, the less the finding is worth
			conf := ocrConfOf(docByID, f)
			if f.Severity == "high" {
				f.Severity = "medium"
			}
			if conf > 0 && conf < ocrWeakConf {
				f.Severity = "low"
			}
			if conf > 0 {
				f.Description += fmt.Sprintf(" [по распознанному тексту скана, достоверность %.0f %%]", conf)
			} else {
				f.Description += " [по распознанному тексту скана]"
			}
			f.Evidence.Rule += "+ocr"
		}
		if findings[i].Fingerprint == "" {
			findings[i].Fingerprint = fingerprint(findings[i].Kind, findings[i].Anchor, findings[i].Description)
		}
	}
	rank := map[string]int{"high": 0, "medium": 1, "low": 2}
	sort.SliceStable(findings, func(i, j int) bool {
		if rank[findings[i].Severity] != rank[findings[j].Severity] {
			return rank[findings[i].Severity] < rank[findings[j].Severity]
		}
		return false
	})
	if len(findings) > opt.MaxFindings {
		res.Notes = append(res.Notes, fmt.Sprintf("Показаны первые %d из %d расхождений (по убыванию критичности).", opt.MaxFindings, len(findings)))
		findings = findings[:opt.MaxFindings]
	}
	for i := range findings {
		findings[i].Description = Relabel(findings[i].Description, opt.Mode)
	}
	res.Findings = findings

	res.Stats.DifferingPages = len(differing)
	if res.Stats.simDen > 0 {
		res.Stats.Similarity = 100 * res.Stats.simNum / res.Stats.simDen
		res.Stats.SimilarityKnown = true
	}
	if res.Stats.DrawingRef+res.Stats.DrawingTarget > 0 {
		res.Notes = append(res.Notes, fmt.Sprintf("Листов-чертежей: ПД %d, РД %d — сравниваются как изображения, текст на них не сверяется.", res.Stats.DrawingRef, res.Stats.DrawingTarget))
	}
	if !res.Stats.SimilarityKnown {
		res.Notes = append(res.Notes, "Схожесть не рассчитана: у сопоставленных томов нет текста для сравнения.")
	}
	if len(m.Pairs) == 0 && len(refVols) > 0 && len(tgtVols) > 0 {
		res.Notes = append(res.Notes, "Ни один том не сопоставлен по содержанию: текстовые слои ПД и РД не пересекаются (значимая часть содержимого — чертежи и таблицы на листах без текста). Сопоставьте вручную.")
	}
	res.Notes = append(res.Notes, fmt.Sprintf("Томов ПД: %d, томов РД: %d, сопоставлено: %d; ПД без пары: %d, РД без пары: %d.",
		len(refVols), len(tgtVols), len(m.Pairs), len(m.RefOnly), len(m.TargetOnly)))
	for i := range res.Notes {
		res.Notes[i] = Relabel(res.Notes[i], opt.Mode)
	}
	return res
}

func mergeLines(vs []*Volume) []Line {
	var out []Line
	for _, v := range vs {
		out = append(out, v.Lines...)
	}
	return out
}

func comparePair(p *Pair, st *Stats) []Finding {
	refLines := p.Ref.Lines
	tgtLines := mergeLines(p.Targets)
	a := make([]string, len(refLines))
	b := make([]string, len(tgtLines))
	for i, l := range refLines {
		a[i] = l.Key
	}
	for i, l := range tgtLines {
		b[i] = l.Key
	}
	if len(a) == 0 {
		// nothing in the reference to compare against (a drawings-only
		// volume): every target line would be reported as "added"
		return nil
	}
	st.RefLines += len(a)
	equal, changes := groupChanges(diffKeys(a, b), a, b)
	st.MatchedLines += len(equal)
	// full pair: how alike are the two; fragment: how much of the target is
	// confirmed by the reference (the rest of the reference is out of scope)
	if p.Fragment {
		st.simNum += float64(len(equal))
		st.simDen += float64(len(b))
	} else {
		st.simNum += 2 * float64(len(equal))
		st.simDen += float64(len(a) + len(b))
	}

	refLbl, tgtLbl := volLabel(p.Ref), ""
	for i, t := range p.Targets {
		if i > 0 {
			tgtLbl += "; "
		}
		tgtLbl += volLabel(t)
	}
	ev := func(r, t string) Evidence {
		return Evidence{Ref: trunc(r, 400), Target: trunc(t, 400), VolumeRef: refLbl, VolumeTgt: tgtLbl, Rule: "line-diff"}
	}

	var out []Finding
	var nChanged, nMissing, nAdded int
	var oChanged, oMissing, oAdded int
	for _, c := range changes {
		for _, pr := range c.pairs {
			nChanged++
			if nChanged > capChanged {
				oChanged++
				continue
			}
			rl, tl := refLines[pr[0]], tgtLines[pr[1]]
			cf := changedFinding(rl, tl, ev(rl.Text, tl.Text))
			cf.OCR = rl.OCR || tl.OCR
			out = append(out, cf)
		}
		for _, i := range c.delOnly {
			if p.Fragment {
				continue // reference content outside the target's scope
			}
			nMissing++
			if nMissing > capMissing {
				oMissing++
				continue
			}
			rl := refLines[i]
			kind := lineKind(rl.Key)
			out = append(out, Finding{
				Kind: kind, Type: "content", Severity: "medium",
				Description: fmt.Sprintf("Есть в ПД, не найдено в РД: «%s» (ПД, стр. %d)", trunc(rl.Text, 160), rl.Page),
				RefDoc:      rl.DocID, PageRef: rl.Page, BoxRef: boxPtr(rl.Box), Anchor: trunc(rl.Text, 200),
				Evidence:    ev(rl.Text, ""),
				Fingerprint: fingerprint(kind, p.Ref.Code, rl.Key, "-"),
				OCR:         rl.OCR,
			})
		}
		for _, j := range c.addOnly {
			nAdded++
			if nAdded > capAdded {
				oAdded++
				continue
			}
			tl := tgtLines[j]
			out = append(out, Finding{
				Kind: KindText, Type: "content", Severity: "low",
				Description: fmt.Sprintf("Есть в РД, нет в ПД (детализация?): «%s» (РД, стр. %d)", trunc(tl.Text, 160), tl.Page),
				TargetDoc:   tl.DocID, PageTarget: tl.Page, BoxTarget: boxPtr(tl.Box), Anchor: trunc(tl.Text, 200),
				Evidence:    ev("", tl.Text),
				Fingerprint: fingerprint(KindText, p.Ref.Code, "-", tl.Key),
				OCR:         tl.OCR,
			})
		}
	}
	if oChanged+oMissing+oAdded > 0 {
		out = append(out, Finding{
			Kind: KindText, Type: "content", Severity: "low",
			Description: fmt.Sprintf("Том «%s»: показаны не все расхождения — пропущено изменённых %d, отсутствующих в РД %d, добавленных в РД %d.", refLbl, oChanged, oMissing, oAdded),
			RefDoc:      p.Ref.DocID, PageRef: p.Ref.FirstPage,
			Evidence:    Evidence{VolumeRef: refLbl, Rule: "cap"},
			Fingerprint: fingerprint("cap", p.Ref.Code, p.Ref.Title),
		})
	}
	covered := map[string]bool{} // lines already reported as an edited pair
	for _, c := range changes {
		for _, pr := range c.pairs {
			covered[lineID(refLines[pr[0]])] = true
			covered[lineID(tgtLines[pr[1]])] = true
		}
	}
	out = append(out, paramFindings(p, refLines, tgtLines, covered)...)
	return out
}

func lineID(l Line) string { return fmt.Sprintf("%s|%d|%s", l.DocID, l.Page, l.Key) }

func boxPtr(b Box) *Box { return &b }

func changedFinding(rl, tl Line, ev Evidence) Finding {
	rk, tk := rl.Key, tl.Key
	f := Finding{
		Kind: lineKind(rl.Key), Type: "content",
		RefDoc: rl.DocID, TargetDoc: tl.DocID, PageRef: rl.Page, PageTarget: tl.Page,
		BoxRef: boxPtr(rl.Box), BoxTarget: boxPtr(tl.Box), Anchor: trunc(rl.Text, 200),
		Evidence: ev, Fingerprint: fingerprint(KindText, rl.Key, tl.Key),
	}
	rn, tn := numbersOf(rk), numbersOf(tk)
	switch {
	case leadNumRe.MatchString(rk) && leadNumRe.MatchString(tk) &&
		leadNumRe.ReplaceAllString(rk, "") == leadNumRe.ReplaceAllString(tk, ""):
		f.Type, f.Severity = "numbering", "low"
		f.Description = fmt.Sprintf("Изменена нумерация: «%s» (ПД, стр. %d) → «%s» (РД, стр. %d)", trunc(rl.Text, 100), rl.Page, trunc(tl.Text, 100), tl.Page)
	case !equalStrings(rn, tn):
		f.Severity = "high"
		f.Description = fmt.Sprintf("Расходятся значения: «%s» (ПД, стр. %d) → «%s» (РД, стр. %d)", trunc(rl.Text, 140), rl.Page, trunc(tl.Text, 140), tl.Page)
	case similarity(rk, tk) < 0.8:
		f.Severity = "medium"
		f.Description = fmt.Sprintf("Изменена формулировка: «%s» (ПД, стр. %d) → «%s» (РД, стр. %d)", trunc(rl.Text, 140), rl.Page, trunc(tl.Text, 140), tl.Page)
	default:
		f.Severity = "low"
		f.Description = fmt.Sprintf("Небольшое отличие: «%s» (ПД, стр. %d) → «%s» (РД, стр. %d)", trunc(rl.Text, 140), rl.Page, trunc(tl.Text, 140), tl.Page)
	}
	return f
}

func paramFindings(p *Pair, refLines, tgtLines []Line, covered map[string]bool) []Finding {
	var out []Finding
	refLbl := volLabel(p.Ref)
	for _, d := range compareParams(refLines, tgtLines, p.Fragment) {
		f := Finding{
			Kind: KindParam, Type: "content",
			Evidence: Evidence{VolumeRef: refLbl, Rule: "key-param:" + d.Name},
		}
		var rh, th *ParamHit
		// every differing value already sits on a line reported as edited:
		// the line finding says it all, a second one would only duplicate it
		if d.Mode == "conflict" && allCovered(d, covered) {
			continue
		}
		if len(d.OnlyRef) > 0 {
			h := d.RefHits[d.OnlyRef[0]]
			rh = &h
		}
		if len(d.OnlyTgt) > 0 {
			h := d.TgtHits[d.OnlyTgt[0]]
			th = &h
		}
		switch d.Mode {
		case "conflict":
			f.Severity = "high"
			f.Description = fmt.Sprintf("%s: в ПД %s, в РД %s", d.Name, strings.Join(d.OnlyRef, ", "), strings.Join(d.OnlyTgt, ", "))
			f.Evidence.Ref, f.Evidence.Target = strings.Join(d.OnlyRef, ", "), strings.Join(d.OnlyTgt, ", ")
		default:
			f.Severity = "low"
			f.Description = fmt.Sprintf("%s: в РД не найдено значение из ПД (%s)", d.Name, strings.Join(d.OnlyRef, ", "))
			f.Evidence.Ref = strings.Join(d.OnlyRef, ", ")
		}
		if rh != nil {
			f.RefDoc, f.PageRef, f.BoxRef, f.Anchor = rh.Line.DocID, rh.Line.Page, boxPtr(rh.Line.Box), trunc(rh.Line.Text, 200)
		}
		if th != nil {
			f.TargetDoc, f.PageTarget, f.BoxTarget = th.Line.DocID, th.Line.Page, boxPtr(th.Line.Box)
			if f.Anchor == "" {
				f.Anchor = trunc(th.Line.Text, 200)
			}
		}
		f.Fingerprint = fingerprint(KindParam, p.Ref.Code, d.Name, strings.Join(d.OnlyRef, ","), strings.Join(d.OnlyTgt, ","))
		out = append(out, f)
	}
	return out
}

// maxDrawingPairs bounds the image comparisons of one volume pair.
const maxDrawingPairs = 60

type sheet struct {
	doc  *Document
	page int
}

func drawingSheets(vs []*Volume, docs map[string]*Document) []sheet {
	var out []sheet
	for _, v := range vs {
		d := docs[v.DocID]
		if d == nil {
			continue
		}
		for _, pn := range v.Drawings {
			out = append(out, sheet{d, pn})
		}
	}
	return out
}

func aspect(s sheet) float64 {
	pg := s.doc.Pages[s.page-1]
	if pg.H == 0 {
		return 0
	}
	return pg.W / pg.H
}

// compareDrawings compares drawing sheets (large sheets and sheets without a
// usable text layer) as images. Sheets can only be paired by position, so
// this is done only when both sides have the same number of drawing sheets
// of the same proportions; otherwise one informational finding asks for a
// visual check — a wrong pairing would produce nothing but noise.
func compareDrawings(p *Pair, docs map[string]*Document, img Imager, st *Stats, skip []Region) []Finding {
	refSheets := drawingSheets([]*Volume{p.Ref}, docs)
	tgtSheets := drawingSheets(p.Targets, docs)
	st.DrawingRef += len(refSheets)
	st.DrawingTarget += len(tgtSheets)
	if len(refSheets) == 0 && len(tgtSheets) == 0 {
		return nil
	}
	label := volLabel(p.Ref)
	visual := func(why string) []Finding {
		f := Finding{
			Kind: KindDrawing, Type: "drawing", Severity: "low",
			Description: fmt.Sprintf("Листы-чертежи тома «%s»: в ПД %d, в РД %d — %s. Проверьте чертежи визуально.", trunc(label, 80), len(refSheets), len(tgtSheets), why),
			Evidence:    Evidence{VolumeRef: label, Rule: "drawings-visual"},
			Fingerprint: fingerprint(KindDrawing, p.Ref.Code, p.Ref.Title, fmt.Sprint(len(refSheets)), fmt.Sprint(len(tgtSheets))),
		}
		if len(refSheets) > 0 {
			f.RefDoc, f.PageRef = refSheets[0].doc.ID, refSheets[0].page
		}
		if len(tgtSheets) > 0 {
			f.TargetDoc, f.PageTarget = tgtSheets[0].doc.ID, tgtSheets[0].page
		}
		return []Finding{f}
	}
	pairs, onlyRef, onlyTgt, by := pairSheets(refSheets, tgtSheets)
	if pairs == nil && img != nil {
		// no numbers to go by and a different count of sheets: a sheet is looked for by what is drawn on it
		pairs, onlyRef, onlyTgt = pairByContent(refSheets, tgtSheets, img)
		by = " (листы сопоставлены по содержимому)"
	}
	if pairs == nil {
		return visual("число листов различается, автоматическое сопоставление недоступно")
	}
	var out []Finding
	if len(onlyRef)+len(onlyTgt) > 0 {
		basis := "по номерам листов в основных надписях"
		if strings.Contains(by, "содержимому") {
			basis = "по содержимому (чертежи не похожи ни на один лист другой стороны)"
		}
		out = append(out, visual(fmt.Sprintf("%s не нашлось пары: только в ПД — %s; только в РД — %s", basis, sheetNos(onlyRef), sheetNos(onlyTgt)))...)
	}
	if img == nil {
		return append(out, visual("сравнение изображений недоступно")...)
	}
	n := len(pairs)
	if n > maxDrawingPairs {
		n = maxDrawingPairs
	}
	skipped := 0
	mismatch := ""
	for i := 0; i < n; i++ {
		rs, ts := pairs[i][0], pairs[i][1]
		ia, err1 := img(rs.doc.ID, rs.page)
		ib, err2 := img(ts.doc.ID, ts.page)
		if err1 != nil || err2 != nil || ia == nil || ib == nil {
			continue
		}
		// the sheets are compared after their contents are aligned (a shifted frame, another scale or paper do not count)
		onA, onB, aligned := AlignedDiff(ia, ib)
		ra, ta := aspect(rs), aspect(ts)
		if !aligned && (ra == 0 || ta == 0 || math.Abs(ra-ta)/ra > 0.05) {
			skipped++
			if mismatch == "" {
				rp, tp := rs.doc.Pages[rs.page-1], ts.doc.Pages[ts.page-1]
				mismatch = fmt.Sprintf(" (например, стр. %d ПД — %s, стр. %d РД — %s)", rs.page, SheetFormat(rp.W, rp.H), ts.page, SheetFormat(tp.W, tp.H))
			}
			continue
		}
		type regionPair struct{ a, b Box }
		var regions []regionPair
		for k := range onA {
			a, b := onA[k], onB[k]
			if len(skip) > 0 && (excluded(skip, SideRef, rs.doc.ID, rs.page, a) || excluded(skip, SideTarget, ts.doc.ID, ts.page, b)) {
				continue // an excluded area of either sheet
			}
			regions = append(regions, regionPair{a, b})
		}
		if len(regions) > 3 {
			regions = regions[:3]
		}
		rule := "raster-diff"
		how := ""
		if aligned {
			rule, how = "raster-diff+aligned", " (листы сравнены после совмещения по содержимому)"
		}
		for k, rg := range regions {
			ba, bb := rg.a, rg.b
			sev := "low"
			if (ba.X1-ba.X0)*(ba.Y1-ba.Y0) >= 0.01 {
				sev = "medium"
			}
			out = append(out, Finding{
				Kind: KindDrawing, Type: "drawing", Severity: sev,
				Description: fmt.Sprintf("Чертёж: область отличается между ПД (стр. %d) и РД (стр. %d)%s%s.", rs.page, ts.page, by, how),
				RefDoc:      rs.doc.ID, TargetDoc: ts.doc.ID, PageRef: rs.page, PageTarget: ts.page,
				BoxRef: &ba, BoxTarget: &bb,
				Evidence:    Evidence{VolumeRef: label, Rule: rule},
				Fingerprint: fingerprint(KindDrawing, p.Ref.Code, fmt.Sprint(i), fmt.Sprint(k), ba.String()),
			})
		}
	}
	if skipped > 0 || len(pairs) > n {
		out = append(out, visual(fmt.Sprintf("часть листов не сопоставлена (разные пропорции: %d%s, сверх лимита %d: %d)", skipped, mismatch, maxDrawingPairs, len(pairs)-n))...)
	}
	return out
}

// sheetNo is the sheet number of a drawing sheet from its title block (0: not read).
func sheetNo(s sheet) int {
	if s.page < 1 || s.page > len(s.doc.Pages) {
		return 0
	}
	no, _ := sheetNumber(s.doc.Pages[s.page-1])
	return no
}

func sheetNos(ss []sheet) string {
	if len(ss) == 0 {
		return "нет"
	}
	var parts []string
	for i, s := range ss {
		if i == 8 {
			parts = append(parts, fmt.Sprintf("… (всего %d)", len(ss)))
			break
		}
		if no := sheetNo(s); no > 0 {
			parts = append(parts, fmt.Sprintf("лист %d (стр. %d)", no, s.page))
		} else {
			parts = append(parts, fmt.Sprintf("стр. %d", s.page))
		}
	}
	return strings.Join(parts, ", ")
}

// contentPairMax: how different two sheets may be (share of the grid) and still be taken for the same sheet.
const contentPairMax = 0.30

// maxContentPairs bounds the sheet pairs tried by their content (every sheet against every sheet).
const maxContentPairs = 2500

// pairByContent pairs sheets by how alike their drawings are once aligned: the best pair first, a sheet
// in one pair only. Sheets with no sufficiently alike sheet on the other side stay unpaired. nil when
// there are too many sheets to try.
func pairByContent(refs, tgts []sheet, img Imager) (pairs [][2]sheet, onlyRef, onlyTgt []sheet) {
	if len(refs) == 0 || len(tgts) == 0 || len(refs)*len(tgts) > maxContentPairs {
		return nil, nil, nil
	}
	// every picture is read once and only its signature is kept: a drawing sheet is a very large image
	sig := func(s sheet) SheetSig {
		im, err := img(s.doc.ID, s.page)
		if err != nil || im == nil {
			return SheetSig{}
		}
		return SignatureOf(im)
	}
	ri, ti := make([]SheetSig, len(refs)), make([]SheetSig, len(tgts))
	for i, s := range refs {
		ri[i] = sig(s)
	}
	for j, s := range tgts {
		ti[j] = sig(s)
	}
	type cand struct {
		i, j  int
		score float64
	}
	var cands []cand
	for i := range refs {
		for j := range tgts {
			if sc := ri[i].Change(ti[j]); sc <= contentPairMax {
				cands = append(cands, cand{i, j, sc})
			}
		}
	}
	sort.SliceStable(cands, func(a, b int) bool { return cands[a].score < cands[b].score })
	usedR, usedT := map[int]bool{}, map[int]bool{}
	for _, c := range cands {
		if usedR[c.i] || usedT[c.j] {
			continue
		}
		usedR[c.i], usedT[c.j] = true, true
		pairs = append(pairs, [2]sheet{refs[c.i], tgts[c.j]})
	}
	if len(pairs) == 0 {
		return nil, nil, nil
	}
	sort.SliceStable(pairs, func(a, b int) bool { return pairs[a][0].page < pairs[b][0].page })
	for i, s := range refs {
		if !usedR[i] {
			onlyRef = append(onlyRef, s)
		}
	}
	for j, s := range tgts {
		if !usedT[j] {
			onlyTgt = append(onlyTgt, s)
		}
	}
	return pairs, onlyRef, onlyTgt
}

// pairSheets pairs the drawing sheets of two volumes: by the sheet numbers of the title blocks when
// most sheets on both sides have a readable and unique one (a sheet that was inserted, removed or moved
// keeps its pair), otherwise by position when both sides have the same number of sheets. nil pairs
// means that no pairing is possible. by says what the pairs rest on, for the text of a finding.
func pairSheets(refs, tgts []sheet) (pairs [][2]sheet, onlyRef, onlyTgt []sheet, by string) {
	numbered := func(ss []sheet) (map[int]sheet, int, bool) {
		m := map[int]sheet{}
		read := 0
		for _, s := range ss {
			no := sheetNo(s)
			if no == 0 {
				continue
			}
			read++
			if _, dup := m[no]; dup {
				return nil, 0, false // the numbers restart inside the volume: they do not identify a sheet
			}
			m[no] = s
		}
		return m, read, true
	}
	rm, rn, rok := numbered(refs)
	tm, tn, tok := numbered(tgts)
	if rok && tok && len(refs) > 0 && len(tgts) > 0 && float64(rn) >= 0.6*float64(len(refs)) && float64(tn) >= 0.6*float64(len(tgts)) {
		for _, r := range refs {
			no := sheetNo(r)
			if t, ok := tm[no]; ok && no > 0 {
				pairs = append(pairs, [2]sheet{r, t})
			} else {
				onlyRef = append(onlyRef, r)
			}
		}
		for _, t := range tgts {
			if no := sheetNo(t); no == 0 || rm[no].doc == nil {
				onlyTgt = append(onlyTgt, t)
			}
		}
		if len(pairs) > 0 {
			return pairs, onlyRef, onlyTgt, " (листы сопоставлены по номерам в основных надписях)"
		}
		pairs, onlyRef, onlyTgt = nil, nil, nil
	}
	if len(refs) == len(tgts) && len(refs) > 0 {
		for i := range refs {
			pairs = append(pairs, [2]sheet{refs[i], tgts[i]})
		}
		return pairs, nil, nil, ""
	}
	return nil, nil, nil, ""
}

// ocrConfOf is the lowest recognition confidence of the pages a finding stands on (0: not known).
func ocrConfOf(docs map[string]*Document, f *Finding) float64 {
	conf := 0.0
	for _, at := range []struct {
		doc  string
		page int
	}{{f.RefDoc, f.PageRef}, {f.TargetDoc, f.PageTarget}} {
		d := docs[at.doc]
		if d == nil || at.page < 1 || at.page > len(d.Pages) {
			continue
		}
		if c := d.Pages[at.page-1].OCRConf; c > 0 && (conf == 0 || c < conf) {
			conf = c
		}
	}
	return conf
}

func allCovered(d paramDiff, covered map[string]bool) bool {
	for _, v := range d.OnlyRef {
		if !covered[lineID(d.RefHits[v].Line)] {
			return false
		}
	}
	for _, v := range d.OnlyTgt {
		if !covered[lineID(d.TgtHits[v].Line)] {
			return false
		}
	}
	return true
}

var (
	posCodeRe = regexp.MustCompile(`^[а-яa-z]{1,3}-?\d+(?:\.\d+)*$`)
	pureNumRe = regexp.MustCompile(`^\d+(?:\.\d+)?$`)
)

// lineKind guesses whether a row is a table row (position code and/or
// several bare numbers) or running text.
func lineKind(key string) string {
	key = leadNumRe.ReplaceAllString(key, "")
	fields := strings.Fields(key)
	nums := 0
	for _, f := range fields {
		if pureNumRe.MatchString(f) {
			nums++
		}
	}
	if nums >= 2 || (nums >= 1 && len(fields) >= 3 && posCodeRe.MatchString(fields[0])) {
		return KindTableRow
	}
	return KindText
}
