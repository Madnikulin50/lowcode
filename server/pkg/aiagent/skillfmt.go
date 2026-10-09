package aiagent

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// The SKILL.md format: a markdown file that starts with a YAML header,
//
//	---
//	name: read-drawings
//	description: When to use it ...
//	requires: [compose.records]
//	---
//	# Instructions ...
//
// Anything else the header holds (license, allowed-tools, metadata) is
// ignored. A skill with files is a directory (or a zip) with SKILL.md at its
// root and the files beside it.

const (
	SkillFileName = "SKILL.md"

	// limits on what is read from a zip or a directory; the library enforces
	// its own, stricter ones when the skill is saved
	maxSkillArchiveFiles = 200
	maxSkillArchiveBytes = 8 << 20
)

type (
	// SkillDoc is a skill as a document: what SKILL.md and a zip carry
	SkillDoc struct {
		Handle      string
		Description string
		Requires    []string
		Body        string
		Resources   []SkillFile
	}

	skillHeader struct {
		Name        string   `yaml:"name"`
		Description string   `yaml:"description"`
		Requires    []string `yaml:"requires,omitempty"`
		// the names other tools use for the same idea
		AllowedTools []string `yaml:"allowed-tools,omitempty"`
	}
)

// ParseSkillMarkdown reads a SKILL.md. The name must make a valid handle and
// the description may not be empty.
func ParseSkillMarkdown(raw []byte) (SkillDoc, error) {
	var doc SkillDoc

	text := strings.TrimPrefix(string(raw), "\ufeff")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return doc, fmt.Errorf("a skill must start with a YAML header between '---' lines (name, description)")
	}
	rest := text[4:]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return doc, fmt.Errorf("the YAML header of the skill is not closed with '---'")
	}
	header, body := rest[:end], rest[end+4:]
	body = strings.TrimPrefix(body, "\n")

	var h skillHeader
	if err := yaml.Unmarshal([]byte(header), &h); err != nil {
		return doc, fmt.Errorf("the YAML header of the skill is not valid: %w", err)
	}

	doc.Handle = SkillHandleFromName(h.Name)
	if doc.Handle == "" {
		return doc, fmt.Errorf("skill name %q does not make a valid handle: use lowercase letters, digits, '_', '-' or '.', starting with a letter", h.Name)
	}
	doc.Description = strings.TrimSpace(h.Description)
	if doc.Description == "" {
		return doc, fmt.Errorf("skill %q has no description - it says when the skill applies, and an agent picks skills by it", doc.Handle)
	}
	doc.Requires = h.Requires
	doc.Body = strings.TrimSpace(body)
	if doc.Body == "" {
		return doc, fmt.Errorf("skill %q has no instructions after the header", doc.Handle)
	}
	return doc, nil
}

// RenderSkillMarkdown writes the SKILL.md for a skill.
func RenderSkillMarkdown(s Skill) ([]byte, error) {
	h := skillHeader{Name: s.Handle, Description: s.Description, Requires: s.Requires}
	head, err := yaml.Marshal(h)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	b.WriteString("---\n")
	b.Write(head)
	b.WriteString("---\n\n")
	b.WriteString(strings.TrimSpace(s.Body))
	b.WriteString("\n")
	return b.Bytes(), nil
}

// SafeSkillPath reports whether a file path stays inside the skill: relative,
// no ".." and not SKILL.md itself.
func SafeSkillPath(p string) (string, bool) {
	p = cleanSkillPath(p)
	if p == "" || strings.HasPrefix(p, "/") || strings.Contains(p, "\x00") {
		return "", false
	}
	p = path.Clean(p)
	if p == "." || p == ".." || strings.HasPrefix(p, "../") || strings.EqualFold(p, SkillFileName) {
		return "", false
	}
	return p, true
}

// ParseSkillZip reads a skill archive: SKILL.md at the root (or inside the one
// top-level directory, the way archives of a skill folder come) and its files.
// Files that are not text are an error rather than being dropped silently.
func ParseSkillZip(data []byte) (SkillDoc, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return SkillDoc{}, fmt.Errorf("not a zip archive: %w", err)
	}
	if len(zr.File) > maxSkillArchiveFiles {
		return SkillDoc{}, fmt.Errorf("the archive has more than %d files", maxSkillArchiveFiles)
	}

	// find the skill's root: the directory holding SKILL.md
	root, found := "", false
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := cleanSkillPath(f.Name)
		if path.Base(name) == SkillFileName && strings.Count(name, "/") <= 1 {
			root, found = path.Dir(name), true
			if root == "." {
				root = ""
				break
			}
		}
	}
	if !found {
		return SkillDoc{}, fmt.Errorf("the archive has no %s", SkillFileName)
	}

	var (
		doc     SkillDoc
		total   int64
		haveDoc bool
		files   []SkillFile
	)
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := cleanSkillPath(f.Name)
		if root != "" {
			if !strings.HasPrefix(name, root+"/") {
				continue
			}
			name = strings.TrimPrefix(name, root+"/")
		}
		if strings.HasPrefix(path.Base(name), ".") || strings.HasPrefix(name, "__MACOSX/") {
			continue
		}

		total += int64(f.UncompressedSize64)
		if total > maxSkillArchiveBytes {
			return SkillDoc{}, fmt.Errorf("the archive holds more than %d MB", maxSkillArchiveBytes>>20)
		}
		rc, err := f.Open()
		if err != nil {
			return SkillDoc{}, err
		}
		content, err := io.ReadAll(io.LimitReader(rc, maxSkillArchiveBytes+1))
		rc.Close()
		if err != nil {
			return SkillDoc{}, err
		}

		if name == SkillFileName {
			if doc, err = ParseSkillMarkdown(content); err != nil {
				return SkillDoc{}, err
			}
			haveDoc = true
			continue
		}

		p, ok := SafeSkillPath(name)
		if !ok {
			return SkillDoc{}, fmt.Errorf("file path %q is not allowed in a skill", f.Name)
		}
		if !utf8.Valid(content) || bytes.IndexByte(content, 0) >= 0 {
			return SkillDoc{}, fmt.Errorf("file %q is not a text file; a skill can only carry text", p)
		}
		files = append(files, SkillFile{Path: p, Mime: skillMime(p), Content: string(content)})
	}
	if !haveDoc {
		return SkillDoc{}, fmt.Errorf("the archive has no %s", SkillFileName)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	doc.Resources = files
	return doc, nil
}

// BuildSkillZip packs a skill: SKILL.md and its files.
func BuildSkillZip(s Skill) ([]byte, error) {
	md, err := RenderSkillMarkdown(s)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	put := func(name string, content []byte) error {
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		_, err = w.Write(content)
		return err
	}
	if err := put(SkillFileName, md); err != nil {
		return nil, err
	}
	for _, f := range s.Resources {
		p, ok := SafeSkillPath(f.Path)
		if !ok {
			continue
		}
		if err := put(p, []byte(f.Content)); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func skillMime(p string) string {
	switch strings.ToLower(path.Ext(p)) {
	case ".md", ".markdown":
		return "text/markdown"
	case ".json":
		return "application/json"
	case ".csv":
		return "text/csv"
	case ".yaml", ".yml":
		return "application/yaml"
	case ".html", ".htm":
		return "text/html"
	}
	return "text/plain"
}

// LoadSkillDir reads skills from a directory: a "<name>.md" file with a
// header is a skill without files, a sub-directory holding SKILL.md is a skill
// with them. Entries that are not valid skills are reported in the second
// result and skipped, so one bad file does not hide the rest.
func LoadSkillDir(dir string) ([]Skill, []error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, []error{err}
	}

	var (
		out  []Skill
		errs []error
	)
	for _, e := range entries {
		full := filepath.Join(dir, e.Name())
		var (
			doc SkillDoc
			err error
		)
		switch {
		case e.IsDir():
			if _, statErr := os.Stat(filepath.Join(full, SkillFileName)); statErr != nil {
				continue
			}
			doc, err = readSkillDir(full)
		case strings.EqualFold(filepath.Ext(e.Name()), ".md"):
			var raw []byte
			if raw, err = os.ReadFile(full); err == nil {
				doc, err = ParseSkillMarkdown(raw)
			}
		default:
			continue
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", full, err))
			continue
		}
		out = append(out, Skill{
			Handle: doc.Handle, Description: doc.Description, Body: doc.Body,
			Requires: doc.Requires, Resources: doc.Resources, Source: SkillSourceFile,
		})
	}
	return out, errs
}

func readSkillDir(dir string) (SkillDoc, error) {
	raw, err := os.ReadFile(filepath.Join(dir, SkillFileName))
	if err != nil {
		return SkillDoc{}, err
	}
	doc, err := ParseSkillMarkdown(raw)
	if err != nil {
		return doc, err
	}

	var total int64
	err = filepath.WalkDir(dir, func(p string, d os.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		if d.IsDir() || strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		rel, _ := filepath.Rel(dir, p)
		rel = filepath.ToSlash(rel)
		if rel == SkillFileName {
			return nil
		}
		if len(doc.Resources) >= maxSkillArchiveFiles {
			return fmt.Errorf("more than %d files", maxSkillArchiveFiles)
		}
		content, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		total += int64(len(content))
		if total > maxSkillArchiveBytes {
			return fmt.Errorf("more than %d MB of files", maxSkillArchiveBytes>>20)
		}
		if !utf8.Valid(content) || bytes.IndexByte(content, 0) >= 0 {
			return fmt.Errorf("file %q is not a text file", rel)
		}
		doc.Resources = append(doc.Resources, SkillFile{Path: rel, Mime: skillMime(rel), Content: string(content)})
		return nil
	})
	return doc, err
}

// DirSkillSource serves skills kept as files. The directory is read again on
// every call: it is small, and an edit shows up without a restart.
type DirSkillSource struct {
	Dir string
}

func (d DirSkillSource) ListSkills(context.Context) ([]Skill, error) {
	skills, _ := LoadSkillDir(d.Dir)
	return skills, nil
}

func (d DirSkillSource) GetSkill(_ context.Context, handle string, version int) (*Skill, error) {
	skills, _ := LoadSkillDir(d.Dir)
	for i := range skills {
		if skills[i].Handle == handle {
			return &skills[i], nil
		}
	}
	return nil, fmt.Errorf("skill %q does not exist", handle)
}
