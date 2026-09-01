package rag

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

func parseXLSX(data []byte) (*ParsedDocument, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("parse xlsx: %w", err)
	}
	shared := readXLSXSharedStrings(zr)
	var sheets []string
	for _, f := range zr.File {
		name := strings.ToLower(f.Name)
		if !strings.HasPrefix(name, "xl/worksheets/sheet") || !strings.HasSuffix(name, ".xml") {
			continue
		}
		raw, err := readZipFile(f)
		if err != nil {
			continue
		}
		text := extractXLSXSheet(raw, shared)
		if strings.TrimSpace(text) == "" {
			continue
		}
		sheets = append(sheets, strings.TrimSpace(text))
	}
	joined := strings.Join(sheets, "\n\n")
	return &ParsedDocument{Text: joined, Title: extractTitle(joined), Kind: "xlsx"}, nil
}

func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(io.LimitReader(rc, 32<<20))
}

type xlsxSST struct {
	Items []xlsxSI `xml:"si"`
}

type xlsxSI struct {
	T      string       `xml:"t"`
	Richer []xlsxSIRun  `xml:"r"`
}

type xlsxSIRun struct {
	T string `xml:"t"`
}

func readXLSXSharedStrings(zr *zip.Reader) []string {
	for _, f := range zr.File {
		if strings.ToLower(f.Name) != "xl/sharedstrings.xml" {
			continue
		}
		raw, err := readZipFile(f)
		if err != nil {
			return nil
		}
		var sst xlsxSST
		if xml.Unmarshal(raw, &sst) != nil {
			return xlsxSharedFallback(raw)
		}
		out := make([]string, 0, len(sst.Items))
		for _, si := range sst.Items {
			if si.T != "" {
				out = append(out, si.T)
				continue
			}
			var b strings.Builder
			for _, r := range si.Richer {
				b.WriteString(r.T)
			}
			out = append(out, b.String())
		}
		return out
	}
	return nil
}

var xlsxTRe = regexp.MustCompile(`<t[^>]*>([^<]*)</t>`)

func xlsxSharedFallback(raw []byte) []string {
	matches := xlsxTRe.FindAllSubmatch(raw, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 {
			out = append(out, string(m[1]))
		}
	}
	return out
}

type xlsxSheet struct {
	Rows []xlsxRow `xml:"sheetData>row"`
}

type xlsxRow struct {
	Cells []xlsxCell `xml:"c"`
}

type xlsxCell struct {
	T  string `xml:"t,attr"`
	V  string `xml:"v"`
	IS *struct {
		T string `xml:"t"`
	} `xml:"is"`
}

func extractXLSXSheet(raw []byte, shared []string) string {
	var sheet xlsxSheet
	if xml.Unmarshal(raw, &sheet) != nil {
		return strings.Join(xlsxSharedFallback(raw), " ")
	}
	var lines []string
	for _, row := range sheet.Rows {
		var cells []string
		for _, c := range row.Cells {
			val := xlsxCellText(c, shared)
			if val != "" {
				cells = append(cells, val)
			}
		}
		if len(cells) > 0 {
			lines = append(lines, strings.Join(cells, "\t"))
		}
	}
	return strings.Join(lines, "\n")
}

func xlsxCellText(c xlsxCell, shared []string) string {
	if c.IS != nil && c.IS.T != "" {
		return strings.TrimSpace(c.IS.T)
	}
	v := strings.TrimSpace(c.V)
	if v == "" {
		return ""
	}
	if c.T == "s" {
		i, err := strconv.Atoi(v)
		if err == nil && i >= 0 && i < len(shared) {
			return strings.TrimSpace(shared[i])
		}
	}
	if c.T == "inlineStr" {
		return v
	}
	return v
}
