package clickhouse

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/dal"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/madnikulin50/lowcode/server/pkg/logger"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms"
)

const (
	SCHEMA = "clickhouse"
)

func init() {
	store.Register(Connect, SCHEMA)
}

func Connect(ctx context.Context, dsn string) (_ store.Storer, err error) {
	var (
		db  *sqlx.DB
		cfg *rdbms.ConnConfig
	)

	if cfg, err = NewConfig(dsn); err != nil {
		return
	}

	if db, err = rdbms.Connect(ctx, logger.Default(), cfg); err != nil {
		return
	}

	s := &rdbms.Store{
		DB: db,

		DAL: dal.Connection(db, Dialect(), DataDefiner(cfg.DBName, db)),

		Dialect:      Dialect(),
		ErrorHandler: errorHandler,

		DataDefiner: DataDefiner(cfg.DBName, db),
		Ping:        db.PingContext,
	}

	s.SetDefaults()

	return s, nil
}

func NewConfig(dsn string) (c *rdbms.ConnConfig, err error) {
	const (
		validScheme = "clickhouse"
	)
	var (
		scheme string
		u      *url.URL
	)

	if u, err = url.Parse(dsn); err != nil {
		return nil, err
	}

	if strings.HasPrefix(dsn, "clickhouse") {
		scheme = u.Scheme
		u.Scheme = validScheme
	} else {
		return nil, fmt.Errorf("expecting valid schema (clickhouse://) at the beginning of the DSN")
	}

	c = &rdbms.ConnConfig{
		DriverName:     scheme,
		DataSourceName: u.String(),
		DBName:         strings.Trim(u.Path, "/"),
		MaskedDSN:      u.Redacted(),
	}

	c.SetDefaults()

	return c, nil
}

func errorHandler(err error) error {
	return err
}
