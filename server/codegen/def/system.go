package def

// Shared attribute shortcuts mirroring codegen/schema/model.cue's
// IdField/HandleField/SortableTimestamp*Field (see federation.go's
// userRefDal for the AttributeUserRef equivalent, reused here as-is).
var (
	sysIDField     = Attribute{ExpIdent: "ID", Unique: true, GoType: "uint64", Dal: &AttributeDal{Type: "ID"}}
	sysHandleField = Attribute{Unique: true, IgnoreCase: true, GoType: "string", Dal: &AttributeDal{Type: "Text", Length: 64}}

	sysTimestampField    = Attribute{Sortable: true, GoType: "time.Time", Dal: &AttributeDal{Type: "Timestamp", Timezone: true}}
	sysTimestampNowField = Attribute{Sortable: true, GoType: "time.Time", Dal: &AttributeDal{Type: "Timestamp", Timezone: true, DefaultCurrentTimestamp: true}}
	sysTimestampNilField = Attribute{Sortable: true, GoType: "*time.Time", Dal: &AttributeDal{Type: "Timestamp", Timezone: true, Nullable: true}}
)

// System mirrors system/*.cue (deleted; see git history at commit
// b1ee51053^ for the originals) the same way Federation/Automation/Compose
// mirror their own component.cue + resource .cue files.
var System = Component{
	Handle: "system",
	Resources: []NamedResource{
		{Handle: "attachment", Resource: Resource{
			Features: Features{Labels: boolPtr(false)},
			Model: Model{
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: sysIDField},
					{Name: "owner_id", Attribute: Attribute{Ident: "ownerID", StoreIdent: "rel_owner", GoType: "uint64", Dal: userRefDal}},
					{Name: "kind", Attribute: Attribute{Sortable: true, Dal: &AttributeDal{}}},
					{Name: "url", Attribute: Attribute{Dal: &AttributeDal{}}},
					{Name: "preview_url", Attribute: Attribute{Dal: &AttributeDal{}}},
					{Name: "name", Attribute: Attribute{Sortable: true, Dal: &AttributeDal{}}},
					{Name: "meta", Attribute: Attribute{GoType: "types.AttachmentMeta", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "created_at", Attribute: sysTimestampNowField},
					{Name: "updated_at", Attribute: sysTimestampNilField},
					{Name: "deleted_at", Attribute: sysTimestampNilField},
				},
				Indexes: map[string]Index{"primary": {Attribute: "id"}},
			},
			Filter: Filter{
				Struct:  map[string]Attribute{"kind": {}},
				ByValue: []string{"kind"},
			},
			Store: &StoreConfig{Lookups: []StoreLookup{{Fields: []string{"id"}}}},
		}},

		{Handle: "application", Resource: Resource{
			Model: Model{
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: sysIDField},
					{Name: "name", Attribute: Attribute{Sortable: true, Dal: &AttributeDal{}}},
					{Name: "enabled", Attribute: Attribute{GoType: "bool", Sortable: true, Dal: &AttributeDal{Type: "Boolean", HasDefault: true, DefaultValue: true}}},
					{Name: "weight", Attribute: Attribute{GoType: "int", Sortable: true, Dal: &AttributeDal{Type: "Number", HasDefault: true, DefaultValue: 0, Meta: map[string]interface{}{"rdbms:type": "integer"}}}},
					{Name: "unify", Attribute: Attribute{GoType: "*types.ApplicationUnify", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "owner_id", Attribute: Attribute{Ident: "ownerID", StoreIdent: "rel_owner", GoType: "uint64", Dal: userRefDal}},
					{Name: "created_at", Attribute: sysTimestampNowField},
					{Name: "updated_at", Attribute: sysTimestampNilField},
					{Name: "deleted_at", Attribute: sysTimestampNilField},
				},
				Indexes: map[string]Index{"primary": {Attribute: "id"}},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"name":        {},
					"flagged_ids": {GoType: "[]uint64"},
					"flags":       {GoType: "[]string"},
					"inc_flags":   {GoType: "uint"},
					"deleted":     {GoType: "filter.State", StoreIdent: "deleted_at"},
				},
				Query:      []string{"name"},
				ByValue:    []string{"name"},
				ByNilState: []string{"deleted"},
			},
			Features: Features{Flags: boolPtr(true)},
			Rbac: &Rbac{Operations: map[string]RbacOperation{
				"read":   {Description: "Read application"},
				"update": {Description: "Update application"},
				"delete": {Description: "Delete application"},
			}},
			Store: &StoreConfig{
				Lookups: []StoreLookup{{Fields: []string{"id"}, Description: "searches for role by ID\n\nIt returns role even if deleted or suspended"}},
				Functions: []StoreFunction{
					{ExpIdent: "ApplicationMetrics", Return: []string{"*types.ApplicationMetrics"}},
					{ExpIdent: "ReorderApplications", Args: []StoreFunctionArg{{Ident: "order", GoType: "[]uint64"}}},
				},
			},
		}},

		{Handle: "apigw-route", Resource: Resource{
			Features: Features{Labels: boolPtr(false)},
			Model: Model{
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: sysIDField},
					{Name: "endpoint", Attribute: Attribute{Sortable: true, Dal: &AttributeDal{}}},
					{Name: "method", Attribute: Attribute{Sortable: true, Dal: &AttributeDal{}}},
					{Name: "enabled", Attribute: Attribute{GoType: "bool", Sortable: true, Dal: &AttributeDal{Type: "Boolean"}}},
					{Name: "meta", Attribute: Attribute{GoType: "types.ApigwRouteMeta", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "group", Attribute: Attribute{Sortable: true, GoType: "uint64", StoreIdent: "rel_group", Dal: &AttributeDal{Type: "Ref", RefModelResType: "corteza::system:apigw-group"}}},
					{Name: "created_at", Attribute: sysTimestampNowField},
					{Name: "updated_at", Attribute: sysTimestampNilField},
					{Name: "deleted_at", Attribute: sysTimestampNilField},
					{Name: "created_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "updated_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "deleted_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
				},
				Indexes: map[string]Index{"primary": {Attribute: "id"}},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"apigw_route_id": {GoType: "[]uint64", Ident: "apigwRouteID", StoreIdent: "id"},
					"route":          {GoType: "string", StoreIdent: "id"},
					"endpoint":       {},
					"method":         {},
					"deleted":        {GoType: "filter.State", StoreIdent: "deleted_at"},
					"disabled":       {GoType: "filter.State", StoreIdent: "enabled"},
				},
				ByValue:      []string{"apigw_route_id", "route", "method"},
				ByNilState:   []string{"deleted"},
				ByFalseState: []string{"disabled"},
			},
			Rbac: &Rbac{Operations: map[string]RbacOperation{
				"read":   {Description: "Read API Gateway route"},
				"update": {Description: "Update API Gateway route"},
				"delete": {Description: "Delete API Gateway route"},
			}},
			Store: &StoreConfig{
				Lookups: []StoreLookup{
					{Fields: []string{"id"}, Description: "searches for route by ID\n\nIt returns route even if deleted or suspended"},
					{Fields: []string{"endpoint"}, Description: "searches for route by endpoint\n\nIt returns route even if deleted or suspended"},
				},
			},
		}},

		{Handle: "apigw-filter", Resource: Resource{
			Features: Features{Labels: boolPtr(false)},
			Model: Model{
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: sysIDField},
					{Name: "route", Attribute: Attribute{Sortable: true, GoType: "uint64", StoreIdent: "rel_route", IdentAlias: []string{"route", "Route", "ApigwRouteID"}, Dal: &AttributeDal{Type: "Ref", RefModelResType: "corteza::system:apigw-route"}}},
					{Name: "weight", Attribute: Attribute{Sortable: true, GoType: "uint64", Dal: &AttributeDal{Type: "Number", Meta: map[string]interface{}{"rdbms:type": "integer"}}}},
					{Name: "kind", Attribute: Attribute{Sortable: true, Dal: &AttributeDal{Type: "Text", Length: 64}}},
					{Name: "ref", Attribute: Attribute{Dal: &AttributeDal{Type: "Text", Length: 64}}},
					{Name: "enabled", Attribute: Attribute{Sortable: true, GoType: "bool", Dal: &AttributeDal{Type: "Boolean"}}},
					{Name: "params", Attribute: Attribute{GoType: "types.ApigwFilterParams", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "created_at", Attribute: sysTimestampNowField},
					{Name: "updated_at", Attribute: sysTimestampNilField},
					{Name: "deleted_at", Attribute: sysTimestampNilField},
					{Name: "created_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "updated_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "deleted_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
				},
				Indexes: map[string]Index{"primary": {Attribute: "id"}},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"apigw_filter_id": {GoType: "[]uint64", Ident: "apigwFilterID", StoreIdent: "id"},
					"route_id":        {GoType: "uint64", Ident: "routeID", StoreIdent: "rel_route"},
					"deleted":         {GoType: "filter.State", StoreIdent: "deleted_at"},
					"disabled":        {GoType: "filter.State", StoreIdent: "enabled"},
				},
				ByValue:      []string{"apigw_filter_id", "route_id"},
				ByNilState:   []string{"deleted"},
				ByFalseState: []string{"disabled"},
			},
			Store: &StoreConfig{
				Lookups: []StoreLookup{
					{Fields: []string{"id"}, Description: "searches for filter by ID"},
					{Fields: []string{"route"}, Description: "searches for filter by route"},
				},
			},
		}},

		{Handle: "auth-client", Resource: Resource{
			Model: Model{
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: sysIDField},
					{Name: "handle", Attribute: sysHandleField},
					{Name: "meta", Attribute: Attribute{GoType: "*types.AuthClientMeta", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "secret", Attribute: Attribute{GoType: "string", Dal: &AttributeDal{Type: "Text", Length: 64}}},
					{Name: "scope", Attribute: Attribute{GoType: "string", Dal: &AttributeDal{Type: "Text", Length: 512}}},
					{Name: "valid_grant", Attribute: Attribute{GoType: "string", Dal: &AttributeDal{Type: "Text", Length: 32}}},
					{Name: "redirect_uri", Attribute: Attribute{GoType: "string", Ident: "redirectURI", Dal: &AttributeDal{}}},
					{Name: "enabled", Attribute: Attribute{Sortable: true, GoType: "bool", Dal: &AttributeDal{Type: "Boolean", HasDefault: true, DefaultValue: false}}},
					{Name: "trusted", Attribute: Attribute{Sortable: true, GoType: "bool", Dal: &AttributeDal{Type: "Boolean", HasDefault: true, DefaultValue: false}}},
					{Name: "valid_from", Attribute: sysTimestampNilField},
					{Name: "expires_at", Attribute: sysTimestampNilField},
					{Name: "security", Attribute: Attribute{GoType: "*types.AuthClientSecurity", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "owned_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "created_at", Attribute: sysTimestampNowField},
					{Name: "updated_at", Attribute: sysTimestampNilField},
					{Name: "deleted_at", Attribute: sysTimestampNilField},
					{Name: "created_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "updated_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "deleted_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
				},
				Indexes: map[string]Index{"primary": {Attribute: "id"}},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"client_id": {GoType: "[]uint64"},
					"handle":    {},
					"deleted":   {GoType: "filter.State", StoreIdent: "deleted_at"},
				},
				ByValue:    []string{"handle"},
				ByNilState: []string{"deleted"},
			},
			Rbac: &Rbac{Operations: map[string]RbacOperation{
				"read":      {Description: "Read authorization client"},
				"update":    {Description: "Update authorization client"},
				"delete":    {Description: "Delete authorization client"},
				"authorize": {Description: "Authorize authorization client"},
			}},
			Store: &StoreConfig{
				Lookups: []StoreLookup{
					{Fields: []string{"id"}, Description: "searches for auth client by ID\n\nIt returns auth clint even if deleted"},
					{Fields: []string{"handle"}, NullConstraint: []string{"deleted_at"}, ConstraintCheck: true, Description: "searches for auth client by ID\n\nIt returns auth clint even if deleted"},
				},
			},
		}},

		{Handle: "auth-confirmed-client", Resource: Resource{
			Features: Features{Labels: boolPtr(false), Paging: boolPtr(false), Sorting: boolPtr(false), CheckFn: boolPtr(false)},
			Model: Model{
				OmitGetterSetter: true,
				Attributes: []NamedAttribute{
					{Name: "user_id", Attribute: Attribute{GoType: "uint64", Ident: "userID", StoreIdent: "rel_user", Dal: &AttributeDal{Type: "Ref", RefModelResType: "corteza::system:user", HasDefault: true, DefaultValue: 0}}},
					{Name: "client_id", Attribute: Attribute{GoType: "uint64", Ident: "clientID", StoreIdent: "rel_client", Dal: &AttributeDal{Type: "Ref", RefModelResType: "corteza::system:auth-client", HasDefault: true, DefaultValue: 0}}},
					{Name: "confirmed_at", Attribute: sysTimestampNowField},
				},
				Indexes: map[string]Index{"primary": {Attributes: []string{"user_id", "client_id"}}},
			},
			Filter: Filter{
				Struct:  map[string]Attribute{"user_id": {GoType: "uint64", Ident: "userID", StoreIdent: "rel_user"}},
				ByValue: []string{"user_id"},
			},
			Store: &StoreConfig{Lookups: []StoreLookup{{Fields: []string{"user_id", "client_id"}}}},
		}},

		{Handle: "auth-session", Resource: Resource{
			Features: Features{Labels: boolPtr(false), Paging: boolPtr(false), Sorting: boolPtr(false), CheckFn: boolPtr(false)},
			Model: Model{
				OmitGetterSetter: true,
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: Attribute{ExpIdent: "ID", GoType: "string", Dal: &AttributeDal{Type: "Text", Length: 64}}},
					{Name: "data", Attribute: Attribute{GoType: "[]byte", Dal: &AttributeDal{Type: "Blob"}}},
					{Name: "user_id", Attribute: Attribute{GoType: "uint64", Ident: "userID", StoreIdent: "rel_user", Dal: &AttributeDal{Type: "Ref", RefModelResType: "corteza::system:user", HasDefault: true, DefaultValue: 0}}},
					{Name: "remote_addr", Attribute: Attribute{Dal: &AttributeDal{}}},
					{Name: "user_agent", Attribute: Attribute{Dal: &AttributeDal{}}},
					{Name: "expires_at", Attribute: sysTimestampField},
					{Name: "created_at", Attribute: sysTimestampNowField},
				},
				Indexes: map[string]Index{
					"primary":    {Attribute: "id"},
					"expires_at": {Attribute: "expires_at"},
				},
			},
			Filter: Filter{
				Struct:  map[string]Attribute{"user_id": {GoType: "uint64", Ident: "userID", StoreIdent: "rel_user"}},
				ByValue: []string{"user_id"},
			},
			Store: &StoreConfig{
				Lookups: []StoreLookup{{Fields: []string{"id"}}},
				Functions: []StoreFunction{
					{ExpIdent: "DeleteExpiredAuthSessions"},
					{ExpIdent: "DeleteAuthSessionsByUserID", Args: []StoreFunctionArg{{Ident: "userID", GoType: "uint64"}}},
				},
			},
		}},

		{Handle: "auth-oa2token", Resource: Resource{
			Features: Features{Labels: boolPtr(false), Paging: boolPtr(false), Sorting: boolPtr(false), CheckFn: boolPtr(false)},
			Model: Model{
				OmitGetterSetter: true,
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: sysIDField},
					{Name: "code", Attribute: Attribute{Dal: &AttributeDal{Type: "Text", Length: 48}}},
					{Name: "access", Attribute: Attribute{Dal: &AttributeDal{Type: "Text", Length: 2048}}},
					{Name: "refresh", Attribute: Attribute{Dal: &AttributeDal{Type: "Text", Length: 48}}},
					{Name: "data", Attribute: Attribute{GoType: "rawJson", Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "remote_addr", Attribute: Attribute{Dal: &AttributeDal{Type: "Text", Length: 64}}},
					{Name: "user_agent", Attribute: Attribute{Dal: &AttributeDal{}}},
					{Name: "client_id", Attribute: Attribute{GoType: "uint64", Ident: "clientID", StoreIdent: "rel_client", Dal: &AttributeDal{Type: "Ref", RefModelResType: "corteza::system:auth-client", HasDefault: true, DefaultValue: 0}}},
					{Name: "user_id", Attribute: Attribute{GoType: "uint64", Ident: "userID", StoreIdent: "rel_user", Dal: &AttributeDal{Type: "Ref", RefModelResType: "corteza::system:user", HasDefault: true, DefaultValue: 0}}},
					{Name: "created_at", Attribute: sysTimestampNowField},
					{Name: "expires_at", Attribute: sysTimestampField},
				},
				Indexes: map[string]Index{
					"primary":   {Attribute: "id"},
					"client_id": {Attribute: "client_id"},
					"code":      {Attribute: "code"},
					"refresh":   {Attribute: "refresh"},
				},
			},
			Filter: Filter{
				Struct:  map[string]Attribute{"user_id": {GoType: "uint64", Ident: "userID"}},
				ByValue: []string{"user_id"},
			},
			Store: &StoreConfig{
				Lookups: []StoreLookup{
					{Fields: []string{"id"}},
					{Fields: []string{"code"}},
					{Fields: []string{"access"}},
					{Fields: []string{"refresh"}},
				},
				Functions: []StoreFunction{
					{ExpIdent: "DeleteExpiredAuthOA2Tokens"},
					{ExpIdent: "DeleteAuthOA2TokenByCode", Args: []StoreFunctionArg{{Ident: "code", GoType: "string"}}},
					{ExpIdent: "DeleteAuthOA2TokenByAccess", Args: []StoreFunctionArg{{Ident: "access", GoType: "string"}}},
					{ExpIdent: "DeleteAuthOA2TokenByRefresh", Args: []StoreFunctionArg{{Ident: "refresh", GoType: "string"}}},
					{ExpIdent: "DeleteAuthOA2TokenByUserID", Args: []StoreFunctionArg{{Ident: "userID", GoType: "uint64"}}},
				},
			},
		}},

		{Handle: "credential", Resource: Resource{
			Features: Features{Labels: boolPtr(false), Paging: boolPtr(false), Sorting: boolPtr(false), CheckFn: boolPtr(false)},
			Model: Model{
				OmitGetterSetter: true,
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: sysIDField},
					{Name: "owner_id", Attribute: Attribute{Ident: "ownerID", StoreIdent: "rel_owner", GoType: "uint64", Dal: userRefDal}},
					{Name: "label", Attribute: Attribute{Dal: &AttributeDal{}}},
					{Name: "kind", Attribute: Attribute{Dal: &AttributeDal{Type: "Text", Length: 128}}},
					{Name: "credentials", Attribute: Attribute{Dal: &AttributeDal{}}},
					{Name: "meta", Attribute: Attribute{GoType: "rawJson", Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "created_at", Attribute: sysTimestampNowField},
					{Name: "updated_at", Attribute: sysTimestampNilField},
					{Name: "deleted_at", Attribute: sysTimestampNilField},
					{Name: "last_used_at", Attribute: sysTimestampNilField},
					{Name: "expires_at", Attribute: sysTimestampNilField},
				},
				Indexes: map[string]Index{
					"primary": {Attribute: "id"},
					"owner_kind": {
						Attributes: []string{"owner_id", "kind"},
						Predicate:  "deleted_at IS NULL",
					},
				},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"owner_id":    {GoType: "uint64", Ident: "ownerID", StoreIdent: "rel_owner"},
					"kind":        {GoType: "string"},
					"credentials": {GoType: "string"},
					"deleted":     {GoType: "filter.State", StoreIdent: "deleted_at"},
				},
				ByValue:    []string{"owner_id", "kind", "credentials"},
				ByNilState: []string{"deleted"},
			},
			Store: &StoreConfig{
				Lookups: []StoreLookup{{Fields: []string{"id"}, Description: "searches for credentials by ID\n\nIt returns credentials even if deleted"}},
			},
		}},

		{Handle: "data-privacy-request", Resource: Resource{
			Features: Features{Labels: boolPtr(false)},
			Model: Model{
				OmitGetterSetter: true,
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: sysIDField},
					{Name: "kind", Attribute: Attribute{GoType: "types.RequestKind", Sortable: true, Dal: &AttributeDal{}}},
					{Name: "status", Attribute: Attribute{GoType: "types.RequestStatus", Sortable: true, Dal: &AttributeDal{Type: "Text", Length: 64}}},
					{Name: "payload", Attribute: Attribute{GoType: "types.DataPrivacyRequestPayloadSet", Dal: &AttributeDal{Type: "JSON"}}},
					{Name: "requested_at", Attribute: sysTimestampField},
					{Name: "requested_by", Attribute: Attribute{GoType: "uint64", Dal: &AttributeDal{Type: "Ref", RefModelResType: "corteza::system:user"}}},
					{Name: "completed_at", Attribute: sysTimestampNilField},
					{Name: "completed_by", Attribute: Attribute{GoType: "uint64", Dal: &AttributeDal{Type: "Ref", RefModelResType: "corteza::system:user"}}},
					{Name: "created_at", Attribute: sysTimestampNowField},
					{Name: "updated_at", Attribute: sysTimestampNilField},
					{Name: "deleted_at", Attribute: sysTimestampNilField},
					{Name: "created_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "updated_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "deleted_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
				},
				Indexes: map[string]Index{"primary": {Attribute: "id"}},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"request_id":   {GoType: "[]uint64", Ident: "requestID", StoreIdent: "id"},
					"requested_by": {GoType: "[]uint64", Ident: "requestedBy"},
					"kind":         {GoType: "[]types.RequestKind"},
					"status":       {GoType: "[]types.RequestStatus"},
				},
				Query:   []string{"kind", "status"},
				ByValue: []string{"kind", "status", "requested_by"},
			},
			Rbac: &Rbac{Operations: map[string]RbacOperation{
				"read":    {Description: "Read data privacy request"},
				"approve": {Description: "Approve/Reject data privacy request"},
			}},
			Store: &StoreConfig{
				Lookups: []StoreLookup{{Fields: []string{"id"}, Description: "searches for data privacy request by ID\n\nIt returns data privacy request even if deleted"}},
			},
		}},

		{Handle: "data-privacy-request-comment", Resource: Resource{
			Features: Features{Labels: boolPtr(false)},
			Model: Model{
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: sysIDField},
					{Name: "request_id", Attribute: Attribute{Ident: "requestID", GoType: "uint64", StoreIdent: "rel_request", Dal: &AttributeDal{Type: "Ref", RefModelResType: "corteza::system:user"}}},
					{Name: "comment", Attribute: Attribute{GoType: "string", Dal: &AttributeDal{}}},
					{Name: "created_at", Attribute: sysTimestampNowField},
					{Name: "updated_at", Attribute: sysTimestampNilField},
					{Name: "deleted_at", Attribute: sysTimestampNilField},
					{Name: "created_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "updated_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "deleted_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
				},
				Indexes: map[string]Index{"primary": {Attribute: "id"}},
			},
			Filter: Filter{
				Struct:  map[string]Attribute{"request_id": {GoType: "[]uint64", Ident: "requestID", StoreIdent: "rel_request"}},
				ByValue: []string{"request_id"},
			},
			Store: &StoreConfig{},
		}},

		{Handle: "queue", Resource: Resource{
			Features: Features{Labels: boolPtr(false)},
			Model: Model{
				Ident: "queue_settings",
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: sysIDField},
					{Name: "consumer", Attribute: Attribute{Sortable: true, GoType: "string", Dal: &AttributeDal{}}},
					{Name: "queue", Attribute: Attribute{Sortable: true, GoType: "string", Dal: &AttributeDal{}}},
					{Name: "meta", Attribute: Attribute{GoType: "types.QueueMeta", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "created_at", Attribute: sysTimestampNowField},
					{Name: "updated_at", Attribute: sysTimestampNilField},
					{Name: "deleted_at", Attribute: sysTimestampNilField},
					{Name: "created_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "updated_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "deleted_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
				},
				Indexes: map[string]Index{"primary": {Attribute: "id"}},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"queue_id": {GoType: "[]uint64", Ident: "queueID", StoreIdent: "id"},
					"query":    {GoType: "string"},
					"deleted":  {GoType: "filter.State", StoreIdent: "deleted_at"},
				},
				Query:      []string{"queue", "consumer"},
				ByValue:    []string{"queue_id"},
				ByNilState: []string{"deleted"},
			},
			Rbac: &Rbac{Operations: map[string]RbacOperation{
				"read":        {Description: "Read queue"},
				"update":      {Description: "Update queue"},
				"delete":      {Description: "Delete queue"},
				"queue.read":  {Description: "Read from queue"},
				"queue.write": {Description: "Write to queue"},
			}},
			Store: &StoreConfig{
				Lookups: []StoreLookup{
					{Fields: []string{"id"}, Description: "searches for queue by ID"},
					{Fields: []string{"queue"}, Description: "searches for queue by queue name"},
				},
			},
		}},

		{Handle: "queue-message", Resource: Resource{
			Features: Features{Labels: boolPtr(false), CheckFn: boolPtr(false)},
			Model: Model{
				OmitGetterSetter: true,
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: sysIDField},
					{Name: "queue", Attribute: Attribute{Sortable: true, Dal: &AttributeDal{}}},
					{Name: "payload", Attribute: Attribute{GoType: "[]byte", Dal: &AttributeDal{Type: "Blob"}}},
					{Name: "created", Attribute: sysTimestampNilField},
					{Name: "processed", Attribute: sysTimestampNilField},
				},
				Indexes: map[string]Index{"primary": {Attribute: "id"}},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"queue":     {},
					"processed": {GoType: "filter.State", StoreIdent: "processed"},
				},
				ByValue:    []string{"queue"},
				ByNilState: []string{"processed"},
			},
			Store: &StoreConfig{},
		}},

		{Handle: "reminder", Resource: Resource{
			Features: Features{Labels: boolPtr(false)},
			Model: Model{
				OmitGetterSetter: true,
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: sysIDField},
					{Name: "resource", Attribute: Attribute{Sortable: true, Dal: &AttributeDal{Type: "Text", Length: 512}}},
					{Name: "payload", Attribute: Attribute{GoType: "rawJson", Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "snooze_count", Attribute: Attribute{GoType: "uint", Dal: &AttributeDal{Type: "Number", Meta: map[string]interface{}{"rdbms:type": "integer"}}}},
					{Name: "assigned_to", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "assigned_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "assigned_at", Attribute: sysTimestampField},
					{Name: "dismissed_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "dismissed_at", Attribute: sysTimestampNilField},
					{Name: "remind_at", Attribute: sysTimestampNilField},
					{Name: "created_at", Attribute: sysTimestampNowField},
					{Name: "updated_at", Attribute: sysTimestampNilField},
					{Name: "deleted_at", Attribute: sysTimestampNilField},
				},
				Indexes: map[string]Index{
					"primary":     {Attribute: "id"},
					"assigned_to": {Attribute: "assigned_to"},
					"resource":    {Attribute: "resource"},
				},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"reminder_id":       {GoType: "[]uint64", Ident: "reminderID", StoreIdent: "id"},
					"resource":          {},
					"assigned_to":       {GoType: "uint64"},
					"scheduled_from":    {GoType: "uint64"},
					"scheduled_until":   {GoType: "uint64"},
					"exclude_dismissed": {GoType: "bool"},
					"include_deleted":   {GoType: "bool"},
					"scheduled_only":    {GoType: "bool"},
					"deleted":           {GoType: "filter.State", StoreIdent: "deleted_at"},
				},
				ByValue: []string{"reminder_id", "assigned_to"},
			},
			Store: &StoreConfig{Lookups: []StoreLookup{{Fields: []string{"id"}}}},
		}},

		{Handle: "notification", Resource: Resource{
			Features: Features{Labels: boolPtr(false)},
			Model: Model{
				OmitGetterSetter: true,
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: sysIDField},
					{Name: "kind", Attribute: Attribute{Sortable: true, GoType: "types.NotificationKind", Dal: &AttributeDal{Type: "Text", Length: 32}}},
					{Name: "config", Attribute: Attribute{GoType: "types.NotificationConfig", Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "recipient", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "created_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "read_at", Attribute: sysTimestampNilField},
					{Name: "created_at", Attribute: sysTimestampNowField},
					{Name: "updated_at", Attribute: sysTimestampNilField},
					{Name: "deleted_at", Attribute: sysTimestampNilField},
				},
				Indexes: map[string]Index{
					"primary":   {Attribute: "id"},
					"recipient": {Attribute: "recipient"},
					"kind":      {Attribute: "kind"},
				},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"notification_id": {GoType: "[]uint64", Ident: "notificationID", StoreIdent: "id"},
					"kind":            {GoType: "[]types.NotificationKind"},
					"recipient":       {GoType: "uint64"},
					"read":            {GoType: "filter.State", StoreIdent: "read_at"},
					"deleted":         {GoType: "filter.State", StoreIdent: "deleted_at"},
				},
				ByValue:    []string{"notification_id", "recipient"},
				ByNilState: []string{"read", "deleted"},
			},
			Store: &StoreConfig{Lookups: []StoreLookup{{Fields: []string{"id"}}}},
		}},

		{Handle: "report", Resource: Resource{
			Model: Model{
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: sysIDField},
					{Name: "handle", Attribute: sysHandleField},
					{Name: "meta", Attribute: Attribute{GoType: "*types.ReportMeta", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "scenarios", Attribute: Attribute{GoType: "types.ReportScenarioSet", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "sources", Attribute: Attribute{GoType: "types.ReportDataSourceSet", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "blocks", Attribute: Attribute{GoType: "types.ReportBlockSet", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "owned_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "created_at", Attribute: sysTimestampNowField},
					{Name: "updated_at", Attribute: sysTimestampNilField},
					{Name: "deleted_at", Attribute: sysTimestampNilField},
					{Name: "created_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "updated_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "deleted_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
				},
				Indexes: map[string]Index{"primary": {Attribute: "id"}},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"report_id": {GoType: "[]uint64", StoreIdent: "id", Ident: "reportID"},
					"handle":    {},
					"deleted":   {GoType: "filter.State", StoreIdent: "deleted_at"},
				},
				Query:      []string{"handle"},
				ByValue:    []string{"handle", "report_id"},
				ByNilState: []string{"deleted"},
			},
			Rbac: &Rbac{Operations: map[string]RbacOperation{
				"read":   {Description: "Read report"},
				"update": {Description: "Update report"},
				"delete": {Description: "Delete report"},
				"run":    {Description: "Run report"},
			}},
			Store: &StoreConfig{
				Lookups: []StoreLookup{
					{Fields: []string{"id"}, Description: "searches for report by ID\n\nIt returns report even if deleted"},
					{Fields: []string{"handle"}, NullConstraint: []string{"deleted_at"}, ConstraintCheck: true, Description: "searches for report by handle\n\nIt returns report if deleted"},
				},
			},
		}},

		{Handle: "resource-translation", Resource: Resource{
			Features: Features{Labels: boolPtr(false), CheckFn: boolPtr(false)},
			Model: Model{
				DefaultSetter: true,
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: sysIDField},
					{Name: "lang", Attribute: Attribute{GoType: "types.Lang", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Type: "Text", Length: 32}}},
					{Name: "resource", Attribute: Attribute{Dal: &AttributeDal{Type: "Text", Length: 256}}},
					{Name: "k", Attribute: Attribute{Dal: &AttributeDal{Type: "Text", Length: 256}}},
					{Name: "message", Attribute: Attribute{Dal: &AttributeDal{}}},
					{Name: "created_at", Attribute: sysTimestampNowField},
					{Name: "updated_at", Attribute: sysTimestampNilField},
					{Name: "deleted_at", Attribute: sysTimestampNilField},
					{Name: "owned_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "created_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "updated_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "deleted_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
				},
				Indexes: map[string]Index{
					"primary": {Attribute: "id"},
					"unique_translation": {
						Fields: []IndexField{
							{Attribute: "lang", Modifiers: []string{"LOWERCASE"}},
							{Attribute: "resource", Modifiers: []string{"LOWERCASE"}},
							{Attribute: "k", Modifiers: []string{"LOWERCASE"}},
						},
					},
				},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"translation_id": {GoType: "[]uint64", Ident: "translationID"},
					"lang":           {GoType: "string"},
					"resource":       {},
					"resourceType":   {},
					"owner_id":       {GoType: "uint64", Ident: "ownerID", StoreIdent: "rel_owner"},
					"deleted":        {GoType: "filter.State", StoreIdent: "deleted_at"},
				},
				ByValue:    []string{"resource", "lang", "translation_id"},
				ByNilState: []string{"deleted"},
			},
			Store: &StoreConfig{
				Lookups: []StoreLookup{{Fields: []string{"id"}, Description: "searches for resource translation by ID\nIt also returns deleted resource translations."}},
				Functions: []StoreFunction{
					{ExpIdent: "TransformResource", Args: []StoreFunctionArg{{Ident: "lang", GoType: "language.Tag"}}, Return: []string{"map[string]map[string]*locale.ResourceTranslation"}},
				},
			},
		}},

		{Handle: "role", Resource: Resource{
			Model: Model{
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: sysIDField},
					{Name: "name", Attribute: Attribute{Sortable: true, Dal: &AttributeDal{}}},
					{Name: "handle", Attribute: sysHandleField},
					{Name: "meta", Attribute: Attribute{GoType: "*types.RoleMeta", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "archived_at", Attribute: sysTimestampNilField},
					{Name: "created_at", Attribute: sysTimestampNowField},
					{Name: "updated_at", Attribute: sysTimestampNilField},
					{Name: "deleted_at", Attribute: sysTimestampNilField},
				},
				Indexes: map[string]Index{"primary": {Attribute: "id"}},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"role_id":       {GoType: "[]uint64", Ident: "roleID", StoreIdent: "id"},
					"member_id":     {GoType: "uint64"},
					"user_group_id": {GoType: "uint64"},
					"resource":      {GoType: "string"},
					"handle":        {},
					"name":          {},
					"deleted":       {GoType: "filter.State", StoreIdent: "deleted_at"},
					"archived":      {GoType: "filter.State", StoreIdent: "archived_at"},
				},
				Query:      []string{"handle", "name"},
				ByValue:    []string{"role_id", "name", "handle"},
				ByNilState: []string{"deleted", "archived"},
			},
			Rbac: &Rbac{Operations: map[string]RbacOperation{
				"read":           {Description: "Read role"},
				"update":         {Description: "Update role"},
				"delete":         {Description: "Delete role"},
				"members.manage": {Description: "Manage members"},
			}},
			Store: &StoreConfig{
				Lookups: []StoreLookup{
					{Fields: []string{"id"}, Description: "searches for role by ID\n\nIt returns role even if deleted or suspended"},
					{Fields: []string{"handle"}, NullConstraint: []string{"deleted_at"}, ConstraintCheck: true, Description: "searches for role by handle\n\nIt returns only valid role (not deleted, not suspended)"},
					{Fields: []string{"name"}, NullConstraint: []string{"deleted_at"}, ConstraintCheck: true, Description: "searches for role by name\n\nIt returns only valid role (not deleted, not suspended)"},
				},
				Functions: []StoreFunction{{ExpIdent: "RoleMetrics", Return: []string{"*types.RoleMetrics"}}},
			},
		}},

		{Handle: "role-member", Resource: Resource{
			Features: Features{Labels: boolPtr(false), Paging: boolPtr(false), Sorting: boolPtr(false), CheckFn: boolPtr(false)},
			Model: Model{
				Attributes: []NamedAttribute{
					{Name: "resource", Attribute: Attribute{GoType: "string", StoreIdent: "rel_resource", Ident: "resource", Dal: &AttributeDal{}}},
					{Name: "role_id", Attribute: Attribute{GoType: "uint64", StoreIdent: "rel_role", Ident: "roleID", Dal: &AttributeDal{Type: "Ref", RefModelResType: "corteza::system:role"}}},
				},
				Indexes: map[string]Index{"primary": {Attributes: []string{"resource", "role_id"}}},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"resource": {GoType: "string", Ident: "resource", StoreIdent: "rel_resource"},
					"role_id":  {GoType: "uint64", Ident: "roleID", StoreIdent: "rel_role"},
				},
				ByValue: []string{"resource", "role_id"},
			},
			Store: &StoreConfig{
				Functions: []StoreFunction{
					{ExpIdent: "TransferRoleMembers", Args: []StoreFunctionArg{{Ident: "src", GoType: "uint64"}, {Ident: "dst", GoType: "uint64"}}},
				},
			},
		}},

		{Handle: "user-group", Resource: Resource{
			Model: Model{
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: sysIDField},
					{Name: "handle", Attribute: sysHandleField},
					{Name: "meta", Attribute: Attribute{GoType: "*types.UserGroupMeta", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "config", Attribute: Attribute{GoType: "*types.UserGroupConfig", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "archived_at", Attribute: sysTimestampNilField},
					{Name: "created_at", Attribute: sysTimestampNowField},
					{Name: "updated_at", Attribute: sysTimestampNilField},
					{Name: "deleted_at", Attribute: sysTimestampNilField},
				},
				Indexes: map[string]Index{"primary": {Attribute: "id"}},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"user_group_id": {GoType: "[]uint64", Ident: "userGroupID", StoreIdent: "id"},
					"member_id":     {GoType: "uint64"},
					"handle":        {},
					"deleted":       {GoType: "filter.State", StoreIdent: "deleted_at"},
					"archived":      {GoType: "filter.State", StoreIdent: "archived_at"},
				},
				Query:      []string{"handle"},
				ByValue:    []string{"user_group_id", "handle"},
				ByNilState: []string{"deleted", "archived"},
			},
			Rbac: &Rbac{Operations: map[string]RbacOperation{
				"read":           {Description: "Read user group"},
				"update":         {Description: "Update user group"},
				"delete":         {Description: "Delete user group"},
				"members.manage": {Description: "Manage members"},
			}},
			Store: &StoreConfig{
				Lookups: []StoreLookup{
					{Fields: []string{"id"}, Description: "searches for user group by ID\n\nIt returns user group even if deleted or suspended"},
					{Fields: []string{"handle"}, NullConstraint: []string{"deleted_at"}, ConstraintCheck: true, Description: "searches for user group by handle\n\nIt returns only valid user group (not deleted, not suspended)"},
				},
			},
		}},

		{Handle: "settings", Resource: Resource{
			Ident:    "settingValue",
			ExpIdent: "SettingValue",
			Features: Features{Labels: boolPtr(false), Paging: boolPtr(false), Sorting: boolPtr(false), CheckFn: boolPtr(false)},
			Model: Model{
				Ident:            "settings",
				OmitGetterSetter: true,
				Attributes: []NamedAttribute{
					{Name: "owned_by", Attribute: Attribute{GoType: "uint64", StoreIdent: "rel_owner", Dal: &AttributeDal{Type: "Ref", RefModelResType: "corteza::system:user"}}},
					{Name: "name", Attribute: Attribute{Dal: &AttributeDal{Type: "Text", Length: 512}}},
					{Name: "value", Attribute: Attribute{GoType: "rawJson", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Type: "JSON"}}},
					{Name: "updated_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "updated_at", Attribute: sysTimestampField},
				},
				Indexes: map[string]Index{
					"primary": {Fields: []IndexField{{Attribute: "owned_by"}, {Attribute: "name"}}},
				},
			},
			Filter: Filter{
				ExpIdent: "SettingsFilter",
				Struct: map[string]Attribute{
					"prefix":   {},
					"owned_by": {GoType: "uint64", StoreIdent: "rel_owner"},
				},
				ByValue: []string{"owned_by"},
			},
			Store: &StoreConfig{
				Lookups: []StoreLookup{{Fields: []string{"name", "owned_by"}, Description: "searches for settings by name and owner"}},
			},
		}},

		{Handle: "template", Resource: Resource{
			Model: Model{
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: sysIDField},
					{Name: "owner_id", Attribute: Attribute{StoreIdent: "rel_owner", Ident: "ownerID", GoType: "uint64", Dal: userRefDal}},
					{Name: "handle", Attribute: sysHandleField},
					{Name: "language", Attribute: Attribute{Sortable: true, GoType: "string", Dal: &AttributeDal{Length: 32}}},
					{Name: "type", Attribute: Attribute{Sortable: true, GoType: "types.DocumentType", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{}}},
					{Name: "partial", Attribute: Attribute{GoType: "bool", Dal: &AttributeDal{Type: "Boolean"}}},
					{Name: "meta", Attribute: Attribute{GoType: "types.TemplateMeta", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "template", Attribute: Attribute{Sortable: true, GoType: "string", Dal: &AttributeDal{}}},
					{Name: "created_at", Attribute: sysTimestampNowField},
					{Name: "updated_at", Attribute: sysTimestampNilField},
					{Name: "deleted_at", Attribute: sysTimestampNilField},
					{Name: "last_used_at", Attribute: sysTimestampNilField},
				},
				Indexes: map[string]Index{
					"primary": {Attribute: "id"},
					"unique_language_handle": {
						Fields: []IndexField{
							{Attribute: "language"},
							{Attribute: "handle"},
						},
						Predicate: "handle != '' AND deleted_at IS NULL",
					},
				},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"template_id": {GoType: "[]uint64", Ident: "templateID", StoreIdent: "id"},
					"handle":      {},
					"type":        {GoType: "string"},
					"owner_id":    {GoType: "uint64", StoreIdent: "rel_owner", Ident: "ownerID"},
					"partial":     {GoType: "bool"},
					"deleted":     {GoType: "filter.State", StoreIdent: "deleted_at"},
				},
				Query:      []string{"handle", "type"},
				ByValue:    []string{"template_id", "handle", "partial", "type", "owner_id"},
				ByNilState: []string{"deleted"},
			},
			Rbac: &Rbac{Operations: map[string]RbacOperation{
				"read":   {Description: "Read template"},
				"update": {Description: "Update template"},
				"delete": {Description: "Delete template"},
				"render": {Description: "Render template"},
			}},
			Store: &StoreConfig{
				Lookups: []StoreLookup{
					{Fields: []string{"id"}, Description: "searches for template by ID\n\nIt also returns deleted templates."},
					{Fields: []string{"handle"}, NullConstraint: []string{"deleted_at"}, ConstraintCheck: true, Description: "searches for template by handle\n\nIt returns only valid templates (not deleted)"},
				},
			},
		}},

		{Handle: "user", Resource: Resource{
			Model: Model{
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: sysIDField},
					{Name: "email", Attribute: Attribute{Sortable: true, Unique: true, IgnoreCase: true, Dal: &AttributeDal{Length: 254}}},
					{Name: "email_confirmed", Attribute: Attribute{GoType: "bool", Dal: &AttributeDal{Type: "Boolean"}}},
					{Name: "user_group_id", Attribute: Attribute{Ident: "userGroupID", GoType: "uint64", StoreIdent: "rel_user_group", Dal: &AttributeDal{Type: "Ref", RefModelResType: "corteza::system:user-group", Nullable: true, HasDefault: true, DefaultValue: 0}}},
					{Name: "username", Attribute: Attribute{Sortable: true, Unique: true, IgnoreCase: true, Dal: &AttributeDal{}}},
					{Name: "roles", Attribute: Attribute{GoType: "[]uint64", NoStore: true, OmitGetter: true, OmitSetter: true}},
					{Name: "name", Attribute: Attribute{Sortable: true, Dal: &AttributeDal{}}},
					{Name: "handle", Attribute: sysHandleField},
					{Name: "kind", Attribute: Attribute{Sortable: true, GoType: "types.UserKind", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Length: 8}}},
					{Name: "meta", Attribute: Attribute{GoType: "*types.UserMeta", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "suspended_at", Attribute: sysTimestampNilField},
					{Name: "created_at", Attribute: sysTimestampNowField},
					{Name: "updated_at", Attribute: sysTimestampNilField},
					{Name: "deleted_at", Attribute: sysTimestampNilField},
				},
				Indexes: map[string]Index{
					"primary":         {Attribute: "id"},
					"unique_email":    {Fields: []IndexField{{Attribute: "email", Modifiers: []string{"LOWERCASE"}}}, Predicate: "email != '' AND deleted_at IS NULL"},
					"unique_handle":   {Fields: []IndexField{{Attribute: "handle", Modifiers: []string{"LOWERCASE"}}}, Predicate: "handle != '' AND deleted_at IS NULL"},
					"unique_username": {Fields: []IndexField{{Attribute: "username", Modifiers: []string{"LOWERCASE"}}}, Predicate: "username != '' AND deleted_at IS NULL"},
				},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"user_id":       {GoType: "[]uint64", Ident: "userID", StoreIdent: "id"},
					"role_id":       {GoType: "[]uint64", Ident: "roleID"},
					"user_group_id": {GoType: "uint64", Ident: "userGroupID"},
					"email":         {GoType: "string"},
					"name":          {},
					"username":      {},
					"handle":        {},
					"kind":          {GoType: "types.UserKind"},
					"allKinds":      {GoType: "bool"},
					"deleted":       {GoType: "filter.State", StoreIdent: "deleted_at"},
					"suspended":     {GoType: "filter.State", StoreIdent: "suspended_at"},
				},
				Query:      []string{"email", "username", "handle", "name"},
				ByValue:    []string{"user_id", "email", "username", "handle"},
				ByNilState: []string{"deleted", "suspended"},
			},
			Rbac: &Rbac{Operations: map[string]RbacOperation{
				"read":               {Description: "Read user"},
				"update":             {Description: "Update user"},
				"delete":             {Description: "Delete user"},
				"suspend":            {Description: "Suspend user"},
				"unsuspend":          {Description: "Unsuspend user"},
				"email.unmask":       {Description: "Unmask email"},
				"name.unmask":        {Description: "Unmask name"},
				"impersonate":        {Description: "Impersonate user"},
				"credentials.manage": {Description: "Manage user's credentials"},
			}},
			Store: &StoreConfig{
				Lookups: []StoreLookup{
					{Fields: []string{"id"}, Description: "searches for user by ID\n\nIt returns user even if deleted or suspended"},
					{Fields: []string{"email"}, NullConstraint: []string{"deleted_at"}, ConstraintCheck: true, Description: "searches for user by email\n\nIt returns only valid user (not deleted, not suspended)"},
					{Fields: []string{"handle"}, NullConstraint: []string{"deleted_at"}, ConstraintCheck: true, Description: "searches for user by handle\n\nIt returns only valid user (not deleted, not suspended)"},
					{Fields: []string{"username"}, NullConstraint: []string{"deleted_at"}, ConstraintCheck: true, Description: "searches for user by username\n\nIt returns only valid user (not deleted, not suspended)"},
				},
				Functions: []StoreFunction{
					{ExpIdent: "CountUsers", Args: []StoreFunctionArg{{Ident: "u", GoType: "types.UserFilter"}}, Return: []string{"uint"}},
					{ExpIdent: "UserMetrics", Return: []string{"*types.UserMetrics"}},
				},
			},
		}},

		{Handle: "dal-connection", Resource: Resource{
			Features: Features{Labels: boolPtr(false)},
			Model: Model{
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: sysIDField},
					{Name: "handle", Attribute: sysHandleField},
					{Name: "type", Attribute: Attribute{Sortable: true, Dal: &AttributeDal{}}},
					{Name: "config", Attribute: Attribute{GoType: "types.ConnectionConfig", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "meta", Attribute: Attribute{GoType: "types.ConnectionMeta", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "created_at", Attribute: sysTimestampNowField},
					{Name: "updated_at", Attribute: sysTimestampNilField},
					{Name: "deleted_at", Attribute: sysTimestampNilField},
					{Name: "created_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "updated_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "deleted_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
				},
				Indexes: map[string]Index{"primary": {Attribute: "id"}},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"dal_connection_id": {GoType: "[]uint64", Ident: "dalConnectionID", StoreIdent: "id"},
					"handle":            {},
					"type":              {},
					"deleted":           {GoType: "filter.State", StoreIdent: "deleted_at"},
				},
				ByValue:    []string{"dal_connection_id", "handle", "type"},
				ByNilState: []string{"deleted"},
			},
			Rbac: &Rbac{Operations: map[string]RbacOperation{
				"read":              {Description: "Read connection"},
				"update":            {Description: "Update connection"},
				"delete":            {Description: "Delete connection"},
				"dal-config.manage": {Description: "Manage DAL configuration"},
			}},
			Store: &StoreConfig{
				Lookups: []StoreLookup{
					{Fields: []string{"id"}, Description: "searches for connection by ID\n\nIt returns connection even if deleted or suspended"},
					{Fields: []string{"handle"}, NullConstraint: []string{"deleted_at"}, ConstraintCheck: true, Description: "searches for connection by handle\n\nIt returns only valid connection (not deleted)"},
				},
			},
		}},

		{Handle: "dal-sensitivity-level", Resource: Resource{
			Features: Features{Labels: boolPtr(false)},
			Model: Model{
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: sysIDField},
					{Name: "handle", Attribute: sysHandleField},
					{Name: "level", Attribute: Attribute{Sortable: true, GoType: "int", Dal: &AttributeDal{Type: "Number", Meta: map[string]interface{}{"rdbms:type": "integer"}}}},
					{Name: "meta", Attribute: Attribute{GoType: "types.DalSensitivityLevelMeta", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "created_at", Attribute: sysTimestampNowField},
					{Name: "updated_at", Attribute: sysTimestampNilField},
					{Name: "deleted_at", Attribute: sysTimestampNilField},
					{Name: "created_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "updated_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "deleted_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
				},
				Indexes: map[string]Index{"primary": {Attribute: "id"}},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"dal_sensitivity_level_id": {GoType: "[]uint64", Ident: "dalSensitivityLevelID", StoreIdent: "id"},
					"handle":                   {},
					"deleted":                  {GoType: "filter.State", StoreIdent: "deleted_at"},
				},
				ByValue:    []string{"dal_sensitivity_level_id", "handle"},
				ByNilState: []string{"deleted"},
			},
			Store: &StoreConfig{
				Lookups: []StoreLookup{{Fields: []string{"id"}, Description: "searches for user by ID\n\nIt returns user even if deleted or suspended"}},
			},
		}},

		{Handle: "dal-schema-alteration", Resource: Resource{
			Features: Features{Labels: boolPtr(false), CheckFn: boolPtr(false)},
			Model: Model{
				Attributes: []NamedAttribute{
					{Name: "id", Attribute: sysIDField},
					{Name: "batchID", Attribute: Attribute{GoType: "uint64", StoreIdent: "batch_id", Dal: &AttributeDal{Type: "ID"}}},
					{Name: "dependsOn", Attribute: Attribute{GoType: "uint64", StoreIdent: "depends_on", Dal: &AttributeDal{Type: "Ref", RefModelResType: "corteza::system:dal-schema-alteration"}}},
					{Name: "resource", Attribute: Attribute{StoreIdent: "resource", Dal: &AttributeDal{Type: "Text", Length: 256}}},
					{Name: "resourceType", Attribute: Attribute{StoreIdent: "resource_type", Dal: &AttributeDal{Type: "Text", Length: 256}}},
					{Name: "connectionID", Attribute: Attribute{GoType: "uint64", StoreIdent: "connection_id", Dal: &AttributeDal{Type: "Ref", RefModelResType: "corteza::system:dal-connection"}}},
					{Name: "kind", Attribute: Attribute{Dal: &AttributeDal{Type: "Text", Length: 256}}},
					{Name: "params", Attribute: Attribute{GoType: "*types.DalSchemaAlterationParams", OmitGetter: true, OmitSetter: true, Dal: &AttributeDal{Type: "JSON", DefaultEmptyObject: true}}},
					{Name: "error", Attribute: Attribute{Dal: &AttributeDal{Type: "Text"}}},
					{Name: "created_at", Attribute: sysTimestampNowField},
					{Name: "updated_at", Attribute: sysTimestampNilField},
					{Name: "deleted_at", Attribute: sysTimestampNilField},
					{Name: "completed_at", Attribute: sysTimestampNilField},
					{Name: "dismissed_at", Attribute: sysTimestampNilField},
					{Name: "created_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "updated_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "deleted_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "completed_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
					{Name: "dismissed_by", Attribute: Attribute{GoType: "uint64", Dal: userRefDal}},
				},
				Indexes: map[string]Index{
					"primary": {Attribute: "id"},
					"unique_alteration": {
						Fields: []IndexField{{Attribute: "id"}, {Attribute: "batchID"}},
					},
				},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"resource":      {GoType: "[]string", Ident: "resource"},
					"resourceType":  {GoType: "string", Ident: "resourceType", StoreIdent: "resource_type"},
					"alteration_id": {GoType: "[]uint64", Ident: "alterationID", StoreIdent: "id"},
					"batch_id":      {GoType: "[]uint64", Ident: "batchID"},
					"kind":          {},
					"deleted":       {GoType: "filter.State", StoreIdent: "deleted_at"},
					"completed":     {GoType: "filter.State", StoreIdent: "completed_at"},
					"dismissed":     {GoType: "filter.State", StoreIdent: "dismissed_at"},
				},
				ByValue:    []string{"kind", "resource", "resourceType", "alteration_id", "batch_id"},
				ByNilState: []string{"deleted", "completed", "dismissed"},
			},
			Store: &StoreConfig{
				Lookups: []StoreLookup{{Fields: []string{"id"}, Description: "searches for resource translation by ID\nIt also returns deleted resource translations."}},
			},
		}},
	},
}
