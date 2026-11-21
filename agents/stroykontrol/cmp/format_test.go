package cmp

import (
	"strings"
	"testing"
)

type fmtPage struct {
	sheet int      // sheet number printed in the title block (0: none)
	code  string   // шифр in the title block ("" → the default)
	body  []string // text of the sheet
}

// fmtDoc builds a document whose pages have a title block with a sheet number and a шифр.
func fmtDoc(id, side string, pages ...fmtPage) *Document {
	d := &Document{ID: id, Name: id + ".pdf", Side: side}
	for i, fp := range pages {
		p := Page{Number: i + 1, W: 595, H: 842, HasText: true}
		add := func(text string, stamp bool, y float64) {
			l := mkLine(text, 0.1, y, 0.8, y+0.02)
			l.DocID, l.Page, l.Stamp = id, i+1, stamp
			p.Lines = append(p.Lines, l)
		}
		for j, r := range fp.body {
			add(r, false, 0.1+float64(j)*0.03)
		}
		code := fp.code
		if code == "" {
			code = "2024-15-КЖ1"
		}
		add("Изм. Кол.уч. Лист №док. Подпись Дата", true, 0.86)
		add(code+" Стадия Лист Листов", true, 0.88)
		if fp.sheet > 0 {
			add("Проверил Иванов 02.24г П "+itoa(fp.sheet), true, 0.90)
		}
		d.Pages = append(d.Pages, p)
	}
	return d
}

func itoa(n int) string {
	return strings.TrimSpace(strings.Repeat(" ", 0) + string(rune('0'+n/100%10)) + string(rune('0'+n/10%10)) + string(rune('0'+n%10)))
}

func body(n int, topic string) []string {
	var out []string
	for i := 0; i < 10; i++ {
		out = append(out, topic+" строка "+string(rune('а'+i))+" "+string(rune('а'+n%20)))
	}
	return out
}

func formatMessages(fs []Finding) string {
	var parts []string
	for _, f := range fs {
		parts = append(parts, f.Evidence.Rule+": "+f.Description)
	}
	return strings.Join(parts, "\n")
}

func volOf(t *testing.T, d *Document) *Volume {
	t.Helper()
	vs := DetectVolumes(d)
	if len(vs) != 1 {
		t.Fatalf("want one volume, got %d", len(vs))
	}
	return vs[0]
}

func TestFormatCleanVolumeHasNoFindings(t *testing.T) {
	d := fmtDoc("d", SideRef, fmtPage{1, "", body(1, "а")}, fmtPage{2, "", body(2, "б")}, fmtPage{3, "", body(3, "в")}, fmtPage{4, "", body(4, "г")}, fmtPage{5, "", body(5, "д")})
	if fs := FormatChecks(d, volOf(t, d), SideRef); len(fs) != 0 {
		t.Fatalf("a clean volume must not be flagged:\n%s", formatMessages(fs))
	}
}

func TestFormatSheetNumberJumpAndRepeat(t *testing.T) {
	d := fmtDoc("d", SideTarget,
		fmtPage{1, "", body(1, "а")}, fmtPage{2, "", body(2, "б")}, fmtPage{4, "", body(3, "в")}, // 3 is missing
		fmtPage{5, "", body(4, "г")}, fmtPage{5, "", body(5, "д")}, // 5 repeats
		fmtPage{7, "", body(6, "е")},
		fmtPage{1, "", body(7, "ж")}, fmtPage{2, "", body(8, "з")}, // a new document in the file starts again from 1: fine
	)
	fs := FormatChecks(d, volOf(t, d), SideTarget)
	msg := formatMessages(fs)
	if !strings.Contains(msg, "не сходится со страницами") || !strings.Contains(msg, "повторяется") {
		t.Fatalf("expected a jump and a repeat:\n%s", msg)
	}
	for _, f := range fs {
		if f.Kind != KindFormat || f.TargetDoc != "d" || f.PageTarget == 0 || f.RefDoc != "" {
			t.Fatalf("a target-side finding must point at the target page only: %+v", f)
		}
	}
}

func TestFormatTooFewTitleBlocksSaysNothing(t *testing.T) {
	d := fmtDoc("d", SideRef, fmtPage{1, "", body(1, "а")}, fmtPage{9, "", body(2, "б")}, fmtPage{0, "", body(3, "в")})
	if fs := FormatChecks(d, volOf(t, d), SideRef); len(fs) != 0 {
		t.Fatalf("%s", formatMessages(fs))
	}
}

func TestFormatStrayCodeAndDuplicateSheet(t *testing.T) {
	same := body(7, "ж")
	d := fmtDoc("d", SideRef,
		fmtPage{1, "", body(1, "а")}, fmtPage{2, "", same}, fmtPage{3, "2024-15-АР1", body(3, "в")}, fmtPage{4, "", same},
		fmtPage{5, "", body(5, "д")}, fmtPage{6, "", body(6, "е")}, fmtPage{7, "", body(8, "з")}, fmtPage{8, "", body(9, "и")},
	)
	vs := DetectVolumes(d)
	if len(vs) == 0 {
		t.Fatal("no volumes")
	}
	fs := FormatChecks(d, vs[0], SideRef)
	msg := formatMessages(fs)
	if !strings.Contains(msg, "дублирует лист на стр. 2") {
		t.Fatalf("duplicate sheet not found:\n%s", msg)
	}
	if !strings.Contains(msg, "шифр «2024-15-АР1»") {
		t.Fatalf("stray шифр not found:\n%s", msg)
	}
}

func TestFormatContentsReferences(t *testing.T) {
	contents := []string{"Содержание тома", "Обозначение Наименование Примечание", "Общие данные стр. 2", "Расчёт фундаментов стр. 3", "Спецификация материалов стр. 4", "Приложения стр. 40"}
	d := fmtDoc("d", SideRef,
		fmtPage{1, "", contents},
		fmtPage{2, "", append([]string{"Общие данные по проекту"}, body(2, "б")...)},
		fmtPage{3, "", append([]string{"Расчёт фундаментов мелкого заложения"}, body(3, "в")...)},
		fmtPage{4, "", append([]string{"Узлы и детали креплений"}, body(4, "г")...)},
	)
	fs := FormatChecks(d, volOf(t, d), SideRef)
	msg := formatMessages(fs)
	if !strings.Contains(msg, "«Спецификация материалов» — стр. 4") && !strings.Contains(msg, "Спецификация материалов") {
		t.Fatalf("a section that is not on its page must be reported:\n%s", msg)
	}
	if !strings.Contains(msg, "ссылается на стр. 40") {
		t.Fatalf("a reference beyond the file must be reported:\n%s", msg)
	}
	if strings.Contains(msg, "Общие данные стр") || strings.Contains(msg, "Расчёт фундаментов") {
		t.Fatalf("correct references must not be flagged:\n%s", msg)
	}
}

func TestFormatChecksRunInsideCompareAndCanBeSwitchedOff(t *testing.T) {
	ref := fmtDoc("r", SideRef, fmtPage{1, "", body(1, "а")}, fmtPage{2, "", body(2, "б")}, fmtPage{3, "", body(3, "в")}, fmtPage{4, "", body(4, "г")})
	tgt := fmtDoc("t", SideTarget, fmtPage{1, "", body(1, "а")}, fmtPage{2, "", body(2, "б")}, fmtPage{4, "", body(3, "в")}, fmtPage{5, "", body(4, "г")}, fmtPage{6, "", body(5, "д")})
	res := Compare([]*Document{ref}, []*Document{tgt}, Options{})
	n := 0
	for _, f := range res.Findings {
		if f.Kind == KindFormat {
			n++
		}
	}
	if n == 0 {
		t.Fatal("the numbering jump on the target side must be found")
	}
	res = Compare([]*Document{ref}, []*Document{tgt}, Options{NoFormat: true})
	for _, f := range res.Findings {
		if f.Kind == KindFormat {
			t.Fatalf("formatting checks were switched off: %+v", f)
		}
	}
}

func TestFormatForeignCodeLikeTextIsNotACodeMistake(t *testing.T) {
	d := fmtDoc("d", SideRef,
		fmtPage{1, "", body(1, "а")}, fmtPage{2, "168.10-н.п", body(2, "б")}, fmtPage{3, "", body(3, "в")},
		fmtPage{4, "", body(4, "г")}, fmtPage{5, "", body(5, "д")}, fmtPage{6, "", body(6, "е")},
	)
	for _, f := range FormatChecks(d, DetectVolumes(d)[0], SideRef) {
		if strings.Contains(f.Description, "шифр") {
			t.Fatalf("not a шифр of this project: %s", f.Description)
		}
	}
}
