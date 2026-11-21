package service

import (
	"context"
	"io"
	"testing"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
	"github.com/madnikulin50/lowcode/server/pkg/objstore"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/drivers/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// memObjects is an in-memory objstore.Store that remembers what was removed
type memObjects struct {
	removed []string
}

func (*memObjects) Original(id uint64, ext string) string  { return "o" + ext }
func (*memObjects) Preview(id uint64, ext string) string   { return "p" + ext }
func (*memObjects) Save(string, io.Reader) error           { return nil }
func (m *memObjects) Remove(name string) error             { m.removed = append(m.removed, name); return nil }
func (*memObjects) Open(string) (io.ReadSeekCloser, error) { return nil, errors.NotFound("n/a") }
func (*memObjects) Healthcheck(context.Context) error      { return nil }

func TestAttachmentVersions(t *testing.T) {
	var (
		req = require.New(t)
		ctx = context.Background()
	)

	s, err := sqlite.ConnectInMemory(ctx)
	req.NoError(err)
	req.NoError(store.Upgrade(ctx, zap.NewNop(), s))
	req.NoError(s.TruncateComposeAttachments(ctx))

	objects := &memObjects{}
	svc := attachment{objects: map[string]objstore.Store{"mem": objects}, legacyDriver: "mem", store: s}

	newFile := func(name string) *types.Attachment {
		a := &types.Attachment{ID: nextID(), Name: name, Kind: types.RecordAttachment, CreatedAt: *now(),
			Url: "u-" + name, PreviewUrl: "pv-" + name, Meta: types.AttachmentMeta{StorageDriver: "mem"}}
		req.NoError(store.CreateComposeAttachment(ctx, s, a))
		return a
	}

	module := func(opts types.ModuleFieldOptions) *types.Module {
		return &types.Module{Fields: types.ModuleFieldSet{{Name: "doc", Kind: "File", Options: opts}}}
	}

	record := func(refs ...uint64) *types.Record {
		r := &types.Record{}
		for i, ref := range refs {
			r.Values = append(r.Values, &types.RecordValue{Name: "doc", Ref: ref, Place: uint(i)})
		}
		return r
	}

	t.Run("refused when versioning is off", func(t *testing.T) {
		a := newFile("a")
		_, err := svc.prepareVersion(ctx, s, module(nil), record(a.ID), "doc", a.ID, &types.Attachment{})
		req.Error(err)
	})

	t.Run("refused for a file that is not a value of the field", func(t *testing.T) {
		a, other := newFile("a"), newFile("other")
		_, err := svc.prepareVersion(ctx, s, module(types.ModuleFieldOptions{"versioning": true}), record(a.ID), "doc", other.ID, &types.Attachment{})
		req.Error(err)
	})

	t.Run("the first version starts the chain", func(t *testing.T) {
		a := newFile("a")
		next := &types.Attachment{}
		dropped, err := svc.prepareVersion(ctx, s, module(types.ModuleFieldOptions{"versioning": true}), record(a.ID), "doc", a.ID, next)
		req.NoError(err)
		req.Empty(dropped)
		req.Equal(a.ID, next.VersionGroup)
		req.Equal(2, next.Version)

		got, err := store.LookupComposeAttachmentByID(ctx, s, a.ID)
		req.NoError(err)
		req.Equal(a.ID, got.VersionGroup)
		req.Equal(1, got.Version)
	})

	t.Run("limit drops the oldest, never the current one, and removes bytes", func(t *testing.T) {
		mod := module(types.ModuleFieldOptions{"versioning": true, "maxVersions": 3})

		v1 := newFile("v1")
		v2 := &types.Attachment{}
		_, err := svc.prepareVersion(ctx, s, mod, record(v1.ID), "doc", v1.ID, v2)
		req.NoError(err)
		v2.ID, v2.Name, v2.Kind, v2.CreatedAt, v2.Url, v2.PreviewUrl = nextID(), "v2", types.RecordAttachment, *now(), "u-v2", "pv-v2"
		v2.Meta.StorageDriver = "mem"
		req.NoError(store.CreateComposeAttachment(ctx, s, v2))

		v3 := &types.Attachment{}
		_, err = svc.prepareVersion(ctx, s, mod, record(v2.ID), "doc", v2.ID, v3)
		req.NoError(err)
		v3.ID, v3.Name, v3.Kind, v3.CreatedAt, v3.Url, v3.PreviewUrl = nextID(), "v3", types.RecordAttachment, *now(), "u-v3", "pv-v3"
		v3.Meta.StorageDriver = "mem"
		req.NoError(store.CreateComposeAttachment(ctx, s, v3))
		req.Equal(3, v3.Version)

		// a 4th version exceeds the limit of 3: v1 goes (the record points to v3)
		v4 := &types.Attachment{}
		dropped, err := svc.prepareVersion(ctx, s, mod, record(v3.ID), "doc", v3.ID, v4)
		req.NoError(err)
		req.Equal(4, v4.Version)
		req.Len(dropped, 1)
		req.Equal(v1.ID, dropped[0].ID)
		v4.ID, v4.Name, v4.Kind, v4.CreatedAt, v4.Url, v4.PreviewUrl = nextID(), "v4", types.RecordAttachment, *now(), "u-v4", "pv-v4"
		v4.Meta.StorageDriver = "mem"
		req.NoError(store.CreateComposeAttachment(ctx, s, v4))

		_, err = store.LookupComposeAttachmentByID(ctx, s, v1.ID)
		req.True(errors.IsNotFound(err))

		svc.removeFiles(dropped)
		req.ElementsMatch([]string{"u-v1", "pv-v1"}, objects.removed)

		// the record points to the oldest version: it is kept, the next one goes
		v5 := &types.Attachment{}
		dropped, err = svc.prepareVersion(ctx, s, mod, record(v2.ID), "doc", v2.ID, v5)
		req.NoError(err)
		req.Len(dropped, 1)
		req.Equal(v3.ID, dropped[0].ID, "v2 is current so v3 is the oldest one that may go")
	})
}
