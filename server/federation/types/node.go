package types

import (
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/filter"
)

var (
	NodeStatusPending       = "pending"
	NodeStatusPairRequested = "pair_requested"
	NodeStatusPaired        = "paired"
	NodeStatusFailed        = "failed"
)

type (
	// Node's `schema` tags are the pilot for a single source of truth
	// between the hand-written struct and codegen/def (see
	// codegen/def/reflectattr.go and codegen/def/federation.go's node
	// resource): the field, its Go type, and its schema/dal metadata are
	// declared once, here, instead of being duplicated as a separate
	// Attribute{} literal that can silently drift from this struct.
	Node struct {
		ID     uint64 `json:"nodeID,string" schema:"col=id,dal=id,unique"`
		Name   string `json:"name" schema:"col=name,dal,sortable"`
		Status string `json:"status" schema:"col=status,dal,sortable"`

		// Base URL of the remote server
		BaseURL string `json:"baseURL" schema:"col=base_url,dal,sortable"`

		Contact string `json:"contact" schema:"col=contact,dal,sortable"`

		// Node ID on the remote server that points back to us
		SharedNodeID uint64 `json:"sharedNodeID,string" schema:"col=shared_node_id,dal=id,sortable"`

		PairToken string `json:"-" schema:"col=pair_token,dal"`
		AuthToken string `json:"-" schema:"col=auth_token,dal"`

		CreatedAt time.Time  `json:"createdAt,omitempty" schema:"col=created_at,dal=timestamp:now,sortable"`
		CreatedBy uint64     `json:"createdBy,string" schema:"col=created_by,dal=userref"`
		UpdatedAt *time.Time `json:"updatedAt,omitempty" schema:"col=updated_at,dal=timestamp:nil,sortable"`
		UpdatedBy uint64     `json:"updatedBy,string,omitempty" schema:"col=updated_by,dal=userref"`
		DeletedAt *time.Time `json:"deletedAt,omitempty" schema:"col=deleted_at,dal=timestamp:nil,sortable"`
		DeletedBy uint64     `json:"deletedBy,string,omitempty" schema:"col=deleted_by,dal=userref"`
	}

	NodeFilter struct {
		Query  string `json:"name"`
		Status string `json:"status"`

		// Check fn is called by store backend for each resource found function can
		// modify the resource and return false if store should not return it
		//
		// Store then loads additional resources to satisfy the paging parameters
		Check func(node *Node) (bool, error) `json:"-"`

		Deleted filter.State `json:"deleted"`

		// Standard helpers for paging and sorting
		filter.Sorting
		filter.Paging
	}
)
