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
func cellMeans(img image.Image) [][]float64 {
	b := img.Bounds()
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
func DiffImages(a, b image.Image) []Box {
	ma, mb := cellMeans(a), cellMeans(b)
	changed := make([][]bool, rasterGrid)
	for y := range changed {
		changed[y] = make([]bool, rasterGrid)
		for x := range changed[y] {
			changed[y][x] = math.Abs(ma[y][x]-mb[y][x]) > rasterCellDelta
		}
	}
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
			if bl.cells >= rasterMinCells {
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
