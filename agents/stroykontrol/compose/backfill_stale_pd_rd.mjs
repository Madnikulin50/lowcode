/**
 * Re-triggers the pd_rd_comparisons AI chain for a specific list of records
 * (typically the ones still stuck on the pre-fix "вердикт ИИ недоступен"
 * comment — see stale_ids.txt), in small batches with a wait-for-completion
 * poll between batches.
 *
 * Why batched, not all-at-once like regen_pd_rd.mjs: firing all records at
 * once queues that many concurrent calls to the local Ollama instance: on a
 * single CPU-bound 8B model those calls batch together and finish in one
 * tight burst, and when ~150 finalize() writes hit Postgres in the same
 * second they blow past max_connections (100) — the finalize script node
 * itself still logs "OK" (it never throws), but the updateRecord() call
 * inside it fails and the record's comment silently stays whatever it was
 * before. Small batches + waiting for each batch to actually reach a
 * terminal status before starting the next keeps concurrent DB writes to
 * BATCH_SIZE at a time, well under the connection limit.
 *
 *   node backfill_stale_pd_rd.mjs stale_pd_rd_ids.txt
 *   # or: node backfill_stale_pd_rd.mjs <<< "512312736396673025
 *   # 512312736402440193"
 *
 * File/stdin format: one recordID per line (blank lines ignored).
 */
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

import { setOf, mintToken, detectBase, apiFactory } from './helpers.mjs'
import { docFile, dxfFile, uploadAttachment, patchFields, pdRdParagraphs } from './filegen.mjs'

function randInt (min, max) { return Math.floor(Math.random() * (max - min + 1)) + min }
function chance (p) { return Math.random() < p }
function sleep (ms) { return new Promise(r => setTimeout(r, ms)) }

const HERE = dirname(fileURLToPath(import.meta.url))
const BATCH_SIZE = Number(process.env.BACKFILL_BATCH_SIZE || 6)
const BATCH_CONCURRENCY = Number(process.env.BACKFILL_CONCURRENCY || 3)
const POLL_INTERVAL_MS = Number(process.env.BACKFILL_POLL_MS || 15000)
const BATCH_TIMEOUT_MS = Number(process.env.BACKFILL_TIMEOUT_MS || 15 * 60 * 1000)
const BATCH_GAP_MS = Number(process.env.BACKFILL_GAP_MS || 5000)

async function mapPool (items, limit, fn) {
  const out = new Array(items.length)
  let i = 0
  async function worker () {
    while (i < items.length) {
      const idx = i++
      out[idx] = await fn(items[idx], idx)
    }
  }
  await Promise.all(Array.from({ length: Math.min(limit, items.length) }, worker))
  return out
}

function val (rec, name) {
  const found = (rec.values || []).find(v => v.name === name)
  return found ? found.value : undefined
}

async function regenOne (ctx, comparison, objectsByID) {
  // Read token/api off ctx at call time, not destructured up front — this
  // script can run for hours (each batch waits up to BATCH_TIMEOUT_MS), way
  // past the ~2h life of a mintToken() JWT (see stroykontrol/README.md's
  // "Token stability" note), so main() refreshes ctx.token/ctx.api before
  // every batch. A stale local copy here would silently keep using the
  // expired token even after that refresh.
  const { base, nsID, m } = ctx
  const comparisonID = String(comparison.recordID)
  const objectID = val(comparison, 'object')
  const o = objectsByID.get(String(objectID)) || { name: `объект ${objectID}`, address: '', code: '' }
  const totalPagesPd = Number(val(comparison, 'total_pages_pd')) || 30
  const totalPagesRd = Number(val(comparison, 'total_pages_rd')) || 30

  let pdFile, rdFile
  if (chance(0.3)) {
    const dxfMeta = { sheetIndex: randInt(0, 5), sheetNo: 1, sheetsTotal: 1, drift: chance(0.5) ? randInt(50, 400) : 0 }
    pdFile = dxfFile(`${o.code || comparisonID}-PD`, 'pd', o, dxfMeta)
    rdFile = dxfFile(`${o.code || comparisonID}-RD`, 'rd', o, dxfMeta)
  } else {
    pdFile = docFile(`${o.code || comparisonID}-PD`, pdRdParagraphs('pd', o, { totalPages: totalPagesPd }))
    rdFile = docFile(`${o.code || comparisonID}-RD`, pdRdParagraphs('rd', o, { totalPages: totalPagesRd }))
  }

  const [pdAtt, rdAtt] = await Promise.all([
    uploadAttachment(base, ctx.token, nsID, m.pd_rd_comparisons, comparisonID, 'pd_file', pdFile),
    uploadAttachment(base, ctx.token, nsID, m.pd_rd_comparisons, comparisonID, 'rd_file', rdFile),
  ])
  await patchFields(ctx.api, nsID, m.pd_rd_comparisons, comparisonID, { pd_file: pdAtt, rd_file: rdAtt })
}

// Waits until every record in `ids` has left new/processing (chain done,
// success or failure), or until BATCH_TIMEOUT_MS elapses — whichever comes
// first. Records still stuck past the timeout are logged and left for a
// future run rather than blocking the whole backfill indefinitely.
async function waitForBatch (ctx, ids) {
  const { nsID, m } = ctx
  const deadline = Date.now() + BATCH_TIMEOUT_MS
  const pending = new Set(ids)
  while (pending.size > 0 && Date.now() < deadline) {
    await sleep(POLL_INTERVAL_MS)
    await Promise.all([...pending].map(async (id) => {
      try {
        const rec = await ctx.api('GET', `/namespace/${nsID}/module/${m.pd_rd_comparisons}/record/${id}`)
        const status = val(rec, 'status')
        if (status && status !== 'new' && status !== 'processing') pending.delete(id)
      } catch (e) {
        // transient fetch failure — leave it in `pending`, retry next tick
      }
    }))
    if (pending.size > 0) console.log(`    ...waiting on ${pending.size}/${ids.length} in this batch`)
  }
  if (pending.size > 0) console.log(`    ! ${pending.size} record(s) still not done after ${Math.round(BATCH_TIMEOUT_MS / 60000)} min, moving on: ${[...pending].join(', ')}`)
}

async function main () {
  const applied = JSON.parse(readFileSync(join(HERE, 'applied.json'), 'utf8'))
  const idsArgPath = process.argv[2]
  const raw = idsArgPath ? readFileSync(idsArgPath, 'utf8') : readFileSync(0, 'utf8')
  const ids = raw.split('\n').map(s => s.trim()).filter(Boolean)
  if (!ids.length) { console.error('no record IDs given'); process.exit(1) }
  console.log(`backfilling ${ids.length} records, batch size ${BATCH_SIZE}`)

  const token = await mintToken()
  const base = await detectBase(token)
  const api = apiFactory(base, token)
  const nsID = applied.namespaceID
  const m = applied.modules

  const objects = setOf(await api('GET', `/namespace/${nsID}/module/${m.objects}/record/?limit=1000`))
  const objectsByID = new Map(objects.map(r => [String(r.recordID), {
    name: val(r, 'name'),
    address: val(r, 'address'),
    code: val(r, 'code'),
  }]))

  const allComparisons = setOf(await api('GET', `/namespace/${nsID}/module/${m.pd_rd_comparisons}/record/?limit=1000`))
  const byID = new Map(allComparisons.map(r => [String(r.recordID), r]))

  const ctx = { api, base, token, nsID, m }
  const batches = []
  for (let i = 0; i < ids.length; i += BATCH_SIZE) batches.push(ids.slice(i, i + BATCH_SIZE))

  for (let b = 0; b < batches.length; b++) {
    const batchIds = batches[b]
    // Refresh before every batch, not just once at startup: each batch can
    // wait up to BATCH_TIMEOUT_MS, so a long run easily outlives the ~2h
    // JWT (this is what killed the first attempt at batch 14 — "token is
    // expired"). Safe to do between batches specifically because the
    // previous batch's uploads/patches have already been awaited — nothing
    // is still in flight on the old token when it gets invalidated.
    ctx.token = await mintToken()
    ctx.api = apiFactory(ctx.base, ctx.token)
    console.log(`\n[batch ${b + 1}/${batches.length}] triggering ${batchIds.length} records: ${batchIds.join(', ')}`)
    await mapPool(batchIds, BATCH_CONCURRENCY, async (id) => {
      const comparison = byID.get(id)
      if (!comparison) { console.log(`  ! ${id} not found, skipping`); return }
      await regenOne(ctx, comparison, objectsByID)
    })
    console.log(`[batch ${b + 1}/${batches.length}] triggered, waiting for chain completion (up to ${Math.round(BATCH_TIMEOUT_MS / 60000)} min)...`)
    await waitForBatch(ctx, batchIds)
    console.log(`[batch ${b + 1}/${batches.length}] done`)
    if (b < batches.length - 1) await sleep(BATCH_GAP_MS)
  }

  console.log(`\nbackfill complete: ${ids.length} records processed in ${batches.length} batches`)
}

main().catch(e => { console.error(e); process.exit(1) })
