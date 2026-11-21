package types

import (
	"github.com/madnikulin50/lowcode/server/pkg/filter"
)

type (
	ModuleMapping struct {
		NodeID             uint64                `json:"nodeID,string" schema:"col=node_id,dal=id,unique"`
		FederationModuleID uint64                `json:"federationModuleID,string" schema:"col=federation_module_id,store=rel_federation_module,dal=id,sortable"`
		ComposeModuleID    uint64                `json:"composeModuleID,string" schema:"col=compose_module_id,store=rel_compose_module,dal=id,sortable"`
		ComposeNamespaceID uint64                `json:"composeNamespaceID,string" schema:"col=compose_namespace_id,store=rel_compose_namespace,dal=id,sortable"`
		FieldMapping       ModuleFieldMappingSet `json:"fields" schema:"col=field_mapping,omit,dal=json:empty"`
	}

	ModuleMappingFilter struct {
		NodeID             uint64 `json:"nodeID"`
		ComposeModuleID    uint64 `json:"composeModuleID"`
		ComposeNamespaceID uint64 `json:"composeNamespaceID"`
		FederationModuleID uint64 `json:"federationModuleID"`
		Query              string `json:"query"`

		Check func(*ModuleMapping) (bool, error) `json:"-"`

		filter.Sorting
		filter.Paging
	}
)
