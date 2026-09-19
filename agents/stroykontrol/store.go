package main

// store abstracts where a comparison record and its pd/rd files come from.
// ComposeClient implements it against a live Compose instance (the normal
// case); FixtureStore implements it against a local directory, so
// stroykontrol-web can run — and be demoed — with no Compose server and no
// API token at all. See --fixtures in main.go and fixtures/README.md.
type store interface {
	// Comparison returns the pd_rd_comparisons fields for recordID.
	Comparison(ns, recordID string) (*comparisonRecord, error)
	// Discrepancies returns pd_rd_discrepancies rows for that comparison.
	Discrepancies(ns, recordID string) ([]discrepancyRecord, error)
	// AttachmentName is a best-effort label lookup: callers fall back to a
	// plain "ПД"/"РД" label if this returns "".
	AttachmentName(ns, attachmentID string) string
	// FetchFile returns the raw bytes + content-type of a comparison's pd or
	// rd file (side is "pd" or "rd").
	FetchFile(ns, recordID, side string) (data []byte, mimetype string, err error)
}

type comparisonRecord struct {
	Title             string
	Status            string
	SimilarityPercent string
	Comment           string
	PdFileID          string // opaque attachment handle; "" means no file
	RdFileID          string
}

type discrepancyRecord struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	PageNumber  string `json:"pageNumber"`
}
