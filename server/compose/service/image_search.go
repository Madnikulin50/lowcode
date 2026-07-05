package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type (
	ImageSearchResult struct {
		Title     string `json:"title"`
		ImageURL  string `json:"image"`
		Thumbnail string `json:"thumbnail"`
		SourceURL string `json:"url"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	}

	ImageSearchService interface {
		Search(ctx context.Context, query string, limit int) ([]ImageSearchResult, error)
	}

	imageSearch struct {
		client  *http.Client
		enabled bool
	}

	ddgImageResponse struct {
		Results []struct {
			Image     string `json:"image"`
			Title     string `json:"title"`
			URL       string `json:"url"`
			Thumbnail string `json:"thumbnail"`
			Height    int    `json:"height"`
			Width     int    `json:"width"`
		} `json:"results"`
	}
)

func ImageSearch(enabled bool) *imageSearch {
	return &imageSearch{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		enabled: enabled,
	}
}

func (s *imageSearch) Search(ctx context.Context, query string, limit int) ([]ImageSearchResult, error) {
	if !s.enabled {
		return nil, fmt.Errorf("image search is disabled")
	}

	if limit <= 0 || limit > 50 {
		limit = 10
	}

	vqd, err := s.getVQD(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get VQD token: %w", err)
	}

	u := fmt.Sprintf("https://duckduckgo.com/i.js?q=%s&vqd=%s&o=json&p=1&f=,,,&l=wt-wt",
		url.QueryEscape(query), url.QueryEscape(vqd))

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("duckduckgo returned status %d: %s", resp.StatusCode, string(body))
	}

	var imgResp ddgImageResponse
	if err := json.Unmarshal(body, &imgResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if limit > len(imgResp.Results) {
		limit = len(imgResp.Results)
	}

	results := make([]ImageSearchResult, 0, limit)
	for _, r := range imgResp.Results[:limit] {
		results = append(results, ImageSearchResult{
			Title:     r.Title,
			ImageURL:  sanitizeURL(r.Image),
			Thumbnail: r.Thumbnail,
			SourceURL: sanitizeURL(r.URL),
			Width:     r.Width,
			Height:    r.Height,
		})
	}

	return results, nil
}

func (s *imageSearch) getVQD(ctx context.Context, query string) (string, error) {
	u := fmt.Sprintf("https://duckduckgo.com/?q=%s&iax=images&ia=images", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`vqd=([\d-]+)`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		re2 := regexp.MustCompile(`"vqd"\s*:\s*"([\d-]+)"`)
		matches = re2.FindStringSubmatch(string(body))
	}
	if len(matches) < 2 {
		return "", fmt.Errorf("could not find VQD token")
	}

	return strings.TrimSpace(matches[1]), nil
}

func sanitizeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "//") {
		raw = "https:" + raw
	}
	return raw
}
