package cmp

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func remarksPDF(t *testing.T, name string) []ExportRow {
	t.Helper()
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skip("pdftotext not installed")
	}
	data, err := os.ReadFile(filepath.Join("testdata", "remarks-pdf", name))
	if err != nil {
		t.Skipf("fixture missing (run fixtures/gen_remarks_pdf.mjs): %v", err)
	}
	rows, err := ParseRemarksFile(data, name, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return rows
}

func TestParseRemarksPDFTable(t *testing.T) {
	rows := remarksPDF(t, "table.pdf")
	if len(rows) != 16 {
		for _, r := range rows {
			t.Logf("%s | %s | %s | %s", r.Number, r.Refs, r.Text, r.Answer)
		}
		t.Fatalf("want 16 remarks over two pages, got %d", len(rows))
	}
	r := rows[0]
	if r.Number != "1" || r.Refs != "2024-15-КЖ2" || !strings.HasSuffix(r.Text, "Привести в соответствие.") || !strings.Contains(r.Answer, "КЖ2 изм. 1") {
		t.Errorf("first remark (wrapped cells): %+v", r)
	}
	if rows[2].Answer != "" || rows[15].Number != "16" || !strings.Contains(rows[15].Answer, "АР-9") {
		t.Errorf("rows: %+v / %+v", rows[2], rows[15])
	}
}

func TestParseRemarksPDFList(t *testing.T) {
	rows := remarksPDF(t, "list.pdf")
	if len(rows) != 5 {
		for _, r := range rows {
			t.Logf("%s | %s | %s | %s | %s", r.Number, r.Section, r.Refs, r.Text, r.Answer)
		}
		t.Fatalf("want 5 numbered remarks, got %d", len(rows))
	}
	if rows[0].Section != "Раздел 4 КЖ" || rows[3].Section != "Раздел 3 АР" {
		t.Errorf("sections: %q %q", rows[0].Section, rows[3].Section)
	}
	if !strings.Contains(rows[0].Answer, "2,6 м") || rows[2].Answer != "" || strings.Contains(rows[0].Text, "Ответ") {
		t.Errorf("answers: %+v", rows[:3])
	}
	if rows[0].Refs != "2024-15-КЖ2" {
		t.Errorf("volume of the remark: %q", rows[0].Refs)
	}
}

func TestParseRemarksPDFWithoutRemarks(t *testing.T) {
	if _, err := ParseRemarksPDF([]Page{fragPage(1, "Пояснительная записка", "1. Общие данные проекта здания")}, "x.pdf"); err == nil {
		t.Error("a PDF that does not speak of remarks is not read as remarks")
	}
	if _, err := ParseRemarksPDF([]Page{{Number: 1}}, "scan.pdf"); err == nil || !strings.Contains(err.Error(), "текстового слоя") {
		t.Errorf("scan: %v", err)
	}
}
