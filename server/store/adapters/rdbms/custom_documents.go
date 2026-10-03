package rdbms

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	composeTypes "github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
)

const documentTable = "compose_document"

type documentRow struct {
	ID           uint64        `db:"id"`
	NamespaceID  uint64        `db:"rel_namespace"`
	Handle       string        `db:"handle"`
	Title        string        `db:"title"`
	Kind         string        `db:"kind"`
	Body         string        `db:"body"`
	AttachmentID sql.NullInt64 `db:"attachment_id"`
	FileName     string        `db:"file_name"`
	Weight       int           `db:"weight"`
	Visible      bool          `db:"visible"`
	CreatedAt    time.Time     `db:"created_at"`
	UpdatedAt    sql.NullTime  `db:"updated_at"`
	DeletedAt    sql.NullTime  `db:"deleted_at"`
}

func (r *documentRow) toDocument() *composeTypes.Document {
	doc := &composeTypes.Document{
		ID:           r.ID,
		NamespaceID:  r.NamespaceID,
		Handle:       r.Handle,
		Title:        r.Title,
		Kind:         r.Kind,
		Body:         r.Body,
		AttachmentID: uint64(r.AttachmentID.Int64),
		FileName:     r.FileName,
		Weight:       r.Weight,
		Visible:      r.Visible,
		CreatedAt:    r.CreatedAt,
	}
	if r.UpdatedAt.Valid {
		doc.UpdatedAt = &r.UpdatedAt.Time
	}
	if r.DeletedAt.Valid {
		doc.DeletedAt = &r.DeletedAt.Time
	}
	return doc
}

func documentSelect(s *Store) *goqu.SelectDataset {
	return s.Dialect.GOQU().From(documentTable).Select(
		"id", "rel_namespace", "handle", "title", "kind", "body",
		"attachment_id", "file_name", "weight", "visible",
		"created_at", "updated_at", "deleted_at",
	)
}

func SearchDocuments(ctx context.Context, s *Store, f composeTypes.DocumentFilter) (composeTypes.DocumentSet, composeTypes.DocumentFilter, error) {
	ex := documentSelect(s)

	switch f.Deleted {
	case filter.StateExclusive:
		ex = ex.Where(goqu.C("deleted_at").IsNotNull())
	case filter.StateInclusive:
	default:
		ex = ex.Where(goqu.C("deleted_at").IsNull())
	}

	if f.NamespaceID > 0 {
		ex = ex.Where(goqu.C("rel_namespace").Eq(f.NamespaceID))
	}
	if f.VisibleOnly {
		ex = ex.Where(goqu.C("visible").Eq(true))
	}
	if h := strings.TrimSpace(f.Handle); h != "" {
		ex = ex.Where(goqu.Func("LOWER", goqu.C("handle")).Eq(strings.ToLower(h)))
	}
	if q := strings.TrimSpace(f.Query); q != "" {
		like := "%" + strings.ToLower(q) + "%"
		ex = ex.Where(goqu.Or(
			goqu.Func("LOWER", goqu.C("title")).Like(like),
			goqu.Func("LOWER", goqu.C("handle")).Like(like),
		))
	}

	ex = ex.Order(goqu.C("weight").Asc(), goqu.C("title").Asc())
	if f.Paging.Limit > 0 {
		ex = ex.Limit(uint(f.Paging.Limit))
	}

	sqlStr, args, err := ex.ToSQL()
	if err != nil {
		return nil, f, err
	}

	rows, err := s.DB.QueryxContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, f, err
	}
	defer rows.Close()

	set := make(composeTypes.DocumentSet, 0)
	for rows.Next() {
		var r documentRow
		if err := rows.StructScan(&r); err != nil {
			return nil, f, err
		}
		set = append(set, r.toDocument())
	}
	return set, f, rows.Err()
}

func LookupDocumentByID(ctx context.Context, s *Store, id uint64) (*composeTypes.Document, error) {
	var r documentRow
	sqlStr, args, err := documentSelect(s).Where(
		goqu.C("id").Eq(id),
		goqu.C("deleted_at").IsNull(),
	).Limit(1).ToSQL()
	if err != nil {
		return nil, err
	}
	if err := sqlx.GetContext(ctx, s.DB, &r, sqlStr, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("document not found")
		}
		return nil, err
	}
	return r.toDocument(), nil
}

func LookupDocumentByHandle(ctx context.Context, s *Store, namespaceID uint64, handle string) (*composeTypes.Document, error) {
	var r documentRow
	sqlStr, args, err := documentSelect(s).Where(
		goqu.C("rel_namespace").Eq(namespaceID),
		goqu.Func("LOWER", goqu.C("handle")).Eq(strings.ToLower(strings.TrimSpace(handle))),
		goqu.C("deleted_at").IsNull(),
	).Limit(1).ToSQL()
	if err != nil {
		return nil, err
	}
	if err := sqlx.GetContext(ctx, s.DB, &r, sqlStr, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("document not found")
		}
		return nil, err
	}
	return r.toDocument(), nil
}

func CreateDocument(ctx context.Context, s *Store, doc *composeTypes.Document) error {
	sqlStr, args, err := s.Dialect.GOQU().Insert(documentTable).Rows(goqu.Record{
		"id":            doc.ID,
		"rel_namespace": doc.NamespaceID,
		"handle":        doc.Handle,
		"title":         doc.Title,
		"kind":          doc.Kind,
		"body":          doc.Body,
		"attachment_id": doc.AttachmentID,
		"file_name":     doc.FileName,
		"weight":        doc.Weight,
		"visible":       doc.Visible,
		"created_at":    doc.CreatedAt,
		"updated_at":    nil,
		"deleted_at":    nil,
	}).ToSQL()
	if err != nil {
		return err
	}
	_, err = s.DB.ExecContext(ctx, sqlStr, args...)
	return err
}

func UpdateDocument(ctx context.Context, s *Store, doc *composeTypes.Document) error {
	now := time.Now().UTC()
	sqlStr, args, err := s.Dialect.GOQU().Update(documentTable).Set(goqu.Record{
		"handle":        doc.Handle,
		"title":         doc.Title,
		"kind":          doc.Kind,
		"body":          doc.Body,
		"attachment_id": doc.AttachmentID,
		"file_name":     doc.FileName,
		"weight":        doc.Weight,
		"visible":       doc.Visible,
		"updated_at":    now,
	}).Where(goqu.C("id").Eq(doc.ID)).ToSQL()
	if err != nil {
		return err
	}
	_, err = s.DB.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return err
	}
	doc.UpdatedAt = &now
	return nil
}

func DeleteDocumentByID(ctx context.Context, s *Store, id uint64) error {
	now := time.Now().UTC()
	sqlStr, args, err := s.Dialect.GOQU().Update(documentTable).Set(
		goqu.Record{"deleted_at": now},
	).Where(
		goqu.C("id").Eq(id),
		goqu.C("deleted_at").IsNull(),
	).ToSQL()
	if err != nil {
		return err
	}
	_, err = s.DB.ExecContext(ctx, sqlStr, args...)
	return err
}
