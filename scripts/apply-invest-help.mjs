#!/usr/bin/env node
/**
 * Write invest page/namespace help into Postgres (config.help / meta.help)
 * and update page.description via SQL so the next Compose GET keeps the short text.
 *
 *   node scripts/apply-invest-help.mjs
 */
import { execFileSync } from 'node:child_process'
import { INVEST_PAGE_DOCS, INVEST_NAMESPACE_DOCS } from '../client3/web/compose/src/help/appDocs.js'

const DSN_CANDIDATES = [...new Set([
  process.env.COMPOSE_DSN,
  'postgres://postgres:Zse45rdx@127.0.0.1:5432/test9?sslmode=disable',
  'postgres://postgres:Zse45rdx@127.0.0.1:5432/test10?sslmode=disable',
].filter(Boolean))]

function psql (dsn, sql) {
  return execFileSync('psql', [dsn, '-v', 'ON_ERROR_STOP=1', '-tA', '-c', sql], { encoding: 'utf8' }).trim()
}

function sqlLiteral (value) {
  return `E'${String(value).replace(/\\/g, '\\\\').replace(/'/g, "\\'")}'`
}

function detectInvest () {
  for (const dsn of DSN_CANDIDATES) {
    try {
      const row = psql(dsn, `SELECT id FROM compose_namespace WHERE slug = 'invest' AND deleted_at IS NULL LIMIT 1`)
      if (row) return { dsn, nsID: row }
    } catch {}
  }
  throw new Error('invest namespace not found. Set COMPOSE_DSN.')
}

function mergeJsonb (dsn, table, id, column, patch) {
  const pairs = Object.entries(patch)
    .map(([k, v]) => `${sqlLiteral(k)}, to_jsonb(${sqlLiteral(v)}::text)`)
    .join(', ')
  psql(dsn, `
    UPDATE ${table}
    SET ${column} = COALESCE(${column}, '{}'::jsonb) || jsonb_build_object(${pairs})
    WHERE id = ${id}
  `)
}

const { dsn, nsID } = detectInvest()
console.log('invest', nsID, dsn.replace(/:[^:@/]+@/, ':***@'))

mergeJsonb(dsn, 'compose_namespace', nsID, 'meta', {
  help: INVEST_NAMESPACE_DOCS.help,
  subtitle: INVEST_NAMESPACE_DOCS.hint,
  description: INVEST_NAMESPACE_DOCS.description,
})
console.log('namespace help')

const pages = psql(dsn, `SELECT handle || '|' || id FROM compose_page WHERE rel_namespace = ${nsID} AND deleted_at IS NULL`)
  .split('\n')
  .filter(Boolean)

let n = 0
for (const line of pages) {
  const sep = line.lastIndexOf('|')
  const handle = line.slice(0, sep)
  const id = line.slice(sep + 1)
  const docs = INVEST_PAGE_DOCS[handle]
  if (!docs) continue
  psql(dsn, `UPDATE compose_page SET description = ${sqlLiteral(docs.description)} WHERE id = ${id}`)
  mergeJsonb(dsn, 'compose_page', id, 'config', { help: docs.help })
  n++
  console.log('  page', handle)
}

console.log(`done: ${n} pages`)
