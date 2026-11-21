package service

import (
	"context"
	"testing"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/drivers/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestResourceVersions(t *testing.T) {
	var (
		req = require.New(t)
		ctx = context.Background()
	)

	s, err := sqlite.ConnectInMemory(ctx)
	req.NoError(err)
	req.NoError(store.Upgrade(ctx, zap.NewNop(), s))

	ns := &types.Namespace{ID: nextID(), Name: "ns", Slug: "ns", CreatedAt: *now()}
	req.NoError(store.CreateComposeNamespace(ctx, s, ns))

	ch := &types.Chart{ID: nextID(), NamespaceID: ns.ID, Handle: "sales", Name: "Sales", CreatedAt: *now()}
	req.NoError(store.CreateComposeChart(ctx, s, ch))

	list := func() types.ResourceVersionSet {
		set, _, err := store.SearchComposeResourceVersions(ctx, s, types.ResourceVersionFilter{
			ResourceType: types.ResourceVersionChart, ResourceID: ch.ID,
		})
		req.NoError(err)
		return set
	}

	t.Run("first save creates version 1", func(t *testing.T) {
		req.NoError(saveResourceVersion(ctx, s, types.ResourceVersionChart, ns.ID, ch.ID, ""))
		set := list()
		req.Len(set, 1)
		req.Equal(1, set[0].Version)
		req.Contains(set[0].Snapshot, `"Sales"`)
	})

	t.Run("identical state is not repeated", func(t *testing.T) {
		req.NoError(saveResourceVersion(ctx, s, types.ResourceVersionChart, ns.ID, ch.ID, ""))
		req.Len(list(), 1)
	})

	t.Run("a change makes the next version", func(t *testing.T) {
		ch.Name = "Revenue"
		req.NoError(store.UpdateComposeChart(ctx, s, ch))
		req.NoError(saveResourceVersion(ctx, s, types.ResourceVersionChart, ns.ID, ch.ID, "renamed"))

		set := list()
		req.Len(set, 2)
		var latest *types.ResourceVersion
		for _, v := range set {
			if latest == nil || v.Version > latest.Version {
				latest = v
			}
		}
		req.Equal(2, latest.Version)
		req.Equal("renamed", latest.Note)
		req.Contains(latest.Snapshot, `"Revenue"`)
	})

	t.Run("oldest versions are dropped over the cap", func(t *testing.T) {
		for i := 0; i < resourceVersionsKept+5; i++ {
			ch.Name = "n" + string(rune('A'+i%26)) + string(rune('a'+i/26))
			req.NoError(store.UpdateComposeChart(ctx, s, ch))
			req.NoError(saveResourceVersion(ctx, s, types.ResourceVersionChart, ns.ID, ch.ID, ""))
		}

		set := list()
		req.Len(set, resourceVersionsKept)
		min, max := set[0].Version, set[0].Version
		for _, v := range set {
			if v.Version < min {
				min = v.Version
			}
			if v.Version > max {
				max = v.Version
			}
		}
		req.Equal(max-resourceVersionsKept+1, min, "kept versions are the newest and contiguous")
	})

	t.Run("unknown resource type", func(t *testing.T) {
		req.Error(saveResourceVersion(ctx, s, "widget", ns.ID, ch.ID, ""))
	})

	t.Run("rule chain versions", func(t *testing.T) {
		rs := rdbmsStore(s)
		req.NotNil(rs)

		rc := &types.RuleChain{ID: nextID(), Handle: "rc_test", Name: "First", EntryNode: "n1",
			Nodes: []byte(`[{"id":"n1","type":"condition"}]`), Edges: []byte(`[]`), CreatedAt: *now()}
		req.NoError(rdbms.CreateRuleChain(ctx, rs, rc))

		req.NoError(saveResourceVersion(ctx, s, types.ResourceVersionRuleChain, 0, rc.ID, ""))
		rc.Name = "Second"
		req.NoError(rdbms.UpdateRuleChain(ctx, rs, rc))
		req.NoError(saveResourceVersion(ctx, s, types.ResourceVersionRuleChain, 0, rc.ID, ""))

		set, _, err := store.SearchComposeResourceVersions(ctx, s, types.ResourceVersionFilter{
			ResourceType: types.ResourceVersionRuleChain, ResourceID: rc.ID,
		})
		req.NoError(err)
		req.Len(set, 2)

		svc := resourceVersion{store: s}
		v, err := svc.Get(ctx, 0, types.ResourceVersionRuleChain, rc.ID, 1)
		req.NoError(err)
		req.Contains(v.Snapshot, `"First"`)

		_, err = svc.Restore(ctx, 0, types.ResourceVersionRuleChain, rc.ID, 1)
		req.Error(err, "restored through the rule chain endpoint")
	})
}
