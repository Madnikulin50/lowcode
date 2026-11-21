package rest

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

const (
	pptxSlideW = 12192000
	pptxSlideH = 6858000
	pptxMargin = 457200

	// same limit as the print version of RecordList
	pptxRowsLimit = 7
	pptxMaxCols   = 8
	pptxMaxCell   = 120
)

const (
	pptxNS = `xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" ` +
		`xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" ` +
		`xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"`

	pptxRelNS = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
	xmlHeader = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n"
)

// pageExportPPTX builds a presentation with one slide per list found on the page.
// Follows the print rules: the first rows only and the "и еще N" tail.
func pageExportPPTX(data pageExportData) ([]byte, error) {
	slides := make([]string, 0, len(data.Sections))

	if len(data.Sections) == 0 {
		slides = append(slides, pptxSlide(data.Title, "", nil))
	}

	for _, s := range data.Sections {
		title := strings.TrimSpace(s.Title)
		if s.Tab != "" {
			if title != "" {
				title = s.Tab + " — " + title
			} else {
				title = s.Tab
			}
		}
		slides = append(slides, pptxSlide(title, data.Title, &s))
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	files := []struct{ name, body string }{
		{"[Content_Types].xml", pptxContentTypes(len(slides))},
		{"_rels/.rels", xmlHeader + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
			`<Relationship Id="rId1" Type="` + pptxRelNS + `/officeDocument" Target="ppt/presentation.xml"/></Relationships>`},
		{"ppt/presentation.xml", pptxPresentation(len(slides))},
		{"ppt/_rels/presentation.xml.rels", pptxPresentationRels(len(slides))},
		{"ppt/slideMasters/slideMaster1.xml", pptxMaster},
		{"ppt/slideMasters/_rels/slideMaster1.xml.rels", pptxRels(
			[3]string{"rId1", "slideLayout", "../slideLayouts/slideLayout1.xml"},
			[3]string{"rId2", "theme", "../theme/theme1.xml"},
		)},
		{"ppt/slideLayouts/slideLayout1.xml", pptxLayout},
		{"ppt/slideLayouts/_rels/slideLayout1.xml.rels", pptxRels(
			[3]string{"rId1", "slideMaster", "../slideMasters/slideMaster1.xml"},
		)},
		{"ppt/theme/theme1.xml", pptxTheme},
	}

	for i, s := range slides {
		n := i + 1
		files = append(files,
			struct{ name, body string }{fmt.Sprintf("ppt/slides/slide%d.xml", n), s},
			struct{ name, body string }{fmt.Sprintf("ppt/slides/_rels/slide%d.xml.rels", n), pptxRels(
				[3]string{"rId1", "slideLayout", "../slideLayouts/slideLayout1.xml"},
			)},
		)
	}

	for _, f := range files {
		w, err := zw.Create(f.name)
		if err != nil {
			return nil, err
		}
		if _, err = w.Write([]byte(f.body)); err != nil {
			return nil, err
		}
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func pptxEsc(s string) string {
	s = strings.Map(func(r rune) rune {
		if r < 0x20 && r != '\t' && r != '\n' && r != '\r' {
			return -1
		}
		return r
	}, s)

	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

func pptxTrim(s string) string {
	r := []rune(s)
	if len(r) > pptxMaxCell {
		return string(r[:pptxMaxCell-1]) + "…"
	}
	return s
}

func pptxRels(rels ...[3]string) string {
	var b strings.Builder
	b.WriteString(xmlHeader + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for _, r := range rels {
		fmt.Fprintf(&b, `<Relationship Id="%s" Type="%s/%s" Target="%s"/>`, r[0], pptxRelNS, r[1], r[2])
	}
	b.WriteString(`</Relationships>`)
	return b.String()
}

func pptxContentTypes(slides int) string {
	var b strings.Builder
	b.WriteString(xmlHeader + `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">` +
		`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>` +
		`<Default Extension="xml" ContentType="application/xml"/>` +
		`<Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/>` +
		`<Override PartName="/ppt/slideMasters/slideMaster1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideMaster+xml"/>` +
		`<Override PartName="/ppt/slideLayouts/slideLayout1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideLayout+xml"/>` +
		`<Override PartName="/ppt/theme/theme1.xml" ContentType="application/vnd.openxmlformats-officedocument.theme+xml"/>`)
	for i := 1; i <= slides; i++ {
		fmt.Fprintf(&b, `<Override PartName="/ppt/slides/slide%d.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>`, i)
	}
	b.WriteString(`</Types>`)
	return b.String()
}

func pptxPresentation(slides int) string {
	var ids strings.Builder
	for i := 1; i <= slides; i++ {
		fmt.Fprintf(&ids, `<p:sldId id="%d" r:id="rId%d"/>`, 255+i, i+2)
	}

	return xmlHeader + `<p:presentation ` + pptxNS + `>` +
		`<p:sldMasterIdLst><p:sldMasterId id="2147483648" r:id="rId1"/></p:sldMasterIdLst>` +
		`<p:sldIdLst>` + ids.String() + `</p:sldIdLst>` +
		fmt.Sprintf(`<p:sldSz cx="%d" cy="%d"/>`, pptxSlideW, pptxSlideH) +
		`<p:notesSz cx="6858000" cy="9144000"/></p:presentation>`
}

func pptxPresentationRels(slides int) string {
	rels := [][3]string{
		{"rId1", "slideMaster", "slideMasters/slideMaster1.xml"},
		{"rId2", "theme", "theme/theme1.xml"},
	}
	for i := 1; i <= slides; i++ {
		rels = append(rels, [3]string{fmt.Sprintf("rId%d", i+2), "slide", fmt.Sprintf("slides/slide%d.xml", i)})
	}
	return pptxRels(rels...)
}

func pptxText(id int, name, text string, x, y, cx, cy, size int, bold bool) string {
	b := 0
	if bold {
		b = 1
	}
	return fmt.Sprintf(`<p:sp><p:nvSpPr><p:cNvPr id="%d" name="%s"/><p:cNvSpPr txBox="1"/><p:nvPr/></p:nvSpPr>`+
		`<p:spPr><a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></p:spPr>`+
		`<p:txBody><a:bodyPr wrap="square" rtlCol="0"><a:normAutofit/></a:bodyPr><a:lstStyle/>`+
		`<a:p><a:r><a:rPr lang="ru-RU" sz="%d" b="%d"><a:solidFill><a:srgbClr val="1F3A4D"/></a:solidFill></a:rPr><a:t>%s</a:t></a:r></a:p></p:txBody></p:sp>`,
		id, name, x, y, cx, cy, size, b, pptxEsc(text))
}

func pptxCell(text string, size int, header bool, span, hMerge bool) string {
	attr := ""
	if span {
		attr = ` gridSpan="%d"`
	}
	if hMerge {
		attr = ` hMerge="1"`
	}

	fill := "FFFFFF"
	bold := 0
	if header {
		fill, bold = "EEEEEE", 1
	}

	line := func(tag string) string {
		return `<a:` + tag + ` w="6350"><a:solidFill><a:srgbClr val="BBBBBB"/></a:solidFill></a:` + tag + `>`
	}

	body := `<a:p><a:endParaRPr lang="ru-RU" sz="1000"/></a:p>`
	if text != "" {
		body = fmt.Sprintf(`<a:p><a:r><a:rPr lang="ru-RU" sz="%d" b="%d"><a:solidFill><a:srgbClr val="222222"/></a:solidFill></a:rPr><a:t>%s</a:t></a:r></a:p>`,
			size, bold, pptxEsc(pptxTrim(text)))
	}

	return `<a:tc` + attr + `><a:txBody><a:bodyPr/><a:lstStyle/>` + body + `</a:txBody>` +
		`<a:tcPr marL="45720" marR="45720" marT="22860" marB="22860">` +
		line("lnL") + line("lnR") + line("lnT") + line("lnB") +
		`<a:solidFill><a:srgbClr val="` + fill + `"/></a:solidFill></a:tcPr></a:tc>`
}

// pptxSlide renders one slide: title, optional page name and a table
func pptxSlide(title, subtitle string, s *pageExportSection) string {
	var shapes strings.Builder

	shapes.WriteString(pptxText(2, "Title", title, pptxMargin, 250000, pptxSlideW-2*pptxMargin, 520000, 2400, true))
	if subtitle != "" {
		shapes.WriteString(pptxText(3, "Subtitle", subtitle, pptxMargin, 760000, pptxSlideW-2*pptxMargin, 320000, 1200, false))
	}

	if s != nil {
		cols := len(s.Headers)
		if cols > pptxMaxCols {
			cols = pptxMaxCols
		}

		if cols == 0 {
			shapes.WriteString(pptxText(4, "Empty", "Записи отсутствуют", pptxMargin, 1300000, pptxSlideW-2*pptxMargin, 400000, 1400, false))
		} else {
			size := 800
			switch {
			case cols <= 4:
				size = 1200
			case cols <= 6:
				size = 1000
			}

			var tbl strings.Builder
			colW := (pptxSlideW - 2*pptxMargin) / cols
			tbl.WriteString(`<a:tbl><a:tblPr firstRow="1"/><a:tblGrid>`)
			for i := 0; i < cols; i++ {
				fmt.Fprintf(&tbl, `<a:gridCol w="%d"/>`, colW)
			}
			tbl.WriteString(`</a:tblGrid>`)

			row := func(cells []string, header bool) {
				tbl.WriteString(`<a:tr h="370840">`)
				for i := 0; i < cols; i++ {
					v := ""
					if i < len(cells) {
						v = cells[i]
					}
					tbl.WriteString(pptxCell(v, size, header, false, false))
				}
				tbl.WriteString(`</a:tr>`)
			}

			// one merged cell across the whole row
			note := func(text string) {
				tbl.WriteString(`<a:tr h="370840">`)
				for i := 0; i < cols; i++ {
					switch {
					case i == 0 && cols > 1:
						tbl.WriteString(strings.Replace(pptxCell(text, size, false, true, false), `gridSpan="%d"`, fmt.Sprintf(`gridSpan="%d"`, cols), 1))
					case i == 0:
						tbl.WriteString(pptxCell(text, size, false, false, false))
					default:
						tbl.WriteString(pptxCell("", size, false, false, true))
					}
				}
				tbl.WriteString(`</a:tr>`)
			}

			row(s.Headers, true)

			rows := s.Rows
			if len(rows) > pptxRowsLimit {
				rows = rows[:pptxRowsLimit]
			}
			for _, r := range rows {
				row(r, false)
			}

			switch {
			case len(s.Rows) == 0:
				note("Записи отсутствуют")
			case s.Total > len(rows):
				note(fmt.Sprintf("и еще %d", s.Total-len(rows)))
			}
			tbl.WriteString(`</a:tbl>`)

			height := 370840 * (len(rows) + 2)
			fmt.Fprintf(&shapes, `<p:graphicFrame><p:nvGraphicFramePr><p:cNvPr id="4" name="Table"/>`+
				`<p:cNvGraphicFramePr><a:graphicFrameLocks noGrp="1"/></p:cNvGraphicFramePr><p:nvPr/></p:nvGraphicFramePr>`+
				`<p:xfrm><a:off x="%d" y="1250000"/><a:ext cx="%d" cy="%d"/></p:xfrm>`+
				`<a:graphic><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/table">%s</a:graphicData></a:graphic></p:graphicFrame>`,
				pptxMargin, colW*cols, height, tbl.String())
		}
	}

	return xmlHeader + `<p:sld ` + pptxNS + `><p:cSld><p:spTree>` +
		`<p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>` +
		`<p:grpSpPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="0" cy="0"/><a:chOff x="0" y="0"/><a:chExt cx="0" cy="0"/></a:xfrm></p:grpSpPr>` +
		shapes.String() +
		`</p:spTree></p:cSld><p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr></p:sld>`
}

const pptxEmptyTree = `<p:cSld><p:spTree><p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>` +
	`<p:grpSpPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="0" cy="0"/><a:chOff x="0" y="0"/><a:chExt cx="0" cy="0"/></a:xfrm></p:grpSpPr></p:spTree></p:cSld>`

var pptxMaster = xmlHeader + `<p:sldMaster ` + pptxNS + `>` + pptxEmptyTree +
	`<p:clrMap bg1="lt1" tx1="dk1" bg2="lt2" tx2="dk2" accent1="accent1" accent2="accent2" accent3="accent3" accent4="accent4" accent5="accent5" accent6="accent6" hlink="hlink" folHlink="folHlink"/>` +
	`<p:sldLayoutIdLst><p:sldLayoutId id="2147483649" r:id="rId1"/></p:sldLayoutIdLst></p:sldMaster>`

var pptxLayout = xmlHeader + `<p:sldLayout ` + pptxNS + ` type="blank" preserve="1">` + pptxEmptyTree +
	`<p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr></p:sldLayout>`

var pptxTheme = xmlHeader + `<a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="Export"><a:themeElements>` +
	`<a:clrScheme name="Export">` +
	`<a:dk1><a:srgbClr val="000000"/></a:dk1><a:lt1><a:srgbClr val="FFFFFF"/></a:lt1>` +
	`<a:dk2><a:srgbClr val="1F3A4D"/></a:dk2><a:lt2><a:srgbClr val="EEEEEE"/></a:lt2>` +
	`<a:accent1><a:srgbClr val="2F6F9F"/></a:accent1><a:accent2><a:srgbClr val="D9822B"/></a:accent2>` +
	`<a:accent3><a:srgbClr val="5B9A5B"/></a:accent3><a:accent4><a:srgbClr val="B5473E"/></a:accent4>` +
	`<a:accent5><a:srgbClr val="7A5C9E"/></a:accent5><a:accent6><a:srgbClr val="888888"/></a:accent6>` +
	`<a:hlink><a:srgbClr val="2F6F9F"/></a:hlink><a:folHlink><a:srgbClr val="7A5C9E"/></a:folHlink></a:clrScheme>` +
	`<a:fontScheme name="Export"><a:majorFont><a:latin typeface="Calibri"/><a:ea typeface=""/><a:cs typeface=""/></a:majorFont>` +
	`<a:minorFont><a:latin typeface="Calibri"/><a:ea typeface=""/><a:cs typeface=""/></a:minorFont></a:fontScheme>` +
	`<a:fmtScheme name="Export">` +
	`<a:fillStyleLst><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:fillStyleLst>` +
	`<a:lnStyleLst><a:ln w="6350"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:ln><a:ln w="12700"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:ln><a:ln w="19050"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:ln></a:lnStyleLst>` +
	`<a:effectStyleLst><a:effectStyle><a:effectLst/></a:effectStyle><a:effectStyle><a:effectLst/></a:effectStyle><a:effectStyle><a:effectLst/></a:effectStyle></a:effectStyleLst>` +
	`<a:bgFillStyleLst><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:bgFillStyleLst>` +
	`</a:fmtScheme></a:themeElements></a:theme>`
