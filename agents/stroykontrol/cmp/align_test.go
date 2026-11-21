package cmp

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"
	"testing"
)

// sheetImage draws a sheet: a frame and some dark blocks given in the fractions of the frame, the whole
// drawing scaled by k and moved by (dx, dy) pixels on a white page of w×h.
func sheetImage(w, h int, k float64, dx, dy int, blocks [][4]float64) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, w, h))
	for i := range img.Pix {
		img.Pix[i] = 255
	}
	fw, fh := int(float64(w-80)*k), int(float64(h-80)*k)
	x0, y0 := 40+dx, 40+dy
	line := func(xa, ya, xb, yb int) {
		for y := ya; y <= yb; y++ {
			for x := xa; x <= xb; x++ {
				if x >= 0 && y >= 0 && x < w && y < h {
					img.SetGray(x, y, color.Gray{0})
				}
			}
		}
	}
	line(x0, y0, x0+fw, y0+2) // frame
	line(x0, y0+fh-2, x0+fw, y0+fh)
	line(x0, y0, x0+2, y0+fh)
	line(x0+fw-2, y0, x0+fw, y0+fh)
	for _, b := range blocks {
		line(x0+int(b[0]*float64(fw)), y0+int(b[1]*float64(fh)), x0+int(b[2]*float64(fw)), y0+int(b[3]*float64(fh)))
	}
	return img
}

var baseBlocks = [][4]float64{{0.1, 0.1, 0.3, 0.2}, {0.5, 0.15, 0.8, 0.3}, {0.2, 0.5, 0.4, 0.7}, {0.6, 0.6, 0.9, 0.85}, {0.1, 0.8, 0.25, 0.9}}

func TestAlignedDiffCancelsAShiftAndAScale(t *testing.T) {
	a := sheetImage(800, 600, 1.0, 0, 0, baseBlocks)
	// the same drawing printed smaller and moved
	b := sheetImage(800, 600, 0.82, 55, 38, baseBlocks)
	if plain := DiffImages(a, b); len(plain) < 3 {
		t.Fatalf("the test needs a plain comparison that is fooled by the shift: %d boxes", len(plain))
	}
	onA, onB, aligned := AlignedDiff(a, b)
	if !aligned || len(onA) != 0 || len(onB) != 0 {
		t.Fatalf("a shifted and rescaled copy is not different: aligned=%v %v %v", aligned, onA, onB)
	}
	// a block added in the second one: found, on both pages, at its own place
	changed := append(append([][4]float64{}, baseBlocks...), [4]float64{0.45, 0.4, 0.55, 0.47})
	c := sheetImage(800, 600, 0.82, 55, 38, changed)
	onA, onB, aligned = AlignedDiff(a, c)
	if !aligned || len(onA) != 1 || len(onB) != 1 {
		t.Fatalf("one added block: aligned=%v %v %v", aligned, onA, onB)
	}
	// on page A the block would be at 0.45..0.55 × 0.4..0.47 of the frame (frame = 40..760 × 40..560)
	want := Box{(40 + 0.45*720) / 800, (40 + 0.4*520) / 600, (40 + 0.55*720) / 800, (40 + 0.47*520) / 600}
	got := onA[0]
	if got.X0 > want.X0+0.03 || got.X1 < want.X1-0.03 || got.Y0 > want.Y0+0.04 || got.Y1 < want.Y1-0.04 {
		t.Errorf("the box on A does not cover the block: got %+v want %+v", got, want)
	}
	// on page B the same block sits at B's own coordinates (smaller and moved)
	wantB := Box{(55 + 40 + 0.45*0.82*720) / 800, (38 + 40 + 0.4*0.82*520) / 600, 0, 0}
	if gb := onB[0]; gb.X0 > wantB.X0+0.04 || gb.Y0 > wantB.Y0+0.05 || gb.X0 < wantB.X0-0.08 {
		t.Errorf("the box on B is in B's page: got %+v near %+v", gb, wantB)
	}
}

func TestAlignedDiffFallsBackWhenThePagesCannotBeAligned(t *testing.T) {
	a := sheetImage(800, 600, 1.0, 0, 0, baseBlocks)
	blank := image.NewGray(image.Rect(0, 0, 800, 600))
	for i := range blank.Pix {
		blank.Pix[i] = 255
	}
	if _, _, aligned := AlignedDiff(a, blank); aligned {
		t.Error("a blank page cannot be aligned")
	}
	// contents of very different proportions: not the same drawing at another scale
	narrow := sheetImage(800, 600, 1.0, 0, 0, nil)
	for y := 0; y < 600; y++ { // cut the frame down to a thin strip
		for x := 300; x < 800; x++ {
			narrow.SetGray(x, y, color.Gray{255})
		}
	}
	if _, _, aligned := AlignedDiff(a, narrow); aligned {
		t.Error("contents of different proportions are compared as whole pages")
	}
	same := sheetImage(800, 600, 1.0, 0, 0, baseBlocks)
	if onA, onB, aligned := AlignedDiff(a, same); !aligned || len(onA)+len(onB) != 0 {
		t.Errorf("identical sheets: %v %v %v", aligned, onA, onB)
	}
	if _, ok := ContentBox(blank); ok {
		t.Error("blank has no content")
	}
}

// drawingDoc is a document of drawing sheets with sheet numbers in the title blocks.
func drawingDoc(id, side string, sheets ...int) *Document {
	var pages []fmtPage
	for range sheets {
		pages = append(pages, fmtPage{0, "", nil})
	}
	for i, n := range sheets {
		pages[i].sheet = n
	}
	d := fmtDoc(id, side, pages...)
	for i := range d.Pages {
		d.Pages[i].Drawing, d.Pages[i].W, d.Pages[i].H = true, 1684, 1191
	}
	return d
}

// imagerOf serves a picture per document and page.
func imagerOf(m map[string]*image.Gray) Imager {
	return func(doc string, page int) (image.Image, error) {
		if img := m[doc+":"+strconv.Itoa(page)]; img != nil {
			return img, nil
		}
		return nil, fmt.Errorf("no picture of %s page %d", doc, page)
	}
}

func drawingFindings(res *Result) (visual, raster []Finding) {
	for _, f := range res.Findings {
		if f.Kind != KindDrawing {
			continue
		}
		if strings.HasPrefix(f.Evidence.Rule, "raster-diff") {
			raster = append(raster, f)
		} else {
			visual = append(visual, f)
		}
	}
	return
}

func TestDrawingSheetsArePairedByTheirNumbers(t *testing.T) {
	pd := drawingDoc("pd", SideRef, 1, 2, 3)
	rd := drawingDoc("rd", SideTarget, 3, 1) // sheet 2 is not released, the order differs
	k := [][4]float64{{0.1, 0.1, 0.3, 0.2}, {0.5, 0.5, 0.7, 0.6}}
	pic := func(extra ...[4]float64) *image.Gray {
		return sheetImage(800, 600, 1.0, 0, 0, append(append([][4]float64{}, k...), extra...))
	}
	other := func() *image.Gray {
		return sheetImage(800, 600, 1.0, 0, 0, [][4]float64{{0.2, 0.6, 0.5, 0.8}, {0.6, 0.1, 0.9, 0.2}})
	}
	imgs := map[string]*image.Gray{
		"pd:1": pic(), "pd:2": other(), "pd:3": other(),
		"rd:1": pic(), // sheet 1 on the РД page 2: unchanged
		"rd:2": sheetImage(800, 600, 1.0, 0, 0, append([][4]float64{{0.2, 0.6, 0.5, 0.8}, {0.6, 0.1, 0.9, 0.2}}, [4]float64{0.4, 0.3, 0.55, 0.4})),
	}
	// РД page 1 is sheet 3 (changed: one more block), page 2 is sheet 1 (as in the ПД)
	imgs["rd:1"], imgs["rd:2"] = imgs["rd:2"], pic()
	res := Compare([]*Document{pd}, []*Document{rd}, Options{NoFormat: true, NoCross: true, Imager: imagerOf(imgs)})
	visual, raster := drawingFindings(res)
	if len(raster) != 1 || raster[0].PageRef != 3 || raster[0].PageTarget != 1 {
		t.Fatalf("only sheet 3 changed (ПД page 3 ↔ РД page 1):\n%+v", raster)
	}
	if !strings.Contains(raster[0].Description, "по номерам") {
		t.Errorf("the finding says the sheets were paired by their numbers: %s", raster[0].Description)
	}
	if raster[0].BoxRef == nil || raster[0].BoxTarget == nil {
		t.Error("a region on both pages")
	}
	if len(visual) != 1 || !strings.Contains(visual[0].Description, "только в ПД — лист 2") {
		t.Errorf("sheet 2 has no pair and is said so: %+v", visual)
	}
}

func TestACopyAtAnotherScaleIsNotADifference(t *testing.T) {
	pd := drawingDoc("pd", SideRef, 1, 2)
	rd := drawingDoc("rd", SideTarget, 1, 2)
	rd.Pages[0].W, rd.Pages[0].H = 1191, 842 // an A3 copy of an A3×… sheet: another paper
	imgs := map[string]*image.Gray{
		"pd:1": sheetImage(800, 600, 1.0, 0, 0, baseBlocks), "pd:2": sheetImage(800, 600, 1.0, 0, 0, baseBlocks[:3]),
		"rd:1": sheetImage(800, 600, 0.8, 50, 30, baseBlocks), "rd:2": sheetImage(800, 600, 1.0, 0, 0, baseBlocks[:3]),
	}
	res := Compare([]*Document{pd}, []*Document{rd}, Options{NoFormat: true, NoCross: true, Imager: imagerOf(imgs)})
	visual, raster := drawingFindings(res)
	if len(raster) != 0 || len(visual) != 0 {
		t.Fatalf("a shifted and rescaled copy is the same drawing:\n%+v\n%+v", raster, visual)
	}
}
