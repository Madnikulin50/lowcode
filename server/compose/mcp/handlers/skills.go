package handlers

import (
	"context"
	"strings"

	automationService "github.com/madnikulin50/lowcode/server/automation/service"
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Skill tools. A skill is a named, versioned set of instructions (with
// optional reference files) that an agent loads when a task calls for it. They
// live in the prompt library next to prompts, with the same versioning, and
// are exchanged as SKILL.md.
func initSkills(ctx context.Context, s *server.MCPServer) {
	s.AddTool(mcp.NewTool("skill_list",
		mcp.WithDescription("List the skills agents can use: the library's (name, what it is for, active version, files) and those loaded from files or remote agents"),
	), handleSkillList)

	s.AddTool(mcp.NewTool("skill_get",
		mcp.WithDescription("Show a skill's instructions and the files it carries. Without a version: the active one."),
		mcp.WithString("handle", mcp.Description("Skill handle"), mcp.Required()),
		mcp.WithNumber("version", mcp.Description("Version number")),
		mcp.WithBoolean("files", mcp.Description("Include the content of the files, not only their names")),
	), handleSkillGet)

	s.AddTool(mcp.NewTool("skill_history",
		mcp.WithDescription("List all versions of a skill, oldest first, with the note saved with each"),
		mcp.WithString("handle", mcp.Description("Skill handle"), mcp.Required()),
	), handleSkillHistory)

	s.AddTool(mcp.NewTool("skill_save",
		mcp.WithDescription("Save a new version of a skill (creating it if the handle is new). The description says WHEN to use the skill - agents choose skills by it. The new version becomes the one in use only if activate=true (the first always is)."),
		mcp.WithString("handle", mcp.Description("Lowercase letters, digits, _ - . ; starts with a letter"), mcp.Required()),
		mcp.WithString("text", mcp.Description("The instructions, in markdown"), mcp.Required()),
		mcp.WithString("description", mcp.Description("When to use the skill (kept from the previous version if omitted)")),
		mcp.WithString("requires", mcp.Description(`Toolkits the skill needs, e.g. "compose.records,workflows" or a JSON array. Omit to keep the previous version's; "[]" clears them.`)),
		mcp.WithString("resources", mcp.Description(`Reference files as a JSON array, e.g. [{"path":"prices.csv","content":"..."}]. Text only. Omit to keep the previous version's; "[]" clears them.`)),
		mcp.WithString("note", mcp.Description("Why this version exists")),
		mcp.WithBoolean("activate", mcp.Description("Make this the version in use")),
	), handleSkillSave)

	s.AddTool(mcp.NewTool("skill_activate",
		mcp.WithDescription("Make a version the one in use - to roll back, or forward to one saved without activation"),
		mcp.WithString("handle", mcp.Description("Skill handle"), mcp.Required()),
		mcp.WithNumber("version", mcp.Description("Version number"), mcp.Required()),
	), handleSkillActivate)

	s.AddTool(mcp.NewTool("skill_delete",
		mcp.WithDescription("Delete a skill with all its versions"),
		mcp.WithString("handle", mcp.Description("Skill handle"), mcp.Required()),
	), handleSkillDelete)

	s.AddTool(mcp.NewTool("skill_import",
		mcp.WithDescription("Save a skill written as SKILL.md: a YAML header (name, description, optional requires) followed by the instructions"),
		mcp.WithString("content", mcp.Description("The SKILL.md text"), mcp.Required()),
		mcp.WithString("handle", mcp.Description("Use this handle instead of the name in the header")),
		mcp.WithBoolean("activate", mcp.Description("Make the imported version the one in use")),
	), handleSkillImport)

	s.AddTool(mcp.NewTool("skill_export",
		mcp.WithDescription("Return a skill as SKILL.md text (a skill with files is exported as a zip from the admin screen)"),
		mcp.WithString("handle", mcp.Description("Skill handle"), mcp.Required()),
		mcp.WithNumber("version", mcp.Description("Version number; default: the active one")),
	), handleSkillExport)
}

func handleSkillList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)

	// the catalog an agent sees: library, files and remote agents together
	catalog := aiagent.ListSkills(ctx)
	if len(catalog) == 0 {
		return textResult("No skills found"), nil
	}
	infos := make([]aiagent.SkillInfo, 0, len(catalog))
	for _, s := range catalog {
		infos = append(infos, s.Info())
	}
	return jsonResult(map[string]interface{}{"skills": infos}), nil
}

func handleSkillGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)

	withFiles, _ := args["files"].(bool)
	handle := strings.TrimSpace(getString(args, "handle"))
	s, err := automationService.GetSkill(ctx, handle, getInt(args, "version"), withFiles)
	if err != nil {
		// not in the library: a file or a remote agent may have it
		if fs, ferr := aiagent.GetSkill(ctx, handle, getInt(args, "version")); ferr == nil {
			return jsonResult(fs), nil
		}
		return errorResult(err), nil
	}
	return jsonResult(s), nil
}

func handleSkillHistory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	h, err := automationService.SkillHistory(ctx, strings.TrimSpace(getString(argsMap(request), "handle")))
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(map[string]interface{}{"versions": h}), nil
}

func handleSkillSave(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)

	requires, err := automationService.ParseSkillRequires(getString(args, "requires"))
	if err != nil {
		return errorResult(err), nil
	}
	files, err := automationService.ParseSkillFiles(getString(args, "resources"))
	if err != nil {
		return errorResult(err), nil
	}
	activate, _ := args["activate"].(bool)

	s, err := automationService.SaveSkillVersion(ctx, automationService.SavePrompt{
		Handle:      strings.TrimSpace(getString(args, "handle")),
		Text:        getString(args, "text"),
		Description: getString(args, "description"),
		Note:        getString(args, "note"),
		Requires:    requires,
		Resources:   files,
		Activate:    activate,
	})
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(s), nil
}

func handleSkillActivate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)

	version := getInt(args, "version")
	if version < 1 {
		return textResult("version is required"), nil
	}
	s, err := automationService.ActivateSkill(ctx, strings.TrimSpace(getString(args, "handle")), version)
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(s), nil
}

func handleSkillDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	handle := strings.TrimSpace(getString(argsMap(request), "handle"))
	if err := automationService.DeleteSkill(ctx, handle); err != nil {
		return errorResult(err), nil
	}
	return textResult("Deleted skill " + handle), nil
}

func handleSkillImport(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)

	activate, _ := args["activate"].(bool)
	s, err := automationService.ImportSkill(ctx, []byte(getString(args, "content")), getString(args, "handle"), "", activate)
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(s), nil
}

func handleSkillExport(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)

	s, err := automationService.GetSkill(ctx, strings.TrimSpace(getString(args, "handle")), getInt(args, "version"), true)
	if err != nil {
		return errorResult(err), nil
	}
	skill := aiagent.Skill{Handle: s.Handle, Description: s.Description, Body: s.Text, Requires: s.Requires}
	md, err := aiagent.RenderSkillMarkdown(skill)
	if err != nil {
		return errorResult(err), nil
	}
	out := string(md)
	if len(s.Resources) > 0 {
		out += "\n\n(this skill also has files: " + strings.Join(fileNames(s.Resources), ", ") + " - export it as a zip to get them)"
	}
	return textResult(out), nil
}

func fileNames(ff []automationService.SkillFileBrief) []string {
	out := make([]string, 0, len(ff))
	for _, f := range ff {
		out = append(out, f.Path)
	}
	return out
}
