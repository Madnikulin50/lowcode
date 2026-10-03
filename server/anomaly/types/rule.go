package types

import (
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/filter"
)

var (
	DetectorZScore       = "zscore"
	DetectorEWMA         = "ewma"
	DetectorCUSUM        = "cusum"
	DetectorRange        = "range"
	DetectorMAD          = "mad"
	DetectorRareCategory = "rare_category"
	DetectorNewCategory  = "new_category"
)

type (
	// Rule declares that a given Compose module field is under anomaly
	// scanning: which detector to run, at what threshold, and (via
	// LastScannedAt) how far the scanner has already gotten.
	//
	// Its `schema` tags follow the AttributesFromStruct pilot introduced by
	// federation/types/node.go - see codegen/def/anomaly.go.
	Rule struct {
		ID uint64 `json:"ruleID,string" schema:"col=id,dal=id,unique"`

		NamespaceID uint64 `json:"namespaceID,string" schema:"col=rel_namespace,dal=id,sortable"`
		ModuleID    uint64 `json:"moduleID,string" schema:"col=rel_module,dal=id,sortable"`
		Field       string `json:"field" schema:"col=field,dal,sortable"`

		Detector  string  `json:"detector" schema:"col=detector,dal,sortable"`
		Threshold float64 `json:"threshold" schema:"col=threshold,dal=number:float"`
		Enabled   bool    `json:"enabled" schema:"col=enabled,dal=bool:false,sortable"`

		// Params holds detector-specific config that doesn't fit a single
		// Threshold float (currently just {min, max} for the range
		// detector). Reuses FindingExplanation's Scan/Value plumbing rather
		// than declaring its own identical JSON map type.
		Params FindingExplanation `json:"params" schema:"col=params,dal=json:empty,omit"`

		// Watermark: records updated after this timestamp haven't been
		// scanned yet by this rule's last run (see anomaly/service/scanner.go).
		LastScannedAt *time.Time `json:"lastScannedAt,omitempty" schema:"col=last_scanned_at,dal=timestamp:nil"`

		CreatedAt time.Time  `json:"createdAt,omitempty" schema:"col=created_at,dal=timestamp:now,sortable"`
		CreatedBy uint64     `json:"createdBy,string" schema:"col=created_by,dal=userref"`
		UpdatedAt *time.Time `json:"updatedAt,omitempty" schema:"col=updated_at,dal=timestamp:nil,sortable"`
		UpdatedBy uint64     `json:"updatedBy,string,omitempty" schema:"col=updated_by,dal=userref"`
		DeletedAt *time.Time `json:"deletedAt,omitempty" schema:"col=deleted_at,dal=timestamp:nil,sortable"`
		DeletedBy uint64     `json:"deletedBy,string,omitempty" schema:"col=deleted_by,dal=userref"`
	}

	RuleFilter struct {
		Query       string `json:"query"`
		NamespaceID uint64 `json:"namespaceID,string"`
		ModuleID    uint64 `json:"moduleID,string"`
		Enabled     bool   `json:"enabled"`

		// Check fn is called by store backend for each resource found; the
		// function can modify the resource and return false if store should
		// not return it.
		Check func(*Rule) (bool, error) `json:"-"`

		Deleted filter.State `json:"deleted"`

		filter.Sorting
		filter.Paging
	}
)
