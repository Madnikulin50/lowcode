package agent

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

type ingestEnvelope struct {
	JobID        string     `json:"jobID"`
	Kind         string     `json:"kind"`
	Status       string     `json:"status"`
	Progress     float64    `json:"progress"`
	Found        int        `json:"found"`
	Error        string     `json:"error,omitempty"`
	ScanningIP   string     `json:"scanningIP,omitempty"`
	Target       string     `json:"target,omitempty"`
	StartedAt    time.Time  `json:"startedAt,omitempty"`
	FinishedAt   *time.Time `json:"finishedAt,omitempty"`
	NamespaceID  uint64     `json:"namespaceID,omitempty"`
	ScanRecordID string     `json:"scanRecordID,omitempty"`
	Items        []Device   `json:"items,omitempty"`
}

func composeJobStatus(agentStatus string) string {
	switch strings.ToLower(strings.TrimSpace(agentStatus)) {
	case "done", "completed", "complete":
		return "completed"
	case "error", "failed", "fail":
		return "failed"
	case "running", "pending":
		return "running"
	default:
		if agentStatus == "" {
			return "running"
		}
		return agentStatus
	}
}

func (a *Agent) notifyCallback(target ScanTarget, id, kind string, items []Device) {
	url := strings.TrimSpace(target.CallbackURL)
	if url == "" {
		return
	}
	if kind == "progress" {
		a.mu.Lock()
		last := a.lastCallback[id]
		if !last.IsZero() && time.Since(last) < 2*time.Second {
			a.mu.Unlock()
			return
		}
		a.lastCallback[id] = time.Now()
		a.mu.Unlock()
		go a.postCallback(target, id, kind, items)
		return
	}
	a.postCallback(target, id, kind, items)
}

func (a *Agent) postCallback(target ScanTarget, id, kind string, items []Device) {
	url := strings.TrimSpace(target.CallbackURL)
	if url == "" {
		return
	}
	st := a.GetStatus(id)
	env := ingestEnvelope{
		JobID:        id,
		Kind:         kind,
		NamespaceID:  target.NamespaceID,
		ScanRecordID: strings.TrimSpace(target.ScanRecordID),
		Items:        items,
	}
	if st != nil {
		env.Status = composeJobStatus(st.Status)
		env.Progress = st.Progress
		env.Found = st.Found
		env.Error = firstNonEmpty(st.Error, st.Message)
		env.ScanningIP = st.ScanningIP
		env.Target = st.Target
		env.StartedAt = st.StartedAt
		env.FinishedAt = st.FinishedAt
		if env.Found == 0 && len(items) > 0 {
			env.Found = len(items)
		}
	} else {
		env.Status = composeJobStatus(kind)
	}
	if kind == "failed" {
		env.Status = "failed"
	}
	if kind == "complete" {
		env.Status = "completed"
	}

	body, err := json.Marshal(env)
	if err != nil {
		log.Printf("callback marshal %s: %v", id[:min(8, len(id))], err)
		return
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		log.Printf("callback request %s: %v", id[:min(8, len(id))], err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if tok := strings.TrimSpace(target.Token); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	client := a.cbClient
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("callback %s %s: %v", kind, id[:min(8, len(id))], err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		log.Printf("callback %s %s: HTTP %d", kind, id[:min(8, len(id))], resp.StatusCode)
	}
}
