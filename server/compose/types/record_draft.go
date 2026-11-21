package types

import (
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/filter"
)

type (
	// RecordDraft is a draft of a record edit that its author keeps on the
	// server, so it follows them between browsers and devices. Drafts are
	// private: they are only ever read, changed and deleted by their author.
	//
	// ChangeID is chosen by the client (it is the ID the client's draft store
	// knows the draft by); it is unique per author. Payload is the draft as the
	// client describes it, as JSON.
	RecordDraft struct {
		ID uint64 `json:"-" schema:"col=id,dal=id,unique"`

		UserID   uint64 `json:"-" schema:"col=user_id,store=rel_user,dal=userref"`
		ChangeID string `json:"changeID" schema:"col=change_id,dal=text:64"`

		NamespaceID uint64 `json:"namespaceID,string" schema:"col=namespace_id,store=rel_namespace,dal=ref:corteza::compose:namespace"`
		ModuleID    uint64 `json:"moduleID,string" schema:"col=module_id,dal=id"`

		// RecordID is 0 for a draft of a record that does not exist yet
		RecordID uint64 `json:"recordID,string" schema:"col=record_id,dal=id"`

		Payload string `json:"-" schema:"col=payload,dal"`

		CreatedAt time.Time  `json:"createdAt,omitempty" schema:"col=created_at,dal=timestamp:now,sortable"`
		UpdatedAt *time.Time `json:"updatedAt,omitempty" schema:"col=updated_at,dal=timestamp:nil,sortable"`
	}

	RecordDraftFilter struct {
		UserID      uint64 `json:"-"`
		ChangeID    string `json:"changeID"`
		NamespaceID uint64 `json:"namespaceID,string"`
		ModuleID    uint64 `json:"moduleID,string"`
		RecordID    uint64 `json:"recordID,string"`

		// Check fn is called by store backend for each resource found function can
		// modify the resource and return false if store should not return it
		//
		// Store then loads additional resources to satisfy the paging parameters
		Check func(*RecordDraft) (bool, error) `json:"-"`

		// Standard helpers for paging and sorting
		filter.Sorting
		filter.Paging
	}
)
