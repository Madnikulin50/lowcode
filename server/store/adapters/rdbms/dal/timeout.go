package dal

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
)

const (
	// pageFetchTimeout caps each list/report SELECT — including every
	// aggregate (chart/report) query, which goes through the same iterator
	// page-fetch path as a plain list. Without a deadline the HTTP handler
	// waits for Postgres forever (lib/pq cancel is not always enough;
	// SET LOCAL statement_timeout aborts on the server).
	//
	// Raised 8s -> 40s (5x): multi-step datasource pipelines (load ->
	// aggregate -> join -> ...) issue several of these page-fetch queries
	// per request, one fresh pageFetchTimeout window each, and a handful of
	// them landing back-to-back on a loaded DB could add up past the old
	// 8s well before the query itself was actually slow.
	pageFetchTimeout = 90 * time.Second
	// countQueryTimeout is the hard cap for COUNT(*) when the caller did not
	// set a shorter deadline.
	countQueryTimeout = 30 * time.Second
)

// boundQueryContext returns ctx with a deadline of at most timeout.
// An existing sooner parent deadline is kept.
func boundQueryContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = pageFetchTimeout
	}
	if dl, ok := ctx.Deadline(); ok {
		remain := time.Until(dl)
		if remain <= timeout {
			return ctx, func() {}
		}
	}
	return context.WithTimeout(ctx, timeout)
}

func isQueryTimeout(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "57014", // query_canceled (statement_timeout / ctx cancel)
			"57000": // operator_intervention
			return true
		}
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "deadline exceeded") ||
		strings.Contains(s, "statement timeout") ||
		strings.Contains(s, "query canceled") ||
		strings.Contains(s, "canceling statement")
}

func statementTimeoutMS(ctx context.Context, fallback time.Duration) int {
	d := fallback
	if dl, ok := ctx.Deadline(); ok {
		remain := time.Until(dl)
		if remain > 0 && remain < d {
			d = remain
		}
	}
	ms := int(d / time.Millisecond)
	if ms < 1 {
		ms = 1
	}
	return ms
}

// queryRowCount runs COUNT with QueryContext deadline plus Postgres
// statement_timeout on the same transaction so the HTTP request cannot wait
// on an unbounded seq-scan if the driver ignores context cancel.
func queryRowCount(ctx context.Context, conn queryRunner, query string, args ...any) (uint, error) {
	var aux struct {
		Count uint `db:"count"`
	}
	scan := func(q sqlx.QueryerContext) error {
		return q.QueryRowxContext(ctx, query, args...).Scan(&aux.Count)
	}

	db, ok := conn.(*sqlx.DB)
	if !ok || !isPostgres(conn) {
		if err := scan(conn); err != nil {
			return 0, err
		}
		return aux.Count, nil
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	ms := statementTimeoutMS(ctx, countQueryTimeout)
	if _, err = tx.ExecContext(ctx, fmt.Sprintf("SET LOCAL statement_timeout = %d", ms)); err != nil {
		return 0, err
	}
	if err = scan(tx); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return aux.Count, nil
}

func isPostgres(conn queryRunner) bool {
	type driverNamer interface{ DriverName() string }
	namer, ok := conn.(driverNamer)
	if !ok {
		return false
	}
	name := strings.ToLower(namer.DriverName())
	return strings.Contains(name, "postgres") || name == "pq"
}

type txRollbacker interface {
	Rollback() error
}

// queryContextWithTimeout runs a SELECT with QueryContext deadline and, when
// conn is *sqlx.DB, SET LOCAL statement_timeout on a transaction that the
// caller must Rollback after rows.Close().
func queryContextWithTimeout(ctx context.Context, conn queryRunner, query string, args ...any) (rows *sql.Rows, tx txRollbacker, err error) {
	db, ok := conn.(*sqlx.DB)
	if !ok || !isPostgres(conn) {
		rows, err = conn.QueryContext(ctx, query, args...)
		return rows, nil, err
	}

	begun, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	ms := statementTimeoutMS(ctx, pageFetchTimeout)
	if _, err = begun.ExecContext(ctx, fmt.Sprintf("SET LOCAL statement_timeout = %d", ms)); err != nil {
		_ = begun.Rollback()
		return nil, nil, err
	}
	rows, err = begun.QueryContext(ctx, query, args...)
	if err != nil {
		_ = begun.Rollback()
		return nil, nil, err
	}
	return rows, begun, nil
}
