package service

import (
	"context"
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/drivers/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestValidateRecordDraft(t *testing.T) {
	ok := func() *types.RecordDraft {
		return &types.RecordDraft{ChangeID: " c1 ", NamespaceID: 1, ModuleID: 2, Payload: `{"a":1}`}
	}

	d := ok()
	require.NoError(t, validateRecordDraft(d))
	require.Equal(t, "c1", d.ChangeID, "trimmed")

	for name, mutate := range map[string]func(*types.RecordDraft){
		"no change id":   func(d *types.RecordDraft) { d.ChangeID = " " },
		"long change id": func(d *types.RecordDraft) { d.ChangeID = strings.Repeat("x", recordDraftMaxChangeID+1) },
		"no namespace":   func(d *types.RecordDraft) { d.NamespaceID = 0 },
		"no module":      func(d *types.RecordDraft) { d.ModuleID = 0 },
		"not json":       func(d *types.RecordDraft) { d.Payload = `{oops` },
		"too large":      func(d *types.RecordDraft) { d.Payload = `"` + strings.Repeat("x", recordDraftMaxPayload) + `"` },
	} {
		d := ok()
		mutate(d)
		require.Error(t, validateRecordDraft(d), name)
	}
}

func TestRecordDrafts_store(t *testing.T) {
	var (
		req = require.New(t)
		ctx = context.Background()
	)

	s, err := sqlite.ConnectInMemory(ctx)
	req.NoError(err)
	req.NoError(store.Upgrade(ctx, zap.NewNop(), s))
	req.NoError(s.TruncateComposeRecordDrafts(ctx))

	svc := recordDraft{store: s}
	alice := auth.SetIdentityToContext(ctx, auth.Authenticated(5))
	bob := auth.SetIdentityToContext(ctx, auth.Authenticated(6))

	draft := func(change string, record uint64) *types.RecordDraft {
		return &types.RecordDraft{ChangeID: change, NamespaceID: 1, ModuleID: 2, RecordID: record, Payload: `{"v":"` + change + `"}`}
	}

	t.Run("the same change id replaces the draft", func(t *testing.T) {
		first, err := saveRecordDraft(alice, s, 5, draft("c1", 0))
		req.NoError(err)
		second, err := saveRecordDraft(alice, s, 5, &types.RecordDraft{ChangeID: "c1", NamespaceID: 1, ModuleID: 2, RecordID: 9, Payload: `{"v":"new"}`})
		req.NoError(err)
		req.Equal(first.ID, second.ID)

		set, err := svc.List(alice, types.RecordDraftFilter{})
		req.NoError(err)
		req.Len(set, 1)
		req.Equal(uint64(9), set[0].RecordID)
		req.JSONEq(`{"v":"new"}`, set[0].Payload)
	})

	t.Run("drafts are private", func(t *testing.T) {
		_, err := saveRecordDraft(bob, s, 6, draft("c1", 0)) // same change id, other author
		req.NoError(err)

		a, err := svc.List(alice, types.RecordDraftFilter{})
		req.NoError(err)
		b, err := svc.List(bob, types.RecordDraftFilter{})
		req.NoError(err)
		req.Len(a, 1)
		req.Len(b, 1)
		req.NotEqual(a[0].ID, b[0].ID)

		// bob cannot delete alice's draft by its change id
		req.NoError(svc.Delete(bob, "c1"))
		a, _ = svc.List(alice, types.RecordDraftFilter{})
		req.Len(a, 1)
		b, _ = svc.List(bob, types.RecordDraftFilter{})
		req.Empty(b)
	})

	t.Run("no user, no drafts", func(t *testing.T) {
		_, err := svc.List(ctx, types.RecordDraftFilter{})
		req.Error(err)
		req.Error(svc.Clear(ctx))
	})

	t.Run("filter by record and clear", func(t *testing.T) {
		_, err := saveRecordDraft(alice, s, 5, draft("c2", 33))
		req.NoError(err)

		set, err := svc.List(alice, types.RecordDraftFilter{RecordID: 33})
		req.NoError(err)
		req.Len(set, 1)
		req.Equal("c2", set[0].ChangeID)

		req.NoError(svc.Clear(alice))
		set, _ = svc.List(alice, types.RecordDraftFilter{})
		req.Empty(set)
	})

	t.Run("the oldest drafts go over the cap", func(t *testing.T) {
		for i := 0; i < recordDraftsKept+3; i++ {
			_, err := saveRecordDraft(alice, s, 5, draft("d"+string(rune('a'+i%26))+string(rune('a'+i/26)), 0))
			req.NoError(err)
		}
		set, err := svc.List(alice, types.RecordDraftFilter{})
		req.NoError(err)
		req.Len(set, recordDraftsKept)
	})
}
