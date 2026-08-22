package agent

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (a *Agent) notifyCallback(st *JobStatus, kind string) {
	if st == nil || strings.TrimSpace(st.CallbackURL) == "" {
		return
	}
	if kind == "progress" {
		a.mu.Lock()
		last := a.lastCallback[st.ID]
		if !last.IsZero() && time.Since(last) < 2*time.Second {
			a.mu.Unlock()
			return
		}
		a.lastCallback[st.ID] = time.Now()
		a.mu.Unlock()
	}
	go a.postCallback(st, kind)
}

func (a *Agent) postCallback(st *JobStatus, kind string) {
	url := strings.TrimSpace(st.CallbackURL)
	env := ingestEnvelope{
		JobID:           st.ID,
		Kind:            kind,
		Status:          string(st.Status),
		Progress:        st.Progress,
		BytesRead:       st.BytesRead,
		BytesWritten:    st.BytesWritten,
		Files:           st.Files,
		Error:           st.Error,
		Message:         st.Message,
		SourceID:        st.SourceID,
		PolicyID:        st.PolicyID,
		S3Bucket:        st.S3Bucket,
		S3Key:           st.S3Key,
		Checksum:        st.Checksum,
		Engine:          st.Engine,
		ResticID:        st.ResticID,
		SnapshotID:      st.SnapshotID,
		CreatedRecordID: st.JobRecordID,
		StartedAt:       st.StartedAt,
		FinishedAt:      st.FinishedAt,
	}
	if st.NamespaceID > 0 {
		env.NamespaceID = strconv.FormatUint(st.NamespaceID, 10)
	}
	body, err := json.Marshal(env)
	if err != nil {
		return
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if st.Token != "" {
		req.Header.Set("Authorization", "Bearer "+st.Token)
	}
	client := a.cbClient
	if kind == "complete" || kind == "failed" {
		c := *client
		c.Timeout = 60 * time.Second
		client = &c
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("callback %s %s: %v", kind, st.ID, err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		log.Printf("callback %s %s: HTTP %d", kind, st.ID, resp.StatusCode)
	}
}
