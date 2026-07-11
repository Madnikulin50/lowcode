package clickhouse

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/madnikulin50/lowcode/server/pkg/dal"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/ddl"
)

type (
	dataDefiner struct {
		dbName string
		conn   *sqlx.DB
		is     *informationSchema
		d      *clickhouseDialect
	}

	reTypeColumn struct {
		Dialect *clickhouseDialect
		Table   string
		Column  string
		Type    *ddl.ColumnType
	}
)

var (
	_ ddl.DataDefiner = new(dataDefiner)
)

func DataDefiner(dbName string, conn *sqlx.DB) *dataDefiner {
	return &dataDefiner{
		dbName: dbName,
		conn:   conn,
		is:     InformationSchema(conn),
		d:      Dialect(),
	}
}

func (dd *dataDefiner) ConvertModel(m *dal.Model) (*ddl.Table, error) {
	t, err := ddl.ConvertModel(m, dd.d)
	if err != nil {
		return nil, err
	}

	orderBy := "id"
	if len(t.Indexes) > 0 {
		pkFields := make([]string, 0)
		for _, idx := range t.Indexes {
			for _, f := range idx.Fields {
				pkFields = append(pkFields, dd.d.QuoteIdent(f.Column))
			}
		}
		if len(pkFields) > 0 {
			orderBy = stringsJoin(pkFields, ", ")
		}
	}

	engine := "MergeTree()"
	partition := ""
	if m.Refs != nil {
		if v, ok := m.Refs["clickhouse:engine"]; ok {
			engine = fmt.Sprintf("%v", v)
		}
		if v, ok := m.Refs["clickhouse:partition_by"]; ok {
			partition = fmt.Sprintf("\nPARTITION BY %v", v)
		}
	}

	t.Meta = map[string]any{
		"engine":            engine,
		"order_by":          orderBy,
		"partition_by_expr": partition,
	}

	return t, nil
}

func (dd *dataDefiner) ConvertAttribute(attr *dal.Attribute) (*ddl.Column, error) {
	return ddl.ConvertAttribute(attr, dd.d)
}

func (dd *dataDefiner) TableCreate(ctx context.Context, t *ddl.Table) error {
	engine := "MergeTree()"
	orderBy := "id"
	partition := ""

	if t.Meta != nil {
		if v, ok := t.Meta["engine"]; ok {
			engine = fmt.Sprintf("%v", v)
		}
		if v, ok := t.Meta["order_by"]; ok {
			orderBy = fmt.Sprintf("%v", v)
		}
		if v, ok := t.Meta["partition_by_expr"]; ok {
			partition = fmt.Sprintf("%v", v)
		}
	}

	suffix := fmt.Sprintf("\nENGINE = %s\nORDER BY (%s)", engine, orderBy)
	if partition != "" {
		suffix += partition
	}

	return ddl.Exec(ctx, dd.conn, &ddl.CreateTable{
		Table:        t,
		Dialect:      dd.d,
		SuffixClause: suffix,
	})
}

func (dd *dataDefiner) TableDrop(ctx context.Context, t string) error {
	return ddl.Exec(ctx, dd.conn, &ddl.DropTable{
		Dialect: dd.d,
		Table:   t,
	})
}

func (dd *dataDefiner) TableLookup(ctx context.Context, t string) (*ddl.Table, error) {
	return dd.is.TableLookup(ctx, t, dd.dbName)
}

func (dd *dataDefiner) ColumnAdd(ctx context.Context, t string, c *ddl.Column) error {
	return ddl.Exec(ctx, dd.conn, &ddl.AddColumn{
		Dialect: dd.d,
		Table:   t,
		Column:  c,
	})
}

func (dd *dataDefiner) ColumnDrop(ctx context.Context, t, col string) error {
	return ddl.Exec(ctx, dd.conn, &ddl.DropColumn{
		Dialect: dd.d,
		Table:   t,
		Column:  col,
	})
}

func (dd *dataDefiner) ColumnRename(ctx context.Context, t string, o string, n string) error {
	return ddl.Exec(ctx, dd.conn, &ddl.RenameColumn{
		Dialect: dd.d,
		Table:   t,
		Old:     o,
		New:     n,
	})
}

func (dd *dataDefiner) ColumnReType(ctx context.Context, t string, col string, tp *ddl.ColumnType) error {
	return ddl.Exec(ctx, dd.conn, &reTypeColumn{
		Dialect: dd.d,
		Table:   t,
		Column:  col,
		Type:    tp,
	})
}

func (dd *dataDefiner) IndexLookup(ctx context.Context, i, t string) (*ddl.Index, error) {
	index, err := dd.is.IndexLookup(ctx, i, t, dd.dbName)
	if err != nil {
		return nil, err
	}
	return index, nil
}

func (dd *dataDefiner) IndexCreate(ctx context.Context, t string, i *ddl.Index) error {
	sql := fmt.Sprintf(
		"ALTER TABLE %s ADD INDEX %s (%s) TYPE minmax GRANULARITY 1",
		dd.d.QuoteIdent(t),
		dd.d.QuoteIdent(i.Ident),
		i.Fields[0].Column,
	)
	_, err := dd.conn.ExecContext(ctx, sql)
	return err
}

func (dd *dataDefiner) IndexDrop(ctx context.Context, t, i string) error {
	return ddl.Exec(ctx, dd.conn, &ddl.DropIndex{
		Dialect:    dd.d,
		Ident:      i,
		TableIdent: t,
	})
}

func (c *reTypeColumn) ToSQL() (sql string, aa []interface{}, err error) {
	return fmt.Sprintf(
		`ALTER TABLE %s MODIFY COLUMN %s %s`,
		c.Dialect.QuoteIdent(c.Table),
		c.Dialect.QuoteIdent(c.Column),
		c.Type.Name,
	), nil, nil
}

func stringsJoin(elems []string, sep string) string {
	if len(elems) == 0 {
		return ""
	}
	out := elems[0]
	for _, e := range elems[1:] {
		out += sep + e
	}
	return out
}
