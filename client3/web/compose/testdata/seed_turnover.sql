-- Inventory turnover analysis: fact + slice for charts/KPI.
-- Depends on: stock_reorder_fact, receipt_margin (or receipt_positions), stores.

DROP TABLE IF EXISTS turnover_slice CASCADE;
DROP TABLE IF EXISTS turnover_fact CASCADE;

CREATE TABLE turnover_fact AS
WITH bounds AS (
  SELECT COALESCE(MAX(dt)::date, CURRENT_DATE) AS as_of FROM receipt_margin
),
sales_30 AS (
  SELECT
    rm.store_id,
    rm.product_id,
    SUM(rm.quantity)::numeric AS qty_30d,
    SUM(rm.revenue)::numeric AS revenue_30d,
    SUM(rm.cogs)::numeric AS cogs_30d,
    MAX(rm.category_name) AS category_name,
    MAX(rm.brand) AS brand
  FROM receipt_margin rm
  CROSS JOIN bounds b
  WHERE rm.dt::date >= b.as_of - 29
    AND rm.dt::date <= b.as_of
  GROUP BY rm.store_id, rm.product_id
),
base AS (
  SELECT
    f.id,
    f.store_id,
    COALESCE(st.store_name, 'store ' || f.store_id::text) AS store_name,
    f.product_id,
    COALESCE(f.product_name, 'SKU ' || f.product_id::text) AS product_name,
    f.category_id,
    COALESCE(s.category_name, 'Категория ' || COALESCE(f.category_id::text, '?')) AS category_name,
    COALESCE(s.brand, '') AS brand,
    COALESCE(f.stock_quantity, 0)::numeric AS stock_quantity,
    COALESCE(f.stock_sum, 0)::numeric AS stock_sum,
    COALESCE(f.avg_daily_qty, 0)::numeric AS avg_daily_qty,
    COALESCE(f.days_of_cover, 0)::numeric AS days_of_cover,
    COALESCE(f.health_level, 'ok') AS health_level,
    COALESCE(s.qty_30d, 0)::numeric AS qty_30d,
    COALESCE(s.revenue_30d, 0)::numeric AS revenue_30d,
    COALESCE(s.cogs_30d, 0)::numeric AS cogs_30d
  FROM stock_reorder_fact f
  LEFT JOIN stores st ON st.store_id = f.store_id
  LEFT JOIN sales_30 s ON s.store_id = f.store_id AND s.product_id = f.product_id
)
SELECT
  id,
  store_id,
  store_name,
  product_id,
  product_name,
  category_id,
  category_name,
  brand,
  stock_quantity,
  stock_sum,
  avg_daily_qty,
  LEAST(days_of_cover, 9999)::numeric AS days_of_cover,
  health_level,
  qty_30d,
  revenue_30d,
  cogs_30d,
  -- turns over 30 days: sold / on-hand
  ROUND(
    CASE WHEN stock_quantity > 0 THEN qty_30d / stock_quantity ELSE 0 END
  , 3) AS turns_30d,
  -- annualized inventory turns from daily demand
  ROUND(
    CASE WHEN stock_quantity > 0 THEN (avg_daily_qty * 365.0) / stock_quantity ELSE 0 END
  , 2) AS turns_year,
  -- DOI capped for charts
  ROUND(LEAST(
    CASE WHEN avg_daily_qty > 0 THEN stock_quantity / avg_daily_qty ELSE 9999 END,
    9999
  )::numeric, 1) AS doi_days,
  CASE
    WHEN COALESCE(qty_30d, 0) = 0 AND COALESCE(stock_quantity, 0) > 0 THEN 'dead'
    WHEN LEAST(days_of_cover, 9999) > 90 THEN 'slow'
    WHEN LEAST(days_of_cover, 9999) > 30 THEN 'normal'
    WHEN LEAST(days_of_cover, 9999) >= 7 THEN 'fast'
    ELSE 'rush'  -- very low cover / high velocity risk
  END AS velocity
FROM base;

CREATE INDEX IF NOT EXISTS turnover_fact_velocity_idx ON turnover_fact (velocity);
CREATE INDEX IF NOT EXISTS turnover_fact_store_idx ON turnover_fact (store_id);
CREATE INDEX IF NOT EXISTS turnover_fact_cat_idx ON turnover_fact (category_id);

CREATE TABLE turnover_slice AS
WITH u AS (
  SELECT
    'kpi'::text AS slice_kind,
    'all'::text AS slice_key,
    'KPI'::text AS slice_label,
    COUNT(*)::numeric AS sku_count,
    COUNT(*) FILTER (WHERE velocity = 'dead')::numeric AS dead_count,
    COUNT(*) FILTER (WHERE velocity = 'slow')::numeric AS slow_count,
    COUNT(*) FILTER (WHERE velocity = 'normal')::numeric AS normal_count,
    COUNT(*) FILTER (WHERE velocity = 'fast')::numeric AS fast_count,
    COUNT(*) FILTER (WHERE velocity = 'rush')::numeric AS rush_count,
    ROUND(AVG(turns_year) FILTER (WHERE stock_quantity > 0 AND turns_year < 500), 2) AS avg_turns,
    ROUND(AVG(doi_days) FILTER (WHERE stock_quantity > 0 AND doi_days < 365), 1) AS avg_doi,
    ROUND(SUM(stock_sum)::numeric, 2) AS stock_value,
    ROUND(SUM(revenue_30d)::numeric, 2) AS revenue_30d,
    ROUND(SUM(stock_sum) FILTER (WHERE velocity IN ('dead', 'slow'))::numeric, 2) AS frozen_value,
    COUNT(*) FILTER (WHERE velocity IN ('dead', 'slow'))::numeric AS problem_count,
    0::numeric AS sort_score
  FROM turnover_fact

  UNION ALL

  SELECT
    'velocity',
    velocity,
    CASE velocity
      WHEN 'dead' THEN 'Мёртвый остаток'
      WHEN 'slow' THEN 'Медленная'
      WHEN 'normal' THEN 'Норма'
      WHEN 'fast' THEN 'Быстрая'
      WHEN 'rush' THEN 'Дефицит / срочно'
      ELSE velocity
    END,
    COUNT(*)::numeric,
    COUNT(*) FILTER (WHERE velocity = 'dead')::numeric,
    COUNT(*) FILTER (WHERE velocity = 'slow')::numeric,
    COUNT(*) FILTER (WHERE velocity = 'normal')::numeric,
    COUNT(*) FILTER (WHERE velocity = 'fast')::numeric,
    COUNT(*) FILTER (WHERE velocity = 'rush')::numeric,
    ROUND(AVG(turns_year) FILTER (WHERE turns_year < 500), 2),
    ROUND(AVG(doi_days) FILTER (WHERE doi_days < 365), 1),
    ROUND(SUM(stock_sum), 2),
    ROUND(SUM(revenue_30d), 2),
    ROUND(SUM(stock_sum) FILTER (WHERE velocity IN ('dead', 'slow')), 2),
    COUNT(*) FILTER (WHERE velocity IN ('dead', 'slow'))::numeric,
    CASE velocity
      WHEN 'rush' THEN 1 WHEN 'fast' THEN 2 WHEN 'normal' THEN 3
      WHEN 'slow' THEN 4 WHEN 'dead' THEN 5 ELSE 9
    END::numeric
  FROM turnover_fact
  GROUP BY velocity

  UNION ALL

  SELECT
    'store',
    store_id::text,
    store_name,
    COUNT(*)::numeric,
    COUNT(*) FILTER (WHERE velocity = 'dead')::numeric,
    COUNT(*) FILTER (WHERE velocity = 'slow')::numeric,
    COUNT(*) FILTER (WHERE velocity = 'normal')::numeric,
    COUNT(*) FILTER (WHERE velocity = 'fast')::numeric,
    COUNT(*) FILTER (WHERE velocity = 'rush')::numeric,
    ROUND(AVG(turns_year) FILTER (WHERE stock_quantity > 0 AND turns_year < 500), 2),
    ROUND(AVG(doi_days) FILTER (WHERE stock_quantity > 0 AND doi_days < 365), 1),
    ROUND(SUM(stock_sum), 2),
    ROUND(SUM(revenue_30d), 2),
    ROUND(SUM(stock_sum) FILTER (WHERE velocity IN ('dead', 'slow')), 2),
    COUNT(*) FILTER (WHERE velocity IN ('dead', 'slow'))::numeric,
    ROUND(AVG(turns_year) FILTER (WHERE stock_quantity > 0 AND turns_year < 500), 2)
  FROM turnover_fact
  GROUP BY store_id, store_name

  UNION ALL

  SELECT
    'category',
    COALESCE(category_id::text, category_name),
    category_name,
    COUNT(*)::numeric,
    COUNT(*) FILTER (WHERE velocity = 'dead')::numeric,
    COUNT(*) FILTER (WHERE velocity = 'slow')::numeric,
    COUNT(*) FILTER (WHERE velocity = 'normal')::numeric,
    COUNT(*) FILTER (WHERE velocity = 'fast')::numeric,
    COUNT(*) FILTER (WHERE velocity = 'rush')::numeric,
    ROUND(AVG(turns_year) FILTER (WHERE stock_quantity > 0 AND turns_year < 500), 2),
    ROUND(AVG(doi_days) FILTER (WHERE stock_quantity > 0 AND doi_days < 365), 1),
    ROUND(SUM(stock_sum), 2),
    ROUND(SUM(revenue_30d), 2),
    ROUND(SUM(stock_sum) FILTER (WHERE velocity IN ('dead', 'slow')), 2),
    COUNT(*) FILTER (WHERE velocity IN ('dead', 'slow'))::numeric,
    ROUND(AVG(doi_days) FILTER (WHERE stock_quantity > 0 AND doi_days < 365), 1)
  FROM turnover_fact
  GROUP BY category_id, category_name

  UNION ALL

  SELECT
    'store_frozen',
    store_id::text,
    store_name,
    COUNT(*) FILTER (WHERE velocity IN ('dead', 'slow'))::numeric,
    COUNT(*) FILTER (WHERE velocity = 'dead')::numeric,
    COUNT(*) FILTER (WHERE velocity = 'slow')::numeric,
    0, 0, 0,
    ROUND(AVG(turns_year) FILTER (WHERE velocity IN ('dead', 'slow') AND turns_year < 500), 2),
    ROUND(AVG(doi_days) FILTER (WHERE velocity IN ('dead', 'slow') AND doi_days < 365), 1),
    ROUND(SUM(stock_sum) FILTER (WHERE velocity IN ('dead', 'slow')), 2),
    ROUND(SUM(revenue_30d), 2),
    ROUND(SUM(stock_sum) FILTER (WHERE velocity IN ('dead', 'slow')), 2),
    COUNT(*) FILTER (WHERE velocity IN ('dead', 'slow'))::numeric,
    ROUND(SUM(stock_sum) FILTER (WHERE velocity IN ('dead', 'slow')), 2)
  FROM turnover_fact
  GROUP BY store_id, store_name

  UNION ALL

  SELECT
    'cat_doi',
    COALESCE(category_id::text, category_name),
    category_name,
    COUNT(*)::numeric,
    0, 0, 0, 0, 0,
    ROUND(AVG(turns_year) FILTER (WHERE stock_quantity > 0 AND turns_year < 500), 2),
    ROUND(AVG(doi_days) FILTER (WHERE stock_quantity > 0 AND doi_days < 365), 1),
    ROUND(SUM(stock_sum), 2),
    ROUND(SUM(revenue_30d), 2),
    ROUND(SUM(stock_sum) FILTER (WHERE velocity IN ('dead', 'slow')), 2),
    COUNT(*) FILTER (WHERE velocity IN ('dead', 'slow'))::numeric,
    ROUND(AVG(doi_days) FILTER (WHERE stock_quantity > 0 AND doi_days < 365), 1)
  FROM turnover_fact
  GROUP BY category_id, category_name
)
SELECT
  row_number() OVER (ORDER BY slice_kind, sort_score DESC NULLS LAST, slice_label)::numeric AS id,
  slice_kind,
  slice_key,
  slice_label,
  ROUND(COALESCE(sku_count, 0), 1) AS sku_count,
  ROUND(COALESCE(dead_count, 0), 1) AS dead_count,
  ROUND(COALESCE(slow_count, 0), 1) AS slow_count,
  ROUND(COALESCE(normal_count, 0), 1) AS normal_count,
  ROUND(COALESCE(fast_count, 0), 1) AS fast_count,
  ROUND(COALESCE(rush_count, 0), 1) AS rush_count,
  ROUND(COALESCE(avg_turns, 0), 2) AS avg_turns,
  ROUND(COALESCE(avg_doi, 0), 1) AS avg_doi,
  ROUND(COALESCE(stock_value, 0), 2) AS stock_value,
  ROUND(COALESCE(revenue_30d, 0), 2) AS revenue_30d,
  ROUND(COALESCE(frozen_value, 0), 2) AS frozen_value,
  ROUND(COALESCE(problem_count, 0), 1) AS problem_count,
  ROUND(COALESCE(sort_score, 0), 2) AS sort_score
FROM u;

CREATE INDEX IF NOT EXISTS turnover_slice_kind_idx ON turnover_slice (slice_kind);
