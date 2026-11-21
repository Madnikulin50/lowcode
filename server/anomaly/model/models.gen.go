package model

// This file is auto-generated version 2.
//
// Changes to this file may cause incorrect behavior and will be lost if
// the code is regenerated from <no value>
//

import (
	"github.com/madnikulin50/lowcode/server/anomaly/types"
	"github.com/madnikulin50/lowcode/server/pkg/dal"
)

var Baseline = &dal.Model{
	Ident:        "anomaly_baselines",
	ResourceType: types.BaselineResourceType,

	Attributes: dal.AttributeSet{
		&dal.Attribute{
			Ident: "ID",
			Type:  &dal.TypeID{},
			Store: &dal.CodecAlias{Ident: "id"},
		},

		&dal.Attribute{
			Ident: "NamespaceID", Sortable: true,
			Type:  &dal.TypeID{},
			Store: &dal.CodecAlias{Ident: "rel_namespace"},
		},

		&dal.Attribute{
			Ident: "ModuleID", Sortable: true,
			Type:  &dal.TypeID{},
			Store: &dal.CodecAlias{Ident: "rel_module"},
		},

		&dal.Attribute{
			Ident: "RuleID", Sortable: true,
			Type: &dal.TypeRef{HasDefault: true,
				DefaultValue: 0,

				RefAttribute: "id",
				RefModel: &dal.ModelRef{
					ResourceType: "corteza::anomaly:rule",
				},
			},
			Store: &dal.CodecAlias{Ident: "rel_rule"},
		},

		&dal.Attribute{
			Ident: "Field", Sortable: true,
			Type:  &dal.TypeText{},
			Store: &dal.CodecAlias{Ident: "field"},
		},

		&dal.Attribute{
			Ident: "Count",
			Type: &dal.TypeNumber{HasDefault: true,
				DefaultValue: 0,
				Precision:    -1, Scale: -1, Meta: map[string]interface{}{"rdbms:type": "integer"},
			},
			Store: &dal.CodecAlias{Ident: "count"},
		},

		&dal.Attribute{
			Ident: "State",
			Type: &dal.TypeJSON{
				DefaultValue: "{}",
			},
			Store: &dal.CodecAlias{Ident: "state"},
		},

		&dal.Attribute{
			Ident: "UpdatedAt", Sortable: true,
			Type:  &dal.TypeTimestamp{Nullable: true, Timezone: true, Precision: -1},
			Store: &dal.CodecAlias{Ident: "updated_at"},
		},
	},

	Indexes: dal.IndexSet{
		&dal.Index{
			Ident: "anomaly_baselines_idxRule",
			Type:  "BTREE",

			Fields: []*dal.IndexField{
				{
					AttributeIdent: "RuleID",
				},
			},
		},

		&dal.Index{
			Ident: "PRIMARY",
			Type:  "BTREE",

			Fields: []*dal.IndexField{
				{
					AttributeIdent: "ID",
				},
			},
		},
	},
}

var Finding = &dal.Model{
	Ident:        "anomaly_findings",
	ResourceType: types.FindingResourceType,

	Attributes: dal.AttributeSet{
		&dal.Attribute{
			Ident: "ID",
			Type:  &dal.TypeID{},
			Store: &dal.CodecAlias{Ident: "id"},
		},

		&dal.Attribute{
			Ident: "NamespaceID", Sortable: true,
			Type:  &dal.TypeID{},
			Store: &dal.CodecAlias{Ident: "rel_namespace"},
		},

		&dal.Attribute{
			Ident: "ModuleID", Sortable: true,
			Type:  &dal.TypeID{},
			Store: &dal.CodecAlias{Ident: "rel_module"},
		},

		&dal.Attribute{
			Ident: "RecordID", Sortable: true,
			Type:  &dal.TypeID{},
			Store: &dal.CodecAlias{Ident: "rel_record"},
		},

		&dal.Attribute{
			Ident: "RuleID", Sortable: true,
			Type: &dal.TypeRef{HasDefault: true,
				DefaultValue: 0,

				RefAttribute: "id",
				RefModel: &dal.ModelRef{
					ResourceType: "corteza::anomaly:rule",
				},
			},
			Store: &dal.CodecAlias{Ident: "rel_rule"},
		},

		&dal.Attribute{
			Ident: "Score", Sortable: true,
			Type:  &dal.TypeNumber{Precision: -1, Scale: -1},
			Store: &dal.CodecAlias{Ident: "score"},
		},

		&dal.Attribute{
			Ident: "Severity", Sortable: true,
			Type:  &dal.TypeText{},
			Store: &dal.CodecAlias{Ident: "severity"},
		},

		&dal.Attribute{
			Ident: "Status", Sortable: true,
			Type:  &dal.TypeText{},
			Store: &dal.CodecAlias{Ident: "status"},
		},

		&dal.Attribute{
			Ident: "Explanation",
			Type: &dal.TypeJSON{
				DefaultValue: "{}",
			},
			Store: &dal.CodecAlias{Ident: "explanation"},
		},

		&dal.Attribute{
			Ident: "CreatedAt", Sortable: true,
			Type: &dal.TypeTimestamp{
				DefaultCurrentTimestamp: true, Timezone: true, Precision: -1,
			},
			Store: &dal.CodecAlias{Ident: "created_at"},
		},

		&dal.Attribute{
			Ident: "UpdatedAt", Sortable: true,
			Type:  &dal.TypeTimestamp{Nullable: true, Timezone: true, Precision: -1},
			Store: &dal.CodecAlias{Ident: "updated_at"},
		},
	},

	Indexes: dal.IndexSet{
		&dal.Index{
			Ident: "anomaly_findings_idxModuleRecord",
			Type:  "BTREE",

			Fields: []*dal.IndexField{
				{
					AttributeIdent: "ModuleID",
				},

				{
					AttributeIdent: "RecordID",
				},
			},
		},

		&dal.Index{
			Ident: "PRIMARY",
			Type:  "BTREE",

			Fields: []*dal.IndexField{
				{
					AttributeIdent: "ID",
				},
			},
		},
	},
}

var Rule = &dal.Model{
	Ident:        "anomaly_rules",
	ResourceType: types.RuleResourceType,

	Attributes: dal.AttributeSet{
		&dal.Attribute{
			Ident: "ID",
			Type:  &dal.TypeID{},
			Store: &dal.CodecAlias{Ident: "id"},
		},

		&dal.Attribute{
			Ident: "NamespaceID", Sortable: true,
			Type:  &dal.TypeID{},
			Store: &dal.CodecAlias{Ident: "rel_namespace"},
		},

		&dal.Attribute{
			Ident: "ModuleID", Sortable: true,
			Type:  &dal.TypeID{},
			Store: &dal.CodecAlias{Ident: "rel_module"},
		},

		&dal.Attribute{
			Ident: "Field", Sortable: true,
			Type:  &dal.TypeText{},
			Store: &dal.CodecAlias{Ident: "field"},
		},

		&dal.Attribute{
			Ident: "Detector", Sortable: true,
			Type:  &dal.TypeText{},
			Store: &dal.CodecAlias{Ident: "detector"},
		},

		&dal.Attribute{
			Ident: "Threshold",
			Type:  &dal.TypeNumber{Precision: -1, Scale: -1},
			Store: &dal.CodecAlias{Ident: "threshold"},
		},

		&dal.Attribute{
			Ident: "Enabled", Sortable: true,
			Type: &dal.TypeBoolean{HasDefault: true,
				DefaultValue: false,
			},
			Store: &dal.CodecAlias{Ident: "enabled"},
		},

		&dal.Attribute{
			Ident: "Params",
			Type: &dal.TypeJSON{
				DefaultValue: "{}",
			},
			Store: &dal.CodecAlias{Ident: "params"},
		},

		&dal.Attribute{
			Ident: "LastScannedAt",
			Type:  &dal.TypeTimestamp{Nullable: true, Timezone: true, Precision: -1},
			Store: &dal.CodecAlias{Ident: "last_scanned_at"},
		},

		&dal.Attribute{
			Ident: "CreatedAt", Sortable: true,
			Type: &dal.TypeTimestamp{
				DefaultCurrentTimestamp: true, Timezone: true, Precision: -1,
			},
			Store: &dal.CodecAlias{Ident: "created_at"},
		},

		&dal.Attribute{
			Ident: "CreatedBy",
			Type: &dal.TypeRef{HasDefault: true,
				DefaultValue: 0,

				RefAttribute: "id",
				RefModel: &dal.ModelRef{
					ResourceType: "corteza::system:user",
				},
			},
			Store: &dal.CodecAlias{Ident: "created_by"},
		},

		&dal.Attribute{
			Ident: "UpdatedAt", Sortable: true,
			Type:  &dal.TypeTimestamp{Nullable: true, Timezone: true, Precision: -1},
			Store: &dal.CodecAlias{Ident: "updated_at"},
		},

		&dal.Attribute{
			Ident: "UpdatedBy",
			Type: &dal.TypeRef{HasDefault: true,
				DefaultValue: 0,

				RefAttribute: "id",
				RefModel: &dal.ModelRef{
					ResourceType: "corteza::system:user",
				},
			},
			Store: &dal.CodecAlias{Ident: "updated_by"},
		},

		&dal.Attribute{
			Ident: "DeletedAt", Sortable: true,
			Type:  &dal.TypeTimestamp{Nullable: true, Timezone: true, Precision: -1},
			Store: &dal.CodecAlias{Ident: "deleted_at"},
		},

		&dal.Attribute{
			Ident: "DeletedBy",
			Type: &dal.TypeRef{HasDefault: true,
				DefaultValue: 0,

				RefAttribute: "id",
				RefModel: &dal.ModelRef{
					ResourceType: "corteza::system:user",
				},
			},
			Store: &dal.CodecAlias{Ident: "deleted_by"},
		},
	},

	Indexes: dal.IndexSet{
		&dal.Index{
			Ident: "anomaly_rules_idxModuleField",
			Type:  "BTREE",

			Fields: []*dal.IndexField{
				{
					AttributeIdent: "ModuleID",
				},

				{
					AttributeIdent: "Field",
				},
			},
		},

		&dal.Index{
			Ident: "PRIMARY",
			Type:  "BTREE",

			Fields: []*dal.IndexField{
				{
					AttributeIdent: "ID",
				},
			},
		},
	},
}

func init() {
	models = append(
		models,
		Baseline,
		Finding,
		Rule,
	)
}
