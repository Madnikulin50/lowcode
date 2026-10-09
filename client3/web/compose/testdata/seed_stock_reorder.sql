-- Stock health + auto-order fact tables for namespace loop.
-- Demand: last 30 days of receipt_positions relative to max(dt).
-- Policy: stock_policy with resolve product+store → product → category+store → category → store → global.

DROP TABLE IF EXISTS purchase_order_line CASCADE;
DROP TABLE IF EXISTS purchase_order CASCADE;
DROP TABLE IF EXISTS stock_health_slice CASCADE;
DROP TABLE IF EXISTS stock_reorder_fact CASCADE;
DROP TABLE IF EXISTS stock_policy CASCADE;

-- ---------------------------------------------------------------------------
-- stock_policy (seed one global default)
-- ---------------------------------------------------------------------------
CREATE TABLE stock_policy (
  id numeric PRIMARY KEY,
  policy_id numeric,
  store_id numeric,
  product_id numeric,
  category_id numeric,
  lead_time_days numeric(15,3) DEFAULT 7,
  review_period_days numeric(15,3) DEFAULT 7,
  target_doc_days numeric(15,3) DEFAULT 21,
  moq numeric(15,3) DEFAULT 12,
  pack_size numeric(15,3) DEFAULT 12,
  service_z numeric(15,3) DEFAULT 1.65,
  min_qty numeric(15,3) DEFAULT 0,
  max_qty numeric(15,3) DEFAULT 0
);

INSERT INTO stock_policy (
  id, policy_id, store_id, product_id, category_id,
  lead_time_days, review_period_days, target_doc_days,
  moq, pack_size, service_z, min_qty, max_qty
) VALUES (
  1, 1, NULL, NULL, NULL,
  7, 7, 21,
  12, 12, 1.65, 0, 0
);

-- ---------------------------------------------------------------------------
-- stock_reorder_fact — one row per product_stock (store × product)
-- ---------------------------------------------------------------------------
CREATE TABLE stock_reorder_fact AS
WITH bounds AS (
  SELECT max(dt)::date AS as_of FROM receipt_positions
),
days AS (
  SELECT generate_series(
    (SELECT as_of - 29 FROM bounds),
    (SELECT as_of FROM bounds),
    '1 day'::interval
  )::date AS d
),
daily AS (
  SELECT
    ps.store_id,
    ps.product_id,
    days.d,
    COALESCE(SUM(rp.quantity), 0) AS qty
  FROM product_stock ps
  CROSS JOIN days
  LEFT JOIN receipt_positions rp
    ON rp.store_id = ps.store_id
   AND rp.product_id = ps.product_id
   AND rp.dt::date = days.d
  GROUP BY ps.store_id, ps.product_id, days.d
),
demand AS (
  SELECT
    store_id,
    product_id,
    ROUND((COALESCE(SUM(qty), 0) / 30.0)::numeric, 6) AS avg_daily_qty,
    ROUND(COALESCE(stddev_samp(qty), 0)::numeric, 6) AS demand_std_daily
  FROM daily
  GROUP BY store_id, product_id
),
costs AS (
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
),
base AS (
  SELECT
    ps.id,
    ps.store_id,
    ps.product_id,
    p.ean,
    p.product_name,
    p.category_id,
    CASE
      WHEN jsonb_typeof(p.values->'supplier_id') = 'array'
        THEN NULLIF(TRIM(COALESCE(p.values->'supplier_id'->>0, '')), '')::numeric
      WHEN jsonb_typeof(p.values->'supplier_id') = 'number'
        THEN (p.values->>'supplier_id')::numeric
      WHEN jsonb_typeof(p.values->'supplier_id') = 'string'
        THEN NULLIF(TRIM(p.values->>'supplier_id'), '')::numeric
      ELSE NULL
    END AS supplier_id,
    ps.stock_quantity,
    ps.stock_sum,
    COALESCE(d.avg_daily_qty, 0) AS avg_daily_qty,
    COALESCE(d.demand_std_daily, 0) AS demand_std_daily,
    COALESCE(c.cost_price, f.cost_price, 0) AS unit_cost,
    ps.date AS stock_date,
    pol.lead_time_days,
    pol.review_period_days,
    pol.target_doc_days,
    pol.moq,
    pol.pack_size,
    pol.service_z,
    pol.min_qty,
    pol.max_qty
  FROM product_stock ps
  JOIN products p ON p.product_id = ps.product_id
  LEFT JOIN demand d
    ON d.store_id = ps.store_id AND d.product_id = ps.product_id
  LEFT JOIN costs c
    ON c.product_id = ps.product_id
   AND ps.date >= c.valid_from
   AND ps.date < c.valid_to
  LEFT JOIN first_cost f ON f.product_id = ps.product_id
  LEFT JOIN LATERAL (
    SELECT *
    FROM stock_policy pol
    WHERE (pol.product_id IS NULL OR pol.product_id = ps.product_id)
      AND (pol.store_id IS NULL OR pol.store_id = ps.store_id)
      AND (pol.category_id IS NULL OR pol.category_id = p.category_id)
    ORDER BY
      CASE
        WHEN pol.product_id IS NOT NULL AND pol.store_id IS NOT NULL THEN 1
        WHEN pol.product_id IS NOT NULL AND pol.store_id IS NULL THEN 2
        WHEN pol.category_id IS NOT NULL AND pol.store_id IS NOT NULL AND pol.product_id IS NULL THEN 3
        WHEN pol.category_id IS NOT NULL AND pol.store_id IS NULL AND pol.product_id IS NULL THEN 4
        WHEN pol.store_id IS NOT NULL AND pol.product_id IS NULL AND pol.category_id IS NULL THEN 5
        WHEN pol.product_id IS NULL AND pol.category_id IS NULL AND pol.store_id IS NULL THEN 6
        ELSE 99
      END,
      pol.id
    LIMIT 1
  ) pol ON TRUE
),
calc AS (
  SELECT
    b.*,
    CASE
      WHEN COALESCE(b.stock_quantity, 0) = 0 THEN 0
      WHEN COALESCE(b.avg_daily_qty, 0) = 0 THEN 9999
      ELSE ROUND((b.stock_quantity / NULLIF(b.avg_daily_qty, 0))::numeric, 3)
    END AS days_of_cover,
    ROUND((
      COALESCE(b.service_z, 1.65)
      * COALESCE(b.demand_std_daily, 0)
      * SQRT(GREATEST(COALESCE(b.lead_time_days, 7) + COALESCE(b.review_period_days, 7), 0))
    )::numeric, 3) AS safety_stock
  FROM base b
),
calc2 AS (
  SELECT
    c.*,
    ROUND((
      COALESCE(c.avg_daily_qty, 0) * (COALESCE(c.lead_time_days, 7) + COALESCE(c.review_period_days, 7))
      + COALESCE(c.safety_stock, 0)
    )::numeric, 3) AS reorder_point,
    CASE
      WHEN COALESCE(c.max_qty, 0) > 0 THEN
        LEAST(
          GREATEST(COALESCE(c.avg_daily_qty, 0) * COALESCE(c.target_doc_days, 21), COALESCE(c.min_qty, 0)),
          c.max_qty
        )
      ELSE
        GREATEST(COALESCE(c.avg_daily_qty, 0) * COALESCE(c.target_doc_days, 21), COALESCE(c.min_qty, 0))
    END AS target_qty
  FROM calc c
),
calc3 AS (
  SELECT
    c.*,
    GREATEST(0, COALESCE(c.target_qty, 0) - COALESCE(c.stock_quantity, 0)) AS raw_need
  FROM calc2 c
),
calc4 AS (
  SELECT
    c.*,
    CASE
      WHEN c.raw_need > 0 THEN
        CEIL(GREATEST(c.raw_need, COALESCE(NULLIF(c.moq, 0), 1)) / NULLIF(GREATEST(COALESCE(c.pack_size, 1), 1), 0))
        * GREATEST(COALESCE(c.pack_size, 1), 1)
      ELSE 0
    END AS reorder_qty
  FROM calc3 c
)
SELECT
  id,
  store_id,
  product_id,
  ean,
  product_name,
  category_id,
  supplier_id,
  ROUND(COALESCE(stock_quantity, 0)::numeric, 3) AS stock_quantity,
  ROUND(COALESCE(stock_sum, 0)::numeric, 3) AS stock_sum,
  ROUND(COALESCE(avg_daily_qty, 0)::numeric, 3) AS avg_daily_qty,
  ROUND(COALESCE(demand_std_daily, 0)::numeric, 3) AS demand_std_daily,
  ROUND(COALESCE(days_of_cover, 0)::numeric, 3) AS days_of_cover,
  ROUND(COALESCE(lead_time_days, 7)::numeric, 3) AS lead_time_days,
  ROUND(COALESCE(review_period_days, 7)::numeric, 3) AS review_period_days,
  ROUND(COALESCE(target_doc_days, 21)::numeric, 3) AS target_doc_days,
  ROUND(COALESCE(moq, 12)::numeric, 3) AS moq,
  ROUND(COALESCE(pack_size, 12)::numeric, 3) AS pack_size,
  ROUND(COALESCE(service_z, 1.65)::numeric, 3) AS service_z,
  ROUND(COALESCE(unit_cost, 0)::numeric, 3) AS unit_cost,
  ROUND(COALESCE(safety_stock, 0)::numeric, 3) AS safety_stock,
  ROUND(COALESCE(reorder_point, 0)::numeric, 3) AS reorder_point,
  ROUND(COALESCE(target_qty, 0)::numeric, 3) AS target_qty,
  ROUND(COALESCE(reorder_qty, 0)::numeric, 3) AS reorder_qty,
  ROUND((COALESCE(reorder_qty, 0) * COALESCE(unit_cost, 0))::numeric, 3) AS order_value,
  CASE
    WHEN COALESCE(stock_quantity, 0) = 0
      OR COALESCE(days_of_cover, 0) < (COALESCE(lead_time_days, 7) + COALESCE(review_period_days, 7))
      THEN 'critical'
    WHEN COALESCE(days_of_cover, 0) < COALESCE(target_doc_days, 21) THEN 'understock'
    WHEN COALESCE(days_of_cover, 0) > COALESCE(target_doc_days, 21) * 2 THEN 'overstock'
    ELSE 'ok'
  END AS health_level,
  stock_date
FROM calc4;

CREATE UNIQUE INDEX stock_reorder_fact_id_idx ON stock_reorder_fact (id);
CREATE INDEX stock_reorder_fact_health_idx ON stock_reorder_fact (health_level);
CREATE INDEX stock_reorder_fact_store_idx ON stock_reorder_fact (store_id);
CREATE INDEX stock_reorder_fact_supplier_idx ON stock_reorder_fact (supplier_id);

ALTER TABLE stock_reorder_fact
  ALTER COLUMN stock_quantity TYPE numeric(15,3),
  ALTER COLUMN stock_sum TYPE numeric(15,3),
  ALTER COLUMN avg_daily_qty TYPE numeric(15,3),
  ALTER COLUMN demand_std_daily TYPE numeric(15,3),
  ALTER COLUMN days_of_cover TYPE numeric(15,3),
  ALTER COLUMN lead_time_days TYPE numeric(15,3),
  ALTER COLUMN review_period_days TYPE numeric(15,3),
  ALTER COLUMN target_doc_days TYPE numeric(15,3),
  ALTER COLUMN moq TYPE numeric(15,3),
  ALTER COLUMN pack_size TYPE numeric(15,3),
  ALTER COLUMN service_z TYPE numeric(15,3),
  ALTER COLUMN unit_cost TYPE numeric(15,3),
  ALTER COLUMN safety_stock TYPE numeric(15,3),
  ALTER COLUMN reorder_point TYPE numeric(15,3),
  ALTER COLUMN target_qty TYPE numeric(15,3),
  ALTER COLUMN reorder_qty TYPE numeric(15,3),
  ALTER COLUMN order_value TYPE numeric(15,3),
  ALTER COLUMN store_id TYPE numeric(15,3),
  ALTER COLUMN product_id TYPE numeric(15,3),
  ALTER COLUMN category_id TYPE numeric(15,3),
  ALTER COLUMN supplier_id TYPE numeric(15,3);

-- ---------------------------------------------------------------------------
-- stock_health_slice
-- ---------------------------------------------------------------------------
CREATE TABLE stock_health_slice AS
WITH u AS (
  SELECT
    'store'::text AS slice_kind,
    s.store_id::text AS slice_key,
    COALESCE(s.store_name, 'store '||f.store_id::text) AS slice_label,
    COUNT(*)::numeric AS sku_count,
    SUM(f.stock_quantity) AS stock_quantity,
    SUM(f.stock_sum) AS stock_sum,
    AVG(f.avg_daily_qty) AS avg_daily_qty,
    AVG(f.days_of_cover) AS days_of_cover,
    SUM(f.reorder_qty) AS reorder_qty,
    SUM(f.order_value) AS order_value,
    COUNT(*) FILTER (WHERE f.health_level = 'critical')::numeric AS critical_count,
    COUNT(*) FILTER (WHERE f.health_level = 'understock')::numeric AS understock_count
  FROM stock_reorder_fact f
  LEFT JOIN stores s ON s.store_id = f.store_id
  GROUP BY s.store_id, s.store_name, f.store_id

  UNION ALL

  SELECT
    'category',
    COALESCE(cat.category_id::text, f.category_id::text),
    COALESCE(cat.category_name, 'category '||COALESCE(f.category_id::text, '?')),
    COUNT(*)::numeric,
    SUM(f.stock_quantity),
    SUM(f.stock_sum),
    AVG(f.avg_daily_qty),
    AVG(f.days_of_cover),
    SUM(f.reorder_qty),
    SUM(f.order_value),
    COUNT(*) FILTER (WHERE f.health_level = 'critical')::numeric,
    COUNT(*) FILTER (WHERE f.health_level = 'understock')::numeric
  FROM stock_reorder_fact f
  LEFT JOIN categories cat ON cat.category_id = f.category_id
  GROUP BY cat.category_id, cat.category_name, f.category_id

  UNION ALL

  SELECT
    'health',
    f.health_level,
    f.health_level,
    COUNT(*)::numeric,
    SUM(f.stock_quantity),
    SUM(f.stock_sum),
    AVG(f.avg_daily_qty),
    AVG(f.days_of_cover),
    SUM(f.reorder_qty),
    SUM(f.order_value),
    COUNT(*) FILTER (WHERE f.health_level = 'critical')::numeric,
    COUNT(*) FILTER (WHERE f.health_level = 'understock')::numeric
  FROM stock_reorder_fact f
  GROUP BY f.health_level

  UNION ALL

  SELECT * FROM (
    SELECT
      'understock_sku'::text,
      f.product_id::text || '@' || f.store_id::text,
      COALESCE(f.product_name, 'SKU '||f.product_id::text)
        || ' · ' || COALESCE(s.store_name, 'store '||f.store_id::text),
      1::numeric,
      f.stock_quantity,
      f.stock_sum,
      f.avg_daily_qty,
      f.days_of_cover,
      f.reorder_qty,
      f.order_value,
      CASE WHEN f.health_level = 'critical' THEN 1 ELSE 0 END::numeric,
      CASE WHEN f.health_level = 'understock' THEN 1 ELSE 0 END::numeric
    FROM stock_reorder_fact f
    LEFT JOIN stores s ON s.store_id = f.store_id
    WHERE f.health_level IN ('understock', 'critical')
    ORDER BY f.days_of_cover ASC NULLS LAST
    LIMIT 100
  ) understock_sku

  UNION ALL

  SELECT
    'kpi',
    'all',
    'KPI',
    COUNT(*) FILTER (WHERE health_level IN ('understock', 'critical'))::numeric,
    SUM(stock_quantity),
    SUM(stock_sum),
    AVG(avg_daily_qty),
    AVG(days_of_cover),
    SUM(reorder_qty),
    SUM(order_value),
    COUNT(*) FILTER (WHERE health_level = 'critical')::numeric,
    COUNT(*) FILTER (WHERE health_level = 'understock')::numeric
  FROM stock_reorder_fact

  UNION ALL

  SELECT * FROM (
    SELECT
      'category_order_top'::text,
      COALESCE(cat.category_id::text, f.category_id::text),
      COALESCE(cat.category_name, 'category '||COALESCE(f.category_id::text, '?')),
      COUNT(*)::numeric,
      SUM(f.stock_quantity),
      SUM(f.stock_sum),
      AVG(f.avg_daily_qty),
      AVG(f.days_of_cover),
      SUM(f.reorder_qty),
      SUM(f.order_value),
      COUNT(*) FILTER (WHERE f.health_level = 'critical')::numeric,
      COUNT(*) FILTER (WHERE f.health_level = 'understock')::numeric
    FROM stock_reorder_fact f
    LEFT JOIN categories cat ON cat.category_id = f.category_id
    WHERE f.reorder_qty > 0
    GROUP BY cat.category_id, cat.category_name, f.category_id
    ORDER BY SUM(f.order_value) DESC
    LIMIT 12
  ) category_order_top
)
SELECT
  row_number() OVER (ORDER BY slice_kind, slice_label)::numeric AS id,
  slice_kind,
  slice_key,
  slice_label,
  ROUND(sku_count::numeric, 3) AS sku_count,
  ROUND(COALESCE(stock_quantity, 0)::numeric, 3) AS stock_quantity,
  ROUND(COALESCE(stock_sum, 0)::numeric, 3) AS stock_sum,
  ROUND(COALESCE(avg_daily_qty, 0)::numeric, 3) AS avg_daily_qty,
  ROUND(COALESCE(days_of_cover, 0)::numeric, 3) AS days_of_cover,
  ROUND(COALESCE(reorder_qty, 0)::numeric, 3) AS reorder_qty,
  ROUND(COALESCE(order_value, 0)::numeric, 3) AS order_value,
  ROUND(COALESCE(critical_count, 0)::numeric, 3) AS critical_count,
  ROUND(COALESCE(understock_count, 0)::numeric, 3) AS understock_count
FROM u;

CREATE UNIQUE INDEX stock_health_slice_id_idx ON stock_health_slice (id);
CREATE INDEX stock_health_slice_kind_idx ON stock_health_slice (slice_kind);

ALTER TABLE stock_health_slice
  ALTER COLUMN sku_count TYPE numeric(15,3),
  ALTER COLUMN stock_quantity TYPE numeric(15,3),
  ALTER COLUMN stock_sum TYPE numeric(15,3),
  ALTER COLUMN avg_daily_qty TYPE numeric(15,3),
  ALTER COLUMN days_of_cover TYPE numeric(15,3),
  ALTER COLUMN reorder_qty TYPE numeric(15,3),
  ALTER COLUMN order_value TYPE numeric(15,3),
  ALTER COLUMN critical_count TYPE numeric(15,3),
  ALTER COLUMN understock_count TYPE numeric(15,3);

-- ---------------------------------------------------------------------------
-- purchase_order + purchase_order_line (submitted, per store × supplier)
-- ---------------------------------------------------------------------------
CREATE TABLE purchase_order AS
WITH need AS (
  SELECT *
  FROM stock_reorder_fact
  WHERE reorder_qty > 0
    AND COALESCE(supplier_id, 0) > 0
),
hdr AS (
  SELECT
    row_number() OVER (ORDER BY supplier_id, store_id)::numeric AS id,
    supplier_id,
    store_id,
    SUM(reorder_qty) AS total_qty,
    SUM(order_value) AS total_sum,
    COUNT(*)::numeric AS line_count,
    MIN(stock_date)::date AS order_date
  FROM need
  GROUP BY supplier_id, store_id
)
SELECT
  id,
  ('PO-' || store_id::bigint || '-' || supplier_id::bigint || '-' || to_char(COALESCE(order_date, CURRENT_DATE), 'YYYYMMDD'))::text AS po_number,
  supplier_id,
  store_id,
  'submitted'::text AS status,
  COALESCE(order_date, CURRENT_DATE)::timestamp AS order_date,
  (COALESCE(order_date, CURRENT_DATE) + INTERVAL '7 days')::timestamp AS expected_date,
  ROUND(total_qty::numeric, 3) AS total_qty,
  ROUND(total_sum::numeric, 3) AS total_sum,
  'rulesgo'::text AS source,
  ROUND(line_count::numeric, 3) AS line_count
FROM hdr;

CREATE UNIQUE INDEX purchase_order_id_idx ON purchase_order (id);
CREATE INDEX purchase_order_store_idx ON purchase_order (store_id);
CREATE INDEX purchase_order_supplier_idx ON purchase_order (supplier_id);

ALTER TABLE purchase_order
  ALTER COLUMN supplier_id TYPE numeric(15,3),
  ALTER COLUMN store_id TYPE numeric(15,3),
  ALTER COLUMN total_qty TYPE numeric(15,3),
  ALTER COLUMN total_sum TYPE numeric(15,3),
  ALTER COLUMN line_count TYPE numeric(15,3);

CREATE TABLE purchase_order_line AS
WITH need AS (
  SELECT f.*,
         ROW_NUMBER() OVER (ORDER BY f.order_value DESC NULLS LAST, f.id) AS rn
  FROM stock_reorder_fact f
  WHERE f.reorder_qty > 0
    AND COALESCE(f.supplier_id, 0) > 0
),
capped AS (
  SELECT * FROM need WHERE rn <= 5000
)
SELECT
  ROW_NUMBER() OVER (ORDER BY po.id, c.product_id)::numeric AS id,
  po.id AS po_id,
  c.product_id,
  c.ean,
  c.product_name,
  c.store_id,
  ROUND(c.reorder_qty::numeric, 3) AS qty_ordered,
  ROUND(c.reorder_qty::numeric, 3) AS qty_suggested,
  ROUND(c.unit_cost::numeric, 3) AS unit_cost,
  ROUND(c.order_value::numeric, 3) AS line_sum,
  ROUND(c.reorder_point::numeric, 3) AS reorder_point,
  ROUND(c.days_of_cover::numeric, 3) AS days_of_cover,
  c.health_level,
  0::numeric AS rule_score
FROM capped c
JOIN purchase_order po
  ON po.supplier_id = c.supplier_id
 AND po.store_id = c.store_id;

CREATE UNIQUE INDEX purchase_order_line_id_idx ON purchase_order_line (id);
CREATE INDEX purchase_order_line_po_idx ON purchase_order_line (po_id);

ALTER TABLE purchase_order_line
  ALTER COLUMN po_id TYPE numeric(15,3),
  ALTER COLUMN product_id TYPE numeric(15,3),
  ALTER COLUMN store_id TYPE numeric(15,3),
  ALTER COLUMN qty_ordered TYPE numeric(15,3),
  ALTER COLUMN qty_suggested TYPE numeric(15,3),
  ALTER COLUMN unit_cost TYPE numeric(15,3),
  ALTER COLUMN line_sum TYPE numeric(15,3),
  ALTER COLUMN reorder_point TYPE numeric(15,3),
  ALTER COLUMN days_of_cover TYPE numeric(15,3),
  ALTER COLUMN rule_score TYPE numeric(15,3);
