#!/usr/bin/env node
/**
 * Fill empty Compose AI prompts in local Postgres (test9 + test10).
 * Existing non-empty prompts are left untouched.
 *
 *   node scripts/fill-ai-prompts.mjs
 */
import { execFileSync } from 'node:child_process'
import { mkdtempSync, writeFileSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { fillStoredPage, isEmptyPrompt, namespacePrompt } from './ai-prompts.mjs'

const DBS = ['test9', 'test10']
const PGPASSWORD = process.env.PGPASSWORD || 'Zse45rdx'
const TAG = 'fill_ai_prompt_json_v1'

function psql (db, sql, viaFile = false) {
  const args = [
    '-h', '127.0.0.1',
    '-U', 'postgres',
    '-d', db,
    '-v', 'ON_ERROR_STOP=1',
    '-t', '-A',
  ]
  if (viaFile) {
    const dir = mkdtempSync(join(tmpdir(), 'fill-ai-prompts-'))
    const file = join(dir, 'update.sql')
    writeFileSync(file, sql)
    try {
      return execFileSync('psql', [...args, '-f', file], {
        env: { ...process.env, PGPASSWORD },
        encoding: 'utf8',
        maxBuffer: 80 * 1024 * 1024,
      }).trim()
    } finally {
      rmSync(dir, { recursive: true, force: true })
    }
  }
  return execFileSync('psql', [...args, '-c', sql], {
    env: { ...process.env, PGPASSWORD },
    encoding: 'utf8',
    maxBuffer: 80 * 1024 * 1024,
  }).trim()
}

function sqlLiteral (value) {
  return `$${TAG}$${value}$${TAG}$`
}

function quoteId (id) {
  return "'" + String(id).replace(/'/g, "''") + "'"
}

function fillDatabase (db) {
  const raw = psql(db, `
    SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json)
    FROM (
      SELECT
        p.id::text AS page_id,
        n.id::text AS ns_id,
        n.slug,
        n.meta AS ns_meta,
        p.handle,
        p.title,
        p.config,
        p.blocks
      FROM compose_page p
      JOIN compose_namespace n ON n.id = p.rel_namespace
      WHERE p.deleted_at IS NULL AND n.deleted_at IS NULL
      ORDER BY n.slug, p.handle, p.id
    ) t
  `)
  const rows = JSON.parse(raw || '[]')
  const stats = { pages: 0, blocks: 0, pagePrompts: 0, namespaces: 0, skippedBlocks: 0 }

  const nsSeen = new Set()
  const statements = []

  for (const row of rows) {
    const nsKey = String(row.ns_id)
    if (!nsSeen.has(nsKey)) {
      nsSeen.add(nsKey)
      const meta = row.ns_meta && typeof row.ns_meta === 'object' ? { ...row.ns_meta } : {}
      if (isEmptyPrompt(meta.prompt)) {
        const text = namespacePrompt(row.slug)
        if (text) {
          meta.prompt = text
          statements.push(
            `UPDATE compose_namespace SET meta = ${sqlLiteral(JSON.stringify(meta))}::jsonb WHERE id = ${quoteId(row.ns_id)};`,
          )
          stats.namespaces++
        }
      }
    }

    let skipped = 0
    for (const b of row.blocks || []) {
      if (!isEmptyPrompt(b.prompt)) skipped++
    }
    stats.skippedBlocks += skipped

    const result = fillStoredPage(row)
    if (!result.changed) continue

    const beforePage = isEmptyPrompt(row.config?.prompt)
    const afterPage = !isEmptyPrompt(result.config.prompt)
    if (beforePage && afterPage) stats.pagePrompts++

    const beforeBlocks = (row.blocks || []).filter((b) => isEmptyPrompt(b.prompt)).length
    const afterEmpty = result.blocks.filter((b) => isEmptyPrompt(b.prompt)).length
    stats.blocks += Math.max(0, beforeBlocks - afterEmpty)
    stats.pages++

    statements.push(
      `UPDATE compose_page SET blocks = ${sqlLiteral(JSON.stringify(result.blocks))}::jsonb, config = ${sqlLiteral(JSON.stringify(result.config))}::jsonb WHERE id = ${quoteId(row.page_id)};`,
    )
  }

  if (statements.length) {
    psql(db, 'BEGIN;\n' + statements.join('\n') + '\nCOMMIT;\n', true)
  }

  return { db, pagesLoaded: rows.length, ...stats, updates: statements.length }
}

const report = []
for (const db of DBS) {
  try {
    report.push(fillDatabase(db))
  } catch (err) {
    report.push({ db, error: err.stderr || err.message })
  }
}
console.log(JSON.stringify(report, null, 2))
