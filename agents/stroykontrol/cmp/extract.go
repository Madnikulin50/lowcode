package cmp

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// IsPDF reports whether data looks like a PDF.
func IsPDF(data []byte) bool {
	return len(data) > 4 && bytes.HasPrefix(data, []byte("%PDF"))
}

var (
	pageRe = regexp.MustCompile(`(?s)<page width="([\d.]+)" height="([\d.]+)">(.*?)</page>`)
	wordRe = regexp.MustCompile(`<word xMin="([\d.]+)" yMin="([\d.]+)" xMax="([\d.]+)" yMax="([\d.]+)">(.*?)</word>`)
)

type word struct {
	x0, y0, x1, y1 float64
	text           string
}

// minTextWords is the number of words below which a page is treated as a
// drawing/scan with no usable text layer.
const minTextWords = 12

// leftMarginPt: words that end within this distance (points) of the left edge
// belong to the vertical margin stamp; content starts at the 20 mm binding margin (57 pt).
const leftMarginPt = 50.0

// ExtractPages runs `pdftotext -bbox` and turns the words into rows of text
// with positions. Results are cached under cacheDir (keyed by content hash)
// when cacheDir is not empty.
func ExtractPages(docID string, data []byte, cacheDir string) ([]Page, error) {
	key := hashKey(data)
	cachePath := ""
	if cacheDir != "" {
		cachePath = filepath.Join(cacheDir, "cmp6-"+key+".json")
		if b, err := os.ReadFile(cachePath); err == nil {
			var pages []Page
			if json.Unmarshal(b, &pages) == nil {
				return withDocID(pages, docID), nil
			}
		}
	}

	tmp, err := os.MkdirTemp("", "cmp-extract-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	in := filepath.Join(tmp, "in.pdf")
	if err := os.WriteFile(in, data, 0o644); err != nil {
		return nil, err
	}
	out, err := exec.Command("pdftotext", "-bbox", in, "-").Output()
	if err != nil {
		msg := ""
		if ee, ok := err.(*exec.ExitError); ok {
			msg = strings.TrimSpace(string(ee.Stderr))
		}
		return nil, fmt.Errorf("pdftotext -bbox: %w (%s)", err, msg)
	}
	pages := ParseBBox(docID, string(out))
	if cachePath != "" {
		if err := os.MkdirAll(cacheDir, 0o755); err == nil {
			if b, err := json.Marshal(pages); err == nil {
				_ = os.WriteFile(cachePath, b, 0o644)
			}
		}
	}
	return pages, nil
}

func withDocID(pages []Page, docID string) []Page {
	for i := range pages {
		for j := range pages[i].Lines {
			pages[i].Lines[j].DocID = docID
		}
	}
	return pages
}

func hashKey(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])[:24]
}

// ParseBBox parses `pdftotext -bbox` output.
func ParseBBox(docID, out string) []Page {
	var pages []Page
	for i, pm := range pageRe.FindAllStringSubmatch(out, -1) {
		w, _ := strconv.ParseFloat(pm[1], 64)
		h, _ := strconv.ParseFloat(pm[2], 64)
		var words []word
		for _, wm := range wordRe.FindAllStringSubmatch(pm[3], -1) {
			x0, _ := strconv.ParseFloat(wm[1], 64)
			y0, _ := strconv.ParseFloat(wm[2], 64)
			x1, _ := strconv.ParseFloat(wm[3], 64)
			y1, _ := strconv.ParseFloat(wm[4], 64)
			t := strings.TrimSpace(html.UnescapeString(wm[5]))
			// the strip along the left edge holds only the vertical margin
			// stamp (Инв. № подл. / Подпись и дата / Взам. инв. №); its words
			// would otherwise be glued onto table rows that share their height
			if t == "" || (w >= 400 && x1 < leftMarginPt) {
				continue
			}
			words = append(words, word{x0, y0, x1, y1, t})
		}
		pages = append(pages, buildPage(docID, i+1, w, h, words))
	}
	return pages
}

// buildPage groups words into visual rows (words whose vertical centres
// fall within the same band) and records each row's box in page fractions.
func buildPage(docID string, number int, w, h float64, words []word) Page {
	p := Page{Number: number, W: w, H: h, HasText: len(words) >= minTextWords}
	if w <= 0 || h <= 0 || len(words) == 0 {
		p.Drawing = true
		return p
	}
	sort.SliceStable(words, func(i, j int) bool {
		ci, cj := (words[i].y0+words[i].y1)/2, (words[j].y0+words[j].y1)/2
		if ci != cj {
			return ci < cj
		}
		return words[i].x0 < words[j].x0
	})
	var rows [][]word
	var curTop, curBot float64
	for _, wd := range words {
		c := (wd.y0 + wd.y1) / 2
		if len(rows) > 0 && c >= curTop && c <= curBot {
			rows[len(rows)-1] = append(rows[len(rows)-1], wd)
			if wd.y0 < curTop {
				curTop = wd.y0
			}
			if wd.y1 > curBot {
				curBot = wd.y1
			}
			continue
		}
		rows = append(rows, []word{wd})
		curTop, curBot = wd.y0, wd.y1
	}
	for _, row := range rows {
		sort.SliceStable(row, func(i, j int) bool { return row[i].x0 < row[j].x0 })
		parts := make([]string, 0, len(row))
		box := Box{X0: 1, Y0: 1}
		for _, wd := range row {
			parts = append(parts, wd.text)
			box = box.union(Box{wd.x0 / w, wd.y0 / h, wd.x1 / w, wd.y1 / h})
		}
		if box.X0 > box.X1 { // union started from X0:1,Y0:1 — guard degenerate input
			continue
		}
		box.X0, box.Y0 = clamp01(box.X0), clamp01(box.Y0)
		box.X1, box.Y1 = clamp01(box.X1), clamp01(box.Y1)
		text := strings.Join(parts, " ")
		p.Lines = append(p.Lines, Line{DocID: docID, Page: number, Text: text, Key: NormKey(text), Box: box})
	}
	p.Drawing = !p.HasText || isDrawing(w, h, p.Lines)
	if !p.Drawing {
		p.Words = make([]Word, len(words))
		for i, wd := range words {
			p.Words[i] = Word{wd.text, Box{clamp01(wd.x0 / w), clamp01(wd.y0 / h), clamp01(wd.x1 / w), clamp01(wd.y1 / h)}}
		}
	}
	for i := range p.Lines {
		p.Lines[i].Stamp = isStampRow(p.Drawing, p.Lines[i])
	}
	p.Cover = number <= 6 && isCover(p.Lines)
	return p
}

// isDrawing: very large sheets (A2 and up) and sheets whose rows are mostly
// short scattered labels are drawings — a row-by-row text diff of them is
// noise, they are compared as images instead.
func isDrawing(w, h float64, lines []Line) bool {
	if math.Max(w, h) > 1500 {
		return true
	}
	if len(lines) < 20 {
		return false
	}
	lens := make([]int, len(lines))
	for i, l := range lines {
		lens[i] = len([]rune(l.Text))
	}
	sort.Ints(lens)
	return lens[len(lens)/2] < 12
}

var (
	// words that only occur in a title block (основная надпись) and in the
	// "Взам. инв. / Подпись и дата / Инв. № подл." margin stamps
	stampRe = regexp.MustCompile(`кол\.?\s?уч|n\s?док|№\s?док|взам\.?\s?инв|подпись и дата|подп\.? и дата|инв\.?\s?(n|№)\s?подл|лист\s*\d+\s*листов|стадия\s*\S{0,3}\s*лист|изм\.?\s+кол`)
	// signature rows of the title block (names and dates change with every edition)
	roleRe = regexp.MustCompile(`проверил|разраб\.|н\.? ?контр|\bгип\b|кгип|рук\.? ?гр|нач\.? ?отд|гл\.? ?(инженер|специалист)`)
	// title page, approval sheet, contents register — front matter, not design content
	coverRe = regexp.MustCompile(`согласовано|утверждаю|рабочая документация|проектная документация|общество с ограниченной|содержание тома|состав тома|ведомость ссылочных|ведомость рабочих чертежей`)
)

// isStampRow: on text pages a row is title block only if it carries title-
// block words (body text near the page bottom is content); on drawing
// sheets the bottom-right corner is the title block by position as well.
func isStampRow(drawing bool, l Line) bool {
	if stampRe.MatchString(l.Key) || (roleRe.MatchString(l.Key) && l.Box.Y0 > 0.75) {
		return true
	}
	return drawing && inTitleBlock(l.Box)
}

func isCover(lines []Line) bool {
	for _, l := range lines {
		if coverRe.MatchString(l.Key) {
			return true
		}
	}
	return false
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// inTitleBlock: the основная надпись sits in the bottom-right corner of the
// sheet (ГОСТ Р 21.101); its content (sheet number, dates, signatures,
// шифр) legitimately differs between stages and must not be compared.
func inTitleBlock(b Box) bool {
	cy := (b.Y0 + b.Y1) / 2
	cx := (b.X0 + b.X1) / 2
	return cy > 0.86 && cx > 0.35
}

var (
	spaceRe   = regexp.MustCompile(`\s+`)
	decCommaR = regexp.MustCompile(`(\d),(\d)`)
	dashFold  = strings.NewReplacer("–", "-", "—", "-", "‐", "-", "−", "-", "ё", "е", "Ё", "е", "«", "\"", "»", "\"", "№", "N")
)

// NormKey is the comparable form of a line: lower case, unified dashes and
// quotes, decimal comma → dot, collapsed whitespace, no edge punctuation.
func NormKey(s string) string {
	s = strings.ToLower(s)
	s = dashFold.Replace(s)
	s = decCommaR.ReplaceAllString(s, "$1.$2")
	s = spaceRe.ReplaceAllString(s, " ")
	return strings.Trim(s, " .,;:")
}
