package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
	"github.com/madnikulin50/lowcode/server/store"
)

// Skills in the prompt library (see promptLibrary.Skills): what a skill is
// made of is checked here, and the library serves them to running agents as
// the "db" source of aiagent skills.

func validSkillRequires(requires []string) error {
	if len(requires) > maxSkillRequires {
		return errors.InvalidData("a skill can require at most %d toolkits", maxSkillRequires)
	}
	seen := map[string]struct{}{}
	for _, name := range requires {
		if !aiagent.PromptHandlePattern.MatchString(name) {
			return errors.InvalidData("toolkit name %q is not valid", name)
		}
		if _, dup := seen[name]; dup {
			return errors.InvalidData("toolkit %q is listed twice", name)
		}
		seen[name] = struct{}{}
	}
	return nil
}

func validSkillFiles(files types.PromptFiles) error {
	if len(files) > maxSkillFiles {
		return errors.InvalidData("a skill can carry at most %d files", maxSkillFiles)
	}
	var total int
	seen := map[string]struct{}{}
	for i := range files {
		f := &files[i]
		p, ok := aiagent.SafeSkillPath(f.Path)
		if !ok {
			return errors.InvalidData("file path %q is not allowed: use a relative path inside the skill, other than %s", f.Path, aiagent.SkillFileName)
		}
		f.Path = p
		if _, dup := seen[p]; dup {
			return errors.InvalidData("file %q is listed twice", p)
		}
		seen[p] = struct{}{}
		if !utf8.ValidString(f.Content) {
			return errors.InvalidData("file %q is not text; a skill can only carry text files", p)
		}
		if len(f.Content) > maxSkillFileLen {
			return errors.InvalidData("file %q is longer than %d KB", p, maxSkillFileLen/1024)
		}
		total += len(f.Content)
		if total > maxSkillResourcesLen {
			return errors.InvalidData("the files of a skill together are longer than %d KB", maxSkillResourcesLen/1024)
		}
	}
	return nil
}

// toAgentSkill is a stored version as an agent sees it
func toAgentSkill(v *types.PromptVersion, full bool) aiagent.Skill {
	s := aiagent.Skill{
		Handle: v.Handle, Description: v.Description, Version: v.Version,
		Requires: v.Requires, Source: aiagent.SkillSourceDB,
	}
	if full {
		s.Body = v.Text
	}
	for _, f := range v.Resources {
		file := aiagent.SkillFile{Path: f.Path, Mime: f.Mime}
		if full {
			file.Content = f.Content
		}
		s.Resources = append(s.Resources, file)
	}
	return s
}

// skillSource serves the library's skills to running agents. Reads need no user
// permission: it is a workflow or a chat asking, not a person - the same as
// resolving "@prompt:" references.
type skillSource struct{ lib *promptLibrary }

func (src skillSource) ListSkills(ctx context.Context) ([]aiagent.Skill, error) {
	all, _, err := store.SearchAutomationPromptVersions(ctx, src.lib.store, types.PromptVersionFilter{})
	if err != nil {
		return nil, err
	}
	var out []aiagent.Skill
	for _, v := range all {
		if v.IsSkill() && v.Active {
			out = append(out, toAgentSkill(v, false))
		}
	}
	return out, nil
}

func (src skillSource) GetSkill(ctx context.Context, handle string, version int) (*aiagent.Skill, error) {
	v, err := src.lib.Skills().find(ctx, src.lib.store, handle, version)
	if err != nil {
		return nil, err
	}
	s := toAgentSkill(v, true)
	return &s, nil
}

// SkillFileBrief is a file of a skill; Content only where it was asked for
type SkillFileBrief struct {
	Path    string `json:"path"`
	Mime    string `json:"mime,omitempty"`
	Size    int    `json:"size"`
	Content string `json:"content,omitempty"`
}

// SkillBrief is a skill version with what makes it a skill
type SkillBrief struct {
	PromptBrief
	Requires  []string         `json:"requires,omitempty"`
	Resources []SkillFileBrief `json:"resources,omitempty"`
}

func skillBriefOf(v *types.PromptVersion, withText, withFiles bool) SkillBrief {
	b := SkillBrief{PromptBrief: briefOf(v, withText), Requires: v.Requires}
	for _, f := range v.Resources {
		fb := SkillFileBrief{Path: f.Path, Mime: f.Mime, Size: len(f.Content)}
		if withFiles {
			fb.Content = f.Content
		}
		b.Resources = append(b.Resources, fb)
	}
	return b
}

func skillLib() (*promptLibrary, error) {
	lib, err := promptLib()
	if err != nil {
		return nil, err
	}
	return lib.Skills(), nil
}

func ListSkills(ctx context.Context) ([]PromptSummary, error) {
	lib, err := skillLib()
	if err != nil {
		return nil, err
	}
	return lib.List(ctx)
}

// GetSkill returns one version (0: the active one) with its instructions and,
// when withFiles is set, the content of its files.
func GetSkill(ctx context.Context, handle string, version int, withFiles bool) (*SkillBrief, error) {
	lib, err := skillLib()
	if err != nil {
		return nil, err
	}
	v, err := lib.Get(ctx, handle, version)
	if err != nil {
		return nil, err
	}
	b := skillBriefOf(v, true, withFiles)
	return &b, nil
}

func SkillHistory(ctx context.Context, handle string) ([]SkillBrief, error) {
	lib, err := skillLib()
	if err != nil {
		return nil, err
	}
	vv, err := lib.History(ctx, handle)
	if err != nil {
		return nil, err
	}
	out := make([]SkillBrief, 0, len(vv))
	for _, v := range vv {
		out = append(out, skillBriefOf(v, false, false))
	}
	return out, nil
}

// SaveSkillVersion stores a new version; in.Text holds the instructions.
func SaveSkillVersion(ctx context.Context, in SavePrompt) (*SkillBrief, error) {
	lib, err := skillLib()
	if err != nil {
		return nil, err
	}
	v, err := lib.Save(ctx, in)
	if err != nil {
		return nil, err
	}
	b := skillBriefOf(v, false, false)
	return &b, nil
}

func ActivateSkill(ctx context.Context, handle string, version int) (*SkillBrief, error) {
	lib, err := skillLib()
	if err != nil {
		return nil, err
	}
	v, err := lib.Activate(ctx, handle, version)
	if err != nil {
		return nil, err
	}
	b := skillBriefOf(v, false, false)
	return &b, nil
}

func DeleteSkill(ctx context.Context, handle string) error {
	lib, err := skillLib()
	if err != nil {
		return err
	}
	return lib.Delete(ctx, handle)
}

// ImportSkill saves a SKILL.md (or a zip with SKILL.md and files) as a new
// version. handle, when given, replaces the name from the header.
func ImportSkill(ctx context.Context, data []byte, handle, note string, activate bool) (*SkillBrief, error) {
	var (
		doc aiagent.SkillDoc
		err error
	)
	if len(data) > 4 && string(data[:2]) == "PK" {
		doc, err = aiagent.ParseSkillZip(data)
	} else {
		doc, err = aiagent.ParseSkillMarkdown(data)
	}
	if err != nil {
		return nil, errors.InvalidData("cannot import the skill: %v", err)
	}
	if h := strings.TrimSpace(handle); h != "" {
		doc.Handle = h
	}

	requires := types.PromptRequires(doc.Requires)
	files := make(types.PromptFiles, 0, len(doc.Resources))
	for _, f := range doc.Resources {
		files = append(files, types.PromptFile{Path: f.Path, Mime: f.Mime, Content: f.Content})
	}
	if note == "" {
		note = "imported"
	}
	return SaveSkillVersion(ctx, SavePrompt{
		Handle: doc.Handle, Text: doc.Body, Description: doc.Description, Note: note,
		Requires: &requires, Resources: &files, Activate: activate,
	})
}

// ExportSkill returns a skill as SKILL.md, or as a zip when it has files (or
// asZip is set), with the file name to offer.
func ExportSkill(ctx context.Context, handle string, version int, asZip bool) (data []byte, name string, err error) {
	lib, err := skillLib()
	if err != nil {
		return nil, "", err
	}
	v, err := lib.Get(ctx, handle, version)
	if err != nil {
		return nil, "", err
	}
	s := toAgentSkill(v, true)
	for i, f := range v.Resources {
		s.Resources[i].Content = f.Content
	}

	if asZip || len(s.Resources) > 0 {
		data, err = aiagent.BuildSkillZip(s)
		return data, fmt.Sprintf("%s.zip", v.Handle), err
	}
	data, err = aiagent.RenderSkillMarkdown(s)
	return data, fmt.Sprintf("%s.md", v.Handle), err
}

// ParseSkillRequires reads the toolkit names a tool call gives: a JSON array
// or names separated by commas or spaces. Empty gives nil - "not given" - and
// "[]" gives an empty list, which clears them.
func ParseSkillRequires(s string) (*types.PromptRequires, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	var out types.PromptRequires
	if strings.HasPrefix(s, "[") {
		if err := json.Unmarshal([]byte(s), &out); err != nil {
			return nil, errors.InvalidData(`requires must be a JSON array of toolkit names or names separated by commas: %v`, err)
		}
		if out == nil {
			out = types.PromptRequires{}
		}
		return &out, nil
	}
	out = types.PromptRequires{}
	for _, name := range strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' || r == ';' }) {
		out = append(out, name)
	}
	return &out, nil
}

// ParseSkillFiles reads a JSON array of {path, content} (mime is optional).
// Empty gives nil, as in ParseSkillRequires.
func ParseSkillFiles(s string) (*types.PromptFiles, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	var out types.PromptFiles
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil, errors.InvalidData(`resources must be a JSON array like [{"path":"prices.csv","content":"..."}]: %v`, err)
	}
	if out == nil {
		out = types.PromptFiles{}
	}
	return &out, nil
}
