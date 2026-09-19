-- Performance indexes for the `incidents` DAL-connection table (retail
-- network namespace: "Управление сетью магазинов").
--
-- Why: the "Открытые инциденты" ("Open Incident Types") chart on the
-- summary page runs
--   SELECT incident_type, count(*) FROM incidents
--   WHERE incident_status != 'Closed' AND incident_status != 'Resolved'
--   GROUP BY incident_type
-- through Compose's generic report/aggregate DAL pipeline. That pipeline
-- (server/compose/service/record.go Report(), the `default` case with a
-- non-empty dimension) is only bounded by an 8s request-scoped context
-- (dalutils.BoundRecordSearchContext) — unlike the plain scalar COUNT(*)
-- shortcut, it does NOT catch a timeout and degrade gracefully; the raw
-- "context deadline exceeded" bubbles straight to the API response and the
-- Metric/Chart block. With no index, this was a full Seq Scan + filter over
-- every row; on a large `incidents` table that's the actual timeout cause.
--
-- Same store_id/status/criticality predicates are also used by the "Selected
-- store" page (per-store Metric/RecordList blocks, all `store_id = ...`)
-- and by the risk_slice query in seed_risk_dashboard.sql, so those benefit
-- from the same indexes.
--
-- Idempotent — safe to re-run.

-- General equality/negation lookups by status (status pie chart, filters).
CREATE INDEX IF NOT EXISTS incidents_status_idx
  ON incidents (incident_status);

-- Targeted partial index matching the "Open Incident Types" chart's exact
-- filter, so Postgres can bitmap/index-scan straight to open incidents
-- instead of scanning (and filtering) the whole table before grouping.
CREATE INDEX IF NOT EXISTS incidents_open_type_idx
  ON incidents (incident_type)
  WHERE incident_status NOT IN ('Closed', 'Resolved');

-- Per-store widgets (Selected store page: Stats metrics, Open Incidents
-- list, risk_slice) all filter on store_id equality.
CREATE INDEX IF NOT EXISTS incidents_store_id_idx
  ON incidents (store_id);

ANALYZE incidents;
