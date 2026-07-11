package clickhouse

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
				aa := make([]any, len(args))
				for a := range args {
					aa[a] = exp.NewCastExpression(exp.NewLiteralExpression("?", args[a]), "String")
				}

				return exp.NewSQLFunctionExpression("concat", aa...)
			},
		},
		"instr": {
			Handler: func(args ...exp.Expression) exp.Expression {
				aa := make([]any, len(args))
				for a := range args {
					aa[a] = exp.NewCastExpression(exp.NewLiteralExpression("?", args[a]), "String")
				}

				return exp.NewSQLFunctionExpression("position", aa[0], aa[1])
			},
		},

		"now": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewSQLFunctionExpression("now")
			},
		},
		"quarter": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewSQLFunctionExpression("toQuarter", args[0])
			},
		},
		"year": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewSQLFunctionExpression("toYear", args[0])
			},
		},
		"month": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewSQLFunctionExpression("toMonth", args[0])
			},
		},
		"this_month": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("? >= toStartOfMonth(now()) AND ? < now()", args[0], args[0])
			},
		},
		"prev_month": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("(? >= toStartOfMonth(now() - INTERVAL 1 MONTH) AND ? < toStartOfMonth(now()))", args[0], args[0])
			},
		},
		"prev_month_truncated": {
			Handler: func(args ...exp.Expression) exp.Expression {
				tm := time.Now()
				interval := fmt.Sprintf("%v", tm.Day())
				return exp.NewLiteralExpression("(? >= toStartOfMonth(now() - INTERVAL 1 MONTH) AND ? < toStartOfMonth(now() - INTERVAL 1 MONTH) + INTERVAL '"+interval+"' DAY)", args[0], args[0])
			},
		},
		"this_year": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("? >= toStartOfYear(now()) AND ? < now()", args[0], args[0])
			},
		},
		"prev_year": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("(? >= toStartOfYear(now() - INTERVAL 1 YEAR) AND ? < toStartOfYear(now()))", args[0], args[0])
			},
		},
		"prev_year_truncated": {
			Handler: func(args ...exp.Expression) exp.Expression {
				tm := time.Now()
				interval := fmt.Sprintf("%v", tm.YearDay())
				return exp.NewLiteralExpression("(? >= toStartOfYear(now() - INTERVAL 1 YEAR) AND ? < toStartOfYear(now() - INTERVAL 1 YEAR) + INTERVAL '"+interval+"' DAY)", args[0], args[0])
			},
		},

		"this_week": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("? >= toStartOfWeek(now()) AND ? < now()", args[0], args[0])
			},
		},
		"prev_week": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("(? >= toStartOfWeek(now() - INTERVAL 1 WEEK) AND ? < toStartOfWeek(now()))", args[0], args[0])
			},
		},
		"prev_week_truncated": {
			Handler: func(args ...exp.Expression) exp.Expression {
				tm := time.Now()
				interval := fmt.Sprintf("%v DAYS + %v HOURS + %v MINUTES",
					int(tm.Weekday()), tm.Hour(), tm.Minute())
				return exp.NewLiteralExpression("(? >= toStartOfWeek(now() - INTERVAL 1 WEEK) AND ? < toStartOfWeek(now() - INTERVAL 1 WEEK) + INTERVAL '"+interval+"')", args[0], args[0])
			},
		},
		"this_quarter": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("? >= toStartOfQuarter(now()) AND ? < now()", args[0], args[0])
			},
		},
		"prev_quarter": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("(? >= toStartOfQuarter(now() - INTERVAL 3 MONTH) AND ? < toStartOfQuarter(now()))", args[0], args[0])
			},
		},
		"prev_quarter_truncated": {
			Handler: func(args ...exp.Expression) exp.Expression {
				tm := time.Now()
				interval := fmt.Sprintf("%v", gvalfnc.DayOfQuarter(tm))
				return exp.NewLiteralExpression("(? >= toStartOfQuarter(now() - INTERVAL 3 MONTH) AND ? < toStartOfQuarter(now() - INTERVAL 3 MONTH) + INTERVAL '"+interval+"' DAY)", args[0], args[0])
			},
		},
		"day_of": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("toStartOfDay(?)", args[0])
			},
		},
		"week_of": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("toStartOfWeek(?)", args[0])
			},
		},
		"month_of": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("toStartOfMonth(?)", args[0])
			},
		},

		"timestamp": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("toDateTime64(?, 3)", args[0])
			},
		},
		"date": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("toDate(?)", args[0])
			},
		},
		"day": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewSQLFunctionExpression("toDayOfMonth", args[0])
			},
		},
		"week": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewSQLFunctionExpression("toWeek", args[0])
			},
		},
		"time": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("toDateTime64(?, 3)", args[0])
			},
		},

		"date_format": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewSQLFunctionExpression("formatDateTime",
					exp.NewLiteralExpression("?", args[0]),
					exp.NewLiteralExpression("?", translateDateFormatParam(args[1])),
				)
			},
		},

		"interval": {
			Handler: func(args ...exp.Expression) exp.Expression {
				return exp.NewLiteralExpression("INTERVAL (? || ' ' || ?)", args[1], args[0])
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
		`%a`, `%a`,
		`%b`, `%b`,
		`%c`, `%c`,
		`%d`, `%d`,
		`%e`, `%e`,
		`%f`, `%f`,
		`%H`, `%H`,
		`%h`, `%h`,
		`%I`, `%I`,
		`%i`, `%i`,
		`%j`, `%j`,
		`%k`, `%k`,
		`%l`, `%l`,
		`%M`, `%M`,
		`%m`, `%m`,
		`%p`, `%p`,
		`%r`, `%r`,
		`%S`, `%S`,
		`%s`, `%s`,
		`%T`, `%T`,
		`%W`, `%W`,
		`%Y`, `%Y`,
		`%y`, `%y`,
		`%%`, `%%`,
	).Replace(s)
}
