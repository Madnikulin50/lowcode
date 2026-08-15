-- Risk dashboard slices: stores, products/stock, incidents.
-- Depends on: stock_reorder_fact, incidents, stores.

DROP TABLE IF EXISTS risk_slice CASCADE;

CREATE TABLE risk_slice AS
WITH stock_by_store AS (
  SELECT
    f.store_id,
    COUNT(*)::numeric AS sku_count,
    COUNT(*) FILTER (WHERE f.health_level = 'critical')::numeric AS stock_critical_count,
    COUNT(*) FILTER (WHERE f.health_level = 'understock')::numeric AS stock_understock_count,
    COUNT(*) FILTER (WHERE f.health_level IN ('critical', 'understock'))::numeric AS stock_risk_count,
    AVG(f.days_of_cover) AS days_of_cover,
    SUM(f.reorder_qty) AS reorder_qty,
    SUM(f.order_value) AS order_value
  FROM stock_reorder_fact f
  GROUP BY f.store_id
),
inc_by_store AS (
  SELECT
    i.store_id,
    COUNT(*)::numeric AS incident_count,
    COUNT(*) FILTER (WHERE i.incident_status IN ('Open', 'In Progress', 'Escalated'))::numeric AS open_incident_count,
    COUNT(*) FILTER (WHERE i.criticality = 'Critical')::numeric AS critical_incident_count,
    COUNT(*) FILTER (WHERE i.incident_status = 'Escalated')::numeric AS escalated_count
  FROM incidents i
  GROUP BY i.store_id
),
store_union AS (
  SELECT store_id FROM stock_by_store
  UNION
  SELECT store_id FROM inc_by_store
),
u AS (
  SELECT
    'kpi'::text AS slice_kind,
    'all'::text AS slice_key,
    'KPI'::text AS slice_label,
    (
      (SELECT COUNT(*) FROM incidents WHERE incident_status IN ('Open', 'In Progress', 'Escalated'))
      + (SELECT COUNT(*) FROM stock_reorder_fact WHERE health_level IN ('critical', 'understock'))
    )::numeric AS risk_score,
    (SELECT COUNT(*)::numeric FROM incidents) AS incident_count,
    (SELECT COUNT(*)::numeric FROM incidents WHERE incident_status IN ('Open', 'In Progress', 'Escalated')) AS open_incident_count,
    (SELECT COUNT(*)::numeric FROM incidents WHERE criticality = 'Critical'
      AND incident_status IN ('Open', 'In Progress', 'Escalated')) AS critical_incident_count,
    (SELECT COUNT(*)::numeric FROM incidents WHERE incident_status = 'Escalated') AS escalated_count,
    (SELECT COUNT(*)::numeric FROM stock_reorder_fact WHERE health_level = 'critical') AS stock_critical_count,
    (SELECT COUNT(*)::numeric FROM stock_reorder_fact WHERE health_level = 'understock') AS stock_understock_count,
    (SELECT COUNT(*)::numeric FROM stock_reorder_fact WHERE health_level IN ('critical', 'understock')) AS stock_risk_count,
    (SELECT COUNT(DISTINCT store_id)::numeric FROM (
        SELECT store_id FROM stock_reorder_fact WHERE health_level IN ('critical', 'understock')
        UNION
        SELECT store_id FROM incidents WHERE incident_status IN ('Open', 'In Progress', 'Escalated')
      ) s
    ) AS store_at_risk_count,
    0::numeric AS sku_count,
    (SELECT AVG(days_of_cover)::numeric FROM stock_reorder_fact) AS days_of_cover,
    (SELECT SUM(reorder_qty)::numeric FROM stock_reorder_fact) AS reorder_qty,
    (SELECT SUM(order_value)::numeric FROM stock_reorder_fact) AS order_value

  UNION ALL

  SELECT
    'store',
    su.store_id::text,
    COALESCE(st.store_name, 'store '||su.store_id::text),
    (
      COALESCE(i.open_incident_count, 0) * 3
      + COALESCE(i.critical_incident_count, 0) * 5
      + COALESCE(i.escalated_count, 0) * 4
      + COALESCE(s.stock_critical_count, 0) * 4
      + COALESCE(s.stock_understock_count, 0) * 2
    ),
    COALESCE(i.incident_count, 0),
    COALESCE(i.open_incident_count, 0),
    COALESCE(i.critical_incident_count, 0),
    COALESCE(i.escalated_count, 0),
    COALESCE(s.stock_critical_count, 0),
    COALESCE(s.stock_understock_count, 0),
    COALESCE(s.stock_risk_count, 0),
    CASE WHEN COALESCE(i.open_incident_count, 0) > 0
           OR COALESCE(s.stock_risk_count, 0) > 0 THEN 1 ELSE 0 END::numeric,
    COALESCE(s.sku_count, 0),
    COALESCE(s.days_of_cover, 0),
    COALESCE(s.reorder_qty, 0),
    COALESCE(s.order_value, 0)
  FROM store_union su
  LEFT JOIN stores st ON st.store_id = su.store_id
  LEFT JOIN stock_by_store s ON s.store_id = su.store_id
  LEFT JOIN inc_by_store i ON i.store_id = su.store_id

  UNION ALL

  SELECT * FROM (
    SELECT
      'product'::text,
      f.product_id::text || '@' || f.store_id::text,
      COALESCE(f.product_name, 'SKU '||f.product_id::text)
        || ' · ' || COALESCE(st.store_name, 'store '||f.store_id::text),
      (
        CASE WHEN f.health_level = 'critical' THEN 40 ELSE 20 END
        + GREATEST(0, 21 - COALESCE(f.days_of_cover, 0))
      )::numeric,
      0::numeric, 0::numeric, 0::numeric, 0::numeric,
      CASE WHEN f.health_level = 'critical' THEN 1 ELSE 0 END::numeric,
      CASE WHEN f.health_level = 'understock' THEN 1 ELSE 0 END::numeric,
      1::numeric,
      1::numeric,
      1::numeric,
      f.days_of_cover,
      f.reorder_qty,
      f.order_value
    FROM stock_reorder_fact f
    LEFT JOIN stores st ON st.store_id = f.store_id
    WHERE f.health_level IN ('critical', 'understock')
    ORDER BY
      CASE WHEN f.health_level = 'critical' THEN 0 ELSE 1 END,
      f.days_of_cover ASC NULLS LAST
    LIMIT 20
  ) product_risk

  UNION ALL

  SELECT
    'stock_health',
    f.health_level,
    CASE f.health_level
      WHEN 'critical' THEN 'Критический остаток'
      WHEN 'understock' THEN 'Недосток'
      WHEN 'overstock' THEN 'Избыток'
      WHEN 'ok' THEN 'Норма'
      ELSE f.health_level
    END,
    COUNT(*)::numeric,
    0, 0, 0, 0,
    COUNT(*) FILTER (WHERE f.health_level = 'critical')::numeric,
    COUNT(*) FILTER (WHERE f.health_level = 'understock')::numeric,
    COUNT(*) FILTER (WHERE f.health_level IN ('critical', 'understock'))::numeric,
    0::numeric,
    COUNT(*)::numeric,
    AVG(f.days_of_cover),
    SUM(f.reorder_qty),
    SUM(f.order_value)
  FROM stock_reorder_fact f
  GROUP BY f.health_level

  UNION ALL

  SELECT
    'incident_type',
    COALESCE(i.incident_type, 'unknown'),
    CASE COALESCE(i.incident_type, 'unknown')
      WHEN 'Power Outage' THEN 'Отключение электроэнергии'
      WHEN 'Spillage' THEN 'Затопление'
      WHEN 'Refrigeration Outage' THEN 'Поломка холодильников'
      WHEN 'Equipment Damage' THEN 'Повреждение снаряжения'
      WHEN 'Security Breach' THEN 'Нарушение безопасности'
      WHEN 'Network Outage' THEN 'Отключение интернет'
      WHEN 'Stock Discrepancy' THEN 'Расхождение остатков'
      WHEN 'Fire Alarm' THEN 'Пожарная тревога'
      WHEN 'Customer Complaint' THEN 'Жалоба клиента'
      WHEN 'POS Failure' THEN 'Отключение POS'
      ELSE COALESCE(i.incident_type, 'unknown')
    END,
    COUNT(*)::numeric,
    COUNT(*)::numeric,
    COUNT(*) FILTER (WHERE i.incident_status IN ('Open', 'In Progress', 'Escalated'))::numeric,
    COUNT(*) FILTER (WHERE i.criticality = 'Critical')::numeric,
    COUNT(*) FILTER (WHERE i.incident_status = 'Escalated')::numeric,
    0, 0, 0, 0, 0, 0, 0, 0
  FROM incidents i
  GROUP BY i.incident_type

  UNION ALL

  SELECT
    'incident_criticality',
    COALESCE(i.criticality, 'unknown'),
    CASE COALESCE(i.criticality, 'unknown')
      WHEN 'Critical' THEN 'Критический'
      WHEN 'High' THEN 'Высокая'
      WHEN 'Low' THEN 'Низкая'
      ELSE COALESCE(i.criticality, 'unknown')
    END,
    COUNT(*)::numeric,
    COUNT(*)::numeric,
    COUNT(*) FILTER (WHERE i.incident_status IN ('Open', 'In Progress', 'Escalated'))::numeric,
    COUNT(*) FILTER (WHERE i.criticality = 'Critical')::numeric,
    COUNT(*) FILTER (WHERE i.incident_status = 'Escalated')::numeric,
    0, 0, 0, 0, 0, 0, 0, 0
  FROM incidents i
  GROUP BY i.criticality

  UNION ALL

  SELECT
    'incident_status',
    COALESCE(i.incident_status, 'unknown'),
    CASE COALESCE(i.incident_status, 'unknown')
      WHEN 'Open' THEN 'Открыт'
      WHEN 'In Progress' THEN 'В процессе'
      WHEN 'Escalated' THEN 'Эскалация'
      WHEN 'Resolved' THEN 'Решен'
      WHEN 'Closed' THEN 'Закрыт'
      ELSE COALESCE(i.incident_status, 'unknown')
    END,
    COUNT(*)::numeric,
    COUNT(*)::numeric,
    COUNT(*) FILTER (WHERE i.incident_status IN ('Open', 'In Progress', 'Escalated'))::numeric,
    COUNT(*) FILTER (WHERE i.criticality = 'Critical')::numeric,
    COUNT(*) FILTER (WHERE i.incident_status = 'Escalated')::numeric,
    0, 0, 0, 0, 0, 0, 0, 0
  FROM incidents i
  GROUP BY i.incident_status

  UNION ALL

  SELECT
    'store_incident',
    i.store_id::text,
    COALESCE(st.store_name, 'store '||i.store_id::text),
    COUNT(*)::numeric,
    COUNT(*)::numeric,
    COUNT(*)::numeric,
    COUNT(*) FILTER (WHERE i.criticality = 'Critical')::numeric,
    COUNT(*) FILTER (WHERE i.incident_status = 'Escalated')::numeric,
    0, 0, 0, 0, 0, 0, 0, 0
  FROM incidents i
  LEFT JOIN stores st ON st.store_id = i.store_id
  WHERE i.incident_status IN ('Open', 'In Progress', 'Escalated')
  GROUP BY i.store_id, st.store_name
)
SELECT
  row_number() OVER (ORDER BY slice_kind, risk_score DESC NULLS LAST, slice_label)::numeric AS id,
  slice_kind,
  slice_key,
  slice_label,
  ROUND(COALESCE(risk_score, 0)::numeric, 3) AS risk_score,
  ROUND(COALESCE(incident_count, 0)::numeric, 3) AS incident_count,
  ROUND(COALESCE(open_incident_count, 0)::numeric, 3) AS open_incident_count,
  ROUND(COALESCE(critical_incident_count, 0)::numeric, 3) AS critical_incident_count,
  ROUND(COALESCE(escalated_count, 0)::numeric, 3) AS escalated_count,
  ROUND(COALESCE(stock_critical_count, 0)::numeric, 3) AS stock_critical_count,
  ROUND(COALESCE(stock_understock_count, 0)::numeric, 3) AS stock_understock_count,
  ROUND(COALESCE(stock_risk_count, 0)::numeric, 3) AS stock_risk_count,
  ROUND(COALESCE(store_at_risk_count, 0)::numeric, 3) AS store_at_risk_count,
  ROUND(COALESCE(sku_count, 0)::numeric, 3) AS sku_count,
  ROUND(COALESCE(days_of_cover, 0)::numeric, 3) AS days_of_cover,
  ROUND(COALESCE(reorder_qty, 0)::numeric, 3) AS reorder_qty,
  ROUND(COALESCE(order_value, 0)::numeric, 3) AS order_value
FROM u;

CREATE UNIQUE INDEX risk_slice_id_idx ON risk_slice (id);
CREATE INDEX risk_slice_kind_idx ON risk_slice (slice_kind);

ALTER TABLE risk_slice
  ALTER COLUMN risk_score TYPE numeric(15,3),
  ALTER COLUMN incident_count TYPE numeric(15,3),
  ALTER COLUMN open_incident_count TYPE numeric(15,3),
  ALTER COLUMN critical_incident_count TYPE numeric(15,3),
  ALTER COLUMN escalated_count TYPE numeric(15,3),
  ALTER COLUMN stock_critical_count TYPE numeric(15,3),
  ALTER COLUMN stock_understock_count TYPE numeric(15,3),
  ALTER COLUMN stock_risk_count TYPE numeric(15,3),
  ALTER COLUMN store_at_risk_count TYPE numeric(15,3),
  ALTER COLUMN sku_count TYPE numeric(15,3),
  ALTER COLUMN days_of_cover TYPE numeric(15,3),
  ALTER COLUMN reorder_qty TYPE numeric(15,3),
  ALTER COLUMN order_value TYPE numeric(15,3);
