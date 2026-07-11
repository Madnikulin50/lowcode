package clickhouse

import (
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9/exp"
)

func DeepIdentJSON(asString bool, jsonDoc exp.Expression, pp ...any) exp.LiteralExpression {
	var (
		sql  strings.Builder
		last = len(pp) - 1
	)

	if asString {
		sql.WriteString("JSONExtractString(?, ")
	} else {
		sql.WriteString("JSONExtractRaw(?, ")
	}

	for i, p := range pp {
		if i > 0 {
			sql.WriteString(", ")
		}

		switch path := p.(type) {
		case string:
			sql.WriteString("'")
			sql.WriteString(strings.ReplaceAll(path, "'", "\\'"))
			sql.WriteString("'")
		case int:
			sql.WriteString(fmt.Sprintf("%d", path))
		default:
			panic("invalid type")
		}

		if i == last {
			sql.WriteString(")")
		}
	}

	return exp.NewLiteralExpression(sql.String(), jsonDoc)
}
