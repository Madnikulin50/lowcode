package service

import (
	"context"
	"strconv"
	"strings"

	automationService "github.com/madnikulin50/lowcode/server/automation/service"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

// chatPromptToolDefs let the assistant work with the prompt library: look at
// prompts, save and activate versions (which change what live workflows say
// to the model, so they ask for the user's "да"), and judge a version by its
// test cases.
func chatPromptToolDefs() []chat.ToolDef {
	return []chat.ToolDef{
		{
			Name:        "list_prompts",
			Description: "List the prompts in the prompt library. AI workflow steps and rule chain AI nodes use one as @prompt:<handle>.",
			Handler: func(ctx context.Context, _ map[string]string) string {
				list, err := automationService.ListPrompts(ctx)
				if err != nil {
					return err.Error()
				}
				if len(list) == 0 {
					return "No prompts found"
				}
				return chatJSON(list)
			},
		},
		{
			Name:        "get_prompt",
			Description: "Show a prompt's text. Without a version: the active one.",
			Params: []chat.ParamDef{
				{Name: "handle", Type: "string", Required: true, Description: "Prompt handle"},
				{Name: "version", Type: "integer", Description: "Version number"},
			},
			Handler: func(ctx context.Context, p map[string]string) string {
				v, _ := strconv.Atoi(strings.TrimSpace(p["version"]))
				got, err := automationService.GetPrompt(ctx, strings.TrimSpace(p["handle"]), v)
				if err != nil {
					return err.Error()
				}
				return chatJSON(got)
			},
		},
		{
			Name:        "save_prompt",
			Description: "Save a new version of a prompt (creating it if the handle is new). Old versions are kept. It only becomes the one in use with activate=true (the first version always is) - so evaluate it first.",
			Mutating:    true,
			Params: []chat.ParamDef{
				{Name: "handle", Type: "string", Required: true, Description: "Lowercase letters, digits, _ - . ; starts with a letter"},
				{Name: "text", Type: "string", Required: true, Description: "The instruction"},
				{Name: "description", Type: "string", Description: "What the prompt is for"},
				{Name: "note", Type: "string", Description: "Why this version exists"},
				{Name: "cases", Type: "objects", Description: `Test cases: [{"name":"outage","inputs":{"ticket":"server down"},"expect":{"priority":"high"}}]`},
				{Name: "activate", Type: "boolean", Description: "true to make this the version in use"},
			},
			Handler: func(ctx context.Context, p map[string]string) string {
				cases, err := automationService.ParsePromptCases(p["cases"])
				if err != nil {
					return err.Error()
				}
				activate, _ := strconv.ParseBool(strings.TrimSpace(p["activate"]))

				saved, err := automationService.SavePromptVersion(ctx, automationService.SavePrompt{
					Handle:      strings.TrimSpace(p["handle"]),
					Text:        p["text"],
					Description: p["description"],
					Note:        p["note"],
					Cases:       cases,
					Activate:    activate,
				})
				if err != nil {
					return err.Error()
				}
				return chatJSON(saved)
			},
		},
		{
			Name:        "activate_prompt",
			Description: "Make a prompt version the one in use - to roll back to an earlier version.",
			Mutating:    true,
			Params: []chat.ParamDef{
				{Name: "handle", Type: "string", Required: true, Description: "Prompt handle"},
				{Name: "version", Type: "integer", Required: true, Description: "Version number"},
			},
			Handler: func(ctx context.Context, p map[string]string) string {
				v, err := strconv.Atoi(strings.TrimSpace(p["version"]))
				if err != nil || v < 1 {
					return "version must be a version number"
				}
				got, err := automationService.ActivatePrompt(ctx, strings.TrimSpace(p["handle"]), v)
				if err != nil {
					return err.Error()
				}
				return chatJSON(got)
			},
		},
		{
			Name:        "evaluate_prompt",
			Description: "Run a prompt's test cases on the real model and report how many pass. Name several versions (\"1,2\") to compare them. Takes a while.",
			Params: []chat.ParamDef{
				{Name: "handle", Type: "string", Required: true, Description: "Prompt handle"},
				{Name: "versions", Type: "string", Description: `Versions to run, e.g. "1,2". Default: the active one`},
			},
			Handler: func(ctx context.Context, p map[string]string) string {
				reports, err := automationService.EvaluatePrompt(ctx, strings.TrimSpace(p["handle"]), p["versions"], "", "", "")
				if err != nil {
					return err.Error()
				}
				return chatJSON(reports)
			},
		},
	}
}
