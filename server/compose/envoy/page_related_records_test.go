package envoy

import (
	"context"
	"testing"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/envoyx"
	"github.com/stretchr/testify/require"
)

func relatedRecordsPage() *types.Page {
	return &types.Page{Blocks: types.PageBlocks{
		{Kind: "Content"},
		{Kind: "RelatedRecords", Options: map[string]interface{}{
			"hideEmptySections": true,
			"relations": []interface{}{
				map[string]interface{}{"moduleID": "111", "refField": "object", "title": "ВОИСР"},
				map[string]interface{}{"moduleID": "0", "refField": "x"}, // no module: no reference, no shift
				map[string]interface{}{"moduleID": "333", "refField": "object"},
			},
		}},
	}}
}

func TestRelatedRecords_ReferencesModulesByRelationPosition(t *testing.T) {
	req := require.New(t)

	refs := decodePageRefs(relatedRecordsPage())
	req.Len(refs, 2)

	first := refs["Blocks.1.Options.relations.0.ModuleID"]
	req.Equal(types.ModuleResourceType, first.ResourceType)
	req.Equal("111", first.Identifiers.FriendlyIdentifier())

	// the relation without a module does not move the next one's key
	third := refs["Blocks.1.Options.relations.2.ModuleID"]
	req.Equal("333", third.Identifiers.FriendlyIdentifier())
	req.NotContains(refs, "Blocks.1.Options.relations.1.ModuleID")
}

func TestRelatedRecords_ReadsModulesGivenByHandleToo(t *testing.T) {
	refs := getPageBlockRelatedRecordsRefs(types.PageBlock{Kind: "RelatedRecords", Options: map[string]interface{}{
		"relations": []interface{}{map[string]interface{}{"module": "positions_voisr", "refField": "object"}},
	}}, 0)

	require.Equal(t, "positions_voisr", refs["Blocks.0.Options.relations.0.ModuleID"].Identifiers.FriendlyIdentifier())
}

type fakeTree struct {
	envoyx.Traverser
	handles map[string]string // module id -> handle
}

func (f fakeTree) ParentForRef(_ *envoyx.Node, ref envoyx.Ref) *envoyx.Node {
	h, ok := f.handles[ref.Identifiers.FriendlyIdentifier()]
	if !ok {
		return nil
	}
	return &envoyx.Node{ResourceType: types.ModuleResourceType, Identifiers: envoyx.MakeIdentifiers(h)}
}

func TestRelatedRecords_ExportsModulesByHandle(t *testing.T) {
	req := require.New(t)

	pg := relatedRecordsPage()
	node := &envoyx.Node{References: decodePageRefs(pg)}

	out, err := YamlEncoder{}.encodePageBlockC(context.Background(), envoyx.EncodeParams{},
		fakeTree{handles: map[string]string{"111": "positions_voisr", "333": "files_id"}}, node, pg, 1, pg.Blocks[1])
	req.NoError(err)

	b := out.(types.PageBlock)
	rels := b.Options["relations"].([]interface{})
	req.Equal("positions_voisr", rels[0].(map[string]interface{})["module"])
	req.NotContains(rels[0], "moduleID")
	req.Equal("object", rels[0].(map[string]interface{})["refField"])

	// no reference, nothing to translate: left as it was
	req.Equal("0", rels[1].(map[string]interface{})["moduleID"])

	req.Equal("files_id", rels[2].(map[string]interface{})["module"])
	req.Equal(true, b.Options["hideEmptySections"])
}

func TestRelatedRecords_ImportSetsTheNewModuleIDs(t *testing.T) {
	req := require.New(t)

	pg := relatedRecordsPage()
	for k, id := range map[string]uint64{
		"Blocks.1.Options.relations.0.ModuleID": 9001,
		"Blocks.1.Options.relations.2.ModuleID": 9003,
	} {
		req.NoError(pg.SetValue(k, 0, id))
	}

	rels := pg.Blocks[1].Options["relations"].([]interface{})
	req.Equal("9001", rels[0].(map[string]interface{})["moduleID"])
	req.Equal("0", rels[1].(map[string]interface{})["moduleID"])
	req.Equal("9003", rels[2].(map[string]interface{})["moduleID"])
	req.Equal("object", rels[2].(map[string]interface{})["refField"], "the rest of a relation is kept")

	// a path that does not exist is not a crash
	req.NoError(pg.SetValue("Blocks.1.Options.relations.7.ModuleID", 0, uint64(1)))
	req.NoError(pg.SetValue("Blocks.1.Options.relations", 0, uint64(1)))
}

// a metric or a feed without a module must not hide the references of the ones after it
func TestPageBlockRefs_ItemWithoutModuleDoesNotStopTheSearch(t *testing.T) {
	req := require.New(t)

	metrics := getPageBlockMetricRefs(types.PageBlock{Kind: "Metric", Options: map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{"moduleID": "111"},
			map[string]interface{}{"moduleID": "0"},
			map[string]interface{}{"module": "orders"},
		},
	}}, 4)
	req.Len(metrics, 2)
	req.Equal("111", metrics["Blocks.4.Options.metrics.0.ModuleID"].Identifiers.FriendlyIdentifier())
	req.Equal("orders", metrics["Blocks.4.Options.metrics.2.ModuleID"].Identifiers.FriendlyIdentifier())

	feeds := getPageBlockCalendarRefs(types.PageBlock{Kind: "Calendar", Options: map[string]interface{}{
		"feeds": []interface{}{
			map[string]interface{}{"options": map[string]interface{}{}},
			map[string]interface{}{"options": map[string]interface{}{"moduleID": "222"}},
		},
	}}, 0)
	req.Len(feeds, 1)
	req.Equal("222", feeds["Blocks.0.Options.feeds.1.ModuleID"].Identifiers.FriendlyIdentifier())
}

func TestPageBlockExport_ItemWithoutModuleIsLeftAlone(t *testing.T) {
	req := require.New(t)

	pg := &types.Page{Blocks: types.PageBlocks{{Kind: "Metric", Options: map[string]interface{}{
		"metrics": []interface{}{
			map[string]interface{}{"moduleID": "0", "label": "none"},
			map[string]interface{}{"moduleID": "111", "label": "orders"},
		},
	}}}}
	node := &envoyx.Node{References: decodePageRefs(pg)}

	out, err := YamlEncoder{}.encodePageBlockC(context.Background(), envoyx.EncodeParams{},
		fakeTree{handles: map[string]string{"111": "orders_mod"}}, node, pg, 0, pg.Blocks[0])
	req.NoError(err)

	mm := out.(types.PageBlock).Options["metrics"].([]interface{})
	req.Equal("0", mm[0].(map[string]interface{})["moduleID"], "no module: kept as it was, not turned into an empty handle")
	req.NotContains(mm[0], "module")
	req.Equal("orders_mod", mm[1].(map[string]interface{})["module"])
}
