package aiagent

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// Skills.
//
// A skill is a named, versioned set of instructions that an agent loads when
// the task calls for it, instead of carrying every procedure in its system
// prompt. The agent sees only a catalog (handle and "when to use it"); the body
// and the reference files come on demand (see skill_tools.go).
//
// Where skills are kept is not this package's business: the automation
// component stores them in the prompt library, a directory can hold them as
// files, a remote agent can publish its own. Each place is a SkillSource; the
// registry below puts them in one catalog, the first source that has a handle
// winning (database, then files, then remote).

const (
	SkillSourceDB     = "db"
	SkillSourceFile   = "file"
	SkillSourceRemote = "remote"
)

var (
	// SkillHandlePattern is what a skill handle looks like (same as a prompt's)
	SkillHandlePattern = PromptHandlePattern

	skillSourceOrder = []string{SkillSourceDB, SkillSourceFile, SkillSourceRemote}
)

type (
	// SkillFile is a text file shipped with a skill
	SkillFile struct {
		Path    string `json:"path" yaml:"path"`
		Mime    string `json:"mime,omitempty" yaml:"mime,omitempty"`
		Content string `json:"content,omitempty" yaml:"content,omitempty"`
	}

	Skill struct {
		Handle      string      `json:"handle"`
		Description string      `json:"description"`
		Body        string      `json:"body,omitempty"`
		Version     int         `json:"version,omitempty"`
		Requires    []string    `json:"requires,omitempty"`
		Resources   []SkillFile `json:"resources,omitempty"`
		Source      string      `json:"source,omitempty"`
	}

	// SkillInfo is a skill as the catalog lists it: no body, no file content
	SkillInfo struct {
		Handle      string   `json:"handle"`
		Description string   `json:"description"`
		Version     int      `json:"version,omitempty"`
		Requires    []string `json:"requires,omitempty"`
		Resources   []string `json:"resources,omitempty"`
		Source      string   `json:"source,omitempty"`
	}

	// SkillSource is a place skills come from. Reads are made on behalf of a
	// running agent, not of a person, so a source does not check permissions.
	SkillSource interface {
		// ListSkills returns the active version of every skill, with bodies
		// and files left out where that is cheaper
		ListSkills(ctx context.Context) ([]Skill, error)
		// GetSkill returns one skill in full; version 0 means the active one
		GetSkill(ctx context.Context, handle string, version int) (*Skill, error)
	}
)

// Ref names exactly what was used, "handle@version" (just the handle for a
// source that has no versions)
func (s Skill) Ref() string {
	if s.Version > 0 {
		return fmt.Sprintf("%s@%d", s.Handle, s.Version)
	}
	return s.Handle
}

func (s Skill) Info() SkillInfo {
	i := SkillInfo{
		Handle: s.Handle, Description: s.Description, Version: s.Version,
		Requires: s.Requires, Source: s.Source,
	}
	for _, f := range s.Resources {
		i.Resources = append(i.Resources, f.Path)
	}
	return i
}

// File returns a shipped file by path
func (s Skill) File(path string) (SkillFile, bool) {
	path = cleanSkillPath(path)
	for _, f := range s.Resources {
		if cleanSkillPath(f.Path) == path {
			return f, true
		}
	}
	return SkillFile{}, false
}

func cleanSkillPath(p string) string {
	p = strings.TrimSpace(strings.ReplaceAll(p, "\\", "/"))
	return strings.TrimPrefix(p, "./")
}

var (
	skillSourcesMu sync.RWMutex
	skillSources   = map[string]SkillSource{}
)

// SetSkillSource registers (or, with nil, removes) a source of skills under
// one of the SkillSource* names.
func SetSkillSource(name string, src SkillSource) {
	skillSourcesMu.Lock()
	defer skillSourcesMu.Unlock()
	if src == nil {
		delete(skillSources, name)
		return
	}
	skillSources[name] = src
}

// orderedSkillSources returns the registered sources, the strongest first.
func orderedSkillSources() []SkillSource {
	skillSourcesMu.RLock()
	defer skillSourcesMu.RUnlock()
	out := make([]SkillSource, 0, len(skillSources))
	for _, name := range skillSourceOrder {
		if src, ok := skillSources[name]; ok {
			out = append(out, src)
		}
	}
	return out
}

// ListSkills is the merged catalog, by handle. A source that fails is skipped
// rather than hiding the others - an agent can still work with what is left.
func ListSkills(ctx context.Context) []Skill {
	seen := map[string]struct{}{}
	var out []Skill
	for _, src := range orderedSkillSources() {
		list, err := src.ListSkills(ctx)
		if err != nil {
			continue
		}
		for _, s := range list {
			if _, dup := seen[s.Handle]; dup {
				continue
			}
			seen[s.Handle] = struct{}{}
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Handle < out[j].Handle })
	return out
}

// GetSkill finds a skill in the strongest source that has it.
func GetSkill(ctx context.Context, handle string, version int) (*Skill, error) {
	handle = strings.TrimSpace(handle)
	if !SkillHandlePattern.MatchString(handle) {
		return nil, fmt.Errorf("skill handle %q is not valid", handle)
	}

	var firstErr error
	for _, src := range orderedSkillSources() {
		s, err := src.GetSkill(ctx, handle, version)
		if err == nil && s != nil {
			return s, nil
		}
		if err != nil && firstErr == nil && !isNotFound(err) {
			firstErr = err
		}
	}
	if firstErr != nil {
		return nil, firstErr
	}
	return nil, fmt.Errorf("skill %q does not exist", handle)
}

func isNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "does not exist")
}

// SkillsFor returns the catalog an agent may use: patterns are handles, "*"
// meaning every skill; none means no skills.
func SkillsFor(ctx context.Context, patterns []string) []Skill {
	if len(patterns) == 0 {
		return nil
	}
	all := ListSkills(ctx)
	for _, p := range patterns {
		if strings.TrimSpace(p) == "*" {
			return all
		}
	}
	allowed := map[string]struct{}{}
	for _, p := range patterns {
		allowed[strings.TrimSpace(p)] = struct{}{}
	}
	var out []Skill
	for _, s := range all {
		if _, ok := allowed[s.Handle]; ok {
			out = append(out, s)
		}
	}
	return out
}

var skillNameCleanup = regexp.MustCompile(`[^a-z0-9_.-]+`)

// SkillHandleFromName turns a free-form name ("Read Drawings") into a handle
// ("read-drawings"); empty when nothing usable is left.
func SkillHandleFromName(name string) string {
	h := skillNameCleanup.ReplaceAllString(strings.ToLower(strings.TrimSpace(name)), "-")
	h = strings.Trim(h, "-_.")
	if len(h) > 64 {
		h = strings.Trim(h[:64], "-_.")
	}
	if !SkillHandlePattern.MatchString(h) {
		return ""
	}
	return h
}

type skillCtxKey struct{}

// ContextWithSkill makes the next agent run use a skill the caller chose
// ("handle" or "handle@version"): its text goes into the prompt and the tools
// it needs are available from the first turn. Empty leaves ctx unchanged.
func ContextWithSkill(ctx context.Context, ref string) context.Context {
	if ref = strings.TrimSpace(ref); ref == "" {
		return ctx
	}
	return context.WithValue(ctx, skillCtxKey{}, ref)
}

// SkillFromContext returns the skill set by ContextWithSkill, or "".
func SkillFromContext(ctx context.Context) string {
	s, _ := ctx.Value(skillCtxKey{}).(string)
	return s
}
