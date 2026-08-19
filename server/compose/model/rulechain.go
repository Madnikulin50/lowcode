package model

import (
	"github.com/madnikulin50/lowcode/server/pkg/dal"
)

func init() {
	models = append(models, &dal.Model{
		Ident: "compose_rule_chain",
		Attributes: dal.AttributeSet{
			&dal.Attribute{
				Ident: "ID",
				Type:  &dal.TypeID{},
				Store: &dal.CodecAlias{Ident: "id"},
			},
			&dal.Attribute{
				Ident: "Handle",
				Type:  &dal.TypeText{Length: 128},
				Store: &dal.CodecAlias{Ident: "handle"},
			},
			&dal.Attribute{
				Ident: "NamespaceID",
				Type:  &dal.TypeID{},
				Store: &dal.CodecAlias{Ident: "rel_namespace"},
			},
			&dal.Attribute{
				Ident: "Name", Sortable: true,
				Type:  &dal.TypeText{},
				Store: &dal.CodecAlias{Ident: "name"},
			},
			&dal.Attribute{
				Ident: "Description",
				Type:  &dal.TypeText{},
				Store: &dal.CodecAlias{Ident: "description"},
			},
			&dal.Attribute{
				Ident: "EntryNode",
				Type:  &dal.TypeText{Length: 128},
				Store: &dal.CodecAlias{Ident: "entry_node"},
			},
			&dal.Attribute{
				Ident: "Nodes",
				Type:  &dal.TypeJSON{DefaultValue: "[]"},
				Store: &dal.CodecAlias{Ident: "nodes"},
			},
			&dal.Attribute{
				Ident: "Edges",
				Type:  &dal.TypeJSON{DefaultValue: "[]"},
				Store: &dal.CodecAlias{Ident: "edges"},
			},
			&dal.Attribute{
				Ident: "Config",
				Type:  &dal.TypeJSON{DefaultValue: "{}"},
				Store: &dal.CodecAlias{Ident: "config"},
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
			&dal.Attribute{
				Ident: "DeletedAt", Sortable: true,
				Type:  &dal.TypeTimestamp{Nullable: true, Timezone: true, Precision: -1},
				Store: &dal.CodecAlias{Ident: "deleted_at"},
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
				Ident:     "compose_rule_chain_uniqueHandle",
				Type:      "BTREE",
				Unique:    true,
				Predicate: "handle != '' AND deleted_at IS NULL",
				Fields: []*dal.IndexField{
					{AttributeIdent: "Handle", Modifiers: []dal.IndexFieldModifier{"LOWERCASE"}},
				},
			},
		},
	})
}
