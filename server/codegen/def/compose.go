package def

import (
	composetypes "github.com/madnikulin50/lowcode/server/compose/types"
)

// Compose is the Go port of compose/component.cue plus its nine resource
// files, in the same order as compose/component.cue's `resources: {...}`
// declaration.
var Compose = Component{
	Handle: "compose",

	Resources: []NamedResource{
		{Handle: "attachment", Resource: Resource{
			Model: Model{
				Ident:      "compose_attachment",
				Attributes: AttributesFromStruct(composetypes.Attachment{}),
				Indexes: map[string]Index{
					"primary":   {Attribute: "id"},
					"namespace": {Attribute: "namespace_id"},
				},
			},
			Features: Features{Labels: boolPtr(false)},
			Filter: Filter{
				Struct: map[string]Attribute{
					"kind":         {},
					"namespace_id": {},
					"page_id":      {GoType: "uint64", Ident: "pageID"},
					"record_id":    {GoType: "uint64", Ident: "recordID"},
					"module_id":    {GoType: "uint64", Ident: "moduleID"},
					"field_name":   {},
				},
				ByValue: []string{"kind", "namespace_id"},
			},
			// attachment.cue defines no rbac block -> excluded from
			// rbac/types_rbac generation.
			Store: &StoreConfig{
				Ident:   "composeAttachment",
				Lookups: []StoreLookup{{Fields: []string{"id"}}},
			},
		}},

		{Handle: "chart", Resource: Resource{
			Parents: []ParentRef{{Handle: "namespace"}},
			Model: Model{
				DefaultSetter: true,
				Ident:         "compose_chart",
				Attributes:    AttributesFromStruct(composetypes.Chart{}),
				Indexes: map[string]Index{
					"primary":   {Attribute: "id"},
					"namespace": {Attribute: "namespace_id"},
					"unique_handle": {
						Fields:    []IndexField{{Attribute: "handle", Modifiers: []string{"LOWERCASE"}}, {Attribute: "namespace_id"}},
						Predicate: "handle != '' AND deleted_at IS NULL",
					},
				},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"chart_id":     {GoType: "[]uint64", Ident: "chartID", StoreIdent: "id"},
					"namespace_id": {},
					"handle":       {},
					"name":         {},
					"deleted":      {GoType: "filter.State", StoreIdent: "deleted_at"},
				},
				Query:      []string{"handle", "name"},
				ByValue:    []string{"handle", "chart_id", "namespace_id"},
				ByNilState: []string{"deleted"},
			},
			Rbac: &Rbac{Operations: map[string]RbacOperation{
				"read":   {},
				"update": {},
				"delete": {},
			}},
			Store: &StoreConfig{
				Ident: "composeChart",
				Lookups: []StoreLookup{
					{Fields: []string{"id"}, Description: "searches for compose chart by ID\n\nIt returns compose chart even if deleted"},
					{Fields: []string{"namespace_id", "handle"}, NullConstraint: []string{"deleted_at"}, Description: "searches for compose chart by handle (case-insensitive)"},
				},
			},
		}},

		{Handle: "module", Resource: Resource{
			Parents: []ParentRef{{Handle: "namespace"}},
			Model: Model{
				DefaultSetter: true,
				Ident:         "compose_module",
				Attributes:    AttributesFromStruct(composetypes.Module{}),
				Indexes: map[string]Index{
					"primary":   {Attribute: "id"},
					"namespace": {Attribute: "namespace_id"},
					"unique_handle": {
						Fields:    []IndexField{{Attribute: "handle", Modifiers: []string{"LOWERCASE"}}, {Attribute: "namespace_id"}},
						Predicate: "handle != '' AND deleted_at IS NULL",
					},
				},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"module_id":    {GoType: "[]uint64", Ident: "moduleID", StoreIdent: "id"},
					"namespace_id": {},
					"handle":       {},
					"name":         {},
					"deleted":      {GoType: "filter.State", StoreIdent: "deleted_at"},
				},
				Query:      []string{"handle", "name"},
				ByValue:    []string{"handle", "module_id", "namespace_id"},
				ByNilState: []string{"deleted"},
			},
			Rbac: &Rbac{Operations: map[string]RbacOperation{
				"read":                {},
				"update":              {},
				"delete":              {},
				"record.create":       {Description: "Create record"},
				"owned-record.create": {Description: "Create record with custom owner"},
				"records.search":      {Description: "List, search or filter records"},
			}},
			Store: &StoreConfig{
				Ident: "composeModule",
				Lookups: []StoreLookup{
					{Fields: []string{"namespace_id", "handle"}, ConstraintCheck: true, NullConstraint: []string{"deleted_at"}, Description: "searches for compose module by handle (case-insensitive)"},
					{Fields: []string{"namespace_id", "name"}, NullConstraint: []string{"deleted_at"}, Description: "searches for compose module by name (case-insensitive)"},
					{Fields: []string{"id"}, Description: "searches for compose module by ID\n\nIt returns compose module even if deleted"},
				},
			},
		}},

		{Handle: "module-field", Resource: Resource{
			Parents: []ParentRef{{Handle: "namespace"}, {Handle: "module"}},
			Model: Model{
				DefaultSetter: true,
				Ident:         "compose_module_field",
				Attributes:    AttributesFromStruct(composetypes.ModuleField{}),
				Indexes: map[string]Index{
					"primary": {Attribute: "id"},
					"module":  {Attribute: "module_id"},
					"unique_name": {
						Fields:    []IndexField{{Attribute: "name", Modifiers: []string{"LOWERCASE"}}, {Attribute: "module_id"}},
						Predicate: "name != '' AND deleted_at IS NULL",
					},
				},
			},
			Features: Features{Labels: boolPtr(false), Paging: boolPtr(false), Sorting: boolPtr(false), CheckFn: boolPtr(false)},
			Filter: Filter{
				Struct: map[string]Attribute{
					"module_id": {GoType: "[]uint64"},
					"deleted":   {GoType: "filter.State", StoreIdent: "deleted_at"},
				},
				ByNilState: []string{"deleted"},
				ByValue:    []string{"module_id"},
			},
			Rbac: &Rbac{Operations: map[string]RbacOperation{
				"record.value.read":   {Description: "Read field value on records"},
				"record.value.update": {Description: "Update field value on records"},
			}},
			Store: &StoreConfig{
				Ident: "composeModuleField",
				Lookups: []StoreLookup{
					{Fields: []string{"module_id", "name"}, ConstraintCheck: true, NullConstraint: []string{"deleted_at"}, Description: "searches for compose module field by name (case-insensitive)"},
					{Fields: []string{"id"}, Description: "searches for compose module field by ID"},
				},
			},
		}},

		{Handle: "namespace", Resource: Resource{
			Model: Model{
				Ident:      "compose_namespace",
				Attributes: AttributesFromStruct(composetypes.Namespace{}),
				Indexes: map[string]Index{
					"primary": {Attribute: "id"},
					"unique_handle": {
						Fields:    []IndexField{{Attribute: "slug", Modifiers: []string{"LOWERCASE"}}},
						Predicate: "slug != '' AND deleted_at IS NULL",
					},
				},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"namespace_id": {GoType: "[]uint64", Ident: "namespaceID", StoreIdent: "id"},
					"slug":         {},
					"name":         {},
					"deleted":      {GoType: "filter.State", StoreIdent: "deleted_at"},
				},
				Query:      []string{"name", "slug"},
				ByValue:    []string{"namespace_id", "name", "slug"},
				ByNilState: []string{"deleted"},
			},
			Rbac: &Rbac{Operations: map[string]RbacOperation{
				"read":           {},
				"update":         {},
				"delete":         {},
				"export":         {Description: "Access to export the entire namespace"},
				"manage":         {Description: "Access to namespace admin panel"},
				"module.create":  {Description: "Create module on namespace"},
				"modules.search": {Description: "List, search or filter module on namespace"},
				"modules.export": {Description: "Export modules on namespace"},
				"chart.create":   {Description: "Create chart on namespace"},
				"charts.search":  {Description: "List, search or filter chart on namespace"},
				"charts.export":  {Description: "Export charts on namespace"},
				"page.create":    {Description: "Create page on namespace"},
				"pages.search":   {Description: "List, search or filter pages on namespace"},
			}},
			Store: &StoreConfig{
				Ident: "composeNamespace",
				Lookups: []StoreLookup{
					{Fields: []string{"slug"}, ConstraintCheck: true, NullConstraint: []string{"deleted_at"}, Description: "searches for namespace by slug (case-insensitive)"},
					{Fields: []string{"id"}, Description: "searches for compose namespace by ID\n\nIt returns compose namespace even if deleted"},
				},
			},
		}},

		{Handle: "page", Resource: Resource{
			Parents: []ParentRef{{Handle: "namespace"}},
			Model: Model{
				DefaultSetter: true,
				Ident:         "compose_page",
				Attributes:    AttributesFromStruct(composetypes.Page{}),
				Indexes: map[string]Index{
					"primary":   {Attribute: "id"},
					"namespace": {Attribute: "namespace_id"},
					"module":    {Attribute: "module_id"},
					"self_id":   {Attribute: "self_id"},
					"unique_handle": {
						Fields:    []IndexField{{Attribute: "handle", Modifiers: []string{"LOWERCASE"}}, {Attribute: "namespace_id"}},
						Predicate: "handle != '' AND deleted_at IS NULL",
					},
				},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"page_id":      {GoType: "[]uint64", Ident: "pageID", StoreIdent: "id"},
					"namespace_id": {},
					"parent_id":    {GoType: "uint64", Ident: "parentID"},
					"module_id":    {},
					"root":         {GoType: "bool"},
					"handle":       {},
					"title":        {},
					"deleted":      {GoType: "filter.State", StoreIdent: "deleted_at"},
				},
				Query:      []string{"handle", "title", "description"},
				ByValue:    []string{"page_id", "handle", "namespace_id", "module_id"},
				ByNilState: []string{"deleted"},
			},
			Rbac: &Rbac{Operations: map[string]RbacOperation{
				"read":                {},
				"update":              {},
				"delete":              {},
				"page-layout.create":  {Description: "Create page layout on namespace"},
				"page-layouts.search": {Description: "List, search or filter page layouts on namespace"},
			}},
			Store: &StoreConfig{
				Ident: "composePage",
				Lookups: []StoreLookup{
					{Fields: []string{"namespace_id", "handle"}, NullConstraint: []string{"deleted_at"}, Description: "searches for page by handle (case-insensitive)"},
					{Fields: []string{"namespace_id", "module_id"}, NullConstraint: []string{"deleted_at"}, Description: "searches for page by moduleID"},
					{Fields: []string{"id"}, Description: "searches for compose page by ID\n\nIt returns compose page even if deleted"},
				},
				Functions: []StoreFunction{
					{
						ExpIdent: "ReorderComposePages",
						Args: []StoreFunctionArg{
							{Ident: "namespace_id", GoType: "uint64"},
							{Ident: "parent_id", GoType: "uint64"},
							{Ident: "page_ids", GoType: "[]uint64"},
						},
					},
				},
			},
		}},

		{Handle: "page-layout", Resource: Resource{
			Parents: []ParentRef{{Handle: "namespace"}, {Handle: "page"}},
			Model: Model{
				DefaultGetter: true,
				DefaultSetter: true,
				Ident:         "compose_page_layout",
				Attributes:    AttributesFromStruct(composetypes.PageLayout{}),
				Indexes: map[string]Index{
					"primary":   {Attribute: "id"},
					"namespace": {Attribute: "namespace_id"},
					"page_id":   {Attribute: "page_id"},
					"parent_id": {Attribute: "parent_id"},
					"unique_handle": {
						Fields:    []IndexField{{Attribute: "handle", Modifiers: []string{"LOWERCASE"}}, {Attribute: "page_id"}, {Attribute: "namespace_id"}},
						Predicate: "handle != '' AND deleted_at IS NULL",
					},
				},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"page_layout_id": {GoType: "[]uint64", Ident: "pageLayoutID", StoreIdent: "id"},
					"namespace_id":   {},
					"page_id":        {},
					"parent_id":      {},
					"default":        {GoType: "bool", Ident: "default"},
					"handle":         {},
					"deleted":        {GoType: "filter.State", StoreIdent: "deleted_at"},
				},
				Query:      []string{"handle"},
				ByValue:    []string{"handle", "parent_id", "namespace_id", "page_id", "page_layout_id"},
				ByNilState: []string{"deleted"},
			},
			Rbac: &Rbac{Operations: map[string]RbacOperation{
				"read":   {},
				"update": {},
				"delete": {},
			}},
			Store: &StoreConfig{
				Ident: "composePageLayout",
				Lookups: []StoreLookup{
					{Fields: []string{"namespace_id", "handle"}, NullConstraint: []string{"deleted_at"}, Description: "searches for page layour by handle (case-insensitive)"},
					{Fields: []string{"namespace_id", "page_id", "handle"}, NullConstraint: []string{"deleted_at"}, Description: "searches for page layour by handle (case-insensitive)"},
					{Fields: []string{"id"}, Description: "searches for compose page layour by ID\n\nIt returns compose page layour even if deleted"},
				},
				Functions: []StoreFunction{
					{
						ExpIdent: "ReorderComposePageLayouts",
						Args: []StoreFunctionArg{
							{Ident: "namespace_id", GoType: "uint64"},
							{Ident: "page_id", GoType: "uint64"},
							{Ident: "page_layout_ids", GoType: "[]uint64"},
						},
					},
				},
			},
		}},

		{Handle: "record", Resource: Resource{
			Parents: []ParentRef{{Handle: "namespace"}, {Handle: "module"}},
			Model: Model{
				DefaultSetter: true,
				DefaultGetter: true,
				Ident:         "compose_record",
				Attributes:    AttributesFromStruct(composetypes.Record{}),
				Indexes: map[string]Index{
					"primary": {Attribute: "id"},
					"idx_compose_record_base": {
						Attributes: []string{"module_id", "namespace_id"},
						Predicate:  "deleted_at IS NULL",
					},
				},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"module_id":    {},
					"namespace_id": {},
					"query":        {GoType: "string"},
					"deleted":      {GoType: "filter.State", StoreIdent: "deleted_at"},
				},
				ByNilState: []string{"deleted"},
			},
			Rbac: &Rbac{Operations: map[string]RbacOperation{
				"read":             {},
				"update":           {},
				"delete":           {},
				"undelete":         {},
				"owner.manage":     {},
				"revisions.search": {},
			}},
			// record.cue defines no store block -> excluded from store
			// generation, but still gets a dal/$component_model entry
			// (that only checks for model.attributes, not store).
		}},

		{Handle: "record-revision", Resource: Resource{
			Model: Model{
				OmitGetterSetter: true,
				Ident:            "compose_record_revisions",
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: Attribute{ExpIdent: "ID", GoType: "uint64", Unique: true, Dal: &AttributeDal{Type: "ID"}}},
					{Name: "timestamp", Attribute: Attribute{StoreIdent: "ts", Sortable: true, GoType: "time.Time", Dal: &AttributeDal{Type: "Timestamp", Timezone: true}}},
					{Name: "rel_resource", Attribute: Attribute{Ident: "resourceID", GoType: "uint64", Dal: &AttributeDal{Type: "ID"}}},
					{Name: "revision", Attribute: Attribute{GoType: "uint", Dal: &AttributeDal{Type: "Number", Meta: map[string]interface{}{"rdbms:type": "integer"}}}},
					{Name: "operation", Attribute: Attribute{Dal: &AttributeDal{}}},
					{Name: "rel_user", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "delta", Attribute: Attribute{GoType: "types.RecordValueSet", Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "comment", Attribute: Attribute{Dal: &AttributeDal{}}},
				},
				Indexes: map[string]Index{
					"primary": {Attribute: "id"},
				},
			},
			// record_revision.cue defines no rbac and no store block.
		}},
	},
}
