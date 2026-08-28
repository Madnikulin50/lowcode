package sdk

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEnvelopeCompatAliases(t *testing.T) {
	fin := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	start := fin.Add(-time.Minute)
	env := Envelope{
		Service:     "cmdb",
		Operation:   "scan",
		ID:          "scan-1",
		Kind:        KindComplete,
		Status:      StatusCompleted,
		Progress:    100,
		NamespaceID: "42",
		RecordID:    "99",
		StartedAt:   start,
		FinishedAt:  &fin,
		Result: map[string]any{
			"found":      1,
			"scanningIP": "10.0.0.1",
			"target":     "10.0.0.0/24",
		},
		Items: []map[string]string{{"ip": "10.0.0.1"}},
	}
	body, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatal(err)
	}
	wantStr := map[string]string{
		"schema":          `"lowcode.agent.v1"`,
		"service":         `"cmdb"`,
		"id":              `"scan-1"`,
		"jobID":           `"scan-1"`,
		"scanID":          `"scan-1"`,
		"status":          `"completed"`,
		"kind":            `"complete"`,
		"namespaceID":     `"42"`,
		"recordID":        `"99"`,
		"createdRecordID": `"99"`,
		"scanRecordID":    `"99"`,
		"jobRecordID":     `"99"`,
		"found":           `1`,
		"scanningIP":      `"10.0.0.1"`,
	}
	for k, want := range wantStr {
		got := string(raw[k])
		if got != want {
			t.Errorf("%s: got %s want %s", k, got, want)
		}
	}
	if _, ok := raw["items"]; !ok {
		t.Fatal("items missing")
	}
}

func TestEnvelopeBackupFlatten(t *testing.T) {
	env := Envelope{
		Service:   "backup",
		Operation: "backup",
		ID:        "job-1",
		Kind:      KindComplete,
		Status:    StatusCompleted,
		RecordID:  "555",
		Result: map[string]any{
			"s3Key":        "a/b.tar.gz",
			"s3Bucket":     "backups",
			"bytesRead":    int64(10),
			"bytesWritten": int64(12),
			"files":        3,
			"checksum":     "abc",
			"engine":       "archive",
		},
	}
	body, _ := json.Marshal(env)
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["s3Key"] != "a/b.tar.gz" || raw["jobRecordID"] != "555" {
		t.Fatalf("flatten %+v", raw)
	}
}

func TestNormalizeStatus(t *testing.T) {
	if NormalizeStatus("done") != StatusCompleted {
		t.Fatal("done")
	}
	if NormalizeStatus("error") != StatusFailed {
		t.Fatal("error")
	}
	if NormalizeStatus("running") != StatusRunning {
		t.Fatal("running")
	}
}

func TestDecodeStartRequest(t *testing.T) {
	req, err := DecodeStartRequest([]byte(`{
		"cidr":"10.0.0.0/24",
		"namespaceID":"509463708777775105",
		"token":"tok",
		"callbackUrl":"http://cb",
		"scanRecordID":"99"
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if req.NamespaceID.Uint64() != 509463708777775105 {
		t.Fatalf("ns %d", req.NamespaceID)
	}
	if req.RecordID != "99" || req.Param("cidr") != "10.0.0.0/24" {
		t.Fatalf("%+v params=%v", req, req.Params)
	}
	req, err = DecodeStartRequest([]byte(`{"sourceID":"1","kind":"restore","namespaceID":42}`))
	if err != nil {
		t.Fatal(err)
	}
	if req.NamespaceID.Uint64() != 42 || req.Param("sourceID") != "1" || req.Param("kind") != "restore" {
		t.Fatalf("%+v", req)
	}
}
