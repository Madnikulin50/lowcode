package cmp

import (
	"strings"
	"testing"
)

func rules(fs []Finding) map[string]int {
	m := map[string]int{}
	for _, f := range fs {
		m[f.Evidence.Rule]++
	}
	return m
}

func five(topic string) []fmtPage {
	var out []fmtPage
	for i := 1; i <= 5; i++ {
		out = append(out, fmtPage{i, "", body(i, topic)})
	}
	return out
}

func TestReferenceToMissingTableAndAppendix(t *testing.T) {
	pages := five("а")
	pages[1].body = append([]string{"Нагрузки приведены в таблице 3 и в приложении Б."}, pages[1].body...)
	pages[2].body = append([]string{"Таблица 1 – Нагрузки"}, pages[2].body...)
	pages[3].body = append([]string{"Приложение А"}, pages[3].body...)
	pages[4].body = append([]string{"Расчёт см. таблицу 1 и приложение А."}, pages[4].body...)
	d := fmtDoc("d", SideRef, pages...)
	fs := FormatChecks(d, volOf(t, d), SideRef)
	r := rules(fs)
	if r["format-ref-table"] != 1 || r["format-ref-appendix"] != 1 {
		t.Fatalf("a table and an appendix that are not in the file, the existing ones are fine:\n%s", formatMessages(fs))
	}
	msg := formatMessages(fs)
	if !strings.Contains(msg, "Таблица 3") || !strings.Contains(msg, "Приложение Б") || strings.Contains(msg, "Таблица 1") || strings.Contains(msg, "Приложение А") {
		t.Errorf("wrong references reported:\n%s", msg)
	}
	for _, f := range fs {
		if f.Severity != "low" || f.PageRef == 0 {
			t.Errorf("a formal finding is low and points at a page: %+v", f)
		}
	}
}

func TestReferenceOutsideTheFileIsNotChecked(t *testing.T) {
	pages := five("б")
	pages[1].body = append([]string{
		"Значения принять по таблице 5 СП 63.13330.",
		"Состав приведён в приложении В тома 2024-15-ПЗ.",
		"Нагрузки по таблице 7 ГОСТ 27751.",
		"Условия приложения к договору соблюдены.", // «приложения к» — not an appendix
	}, pages[1].body...)
	d := fmtDoc("d", SideRef, pages...)
	if fs := FormatChecks(d, volOf(t, d), SideRef); len(fs) != 0 {
		t.Fatalf("a reference that leads out of the file is not a finding:\n%s", formatMessages(fs))
	}
}

func TestContentsLineIsNotAHeading(t *testing.T) {
	pages := five("в")
	// the contents lists «Приложение А» with a page, but the appendix itself is not in the file
	pages[0].body = append([]string{
		"Содержание тома",
		"Общие данные стр. 2",
		"Расчёт нагрузок стр. 3",
		"Конструкции стр. 4",
		"Приложение А Расчёт осадки стр. 5",
	}, pages[0].body...)
	pages[3].body = append([]string{"В расчёте использованы данные приложения А."}, pages[3].body...)
	d := fmtDoc("d", SideRef, pages...)
	r := rules(FormatChecks(d, volOf(t, d), SideRef))
	if r["format-appendix"] != 1 || r["format-ref-appendix"] != 1 {
		t.Fatalf("the declared appendix is missing and referenced: %v", r)
	}
}

func TestEmptySection(t *testing.T) {
	pages := five("г")
	pages[1].body = []string{
		"1 Общие данные", "Текст общих данных, достаточно длинный для раздела.",
		"2 Исходные данные", "3 Конструктивные решения", "Фундаменты выполнить монолитными.",
		"3.1 Основания", "3.2 Фундаменты", "Плита толщиной 600 мм.",
		"4 Спецификация",
	}
	d := fmtDoc("d", SideRef, pages...)
	fs := FormatChecks(d, volOf(t, d), SideRef)
	msg := formatMessages(fs)
	if rules(fs)["format-empty-section"] < 2 || !strings.Contains(msg, "2 Исходные данные") || !strings.Contains(msg, "3.1 Основания") {
		t.Fatalf("sections 2 and 3.1 have no text:\n%s", msg)
	}
	if strings.Contains(msg, "«1 Общие данные") || strings.Contains(msg, "«3 Конструктивные") || strings.Contains(msg, "«3.2 Фундаменты") {
		t.Errorf("sections with text (or with subsections) are fine:\n%s", msg)
	}
}

func TestDrawingListAgainstTitleBlocks(t *testing.T) {
	pages := []fmtPage{
		{1, "", []string{"Ведомость рабочих чертежей основного комплекта", "Лист Наименование Примечание", "2 План фундаментов", "3 Армирование плиты", "4 Разрез 1-1", "5 Спецификация"}},
		{2, "", body(2, "а")},
		{3, "", body(3, "б")},
		{5, "", body(5, "в")}, // sheet 4 is missing
		{6, "", body(6, "г")},
	}
	d := fmtDoc("d", SideRef, pages...)
	fs := FormatChecks(d, volOf(t, d), SideRef)
	msg := formatMessages(fs)
	if rules(fs)["format-sheet-list"] != 1 || !strings.Contains(msg, "лист 4") {
		t.Fatalf("only sheet 4 is listed and absent:\n%s", msg)
	}
	// unreadable title blocks: nothing is judged
	for i := range pages {
		pages[i].sheet = 0
	}
	pages[0].sheet = 1
	d = fmtDoc("d", SideRef, pages...)
	if fs := FormatChecks(d, volOf(t, d), SideRef); rules(fs)["format-sheet-list"] != 0 {
		t.Errorf("a list cannot be checked when the sheet numbers are not readable:\n%s", formatMessages(fs))
	}
}
