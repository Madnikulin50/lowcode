package resource

import (
	"testing"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/stretchr/testify/require"
)

// the modules whose records a RelatedRecords block lists are dependencies of
// its page: they have to exist before the page is imported
func TestComposePage_RelatedRecordsRefersToItsModules(t *testing.T) {
	req := require.New(t)

	pg := &types.Page{Handle: "object_page", Blocks: types.PageBlocks{
		{Kind: "RelatedRecords", Options: map[string]interface{}{
			"relations": []interface{}{
				map[string]interface{}{"moduleID": "111", "refField": "object"},
				map[string]interface{}{"module": "files_id", "refField": "object"},
				map[string]interface{}{"moduleID": "0", "refField": "x"},
			},
		}},
	}}

	r := NewComposePage(pg, MakeRef(types.NamespaceResourceType, MakeIdentifiers("ns")), nil, nil)

	req.Len(r.ModRefs, 2)
	req.Len(r.BlockRefs[0], 2)
	req.Equal(types.ModuleResourceType, r.ModRefs[0].ResourceType)
	req.Contains(r.ModRefs[0].Identifiers.StringSlice(), "111")
	req.Contains(r.ModRefs[1].Identifiers.StringSlice(), "files_id")
}
