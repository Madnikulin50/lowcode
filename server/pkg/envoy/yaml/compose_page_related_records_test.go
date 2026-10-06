package yaml

import (
	"testing"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/stretchr/testify/require"
)

// modules are written by handle, in the order of the relations that name one
func TestComposePageBlock_RelatedRecordsWritesModuleHandles(t *testing.T) {
	req := require.New(t)

	c := &composePageBlock{
		res: &types.PageBlock{Kind: "RelatedRecords", Options: map[string]interface{}{
			"relations": []interface{}{
				map[string]interface{}{"moduleID": "111", "refField": "object"},
				map[string]interface{}{"moduleID": "0", "refField": "x"},
				map[string]interface{}{"moduleID": "333", "refField": "object"},
			},
		}},
		refMod: []string{"positions_voisr", "files_id"},
	}

	_, err := c.MarshalYAML()
	req.NoError(err)

	rels := c.res.Options["relations"].([]interface{})
	req.Equal("positions_voisr", rels[0].(map[string]interface{})["module"])
	req.NotContains(rels[0], "moduleID")
	req.Equal("0", rels[1].(map[string]interface{})["moduleID"], "no module, nothing to write")
	req.Equal("files_id", rels[2].(map[string]interface{})["module"])
}

// the references are collected only for the items that name a module, so they
// have to be matched to those items and not to every item by position
func TestComposePageBlock_MetricsAndFeedsWithoutModuleDoNotShiftTheRest(t *testing.T) {
	req := require.New(t)

	m := &composePageBlock{
		res: &types.PageBlock{Kind: "Metric", Options: map[string]interface{}{
			"metrics": []interface{}{
				map[string]interface{}{"moduleID": "0"},
				map[string]interface{}{"moduleID": "111"},
				map[string]interface{}{"moduleID": "222"},
			},
		}},
		refMod: []string{"orders", "clients"},
	}
	_, err := m.MarshalYAML()
	req.NoError(err)
	mm := m.res.Options["metrics"].([]interface{})
	req.NotContains(mm[0], "module")
	req.Equal("orders", mm[1].(map[string]interface{})["module"])
	req.Equal("clients", mm[2].(map[string]interface{})["module"])

	c := &composePageBlock{
		res: &types.PageBlock{Kind: "Calendar", Options: map[string]interface{}{
			"feeds": []interface{}{
				map[string]interface{}{"options": map[string]interface{}{}},
				map[string]interface{}{"options": map[string]interface{}{"moduleID": "9"}},
			},
		}},
		refMod: []string{"events"},
	}
	_, err = c.MarshalYAML() // used to index refMod[0] for the feed without a module, and fail for the other
	req.NoError(err)
	ff := c.res.Options["feeds"].([]interface{})
	req.Equal("events", ff[1].(map[string]interface{})["options"].(map[string]interface{})["module"])
}
