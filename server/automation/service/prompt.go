package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
	"github.com/madnikulin50/lowcode/server/store"
)

// The prompt library: named, versioned instructions that AI steps and rule
// chain AI nodes refer to as "@prompt:<handle>" (see aiagent.ResolvePrompt).
//
// Editing a prompt never changes an old version - it saves a new one - so a
// flow pinned to "@prompt:x@3" keeps behaving as tested, and a bad edit is
// undone by activating the previous version.
//
// Skills are kept in the same table (kind "skill") and go through the same
// versioning; Skills() gives the view of the library that only sees them, so a
// handle is either a prompt or a skill and never both.
//
// Who may do what follows the workflows it serves: whoever may see workflows
// may read prompts, whoever may create workflows may change them.

const (
	maxPromptLen     = 20000
	maxPromptCases   = 100
	maxPromptVersion = 200 // per handle; older, inactive ones are pruned beyond it

	// skills carry more than a prompt: instructions plus reference files
	maxSkillLen          = 50000
	maxSkillFiles        = 50
	maxSkillFileLen      = 100 * 1024
	maxSkillResourcesLen = 512 * 1024
	maxSkillRequires     = 20
)

type (
	promptLibrary struct {
		store store.Storer
		ac    promptAccessController
		mux   *sync.Mutex // serialises version numbering; shared by both views
		kind  string      // "" or "prompt": prompts only; "skill": skills only
	}

	promptAccessController interface {
		CanSearchWorkflows(context.Context) bool
		CanCreateWorkflow(context.Context) bool
	}

	// PromptSummary is a prompt as the library lists it
	PromptSummary struct {
		Handle      string    `json:"handle"`
		Description string    `json:"description,omitempty"`
		Active      int       `json:"activeVersion"`
		Versions    int       `json:"versions"`
		Cases       int       `json:"cases"`
		Requires    []string  `json:"requires,omitempty"`
		Resources   int       `json:"resources,omitempty"`
		UpdatedAt   time.Time `json:"updatedAt"`
	}

	// SavePrompt is a new version of a prompt (creating the prompt if the
	// handle is new).
	SavePrompt struct {
		Handle      string
		Text        string
		Description string // empty keeps the previous one
		Note        string // why this version exists
		// Cases replaces the golden examples; nil keeps the previous version's
		Cases *types.PromptCases
		// Activate makes the new version the one "@prompt:<handle>" follows
		Activate bool

		// Skills only; nil keeps the previous version's
		Requires  *types.PromptRequires
		Resources *types.PromptFiles
	}
)

var DefaultPrompts *promptLibrary

func PromptLibrary(s store.Storer, ac promptAccessController) *promptLibrary {
	return &promptLibrary{store: s, ac: ac, mux: &sync.Mutex{}}
}

// Skills is the view of the library that holds skills instead of prompts.
func (lib *promptLibrary) Skills() *promptLibrary {
	view := *lib
	view.kind = types.PromptKindSkill
	return &view
}

func (lib *promptLibrary) isSkills() bool { return lib.kind == types.PromptKindSkill }

// noun is what this view of the library holds, for messages
func (lib *promptLibrary) noun() string {
	if lib.isSkills() {
		return "skill"
	}
	return "prompt"
}

// mine reports whether a version belongs to this view
func (lib *promptLibrary) mine(v *types.PromptVersion) bool {
	return v.IsSkill() == lib.isSkills()
}

func (lib *promptLibrary) canRead(ctx context.Context) error {
	if lib.ac == nil || !lib.ac.CanSearchWorkflows(ctx) {
		return WorkflowErrNotAllowedToSearch()
	}
	return nil
}

func (lib *promptLibrary) canWrite(ctx context.Context) error {
	if lib.ac == nil || !lib.ac.CanCreateWorkflow(ctx) {
		return WorkflowErrNotAllowedToCreate()
	}
	return nil
}

func validPromptHandle(handle string) error {
	if !aiagent.PromptHandlePattern.MatchString(handle) {
		return errors.InvalidData("prompt handle %q is not valid: use lowercase letters, digits, '_', '-' or '.', starting with a letter (up to 64 characters)", handle)
	}
	return nil
}

// versions returns the versions of a handle that belong to this view
func (lib *promptLibrary) versions(ctx context.Context, s store.Storer, handle string) (types.PromptVersionSet, error) {
	set, _, err := store.SearchAutomationPromptVersions(ctx, s, types.PromptVersionFilter{Handle: handle})
	if err != nil {
		return nil, err
	}
	mine := set[:0]
	for _, v := range set {
		if lib.mine(v) {
			mine = append(mine, v)
		}
	}
	set = mine
	sort.Slice(set, func(i, j int) bool { return set[i].Version < set[j].Version })
	return set, nil
}

// List returns one summary per prompt, by handle.
func (lib *promptLibrary) List(ctx context.Context) ([]PromptSummary, error) {
	if err := lib.canRead(ctx); err != nil {
		return nil, err
	}

	all, _, err := store.SearchAutomationPromptVersions(ctx, lib.store, types.PromptVersionFilter{})
	if err != nil {
		return nil, err
	}

	byHandle := map[string]*PromptSummary{}
	for _, v := range all {
		if !lib.mine(v) {
			continue
		}
		s := byHandle[v.Handle]
		if s == nil {
			s = &PromptSummary{Handle: v.Handle}
			byHandle[v.Handle] = s
		}
		s.Versions++
		if v.Active {
			s.Active, s.Description, s.Cases = v.Version, v.Description, len(v.Cases)
			s.Requires, s.Resources = v.Requires, len(v.Resources)
		}
		if v.CreatedAt.After(s.UpdatedAt) {
			s.UpdatedAt = v.CreatedAt
		}
	}

	out := make([]PromptSummary, 0, len(byHandle))
	for _, s := range byHandle {
		out = append(out, *s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Handle < out[j].Handle })
	return out, nil
}

// Get returns one version of a prompt; version 0 means the active one.
func (lib *promptLibrary) Get(ctx context.Context, handle string, version int) (*types.PromptVersion, error) {
	if err := lib.canRead(ctx); err != nil {
		return nil, err
	}
	return lib.find(ctx, lib.store, handle, version)
}

func (lib *promptLibrary) find(ctx context.Context, s store.Storer, handle string, version int) (*types.PromptVersion, error) {
	vv, err := lib.versions(ctx, s, handle)
	if err != nil {
		return nil, err
	}
	for _, v := range vv {
		if (version == 0 && v.Active) || (version != 0 && v.Version == version) {
			return v, nil
		}
	}
	if version == 0 {
		if len(vv) == 0 {
			return nil, errors.NotFound("%s %q does not exist", lib.noun(), handle)
		}
		return nil, errors.NotFound("%s %q has no active version", lib.noun(), handle)
	}
	return nil, errors.NotFound("%s %q has no version %d", lib.noun(), handle, version)
}

// History returns every version of a prompt, oldest first.
func (lib *promptLibrary) History(ctx context.Context, handle string) (types.PromptVersionSet, error) {
	if err := lib.canRead(ctx); err != nil {
		return nil, err
	}
	vv, err := lib.versions(ctx, lib.store, handle)
	if err != nil {
		return nil, err
	}
	if len(vv) == 0 {
		return nil, errors.NotFound("%s %q does not exist", lib.noun(), handle)
	}
	return vv, nil
}

// Save stores a new version of a prompt.
func (lib *promptLibrary) Save(ctx context.Context, in SavePrompt) (*types.PromptVersion, error) {
	if err := lib.canWrite(ctx); err != nil {
		return nil, err
	}

	in.Handle = strings.TrimSpace(in.Handle)
	if err := validPromptHandle(in.Handle); err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Text) == "" {
		return nil, errors.InvalidData("prompt text is empty")
	}
	limit := maxPromptLen
	if lib.isSkills() {
		limit = maxSkillLen
	}
	if len([]rune(in.Text)) > limit {
		return nil, errors.InvalidData("%s text is longer than %d characters", lib.noun(), limit)
	}
	if !lib.isSkills() && (in.Requires != nil || in.Resources != nil) {
		return nil, errors.InvalidData("only a skill can require toolkits or carry files")
	}
	if in.Requires != nil {
		if err := validSkillRequires(*in.Requires); err != nil {
			return nil, err
		}
	}
	if in.Resources != nil {
		if err := validSkillFiles(*in.Resources); err != nil {
			return nil, err
		}
	}
	if in.Cases != nil && len(*in.Cases) > maxPromptCases {
		return nil, errors.InvalidData("a prompt can have at most %d test cases", maxPromptCases)
	}

	ident := auth.GetIdentityFromContext(ctx)
	if ident == nil || ident.Identity() == 0 {
		return nil, errors.InvalidData("a signed-in user is required to save a prompt")
	}

	lib.mux.Lock()
	defer lib.mux.Unlock()

	var saved *types.PromptVersion
	err := store.Tx(ctx, lib.store, func(ctx context.Context, s store.Storer) error {
		vv, err := lib.versions(ctx, s, in.Handle)
		if err != nil {
			return err
		}
		if len(vv) == 0 {
			// a handle is a prompt or a skill, not both
			if taken, err := lib.takenByOther(ctx, s, in.Handle); err != nil {
				return err
			} else if taken {
				return errors.InvalidData("handle %q is already used by a %s", in.Handle, lib.other())
			}
		}

		next := &types.PromptVersion{
			ID:        nextID(),
			Handle:    in.Handle,
			Version:   1,
			Text:      in.Text,
			Note:      in.Note,
			Active:    in.Activate || len(vv) == 0, // the first version is always live
			CreatedAt: *now(),
			CreatedBy: ident.Identity(),
		}
		if lib.isSkills() {
			next.Kind = types.PromptKindSkill
		}

		if len(vv) > 0 {
			latest := vv[len(vv)-1]
			next.Version = latest.Version + 1
			next.Description = latest.Description
			next.Cases = latest.Cases
			next.Requires = latest.Requires
			next.Resources = latest.Resources
		}
		if in.Description != "" {
			next.Description = in.Description
		}
		if in.Cases != nil {
			next.Cases = *in.Cases
		}
		if in.Requires != nil {
			next.Requires = *in.Requires
		}
		if in.Resources != nil {
			next.Resources = *in.Resources
		}

		if next.Active {
			for _, v := range vv {
				if v.Active {
					v.Active = false
					if err := store.UpdateAutomationPromptVersion(ctx, s, v); err != nil {
						return err
					}
				}
			}
		}
		if err := store.CreateAutomationPromptVersion(ctx, s, next); err != nil {
			return err
		}

		// keep the history bounded: drop the oldest inactive versions
		vv = append(vv, next)
		for len(vv) > maxPromptVersion {
			dropped := false
			for i, v := range vv {
				if !v.Active {
					if err := store.DeleteAutomationPromptVersionByID(ctx, s, v.ID); err != nil {
						return err
					}
					vv = append(vv[:i], vv[i+1:]...)
					dropped = true
					break
				}
			}
			if !dropped {
				break
			}
		}

		saved = next
		return nil
	})
	return saved, err
}

// Activate makes a version the one "@prompt:<handle>" follows - the way to
// roll back, or to roll forward after testing a version that was saved
// without being activated.
func (lib *promptLibrary) Activate(ctx context.Context, handle string, version int) (*types.PromptVersion, error) {
	if err := lib.canWrite(ctx); err != nil {
		return nil, err
	}

	lib.mux.Lock()
	defer lib.mux.Unlock()

	var activated *types.PromptVersion
	err := store.Tx(ctx, lib.store, func(ctx context.Context, s store.Storer) error {
		vv, err := lib.versions(ctx, s, handle)
		if err != nil {
			return err
		}

		var target *types.PromptVersion
		for _, v := range vv {
			if v.Version == version {
				target = v
			}
		}
		if target == nil {
			return errors.NotFound("%s %q has no version %d", lib.noun(), handle, version)
		}

		for _, v := range vv {
			if v.Active && v != target {
				v.Active = false
				if err := store.UpdateAutomationPromptVersion(ctx, s, v); err != nil {
					return err
				}
			}
		}
		if !target.Active {
			target.Active = true
			if err := store.UpdateAutomationPromptVersion(ctx, s, target); err != nil {
				return err
			}
		}
		activated = target
		return nil
	})
	return activated, err
}

// Delete removes a prompt with all its versions. Steps that still refer to it
// will fail until it is saved again - on purpose, see aiagent.ResolvePrompt.
func (lib *promptLibrary) Delete(ctx context.Context, handle string) error {
	if err := lib.canWrite(ctx); err != nil {
		return err
	}

	lib.mux.Lock()
	defer lib.mux.Unlock()

	return store.Tx(ctx, lib.store, func(ctx context.Context, s store.Storer) error {
		vv, err := lib.versions(ctx, s, handle)
		if err != nil {
			return err
		}
		if len(vv) == 0 {
			return errors.NotFound("%s %q does not exist", lib.noun(), handle)
		}
		for _, v := range vv {
			if err := store.DeleteAutomationPromptVersionByID(ctx, s, v.ID); err != nil {
				return err
			}
		}
		return nil
	})
}

func (lib *promptLibrary) other() string {
	if lib.isSkills() {
		return "prompt"
	}
	return "skill"
}

// takenByOther reports whether the handle belongs to the other kind
func (lib *promptLibrary) takenByOther(ctx context.Context, s store.Storer, handle string) (bool, error) {
	set, _, err := store.SearchAutomationPromptVersions(ctx, s, types.PromptVersionFilter{Handle: handle})
	if err != nil {
		return false, err
	}
	for _, v := range set {
		if !lib.mine(v) {
			return true, nil
		}
	}
	return false, nil
}

// resolve is the library as aiagent sees it: reads need no user permission,
// because it is a running workflow or chain asking, not a person.
func (lib *promptLibrary) resolve(ctx context.Context, handle string, version int) (*aiagent.ResolvedPrompt, error) {
	v, err := lib.find(ctx, lib.store, handle, version)
	if err != nil {
		return nil, err
	}
	return &aiagent.ResolvedPrompt{Text: v.Text, Ref: fmt.Sprintf("%s@%d", v.Handle, v.Version)}, nil
}

// Eval runs the golden examples over the given versions (none given: the
// active one) and returns a report per version, so two versions can be
// compared on the same cases. cases, when not empty, replaces the stored ones
// for this run only.
//
// run is how the model is called; agent and model are optional.
//
// onCase, when not nil, is told about each case as it is judged (with the
// version it belongs to).
func (lib *promptLibrary) Eval(ctx context.Context, run aiagent.Runner, handle string, versions []int, agent, model string, cases types.PromptCases, onCase func(version int, res aiagent.EvalResult)) (map[int]*aiagent.EvalReport, error) {
	if err := lib.canRead(ctx); err != nil {
		return nil, err
	}
	if len(versions) == 0 {
		versions = []int{0}
	}

	out := make(map[int]*aiagent.EvalReport, len(versions))
	for _, n := range versions {
		v, err := lib.find(ctx, lib.store, handle, n)
		if err != nil {
			return nil, err
		}

		use := cases
		if len(use) == 0 {
			use = v.Cases
		}
		if len(use) == 0 {
			return nil, errors.InvalidData("prompt %q version %d has no test cases to evaluate against", handle, v.Version)
		}

		spec := aiagent.EvalSpec{
			Agent:     agent,
			Model:     model,
			Prompt:    v.Text,
			PromptRef: fmt.Sprintf("%s@%d", v.Handle, v.Version),
		}
		for _, c := range use {
			spec.Cases = append(spec.Cases, aiagent.EvalCase{Name: c.Name, Inputs: c.Inputs, Expect: c.Expect, Contains: c.Contains})
		}

		if onCase != nil {
			version := v.Version
			spec.OnCase = func(res aiagent.EvalResult) { onCase(version, res) }
		}
		out[v.Version] = aiagent.RunEval(ctx, run, spec)
	}
	return out, nil
}
