package chat

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/require"
)

func TestParamString(t *testing.T) {
	for _, c := range []struct {
		in   any
		want string
	}{
		{nil, ""},
		{"text", "text"},
		{true, "true"},
		{false, "false"},
		{float64(3), "3"},
		{float64(1000000), "1000000"}, // %v would print 1e+06
		{3.5, "3.5"},
		{float64(-2), "-2"},
		{json.Number("12"), "12"},
		{[]any{"a", "b"}, `["a","b"]`},
		{map[string]any{"k": 1.0}, `{"k":1}`}, // %v would print map[k:1]
		{[]any{}, `[]`},
	} {
		require.Equal(t, c.want, ParamString(c.in), "%#v", c.in)
	}
}

func TestParamInfo_TypesTheSchema(t *testing.T) {
	for _, c := range []struct {
		typ  string
		want schema.DataType
	}{
		{"", schema.String},
		{"string", schema.String},
		{"json", schema.String}, // text that contains JSON, as it always was
		{"unknown", schema.String},
		{"number", schema.Number},
		{"integer", schema.Integer},
		{"boolean", schema.Boolean},
		{"array", schema.Array},
		{"object", schema.Object},
	} {
		info := paramInfo(ParamDef{Name: "p", Type: c.typ, Required: true, Description: "d"})
		require.Equal(t, c.want, info.Type, "type %q", c.typ)
		require.True(t, info.Required)
		require.Equal(t, "d", info.Desc)
	}
	require.NotNil(t, paramInfo(ParamDef{Type: "array"}).ElemInfo, "an array schema needs its element type")
}

func TestToolInfos_CarryTheTypes(t *testing.T) {
	defs := []ToolDef{{
		Name: "t",
		Params: []ParamDef{
			{Name: "count", Type: "integer"},
			{Name: "ratio", Type: "number"},
			{Name: "async", Type: "boolean"},
			{Name: "ids", Type: "array"},
			{Name: "filter", Type: "object"},
			{Name: "name"},
		},
	}}

	infos, err := ToToolInfos(defs)
	require.NoError(t, err)
	require.Len(t, infos, 1)

	// the schema a model actually receives, as JSON schema
	js, err := infos[0].ParamsOneOf.ToJSONSchema()
	require.NoError(t, err)
	raw, _ := json.Marshal(js)

	var got struct {
		Properties map[string]struct {
			Type string `json:"type"`
		} `json:"properties"`
	}
	require.NoError(t, json.Unmarshal(raw, &got))
	require.Equal(t, "integer", got.Properties["count"].Type)
	require.Equal(t, "number", got.Properties["ratio"].Type)
	require.Equal(t, "boolean", got.Properties["async"].Type)
	require.Equal(t, "array", got.Properties["ids"].Type)
	require.Equal(t, "object", got.Properties["filter"].Type)
	require.Equal(t, "string", got.Properties["name"].Type)

	// the adapter used for invocation describes the same schema
	adapted, err := (&toolAdapter{def: defs[0]}).Info(context.Background())
	require.NoError(t, err)
	js2, _ := adapted.ParamsOneOf.ToJSONSchema()
	raw2, _ := json.Marshal(js2)
	require.JSONEq(t, string(raw), string(raw2))
}

func TestToolAdapter_HandlerGetsTypedArgumentsAsText(t *testing.T) {
	var got map[string]string
	a := &toolAdapter{def: ToolDef{Name: "t", Handler: func(_ context.Context, p map[string]string) string {
		got = p
		return "ok"
	}}}

	_, err := a.InvokableRun(context.Background(),
		`{"count": 5, "ratio": 0.25, "async": true, "ids": ["a","b"], "filter": {"status": "open", "n": 2}, "name": "x"}`)
	require.NoError(t, err)

	require.Equal(t, "5", got["count"])
	require.Equal(t, "0.25", got["ratio"])
	require.Equal(t, "true", got["async"])
	require.Equal(t, `["a","b"]`, got["ids"])
	require.JSONEq(t, `{"status":"open","n":2}`, got["filter"])
	require.Equal(t, "x", got["name"])
}

func TestToolSystemPrompt_ShowsNonStringTypes(t *testing.T) {
	out := ToolSystemPrompt([]ToolDef{{Name: "t", Description: "d", Params: []ParamDef{
		{Name: "count", Type: "integer", Required: true, Description: "how many"},
		{Name: "name", Description: "who"},
	}}})
	require.Contains(t, out, "count [integer] (required): how many")
	require.Contains(t, out, "- name: who", "plain strings stay uncluttered")
}

func TestParamInfo_ArrayOfObjects(t *testing.T) {
	info := paramInfo(ParamDef{Name: "nodes", Type: "objects"})
	require.Equal(t, schema.Array, info.Type)
	require.Equal(t, schema.Object, info.ElemInfo.Type)

	var got map[string]string
	a := &toolAdapter{def: ToolDef{Name: "t", Handler: func(_ context.Context, p map[string]string) string { got = p; return "" }}}
	_, err := a.InvokableRun(context.Background(), `{"nodes":[{"id":"a","type":"mail","config":{"to":"x"}}]}`)
	require.NoError(t, err)
	require.JSONEq(t, `[{"id":"a","type":"mail","config":{"to":"x"}}]`, got["nodes"], "the handler can json.Unmarshal what it receives")
}
