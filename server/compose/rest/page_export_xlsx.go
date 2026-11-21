package rest

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/xuri/excelize/v2"
)

// pageExportXLSX builds a workbook with one sheet per list found on the page.
// Unlike print, all loaded rows are exported.
func pageExportXLSX(data pageExportData) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	borders := []excelize.Border{
		{Type: "left", Color: "#B7B7B7", Style: 1},
		{Type: "right", Color: "#B7B7B7", Style: 1},
		{Type: "top", Color: "#B7B7B7", Style: 1},
		{Type: "bottom", Color: "#B7B7B7", Style: 1},
	}

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#DDE6F0"}, Pattern: 1},
		Alignment: &excelize.Alignment{WrapText: true, Vertical: "center", Horizontal: "center"},
		Border:    borders,
	})
	if err != nil {
		return nil, err
	}

	bodyStyle, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{WrapText: true, Vertical: "top"},
		Border:    borders,
	})
	if err != nil {
		return nil, err
	}

	titleStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14},
	})
	if err != nil {
		return nil, err
	}

	if len(data.Sections) == 0 {
		sheet := f.GetSheetName(0)
		_ = f.SetCellValue(sheet, "A1", data.Title)
		_ = f.SetCellValue(sheet, "A2", "Записи отсутствуют")
		return xlsxBytes(f)
	}

	used := map[string]bool{}

	for i, s := range data.Sections {
		name := xlsxSheetName(s, i+1, used)

		if i == 0 {
			if err = f.SetSheetName(f.GetSheetName(0), name); err != nil {
				return nil, err
			}
		} else if _, err = f.NewSheet(name); err != nil {
			return nil, err
		}

		row := 1
		if s.Title != "" {
			_ = f.SetCellValue(name, "A1", s.Title)
			_ = f.SetCellStyle(name, "A1", "A1", titleStyle)
			row = 3
		}

		headerRow := row
		for c, h := range s.Headers {
			cell, _ := excelize.CoordinatesToCellName(c+1, row)
			_ = f.SetCellValue(name, cell, h)
			_ = f.SetCellStyle(name, cell, cell, headerStyle)
		}
		row++

		widths := make([]int, len(s.Headers))
		for c, h := range s.Headers {
			widths[c] = utf8.RuneCountInString(h)
		}

		for _, r := range s.Rows {
			for c, v := range r {
				cell, _ := excelize.CoordinatesToCellName(c+1, row)
				_ = f.SetCellValue(name, cell, v)
				_ = f.SetCellStyle(name, cell, cell, bodyStyle)
				if n := utf8.RuneCountInString(v); c < len(widths) && n > widths[c] {
					widths[c] = n
				}
			}
			row++
		}

		switch {
		case len(s.Rows) == 0:
			_ = f.SetCellValue(name, fmt.Sprintf("A%d", row), "Записи отсутствуют")
		case s.Total > len(s.Rows):
			_ = f.SetCellValue(name, fmt.Sprintf("A%d", row), fmt.Sprintf("и еще %d", s.Total-len(s.Rows)))
		}

		for c, w := range widths {
			col, _ := excelize.ColumnNumberToName(c + 1)
			if w < 10 {
				w = 10
			}
			if w > 60 {
				w = 60
			}
			_ = f.SetColWidth(name, col, col, float64(w)+2)
		}

		if len(s.Headers) > 0 {
			_ = f.SetPanes(name, &excelize.Panes{
				Freeze:      true,
				YSplit:      headerRow,
				TopLeftCell: fmt.Sprintf("A%d", headerRow+1),
				ActivePane:  "bottomLeft",
			})
		}
	}

	return xlsxBytes(f)
}

func xlsxBytes(f *excelize.File) ([]byte, error) {
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// xlsxSheetName builds a unique valid (<= 31 chars, no special symbols) sheet name
func xlsxSheetName(s pageExportSection, n int, used map[string]bool) string {
	base := strings.TrimSpace(s.Tab + " " + s.Title)
	if s.Tab != "" && s.Title != "" {
		base = s.Tab + " - " + s.Title
	}
	if base == "" {
		base = fmt.Sprintf("Список %d", n)
	}

	base = strings.Map(func(r rune) rune {
		if strings.ContainsRune(`[]:*?/\`, r) {
			return ' '
		}
		return r
	}, base)

	for i := 0; ; i++ {
		name := base
		suffix := ""
		if i > 0 {
			suffix = fmt.Sprintf(" (%d)", i+1)
		}

		runes := []rune(name)
		if len(runes) > 31-len([]rune(suffix)) {
			runes = runes[:31-len([]rune(suffix))]
		}
		name = strings.TrimSpace(string(runes)) + suffix

		if !used[strings.ToLower(name)] {
			used[strings.ToLower(name)] = true
			return name
		}
	}
}
