package cmp

import (
	"image"
	"image/color"
	"math"
	"sort"
)

// Imager returns the rendered image of a document page (1-based). Optional:
// without it, pages with no text layer are only reported, not compared.
type Imager func(docID string, page int) (image.Image, error)

const (
	rasterGrid      = 48   // cells per side
	rasterCellDelta = 26.0 // mean gray difference (0..255) that marks a cell as changed
	rasterMinCells  = 2    // smaller blobs are scanner/anti-aliasing noise
	rasterMaxBoxes  = 6
)

func gray(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	return (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 257
}

// cellMeans averages grayscale over a rasterGrid×rasterGrid grid, sampling
// in page fractions so images of different pixel sizes line up.
func cellMeans(img image.Image) [][]float64 { return cellMeansIn(img, Box{0, 0, 1, 1}) }

// cellMeansIn is cellMeans over a part of the page (in page fractions): the grid covers that part only.
func cellMeansIn(img image.Image, area Box) [][]float64 {
	ib := img.Bounds()
	ax0, ay0 := ib.Min.X+int(area.X0*float64(ib.Dx())), ib.Min.Y+int(area.Y0*float64(ib.Dy()))
	b := image.Rect(ax0, ay0, ib.Min.X+int(area.X1*float64(ib.Dx())), ib.Min.Y+int(area.Y1*float64(ib.Dy())))
	if b.Dx() < rasterGrid || b.Dy() < rasterGrid {
		b = ib // too small an area to sample: the whole page
	}
	w, h := b.Dx(), b.Dy()
	out := make([][]float64, rasterGrid)
	for gy := range out {
		out[gy] = make([]float64, rasterGrid)
		for gx := range out[gy] {
			x0, x1 := gx*w/rasterGrid, (gx+1)*w/rasterGrid
			y0, y1 := gy*h/rasterGrid, (gy+1)*h/rasterGrid
			step := 1
			if x1-x0 > 24 {
				step = (x1 - x0) / 12
			}
			var sum float64
			var n int
			for y := y0; y < y1; y += step {
				for x := x0; x < x1; x += step {
					sum += gray(img.At(b.Min.X+x, b.Min.Y+y))
					n++
				}
			}
			if n > 0 {
				out[gy][gx] = sum / float64(n)
			} else {
				out[gy][gx] = 255
			}
		}
	}
	return out
}

// DiffImages finds regions where two page renderings differ: grid cells
// whose mean brightness differs by more than a threshold are grouped into
// 8-connected blobs; each blob larger than the noise floor becomes a box in
// page fractions. Boxes are ordered by area, largest first.
func DiffImages(a, b image.Image) []Box { return changedBlobs(cellMeans(a), cellMeans(b)) }

// changedBlobs groups the cells that differ between two grids into boxes (fractions of the area the grids cover).
func changedBlobs(ma, mb [][]float64) []Box { return blobsOf(diffCells(ma, mb, 0), 1) }

// diffCells marks the cells whose brightness differs by more than the threshold. With tol > 0 a cell
// counts as different only if its brightness is outside the range the other grid takes within tol cells
// around it (widened by the threshold), and the same the other way round: two renderings of one
// drawing never agree to the pixel (resampling, anti-aliasing), and the edge of a line or a filled
// area then falls into the neighbouring cell with a brightness in between, without anything having changed.
func diffCells(ma, mb [][]float64, tol int) [][]bool {
	changed := make([][]bool, rasterGrid)
	span := func(m [][]float64, y, x int) (lo, hi float64) {
		lo, hi = 255, 0
		for dy := -tol; dy <= tol; dy++ {
			for dx := -tol; dx <= tol; dx++ {
				ny, nx := y+dy, x+dx
				if ny < 0 || nx < 0 || ny >= rasterGrid || nx >= rasterGrid {
					continue
				}
				lo, hi = math.Min(lo, m[ny][nx]), math.Max(hi, m[ny][nx])
			}
		}
		return
	}
	outside := func(v, lo, hi float64) bool { return v < lo-rasterCellDelta || v > hi+rasterCellDelta }
	for y := range changed {
		changed[y] = make([]bool, rasterGrid)
		for x := range changed[y] {
			if math.Abs(ma[y][x]-mb[y][x]) <= rasterCellDelta {
				continue
			}
			if tol == 0 {
				changed[y][x] = true
				continue
			}
			lo, hi := span(mb, y, x)
			lo2, hi2 := span(ma, y, x)
			changed[y][x] = outside(ma[y][x], lo, hi) || outside(mb[y][x], lo2, hi2)
		}
	}
	return changed
}

// blobsOf groups changed cells into 8-connected blobs; each blob larger than the noise floor becomes a
// box (fractions of the area the grid covers), the largest first.
func blobsOf(changed [][]bool, minSide int) []Box {
	seen := make([][]bool, rasterGrid)
	for y := range seen {
		seen[y] = make([]bool, rasterGrid)
	}
	type blob struct {
		box   Box
		cells int
	}
	var blobs []blob
	for y := 0; y < rasterGrid; y++ {
		for x := 0; x < rasterGrid; x++ {
			if !changed[y][x] || seen[y][x] {
				continue
			}
			stack := [][2]int{{x, y}}
			seen[y][x] = true
			bl := blob{box: Box{X0: 1, Y0: 1}}
			for len(stack) > 0 {
				c := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				bl.cells++
				bl.box = bl.box.union(Box{
					float64(c[0]) / rasterGrid, float64(c[1]) / rasterGrid,
					float64(c[0]+1) / rasterGrid, float64(c[1]+1) / rasterGrid,
				})
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						nx, ny := c[0]+dx, c[1]+dy
						if nx < 0 || ny < 0 || nx >= rasterGrid || ny >= rasterGrid {
							continue
						}
						if changed[ny][nx] && !seen[ny][nx] {
							seen[ny][nx] = true
							stack = append(stack, [2]int{nx, ny})
						}
					}
				}
			}
			// a thin strip along the frame or a line is what a misaligned edge leaves, not a change of the drawing
			cw := int(math.Round((bl.box.X1 - bl.box.X0) * rasterGrid))
			ch := int(math.Round((bl.box.Y1 - bl.box.Y0) * rasterGrid))
			if bl.cells >= rasterMinCells && cw >= minSide && ch >= minSide {
				blobs = append(blobs, bl)
			}
		}
	}
	sort.Slice(blobs, func(i, j int) bool { return blobs[i].cells > blobs[j].cells })
	if len(blobs) > rasterMaxBoxes {
		blobs = blobs[:rasterMaxBoxes]
	}
	out := make([]Box, len(blobs))
	for i, b := range blobs {
		out[i] = b.box
	}
	return out
}

// Alignment of two renderings of a sheet (ТЗ 4.2.5): the same drawing printed or exported twice is
// rarely in the same place — a shifted frame, another scale, another paper — and a plain grid
// comparison then reports the whole sheet. The content of each page (everything that is not
// background) is taken as its area, and the two areas are compared cell for cell: a shift and a
// uniform change of scale cancel out. Added or removed content inside the frame shows as a difference.

const (
	contentInk      = 215.0 // a pixel darker than this is content
	contentMaxSkew  = 0.15  // the content areas may differ in proportion by this much to be aligned
	contentMinShare = 0.04  // the content must span at least this part of the page in both directions
	// a region of an aligned comparison is at least this many cells wide and high: thinner is the edge of
	// a frame or a line that two renderings never place alike
	alignedMinSide = 2
)

// ContentBox is the bounding box (page fractions) of what is drawn on a page; false for a blank page.
func ContentBox(img image.Image) (Box, bool) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < rasterGrid || h < rasterGrid {
		return Box{}, false
	}
	step := 1
	if m := math.Min(float64(w), float64(h)); m > 900 {
		step = int(m / 900)
	}
	x0, y0, x1, y1 := w, h, -1, -1
	for y := 0; y < h; y += step {
		for x := 0; x < w; x += step {
			if gray(img.At(b.Min.X+x, b.Min.Y+y)) < contentInk {
				if x < x0 {
					x0 = x
				}
				if x > x1 {
					x1 = x
				}
				if y < y0 {
					y0 = y
				}
				if y > y1 {
					y1 = y
				}
			}
		}
	}
	if x1 < 0 {
		return Box{}, false
	}
	box := Box{float64(x0) / float64(w), float64(y0) / float64(h), float64(x1+step) / float64(w), float64(y1+step) / float64(h)}
	if box.X1-box.X0 < contentMinShare || box.Y1-box.Y0 < contentMinShare {
		return Box{}, false
	}
	return box, true
}

// AlignedDiff compares two renderings of a sheet after aligning their content. It returns the differing
// regions on each page (the same regions, each in the fractions of its own page) and whether the
// pages could be aligned; when not (a blank page, contents of very different proportions) it falls
// back to the comparison of the whole pages.
func AlignedDiff(a, b image.Image) (onA, onB []Box, aligned bool) {
	ca, okA := ContentBox(a)
	cb, okB := ContentBox(b)
	wa, ha := ca.X1-ca.X0, ca.Y1-ca.Y0
	wb, hb := cb.X1-cb.X0, cb.Y1-cb.Y0
	// the proportions are those of the pictures, not of the fractions: a page is not a square
	ra := (wa * float64(a.Bounds().Dx())) / (ha * float64(a.Bounds().Dy()))
	rb := (wb * float64(b.Bounds().Dx())) / (hb * float64(b.Bounds().Dy()))
	if !okA || !okB || math.Abs(ra-rb)/ra > contentMaxSkew {
		boxes := DiffImages(a, b)
		return boxes, boxes, false
	}
	for _, f := range blobsOf(diffCells(cellMeansIn(a, ca), cellMeansIn(b, cb), 1), alignedMinSide) {
		onA = append(onA, Box{ca.X0 + f.X0*wa, ca.Y0 + f.Y0*ha, ca.X0 + f.X1*wa, ca.Y0 + f.Y1*ha})
		onB = append(onB, Box{cb.X0 + f.X0*wb, cb.Y0 + f.Y0*hb, cb.X0 + f.X1*wb, cb.Y0 + f.Y1*hb})
	}
	return onA, onB, true
}

// SheetSig is what is kept of a rendering of a sheet to tell how alike it is to another: the box of its
// content, the proportion of that box and the grid of brightness over it. The picture itself can be let go.
type SheetSig struct {
	box    Box
	aspect float64
	grid   [][]float64
	ok     bool
}

// SignatureOf reads the signature of a rendering.
func SignatureOf(img image.Image) SheetSig {
	c, ok := ContentBox(img)
	if !ok {
		return SheetSig{}
	}
	return SheetSig{
		box: c, ok: true, grid: cellMeansIn(img, c),
		aspect: ((c.X1 - c.X0) * float64(img.Bounds().Dx())) / ((c.Y1 - c.Y0) * float64(img.Bounds().Dy())),
	}
}

// Change is the share of the sheet (0..1) that differs between two signatures once the contents are aligned;
// 1 when they cannot be aligned.
func (a SheetSig) Change(b SheetSig) float64 {
	if !a.ok || !b.ok || math.Abs(a.aspect-b.aspect)/a.aspect > contentMaxSkew {
		return 1
	}
	n := 0
	for _, row := range diffCells(a.grid, b.grid, 1) {
		for _, c := range row {
			if c {
				n++
			}
		}
	}
	return float64(n) / float64(rasterGrid*rasterGrid)
}

// AlignedChange is the share of the sheet (0..1) that differs between two renderings after their contents
// are aligned. 1 when they cannot be aligned. It tells how alike two sheets are, to find the pair of a sheet.
func AlignedChange(a, b image.Image) float64 { return SignatureOf(a).Change(SignatureOf(b)) }
