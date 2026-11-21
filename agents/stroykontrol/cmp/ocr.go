package cmp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// OCR reads pages that have no text layer (scans) with tesseract, so they can
// be compared like any other text page. Large sheets (A2 and up) are drawings
// and are never recognized.
type OCR struct {
	Bin      string // tesseract binary ("" disables OCR)
	Lang     string // "rus"
	DPI      int    // rendering resolution for recognition
	MaxPages int    // recognized pages per document, 0 = 40
	CacheDir string // per-page results, keyed by content hash
}

// NewOCR returns a recognizer; it is disabled when the binary is missing.
func NewOCR(bin, lang, cacheDir string) *OCR {
	if bin == "" {
		bin = "tesseract"
	}
	if _, err := exec.LookPath(bin); err != nil {
		return &OCR{}
	}
	return &OCR{Bin: bin, Lang: lang, DPI: 200, CacheDir: cacheDir}
}

// Enabled reports whether recognition is available.
func (o *OCR) Enabled() bool { return o != nil && o.Bin != "" }

// minOCRConf: tesseract's word confidence (0..100) below which a word is noise.
const minOCRConf = 35

// words recognizes one page (1-based) of a PDF and returns words in PDF
// points (scaled to the page's own size).
func (o *OCR) words(ctx context.Context, data []byte, page int, wpt, hpt float64) ([]word, float64, error) {
	sum := sha256.Sum256(data)
	key := hex.EncodeToString(sum[:])[:24]
	cache := ""
	if o.CacheDir != "" {
		cache = filepath.Join(o.CacheDir, fmt.Sprintf("ocr2-%s-%d-%d.json", key, o.DPI, page))
		if b, err := os.ReadFile(cache); err == nil {
			var c cachedOCR
			if json.Unmarshal(b, &c) == nil {
				return c.words(), c.Conf, nil
			}
		}
	}
	tmp, err := os.MkdirTemp("", "cmp-ocr-*")
	if err != nil {
		return nil, 0, err
	}
	defer os.RemoveAll(tmp)
	pdf := filepath.Join(tmp, "in.pdf")
	if err := os.WriteFile(pdf, data, 0o644); err != nil {
		return nil, 0, err
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	img := filepath.Join(tmp, "p")
	if out, err := exec.CommandContext(ctx, "pdftoppm", "-r", strconv.Itoa(o.DPI), "-f", strconv.Itoa(page), "-l", strconv.Itoa(page), "-png", "-singlefile", pdf, img).CombinedOutput(); err != nil {
		return nil, 0, fmt.Errorf("pdftoppm: %v: %s", err, strings.TrimSpace(string(out)))
	}
	out, err := exec.CommandContext(ctx, o.Bin, img+".png", "stdout", "-l", o.Lang, "--psm", "6", "tsv").Output()
	if err != nil {
		return nil, 0, fmt.Errorf("tesseract: %w", err)
	}
	ws, conf := parseTSV(string(out), wpt, hpt)
	if cache != "" {
		if err := os.MkdirAll(o.CacheDir, 0o755); err == nil {
			if b, err := json.Marshal(newCachedOCR(ws, conf)); err == nil {
				_ = os.WriteFile(cache, b, 0o644)
			}
		}
	}
	return ws, conf, nil
}

// cachedOCR is the on-disk form of a recognized page (word has unexported fields).
type cachedOCR struct {
	Conf  float64      `json:"conf"`
	Words []cachedWord `json:"words"`
}

type cachedWord struct {
	X0, Y0, X1, Y1 float64
	T              string
}

func newCachedOCR(ws []word, conf float64) cachedOCR {
	c := cachedOCR{Conf: conf}
	for _, w := range ws {
		c.Words = append(c.Words, cachedWord{w.x0, w.y0, w.x1, w.y1, w.text})
	}
	return c
}

func (c cachedOCR) words() []word {
	out := make([]word, len(c.Words))
	for i, w := range c.Words {
		out[i] = word{w.X0, w.Y0, w.X1, w.Y1, w.T}
	}
	return out
}

// parseTSV turns tesseract's TSV into words. Pixel coordinates are scaled
// to the page's size in points using the image size, which tesseract
// reports on its level-1 (page) row.
func parseTSV(tsv string, wpt, hpt float64) ([]word, float64) {
	var imgW, imgH, confSum float64
	var ws []word
	for i, line := range strings.Split(tsv, "\n") {
		f := strings.Split(strings.TrimRight(line, "\r"), "\t")
		if i == 0 || len(f) < 12 {
			continue
		}
		level, _ := strconv.Atoi(f[0])
		l, _ := strconv.ParseFloat(f[6], 64)
		t, _ := strconv.ParseFloat(f[7], 64)
		w, _ := strconv.ParseFloat(f[8], 64)
		h, _ := strconv.ParseFloat(f[9], 64)
		if level == 1 {
			imgW, imgH = w, h
			continue
		}
		conf, _ := strconv.ParseFloat(f[10], 64)
		text := strings.TrimSpace(strings.Join(f[11:], "\t"))
		if level != 5 || text == "" || conf < minOCRConf || imgW == 0 || imgH == 0 {
			continue
		}
		ws = append(ws, word{
			x0: l / imgW * wpt, y0: t / imgH * hpt,
			x1: (l + w) / imgW * wpt, y1: (t + h) / imgH * hpt,
			text: text,
		})
		confSum += conf
	}
	if len(ws) == 0 {
		return nil, 0
	}
	return ws, confSum / float64(len(ws))
}

// ocrMinWords: fewer recognized words than this is a drawing with a few labels, not a text page.
const ocrMinWords = 20

// A recognized page is text only if tesseract is confident and most of what it
// read looks like words or numbers; a drawing's lines and symbols produce
// low-confidence junk that must not be compared.
const (
	ocrMinConf     = 78.0
	ocrMinWordLike = 0.6
)

func wordLikeShare(ws []word) float64 {
	if len(ws) == 0 {
		return 0
	}
	n := 0
	for _, w := range ws {
		letters, digits := 0, 0
		for _, r := range w.text {
			switch {
			case unicode.IsLetter(r):
				letters++
			case unicode.IsDigit(r):
				digits++
			}
		}
		if letters >= 3 || (digits >= 1 && letters+digits == len([]rune(w.text))) {
			n++
		}
	}
	return float64(n) / float64(len(ws))
}

// ocrMaxSide: sheets larger than this (in points) are drawings.
const ocrMaxSide = 1500

// ExtractPagesOCR is ExtractPages plus recognition of the scanned pages
// (no text layer, not a large drawing sheet). It returns how many pages
// were recognized and how many scans were left out because of the limit.
func ExtractPagesOCR(ctx context.Context, docID string, data []byte, cacheDir string, o *OCR) (pages []Page, recognized, skipped int, err error) {
	pages, err = ExtractPages(docID, data, cacheDir)
	if err != nil || !o.Enabled() {
		return pages, 0, 0, err
	}
	limit := o.MaxPages
	if limit <= 0 {
		limit = 40
	}
	for i, p := range pages {
		if p.HasText || math.Max(p.W, p.H) > ocrMaxSide || p.W <= 0 || p.H <= 0 {
			continue
		}
		if recognized >= limit {
			skipped++
			continue
		}
		ws, conf, werr := o.words(ctx, data, p.Number, p.W, p.H)
		if werr != nil {
			return pages, recognized, skipped, werr
		}
		if len(ws) < ocrMinWords || conf < ocrMinConf || wordLikeShare(ws) < ocrMinWordLike {
			// a blank, graphic or badly recognized sheet: stays a drawing page; a sheet with many words that
			// were not trusted is said so, a sheet with a few labels is just a drawing
			if len(ws) >= ocrMinWords {
				pages[i].OCRReject, pages[i].OCRConf = true, conf
			}
			continue
		}
		np := buildPage(docID, p.Number, p.W, p.H, ws)
		if np.Drawing {
			continue // recognized labels of a drawing are not text to compare
		}
		np.OCR = true
		np.OCRConf = conf
		np.Raster = true
		for j := range np.Lines {
			np.Lines[j].OCR = true
		}
		pages[i] = np
		recognized++
	}
	return pages, recognized, skipped, nil
}
