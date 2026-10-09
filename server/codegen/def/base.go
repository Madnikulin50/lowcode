// Package def replaces the CUE schema (codegen/schema/*.cue) and component
// definitions (compose/*.cue, automation/*.cue, federation/*.cue) with plain
// Go values. It mirrors the field derivation rules of codegen/schema/*.cue
// exactly, so that the JSON payload it produces can be fed, unchanged, into
// the existing codegen/tool (jsontplexec) binary and templates.
package def

import "strings"

// splitWords mirrors the CUE pattern of replacing separator characters with
// spaces and splitting into words, e.g. "shared_node_id" -> ["shared","node","id"].
func splitWords(s string, seps ...string) []string {
	for _, sep := range seps {
		s = strings.ReplaceAll(s, sep, " ")
	}
	return strings.Fields(s)
}

// titleFirst mirrors CUE's strings.ToTitle applied to a single word: it
// upper-cases only the first rune and leaves the rest untouched.
func titleFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	return strings.ToUpper(string(r[0])) + string(r[1:])
}

// pascal mirrors strings.ToTitle(words joined by space) with spaces removed:
// every word gets its first rune upper-cased, the rest of each word is left
// as-is, and the words are concatenated.
func pascal(words []string) string {
	var b strings.Builder
	for _, w := range words {
		b.WriteString(titleFirst(w))
	}
	return b.String()
}

// camel mirrors CUE's strings.ToCamel: lower-case only the first rune.
func camel(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	return strings.ToLower(string(r[0])) + string(r[1:])
}

// Base mirrors codegen/schema/model.cue's #_base: handle -> ident/expIdent
// derivation, with explicit overrides taking precedence over the default.
type Base struct {
	Handle   string
	Ident    string // optional override
	ExpIdent string // optional override
}

// ResolvedBase is the fully defaulted form of Base, ready to be marshaled.
type ResolvedBase struct {
	Handle         string `json:"handle"`
	Ident          string `json:"ident"`
	ExpIdent       string `json:"expIdent"`
	IdentPlural    string `json:"identPlural"`
	ExpIdentPlural string `json:"expIdentPlural"`
}

func (b Base) Resolve() ResolvedBase {
	p := pascal(splitWords(b.Handle, "-", "_", "."))

	ident := b.Ident
	if ident == "" {
		ident = camel(p)
	}

	expIdent := b.ExpIdent
	if expIdent == "" {
		expIdent = p
	}

	return ResolvedBase{
		Handle:         b.Handle,
		Ident:          ident,
		ExpIdent:       expIdent,
		IdentPlural:    ident + "s",
		ExpIdentPlural: expIdent + "s",
	}
}
