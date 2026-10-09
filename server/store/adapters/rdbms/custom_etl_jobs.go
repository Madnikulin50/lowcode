package rdbms

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	composeTypes "github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
)

const etlJobTable = "compose_etl_job"

type etlJobRow struct {
	ID          uint64         `db:"id"`
	ModuleID    uint64         `db:"module_id"`
	NamespaceID uint64         `db:"namespace_id"`
	Name        string         `db:"name"`
	Enabled     bool           `db:"enabled"`
	Schedule    string         `db:"schedule"`
	SourceJSON  sql.NullString `db:"source"`
	LastRunAt   sql.NullTime   `db:"last_run_at"`
	LastStatus  string         `db:"last_status"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   sql.NullTime   `db:"updated_at"`
	DeletedAt   sql.NullTime   `db:"deleted_at"`
}

func (r *etlJobRow) toETLJob() (*composeTypes.ETLJob, error) {
	j := &composeTypes.ETLJob{
		ID:          r.ID,
		ModuleID:    r.ModuleID,
		NamespaceID: r.NamespaceID,
		Name:        r.Name,
		Enabled:     r.Enabled,
		Schedule:    r.Schedule,
		LastStatus:  r.LastStatus,
		CreatedAt:   r.CreatedAt,
	}
	if r.SourceJSON.Valid {
		if err := json.Unmarshal([]byte(r.SourceJSON.String), &j.Source); err != nil {
			return nil, err
		}
	}
	if r.LastRunAt.Valid {
		j.LastRunAt = &r.LastRunAt.Time
	}
	if r.UpdatedAt.Valid {
		j.UpdatedAt = &r.UpdatedAt.Time
	}
	if r.DeletedAt.Valid {
		j.DeletedAt = &r.DeletedAt.Time
	}
	return j, nil
}

func jobToRow(j *composeTypes.ETLJob) (*etlJobRow, error) {
	srcJSON, err := json.Marshal(j.Source)
	if err != nil {
		return nil, err
	}
	r := &etlJobRow{
		ID:          j.ID,
		ModuleID:    j.ModuleID,
		NamespaceID: j.NamespaceID,
		Name:        j.Name,
		Enabled:     j.Enabled,
		Schedule:    j.Schedule,
		SourceJSON:  sql.NullString{String: string(srcJSON), Valid: true},
		LastStatus:  j.LastStatus,
		CreatedAt:   j.CreatedAt,
	}
	if j.LastRunAt != nil {
		r.LastRunAt = sql.NullTime{Time: *j.LastRunAt, Valid: true}
	}
	if j.UpdatedAt != nil {
		r.UpdatedAt = sql.NullTime{Time: *j.UpdatedAt, Valid: true}
	}
	if j.DeletedAt != nil {
		r.DeletedAt = sql.NullTime{Time: *j.DeletedAt, Valid: true}
	}
	return r, nil
}

func SearchETLJobs(ctx context.Context, s *Store, f composeTypes.ETLJobFilter) (composeTypes.ETLJobSet, composeTypes.ETLJobFilter, error) {
	ex := s.Dialect.GOQU().From(etlJobTable).Select(
		"id", "module_id", "namespace_id", "name", "enabled",
		"schedule", "source", "last_run_at", "last_status",
		"created_at", "updated_at", "deleted_at",
	)
	ex = ex.Where(goqu.C("deleted_at").IsNull())

	if f.NamespaceID > 0 {
		ex = ex.Where(goqu.C("namespace_id").Eq(f.NamespaceID))
	}
	if f.ModuleID > 0 {
		ex = ex.Where(goqu.C("module_id").Eq(f.ModuleID))
	}

	if f.Paging.Limit == 0 {
		f.Paging.Limit = 20
	}
	ex = ex.Limit(uint(f.Paging.Limit))

	sqlStr, args, err := ex.ToSQL()
	if err != nil {
		return nil, f, err
	}

	rows, err := s.DB.QueryxContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, f, err
	}
	defer rows.Close()

	set := make(composeTypes.ETLJobSet, 0)
	for rows.Next() {
		var r etlJobRow
		if err := rows.StructScan(&r); err != nil {
			return nil, f, err
		}
		j, err := r.toETLJob()
		if err != nil {
			return nil, f, err
		}
		set = append(set, j)
	}

	return set, f, rows.Err()
}

func LookupETLJobByID(ctx context.Context, s *Store, id uint64) (*composeTypes.ETLJob, error) {
	var r etlJobRow

	sqlStr, args, err := s.Dialect.GOQU().From(etlJobTable).Select(
		"id", "module_id", "namespace_id", "name", "enabled",
		"schedule", "source", "last_run_at", "last_status",
		"created_at", "updated_at", "deleted_at",
	).Where(goqu.C("id").Eq(id)).Limit(1).ToSQL()
	if err != nil {
		return nil, err
	}

	if err := sqlx.GetContext(ctx, s.DB, &r, sqlStr, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("etl job not found")
		}
		return nil, err
	}

	return r.toETLJob()
}

func CreateETLJob(ctx context.Context, s *Store, j *composeTypes.ETLJob) error {
	srcJSON, err := json.Marshal(j.Source)
	if err != nil {
		return err
	}

	sqlStr, args, err := s.Dialect.GOQU().Insert(etlJobTable).Rows(
		goqu.Record{
			"id":           j.ID,
			"module_id":    j.ModuleID,
			"namespace_id": j.NamespaceID,
			"name":         j.Name,
			"enabled":      j.Enabled,
			"schedule":     j.Schedule,
			"source":       string(srcJSON),
			"last_run_at":  nil,
			"last_status":  j.LastStatus,
			"created_at":   j.CreatedAt,
			"updated_at":   nil,
			"deleted_at":   nil,
		},
	).ToSQL()
	if err != nil {
		return err
	}

	_, err = s.DB.ExecContext(ctx, sqlStr, args...)
	return err
}

func UpdateETLJob(ctx context.Context, s *Store, j *composeTypes.ETLJob) error {
	srcJSON, err := json.Marshal(j.Source)
	if err != nil {
		return err
	}

	rec := goqu.Record{
		"module_id":    j.ModuleID,
		"namespace_id": j.NamespaceID,
		"name":         j.Name,
		"enabled":      j.Enabled,
		"schedule":     j.Schedule,
		"source":       string(srcJSON),
		"last_status":  j.LastStatus,
		"updated_at":   time.Now(),
	}

	if j.LastRunAt != nil {
		rec["last_run_at"] = *j.LastRunAt
	}

	sqlStr, args, err := s.Dialect.GOQU().Update(etlJobTable).Set(rec).Where(goqu.C("id").Eq(j.ID)).ToSQL()
	if err != nil {
		return err
	}

	_, err = s.DB.ExecContext(ctx, sqlStr, args...)
	return err
}

func DeleteETLJob(ctx context.Context, s *Store, id uint64) error {
	now := time.Now()
	sqlStr, args, err := s.Dialect.GOQU().Update(etlJobTable).Set(
		goqu.Record{"deleted_at": now},
	).Where(goqu.C("id").Eq(id)).ToSQL()
	if err != nil {
		return err
	}

	_, err = s.DB.ExecContext(ctx, sqlStr, args...)
	return err
}
