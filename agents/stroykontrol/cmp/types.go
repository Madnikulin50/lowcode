// Package cmp compares two sets of PDF documents — the reference stage (П,
// эталон) and the target stage (Р/РД) — page by page and produces findings
// with positions. It is deterministic: no LLM takes part in deciding whether
// something differs. The runner (run.go) persists results into Compose.
package cmp

import (
	"encoding/json"
	"fmt"
)

// Sides of a comparison.
const (
	SideRef    = "ref"    // стадия П — эталон
	SideTarget = "target" // стадия Р/РД
)

// Finding kinds.
const (
	KindText       = "text"
	KindTableRow   = "table_row"
	KindParam      = "param"
	KindDrawing    = "drawing"
	KindVOR        = "vor" // bill of quantities / specification consistency
	KindNotRelease = "not_released"
	KindScopeDiff  = "scope_diff"
)

// Box is a rectangle in page fractions (0..1, origin top-left).
type Box struct {
	X0, Y0, X1, Y1 float64
}

// String is the form stored in Compose: "[x0,y0,x1,y1]".
func (b *Box) String() string {
	if b == nil {
		return ""
	}
	return fmt.Sprintf("[%.4f,%.4f,%.4f,%.4f]", b.X0, b.Y0, b.X1, b.Y1)
}

// ParseBox parses the stored form; nil when empty or malformed.
func ParseBox(s string) *Box {
	if s == "" {
		return nil
	}
	var a []float64
	if err := json.Unmarshal([]byte(s), &a); err != nil || len(a) != 4 {
		return nil
	}
	return &Box{a[0], a[1], a[2], a[3]}
}

func (b Box) union(o Box) Box {
	if o.X0 < b.X0 {
		b.X0 = o.X0
	}
	if o.Y0 < b.Y0 {
		b.Y0 = o.Y0
	}
	if o.X1 > b.X1 {
		b.X1 = o.X1
	}
	if o.Y1 > b.Y1 {
		b.Y1 = o.Y1
	}
	return b
}

// Line is one visual row of text on a page.
type Line struct {
	DocID string
	Page  int // 1-based within the document
	Text  string
	Key   string // normalized comparison key
	Box   Box
	Stamp bool // inside the title block (основная надпись): not compared
	OCR   bool // recognized from an image, not read from a text layer
}

// Page is one PDF page.
type Page struct {
	Number  int
	W, H    float64
	Lines   []Line
	HasText bool
	// Drawing: a sheet whose text layer is only scattered labels (or none) —
	// compared as an image, never line by line.
	Drawing bool
	// Cover: title/approval page at the start of a document; its dates,
	// signatures and edition marks differ by nature and are not compared.
	Cover bool
	// OCR: the text comes from recognition of a scan.
	OCR bool
	// Words with their boxes, kept for table pages (not for drawing sheets).
	Words []Word `json:",omitempty"`
}

// Word is one word of a page with its box in page fractions.
type Word struct {
	T string
	B Box
}

// Document is one input file of either side.
type Document struct {
	ID   string
	Name string
	Side string
	// ManualCode/ManualTitle, when set, make the whole file one volume with
	// that code (user override of the automatic detection).
	ManualCode  string
	ManualTitle string
	Pages       []Page
}

// Volume is a contiguous run of pages of one document that belongs to one
// том (by шифр in the title block).
type Volume struct {
	DocID     string
	DocName   string
	Side      string
	Index     int
	Code      string // normalized шифр
	RawCode   string
	Title     string
	FirstPage int
	LastPage  int
	Lines     []Line // compared lines (title block, cover and drawing sheets excluded)
	Drawings  []int  // page numbers of drawing sheets (compared as images)
	Items     []Item // positions of the ВОР and specification tables (compared as tables)
}

// Evidence is the explanation attached to a finding: which fragments the
// conclusion rests on.
type Evidence struct {
	Ref       string `json:"ref,omitempty"`
	Target    string `json:"target,omitempty"`
	VolumeRef string `json:"volumeRef,omitempty"`
	VolumeTgt string `json:"volumeTarget,omitempty"`
	Rule      string `json:"rule,omitempty"`
}

// Finding is one discrepancy.
type Finding struct {
	Kind        string
	Type        string // legacy discrepancy_type: content|numbering|missing_page|extra_page|drawing
	Severity    string // low|medium|high
	Description string
	RefDoc      string
	TargetDoc   string
	PageRef     int
	PageTarget  int
	BoxRef      *Box
	BoxTarget   *Box
	Anchor      string
	Evidence    Evidence
	Fingerprint string
	OCR         bool // rests on recognized (scanned) text, less reliable
}

// VolumeInfo describes a detected volume for persistence/display.
type VolumeInfo struct {
	DocID string
	Index int
	Code  string
	Title string
	From  int
	To    int
}

// PairInfo says which volumes were compared with which.
type PairInfo struct {
	Ref      VolumeInfo
	Targets  []VolumeInfo
	Fragment bool
	Score    float64
}

// Stats summarizes a run.
type Stats struct {
	RefPages       int
	TargetPages    int
	RefLines       int // lines of paired reference volumes
	MatchedLines   int
	ComparedPages  int // reference pages inside paired volumes
	DifferingPages int
	DrawingRef     int // drawing sheets in paired reference volumes
	DrawingTarget  int
	TableItems     int // ВОР/specification positions read in paired volumes (reference + target)
	// Similarity is a percentage; SimilarityKnown is false when there was no
	// text to compare (drawings only, or nothing paired).
	Similarity      float64
	SimilarityKnown bool
	simNum, simDen  float64
}

// Result is the full outcome of Compare.
type Result struct {
	Findings []Finding
	Volumes  map[string][]VolumeInfo // docID → volumes
	Pairs    []PairInfo
	Stats    Stats
	Notes    []string
}
