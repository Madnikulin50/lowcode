package tests

import (
	"context"
	"testing"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/madnikulin50/lowcode/server/pkg/id"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/stretchr/testify/require"
)

func testComposeResourceVersions(t *testing.T, s store.ComposeResourceVersions) {
	var (
		ctx = context.Background()

		makeNew = func(rtype string, resID uint64, version int) *types.ResourceVersion {
			return &types.ResourceVersion{
				ID:           id.Next(),
				NamespaceID:  id.Next(),
				ResourceType: rtype,
				ResourceID:   resID,
				Version:      version,
				Snapshot:     `{"handle":"home","title":"Дом\nline two"}`,
				CreatedAt:    *now(),
			}
		}
	)

	t.Run("create and lookup keep the snapshot", func(t *testing.T) {
		req := require.New(t)
		req.NoError(s.TruncateComposeResourceVersions(ctx))

		v := makeNew(types.ResourceVersionPage, 1, 1)
		req.NoError(s.CreateComposeResourceVersion(ctx, v))

		got, err := s.LookupComposeResourceVersionByID(ctx, v.ID)
		req.NoError(err)
		req.Equal(v.Snapshot, got.Snapshot)
		req.Equal(types.ResourceVersionPage, got.ResourceType)
		req.Equal(1, got.Version)
	})

	t.Run("search by resource, newest first", func(t *testing.T) {
		req := require.New(t)
		req.NoError(s.TruncateComposeResourceVersions(ctx))

		for i := 1; i <= 3; i++ {
			req.NoError(s.CreateComposeResourceVersion(ctx, makeNew(types.ResourceVersionPage, 7, i)))
		}
		req.NoError(s.CreateComposeResourceVersion(ctx, makeNew(types.ResourceVersionChart, 7, 1)))
		req.NoError(s.CreateComposeResourceVersion(ctx, makeNew(types.ResourceVersionPage, 8, 1)))

		set, _, err := s.SearchComposeResourceVersions(ctx, types.ResourceVersionFilter{
			ResourceType: types.ResourceVersionPage,
			ResourceID:   7,
			Sorting:      filter.Sorting{Sort: filter.SortExprSet{{Column: "version", Descending: true}}},
		})
		req.NoError(err)
		req.Len(set, 3)
		req.Equal(3, set[0].Version)
		req.Equal(1, set[2].Version)
	})

	t.Run("delete by id", func(t *testing.T) {
		req := require.New(t)
		req.NoError(s.TruncateComposeResourceVersions(ctx))

		v := makeNew(types.ResourceVersionModule, 1, 1)
		req.NoError(s.CreateComposeResourceVersion(ctx, v))
		req.NoError(s.DeleteComposeResourceVersionByID(ctx, v.ID))

		_, err := s.LookupComposeResourceVersionByID(ctx, v.ID)
		req.ErrorIs(err, store.ErrNotFound)
	})
}
