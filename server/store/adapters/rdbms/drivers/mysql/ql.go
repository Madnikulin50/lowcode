package mysql

import (
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/ql"
)

var (
	ref2exp = ql.ExprHandlerMap{
		"std": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewSQLFunctionExpression("STD", args[0])
			},
		},
	}.ExprHandlers()
)
