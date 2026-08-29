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
	RecordID     string   `json:"recordID"`
	Token        string   `json:"token"`
	CPIThreshold float64  `json:"cpiThreshold"`
	Decision     string   `json:"decision"`
	UserID       string   `json:"userID"`
	Comment      string   `json:"comment"`
}

func (r *JobRequest) Normalize() {
	r.ProjectID = cleanID(r.ProjectID)
	r.RecordID = cleanID(r.RecordID)
	r.UserID = cleanID(r.UserID)
}

func cleanID(s string) string {
	s = strings.TrimSpace(s)
	switch s {
	case "0", "auto", "{{projectID}}", "{{recordID}}", "{{userID}}", "${recordID}", "${userID}", "${project}":
		return ""
	}
	if strings.Contains(s, "{{") || strings.Contains(s, "${") {
		return ""
	}
	return s
}

func ParseID(s string) uint64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" {
		return 0
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return n
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
	ID       uint64
	Project  string
	Title    string
	Status   string
	DueDate  time.Time
	Assignee string
	Author   string
	Contract string
	File     string
	Number   string
}

type RFC struct {
	ID          uint64
	Project     string
	Title       string
	Status      string
	DeltaBudget float64
	DeltaDays   float64
	EACBefore   float64
	EACAfter    float64
	Simulated   bool
	EndAfter    time.Time
	Author      string
}

type Project struct {
	ID               uint64
	Name             string
	BudgetPlanned    float64
	BudgetActual     float64
	EAC              float64
	EndPlanned       time.Time
	ConstructionType string
}

type Approval struct {
	ID       uint64
	Document string
	Approver string
	Decision string
	Step     float64
	Role     string
}

type Member struct {
	User string
	Role string
}

type BudgetLine struct {
	Project string
	Article string
	Reserve float64
	Actual  float64
	Planned float64
}

type EVMResult struct {
	PV  float64 `json:"pv"`
	EV  float64 `json:"ev"`
	AC  float64 `json:"ac"`
	SPI float64 `json:"spi"`
	CPI float64 `json:"cpi"`
	EAC float64 `json:"eac"`
	BAC float64 `json:"bac"`
}

type Alert struct {
	Kind    string `json:"kind"`
	Title   string `json:"title"`
	Project string `json:"project,omitempty"`
	WBS     string `json:"wbs,omitempty"`
	Detail  string `json:"detail"`
	Email   string `json:"email,omitempty"`
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
