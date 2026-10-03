package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// composeBackend adapts ComposeClient to idcheck.Backend (module handles →
// IDs, write calls, paged search).
type composeBackend struct{ c *ComposeClient }

func toIDRecord(r record) idcheck.Record {
	out := idcheck.Record{ID: r.RecordID, Values: map[string][]string{}}
	for _, v := range r.Values {
		if v.Value == nil {
			continue
		}
		out.Values[v.Name] = append(out.Values[v.Name], fmt.Sprintf("%v", v.Value))
	}
	return out
}

func (b composeBackend) GetRecord(ns, module, id string) (*idcheck.Record, error) {
	r, err := b.c.getRecord(ns, module, id)
	if err != nil {
		return nil, err
	}
	rec := toIDRecord(*r)
	return &rec, nil
}

func (b composeBackend) Search(ns, module, filter string) ([]idcheck.Record, error) {
	modID, err := b.c.resolveModuleID(ns, module)
	if err != nil {
		return nil, err
	}
	var out []idcheck.Record
	cursor := ""
	for page := 0; page < 200; page++ {
		q := url.Values{"limit": {"500"}}
		if filter != "" {
			q.Set("query", filter)
		}
		if cursor != "" {
			q.Set("pageCursor", cursor)
		}
		var rs struct {
			Filter struct {
				NextPage string `json:"nextPage"`
			} `json:"filter"`
			Set []record `json:"set"`
		}
		if err := b.c.get(fmt.Sprintf("/namespace/%s/module/%s/record/?%s", ns, modID, q.Encode()), &rs); err != nil {
			return nil, err
		}
		for _, r := range rs.Set {
			out = append(out, toIDRecord(r))
		}
		if rs.Filter.NextPage == "" || len(rs.Set) == 0 {
			break
		}
		cursor = rs.Filter.NextPage
	}
	return out, nil
}

func valuesPayload(v map[string][]string) []recordValue {
	var out []recordValue
	for name, vals := range v {
		for _, x := range vals {
			if x != "" {
				out = append(out, recordValue{Name: name, Value: x})
			}
		}
	}
	return out
}

func (b composeBackend) Create(ns, module string, values idcheck.Values) (string, error) {
	modID, err := b.c.resolveModuleID(ns, module)
	if err != nil {
		return "", err
	}
	var rec record
	err = b.c.send(http.MethodPost, fmt.Sprintf("/namespace/%s/module/%s/record/", ns, modID),
		map[string]any{"values": valuesPayload(values)}, &rec)
	return rec.RecordID, err
}

func (b composeBackend) Update(ns, module, id string, values idcheck.Values) error {
	modID, err := b.c.resolveModuleID(ns, module)
	if err != nil {
		return err
	}
	var cur struct {
		record
		UpdatedAt string `json:"updatedAt"`
	}
	path := fmt.Sprintf("/namespace/%s/module/%s/record/%s", ns, modID, id)
	if err := b.c.get(path, &cur); err != nil {
		return err
	}
	merged := toIDRecord(cur.record).Values
	for k, v := range values {
		merged[k] = v
	}
	body := map[string]any{"values": valuesPayload(merged)}
	if cur.UpdatedAt != "" {
		body["updatedAt"] = cur.UpdatedAt
	}
	return b.c.send(http.MethodPost, path, body, nil)
}

func (b composeBackend) Delete(ns, module, id string) error {
	modID, err := b.c.resolveModuleID(ns, module)
	if err != nil {
		return err
	}
	return b.c.send(http.MethodDelete, fmt.Sprintf("/namespace/%s/module/%s/record/%s", ns, modID, id), nil, nil)
}

func (b composeBackend) Download(ns, attachmentID string) ([]byte, string, error) {
	att, err := b.c.getAttachment(ns, attachmentID)
	if err != nil {
		return nil, "", err
	}
	data, _, err := b.c.downloadOriginal(ns, attachmentID)
	return data, att.Name, err
}

// send performs a JSON write call with the same token-refresh retry as get.
func (c *ComposeClient) send(method, path string, body, out interface{}) error {
	tok := c.getToken()
	err := c.doSend(method, path, tok, body, out)
	if isExpiredToken(err) && c.refreshToken(tok) {
		err = c.doSend(method, path, c.getToken(), body, out)
	}
	return err
}

func (c *ComposeClient) doSend(method, path, token string, body, out interface{}) error {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.base+path, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var env apiEnvelope
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("decode %s %s: %w (body: %s)", method, path, err, truncate(raw, 300))
		}
	}
	if env.Error != nil && env.Error.Message != "" {
		return fmt.Errorf("%s %s: %s", method, path, env.Error.Message)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("%s %s: HTTP %d", method, path, resp.StatusCode)
	}
	if out != nil && len(env.Response) > 0 {
		return json.Unmarshal(env.Response, out)
	}
	return nil
}
