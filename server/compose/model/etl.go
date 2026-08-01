package model

import (
	"github.com/madnikulin50/lowcode/server/pkg/dal"
)

func init() {
	models = append(models, &dal.Model{
		Ident: "compose_etl_job",
		Attributes: dal.AttributeSet{
			&dal.Attribute{
				Ident: "ID",
				Type:  &dal.TypeID{},
				Store: &dal.CodecAlias{Ident: "id"},
			},
			&dal.Attribute{
				Ident: "ModuleID",
				Type:  &dal.TypeID{},
				Store: &dal.CodecAlias{Ident: "module_id"},
			},
			&dal.Attribute{
				Ident: "NamespaceID",
				Type:  &dal.TypeID{},
				Store: &dal.CodecAlias{Ident: "namespace_id"},
			},
			&dal.Attribute{
				Ident: "Name", Sortable: true,
				Type:  &dal.TypeText{},
				Store: &dal.CodecAlias{Ident: "name"},
			},
			&dal.Attribute{
				Ident: "Enabled",
				Type:  &dal.TypeBoolean{},
				Store: &dal.CodecAlias{Ident: "enabled"},
			},
			&dal.Attribute{
				Ident: "Schedule",
				Type:  &dal.TypeText{},
				Store: &dal.CodecAlias{Ident: "schedule"},
			},
			&dal.Attribute{
				Ident: "Source",
				Type:  &dal.TypeJSON{DefaultValue: "{}"},
				Store: &dal.CodecAlias{Ident: "source"},
			},
			&dal.Attribute{
				Ident: "LastRunAt", Sortable: true,
				Type:  &dal.TypeTimestamp{Nullable: true, Timezone: true, Precision: -1},
				Store: &dal.CodecAlias{Ident: "last_run_at"},
			},
			&dal.Attribute{
				Ident: "LastStatus",
				Type:  &dal.TypeText{},
				Store: &dal.CodecAlias{Ident: "last_status"},
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
				Ident: "idx_etl_namespace_module",
				Type:  "BTREE",
				Fields: []*dal.IndexField{
					{AttributeIdent: "NamespaceID"},
					{AttributeIdent: "ModuleID"},
				},
			},
		},
	})
}
