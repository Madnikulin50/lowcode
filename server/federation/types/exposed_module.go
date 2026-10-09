package types

import (
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/filter"
)

type (
	ExposedModule struct {
		ID                 uint64         `json:"moduleID,string" schema:"col=id,dal=id,unique"`
		NodeID             uint64         `json:"nodeID,string" schema:"col=node_id,store=rel_node,dal=ref:corteza::federation:node:default0,sortable"`
		ComposeModuleID    uint64         `json:"composeModuleID,string" schema:"col=compose_module_id,store=rel_compose_module,dal=id"`
		ComposeNamespaceID uint64         `json:"composeNamespaceID,string" schema:"col=compose_namespace_id,store=rel_compose_namespace,dal=id"`
		Handle             string         `json:"handle" schema:"col=handle,dal=text:64,unique,ignoreCase"`
		Name               string         `json:"name" schema:"col=name,dal,sortable"`
		Fields             ModuleFieldSet `json:"fields" schema:"col=fields,omit,dal=json:empty"`

		CreatedAt time.Time  `json:"createdAt,omitempty" schema:"col=created_at,dal=timestamp:now,sortable"`
		CreatedBy uint64     `json:"createdBy,string" schema:"col=created_by,dal=userref"`
		UpdatedAt *time.Time `json:"updatedAt,omitempty" schema:"col=updated_at,dal=timestamp:nil,sortable"`
		UpdatedBy uint64     `json:"updatedBy,string,omitempty" schema:"col=updated_by,dal=userref"`
		DeletedAt *time.Time `json:"deletedAt,omitempty" schema:"col=deleted_at,dal=timestamp:nil,sortable"`
		DeletedBy uint64     `json:"deletedBy,string,omitempty" schema:"col=deleted_by,dal=userref"`
	}

	ExposedModuleFilter struct {
		NodeID             uint64 `json:"nodeID,string"`
		ComposeModuleID    uint64 `json:"composeModuleID,string"`
		ComposeNamespaceID uint64 `json:"composeNamespaceID,string"`

		LastSync uint64 `json:"lastSync"`
		Handle   string `json:"handle"`
		Name     string `json:"name"`
		Query    string `json:"query"`

		Check func(*ExposedModule) (bool, error) `json:"-"`

		filter.Sorting
		filter.Paging
	}
)
