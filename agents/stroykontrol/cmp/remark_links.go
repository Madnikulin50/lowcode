package cmp

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// Привязка замечания экспертизы к месту в документах (ТЗ 3.5): тома, листы и страницы, на которые
// ссылается замечание, и области находок сравнения, с которыми оно сопоставлено. Лист — номер из
// основной надписи; страница — номер страницы PDF; связь между ними — карта листов документа.

// SheetMap reads the sheet numbers of the title blocks: PDF page → sheet number.
func SheetMap(pages []Page) map[int]int {
	out := map[int]int{}
	for i, p := range pages {
		if p.Cover {
			continue
		}
		if no, ok := sheetNumber(p); ok {
			out[i+1] = no
		}
	}
	return out
}

// FormatSheetMap writes a sheet map as "страница:лист,…" (the form it has in a document record).
func FormatSheetMap(m map[int]int) string {
	pages := make([]int, 0, len(m))
	for p := range m {
		pages = append(pages, p)
	}
	sort.Ints(pages)
	var b strings.Builder
	for i, p := range pages {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, "%d:%d", p, m[p])
	}
	return b.String()
}

// ParseSheetMap is the reverse of FormatSheetMap.
func ParseSheetMap(s string) map[int]int {
	out := map[int]int{}
	for _, part := range strings.Split(s, ",") {
		p, n, ok := strings.Cut(strings.TrimSpace(part), ":")
		if !ok {
			continue
		}
		pg, e1 := strconv.Atoi(p)
		sh, e2 := strconv.Atoi(n)
		if e1 == nil && e2 == nil && pg > 0 {
			out[pg] = sh
		}
	}
	return out
}

// Kinds of a location of a remark.
const (
	LocPlace   = "place"   // the place the remark's own «Ссылка на материалы» names
	LocFinding = "finding" // the place of a comparison finding the remark was matched to
)

// RemarkLocation is where a remark leads in the documents.
type RemarkLocation struct {
	Kind string `json:"kind"`
	// place: the document, its stage and the place the reference names
	DocID string `json:"docID,omitempty"`
	Doc   string `json:"doc,omitempty"`  // file name
	Side  string `json:"side,omitempty"` // ref = исходная редакция (ПД), target = исправленная (РД)
	Page  int    `json:"page,omitempty"` // PDF page (0: unknown, open from the start)
	Sheet int    `json:"sheet,omitempty"`
	Note  string `json:"note,omitempty"` // why the place is approximate; for a fragment, how it was found
	Box   string `json:"box,omitempty"`  // fragment: the region on the page, "[x0,y0,x1,y1]" in page fractions
	// finding: the comparison, the finding and its places on both sides
	Comparison string `json:"comparison,omitempty"`
	FindingID  string `json:"findingID,omitempty"`
	Text       string `json:"text,omitempty"`
	RefDoc     string `json:"refDoc,omitempty"`
	RefPage    int    `json:"refPage,omitempty"`
	RefBox     string `json:"refBox,omitempty"`
	TargetDoc  string `json:"targetDoc,omitempty"`
	TargetPage int    `json:"targetPage,omitempty"`
	TargetBox  string `json:"targetBox,omitempty"`
}

// LocDoc is a loaded document a reference may lead into.
type LocDoc struct {
	ID, Name, Side, Codes, Title string
	Comparison                   string // the comparison the file belongs to ("" — a file of a set only)
	Pages                        int
	Sheets                       map[int]int // PDF page → sheet number
}

// FoundAt is a finding a remark was matched to.
type FoundAt struct {
	ID, Comparison, Text string
	RefDoc, TargetDoc    string
	RefPage, TargetPage  int
	RefBox, TargetBox    string
}

func pageOfSheet(m map[int]int, sheet int) int {
	best := 0
	for pg, sh := range m {
		if sh == sheet && (best == 0 || pg < best) {
			best = pg
		}
	}
	return best
}

// LocateRemark turns the references of a remark and its matched findings into places in the
// documents. A reference with a page goes to that page; one with a sheet goes to the page whose
// title block carries that number (when the document's sheets were read — otherwise the place is
// the start of the volume, said so in Note); one with neither opens the volume at its start.
func LocateRemark(refs []VolumeRef, docs []LocDoc, findings []FoundAt) []RemarkLocation {
	var out []RemarkLocation
	seen := map[string]bool{}
	for _, ref := range refs {
		for _, d := range docs {
			codes := docCodes(d.Codes)
			hit := len(codes) == 0 && MatchVolume(ref, "", d.Title)
			for _, c := range codes {
				if MatchVolume(ref, c, d.Title) {
					hit = true
				}
			}
			if !hit {
				continue
			}
			loc := RemarkLocation{Kind: LocPlace, DocID: d.ID, Doc: d.Name, Side: d.Side, Sheet: ref.Sheet, Comparison: d.Comparison}
			switch {
			case ref.Page > 0:
				loc.Page = ref.Page
				if d.Pages > 0 && ref.Page > d.Pages {
					loc.Page, loc.Note = 0, fmt.Sprintf("в файле %d стр., а замечание называет стр. %d", d.Pages, ref.Page)
				}
			case ref.Sheet > 0:
				if pg := pageOfSheet(d.Sheets, ref.Sheet); pg > 0 {
					loc.Page = pg
				} else if len(d.Sheets) == 0 {
					loc.Note = fmt.Sprintf("лист %d: номера листов файла не прочитаны, открыт с начала", ref.Sheet)
				} else {
					loc.Note = fmt.Sprintf("лист %d: в основных надписях файла такого номера нет", ref.Sheet)
				}
			}
			key := fmt.Sprintf("%s|%d|%d", d.ID, loc.Page, loc.Sheet)
			if !seen[key] {
				seen[key] = true
				out = append(out, loc)
			}
		}
	}
	for _, f := range findings {
		out = append(out, RemarkLocation{
			Kind: LocFinding, Comparison: f.Comparison, FindingID: f.ID, Text: trunc(f.Text, 160),
			RefDoc: f.RefDoc, RefPage: f.RefPage, RefBox: f.RefBox,
			TargetDoc: f.TargetDoc, TargetPage: f.TargetPage, TargetBox: f.TargetBox,
		})
	}
	return out
}

// LocationsJSON is the value of the field links_json.
func LocationsJSON(l []RemarkLocation) string {
	if len(l) == 0 {
		return ""
	}
	b, _ := json.Marshal(l)
	return string(b)
}

// locDocsFrom reads the loaded documents of the comparisons of an object.
func locDocsFrom(recs []idcheck.Record, side func(docID string) string) []LocDoc {
	var out []LocDoc
	for _, d := range recs {
		pages, _ := strconv.Atoi(d.Get("pages"))
		out = append(out, LocDoc{
			ID: d.ID, Name: strings.TrimSpace(d.Get("volume_code") + " " + d.Get("volume_title")), Side: firstNonEmpty(d.Get("side"), side(d.ID)), Codes: d.Get("volume_code"), Title: d.Get("volume_title"),
			Pages: pages, Sheets: ParseSheetMap(d.Get("sheet_map")), Comparison: d.Get("comparison"),
		})
	}
	return out
}

func atoi(s string) int { n, _ := strconv.Atoi(s); return n }
