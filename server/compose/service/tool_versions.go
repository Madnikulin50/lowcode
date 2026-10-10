package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// VersionToolDefs are the tools for the version history of pages, modules and
// charts, shared by the assistant toolkit (compose.schema) and MCP. Restoring
// changes the resource, so it is Mutating (asks for the user's "да").
func VersionToolDefs() []Def {
	kindParam := Param{Name: "kind", Type: "string", Required: true, Description: "page, module or chart"}
	idParams := []Param{
		{Name: "namespaceID", Type: "string", Required: true, Description: "Namespace ID"},
		kindParam,
		{Name: "resourceID", Type: "string", Required: true, Description: "ID of the page, module or chart"},
	}

	return []Def{
		{
			Name:        "list_resource_versions",
			Description: "List the saved versions (change history, newest first) of a page, module or chart",
			Params:      idParams,
			Handler: func(ctx context.Context, args map[string]string) (any, error) {
				set, err := DefaultResourceVersion.List(ctx, parseUint64(args["namespaceID"]), strings.TrimSpace(args["kind"]), parseUint64(args["resourceID"]))
				if err != nil {
					return nil, err
				}
				if len(set) == 0 {
					return "No versions found", nil
				}
				return set, nil
			},
		},
		{
			Name:        "list_file_versions",
			Description: "List the versions (newest first) of a file in a File field of a record. Only fields with versioning enabled have versions; the result tells which one the record uses now.",
			Params: []Param{
				{Name: "namespaceID", Type: "string", Required: true, Description: "Namespace ID"},
				{Name: "moduleID", Type: "string", Required: true, Description: "Module ID"},
				{Name: "recordID", Type: "string", Required: true, Description: "Record ID"},
				{Name: "fieldName", Type: "string", Required: true, Description: "Name of the File field"},
				{Name: "attachmentID", Type: "string", Required: true, Description: "ID of the file (any of its versions), as stored in the record"},
			},
			Handler: func(ctx context.Context, args map[string]string) (any, error) {
				vv, err := DefaultAttachment.ListVersions(ctx,
					parseUint64(args["namespaceID"]), parseUint64(args["moduleID"]), parseUint64(args["recordID"]),
					strings.TrimSpace(args["fieldName"]), parseUint64(args["attachmentID"]))
				if err != nil {
					return nil, err
				}

				type row struct {
					AttachmentID string `json:"attachmentID"`
					Version      int    `json:"version"`
					Name         string `json:"name"`
					Size         int64  `json:"size"`
					CreatedAt    string `json:"createdAt"`
					Current      bool   `json:"current"`
				}

				current := map[uint64]bool{}
				for _, id := range vv.Current {
					current[id] = true
				}

				out := make([]row, 0, len(vv.Set))
				for _, a := range vv.Set {
					out = append(out, row{
						AttachmentID: strconv.FormatUint(a.ID, 10),
						Version:      a.Version,
						Name:         a.Name,
						Size:         a.Meta.Original.Size,
						CreatedAt:    a.CreatedAt.Format("2006-01-02 15:04:05"),
						Current:      current[a.ID],
					})
				}
				return out, nil
			},
		},
		{
			Name:        "restore_resource_version",
			Description: "Restore a page, module or chart to a saved version. The result is saved as a new version, so nothing is lost. Restoring a module removes fields that did not exist in that version.",
			Mutating:    true,
			Params: append(append([]Param(nil), idParams...),
				Param{Name: "version", Type: "integer", Required: true, Description: "Version number from list_resource_versions"}),
			Handler: func(ctx context.Context, args map[string]string) (any, error) {
				version, err := strconv.Atoi(strings.TrimSpace(args["version"]))
				if err != nil || version <= 0 {
					return nil, fmt.Errorf("version must be a positive number")
				}
				return DefaultResourceVersion.Restore(ctx, parseUint64(args["namespaceID"]), strings.TrimSpace(args["kind"]), parseUint64(args["resourceID"]), version)
			},
		},
	}
}
