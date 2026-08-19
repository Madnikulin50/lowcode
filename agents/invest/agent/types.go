package agent

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ListenAddr  string
	CortezaAPI  string
	Token       string
	NamespaceID uint64
	HTTPTimeout time.Duration
}

type flexUint uint64

func (id *flexUint) UnmarshalJSON(b []byte) error {
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
		*id = flexUint(n)
		return nil
	}
	var n uint64
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	*id = flexUint(n)
	return nil
}

type JobRequest struct {
	NamespaceID  flexUint `json:"namespaceID"`
	ProjectID    string   `json:"projectID"`
	Token        string   `json:"token"`
	CPIThreshold float64  `json:"cpiThreshold"`
}

func (r *JobRequest) Normalize() {
	switch r.ProjectID {
	case "0", "auto", "{{projectID}}", "${recordID}":
		r.ProjectID = ""
	}
}

type WBSItem struct {
	ID              uint64
	ProjectID       string
	Code            string
	Name            string
	ParentID        string
	Predecessors    []string
	StartPlanned    time.Time
	EndPlanned      time.Time
	BudgetPlanned   float64
	ActualCost      float64
	PercentComplete float64
	PV              float64
	EV              float64
	SPI             float64
	CPI             float64
	EAC             float64
	IsCritical      bool
	TotalFloat      float64
}

type ProgressFact struct {
	WBSID   string
	Percent float64
	Cost    float64
}

type Document struct {
	ID      uint64
	Project string
	Title   string
	Status  string
	DueDate time.Time
}

type EVMResult struct {
	PV  float64
	EV  float64
	AC  float64
	SPI float64
	CPI float64
	EAC float64
	BAC float64
}

type Alert struct {
	Kind    string `json:"kind"`
	Title   string `json:"title"`
	Project string `json:"project,omitempty"`
	WBS     string `json:"wbs,omitempty"`
	Detail  string `json:"detail"`
}

type Activity struct {
	ID           string
	DurationDays float64
	Predecessors []string
	ES, EF       float64
	LS, LF       float64
	Float        float64
	Critical     bool
}
