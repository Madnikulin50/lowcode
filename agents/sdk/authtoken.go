package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// SelfToken returns a Lowcode API access token for this agent's own use.
//
// Precedence:
//  1. TOKEN env var, if set (a token minted elsewhere — the old manual
//     workflow still works exactly as before).
//  2. Self-enrollment: POST {cortezaAPI}/agents/enroll with the
//     AGENT_SHARED_SECRET env var. The server hands back a working token
//     for its built-in service identity if the secret matches — no live
//     user session, no external provisioning script, no OAuth2 client to
//     register by hand. Works from a cold `docker-compose up` as long as
//     AGENT_SHARED_SECRET is the same string on the server and the agent.
//
// Both are entirely optional: with neither TOKEN nor AGENT_SHARED_SECRET
// set, this returns "" and the agent runs unauthenticated, exactly like
// before this existed. Callers already treat an empty token that way.
func SelfToken(cortezaAPI string) string {
	if t := strings.TrimSpace(os.Getenv("TOKEN")); t != "" {
		return t
	}

	secret := strings.TrimSpace(os.Getenv("AGENT_SHARED_SECRET"))
	if secret == "" {
		return ""
	}

	endpoint := strings.TrimRight(cortezaAPI, "/") + "/agents/enroll"

	// The server may still be starting up (docker-compose has no strict
	// health-gated ordering here), so retry connection/5xx failures for a
	// while instead of failing the agent's own startup on the first
	// connection refused. A rejected secret won't fix itself on retry, so
	// that fails fast instead.
	deadline := time.Now().Add(60 * time.Second)
	var lastErr error
	for {
		token, retry, err := fetchEnrollToken(endpoint, secret)
		if err == nil && token != "" {
			return token
		}
		lastErr = err
		if !retry || !time.Now().Before(deadline) {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if lastErr != nil {
		fmt.Fprintf(os.Stderr, "sdk: self-enroll via AGENT_SHARED_SECRET failed: %v\n", lastErr)
	}
	return ""
}

// fetchEnrollToken returns (token, retry, err). retry is true when the
// failure looks transient (connection refused, 5xx — server not ready
// yet) and worth retrying; false for a definitive rejection (bad secret),
// which retrying won't change.
func fetchEnrollToken(endpoint, secret string) (string, bool, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, nil)
	if err != nil {
		return "", false, err
	}
	req.Header.Set("X-Agent-Secret", secret)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", true, err
	}
	defer res.Body.Close()

	switch {
	case res.StatusCode == http.StatusUnauthorized:
		return "", false, fmt.Errorf("server rejected AGENT_SHARED_SECRET (401)")
	case res.StatusCode >= http.StatusInternalServerError:
		return "", true, fmt.Errorf("enroll endpoint returned %d (server may still be starting)", res.StatusCode)
	case res.StatusCode != http.StatusOK:
		return "", false, fmt.Errorf("enroll endpoint returned %d", res.StatusCode)
	}

	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", false, fmt.Errorf("decode enroll response: %w", err)
	}
	if out.AccessToken == "" {
		return "", false, fmt.Errorf("enroll response missing access_token")
	}
	return out.AccessToken, false, nil
}
