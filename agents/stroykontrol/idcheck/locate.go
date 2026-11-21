package idcheck

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Word is one OCR word with its box in page-relative coordinates (0..1),
// so boxes stay valid whatever resolution the viewer renders the page at.
type Word struct {
	Text string  `json:"t"`
	Line int     `json:"l"` // block*10000+par*100+line — words on one line share it
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	W    float64 `json:"w"`
	H    float64 `json:"h"`
}

// Box is a highlight rectangle on a page (relative coordinates).
type Box struct {
	X, Y, W, H float64
}

// String encodes a box for storage in a record field: "x,y,w,h".
func (b Box) String() string {
	return fmt.Sprintf("%.4f,%.4f,%.4f,%.4f", b.X, b.Y, b.W, b.H)
}

// Words OCRs a page image into positioned words (tesseract TSV), cached
// next to the image as .tsv.
func (t Tesseract) Words(ctx context.Context, img string) ([]Word, error) {
	if !t.Available() {
		return nil, nil
	}
	cache := strings.TrimSuffix(img, ".jpg") + ".tsv"
	raw, err := os.ReadFile(cache)
	if err != nil {
		raw, err = exec.CommandContext(ctx, t.Bin, img, "stdout", "-l", t.Lang, "--psm", "6", "tsv").Output()
		if err != nil {
			return nil, fmt.Errorf("tesseract tsv: %w", err)
		}
		_ = os.WriteFile(cache, raw, 0o644)
	}
	return parseTSV(string(raw)), nil
}

// parseTSV reads tesseract's TSV: level page block par line word left top
// width height conf text. The level-1 row carries the page size.
func parseTSV(s string) []Word {
	var out []Word
	var pw, ph float64
	sc := bufio.NewScanner(strings.NewReader(s))
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		f := strings.Split(sc.Text(), "\t")
		if len(f) < 12 || f[0] == "level" {
			continue
		}
		num := func(i int) float64 { v, _ := strconv.ParseFloat(f[i], 64); return v }
		switch f[0] {
		case "1":
			pw, ph = num(8), num(9)
		case "5":
			txt := strings.TrimSpace(f[11])
			if txt == "" || pw == 0 || ph == 0 {
				continue
			}
			out = append(out, Word{
				Text: txt,
				Line: int(num(2))*10000 + int(num(3))*100 + int(num(4)),
				X:    num(6) / pw, Y: num(7) / ph, W: num(8) / pw, H: num(9) / ph,
			})
		}
	}
	return out
}

// FindText locates a phrase among a page's words: the shortest run of
// consecutive words whose compacted concatenation contains the compacted
// needle. Returns the union box and true when found.
func FindText(words []Word, needle string) (Box, bool) {
	n := compact(needle)
	if len([]rune(n)) < 3 {
		return Box{}, false
	}
	best, bestLen := Box{}, 0
	for i := range words {
		acc := ""
		for j := i; j < len(words) && j < i+25; j++ {
			acc += compact(words[j].Text)
			if strings.Contains(acc, n) {
				if bestLen == 0 || j-i+1 < bestLen {
					best, bestLen = union(words[i:j+1]), j-i+1
				}
				break
			}
			if len([]rune(acc)) > len([]rune(n))+40 {
				break
			}
		}
	}
	if bestLen > 0 {
		return best, true
	}
	return fuzzyFind(words, n)
}

// fuzzyFind tolerates OCR slips in longer needles (act numbers, codes):
// the run of consecutive words closest to the needle by edit distance,
// accepted within ~15% of its length.
func fuzzyFind(words []Word, n string) (Box, bool) {
	nl := len([]rune(n))
	if nl < 8 {
		return Box{}, false
	}
	limit := max(2, nl*15/100)
	bestD, best := limit+1, Box{}
	for i := range words {
		acc := ""
		for j := i; j < len(words) && j < i+15; j++ {
			acc += compact(words[j].Text)
			al := len([]rune(acc))
			if al >= nl-limit {
				if d := levenshtein(acc, n); d < bestD {
					bestD, best = d, union(words[i:j+1])
				}
			}
			if al > nl+limit {
				break
			}
		}
	}
	return best, bestD <= limit
}

func union(ws []Word) Box {
	x0, y0, x1, y1 := 1.0, 1.0, 0.0, 0.0
	for _, w := range ws {
		x0, y0 = min(x0, w.X), min(y0, w.Y)
		x1, y1 = max(x1, w.X+w.W), max(y1, w.Y+w.H)
	}
	return Box{X: x0, Y: y0, W: x1 - x0, H: y1 - y0}
}

// lineBox widens a box to the whole OCR line it sits on — a value alone is
// a few millimetres wide, the line gives the reader context.
func lineBox(words []Word, b Box) Box {
	cy := b.Y + b.H/2
	var line []Word
	for _, w := range words {
		if cy >= w.Y && cy <= w.Y+w.H {
			line = append(line, w)
		}
	}
	if len(line) == 0 {
		return b
	}
	lb := union(line)
	return union([]Word{{X: lb.X, Y: min(lb.Y, b.Y), W: lb.W, H: max(lb.Y+lb.H, b.Y+b.H) - min(lb.Y, b.Y)}})
}

// fieldAnchors are labels printed next to a field — used when the value
// itself was not found (it's missing, or OCR mangled it).
var fieldAnchors = map[string][]string{
	"number":         {"АКТ"},
	"date":           {"АКТ"},
	"volume":         {"Vпр", "Vф", "Мф", "предъявлены следующие работы"},
	"vol_project":    {"Vпр", "предъявлены следующие работы"},
	"vol_fact":       {"Vф", "предъявлены следующие работы"},
	"p2_rd_code":     {"Работы выполнены по проектной"},
	"p6_rd_code":     {"Работы выполнены в соответствии"},
	"p6":             {"Работы выполнены в соответствии"},
	"p6_sufficiency": {"Работы выполнены в соответствии"},
	"p6_project":     {"Работы выполнены в соответствии"},
	"p3":             {"При выполнении работ применены"},
	"appendix":       {"Приложения"},
	"attached":       {"Приложения"},
	"appendix_date":  {"Приложения"},
	"order":          {"приказ"},
	"order_date":     {"приказ"},
	"nrs_id":         {"идентификационный"},
	"nrs":            {"идентификационный"},
	"inn":            {"ИНН"},
	"ogrn":           {"ОГРН"},
	"egrul":          {"ОГРН"},
	"name":           {"ОГРН"},
	"status":         {"ОГРН"},
	"address":        {"ОГРН"},
	"scheme_rd_code": {},
}

// LocateFinding picks the page and box for a finding using the OCR words
// of the act's pages. Search order: the reported value (Actual), the
// expected value, then the field's printed label; act-level fields are
// searched on the act pages, page-bound findings on their own page.
func LocateFinding(f *Finding, pages []Page, words map[int][]Word, actPages []int) (page int, box Box, ok bool) {
	candidates := []int{}
	if f.Page > 0 {
		candidates = append(candidates, f.Page)
	} else {
		candidates = append(candidates, actPages...)
	}
	var needles []string
	for _, s := range []string{f.Actual, f.Expected} {
		s = strings.TrimSpace(s)
		if s == "" || strings.HasPrefix(s, "не ") || strings.HasPrefix(s, "действует") {
			continue
		}
		needles = append(needles, s)
		// act numbers: OCR mangles the head ("№2/2.14.1.9/ОП№6" → "иЛОПА6"),
		// the tail ("ИССО 2.2/СКТРАСТ") usually survives
		if parts := strings.Split(s, "/"); len(parts) >= 3 {
			needles = append(needles, strings.Join(parts[len(parts)-2:], "/"))
		}
		// "проект 26.3 / факт 25" → also try the bare numbers
		for _, tok := range strings.Fields(s) {
			if _, isNum := ParseNum(tok); isNum && strings.ContainsAny(tok, ".,") {
				needles = append(needles, strings.ReplaceAll(tok, ".", ","), tok)
			}
		}
	}
	for _, n := range needles {
		for _, p := range candidates {
			if b, found := FindText(words[p], n); found {
				return p, lineBox(words[p], b), true
			}
		}
	}
	for _, a := range fieldAnchors[f.Field] {
		for _, p := range candidates {
			if b, found := FindText(words[p], a); found {
				return p, lineBox(words[p], b), true
			}
		}
	}
	if len(candidates) > 0 {
		return candidates[0], Box{}, false
	}
	return 0, Box{}, false
}
