package def

import "strings"

// This file replaces codegen/server.store.cue: it builds the same JSON
// payload shape (the "_StoreResource" per-resource result, plus the shared
// imports/types payload) for the bulk store/store-rdbms/store-tests task.
//
// Only a single component's contribution is built (BuildStoreTasks takes
// one ResolvedComponent), since not every component has been ported yet;
// see the caller for how that's combined/verified.

// storeBase mirrors codegen/server.store.cue's _StoreResource.result.api's
// "_base" struct, embedded (JSON-flattened) into each api sub-block.
type storeBase struct {
	Ident         string `json:"ident"`
	ExpStoreIdent string `json:"expStoreIdent"`
	GoType        string `json:"goType"`
	GoFilterType  string `json:"goFilterType"`
	AuxIdent      string `json:"auxIdent"`
}

type storeStructField struct {
	Ident      string `json:"ident"`
	ExpIdent   string `json:"expIdent"`
	StoreIdent string `json:"storeIdent"`
	Name       string `json:"name"`
	PrimaryKey bool   `json:"primaryKey"`
	IgnoreCase bool   `json:"ignoreCase"`
	GoType     string `json:"goType"`
}

type storeFilterPayload struct {
	Query        []ResolvedAttribute `json:"query"`
	ByNilState   []ResolvedAttribute `json:"byNilState"`
	ByFalseState []ResolvedAttribute `json:"byFalseState"`
	ByValue      []ResolvedAttribute `json:"byValue"`
	ByLabel      bool                `json:"byLabel"`
	ByFlag       bool                `json:"byFlag"`
}

type storeDeleteByPK struct {
	ExpFnIdent string              `json:"expFnIdent"`
	Attributes []ResolvedAttribute `json:"attributes"`
}

type lookupArgPayload struct {
	Ident      string `json:"ident"`
	StoreIdent string `json:"storeIdent"`
	GoType     string `json:"goType"`
	IgnoreCase bool   `json:"ignoreCase"`
}

type storeLookupPayload struct {
	storeBase
	ExpFnIdent        string             `json:"expFnIdent"`
	Description       string             `json:"description,omitempty"`
	Args              []lookupArgPayload `json:"args"`
	NullConstraint    []string           `json:"nullConstraint"`
	ReturnType        string             `json:"returnType"`
	CollectionFnIdent string             `json:"collectionFnIdent"`
}

type storeSortableFieldsPayload struct {
	storeBase
	FnIdent string            `json:"fnIdent"`
	Fields  map[string]string `json:"fields"`
}

type cursorFieldPayload struct {
	Ident      string `json:"ident"`
	ExpIdent   string `json:"expIdent"`
	PrimaryKey bool   `json:"primaryKey"`
	Unique     bool   `json:"unique"`
	Descending bool   `json:"descending"`
}

type storeCollectCursorPayload struct {
	storeBase
	FnIdent     string               `json:"fnIdent"`
	Fields      []cursorFieldPayload `json:"fields"`
	PrimaryKeys []cursorFieldPayload `json:"primaryKeys"`
}

type nullConstraintEntry struct {
	ExpIdent string `json:"expIdent"`
}

type storeCheckEntry struct {
	LookupFnIdent  string                `json:"lookupFnIdent"`
	Fields         []ResolvedAttribute   `json:"fields"`
	NullConstraint []nullConstraintEntry `json:"nullConstraint"`
}

type storeCheckConstraintsPayload struct {
	storeBase
	FnIdent string            `json:"fnIdent"`
	Checks  []storeCheckEntry `json:"checks"`
}

type functionArgPayload struct {
	Ident  string `json:"ident"`
	GoType string `json:"goType"`
	Spread bool   `json:"spread"`
}

type functionPayload struct {
	storeBase
	ExpFnIdent  string               `json:"expFnIdent"`
	Description string               `json:"description,omitempty"`
	Args        []functionArgPayload `json:"args"`
	Return      []string             `json:"return"`
}

type storeApiPayload struct {
	storeBase
	DeleteByPK          *storeDeleteByPK             `json:"deleteByPK,omitempty"`
	Lookups             []storeLookupPayload         `json:"lookups"`
	Functions           []functionPayload            `json:"functions"`
	SortableFields      storeSortableFieldsPayload   `json:"sortableFields"`
	CollectCursorValues storeCollectCursorPayload    `json:"collectCursorValues"`
	CheckConstraints    storeCheckConstraintsPayload `json:"checkConstraints"`
}

type storeResultPayload struct {
	Ident          string `json:"ident"`
	IdentPlural    string `json:"identPlural"`
	ExpIdent       string `json:"expIdent"`
	ExpIdentPlural string `json:"expIdentPlural"`
	ModelIdent     string `json:"modelIdent"`
	GoType         string `json:"goType"`
	GoSetType      string `json:"goSetType"`
	GoFilterType   string `json:"goFilterType"`

	Struct    []storeStructField `json:"struct"`
	Filter    storeFilterPayload `json:"filter"`
	AuxIdent  string             `json:"auxIdent"`
	AuxStruct []storeStructField `json:"auxStruct"`

	Features ResolvedFeatures `json:"features"`

	Api *storeApiPayload `json:"api,omitempty"`
}

func containsStr(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

// buildSortableFields mirrors _StoreResource.result.api.sortableFields.fields.
func buildSortableFields(attrs []ResolvedAttribute, pkSet map[string]bool) map[string]string {
	fields := map[string]string{}
	for _, a := range attrs {
		if a.Sortable || a.Unique || pkSet[a.Name] {
			fields[strings.ToLower(a.Name)] = a.Name
			fields[strings.ToLower(a.Ident)] = a.Name
		}
	}
	return fields
}

// buildCursorFields mirrors _StoreResource.result.api.collectCursorValues.
func buildCursorFields(attrs []ResolvedAttribute, pkSet map[string]bool) (fields, primaryKeys []cursorFieldPayload) {
	for _, a := range attrs {
		isPK := pkSet[a.Name]
		if a.Sortable || a.Unique || isPK {
			fields = append(fields, cursorFieldPayload{
				Ident: a.Ident, ExpIdent: a.ExpIdent, PrimaryKey: isPK, Unique: a.Unique, Descending: a.Descending,
			})
		}
		if isPK {
			primaryKeys = append(primaryKeys, cursorFieldPayload{Ident: a.Ident, ExpIdent: a.ExpIdent, Descending: a.Descending})
		}
	}
	return
}

// buildStoreResult mirrors codegen/server.store.cue's _StoreResource,
// evaluated for one resource that has a store block.
func buildStoreResult(res ResolvedResource, typePkg string) storeResultPayload {
	var pkNames []string
	if idx, ok := res.Model.Indexes["primary"]; ok {
		for _, f := range idx.fields() {
			pkNames = append(pkNames, f.Attribute)
		}
	}
	pkSet := make(map[string]bool, len(pkNames))
	for _, n := range pkNames {
		pkSet[n] = true
	}

	replaceTypesPkg := func(goType string) string {
		return strings.Replace(goType, "types.", typePkg+".", 1)
	}

	structFields := make([]storeStructField, 0, len(res.Model.AttributesOrdered))
	for _, a := range res.Model.AttributesOrdered {
		if !a.Store {
			continue
		}
		structFields = append(structFields, storeStructField{
			Ident:      a.Ident,
			ExpIdent:   a.ExpIdent,
			StoreIdent: a.StoreIdent,
			Name:       a.Name,
			PrimaryKey: pkSet[a.Name],
			IgnoreCase: a.IgnoreCase,
			GoType:     replaceTypesPkg(a.GoType),
		})
	}

	goType := typePkg + "." + res.ExpIdent
	goFilterType := typePkg + "." + res.Filter.ExpIdent
	// auxIdent is derived from the store-level (prefixed) expIdent, unlike
	// goType/goFilterType which use the resource's own expIdent - mirrors
	// codegen/server.store.cue's result.auxIdent: "aux\(expIdent)" resolving
	// against result.expIdent (= res.store.expIdent), not the outer res.
	auxIdent := "aux" + res.Store.ExpIdent

	base := storeBase{
		Ident:         res.Store.Ident,
		ExpStoreIdent: res.Store.ExpIdentPlural,
		GoType:        goType,
		GoFilterType:  goFilterType,
		AuxIdent:      auxIdent,
	}

	api := &storeApiPayload{storeBase: base}

	if len(pkNames) > 0 {
		attrs := make([]ResolvedAttribute, len(pkNames))
		expIdents := ""
		for i, n := range pkNames {
			attrs[i] = res.Model.Attributes[n]
			expIdents += attrs[i].ExpIdent
		}
		api.DeleteByPK = &storeDeleteByPK{
			ExpFnIdent: "Delete" + res.Store.ExpIdent + "By" + expIdents,
			Attributes: attrs,
		}
	}

	for _, l := range res.Store.Lookups {
		args := make([]lookupArgPayload, len(l.Args))
		for i, a := range l.Args {
			args[i] = lookupArgPayload{Ident: a.Ident, StoreIdent: a.StoreIdent, GoType: a.GoType, IgnoreCase: a.IgnoreCase}
		}
		api.Lookups = append(api.Lookups, storeLookupPayload{
			storeBase:         base,
			ExpFnIdent:        l.ExpFnIdent,
			Description:       l.Description,
			Args:              args,
			NullConstraint:    l.NullConstraint,
			ReturnType:        goType,
			CollectionFnIdent: res.Store.Ident + "Collection",
		})
	}

	for _, f := range res.Store.Functions {
		args := make([]functionArgPayload, len(f.Args))
		for i, a := range f.Args {
			args[i] = functionArgPayload{Ident: a.Ident, GoType: replaceTypesPkg(a.GoType), Spread: a.Spread}
		}

		ret := make([]string, len(f.Return))
		for i, r := range f.Return {
			ret[i] = replaceTypesPkg(r)
		}

		api.Functions = append(api.Functions, functionPayload{
			storeBase:   base,
			ExpFnIdent:  f.ExpFnIdent,
			Description: f.Description,
			Args:        args,
			Return:      ret,
		})
	}

	fields, primaryKeys := buildCursorFields(res.Model.AttributesOrdered, pkSet)

	// fnIdent fields below are derived from the store-level (prefixed)
	// expIdent, same as auxIdent above (see comment on auxIdent).
	api.SortableFields = storeSortableFieldsPayload{
		storeBase: base,
		FnIdent:   "sortable" + res.Store.ExpIdent + "Fields",
		Fields:    buildSortableFields(res.Model.AttributesOrdered, pkSet),
	}
	api.CollectCursorValues = storeCollectCursorPayload{
		storeBase:   base,
		FnIdent:     "collect" + res.Store.ExpIdent + "CursorValues",
		Fields:      fields,
		PrimaryKeys: primaryKeys,
	}
	var checks []storeCheckEntry
	for _, l := range res.Store.Lookups {
		if !l.ConstraintCheck {
			continue
		}

		var nullConstraint []nullConstraintEntry
		for _, a := range res.Model.AttributesOrdered {
			if containsStr(l.NullConstraint, a.Name) {
				nullConstraint = append(nullConstraint, nullConstraintEntry{ExpIdent: a.ExpIdent})
			}
		}

		checks = append(checks, storeCheckEntry{
			LookupFnIdent:  l.ExpFnIdent,
			Fields:         l.Args,
			NullConstraint: nullConstraint,
		})
	}

	api.CheckConstraints = storeCheckConstraintsPayload{
		storeBase: base,
		FnIdent:   "check" + res.Store.ExpIdent + "Constraints",
		Checks:    checks,
	}

	return storeResultPayload{
		Ident:          res.Store.Ident,
		IdentPlural:    res.Store.IdentPlural,
		ExpIdent:       res.Store.ExpIdent,
		ExpIdentPlural: res.Store.ExpIdentPlural,
		ModelIdent:     res.Model.Ident,
		GoType:         goType,
		GoSetType:      goType + "Set",
		GoFilterType:   goFilterType,
		Struct:         structFields,
		Filter: storeFilterPayload{
			Query:        res.Filter.Query,
			ByNilState:   res.Filter.ByNilState,
			ByFalseState: res.Filter.ByFalseState,
			ByValue:      res.Filter.ByValue,
			ByLabel:      res.Features.Labels,
			ByFlag:       res.Features.Flags,
		},
		AuxIdent:  auxIdent,
		AuxStruct: structFields,
		Features:  res.Features,
		Api:       api,
	}
}

// BuildStoreTasks mirrors codegen/server.store.cue's single bulk task,
// across every given (ported) component - not every component has been
// ported yet; see the file comment.
func BuildStoreTasks(components ...ResolvedComponent) []Task {
	imports := make(map[string]string)
	types := make(map[string]storeResultPayload)

	for _, cmp := range components {
		typePkg := cmp.Ident + "Type"
		imports["github.com/madnikulin50/lowcode/server/"+cmp.Ident+"/types"] = typePkg

		for _, res := range cmp.Resources {
			if res.Store == nil {
				continue
			}
			types[res.Store.Ident] = buildStoreResult(res, typePkg)
		}
	}

	payload := map[string]interface{}{
		"imports": imports,
		"types":   types,
	}

	return []Task{
		{
			// store/adapters/rdbms/rdbms.gen.go and store/tests/all_test.go
			// are intentionally excluded: they're marked "Formerly
			// generated from CUE; now maintained by hand." in the repo -
			// codegen no longer touches them at all.
			Bulk: []IOSpec{
				{Template: "gocode/store/interfaces.go.tpl", Output: "store/interfaces.gen.go", Syntax: "go"},
				{Template: "gocode/store/rdbms/aux_types.go.tpl", Output: "store/adapters/rdbms/aux_types.gen.go", Syntax: "go"},
				{Template: "gocode/store/rdbms/queries.go.tpl", Output: "store/adapters/rdbms/queries.gen.go", Syntax: "go"},
				{Template: "gocode/store/rdbms/filters.go.tpl", Output: "store/adapters/rdbms/filters.gen.go", Syntax: "go"},
			},
			Payload: payload,
		},
	}
}
