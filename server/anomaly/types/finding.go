package types

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/madnikulin50/lowcode/server/pkg/sql"
)

var (
	SeverityLow    = "low"
	SeverityMedium = "medium"
	SeverityHigh   = "high"

	StatusNew           = "new"
	StatusAcknowledged  = "acknowledged"
	StatusResolved      = "resolved"
	StatusFalsePositive = "false_positive"
)

type (
	// Finding is one scanner-detected anomaly on a single record, for a
	// single Rule. Re-flagging the same record+rule updates the existing
	// Finding (see anomaly/service/scanner.go) instead of piling up
	// duplicates.
	Finding struct {
		ID uint64 `json:"findingID,string" schema:"col=id,dal=id,unique"`

		NamespaceID uint64 `json:"namespaceID,string" schema:"col=rel_namespace,dal=id,sortable"`
		ModuleID    uint64 `json:"moduleID,string" schema:"col=rel_module,dal=id,sortable"`
		RecordID    uint64 `json:"recordID,string" schema:"col=rel_record,dal=id,sortable"`
		RuleID      uint64 `json:"ruleID,string" schema:"col=rel_rule,dal=ref:corteza::anomaly:rule:default0,sortable"`

		Score    float64 `json:"score" schema:"col=score,dal=number:float,sortable"`
		Severity string  `json:"severity" schema:"col=severity,dal,sortable"`
		Status   string  `json:"status" schema:"col=status,dal,sortable"`

		// Human-readable breakdown of why this record was flagged (field
		// value, baseline mean/stddev, detector) - rendered as-is by the
		// record-list badge tooltip and the Anomaly Center drill-down, so
		// the UI needs no per-detector rendering logic of its own.
		Explanation FindingExplanation `json:"explanation" schema:"col=explanation,dal=json:empty,omit"`

		CreatedAt time.Time  `json:"createdAt,omitempty" schema:"col=created_at,dal=timestamp:now,sortable"`
		UpdatedAt *time.Time `json:"updatedAt,omitempty" schema:"col=updated_at,dal=timestamp:nil,sortable"`
	}

	FindingFilter struct {
		NamespaceID uint64 `json:"namespaceID,string"`
		ModuleID    uint64 `json:"moduleID,string"`
		RecordID    uint64 `json:"recordID,string"`
		Status      string `json:"status"`
		Severity    string `json:"severity"`

		Check func(*Finding) (bool, error) `json:"-"`

		filter.Sorting
		filter.Paging
	}

	// FindingExplanation is a plain JSON-object map, but needs its own named
	// type (rather than a bare map[string]any) so it can implement
	// sql.Scanner/driver.Valuer - every other JSON-column type in this
	// codebase does the same (see e.g. compose/types/namespace.go's
	// NamespaceMeta.Scan/Value); a raw map[string]any has neither and the
	// rdbms store can't persist it as-is.
	FindingExplanation map[string]any
)

func (e *FindingExplanation) Scan(src any) error          { return sql.ParseJSON(src, e) }
func (e FindingExplanation) Value() (driver.Value, error) { return json.Marshal(e) }
