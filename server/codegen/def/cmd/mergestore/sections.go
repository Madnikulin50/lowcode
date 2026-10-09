package main

import (
	"fmt"
	"sort"
	"strings"
)

// extractParenBlock finds the line that is exactly openLine (trimmed) at
// depth 0, and the following line that is exactly ")" at depth 1 (i.e. the
// matching close), and splits src into:
//
//	before: everything up to and including the openLine
//	inner:  everything strictly between (no leading/trailing blank trim)
//	after:  the closing ")" line and everything after it
func extractParenBlock(src, openLine string) (before, inner, after string, err error) {
	lines := strings.Split(src, "\n")
	depths := lineDepths(src)

	start := -1
	for i, l := range lines {
		if strings.TrimSpace(l) == openLine && depths[i] == 0 {
			start = i
			break
		}
	}
	if start == -1 {
		return "", "", "", fmt.Errorf("could not find opening line %q", openLine)
	}

	end := -1
	for i := start + 1; i < len(lines); i++ {
		if depths[i] == 1 && strings.TrimSpace(lines[i]) == ")" {
			end = i
			break
		}
	}
	if end == -1 {
		return "", "", "", fmt.Errorf("could not find matching close for %q", openLine)
	}

	before = strings.Join(lines[:start+1], "\n") + "\n"
	inner = strings.Join(lines[start+1:end], "\n")
	after = strings.Join(lines[end:], "\n")
	return before, inner, after, nil
}

// splitHeader pulls apart "package X\n\n<comment>\n\nimport (\n...\n)\n\n"
// from the rest of the file. Returns the part before the import block
// (package + header comment), the raw import lines (trimmed, one per
// line, blank lines dropped), and everything after the import block.
func splitHeader(src string) (preamble string, importLines []string, rest string, err error) {
	before, inner, after, err := extractParenBlock(src, "import (")
	if err != nil {
		return "", nil, "", err
	}

	for _, l := range strings.Split(inner, "\n") {
		if strings.TrimSpace(l) == "" {
			continue
		}
		importLines = append(importLines, strings.TrimSpace(l))
	}

	// after starts with the ")" line; drop it and the blank line that follows.
	afterLines := strings.Split(after, "\n")
	rest = strings.TrimLeft(strings.Join(afterLines[1:], "\n"), "\n")

	return before, importLines, rest, nil
}

// importPath extracts the quoted import path from an import line, which is
// either `"path"` or `alias "path"`.
func importPath(line string) string {
	i := strings.LastIndex(line, "\"")
	j := strings.LastIndex(line[:i], "\"")
	if i < 0 || j < 0 {
		return line
	}
	return line[j : i+1]
}

// mergeImports unions old and fresh import lines by import path (fresh
// wins on conflict), sorted by path - matching the map[path]alias
// ordering codegen/def's BuildStoreTasks already produces.
func mergeImports(oldLines, freshLines []string) []string {
	byPath := map[string]string{}
	for _, l := range oldLines {
		byPath[importPath(l)] = l
	}
	for _, l := range freshLines {
		byPath[importPath(l)] = l
	}

	paths := make([]string, 0, len(byPath))
	for p := range byPath {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	out := make([]string, len(paths))
	for i, p := range paths {
		out[i] = byPath[p]
	}
	return out
}

// renderImportBlock renders import lines plus the closing ")\n\n". The
// caller's preamble (from splitHeader) already ends with "import (\n".
func renderImportBlock(lines []string) string {
	var b strings.Builder
	for _, l := range lines {
		b.WriteString("\t")
		b.WriteString(l)
		b.WriteString("\n")
	}
	b.WriteString(")\n\n")
	return b.String()
}
