package datasource

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cast"
)

type (
	RawRecord      map[string]RawRecordValue
	RawRecordValue struct {
		Name   string
		Values []string
	}
)

func (ar RawRecord) SetValue(name string, pos uint, value any) (err error) {
	aux := ar[name]

	aux.Name = name

	if pos >= uint(len(aux.Values)) {
		aux.Values = append(aux.Values, make([]string, pos-uint(len(aux.Values))+1)...)
	}

	aux.Values[pos] = cast.ToString(value)

	ar[name] = aux

	return
}

func (r *RawRecord) UnmarshalJSON(data []byte) error {
	raw := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	result := make(RawRecord)
	for k, v := range raw {
		var single string
		if err := json.Unmarshal(v, &single); err == nil {
			result[k] = RawRecordValue{
				Name:   k,
				Values: []string{single},
			}
			continue
		}

		var num json.Number
		if err := json.Unmarshal(v, &num); err == nil {
			result[k] = RawRecordValue{
				Name:   k,
				Values: []string{num.String()},
			}
			continue
		}

		var multi []string
		if err := json.Unmarshal(v, &multi); err == nil {
			result[k] = RawRecordValue{
				Name:   k,
				Values: multi,
			}
			continue
		}

		var multiNum []json.Number
		if err := json.Unmarshal(v, &multiNum); err == nil {
			vals := make([]string, 0, len(multiNum))
			for _, n := range multiNum {
				vals = append(vals, n.String())
			}
			result[k] = RawRecordValue{
				Name:   k,
				Values: vals,
			}
			continue
		}

		return fmt.Errorf("unexpected format for key %q", k)
	}

	*r = result
	return nil
}
