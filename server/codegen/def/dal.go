package def

// AttributeDal mirrors codegen/schema/model.cue's #ModelAttributeDal.
// Only the type variants actually used by a ported component are
// implemented so far (ID, Ref, Text, Timestamp, JSON, Number, Boolean);
// extend the switch in resolve() when a new component needs another dal type.
type AttributeDal struct {
	Type string // "ID" | "Ref" | "Text" | "Timestamp" | "JSON" | "Number" | "Boolean", default "Text"

	Nullable bool

	// Ref
	RefModelResType string
	RefAttribute    string // default "id"

	// Ref/ID/JSON/Number/Boolean: presence of a default value (mirrors
	// CUE's `default?`)
	HasDefault   bool
	DefaultValue interface{}

	// Text
	Length int

	// Timestamp
	Timezone                bool
	DefaultCurrentTimestamp bool

	// JSON
	DefaultEmptyObject bool

	// Number
	Meta map[string]interface{}
}

// resolve mirrors #ModelAttributeDal's defaulting. The result is a plain
// map so that only the fields relevant to the concrete dal type are present
// in the JSON payload, exactly like CUE's conditional (`if type == ...`)
// struct fields.
func (d AttributeDal) resolve() map[string]interface{} {
	typ := d.Type
	if typ == "" {
		typ = "Text"
	}

	out := map[string]interface{}{
		"type":   typ,
		"fqType": "dal.Type" + typ,
	}
	if d.Nullable {
		out["nullable"] = true
	}

	switch typ {
	case "ID":
		if d.HasDefault {
			out["hasDefault"] = true
			out["default"] = d.DefaultValue
		}
	case "Ref":
		refAttribute := d.RefAttribute
		if refAttribute == "" {
			refAttribute = "id"
		}
		out["attribute"] = refAttribute
		out["refModelResType"] = d.RefModelResType
		if d.HasDefault {
			out["hasDefault"] = true
			out["default"] = d.DefaultValue
		}
	case "Timestamp", "Time":
		out["timezone"] = d.Timezone
		out["precision"] = -1
		if d.DefaultCurrentTimestamp {
			out["defaultCurrentTimestamp"] = true
		}
	case "Text":
		if d.Length != 0 {
			out["length"] = d.Length
		}
	case "JSON":
		if d.DefaultEmptyObject {
			out["defaultEmptyObject"] = true
		}
		if d.HasDefault {
			out["hasDefault"] = true
			out["default"] = d.DefaultValue
		}
	case "Number":
		out["precision"] = -1
		out["scale"] = -1
		if d.HasDefault {
			out["hasDefault"] = true
			out["default"] = d.DefaultValue
		}
		if d.Meta != nil {
			out["meta"] = d.Meta
		}
	case "Boolean":
		if d.HasDefault {
			out["hasDefault"] = true
			out["default"] = d.DefaultValue
		}
	}

	return out
}

// IndexField mirrors codegen/schema/model.cue's #ModelIndexField.
type IndexField struct {
	Attribute string
	Modifiers []string // e.g. "LOWERCASE"
	Sort      string   // "ASC" | "DESC"
	Nulls     string   // "FIRST" | "LAST"
}

// Index mirrors codegen/schema/model.cue's #ModelIndex (the subset needed
// by the dal/$component_model template: a single- or multi-attribute BTREE
// index, "primary" being recognized by its (lower-cased) name). Attribute/
// Attributes are shorthands for a plain Fields list (no modifiers); Fields
// is used directly when a field needs modifiers (e.g. LOWERCASE).
type Index struct {
	Attribute  string   // single-attribute shorthand
	Attributes []string // multi-attribute shorthand
	Fields     []IndexField
	Predicate  string
}

func (idx Index) fields() []IndexField {
	if len(idx.Fields) > 0 {
		return idx.Fields
	}
	if idx.Attribute != "" {
		return []IndexField{{Attribute: idx.Attribute}}
	}
	out := make([]IndexField, len(idx.Attributes))
	for i, a := range idx.Attributes {
		out[i] = IndexField{Attribute: a}
	}
	return out
}
