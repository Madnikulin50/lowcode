package def

import "strings"

// Attribute mirrors codegen/schema/model.cue's #ModelAttribute. Only the
// fields needed to drive the types/resources, types/getters_setters and
// rbac/types_rbac templates are modeled so far; store/dal/envoy fields will
// be added when those templates are ported.
type Attribute struct {
	Ident      string   // optional override
	ExpIdent   string   // optional override
	GoType     string   // default "string"
	GoCastFnc  string   // optional override
	IdentAlias []string // optional override, default [ident, expIdent]
	OmitGetter bool
	OmitSetter bool

	StoreIdent string // optional override, default name
	Sortable   bool
	Unique     bool
	Descending bool
	IgnoreCase bool

	// NoStore mirrors `store: false` (the CUE default is true): excludes
	// this attribute from the store templates' struct/auxStruct, without
	// affecting dal/$component_model generation (that only checks `dal?`).
	NoStore bool

	// Dal is nil when the attribute has no `dal:` block at all (mirrors
	// CUE's optional `dal?:` being absent), which excludes it from
	// dal/$component_model generation.
	Dal *AttributeDal
}

type ResolvedAttribute struct {
	Name       string                 `json:"name"`
	Ident      string                 `json:"ident"`
	ExpIdent   string                 `json:"expIdent"`
	GoType     string                 `json:"goType"`
	GoCastFnc  string                 `json:"goCastFnc"`
	IdentAlias []string               `json:"identAlias"`
	OmitGetter bool                   `json:"omitGetter"`
	OmitSetter bool                   `json:"omitSetter"`
	StoreIdent string                 `json:"storeIdent"`
	Sortable   bool                   `json:"sortable"`
	Unique     bool                   `json:"unique"`
	Descending bool                   `json:"descending"`
	IgnoreCase bool                   `json:"ignoreCase"`
	Store      bool                   `json:"store"`
	Dal        map[string]interface{} `json:"dal,omitempty"`
}

// Resolve mirrors #ModelAttribute's defaulting, given the attribute's map
// key (e.g. "shared_node_id").
func (a Attribute) Resolve(name string) ResolvedAttribute {
	ident := a.Ident
	if ident == "" {
		ident = camel(pascal(splitWords(name, "_", ".")))
	}

	expIdent := a.ExpIdent
	if expIdent == "" {
		expIdent = titleFirst(ident)
	}

	goType := a.GoType
	if goType == "" {
		goType = "string"
	}

	goCastFnc := a.GoCastFnc
	if goCastFnc == "" {
		switch goType {
		case "*time.Time":
			goCastFnc = "TimePtr"
		case "time.Time":
			goCastFnc = "Time"
		case "map[string]any":
			goCastFnc = "Meta"
		default:
			goCastFnc = titleFirst(goType)
		}
	}

	identAlias := a.IdentAlias
	if identAlias == nil {
		identAlias = []string{ident, expIdent}
	}

	storeIdent := a.StoreIdent
	if storeIdent == "" {
		storeIdent = name
	}

	var dal map[string]interface{}
	if a.Dal != nil {
		dal = a.Dal.resolve()
	}

	return ResolvedAttribute{
		Name:       name,
		Ident:      ident,
		ExpIdent:   expIdent,
		GoType:     goType,
		GoCastFnc:  goCastFnc,
		IdentAlias: identAlias,
		OmitGetter: a.OmitGetter,
		OmitSetter: a.OmitSetter,
		StoreIdent: storeIdent,
		Sortable:   a.Sortable,
		Unique:     a.Unique,
		Descending: a.Descending,
		IgnoreCase: a.IgnoreCase,
		Store:      !a.NoStore,
		Dal:        dal,
	}
}

// NamedAttribute pairs an attribute with its map key (its name in the CUE
// `attributes: {id: ..., name: ...}` struct). A slice (not a map) is used
// because CUE's `[for attr in res.model.attributes {...}]` comprehension
// (used by the dal/$component_model template payload) preserves the
// source's declaration order, which a Go map could not reproduce
// deterministically. The types/getters_setters template, by contrast,
// consumes model.attributes as a JSON object and therefore sees them
// alphabetized by Go's text/template map-range - see ResolvedModel.Attributes.
type NamedAttribute struct {
	Name      string
	Attribute Attribute
}

// Model mirrors codegen/schema/model.cue's #Model (the subset needed so far).
type Model struct {
	Ident            string
	Attributes       []NamedAttribute
	Indexes          map[string]Index
	OmitGetterSetter bool
	DefaultGetter    bool
	DefaultSetter    bool
}

type ResolvedModel struct {
	Ident      string                       `json:"ident"`
	Attributes map[string]ResolvedAttribute `json:"attributes"`

	// AttributesOrdered mirrors the same attributes, but in source
	// declaration order - needed by the dal/$component_model payload.
	AttributesOrdered []ResolvedAttribute `json:"-"`
	Indexes           map[string]Index    `json:"-"`

	OmitGetterSetter bool `json:"omitGetterSetter"`
	DefaultGetter    bool `json:"defaultGetter"`
	DefaultSetter    bool `json:"defaultSetter"`
}

func (m Model) Resolve(resourceHandle string) ResolvedModel {
	attrs := make(map[string]ResolvedAttribute, len(m.Attributes))
	ordered := make([]ResolvedAttribute, len(m.Attributes))
	for i, na := range m.Attributes {
		resolved := na.Attribute.Resolve(na.Name)
		attrs[na.Name] = resolved
		ordered[i] = resolved
	}

	ident := m.Ident
	if ident == "" {
		// mirrors codegen/schema/resource.cue: model.ident default is the
		// resource handle (plural), dashes turned into underscores verbatim
		// (no case transform).
		ident = strings.ReplaceAll(resourceHandle, "-", "_") + "s"
	}

	return ResolvedModel{
		Ident:             ident,
		Attributes:        attrs,
		AttributesOrdered: ordered,
		Indexes:           m.Indexes,
		OmitGetterSetter:  m.OmitGetterSetter,
		DefaultGetter:     m.DefaultGetter,
		DefaultSetter:     m.DefaultSetter,
	}
}
