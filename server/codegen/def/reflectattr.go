package def

import (
	"reflect"
	"strconv"
	"strings"
)

// AttributesFromStruct is a pilot for eliminating the dual source of truth
// between codegen/def's Attribute literals and the hand-written type
// struct itself - the exact class of bug that let federation/types/node.go
// (nee compose/types/page.go) drift from its schema (a field declared in
// the schema with no matching struct field, or vice versa).
//
// It builds a resource's ordered attribute list by reflecting over the
// hand-written struct (v, e.g. types.Node{}) and reading `schema:"..."`
// tags on its fields, in field declaration order - see
// federation/types/node.go for the annotated example. Fields with no
// `schema` tag are skipped (e.g. json-only helper fields, computed
// getters).
//
// Tag grammar: `schema:"directive,directive,..."`, each directive either a
// bare flag (`sortable`, `unique`, `descending`, `ignoreCase`, `nostore`,
// `omit`) or `key=value`:
//   - col=<name>      required; the snake_case schema/store column name
//   - store=<name>    optional override for the store column, when it
//     differs from col (e.g. a "rel_module" attribute stored as
//     "rel_compose_module")
//   - ident=<name>    optional override for the camelCase ident (rarely
//     needed - see identFromGoName)
//   - goType=<type>   optional override for the reflected Go type string
//     (rarely needed - reflect already renders e.g. "types.ModuleFieldSet"
//     or "*time.Time" correctly; use this for a locally-aliased type whose
//     reflect name doesn't match, e.g. compose's `rawJson`)
//   - alias=a|b|c     optional override for identAlias (pipe-separated),
//     default [ident, expIdent]
//   - dal=<spec>      optional; absence means no dal block (excluded from
//     dal/$component_model generation, like module.fields). <spec> is one
//     of: id | userref | ref:<resType>[:default0] | timestamp[:now|:nil] |
//     text[:<length>] | json[:empty] | number[:default0] | bool[:true|:false]
//   - or empty ("dal" bare) for a plain Text column. number always
//     carries the {"rdbms:type":"integer"} meta hint, matching every
//     Number attribute in the ported schema so far.
//
// ExpIdent is never tagged: it's always exactly the Go field name, since
// the generated code accesses the field directly (r.<ExpIdent>) and
// wouldn't compile otherwise.
func AttributesFromStruct(v interface{}) []NamedAttribute {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var out []NamedAttribute
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag, ok := f.Tag.Lookup("schema")
		if !ok {
			continue
		}

		d := parseSchemaTag(tag)
		col, ok := d["col"]
		if !ok {
			panic("AttributesFromStruct: field " + t.Name() + "." + f.Name + " is missing schema:\"col=...\"")
		}

		out = append(out, NamedAttribute{Name: col, Attribute: attributeFromTag(f, d)})
	}
	return out
}

func parseSchemaTag(tag string) map[string]string {
	m := map[string]string{}
	for _, part := range strings.Split(tag, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if i := strings.Index(part, "="); i >= 0 {
			m[part[:i]] = part[i+1:]
		} else {
			m[part] = ""
		}
	}
	return m
}

func hasFlag(d map[string]string, key string) bool {
	_, ok := d[key]
	return ok
}

func attributeFromTag(f reflect.StructField, d map[string]string) Attribute {
	goType := f.Type.String()
	if override, ok := d["goType"]; ok {
		goType = override
	}

	a := Attribute{
		ExpIdent:   f.Name, // must equal the Go field name - see doc comment
		GoType:     goType,
		Sortable:   hasFlag(d, "sortable"),
		Unique:     hasFlag(d, "unique"),
		Descending: hasFlag(d, "descending"),
		IgnoreCase: hasFlag(d, "ignoreCase"),
		NoStore:    hasFlag(d, "nostore"),
	}
	if hasFlag(d, "omit") {
		a.OmitGetter = true
		a.OmitSetter = true
	}
	if ident, ok := d["ident"]; ok {
		a.Ident = ident
	} else {
		a.Ident = identFromGoName(f.Name)
	}
	if alias, ok := d["alias"]; ok {
		a.IdentAlias = strings.Split(alias, "|")
	}
	if store, ok := d["store"]; ok {
		a.StoreIdent = store
	}
	if dal, ok := d["dal"]; ok {
		a.Dal = parseDalDirective(dal)
	}
	return a
}

// identFromGoName derives the default camelCase ident from a Go field
// name: lowercase the first rune (e.g. "BaseURL" -> "baseURL"), except
// when the whole name is an acronym (e.g. "ID" -> "id", not "iD") - the
// only case this doesn't handle is an acronym-prefixed name (e.g. a
// hypothetical "IDPrefix"), which none of federation.Node's fields need;
// use the `ident=` tag override for that.
func identFromGoName(name string) string {
	if name == strings.ToUpper(name) {
		return strings.ToLower(name)
	}
	r := []rune(name)
	return strings.ToLower(string(r[0])) + string(r[1:])
}

// parseDalDirective splits on only the FIRST ":" (kind vs. rest) - a
// RefModelResType like "corteza::federation:node" contains colons of its
// own, so a naive strings.Split(spec, ":") truncates it (this broke the
// first version of this pilot: ref:corteza::federation:node:default0 was
// silently parsed as RefModelResType "corteza").
func parseDalDirective(spec string) *AttributeDal {
	if spec == "" {
		return &AttributeDal{}
	}

	kind, rest, hasRest := strings.Cut(spec, ":")
	switch kind {
	case "id":
		return &AttributeDal{Type: "ID"}
	case "userref":
		return userRefDal
	case "ref":
		resType := rest
		hasDefault := false
		if r, ok := strings.CutSuffix(resType, ":default0"); ok {
			resType = r
			hasDefault = true
		}
		d := &AttributeDal{Type: "Ref", RefModelResType: resType}
		if hasDefault {
			d.HasDefault = true
			d.DefaultValue = 0
		}
		return d
	case "timestamp":
		d := &AttributeDal{Type: "Timestamp", Timezone: true}
		if hasRest {
			switch rest {
			case "now":
				d.DefaultCurrentTimestamp = true
			case "nil":
				d.Nullable = true
			}
		}
		return d
	case "text":
		d := &AttributeDal{Type: "Text"}
		if hasRest {
			n, _ := strconv.Atoi(rest)
			d.Length = n
		}
		return d
	case "json":
		d := &AttributeDal{Type: "JSON"}
		if hasRest && rest == "empty" {
			d.DefaultEmptyObject = true
		}
		return d
	case "number":
		d := &AttributeDal{Type: "Number", Meta: map[string]interface{}{"rdbms:type": "integer"}}
		if hasRest && rest == "default0" {
			d.HasDefault = true
			d.DefaultValue = 0
		}
		return d
	case "bool":
		d := &AttributeDal{Type: "Boolean"}
		if hasRest {
			b, _ := strconv.ParseBool(rest)
			d.HasDefault = true
			d.DefaultValue = b
		}
		return d
	default:
		panic("parseDalDirective: unknown dal spec " + spec)
	}
}
