package postgres

import (
	"strings"
	"testing"

	"github.com/doug-martin/goqu/v9/exp"
	"github.com/madnikulin50/lowcode/server/pkg/dal"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/ddl"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/ql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestColumnFits(t *testing.T) {
	tcc := []struct {
		name     string
		target   *ddl.Column
		assert   *ddl.Column
		expected bool
	}{
		{
			name: "exact match (text)",
			target: &ddl.Column{
				Type: &ddl.ColumnType{
					Name: "text",
				},
			},
			assert: &ddl.Column{
				Type: &ddl.ColumnType{
					Name: "text",
				},
			},
			expected: true,
		},
		{
			name: "fits somewhere",
			target: &ddl.Column{
				Type: &ddl.ColumnType{
					Name: "text",
				},
			},
			assert: &ddl.Column{
				Type: &ddl.ColumnType{
					Name: "numeric(1,2)",
				},
			},
			expected: true,
		},
		{
			name: "doesn't fit",
			target: &ddl.Column{
				Type: &ddl.ColumnType{
					Name: "numeric(1,2)",
				},
			},
			assert: &ddl.Column{
				Type: &ddl.ColumnType{
					Name: "text",
				},
			},
			expected: false,
		},

		{
			name: "numeric fits",
			target: &ddl.Column{
				Type: &ddl.ColumnType{Name: "numeric(1,2)"},
			},
			assert: &ddl.Column{
				Type: &ddl.ColumnType{Name: "numeric(1,2)"},
			},
			expected: true,
		},

		{
			name: "numeric doesn't fit",
			target: &ddl.Column{
				Type: &ddl.ColumnType{Name: "numeric(1,2)"},
			},
			assert: &ddl.Column{
				Type: &ddl.ColumnType{Name: "numeric(2,3)"},
			},
			expected: false,
		},

		{
			name: "unconstrained numeric fits any numeric(p,s)",
			target: &ddl.Column{
				Type: &ddl.ColumnType{Name: "numeric"},
			},
			assert: &ddl.Column{
				Type: &ddl.ColumnType{Name: "numeric(15,2)"},
			},
			expected: true,
		},

		{
			name: "varchar fits",
			target: &ddl.Column{
				Type: &ddl.ColumnType{Name: "varchar(42)"},
			},
			assert: &ddl.Column{
				Type: &ddl.ColumnType{Name: "varchar(42)"},
			},
			expected: true,
		},

		{
			name: "varchar doesn't fit",
			target: &ddl.Column{
				Type: &ddl.ColumnType{Name: "varchar(42)"},
			},
			assert: &ddl.Column{
				Type: &ddl.ColumnType{Name: "varchar(84)"},
			},
			expected: false,
		},

		{
			name: "timestamp fits into timestamptz",
			target: &ddl.Column{
				Type: &ddl.ColumnType{Name: "timestamptz"},
			},
			assert: &ddl.Column{
				Type: &ddl.ColumnType{Name: "timestamp"},
			},
			expected: true,
		},
		{
			name: "timestamptz doesn't fit into timestamp",
			target: &ddl.Column{
				Type: &ddl.ColumnType{Name: "timestamp"},
			},
			assert: &ddl.Column{
				Type: &ddl.ColumnType{Name: "timestamptz"},
			},
			expected: false,
		},

		{
			name: "time fits into timetz",
			target: &ddl.Column{
				Type: &ddl.ColumnType{Name: "timetz"},
			},
			assert: &ddl.Column{
				Type: &ddl.ColumnType{Name: "time"},
			},
			expected: true,
		},
		{
			name: "timetz doesn't fit into time",
			target: &ddl.Column{
				Type: &ddl.ColumnType{Name: "time"},
			},
			assert: &ddl.Column{
				Type: &ddl.ColumnType{Name: "timetz"},
			},
			expected: false,
		},
	}

	d := postgresDialect{}

	for _, c := range tcc {
		t.Run(c.name, func(t *testing.T) {
			out := d.ColumnFits(c.target, c.assert)
			assert.Equal(t, c.expected, out)
		})
	}

}

func TestJSONRefEqUsesJsonbContainment(t *testing.T) {
	attr := &dal.Attribute{
		Ident: "device",
		Type:  &dal.TypeRef{},
		Store: &dal.CodecRecordValueSetJSON{Ident: "values"},
	}
	model := &dal.Model{Ident: "compose_record"}
	n := &ql.ASTNode{
		Ref: "eq",
		Args: []*ql.ASTNode{
			{Symbol: "device", Meta: map[string]any{"dal.Attribute": attr, "dal.Model": model}},
			{},
		},
	}

	expr, err := dialect.ExprHandler(n, exp.NewLiteralExpression("CAST_IGNORED"), exp.NewLiteralExpression("?", uint64(509728716461178881)))
	require.NoError(t, err)

	sql, args, err := dialect.GOQU().Select(expr).ToSQL()
	require.NoError(t, err)
	require.Contains(t, sql, "@>")
	require.Contains(t, sql, "jsonb_build_object")
	require.Contains(t, sql, "jsonb_build_array")
	require.Contains(t, sql, "to_jsonb")
	require.Contains(t, sql, "::text")
	require.NotContains(t, sql, `->'device'->>0`)
	require.NotContains(t, strings.ToUpper(sql), "BIGINT")
	require.NotContains(t, strings.ToUpper(sql), "CASE")
	require.Equal(t, []any{"device", uint64(509728716461178881)}, args)
	require.Regexp(t, `jsonb_build_object\(\$\d+::text, jsonb_build_array\(to_jsonb\(\$\d+::text\)\)\)`, sql)
}

func TestJSONRefEqSkipsColumnStoredTypeID(t *testing.T) {
	attr := &dal.Attribute{
		Ident: "moduleID",
		Type:  &dal.TypeID{},
		Store: &dal.CodecAlias{Ident: "rel_module"},
	}
	model := &dal.Model{Ident: "compose_record"}
	n := &ql.ASTNode{
		Ref: "eq",
		Args: []*ql.ASTNode{
			{Symbol: "moduleID", Meta: map[string]any{"dal.Attribute": attr, "dal.Model": model}},
			{},
		},
	}

	expr, err := dialect.ExprHandler(n, exp.NewLiteralExpression("CAST_IGNORED"), exp.NewLiteralExpression("?", uint64(509463708787081217)))
	require.NoError(t, err)

	sql, _, err := dialect.GOQU().Select(expr).ToSQL()
	require.NoError(t, err)
	require.NotContains(t, sql, "@>")
	require.NotContains(t, sql, "jsonb_build_object")
}
