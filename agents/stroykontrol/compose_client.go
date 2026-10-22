package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
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
	remint    func() (string, error)
	remintMu  sync.Mutex
	remintErr error // why the last refresh failed, shown to the user with the expired-token error

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
		c.remintErr = err
		return false
	}
	c.remintErr = nil
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
	if isExpiredToken(err) {
		if c.refreshToken(tok) {
			err = c.doGet(path, c.getToken(), out)
		} else if c.remintErr != nil {
			err = fmt.Errorf("%w — токен не удалось обновить автоматически: %v. Перезапустите агент или выполните ./mint-token.sh", err, c.remintErr)
		}
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
		ObjectID:          rec.val("object"),
		SetsLabel:         rec.val("sets_label"),
		Kind:              modeOf(rec.val("kind")),
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
			Comment:     d.val("reviewer_comment"),
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

func (c *ComposeClient) SetDiscrepancyStatus(ns, recordID, discrepancyID, status, comment string) error {
	d, err := c.getRecord(ns, discrepancyModule, discrepancyID)
	if err != nil {
		return err
	}
	if d.val("comparison") != recordID {
		return fmt.Errorf("discrepancy %s does not belong to comparison %s", discrepancyID, recordID)
	}
	return composeBackend{c}.Update(ns, discrepancyModule, discrepancyID, idcheck.V("status", status, "reviewer_comment", comment))
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

func (c *ComposeClient) RegisterIssued(ns, comparisonID string, in cmp.IssueInput, res cmp.IssueResult) (*cmp.IssueRegistered, error) {
	rec, err := c.getRecord(ns, comparisonModule, comparisonID)
	if err != nil {
		return nil, err
	}
	return cmp.IssueToRegistry(composeBackend{c}, ns, rec.val("object"), in, res)
}

func (c *ComposeClient) IssuedNumbers(ns, comparisonID string) (map[string]string, error) {
	rec, err := c.getRecord(ns, comparisonModule, comparisonID)
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	if rec.val("object") == "" {
		return out, nil
	}
	recs, err := composeBackend{c}.Search(ns, cmp.ModExpertise, fmt.Sprintf("object = '%s'", rec.val("object")))
	if err != nil {
		return nil, err
	}
	for _, r := range recs {
		if r.Get("origin") == cmp.OriginComparison && r.Get("source_comparison") == comparisonID {
			out[r.Get("group_key")] = r.Get("number")
		}
	}
	return out, nil
}

func (c *ComposeClient) ObjectName(ns, objectID string) string {
	if objectID == "" {
		return ""
	}
	if o, err := c.getRecord(ns, cmp.ModObjects, objectID); err == nil {
		return o.val("name")
	}
	return ""
}

// PageID is the ID of the page with the handle in the namespace ("" if there is none); the
// viewers use it to link to record cards in the Compose UI.
func (c *ComposeClient) PageID(ns, handle string) string {
	key := "page:" + ns + "/" + handle
	c.moduleIDsMu.Lock()
	if id, ok := c.moduleIDs[key]; ok {
		c.moduleIDsMu.Unlock()
		return id
	}
	c.moduleIDsMu.Unlock()
	var out struct {
		Set []struct {
			PageID string `json:"pageID"`
			Handle string `json:"handle"`
		} `json:"set"`
	}
	if err := c.get(fmt.Sprintf("/namespace/%s/page/?handle=%s&limit=20", ns, url.QueryEscape(handle)), &out); err != nil {
		return ""
	}
	for _, p := range out.Set {
		if p.Handle == handle {
			c.moduleIDsMu.Lock()
			c.moduleIDs[key] = p.PageID
			c.moduleIDsMu.Unlock()
			return p.PageID
		}
	}
	return ""
}

// modeOf maps the "kind" field of a comparison to a cmp mode.
func modeOf(kind string) string {
	if kind == cmp.ModeVersions {
		return cmp.ModeVersions
	}
	return cmp.ModeStages
}

// Pairing implements store.
func (c *ComposeClient) Pairing(ns, recordID string) (*pairingOut, error) {
	rec, err := c.getRecord(ns, comparisonModule, recordID)
	if err != nil {
		return nil, err
	}
	docs, err := c.Documents(ns, recordID)
	if err != nil {
		return nil, err
	}
	p, err := cmp.ParsePairing(rec.val("pairing"))
	if err != nil {
		return nil, err
	}
	out := &pairingOut{Documents: docs, Pairing: p}
	if out.Pairing == nil {
		out.Pairing = []cmp.ManualPair{}
	}
	if out.Excluded, err = cmp.ParseRegions(rec.val("excluded_regions")); err != nil {
		return nil, err
	}
	if out.Excluded == nil {
		out.Excluded = []cmp.Region{}
	}
	if raw := rec.val("pairs_json"); raw != "" {
		var st cmp.PairsState
		if json.Unmarshal([]byte(raw), &st) == nil {
			out.State = &st
		}
	}
	return out, nil
}

// SetPairing implements store.
func (c *ComposeClient) SetPairing(ns, recordID string, p []cmp.ManualPair) error {
	v := idcheck.V("pairing", cmp.PairingJSON(p))
	if len(p) == 0 {
		v["pairing"] = []string{""} // an explicit empty value clears the stored one
	}
	return composeBackend{c}.Update(ns, comparisonModule, recordID, v)
}

// ReportIndex implements store: the working objects with their comparisons and the size of their registries.
func (c *ComposeClient) ReportIndex(ns string) ([]reportObject, error) {
	b := composeBackend{c}
	objs, err := b.Search(ns, cmp.ModObjects, "")
	if err != nil {
		return nil, err
	}
	byID := map[string]*reportObject{}
	var out []*reportObject
	for _, o := range objs {
		if o.Get("data_kind") == "demo" {
			continue
		}
		ro := &reportObject{ID: o.ID, Name: o.Get("name"), Code: o.Get("code"), Comparisons: []reportComparison{}}
		byID[o.ID] = ro
		out = append(out, ro)
	}
	if comps, err := b.Search(ns, cmp.ModComparisons, ""); err == nil {
		for _, cp := range comps {
			if ro := byID[cp.Get("object")]; ro != nil {
				ro.Comparisons = append(ro.Comparisons, reportComparison{ID: cp.ID, Title: cp.Get("title"), Kind: cp.Get("kind"), Iteration: cp.Get("iteration_no"), Status: cp.Get("status")})
			}
		}
	}
	if rems, err := b.Search(ns, cmp.ModExpertise, ""); err == nil {
		for _, rm := range rems {
			if ro := byID[rm.Get("object")]; ro != nil && rm.Get("archived") == "" {
				ro.Remarks++
			}
		}
	}
	if pk, err := b.Search(ns, cmp.ModCorrection, ""); err == nil {
		for _, p := range pk {
			if ro := byID[p.Get("object")]; ro != nil {
				ro.Packages++
			}
		}
	}
	res := make([]reportObject, 0, len(out))
	for _, ro := range out {
		sort.Slice(ro.Comparisons, func(i, j int) bool { return ro.Comparisons[i].Title < ro.Comparisons[j].Title })
		res = append(res, *ro)
	}
	sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })
	return res, nil
}

// SetExclusions implements store.
func (c *ComposeClient) SetExclusions(ns, recordID string, r []cmp.Region) error {
	v := idcheck.V("excluded_regions", cmp.RegionsJSON(r))
	if len(r) == 0 {
		v["excluded_regions"] = []string{""} // an explicit empty value clears the stored one
	}
	return composeBackend{c}.Update(ns, comparisonModule, recordID, v)
}

// downloadDocument returns the file of a Compose document (the Docs section).
func (c *ComposeClient) downloadDocument(ns, documentID string) ([]byte, string, error) {
	u := fmt.Sprintf("%s/namespace/%s/document/%s/file", c.base, ns, documentID)
	tok := c.getToken()
	data, mimetype, err := c.doDownload(u, documentID, tok)
	if isExpiredToken(err) && c.refreshToken(tok) {
		data, mimetype, err = c.doDownload(u, documentID, c.getToken())
	}
	return data, mimetype, err
}
