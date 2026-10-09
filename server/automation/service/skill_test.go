package service

import (
	"testing"

	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/stretchr/testify/require"
)

func TestSkillLibrary_SeparateFromPrompts(t *testing.T) {
	lib, ctx := newPromptLib(t, fakePromptAC{true, true})
	skills := lib.Skills()
	req := require.New(t)

	requires := types.PromptRequires{"compose.records"}
	files := types.PromptFiles{{Path: "refs/prices.csv", Content: "a,b"}}

	v1, err := skills.Save(ctx, SavePrompt{
		Handle: "estimate", Text: "Do the estimate.", Description: "When estimating",
		Requires: &requires, Resources: &files,
	})
	req.NoError(err)
	req.True(v1.IsSkill())
	req.Equal(1, v1.Version)

	// the second version carries requires and files over
	v2, err := skills.Save(ctx, SavePrompt{Handle: "estimate", Text: "Better."})
	req.NoError(err)
	req.Equal(requires, v2.Requires)
	req.Equal(files, v2.Resources)

	// prompts and skills do not see each other
	_, err = lib.Save(ctx, SavePrompt{Handle: "triage", Text: "Rate urgency."})
	req.NoError(err)

	pp, err := lib.List(ctx)
	req.NoError(err)
	req.Len(pp, 1)
	req.Equal("triage", pp[0].Handle)

	ss, err := skills.List(ctx)
	req.NoError(err)
	req.Len(ss, 1)
	req.Equal("estimate", ss[0].Handle)
	req.Equal(1, ss[0].Resources)

	_, err = lib.Get(ctx, "estimate", 0)
	req.Error(err, "a skill is not a prompt")
	_, err = skills.Get(ctx, "triage", 0)
	req.Error(err)

	// a handle is one or the other
	_, err = lib.Save(ctx, SavePrompt{Handle: "estimate", Text: "x"})
	req.ErrorContains(err, "already used by a skill")
	_, err = skills.Save(ctx, SavePrompt{Handle: "triage", Text: "x"})
	req.ErrorContains(err, "already used by a prompt")

	// and "@prompt:estimate" must not pick up a skill
	_, err = lib.resolve(ctx, "estimate", 0)
	req.Error(err)
}

func TestSkillLibrary_Validation(t *testing.T) {
	lib, ctx := newPromptLib(t, fakePromptAC{true, true})
	skills := lib.Skills()
	req := require.New(t)

	for name, files := range map[string]types.PromptFiles{
		"escaping path": {{Path: "../x", Content: "a"}},
		"skill.md":      {{Path: "SKILL.md", Content: "a"}},
		"duplicate":     {{Path: "a", Content: "1"}, {Path: "a", Content: "2"}},
		"too big":       {{Path: "a", Content: string(make([]byte, maxSkillFileLen+1))}},
	} {
		f := files
		_, err := skills.Save(ctx, SavePrompt{Handle: "s1", Text: "x", Resources: &f})
		req.Error(err, name)
	}

	bad := types.PromptRequires{"Not Valid"}
	_, err := skills.Save(ctx, SavePrompt{Handle: "s1", Text: "x", Requires: &bad})
	req.Error(err)

	// a plain prompt cannot carry skill parts
	ok := types.PromptRequires{"compose.records"}
	_, err = lib.Save(ctx, SavePrompt{Handle: "p1", Text: "x", Requires: &ok})
	req.Error(err)
}

func TestSkill_ImportExport(t *testing.T) {
	lib, ctx := newPromptLib(t, fakePromptAC{true, true})
	DefaultPrompts = lib
	t.Cleanup(func() { DefaultPrompts = nil })
	req := require.New(t)

	md := []byte("---\nname: Check Fire Safety\ndescription: When fire rules matter.\nrequires: [compose.records]\n---\nCheck the rules.")
	b, err := ImportSkill(ctx, md, "", "", true)
	req.NoError(err)
	req.Equal("check-fire-safety", b.Handle)
	req.True(b.Active)
	req.Equal("imported", b.Note)

	got, err := GetSkill(ctx, "check-fire-safety", 0, false)
	req.NoError(err)
	req.Equal("Check the rules.", got.Text)
	req.Equal([]string{"compose.records"}, got.Requires)

	// no files: exported as markdown, and it imports back
	data, name, err := ExportSkill(ctx, "check-fire-safety", 0, false)
	req.NoError(err)
	req.Equal("check-fire-safety.md", name)
	again, err := ImportSkill(ctx, data, "copy-of-it", "", false)
	req.NoError(err)
	req.Equal("copy-of-it", again.Handle)

	// with files: a zip
	files := types.PromptFiles{{Path: "a.txt", Content: "hi"}}
	_, err = SaveSkillVersion(ctx, SavePrompt{Handle: "check-fire-safety", Text: "v2", Resources: &files, Activate: true})
	req.NoError(err)
	data, name, err = ExportSkill(ctx, "check-fire-safety", 0, false)
	req.NoError(err)
	req.Equal("check-fire-safety.zip", name)
	z, err := ImportSkill(ctx, data, "zipped", "", true)
	req.NoError(err)
	req.Len(z.Resources, 1)
}
