package rdbms

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	composeTypes "github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
)

const ruleChainTable = "compose_rule_chain"

type ruleChainRow struct {
	ID          uint64       `db:"id"`
	Handle      string       `db:"handle"`
	NamespaceID uint64       `db:"rel_namespace"`
	Name        string       `db:"name"`
	Description string       `db:"description"`
	EntryNode   string       `db:"entry_node"`
	NodesJSON   []byte       `db:"nodes"`
	EdgesJSON   []byte       `db:"edges"`
	ConfigJSON  []byte       `db:"config"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   sql.NullTime `db:"updated_at"`
	DeletedAt   sql.NullTime `db:"deleted_at"`
}

func (r *ruleChainRow) toRuleChain() *composeTypes.RuleChain {
	rc := &composeTypes.RuleChain{
		ID:          r.ID,
		Handle:      r.Handle,
		NamespaceID: r.NamespaceID,
		Name:        r.Name,
		Description: r.Description,
		EntryNode:   r.EntryNode,
		CreatedAt:   r.CreatedAt,
		Nodes:       jsonbOr(r.NodesJSON, "[]"),
		Edges:       jsonbOr(r.EdgesJSON, "[]"),
		Config:      jsonbOr(r.ConfigJSON, "{}"),
	}
	if r.UpdatedAt.Valid {
		rc.UpdatedAt = &r.UpdatedAt.Time
	}
	if r.DeletedAt.Valid {
		rc.DeletedAt = &r.DeletedAt.Time
	}
	return rc
}

func jsonbOr(raw []byte, fallback string) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage(fallback)
	}
	return json.RawMessage(raw)
}

func jsonbArg(raw json.RawMessage, fallback string) string {
	s := strings.TrimSpace(string(raw))
	if s == "" || s == "null" {
		return fallback
	}
	if !json.Valid([]byte(s)) {
		return fallback
	}
	return s
}

func ruleChainSelect(s *Store) *goqu.SelectDataset {
	return s.Dialect.GOQU().From(ruleChainTable).Select(
		"id", "handle", "rel_namespace", "name", "description", "entry_node",
		"nodes", "edges", "config", "created_at", "updated_at", "deleted_at",
	)
}

func SearchRuleChains(ctx context.Context, s *Store, f composeTypes.RuleChainFilter) (composeTypes.RuleChainSet, composeTypes.RuleChainFilter, error) {
	ex := ruleChainSelect(s)

	switch f.Deleted {
	case filter.StateExclusive:
		ex = ex.Where(goqu.C("deleted_at").IsNotNull())
	case filter.StateInclusive:
		// include deleted and live rows
	default:
		ex = ex.Where(goqu.C("deleted_at").IsNull())
	}

	if f.NamespaceID > 0 {
		ex = ex.Where(goqu.C("rel_namespace").Eq(f.NamespaceID))
	}
	if h := strings.TrimSpace(f.Handle); h != "" {
		ex = ex.Where(goqu.Func("LOWER", goqu.C("handle")).Eq(strings.ToLower(h)))
	}
	if q := strings.TrimSpace(f.Query); q != "" {
		like := "%" + strings.ToLower(q) + "%"
		ex = ex.Where(goqu.Or(
			goqu.Func("LOWER", goqu.C("name")).Like(like),
			goqu.Func("LOWER", goqu.C("handle")).Like(like),
			goqu.Func("LOWER", goqu.C("description")).Like(like),
		))
	}

	ex = ex.Order(goqu.C("name").Asc())
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

	set := make(composeTypes.RuleChainSet, 0)
	for rows.Next() {
		var r ruleChainRow
		if err := rows.StructScan(&r); err != nil {
			return nil, f, err
		}
		set = append(set, r.toRuleChain())
	}
	return set, f, rows.Err()
}

func LookupRuleChainByHandle(ctx context.Context, s *Store, handle string) (*composeTypes.RuleChain, error) {
	var r ruleChainRow
	sqlStr, args, err := ruleChainSelect(s).Where(
		goqu.Func("LOWER", goqu.C("handle")).Eq(strings.ToLower(strings.TrimSpace(handle))),
		goqu.C("deleted_at").IsNull(),
	).Limit(1).ToSQL()
	if err != nil {
		return nil, err
	}
	if err := sqlx.GetContext(ctx, s.DB, &r, sqlStr, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("rule chain not found")
		}
		return nil, err
	}
	return r.toRuleChain(), nil
}

func LookupRuleChainByID(ctx context.Context, s *Store, id uint64) (*composeTypes.RuleChain, error) {
	var r ruleChainRow
	sqlStr, args, err := ruleChainSelect(s).Where(goqu.C("id").Eq(id)).Limit(1).ToSQL()
	if err != nil {
		return nil, err
	}
	if err := sqlx.GetContext(ctx, s.DB, &r, sqlStr, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("rule chain not found")
		}
		return nil, err
	}
	return r.toRuleChain(), nil
}

func CreateRuleChain(ctx context.Context, s *Store, rc *composeTypes.RuleChain) error {
	sqlStr, args, err := s.Dialect.GOQU().Insert(ruleChainTable).Rows(
		goqu.Record{
			"id":            rc.ID,
			"handle":        rc.Handle,
			"rel_namespace": rc.NamespaceID,
			"name":          rc.Name,
			"description":   rc.Description,
			"entry_node":    rc.EntryNode,
			"nodes":         jsonbArg(rc.Nodes, "[]"),
			"edges":         jsonbArg(rc.Edges, "[]"),
			"config":        jsonbArg(rc.Config, "{}"),
			"created_at":    rc.CreatedAt,
			"updated_at":    nil,
			"deleted_at":    nil,
		},
	).ToSQL()
	if err != nil {
		return err
	}
	_, err = s.DB.ExecContext(ctx, sqlStr, args...)
	return err
}

func UpdateRuleChain(ctx context.Context, s *Store, rc *composeTypes.RuleChain) error {
	now := time.Now().UTC()
	sqlStr, args, err := s.Dialect.GOQU().Update(ruleChainTable).Set(goqu.Record{
		"handle":        rc.Handle,
		"rel_namespace": rc.NamespaceID,
		"name":          rc.Name,
		"description":   rc.Description,
		"entry_node":    rc.EntryNode,
		"nodes":         jsonbArg(rc.Nodes, "[]"),
		"edges":         jsonbArg(rc.Edges, "[]"),
		"config":        jsonbArg(rc.Config, "{}"),
		"updated_at":    now,
		"deleted_at":    nil,
	}).Where(goqu.C("id").Eq(rc.ID)).ToSQL()
	if err != nil {
		return err
	}
	_, err = s.DB.ExecContext(ctx, sqlStr, args...)
	return err
}

func DeleteRuleChainByHandle(ctx context.Context, s *Store, handle string) error {
	now := time.Now().UTC()
	sqlStr, args, err := s.Dialect.GOQU().Update(ruleChainTable).Set(
		goqu.Record{"deleted_at": now},
	).Where(
		goqu.Func("LOWER", goqu.C("handle")).Eq(strings.ToLower(strings.TrimSpace(handle))),
		goqu.C("deleted_at").IsNull(),
	).ToSQL()
	if err != nil {
		return err
	}
	_, err = s.DB.ExecContext(ctx, sqlStr, args...)
	return err
}
