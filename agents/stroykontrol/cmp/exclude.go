package cmp

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Исключаемые области (ТЗ 3.3.9): специалист указывает на листе область, которую не нужно сравнивать —
// штамп согласования, облако изменений, чужую вставку. Строки текста внутри области в сравнение не
// попадают, а расхождения чертежей, чей центр лежит в области, не фиксируются. Основная надпись
// и так не сравнивается. Исходные файлы не меняются.

// Region is an area of a page (in page fractions) that is left out of the comparison.
type Region struct {
	// Side: ref, target or both (empty = both)
	Side string `json:"side,omitempty"`
	// Doc: the document the region is on ("" = every document of the side)
	Doc string `json:"doc,omitempty"`
	// Page: the page (1-based; 0 = every page of the document)
	Page int `json:"page,omitempty"`
	// Box: "[x0,y0,x1,y1]" (or without the brackets) in page fractions (0..1)
	Box   string `json:"box"`
	Label string `json:"label,omitempty"`
}

// regionBox reads the box of a region, with or without the brackets of the stored form.
func regionBox(s string) *Box {
	s = strings.TrimSpace(s)
	if s != "" && !strings.HasPrefix(s, "[") {
		s = "[" + s + "]"
	}
	return ParseBox(s)
}

// ParseRegions reads the stored JSON ("" is no regions) and checks every region.
func ParseRegions(s string) ([]Region, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	var out []Region
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil, fmt.Errorf("исключаемые области: %w", err)
	}
	if err := ValidateRegions(out); err != nil {
		return nil, err
	}
	return out, nil
}

// ValidateRegions checks the regions a client sent.
func ValidateRegions(rs []Region) error {
	for i, r := range rs {
		switch r.Side {
		case "", SideRef, SideTarget, "both":
		default:
			return fmt.Errorf("область %d: сторона «%s» неизвестна (ref | target | both)", i+1, r.Side)
		}
		if r.Page < 0 {
			return fmt.Errorf("область %d: страница %d", i+1, r.Page)
		}
		b := regionBox(r.Box)
		if b == nil || b.X0 < 0 || b.Y0 < 0 || b.X1 > 1.0001 || b.Y1 > 1.0001 || b.X1 <= b.X0 || b.Y1 <= b.Y0 {
			return fmt.Errorf("область %d: нужен прямоугольник «x0,y0,x1,y1» в долях страницы (0…1), получено «%s»", i+1, r.Box)
		}
	}
	return nil
}

// RegionsJSON is the stored form.
func RegionsJSON(rs []Region) string {
	if len(rs) == 0 {
		return ""
	}
	b, _ := json.Marshal(rs)
	return string(b)
}

func (r Region) appliesTo(side, doc string, page int) bool {
	if r.Side != "" && r.Side != "both" && r.Side != side {
		return false
	}
	if r.Doc != "" && r.Doc != doc {
		return false
	}
	return r.Page == 0 || r.Page == page
}

func (b Box) holds(cx, cy float64) bool { return cx >= b.X0 && cx <= b.X1 && cy >= b.Y0 && cy <= b.Y1 }

// excluded reports whether the center of a box lies in a region that applies to the page.
func excluded(rs []Region, side, doc string, page int, b Box) bool {
	cx, cy := (b.X0+b.X1)/2, (b.Y0+b.Y1)/2
	for _, r := range rs {
		if !r.appliesTo(side, doc, page) {
			continue
		}
		if rb := regionBox(r.Box); rb != nil && rb.holds(cx, cy) {
			return true
		}
	}
	return false
}

// withoutRegions returns the documents with the lines and words inside the regions taken out. The
// documents given are not changed (their pages are shared with the cache of the extraction); only
// the pages a region touches are copied. Title-block lines stay: they are not compared anyway and
// the volume detection reads them.
func withoutRegions(docs []*Document, side string, rs []Region) ([]*Document, int) {
	if len(rs) == 0 {
		return docs, 0
	}
	out := make([]*Document, len(docs))
	dropped := 0
	for i, d := range docs {
		nd := *d
		nd.Pages = append([]Page(nil), d.Pages...)
		for pi := range nd.Pages {
			pg := &nd.Pages[pi]
			var lines []Line
			cut := false
			for _, l := range pg.Lines {
				if !l.Stamp && excluded(rs, side, d.ID, pi+1, l.Box) {
					cut = true
					dropped++
					continue
				}
				lines = append(lines, l)
			}
			if !cut {
				continue
			}
			pg.Lines = lines
			var words []Word
			for _, w := range pg.Words {
				if !excluded(rs, side, d.ID, pi+1, w.B) {
					words = append(words, w)
				}
			}
			pg.Words = words
		}
		out[i] = &nd
	}
	return out, dropped
}
