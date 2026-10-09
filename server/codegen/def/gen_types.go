package def

// This file replaces codegen/server.types.cue and codegen/server.rbac_types.cue:
// it builds the same JSON payload shape those CUE files used to produce,
// from a ResolvedComponent, for the same three templates.

type resourceTypeConst struct {
	Const string `json:"const"`
	Type  string `json:"type"`
}

// BuildTypesTasks mirrors codegen/server.types.cue.
func BuildTypesTasks(cmp ResolvedComponent) []Task {
	consts := make([]resourceTypeConst, 0, len(cmp.Resources)+1)
	for _, res := range cmp.Resources {
		consts = append(consts, resourceTypeConst{
			Const: res.ExpIdent + "ResourceType",
			Type:  res.Fqrt,
		})
	}
	consts = append(consts, resourceTypeConst{Const: "ComponentResourceType", Type: cmp.Fqrt})

	return []Task{
		goTask(
			"gocode/types/$component_resources.go.tpl",
			cmp.Ident+"/types/resources.gen.go",
			map[string]interface{}{
				"package":  "types",
				"cmpIdent": cmp.Ident,
				"types":    consts,
			},
		),
		goTask(
			"gocode/types/$component_getters_setters.go.tpl",
			cmp.Ident+"/types/getters_setters.gen.go",
			map[string]interface{}{
				"package":   "types",
				"cmpIdent":  cmp.Ident,
				"resources": cmp.Resources,
			},
		),
	}
}

type rbacTypeEntry struct {
	Const      string           `json:"const"`
	Type       string           `json:"type"`
	ResFunc    string           `json:"resFunc"`
	TplFunc    string           `json:"tplFunc"`
	AttFunc    string           `json:"attFunc"`
	GoType     string           `json:"goType"`
	Component  bool             `json:"component,omitempty"`
	References []ResolvedParent `json:"references"`
}

// BuildRbacTypesTask mirrors codegen/server.rbac_types.cue.
func BuildRbacTypesTask(cmp ResolvedComponent) Task {
	types := make([]rbacTypeEntry, 0, len(cmp.Resources)+1)
	for _, res := range cmp.Resources {
		if res.Rbac == nil {
			continue
		}

		references := append([]ResolvedParent{}, res.Parents...)
		references = append(references, ResolvedParent{Param: "id", RefField: "ID"})

		types = append(types, rbacTypeEntry{
			Const:      res.ExpIdent + "ResourceType",
			Type:       res.Fqrt,
			ResFunc:    res.ExpIdent + "RbacResource",
			TplFunc:    res.ExpIdent + "RbacResourceTpl",
			AttFunc:    res.ExpIdent + "RbacAttributes",
			GoType:     res.ExpIdent,
			References: references,
		})
	}

	types = append(types, rbacTypeEntry{
		Const:     "ComponentResourceType",
		Type:      cmp.Fqrt,
		ResFunc:   "ComponentRbacResource",
		TplFunc:   "ComponentRbacResourceTpl",
		AttFunc:   "ComponentRbacAttributes",
		GoType:    "Component",
		Component: true,
	})

	return goTask(
		"gocode/rbac/$component_types_rbac.go.tpl",
		cmp.Ident+"/types/rbac.gen.go",
		map[string]interface{}{
			"package":  "types",
			"cmpIdent": cmp.Ident,
			"types":    types,
		},
	)
}
