-- Помечает демо-данные: объекты-примеры (код вида OBJ-0001, их создаёт seed.mjs) и всё, что к ним
-- относится, получают data_kind = 'demo'. Рабочие списки и метрики такие записи скрывают
-- (см. REAL в pages.mjs); открыть их можно в «Служебное → Демо-данные».
-- Откат: UPDATE compose_record SET values = values - 'data_kind' для тех же записей.
--
--   psql … -f mark_demo.sql            # пробный прогон (ROLLBACK)
--   psql … -v apply=1 -f mark_demo.sql # применить
\set ON_ERROR_STOP on
BEGIN;

CREATE TEMP TABLE _mod AS
  SELECT m.handle, m.id FROM compose_module m
  JOIN compose_namespace n ON n.id = m.rel_namespace AND n.slug = 'stroykontrol' AND n.deleted_at IS NULL
  WHERE m.deleted_at IS NULL;

CREATE TEMP TABLE _demo_obj AS
  SELECT r.id FROM compose_record r JOIN _mod ON _mod.id = r.rel_module AND _mod.handle = 'objects'
  WHERE r.deleted_at IS NULL AND r.values->'code'->>0 ~ '^OBJ-[0-9]+$';

CREATE TEMP TABLE _demo_cmp AS
  SELECT r.id FROM compose_record r JOIN _mod ON _mod.id = r.rel_module AND _mod.handle = 'pd_rd_comparisons'
  WHERE r.deleted_at IS NULL AND r.values->'object'->>0 IN (SELECT id::text FROM _demo_obj);

CREATE TEMP TABLE _demo_item AS
  SELECT r.id FROM compose_record r JOIN _mod ON _mod.id = r.rel_module AND _mod.handle = 'voisr_items'
  WHERE r.deleted_at IS NULL AND r.values->'object'->>0 IN (SELECT id::text FROM _demo_obj);

CREATE TEMP TABLE _stamped (what text, n int);

WITH u AS (UPDATE compose_record SET values = values || '{"data_kind":["demo"]}'::jsonb WHERE id IN (SELECT id FROM _demo_obj) RETURNING id)
INSERT INTO _stamped SELECT 'objects', count(*) FROM u;

WITH u AS (UPDATE compose_record SET values = values || '{"data_kind":["demo"]}'::jsonb WHERE id IN (SELECT id FROM _demo_cmp) RETURNING id)
INSERT INTO _stamped SELECT 'pd_rd_comparisons', count(*) FROM u;

WITH u AS (UPDATE compose_record r SET values = values || '{"data_kind":["demo"]}'::jsonb
  FROM _mod WHERE _mod.id = r.rel_module AND _mod.handle = 'pd_rd_discrepancies' AND r.deleted_at IS NULL
    AND r.values->'comparison'->>0 IN (SELECT id::text FROM _demo_cmp) RETURNING r.id)
INSERT INTO _stamped SELECT 'pd_rd_discrepancies', count(*) FROM u;

WITH u AS (UPDATE compose_record r SET values = values || '{"data_kind":["demo"]}'::jsonb
  FROM _mod WHERE _mod.id = r.rel_module AND _mod.handle = 'expertise_remarks' AND r.deleted_at IS NULL
    AND r.values->'object'->>0 IN (SELECT id::text FROM _demo_obj) RETURNING r.id)
INSERT INTO _stamped SELECT 'expertise_remarks', count(*) FROM u;

WITH u AS (UPDATE compose_record r SET values = values || '{"data_kind":["demo"]}'::jsonb
  FROM _mod WHERE _mod.id = r.rel_module AND _mod.handle = 'packages' AND r.deleted_at IS NULL
    AND r.values->'object_ref'->>0 IN (SELECT id::text FROM _demo_obj) RETURNING r.id)
INSERT INTO _stamped SELECT 'packages', count(*) FROM u;

WITH u AS (UPDATE compose_record r SET values = values || '{"data_kind":["demo"]}'::jsonb
  FROM _mod WHERE _mod.id = r.rel_module AND _mod.handle = 'voisr_check_results' AND r.deleted_at IS NULL
    AND r.values->'voisr_item'->>0 IN (SELECT id::text FROM _demo_item) RETURNING r.id)
INSERT INTO _stamped SELECT 'voisr_check_results', count(*) FROM u;

SELECT what AS "помечено демо", n FROM _stamped ORDER BY 1;

\if :{?apply}
  COMMIT;
  \echo 'APPLIED'
\else
  ROLLBACK;
  \echo 'DRY RUN — rolled back (pass -v apply=1 to apply)'
\endif
