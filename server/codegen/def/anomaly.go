package def

import (
	anomalytypes "github.com/madnikulin50/lowcode/server/anomaly/types"
)

// Anomaly is a new (not ported from any .cue source - .cue is being removed,
// see codegen/def/cmd/gengo/main.go's doc comment) component for the
// anomaly-detection engine: which Compose module fields are being watched
// (Rule), each watched field's running mean/variance (Baseline), and the
// anomalies the scanner has flagged (Finding). It follows federation.go's
// AttributesFromStruct pilot from the start, rather than a parallel
// Attribute{} literal list.
var Anomaly = Component{
	Handle: "anomaly",

	Resources: []NamedResource{
		{Handle: "rule", Resource: Resource{
			Model: Model{
				Ident:      "anomaly_rules",
				Attributes: AttributesFromStruct(anomalytypes.Rule{}),
				Indexes: map[string]Index{
					"primary":          {Attribute: "id"},
					"idx_module_field": {Attributes: []string{"rel_module", "field"}},
				},
			},
			Features: Features{Labels: boolPtr(false)},
			Filter: Filter{
				Struct: map[string]Attribute{
					"rel_namespace": {},
					"rel_module":    {},
					"field":         {},
					"enabled":       {},
					"deleted":       {GoType: "filter.State", StoreIdent: "deleted_at"},
				},
				Query:      []string{"field"},
				ByValue:    []string{"rel_namespace", "rel_module", "enabled"},
				ByNilState: []string{"deleted"},
			},
			Store: &StoreConfig{
				Ident: "anomalyRule",
				Lookups: []StoreLookup{
					{Fields: []string{"id"}, Description: "searches for anomaly rule by ID\n\nIt returns anomaly rule"},
					{Fields: []string{"rel_module", "field"}, Description: "searches for anomaly rule by module and field\n\nIt returns anomaly rule"},
				},
			},
		}},

		{Handle: "baseline", Resource: Resource{
			Model: Model{
				Ident:      "anomaly_baselines",
				Attributes: AttributesFromStruct(anomalytypes.Baseline{}),
				Indexes: map[string]Index{
					"primary":  {Attribute: "id"},
					"idx_rule": {Attribute: "rel_rule"},
				},
			},
			Features: Features{Labels: boolPtr(false)},
			Filter: Filter{
				Struct: map[string]Attribute{
					"rel_namespace": {},
					"rel_module":    {},
					"rel_rule":      {},
					"field":         {},
				},
				ByValue: []string{"rel_namespace", "rel_module", "rel_rule", "field"},
			},
			// No Rbac block: baselines are internal scanner state, never
			// exposed directly - gated by the owning module's permissions
			// wherever a Finding referencing them is read.
			Store: &StoreConfig{
				Ident: "anomalyBaseline",
				Lookups: []StoreLookup{
					{Fields: []string{"rel_rule"}, Description: "searches for anomaly baseline by rule - baselines are scoped per rule since different rules on the same field can use incompatible detector state shapes\n\nIt returns anomaly baseline"},
				},
			},
		}},

		{Handle: "finding", Resource: Resource{
			Model: Model{
				Ident:      "anomaly_findings",
				Attributes: AttributesFromStruct(anomalytypes.Finding{}),
				Indexes: map[string]Index{
					"primary":           {Attribute: "id"},
					"idx_module_record": {Attributes: []string{"rel_module", "rel_record"}},
				},
			},
			Features: Features{Labels: boolPtr(false)},
			Filter: Filter{
				Struct: map[string]Attribute{
					"rel_namespace": {},
					"rel_module":    {},
					"rel_record":    {},
					"status":        {},
					"severity":      {},
				},
				ByValue: []string{"rel_namespace", "rel_module", "rel_record", "status", "severity"},
			},
			// No Rbac block: access is gated by the referenced Compose
			// module/namespace's own CanReadRecord/CanSearchRecordsOnModule
			// (see anomaly/rest/finding.go), not a new set of operations.
			Store: &StoreConfig{
				Ident: "anomalyFinding",
				Lookups: []StoreLookup{
					{Fields: []string{"id"}, Description: "searches for anomaly finding by ID\n\nIt returns anomaly finding"},
					{Fields: []string{"rel_rule", "rel_record"}, Description: "searches for an existing finding by rule and record, to update it instead of creating a duplicate\n\nIt returns anomaly finding"},
				},
			},
		}},
	},
}
