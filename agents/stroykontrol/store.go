package main

import "github.com/madnikulin50/lowcode/agents/stroykontrol/cmp"

// store abstracts where a comparison record and its files come from.
// ComposeClient implements it against a live Compose instance (the normal
// case); FixtureStore implements it against a local directory, so
// stroykontrol-web can run — and be demoed — with no Compose server and no
// API token at all. See --fixtures in main.go and fixtures/README.md.
type store interface {
	// Comparison returns the pd_rd_comparisons fields for recordID.
	Comparison(ns, recordID string) (*comparisonRecord, error)
	// Discrepancies returns pd_rd_discrepancies rows for that comparison.
	Discrepancies(ns, recordID string) ([]discrepancyRecord, error)
	// Documents lists the files of a comparison: its package (reference
	// side = ПД, target side = РД/Р), or — for older comparisons — the
	// single pd_file/rd_file.
	Documents(ns, recordID string) ([]documentRecord, error)
	// FetchDocument returns the raw bytes + content-type of one document of
	// the comparison. docID comes from Documents(); implementations must
	// refuse documents that do not belong to recordID.
	FetchDocument(ns, recordID, docID string) (data []byte, mimetype string, err error)
	// SetDiscrepancyStatus records the reviewer's decision (new|confirmed|rejected).
	SetDiscrepancyStatus(ns, recordID, discrepancyID, status string) error
	// Remarks lists the remark cards of a comparison.
	Remarks(ns, recordID string) ([]cmp.Remark, error)
	// Expertise lists the imported expertise remarks of a comparison with
	// their matches to the findings.
	Expertise(ns, recordID string) ([]expertiseRecord, error)
	// SetExpertiseStatus records the specialist's own status of a remark (open|closed|disputed).
	SetExpertiseStatus(ns, recordID, remarkID, status string) error
	// SaveRemark creates (ID == "") or updates a remark card.
	SaveRemark(ns, recordID string, r cmp.Remark) (cmp.Remark, error)
}

type comparisonRecord struct {
	Title             string
	Status            string
	SimilarityPercent string
	Comment           string
	// iteration: this comparison re-checks the same pair after corrections
	ParentID    string
	ParentTitle string
	IterationNo string
}

type documentRecord struct {
	ID          string `json:"id"`
	Side        string `json:"side"` // ref | target
	Name        string `json:"name"`
	VolumeCode  string `json:"volumeCode"`
	VolumeTitle string `json:"volumeTitle"`
	Pages       string `json:"pages"`
}

// discrepancyRecord is the wire/fixture form of a discrepancy. The first
// four fields are the original schema (still accepted in fixtures); the
// rest carry the position of the finding on the documents.
type discrepancyRecord struct {
	ID          string `json:"id,omitempty"`
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	PageNumber  string `json:"pageNumber"`
	Kind        string `json:"kind,omitempty"`
	Status      string `json:"status,omitempty"`
	DocRef      string `json:"docRef,omitempty"`
	DocTarget   string `json:"docTarget,omitempty"`
	PageRef     string `json:"pageRef,omitempty"`
	PageTarget  string `json:"pageTarget,omitempty"`
	BboxRef     string `json:"bboxRef,omitempty"`
	BboxTarget  string `json:"bboxTarget,omitempty"`
	Anchor      string `json:"anchor,omitempty"`
	Evidence    string `json:"evidence,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	// IterState: new | persisting | resolved, relative to the previous iteration
	IterState string `json:"iterState,omitempty"`
}

// expertiseRecord is an expertise remark (замечание экспертизы) with the
// designer's answer and what the engine found for it in the documents.
type expertiseRecord struct {
	ID            string   `json:"id"`
	Number        string   `json:"number"`
	Text          string   `json:"text"`
	Answer        string   `json:"answer"`
	Location      string   `json:"location"`
	Source        string   `json:"source"`
	Verdict       string   `json:"verdict"` // still_differs | resolved | no_match
	Score         string   `json:"score"`
	Note          string   `json:"note"`
	Status        string   `json:"status"` // open | closed | disputed
	Findings      []string `json:"findings"`
	ClaimMismatch bool     `json:"claimMismatch"`
}
