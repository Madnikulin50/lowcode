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
	// Expertise lists the remarks of the comparison's object that concern
	// this comparison (linked to its documents or matched to its findings).
	Expertise(ns, comparisonID string) ([]remarkRecord, error)
	// Board returns the object's registry of expertise remarks.
	Board(ns, objectID string) (*boardOut, error)
	// RemarkHistory returns the iterations/changes of one remark, oldest first.
	RemarkHistory(ns, objectID, remarkID string) ([]remarkHistoryRecord, error)
	// UpdateRemark applies the specialist's changes (status, assignee, due date, note).
	UpdateRemark(ns, objectID, remarkID string, p remarkPatch) (*remarkRecord, error)
	// RegisterIssued puts the list made from the comparison into the registry of
	// the object's remarks (creating, updating, closing).
	RegisterIssued(ns, comparisonID string, in cmp.IssueInput, res cmp.IssueResult) (*cmp.IssueRegistered, error)
	// IssuedNumbers maps the group keys of the comparison's registered remarks to their numbers.
	IssuedNumbers(ns, comparisonID string) (map[string]string, error)
	// ObjectName is the name of the object ("" if unknown).
	ObjectName(ns, objectID string) string
	// Pairing returns the documents of the comparison, its manual pairing and the pairs the last run made.
	Pairing(ns, recordID string) (*pairingOut, error)
	// SetPairing stores the manual pairing (the next run uses it).
	SetPairing(ns, recordID string, p []cmp.ManualPair) error
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
	ObjectID    string // the object (project) the comparison belongs to
	SetsLabel   string // which versions of the document sets are compared: "ПД v2 ↔ РД v1"
	Kind        string // cmp.ModeStages | cmp.ModeVersions
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

// remarkRecord is a remark of the expertise as the registry holds it.
type remarkRecord struct {
	ID               string          `json:"id"`
	Key              string          `json:"key"`
	Number           string          `json:"number"`
	Batch            string          `json:"batch"`
	Text             string          `json:"text"`
	Refs             string          `json:"refs"`
	VolumeRefs       []cmp.VolumeRef `json:"volumeRefs"`
	Section          string          `json:"section"`
	Basis            string          `json:"basis"`
	Answer           string          `json:"answer"`
	RejectReason     string          `json:"rejectReason"`
	PrevAnswer       string          `json:"prevAnswer"`
	PrevRejectReason string          `json:"prevRejectReason"`
	Discipline       string          `json:"discipline"`
	Expert           string          `json:"expert"`
	Criticality      string          `json:"criticality"`
	Proposals        string          `json:"proposals"`
	Expected         string          `json:"expected"`
	Iterations       int             `json:"iterations"`
	ChangedAt        string          `json:"changedAt"`
	SourceStatus     string          `json:"sourceStatus"`
	Status           string          `json:"status"` // the working status
	Assignee         string          `json:"assignee"`
	DueDate          string          `json:"dueDate"`
	Note             string          `json:"note"`
	Verdict          string          `json:"verdict"`
	Score            string          `json:"score"`
	MatchNote        string          `json:"matchNote"`
	ClaimMismatch    bool            `json:"claimMismatch"`
	Findings         []string        `json:"findings"`  // matched discrepancies
	Documents        []string        `json:"documents"` // linked cmp_documents
	Archived         bool            `json:"archived"`
	Origin           string          `json:"origin"` // "comparison": a remark made from a ПД/РД comparison
	// AI reading of the remark by the fragments of its volumes (never changes the working status)
	AIVerdict string `json:"aiVerdict"`
	AINote    string `json:"aiNote"`
	AIAt      string `json:"aiAt"`
}

type remarkHistoryRecord struct {
	Iteration    int    `json:"iteration"`
	Answer       string `json:"answer"`
	RejectReason string `json:"rejectReason"`
	Status       string `json:"status"`
	SourceStatus string `json:"sourceStatus"`
	At           string `json:"at"`
	Author       string `json:"author"`
	Source       string `json:"source"`
}

type importRecord struct {
	ID            string `json:"id"`
	File          string `json:"file"`
	At            string `json:"at"`
	Total         int    `json:"total"`
	New           int    `json:"new"`
	Changed       int    `json:"changed"`
	StatusChanges int    `json:"statusChanges"`
	Reopened      int    `json:"reopened"`
	Archived      int    `json:"archived"`
	Details       string `json:"details"`
}

type boardComparison struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// boardOut is everything the remarks workbench shows for an object.
type boardOut struct {
	ObjectID    string            `json:"objectID"`
	ObjectName  string            `json:"objectName"`
	Remarks     []remarkRecord    `json:"remarks"`
	Imports     []importRecord    `json:"imports"`
	Comparisons []boardComparison `json:"comparisons"`
	Documents   map[string]string `json:"documents"` // linked document ID → name
}

// remarkPatch is what the specialist may change by hand; nil = leave.
type remarkPatch struct {
	Status   *string `json:"status"`
	Assignee *string `json:"assignee"`
	DueDate  *string `json:"dueDate"`
	Note     *string `json:"note"`
}
