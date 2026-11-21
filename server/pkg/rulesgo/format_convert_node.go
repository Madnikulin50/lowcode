// format.convert is a universal data-format converter node: JSON, XML, CSV
// and XLSX all get parsed down to the same canonical shape - a flat array
// of records ([]map[string]interface{}) - and re-serialized to whichever
// format is requested. That's the "мультиформатность обмена данными"
// (JSON/XML/Excel/CSV) piece from the platform review: one node instead of
// bespoke glue code per format pair.
//
// XLSX gets no external dependency, same call as pkg/rag/parser_xlsx.go
// elsewhere in this codebase: it's a zip of a few small XML parts, so
// archive/zip + encoding/xml is enough for both reading and writing it.
package rulesgo

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

type FormatConvertConfig struct {
	Input        string `json:"input"`        // template - the raw source data
	InputFormat  string `json:"inputFormat"`  // "json" | "xml" | "csv" | "xlsx"
	OutputFormat string `json:"outputFormat"` // "json" | "xml" | "csv" | "xlsx"

	// Input is base64-encoded before templating is resolved on it - needed
	// for binary xlsx input coming from e.g. an HTTP response body or file
	// attachment captured as a base64 string in the chain context.
	InputBase64 bool `json:"inputBase64,omitempty"`

	Delimiter string `json:"delimiter,omitempty"` // CSV, default ","
	RootTag   string `json:"rootTag,omitempty"`   // XML, default "root"
	RecordTag string `json:"recordTag,omitempty"` // XML, default "item"
	SheetName string `json:"sheetName,omitempty"` // XLSX, default "Sheet1"
}

type formatConvertExecutor struct{}

func (n *formatConvertExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[FormatConvertConfig](node.Config)
	if err != nil {
		return nil, err
	}
	cfg.Input = resolveTemplateValue(cfg.Input, ec)
	cfg.InputFormat = strings.ToLower(strings.TrimSpace(resolveTemplateValue(cfg.InputFormat, ec)))
	cfg.OutputFormat = strings.ToLower(strings.TrimSpace(resolveTemplateValue(cfg.OutputFormat, ec)))

	if cfg.InputFormat == "" {
		return nil, fmt.Errorf("format.convert: inputFormat is required")
	}
	if cfg.OutputFormat == "" {
		return nil, fmt.Errorf("format.convert: outputFormat is required")
	}

	raw := []byte(cfg.Input)
	if cfg.InputBase64 {
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(cfg.Input))
		if err != nil {
			return nil, fmt.Errorf("format.convert: input is not valid base64: %w", err)
		}
		raw = decoded
	}

	delimiter := ','
	if cfg.Delimiter != "" {
		delimiter = []rune(cfg.Delimiter)[0]
	}
	rootTag := cfg.RootTag
	if rootTag == "" {
		rootTag = "root"
	}
	recordTag := cfg.RecordTag
	if recordTag == "" {
		recordTag = "item"
	}
	sheetName := cfg.SheetName
	if sheetName == "" {
		sheetName = "Sheet1"
	}

	var records []map[string]interface{}
	switch cfg.InputFormat {
	case "json":
		records, err = parseJSONRecords(raw)
	case "csv":
		records, err = parseCSVRecords(raw, delimiter)
	case "xml":
		records, err = parseXMLRecords(raw, recordTag)
	case "xlsx":
		records, err = parseXLSXRecords(raw, sheetName)
	default:
		return nil, fmt.Errorf("format.convert: unsupported inputFormat %q", cfg.InputFormat)
	}
	if err != nil {
		return nil, fmt.Errorf("format.convert: parse %s: %w", cfg.InputFormat, err)
	}

	var out []byte
	binary := false
	switch cfg.OutputFormat {
	case "json":
		out, err = writeJSONRecords(records)
	case "csv":
		out, err = writeCSVRecords(records, delimiter)
	case "xml":
		out, err = writeXMLRecords(records, rootTag, recordTag)
	case "xlsx":
		out, err = writeXLSXRecords(records, sheetName)
		binary = true
	default:
		return nil, fmt.Errorf("format.convert: unsupported outputFormat %q", cfg.OutputFormat)
	}
	if err != nil {
		return nil, fmt.Errorf("format.convert: write %s: %w", cfg.OutputFormat, err)
	}

	result := map[string]interface{}{
		"count":   len(records),
		"records": records,
	}
	if binary {
		result["output"] = base64.StdEncoding.EncodeToString(out)
		result["outputBase64"] = true
	} else {
		result["output"] = string(out)
		result["outputBase64"] = false
	}
	return result, nil
}

// --- canonical record helpers ---

// recordKeys returns the union of all keys across records, sorted for a
// deterministic column/field order (JSON preserves per-record shape, but
// CSV/XML need one stable order across all rows).
func recordKeys(records []map[string]interface{}) []string {
	seen := map[string]bool{}
	for _, rec := range records {
		for k := range rec {
			seen[k] = true
		}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func stringifyValue(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// --- JSON ---

func parseJSONRecords(raw []byte) ([]map[string]interface{}, error) {
	var arr []map[string]interface{}
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, nil
	}
	var single map[string]interface{}
	if err := json.Unmarshal(raw, &single); err != nil {
		return nil, fmt.Errorf("expected a JSON array of objects or one object: %w", err)
	}
	return []map[string]interface{}{single}, nil
}

func writeJSONRecords(records []map[string]interface{}) ([]byte, error) {
	if records == nil {
		records = []map[string]interface{}{}
	}
	return json.Marshal(records)
}

// --- CSV ---

func parseCSVRecords(raw []byte, delimiter rune) ([]map[string]interface{}, error) {
	r := csv.NewReader(bytes.NewReader(raw))
	r.Comma = delimiter
	r.FieldsPerRecord = -1 // tolerate ragged rows instead of failing outright

	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	header := rows[0]
	records := make([]map[string]interface{}, 0, len(rows)-1)
	for _, row := range rows[1:] {
		rec := make(map[string]interface{}, len(header))
		for i, col := range header {
			if i < len(row) {
				rec[col] = row[i]
			} else {
				rec[col] = ""
			}
		}
		records = append(records, rec)
	}
	return records, nil
}

func writeCSVRecords(records []map[string]interface{}, delimiter rune) ([]byte, error) {
	keys := recordKeys(records)

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Comma = delimiter

	if err := w.Write(keys); err != nil {
		return nil, err
	}
	for _, rec := range records {
		row := make([]string, len(keys))
		for i, k := range keys {
			row[i] = stringifyValue(rec[k])
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// --- XML ---
//
// Records are matched by tag name (recordTag) anywhere in the document, one
// level of children become fields (nested elements aren't recursed into -
// this is a converter for flat business records, not arbitrary trees), and
// attributes on the record element become "@name" fields.

type xmlNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content string     `xml:",chardata"`
	Nodes   []xmlNode  `xml:",any"`
}

func parseXMLRecords(raw []byte, recordTag string) ([]map[string]interface{}, error) {
	var root xmlNode
	if err := xml.Unmarshal(raw, &root); err != nil {
		return nil, err
	}

	var records []map[string]interface{}
	var walk func(n xmlNode)
	walk = func(n xmlNode) {
		if n.XMLName.Local == recordTag {
			rec := make(map[string]interface{}, len(n.Attrs)+len(n.Nodes))
			for _, a := range n.Attrs {
				rec["@"+a.Name.Local] = a.Value
			}
			for _, child := range n.Nodes {
				rec[child.XMLName.Local] = strings.TrimSpace(child.Content)
			}
			records = append(records, rec)
		}
		for _, child := range n.Nodes {
			walk(child)
		}
	}
	walk(root)
	return records, nil
}

func writeXMLRecords(records []map[string]interface{}, rootTag, recordTag string) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	fmt.Fprintf(&buf, "<%s>", rootTag)

	keys := recordKeys(records)
	for _, rec := range records {
		fmt.Fprintf(&buf, "<%s>", recordTag)
		for _, k := range keys {
			buf.WriteString("<")
			buf.WriteString(k)
			buf.WriteString(">")
			xml.EscapeText(&buf, []byte(stringifyValue(rec[k])))
			buf.WriteString("</")
			buf.WriteString(k)
			buf.WriteString(">")
		}
		fmt.Fprintf(&buf, "</%s>", recordTag)
	}
	fmt.Fprintf(&buf, "</%s>", rootTag)
	return buf.Bytes(), nil
}

// --- XLSX ---
//
// A minimal but valid .xlsx: every cell is written as an inline string
// (t="inlineStr"), which sidesteps needing a shared-strings table and keeps
// both the writer and the reader simple. Excel/LibreOffice/Google Sheets all
// read this fine; they also auto-detect numeric-looking inline strings on
// open, so round-tripping through spreadsheet software isn't lossy for the
// common case.

func writeXLSXRecords(records []map[string]interface{}, sheetName string) ([]byte, error) {
	keys := recordKeys(records)

	var sheet bytes.Buffer
	sheet.WriteString(xml.Header)
	sheet.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)

	writeRow := func(rowNum int, values []string) {
		fmt.Fprintf(&sheet, `<row r="%d">`, rowNum)
		for i, v := range values {
			ref := columnLetter(i+1) + strconv.Itoa(rowNum)
			fmt.Fprintf(&sheet, `<c r="%s" t="inlineStr"><is><t>`, ref)
			xml.EscapeText(&sheet, []byte(v))
			sheet.WriteString(`</t></is></c>`)
		}
		sheet.WriteString(`</row>`)
	}

	writeRow(1, keys)
	for i, rec := range records {
		row := make([]string, len(keys))
		for j, k := range keys {
			row[j] = stringifyValue(rec[k])
		}
		writeRow(i+2, row)
	}
	sheet.WriteString(`</sheetData></worksheet>`)

	var zbuf bytes.Buffer
	zw := zip.NewWriter(&zbuf)

	files := map[string]string{
		"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
</Types>`,
		"_rels/.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`,
		"xl/workbook.xml": fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name=%q sheetId="1" r:id="rId1"/>
  </sheets>
</workbook>`, sheetName),
		"xl/_rels/workbook.xml.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
</Relationships>`,
		"xl/worksheets/sheet1.xml": sheet.String(),
	}

	// deterministic order (map iteration would make the zip's own bytes
	// vary run to run, which is a nuisance to diff/test against)
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		fw, err := zw.Create(name)
		if err != nil {
			return nil, err
		}
		if _, err := fw.Write([]byte(files[name])); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return zbuf.Bytes(), nil
}

// columnLetter converts a 1-based column index to its spreadsheet letter
// (1 -> A, 26 -> Z, 27 -> AA, ...).
func columnLetter(col int) string {
	var letters []byte
	for col > 0 {
		col--
		letters = append([]byte{byte('A' + col%26)}, letters...)
		col /= 26
	}
	return string(letters)
}

func parseXLSXRecords(raw []byte, sheetName string) ([]map[string]interface{}, error) {
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return nil, fmt.Errorf("not a valid xlsx (zip): %w", err)
	}

	shared, err := readXLSXSharedStrings(zr)
	if err != nil {
		return nil, err
	}

	sheetFile, err := findXLSXSheet(zr, sheetName)
	if err != nil {
		return nil, err
	}
	sheetXML, err := readZipFile(sheetFile)
	if err != nil {
		return nil, err
	}

	rows, err := parseXLSXSheetRows(sheetXML, shared)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	header := rows[0]
	records := make([]map[string]interface{}, 0, len(rows)-1)
	for _, row := range rows[1:] {
		rec := make(map[string]interface{}, len(header))
		for i, col := range header {
			if i < len(row) {
				rec[col] = row[i]
			} else {
				rec[col] = ""
			}
		}
		records = append(records, rec)
	}
	return records, nil
}

// findXLSXSheet picks sheet1.xml by convention (this converter always
// writes exactly one sheet); sheetName is accepted for symmetry with the
// writer but every worksheet part is a candidate if there's only one.
func findXLSXSheet(zr *zip.Reader, sheetName string) (*zip.File, error) {
	var sheets []*zip.File
	for _, f := range zr.File {
		name := strings.ToLower(f.Name)
		if strings.HasPrefix(name, "xl/worksheets/") && strings.HasSuffix(name, ".xml") {
			sheets = append(sheets, f)
		}
	}
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no worksheet found in xlsx")
	}
	sort.Slice(sheets, func(i, j int) bool { return sheets[i].Name < sheets[j].Name })
	return sheets[0], nil
}

func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

type xlsxSST struct {
	XMLName xml.Name      `xml:"sst"`
	SI      []xlsxSSTItem `xml:"si"`
}
type xlsxSSTItem struct {
	T string `xml:"t"`
}

func readXLSXSharedStrings(zr *zip.Reader) ([]string, error) {
	for _, f := range zr.File {
		if strings.EqualFold(f.Name, "xl/sharedStrings.xml") {
			raw, err := readZipFile(f)
			if err != nil {
				return nil, err
			}
			var sst xlsxSST
			if err := xml.Unmarshal(raw, &sst); err != nil {
				return nil, fmt.Errorf("parse sharedStrings.xml: %w", err)
			}
			out := make([]string, len(sst.SI))
			for i, si := range sst.SI {
				out[i] = si.T
			}
			return out, nil
		}
	}
	return nil, nil // valid: a workbook using only inline strings has none
}

type xlsxWorksheet struct {
	SheetData struct {
		Rows []struct {
			Cells []struct {
				R  string `xml:"r,attr"` // cell ref, e.g. "B3"
				T  string `xml:"t,attr"` // type: "s" (shared), "inlineStr", "" (number)
				V  string `xml:"v"`
				Is struct {
					T string `xml:"t"`
				} `xml:"is"`
			} `xml:"c"`
		} `xml:"row"`
	} `xml:"sheetData"`
}

func parseXLSXSheetRows(sheetXML []byte, shared []string) ([][]string, error) {
	var ws xlsxWorksheet
	if err := xml.Unmarshal(sheetXML, &ws); err != nil {
		return nil, fmt.Errorf("parse worksheet xml: %w", err)
	}

	rows := make([][]string, 0, len(ws.SheetData.Rows))
	for _, row := range ws.SheetData.Rows {
		var cells []string
		for _, c := range row.Cells {
			colIdx := columnIndexFromRef(c.R)
			for len(cells) <= colIdx {
				cells = append(cells, "")
			}
			switch c.T {
			case "s":
				if idx, err := strconv.Atoi(c.V); err == nil && idx >= 0 && idx < len(shared) {
					cells[colIdx] = shared[idx]
				}
			case "inlineStr":
				cells[colIdx] = c.Is.T
			default:
				cells[colIdx] = c.V
			}
		}
		rows = append(rows, cells)
	}
	return rows, nil
}

// columnIndexFromRef extracts the 0-based column index from a cell
// reference like "C7" (-> 2). Falls back to 0 for anything unparsable.
func columnIndexFromRef(ref string) int {
	idx := 0
	for _, r := range ref {
		if r < 'A' || r > 'Z' {
			break
		}
		idx = idx*26 + int(r-'A'+1)
	}
	if idx == 0 {
		return 0
	}
	return idx - 1
}
