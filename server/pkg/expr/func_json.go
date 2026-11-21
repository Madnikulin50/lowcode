package expr

import (
	"encoding/json"
	"github.com/PaesslerAG/gval"
)

func JsonFunctions() []gval.Language {
	return []gval.Language{
		// TODO: Json decoding, parsing, stringify, JQ, JSONPath
		gval.Function("toJSON", toJSON),
		gval.Function("toPlainJSON", toPlainJSON),
	}
}

func toJSON(f interface{}) string {
	if _, is := f.(json.Marshaler); !is {
		f = UntypedValue(f)
	}
	b, _ := json.Marshal(f)
	return string(b)
}

// toPlainJSON is toJSON without the type envelope: variables come out as
// ordinary JSON ({"severity":"high"}) instead of {"severity":{"@value":"high",
// "@type":"String"}}. That is the form to hand to an LLM or an external API;
// toJSON keeps the typed form, which round-trips back into workflow variables.
func toPlainJSON(f interface{}) string {
	switch v := f.(type) {
	case *Vars:
		f = v.Dict()
	case TypedValue:
		f = UntypedValue(v)
	default:
		f = UntypedValue(f)
	}
	b, _ := json.Marshal(f)
	return string(b)
}
