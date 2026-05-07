package app

// Registers all supported store backends
import (
	_ "github.com/madnikulin50/lowcode/server/store/adapters/rdbms/drivers/mssql"
	_ "github.com/madnikulin50/lowcode/server/store/adapters/rdbms/drivers/mysql"
	_ "github.com/madnikulin50/lowcode/server/store/adapters/rdbms/drivers/postgres"
	_ "github.com/madnikulin50/lowcode/server/store/adapters/rdbms/drivers/sqlite"
)
