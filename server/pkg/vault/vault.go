// Package vault resolves secret references against a HashiCorp Vault KV v2
// store, so connector configs (DB/Redis/REST/RabbitMQ credentials, ...) can
// point at a secret instead of storing it in plaintext.
//
// This is intentionally a small, hand-rolled client (a couple of HTTP calls)
// rather than a dependency on github.com/hashicorp/vault/api - the KV v2
// "read" operation used here is trivial, and staying stdlib-only keeps this
// package trivially testable (see vault_test.go) and avoids pulling in the
// full Vault SDK's dependency tree for what amounts to one GET request.
package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Client is a minimal Vault KV v2 reader.
type Client struct {
	Addr  string // e.g. https://vault.example.com:8200
	Token string
	Mount string // KV v2 mount point, default "secret"

	HTTPClient *http.Client

	cacheTTL time.Duration
	mu       sync.Mutex
	cache    map[string]cacheEntry
}

type cacheEntry struct {
	value   string
	expires time.Time
}

// Ref is a parsed secret reference.
type Ref struct {
	Mount string // empty = use the client's default Mount
	Path  string
	Key   string
}

// ParseRef parses a secret reference in the form "[mount:]path#key", e.g.:
//
//	compose/connectors/512#password
//	secret:compose/connectors/512#password
//	vault:secret:compose/connectors/512#password
//
// The "vault:" prefix is optional and purely cosmetic (useful to make a
// config field self-documenting at a glance).
func ParseRef(raw string) (Ref, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "vault:")
	if raw == "" {
		return Ref{}, fmt.Errorf("vault: empty secret reference")
	}

	pathAndMount, key, ok := strings.Cut(raw, "#")
	if !ok || key == "" || pathAndMount == "" {
		return Ref{}, fmt.Errorf("vault: secret reference %q must be in the form [mount:]path#key", raw)
	}

	if mount, path, ok := strings.Cut(pathAndMount, ":"); ok && mount != "" && path != "" {
		return Ref{Mount: mount, Path: path, Key: key}, nil
	}
	return Ref{Path: pathAndMount, Key: key}, nil
}

// FromEnv builds a Client from VAULT_ADDR / VAULT_TOKEN / VAULT_KV_MOUNT /
// VAULT_CACHE_TTL_SECONDS. A Client with an empty Addr is "unconfigured":
// Configured() returns false and Resolve returns a clear error instead of
// making a request.
func FromEnv() *Client {
	ttl := 60 * time.Second
	if raw := strings.TrimSpace(os.Getenv("VAULT_CACHE_TTL_SECONDS")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 0 {
			ttl = time.Duration(n) * time.Second
		}
	}
	mount := strings.TrimSpace(os.Getenv("VAULT_KV_MOUNT"))
	if mount == "" {
		mount = "secret"
	}
	return &Client{
		Addr:     strings.TrimSpace(os.Getenv("VAULT_ADDR")),
		Token:    strings.TrimSpace(os.Getenv("VAULT_TOKEN")),
		Mount:    mount,
		cacheTTL: ttl,
	}
}

// Configured reports whether the client has enough to make requests.
func (c *Client) Configured() bool {
	return c != nil && c.Addr != ""
}

// Resolve fetches the plaintext value for a secret reference, using a short
// TTL cache to avoid round-tripping to Vault on every connector call.
func (c *Client) Resolve(ctx context.Context, refStr string) (string, error) {
	ref, err := ParseRef(refStr)
	if err != nil {
		return "", err
	}

	mount := ref.Mount
	if mount == "" {
		mount = c.Mount
	}
	if mount == "" {
		mount = "secret"
	}
	path := strings.Trim(ref.Path, "/")

	cacheKey := mount + "/" + path + "#" + ref.Key
	if v, ok := c.cacheGet(cacheKey); ok {
		return v, nil
	}

	if !c.Configured() {
		return "", fmt.Errorf("vault: not configured (VAULT_ADDR is not set)")
	}

	url := strings.TrimRight(c.Addr, "/") + "/v1/" + mount + "/data/" + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("vault: build request: %w", err)
	}
	if c.Token != "" {
		req.Header.Set("X-Vault-Token", c.Token)
	}

	hc := c.HTTPClient
	if hc == nil {
		hc = http.DefaultClient
	}
	resp, err := hc.Do(req)
	if err != nil {
		return "", fmt.Errorf("vault: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("vault: secret not found at %s/data/%s", mount, path)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("vault: unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("vault: invalid response: %w", err)
	}

	raw, ok := payload.Data.Data[ref.Key]
	if !ok {
		return "", fmt.Errorf("vault: key %q not found at %s/data/%s", ref.Key, mount, path)
	}

	value := fmt.Sprintf("%v", raw)
	c.cacheSet(cacheKey, value)
	return value, nil
}

func (c *Client) cacheGet(key string) (string, bool) {
	if c == nil || c.cacheTTL <= 0 {
		return "", false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.cache[key]
	if !ok || time.Now().After(e.expires) {
		return "", false
	}
	return e.value, true
}

func (c *Client) cacheSet(key, value string) {
	if c == nil || c.cacheTTL <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		c.cache = make(map[string]cacheEntry)
	}
	c.cache[key] = cacheEntry{value: value, expires: time.Now().Add(c.cacheTTL)}
}

// --- package-level default client, configured from the environment ---

var defaultClient = FromEnv()

// Default returns the process-wide, env-configured client.
func Default() *Client { return defaultClient }

// SetDefault overrides the process-wide default client - most useful in
// tests, to point Resolve/Configured/ResolveField at a fake Vault server
// instead of whatever FromEnv() picked up. Returns a restore func.
func SetDefault(c *Client) (restore func()) {
	prev := defaultClient
	defaultClient = c
	return func() { defaultClient = prev }
}

// Configured reports whether the default client is usable.
func Configured() bool { return defaultClient.Configured() }

// Resolve resolves a secret reference using the default client.
func Resolve(ctx context.Context, ref string) (string, error) {
	return defaultClient.Resolve(ctx, ref)
}

// ResolveField is the common case for a config field that can be either a
// plaintext value or point at Vault: if ref is empty, plaintext is returned
// unchanged (no Vault call at all); otherwise the secret is resolved and
// takes priority over plaintext.
func ResolveField(ctx context.Context, ref, plaintext string) (string, error) {
	if strings.TrimSpace(ref) == "" {
		return plaintext, nil
	}
	return Resolve(ctx, ref)
}
