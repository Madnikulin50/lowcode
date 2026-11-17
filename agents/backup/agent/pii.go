package agent

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Контентное детектирование ПДн: для небольших файлов читаем начало
// содержимого и ищем структурные паттерны (СНИЛС, ИНН, паспорт, телефон,
// email, банковская карта) плюс ключевые слова. Срабатывание эскалирует
// находку до personal_special/critical — это уже не «угадывание по имени»,
// а подтверждённый риск.

const (
	// piiMaxFileMB — файлы больше не читаем вообще.
	piiMaxFileMB = 2 << 20
	// piiMaxTextBytes — сколько текста нормируем для анализа.
	piiMaxTextBytes = 512 << 10
)

// textExts — читаем как текст напрямую.
var textExts = map[string]bool{
	"txt": true, "csv": true, "log": true, "md": true, "json": true,
	"xml": true, "yml": true, "yaml": true, "ini": true, "cfg": true,
	"conf": true, "sql": true, "sh": true, "py": true, "js": true,
	"ts": true, "go": true, "java": true, "php": true, "html": true,
	"htm": true, "rst": true, "tsv": true,
}

// officeZipExts — zip-контейнеры с XML-текстом внутри.
var officeZipExts = map[string]bool{
	"docx": true, "xlsx": true, "pptx": true, "odt": true, "ods": true, "odp": true,
}

var (
	snilsRe = regexp.MustCompile(`\b\d{3}-\d{3}-\d{3}[ -]?\d{2}\b`)
	innRe   = regexp.MustCompile(`\b(инн|inn)\s*[:№#]?\s*[A-ZА-Я0-9/-]*\s*\d{10}(\d{2})?\b`)
	passRe  = regexp.MustCompile(`(?i)\b(паспорт|passport)\b[^\n]{0,40}?\b\d{4}[\s-]?\d{6}\b|\b\d{4}[\s-]?\d{6}\b[^\n]{0,20}?\b(паспорт|выдан)\b`)
	phoneRe = regexp.MustCompile(`(?:\+?7|8)[\s(-]?\d{3}[\s)-]?\d{3}[\s-]?\d{2}[\s-]?\d{2}`)
	emailRe = regexp.MustCompile(`(?i)\b[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}\b`)
	cardRe  = regexp.MustCompile(`\b\d{13,19}\b`)
)

var piiKeywordsRe = regexp.MustCompile(`(?i)персональн|снилс|\bинн\b|медицин|диагноз|зарплат|трудовой|полис|номер\s+карты|банковск`)

// scanFilePII возвращает список сработавших индикаторов. ok=false — файл
// не подлежал сканированию (большой, бинарный, неподходящий тип) — это не
// «ПДн не найдено», а «не смотрели». Только регулярные файлы: открытие
// FIFO/сокета/устройства может заблокироваться навсегда.
func scanFilePII(fullPath string) (hits []string, ok bool) {
	info, err := os.Stat(fullPath)
	if err != nil || !info.Mode().IsRegular() || info.Size() > piiMaxFileMB {
		return nil, false
	}
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(fullPath)), ".")

	var text []byte
	switch {
	case textExts[ext] || ext == "":
		raw, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, false
		}
		if !mostlyText(raw) {
			return nil, false
		}
		text = raw
	case officeZipExts[ext]:
		text = officeZipText(fullPath)
		if len(text) == 0 {
			return nil, false
		}
	default:
		return nil, false
	}
	if len(text) > piiMaxTextBytes {
		text = text[:piiMaxTextBytes]
	}

	structural := 0
	if snilsRe.Find(text) != nil {
		hits = append(hits, "снилс")
		structural++
	}
	if innRe.Find(text) != nil {
		hits = append(hits, "инн")
		structural++
	}
	if passRe.Find(text) != nil {
		hits = append(hits, "паспорт")
		structural++
	}
	if phoneRe.Find(text) != nil {
		hits = append(hits, "телефон")
		structural++
	}
	if emailRe.Find(text) != nil {
		hits = append(hits, "email")
		structural++
	}
	if cardRe.Find(text) != nil && luhnHit(cardRe.FindAll(text, -1)) {
		hits = append(hits, "карта")
		structural++
	}
	keywords := len(piiKeywordsRe.FindAll(text, -1))

	switch {
	case structural >= 2, structural >= 1 && keywords >= 1, keywords >= 3:
		if keywords > 0 && !contains(hits, "ключевые слова") {
			hits = append(hits, "ключевые слова")
		}
		return hits, true
	}
	return nil, false
}

// piiScanTimeout — жёсткий дедлайн на один файл: чтение (FUSE, сетевая ФС)
// не должно вешать аудит. Горутина, превысившая дедлайн, остаётся
// заблокированной в ОС — редкий случай, приемлемая утечка.
const piiScanTimeout = 5 * time.Second

// scanFilePIITimed — scanFilePII с дедлайном.
func scanFilePIITimed(fullPath string) (hits []string, ok bool) {
	type res struct {
		hits []string
		ok   bool
	}
	ch := make(chan res, 1)
	go func() {
		h, o := scanFilePII(fullPath)
		ch <- res{h, o}
	}()
	select {
	case r := <-ch:
		return r.hits, r.ok
	case <-time.After(piiScanTimeout):
		return nil, false
	}
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

// mostlyText — грубая проверка, что начало файла похоже на текст.
func mostlyText(raw []byte) bool {
	if len(raw) == 0 {
		return false
	}
	n := min(len(raw), 4096)
	printable := 0
	for _, b := range raw[:n] {
		// control chars кроме \n\r\t непечатны; всё ≥0x80 считаем текстом
		// (UTF-8: кириллица — двухбайтовая, второй байт 0x80–0xBF)
		if b == '\n' || b == '\r' || b == '\t' || (b >= 0x20 && b < 0x7f) || b >= 0x80 {
			printable++
		}
	}
	return printable*100/n >= 85
}

// officeZipText вытаскивает текст из xml-частей офисного контейнера.
func officeZipText(fullPath string) []byte {
	r, err := zip.OpenReader(fullPath)
	if err != nil {
		return nil
	}
	defer r.Close()
	var buf bytes.Buffer
	for _, f := range r.File {
		name := strings.ToLower(f.Name)
		if !strings.HasSuffix(name, ".xml") && !strings.HasSuffix(name, ".rdf") {
			continue
		}
		if f.UncompressedSize64 > 1<<20 {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		_, _ = io.Copy(&buf, io.LimitReader(rc, 256<<10))
		rc.Close()
		if buf.Len() > piiMaxTextBytes {
			break
		}
	}
	return buf.Bytes()
}

// luhnHit проверяет цифровые последовательности по алгоритму Луна.
func luhnHit(runs [][]byte) bool {
	for _, run := range runs {
		digits := make([]int, 0, len(run))
		for _, b := range run {
			if b >= '0' && b <= '9' {
				digits = append(digits, int(b-'0'))
			}
		}
		if len(digits) < 13 || len(digits) > 19 {
			continue
		}
		sum := 0
		for i, d := range digits {
			if (len(digits)-i)%2 == 0 {
				d *= 2
				if d > 9 {
					d -= 9
				}
			}
			sum += d
		}
		if sum%10 == 0 {
			return true
		}
	}
	return false
}
