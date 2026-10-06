package service

import (
	"context"
	"strconv"
	"strings"

	automationService "github.com/madnikulin50/lowcode/server/automation/service"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

// chatSkillToolDefs let the assistant author skills: instructions that agents
// load on demand (the list_skills / load_skill tools an agent gets are a
// different thing - they read the catalog, these change it). Saving or
// activating a skill changes what live agents are told to do, so those ask
// for the user's "да".
func chatSkillToolDefs() []chat.ToolDef {
	return []chat.ToolDef{
		{
			Name:        "list_skill_library",
			Description: "List the skills in the skill library: handle, what it is for, active version, number of versions and files",
			Handler: func(ctx context.Context, _ map[string]string) string {
				list, err := automationService.ListSkills(ctx)
				if err != nil {
					return err.Error()
				}
				if len(list) == 0 {
					return "No skills found"
				}
				return chatJSON(list)
			},
		},
		{
			Name:        "get_skill",
			Description: "Show a skill: its instructions, required toolkits and files. Without a version: the active one.",
			Params: []chat.ParamDef{
				{Name: "handle", Type: "string", Required: true, Description: "Skill handle"},
				{Name: "version", Type: "integer", Description: "Version number"},
				{Name: "files", Type: "boolean", Description: "true to include the content of the files"},
			},
			Handler: func(ctx context.Context, p map[string]string) string {
				v, _ := strconv.Atoi(strings.TrimSpace(p["version"]))
				withFiles, _ := strconv.ParseBool(strings.TrimSpace(p["files"]))
				got, err := automationService.GetSkill(ctx, strings.TrimSpace(p["handle"]), v, withFiles)
				if err != nil {
					return err.Error()
				}
				return chatJSON(got)
			},
		},
		{
			Name:        "save_skill",
			Description: "Save a new version of a skill (creating it if the handle is new). Old versions are kept. It only becomes the one in use with activate=true (the first version always is). The description must say WHEN to use the skill: agents pick skills by it.",
			Mutating:    true,
			Params: []chat.ParamDef{
				{Name: "handle", Type: "string", Required: true, Description: "Lowercase letters, digits, _ - . ; starts with a letter"},
				{Name: "text", Type: "string", Required: true, Description: "The instructions, in markdown"},
				{Name: "description", Type: "string", Description: "When to use the skill"},
				{Name: "requires", Type: "array", Description: `Toolkits the skill needs, e.g. ["compose.records","workflows"]`},
				{Name: "resources", Type: "objects", Description: `Reference files (text): [{"path":"prices.csv","content":"..."}]`},
				{Name: "note", Type: "string", Description: "Why this version exists"},
				{Name: "activate", Type: "boolean", Description: "true to make this the version in use"},
			},
			Handler: func(ctx context.Context, p map[string]string) string {
				requires, err := automationService.ParseSkillRequires(p["requires"])
				if err != nil {
					return err.Error()
				}
				files, err := automationService.ParseSkillFiles(p["resources"])
				if err != nil {
					return err.Error()
				}
				activate, _ := strconv.ParseBool(strings.TrimSpace(p["activate"]))

				saved, err := automationService.SaveSkillVersion(ctx, automationService.SavePrompt{
					Handle:      strings.TrimSpace(p["handle"]),
					Text:        p["text"],
					Description: p["description"],
					Note:        p["note"],
					Requires:    requires,
					Resources:   files,
					Activate:    activate,
				})
				if err != nil {
					return err.Error()
				}
				return chatJSON(saved)
			},
		},
		{
			Name:        "import_skill",
			Description: "Save a skill written as SKILL.md: a YAML header (name, description, optional requires) and the instructions",
			Mutating:    true,
			Params: []chat.ParamDef{
				{Name: "content", Type: "string", Required: true, Description: "The SKILL.md text"},
				{Name: "handle", Type: "string", Description: "Use this handle instead of the name in the header"},
				{Name: "activate", Type: "boolean", Description: "true to make it the version in use"},
			},
			Handler: func(ctx context.Context, p map[string]string) string {
				activate, _ := strconv.ParseBool(strings.TrimSpace(p["activate"]))
				saved, err := automationService.ImportSkill(ctx, []byte(p["content"]), p["handle"], "", activate)
				if err != nil {
					return err.Error()
				}
				return chatJSON(saved)
			},
		},
		{
			Name:        "activate_skill",
			Description: "Make a skill version the one in use - to roll back to an earlier version.",
			Mutating:    true,
			Params: []chat.ParamDef{
				{Name: "handle", Type: "string", Required: true, Description: "Skill handle"},
				{Name: "version", Type: "integer", Required: true, Description: "Version number"},
			},
			Handler: func(ctx context.Context, p map[string]string) string {
				v, err := strconv.Atoi(strings.TrimSpace(p["version"]))
				if err != nil || v < 1 {
					return "version must be a version number"
				}
				got, err := automationService.ActivateSkill(ctx, strings.TrimSpace(p["handle"]), v)
				if err != nil {
					return err.Error()
				}
				return chatJSON(got)
			},
		},
	}
}
