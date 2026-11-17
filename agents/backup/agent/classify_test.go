package agent

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func now() time.Time { return time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC) }

func TestAssessFile(t *testing.T) {
	cases := []struct {
		rel      string
		mod      time.Time
		wantCrit string
		wantCls  string
	}{
		// ПДн в имени — всегда critical
		{"скан_паспорта_иванова.jpg", now().AddDate(0, -2, 0), CritCritical, "personal_special"},
		{"список_персональных_данных.xlsx", now().AddDate(0, -2, 0), CritCritical, "personal_special"},
		// Ключевой материал
		{"vpn/server.key", now().AddDate(0, -1, 0), CritCritical, "confidential"},
		{".env", now().AddDate(0, -1, 0), CritCritical, "confidential"},
		// Базы и дампы — high; каталог 1С считаем ПДн-зоной (зарплаты/кадры)
		{"1c/base.1cd", now().AddDate(0, -1, 0), CritHigh, "personal_special"},
		{"db/dump.sql", now().AddDate(0, -1, 0), CritHigh, "confidential"},
		// Каталог кадры — high и personal_special
		{"кадры/заявление.docx", now().AddDate(0, -1, 0), CritHigh, "personal_special"},
		// Офисный документ — medium (confidential)
		{"docs/договор_аренды.docx", now().AddDate(0, -1, 0), CritMedium, "confidential"},
		{"docs/отчет.pdf", now().AddDate(0, -6, 0), CritMedium, "confidential"},
		// Активность поднимает, залежь понижает
		{"docs/презентация.pptx", now().AddDate(0, 0, -3), CritHigh, "confidential"},
		{"docs/старый_договор.docx", now().AddDate(-3, 0, 0), CritLow, "confidential"},
		// Код и медиа — low
		{"src/main.go", now().AddDate(0, -1, 0), CritLow, "internal"},
		{"фото/пейзаж.jpg", now().AddDate(0, -1, 0), CritLow, "internal"},
		// Финансовые маркеры
		{"бухгалтерия/счет_№15.xls", now().AddDate(0, -2, 0), CritMedium, "confidential"},
	}
	for _, tc := range cases {
		got := assessFile(tc.rel, 1024, tc.mod, now())
		if got.Criticality != tc.wantCrit || got.Class != tc.wantCls {
			t.Errorf("%s: got %s/%s want %s/%s (%v)", tc.rel, got.Criticality, got.Class, tc.wantCrit, tc.wantCls, got.Reasons)
		}
	}
}

func TestEscalatePII(t *testing.T) {
	fa := assessFile("docs/анкета.docx", 1024, now().AddDate(0, -1, 0), now())
	fa.escalatePII([]string{"снилс", "паспорт"})
	if fa.Criticality != CritCritical || fa.Class != "personal_special" {
		t.Fatalf("escalate: %s/%s", fa.Criticality, fa.Class)
	}
	if !strings.Contains(strings.Join(fa.Reasons, ";"), "содержимое:снилс+паспорт") {
		t.Fatalf("reasons: %v", fa.Reasons)
	}
}

func TestScanFilePII(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}

	pdn := write("anketa.txt", "ФИО: Иванов Иван\nСНИЛС: 123-456-789 01\nПаспорт 4510 123456 выдан ОВД\nтел +7 (916) 123-45-67\n")
	hits, ok := scanFilePII(pdn)
	if !ok || len(hits) < 2 {
		t.Fatalf("pdn: ok=%v hits=%v", ok, hits)
	}

	clean := write("notes.txt", "заметки про погоду и планы на выходные, ничего интересного\n")
	if hits, ok := scanFilePII(clean); ok {
		t.Fatalf("clean file flagged: %v", hits)
	}

	binary := write("blob.bin", string([]byte{0, 1, 2, 3, 255, 254, 200, 100, 0, 0, 0, 0, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7}))
	if hits, ok := scanFilePII(binary); ok {
		t.Fatalf("binary flagged: %v", hits)
	}

	// FIFO не читаем вообще (ReadFile повис бы навсегда)
	fifo := filepath.Join(dir, "pipe.fifo")
	if err := syscall.Mkfifo(fifo, 0o600); err == nil {
		done := make(chan struct{})
		go func() {
			if _, ok := scanFilePIITimed(fifo); ok {
				t.Error("fifo must not be scanned")
			}
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatal("fifo scan blocked")
		}
	}
}

func TestLuhn(t *testing.T) {
	// 4539 1488 0343 6467 — валидный тестовый номер карты (диапазон 4051..)
	if !luhnHit([][]byte{[]byte("4539148803436467")}) {
		t.Fatal("valid luhn number not detected")
	}
	if luhnHit([][]byte{[]byte("4539148803436461")}) {
		t.Fatal("invalid luhn number detected")
	}
}

func TestScanOfficeZip(t *testing.T) {
	// docx со встроенным СНИЛС в document.xml
	dir := t.TempDir()
	p := filepath.Join(dir, "документ.docx")
	if err := writeTestZip(p, map[string]string{
		"word/document.xml": `<w:document>Заявление сотрудника, СНИЛС 234-567-890 12, ИНН 7707083893</w:document>`,
	}); err != nil {
		t.Fatal(err)
	}

	hits, ok := scanFilePII(p)
	if !ok || !contains(hits, "снилс") {
		t.Fatalf("docx: ok=%v hits=%v", ok, hits)
	}
}

func writeTestZip(path string, files map[string]string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(f)
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			f.Close()
			return err
		}
		if _, err := w.Write([]byte(body)); err != nil {
			f.Close()
			return err
		}
	}
	if err := zw.Close(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
