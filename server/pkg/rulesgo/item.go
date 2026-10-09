package rulesgo

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// NormalizeIngestEnvelope makes webhook/poller input safe for CRUD/foreach:
// snowflake IDs as strings, scanRecordID↔createdRecordID aliases, items as []maps.
func NormalizeIngestEnvelope(in map[string]interface{}) map[string]interface{} {
	if in == nil {
		in = map[string]interface{}{}
	}
	for _, k := range []string{"namespaceID", "scanRecordID", "createdRecordID", "recordID", "jobID", "scanID", "moduleID"} {
		v, ok := in[k]
		if !ok || v == nil {
			continue
		}
		if s := snowflakeString(v); s != "" {
			in[k] = s
			continue
		}
		if _, isFloat := v.(float64); isFloat && (k == "namespaceID" || k == "moduleID") {
			delete(in, k)
		}
	}
	if emptyAny(in["scanRecordID"]) && !emptyAny(in["createdRecordID"]) {
		in["scanRecordID"] = in["createdRecordID"]
	}
	if emptyAny(in["createdRecordID"]) && !emptyAny(in["scanRecordID"]) {
		in["createdRecordID"] = in["scanRecordID"]
	}
	if items := jsonToItems(in["items"]); len(items) > 0 {
		in["items"] = items
	} else if items := jsonToItems(in["devices"]); len(items) > 0 {
		in["items"] = items
	} else if items := jsonToItems(in); len(items) > 0 {
		in["items"] = items
	}
	return in
}

func snowflakeString(v interface{}) string {
	switch t := v.(type) {
	case json.Number:
		s := strings.TrimSpace(string(t))
		if s != "" && s != "0" {
			return s
		}
	case string:
		s := strings.TrimSpace(t)
		if s != "" && s != "<nil>" && s != "0" {
			return s
		}
	case uint64:
		if t > 0 {
			return strconv.FormatUint(t, 10)
		}
	case int:
		if t > 0 {
			return strconv.FormatInt(int64(t), 10)
		}
	case int64:
		if t > 0 {
			return strconv.FormatInt(t, 10)
		}
	case float64:
		if t > 0 && t < float64(uint64(1)<<53) {
			return strconv.FormatUint(uint64(t), 10)
		}
	}
	return ""
}

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
	case map[string]interface{}:
		for _, k := range []string{"items", "devices", "set", "response"} {
			if items := CollectItems(t[k]); len(items) > 0 {
				return items
			}
		}
		return nil
	case string:
		s := strings.TrimSpace(t)
		if s == "" || s == "null" {
			return nil
		}
		if s[0] == '[' || s[0] == '{' {
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
		if _, ok := any.(map[string]interface{}); ok {
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
