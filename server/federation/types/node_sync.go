package types

import (
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/filter"
)

var (
	NodeSyncTypeStructure = "sync_structure"
	NodeSyncTypeData      = "sync_data"
	NodeSyncStatusSuccess = "success"
	NodeSyncStatusError   = "error"
)

type (
	NodeSync struct {
		NodeID     uint64 `json:"nodeID,string" schema:"col=rel_node,dal=id,sortable"`
		ModuleID   uint64 `json:"moduleID,string" schema:"col=rel_module,store=rel_compose_module,dal=id,sortable"`
		SyncStatus string `json:"syncStatus" schema:"col=sync_status,dal,sortable"`
		SyncType   string `json:"syncType" schema:"col=sync_type,dal,sortable"`

		TimeOfAction time.Time `json:"timeOfAction" schema:"col=time_of_action,dal=timestamp,sortable"`
	}

	NodeSyncFilter struct {
		NodeID     uint64 `json:"nodeID"`
		RelNodeID  uint64 `json:"relNodeID"`
		ModuleID   uint64 `json:"moduleID"`
		SyncStatus string `json:"syncStatus"`
		SyncType   string `json:"syncType"`

		Query string `json:"name"`

		Check func(*NodeSync) (bool, error) `json:"-"`

		filter.Sorting
		filter.Paging
	}
)
