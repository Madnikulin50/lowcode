package clickhouse

import (
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/doug-martin/goqu/v9/sqlgen"
	"github.com/madnikulin50/lowcode/server/pkg/dal"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/ddl"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/drivers"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/ql"
	"github.com/spf13/cast"
)

type (
	clickhouseDialect struct{}
)

var (
	_ drivers.Dialect = &clickhouseDialect{}

	dialect            = &clickhouseDialect{}
	goquDialectWrapper = goqu.Dialect("mysql")
	goquDialectOptions = mysql.DialectOptions()
	quoteIdent         = string(mysql.DialectOptions().QuoteRune)

	nuances = drivers.Nuances{
		HavingClauseMustUseAlias: true,
	}
)

func Dialect() *clickhouseDialect {
	return dialect
}

func (clickhouseDialect) Nuances() drivers.Nuances {
	return nuances
}

func (clickhouseDialect) GOQU() goqu.DialectWrapper                 { return goquDialectWrapper }
func (clickhouseDialect) DialectOptions() *sqlgen.SQLDialectOptions { return goquDialectOptions }
func (clickhouseDialect) QuoteIdent(i string) string                { return quoteIdent + i + quoteIdent }

func (d clickhouseDialect) IndexFieldModifiers(attr *dal.Attribute, mm ...dal.IndexFieldModifier) (string, error) {
	return drivers.IndexFieldModifiers(attr, d.QuoteIdent, mm...)
}

func (d clickhouseDialect) JsonQuote(expr exp.Expression) exp.Expression {
	return expr
}

func (d clickhouseDialect) JsonExtract(ident exp.Expression, pp ...any) (exp.Expression, error) {
	return DeepIdentJSON(false, ident, pp...), nil
}

func (d clickhouseDialect) JsonExtractUnquote(ident exp.Expression, pp ...any) (exp.Expression, error) {
	return DeepIdentJSON(true, ident, pp...), nil
}

func (d clickhouseDialect) AggregateBase(t drivers.TableCodec, groupBy []dal.AggregateAttr, out []dal.AggregateAttr) (slct *goqu.SelectDataset) {
	var (
		cols = t.Columns()

		q = d.GOQU().
			From(t.Ident())
	)

	for _, g := range groupBy {
		if g.MultiValue {
			colName := g.RawExpr
			xpr, err := t.AttributeExpressionQuoted(colName)
			if err != nil {
				q = q.SetError(err)
				return q
			}

			q = q.From(
				t.Ident(),
				goqu.Func("JSONExtractArrayRaw", xpr).As(colName),
			)
		}
	}

	if len(cols) == 0 {
		return q.SetError(fmt.Errorf("can not create SELECT without columns"))
	}

	q = q.Select(t.Ident().Col(cols[0].Name()))
	for _, col := range cols[1:] {
		q = q.SelectAppend(t.Ident().Col(col.Name()))
	}

	return q
}

func (d clickhouseDialect) JsonArrayContains(needle, haystack exp.Expression) (exp.Expression, error) {
	return exp.NewLiteralExpression("has(JSONExtractArrayRaw(?), JSONExtractRaw(?))", haystack, needle), nil
}

func (d clickhouseDialect) TableCodec(m *dal.Model) drivers.TableCodec {
	return drivers.NewTableCodec(m, d)
}

func (d clickhouseDialect) TypeWrap(dt dal.Type) drivers.Type {
	switch c := dt.(type) {
	case *dal.TypeTimestamp:
		return &TypeTimestamp{c}
	case *dal.TypeBoolean:
		return &TypeBoolean{c}
	}

	return drivers.TypeWrap(dt)
}

func (clickhouseDialect) AttributeCast(attr *dal.Attribute, val exp.Expression) (expr exp.Expression, err error) {
	switch attr.Type.(type) {
	case *dal.TypeText:
		expr = exp.NewCastExpression(val, "String")

	case *dal.TypeBoolean:
		expr = exp.NewCastExpression(val, "String")
		expr = exp.NewBooleanExpression(exp.EqOp, expr, exp.NewLiteralExpression(`'true'`))

	default:
		return drivers.AttributeCast(attr, val)
	}

	return
}

func (clickhouseDialect) AttributeExpression(attr *dal.Attribute, modelIdent string, ident string) (expr exp.Expression, err error) {
	identExpr := exp.NewIdentifierExpression("", modelIdent, ident)

	if attr.Type.Type() == dal.AttributeTypeTimestamp {
		return exp.NewLiteralExpression("date_trunc(?, ?)", "second", identExpr), nil
	}

	return exp.NewLiteralExpression("?", identExpr), nil
}

func (clickhouseDialect) AttributeToColumn(attr *dal.Attribute) (col *ddl.Column, err error) {
	col = &ddl.Column{
		Ident:   attr.StoreIdent(),
		Comment: attr.Label,
		Type: &ddl.ColumnType{
			Null: attr.Type.IsNullable(),
		},
	}

	switch t := attr.Type.(type) {
	case *dal.TypeID:
		col.Type.Name = "UInt64"
		col.Default = ddl.DefaultID(t.HasDefault, t.DefaultValue)
	case *dal.TypeRef:
		col.Type.Name = "UInt64"
		col.Default = ddl.DefaultID(t.HasDefault, t.DefaultValue)

	case *dal.TypeTimestamp:
		col.Type.Name = "DateTime64(3)"

		if t.Timezone {
			col.Type.Name = "DateTime64(3)"
		}

		if t.Precision >= 0 {
			col.Type.Name = fmt.Sprintf("DateTime64(%d)", t.Precision)
		}

		col.Default = ddl.DefaultValueCurrentTimestamp(t.DefaultCurrentTimestamp)

	case *dal.TypeTime:
		col.Type.Name = "DateTime64(3)"

		if t.Precision >= 0 {
			col.Type.Name = fmt.Sprintf("DateTime64(%d)", t.Precision)
		}

		col.Default = ddl.DefaultValueCurrentTimestamp(t.DefaultCurrentTimestamp)
	case *dal.TypeDate:
		col.Type.Name = "Date"
		col.Default = ddl.DefaultValueCurrentTimestamp(t.DefaultCurrentTimestamp)

	case *dal.TypeNumber:
		if numType := cast.ToString(t.Meta["rdbms:type"]); numType != "" {
			col.Type.Name = numType
			col.Default = ddl.DefaultNumber(t.HasDefault, 0, t.DefaultValue)
			break
		}

		col.Type.Name = "Float64"

		switch {
		case t.Precision > 0 && t.Scale > 0:
			col.Type.Name = fmt.Sprintf("Decimal(%d, %d)", t.Precision, t.Scale)
		case t.Precision > 0:
			col.Type.Name = fmt.Sprintf("Decimal(%d, 0)", t.Precision)
		}

		col.Default = ddl.DefaultNumber(t.HasDefault, t.Precision, t.DefaultValue)

	case *dal.TypeText:
		if t.Length > 0 {
			col.Type.Name = fmt.Sprintf("FixedString(%d)", t.Length)
		} else {
			col.Type.Name = "String"
		}

		if t.HasDefault {
			col.Default = fmt.Sprintf("%q", t.DefaultValue)
		}

	case *dal.TypeJSON:
		col.Type.Name = "String"
		if col.Default, err = ddl.DefaultJSON(t.HasDefault, t.DefaultValue); err != nil {
			return nil, err
		}

	case *dal.TypeGeometry:
		col.Type.Name = "String"

	case *dal.TypeBlob:
		col.Type.Name = "String"

	case *dal.TypeBoolean:
		col.Type.Name = "UInt8"
		col.Default = ddl.DefaultBoolean(t.HasDefault, t.DefaultValue)

	case *dal.TypeUUID:
		col.Type.Name = "UUID"

	case *dal.TypeEnum:
		col.Type.Name = "String"

	default:
		return nil, fmt.Errorf("unsupported column type: %s ", t.Type())
	}

	return
}

func (clickhouseDialect) ColumnFits(target, assert *ddl.Column) bool {
	targetType, targetName, targetMeta := ddl.ParseColumnTypes(target)
	assertType, assertName, assertMeta := ddl.ParseColumnTypes(assert)

	if assertType == targetType {
		return true
	}

	matches := map[string]map[string]bool{
		"uint64": {
			"string":  true,
			"int64":   true,
			"int32":   true,
			"float64": true,
		},
		"int64": {
			"string": true,
			"uint64": true,
			"int32":  true,
		},
		"float64": {
			"string":  true,
			"decimal": true,
		},
		"decimal": {
			"string":  true,
			"float64": true,
		},
		"datetime64": {
			"string": true,
		},
		"datetime": {
			"string":     true,
			"datetime64": true,
		},
		"date": {
			"string": true,
		},
		"string": {},
		"fixedstring": {
			"string": true,
		},
		"uint8": {
			"uint64": true,
		},
		"uuid": {
			"string": true,
		},
	}

	baseMatch := assertName == targetName || matches[assertName][targetName]

	switch {
	case assertName == "fixedstring" && targetName == "fixedstring":
		for i := len(assertMeta); i < 1; i++ {
			assertMeta = append(assertMeta, "0")
		}
		for i := len(targetMeta); i < 1; i++ {
			targetMeta = append(targetMeta, "0")
		}

		return baseMatch && cast.ToInt(assertMeta[0]) <= cast.ToInt(targetMeta[0])

	case assertName == "decimal" && targetName == "decimal":
		for i := len(assertMeta); i < 2; i++ {
			assertMeta = append(assertMeta, "0")
		}
		for i := len(targetMeta); i < 2; i++ {
			targetMeta = append(targetMeta, "0")
		}

		return baseMatch && cast.ToInt(assertMeta[0]) <= cast.ToInt(targetMeta[0]) && cast.ToInt(assertMeta[1]) <= cast.ToInt(targetMeta[1])
	}

	return baseMatch
}

func (d clickhouseDialect) ExprHandler(n *ql.ASTNode, args ...exp.Expression) (expr exp.Expression, err error) {
	switch ref := strings.ToLower(n.Ref); ref {
	case "concat":
		return castColumnDataToText("concat", args...)

	case "in":
		return drivers.OpHandlerIn(d, n, args...)

	case "nin":
		return drivers.OpHandlerNotIn(d, n, args...)

	case "like", "nlike":
		if dalType, ok := n.Args[0].Meta["dal.Attribute"].(*dal.Attribute); ok {
			col, err := d.AttributeToColumn(dalType)
			if err != nil {
				return nil, err
			}

			if col.Type.Name == "UInt64" {
				op := "LIKE"
				if ref == "nlike" {
					op = "NOT LIKE"
				}
				return castColumnDataToText(op, args...)
			}
		}
	}

	return ref2exp.RefHandler(n, args...)
}

func (d clickhouseDialect) ValHandler(n *ql.ASTNode) (out exp.Expression, err error) {
	switch v := n.Value.V.(type) {
	case *expr.Boolean:
		if cast.ToBool(v.Get()) {
			out = exp.NewLiteralExpression("1")
		} else {
			out = exp.NewLiteralExpression("0")
		}
	default:
		out = exp.NewLiteralExpression("?", n.Value.V.Get())
	}

	return
}

func (d clickhouseDialect) OrderedExpression(expr exp.Expression, dir exp.SortDirection, nst exp.NullSortType) exp.OrderedExpression {
	return exp.NewOrderedExpression(expr, dir, nst)
}

func castColumnDataToText(op string, args ...exp.Expression) (expr exp.Expression, err error) {
	aa := make([]any, len(args))
	for a := range args {
		aa[a] = exp.NewCastExpression(exp.NewLiteralExpression("?", args[a]), "String")
	}

	return exp.NewSQLFunctionExpression(op, aa...), nil
}
