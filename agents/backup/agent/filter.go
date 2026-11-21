package agent

import (
	"context"
	"path"
	"strings"
)

// Exclude-паттерны политики: что не бэкапить. Один паттерн на строку
// (запятая тоже разделитель — удобно из однострочного поля Compose).
//
// Семантика:
//   - glob path.Match по относительному пути ("docs/*.tmp");
//   - паттерн без '/' матчит и по base name ("*.tmp");
//   - имя каталога без слэша матчит всё поддерево ("node_modules");
//   - суффикс "/**" эквивалентен имени каталога ("cache/**" == "cache").

func parseExclude(spec string) []string {
	var out []string
	for _, line := range strings.Split(spec, "\n") {
		for _, pat := range strings.Split(line, ",") {
			pat = strings.TrimSpace(pat)
			if pat == "" || strings.HasPrefix(pat, "#") {
				continue
			}
			out = append(out, pat)
		}
	}
	return out
}

func matchExclude(rel string, patterns []string) bool {
	rel = strings.Trim(path.Clean(strings.ReplaceAll(rel, "\\", "/")), "/")
	base := path.Base(rel)
	for _, pat := range patterns {
		p := strings.Trim(strings.ReplaceAll(pat, "\\", "/"), "/")
		if p == "" {
			continue
		}
		if strings.HasSuffix(p, "/**") {
			p = strings.TrimSuffix(p, "/**")
			if rel == p || strings.HasPrefix(rel, p+"/") {
				return true
			}
			continue
		}
		if ok, _ := path.Match(p, rel); ok {
			return true
		}
		if !strings.Contains(p, "/") {
			// base name / имя каталога на любом уровне
			if ok, _ := path.Match(p, base); ok {
				return true
			}
			for _, part := range strings.Split(rel, "/") {
				if ok, _ := path.Match(p, part); ok {
					return true
				}
			}
		}
	}
	return false
}

// excludeWalker оборачивает FileWalker и пропускает исключённые пути.
type excludeWalker struct {
	inner    FileWalker
	patterns []string
}

func newExcludeWalker(inner FileWalker, spec string) FileWalker {
	patterns := parseExclude(spec)
	if len(patterns) == 0 {
		return inner
	}
	return &excludeWalker{inner: inner, patterns: patterns}
}

func (e *excludeWalker) Walk(ctx context.Context, fn func(FileEntry, openFn) error) error {
	return e.inner.Walk(ctx, func(entry FileEntry, open openFn) error {
		if matchExclude(entry.RelPath, e.patterns) {
			return nil
		}
		return fn(entry, open)
	})
}
