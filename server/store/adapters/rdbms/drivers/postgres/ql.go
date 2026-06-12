package postgres

import (
	"fmt"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9/exp"
	"github.com/madnikulin50/lowcode/server/pkg/gvalfnc"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/ql"
)

var (
	ref2exp = ql.ExprHandlerMap{
		"concat": {
			Handler: func(args ...exp.Expression) exp.Expression {
				// need to force text type on all arguments
				aa := make([]any, len(args))
				for a := range args {
					aa[a] = exp.NewCastExpression(exp.NewLiteralExpression("?", args[a]), "TEXT")
				}

				return exp.NewSQLFunctionExpression("CONCAT", aa...)
			},
		},
		"instr": {
			Handler: func(args ...exp.Expression) exp.Expression {
				// need to force text type on all arguments
				aa := make([]any, len(args))
				for a := range args {
					aa[a] = exp.NewCastExpression(exp.NewLiteralExpression("?", args[a]), "TEXT")
				}

				return exp.NewSQLFunctionExpression("INSTR", args[0], args[1])
			},
		},

		// filtering
		"now": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewSQLFunctionExpression("NOW")
			},
		},
		"quarter": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewSQLFunctionExpression("EXTRACT",
					exp.NewLiteralExpression("QUARTER FROM ?", args[0]),
				)
			},
		},
		"year": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewSQLFunctionExpression("EXTRACT",
					exp.NewLiteralExpression("YEAR FROM ?", args[0]),
				)
			},
		},
		"month": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewSQLFunctionExpression("EXTRACT",
					exp.NewLiteralExpression("MONTH FROM ?", args[0]),
				)
			},
		},
		"this_month": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("?::timestamp > DATE_TRUNC('month', NOW()::timestamp)::timestamp AND ?::timestamp < NOW()::timestamp", args[0], args[0])
			},
		},
		"prev_month": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("(?::timestamp >= date_trunc('month', NOW()::timestamp - interval '1 month')::timestamp  AND ?::timestamp < date_trunc('month', NOW()::timestamp)::timestamp)", args[0], args[0])
			},
		},
		"prev_month_truncated": {
			Handler: func(args ...exp.Expression) exp.Expression {
				tm := time.Now()
				interval := fmt.Sprintf("%v days", tm.Day())
				return exp.NewLiteralExpression("(?::timestamp >= date_trunc('month', NOW()::timestamp - interval '1 month')::timestamp AND ?::timestamp < (date_trunc('month', NOW()::timestamp - interval '1 month')+ interval '"+interval+"')::timestamp)", args[0], args[0])
			},
		},
		"this_year": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("?::timestamp > DATE_TRUNC('year', NOW()::timestamp)::timestamp AND ?::timestamp < NOW()::timestamp", args[0], args[0])
			},
		},
		"prev_year": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("(?::timestamp >= date_trunc('year', NOW()::timestamp - interval '1 year')::timestamp  AND ?::timestamp < date_trunc('year', NOW()::timestamp)::timestamp)", args[0], args[0])
			},
		},
		"prev_year_truncated": {
			Handler: func(args ...exp.Expression) exp.Expression {
				tm := time.Now()
				interval := fmt.Sprintf("%v days", tm.YearDay())
				return exp.NewLiteralExpression("(?::timestamp >= date_trunc('year', NOW()::timestamp - interval '1 year')::timestamp AND ?::timestamp < (date_trunc('year', NOW()::timestamp - interval '1 year')+ interval '"+interval+"')::timestamp)", args[0], args[0])
			},
		},

		"this_week": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("?::timestamp > DATE_TRUNC('week', NOW()::timestamp)::timestamp AND ?::timestamp < NOW()::timestamp", args[0], args[0])
			},
		},
		"prev_week": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("(?::timestamp >= date_trunc('week', NOW()::timestamp - interval '1 week')::timestamp  AND ?::timestamp < date_trunc('week', NOW()::timestamp)::timestamp)", args[0], args[0])
			},
		},
		"prev_week_truncated": {
			Handler: func(args ...exp.Expression) exp.Expression {
				tm := time.Now()
				interval := fmt.Sprintf("%v days + %v hours + %v minutes",
					int(tm.Weekday()), tm.Hour(), tm.Minute())
				return exp.NewLiteralExpression("(?::timestamp >= date_trunc('week', NOW()::timestamp - interval '1 week')::timestamp AND ?::timestamp < (date_trunc('week', NOW()::timestamp - interval '1 week')+ interval '"+interval+"')::timestamp)", args[0], args[0])
			},
		},
		"this_quarter": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("?::timestamp > DATE_TRUNC('quarter', NOW()::timestamp)::timestamp AND ?::timestamp < NOW()::timestamp", args[0], args[0])
			},
		},
		"prev_quarter": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("(?::timestamp >= date_trunc('quarter', NOW()::timestamp - interval '90 days')::timestamp  AND ?::timestamp < date_trunc('quarter', NOW()::timestamp)::timestamp)", args[0], args[0])
			},
		},
		"prev_quarter_truncated": {
			Handler: func(args ...exp.Expression) exp.Expression {
				tm := time.Now()
				interval := fmt.Sprintf("%v days", gvalfnc.DayOfQuarter(tm))
				return exp.NewLiteralExpression("(?::timestamp >= date_trunc('quarter', NOW()::timestamp - interval '90 days')::timestamp AND ?::timestamp < (date_trunc('quarter', NOW()::timestamp - interval '90 days')+ interval '"+interval+"')::timestamp)", args[0], args[0])
			},
		},
		"day_of": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("date_trunc('day', ?::timestamp)::timestamp", args[0])
			},
		},
		"week_of": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("date_trunc('week', ?::timestamp)::timestamp", args[0])
			},
		},
		"month_of": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("date_trunc('month', ?::timestamp)::timestamp", args[0])
			},
		},

		"timestamp": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("?::TIMESTAMPTZ", args[0])
			},
		},
		"date": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("?::DATE", args[0])
			},
		},
		"day": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewSQLFunctionExpression("EXTRACT",
					exp.NewLiteralExpression("DAY FROM ?", args[0]),
				)
			},
		},
		"week": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewSQLFunctionExpression("EXTRACT",
					exp.NewLiteralExpression("WEEK FROM ?", args[0]),
				)
			},
		},
		"time": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("DATE_TRUNC('second', ?::TIME)::TIME", args[0])
			},
		},

		// @todo replace given argument before constructing sql
		"date_format": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewSQLFunctionExpression("TO_CHAR",
					exp.NewLiteralExpression("?::TIMESTAMPTZ", args[0]),
					exp.NewLiteralExpression("?::TEXT", translateDateFormatParam(args[1])),
				)
			},
		},

		"interval": {
			Handler: func(args ...exp.Expression) exp.Expression {
				// The problem here is that PGSQL, to my findings, doesn't have functions to add/sub dates
				// like MySQL for example.
				//
				// We need to construct an expression in the lines of `INTERVAL 'N UNIT'` which
				// then becomes, for example, d + INTERVAL 'N UNIT'.
				//
				// The problem #2 is that we can't just use value placeholders in string literals
				// nor is there a 2 arg function to make an interval. There is a make_interval function
				// but that one won't do.
				//
				// So...
				// (?  || 'S') makes the interval label a plural because MySQL uses singular and that
				// is what QL supports. PgSQL uses plural, such as years, months, and days.
				//
				// The rest of the expression is just to construct the string which can then be casted to INTERVAL
				// which can then be used in date math, which is done with regular math operators.
				return exp.NewLiteralExpression("(?::INTEGER || ' ' || (?  || 'S'))::INTERVAL", args[1], args[0])
			},
		},

		"date_add": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("(? + ?)", args[0], args[1])
			},
		},

		"date_sub": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("(? - ?)", args[0], args[1])
			},
		},

		// functions currently unsupported in PostgreSQL store backend
		// "STD": {
		//	Handler: func(args ...exp.Expression) exp.Expression {
		//		return exp.NewLiteralExpression("")
		//	},
		// },
	}.ExprHandlers()
)

func translateDateFormatParam(e interface{}) interface{} {
	le, ok := e.(exp.LiteralExpression)
	if !ok {
		return e
	}

	args := le.Args()
	if len(args) > 0 {
		return dateFormatReplacer(fmt.Sprintf("%s", args[0]))
	}

	return e
}

func dateFormatReplacer(s string) string {
	return strings.NewReplacer(
		// @todo Doing ...%dT%H... (for iso timestamp) pgsql doesn't format it correctly
		// so I'm covering this edge case.
		// We should fix this properly when we redo record storage.
		`%dT%H`, `DD"T"HH24`,

		`%a`, `Dy`,
		`%b`, `Mon`,
		`%c`, `FMMM`,
		`%d`, `DD`,
		`%e`, `FMDD`,
		`%f`, `US`,
		`%H`, `HH24`,
		`%h`, `HH12`,
		`%I`, `HH12`,
		`%i`, `MI`,
		`%j`, `DDD`,
		`%k`, `FMHH24`,
		`%l`, `FMHH12`,
		`%M`, `FMMonth`,
		`%m`, `MM`,
		`%p`, `AM`,
		`%r`, `HH12:MI:SS AM`,
		`%S`, `SS`,
		`%s`, `SS`,
		`%T`, `HH24:MI:SS`,
		`%W`, `FMDay`,
		`%Y`, `YYYY`,
		`%y`, `YY`,
		`%%`, `%`,
	).Replace(s)
}
