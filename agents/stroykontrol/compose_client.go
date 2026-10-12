package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/cmp"
	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// ComposeClient talks to the Lowcode/Corteza Compose REST API on the agent's
// own behalf (a stored service token), the same way agents/cmdb's web UI
// does — the browser talks only to this agent, never directly to Compose.
type ComposeClient struct {
	base string // e.g. http://localhost:3333/compose
	http *http.Client

	tokenMu sync.RWMutex
	token   string

	// remint mints a fresh token (e.g. re-running mintToken(), see run.sh/
	// mint-token.sh) when the stored one has expired — the dev token is
	// short-lived (~2h) and the agent otherwise holds it for its whole
	// process lifetime (see README's "Token stability" section). Optional;
	// nil disables auto-refresh and preserves the old fail-hard behaviour.
	remint   func() (string, error)
	remintMu sync.Mutex

	moduleIDsMu sync.Mutex
	moduleIDs   map[string]string // "namespaceID/handle" -> numeric moduleID
}

func NewComposeClient(base, token string, remint func() (string, error)) *ComposeClient {
	return &ComposeClient{
		base:      strings.TrimRight(base, "/"),
		token:     token,
		remint:    remint,
		http:      &http.Client{Timeout: 30 * time.Second},
		moduleIDs: map[string]string{},
	}
}

func (c *ComposeClient) getToken() string {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	return c.token
}

func (c *ComposeClient) setToken(t string) {
	c.tokenMu.Lock()
	c.token = t
	c.tokenMu.Unlock()
}

// isExpiredToken reports whether err looks like an expired-token response
// worth a re-mint-and-retry: either Compose's own JSON error message (the
// /api/... envelope path, see get()) or a bare HTTP 401 (the raw attachment
// download path, see downloadOriginal(), which gets no JSON envelope).
func isExpiredToken(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// "unauthorized": the dev token was rotated by another mintToken() call
	// (see README "Token stability") — re-minting recovers that too.
	return strings.Contains(msg, "token is expired") || strings.Contains(msg, "HTTP 401") || strings.Contains(msg, "unauthorized")
}

// refreshToken re-mints the token, but only if it still matches failedToken
// — if a concurrent request already refreshed it while this one was waiting
// for the lock, that refresh is reused instead of minting again. mintToken()
// can rotate/invalidate whatever was minted just before it (see README), so
// overlapping mints from concurrent requests must be avoided. Returns true
// if the caller should retry with the (possibly just-refreshed) token.
func (c *ComposeClient) refreshToken(failedToken string) bool {
	if c.remint == nil {
		return false
	}
	c.remintMu.Lock()
	defer c.remintMu.Unlock()
	if c.getToken() != failedToken {
		return true
	}
	tok, err := c.remint()
	if err != nil || tok == "" {
		log.Printf("token refresh failed: %v", err)
		return false
	}
	c.setToken(tok)
	return true
}

// resolveModuleID turns a module handle into the numeric ID the record
// routes actually require (they don't accept handles in the path), caching
// the result — module IDs don't change at runtime.
func (c *ComposeClient) resolveModuleID(namespaceID, handle string) (string, error) {
	key := namespaceID + "/" + handle
	c.moduleIDsMu.Lock()
	if id, ok := c.moduleIDs[key]; ok {
		c.moduleIDsMu.Unlock()
		return id, nil
	}
	c.moduleIDsMu.Unlock()

	var out struct {
		Set []struct {
			ModuleID string `json:"moduleID"`
			Handle   string `json:"handle"`
		} `json:"set"`
	}
	q := "?handle=" + url.QueryEscape(handle) + "&limit=1"
	if err := c.get(fmt.Sprintf("/namespace/%s/module/%s", namespaceID, q), &out); err != nil {
		return "", err
	}
	if len(out.Set) == 0 {
		return "", fmt.Errorf("module %q not found in namespace %s", handle, namespaceID)
	}
	id := out.Set[0].ModuleID

	c.moduleIDsMu.Lock()
	c.moduleIDs[key] = id
	c.moduleIDsMu.Unlock()
	return id, nil
}

type apiEnvelope struct {
	Error    *apiError       `json:"error"`
	Response json.RawMessage `json:"response"`
}

type apiError struct {
	Message string `json:"message"`
}

func (c *ComposeClient) get(path string, out interface{}) error {
	tok := c.getToken()
	err := c.doGet(path, tok, out)
	if isExpiredToken(err) && c.refreshToken(tok) {
		err = c.doGet(path, c.getToken(), out)
	}
	return err
}

func (c *ComposeClient) doGet(path, token string, out interface{}) error {
	req, err := http.NewRequest(http.MethodGet, c.base+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return fmt.Errorf("decode %s: %w (body: %s)", path, err, truncate(body, 300))
	}
	if env.Error != nil && env.Error.Message != "" {
		return fmt.Errorf("%s: %s", path, env.Error.Message)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("%s: HTTP %d", path, resp.StatusCode)
	}
	if out != nil && len(env.Response) > 0 {
		return json.Unmarshal(env.Response, out)
	}
	return nil
}

type attachmentMeta struct {
	AttachmentID string `json:"attachmentID"`
	Name         string `json:"name"`
	Url          string `json:"url"`
}

// getAttachment reads attachment metadata; its Url comes back freshly
// signed (sign+userID) for the agent's own identity — record-kind
// attachments reject plain Bearer auth on /original without that signature.
func (c *ComposeClient) getAttachment(namespaceID, attachmentID string) (*attachmentMeta, error) {
	var a attachmentMeta
	if err := c.get(fmt.Sprintf("/namespace/%s/attachment/record/%s", namespaceID, attachmentID), &a); err != nil {
		return nil, err
	}
	return &a, nil
}

// downloadOriginal fetches the raw content of a record attachment.
func (c *ComposeClient) downloadOriginal(namespaceID, attachmentID string) ([]byte, string, error) {
	att, err := c.getAttachment(namespaceID, attachmentID)
	if err != nil {
		return nil, "", fmt.Errorf("attachment metadata: %w", err)
	}
	if att.Url == "" {
		return nil, "", fmt.Errorf("attachment %s has no url", attachmentID)
	}
	// att.Url is relative to the Compose API base (matches how the webapp
	// itself resolves it: window.CortezaAPI + url), not the bare origin.
	u := att.Url
	if strings.HasPrefix(u, "/") {
		u = c.base + u
	}
	tok := c.getToken()
	data, mimetype, err := c.doDownload(u, attachmentID, tok)
	if isExpiredToken(err) && c.refreshToken(tok) {
		data, mimetype, err = c.doDownload(u, attachmentID, c.getToken())
	}
	return data, mimetype, err
}

func (c *ComposeClient) doDownload(u, attachmentID, token string) ([]byte, string, error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode >= 400 {
		return nil, "", fmt.Errorf("download %s: HTTP %d (%s)", attachmentID, resp.StatusCode, u)
	}
	return data, resp.Header.Get("Content-Type"), nil
}

type recordValue struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

type record struct {
	RecordID string        `json:"recordID"`
	Values   []recordValue `json:"values"`
}

func (r *record) val(name string) string {
	for _, v := range r.Values {
		if v.Name == name {
			if v.Value == nil {
				return ""
			}
			return fmt.Sprintf("%v", v.Value)
		}
	}
	return ""
}

func (c *ComposeClient) getRecord(namespaceID, moduleHandle, recordID string) (*record, error) {
	modID, err := c.resolveModuleID(namespaceID, moduleHandle)
	if err != nil {
		return nil, err
	}
	var rec record
	if err := c.get(fmt.Sprintf("/namespace/%s/module/%s/record/%s", namespaceID, modID, recordID), &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

type recordSet struct {
	Set []record `json:"set"`
}

func (c *ComposeClient) searchRecords(namespaceID, moduleHandle, filter string) ([]record, error) {
	modID, err := c.resolveModuleID(namespaceID, moduleHandle)
	if err != nil {
		return nil, err
	}
	var rs recordSet
	q := ""
	if filter != "" {
		q = "?query=" + url.QueryEscape(filter) + "&limit=500"
	} else {
		q = "?limit=500"
	}
	if err := c.get(fmt.Sprintf("/namespace/%s/module/%s/record/%s", namespaceID, modID, q), &rs); err != nil {
		return nil, err
	}
	return rs.Set, nil
}

// --- store implementation (see store.go) --------------------------------
//
// The methods below are the only ones main.go's handlers call; getRecord/
// searchRecords/getAttachment/downloadOriginal above stay Compose-specific
// plumbing. FixtureStore (fixture_store.go) implements the same store
// interface without any of that plumbing.

func (c *ComposeClient) Comparison(ns, recordID string) (*comparisonRecord, error) {
	rec, err := c.getRecord(ns, comparisonModule, recordID)
	if err != nil {
		return nil, err
	}
	out := &comparisonRecord{
		Title:             rec.val("title"),
		Status:            rec.val("status"),
		SimilarityPercent: rec.val("similarity_percent"),
		Comment:           rec.val("comment"),
		ParentID:          rec.val("parent_comparison"),
		IterationNo:       rec.val("iteration_no"),
	}
	if out.ParentID != "" && out.ParentID != recordID {
		if p, err := c.getRecord(ns, comparisonModule, out.ParentID); err == nil {
			out.ParentTitle = p.val("title")
		}
	} else {
		out.ParentID = ""
	}
	return out, nil
}

func (c *ComposeClient) Discrepancies(ns, recordID string) ([]discrepancyRecord, error) {
	recs, err := c.searchRecords(ns, discrepancyModule, fmt.Sprintf("comparison = '%s'", recordID))
	if err != nil {
		return nil, err
	}
	out := make([]discrepancyRecord, 0, len(recs))
	for _, d := range recs {
		out = append(out, discrepancyRecord{
			ID:          d.RecordID,
			Type:        d.val("discrepancy_type"),
			Severity:    d.val("severity"),
			Description: d.val("description"),
			PageNumber:  d.val("page_number"),
			Kind:        d.val("kind"),
			Status:      d.val("status"),
			DocRef:      d.val("doc_ref"),
			DocTarget:   d.val("doc_target"),
			PageRef:     d.val("page_pd"),
			PageTarget:  d.val("page_rd"),
			BboxRef:     d.val("bbox_pd"),
			BboxTarget:  d.val("bbox_rd"),
			Anchor:      d.val("anchor_text"),
			Evidence:    d.val("evidence"),
			Fingerprint: d.val("fingerprint"),
			IterState:   d.val("iter_state"),
		})
	}
	return out, nil
}

// attachmentName is a best-effort label lookup: the UI falls back to a plain
// "ПД"/"РД" label if this fails — it is a label, not worth failing a request.
func (c *ComposeClient) attachmentName(ns, attachmentID string) string {
	if attachmentID == "" {
		return ""
	}
	att, err := c.getAttachment(ns, attachmentID)
	if err != nil {
		log.Printf("attachment name lookup failed for %s: %v", attachmentID, err)
		return ""
	}
	return att.Name
}

const (
	legacyPD = "legacy:pd"
	legacyRD = "legacy:rd"
)

// Documents lists the package of a comparison (cmp_documents records) and,
// for comparisons created before packs existed, the single pd_file/rd_file.
func (c *ComposeClient) Documents(ns, recordID string) ([]documentRecord, error) {
	var out []documentRecord
	cmpRec, err := c.getRecord(ns, comparisonModule, recordID)
	if err != nil {
		return nil, err
	}
	if id := cmpRec.val("pd_file"); id != "" {
		out = append(out, documentRecord{ID: legacyPD, Side: cmp.SideRef, Name: c.attachmentName(ns, id)})
	}
	recs, err := c.searchRecords(ns, cmp.ModDocuments, fmt.Sprintf("comparison = '%s'", recordID))
	if err != nil {
		log.Printf("cmp_documents lookup failed: %v", err)
		recs = nil
	}
	// a later iteration inherits the reference set (ПД) of the previous one
	hasRef := false
	for _, d := range recs {
		hasRef = hasRef || (d.val("side") == cmp.SideRef && d.val("file") != "")
	}
	if !hasRef {
		for _, anc := range cmp.Ancestors(composeBackend{c}, ns, recordID)[1:] {
			ar, err := c.searchRecords(ns, cmp.ModDocuments, fmt.Sprintf("comparison = '%s'", anc))
			if err != nil {
				continue
			}
			var inherited []record
			for _, d := range ar {
				if d.val("side") == cmp.SideRef && d.val("file") != "" {
					inherited = append(inherited, d)
				}
			}
			if len(inherited) > 0 {
				recs = append(inherited, recs...)
				break
			}
		}
	}
	for _, d := range recs {
		att := d.val("file")
		if att == "" {
			continue
		}
		side := d.val("side")
		if side != cmp.SideRef {
			side = cmp.SideTarget
		}
		out = append(out, documentRecord{
			ID: d.RecordID, Side: side, Name: c.attachmentName(ns, att),
			VolumeCode: d.val("volume_code"), VolumeTitle: d.val("volume_title"), Pages: d.val("pages"),
		})
	}
	if id := cmpRec.val("rd_file"); id != "" {
		out = append(out, documentRecord{ID: legacyRD, Side: cmp.SideTarget, Name: c.attachmentName(ns, id)})
	}
	return out, nil
}

// FetchDocument downloads one document. A package document is looked up by
// its record and must belong to recordID, so the endpoint cannot be used to
// read arbitrary attachments of the namespace.
func (c *ComposeClient) FetchDocument(ns, recordID, docID string) ([]byte, string, error) {
	field := ""
	switch docID {
	case legacyPD:
		field = "pd_file"
	case legacyRD:
		field = "rd_file"
	}
	if field != "" {
		rec, err := c.getRecord(ns, comparisonModule, recordID)
		if err != nil {
			return nil, "", err
		}
		attID := rec.val(field)
		if attID == "" {
			return nil, "", fmt.Errorf("record has no %s attachment", field)
		}
		return c.downloadOriginal(ns, attID)
	}
	d, err := c.getRecord(ns, cmp.ModDocuments, docID)
	if err != nil {
		return nil, "", err
	}
	if !c.inChain(ns, recordID, d.val("comparison")) {
		return nil, "", fmt.Errorf("document %s does not belong to comparison %s", docID, recordID)
	}
	attID := d.val("file")
	if attID == "" {
		return nil, "", fmt.Errorf("document %s has no file", docID)
	}
	return c.downloadOriginal(ns, attID)
}

// inChain reports whether owner is the comparison or one of its previous
// iterations (whose reference documents it inherits).
func (c *ComposeClient) inChain(ns, recordID, owner string) bool {
	for _, id := range cmp.Ancestors(composeBackend{c}, ns, recordID) {
		if id == owner {
			return true
		}
	}
	return false
}

func (c *ComposeClient) SetDiscrepancyStatus(ns, recordID, discrepancyID, status string) error {
	d, err := c.getRecord(ns, discrepancyModule, discrepancyID)
	if err != nil {
		return err
	}
	if d.val("comparison") != recordID {
		return fmt.Errorf("discrepancy %s does not belong to comparison %s", discrepancyID, recordID)
	}
	return composeBackend{c}.Update(ns, discrepancyModule, discrepancyID, idcheck.V("status", status))
}

func remarkFromRecord(r record) cmp.Remark {
	var ids []string
	for _, v := range r.Values {
		if v.Name == "discrepancies" && v.Value != nil {
			ids = append(ids, fmt.Sprintf("%v", v.Value))
		}
	}
	return cmp.Remark{
		ID: r.RecordID, DiscrepancyIDs: ids,
		DocumentRef: r.val("document_ref"), Sheet: r.val("sheet"),
		WhatDiverges: r.val("what_diverges"), MustAlign: r.val("must_align"),
		DraftText: r.val("draft_text"), FinalText: r.val("final_text"), Status: r.val("status"),
	}
}

func (c *ComposeClient) Remarks(ns, recordID string) ([]cmp.Remark, error) {
	recs, err := c.searchRecords(ns, cmp.ModRemarks, fmt.Sprintf("comparison = '%s'", recordID))
	if err != nil {
		return nil, err
	}
	out := make([]cmp.Remark, 0, len(recs))
	for _, r := range recs {
		out = append(out, remarkFromRecord(r))
	}
	return out, nil
}

func (c *ComposeClient) SaveRemark(ns, recordID string, r cmp.Remark) (cmp.Remark, error) {
	v := idcheck.V(
		"comparison", recordID,
		"document_ref", r.DocumentRef, "sheet", r.Sheet,
		"what_diverges", r.WhatDiverges, "must_align", r.MustAlign,
		"draft_text", r.DraftText, "final_text", r.FinalText, "status", r.Status,
	)
	if len(r.DiscrepancyIDs) > 0 {
		v["discrepancies"] = r.DiscrepancyIDs
	}
	b := composeBackend{c}
	if r.ID == "" {
		id, err := b.Create(ns, cmp.ModRemarks, v)
		if err != nil {
			return r, err
		}
		r.ID = id
		return r, nil
	}
	existing, err := c.getRecord(ns, cmp.ModRemarks, r.ID)
	if err != nil {
		return r, err
	}
	if existing.val("comparison") != recordID {
		return r, fmt.Errorf("remark %s does not belong to comparison %s", r.ID, recordID)
	}
	// Update overlays non-empty values only: a field cannot be blanked from the UI.
	return r, b.Update(ns, cmp.ModRemarks, r.ID, v)
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}

func (c *ComposeClient) Expertise(ns, recordID string) ([]expertiseRecord, error) {
	recs, err := c.searchRecords(ns, cmp.ModExpertise, fmt.Sprintf("comparison = '%s'", recordID))
	if err != nil {
		return nil, err
	}
	out := make([]expertiseRecord, 0, len(recs))
	for _, r := range recs {
		e := expertiseRecord{
			ID: r.RecordID, Number: r.val("number"), Text: r.val("text"), Answer: r.val("answer"),
			Location: r.val("location"), Source: r.val("source_file"), Verdict: r.val("verdict"),
			Score: r.val("match_score"), Note: r.val("match_note"), Status: r.val("status"),
			ClaimMismatch: r.val("claim_mismatch") != "",
		}
		for _, v := range r.Values {
			if v.Name == "matched_discrepancies" && v.Value != nil {
				e.Findings = append(e.Findings, fmt.Sprintf("%v", v.Value))
			}
		}
		out = append(out, e)
	}
	sortExpertise(out)
	return out, nil
}

func (c *ComposeClient) SetExpertiseStatus(ns, recordID, remarkID, status string) error {
	r, err := c.getRecord(ns, cmp.ModExpertise, remarkID)
	if err != nil {
		return err
	}
	if r.val("comparison") != recordID {
		return fmt.Errorf("remark %s does not belong to comparison %s", remarkID, recordID)
	}
	return composeBackend{c}.Update(ns, cmp.ModExpertise, remarkID, idcheck.V("status", status))
}
