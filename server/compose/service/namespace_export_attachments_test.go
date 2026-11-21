package service

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/stretchr/testify/require"
)

func TestRemapRecordFileJSONL(t *testing.T) {
	mod := &types.Module{Fields: types.ModuleFieldSet{
		{Name: "title", Kind: "String"},
		{Name: "file", Kind: "File"},
	}}
	idMap := map[uint64]uint64{100: 200}

	t.Run("string id", func(t *testing.T) {
		req := require.New(t)
		out, err := RemapRecordFileJSONL(strings.NewReader("{\"title\":\"x\",\"file\":\"100\"}\n"), mod, idMap)
		req.NoError(err)
		raw, err := io.ReadAll(out)
		req.NoError(err)
		req.Contains(string(raw), `"file":"200"`)
		req.Contains(string(raw), `"title":"x"`)
	})

	t.Run("numeric id preserves snowflake digits", func(t *testing.T) {
		req := require.New(t)
		const oldID uint64 = 492476413458710529
		const newID uint64 = 496736350624153601
		src := `{"file":` + json.Number("492476413458710529").String() + "}\n"
		out, err := RemapRecordFileJSONL(strings.NewReader(src), mod, map[uint64]uint64{oldID: newID})
		req.NoError(err)
		raw, err := io.ReadAll(out)
		req.NoError(err)
		req.Contains(string(raw), `"file":"496736350624153601"`)
	})

	t.Run("unknown id dropped", func(t *testing.T) {
		req := require.New(t)
		out, err := RemapRecordFileJSONL(strings.NewReader("{\"file\":\"999\"}\n"), mod, idMap)
		req.NoError(err)
		raw, err := io.ReadAll(out)
		req.NoError(err)
		req.NotContains(string(raw), `"file"`)
	})
}
