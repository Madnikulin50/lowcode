package rag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Embedder struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

type embedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type embedResponse struct {
	Embedding []float64 `json:"embedding"`
}

func NewEmbedder(baseURL, model string) *Embedder {
	return &Embedder{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
}

func (e *Embedder) Embed(text string) ([]float32, error) {
	reqBody := embedRequest{
		Model:  e.model,
		Prompt: text,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("embed: marshal: %w", err)
	}
	url := e.baseURL + "/api/embeddings"
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("embed: request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embed: post: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("embed: read: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embed: status %d: %s", resp.StatusCode, string(respBody))
	}
	var er embedResponse
	if err := json.Unmarshal(respBody, &er); err != nil {
		return nil, fmt.Errorf("embed: unmarshal: %w", err)
	}
	emb := make([]float32, len(er.Embedding))
	for i, v := range er.Embedding {
		emb[i] = float32(v)
	}
	return emb, nil
}
