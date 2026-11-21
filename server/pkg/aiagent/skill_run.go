package aiagent

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/cloudwego/eino/schema"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

// What an agent does with skills during one run.
//
// The system prompt carries only the catalog (handle and "when to use it").
// The agent asks for a skill with load_skill; the answer is the instructions,
// the list of files that come with it and - the point of "requires" - the
// toolkits the skill needs, which join the agent's tools from the next turn
// on. A run that never loads a skill pays for the catalog only.
//
// A skill does not widen what a call may do: tools it brings are checked for
// confirmation like any other (a Mutating tool still needs the caller's yes),
// and the agent can load only skills its spec lists.

const (
	ToolListSkills        = "list_skills"
	ToolLoadSkill         = "load_skill"
	ToolReadSkillResource = "read_skill_resource"
)

// SkillRun is the skills state of one agent run.
type SkillRun struct {
	mu sync.Mutex

	catalog map[string]Skill // what this agent may load, by handle
	loaded  map[string]string
	order   []string // handles in the order they were loaded

	tools   map[string]struct{} // names of tools activated by skills
	pending []chat.ToolDef      // activated, not yet handed to the loop
}

// NewSkillRun starts the skills state for an agent that may use the catalog.
func NewSkillRun(catalog []Skill) *SkillRun {
	r := &SkillRun{
		catalog: make(map[string]Skill, len(catalog)),
		loaded:  map[string]string{},
		tools:   map[string]struct{}{},
	}
	for _, s := range catalog {
		r.catalog[s.Handle] = s
	}
	return r
}

type skillRunKey struct{}

func ContextWithSkillRun(ctx context.Context, r *SkillRun) context.Context {
	if r == nil {
		return ctx
	}
	return context.WithValue(ctx, skillRunKey{}, r)
}

func SkillRunFromContext(ctx context.Context) *SkillRun {
	r, _ := ctx.Value(skillRunKey{}).(*SkillRun)
	return r
}

// Loaded returns the skills loaded so far as "handle@version", in order.
func (r *SkillRun) Loaded() []string {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, 0, len(r.order))
	for _, h := range r.order {
		out = append(out, r.loaded[h])
	}
	return out
}

// Handles returns the handles loaded so far.
func (r *SkillRun) Handles() []string {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.order...)
}

// Allows reports whether the agent may load the skill.
func (r *SkillRun) Allows(handle string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.catalog[handle]
	return ok
}

func (r *SkillRun) markLoaded(s *Skill) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.loaded[s.Handle]; !ok {
		r.order = append(r.order, s.Handle)
	}
	r.loaded[s.Handle] = s.Ref()
}

func (r *SkillRun) isLoaded(handle string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.loaded[handle]
	return ok
}

// Activate adds the tools of the toolkits to the run. have is what the agent
// already has; it returns the tools that are new, and the toolkits that do not
// exist (a remote toolkit may be down). Active tools persist: calling it again
// for the same toolkit adds nothing.
func (r *SkillRun) Activate(kits []string, have []chat.ToolDef) (added []chat.ToolDef, missing []string) {
	if len(kits) == 0 {
		return nil, nil
	}
	known := make(map[string]struct{}, len(have))
	for _, t := range have {
		known[t.Name] = struct{}{}
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	for _, name := range kits {
		kit, ok := DefaultCatalog().Get(name)
		if !ok {
			missing = append(missing, name)
			continue
		}
		for _, t := range kit.Tools {
			if t.Name == "" {
				continue
			}
			if _, dup := known[t.Name]; dup {
				continue
			}
			known[t.Name] = struct{}{}
			r.tools[t.Name] = struct{}{}
			added = append(added, t)
			r.pending = append(r.pending, t)
		}
	}
	return added, missing
}

// Drain returns the tools activated since the last call.
func (r *SkillRun) Drain() []chat.ToolDef {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	out := r.pending
	r.pending = nil
	return out
}

// Restore replays loaded skills in a new run - the confirmation of a mutating
// call happens in a later request, and the tools the skills brought must be
// there again to be executed.
func (r *SkillRun) Restore(ctx context.Context, handles []string, have []chat.ToolDef) []chat.ToolDef {
	var all []chat.ToolDef
	for _, h := range handles {
		s, err := GetSkill(ctx, h, 0)
		if err != nil || !r.Allows(h) {
			continue
		}
		r.markLoaded(s)
		added, _ := r.Activate(s.Requires, append(append([]chat.ToolDef(nil), have...), all...))
		all = append(all, added...)
	}
	r.Drain()
	return all
}

// SkillTools are the tools an agent with skills gets. have is the agent's tool
// list, so a skill does not add a tool the agent already has.
func SkillTools(run *SkillRun, have func() []chat.ToolDef) []chat.ToolDef {
	return []chat.ToolDef{
		{
			Name:        ToolListSkills,
			Description: "List the skills you can load, with when each applies",
			Handler: func(ctx context.Context, _ map[string]string) string {
				return run.catalogText()
			},
		},
		{
			Name:        ToolLoadSkill,
			Description: "Load a skill: its instructions, the files that come with it and the tools it needs. Do this before starting a task a skill is meant for, then follow the instructions.",
			Params: []chat.ParamDef{
				{Name: "handle", Required: true, Description: "Skill handle from the catalog"},
				{Name: "version", Type: chat.ParamTypeInteger, Description: "Version number; default: the one in use"},
			},
			Handler: func(ctx context.Context, p map[string]string) string {
				return run.load(ctx, strings.TrimSpace(p["handle"]), p["version"], have())
			},
		},
		{
			Name:        ToolReadSkillResource,
			Description: "Read a file that comes with a skill you have loaded",
			Params: []chat.ParamDef{
				{Name: "handle", Required: true, Description: "Skill handle"},
				{Name: "path", Required: true, Description: "File path as listed by load_skill"},
			},
			Handler: func(ctx context.Context, p map[string]string) string {
				return run.readResource(ctx, strings.TrimSpace(p["handle"]), p["path"])
			},
		},
	}
}

func (r *SkillRun) catalogText() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.catalog) == 0 {
		return "No skills available"
	}
	handles := make([]string, 0, len(r.catalog))
	for h := range r.catalog {
		handles = append(handles, h)
	}
	sort.Strings(handles)
	var b strings.Builder
	for _, h := range handles {
		fmt.Fprintf(&b, "- %s: %s\n", h, r.catalog[h].Description)
	}
	return strings.TrimRight(b.String(), "\n")
}

// load is the load_skill tool
func (r *SkillRun) load(ctx context.Context, handle, version string, have []chat.ToolDef) string {
	if handle == "" {
		return "handle is required"
	}
	if !r.Allows(handle) {
		return fmt.Sprintf("skill %q is not available. Skills you can load:\n%s", handle, r.catalogText())
	}

	var n int
	if v := strings.TrimSpace(version); v != "" {
		if _, err := fmt.Sscanf(v, "%d", &n); err != nil || n < 0 {
			return fmt.Sprintf("%q is not a version number", version)
		}
	}
	s, err := GetSkill(ctx, handle, n)
	if err != nil {
		return "cannot load skill: " + err.Error()
	}
	r.markLoaded(s)
	return r.render(s, have)
}

// render is what the model reads after loading a skill: instructions, files and
// the tools that just became available.
func (r *SkillRun) render(s *Skill, have []chat.ToolDef) string {
	added, missing := r.Activate(s.Requires, have)

	var b strings.Builder
	fmt.Fprintf(&b, "# Skill %s\n\n%s\n", s.Ref(), strings.TrimSpace(s.Body))

	if len(s.Resources) > 0 {
		b.WriteString("\n## Files\nRead one with read_skill_resource when the instructions call for it:\n")
		for _, f := range s.Resources {
			if f.Content != "" {
				fmt.Fprintf(&b, "- %s (%d bytes)\n", f.Path, len(f.Content))
			} else {
				fmt.Fprintf(&b, "- %s\n", f.Path)
			}
		}
	}
	if len(added) > 0 {
		b.WriteString("\n## Tools now available\n")
		b.WriteString(toolBriefs(added))
	}
	if len(missing) > 0 {
		fmt.Fprintf(&b, "\n(toolkits not available right now: %s)\n", strings.Join(missing, ", "))
	}
	return b.String()
}

func toolBriefs(tools []chat.ToolDef) string {
	var b strings.Builder
	for _, t := range tools {
		fmt.Fprintf(&b, "- %s", t.Name)
		if len(t.Params) > 0 {
			names := make([]string, 0, len(t.Params))
			for _, p := range t.Params {
				n := p.Name
				if p.Required {
					n += "*"
				}
				names = append(names, n)
			}
			fmt.Fprintf(&b, "(%s)", strings.Join(names, ", "))
		}
		if t.Description != "" {
			fmt.Fprintf(&b, " - %s", t.Description)
		}
		b.WriteString("\n")
	}
	return b.String()
}

// readResource is the read_skill_resource tool
func (r *SkillRun) readResource(ctx context.Context, handle, path string) string {
	if handle == "" || strings.TrimSpace(path) == "" {
		return "handle and path are required"
	}
	if !r.isLoaded(handle) {
		return fmt.Sprintf("load skill %q first", handle)
	}
	r.mu.Lock()
	ref := r.loaded[handle]
	r.mu.Unlock()

	var version int
	if i := strings.LastIndex(ref, "@"); i >= 0 {
		fmt.Sscanf(ref[i+1:], "%d", &version)
	}
	s, err := GetSkill(ctx, handle, version)
	if err != nil {
		return "cannot read: " + err.Error()
	}
	f, ok := s.File(path)
	if !ok {
		names := make([]string, 0, len(s.Resources))
		for _, rf := range s.Resources {
			names = append(names, rf.Path)
		}
		return fmt.Sprintf("skill %q has no file %q. Files: %s", handle, path, strings.Join(names, ", "))
	}
	return f.Content
}

// CatalogPrompt is the part of the system prompt that tells the agent about
// its skills.
func (r *SkillRun) CatalogPrompt() string {
	if r == nil || len(r.catalog) == 0 {
		return ""
	}
	return "## Skills\n" +
		"You can load skills: instructions for specific kinds of tasks. Before you start a task one of them covers, " +
		"call " + ToolLoadSkill + " with its handle and follow what it says. Skills:\n" +
		r.catalogText()
}

// Preload loads a skill the caller chose (a workflow step or rule chain node
// with a skill set): its text goes straight into the prompt and its tools are
// available from the first turn. Returns the text, and the tools it brought.
func (r *SkillRun) Preload(ctx context.Context, handle string, have []chat.ToolDef) (string, []chat.ToolDef, error) {
	handle, version := SplitSkillRef(handle)
	s, err := GetSkill(ctx, handle, version)
	if err != nil {
		return "", nil, err
	}
	r.mu.Lock()
	r.catalog[s.Handle] = *s // a chosen skill is allowed even if the agent lists none
	r.mu.Unlock()
	r.markLoaded(s)
	text := r.render(s, have)
	return text, r.Drain(), nil
}

// SplitSkillRef reads "handle" or "handle@version".
func SplitSkillRef(ref string) (string, int) {
	ref = strings.TrimSpace(ref)
	if i := strings.LastIndex(ref, "@"); i > 0 {
		var n int
		if _, err := fmt.Sscanf(ref[i+1:], "%d", &n); err == nil && n > 0 {
			return ref[:i], n
		}
	}
	return ref, 0
}

// withExtraTools adds tools to a run's options, with their confirmation rules:
// a Mutating tool a skill brought needs the caller's yes like any other. The
// rest of the options are untouched.
func withExtraTools(opt Options, extra []chat.ToolDef) Options {
	opt.Tools = Flatten(ToolKit{Name: "_run", Tools: opt.Tools}, ToolKit{Name: "_skill", Tools: extra})

	mine := make(map[string]struct{}, len(extra))
	for _, t := range extra {
		mine[t.Name] = struct{}{}
	}
	gate := NeedsConfirmFromToolDefs(extra)
	prev := opt.NeedsConfirm
	opt.NeedsConfirm = func(calls []Call) bool {
		if prev != nil && prev(calls) {
			return true
		}
		var ours []Call
		for _, c := range calls {
			if _, ok := mine[c.Name]; ok {
				ours = append(ours, c)
			}
		}
		return len(ours) > 0 && gate(ours)
	}
	return opt
}

// approvableTools are the tools a human may approve a call to on behalf of an
// agent with skills: its own, plus those of toolkits the skills it may load
// need. The confirmation happens in a later request, where the run's loaded
// skills are gone, so this is the bound instead of the exact set.
func (a *Agent) approvableTools(ctx context.Context, tools []chat.ToolDef) []chat.ToolDef {
	if len(a.cfg.Skills) == 0 {
		return tools
	}
	all := append([]chat.ToolDef(nil), tools...)
	for _, s := range SkillsFor(ctx, a.cfg.Skills) {
		for _, kit := range s.Requires {
			if k, ok := DefaultCatalog().Get(kit); ok {
				all = Flatten(ToolKit{Name: "_a", Tools: all}, ToolKit{Name: "_b", Tools: k.Tools})
			}
		}
	}
	return all
}

// skillsOf finds the run's skills state: on the options, or on the context
// for callers that set it there.
func skillsOf(ctx context.Context, opt Options) *SkillRun {
	if opt.Skills != nil {
		return opt.Skills
	}
	return SkillRunFromContext(ctx)
}

// HasCatalog reports whether the agent may load any skill.
func (r *SkillRun) HasCatalog() bool {
	if r == nil {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.catalog) > 0
}

// AttachSkills gives a run built from raw Options (compose chat) the skills
// the patterns allow: the load tools, the catalog in the first system message
// and the state the loop reads. restore lists skills loaded earlier in the same
// conversation (the run that stopped for confirmation), so the tools they
// brought can be executed now. With no skills available the options come back
// unchanged.
func AttachSkills(ctx context.Context, opt Options, patterns []string, restore []string) Options {
	if len(patterns) == 0 || len(opt.Tools) == 0 {
		return opt
	}
	run := NewSkillRun(SkillsFor(ctx, patterns))
	if !run.HasCatalog() {
		return opt
	}

	tools := opt.Tools
	current := tools
	meta := SkillTools(run, func() []chat.ToolDef { return current })
	tools = Flatten(ToolKit{Name: "_base", Tools: tools}, ToolKit{Name: "_skills", Tools: meta})
	if len(restore) > 0 {
		tools = Flatten(ToolKit{Name: "_base", Tools: tools}, ToolKit{Name: "_restored", Tools: run.Restore(ctx, restore, tools)})
	}
	current = tools

	opt.Tools = tools
	opt.NeedsConfirm = NeedsConfirmFromToolDefs(tools)
	opt.Skills = run

	if cat := run.CatalogPrompt(); cat != "" && len(opt.Messages) > 0 && opt.Messages[0].Role == schema.System {
		first := *opt.Messages[0]
		first.Content += "\n\n" + cat
		msgs := append([]*schema.Message{&first}, opt.Messages[1:]...)
		opt.Messages = msgs
	}
	return opt
}
