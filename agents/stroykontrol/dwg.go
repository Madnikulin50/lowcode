package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DWG support: a DWG sheet is converted to PDF once, right where the agent
// downloads a document (ComposeClient.downloadOriginal/downloadDocument and
// the fixture store), and from then on it is a PDF everywhere — page images,
// text layer, title-block sheet numbers, deterministic comparison, OCR, the
// assistant. That is what keeps DWG on a par with PDF without a second
// pipeline to maintain.

// dwgConv is nil when conversion is disabled; DWG files then pass through
// untouched and the viewer reports them as unsupported ("other").
var dwgConv *dwgConverter

// IsDWG recognizes a DWG by its version magic ("AC1009" … "AC1032") and, for
// uploads whose content is not available yet, by the mimetype.
func IsDWG(mimetype string, data []byte) bool {
	if len(data) >= 6 && bytes.HasPrefix(data, []byte("AC10")) && data[4] >= '0' && data[4] <= '9' && data[5] >= '0' && data[5] <= '9' {
		return true
	}
	mt := strings.ToLower(mimetype)
	return strings.Contains(mt, "dwg") || strings.Contains(mt, "acad")
}

type dwgConverter struct {
	cmdTemplate string // command line with {in} and {out} placeholders
	cacheDir    string
	timeout     time.Duration
	mu          sync.Mutex // conversion is rare and heavy: one at a time
}

// NewDWGConverter returns nil for an empty command template.
func NewDWGConverter(cmdTemplate, cacheDir string) *dwgConverter {
	cmdTemplate = strings.TrimSpace(cmdTemplate)
	if cmdTemplate == "" {
		return nil
	}
	dir := filepath.Join(cacheDir, "dwg")
	_ = os.MkdirAll(dir, 0o755)
	return &dwgConverter{cmdTemplate: cmdTemplate, cacheDir: dir, timeout: 3 * time.Minute}
}

// available reports whether the converter binary can be started.
func (c *dwgConverter) available() bool {
	fields := strings.Fields(c.cmdTemplate)
	if len(fields) == 0 {
		return false
	}
	_, err := exec.LookPath(fields[0])
	return err == nil
}

// ToPDF converts DWG bytes to PDF bytes, caching by content hash.
func (c *dwgConverter) ToPDF(data []byte) ([]byte, error) {
	cachePath := filepath.Join(c.cacheDir, contentKey(data)+".pdf")
	if b, err := os.ReadFile(cachePath); err == nil && IsPDF("", b) {
		return b, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if b, err := os.ReadFile(cachePath); err == nil && IsPDF("", b) {
		return b, nil
	}

	tmp, err := os.MkdirTemp("", "stroykontrol-dwg-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	in, out := filepath.Join(tmp, "in.dwg"), filepath.Join(tmp, "out.pdf")
	if err := os.WriteFile(in, data, 0o644); err != nil {
		return nil, err
	}

	args := strings.Fields(c.cmdTemplate)
	for i, a := range args {
		a = strings.ReplaceAll(a, "{in}", in)
		args[i] = strings.ReplaceAll(a, "{out}", out)
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = tmp
	if msg, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("dwg converter %s: %w (%s)", args[0], err, truncate(msg, 400))
	}
	pdf, err := os.ReadFile(out)
	if err != nil {
		return nil, fmt.Errorf("dwg converter %s produced no output: %w", args[0], err)
	}
	if !IsPDF("", pdf) {
		return nil, fmt.Errorf("dwg converter %s output is not a PDF", args[0])
	}
	if err := os.WriteFile(cachePath, pdf, 0o644); err != nil {
		log.Printf("dwg: cache write: %v", err)
	}
	return pdf, nil
}

// normalizeDoc turns a downloaded DWG into a PDF; everything else is returned
// as is. A failed conversion is logged and the original is returned, so one
// bad drawing degrades to "unsupported" instead of failing the whole run.
func normalizeDoc(data []byte, mimetype string) ([]byte, string) {
	if dwgConv == nil || IsPDF(mimetype, data) || !IsDWG(mimetype, data) {
		return data, mimetype
	}
	pdf, err := dwgConv.ToPDF(data)
	if err != nil {
		log.Printf("dwg: %v", err)
		return data, mimetype
	}
	return pdf, "application/pdf"
}
