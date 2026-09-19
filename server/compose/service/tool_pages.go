package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/madnikulin50/lowcode/server/compose/types"
)

// PageToolDefs is the single definition of the page CRUD tools, shared by
// chat (a curated subset, see Chat() in chat.go) and MCP (all of them,
// registered verbatim — see mcp/handlers/pages.go).
func PageToolDefs() []Def {
	return []Def{
		{
			Name:        "read_page",
			Description: "Read a page by ID, including its blocks",
			Params: []Param{
				{Name: "namespaceID", Type: "string", Required: true, Description: "Namespace ID"},
				{Name: "pageID", Type: "string", Required: true, Description: "Page ID"},
			},
			Handler: func(ctx context.Context, args map[string]string) (any, error) {
				ns := parseUint64(args["namespaceID"])
				pID := parseUint64(args["pageID"])
				p, err := DefaultPage.FindByID(ctx, ns, pID)
				if err != nil {
					return nil, fmt.Errorf("page not found: %w", err)
				}
				return p, nil
			},
		},
		{
			Name:        "list_pages",
			Description: "List all pages in the current namespace",
			Params: []Param{
				{Name: "namespaceID", Type: "string", Required: true, Description: "Namespace ID"},
			},
			Handler: func(ctx context.Context, args map[string]string) (any, error) {
				ns := parseUint64(args["namespaceID"])
				set, _, err := DefaultPage.Find(ctx, types.PageFilter{NamespaceID: ns})
				if err != nil {
					return nil, fmt.Errorf("failed to list pages: %w", err)
				}
				return set, nil
			},
			Render: renderPageList,
		},
		{
			Name:        "search_pages",
			Description: "Search pages by title",
			Params: []Param{
				{Name: "namespaceID", Type: "string", Required: true, Description: "Namespace ID"},
				{Name: "query", Type: "string", Required: true, Description: "Search query"},
			},
			Handler: func(ctx context.Context, args map[string]string) (any, error) {
				ns := parseUint64(args["namespaceID"])
				set, _, err := DefaultPage.Find(ctx, types.PageFilter{NamespaceID: ns, Query: args["query"]})
				if err != nil {
					return nil, fmt.Errorf("failed to search pages: %w", err)
				}
				return set, nil
			},
		},
		{
			Name:        "create_page",
			Description: "Create a new page",
			Mutating:    true,
			Params: []Param{
				{Name: "namespaceID", Type: "string", Required: true, Description: "Namespace ID"},
				{Name: "title", Type: "string", Required: true, Description: "Page title"},
				{Name: "handle", Type: "string", Required: false, Description: "URL-safe handle"},
				{Name: "description", Type: "string", Required: false, Description: "Page description"},
				{Name: "help", Type: "string", Required: false, Description: "Optional Markdown help shown to users"},
				{Name: "moduleID", Type: "string", Required: false, Description: "Module ID this page is for"},
				{Name: "selfID", Type: "string", Required: false, Description: "Parent page ID (0 for root)"},
				{Name: "blocks", Type: "json", Required: false, Description: `JSON array of page blocks. Example: [{"kind":"RecordList","options":{"fields":["field1","field2"]}}]`},
			},
			Handler: createPageTool,
			Render:  renderPageCreated,
		},
		{
			Name:        "update_page",
			Description: "Update a page's title, handle, description, help, or blocks",
			Mutating:    true,
			Params: []Param{
				{Name: "namespaceID", Type: "string", Required: true, Description: "Namespace ID"},
				{Name: "pageID", Type: "string", Required: true, Description: "Page ID"},
				{Name: "title", Type: "string", Required: false, Description: "New page title"},
				{Name: "handle", Type: "string", Required: false, Description: "New URL-safe handle"},
				{Name: "description", Type: "string", Required: false, Description: "New page description"},
				{Name: "help", Type: "string", Required: false, Description: "Optional Markdown help shown to users"},
				{Name: "blocks", Type: "json", Required: false, Description: "JSON array of page blocks"},
			},
			Handler: func(ctx context.Context, args map[string]string) (any, error) {
				ns := parseUint64(args["namespaceID"])
				pID := parseUint64(args["pageID"])
				p, err := DefaultPage.FindByID(ctx, ns, pID)
				if err != nil {
					return nil, fmt.Errorf("page not found: %w", err)
				}
				if v := args["title"]; v != "" {
					p.Title = v
				}
				if v := args["handle"]; v != "" {
					p.Handle = v
				}
				if v := args["description"]; v != "" {
					p.Description = v
				}
				if v := args["help"]; v != "" {
					p.Config.Help = v
				}
				if v := args["blocks"]; v != "" {
					var blocks types.PageBlocks
					if err := json.Unmarshal([]byte(v), &blocks); err != nil {
						return nil, fmt.Errorf("invalid blocks JSON: %w", err)
					}
					p.Blocks = blocks
				}
				updated, err := DefaultPage.Update(ctx, p)
				if err != nil {
					return nil, fmt.Errorf("failed to update page: %w", err)
				}
				return updated, nil
			},
		},
		{
			Name:        "delete_page",
			Description: "Delete a page by ID",
			Mutating:    true,
			Params: []Param{
				{Name: "namespaceID", Type: "string", Required: true, Description: "Namespace ID"},
				{Name: "pageID", Type: "string", Required: true, Description: "Page ID"},
			},
			Handler: func(ctx context.Context, args map[string]string) (any, error) {
				ns := parseUint64(args["namespaceID"])
				pID := parseUint64(args["pageID"])
				if err := DefaultPage.DeleteByID(ctx, ns, pID, types.PageChildrenOnDeleteForce); err != nil {
					return nil, fmt.Errorf("failed to delete page: %w", err)
				}
				return "Page deleted", nil
			},
		},
	}
}

func createPageTool(ctx context.Context, args map[string]string) (any, error) {
	ns := parseUint64(args["namespaceID"])
	title := args["title"]
	if ns == 0 || title == "" {
		return nil, fmt.Errorf("missing required parameters (namespaceID='%s', title='%s')", args["namespaceID"], title)
	}

	var blocks types.PageBlocks
	if b := args["blocks"]; b != "" {
		if err := json.Unmarshal([]byte(b), &blocks); err != nil {
			return nil, fmt.Errorf("invalid blocks JSON for page '%s': %w", title, err)
		}
	}

	p := &types.Page{
		NamespaceID: ns,
		SelfID:      parseUint64(args["selfID"]),
		ModuleID:    parseUint64(args["moduleID"]),
		Title:       title,
		Handle:      args["handle"],
		Description: args["description"],
		Visible:     true,
		Weight:      0,
		Blocks:      blocks,
		Config:      types.PageConfig{Help: args["help"]},
	}

	created, err := DefaultPage.Create(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("failed to create page '%s': %w", title, err)
	}
	return created, nil
}

func renderPageCreated(data any) string {
	p := data.(*types.Page)
	return fmt.Sprintf("✅ Page '%s' created! ID: %d, Handle: %s", p.Title, p.ID, p.Handle)
}

func renderPageList(data any) string {
	set := data.(types.PageSet)
	if len(set) == 0 {
		return "No pages found in this namespace."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "📄 **Pages (%d):**\n\n", len(set))
	for _, p := range set {
		fmt.Fprintf(&b, "### %s\n", p.Title)
		fmt.Fprintf(&b, "- **ID:** %d\n", p.ID)
		fmt.Fprintf(&b, "- **Handle:** `%s`\n", p.Handle)
		if p.Description != "" {
			fmt.Fprintf(&b, "- **Description:** %s\n", p.Description)
		}
		if p.ModuleID > 0 {
			fmt.Fprintf(&b, "- **ModuleID:** %d\n", p.ModuleID)
		}
		if p.Visible {
			fmt.Fprintf(&b, "- **Visible:** ✅\n")
		}
		fmt.Fprintf(&b, "- **Blocks (%d):**\n", len(p.Blocks))
		for _, blk := range p.Blocks {
			title := blk.Title
			if title == "" {
				title = blk.Kind
			}
			fmt.Fprintf(&b, "  - `%s`", title)
			if blk.Kind != "" && blk.Kind != title {
				fmt.Fprintf(&b, " (_%s_)", blk.Kind)
			}
			if blk.Description != "" {
				fmt.Fprintf(&b, ": %s", blk.Description)
			}
			fmt.Fprintf(&b, "\n")
		}
		fmt.Fprintf(&b, "\n")
	}
	return b.String()
}
