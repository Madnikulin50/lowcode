-- Переносит пространство «Проверка ИД (АОСР)» (slug stroykontrol-id) в «Стройконтроль ПТО»
-- (slug stroykontrol) без потери данных: меняется только rel_namespace; ID модулей, полей,
-- записей, страниц, вложений и цепочек остаются прежними, поэтому ссылки record→record целы.
--
--   PGPASSWORD=… psql -h 127.0.0.1 -U postgres -d test11 -v ON_ERROR_STOP=1 -f migrate_id_into_main.sql           # пробный прогон (ROLLBACK)
--   PGPASSWORD=… psql … -v apply=1 -f migrate_id_into_main.sql                                                   # применить
--
-- Перед применением — дамп (см. backup/). Скрипт идемпотентен: если старого пространства уже нет,
-- ничего не делает. После применения перезапустите сервер Compose (кэш пространств/модулей).

\set ON_ERROR_STOP on
BEGIN;

CREATE TEMP TABLE _ns AS
  SELECT (SELECT id FROM compose_namespace WHERE slug = 'stroykontrol-id' AND deleted_at IS NULL) AS old_id,
         (SELECT id FROM compose_namespace WHERE slug = 'stroykontrol'    AND deleted_at IS NULL) AS new_id;

DO $$
DECLARE o numeric; n numeric;
BEGIN
  SELECT old_id, new_id INTO o, n FROM _ns;
  IF n IS NULL THEN RAISE EXCEPTION 'namespace stroykontrol not found'; END IF;
  IF o IS NULL THEN RAISE NOTICE 'namespace stroykontrol-id is already merged — nothing to do'; END IF;
  -- handle collisions would break the unique (namespace, handle) indexes
  IF o IS NOT NULL AND EXISTS (SELECT 1 FROM compose_module a JOIN compose_module b ON a.handle = b.handle AND b.rel_namespace = n AND b.deleted_at IS NULL WHERE a.rel_namespace = o AND a.deleted_at IS NULL) THEN
    RAISE EXCEPTION 'module handle collision between namespaces';
  END IF;
  IF o IS NOT NULL AND EXISTS (SELECT 1 FROM compose_page a JOIN compose_page b ON a.handle = b.handle AND b.rel_namespace = n AND b.deleted_at IS NULL WHERE a.rel_namespace = o AND a.deleted_at IS NULL) THEN
    RAISE EXCEPTION 'page handle collision between namespaces';
  END IF;
  IF o IS NOT NULL AND EXISTS (SELECT 1 FROM compose_rule_chain a JOIN compose_rule_chain b ON a.handle = b.handle AND b.rel_namespace = n AND b.deleted_at IS NULL WHERE a.rel_namespace = o AND a.deleted_at IS NULL) THEN
    RAISE EXCEPTION 'rule chain handle collision between namespaces';
  END IF;
END $$;

-- before: rows per table in both namespaces
CREATE TEMP TABLE _before AS
  SELECT 'module' t, count(*) c FROM compose_module WHERE rel_namespace IN (SELECT old_id FROM _ns UNION SELECT new_id FROM _ns)
  UNION ALL SELECT 'record', count(*) FROM compose_record WHERE rel_namespace IN (SELECT old_id FROM _ns UNION SELECT new_id FROM _ns)
  UNION ALL SELECT 'page', count(*) FROM compose_page WHERE rel_namespace IN (SELECT old_id FROM _ns UNION SELECT new_id FROM _ns)
  UNION ALL SELECT 'layout', count(*) FROM compose_page_layout WHERE rel_namespace IN (SELECT old_id FROM _ns UNION SELECT new_id FROM _ns)
  UNION ALL SELECT 'attachment', count(*) FROM compose_attachment WHERE rel_namespace IN (SELECT old_id FROM _ns UNION SELECT new_id FROM _ns)
  UNION ALL SELECT 'rule_chain', count(*) FROM compose_rule_chain WHERE rel_namespace IN (SELECT old_id FROM _ns UNION SELECT new_id FROM _ns)
  UNION ALL SELECT 'rule_chain_run', count(*) FROM compose_rule_chain_run WHERE rel_namespace IN (SELECT old_id FROM _ns UNION SELECT new_id FROM _ns)
  UNION ALL SELECT 'resource_versions', count(*) FROM compose_resource_versions WHERE rel_namespace IN (SELECT old_id FROM _ns UNION SELECT new_id FROM _ns)
  UNION ALL SELECT 'translations', count(*) FROM resource_translations WHERE resource LIKE 'compose:%';

UPDATE compose_module            SET rel_namespace = (SELECT new_id FROM _ns) WHERE rel_namespace = (SELECT old_id FROM _ns);
UPDATE compose_record            SET rel_namespace = (SELECT new_id FROM _ns) WHERE rel_namespace = (SELECT old_id FROM _ns);
UPDATE compose_page              SET rel_namespace = (SELECT new_id FROM _ns) WHERE rel_namespace = (SELECT old_id FROM _ns);
UPDATE compose_page_layout       SET rel_namespace = (SELECT new_id FROM _ns) WHERE rel_namespace = (SELECT old_id FROM _ns);
UPDATE compose_attachment        SET rel_namespace = (SELECT new_id FROM _ns) WHERE rel_namespace = (SELECT old_id FROM _ns);
UPDATE compose_rule_chain        SET rel_namespace = (SELECT new_id FROM _ns) WHERE rel_namespace = (SELECT old_id FROM _ns);
UPDATE compose_rule_chain_run    SET rel_namespace = (SELECT new_id FROM _ns) WHERE rel_namespace = (SELECT old_id FROM _ns);
UPDATE compose_resource_versions SET rel_namespace = (SELECT new_id FROM _ns) WHERE rel_namespace = (SELECT old_id FROM _ns);
UPDATE compose_record_drafts     SET rel_namespace = (SELECT new_id FROM _ns) WHERE rel_namespace = (SELECT old_id FROM _ns);
UPDATE compose_chart             SET rel_namespace = (SELECT new_id FROM _ns) WHERE rel_namespace = (SELECT old_id FROM _ns);
UPDATE compose_document          SET rel_namespace = (SELECT new_id FROM _ns) WHERE rel_namespace = (SELECT old_id FROM _ns);

-- translations keep the namespace ID inside the resource string: compose:<type>/<ns>/<id>
UPDATE resource_translations
   SET resource = replace(resource, '/' || (SELECT old_id FROM _ns)::text || '/', '/' || (SELECT new_id FROM _ns)::text || '/')
 WHERE (SELECT old_id FROM _ns) IS NOT NULL
   AND resource LIKE 'compose:%/' || (SELECT old_id FROM _ns)::text || '/%';

UPDATE compose_namespace SET deleted_at = now() WHERE id = (SELECT old_id FROM _ns);

-- after: totals must equal "before"
CREATE TEMP TABLE _after AS
  SELECT 'module' t, count(*) c FROM compose_module WHERE rel_namespace = (SELECT new_id FROM _ns)
  UNION ALL SELECT 'record', count(*) FROM compose_record WHERE rel_namespace = (SELECT new_id FROM _ns)
  UNION ALL SELECT 'page', count(*) FROM compose_page WHERE rel_namespace = (SELECT new_id FROM _ns)
  UNION ALL SELECT 'layout', count(*) FROM compose_page_layout WHERE rel_namespace = (SELECT new_id FROM _ns)
  UNION ALL SELECT 'attachment', count(*) FROM compose_attachment WHERE rel_namespace = (SELECT new_id FROM _ns)
  UNION ALL SELECT 'rule_chain', count(*) FROM compose_rule_chain WHERE rel_namespace = (SELECT new_id FROM _ns)
  UNION ALL SELECT 'rule_chain_run', count(*) FROM compose_rule_chain_run WHERE rel_namespace = (SELECT new_id FROM _ns)
  UNION ALL SELECT 'resource_versions', count(*) FROM compose_resource_versions WHERE rel_namespace = (SELECT new_id FROM _ns)
  UNION ALL SELECT 'translations', count(*) FROM resource_translations WHERE resource LIKE 'compose:%';

SELECT b.t AS "table", b.c AS before, a.c AS after, CASE WHEN a.c = b.c THEN 'ok' ELSE 'MISMATCH' END AS check
  FROM _before b JOIN _after a USING (t) ORDER BY 1;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM _before b JOIN _after a USING (t) WHERE a.c <> b.c) THEN
    RAISE EXCEPTION 'row counts differ after the move — rolled back';
  END IF;
END $$;

\if :{?apply}
  COMMIT;
  \echo 'APPLIED'
\else
  ROLLBACK;
  \echo 'DRY RUN — rolled back (pass -v apply=1 to apply)'
\endif
