package types

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/madnikulin50/lowcode/server/pkg/sql"
)

type (
	// Baseline is one Detector's accumulated state for a single Rule.
	// Scoped per rule (not just per module+field): different rules on the
	// same field can use different detectors, and their state shapes are
	// mutually incompatible (mean/variance vs. a value window vs. a
	// frequency map - see anomaly/service/detector.go), so sharing one row
	// per field would let one rule corrupt another's baseline.
	Baseline struct {
		ID uint64 `json:"baselineID,string" schema:"col=id,dal=id,unique"`

		NamespaceID uint64 `json:"namespaceID,string" schema:"col=rel_namespace,dal=id,sortable"`
		ModuleID    uint64 `json:"moduleID,string" schema:"col=rel_module,dal=id,sortable"`
		RuleID      uint64 `json:"ruleID,string" schema:"col=rel_rule,dal=ref:corteza::anomaly:rule:default0,sortable"`
		Field       string `json:"field" schema:"col=field,dal,sortable"`

		// Count is how many observations have been folded into State so
		// far - generic across every detector, used to gate detection
		// until a baseline has "warmed up" (see minBaselineCount /
		// minCategoricalCount in anomaly/service/detector.go).
		Count uint64 `json:"count" schema:"col=count,dal=number:default0"`

		// State is opaque to everything except the Detector that owns it:
		// mean/variance for zscore/EWMA, mean/variance plus cumulative sums
		// for CUSUM, a bounded value window for MAD, a value->count
		// frequency map for the categorical detectors. See BaselineState.
		State BaselineState `json:"state" schema:"col=state,dal=json:empty,omit"`

		UpdatedAt *time.Time `json:"updatedAt,omitempty" schema:"col=updated_at,dal=timestamp:nil,sortable"`
	}

	BaselineFilter struct {
		NamespaceID uint64 `json:"namespaceID,string"`
		ModuleID    uint64 `json:"moduleID,string"`
		RuleID      uint64 `json:"ruleID,string"`
		Field       string `json:"field"`

		Check func(*Baseline) (bool, error) `json:"-"`

		filter.Sorting
		filter.Paging
	}

	// BaselineState is a plain JSON-object map - needs its own named type
	// (rather than a bare map[string]any) to implement sql.Scanner/
	// driver.Valuer, same reasoning as Finding.Explanation in finding.go:
	// a bare map[string]any can't be persisted by the rdbms store as-is.
	BaselineState map[string]any
)

func (s *BaselineState) Scan(src any) error          { return sql.ParseJSON(src, s) }
func (s BaselineState) Value() (driver.Value, error) { return json.Marshal(s) }
