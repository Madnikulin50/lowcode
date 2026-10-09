package def

// NamedResource pairs a resource with its handle (its key in the CUE
// `resources: {"node": node, ...}` struct). A slice (not a map) is used
// because the generated output lists resources in declaration order, which
// a Go map could not reproduce deterministically.
type NamedResource struct {
	Handle   string
	Resource Resource
}

// Component mirrors codegen/schema/component.cue's #component (the subset
// needed by the types/resources, types/getters_setters and
// rbac/types_rbac templates).
type Component struct {
	Handle    string
	Platform  string // optional override, default "corteza"
	Resources []NamedResource
}

type ResolvedComponent struct {
	Handle    string             `json:"handle"`
	Ident     string             `json:"ident"`
	ExpIdent  string             `json:"expIdent"`
	Fqrt      string             `json:"fqrt"`
	Resources []ResolvedResource `json:"resources"`
}

func (c Component) Resolve() ResolvedComponent {
	platform := c.Platform
	if platform == "" {
		platform = "corteza"
	}

	base := Base{Handle: c.Handle}.Resolve()

	resources := make([]ResolvedResource, len(c.Resources))
	for i, nr := range c.Resources {
		resources[i] = nr.Resource.Resolve(nr.Handle, c.Handle, platform)
	}

	return ResolvedComponent{
		Handle:    base.Handle,
		Ident:     base.Ident,
		ExpIdent:  base.ExpIdent,
		Fqrt:      platform + "::" + c.Handle,
		Resources: resources,
	}
}
