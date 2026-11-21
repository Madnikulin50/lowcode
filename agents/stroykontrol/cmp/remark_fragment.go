package cmp

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
)

// Фрагмент, названный в замечании (ТЗ 3.4.2, 3.5.1): замечание называет место текстом — цитатой в
// кавычках, значением, маркой, — а не номером страницы. Агент ищет это место в привязанных томах по
// текстовому слою: сначала цитаты из замечания, затем окно в несколько строк, у которого больше всего
// общих слов и (сильнее) значений и марок с замечанием. Лучшее место на каждой стороне (исходная и
// исправленная редакции) становится местом замечания с рамкой на странице. Страница, на которую
// ссылается само замечание, получает преимущество. Чертежи с «обведённым» текстом и сканы без текста
// не участвуют (п. 2.4.7).

// LocFragment is the kind of a location found by the text of the remark.
const LocFragment = "fragment"

// FragDoc is a linked document with its pages.
type FragDoc struct {
	LocDoc
	Pages []Page
}

const (
	fragMinScore   = 0.3 // a window that shares only words with the remark
	fragValueScore = 0.2 // a window that shares a value or a mark too
	fragWindow     = 3   // lines of a text page in one window
)

var quoteRe = regexp.MustCompile(`[«"„“]([^«»"„“”]{6,200})[»"“”]`)

// remarkQuotes are the phrases the remark quotes (two words at least).
func remarkQuotes(text string) []string {
	var out []string
	for _, m := range quoteRe.FindAllStringSubmatch(text, -1) {
		q := normQuote(m[1])
		if len(strings.Fields(q)) >= 2 {
			out = append(out, q)
		}
	}
	return out
}

func normQuote(s string) string {
	s = strings.ToLower(strings.ReplaceAll(s, "ё", "е"))
	return strings.Join(strings.Fields(s), " ")
}

type fragHit struct {
	doc   *FragDoc
	page  int
	box   Box
	text  string
	score float64
	quote bool
	words []string
}

// LocateFragments finds the places the remark's text names in the linked documents: the best one
// on each side. refs are the remark's references (their pages are preferred).
func LocateFragments(text string, refs []VolumeRef, docs []FragDoc) []RemarkLocation {
	qt := remarkTokens(text)
	quotes := remarkQuotes(text)
	wq := total(qt)
	if wq == 0 && len(quotes) == 0 {
		return nil
	}
	best := map[string]*fragHit{} // side → the best place
	for di := range docs {
		d := &docs[di]
		named := namedPages(refs, d)
		for _, pg := range d.Pages {
			if !pg.HasText || pg.Cover {
				continue
			}
			var lines []Line
			for _, l := range pg.Lines {
				if !l.Stamp && !skipLine(l.Key) {
					lines = append(lines, l)
				}
			}
			win := fragWindow
			if pg.Drawing {
				win = 1 // scattered labels: one label is the place
			}
			bonus := 1.0
			if named[pg.Number] {
				bonus = 1.5
			}
			for i := range lines {
				for j := i + 1; j <= len(lines) && j <= i+win; j++ {
					h := scoreWindow(lines[i:j], qt, wq, quotes)
					if h == nil {
						continue
					}
					h.score *= bonus
					h.doc, h.page = d, pg.Number
					if cur := best[d.Side]; cur == nil || better(h, cur) {
						best[d.Side] = h
					}
				}
			}
		}
	}
	sides := make([]string, 0, len(best))
	for s := range best {
		sides = append(sides, s)
	}
	sort.Slice(sides, func(i, j int) bool { return sideOrder(sides[i]) < sideOrder(sides[j]) })
	var out []RemarkLocation
	for _, s := range sides {
		h := best[s]
		box := h.box
		loc := RemarkLocation{Kind: LocFragment, DocID: h.doc.ID, Doc: h.doc.Name, Side: h.doc.Side, Comparison: h.doc.Comparison,
			Page: h.page, Sheet: h.doc.Sheets[h.page], Box: box.String(), Text: trunc(h.text, 200)}
		if h.quote {
			loc.Note = "по цитате из замечания"
		} else {
			loc.Note = fmt.Sprintf("по совпадению: %s", strings.Join(h.words, ", "))
		}
		out = append(out, loc)
	}
	return out
}

func sideOrder(s string) int {
	switch s {
	case SideRef:
		return 0
	case SideTarget:
		return 1
	}
	return 2
}

// better prefers a quotation, then the score, then the shorter window.
func better(a, b *fragHit) bool {
	if a.quote != b.quote {
		return a.quote
	}
	if math.Abs(a.score-b.score) > 1e-9 {
		return a.score > b.score
	}
	return len(a.text) < len(b.text)
}

func scoreWindow(lines []Line, qt weighted, wq float64, quotes []string) *fragHit {
	var parts []string
	box := lines[0].Box
	for i, l := range lines {
		parts = append(parts, l.Text)
		if i > 0 {
			box = box.union(l.Box)
		}
	}
	text := strings.Join(parts, " ")
	h := &fragHit{box: box, text: text}
	norm := normQuote(text)
	for _, q := range quotes {
		if strings.Contains(norm, q) {
			h.quote, h.score = true, 1 // of the windows that hold it, the shortest wins (better)
			return h
		}
	}
	wt := remarkTokens(text)
	tw := total(wt)
	if tw == 0 || wq == 0 {
		return nil
	}
	s, common := shared(qt, wt)
	if len(common) < 2 {
		return nil
	}
	value := false
	for _, c := range common {
		if qt[c] > 1 {
			value = true
		}
	}
	h.score = s / math.Sqrt(wq*tw)
	if h.score < fragMinScore && !(value && h.score >= fragValueScore) {
		return nil
	}
	h.words = common
	return h
}

// namedPages are the pages of the document the references of the remark name.
func namedPages(refs []VolumeRef, d *FragDoc) map[int]bool {
	out := map[int]bool{}
	for _, l := range LocateRemark(refs, []LocDoc{d.LocDoc}, nil) {
		if l.Page > 0 {
			out[l.Page] = true
		}
	}
	return out
}

// fragmentDocs are the documents searched for the fragment of a remark: the volumes it is linked to
// and the documents of the findings it was matched to.
func fragmentDocs(linked, findings []string, docOf map[string][2]string) []string {
	seen := map[string]bool{}
	var out []string
	add := func(id string) {
		if id != "" && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	for _, id := range linked {
		add(id)
	}
	for _, f := range findings {
		add(docOf[f][0])
		add(docOf[f][1])
	}
	return out
}
