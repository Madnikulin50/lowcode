package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Rasterizer renders PDF pages to JPEG images via poppler's pdftoppm,
// caching output on disk keyed by content hash so the same file is never
// rasterized twice. Non-PDF documents (docx, …) are not supported here —
// there's no page-image concept to render without a PDF conversion step
// (no LibreOffice on this host); the viewer falls back to a text-only diff
// for those.
type Rasterizer struct {
	cacheDir string
	mu       sync.Mutex
	inFlight map[string]*sync.WaitGroup
}

func NewRasterizer(cacheDir string) *Rasterizer {
	_ = os.MkdirAll(cacheDir, 0o755)
	return &Rasterizer{cacheDir: cacheDir, inFlight: map[string]*sync.WaitGroup{}}
}

func contentKey(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])[:24]
}

// IsPDF checks the usual magic bytes / mimetype.
func IsPDF(mimetype string, data []byte) bool {
	if strings.Contains(strings.ToLower(mimetype), "pdf") {
		return true
	}
	return len(data) > 4 && string(data[:4]) == "%PDF"
}

// IsDOCX checks mimetype first, then (since Compose doesn't always know the
// mimetype of an uploaded file, and a renamed/octet-stream upload is common)
// falls back to sniffing the zip's central directory for word/document.xml —
// DOCX/XLSX/PPTX/ODT are all zip containers, so the mimetype alone isn't
// reliable proof either way.
func IsDOCX(mimetype string, data []byte) bool {
	if strings.Contains(strings.ToLower(mimetype), "wordprocessingml") {
		return true
	}
	if len(data) < 4 || data[0] != 'P' || data[1] != 'K' {
		return false
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return false
	}
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			return true
		}
	}
	return false
}

// Pages returns the cache directory holding page-N.jpg files for this PDF's
// content, rasterizing on first request. Safe for concurrent callers with
// the same content (de-duplicates via inFlight).
func (r *Rasterizer) Pages(data []byte) (dir string, pageCount int, err error) {
	key := contentKey(data)
	dir = filepath.Join(r.cacheDir, key)

	if n := r.countPages(dir); n > 0 {
		return dir, n, nil
	}

	r.mu.Lock()
	wg, running := r.inFlight[key]
	if !running {
		wg = &sync.WaitGroup{}
		wg.Add(1)
		r.inFlight[key] = wg
	}
	r.mu.Unlock()

	if running {
		wg.Wait()
		if n := r.countPages(dir); n > 0 {
			return dir, n, nil
		}
		return dir, 0, fmt.Errorf("rasterize: no pages produced for %s", key)
	}

	defer func() {
		r.mu.Lock()
		delete(r.inFlight, key)
		r.mu.Unlock()
		wg.Done()
	}()

	if err = r.render(data, dir); err != nil {
		return dir, 0, err
	}
	pageCount = r.countPages(dir)
	if pageCount == 0 {
		return dir, 0, fmt.Errorf("rasterize: pdftoppm produced no pages")
	}
	return dir, pageCount, nil
}

func (r *Rasterizer) render(data []byte, dir string) error {
	tmp, err := os.MkdirTemp("", "stroykontrol-raster-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	pdfPath := filepath.Join(tmp, "in.pdf")
	if err := os.WriteFile(pdfPath, data, 0o644); err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// -jpeg -r 100: matches the resolution already used elsewhere in this
	// project's docx/pdf tooling; good enough to spot real differences
	// without producing huge images.
	cmd := exec.Command("pdftoppm", "-jpeg", "-r", "100", pdfPath, filepath.Join(dir, "page"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pdftoppm: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (r *Rasterizer) countPages(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".jpg") {
			n++
		}
	}
	return n
}

// Text returns per-page extracted text for a PDF (used by the "text
// comparison" view mode), caching it alongside the rasterized page images
// under the same content-hash directory. Poppler's pdftotext -layout keeps
// column/table spacing reasonably intact and separates pages with form-feed
// characters by default.
func (r *Rasterizer) Text(data []byte) ([]string, error) {
	key := contentKey(data)
	dir := filepath.Join(r.cacheDir, key)
	cachePath := filepath.Join(dir, "text.json")

	if b, err := os.ReadFile(cachePath); err == nil {
		var pages []string
		if json.Unmarshal(b, &pages) == nil {
			return pages, nil
		}
	}

	tmp, err := os.MkdirTemp("", "stroykontrol-text-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)

	pdfPath := filepath.Join(tmp, "in.pdf")
	if err := os.WriteFile(pdfPath, data, 0o644); err != nil {
		return nil, err
	}

	cmd := exec.Command("pdftotext", "-layout", pdfPath, "-")
	out, err := cmd.Output()
	if err != nil {
		msg := ""
		if ee, ok := err.(*exec.ExitError); ok {
			msg = strings.TrimSpace(string(ee.Stderr))
		}
		return nil, fmt.Errorf("pdftotext: %w (%s)", err, msg)
	}

	pages := strings.Split(string(out), "\f")
	// pdftotext emits a trailing form feed after the last page.
	if len(pages) > 0 && strings.TrimSpace(pages[len(pages)-1]) == "" {
		pages = pages[:len(pages)-1]
	}

	if err := os.MkdirAll(dir, 0o755); err == nil {
		if b, err := json.Marshal(pages); err == nil {
			_ = os.WriteFile(cachePath, b, 0o644)
		}
	}
	return pages, nil
}

// PagePath resolves the on-disk path for a 1-indexed page number.
// pdftoppm zero-pads page numbers to the width of the page count.
func (r *Rasterizer) PagePath(dir string, pageCount, page int) (string, error) {
	width := len(fmt.Sprintf("%d", pageCount))
	if width < 2 {
		width = 2
	}
	name := fmt.Sprintf("page-%0*d.jpg", width, page)
	full := filepath.Join(dir, name)
	if _, err := os.Stat(full); err == nil {
		return full, nil
	}
	// pdftoppm drops padding entirely for single-page docs in some versions
	entries, _ := os.ReadDir(dir)
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".jpg") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	if page >= 1 && page <= len(names) {
		return filepath.Join(dir, names[page-1]), nil
	}
	return "", fmt.Errorf("page %d not found in %s", page, dir)
}
