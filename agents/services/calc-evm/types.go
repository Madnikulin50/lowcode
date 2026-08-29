package calcevm

import "time"

// Item is one work package. IDs are strings so HTTP/JSON and Compose records match.
type Item struct {
	ID              string    `json:"id"`
	ProjectID       string    `json:"projectID,omitempty"`
	BudgetPlanned   float64   `json:"budgetPlanned"`
	PercentComplete float64   `json:"percentComplete"`
	ActualCost      float64   `json:"actualCost"`
	StartPlanned    time.Time `json:"startPlanned,omitempty"`
	EndPlanned      time.Time `json:"endPlanned,omitempty"`
	PV              float64   `json:"pv"`
	EV              float64   `json:"ev"`
	SPI             float64   `json:"spi"`
	CPI             float64   `json:"cpi"`
	EAC             float64   `json:"eac"`
}

type Fact struct {
	WBSID     string  `json:"wbsID"`
	ProjectID string  `json:"projectID,omitempty"`
	Percent   float64 `json:"percent"`
	Cost      float64 `json:"cost"`
}

type Result struct {
	PV  float64 `json:"pv"`
	EV  float64 `json:"ev"`
	AC  float64 `json:"ac"`
	SPI float64 `json:"spi"`
	CPI float64 `json:"cpi"`
	EAC float64 `json:"eac"`
	BAC float64 `json:"bac"`
}

type Input struct {
	Now       time.Time `json:"now,omitempty"`
	ProjectID string    `json:"projectID,omitempty"`
	Items     []Item    `json:"items"`
	Facts     []Fact    `json:"facts,omitempty"`
}

type Output struct {
	WBS       int             `json:"wbs"`
	Updated   int             `json:"updated"`
	ProjectID string          `json:"projectID,omitempty"`
	Project   Result          `json:"project"`
	Projects  []ProjectRollup `json:"projects,omitempty"`
	SPI       float64         `json:"spi"`
	CPI       float64         `json:"cpi"`
	EAC       float64         `json:"eac"`
	AC        float64         `json:"ac"`
	Items     []Item          `json:"items"`
}

type ProjectRollup struct {
	ProjectID string `json:"projectID"`
	Result
}
