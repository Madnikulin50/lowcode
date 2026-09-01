package rag

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

func TestParseDocument_DocxAndPDFKinds(t *testing.T) {
	docx := mustZip(t, map[string]string{
		"word/document.xml": `<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>Договор подряда</w:t></w:r></w:p></w:body></w:document>`,
	})
	doc, err := ParseDocument(docx, "a.docx", "")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Kind != "docx" || !strings.Contains(doc.Text, "Договор") {
		t.Fatalf("docx: %+v", doc)
	}
}

func TestParseDocument_XLSX(t *testing.T) {
	raw := mustZip(t, map[string]string{
		"xl/sharedStrings.xml": `<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><si><t>Смета</t></si><si><t>1200000</t></si></sst>`,
		"xl/worksheets/sheet1.xml": `<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row><c t="s"><v>0</v></c><c t="s"><v>1</v></c></row></sheetData></worksheet>`,
	})
	doc, err := ParseDocument(raw, "budget.xlsx", "")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Kind != "xlsx" {
		t.Fatalf("kind %s", doc.Kind)
	}
	if !strings.Contains(doc.Text, "Смета") || !strings.Contains(doc.Text, "1200000") {
		t.Fatalf("xlsx text: %q", doc.Text)
	}
}

func TestParseDocument_DXF(t *testing.T) {
	src := "  0\nSECTION\n  2\nENTITIES\n  0\nTEXT\n  8\nSTAMP\n  1\nКорпус А\n  0\nMTEXT\n  1\nЛист 1\n  0\nENDSEC\n  0\nEOF\n"
	doc, err := ParseDocument([]byte(src), "plan.dxf", "")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Kind != "dxf" {
		t.Fatalf("kind %s", doc.Kind)
	}
	if !strings.Contains(doc.Text, "Корпус А") || !strings.Contains(doc.Text, "Лист 1") {
		t.Fatalf("dxf text: %q", doc.Text)
	}
}

func TestParseDocument_IFC(t *testing.T) {
	src := `ISO-10303-21;
HEADER;
ENDSEC;
DATA;
#1=IFCPROJECT('2b$id',$,'ЖК Север',$,$,$,$,$,$);
#2=IFCBUILDINGSTOREY('x',$,'Этаж 3',$,$,$,$,$,.ELEMENT.,0.);
ENDSEC;
END-ISO-10303-21;
`
	doc, err := ParseDocument([]byte(src), "model.ifc", "")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Kind != "ifc" {
		t.Fatalf("kind %s", doc.Kind)
	}
	if !strings.Contains(doc.Text, "ЖК Север") || !strings.Contains(doc.Text, "Этаж 3") {
		t.Fatalf("ifc text: %q", doc.Text)
	}
}

func TestParseDocument_DWGHarvest(t *testing.T) {
	raw := append([]byte("AC1027"), bytes.Repeat([]byte{0, 1, 2}, 40)...)
	raw = append(raw, []byte("Layer_Walls\x00\x00Title Block Main\x00")...)
	doc, err := ParseDocument(raw, "a.dwg", "")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Kind != "dwg" || !doc.Partial {
		t.Fatalf("dwg: %+v", doc)
	}
	if !strings.Contains(doc.Text, "Layer_Walls") {
		t.Fatalf("harvest missing: %q", doc.Text)
	}
}

func TestParseDocument_PLNPartial(t *testing.T) {
	raw := append([]byte("GRAPHISOFT"), bytes.Repeat([]byte{0xff}, 80)...)
	raw = append(raw, []byte("Building_North\x00")...)
	doc, err := ParseDocument(raw, "house.pln", "")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Kind != "pln" || !doc.Partial {
		t.Fatalf("pln: %+v", doc)
	}
}

func TestSanitizeAndTruncate(t *testing.T) {
	s := SanitizeExtractedText("see {{recordID}} and }}")
	if strings.Contains(s, "{{") || strings.Contains(s, "}}") {
		t.Fatalf("not sanitized: %q", s)
	}
	got := TruncateRunes("абвгд", 4)
	if got != "абв…" {
		t.Fatalf("truncate %q", got)
	}
}

func TestExtractRankPrefersIFC(t *testing.T) {
	if ExtractRank("ifc") <= ExtractRank("pln") {
		t.Fatal("ifc should rank above pln")
	}
}

func mustZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
