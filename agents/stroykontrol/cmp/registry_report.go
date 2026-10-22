package cmp

import (
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// Выгрузка реестра замечаний экспертизы в Excel (ТЗ 3.9.1–3.9.2): реестр с исходным статусом
// экспертизы, статусом выполнения, результатом проверки исправления и решением специалиста;
// история итераций (ответы, причины отклонения, назначения, комментарии, решения); проверка
// устранения; динамика по дням.

// RegistryRemark is one row of the registry.
type RegistryRemark struct {
	Number, Origin, Criticality, Section, Discipline, Expert string
	Text, Refs, Answer, RejectReason                         string
	SourceStatus, Status                                     string // the expertise's own status, the working status
	Verdict, VerdictNote, AIVerdict, AINote                  string // check against the documents, AI reading
	Outcome, OutcomeFinal, OutcomeComment, OutcomeBy         string // result of the check: the system's, the specialist's
	Assignee, DueDate, Note, ChangedAt, EstimateH            string
	Iterations                                               int
	Overdue                                                  bool
	Places                                                   string // where the remark leads in the documents
}

// RegistryHistory is one entry of the history of a remark.
type RegistryHistory struct {
	Number, At, Kind, Source, Author, Status string
	Iteration                                int
	Answer, RejectReason, Note               string
}

// RegistryReport is what the export is built from.
type RegistryReport struct {
	ObjectName string
	Date       time.Time
	Filters    []string // the filters the rows were selected by, in words
	Remarks    []RegistryRemark
	History    []RegistryHistory
	Dynamics   []StatusSnapshot
}

// VerdictLabels names the results of the check against the documents.
var VerdictLabels = map[string]string{
	VerdictStillDiffers: "Расхождение сохраняется",
	VerdictResolved:     "Устранено в этой итерации",
	VerdictNoMatch:      "Связанных расхождений не найдено",
	VerdictNoDocument:   "Тома не загружены",
}

// AILabels names the readings of the model.
var AILabels = map[string]string{
	AIFixed: "Устранено", AIPartial: "Частично", AINotFixed: "Не устранено", AIUnclear: "Судить нельзя",
}

// HistKindLabels names the kinds of history entries.
var HistKindLabels = map[string]string{
	HistImport: "Выгрузка экспертизы", HistStatus: "Смена статуса", HistAssignee: "Назначение", HistDue: "Срок",
	HistNote: "Заметка", HistDecision: "Решение по проверке", HistAuto: "Агент", HistComparison: "Из сравнения",
}

func label(m map[string]string, k string) string {
	if k == "" {
		return ""
	}
	if v, ok := m[k]; ok {
		return v
	}
	return k
}

func yesNo(b bool) string {
	if b {
		return "да"
	}
	return ""
}

// RegistryXLSX writes the registry export.
func RegistryXLSX(r RegistryReport) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	bold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}, Alignment: &excelize.Alignment{WrapText: true, Vertical: "top"}})
	wrap, _ := f.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{WrapText: true, Vertical: "top"}})
	put := func(sheet string, row int, vals []any) {
		for c, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(c+1, row)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}
	table := func(sheet string, heads []string, rows [][]any, widths []float64) error {
		if sheet != "Сводка" {
			if _, err := f.NewSheet(sheet); err != nil {
				return err
			}
		}
		hs := make([]any, len(heads))
		for i, h := range heads {
			hs[i] = h
		}
		put(sheet, 1, hs)
		_ = f.SetRowStyle(sheet, 1, 1, bold)
		for i, row := range rows {
			put(sheet, i+2, row)
		}
		if len(rows) > 0 {
			last, _ := excelize.CoordinatesToCellName(len(heads), len(rows)+1)
			_ = f.SetCellStyle(sheet, "A2", last, wrap)
		}
		for i, w := range widths {
			col, _ := excelize.ColumnNumberToName(i + 1)
			_ = f.SetColWidth(sheet, col, col, w)
		}
		_ = f.SetPanes(sheet, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
		if len(rows) > 0 {
			last, _ := excelize.CoordinatesToCellName(len(heads), len(rows)+1)
			_ = f.AutoFilter(sheet, "A1:"+last, nil)
		}
		return nil
	}

	// Сводка
	const sum = "Сводка"
	f.SetSheetName("Sheet1", sum)
	byStatus := map[string]int{}
	byOutcome := map[string]int{}
	overdue, again := 0, 0
	for _, x := range r.Remarks {
		byStatus[x.Status]++
		o := firstNonEmpty(x.OutcomeFinal, x.Outcome)
		byOutcome[o]++
		if x.Overdue {
			overdue++
		}
		if x.Iterations >= 2 {
			again++
		}
	}
	rows := [][]any{
		{"Отчёт", "Реестр замечаний экспертизы"},
		{"Объект", r.ObjectName},
		{"Дата выгрузки", r.Date.Format("02.01.2006 15:04")},
		{"Отбор", firstNonEmpty(strings.Join(r.Filters, "; "), "без отбора")},
		{"Замечаний в выгрузке", len(r.Remarks)},
	}
	for _, code := range RemarkStatusOrder {
		rows = append(rows, []any{"  " + RemarkStatusLabels[code], byStatus[code]})
	}
	rows = append(rows, []any{"Просрочено", overdue}, []any{"Повторных итераций (2 и более)", again})
	for _, code := range []string{OutcomeFixed, OutcomePartial, OutcomeNotFixed, OutcomeCheck} {
		rows = append(rows, []any{"Результат проверки: " + OutcomeLabels[code], byOutcome[code]})
	}
	rows = append(rows, []any{"Примечание", "Статус экспертизы берётся из выгрузки и не меняется по результатам проверки; результат проверки — вывод системы (по документам и разбору ИИ) и решение специалиста — хранится отдельно."})
	for i, row := range rows {
		put(sum, i+1, row)
	}
	_ = f.SetCellStyle(sum, "A1", fmt.Sprintf("B%d", len(rows)), wrap)
	_ = f.SetColWidth(sum, "A", "A", 36)
	_ = f.SetColWidth(sum, "B", "B", 90)

	// Реестр
	var rr [][]any
	for _, x := range r.Remarks {
		rr = append(rr, []any{
			x.Number, label(map[string]string{"expertise": "Экспертиза", OriginComparison: "Сравнение ПД/РД"}, x.Origin),
			label(map[string]string{"critical": "Критичное", "noncritical": "Некритичное"}, x.Criticality), x.Section, x.Discipline, x.Expert,
			x.Text, x.Refs, x.Answer, x.RejectReason, x.Iterations,
			x.SourceStatus, label(RemarkStatusLabels, x.Status),
			label(VerdictLabels, x.Verdict), label(AILabels, x.AIVerdict), label(OutcomeLabels, x.Outcome),
			label(OutcomeLabels, x.OutcomeFinal), x.OutcomeComment,
			x.Assignee, x.DueDate, yesNo(x.Overdue), x.EstimateH, x.Note, x.ChangedAt, x.Places,
		})
	}
	if err := table("Реестр", []string{"№", "Источник", "Критичность", "Раздел", "НД", "Эксперт", "Вывод о несоответствии", "Ссылка на материалы", "Ответ проектировщика", "Причина отклонения", "Итерация",
		"Статус экспертизы (выгрузка)", "Статус выполнения", "Проверка по документам", "Разбор ИИ", "Результат проверки (система)", "Результат проверки (специалист)", "Комментарий к решению",
		"Ответственный", "Срок", "Просрочено", "Оценка, ч", "Заметка ПТО", "Изменено в экспертизе", "Где в документах"}, rr,
		[]float64{10, 14, 13, 24, 8, 18, 60, 36, 50, 36, 9, 24, 22, 24, 16, 20, 20, 36, 18, 12, 11, 9, 36, 18, 44}); err != nil {
		return nil, err
	}

	// История итераций
	var hr [][]any
	for _, h := range r.History {
		hr = append(hr, []any{h.Number, h.Iteration, h.At, label(HistKindLabels, firstNonEmpty(h.Kind, HistImport)), h.Source, h.Author, label(RemarkStatusLabels, h.Status), h.Answer, h.RejectReason, h.Note})
	}
	if err := table("История итераций", []string{"№ замечания", "Итерация", "Когда", "Вид записи", "Источник", "Кто", "Статус", "Ответ", "Причина отклонения", "Комментарий"}, hr,
		[]float64{12, 9, 18, 22, 36, 16, 22, 50, 36, 40}); err != nil {
		return nil, err
	}

	// Проверка устранения
	var cr [][]any
	for _, x := range r.Remarks {
		cr = append(cr, []any{x.Number, label(VerdictLabels, x.Verdict), x.VerdictNote, label(AILabels, x.AIVerdict), x.AINote, label(OutcomeLabels, x.Outcome),
			label(OutcomeLabels, x.OutcomeFinal), x.OutcomeComment, x.OutcomeBy})
	}
	if err := table("Проверка устранения", []string{"№ замечания", "Проверка по документам", "Пояснение проверки", "Разбор ИИ", "Пояснение ИИ (с цитатами)", "Результат (система)", "Результат (специалист)", "Комментарий специалиста", "Решил"}, cr,
		[]float64{12, 26, 50, 16, 60, 20, 20, 40, 16}); err != nil {
		return nil, err
	}

	// Динамика
	if len(r.Dynamics) > 0 {
		var dr [][]any
		for _, d := range r.Dynamics {
			row := []any{d.Date, d.Total}
			for _, code := range RemarkStatusOrder {
				row = append(row, d.ByStatus[code])
			}
			dr = append(dr, append(row, d.Overdue))
		}
		heads := []string{"День", "Всего"}
		for _, code := range RemarkStatusOrder {
			heads = append(heads, RemarkStatusLabels[code])
		}
		if err := table("Динамика", append(heads, "Просрочено"), dr, []float64{12, 9, 20, 22, 26, 22, 12, 12}); err != nil {
			return nil, err
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
