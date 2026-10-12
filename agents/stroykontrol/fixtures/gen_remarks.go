//go:build ignore

// Generates fixtures/pack-1/remarks.xlsx and fixtures/pack-1-v2/remarks.xlsx:
// a table of expertise remarks with the designer's answers, as it arrives
// from the expertise. Run: go run fixtures/gen_remarks.go
package main

import (
	"log"
	"path/filepath"

	"github.com/xuri/excelize/v2"
)

func write(path string, rows [][]string) {
	f := excelize.NewFile()
	const sheet = "Замечания"
	f.SetSheetName("Sheet1", sheet)
	for r, row := range rows {
		for c, v := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+1)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}
	_ = f.SetColWidth(sheet, "B", "D", 48)
	if err := f.SaveAs(path); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote", path)
}

func main() {
	rows := [][]string{
		{"Замечания экспертизы по разделу «Конструктивные решения»"},
		{},
		{"№ п/п", "Раздел, лист", "Содержание замечания", "Ответ проектировщика"},
		{"Раздел КЖ. Фундаменты"},
		{"1", "КЖ1, лист 1", "Глубина заложения фундамента 2,6 м не соответствует стадии П (2,4 м). Обосновать.", "Замечание рассмотрено. В рабочую документацию внесены корректировки, глубина заложения соответствует стадии П."},
		{"2", "КЖ1, лист 2", "Количество погонных метров ленточного фундамента Ф-2 (52 м) отличается от стадии П (48 м).", "Количество уточнено по результатам детальной проработки."},
		{"3", "КЖ1, лист 1", "Класс бетона фундаментов В30 не соответствует проекту стадии П (В25).", "Учтено, принят класс В25."},
		{"4", "АР, лист 2", "Не представлены расчёты теплопередачи ограждающих конструкций.", ""},
	}
	write(filepath.Join("fixtures", "pack-1", "remarks.xlsx"), rows)
	write(filepath.Join("fixtures", "pack-1-v2", "remarks.xlsx"), rows)
}
