package types

import (
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/filter"
)

type (
	// ResourceVersion is one saved snapshot of a compose resource (page,
	// module or chart). Versions are immutable and numbered per resource from
	// 1; restoring one applies its snapshot as a regular update, which in turn
	// saves a new version.
	ResourceVersion struct {
		ID uint64 `json:"versionID,string" schema:"col=id,dal=id,unique"`

		NamespaceID  uint64 `json:"namespaceID,string" schema:"col=namespace_id,store=rel_namespace,dal=ref:corteza::compose:namespace"`
		ResourceType string `json:"resourceType" schema:"col=resource_type,dal=text:16,sortable"`
		ResourceID   uint64 `json:"resourceID,string" schema:"col=resource_id,dal=id"`
		Version      int    `json:"version" schema:"col=version,dal=number:default0,sortable"`

		// Snapshot is the resource serialized as JSON
		Snapshot string `json:"snapshot,omitempty" schema:"col=snapshot,dal"`
		Note     string `json:"note,omitempty" schema:"col=note,dal"`

		CreatedAt time.Time `json:"createdAt,omitempty" schema:"col=created_at,dal=timestamp:now,sortable"`
		CreatedBy uint64    `json:"createdBy,string" schema:"col=created_by,dal=userref"`
	}

	ResourceVersionFilter struct {
		VersionID    []uint64 `json:"versionID"`
		NamespaceID  uint64   `json:"namespaceID,string"`
		ResourceType string   `json:"resourceType"`
		ResourceID   uint64   `json:"resourceID,string"`

		// Check fn is called by store backend for each resource found function can
		// modify the resource and return false if store should not return it
		//
		// Store then loads additional resources to satisfy the paging parameters
		Check func(*ResourceVersion) (bool, error) `json:"-"`

		// Standard helpers for paging and sorting
		filter.Sorting
		filter.Paging
	}
)

const (
	ResourceVersionPage   = "page"
	ResourceVersionModule = "module"
	ResourceVersionChart  = "chart"
)
