package envoy

import (
	"context"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/madnikulin50/lowcode/server/pkg/envoyx"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

const (
	pdfFontFamily = "gofont"

	pdfMargin      = 8.0
	pdfLineHeight  = 4.0
	pdfCellPadding = 1.0
	// Guards against a single huge text value producing an unreadable page
	pdfMaxCellRunes = 400
)

type (
	PdfEncoder struct{}

	pdfTable struct {
		doc    *fpdf.Fpdf
		widths []float64
		header []string
	}
)

func (e PdfEncoder) Encode(ctx context.Context, p envoyx.EncodeParams, rt string, nodes envoyx.NodeSet, tt envoyx.Traverser) (err error) {
	if rt != ComposeRecordDatasourceAuxType {
		return
	}

	w, err := getEncoderWriter(p, "pdf")
	if err != nil {
		return
	}

	// Go fonts cover Latin and Cyrillic, so no external font files are needed
	doc := fpdf.New("L", "mm", "A4", "")
	doc.SetMargins(pdfMargin, pdfMargin, pdfMargin)
	doc.SetAutoPageBreak(false, pdfMargin)
	doc.SetCellMargin(pdfCellPadding)
	doc.AddUTF8FontFromBytes(pdfFontFamily, "", goregular.TTF)
	doc.AddUTF8FontFromBytes(pdfFontFamily, "B", gobold.TTF)
	doc.AddPage()

	t := &pdfTable{doc: doc}

	for _, n := range nodes {
		err = streamRecordTable(ctx, p, n, t.writeHeader, t.writeRow)
		if err != nil {
			return
		}
	}

	if doc.PageNo() == 1 && t.header == nil {
		doc.SetFont(pdfFontFamily, "", 10)
		doc.Cell(0, 8, "No records")
	}

	if err = doc.Error(); err != nil {
		return
	}

	return doc.Output(w)
}

func (t *pdfTable) writeHeader(header []string) error {
	t.header = header

	pageW, _ := t.doc.GetPageSize()
	usable := pageW - 2*pdfMargin

	t.widths = make([]float64, len(header))
	for i := range t.widths {
		t.widths[i] = usable / float64(len(header))
	}

	// Many columns need smaller text to remain readable
	size := 8.0
	switch {
	case len(header) > 18:
		size = 5
	case len(header) > 12:
		size = 6
	case len(header) > 8:
		size = 7
	}
	t.doc.SetFont(pdfFontFamily, "B", size)

	t.drawRow(header, true)
	return nil
}

func (t *pdfTable) writeRow(row []string) error {
	t.drawRow(row, false)
	return nil
}

func (t *pdfTable) drawRow(cells []string, isHeader bool) {
	doc := t.doc

	style := ""
	if isHeader {
		style = "B"
	}
	size, _ := doc.GetFontSize()
	doc.SetFont(pdfFontFamily, style, size)

	_, pageH := doc.GetPageSize()
	maxHeight := pageH - 2*pdfMargin

	lines := make([][]string, len(cells))
	maxLines := 1
	for i, c := range cells {
		if r := []rune(c); len(r) > pdfMaxCellRunes {
			c = string(r[:pdfMaxCellRunes]) + "..."
		}

		lines[i] = t.splitLines(c, t.widths[i]-2*pdfCellPadding)
		if len(lines[i]) > maxLines {
			maxLines = len(lines[i])
		}
	}

	height := float64(maxLines) * pdfLineHeight
	if height > maxHeight {
		maxLines = int(maxHeight / pdfLineHeight)
		height = float64(maxLines) * pdfLineHeight
	}

	if doc.GetY()+height > pageH-pdfMargin {
		doc.AddPage()
		if !isHeader && t.header != nil {
			t.drawRow(t.header, true)
			doc.SetFont(pdfFontFamily, "", size)
		}
	}

	if isHeader {
		doc.SetFillColor(230, 230, 230)
	}

	x, y := pdfMargin, doc.GetY()
	for i := range cells {
		doc.Rect(x, y, t.widths[i], height, map[bool]string{true: "DF", false: "D"}[isHeader])

		for li, line := range lines[i] {
			if li >= maxLines {
				break
			}
			doc.SetXY(x, y+float64(li)*pdfLineHeight)
			doc.CellFormat(t.widths[i], pdfLineHeight, line, "", 0, "L", false, 0, "")
		}

		x += t.widths[i]
	}

	doc.SetXY(pdfMargin, y+height)
}

// splitLines wraps text by explicit newlines first, then by cell width
func (t *pdfTable) splitLines(text string, width float64) (out []string) {
	for _, part := range strings.Split(strings.ReplaceAll(text, "\r", ""), "\n") {
		if part == "" {
			out = append(out, "")
			continue
		}

		out = append(out, t.doc.SplitText(part, width)...)
	}

	return
}
