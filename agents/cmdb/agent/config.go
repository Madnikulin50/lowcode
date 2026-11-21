package agent

import "time"

type Config struct {
	ListenAddr     string
	CortezaAPI     string
	CortezaAuth    string
	LLMModel       string
	LLMBaseURL     string
	Concurrency    int
	PingTimeout    time.Duration
	ScanPorts      []int
	Token          string
	NamespaceID    uint64
	ScanInterval   time.Duration
	StatusInterval time.Duration
	AutoCIDRs      []string
}
