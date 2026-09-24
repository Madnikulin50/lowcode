package main

import (
	"fmt"
	"strings"
)

// --- aux_types.gen.go -------------------------------------------------------

func mergeAuxTypesFile(oldSrc, freshSrc string) (string, error) {
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

	oldStructs, err := parseKeyedTopLevel(oldTypeInner, reStructDecl)
	if err != nil {
		return "", fmt.Errorf("old structs: %w", err)
	}
	freshStructs, err := parseKeyedTopLevel(freshTypeInner, reStructDecl)
	if err != nil {
		return "", fmt.Errorf("fresh structs: %w", err)
	}
	mergedStructs := mergeKeyed(oldStructs, freshStructs)

	oldFuncs, err := parseKeyedTopLevel(afterTypeBlockBody(oldTypeAfter), reAuxReceiver)
	if err != nil {
		return "", fmt.Errorf("old funcs: %w", err)
	}
	freshFuncs, err := parseKeyedTopLevel(afterTypeBlockBody(freshTypeAfter), reAuxReceiver)
	if err != nil {
		return "", fmt.Errorf("fresh funcs: %w", err)
	}
	mergedFuncs := mergeKeyed(groupAdjacent(oldFuncs), groupAdjacent(freshFuncs))

	var out strings.Builder
	out.WriteString(oldPre)
	out.WriteString(renderImportBlock(mergeImports(oldImports, freshImports)))
	out.WriteString(oldTypeBefore)
	out.WriteString("\n")
	out.WriteString(joinBlocks(mergedStructs))
	out.WriteString("\n)\n\n")
	out.WriteString(joinBlocks(mergedFuncs))
	out.WriteString("\n")

	return out.String(), nil
}

// --- queries.gen.go ----------------------------------------------------------

func mergeQueriesFile(oldSrc, freshSrc string) (string, error) {
	oldPre, oldImports, oldRest, err := splitHeader(oldSrc)
	if err != nil {
		return "", fmt.Errorf("old: %w", err)
	}
	_, freshImports, freshRest, err := splitHeader(freshSrc)
	if err != nil {
		return "", fmt.Errorf("fresh: %w", err)
	}

	oldVarBefore, oldVarInner, _, err := extractParenBlock(oldRest, "var (")
	if err != nil {
		return "", fmt.Errorf("old var block: %w", err)
	}
	_, freshVarInner, _, err := extractParenBlock(freshRest, "var (")
	if err != nil {
		return "", fmt.Errorf("fresh var block: %w", err)
	}

	oldUnits, err := parseKeyedTopLevel(oldVarInner, reQueryVarName)
	if err != nil {
		return "", fmt.Errorf("old query vars: %w", err)
	}
	freshUnits, err := parseKeyedTopLevel(freshVarInner, reQueryVarName)
	if err != nil {
		return "", fmt.Errorf("fresh query vars: %w", err)
	}
	merged := mergeKeyed(groupAdjacent(oldUnits), groupAdjacent(freshUnits))

	var out strings.Builder
	out.WriteString(oldPre)
	out.WriteString(renderImportBlock(mergeImports(oldImports, freshImports)))
	out.WriteString(oldVarBefore)
	out.WriteString("\n")
	out.WriteString(joinBlocks(merged))
	out.WriteString("\n)\n")

	return out.String(), nil
}

// --- filters.gen.go -----------------------------------------------------------

const filterFieldsPreambleAnchor = "// Filter extensions for search/query functions"

func mergeFiltersFile(oldSrc, freshSrc string) (string, error) {
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

	oldStructPreamble, oldFields, err := extractedFiltersFields(oldTypeInner)
	if err != nil {
		return "", fmt.Errorf("old extendedFilters fields: %w", err)
	}
	_, freshFields, err := extractedFiltersFields(freshTypeInner)
	if err != nil {
		return "", fmt.Errorf("fresh extendedFilters fields: %w", err)
	}
	mergedFields := mergeKeyed(oldFields, freshFields)

	oldFuncs, err := parseKeyedTopLevel(afterTypeBlockBody(oldTypeAfter), reFilterFunc, reFuncName)
	if err != nil {
		return "", fmt.Errorf("old filter funcs: %w", err)
	}
	freshFuncs, err := parseKeyedTopLevel(afterTypeBlockBody(freshTypeAfter), reFilterFunc, reFuncName)
	if err != nil {
		return "", fmt.Errorf("fresh filter funcs: %w", err)
	}
	mergedFuncs := mergeKeyed(oldFuncs, freshFuncs)

	// mergedFields' text already carries its original "\t\t"-deep indentation
	// verbatim from the source file (splitTopLevelBlocks doesn't strip it),
	// so it's joined as-is - re-indenting it here would double up the tabs
	// on every merge (harmless once gofmt reformats, but still wrong).
	var extendedFilters strings.Builder
	extendedFilters.WriteString(oldStructPreamble)
	for _, f := range mergedFields {
		extendedFilters.WriteString("\n")
		extendedFilters.WriteString(f.text)
		extendedFilters.WriteString("\n")
	}
	extendedFilters.WriteString("\t}")

	var out strings.Builder
	out.WriteString(oldPre)
	out.WriteString(renderImportBlock(mergeImports(oldImports, freshImports)))
	out.WriteString(oldTypeBefore)
	out.WriteString(extendedFilters.String())
	out.WriteString("\n)\n\n")
	out.WriteString(joinBlocks(mergedFuncs))
	out.WriteString("\n")

	return out.String(), nil
}

// extractedFiltersFields parses the (single) extendedFilters struct body
// and returns:
//   - the struct's fixed preamble verbatim (any leading doc comment, the
//     "extendedFilters struct {" line, and the "// Filter extensions..."
//     comment that precedes the first field)
//   - its per-resource field declarations, keyed by field name
//
// Struct bodies close with "}", not ")", so this doesn't reuse
// extractParenBlock (which is paren-specific) - it locates the
// "extendedFilters struct {" line directly instead.
func extractedFiltersFields(typeInner string) (preamble string, fields []keyedBlock, err error) {
	blocks := splitTopLevelBlocks(typeInner)
	if len(blocks) != 1 {
		return "", nil, fmt.Errorf("expected exactly one top-level struct in filters type block, got %d", len(blocks))
	}
	structBody := blocks[0]

	lines := strings.Split(structBody, "\n")
	declLine := -1
	for i, l := range lines {
		if strings.TrimSpace(l) == "extendedFilters struct {" {
			declLine = i
			break
		}
	}
	if declLine == -1 {
		return "", nil, fmt.Errorf("could not find \"extendedFilters struct {\" line")
	}

	anchorLine := -1
	for i := declLine + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == filterFieldsPreambleAnchor {
			anchorLine = i
			break
		}
	}
	if anchorLine == -1 {
		return "", nil, fmt.Errorf("could not find extendedFilters preamble anchor")
	}

	preamble = strings.Join(lines[:anchorLine+1], "\n") + "\n"
	inner := strings.Join(lines[anchorLine+1:len(lines)-1], "\n")

	fields, err = parseKeyedTopLevel(inner, reFilterField)
	return preamble, fields, err
}
