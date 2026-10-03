package model

import (
	"github.com/madnikulin50/lowcode/server/pkg/dal"
)

func init() {
	models = append(models, &dal.Model{
		Ident: "compose_document",
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
				Ident: "Handle",
				Type:  &dal.TypeText{Length: 128},
				Store: &dal.CodecAlias{Ident: "handle"},
			},
			&dal.Attribute{
				Ident: "Title", Sortable: true,
				Type:  &dal.TypeText{},
				Store: &dal.CodecAlias{Ident: "title"},
			},
			&dal.Attribute{
				Ident: "Kind",
				Type:  &dal.TypeText{Length: 32},
				Store: &dal.CodecAlias{Ident: "kind"},
			},
			&dal.Attribute{
				Ident: "Body",
				Type:  &dal.TypeText{},
				Store: &dal.CodecAlias{Ident: "body"},
			},
			&dal.Attribute{
				Ident: "AttachmentID",
				Type:  &dal.TypeID{Nullable: true},
				Store: &dal.CodecAlias{Ident: "attachment_id"},
			},
			&dal.Attribute{
				Ident: "FileName",
				Type:  &dal.TypeText{},
				Store: &dal.CodecAlias{Ident: "file_name"},
			},
			&dal.Attribute{
				Ident: "Weight", Sortable: true,
				Type: &dal.TypeNumber{
					Precision: 10, HasDefault: true,
					Meta: map[string]any{"rdbms:type": "INTEGER"},
				},
				Store: &dal.CodecAlias{Ident: "weight"},
			},
			&dal.Attribute{
				Ident: "Visible",
				Type:  &dal.TypeBoolean{HasDefault: true, DefaultValue: true},
				Store: &dal.CodecAlias{Ident: "visible"},
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
				Ident: "compose_document_idxNamespace",
				Type:  "BTREE",
				Fields: []*dal.IndexField{
					{AttributeIdent: "NamespaceID"},
					{AttributeIdent: "Weight"},
				},
			},
			&dal.Index{
				Ident:     "compose_document_uniqueHandle",
				Type:      "BTREE",
				Unique:    true,
				Predicate: "handle != '' AND deleted_at IS NULL",
				Fields: []*dal.IndexField{
					{AttributeIdent: "NamespaceID"},
					{AttributeIdent: "Handle", Modifiers: []dal.IndexFieldModifier{"LOWERCASE"}},
				},
			},
		},
	})
}
