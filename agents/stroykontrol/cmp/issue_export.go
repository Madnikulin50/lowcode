package cmp

import (
	"archive/zip"
	"bytes"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// IssueMeta describes the list in an export.
type IssueMeta struct {
	ObjectName      string
	ComparisonTitle string
	Date            time.Time
	Mode            string // ModeStages (default) | ModeVersions
}

// exportHeadings are the columns of an expertise export, in its order: a list
// made here can be sent to the designer and, with the answers, loaded back.
var exportHeadings = []string{
	"Вывод о несоответствии", "Ссылка на материалы", "Раздел документации", "Основание", "Ответ на замечание", "Причина отклонения",
	"Предыдущий ответ на замечание", "Предыдущая причина отклонения", "НД", "Эксперт-соисполнитель", "Критичность",
	"Предложения по отработке замечаний", "Описание ожидаемого результата", "Итераций", "Дата последнего изменения", "Статус", "№ замечания",
}

func critLabel(c string) string {
	if c == "critical" {
		return "Критичное"
	}
	return "Некритичное"
}

// IssuedXLSX writes the list as an Excel table with the headings of an
// expertise export («Выгрузка данных»).
func IssuedXLSX(meta IssueMeta, list []Issued) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	const sheet = "Выгрузка данных"
	f.SetSheetName("Sheet1", sheet)
	for c, h := range exportHeadings {
		cell, _ := excelize.CoordinatesToCellName(c+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	date := meta.Date
	if date.IsZero() {
		date = time.Now()
	}
	for r, is := range list {
		row := []any{is.Text, is.Refs, is.Section, is.Basis, "", "", "", "", "", "", critLabel(is.Criticality), "", "", 1,
			date.Format("02.01.2006 15:04"), RemarkStatusLabels[RemarkAwaiting], is.Number}
		for c, v := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}
	wrap, _ := f.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{WrapText: true, Vertical: "top"}})
	bold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}, Alignment: &excelize.Alignment{WrapText: true, Vertical: "top"}})
	_ = f.SetRowStyle(sheet, 1, 1, bold)
	if len(list) > 0 {
		_ = f.SetCellStyle(sheet, "A2", fmt.Sprintf("Q%d", len(list)+1), wrap)
	}
	for col, w := range map[string]float64{"A": 70, "B": 40, "C": 28, "D": 32, "K": 14, "O": 18, "P": 28, "Q": 12} {
		_ = f.SetColWidth(sheet, col, col, w)
	}
	_ = f.SetPanes(sheet, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func xmlText(s string) string {
	var b strings.Builder
	for i, line := range strings.Split(s, "\n") {
		if i > 0 {
			b.WriteString(`<w:br/>`)
		}
		fmt.Fprintf(&b, `<w:t xml:space="preserve">%s</w:t>`, html.EscapeString(line))
	}
	return b.String()
}

func docxPara(text string, bold bool, size int) string {
	rpr := fmt.Sprintf(`<w:rPr>%s<w:sz w:val="%d"/></w:rPr>`, map[bool]string{true: "<w:b/>", false: ""}[bold], size)
	return fmt.Sprintf(`<w:p><w:r>%s%s</w:r></w:p>`, rpr, xmlText(text))
}

func docxCell(text string, w int, bold bool) string {
	rpr := fmt.Sprintf(`<w:rPr>%s<w:sz w:val="20"/></w:rPr>`, map[bool]string{true: "<w:b/>", false: ""}[bold])
	return fmt.Sprintf(`<w:tc><w:tcPr><w:tcW w:w="%d" w:type="dxa"/></w:tcPr><w:p><w:r>%s%s</w:r></w:p></w:tc>`, w, rpr, xmlText(text))
}

// IssuedDOCX writes the list as a Word letter with a table of remarks.
func IssuedDOCX(meta IssueMeta, list []Issued) ([]byte, error) {
	date := meta.Date
	if date.IsZero() {
		date = time.Now()
	}
	widths := []int{700, 3300, 7400, 1700}
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`)
	title, ref := "Замечания по результатам сравнения проектной и рабочей документации", "Эталон — проектная документация стадии П. Рабочая документация должна ей соответствовать либо отклонения должны быть обоснованы."
	if meta.Mode == ModeVersions {
		title, ref = "Замечания по результатам сравнения версий документации", "Эталон — предыдущая версия документации. Новая версия должна ей соответствовать либо изменения должны быть обоснованы."
	}
	b.WriteString(docxPara(title, true, 28))
	if meta.ObjectName != "" {
		b.WriteString(docxPara("Объект: "+meta.ObjectName, false, 22))
	}
	if meta.ComparisonTitle != "" {
		b.WriteString(docxPara("Сравнение: "+meta.ComparisonTitle, false, 22))
	}
	b.WriteString(docxPara("Дата формирования: "+date.Format("02.01.2006"), false, 22))
	b.WriteString(docxPara(ref, false, 22))
	b.WriteString(`<w:tbl><w:tblPr><w:tblW w:w="0" w:type="auto"/><w:tblBorders>`)
	for _, side := range []string{"top", "left", "bottom", "right", "insideH", "insideV"} {
		fmt.Fprintf(&b, `<w:%s w:val="single" w:sz="4" w:space="0" w:color="808080"/>`, side)
	}
	b.WriteString(`</w:tblBorders></w:tblPr><w:tblGrid>`)
	for _, w := range widths {
		fmt.Fprintf(&b, `<w:gridCol w:w="%d"/>`, w)
	}
	b.WriteString(`</w:tblGrid>`)
	heads := []string{"№", "Ссылка на материалы", "Замечание", "Критичность"}
	b.WriteString(`<w:tr><w:trPr><w:tblHeader/></w:trPr>`)
	for i, h := range heads {
		b.WriteString(docxCell(h, widths[i], true))
	}
	b.WriteString(`</w:tr>`)
	for _, is := range list {
		b.WriteString(`<w:tr>`)
		for i, v := range []string{is.Number, is.Refs, is.Text, critLabel(is.Criticality)} {
			b.WriteString(docxCell(v, widths[i], false))
		}
		b.WriteString(`</w:tr>`)
	}
	b.WriteString(`</w:tbl><w:sectPr><w:pgSz w:w="16838" w:h="11906" w:orient="landscape"/><w:pgMar w:top="1000" w:right="1000" w:bottom="1000" w:left="1000"/></w:sectPr></w:body></w:document>`)

	return docxZip(b.String())
}

// docxZip packs a document body into a minimal .docx file.
func docxZip(body string) ([]byte, error) {
	var out bytes.Buffer
	zw := zip.NewWriter(&out)
	files := []struct{ name, body string }{
		{"[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`},
		{"_rels/.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`},
		{"word/document.xml", body},
	}
	for _, f := range files {
		w, err := zw.Create(f.name)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write([]byte(f.body)); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
