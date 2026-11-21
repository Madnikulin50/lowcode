package rulesgo

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func runFormatConvert(t *testing.T, cfg map[string]interface{}) map[string]interface{} {
	t.Helper()
	raw, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	n := &formatConvertExecutor{}
	out, err := n.Execute(context.Background(), ChainNode{Config: raw}, &ExecutionContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return out
}

func TestFormatConvert_JSONToCSV(t *testing.T) {
	out := runFormatConvert(t, map[string]interface{}{
		"input":        `[{"name":"Товар А","qty":10},{"name":"Товар Б","qty":5}]`,
		"inputFormat":  "json",
		"outputFormat": "csv",
	})
	csvOut, _ := out["output"].(string)
	lines := strings.Split(strings.TrimRight(csvOut, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected header + 2 rows, got %d lines: %q", len(lines), csvOut)
	}
	if lines[0] != "name,qty" {
		t.Fatalf("unexpected header: %q", lines[0])
	}
	if lines[1] != "Товар А,10" || lines[2] != "Товар Б,5" {
		t.Fatalf("unexpected rows: %q / %q", lines[1], lines[2])
	}
	if out["count"] != 2 {
		t.Fatalf("expected count=2, got %#v", out["count"])
	}
}

func TestFormatConvert_CSVToJSON(t *testing.T) {
	out := runFormatConvert(t, map[string]interface{}{
		"input":        "sku,name,price\nA1,Widget,9.99\nA2,Gadget,19.99\n",
		"inputFormat":  "csv",
		"outputFormat": "json",
	})
	var records []map[string]interface{}
	if err := json.Unmarshal([]byte(out["output"].(string)), &records); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	if records[0]["sku"] != "A1" || records[0]["price"] != "9.99" {
		t.Fatalf("unexpected first record: %#v", records[0])
	}
}

func TestFormatConvert_CSVCustomDelimiter(t *testing.T) {
	out := runFormatConvert(t, map[string]interface{}{
		"input":        "a;b\n1;2\n",
		"inputFormat":  "csv",
		"outputFormat": "json",
		"delimiter":    ";",
	})
	var records []map[string]interface{}
	json.Unmarshal([]byte(out["output"].(string)), &records)
	if len(records) != 1 || records[0]["a"] != "1" || records[0]["b"] != "2" {
		t.Fatalf("unexpected records: %#v", records)
	}
}

func TestFormatConvert_JSONToXML(t *testing.T) {
	out := runFormatConvert(t, map[string]interface{}{
		"input":        `[{"code":"1","name":"A & B"}]`,
		"inputFormat":  "json",
		"outputFormat": "xml",
		"rootTag":      "orders",
		"recordTag":    "order",
	})
	xmlOut, _ := out["output"].(string)
	if !strings.Contains(xmlOut, "<orders>") || !strings.Contains(xmlOut, "<order>") {
		t.Fatalf("expected custom root/record tags, got: %s", xmlOut)
	}
	if !strings.Contains(xmlOut, "<code>1</code>") {
		t.Fatalf("expected code field, got: %s", xmlOut)
	}
	if !strings.Contains(xmlOut, "A &amp; B") {
		t.Fatalf("expected XML-escaped ampersand, got: %s", xmlOut)
	}
}

func TestFormatConvert_XMLToJSON_WithAttributes(t *testing.T) {
	out := runFormatConvert(t, map[string]interface{}{
		"input": `<catalog>
			<item id="42"><name>Товар</name><price>100</price></item>
			<item id="43"><name>Другой</name><price>200</price></item>
		</catalog>`,
		"inputFormat":  "xml",
		"outputFormat": "json",
		"recordTag":    "item",
	})
	var records []map[string]interface{}
	if err := json.Unmarshal([]byte(out["output"].(string)), &records); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d: %#v", len(records), records)
	}
	if records[0]["@id"] != "42" || records[0]["name"] != "Товар" || records[0]["price"] != "100" {
		t.Fatalf("unexpected first record: %#v", records[0])
	}
}

func TestFormatConvert_XLSXRoundTrip(t *testing.T) {
	// write records to xlsx bytes ourselves (base64 output), then feed that
	// straight back in as xlsx input - proves the reader understands
	// exactly what the writer produces without needing a real Excel file.
	writeOut := runFormatConvert(t, map[string]interface{}{
		"input":        `[{"Код":"1","Наименование":"Товар А"},{"Код":"2","Наименование":"Товар Б"}]`,
		"inputFormat":  "json",
		"outputFormat": "xlsx",
		"sheetName":    "Номенклатура",
	})
	xlsxB64, ok := writeOut["output"].(string)
	if !ok || xlsxB64 == "" {
		t.Fatalf("expected base64 xlsx output, got %#v", writeOut["output"])
	}
	if writeOut["outputBase64"] != true {
		t.Fatalf("expected outputBase64=true for xlsx, got %#v", writeOut["outputBase64"])
	}
	if _, err := base64.StdEncoding.DecodeString(xlsxB64); err != nil {
		t.Fatalf("output is not valid base64: %v", err)
	}

	readOut := runFormatConvert(t, map[string]interface{}{
		"input":        xlsxB64,
		"inputFormat":  "xlsx",
		"inputBase64":  true,
		"outputFormat": "json",
		"sheetName":    "Номенклатура",
	})
	var records []map[string]interface{}
	if err := json.Unmarshal([]byte(readOut["output"].(string)), &records); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records round-tripped through xlsx, got %d: %#v", len(records), records)
	}
	if records[0]["Код"] != "1" || records[0]["Наименование"] != "Товар А" {
		t.Fatalf("unexpected first record: %#v", records[0])
	}
	if records[1]["Код"] != "2" || records[1]["Наименование"] != "Товар Б" {
		t.Fatalf("unexpected second record: %#v", records[1])
	}
}

func TestFormatConvert_RequiresFormats(t *testing.T) {
	n := &formatConvertExecutor{}
	if _, err := n.Execute(context.Background(), ChainNode{Config: []byte(`{"input":"x","outputFormat":"json"}`)}, &ExecutionContext{}); err == nil {
		t.Fatal("expected error for missing inputFormat")
	}
	if _, err := n.Execute(context.Background(), ChainNode{Config: []byte(`{"input":"x","inputFormat":"json"}`)}, &ExecutionContext{}); err == nil {
		t.Fatal("expected error for missing outputFormat")
	}
	if _, err := n.Execute(context.Background(), ChainNode{Config: []byte(`{"input":"x","inputFormat":"yaml","outputFormat":"json"}`)}, &ExecutionContext{}); err == nil {
		t.Fatal("expected error for unsupported inputFormat")
	}
}

func TestFormatConvert_InvalidBase64Errors(t *testing.T) {
	n := &formatConvertExecutor{}
	_, err := n.Execute(context.Background(), ChainNode{Config: []byte(`{
		"input": "not-valid-base64!!!",
		"inputFormat": "xlsx",
		"inputBase64": true,
		"outputFormat": "json"
	}`)}, &ExecutionContext{})
	if err == nil {
		t.Fatal("expected error for invalid base64 input")
	}
}

func TestColumnLetter(t *testing.T) {
	cases := map[int]string{1: "A", 2: "B", 26: "Z", 27: "AA", 28: "AB", 52: "AZ", 53: "BA"}
	for col, want := range cases {
		if got := columnLetter(col); got != want {
			t.Errorf("columnLetter(%d) = %q, want %q", col, got, want)
		}
	}
}

func TestColumnIndexFromRef(t *testing.T) {
	cases := map[string]int{"A1": 0, "B3": 1, "Z10": 25, "AA1": 26, "AB5": 27}
	for ref, want := range cases {
		if got := columnIndexFromRef(ref); got != want {
			t.Errorf("columnIndexFromRef(%q) = %d, want %d", ref, got, want)
		}
	}
}
