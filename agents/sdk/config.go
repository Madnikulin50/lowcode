package sdk

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	Handle      string
	Name        string
	Version     string
	Listen      string
	PublicURL   string
	Slug        string
	CortezaAPI  string
	Token       string
	NamespaceID uint64
	Heartbeat   HeartbeatConfig
}

type HeartbeatConfig struct {
	Module   string
	Interval time.Duration
}

type Identity struct {
	Name         string
	PublicURL    string
	Hostname     string
	Capabilities []string
}

func Hostname() string {
	h, err := os.Hostname()
	if err != nil || strings.TrimSpace(h) == "" {
		return "agent"
	}
	return h
}

func (c Config) Identity(caps []string) Identity {
	name := c.Name
	if name == "" {
		name = c.Handle
	}
	return Identity{
		Name:         name,
		PublicURL:    strings.TrimRight(c.PublicURL, "/"),
		Hostname:     Hostname(),
		Capabilities: caps,
	}
}
