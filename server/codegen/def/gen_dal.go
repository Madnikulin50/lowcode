package def

import "strings"

// This file replaces codegen/server.dal_models.cue's per-component tasks
// (the "system/model/corteza.gen.go" app.resources task is not ported: it
// has no equivalent in a Go-defined component yet).

type dalModelPayload struct {
	Var        string                   `json:"var"`
	ResType    string                   `json:"resType"`
	Ident      string                   `json:"ident"`
	Attributes []ResolvedAttribute      `json:"attributes"`
	Indexes    map[string]dalIndexEntry `json:"indexes"`
}

type dalIndexEntry struct {
	Ident     string          `json:"ident"`
	Type      string          `json:"type"`
	Unique    bool            `json:"unique,omitempty"`
	Predicate string          `json:"predicate,omitempty"`
	Fields    []dalIndexField `json:"fields"`
}

type dalIndexField struct {
	Attribute string   `json:"attribute"`
	Modifiers []string `json:"modifiers,omitempty"`
	Sort      string   `json:"sort,omitempty"`
	Nulls     string   `json:"nulls,omitempty"`
}

// buildDalIndexes mirrors #ModelIndex's defaulting plus
// codegen/server.dal_models.cue's primary/non-primary branching.
func buildDalIndexes(modelIdent string, indexes map[string]Index, attrs map[string]ResolvedAttribute) map[string]dalIndexEntry {
	if len(indexes) == 0 {
		return nil
	}

	out := make(map[string]dalIndexEntry, len(indexes))
	for name, idx := range indexes {
		primary := strings.ToLower(name) == "primary"
		unique := primary || strings.Contains(name, "unique")

		entry := dalIndexEntry{Type: "BTREE", Predicate: idx.Predicate}
		if primary {
			entry.Ident = "PRIMARY"
		} else {
			entry.Ident = modelIdent + "_" + camel(pascal(splitWords(name, "_", ".")))
			entry.Unique = unique
		}

		for _, f := range idx.fields() {
			entry.Fields = append(entry.Fields, dalIndexField{
				Attribute: attrs[f.Attribute].ExpIdent,
				Modifiers: f.Modifiers,
				Sort:      f.Sort,
				Nulls:     f.Nulls,
			})
		}

		out[name] = entry
	}
	return out
}

// BuildDalTasks mirrors the two per-component tasks in
// codegen/server.dal_models.cue: "<cmp>/model/models.gen.go" and
// "<cmp>/model/init.gen.go".
func BuildDalTasks(cmp ResolvedComponent) []Task {
	models := make(map[string]dalModelPayload, len(cmp.Resources))
	for _, res := range cmp.Resources {
		attrs := make([]ResolvedAttribute, 0, len(res.Model.AttributesOrdered))
		for _, a := range res.Model.AttributesOrdered {
			if a.Dal != nil {
				attrs = append(attrs, a)
			}
		}

		models[res.Ident] = dalModelPayload{
			Var:        res.ExpIdent,
			ResType:    "types." + res.ExpIdent + "ResourceType",
			Ident:      res.Model.Ident,
			Attributes: attrs,
			Indexes:    buildDalIndexes(res.Model.Ident, res.Model.Indexes, res.Model.Attributes),
		}
	}

	return []Task{
		goTask(
			"gocode/dal/$component_model.go.tpl",
			cmp.Ident+"/model/models.gen.go",
			map[string]interface{}{
				"package": "model",
				"imports": []string{`"github.com/madnikulin50/lowcode/server/` + cmp.Ident + `/types"`},
				"models":  models,
			},
		),
		goTask(
			"gocode/dal/$component_init.go.tpl",
			cmp.Ident+"/model/init.gen.go",
			map[string]interface{}{
				"package": "model",
			},
		),
	}
}
