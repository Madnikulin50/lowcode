package def

import (
	automationtypes "github.com/madnikulin50/lowcode/server/automation/types"
)

// Automation is the Go port of automation/component.cue plus its three
// resource files (workflow.cue, session.cue, trigger.cue), in the same
// order as automation/component.cue's `resources: {...}` declaration.
var Automation = Component{
	Handle: "automation",

	Resources: []NamedResource{
		{Handle: "workflow", Resource: Resource{
			Model: Model{
				Ident:      "automation_workflows",
				Attributes: AttributesFromStruct(automationtypes.Workflow{}),
				Indexes: map[string]Index{
					"primary": {Attribute: "id"},
				},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"workflow_id":  {GoType: "[]string", Ident: "workflowID", StoreIdent: "id"},
					"handle":       {},
					"sub_workflow": {GoType: "filter.State"},
					"deleted":      {GoType: "filter.State", StoreIdent: "deleted_at"},
					"disabled":     {GoType: "filter.State", StoreIdent: "enabled"},
				},
				Query:        []string{"handle"},
				ByValue:      []string{"workflow_id", "handle"},
				ByNilState:   []string{"deleted"},
				ByFalseState: []string{"disabled"},
			},
			Rbac: &Rbac{Operations: map[string]RbacOperation{
				"read":            {Description: "Read workflow"},
				"update":          {Description: "Update workflow"},
				"delete":          {Description: "Delete workflow"},
				"undelete":        {Description: "Undelete workflow"},
				"execute":         {Description: "Execute workflow"},
				"triggers.manage": {Description: "Manage workflow triggers"},
				"sessions.manage": {Description: "Manage workflow sessions"},
			}},
			Store: &StoreConfig{
				Ident: "automationWorkflow",
				Lookups: []StoreLookup{
					{Fields: []string{"id"}, Description: "searches for workflow by ID\n\nIt returns workflow even if deleted"},
					{Fields: []string{"handle"}, NullConstraint: []string{"deleted_at"}, ConstraintCheck: true, Description: "searches for workflow by their handle\n\nIt returns only valid workflows"},
				},
			},
		}},

		{Handle: "session", Resource: Resource{
			Model: Model{
				Ident:      "automation_sessions",
				Attributes: AttributesFromStruct(automationtypes.Session{}),
				Indexes: map[string]Index{
					"primary":       {Attribute: "id"},
					"completed_at":  {Attribute: "completed_at"},
					"created_at":    {Attribute: "created_at"},
					"event_type":    {Attribute: "event_type"},
					"resource_type": {Attribute: "resource_type"},
					"status":        {Attribute: "status"},
					"suspended_at":  {Attribute: "suspended_at"},
				},
			},
			Features: Features{Labels: boolPtr(false)},
			Filter: Filter{
				Struct: map[string]Attribute{
					"session_id":    {GoType: "[]string", StoreIdent: "id", Ident: "sessionID"},
					"completed":     {GoType: "*time.Time", StoreIdent: "completed_at"},
					"created_by":    {GoType: "[]string"},
					"status":        {GoType: "[]uint"},
					"workflow_id":   {GoType: "[]string"},
					"event_type":    {},
					"resource_type": {},
				},
				ByValue:    []string{"status", "session_id", "workflow_id", "event_type", "resource_type", "created_by"},
				ByNilState: []string{"completed"},
			},
			// session.cue defines no rbac block -> excluded from
			// rbac/types_rbac generation.
			Store: &StoreConfig{
				Ident: "automationSession",
				Lookups: []StoreLookup{
					{Fields: []string{"id"}, Description: "searches for session by ID\n\nIt returns session even if deleted"},
				},
			},
		}},

		{Handle: "trigger", Resource: Resource{
			Model: Model{
				Ident:      "automation_triggers",
				Attributes: AttributesFromStruct(automationtypes.Trigger{}),
				Indexes: map[string]Index{
					"primary": {Attribute: "id"},
				},
			},
			Filter: Filter{
				Struct: map[string]Attribute{
					"deleted":       {GoType: "filter.State", StoreIdent: "deleted_at"},
					"disabled":      {GoType: "filter.State", StoreIdent: "enabled"},
					"trigger_id":    {GoType: "[]uint64", Ident: "triggerID", StoreIdent: "id"},
					"workflow_id":   {GoType: "[]uint64"},
					"event_type":    {},
					"resource_type": {},
				},
				ByValue:      []string{"trigger_id", "workflow_id", "event_type", "resource_type"},
				ByNilState:   []string{"deleted"},
				ByFalseState: []string{"disabled"},
			},
			// trigger.cue defines no rbac block -> excluded from
			// rbac/types_rbac generation.
			Store: &StoreConfig{
				Ident: "automationTrigger",
				Lookups: []StoreLookup{
					{Fields: []string{"id"}, Description: "searches for trigger by ID\n\nIt returns trigger even if deleted"},
				},
			},
		}},
	},
}
