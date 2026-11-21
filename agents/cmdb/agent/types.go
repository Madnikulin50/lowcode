package agent

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

type FlexUint64 uint64

func (id *FlexUint64) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "" || s == "null" || s == `""` {
		*id = 0
		return nil
	}
	if s[0] == '"' {
		var str string
		if err := json.Unmarshal(b, &str); err != nil {
			return err
		}
		str = strings.TrimSpace(str)
		if str == "" {
			*id = 0
			return nil
		}
		n, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return err
		}
		*id = FlexUint64(n)
		return nil
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	*id = FlexUint64(n)
	return nil
}

type ScanTarget struct {
	ID           string     `json:"id"`
	CIDR         string     `json:"cidr"`
	NamespaceID  FlexUint64 `json:"namespaceID"`
	ModuleID     FlexUint64 `json:"moduleID,omitempty"`
	Token        string     `json:"token,omitempty"`
	API          string     `json:"api,omitempty"`
	ScanRecordID string     `json:"scanRecordID,omitempty"`
	CallbackURL  string     `json:"callbackUrl,omitempty"`
}

type ScanStatus struct {
	ID         string     `json:"id"`
	Target     string     `json:"target"`
	Progress   float64    `json:"progress"`
	Status     string     `json:"status"`
	StartedAt  time.Time  `json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
	Found      int        `json:"found"`
	ModuleID   uint64     `json:"moduleID,omitempty"`
	Error      string     `json:"error,omitempty"`
	Message    string     `json:"message,omitempty"`
	ScanningIP string     `json:"scanningIP,omitempty"`
	TotalIPs   int        `json:"totalIPs,omitempty"`
	ScannedIPs int        `json:"scannedIPs,omitempty"`
	CIDRs      []string   `json:"cidrs,omitempty"`
	Items      []Device   `json:"items,omitempty"`
}

type Device struct {
	IP              string          `json:"ip"`
	MAC             string          `json:"mac,omitempty"`
	Hostname        string          `json:"hostname,omitempty"`
	Vendor          string          `json:"vendor,omitempty"`
	Model           string          `json:"model,omitempty"`
	DeviceType      string          `json:"deviceType,omitempty"`
	OS              string          `json:"os,omitempty"`
	Domain          string          `json:"domain,omitempty"`
	OpenPorts       []Port          `json:"openPorts,omitempty"`
	Services        []string        `json:"services,omitempty"`
	Shares          []string        `json:"shares,omitempty"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities,omitempty"`
	LastSeen        string          `json:"lastSeen"`
	Status          string          `json:"status"`
	RecordID        uint64          `json:"recordID,omitempty"`
}

type Port struct {
	Port    int    `json:"port"`
	Proto   string `json:"proto"`
	Service string `json:"service,omitempty"`
	Version string `json:"version,omitempty"`
	Banner  string `json:"banner,omitempty"`
}
