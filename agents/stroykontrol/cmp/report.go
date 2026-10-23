package cmp

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// Отчёт о сравнении: сводка и все расхождения по томам и листам с критичностью,
// решением специалиста, итерацией и номерами замечаний из реестра. Это справка
// для специалиста и для передачи, не список замечаний (его делает BuildIssue).

// ReportDoc is a source document of the comparison.
type ReportDoc struct {
	ID          string
	Side        string
	Name        string
	VolumeCode  string
	VolumeTitle string
	Pages       string
}

// ReportIteration is the re-check summary of a comparison.
type ReportIteration struct {
	No          string
	ParentTitle string
	Resolved    int
	Persisting  int
	New         int
}

// ReportInput is everything a report is made of.
type ReportInput struct {
	Title      string
	ObjectName string
	SetsLabel  string // "ПД v2 ↔ РД v1"
	Status     string
	Similarity string
	Comment    string
	Date       time.Time
	// RefName / TargetName are how the two sides are called ("ПД"/"РД", or "До"/"После")
	RefName, TargetName string
	Documents           []ReportDoc
	Findings            []FindingIn
	RemarkNo            map[string]string // finding ID → number of the remark it went into
	Iteration           *ReportIteration
	Params              []ParamRow    // key parameters per volume and element (cross-volume check)
	SheetChanges        []SheetChange // the statuses of the sheets of the new edition (versions)
}

// ReportFinding is a finding as the report shows it.
type ReportFinding struct {
	No          int
	Severity    string // high | medium | low
	Kind        string
	Decision    string // new | confirmed | rejected
	IterState   string
	Where       string // "ПД: том, стр. 3 · РД: том, стр. 5"
	Volume      string
	Description string
	RemarkNo    string
}

// ReportVolume is the findings count of one pair of documents.
type ReportVolume struct {
	Ref, Target            string
	Total                  int
	High, Medium, Low      int
	Confirmed, Rejected    int
	Open                   int // not rejected, not resolved in a later iteration
	FirstSheetRef, FirstTg string
}

// Report is the built report.
type Report struct {
	Title, ObjectName, SetsLabel, Status, Similarity, Comment, Date string
	RefName, TargetName                                             string
	Ref, Target                                                     []ReportDoc
	Counts                                                          struct{ Total, High, Medium, Low, Confirmed, Rejected, New, Resolved int }
	Iteration                                                       *ReportIteration
	Volumes                                                         []ReportVolume
	Findings                                                        []ReportFinding // defects
	Info                                                            []ReportFinding // not released / out of scope
	Format                                                          []ReportFinding // formatting checks
	Params                                                          []ParamRow
	SheetChanges                                                    []SheetChange
}

// KindLabel is the Russian name of a finding kind.
func KindLabel(kind string) string {
	switch kind {
	case KindText:
		return "Текст"
	case KindTableRow:
		return "Строка таблицы"
	case KindParam:
		return "Ключевой параметр"
	case KindDrawing:
		return "Чертёж"
	case KindVOR:
		return "ВОР / спецификация"
	case KindNotRelease:
		return "Не выпущено"
	case KindScopeDiff:
		return "Вне объёма"
	case KindFormat:
		return "Оформление"
	case KindCrossVolume:
		return "Между томами"
	case KindElement:
		return "Марки элементов"
	case KindChangeReg:
		return "Регистрация изменений"
	}
	return kind
}

// SeverityLabel is the Russian name of a severity.
func SeverityLabel(s string) string {
	switch s {
	case "high":
		return "Высокая"
	case "medium":
		return "Средняя"
	}
	return "Низкая"
}

// DecisionLabel is the Russian name of the specialist's decision.
func DecisionLabel(s string) string {
	switch s {
	case StatusConfirmed:
		return "Подтверждено"
	case StatusRejected:
		return "Отклонено"
	case StatusClarify:
		return "Требует уточнения"
	}
	return "К проверке"
}

func iterLabel(s string) string {
	switch s {
	case IterResolved:
		return "устранено"
	case IterNew:
		return "новое"
	case IterPersisting:
		return "не устранено"
	}
	return ""
}

// BuildReport arranges the findings of a comparison into a report.
func BuildReport(in ReportInput) Report {
	r := Report{Title: in.Title, ObjectName: in.ObjectName, SetsLabel: in.SetsLabel, Status: in.Status, Similarity: in.Similarity,
		Params: in.Params, SheetChanges: in.SheetChanges, Comment: in.Comment, RefName: firstNonEmpty(in.RefName, "ПД"), TargetName: firstNonEmpty(in.TargetName, "РД"), Iteration: in.Iteration}
	date := in.Date
	if date.IsZero() {
		date = time.Now()
	}
	r.Date = date.Format("02.01.2006")
	name := map[string]string{}
	for _, d := range in.Documents {
		name[d.ID] = d.Name
		if d.Side == SideRef {
			r.Ref = append(r.Ref, d)
		} else {
			r.Target = append(r.Target, d)
		}
	}
	docName := func(id string) string {
		if n := name[id]; n != "" {
			return n
		}
		return id
	}

	rank := map[string]int{"high": 0, "medium": 1, "low": 2}
	vols := map[string]*ReportVolume{}
	var volOrder []string
	for _, f := range in.Findings {
		var ev Evidence
		if f.Evidence != "" {
			_ = json.Unmarshal([]byte(f.Evidence), &ev)
		}
		pr, pt := pageOf(f)
		var where []string
		if f.DocRef != "" || pr != "" {
			where = append(where, fmt.Sprintf("%s: %s%s", r.RefName, docName(f.DocRef), pageSuffix(pr)))
		}
		if f.DocTarget != "" || pt != "" {
			where = append(where, fmt.Sprintf("%s: %s%s", r.TargetName, docName(f.DocTarget), pageSuffix(pt)))
		}
		rf := ReportFinding{Severity: f.Severity, Kind: f.Kind, Decision: firstNonEmpty(f.Status, StatusNew), IterState: f.IterState,
			Where: strings.Join(where, " · "), Volume: firstNonEmpty(ev.VolumeRef, ev.VolumeTgt), Description: cleanDescription(f.Description),
			RemarkNo: in.RemarkNo[f.ID]}
		switch {
		case informational(f.Kind):
			r.Info = append(r.Info, rf)
			continue
		case f.Kind == KindFormat:
			r.Format = append(r.Format, rf)
			continue
		}
		r.Findings = append(r.Findings, rf)

		c := &r.Counts
		c.Total++
		switch f.Severity {
		case "high":
			c.High++
		case "medium":
			c.Medium++
		default:
			c.Low++
		}
		switch rf.Decision {
		case StatusConfirmed:
			c.Confirmed++
		case StatusRejected:
			c.Rejected++
		}
		switch f.IterState {
		case IterNew:
			c.New++
		case IterResolved:
			c.Resolved++
		}

		key := f.DocRef + "|" + f.DocTarget
		v := vols[key]
		if v == nil {
			v = &ReportVolume{Ref: docName(f.DocRef), Target: docName(f.DocTarget)}
			vols[key] = v
			volOrder = append(volOrder, key)
		}
		v.Total++
		switch f.Severity {
		case "high":
			v.High++
		case "medium":
			v.Medium++
		default:
			v.Low++
		}
		switch rf.Decision {
		case StatusConfirmed:
			v.Confirmed++
		case StatusRejected:
			v.Rejected++
		}
		if rf.Decision != StatusRejected && f.IterState != IterResolved {
			v.Open++
		}
	}
	less := func(a, b ReportFinding) bool {
		if rank[a.Severity] != rank[b.Severity] {
			return rank[a.Severity] < rank[b.Severity]
		}
		return false
	}
	sort.SliceStable(r.Findings, func(i, j int) bool { return less(r.Findings[i], r.Findings[j]) })
	for i := range r.Findings {
		r.Findings[i].No = i + 1
	}
	for i := range r.Info {
		r.Info[i].No = i + 1
	}
	for i := range r.Format {
		r.Format[i].No = i + 1
	}
	for _, k := range volOrder {
		r.Volumes = append(r.Volumes, *vols[k])
	}
	sort.SliceStable(r.Volumes, func(i, j int) bool { return r.Volumes[i].Open > r.Volumes[j].Open })
	return r
}

func pageSuffix(p string) string {
	if p == "" {
		return ""
	}
	return ", стр. " + p
}

// summaryRows are the key/value lines of the report head, shared by Excel and Word.
func (r Report) summaryRows() [][2]string {
	rows := [][2]string{{"Сравнение", r.Title}}
	if r.ObjectName != "" {
		rows = append(rows, [2]string{"Объект", r.ObjectName})
	}
	if r.SetsLabel != "" {
		rows = append(rows, [2]string{"Версии комплектов", r.SetsLabel})
	}
	rows = append(rows, [2]string{"Дата отчёта", r.Date})
	if r.Similarity != "" {
		rows = append(rows, [2]string{"Схожесть", r.Similarity + " %"})
	}
	rows = append(rows,
		[2]string{"Расхождений", fmt.Sprint(r.Counts.Total)},
		[2]string{"По критичности", fmt.Sprintf("высокая %d, средняя %d, низкая %d", r.Counts.High, r.Counts.Medium, r.Counts.Low)},
		[2]string{"Решения специалиста", fmt.Sprintf("подтверждено %d, отклонено %d, к проверке %d", r.Counts.Confirmed, r.Counts.Rejected, r.Counts.Total-r.Counts.Confirmed-r.Counts.Rejected)},
	)
	if r.Iteration != nil {
		rows = append(rows, [2]string{"Итерация", fmt.Sprintf("%s (предыдущая: %s): устранено %d, не устранено %d, новых %d",
			firstNonEmpty(r.Iteration.No, "—"), firstNonEmpty(r.Iteration.ParentTitle, "—"), r.Iteration.Resolved, r.Iteration.Persisting, r.Iteration.New)})
	}
	var refs, tgts []string
	for _, d := range r.Ref {
		refs = append(refs, docLine(d))
	}
	for _, d := range r.Target {
		tgts = append(tgts, docLine(d))
	}
	rows = append(rows, [2]string{r.RefName + " (файлы)", strings.Join(refs, "\n")}, [2]string{r.TargetName + " (файлы)", strings.Join(tgts, "\n")})
	if r.Comment != "" {
		rows = append(rows, [2]string{"Комментарий", r.Comment})
	}
	return rows
}

func docLine(d ReportDoc) string {
	s := d.Name
	if d.VolumeCode != "" {
		s += " — " + d.VolumeCode
	}
	if d.Pages != "" {
		s += fmt.Sprintf(" (%s стр.)", d.Pages)
	}
	return s
}

func findingRow(f ReportFinding) []any {
	return []any{f.No, SeverityLabel(f.Severity), KindLabel(f.Kind), DecisionLabel(f.Decision), iterLabel(f.IterState), f.Where, f.Description, f.RemarkNo}
}

var findingHeads = []string{"№", "Критичность", "Вид", "Решение", "Итерация", "Где", "Описание", "Замечание №"}

// ReportXLSX writes the report as an Excel file: Сводка, По томам, Расхождения (+ Оформление, Не выпущено).
func ReportXLSX(r Report) ([]byte, error) {
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
	table := func(sheet string, heads []string, rows [][]any, widths map[string]float64) {
		heads2 := make([]any, len(heads))
		for i, h := range heads {
			heads2[i] = h
		}
		put(sheet, 1, heads2)
		_ = f.SetRowStyle(sheet, 1, 1, bold)
		for i, row := range rows {
			put(sheet, i+2, row)
		}
		if len(rows) > 0 {
			last, _ := excelize.CoordinatesToCellName(len(heads), len(rows)+1)
			_ = f.SetCellStyle(sheet, "A2", last, wrap)
		}
		for col, w := range widths {
			_ = f.SetColWidth(sheet, col, col, w)
		}
		_ = f.SetPanes(sheet, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	}

	const sum = "Сводка"
	f.SetSheetName("Sheet1", sum)
	for i, kv := range r.summaryRows() {
		put(sum, i+1, []any{kv[0], kv[1]})
	}
	_ = f.SetCellStyle(sum, "A1", fmt.Sprintf("B%d", len(r.summaryRows())), wrap)
	_ = f.SetColWidth(sum, "A", "A", 26)
	_ = f.SetColWidth(sum, "B", "B", 100)

	if _, err := f.NewSheet("По томам"); err != nil {
		return nil, err
	}
	var vrows [][]any
	for _, v := range r.Volumes {
		vrows = append(vrows, []any{v.Ref, v.Target, v.Total, v.High, v.Medium, v.Low, v.Confirmed, v.Rejected, v.Open})
	}
	table("По томам", []string{r.RefName, r.TargetName, "Расхождений", "Высокая", "Средняя", "Низкая", "Подтверждено", "Отклонено", "Открыто"}, vrows,
		map[string]float64{"A": 46, "B": 46, "C": 13, "D": 10, "E": 10, "F": 10, "G": 14, "H": 11, "I": 10})

	widths := map[string]float64{"A": 6, "B": 13, "C": 18, "D": 15, "E": 14, "F": 52, "G": 80, "H": 12}
	if _, err := f.NewSheet("Расхождения"); err != nil {
		return nil, err
	}
	var frows [][]any
	for _, x := range r.Findings {
		frows = append(frows, findingRow(x))
	}
	table("Расхождения", findingHeads, frows, widths)
	if len(r.Format) > 0 {
		if _, err := f.NewSheet("Оформление"); err != nil {
			return nil, err
		}
		var rows [][]any
		for _, x := range r.Format {
			rows = append(rows, findingRow(x))
		}
		table("Оформление", findingHeads, rows, widths)
	}
	if len(r.Info) > 0 {
		if _, err := f.NewSheet("Не выпущено, вне объёма"); err != nil {
			return nil, err
		}
		var rows [][]any
		for _, x := range r.Info {
			rows = append(rows, findingRow(x))
		}
		table("Не выпущено, вне объёма", findingHeads, rows, widths)
	}
	if len(r.Params) > 0 {
		if _, err := f.NewSheet("Параметры по томам"); err != nil {
			return nil, err
		}
		var rows [][]any
		for _, p := range r.Params {
			side := r.RefName
			if p.Side != SideRef {
				side = r.TargetName
			}
			rows = append(rows, []any{side, p.Volume, p.Kind, p.Element, strings.Join(p.Values, ", "), p.Page})
		}
		table("Параметры по томам", []string{"Сторона", "Том", "Параметр", "Элемент", "Значения", "Стр."}, rows,
			map[string]float64{"A": 10, "B": 60, "C": 22, "D": 18, "E": 24, "F": 8})
	}
	if len(r.SheetChanges) > 0 {
		if _, err := f.NewSheet("Изменения листов"); err != nil {
			return nil, err
		}
		var rows [][]any
		for _, s := range r.SheetChanges {
			check := "соответствует"
			if !s.OK {
				check = "расхождение"
			}
			rows = append(rows, []any{s.Volume, s.Sheet, s.Change, SheetStateLabels[s.Fact], SheetStateLabels[s.Declared], SheetStateLabels[s.Stamp], s.Diffs, check, s.Note})
		}
		table("Изменения листов", []string{"Том", "Лист", "Изм.", "Фактически", "По таблице регистрации", "По основной надписи", "Расхождений", "Проверка", "Пояснение"}, rows,
			map[string]float64{"A": 40, "B": 7, "C": 6, "D": 16, "E": 20, "F": 18, "G": 12, "H": 14, "I": 70})
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func docxTable(b *strings.Builder, widths []int, heads []string, rows [][]string) {
	b.WriteString(`<w:tbl><w:tblPr><w:tblW w:w="0" w:type="auto"/><w:tblBorders>`)
	for _, side := range []string{"top", "left", "bottom", "right", "insideH", "insideV"} {
		fmt.Fprintf(b, `<w:%s w:val="single" w:sz="4" w:space="0" w:color="808080"/>`, side)
	}
	b.WriteString(`</w:tblBorders></w:tblPr><w:tblGrid>`)
	for _, w := range widths {
		fmt.Fprintf(b, `<w:gridCol w:w="%d"/>`, w)
	}
	b.WriteString(`</w:tblGrid><w:tr><w:trPr><w:tblHeader/></w:trPr>`)
	for i, h := range heads {
		b.WriteString(docxCell(h, widths[i], true))
	}
	b.WriteString(`</w:tr>`)
	for _, row := range rows {
		b.WriteString(`<w:tr>`)
		for i, v := range row {
			b.WriteString(docxCell(v, widths[i], false))
		}
		b.WriteString(`</w:tr>`)
	}
	b.WriteString(`</w:tbl>`)
}

func findingStrings(list []ReportFinding) [][]string {
	var rows [][]string
	for _, f := range list {
		rows = append(rows, []string{fmt.Sprint(f.No), SeverityLabel(f.Severity), KindLabel(f.Kind), DecisionLabel(f.Decision), f.Where, f.Description, f.RemarkNo})
	}
	return rows
}

// ReportDOCX writes the report as a Word document.
func ReportDOCX(r Report) ([]byte, error) {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`)
	b.WriteString(docxPara("Отчёт о сравнении документации", true, 30))
	for _, kv := range r.summaryRows() {
		b.WriteString(docxPara(kv[0]+": "+kv[1], false, 22))
	}
	b.WriteString(docxPara("Эталон — "+r.RefName+"; "+r.TargetName+" проверяется на соответствие. Отчёт носит справочный характер: решение о замечаниях принимает специалист.", false, 20))

	if len(r.Volumes) > 0 {
		b.WriteString(docxPara("Расхождения по документам", true, 24))
		var rows [][]string
		for _, v := range r.Volumes {
			rows = append(rows, []string{v.Ref, v.Target, fmt.Sprint(v.Total), fmt.Sprintf("%d / %d / %d", v.High, v.Medium, v.Low), fmt.Sprint(v.Open)})
		}
		docxTable(&b, []int{4200, 4200, 1300, 2200, 1100}, []string{r.RefName, r.TargetName, "Всего", "Выс. / ср. / низк.", "Открыто"}, rows)
	}
	heads := []string{"№", "Критичность", "Вид", "Решение", "Где", "Описание", "Замечание №"}
	widths := []int{500, 1300, 1500, 1400, 3600, 5200, 1100}
	if len(r.Findings) > 0 {
		b.WriteString(docxPara("Расхождения", true, 24))
		docxTable(&b, widths, heads, findingStrings(r.Findings))
	}
	if len(r.Format) > 0 {
		b.WriteString(docxPara("Оформление", true, 24))
		docxTable(&b, widths, heads, findingStrings(r.Format))
	}
	if len(r.Info) > 0 {
		b.WriteString(docxPara("Не выпущено и вне объёма (не считается ошибкой)", true, 24))
		docxTable(&b, widths, heads, findingStrings(r.Info))
	}
	b.WriteString(`<w:sectPr><w:pgSz w:w="16838" w:h="11906" w:orient="landscape"/><w:pgMar w:top="1000" w:right="1000" w:bottom="1000" w:left="1000"/></w:sectPr></w:body></w:document>`)
	return docxZip(b.String())
}
