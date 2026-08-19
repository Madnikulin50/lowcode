package rulesgo

import (
	"encoding/json"
	"fmt"
	"strings"
)

func CollectItems(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case []interface{}:
		return t
	case []map[string]interface{}:
		out := make([]interface{}, len(t))
		for i := range t {
			out[i] = t[i]
		}
		return out
	case json.RawMessage:
		var any interface{}
		if json.Unmarshal(t, &any) != nil {
			return nil
		}
		return CollectItems(any)
	case string:
		s := strings.TrimSpace(t)
		if s == "" || s == "null" {
			return nil
		}
		if s[0] == '[' {
			var any interface{}
			if json.Unmarshal([]byte(s), &any) != nil {
				return nil
			}
			return CollectItems(any)
		}
		return nil
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return nil
		}
		var any interface{}
		if json.Unmarshal(b, &any) != nil {
			return nil
		}
		if _, ok := any.([]interface{}); ok {
			return CollectItems(any)
		}
		return nil
	}
}

func FlattenItem(ec *ExecutionContext, prefix string, item interface{}) {
	if prefix == "" {
		prefix = "item"
	}
	ec.Set(prefix, item)
	m := asStringMap(item)
	if m == nil {
		return
	}
	for k, v := range m {
		key := prefix + "." + k
		switch v.(type) {
		case map[string]interface{}, []interface{}:
			b, err := json.Marshal(v)
			if err != nil {
				ec.Set(key, fmt.Sprintf("%v", v))
				continue
			}
			ec.Set(key, string(b))
		default:
			ec.Set(key, v)
		}
	}
}

func asStringMap(v interface{}) map[string]interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		return t
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return nil
		}
		var m map[string]interface{}
		if json.Unmarshal(b, &m) != nil {
			return nil
		}
		return m
	}
}
