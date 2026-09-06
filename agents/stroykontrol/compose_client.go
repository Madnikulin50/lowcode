package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ComposeClient talks to the Lowcode/Corteza Compose REST API on the agent's
// own behalf (a stored service token), the same way agents/cmdb's web UI
// does — the browser talks only to this agent, never directly to Compose.
type ComposeClient struct {
	base  string // e.g. http://localhost:3333/compose
	token string
	http  *http.Client

	moduleIDsMu sync.Mutex
	moduleIDs   map[string]string // "namespaceID/handle" -> numeric moduleID
}

func NewComposeClient(base, token string) *ComposeClient {
	return &ComposeClient{
		base:      strings.TrimRight(base, "/"),
		token:     token,
		http:      &http.Client{Timeout: 30 * time.Second},
		moduleIDs: map[string]string{},
	}
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
	req, err := http.NewRequest(http.MethodGet, c.base+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
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
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
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

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}
