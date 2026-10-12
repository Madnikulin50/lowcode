package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/cmp"
)

var (
	cmpRunner   *cmp.Runner
	suggester   *cmp.Suggester
	cmpCacheDir string
	cmpOCR      *cmp.OCR
)

// rasterPage renders one page of a PDF through the shared page cache, for
// the image comparison of drawing pages.
func rasterPage(data []byte, page int) (image.Image, error) {
	dir, count, err := raster.Pages(data)
	if err != nil {
		return nil, err
	}
	path, err := raster.PagePath(dir, count, page)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return jpeg.Decode(f)
}

func itoaOrEmpty(n int) string {
	if n <= 0 {
		return ""
	}
	return strconv.Itoa(n)
}

// findingRecord is the wire form of an engine finding.
func findingRecord(f cmp.Finding, id string) discrepancyRecord {
	page := f.PageRef
	if page == 0 {
		page = f.PageTarget
	}
	ev, _ := json.Marshal(f.Evidence)
	return discrepancyRecord{
		ID: id, Type: f.Type, Severity: f.Severity, Description: "[детерм.] " + f.Description,
		PageNumber: itoaOrEmpty(page), Kind: f.Kind, Status: cmp.StatusNew,
		DocRef: f.RefDoc, DocTarget: f.TargetDoc,
		PageRef: itoaOrEmpty(f.PageRef), PageTarget: itoaOrEmpty(f.PageTarget),
		BboxRef: f.BoxRef.String(), BboxTarget: f.BoxTarget.String(),
		Anchor: f.Anchor, Evidence: string(ev), Fingerprint: f.Fingerprint,
	}
}

// resolveDoc picks the document a request is about: ?doc=<id>, or the
// legacy ?side=pd|rd of single-file comparisons.
func resolveDoc(r *http.Request) (ns, recordID, doc string, err error) {
	q := r.URL.Query()
	ns, recordID, doc = nsFromQuery(r), q.Get("recordID"), q.Get("doc")
	if ns == "" || recordID == "" {
		return "", "", "", fmt.Errorf("recordID and namespaceID are required")
	}
	if doc == "" {
		switch q.Get("side") {
		case "pd":
			doc = legacyPD
		case "rd":
			doc = legacyRD
		default:
			return "", "", "", fmt.Errorf("doc is required")
		}
	}
	return ns, recordID, doc, nil
}

func decodeJSON(r *http.Request, v any) error {
	if r.Body == nil {
		return fmt.Errorf("empty body")
	}
	return json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20)).Decode(v)
}

// handleCmpRun queues a package comparison. Called by the Compose rule chain
// "stroykontrol-compare-pd-rd" (http node) with ?recordID=&namespaceID=.
func handleCmpRun(w http.ResponseWriter, r *http.Request) {
	if cmpRunner == nil {
		writeJSON(w, http.StatusOK, map[string]any{"queued": false, "note": "comparison runs on demand in --fixtures mode"})
		return
	}
	ns, rec := nsFromQuery(r), strings.TrimSpace(r.URL.Query().Get("recordID"))
	if ns == "" || rec == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("recordID and namespaceID are required"))
		return
	}
	queued := cmpRunner.Enqueue(ns, rec)
	writeJSON(w, http.StatusAccepted, map[string]any{"recordID": rec, "queued": queued, "status": cmpRunner.Status(rec)})
}

func handleCmpStatus(w http.ResponseWriter, r *http.Request) {
	rec := r.URL.Query().Get("recordID")
	st := ""
	if cmpRunner != nil {
		st = cmpRunner.Status(rec)
	}
	writeJSON(w, http.StatusOK, map[string]any{"recordID": rec, "status": st})
}

var validStatuses = map[string]bool{cmp.StatusNew: true, cmp.StatusConfirmed: true, cmp.StatusRejected: true}

func handleDiscrepancyStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NamespaceID string `json:"namespaceID"`
		RecordID    string `json:"recordID"`
		ID          string `json:"id"`
		Status      string `json:"status"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	ns := firstNonEmpty(body.NamespaceID, defaultNamespaceID)
	if ns == "" || body.RecordID == "" || body.ID == "" || !validStatuses[body.Status] {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("recordID, id and status (new|confirmed|rejected) are required"))
		return
	}
	if err := compose.SetDiscrepancyStatus(ns, body.RecordID, body.ID, body.Status); err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": body.ID, "status": body.Status})
}

// handleRemarkCreate drafts a remark card from the selected discrepancies,
// deterministically (no LLM).
func handleRemarkCreate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NamespaceID    string   `json:"namespaceID"`
		RecordID       string   `json:"recordID"`
		DiscrepancyIDs []string `json:"discrepancyIDs"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	ns := firstNonEmpty(body.NamespaceID, defaultNamespaceID)
	if ns == "" || body.RecordID == "" || len(body.DiscrepancyIDs) == 0 {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("recordID and discrepancyIDs are required"))
		return
	}
	discs, err := compose.Discrepancies(ns, body.RecordID)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	docs, _ := compose.Documents(ns, body.RecordID)
	names := map[string]string{}
	for _, d := range docs {
		names[d.ID] = d.Name
	}
	want := map[string]bool{}
	for _, id := range body.DiscrepancyIDs {
		want[id] = true
	}
	var facts []cmp.Fact
	var ids []string
	for _, d := range discs {
		if !want[d.ID] {
			continue
		}
		ids = append(ids, d.ID)
		f := cmp.Fact{Description: d.Description, DocRef: names[d.DocRef], DocTarget: names[d.DocTarget], PageRef: d.PageRef, PageTarget: d.PageTarget, Kind: d.Kind}
		if f.DocRef == "" && d.PageRef != "" {
			f.DocRef = "ПД"
		}
		if f.DocTarget == "" && d.PageTarget != "" {
			f.DocTarget = "РД"
		}
		_ = json.Unmarshal([]byte(d.Evidence), &f.Evidence)
		facts = append(facts, f)
	}
	if len(facts) == 0 {
		writeErr(w, http.StatusNotFound, fmt.Errorf("no such discrepancies"))
		return
	}
	card := cmp.BuildRemark(facts)
	card.DiscrepancyIDs = ids
	saved, err := compose.SaveRemark(ns, body.RecordID, card)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

var validRemarkStatuses = map[string]bool{"draft": true, "issued": true, "closed": true}

// handleRemarkSave stores the specialist's edits of a card.
func handleRemarkSave(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NamespaceID string     `json:"namespaceID"`
		RecordID    string     `json:"recordID"`
		Remark      cmp.Remark `json:"remark"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	ns := firstNonEmpty(body.NamespaceID, defaultNamespaceID)
	if ns == "" || body.RecordID == "" || body.Remark.ID == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("recordID and remark.id are required"))
		return
	}
	if body.Remark.Status != "" && !validRemarkStatuses[body.Remark.Status] {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("status must be draft|issued|closed"))
		return
	}
	saved, err := compose.SaveRemark(ns, body.RecordID, body.Remark)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

// handleRemarkSuggest asks the local model for a draft wording of a card.
// Only runs on this explicit request; the draft lands in draftText and the
// specialist decides what goes into finalText.
func handleRemarkSuggest(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NamespaceID string `json:"namespaceID"`
		RecordID    string `json:"recordID"`
		ID          string `json:"id"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	ns := firstNonEmpty(body.NamespaceID, defaultNamespaceID)
	if ns == "" || body.RecordID == "" || body.ID == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("recordID and id are required"))
		return
	}
	if !suggester.Enabled() {
		writeErr(w, http.StatusServiceUnavailable, fmt.Errorf("модель для формулировок не настроена (--remark-model)"))
		return
	}
	cards, err := compose.Remarks(ns, body.RecordID)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	var card *cmp.Remark
	for i := range cards {
		if cards[i].ID == body.ID {
			card = &cards[i]
		}
	}
	if card == nil {
		writeErr(w, http.StatusNotFound, fmt.Errorf("remark %s not found", body.ID))
		return
	}
	text, err := suggester.Suggest(r.Context(), *card)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	card.DraftText = text
	saved, err := compose.SaveRemark(ns, body.RecordID, *card)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

// sortExpertise orders remarks by number, numerically when they are numbers.
func sortExpertise(l []expertiseRecord) {
	sort.SliceStable(l, func(i, j int) bool {
		a, ea := strconv.Atoi(l[i].Number)
		b, eb := strconv.Atoi(l[j].Number)
		if ea == nil && eb == nil {
			return a < b
		}
		return l[i].Number < l[j].Number
	})
}

var validExpertiseStatuses = map[string]bool{"open": true, "closed": true, "disputed": true}

func handleExpertiseStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NamespaceID string `json:"namespaceID"`
		RecordID    string `json:"recordID"`
		ID          string `json:"id"`
		Status      string `json:"status"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	ns := firstNonEmpty(body.NamespaceID, defaultNamespaceID)
	if ns == "" || body.RecordID == "" || body.ID == "" || !validExpertiseStatuses[body.Status] {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("recordID, id and status (open|closed|disputed) are required"))
		return
	}
	if err := compose.SetExpertiseStatus(ns, body.RecordID, body.ID, body.Status); err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": body.ID, "status": body.Status})
}

// handleExpertiseImport reads the Excel tables attached to the comparison
// and matches the remarks with its findings. Called by the viewer and by the
// rule chain "stroykontrol-import-remarks".
func handleExpertiseImport(w http.ResponseWriter, r *http.Request) {
	if cmpRunner == nil {
		writeJSON(w, http.StatusOK, map[string]any{"note": "in --fixtures mode remarks are read from remarks*.xlsx of the case on every request"})
		return
	}
	ns, rec := nsFromQuery(r), strings.TrimSpace(r.URL.Query().Get("recordID"))
	if ns == "" || rec == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("recordID and namespaceID are required"))
		return
	}
	sum, err := cmpRunner.ImportRemarks(r.Context(), ns, rec)
	if err != nil {
		writeErr(w, http.StatusUnprocessableEntity, err)
		return
	}
	writeJSON(w, http.StatusOK, sum)
}
