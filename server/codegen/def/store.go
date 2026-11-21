package def

import "strings"

// Features mirrors codegen/schema/resource.cue's #Resource.features.
// Labels/Paging/Sorting/CheckFn default to true, Flags defaults to false;
// *bool lets a resource override a true-by-default flag to false.
type Features struct {
	Labels  *bool
	Flags   *bool
	Paging  *bool
	Sorting *bool
	CheckFn *bool
}

type ResolvedFeatures struct {
	Labels  bool `json:"labels"`
	Flags   bool `json:"flags"`
	Paging  bool `json:"paging"`
	Sorting bool `json:"sorting"`
	CheckFn bool `json:"checkFn"`
}

func boolPtr(b bool) *bool { return &b }

func boolOr(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}

func (f Features) resolve() ResolvedFeatures {
	return ResolvedFeatures{
		Labels:  boolOr(f.Labels, true),
		Flags:   boolOr(f.Flags, false),
		Paging:  boolOr(f.Paging, true),
		Sorting: boolOr(f.Sorting, true),
		CheckFn: boolOr(f.CheckFn, true),
	}
}

// Filter mirrors codegen/schema/resource.cue's #Resource.filter (the subset
// needed by the store templates). Struct is its own small attribute set
// (defaulted the same way as model attributes), separate from
// Model.Attributes; Query resolves against the resource's model attributes
// instead (see codegen/server.store.cue's _StoreResource.result.filter).
//
// A Struct entry's GoType/Ident/StoreIdent, when left unset, default from
// the model attribute of the same name (if one exists) - the common case
// being a filter field that just reuses the model attribute's type/naming
// unchanged. Sortable/Unique/Descending/IgnoreCase/Dal are deliberately
// NOT inherited: those are model/DAL-level concerns that a filter struct
// field has no equivalent of (a filter's own Attribute always resolves
// them to their own zero value regardless). A genuine override - a
// list-type filter field like []uint64 for a singular uint64 model
// attribute, or a filter-only field with no model attribute at all (e.g.
// "deleted") - still needs its differing field(s) stated explicitly.
type Filter struct {
	ExpIdent     string // optional override, default resourceExpIdent+"Filter"
	Struct       map[string]Attribute
	Query        []string
	ByNilState   []string
	ByFalseState []string
	ByValue      []string
}

type ResolvedFilter struct {
	ExpIdent     string
	Query        []ResolvedAttribute
	ByNilState   []ResolvedAttribute
	ByFalseState []ResolvedAttribute
	ByValue      []ResolvedAttribute
}

func pickAttrs(names []string, from map[string]ResolvedAttribute) []ResolvedAttribute {
	out := make([]ResolvedAttribute, 0, len(names))
	for _, n := range names {
		out = append(out, from[n])
	}
	return out
}

func (f Filter) resolve(resourceExpIdent string, modelAttrs map[string]ResolvedAttribute) ResolvedFilter {
	structAttrs := make(map[string]ResolvedAttribute, len(f.Struct))
	for name, a := range f.Struct {
		if ma, ok := modelAttrs[name]; ok {
			if a.GoType == "" {
				a.GoType = ma.GoType
			}
			if a.Ident == "" {
				a.Ident = ma.Ident
			}
			if a.StoreIdent == "" {
				a.StoreIdent = ma.StoreIdent
			}
		}
		structAttrs[name] = a.Resolve(name)
	}

	expIdent := f.ExpIdent
	if expIdent == "" {
		expIdent = resourceExpIdent + "Filter"
	}

	return ResolvedFilter{
		ExpIdent:     expIdent,
		Query:        pickAttrs(f.Query, modelAttrs),
		ByNilState:   pickAttrs(f.ByNilState, structAttrs),
		ByFalseState: pickAttrs(f.ByFalseState, structAttrs),
		ByValue:      pickAttrs(f.ByValue, structAttrs),
	}
}

// StoreLookup mirrors one entry of codegen/schema/resource.cue's
// #Resource.store.api.lookups.
type StoreLookup struct {
	Fields          []string
	Description     string
	NullConstraint  []string
	ConstraintCheck bool
}

type ResolvedStoreLookup struct {
	ExpFnIdent      string
	Description     string
	Args            []ResolvedAttribute
	NullConstraint  []string
	ConstraintCheck bool
}

// StoreFunctionArg mirrors one entry of a store function's args.
type StoreFunctionArg struct {
	Ident  string
	GoType string
	Spread bool
}

// StoreFunction mirrors one entry of codegen/schema/resource.cue's
// #Resource.store.api.functions.
type StoreFunction struct {
	ExpIdent    string
	Description string
	Args        []StoreFunctionArg
	Return      []string
}

type ResolvedStoreFunctionArg struct {
	Ident  string
	GoType string // not yet typePkg-adjusted, see gen_store.go
	Spread bool
}

type ResolvedStoreFunction struct {
	ExpFnIdent  string
	Description string
	Args        []ResolvedStoreFunctionArg
	Return      []string // not yet typePkg-adjusted, see gen_store.go
}

// StoreConfig mirrors codegen/schema/resource.cue's #Resource.store.
type StoreConfig struct {
	Ident     string
	Lookups   []StoreLookup
	Functions []StoreFunction
}

type ResolvedStore struct {
	Ident          string
	IdentPlural    string
	ExpIdent       string
	ExpIdentPlural string
	Lookups        []ResolvedStoreLookup
	Functions      []ResolvedStoreFunction
}

// formatLookupDescription mirrors codegen/server.store.cue's lookup
// description formatting: "// {expFnIdent} " followed by the (already
// CUE-dedented) description text, each line prefixed with "// ". The CUE
// schema defaults description to "" rather than leaving it absent, so this
// always formats (even an empty description yields "// {expFnIdent} ",
// trailing space and all - gofmt trims it on the way out). The go/format
// pass that codegen/tool applies afterwards (unchanged) then turns a lone
// short line into a "# " doc-comment heading, exactly as it did for the
// CUE-generated output.
func formatLookupDescription(expFnIdent, description string) string {
	return "// " + expFnIdent + " " + strings.Join(strings.Split(description, "\n"), "\n// ")
}

func (s *StoreConfig) resolve(resourceIdent string, modelAttrs map[string]ResolvedAttribute) *ResolvedStore {
	if s == nil {
		return nil
	}

	ident := s.Ident
	if ident == "" {
		ident = resourceIdent
	}
	identPlural := ident + "s"
	expIdent := titleFirst(ident)
	expIdentPlural := expIdent + "s"

	lookups := make([]ResolvedStoreLookup, len(s.Lookups))
	for i, l := range s.Lookups {
		expFields := ""
		args := make([]ResolvedAttribute, len(l.Fields))
		for j, name := range l.Fields {
			attr := modelAttrs[name]
			args[j] = attr
			expFields += titleFirst(attr.ExpIdent)
		}

		expFnIdent := "Lookup" + expIdent + "By" + expFields

		lookups[i] = ResolvedStoreLookup{
			ExpFnIdent:      expFnIdent,
			Description:     formatLookupDescription(expFnIdent, l.Description),
			Args:            args,
			NullConstraint:  l.NullConstraint,
			ConstraintCheck: l.ConstraintCheck,
		}
	}

	functions := make([]ResolvedStoreFunction, len(s.Functions))
	for i, f := range s.Functions {
		args := make([]ResolvedStoreFunctionArg, len(f.Args))
		for j, a := range f.Args {
			args[j] = ResolvedStoreFunctionArg{Ident: a.Ident, GoType: a.GoType, Spread: a.Spread}
		}

		functions[i] = ResolvedStoreFunction{
			ExpFnIdent:  f.ExpIdent,
			Description: formatLookupDescription(f.ExpIdent, f.Description),
			Args:        args,
			Return:      f.Return,
		}
	}

	return &ResolvedStore{
		Ident:          ident,
		IdentPlural:    identPlural,
		ExpIdent:       expIdent,
		ExpIdentPlural: expIdentPlural,
		Lookups:        lookups,
		Functions:      functions,
	}
}
