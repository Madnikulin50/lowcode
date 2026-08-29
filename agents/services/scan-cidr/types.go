package scancidr

import "time"

type Device struct {
	IP        string   `json:"ip"`
	MAC       string   `json:"mac,omitempty"`
	Hostname  string   `json:"hostname,omitempty"`
	OpenPorts []Port   `json:"openPorts,omitempty"`
	Services  []string `json:"services,omitempty"`
	LastSeen  string   `json:"lastSeen"`
	Status    string   `json:"status"`
}

type Port struct {
	Port    int    `json:"port"`
	Proto   string `json:"proto"`
	Service string `json:"service,omitempty"`
	Version string `json:"version,omitempty"`
	Banner  string `json:"banner,omitempty"`
}

type Config struct {
	Concurrency int
	Timeout     time.Duration
	Ports       []int
}

func DefaultConfig() Config {
	return Config{
		Concurrency: 32,
		Timeout:     400 * time.Millisecond,
		Ports: []int{
			22, 80, 443, 8080, 8443, 3389, 445, 139, 21, 23, 25, 53,
			110, 143, 389, 3306, 5432, 6379, 3333, 5555, 7000,
		},
	}
}
