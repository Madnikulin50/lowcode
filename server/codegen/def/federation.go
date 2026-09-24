package def

import (
	federationtypes "github.com/madnikulin50/lowcode/server/federation/types"
)

// Federation is the Go port of federation/component.cue plus its five
// resource files (node.cue, node_sync.cue, module_exposed.cue,
// shared_module.cue, module_mapping.cue). Field-for-field it mirrors those
// .cue sources, in the same order (order matters for dal/$component_model
// generation - see NamedAttribute); only attributes/options actually
// consumed by the ported templates (types/resources, types/getters_setters,
// rbac/types_rbac, dal/$component_model, dal/$component_init) are carried
// over so far.
var Federation = Component{
	Handle: "federation",

	Resources: []NamedResource{
		{Handle: "node", Resource: Resource{
			Model: Model{
				Ident: "federation_nodes",
				// Pilot: attributes come from federation/types.Node's own
				// `schema` struct tags (see codegen/def/reflectattr.go)
				// instead of a parallel Attribute{} literal list, so the
				// struct and its schema can't drift apart the way
				// compose/types/page.go's Prompt field once did.
				Attributes: AttributesFromStruct(federationtypes.Node{}),
				Indexes: map[string]Index{
					"primary": {Attribute: "id"},
				},
			},
			Features: Features{Labels: boolPtr(false)},
			Filter: Filter{
				Struct: map[string]Attribute{
					"name":     {},
					"base_url": {},
					"status":   {},
					"deleted":  {GoType: "filter.State", StoreIdent: "deleted_at"},
				},
				Query:      []string{"name", "base_url"},
				ByNilState: []string{"deleted"},
			},
			Rbac: &Rbac{Operations: map[string]RbacOperation{
				"manage":        {Description: "Manage federation node"},
				"module.create": {Description: "Create shared module"},
			}},
			Store: &StoreConfig{
				Ident: "federationNode",
				Lookups: []StoreLookup{
					{Fields: []string{"id"}, Description: "searches for federation node by ID\n\nIt returns federation node"},
					{Fields: []string{"base_url", "shared_node_id"}, Description: "searches for node by shared-node-id and base-url"},
					{Fields: []string{"shared_node_id"}, Description: "searches for node by shared-node-id"},
				},
			},
		}},

		{Handle: "node-sync", Resource: Resource{
			Model: Model{
				Ident:      "federation_nodes_sync",
				Attributes: AttributesFromStruct(federationtypes.NodeSync{}),
				Indexes: map[string]Index{
					"idx_rel_node": {Attribute: "rel_node"},
				},
			},
			Features: Features{Labels: boolPtr(false)},
			Filter: Filter{
				Struct: map[string]Attribute{
					"rel_node":    {},
					"rel_module":  {},
					"sync_status": {},
					"sync_type":   {},
				},
				ByValue: []string{"rel_node", "rel_module", "sync_status", "sync_type"},
			},
			// node_sync.cue defines no rbac block -> excluded from
			// rbac/types_rbac generation.
			Store: &StoreConfig{
				Ident: "federationNodeSync",
				Lookups: []StoreLookup{
					{Fields: []string{"rel_node"}, Description: "searches for sync activity by node ID\n\nIt returns sync activity"},
					{Fields: []string{"rel_node", "rel_module", "sync_type", "sync_status"}, Description: "searches for activity by node, type and status\n\nIt returns sync activity"},
				},
			},
		}},

		{Handle: "exposed-module", Resource: Resource{
			Parents: []ParentRef{{Handle: "node"}},
			Model: Model{
				Ident:      "federation_module_exposed",
				Attributes: AttributesFromStruct(federationtypes.ExposedModule{}),
				Indexes: map[string]Index{
					"primary": {Attribute: "id"},
				},
			},
			Features: Features{Labels: boolPtr(false)},
			Filter: Filter{
				Struct: map[string]Attribute{
					"node_id":              {},
					"compose_module_id":    {},
					"compose_namespace_id": {},
				},
				ByValue: []string{"compose_module_id", "compose_namespace_id", "node_id"},
			},
			Rbac: &Rbac{Operations: map[string]RbacOperation{
				"manage": {Description: "Manage exposed module module"},
			}},
			Store: &StoreConfig{
				Ident: "federationExposedModule",
				Lookups: []StoreLookup{
					{Fields: []string{"id"}, Description: "searches for federation module by ID\n\nIt returns federation module"},
				},
			},
		}},

		{Handle: "shared-module", Resource: Resource{
			Parents: []ParentRef{{Handle: "node"}},
			Model: Model{
				Ident:      "federation_module_shared",
				Attributes: AttributesFromStruct(federationtypes.SharedModule{}),
				Indexes: map[string]Index{
					"primary": {Attribute: "id"},
				},
			},
			Features: Features{Labels: boolPtr(false)},
			Filter: Filter{
				Struct: map[string]Attribute{
					"node_id":                       {},
					"handle":                        {},
					"name":                          {},
					"external_federation_module_id": {},
				},
				Query:   []string{"name", "handle"},
				ByValue: []string{"handle", "node_id", "name", "external_federation_module_id"},
			},
			Rbac: &Rbac{Operations: map[string]RbacOperation{
				"map": {Description: "Map shared module"},
			}},
			Store: &StoreConfig{
				Ident: "federationSharedModule",
				Lookups: []StoreLookup{
					{Fields: []string{"id"}, Description: "searches for shared federation module by ID\n\nIt returns shared federation module"},
				},
			},
		}},

		{Handle: "module-mapping", Resource: Resource{
			Parents: []ParentRef{{Handle: "node"}},
			Model: Model{
				Ident:      "federation_module_mapping",
				Attributes: AttributesFromStruct(federationtypes.ModuleMapping{}),
				Indexes: map[string]Index{
					"unique_module_compose_module": {Attributes: []string{"federation_module_id", "compose_module_id", "compose_namespace_id"}},
				},
			},
			Features: Features{Labels: boolPtr(false)},
			Filter: Filter{
				Struct: map[string]Attribute{
					"compose_module_id":    {},
					"compose_namespace_id": {},
					"federation_module_id": {},
				},
				ByValue: []string{"compose_module_id", "compose_namespace_id", "federation_module_id"},
			},
			// module_mapping.cue defines no rbac block -> excluded from
			// rbac/types_rbac generation.
			Store: &StoreConfig{
				Ident: "federationModuleMapping",
				Lookups: []StoreLookup{
					{Fields: []string{"federation_module_id", "compose_module_id", "compose_namespace_id"}, Description: "searches for module mapping by federation module id and compose module id\n\nIt returns module mapping"},
					{Fields: []string{"federation_module_id"}, Description: "searches for module mapping by federation module id\n\nIt returns module mapping"},
				},
			},
		}},
	},
}

// userRefDal mirrors codegen/schema/model.cue's AttributeUserRef.dal.
var userRefDal = &AttributeDal{Type: "Ref", RefModelResType: "corteza::system:user", HasDefault: true, DefaultValue: 0}
