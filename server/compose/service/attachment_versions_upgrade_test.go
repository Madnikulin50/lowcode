package service

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/drivers/sqlite"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// An installation that already has attachments gets the version columns from
// the upgrade fix, and its old rows read as "not versioned".
func TestAttachmentVersions_upgradeExistingTable(t *testing.T) {
	var (
		req = require.New(t)
		ctx = context.Background()
		db  = filepath.Join(t.TempDir(), "old.db")
	)

	// the table as it was before versioning, with a row in it
	raw, err := sql.Open("sqlite3", db)
	req.NoError(err)
	_, err = raw.Exec(`CREATE TABLE compose_attachment (
		id INTEGER NOT NULL PRIMARY KEY,
		rel_owner INTEGER NOT NULL DEFAULT 0,
		kind TEXT NOT NULL DEFAULT '',
		url TEXT NOT NULL DEFAULT '',
		preview_url TEXT NOT NULL DEFAULT '',
		name TEXT NOT NULL DEFAULT '',
		meta TEXT NOT NULL DEFAULT '{}',
		rel_namespace INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NULL,
		deleted_at TIMESTAMP NULL)`)
	req.NoError(err)
	_, err = raw.Exec(`INSERT INTO compose_attachment (id, name) VALUES (7, 'old')`)
	req.NoError(err)
	req.NoError(raw.Close())

	s, err := sqlite.Connect(ctx, "sqlite3://file:"+db)
	req.NoError(err)
	req.NoError(store.Upgrade(ctx, zap.NewNop(), s))

	old, err := store.LookupComposeAttachmentByID(ctx, s, 7)
	req.NoError(err)
	req.Equal("old", old.Name)
	req.Zero(old.VersionGroup)
	req.Zero(old.Version)

	old.VersionGroup, old.Version = 7, 1
	req.NoError(store.UpdateComposeAttachment(ctx, s, old))
	got, err := store.LookupComposeAttachmentByID(ctx, s, 7)
	req.NoError(err)
	req.Equal(uint64(7), got.VersionGroup)
	req.Equal(1, got.Version)

	set, _, err := store.SearchComposeAttachments(ctx, s, types.AttachmentFilter{VersionGroup: 7})
	req.NoError(err)
	req.Len(set, 1)
}
