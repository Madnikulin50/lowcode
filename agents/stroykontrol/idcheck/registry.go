package idcheck

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// openSheetRows opens the first sheet that satisfies accept (or the first
// sheet) and returns its rows with raw (unformatted) cell values, so dates
// stay serial numbers and quantities keep full precision.
func openSheetRows(data []byte, accept func(rows [][]string) bool) (string, [][]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return "", nil, fmt.Errorf("open xlsx: %w", err)
	}
	defer f.Close()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return "", nil, fmt.Errorf("xlsx has no sheets")
	}
	var firstName string
	var first [][]string
	for i, name := range sheets {
		rows, err := f.GetRows(name, excelize.Options{RawCellValue: true})
		if err != nil {
			continue
		}
		if i == 0 {
			firstName, first = name, rows
		}
		if accept == nil || accept(rows) {
			return name, rows, nil
		}
	}
	return firstName, first, nil
}

func cell(rows [][]string, r, c int) string {
	if r < 0 || r >= len(rows) || c < 0 || c >= len(rows[r]) {
		return ""
	}
	return strings.TrimSpace(rows[r][c])
}

// findNumberingRow locates the header row that numbers the columns
// 1,2,3,…,n (both the registry and КС-2 have one) and returns a map
// columnNumber → zero-based spreadsheet column index.
func findNumberingRow(rows [][]string, minCols int) (int, map[int]int) {
	for r := 0; r < len(rows) && r < 60; r++ {
		cols := map[int]int{}
		expect := 1
		for c, v := range rows[r] {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			n, err := strconv.Atoi(strings.TrimSuffix(v, ".0"))
			if err != nil || n != expect {
				break
			}
			cols[n] = c
			expect++
		}
		if len(cols) >= minCols {
			return r, cols
		}
	}
	return -1, nil
}

var positionRe = regexp.MustCompile(`^\d+(\.\d+)+$|^\d+$`)

// ParseRegistry reads the ИД transfer registry. Columns are located by the
// numbering row (1..17), not by letters, because the template merges cells
// differently from one registry to the next.
func ParseRegistry(data []byte) (*Registry, error) {
	_, rows, err := openSheetRows(data, func(rows [][]string) bool {
		r, _ := findNumberingRow(rows, 13)
		return r >= 0
	})
	if err != nil {
		return nil, err
	}
	hdr, cols := findNumberingRow(rows, 13)
	if hdr < 0 {
		return nil, fmt.Errorf("registry: column numbering row (1..13) not found")
	}
	reg := &Registry{}
	for r := 0; r < hdr; r++ {
		for c := range rows[r] {
			v := cell(rows, r, c)
			switch {
			case strings.HasPrefix(v, "Реестр"):
				reg.Title = v
			case strings.HasPrefix(v, "КС-2"):
				reg.KS2Ref = v
			case strings.HasPrefix(v, "с ") && strings.Contains(v, " по "):
				reg.Period = v
			case strings.HasPrefix(v, "Заказчик"):
				reg.Customer = strings.TrimSpace(strings.TrimPrefix(v, "Заказчик:"))
			}
		}
	}
	col := func(r, n int) string {
		c, ok := cols[n]
		if !ok {
			return ""
		}
		return cell(rows, r, c)
	}
	currentPos := ""
	for r := hdr + 1; r < len(rows); r++ {
		row := RegistryRow{
			Row:        r + 1,
			Seq:        col(r, 1),
			Position:   col(r, 2),
			WorkName:   col(r, 3),
			DocName:    col(r, 4),
			DocNumber:  col(r, 5),
			DocDateRaw: col(r, 6),
			Unit:       col(r, 8),
			QtyVRC:     numPtr(col(r, 9)),
			QtyTotal:   numPtr(col(r, 10)),
			QtyKS2:     numPtr(col(r, 11)),
			QtyProject: numPtr(col(r, 12)),
			QtyFact:    numPtr(col(r, 13)),
		}
		if row.Position == "" && row.DocNumber == "" && row.WorkName == "" {
			continue
		}
		if !positionRe.MatchString(row.Position) && row.DocNumber == "" {
			// signatures / footnotes below the table ("* Представитель ...")
			continue
		}
		row.DocDate, _ = ParseDate(row.DocDateRaw)
		switch {
		case row.DocNumber != "" || row.DocName != "":
			row.Kind = RowAct
			if row.Position == "" {
				row.Position = currentPos
			}
		case row.Unit != "" || row.QtyVRC != nil || row.QtyKS2 != nil:
			row.Kind = RowPosition
			currentPos = row.Position
		default:
			row.Kind = RowSection
		}
		reg.Rows = append(reg.Rows, row)
	}
	return reg, nil
}

// Positions returns the position rows keyed by position code.
func (r *Registry) Positions() map[string]*RegistryRow {
	out := map[string]*RegistryRow{}
	for i := range r.Rows {
		if r.Rows[i].Kind == RowPosition {
			out[r.Rows[i].Position] = &r.Rows[i]
		}
	}
	return out
}

// Acts returns the act rows.
func (r *Registry) Acts() []*RegistryRow {
	var out []*RegistryRow
	for i := range r.Rows {
		if r.Rows[i].Kind == RowAct {
			out = append(out, &r.Rows[i])
		}
	}
	return out
}

// ParseKS2 reads a КС-2 workbook: the numbering row is 1..8 (№ по порядку,
// позиция по смете, наименование, расценка, ед., количество, цена, стоимость).
func ParseKS2(data []byte) (*KS2, error) {
	_, rows, err := openSheetRows(data, func(rows [][]string) bool {
		r, _ := findNumberingRow(rows, 8)
		return r >= 0
	})
	if err != nil {
		return nil, err
	}
	hdr, cols := findNumberingRow(rows, 8)
	if hdr < 0 {
		return nil, fmt.Errorf("КС-2: column numbering row (1..8) not found")
	}
	ks := &KS2{}
	// header block: "Номер документа | Дата составления | Отчетный период с | по"
	for r := 0; r < hdr; r++ {
		for c := range rows[r] {
			if cell(rows, r, c) == "Номер документа" && r+2 < len(rows) {
				ks.Number = cell(rows, r+2, c)
				ks.Date, _ = ParseDate(cell(rows, r+2, c+1))
				ks.PeriodFrom, _ = ParseDate(cell(rows, r+2, c+2))
				ks.PeriodTo, _ = ParseDate(cell(rows, r+2, c+3))
			}
		}
	}
	col := func(r, n int) string { return cell(rows, r, cols[n]) }
	for r := hdr + 1; r < len(rows); r++ {
		pos := col(r, 2)
		if !positionRe.MatchString(pos) {
			continue
		}
		qty, ok := ParseNum(col(r, 6))
		if !ok {
			continue
		}
		price, _ := ParseNum(col(r, 7))
		amount, _ := ParseNum(col(r, 8))
		ks.Rows = append(ks.Rows, KS2Row{
			Row: r + 1, Seq: col(r, 1), Position: pos, Name: col(r, 3),
			Unit: col(r, 5), Qty: qty, Price: price, Amount: amount,
		})
	}
	return ks, nil
}
