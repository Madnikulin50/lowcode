// Package dbblob is an objstore.Store backend that keeps file content in the
// database instead of on disk or in S3-compatible storage. It is selected
// per module File-field (or as the system-wide default) alongside "plain"
// and "minio" — see compose/service/attachment.go for the resolver.
package dbblob

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"path"

	_ "github.com/lib/pq"
)

const tableName = "sys_object_store_blob"

type store struct {
	db        *sql.DB
	namespace string

	originalFn func(id uint64, ext string) string
	previewFn  func(id uint64, ext string) string
}

var (
	defPreviewFn = func(id uint64, ext string) string {
		return fmt.Sprintf("%d_preview.%s", id, ext)
	}

	defOriginalFn = func(id uint64, ext string) string {
		return fmt.Sprintf("%d.%s", id, ext)
	}
)

// New opens (or reuses, via database/sql's own pooling) a connection to dsn
// and ensures the blob table exists. namespace is the same per-service path
// prefix plain/minio use ("compose", "system", ...) — kept only to validate
// filenames, not to partition the table (the filename already encodes it).
func New(dsn, namespace string) (*store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("dbblob: could not open connection: %w", err)
	}

	// Attachment I/O is incidental, not the app's primary workload — keep
	// this pool small so it can't starve the main store's connections.
	db.SetMaxOpenConns(4)

	if err := ensureTable(context.Background(), db); err != nil {
		return nil, err
	}

	return &store{
		db:        db,
		namespace: namespace,

		originalFn: defOriginalFn,
		previewFn:  defPreviewFn,
	}, nil
}

func ensureTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS `+tableName+` (
		filename   text PRIMARY KEY,
		content    bytea NOT NULL,
		size       bigint NOT NULL,
		created_at timestamptz NOT NULL DEFAULT now(),
		updated_at timestamptz
	)`)
	if err != nil {
		return fmt.Errorf("dbblob: could not ensure table: %w", err)
	}
	return nil
}

func (s *store) check(filename string) error {
	if len(filename) == 0 {
		return fmt.Errorf("invalid filename when trying to store file: '%s' (for %s)", filename, s.namespace)
	}

	if len(filename) <= len(s.namespace)+1 || filename[:len(s.namespace)+1] != s.namespace+"/" {
		return fmt.Errorf("invalid namespace when trying to store file: '%s' (for %s)", filename, s.namespace)
	}

	return nil
}

func (s *store) Original(id uint64, ext string) string {
	return path.Join(s.namespace, s.originalFn(id, ext))
}

func (s *store) Preview(id uint64, ext string) string {
	return path.Join(s.namespace, s.previewFn(id, ext))
}

func (s *store) Save(filename string, contents io.Reader) error {
	if err := s.check(filename); err != nil {
		return err
	}

	data, err := io.ReadAll(contents)
	if err != nil {
		return fmt.Errorf("dbblob: could not read content for %s: %w", filename, err)
	}

	_, err = s.db.Exec(`
		INSERT INTO `+tableName+` (filename, content, size, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (filename) DO UPDATE SET
			content = EXCLUDED.content,
			size = EXCLUDED.size,
			updated_at = now()
	`, filename, data, len(data))
	if err != nil {
		return fmt.Errorf("dbblob: could not save %s: %w", filename, err)
	}

	return nil
}

func (s *store) Remove(filename string) error {
	if err := s.check(filename); err != nil {
		return err
	}

	_, err := s.db.Exec(`DELETE FROM `+tableName+` WHERE filename = $1`, filename)
	if err != nil {
		return fmt.Errorf("dbblob: could not remove %s: %w", filename, err)
	}

	return nil
}

func (s *store) Open(filename string) (io.ReadSeekCloser, error) {
	if err := s.check(filename); err != nil {
		return nil, err
	}

	var data []byte
	err := s.db.QueryRow(`SELECT content FROM `+tableName+` WHERE filename = $1`, filename).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("dbblob: %s not found", filename)
	}
	if err != nil {
		return nil, fmt.Errorf("dbblob: could not open %s: %w", filename, err)
	}

	return readSeekCloser{bytes.NewReader(data)}, nil
}

func (s *store) Healthcheck(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// readSeekCloser adapts *bytes.Reader (Read+Seek) to io.ReadSeekCloser with a
// no-op Close — the content is already fully loaded in memory.
type readSeekCloser struct {
	*bytes.Reader
}

func (readSeekCloser) Close() error { return nil }
