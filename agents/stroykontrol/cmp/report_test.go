package cmp

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func sampleReport() Report {
	return BuildReport(ReportInput{
		Title: "ИССО 3.10 · КЖ1.13", ObjectName: "ИССО 3.10", SetsLabel: "ПД v1 ↔ РД Изм. 2", Similarity: "41.5", Status: "done",
		Documents: []ReportDoc{
			{ID: "d1", Side: SideRef, Name: "pd.pdf", VolumeCode: "КЖ1", Pages: "10"},
			{ID: "d2", Side: SideTarget, Name: "rd.pdf", VolumeCode: "КЖ1", Pages: "12"},
		},
		Findings: []FindingIn{
			{ID: "f1", Kind: KindText, Severity: "medium", Description: "[детерм.] Изменён текст", Status: StatusConfirmed, DocRef: "d1", DocTarget: "d2", PageRef: "3", PageTarget: "4"},
			{ID: "f2", Kind: KindParam, Severity: "high", Description: "Класс бетона B25 → B30", Status: StatusNew, DocRef: "d1", DocTarget: "d2", PageRef: "5", PageTarget: "6", IterState: IterPersisting},
			{ID: "f3", Kind: KindText, Severity: "low", Description: "Пустяк", Status: StatusRejected, DocRef: "d1", DocTarget: "d2", PageRef: "7", PageTarget: "8"},
			{ID: "f4", Kind: KindNotRelease, Severity: "low", Description: "Нет тома ОВ", Status: StatusNew, DocRef: "d1"},
			{ID: "f5", Kind: KindFormat, Severity: "low", Description: "Пропущен лист 4", Status: StatusNew, DocTarget: "d2"},
		},
		RemarkNo:  map[string]string{"f1": "ПТО-1.1"},
		Iteration: &ReportIteration{No: "2", ParentTitle: "первая", Resolved: 3, Persisting: 1, New: 0},
		Params:    []ParamRow{{Side: SideRef, Volume: "КЖ1", Kind: "Класс бетона", Element: "фундамент", Values: []string{"B25"}, Page: 3}},
	})
}

func TestBuildReportGroupsAndCounts(t *testing.T) {
	r := sampleReport()
	if r.Counts.Total != 3 || r.Counts.High != 1 || r.Counts.Medium != 1 || r.Counts.Low != 1 || r.Counts.Confirmed != 1 || r.Counts.Rejected != 1 {
		t.Fatalf("counts: %+v", r.Counts)
	}
	if len(r.Info) != 1 || len(r.Format) != 1 {
		t.Fatalf("info=%d format=%d", len(r.Info), len(r.Format))
	}
	if r.Findings[0].Severity != "high" || r.Findings[0].No != 1 {
		t.Fatalf("findings must go by severity: %+v", r.Findings)
	}
	if got := r.Findings[1].RemarkNo; got != "ПТО-1.1" {
		t.Fatalf("remark number %q", got)
	}
	if !strings.Contains(r.Findings[1].Where, "ПД: pd.pdf, стр. 3") || !strings.Contains(r.Findings[1].Where, "РД: rd.pdf, стр. 4") {
		t.Fatalf("where: %q", r.Findings[1].Where)
	}
	if strings.Contains(r.Findings[1].Description, "[детерм.]") {
		t.Fatalf("description not cleaned: %q", r.Findings[1].Description)
	}
	if len(r.Volumes) != 1 || r.Volumes[0].Total != 3 || r.Volumes[0].Open != 2 { // the rejected one is not open
		t.Fatalf("volumes: %+v", r.Volumes)
	}
}

func TestReportXLSX(t *testing.T) {
	data, err := ReportXLSX(sampleReport())
	if err != nil {
		t.Fatal(err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	sheets := strings.Join(f.GetSheetList(), "|")
	for _, want := range []string{"Сводка", "По томам", "Расхождения", "Оформление", "Не выпущено, вне объёма", "Параметры по томам"} {
		if !strings.Contains(sheets, want) {
			t.Fatalf("sheet %q missing in %s", want, sheets)
		}
	}
	rows, _ := f.GetRows("Расхождения")
	if len(rows) != 4 || rows[1][1] != "Высокая" {
		t.Fatalf("rows: %v", rows)
	}
	sum, _ := f.GetRows("Сводка")
	joined := ""
	for _, r := range sum {
		joined += strings.Join(r, " ") + "\n"
	}
	if !strings.Contains(joined, "ПД v1 ↔ РД Изм. 2") || !strings.Contains(joined, "подтверждено 1, отклонено 1, к проверке 1") {
		t.Fatalf("summary: %s", joined)
	}
}

func TestReportDOCX(t *testing.T) {
	data, err := ReportDOCX(sampleReport())
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	var body string
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			rc, _ := f.Open()
			b, _ := io.ReadAll(rc)
			rc.Close()
			body = string(b)
		}
	}
	for _, want := range []string{"Отчёт о сравнении документации", "Класс бетона B25", "Оформление", "Не выпущено и вне объёма", "ПТО-1.1"} {
		if !strings.Contains(body, want) {
			t.Fatalf("docx lacks %q", want)
		}
	}
}
