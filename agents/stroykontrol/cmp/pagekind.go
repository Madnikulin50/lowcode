package cmp

import (
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

// Свойства страницы PDF (ТЗ 4.2.6): формат и ориентация листа, растровая страница или векторная,
// достоверность распознавания скана.

// ocrWeakConf: a recognized page below this mean confidence (0..100) is said to be weakly recognized.
const ocrWeakConf = 85.0

// paperSides are the sides of the A-series in points (1/72 inch), short side first.
var paperSides = []struct {
	name       string
	short, lng float64
}{
	{"A4", 595, 842}, {"A3", 842, 1191}, {"A2", 1191, 1684}, {"A1", 1684, 2384}, {"A0", 2384, 3370},
}

// SheetFormat names the paper format and orientation of a page: "A3 альбомный", "A4 книжный". CAD
// sheets are often an extended format ("A4×3"), and say so; anything else is given in points.
func SheetFormat(w, h float64) string {
	if w <= 0 || h <= 0 {
		return ""
	}
	short, long := math.Min(w, h), math.Max(w, h)
	orient := "книжный"
	if w > h {
		orient = "альбомный"
	}
	near := func(a, b float64) bool { return math.Abs(a-b) <= 0.06*b }
	for _, p := range paperSides {
		if near(short, p.short) && near(long, p.lng) {
			return p.name + " " + orient
		}
	}
	// an extended format (ГОСТ 2.301): the long side of an A-format by a whole multiple (3…9) of its short side — «А4×3» is 297×630
	for _, p := range paperSides {
		if !near(short, p.lng) {
			continue
		}
		if k := math.Round(long / p.short); k >= 3 && k <= 9 && near(long, k*p.short) {
			return fmt.Sprintf("%s×%d %s", p.name, int(k), orient)
		}
	}
	return fmt.Sprintf("%.0f×%.0f пт %s", w, h, orient)
}

// markRaster marks the pages that carry a full-size image (pdfimages -list). Without pdfimages
// nothing is marked: the pages are then taken for vector ones.
func markRaster(pages []Page, pdfPath string) {
	out, err := exec.Command("pdfimages", "-list", pdfPath).Output()
	if err != nil {
		return
	}
	for i, line := range strings.Split(string(out), "\n") {
		f := strings.Fields(line)
		if i < 2 || len(f) < 5 || f[2] != "image" {
			continue
		}
		pg, e1 := strconv.Atoi(f[0])
		w, e2 := strconv.Atoi(f[3])
		h, e3 := strconv.Atoi(f[4])
		if e1 != nil || e2 != nil || e3 != nil || pg < 1 || pg > len(pages) || w < 600 || h < 600 {
			continue // logos and stamps are small images
		}
		pages[pg-1].Raster = true
	}
}

// scanNotes tells the specialist where the text of a document rests on recognition and how far to
// trust it: pages read from a scan with a weak confidence, scans that were not trusted at all, and the
// number of raster pages. Nothing is said about documents that are plain vector PDFs.
func scanNotes(docs []*Document) []string {
	var out []string
	for _, d := range docs {
		var weak, rejected []int
		raster := 0
		for i, p := range d.Pages {
			if p.Raster {
				raster++
			}
			switch {
			case p.OCRReject:
				rejected = append(rejected, i+1)
			case p.OCR && p.OCRConf > 0 && p.OCRConf < ocrWeakConf:
				weak = append(weak, i+1)
			}
		}
		if len(weak) > 0 {
			out = append(out, fmt.Sprintf("«%s»: стр. %s распознаны со сниженной достоверностью (ниже %.0f %%) — находки по ним ненадёжны, проверьте по скану.", d.Name, pageList(weak), ocrWeakConf))
		}
		if len(rejected) > 0 {
			out = append(out, fmt.Sprintf("«%s»: стр. %s — скан не распознан (низкая достоверность текста), листы не сравнивались по тексту.", d.Name, pageList(rejected)))
		}
		if raster > 0 && raster == len(d.Pages) && len(d.Pages) > 1 {
			out = append(out, fmt.Sprintf("«%s»: все страницы растровые (скан).", d.Name))
		}
	}
	return out
}

func pageList(p []int) string {
	var parts []string
	for i, n := range p {
		if i == 8 {
			parts = append(parts, fmt.Sprintf("… (всего %d)", len(p)))
			break
		}
		parts = append(parts, strconv.Itoa(n))
	}
	return strings.Join(parts, ", ")
}
