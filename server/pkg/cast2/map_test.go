package cast2

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	var (
		req = require.New(t)
	)

	{
		target := make(map[string]any)
		req.NoError(Meta([]byte(`{"a":"b"}`), &target))
		req.Equal(map[string]any{"a": "b"}, target)
	}

	{
		target := make(map[string]any)
		req.NoError(Meta(`{"a":"b"}`, &target))
		req.Equal(map[string]any{"a": "b"}, target)
	}

	{
		target := make(map[string]any)
		req.NoError(Meta(map[string]any{"a": "b"}, &target))
		req.Equal(map[string]any{"a": "b"}, target)
	}
}

func TestMetaEmpty(t *testing.T) {
	req := require.New(t)
	empty := map[string]any{}

	cases := []struct {
		name string
		in   any
	}{
		{"nil", nil},
		{"empty bytes", []byte{}},
		{"empty string", ""},
		{"whitespace bytes", []byte("  \n\t")},
		{"whitespace string", "  \n\t"},
		{"json null bytes", []byte("null")},
		{"json null string", "null"},
		{"json.RawMessage empty", json.RawMessage{}},
		{"json.RawMessage null", json.RawMessage("null")},
		{"nil map", map[string]any(nil)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var target map[string]any
			req.NoError(Meta(tc.in, &target))
			req.Equal(empty, target)
			req.NotNil(target)
		})
	}
}

func TestMetaCorruptJSON(t *testing.T) {
	req := require.New(t)

	var target map[string]any
	err := Meta([]byte(`{`), &target)
	req.Error(err)
	req.Contains(err.Error(), "can not cast to Meta")

	err = Meta(`not-json`, &target)
	req.Error(err)
	req.Contains(err.Error(), "can not cast to Meta")
}
