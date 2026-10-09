package expr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Example_toJSON() {
	var (
		p = map[string]interface{}{
			"vars": Must(NewVars(
				&Vars{value: map[string]TypedValue{
					"k1": &String{value: "v1"},
					"k2": &String{value: "v2"},
				}},
			)),
		}
	)

	eval(`toJSON(vars)`, p)

	// output:
	// {"k1":{"@value":"v1","@type":"String"},"k2":{"@value":"v2","@type":"String"}}
}

func Example_kv_toJSON() {
	var (
		p = map[string]interface{}{
			"kv": Must(NewKV(
				&KV{value: map[string]string{
					"k1": "v1",
					"k2": "v2",
				}},
			)),
		}
	)

	eval(`toJSON(kv)`, p)

	// output:
	// {"k1":"v1","k2":"v2"}
}

func TestToPlainJSON(t *testing.T) {
	vars, err := NewVars(map[string]interface{}{
		"severity": "high",
		"score":    3.5,
		"nested":   map[string]interface{}{"detector": "zscore"},
	})
	require.NoError(t, err)

	got := toPlainJSON(vars)
	require.JSONEq(t, `{"severity":"high","score":3.5,"nested":{"detector":"zscore"}}`, got)
	require.NotContains(t, got, "@type")

	// the typed form is unchanged
	require.Contains(t, toJSON(vars), "@type")

	require.Equal(t, `"x"`, toPlainJSON(&String{value: "x"}))
	require.Equal(t, `[1,2]`, toPlainJSON([]int{1, 2}))
}
