package sdk

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// FlexID accepts Compose snowflake IDs as JSON strings or numbers.
type FlexID uint64

func (id *FlexID) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "" || s == "null" || s == `""` {
		*id = 0
		return nil
	}
	if s[0] == '"' {
		var str string
		if err := json.Unmarshal(b, &str); err != nil {
			return err
		}
		str = strings.TrimSpace(str)
		if str == "" {
			*id = 0
			return nil
		}
		n, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return err
		}
		*id = FlexID(n)
		return nil
	}
	var n uint64
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	*id = FlexID(n)
	return nil
}

func (id FlexID) MarshalJSON() ([]byte, error) {
	if id == 0 {
		return []byte(`""`), nil
	}
	return json.Marshal(strconv.FormatUint(uint64(id), 10))
}

func (id FlexID) Uint64() uint64 { return uint64(id) }

func (id FlexID) String() string {
	if id == 0 {
		return ""
	}
	return strconv.FormatUint(uint64(id), 10)
}

func ParseID(v any) uint64 {
	switch t := v.(type) {
	case nil:
		return 0
	case FlexID:
		return uint64(t)
	case uint64:
		return t
	case int:
		if t < 0 {
			return 0
		}
		return uint64(t)
	case int64:
		if t < 0 {
			return 0
		}
		return uint64(t)
	case float64:
		if t <= 0 {
			return 0
		}
		return uint64(t)
	case json.Number:
		n, _ := t.Int64()
		if n < 0 {
			return 0
		}
		return uint64(n)
	case string:
		s := strings.TrimSpace(t)
		if s == "" || s == "0" {
			return 0
		}
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			fmt.Sscanf(s, "%d", &n)
		}
		return n
	default:
		s := strings.TrimSpace(fmt.Sprintf("%v", t))
		n, _ := strconv.ParseUint(s, 10, 64)
		return n
	}
}
