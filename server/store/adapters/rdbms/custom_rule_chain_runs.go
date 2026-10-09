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
)

const ruleChainRunTable = "compose_rule_chain_run"

type ruleChainRunRow struct {
	ID          uint64         `db:"id"`
	NamespaceID uint64         `db:"rel_namespace"`
	ChainID     string         `db:"chain_id"`
	TriggerType string         `db:"trigger_type"`
	Success     bool           `db:"success"`
	Error       sql.NullString `db:"error"`
	InputJSON   []byte         `db:"input"`
	OutputJSON  []byte         `db:"output"`
	NodesJSON   []byte         `db:"nodes"`
	StartedAt   time.Time      `db:"started_at"`
	FinishedAt  sql.NullTime   `db:"finished_at"`
	DurationMs  int64          `db:"duration_ms"`
	CreatedAt   time.Time      `db:"created_at"`
}

func (r *ruleChainRunRow) toRuleChainRun() *composeTypes.RuleChainRun {
	run := &composeTypes.RuleChainRun{
		ID:          r.ID,
		NamespaceID: r.NamespaceID,
		ChainID:     r.ChainID,
		TriggerType: r.TriggerType,
		Success:     r.Success,
		DurationMs:  r.DurationMs,
		StartedAt:   r.StartedAt,
		CreatedAt:   r.CreatedAt,
		Input:       jsonbOr(r.InputJSON, "{}"),
		Output:      jsonbOr(r.OutputJSON, "{}"),
		Nodes:       jsonbOr(r.NodesJSON, "[]"),
	}
	if r.Error.Valid {
		run.Error = r.Error.String
	}
	if r.FinishedAt.Valid {
		run.FinishedAt = &r.FinishedAt.Time
	}
	return run
}

func ruleChainRunSelect(s *Store) *goqu.SelectDataset {
	return s.Dialect.GOQU().From(ruleChainRunTable).Select(
		"id", "rel_namespace", "chain_id", "trigger_type", "success", "error",
		"input", "output", "nodes", "started_at", "finished_at", "duration_ms", "created_at",
	)
}

// InsertRuleChainRun stores one rule chain execution record (run-log entry).
func InsertRuleChainRun(ctx context.Context, s *Store, run *composeTypes.RuleChainRun) error {
	var finishedAt any
	if run.FinishedAt != nil {
		finishedAt = *run.FinishedAt
	}

	sqlStr, args, err := s.Dialect.GOQU().Insert(ruleChainRunTable).Rows(
		goqu.Record{
			"id":            run.ID,
			"rel_namespace": run.NamespaceID,
			"chain_id":      run.ChainID,
			"trigger_type":  run.TriggerType,
			"success":       run.Success,
			"error":         nullableString(run.Error),
			"input":         jsonbArg(run.Input, "{}"),
			"output":        jsonbArg(run.Output, "{}"),
			"nodes":         jsonbArg(run.Nodes, "[]"),
			"started_at":    run.StartedAt,
			"finished_at":   finishedAt,
			"duration_ms":   run.DurationMs,
			"created_at":    run.CreatedAt,
		},
	).ToSQL()
	if err != nil {
		return err
	}
	_, err = s.DB.ExecContext(ctx, sqlStr, args...)
	return err
}

// SearchRuleChainRuns returns persisted run-log entries, newest first.
func SearchRuleChainRuns(ctx context.Context, s *Store, f composeTypes.RuleChainRunFilter) (composeTypes.RuleChainRunSet, composeTypes.RuleChainRunFilter, error) {
	ex := ruleChainRunSelect(s)

	if h := strings.TrimSpace(f.ChainID); h != "" {
		ex = ex.Where(goqu.C("chain_id").Eq(h))
	}
	if f.NamespaceID > 0 {
		ex = ex.Where(goqu.C("rel_namespace").Eq(f.NamespaceID))
	}
	if tt := strings.TrimSpace(f.TriggerType); tt != "" {
		ex = ex.Where(goqu.C("trigger_type").Eq(tt))
	}
	if f.Success != nil {
		ex = ex.Where(goqu.C("success").Eq(*f.Success))
	}

	ex = ex.Order(goqu.C("started_at").Desc())

	limit := f.Limit
	if limit <= 0 || limit > 1000 {
		limit = 50
	}
	ex = ex.Limit(uint(limit))
	if f.Offset > 0 {
		ex = ex.Offset(uint(f.Offset))
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

	set := make(composeTypes.RuleChainRunSet, 0)
	for rows.Next() {
		var r ruleChainRunRow
		if err := rows.StructScan(&r); err != nil {
			return nil, f, err
		}
		set = append(set, r.toRuleChainRun())
	}
	return set, f, rows.Err()
}

// LookupRuleChainRunByID returns a single run-log entry with its full node trace.
func LookupRuleChainRunByID(ctx context.Context, s *Store, id uint64) (*composeTypes.RuleChainRun, error) {
	var r ruleChainRunRow
	sqlStr, args, err := ruleChainRunSelect(s).Where(goqu.C("id").Eq(id)).Limit(1).ToSQL()
	if err != nil {
		return nil, err
	}
	if err := sqlx.GetContext(ctx, s.DB, &r, sqlStr, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("rule chain run not found")
		}
		return nil, err
	}
	return r.toRuleChainRun(), nil
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// PruneRuleChainRuns deletes run-log entries older than `before`, keeping the
// table bounded. Intended to be called periodically (e.g. from a cron job).
func PruneRuleChainRuns(ctx context.Context, s *Store, before time.Time) (int64, error) {
	sqlStr, args, err := s.Dialect.GOQU().Delete(ruleChainRunTable).Where(
		goqu.C("started_at").Lt(before),
	).ToSQL()
	if err != nil {
		return 0, err
	}
	res, err := s.DB.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
