package def

import (
	"reflect"
	"testing"
)

func TestIdentFromGoName(t *testing.T) {
	cases := map[string]string{
		"ID":          "id",
		"Handle":      "handle",
		"NamespaceID": "namespaceID",
		"OwnedBy":     "ownedBy",
		"CreatedAt":   "createdAt",
		"URL":         "url", // whole-name acronym
	}
	for in, want := range cases {
		if got := identFromGoName(in); got != want {
			t.Errorf("identFromGoName(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestParseDalDirectiveRefColons guards against the regression found this
// session: a naive strings.Split(spec, ":") truncated a RefModelResType
// that itself contains "::" (e.g. federation's "corteza::federation:node"),
// silently dropping everything after the first colon.
func TestParseDalDirectiveRefColons(t *testing.T) {
	cases := []struct {
		name           string
		spec           string
		wantResType    string
		wantHasDefault bool
	}{
		{
			name:        "simple resource type",
			spec:        "ref:corteza::compose:namespace",
			wantResType: "corteza::compose:namespace",
		},
		{
			name:           "resource type with default0 suffix",
			spec:           "ref:corteza::federation:node:default0",
			wantResType:    "corteza::federation:node",
			wantHasDefault: true,
		},
		{
			name:        "single-segment resource type",
			spec:        "ref:corteza::compose:module",
			wantResType: "corteza::compose:module",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := parseDalDirective(c.spec)
			if d.Type != "Ref" {
				t.Fatalf("Type = %q, want Ref", d.Type)
			}
			if d.RefModelResType != c.wantResType {
				t.Errorf("RefModelResType = %q, want %q", d.RefModelResType, c.wantResType)
			}
			if d.HasDefault != c.wantHasDefault {
				t.Errorf("HasDefault = %v, want %v", d.HasDefault, c.wantHasDefault)
			}
		})
	}
}

func TestParseDalDirectiveKinds(t *testing.T) {
	if got := parseDalDirective("id"); got.Type != "ID" {
		t.Errorf("id: Type = %q, want ID", got.Type)
	}
	if got := parseDalDirective("userref"); got != userRefDal {
		t.Errorf("userref: got %#v, want the shared userRefDal instance", got)
	}
	if got := parseDalDirective("timestamp:now"); !got.DefaultCurrentTimestamp {
		t.Errorf("timestamp:now: DefaultCurrentTimestamp = false, want true")
	}
	if got := parseDalDirective("timestamp:nil"); !got.Nullable {
		t.Errorf("timestamp:nil: Nullable = false, want true")
	}
	if got := parseDalDirective("text:64"); got.Length != 64 {
		t.Errorf("text:64: Length = %d, want 64", got.Length)
	}
	if got := parseDalDirective("json:empty"); !got.DefaultEmptyObject {
		t.Errorf("json:empty: DefaultEmptyObject = false, want true")
	}
	if got := parseDalDirective("number:default0"); !got.HasDefault || got.DefaultValue != 0 {
		t.Errorf("number:default0: HasDefault=%v DefaultValue=%v, want true/0", got.HasDefault, got.DefaultValue)
	}
	if got := parseDalDirective("number"); got.Meta["rdbms:type"] != "integer" {
		t.Errorf(`number: Meta["rdbms:type"] = %v, want "integer"`, got.Meta["rdbms:type"])
	}
	if got := parseDalDirective("bool:true"); !got.HasDefault || got.DefaultValue != true {
		t.Errorf("bool:true: HasDefault=%v DefaultValue=%v, want true/true", got.HasDefault, got.DefaultValue)
	}
}

func TestParseDalDirectiveUnknownPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected a panic for an unknown dal kind")
		}
	}()
	parseDalDirective("bogus:thing")
}

type reflectFixture struct {
	ID       uint64 `schema:"col=id,dal=id,unique"`
	Handle   string `schema:"col=handle,dal=text:64,sortable,ignoreCase"`
	Internal string // no schema tag - must be skipped
}

func TestAttributesFromStruct(t *testing.T) {
	got := AttributesFromStruct(reflectFixture{})

	want := []string{"id", "handle"}
	if len(got) != len(want) {
		t.Fatalf("got %d attributes, want %d: %#v", len(got), len(want), got)
	}
	for i, name := range want {
		if got[i].Name != name {
			t.Errorf("attribute %d: Name = %q, want %q", i, got[i].Name, name)
		}
	}

	if got[0].Attribute.ExpIdent != "ID" {
		t.Errorf("id attribute ExpIdent = %q, want %q", got[0].Attribute.ExpIdent, "ID")
	}
	if !got[0].Attribute.Unique {
		t.Errorf("id attribute should be Unique")
	}

	handle := got[1].Attribute
	if !handle.Sortable || !handle.IgnoreCase {
		t.Errorf("handle attribute Sortable=%v IgnoreCase=%v, want both true", handle.Sortable, handle.IgnoreCase)
	}
	if handle.Dal == nil || handle.Dal.Type != "Text" || handle.Dal.Length != 64 {
		t.Errorf("handle attribute Dal = %#v, want Text length 64", handle.Dal)
	}
}

type reflectFixtureMissingCol struct {
	Broken string `schema:"sortable"`
}

func TestAttributesFromStructPanicsWithoutCol(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected a panic when schema tag has no col=")
		}
	}()
	AttributesFromStruct(reflectFixtureMissingCol{})
}

func TestParseSchemaTag(t *testing.T) {
	got := parseSchemaTag("col=handle, sortable ,ignoreCase,alias=a|b")
	want := map[string]string{"col": "handle", "sortable": "", "ignoreCase": "", "alias": "a|b"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseSchemaTag = %#v, want %#v", got, want)
	}
}
