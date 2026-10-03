package types

import (
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/filter"
)

const (
	DocumentKindMarkdown = "markdown"
	DocumentKindPDF      = "pdf"
)

type (
	Document struct {
		ID           uint64 `json:"documentID,string"`
		NamespaceID  uint64 `json:"namespaceID,string"`
		Handle       string `json:"handle,omitempty"`
		Title        string `json:"title"`
		Kind         string `json:"kind"`
		Body         string `json:"body,omitempty"`
		AttachmentID uint64 `json:"attachmentID,string,omitempty"`
		FileName     string `json:"fileName,omitempty"`
		Weight       int    `json:"weight"`
		Visible      bool   `json:"visible"`

		CreatedAt time.Time  `json:"createdAt,omitempty"`
		UpdatedAt *time.Time `json:"updatedAt,omitempty"`
		DeletedAt *time.Time `json:"deletedAt,omitempty"`
	}

	DocumentFilter struct {
		NamespaceID uint64 `json:"namespaceID,string"`
		Handle      string `json:"handle,omitempty"`
		Query       string `json:"query,omitempty"`
		// VisibleOnly drops unpublished documents.
		VisibleOnly bool `json:"-"`

		Deleted filter.State `json:"deleted"`

		filter.Sorting
		filter.Paging
	}

	DocumentSet []*Document
)
