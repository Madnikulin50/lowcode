package sdk

import (
	"encoding/json"
	"strings"
	"time"
)

const SchemaV1 = "lowcode.agent.v1"

type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

type EventKind string

const (
	KindProgress EventKind = "progress"
	KindComplete EventKind = "complete"
	KindFailed   EventKind = "failed"
)

// Envelope is the wire contract for job status and ingest callbacks.
// MarshalJSON always writes canonical fields plus aliases that existing
// cmdb/backup ingest chains still read (scanID, createdRecordID, found, …).
type Envelope struct {
	Schema      string         `json:"schema"`
	Service     string         `json:"service"`
	Operation   string         `json:"operation,omitempty"`
	ID          string         `json:"id"`
	JobID       string         `json:"jobID"`
	Kind        EventKind      `json:"kind,omitempty"`
	Status      Status         `json:"status"`
	Progress    float64        `json:"progress"`
	Message     string         `json:"message,omitempty"`
	Error       string         `json:"error,omitempty"`
	NamespaceID string         `json:"namespaceID,omitempty"`
	RecordID    string         `json:"recordID,omitempty"`
	StartedAt   time.Time      `json:"startedAt,omitempty"`
	FinishedAt  *time.Time     `json:"finishedAt,omitempty"`
	Result      map[string]any `json:"result,omitempty"`
	Items       any            `json:"items,omitempty"`
}

func NormalizeStatus(s string) Status {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "done", "completed", "complete":
		return StatusCompleted
	case "error", "failed", "fail":
		return StatusFailed
	case "pending":
		return StatusPending
	default:
		return StatusRunning
	}
}

func KindFromStatus(st Status) EventKind {
	switch st {
	case StatusFailed:
		return KindFailed
	case StatusCompleted:
		return KindComplete
	default:
		return KindProgress
	}
}

func (e Envelope) MarshalJSON() ([]byte, error) {
	id := firstNonEmpty(e.ID, e.JobID)
	jobID := firstNonEmpty(e.JobID, e.ID)
	schema := e.Schema
	if schema == "" {
		schema = SchemaV1
	}
	status := e.Status
	if status == "" {
		status = NormalizeStatus(string(e.Kind))
	}
	m := map[string]any{
		"schema":   schema,
		"service":  e.Service,
		"id":       id,
		"jobID":    jobID,
		"scanID":   id,
		"status":   string(status),
		"progress": e.Progress,
	}
	if e.Operation != "" {
		m["operation"] = e.Operation
	}
	if e.Kind != "" {
		m["kind"] = string(e.Kind)
	}
	if e.Message != "" {
		m["message"] = e.Message
	}
	if e.Error != "" {
		m["error"] = e.Error
	}
	if e.NamespaceID != "" {
		m["namespaceID"] = e.NamespaceID
	}
	if e.RecordID != "" {
		m["recordID"] = e.RecordID
		m["createdRecordID"] = e.RecordID
		m["scanRecordID"] = e.RecordID
		m["jobRecordID"] = e.RecordID
	}
	if !e.StartedAt.IsZero() {
		m["startedAt"] = e.StartedAt
	}
	if e.FinishedAt != nil {
		m["finishedAt"] = e.FinishedAt
	}
	if e.Result != nil {
		m["result"] = e.Result
		for k, v := range e.Result {
			if _, exists := m[k]; !exists {
				m[k] = v
			}
		}
	}
	if e.Items != nil {
		m["items"] = e.Items
	}
	return json.Marshal(m)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
