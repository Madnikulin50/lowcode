package def

// ParentRef mirrors one entry of codegen/schema/resource.cue's
// #Resource.parents.
type ParentRef struct {
	Handle   string
	RefField string // optional override, default expIdent+"ID"
	Param    string // optional override, default ident+"ID"
}

type ResolvedParent struct {
	Handle   string `json:"handle"`
	Ident    string `json:"ident"`
	ExpIdent string `json:"expIdent"`
	RefField string `json:"refField"`
	Param    string `json:"param"`
}

func (p ParentRef) Resolve() ResolvedParent {
	b := Base{Handle: p.Handle}.Resolve()

	refField := p.RefField
	if refField == "" {
		refField = b.ExpIdent + "ID"
	}

	param := p.Param
	if param == "" {
		param = b.Ident + "ID"
	}

	return ResolvedParent{
		Handle:   b.Handle,
		Ident:    b.Ident,
		ExpIdent: b.ExpIdent,
		RefField: refField,
		Param:    param,
	}
}

// Resource mirrors codegen/schema/resource.cue's #Resource (the subset
// needed by the types/resources, types/getters_setters and
// rbac/types_rbac templates). Handle is not part of the struct: like in
// CUE, it comes from the resource's key in the owning Component.Resources
// map.
type Resource struct {
	// Ident/ExpIdent optionally override the resource's own identifiers
	// (mirrors #_base's ident/expIdent override, embedded into #Resource) -
	// default is derived from handle, same as everywhere else #_base is used.
	Ident    string
	ExpIdent string

	Model    Model
	Parents  []ParentRef
	Features Features
	Filter   Filter

	// Rbac is nil when the resource defines no rbac block at all (mirrors
	// CUE's optional `rbac?:` field being absent), which excludes the
	// resource from rbac/types_rbac generation.
	Rbac *Rbac

	// Store is nil when the resource defines no store block at all (mirrors
	// CUE's optional `store?:` field being absent), which excludes the
	// resource from store generation.
	Store *StoreConfig
}

type ResolvedResource struct {
	Handle   string `json:"handle"`
	Ident    string `json:"ident"`
	ExpIdent string `json:"expIdent"`
	Fqrt     string `json:"fqrt"`

	Model    ResolvedModel    `json:"model"`
	Parents  []ResolvedParent `json:"parents"`
	Rbac     *Rbac            `json:"rbac,omitempty"`
	Features ResolvedFeatures `json:"-"`
	Filter   ResolvedFilter   `json:"-"`
	Store    *ResolvedStore   `json:"-"`
}

// Resolve mirrors #Resource's defaulting. handle is the resource's key in
// the component's resources map (e.g. "node-sync"); component and platform
// are the owning component's handle and platform ident.
func (r Resource) Resolve(handle, component, platform string) ResolvedResource {
	base := Base{Handle: handle, Ident: r.Ident, ExpIdent: r.ExpIdent}.Resolve()

	parents := make([]ResolvedParent, len(r.Parents))
	for i, p := range r.Parents {
		parents[i] = p.Resolve()
	}

	model := r.Model.Resolve(handle)

	return ResolvedResource{
		Handle:   base.Handle,
		Ident:    base.Ident,
		ExpIdent: base.ExpIdent,
		Fqrt:     platform + "::" + component + ":" + handle,
		Model:    model,
		Parents:  parents,
		Rbac:     r.Rbac,
		Features: r.Features.resolve(),
		Filter:   r.Filter.resolve(base.ExpIdent, model.Attributes),
		Store:    r.Store.resolve(base.Ident, model.Attributes),
	}
}
