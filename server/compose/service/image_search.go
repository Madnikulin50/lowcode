package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
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
	// getVQD's response very likely sets a session cookie DuckDuckGo expects
	// back on the follow-up i.js request that actually carries that VQD
	// token — without it, the second request looks like it's presenting a
	// token that isn't its own, which is exactly the shape of thing
	// DuckDuckGo's anomaly detection (the 403 "please let us know" page)
	// flags. jar can only fail on a bad PublicSuffixList, which we don't
	// set — the returned error is always nil here.
	jar, _ := cookiejar.New(nil)

	return &imageSearch{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Jar:     jar,
			// DuckDuckGo (an unofficial, scraped endpoint — there's no
			// official API/key for this) sometimes closes a pooled
			// keep-alive connection from its side; Go's default transport
			// then hands the next request a dead connection and read fails
			// with a bare EOF before any response is seen. These are
			// one-off, infrequent search requests — reuse buys us nothing
			// here — so disable keep-alives entirely rather than chase
			// stale-connection races.
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
		},
		enabled: enabled,
	}
}

// browserHeaders makes the request look like an ordinary browser hit rather
// than a bare Go http.Client — DuckDuckGo's bot filtering appears to key off
// more than just User-Agent.
func browserHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,ru;q=0.8")
}

// doWithRetry retries a small, fixed number of times on transport-level
// failures only (connection reset, EOF before any response, timeouts) — the
// class of error a dead pooled connection or a transient network blip
// produces, not on the request being rejected outright. It never retries
// once a response was actually received, successful or not, since that's a
// real answer, not a network hiccup.
func doWithRetry(ctx context.Context, client *http.Client, req *http.Request) (resp *http.Response, err error) {
	backoffs := []time.Duration{0, 300 * time.Millisecond, 900 * time.Millisecond}
	for attempt, backoff := range backoffs {
		if backoff > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}
		resp, err = client.Do(req)
		if err == nil {
			return resp, nil
		}
		if attempt == len(backoffs)-1 {
			return nil, err
		}
	}
	return nil, err
}

func (s *imageSearch) Search(ctx context.Context, query string, limit int) ([]ImageSearchResult, error) {
	if !s.enabled {
		return nil, fmt.Errorf("image search is disabled")
	}

	if limit <= 0 || limit > 50 {
		limit = 10
	}

	vqd, refererURL, err := s.getVQD(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get VQD token: %w", err)
	}

	u := fmt.Sprintf("https://duckduckgo.com/i.js?q=%s&vqd=%s&o=json&p=1&f=,,,&l=wt-wt",
		url.QueryEscape(query), url.QueryEscape(vqd))

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	browserHeaders(req)
	// A real browser arrives at i.js from the search page it just rendered
	// — send the same signal.
	req.Header.Set("Referer", refererURL)

	resp, err := doWithRetry(ctx, s.client, req)
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

// getVQD returns the page's VQD token and the page's own URL (for the
// follow-up i.js request's Referer — see Search).
func (s *imageSearch) getVQD(ctx context.Context, query string) (string, string, error) {
	u := fmt.Sprintf("https://duckduckgo.com/?q=%s&iax=images&ia=images", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return "", "", err
	}
	browserHeaders(req)

	resp, err := doWithRetry(ctx, s.client, req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("duckduckgo returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	re := regexp.MustCompile(`vqd=([\d-]+)`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		re2 := regexp.MustCompile(`"vqd"\s*:\s*"([\d-]+)"`)
		matches = re2.FindStringSubmatch(string(body))
	}
	if len(matches) < 2 {
		return "", "", fmt.Errorf("could not find VQD token")
	}

	return strings.TrimSpace(matches[1]), u, nil
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
