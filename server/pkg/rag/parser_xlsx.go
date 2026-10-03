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
	files := map[string]*zip.File{}
	for _, f := range zr.File {
		files[strings.ToLower(f.Name)] = f
	}
	var sheets []string
	for _, ref := range xlsxSheetOrder(zr, files) {
		f := files[strings.ToLower(ref.path)]
		if f == nil {
			continue
		}
		raw, err := readZipFile(f)
		if err != nil {
			continue
		}
		text := strings.TrimSpace(extractXLSXSheet(raw, shared))
		if text == "" {
			continue
		}
		// Sheet header keeps discrepancies attributable to a sheet; it is a
		// line of its own so row lines still match across renamed sheets.
		if ref.name != "" {
			text = "Лист: " + ref.name + "\n" + text
		}
		sheets = append(sheets, text)
	}
	joined := strings.Join(sheets, "\n\n")
	return &ParsedDocument{Text: joined, Title: extractTitle(joined), Kind: "xlsx"}, nil
}

type xlsxSheetRef struct {
	name string
	path string
}

type xlsxWorkbook struct {
	Sheets []struct {
		Name string `xml:"name,attr"`
		RID  string `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
	} `xml:"sheets>sheet"`
}

type xlsxRels struct {
	Rels []struct {
		ID     string `xml:"Id,attr"`
		Target string `xml:"Target,attr"`
	} `xml:"Relationship"`
}

// xlsxSheetOrder lists worksheets in workbook (tab) order with their names.
// Falls back to the archive's own order when workbook.xml/rels are missing
// or unreadable, so a slightly malformed file still yields its text.
func xlsxSheetOrder(zr *zip.Reader, files map[string]*zip.File) []xlsxSheetRef {
	var out []xlsxSheetRef
	wbf, relf := files["xl/workbook.xml"], files["xl/_rels/workbook.xml.rels"]
	if wbf != nil && relf != nil {
		var wb xlsxWorkbook
		var rels xlsxRels
		wbRaw, err1 := readZipFile(wbf)
		relRaw, err2 := readZipFile(relf)
		if err1 == nil && err2 == nil && xml.Unmarshal(wbRaw, &wb) == nil && xml.Unmarshal(relRaw, &rels) == nil {
			targets := map[string]string{}
			for _, r := range rels.Rels {
				t := strings.TrimPrefix(r.Target, "/")
				if !strings.HasPrefix(t, "xl/") {
					t = "xl/" + t
				}
				targets[r.ID] = t
			}
			for _, sh := range wb.Sheets {
				if t, ok := targets[sh.RID]; ok && strings.Contains(strings.ToLower(t), "worksheets/") {
					out = append(out, xlsxSheetRef{name: sh.Name, path: t})
				}
			}
		}
	}
	if len(out) > 0 {
		return out
	}
	for _, f := range zr.File {
		name := strings.ToLower(f.Name)
		if strings.HasPrefix(name, "xl/worksheets/sheet") && strings.HasSuffix(name, ".xml") {
			out = append(out, xlsxSheetRef{path: f.Name})
		}
	}
	return out
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
	T      string      `xml:"t"`
	Richer []xlsxSIRun `xml:"r"`
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
