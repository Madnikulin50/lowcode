package aiagent

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const sampleSkill = `---
name: Read Drawings
description: Use when the user uploads architectural drawings and wants quantities.
requires: [compose.records]
license: MIT
---

# Steps

1. Read the sheet list.
`

func TestParseSkillMarkdown(t *testing.T) {
	req := require.New(t)

	doc, err := ParseSkillMarkdown([]byte(sampleSkill))
	req.NoError(err)
	req.Equal("read-drawings", doc.Handle)
	req.Contains(doc.Description, "architectural drawings")
	req.Equal([]string{"compose.records"}, doc.Requires)
	req.Equal("# Steps\n\n1. Read the sheet list.", doc.Body)

	for name, src := range map[string]string{
		"no header":      "# just text",
		"unclosed":       "---\nname: x\n",
		"no description": "---\nname: x1\n---\nbody",
		"bad name":       "---\nname: '!!!'\ndescription: d\n---\nbody",
		"no body":        "---\nname: x1\ndescription: d\n---\n",
	} {
		_, err := ParseSkillMarkdown([]byte(src))
		req.Error(err, name)
	}

	// windows line endings and a BOM
	_, err = ParseSkillMarkdown([]byte("\xef\xbb\xbf---\r\nname: a1\r\ndescription: d\r\n---\r\nbody\r\n"))
	req.NoError(err)
}

func TestSkillZipRoundTrip(t *testing.T) {
	req := require.New(t)

	in := Skill{
		Handle: "read-drawings", Description: "when", Body: "Do it.",
		Requires:  []string{"compose.records"},
		Resources: []SkillFile{{Path: "refs/prices.csv", Content: "a,b\n1,2\n"}},
	}
	data, err := BuildSkillZip(in)
	req.NoError(err)

	doc, err := ParseSkillZip(data)
	req.NoError(err)
	req.Equal("read-drawings", doc.Handle)
	req.Equal("Do it.", doc.Body)
	req.Equal([]SkillFile{{Path: "refs/prices.csv", Mime: "text/csv", Content: "a,b\n1,2\n"}}, doc.Resources)
}

func zipOf(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range files {
		w, err := zw.Create(name)
		require.NoError(t, err)
		_, _ = w.Write([]byte(content))
	}
	require.NoError(t, zw.Close())
	return buf.Bytes()
}

func TestParseSkillZip_Rejects(t *testing.T) {
	req := require.New(t)

	// archive of a folder: SKILL.md one level down is fine
	doc, err := ParseSkillZip(zipOf(t, map[string]string{"my/SKILL.md": sampleSkill, "my/notes.txt": "hi"}))
	req.NoError(err)
	req.Equal("notes.txt", doc.Resources[0].Path)

	_, err = ParseSkillZip(zipOf(t, map[string]string{"readme.txt": "x"}))
	req.Error(err, "no SKILL.md")

	_, err = ParseSkillZip(zipOf(t, map[string]string{"SKILL.md": sampleSkill, "bin.dat": "a\x00b"}))
	req.Error(err, "binary file")

	_, err = ParseSkillZip(zipOf(t, map[string]string{"SKILL.md": sampleSkill, "../evil.txt": "x"}))
	req.Error(err, "path escaping the skill")

	_, err = ParseSkillZip([]byte("not a zip"))
	req.Error(err)
}

func TestSafeSkillPath(t *testing.T) {
	for in, ok := range map[string]bool{
		"refs/a.md": true, "./a.md": true, "a\\b.md": true,
		"": false, "/etc/passwd": false, "../a": false, "a/../../b": false, "SKILL.md": false, "skill.md": false,
	} {
		_, got := SafeSkillPath(in)
		require.Equal(t, ok, got, in)
	}
}

func TestLoadSkillDir(t *testing.T) {
	req := require.New(t)
	dir := t.TempDir()

	req.NoError(os.WriteFile(filepath.Join(dir, "one.md"), []byte(sampleSkill), 0o600))
	req.NoError(os.WriteFile(filepath.Join(dir, "broken.md"), []byte("# no header"), 0o600))
	req.NoError(os.WriteFile(filepath.Join(dir, "ignored.txt"), []byte("x"), 0o600))

	sub := filepath.Join(dir, "folder")
	req.NoError(os.MkdirAll(filepath.Join(sub, "refs"), 0o700))
	req.NoError(os.WriteFile(filepath.Join(sub, "SKILL.md"), []byte("---\nname: with-files\ndescription: d\n---\nbody"), 0o600))
	req.NoError(os.WriteFile(filepath.Join(sub, "refs", "a.md"), []byte("ref"), 0o600))

	skills, errs := LoadSkillDir(dir)
	req.Len(errs, 1, "the broken one is reported, not fatal")
	req.Len(skills, 2)

	byHandle := map[string]Skill{}
	for _, s := range skills {
		byHandle[s.Handle] = s
		req.Equal(SkillSourceFile, s.Source)
	}
	req.Contains(byHandle, "read-drawings")
	req.Equal("refs/a.md", byHandle["with-files"].Resources[0].Path)
}

type fakeSkillSource struct{ skills []Skill }

func (f fakeSkillSource) ListSkills(context.Context) ([]Skill, error) { return f.skills, nil }
func (f fakeSkillSource) GetSkill(_ context.Context, h string, _ int) (*Skill, error) {
	for i := range f.skills {
		if f.skills[i].Handle == h {
			return &f.skills[i], nil
		}
	}
	return nil, os.ErrNotExist
}

func TestSkillCatalog_SourcePrecedence(t *testing.T) {
	req := require.New(t)
	t.Cleanup(func() {
		SetSkillSource(SkillSourceDB, nil)
		SetSkillSource(SkillSourceFile, nil)
	})

	SetSkillSource(SkillSourceFile, fakeSkillSource{[]Skill{{Handle: "a", Description: "from file"}, {Handle: "b", Description: "only file"}}})
	SetSkillSource(SkillSourceDB, fakeSkillSource{[]Skill{{Handle: "a", Description: "from db"}}})

	list := ListSkills(context.Background())
	req.Len(list, 2)
	req.Equal("from db", list[0].Description, "database wins over files")

	s, err := GetSkill(context.Background(), "a", 0)
	req.NoError(err)
	req.Equal("from db", s.Description)

	_, err = GetSkill(context.Background(), "zzz", 0)
	req.ErrorContains(err, "does not exist")

	req.Len(SkillsFor(context.Background(), nil), 0)
	req.Len(SkillsFor(context.Background(), []string{"*"}), 2)
	req.Len(SkillsFor(context.Background(), []string{"b"}), 1)
}
