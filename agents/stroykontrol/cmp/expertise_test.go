package cmp

import (
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func xlsxBytes(t *testing.T, sheet string, rows [][]string) []byte {
	t.Helper()
	f := excelize.NewFile()
	f.SetSheetName("Sheet1", sheet)
	for r, row := range rows {
		for c, v := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+1)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestParseExpertiseFindsColumnsByHeading(t *testing.T) {
	data := xlsxBytes(t, "Замечания", [][]string{
		{"Замечания экспертизы по объекту «Эстакада»"},
		{},
		{"№ п/п", "Раздел, том, лист", "Содержание замечания", "Ответ проектировщика"},
		{"Раздел 4. Искусственные сооружения"},
		{"1", "КЖ1.13, лист 3", "Количество свай марки С15-35Т6 для опоры №17п на 5 шт. больше, чем в проекте стадии П", "Замечание рассмотрено. Количество свай соответствует стадии П."},
		{"", "", "", "Корректировки внесены на листе 3."},
		{"2", "КЖ1.13, лист 5", "Глубина заложения фундамента 2,6 м не соответствует стадии П", ""},
		{"", "", "(обосновать расчётом)", ""},
		{},
		{"3", "", "Не указана марка бетона", "Принято."},
	})
	got, err := ParseExpertise(data, "zam.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("want 3 remarks, got %d: %+v", len(got), got)
	}
	if got[0].Number != "1" || !strings.Contains(got[0].Text, "С15-35Т6") || got[0].Where != "КЖ1.13, лист 3" || got[0].Section != "Раздел 4. Искусственные сооружения" {
		t.Errorf("remark 1: %+v", got[0])
	}
	if !strings.Contains(got[0].Answer, "Количество свай соответствует") || !strings.Contains(got[0].Answer, "листе 3") {
		t.Errorf("an answer-only row continues the previous remark: %q", got[0].Answer)
	}
	if !strings.Contains(got[1].Text, "обосновать расчётом") {
		t.Errorf("wrapped text must be joined: %q", got[1].Text)
	}
	if got[2].Number != "3" || got[2].Answer != "Принято." {
		t.Errorf("remark 3: %+v", got[2])
	}
	if _, err := ParseExpertise(xlsxBytes(t, "x", [][]string{{"a", "b"}, {"1", "2"}}), "x.xlsx"); err == nil {
		t.Error("a table without a remark column must be refused with a reason")
	}
}

func TestMergeExpertsByNumber(t *testing.T) {
	got := MergeExperts(
		[]Expert{{Number: "1", Text: "Текст 1"}, {Number: "2", Text: "Текст 2"}},
		[]Expert{{Number: "2", Answer: "Ответ 2"}, {Number: "3", Text: "Только в ответах", Answer: "Ответ 3"}},
	)
	if len(got) != 3 || got[1].Text != "Текст 2" || got[1].Answer != "Ответ 2" || got[2].Number != "3" {
		t.Errorf("%+v", got)
	}
}

func TestMatchRemarksToFindings(t *testing.T) {
	cands := []Candidate{
		{ID: "f-depth", Severity: "high", Text: "Расходятся значения: «1.3 Глубина заложения фундамента: 2,4 м.» (ПД, стр. 1) → «1.3 Глубина заложения фундамента: 2,6 м.» (РД, стр. 1)"},
		{ID: "f-qty", Severity: "high", Text: "Расходятся значения: «Ф-2 Ленточный фундамент м 48» → «Ф-2 Ленточный фундамент м 52»"},
		{ID: "f-num", Severity: "low", Text: "Изменена нумерация: «2.1 Полы: керамогранит» → «2.5 Полы: керамогранит»"},
		{ID: "f-wall", IterState: IterResolved, Severity: "high", Text: "Расходятся значения: «Толщина стен подвала: 300 мм» → «Толщина стен подвала: 350 мм»"},
	}
	ms := MatchRemarks([]Expert{
		{Number: "1", Text: "Глубина заложения фундамента 2,6 м не соответствует проекту стадии П (2,4 м)", Answer: "Откорректировано в соответствии со стадией П."},
		{Number: "2", Text: "Количество свай марки С15-35Т6 для опоры №17п на 5 шт. больше, чем в проекте"},
		{Number: "3", Text: "Толщина стен подвала увеличена до 350 мм без обоснования"},
		{Number: "4", Text: "Не указан класс бетона для ленточного фундамента Ф-2", Where: "лист 2"},
	}, cands)
	if ms[0].Verdict != VerdictStillDiffers || len(ms[0].FindingIDs) == 0 || ms[0].FindingIDs[0] != "f-depth" {
		t.Errorf("remark 1: %+v", ms[0])
	}
	if !ms[0].ClaimMissing || !strings.Contains(ms[0].Note, "заявляет правку") {
		t.Errorf("an answer that claims a correction while the difference remains must be flagged: %+v", ms[0])
	}
	if ms[1].Verdict != VerdictNoMatch || len(ms[1].FindingIDs) != 0 {
		t.Errorf("remark 2 has no related finding: %+v", ms[1])
	}
	if ms[2].Verdict != VerdictResolved || ms[2].FindingIDs[0] != "f-wall" {
		t.Errorf("remark 3 is about a finding fixed in this iteration: %+v", ms[2])
	}
	// shares only the code "ф-2" with the table finding; the numbering finding is unrelated
	if ms[3].Verdict != VerdictStillDiffers || ms[3].FindingIDs[0] != "f-qty" {
		t.Errorf("a code shared with a finding links the remark to it: %+v", ms[3])
	}
	for _, m := range ms {
		for _, id := range m.FindingIDs {
			if id == "f-num" {
				t.Errorf("unrelated finding matched remark %s", m.Number)
			}
		}
	}
}
