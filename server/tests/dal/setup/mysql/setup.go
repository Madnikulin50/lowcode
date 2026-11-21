package mysql

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/madnikulin50/lowcode/server/pkg/logger"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/drivers/mysql"
)

func Setup(ctx context.Context, dsn string) (_ *sqlx.DB, err error) {
	var (
		cfg *rdbms.ConnConfig
	)

	cfg, err = mysql.NewConfig(dsn)
	if err != nil {
		return
	}

	return rdbms.Connect(ctx, logger.MakeDebugLogger(), cfg)
}
