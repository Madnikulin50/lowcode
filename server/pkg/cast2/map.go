package cast2

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/modern-go/reflect2"
)

// Meta casts a value to a map[string]any.
// Empty, whitespace-only, and JSON-null inputs decode as an empty map.
// Corrupt JSON still returns an error (do not swallow garbage).
func Meta(in any, out *map[string]any) (err error) {
	if reflect2.IsNil(in) {
		*out = map[string]any{}
		return nil
	}

	switch aux := in.(type) {
	case []byte:
		err = unmarshalMeta(aux, out)
	case json.RawMessage:
		err = unmarshalMeta(aux, out)
	case string:
		err = unmarshalMeta([]byte(aux), out)
	case map[string]any:
		if aux == nil {
			*out = map[string]any{}
		} else {
			*out = aux
		}
	default:
		err = fmt.Errorf("unsupported type: %T", in)
	}

	if err == nil {
		return
	}

	return fmt.Errorf("can not cast to Meta: %w", err)
}

func unmarshalMeta(raw []byte, out *map[string]any) error {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		*out = map[string]any{}
		return nil
	}
	return json.Unmarshal(raw, out)
}
