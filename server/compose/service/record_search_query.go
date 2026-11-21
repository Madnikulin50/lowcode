package service

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/madnikulin50/lowcode/server/compose/types"
)

const (
	defaultRecordSearchLimit uint = 50
	maxRecordSearchLimit     uint = 200
)

var RecordFieldIdent = regexp.MustCompile(`^[A-Za-z][0-9A-Za-z_-]*$`)

func SanitizeRecordSearchQuery(query string) string {
	var b strings.Builder
	for _, r := range query {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) || r == '-' || r == '.' {
			b.WriteRune(r)
		}
	}
	s := strings.TrimSpace(strings.Join(strings.Fields(b.String()), " "))
	s = strings.ReplaceAll(s, `'`, `''`)
	return s
}

func ClampRecordSearchLimit(n uint) uint {
	if n == 0 {
		return defaultRecordSearchLimit
	}
	if n > maxRecordSearchLimit {
		return maxRecordSearchLimit
	}
	return n
}

func BuildRecordTextSearchQL(fields types.ModuleFieldSet, query string) string {
	q := SanitizeRecordSearchQuery(query)
	if q == "" {
		return ""
	}
	textKinds := map[string]bool{"String": true, "Text": true, "URL": true, "Email": true}
	var conditions []string
	for _, f := range fields {
		if f == nil || !textKinds[f.Kind] || !RecordFieldIdent.MatchString(f.Name) {
			continue
		}
		conditions = append(conditions, fmt.Sprintf("%s LIKE '%%%s%%'", f.Name, q))
	}
	return strings.Join(conditions, " OR ")
}
