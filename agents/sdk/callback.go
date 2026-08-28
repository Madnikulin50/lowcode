package sdk

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Callback struct {
	client    *http.Client
	mu        sync.Mutex
	last      map[string]time.Time
	minGap    time.Duration
	finalWait time.Duration
}

func NewCallback() *Callback {
	return &Callback{
		client:    &http.Client{Timeout: 15 * time.Second},
		last:      map[string]time.Time{},
		minGap:    2 * time.Second,
		finalWait: 90 * time.Second,
	}
}

func (c *Callback) Notify(j *Job, kind EventKind) {
	if j == nil || strings.TrimSpace(j.CallbackURL) == "" {
		return
	}
	if kind == KindProgress {
		c.mu.Lock()
		last := c.last[j.ID]
		if !last.IsZero() && time.Since(last) < c.minGap {
			c.mu.Unlock()
			return
		}
		c.last[j.ID] = time.Now()
		c.mu.Unlock()
		go c.post(j, kind)
		return
	}
	c.post(j, kind)
}

func (c *Callback) NotifyEnv(id, url, token string, env Envelope) {
	if strings.TrimSpace(url) == "" {
		return
	}
	kind := env.Kind
	if kind == KindProgress {
		c.mu.Lock()
		last := c.last[id]
		if !last.IsZero() && time.Since(last) < c.minGap {
			c.mu.Unlock()
			return
		}
		c.last[id] = time.Now()
		c.mu.Unlock()
		go c.postEnv(url, token, env, kind)
		return
	}
	c.postEnv(url, token, env, kind)
}

func (c *Callback) post(j *Job, kind EventKind) {
	url := strings.TrimSpace(j.CallbackURL)
	if url == "" {
		return
	}
	c.postEnv(url, j.Token, j.Envelope(kind), kind)
}

func (c *Callback) postEnv(url, token string, env Envelope, kind EventKind) {
	body, err := json.Marshal(env)
	if err != nil {
		log.Printf("sdk callback marshal: %v", err)
		return
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		log.Printf("sdk callback request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if tok := strings.TrimSpace(token); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	client := c.client
	if kind == KindComplete || kind == KindFailed {
		cp := *client
		cp.Timeout = c.finalWait
		client = &cp
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("sdk callback %s %s: %v", kind, env.ID, err)
		return
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	if resp.StatusCode >= 400 {
		log.Printf("sdk callback %s %s: HTTP %d", kind, env.ID, resp.StatusCode)
	}
}
