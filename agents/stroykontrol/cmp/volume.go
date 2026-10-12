package cmp

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// A шифр looks like "2024-15-КЖ1", "ПД-123/2024-АР", "01-24-ПЗ.1": groups of
// letters/digits joined by - . / with at least one digit and one letter.
var codeRe = regexp.MustCompile(`[0-9A-Za-zА-Яа-я]{1,8}(?:[-./][0-9A-Za-zА-Яа-я]{1,8}){1,5}`)

// latinFold folds Latin look-alikes onto Cyrillic so "KЖ1" and "КЖ1" match.
var latinFold = strings.NewReplacer(
	"A", "А", "B", "В", "C", "С", "E", "Е", "H", "Н", "K", "К", "M", "М",
	"O", "О", "P", "Р", "T", "Т", "X", "Х", "Y", "У",
)

// NormCode canonicalizes a шифр for comparison.
func NormCode(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = latinFold.Replace(s)
	s = strings.NewReplacer("–", "-", "—", "-", "/", "-", " ", "").Replace(s)
	return s
}

func plausibleCode(c string) bool {
	if n := len([]rune(c)); n < 5 || n > 32 {
		return false
	}
	hasDigit, letters, seps := false, 0, 0
	for _, r := range c {
		switch {
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsLetter(r):
			letters++
		case r == '-' || r == '.' || r == '/':
			seps++
		}
	}
	// a real шифр has a digit, a discipline mark of at least two letters and
	// at least two parts ("2024-15-КЖ1"); this rejects dates ("02.24г"),
	// table codes ("Т380-71") and plain words with a dash ("13-Расчеты")
	if !hasDigit || letters < 2 || seps < 2 {
		return false
	}
	// dates ("12.03.2024") and section numbers never contain letters, so
	// the check above already drops them; also drop things like "п.4.2-а".
	return true
}

// pageCode returns the шифр found in the title block of a page ("" if none).
// Only title-block rows count: shifr-like tokens elsewhere (table cells,
// references, drawing labels) are not volume marks.
func pageCode(p Page) (raw string) {
	score := map[string]float64{}
	first := map[string]string{}
	for _, l := range p.Lines {
		// the title block's own row (stage / sheet / sheets), which also holds the шифр
		if !l.Stamp || !stampRe.MatchString(l.Key) {
			continue
		}
		for _, m := range codeRe.FindAllString(l.Text, -1) {
			n := NormCode(m)
			if !plausibleCode(n) {
				continue
			}
			score[n]++
			if _, ok := first[n]; !ok {
				first[n] = m
			}
		}
	}
	best, bs := "", 0.0
	keys := make([]string, 0, len(score))
	for k := range score {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if score[k] > bs {
			best, bs = k, score[k]
		}
	}
	return first[best]
}

// coverCode is the шифр printed on the title page (e.g. "03102022/РД-3-ПИК1-НК2"),
// used to name a volume whose sheets have no readable title block.
func coverCode(d *Document) string {
	score := map[string]int{}
	first := map[string]string{}
	for i := 0; i < len(d.Pages) && i < 3; i++ {
		for _, l := range d.Pages[i].Lines {
			text := stationRe.ReplaceAllString(l.Text, " ")
			for _, m := range codeRe.FindAllString(text, -1) {
				n := NormCode(m)
				if plausibleCode(n) && len(n) >= 8 && strings.ContainsAny(m, "/-") && !strings.Contains(m, "..") {
					score[n]++
					if _, ok := first[n]; !ok {
						first[n] = m
					}
				}
			}
		}
	}
	best, bs := "", 0
	for k, v := range score {
		if v > bs || (v == bs && k < best) {
			best, bs = k, v
		}
	}
	return first[best]
}

// road stationing ("ПК33+94-ПК47+20") looks like a шифр but is not one
var stationRe = regexp.MustCompile(`ПК\s?\d+\s?\+\s?\d+(?:\s?[-–]\s?ПК\s?\d+\s?\+\s?\d+)?`)

var titleRe = regexp.MustCompile(`(?i)^(том|раздел|подраздел|книга)\s*[\dIVXLа-я.]*[\s:.-]`)

// pageTitle picks a descriptive heading for the first page of a volume.
func pageTitle(p Page) string {
	for _, l := range p.Lines {
		if l.Stamp {
			continue
		}
		if titleRe.MatchString(l.Text) && len(l.Text) <= 120 {
			return l.Text
		}
	}
	for _, l := range p.Lines {
		if !l.Stamp && len([]rune(l.Text)) >= 6 && len([]rune(l.Text)) <= 100 {
			return l.Text
		}
	}
	return ""
}

// minVolumePages: a stretch of sheets shorter than this is not a volume of
// its own — a stray code on one sheet must not split a document.
const minVolumePages = 2

// DetectVolumes splits a document into volumes by the шифр of the title
// block. A шифр counts only if it is read from the title block on at least
// two sheets; stretches shorter than minVolumePages are absorbed by their
// neighbours. A file with one шифр, or none, is a single volume — the usual
// case, since drawing sets rarely carry a readable title block.
func DetectVolumes(d *Document) []*Volume {
	nameBase := strings.TrimSuffix(d.Name, ".pdf")
	if len(d.Pages) == 0 {
		return nil
	}
	if d.ManualCode != "" {
		return []*Volume{buildVolume(d, 0, d.ManualCode, d.ManualTitle, 1, len(d.Pages), nameBase)}
	}

	raw := make([]string, len(d.Pages))
	count := map[string]int{}
	for i, p := range d.Pages {
		raw[i] = pageCode(p)
		if raw[i] != "" {
			count[NormCode(raw[i])]++
		}
	}
	codes := make([]string, len(d.Pages))
	for i := range codes {
		if raw[i] != "" && count[NormCode(raw[i])] >= 2 {
			codes[i] = raw[i]
		}
	}
	// sheets without a trusted code continue the previous volume; leading
	// ones adopt the first code that follows.
	for i := 1; i < len(codes); i++ {
		if codes[i] == "" {
			codes[i] = codes[i-1]
		}
	}
	for i := len(codes) - 2; i >= 0; i-- {
		if codes[i] == "" {
			codes[i] = codes[i+1]
		}
	}

	type run struct {
		code     string
		from, to int // 0-based, inclusive
	}
	var runs []run
	for i, c := range codes {
		if len(runs) > 0 && NormCode(runs[len(runs)-1].code) == NormCode(c) {
			runs[len(runs)-1].to = i
			continue
		}
		runs = append(runs, run{c, i, i})
	}
	// absorb short runs into the previous one (or the next, for the first)
	for changed := true; changed && len(runs) > 1; {
		changed = false
		for i, r := range runs {
			if r.to-r.from+1 >= minVolumePages {
				continue
			}
			if i > 0 {
				runs[i-1].to = r.to
				runs = append(runs[:i], runs[i+1:]...)
			} else {
				runs[1].from = r.from
				runs = runs[1:]
			}
			changed = true
			break
		}
		// neighbouring runs that ended up with the same code merge
		for i := len(runs) - 1; i > 0; i-- {
			if NormCode(runs[i].code) == NormCode(runs[i-1].code) {
				runs[i-1].to = runs[i].to
				runs = append(runs[:i], runs[i+1:]...)
			}
		}
	}

	var vols []*Volume
	for _, r := range runs {
		code := r.code
		if code == "" {
			code = coverCode(d)
		}
		vols = append(vols, buildVolume(d, len(vols), code, "", r.from+1, r.to+1, nameBase))
	}
	return vols
}

func buildVolume(d *Document, idx int, rawCode, title string, from, to int, nameBase string) *Volume {
	v := &Volume{
		DocID: d.ID, DocName: d.Name, Side: d.Side, Index: idx,
		Code: NormCode(rawCode), RawCode: rawCode, Title: title,
		FirstPage: from, LastPage: to,
	}
	if v.Title == "" {
		v.Title = pageTitle(d.Pages[from-1])
	}
	if v.Title == "" {
		v.Title = nameBase
	}
	items, tablePages := volumeItems(d, from, to)
	v.Items = items
	for pn := from; pn <= to; pn++ {
		pg := d.Pages[pn-1]
		if tablePages[pn] {
			continue // compared as a table of positions, not line by line
		}
		if pg.Drawing {
			v.Drawings = append(v.Drawings, pn)
			continue
		}
		if pg.Cover {
			continue
		}
		for _, l := range pg.Lines {
			if l.Stamp || skipLine(l.Key) {
				continue
			}
			v.Lines = append(v.Lines, l)
		}
	}
	return v
}

var (
	pageNumRe = regexp.MustCompile(`^(стр\.?|лист|л\.)?\s*\d{1,3}$`)
	// dates, years and margin-stamp remnants: change from edition to edition
	// by nature and say nothing about the design
	dateOnlyRe = regexp.MustCompile(`^(["«]\s*["»]\s*)?(\d{1,2}\.\d{2}\.\d{2,4}|\d{4}\s*г?\.?)$`)
	marginRe   = regexp.MustCompile(`^(взам\.? ?инв|инв\.? ?(n|№)|подп\.? и дата|подпись и дата)`)
)

// skipLine drops lines that carry no comparable content: bare page
// numbers, rows with no letters at all, dates and one-/two-character noise.
func skipLine(key string) bool {
	if len([]rune(key)) < 3 || pageNumRe.MatchString(key) || dateOnlyRe.MatchString(key) || marginRe.MatchString(key) {
		return true
	}
	for _, r := range key {
		if unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// Series is the letter part of a шифр ("КЖ", "АР", "ПЗ") — the discipline.
func Series(code string) string {
	var b strings.Builder
	for _, r := range code {
		if unicode.IsLetter(r) {
			b.WriteRune(r)
		}
	}
	s := b.String()
	// шифры often start with a project prefix "ПД"; keep the last 2-3
	// letters block, which is the section mark.
	rs := []rune(s)
	if len(rs) > 3 {
		return string(rs[len(rs)-2:])
	}
	return s
}
