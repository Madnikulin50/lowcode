package service

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
)

// StockReorderSummary is returned by RunStockReorder after rebuilding submitted POs.
type StockReorderSummary struct {
	OrderCount int64   `json:"orderCount" db:"order_count"`
	LineCount  int64   `json:"lineCount" db:"line_count"`
	TotalQty   float64 `json:"totalQty" db:"total_qty"`
	TotalSum   float64 `json:"totalSum" db:"total_sum"`
}

// RunStockReorder deletes existing source='rulesgo' purchase orders (+ lines) and
// recreates them from stock_reorder_fact (reorder_qty > 0, supplier_id > 0),
// grouped by store × supplier, status=submitted. Fact/slice tables are not touched.
func RunStockReorder(ctx context.Context, namespaceID uint64) (*StockReorderSummary, error) {
	_ = namespaceID // routing / auth context; physical tables are shared

	rs := rdbmsStore(DefaultStore)
	if rs == nil {
		return nil, errors.Internal("store type not supported")
	}
	return runStockReorderSQL(ctx, rs.DB)
}

func runStockReorderSQL(ctx context.Context, db sqlx.ExtContext) (*StockReorderSummary, error) {
	if db == nil {
		return nil, errors.Internal("database connection unavailable")
	}

	if _, err := db.ExecContext(ctx, `
DELETE FROM purchase_order_line
 WHERE po_id IN (SELECT id FROM purchase_order WHERE source = 'rulesgo')`); err != nil {
		return nil, fmt.Errorf("delete purchase_order_line: %w", err)
	}

	if _, err := db.ExecContext(ctx, `
DELETE FROM purchase_order WHERE source = 'rulesgo'`); err != nil {
		return nil, fmt.Errorf("delete purchase_order: %w", err)
	}

	if _, err := db.ExecContext(ctx, `
INSERT INTO purchase_order (
  id, po_number, supplier_id, store_id, status,
  order_date, expected_date, total_qty, total_sum, source, line_count
)
WITH need AS (
  SELECT *
  FROM stock_reorder_fact
  WHERE reorder_qty > 0
    AND COALESCE(supplier_id, 0) > 0
),
hdr AS (
  SELECT
    supplier_id,
    store_id,
    SUM(reorder_qty) AS total_qty,
    SUM(order_value) AS total_sum,
    COUNT(*)::numeric AS line_count,
    MIN(stock_date)::date AS order_date
  FROM need
  GROUP BY supplier_id, store_id
),
idbase AS (
  SELECT COALESCE(MAX(id), 0) AS base FROM purchase_order
)
SELECT
  idbase.base + row_number() OVER (ORDER BY hdr.supplier_id, hdr.store_id),
  ('PO-' || hdr.store_id::bigint || '-' || hdr.supplier_id::bigint || '-' ||
    to_char(COALESCE(hdr.order_date, CURRENT_DATE), 'YYYYMMDD'))::text,
  hdr.supplier_id,
  hdr.store_id,
  'submitted',
  COALESCE(hdr.order_date, CURRENT_DATE)::timestamp,
  (COALESCE(hdr.order_date, CURRENT_DATE) + INTERVAL '7 days')::timestamp,
  ROUND(hdr.total_qty::numeric, 3),
  ROUND(hdr.total_sum::numeric, 3),
  'rulesgo',
  ROUND(hdr.line_count::numeric, 3)
FROM hdr
CROSS JOIN idbase`); err != nil {
		return nil, fmt.Errorf("insert purchase_order: %w", err)
	}

	if _, err := db.ExecContext(ctx, `
INSERT INTO purchase_order_line (
  id, po_id, product_id, ean, product_name, store_id,
  qty_ordered, qty_suggested, unit_cost, line_sum,
  reorder_point, days_of_cover, health_level, rule_score
)
WITH need AS (
  SELECT f.*,
         ROW_NUMBER() OVER (ORDER BY f.order_value DESC NULLS LAST, f.id) AS rn
  FROM stock_reorder_fact f
  WHERE f.reorder_qty > 0
    AND COALESCE(f.supplier_id, 0) > 0
),
capped AS (
  SELECT * FROM need WHERE rn <= 5000
),
idbase AS (
  SELECT COALESCE(MAX(id), 0) AS base FROM purchase_order_line
)
SELECT
  idbase.base + ROW_NUMBER() OVER (ORDER BY po.id, c.product_id),
  po.id,
  c.product_id,
  c.ean,
  c.product_name,
  c.store_id,
  ROUND(c.reorder_qty::numeric, 3),
  ROUND(c.reorder_qty::numeric, 3),
  ROUND(c.unit_cost::numeric, 3),
  ROUND(c.order_value::numeric, 3),
  ROUND(c.reorder_point::numeric, 3),
  ROUND(c.days_of_cover::numeric, 3),
  c.health_level,
  0
FROM capped c
JOIN purchase_order po
  ON po.supplier_id = c.supplier_id
 AND po.store_id = c.store_id
 AND po.source = 'rulesgo'
CROSS JOIN idbase`); err != nil {
		return nil, fmt.Errorf("insert purchase_order_line: %w", err)
	}

	rows, err := db.QueryxContext(ctx, `
SELECT
  (SELECT COUNT(*)::bigint FROM purchase_order WHERE source = 'rulesgo') AS order_count,
  (SELECT COUNT(*)::bigint FROM purchase_order_line pol
     WHERE pol.po_id IN (SELECT id FROM purchase_order WHERE source = 'rulesgo')) AS line_count,
  (SELECT COALESCE(SUM(total_qty), 0)::float8 FROM purchase_order WHERE source = 'rulesgo') AS total_qty,
  (SELECT COALESCE(SUM(total_sum), 0)::float8 FROM purchase_order WHERE source = 'rulesgo') AS total_sum`)
	if err != nil {
		return nil, fmt.Errorf("summary: %w", err)
	}
	defer rows.Close()

	var summary StockReorderSummary
	if !rows.Next() {
		return nil, fmt.Errorf("summary: empty result")
	}
	if err := rows.StructScan(&summary); err != nil {
		return nil, fmt.Errorf("summary: %w", err)
	}

	return &summary, nil
}
