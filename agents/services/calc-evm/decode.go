package calcevm

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func InputFromParams(projectID string, params map[string]any) (Input, error) {
	if params == nil {
		params = map[string]any{}
	}
	in := Input{
		ProjectID: firstNonEmpty(projectID, stringify(params["projectID"]), stringify(params["project"])),
		Items:     coerceItems(params["items"]),
		Facts:     coerceFacts(params["facts"]),
	}
	if t := stringify(params["now"]); t != "" {
		if now, ok := parseTime(t); ok {
			in.Now = now
		}
	}
	if len(in.Items) == 0 {
		return in, fmt.Errorf("items required")
	}
	return in, nil
}

func coerceItems(v any) []Item {
	var out []Item
	for _, m := range asMaps(v) {
		it := Item{
			ID:              firstNonEmpty(stringify(m["id"]), stringify(m["recordID"]), stringify(m["record_id"])),
			ProjectID:       firstNonEmpty(stringify(m["projectID"]), stringify(m["project"])),
			BudgetPlanned:   parseFloat(firstNonNil(m["budgetPlanned"], m["budget_planned"])),
			PercentComplete: parseFloat(firstNonNil(m["percentComplete"], m["percent_complete"])),
			ActualCost:      parseFloat(firstNonNil(m["actualCost"], m["actual_cost"])),
		}
		if t, ok := parseTime(firstNonNil(m["startPlanned"], m["start_planned"])); ok {
			it.StartPlanned = t
		}
		if t, ok := parseTime(firstNonNil(m["endPlanned"], m["end_planned"])); ok {
			it.EndPlanned = t
		}
		if it.ID == "" && it.BudgetPlanned == 0 && it.ProjectID == "" {
			continue
		}
		out = append(out, it)
	}
	return out
}

func coerceFacts(v any) []Fact {
	var out []Fact
	for _, m := range asMaps(v) {
		f := Fact{
			WBSID:     firstNonEmpty(stringify(m["wbsID"]), stringify(m["wbs"])),
			ProjectID: firstNonEmpty(stringify(m["projectID"]), stringify(m["project"])),
			Percent:   parseFloat(m["percent"]),
			Cost:      parseFloat(m["cost"]),
		}
		if f.WBSID == "" && f.Percent == 0 && f.Cost == 0 {
			continue
		}
		out = append(out, f)
	}
	return out
}

func asMaps(v any) []map[string]any {
	if v == nil {
		return nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var maps []map[string]any
	if json.Unmarshal(raw, &maps) == nil {
		return maps
	}
	var one map[string]any
	if json.Unmarshal(raw, &one) == nil {
		return []map[string]any{one}
	}
	return nil
}

func stringify(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case json.Number:
		return strings.TrimSpace(string(t))
	default:
		b, _ := json.Marshal(t)
		s := string(b)
		if len(s) >= 2 && s[0] == '"' {
			var out string
			if json.Unmarshal(b, &out) == nil {
				return strings.TrimSpace(out)
			}
		}
		return strings.TrimSpace(s)
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" && v != "0" && v != "<nil>" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func firstNonNil(vals ...any) any {
	for _, v := range vals {
		if v != nil {
			return v
		}
	}
	return nil
}

func parseFloat(v any) float64 {
	if v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case json.Number:
		n, _ := t.Float64()
		return n
	default:
		s := stringify(t)
		if s == "" {
			return 0
		}
		n, _ := strconv.ParseFloat(s, 64)
		return n
	}
}

func parseTime(v any) (time.Time, bool) {
	s := stringify(v)
	if s == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02", "2006-01-02T15:04:05", "2006-01-02 15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
