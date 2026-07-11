package clickhouse

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/jmoiron/sqlx"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/ddl"
)

type (
	informationSchema struct {
		conn *sqlx.DB
	}
)

func InformationSchema(conn *sqlx.DB) *informationSchema {
	return &informationSchema{
		conn: conn,
	}
}

func (i *informationSchema) TableLookup(ctx context.Context, table, dbname string) (*ddl.Table, error) {
	var (
		oneTable = i.columnSelect().Where(
			exp.ParseIdentifier("table").Eq(table),
			exp.ParseIdentifier("database").Eq(dbname),
		)
	)

	if out, err := i.scanColumns(ctx, oneTable); err != nil {
		return nil, err
	} else if len(out) > 0 {
		return out[0], nil
	} else {
		return nil, errors.NotFound("table does not exist")
	}
}

func (i *informationSchema) columnSelect() *goqu.SelectDataset {
	return dialect.GOQU().Select(
		"table",
		"name",
		"type",
		"is_in_partition_key",
		"position",
	).
		From("system.columns").
		Order(
			exp.NewOrderedExpression(exp.ParseIdentifier("table"), exp.AscDir, exp.NoNullsSortType),
			exp.NewOrderedExpression(exp.ParseIdentifier("position"), exp.AscDir, exp.NoNullsSortType),
		)
}

func (i *informationSchema) scanColumns(ctx context.Context, sd *goqu.SelectDataset) (out []*ddl.Table, err error) {
	var (
		at  int
		has bool
		n2p = make(map[string]int)

		typeMapping = map[string]string{
			"DateTime":    "datetime",
			"DateTime64":  "datetime64",
			"FixedString": "fixedstring",
			"UInt8":       "uint8",
			"UInt16":      "uint16",
			"UInt32":      "uint32",
			"UInt64":      "uint64",
			"Int8":        "int8",
			"Int16":       "int16",
			"Int32":       "int32",
			"Int64":       "int64",
			"Float32":     "float32",
			"Float64":     "float64",
		}

		aux = make([]struct {
			Table     string `db:"table"`
			Name      string `db:"name"`
			Type      string `db:"type"`
			Partition any    `db:"is_in_partition_key"`
			Position  any    `db:"position"`
		}, 0)
	)

	if err = ddl.Structs(ctx, i.conn, sd, &aux); err != nil {
		return
	}

	out = make([]*ddl.Table, 0, 10)

	for _, v := range aux {
		if at, has = n2p[v.Table]; !has {
			at = len(out)
			n2p[v.Table] = at
			out = append(out, &ddl.Table{Ident: v.Table})
		}

		tn, has := typeMapping[v.Type]
		if !has {
			tn = v.Type
		}

		out[at].Columns = append(out[at].Columns, &ddl.Column{
			Ident: v.Name,
			Type: &ddl.ColumnType{
				Name: tn,
				Null: true,
			},
		})
	}

	return
}

func (i *informationSchema) IndexLookup(ctx context.Context, index, table, dbname string) (*ddl.Index, error) {
	var (
		oneIndex = i.indexSelect().Where(
			exp.ParseIdentifier("name").Eq(index),
			exp.ParseIdentifier("table").Eq(table),
			exp.ParseIdentifier("database").Eq(dbname),
		)
	)

	if out, err := i.scanIndexes(ctx, oneIndex); err != nil {
		return nil, err
	} else if len(out) > 0 {
		return out[0], nil
	} else {
		return nil, errors.NotFound("index does not exist")
	}
}

func (i *informationSchema) indexSelect() *goqu.SelectDataset {
	return dialect.GOQU().Select(
		"table",
		"name",
		"type",
		"expr",
	).
		From("system.data_skipping_indices").
		Order(
			exp.NewOrderedExpression(exp.ParseIdentifier("table"), exp.AscDir, exp.NoNullsSortType),
			exp.NewOrderedExpression(exp.ParseIdentifier("name"), exp.AscDir, exp.NoNullsSortType),
		)
}

func (i *informationSchema) scanIndexes(ctx context.Context, sd *goqu.SelectDataset) (out []*ddl.Index, err error) {
	var (
		aux = make([]struct {
			Table string `db:"table"`
			Name  string `db:"name"`
			Type  string `db:"type"`
			Expr  string `db:"expr"`
		}, 0)
	)

	if err = ddl.Structs(ctx, i.conn, sd, &aux); err != nil {
		return
	}

	out = make([]*ddl.Index, len(aux))
	for i, a := range aux {
		fields := make([]*ddl.IndexField, 0)
		fields = append(fields, &ddl.IndexField{
			Column: a.Expr,
		})

		out[i] = &ddl.Index{
			Ident:      a.Name,
			TableIdent: a.Table,
			Type:       a.Type,
			Fields:     fields,
		}
	}

	return
}
