package idcheck

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// PageImages rasterizes a PDF with pdftoppm at the given DPI, caching the
// JPEGs on disk by content hash + DPI. Returns page image paths in order.
func PageImages(ctx context.Context, cacheDir string, data []byte, dpi int) ([]string, error) {
	sum := sha256.Sum256(data)
	dir := filepath.Join(cacheDir, hex.EncodeToString(sum[:])[:24], fmt.Sprintf("r%d", dpi))
	if pages := listJPEG(dir); len(pages) > 0 {
		return pages, nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	src := filepath.Join(dir, "src.pdf")
	if err := os.WriteFile(src, data, 0o644); err != nil {
		return nil, err
	}
	defer os.Remove(src)
	cmd := exec.CommandContext(ctx, "pdftoppm", "-r", fmt.Sprint(dpi), "-jpeg", "-jpegopt", "quality=85", src, filepath.Join(dir, "p"))
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pdftoppm: %v: %s", err, out)
	}
	pages := listJPEG(dir)
	if len(pages) == 0 {
		return nil, fmt.Errorf("pdftoppm produced no pages")
	}
	return pages, nil
}

var pageFileRe = regexp.MustCompile(`^p-\d+\.jpg$`)

// listJPEG lists pdftoppm's page images only ("p-07.jpg") — derived files
// such as the title-block crop ("p-03-stamp.jpg") live in the same folder.
func listJPEG(dir string) []string {
	m, _ := filepath.Glob(filepath.Join(dir, "p-*.jpg"))
	out := m[:0]
	for _, p := range m {
		if pageFileRe.MatchString(filepath.Base(p)) {
			out = append(out, p)
		}
	}
	sort.Strings(out) // pdftoppm zero-pads page numbers to the same width
	return out
}

// Tesseract runs the tesseract CLI. A zero value (Bin == "") means OCR is
// unavailable; extraction then relies on the vision model alone.
type Tesseract struct {
	Bin  string
	Lang string
}

// NewTesseract returns a usable Tesseract if the binary and language data
// are installed, otherwise a zero value.
func NewTesseract(bin, lang string) Tesseract {
	if bin == "" {
		bin = "tesseract"
	}
	path, err := exec.LookPath(bin)
	if err != nil {
		return Tesseract{}
	}
	out, _ := exec.Command(path, "--list-langs").CombinedOutput()
	if !strings.Contains(string(out), lang) {
		return Tesseract{}
	}
	return Tesseract{Bin: path, Lang: lang}
}

// Available reports whether OCR can run.
func (t Tesseract) Available() bool { return t.Bin != "" }

// Text OCRs one page image; the result is cached next to the image.
func (t Tesseract) Text(ctx context.Context, img string) (string, error) {
	if !t.Available() {
		return "", nil
	}
	cache := strings.TrimSuffix(img, ".jpg") + ".txt"
	if b, err := os.ReadFile(cache); err == nil {
		return string(b), nil
	}
	out, err := exec.CommandContext(ctx, t.Bin, img, "stdout", "-l", t.Lang, "--psm", "6").Output()
	if err != nil {
		return "", fmt.Errorf("tesseract: %w", err)
	}
	_ = os.WriteFile(cache, out, 0o644)
	return string(out), nil
}

// pageKeywords maps page kinds to lowercase markers, in priority order: the
// act's own pages mention schemes/certificates/letters in п.4 and the
// appendix list, so act markers are tested first.
var pageKeywords = []struct {
	kind PageKind
	keys []string
}{
	{PageAct, []string{"объект капитального строительства", "работы выполнены по проектной", "при выполнении работ применены", "освидетельствования скрытых работ", "освидетельствования ответственных конструкций", "освидетельствования участков сетей", "акт освидетельствования"}},
	{PageScheme, []string{"исполнительная схема", "исполнительный чертеж", "исполнительная съемка", "исполнительная геодезическая"}},
	{PageRegister, []string{"реестр документов", "реестр исполнительной", "реестр сертификатов", "реестр документов о качестве"}},
	{PageProtocol, []string{"протокол"}},
	{PageCertificate, []string{"сертификат", "паспорт", "декларация о соответствии", "документ о качестве"}},
	{PageLetter, []string{"исх.", "исх №", "уважаем", "письмо"}},
}

// ClassifyText classifies a page from its OCR text.
func ClassifyText(text string) PageKind {
	t := strings.ToLower(text)
	for _, pk := range pageKeywords {
		for _, k := range pk.keys {
			if strings.Contains(t, k) {
				return pk.kind
			}
		}
	}
	return PageOther
}
