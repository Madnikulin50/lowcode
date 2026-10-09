package model

import (
	"github.com/madnikulin50/lowcode/server/pkg/dal"
)

func init() {
	models = append(models, &dal.Model{
		Ident: "compose_rule_chain_run",
		Attributes: dal.AttributeSet{
			&dal.Attribute{
				Ident: "ID",
				Type:  &dal.TypeID{},
				Store: &dal.CodecAlias{Ident: "id"},
			},
			&dal.Attribute{
				Ident: "NamespaceID",
				Type:  &dal.TypeID{},
				Store: &dal.CodecAlias{Ident: "rel_namespace"},
			},
			&dal.Attribute{
				Ident: "ChainID", Sortable: true,
				Type:  &dal.TypeText{Length: 191},
				Store: &dal.CodecAlias{Ident: "chain_id"},
			},
			&dal.Attribute{
				Ident: "TriggerType",
				Type:  &dal.TypeText{Length: 32},
				Store: &dal.CodecAlias{Ident: "trigger_type"},
			},
			&dal.Attribute{
				Ident: "Success",
				Type:  &dal.TypeBoolean{},
				Store: &dal.CodecAlias{Ident: "success"},
			},
			&dal.Attribute{
				Ident: "Error",
				Type:  &dal.TypeText{Nullable: true},
				Store: &dal.CodecAlias{Ident: "error"},
			},
			&dal.Attribute{
				Ident: "Input",
				Type:  &dal.TypeJSON{DefaultValue: "{}"},
				Store: &dal.CodecAlias{Ident: "input"},
			},
			&dal.Attribute{
				Ident: "Output",
				Type:  &dal.TypeJSON{DefaultValue: "{}"},
				Store: &dal.CodecAlias{Ident: "output"},
			},
			&dal.Attribute{
				Ident: "Nodes",
				Type:  &dal.TypeJSON{DefaultValue: "[]"},
				Store: &dal.CodecAlias{Ident: "nodes"},
			},
			&dal.Attribute{
				Ident: "StartedAt", Sortable: true,
				Type:  &dal.TypeTimestamp{Timezone: true, Precision: -1},
				Store: &dal.CodecAlias{Ident: "started_at"},
			},
			&dal.Attribute{
				Ident: "FinishedAt", Sortable: true,
				Type:  &dal.TypeTimestamp{Nullable: true, Timezone: true, Precision: -1},
				Store: &dal.CodecAlias{Ident: "finished_at"},
			},
			&dal.Attribute{
				Ident: "DurationMs",
				Type:  &dal.TypeNumber{},
				Store: &dal.CodecAlias{Ident: "duration_ms"},
			},
			&dal.Attribute{
				Ident: "CreatedAt", Sortable: true,
				Type: &dal.TypeTimestamp{
					DefaultCurrentTimestamp: true, Timezone: true, Precision: -1,
				},
				Store: &dal.CodecAlias{Ident: "created_at"},
			},
		},
		Indexes: dal.IndexSet{
			&dal.Index{
				Ident: "PRIMARY",
				Type:  "BTREE",
				Fields: []*dal.IndexField{
					{AttributeIdent: "ID"},
				},
			},
			&dal.Index{
				Ident: "compose_rule_chain_run_idxChain",
				Type:  "BTREE",
				Fields: []*dal.IndexField{
					{AttributeIdent: "ChainID"},
					{AttributeIdent: "StartedAt"},
				},
			},
			&dal.Index{
				Ident: "compose_rule_chain_run_idxNamespace",
				Type:  "BTREE",
				Fields: []*dal.IndexField{
					{AttributeIdent: "NamespaceID"},
					{AttributeIdent: "StartedAt"},
				},
			},
		},
	})
}
