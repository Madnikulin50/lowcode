package rag

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func parseDXF(data []byte) (*ParsedDocument, error) {
	if isBinaryDXF(data) {
		if conv, err := convertDWGToDXF(context.Background(), data, "in.dxf"); err == nil && len(conv) > 0 {
			return parseDXF(conv)
		}
		doc, err := harvestDocument(data, "dxf", true)
		if err != nil {
			return nil, err
		}
		doc.Partial = true
		return doc, nil
	}
	text := extractDXFText(string(data))
	return &ParsedDocument{Text: text, Title: extractTitle(text), Kind: "dxf"}, nil
}

func isBinaryDXF(data []byte) bool {
	return bytes.Contains(data[:min(len(data), 32)], []byte("AutoCAD Binary DXF"))
}

func extractDXFText(src string) string {
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	var (
		out     []string
		entity  string
		pending string
	)
	flushPending := func() {
		if pending != "" {
			out = append(out, pending)
			pending = ""
		}
	}
	for i := 0; i+1 < len(lines); i += 2 {
		code := strings.TrimSpace(lines[i])
		val := strings.TrimSpace(lines[i+1])
		if code == "0" {
			flushPending()
			entity = strings.ToUpper(val)
			continue
		}
		switch code {
		case "1", "3", "4":
			if val == "" {
				continue
			}
			if entity == "MTEXT" && pending != "" {
				pending += " " + val
			} else {
				flushPending()
				pending = val
			}
		case "2":
			if val != "" && (entity == "LAYER" || entity == "TABLE" || entity == "BLOCK" || entity == "STYLE") {
				out = append(out, entity+": "+val)
			}
		case "8":
			if val != "" && (entity == "TEXT" || entity == "MTEXT" || entity == "DIMENSION" || entity == "INSERT") {
				out = append(out, "LAYER "+val)
			}
		}
	}
	flushPending()
	return strings.Join(uniqueNonEmpty(out), "\n")
}

func parseIFC(data []byte) (*ParsedDocument, error) {
	src := string(data)
	if !bytes.Contains(data[:min(len(data), 64)], []byte("ISO-10303")) && !strings.Contains(strings.ToUpper(src[:min(len(src), 200)]), "ISO-10303") {
		if looksBinary(data) {
			return harvestDocument(data, "ifc", true)
		}
	}
	text := extractIFCText(src)
	return &ParsedDocument{Text: text, Title: extractTitle(text), Kind: "ifc"}, nil
}

var ifcLineRe = regexp.MustCompile(`(?i)#\d+\s*=\s*(IFC[A-Z0-9_]+)\s*\((.*)\)\s*;`)
var ifcStringRe = regexp.MustCompile(`'([^']*)'`)

func extractIFCText(src string) string {
	var out []string
	for _, m := range ifcLineRe.FindAllStringSubmatch(src, 8000) {
		ent := strings.ToUpper(m[1])
		switch ent {
		case "IFCPROJECT", "IFCSITE", "IFCBUILDING", "IFCBUILDINGSTOREY", "IFCSPACE",
			"IFCBUILDINGELEMENTPROXY", "IFCWALL", "IFCSLAB", "IFCDOOR", "IFCWINDOW",
			"IFCPROPERTYSINGLEVALUE", "IFCLABEL", "IFCTEXT", "IFCIDENTIFIER",
			"IFCORGANIZATION", "IFCPERSON", "IFCAPPLICATION":
		default:
			if !strings.Contains(ent, "PROPERTY") && !strings.Contains(ent, "ANNOTATION") {
				continue
			}
		}
		strs := ifcStringRe.FindAllStringSubmatch(m[2], 8)
		var parts []string
		for _, s := range strs {
			v := strings.TrimSpace(s[1])
			if v == "" || v == "$" {
				continue
			}
			parts = append(parts, v)
		}
		if len(parts) == 0 {
			continue
		}
		out = append(out, ent+": "+strings.Join(parts, " · "))
	}
	return strings.Join(uniqueNonEmpty(out), "\n")
}

func parseIFCZip(data []byte) (*ParsedDocument, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return harvestDocument(data, "ifczip", true)
	}
	for _, f := range zr.File {
		n := strings.ToLower(f.Name)
		if strings.HasSuffix(n, ".ifc") || strings.HasSuffix(n, ".ifcxml") {
			raw, err := readZipFile(f)
			if err != nil {
				continue
			}
			doc, err := parseIFC(raw)
			if err == nil && doc != nil {
				doc.Kind = "ifczip"
				return doc, nil
			}
		}
	}
	return harvestZipText(zr, "ifczip")
}

func parseBIMX(data []byte) (*ParsedDocument, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return harvestDocument(data, "bimx", true)
	}
	for _, f := range zr.File {
		n := strings.ToLower(f.Name)
		raw, err := readZipFile(f)
		if err != nil {
			continue
		}
		switch {
		case strings.HasSuffix(n, ".ifc") || strings.HasSuffix(n, ".ifcxml"):
			doc, err := parseIFC(raw)
			if err == nil && doc != nil && strings.TrimSpace(doc.Text) != "" {
				doc.Kind = "bimx"
				return doc, nil
			}
		case strings.HasSuffix(n, ".xml") || strings.HasSuffix(n, ".json") || strings.HasSuffix(n, ".txt"):
			if t := strings.TrimSpace(string(raw)); len(t) > 20 && !looksBinary(raw) {
				return &ParsedDocument{Text: t, Title: extractTitle(t), Kind: "bimx", Partial: true}, nil
			}
		}
	}
	return harvestZipText(zr, "bimx")
}

func harvestZipText(zr *zip.Reader, kind string) (*ParsedDocument, error) {
	var parts []string
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		raw, err := readZipFile(f)
		if err != nil {
			continue
		}
		doc, _ := harvestDocument(raw, kind, true)
		if doc != nil && strings.TrimSpace(doc.Text) != "" {
			parts = append(parts, doc.Text)
		}
	}
	text := strings.Join(parts, "\n")
	return &ParsedDocument{Text: text, Title: extractTitle(text), Kind: kind, Partial: true}, nil
}

func parseDWG(data []byte, filename string) (*ParsedDocument, error) {
	if conv, err := convertDWGToDXF(context.Background(), data, filename); err == nil && len(conv) > 0 {
		doc, err := parseDXF(conv)
		if err == nil && doc != nil {
			doc.Kind = "dwg"
			return doc, nil
		}
	}
	doc, err := harvestDocument(data, "dwg", true)
	if err != nil {
		return nil, err
	}
	if len(data) >= 6 && bytes.HasPrefix(data, []byte("AC10")) {
		ver := string(data[:6])
		doc.Text = strings.TrimSpace("DWG " + ver + "\n" + doc.Text)
	}
	doc.Partial = true
	return doc, nil
}

func parseArchiCAD(data []byte, kind string) (*ParsedDocument, error) {
	doc, err := harvestDocument(data, kind, true)
	if err != nil {
		return nil, err
	}
	doc.Partial = true
	if strings.TrimSpace(doc.Text) == "" {
		doc.Text = "ArchiCAD " + kind + ": native binary, no text. Upload IFC (or PDF/DXF) alongside for a full summary."
	}
	return doc, nil
}

func convertDWGToDXF(ctx context.Context, data []byte, filename string) ([]byte, error) {
	bin := dwgConverterBin()
	if bin == "" {
		return nil, fmt.Errorf("no dwg converter on PATH")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	cctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	dir, err := os.MkdirTemp("", "dwg2dxf_*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	inName := "in.dwg"
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".dxf" || ext == ".dwg" {
		inName = "in" + ext
	}
	inPath := filepath.Join(dir, inName)
	outPath := filepath.Join(dir, "out.dxf")
	if err := os.WriteFile(inPath, data, 0o600); err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(cctx, bin, "-o", outPath, inPath)
	if _, err := cmd.CombinedOutput(); err != nil {
		cmd = exec.CommandContext(cctx, bin, inPath)
		cmd.Dir = dir
		if _, err2 := cmd.CombinedOutput(); err2 != nil {
			return nil, err
		}
		alt := strings.TrimSuffix(inPath, filepath.Ext(inPath)) + ".dxf"
		if raw, err := os.ReadFile(alt); err == nil {
			return raw, nil
		}
		return nil, err
	}
	return os.ReadFile(outPath)
}

func dwgConverterBin() string {
	for _, key := range []string{"DWG2DXF", "DWG_CONVERTER"} {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	if p, err := exec.LookPath("dwg2dxf"); err == nil {
		return p
	}
	return ""
}

func uniqueNonEmpty(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
