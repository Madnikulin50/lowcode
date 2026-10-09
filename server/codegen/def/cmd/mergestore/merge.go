package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// keyedBlock is one top-level chunk of generated code tagged with the
// resource/field identifier it belongs to.
type keyedBlock struct {
	key  string
	text string
}

// groupAdjacent merges consecutive units that share the same key into one
// block (preserving their internal order and the blank line between them).
// Per-resource output in these templates is always contiguous, so this
// turns "8 vars for federationNode, 8 vars for federationNodeSync, ..."
// into one block per resource.
func groupAdjacent(units []keyedBlock) []keyedBlock {
	var out []keyedBlock
	for _, u := range units {
		if len(out) > 0 && out[len(out)-1].key == u.key {
			out[len(out)-1].text += "\n\n" + u.text
			continue
		}
		out = append(out, u)
	}
	return out
}

// mergeKeyed unions old and fresh keyed blocks (fresh wins on key
// conflict), sorted by key to match the alphabetical ordering the
// templates always produce (Go map/range sorts by key).
func mergeKeyed(old, fresh []keyedBlock) []keyedBlock {
	byKey := map[string]string{}
	var keys []string
	add := func(bs []keyedBlock) {
		for _, b := range bs {
			if _, ok := byKey[b.key]; !ok {
				keys = append(keys, b.key)
			}
			byKey[b.key] = b.text
		}
	}
	add(old)
	add(fresh)
	sort.Strings(keys)

	out := make([]keyedBlock, len(keys))
	for i, k := range keys {
		out[i] = keyedBlock{key: k, text: byKey[k]}
	}
	return out
}

func joinBlocks(bs []keyedBlock) string {
	parts := make([]string, len(bs))
	for i, b := range bs {
		parts[i] = b.text
	}
	return strings.Join(parts, "\n\n")
}

var (
	reInterfaceDecl = regexp.MustCompile(`^(\w+) interface \{`)
	reWrapperParam  = regexp.MustCompile(`, s ([A-Za-z][A-Za-z0-9]*)[,)]`)
	reStructDecl    = regexp.MustCompile(`^(\w+) struct \{`)
	reAuxReceiver   = regexp.MustCompile(`func \(aux \*?(\w+)\)`)
	reQueryVarName  = regexp.MustCompile(`(?m)^(\w+?)(Table|SelectQuery|InsertQuery|UpsertQuery|UpdateQuery|DeleteQuery|TruncateQuery|PrimaryKeys) = `)
	reFilterField   = regexp.MustCompile(`^(\w+) func\(`)
	reFilterFunc    = regexp.MustCompile(`^func (\w+)Filter\(`)
	reFuncName      = regexp.MustCompile(`^func (\w+)\(`)
)

// identFromPlural reconstructs the lowercase, singular store ident from an
// exported plural name (e.g. "ComposeModuleFields" -> "composeModuleField").
// This is the true sort key these templates use (Go template map-range
// sorts by the JSON payload's map key, which is res.store.ident) - the
// plural display name itself does NOT sort the same way in every case
// (e.g. "ComposeModuleFields" < "ComposeModules" as plain strings, but
// "composeModuleField" > "composeModule" as idents, since pluralizing
// "composeModule" to "composeModules" breaks the prefix relationship that
// "composeModule"/"composeModuleField" have). expIdentPlural is always
// exactly expIdent+"s" (see codegen/def's Base/StoreConfig), so this
// reversal is exact, not a heuristic.
func identFromPlural(plural string) string {
	singular := strings.TrimSuffix(plural, "s")
	if singular == "" {
		return plural
	}
	r := []rune(singular)
	return strings.ToLower(string(r[0])) + string(r[1:])
}

func firstMatch(re *regexp.Regexp, s string) (string, bool) {
	// search line by line so a leading doc comment doesn't confuse `^`
	for _, line := range strings.Split(s, "\n") {
		if m := re.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			return m[1], true
		}
	}
	return "", false
}

// --- interfaces.gen.go -----------------------------------------------------

// storerPreambleAnchor is the last line of the Storer interface's fixed
// (non-generated) method set, right before the embedded per-resource
// interface names start.
const storerPreambleAnchor = "Healthcheck(context.Context) error"

func mergeInterfacesFile(oldSrc, freshSrc string) (string, error) {
	oldPre, oldImports, oldRest, err := splitHeader(oldSrc)
	if err != nil {
		return "", fmt.Errorf("old: %w", err)
	}
	_, freshImports, freshRest, err := splitHeader(freshSrc)
	if err != nil {
		return "", fmt.Errorf("fresh: %w", err)
	}

	oldTypeBefore, oldTypeInner, oldTypeAfter, err := extractParenBlock(oldRest, "type (")
	if err != nil {
		return "", fmt.Errorf("old type block: %w", err)
	}
	_, freshTypeInner, freshTypeAfter, err := extractParenBlock(freshRest, "type (")
	if err != nil {
		return "", fmt.Errorf("fresh type block: %w", err)
	}

	oldStorer, oldIfaces, err := splitInterfaceTypeBlock(oldTypeInner)
	if err != nil {
		return "", fmt.Errorf("old interfaces: %w", err)
	}
	freshStorer, freshIfaces, err := splitInterfaceTypeBlock(freshTypeInner)
	if err != nil {
		return "", fmt.Errorf("fresh interfaces: %w", err)
	}

	mergedIfaces := mergeKeyed(oldIfaces, freshIfaces)

	mergedPlurals := map[string]string{} // ident -> plural display text
	for _, p := range oldStorer.plurals {
		mergedPlurals[identFromPlural(p)] = p
	}
	for _, p := range freshStorer.plurals {
		mergedPlurals[identFromPlural(p)] = p
	}
	idents := make([]string, 0, len(mergedPlurals))
	for id := range mergedPlurals {
		idents = append(idents, id)
	}
	sort.Strings(idents)
	pluralList := make([]string, len(idents))
	for i, id := range idents {
		pluralList[i] = mergedPlurals[id]
	}

	var storerBlock strings.Builder
	storerBlock.WriteString(oldStorer.preamble)
	for _, p := range pluralList {
		storerBlock.WriteString("\t\t")
		storerBlock.WriteString(p)
		storerBlock.WriteString("\n")
	}
	storerBlock.WriteString("\t}")

	var typeInner strings.Builder
	typeInner.WriteString(storerBlock.String())
	for _, b := range mergedIfaces {
		typeInner.WriteString("\n\n")
		typeInner.WriteString(b.text)
	}

	// wrapper func section (everything after the type block)
	oldFuncs, err := parseWrapperFuncs(afterTypeBlockBody(oldTypeAfter))
	if err != nil {
		return "", fmt.Errorf("old wrapper funcs: %w", err)
	}
	freshFuncs, err := parseWrapperFuncs(afterTypeBlockBody(freshTypeAfter))
	if err != nil {
		return "", fmt.Errorf("fresh wrapper funcs: %w", err)
	}

	mergedFuncs := mergeKeyed(groupAdjacent(oldFuncs), groupAdjacent(freshFuncs))

	var out strings.Builder
	out.WriteString(oldPre)
	out.WriteString(renderImportBlock(mergeImports(oldImports, freshImports)))
	out.WriteString(oldTypeBefore)
	out.WriteString(typeInner.String())
	out.WriteString("\n)\n\n")
	out.WriteString(joinBlocks(mergedFuncs))
	out.WriteString("\n")

	return out.String(), nil
}

type storerBlockParts struct {
	preamble string
	plurals  []string
}

func splitInterfaceTypeBlock(inner string) (storerBlockParts, []keyedBlock, error) {
	blocks := splitTopLevelBlocks(inner)
	if len(blocks) == 0 {
		return storerBlockParts{}, nil, fmt.Errorf("empty type block")
	}

	storerRaw := blocks[0]
	lines := strings.Split(storerRaw, "\n")
	anchor := -1
	for i, l := range lines {
		if strings.TrimSpace(l) == storerPreambleAnchor {
			anchor = i
			break
		}
	}
	if anchor == -1 {
		return storerBlockParts{}, nil, fmt.Errorf("could not find Storer preamble anchor")
	}

	preamble := strings.Join(lines[:anchor+1], "\n") + "\n"
	var plurals []string
	for _, l := range lines[anchor+1 : len(lines)-1] {
		t := strings.TrimSpace(l)
		if t != "" {
			plurals = append(plurals, t)
		}
	}

	var ifaces []keyedBlock
	for _, b := range blocks[1:] {
		plural, ok := firstMatch(reInterfaceDecl, b)
		if !ok {
			return storerBlockParts{}, nil, fmt.Errorf("could not extract interface name from block: %.60s", b)
		}
		ifaces = append(ifaces, keyedBlock{key: identFromPlural(plural), text: b})
	}

	return storerBlockParts{preamble: preamble, plurals: plurals}, ifaces, nil
}

// afterTypeBlockBody strips the ")" line (and one blank line) that
// extractParenBlock's "after" return still carries at the front.
func afterTypeBlockBody(after string) string {
	lines := strings.Split(after, "\n")
	return strings.TrimLeft(strings.Join(lines[1:], "\n"), "\n")
}

func parseWrapperFuncs(src string) ([]keyedBlock, error) {
	blocks, err := parseKeyedTopLevel(src, reWrapperParam)
	if err != nil {
		return nil, err
	}
	// keys here are plural display names (from the "s <Plural>" wrapper
	// param) - convert to the true ident sort key, same as interface blocks.
	for i := range blocks {
		blocks[i].key = identFromPlural(blocks[i].key)
	}
	return blocks, nil
}

// parseKeyedTopLevel splits src into top-level blocks and tags each with a
// key extracted via the first matching regexp in res (tried in order,
// first match wins). Each regexp must have exactly one capture group.
func parseKeyedTopLevel(src string, res ...*regexp.Regexp) ([]keyedBlock, error) {
	units := splitTopLevelBlocks(src)
	out := make([]keyedBlock, 0, len(units))
	for _, u := range units {
		var (
			key string
			ok  bool
		)
		for _, re := range res {
			if key, ok = firstMatch(re, u); ok {
				break
			}
		}
		if !ok {
			return nil, fmt.Errorf("could not extract key from block: %.80s", u)
		}
		out = append(out, keyedBlock{key: key, text: u})
	}
	return out, nil
}
