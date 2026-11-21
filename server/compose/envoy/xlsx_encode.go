package envoy

import (
	"context"
	"fmt"

	"github.com/madnikulin50/lowcode/server/pkg/envoyx"
	"github.com/xuri/excelize/v2"
)

const xlsxSheetName = "Records"

type (
	XlsxEncoder struct{}
)

func (e XlsxEncoder) Encode(ctx context.Context, p envoyx.EncodeParams, rt string, nodes envoyx.NodeSet, tt envoyx.Traverser) (err error) {
	if rt != ComposeRecordDatasourceAuxType {
		return
	}

	w, err := getEncoderWriter(p, "xlsx")
	if err != nil {
		return
	}

	f := excelize.NewFile()
	defer f.Close()

	if err = f.SetSheetName("Sheet1", xlsxSheetName); err != nil {
		return
	}

	sw, err := f.NewStreamWriter(xlsxSheetName)
	if err != nil {
		return
	}

	borderColor := "BFBFBF"
	borders := []excelize.Border{
		{Type: "left", Color: borderColor, Style: 1},
		{Type: "right", Color: borderColor, Style: 1},
		{Type: "top", Color: borderColor, Style: 1},
		{Type: "bottom", Color: borderColor, Style: 1},
	}

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"E7ECF3"}, Pattern: 1},
		Border:    borders,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	if err != nil {
		return
	}

	cellStyle, err := f.NewStyle(&excelize.Style{
		Border:    borders,
		Alignment: &excelize.Alignment{Vertical: "top", WrapText: true},
	})
	if err != nil {
		return
	}

	// Stream writer requires rows to be written in ascending order
	rowNum := 1
	writeRow := func(values []string, styleID int) error {
		cells := make([]interface{}, len(values))
		for i, v := range values {
			cells[i] = excelize.Cell{StyleID: styleID, Value: v}
		}

		cell, err := excelize.CoordinatesToCellName(1, rowNum)
		if err != nil {
			return err
		}
		rowNum++

		return sw.SetRow(cell, cells)
	}

	for _, n := range nodes {
		err = streamRecordTable(ctx, p, n,
			func(header []string) error {
				// column widths must be set before the first row is written
				if err := sw.SetColWidth(1, len(header), 24); err != nil {
					return err
				}
				return writeRow(header, headerStyle)
			},
			func(row []string) error {
				return writeRow(row, cellStyle)
			},
		)
		if err != nil {
			return
		}
	}

	if err = sw.Flush(); err != nil {
		return fmt.Errorf("could not flush xlsx stream: %w", err)
	}

	return f.Write(w)
}
