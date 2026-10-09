-- Gross-margin fact + slices for namespace loop.
-- As-of cost: cost_price valid from date until next cost (fallback: first cost).

CREATE INDEX IF NOT EXISTS cost_price_product_date_idx ON cost_price (product_id, date DESC);

DROP MATERIALIZED VIEW IF EXISTS receipt_margin_slice CASCADE;
DROP MATERIALIZED VIEW IF EXISTS receipt_margin_kpi CASCADE;
DROP MATERIALIZED VIEW IF EXISTS receipt_margin CASCADE;
DROP TABLE IF EXISTS receipt_margin_slice CASCADE;
DROP TABLE IF EXISTS receipt_margin_kpi CASCADE;
DROP TABLE IF EXISTS receipt_margin CASCADE;

CREATE TABLE receipt_margin AS
WITH costs AS (
  SELECT
    product_id,
    date AS valid_from,
    cost_price,
    COALESCE(lead(date) OVER (PARTITION BY product_id ORDER BY date), TIMESTAMP '9999-12-31') AS valid_to
  FROM cost_price
),
first_cost AS (
  SELECT DISTINCT ON (product_id) product_id, cost_price
  FROM cost_price
  ORDER BY product_id, date ASC
)
SELECT
  r.id,
  r.store_id,
  s.store_name,
  s.store_type,
  r.product_id,
  p.product_name,
  p.brand,
  p.category_id,
  cat.category_name,
  r.dt,
  r.quantity,
  r.position_sum AS revenue,
  r.vat,
  r.discount,
  COALESCE(c.cost_price, f.cost_price) AS unit_cost,
  r.quantity * COALESCE(c.cost_price, f.cost_price) AS cogs,
  r.position_sum - r.quantity * COALESCE(c.cost_price, f.cost_price) AS gross_profit,
  CASE WHEN r.position_sum IS NULL OR r.position_sum = 0 THEN NULL
       ELSE round(100 * (r.position_sum - r.quantity * COALESCE(c.cost_price, f.cost_price)) / r.position_sum, 2)
  END AS margin_pct
FROM receipt_positions r
JOIN products p ON p.product_id = r.product_id
JOIN categories cat ON cat.category_id = p.category_id
JOIN stores s ON s.store_id = r.store_id
LEFT JOIN costs c
  ON c.product_id = r.product_id
 AND r.dt >= c.valid_from
 AND r.dt < c.valid_to
LEFT JOIN first_cost f ON f.product_id = r.product_id;

CREATE UNIQUE INDEX receipt_margin_id_idx ON receipt_margin (id);
CREATE INDEX receipt_margin_dt_idx ON receipt_margin (dt);
CREATE INDEX receipt_margin_cat_idx ON receipt_margin (category_name);
CREATE INDEX receipt_margin_brand_idx ON receipt_margin (brand);
CREATE INDEX receipt_margin_store_idx ON receipt_margin (store_name);

-- CREATE TABLE AS leaves computed numerics unconstrained; Corteza Number fields
-- map to NUMERIC(15, scale) and ColumnFits would otherwise emit attributeReType
-- and skip loading the model into the DAL connection cache.
ALTER TABLE receipt_margin
  ALTER COLUMN cogs TYPE numeric(15,3),
  ALTER COLUMN gross_profit TYPE numeric(15,3),
  ALTER COLUMN margin_pct TYPE numeric(15,3);

CREATE TABLE receipt_margin_kpi AS
SELECT
  1::numeric AS id,
  round(sum(revenue), 2) AS revenue,
  round(sum(cogs), 2) AS cogs,
  round(sum(gross_profit), 2) AS gross_profit,
  round(100 * sum(gross_profit) / NULLIF(sum(revenue), 0), 2) AS margin_pct,
  round(sum(vat), 2) AS vat,
  round(sum(discount), 2) AS discount,
  count(*)::numeric AS line_count,
  count(DISTINCT store_id)::numeric AS store_count,
  count(DISTINCT product_id)::numeric AS sku_count
FROM receipt_margin;

CREATE UNIQUE INDEX receipt_margin_kpi_id_idx ON receipt_margin_kpi (id);

ALTER TABLE receipt_margin_kpi
  ALTER COLUMN revenue TYPE numeric(15,3),
  ALTER COLUMN cogs TYPE numeric(15,3),
  ALTER COLUMN gross_profit TYPE numeric(15,3),
  ALTER COLUMN margin_pct TYPE numeric(15,3),
  ALTER COLUMN vat TYPE numeric(15,3),
  ALTER COLUMN discount TYPE numeric(15,3),
  ALTER COLUMN line_count TYPE numeric(15,3),
  ALTER COLUMN store_count TYPE numeric(15,3),
  ALTER COLUMN sku_count TYPE numeric(15,3);

CREATE TABLE receipt_margin_slice AS
WITH u AS (
  SELECT 'category'::text AS slice_kind, category_name AS slice_name, 0 AS slice_order,
         sum(revenue) AS revenue, sum(cogs) AS cogs, sum(gross_profit) AS gross_profit, count(*) AS line_count
  FROM receipt_margin GROUP BY category_name
  UNION ALL
  SELECT 'brand', brand, 0,
         sum(revenue), sum(cogs), sum(gross_profit), count(*)
  FROM receipt_margin GROUP BY brand
  UNION ALL
  SELECT 'brand_top', brand, 0, revenue, cogs, gross_profit, line_count
  FROM (
    SELECT brand,
           sum(revenue) AS revenue, sum(cogs) AS cogs, sum(gross_profit) AS gross_profit, count(*) AS line_count
    FROM receipt_margin GROUP BY brand
    ORDER BY sum(revenue) DESC
    LIMIT 12
  ) brand_top
  UNION ALL
  SELECT 'store', store_name, 0,
         sum(revenue), sum(cogs), sum(gross_profit), count(*)
  FROM receipt_margin GROUP BY store_name
  UNION ALL
  SELECT 'store_type', store_type, 0,
         sum(revenue), sum(cogs), sum(gross_profit), count(*)
  FROM receipt_margin GROUP BY store_type
  UNION ALL
  SELECT 'month', to_char(date_trunc('month', dt), 'YYYY-MM'), 0,
         sum(revenue), sum(cogs), sum(gross_profit), count(*)
  FROM receipt_margin GROUP BY date_trunc('month', dt)
  UNION ALL
  SELECT 'sku', brand || ' · ' || product_name, 0,
         sum(revenue), sum(cogs), sum(gross_profit), count(*)
  FROM receipt_margin GROUP BY brand, product_name
  UNION ALL
  SELECT 'bridge', x.slice_name, x.slice_order, x.amount, 0, 0, 0
  FROM receipt_margin_kpi k
  CROSS JOIN LATERAL (VALUES
    ('1. Выручка', 1, k.revenue),
    ('2. Себестоимость', 2, k.cogs),
    ('3. Валовая прибыль', 3, k.gross_profit)
  ) AS x(slice_name, slice_order, amount)
)
SELECT
  row_number() OVER (ORDER BY slice_kind, slice_name)::numeric AS id,
  slice_kind,
  slice_name,
  slice_order,
  round(revenue, 2) AS revenue,
  round(cogs, 2) AS cogs,
  round(gross_profit, 2) AS gross_profit,
  CASE WHEN revenue IS NULL OR revenue = 0 THEN NULL
       ELSE round(100 * gross_profit / revenue, 2)
  END AS margin_pct,
  line_count::numeric AS line_count
FROM u;

CREATE UNIQUE INDEX receipt_margin_slice_id_idx ON receipt_margin_slice (id);
CREATE INDEX receipt_margin_slice_kind_idx ON receipt_margin_slice (slice_kind);

ALTER TABLE receipt_margin_slice
  ALTER COLUMN revenue TYPE numeric(15,3),
  ALTER COLUMN cogs TYPE numeric(15,3),
  ALTER COLUMN gross_profit TYPE numeric(15,3),
  ALTER COLUMN margin_pct TYPE numeric(15,3),
  ALTER COLUMN line_count TYPE numeric(15,3);
