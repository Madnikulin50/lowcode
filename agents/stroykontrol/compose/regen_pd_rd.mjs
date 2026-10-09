/**
 * Regenerates the pd_file/rd_file attachments on every existing pd_rd_comparisons
 * record: replaces the old stub docx (a handful of lines) with a fuller demo
 * docx of at least MIN_LINES=30 paragraphs each (see pdRdParagraphs() in
 * filegen.mjs) — title block, table of contents and a sheet-by-sheet listing.
 * ~30% of pairs get a synthetic DXF чертёж (dxfFile()) instead, so drawing
 * content and 'drawing'-type discrepancies are exercised too.
 * Does not touch objects, discrepancies, smeta or ВОИСР data — only the two
 * File fields on pd_rd_comparisons.
 *
 *   COMPOSE_DSN=postgres://postgres:Zse45rdx@127.0.0.1:5432/test11?sslmode=disable \
 *   node regen_pd_rd.mjs
 */
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

import { setOf, mintToken, detectBase, apiFactory } from './helpers.mjs'
import { docFile, dxfFile, uploadAttachment, patchFields, pdRdParagraphs } from './filegen.mjs'

function randInt (min, max) { return Math.floor(Math.random() * (max - min + 1)) + min }
function chance (p) { return Math.random() < p }

const HERE = dirname(fileURLToPath(import.meta.url))
const CONCURRENCY = Number(process.env.REGEN_CONCURRENCY || 6)

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
  const { api, base, token, nsID, m } = ctx
  const comparisonID = String(comparison.recordID)
  const objectID = val(comparison, 'object')
  const o = objectsByID.get(String(objectID)) || { name: `объект ${objectID}`, address: '', code: '' }
  const totalPagesPd = Number(val(comparison, 'total_pages_pd')) || 30
  const totalPagesRd = Number(val(comparison, 'total_pages_rd')) || 30

  // ~30% real чертежи (dxfFile, see filegen.mjs) so 'drawing'-type
  // discrepancies (chains.mjs / seed.mjs) have actual drawing content to
  // point at instead of only ever getting docx text stubs.
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
    uploadAttachment(base, token, nsID, m.pd_rd_comparisons, comparisonID, 'pd_file', pdFile),
    uploadAttachment(base, token, nsID, m.pd_rd_comparisons, comparisonID, 'rd_file', rdFile),
  ])
  await patchFields(api, nsID, m.pd_rd_comparisons, comparisonID, { pd_file: pdAtt, rd_file: rdAtt })
}

async function main () {
  const applied = JSON.parse(readFileSync(join(HERE, 'applied.json'), 'utf8'))
  const token = await mintToken()
  const base = await detectBase(token)
  const api = apiFactory(base, token)
  const nsID = applied.namespaceID
  const m = applied.modules
  console.log('API', base, 'ns', nsID)

  const objects = setOf(await api('GET', `/namespace/${nsID}/module/${m.objects}/record/?limit=1000`))
  const objectsByID = new Map(objects.map(r => [String(r.recordID), {
    name: val(r, 'name'),
    address: val(r, 'address'),
    code: val(r, 'code'),
  }]))
  console.log('objects loaded:', objectsByID.size)

  const comparisons = setOf(await api('GET', `/namespace/${nsID}/module/${m.pd_rd_comparisons}/record/?limit=1000`))
  console.log('pd_rd_comparisons to regenerate:', comparisons.length)

  const ctx = { api, base, token, nsID, m }
  let done = 0
  await mapPool(comparisons, CONCURRENCY, async (comparison) => {
    await regenOne(ctx, comparison, objectsByID)
    done++
    if (done % 20 === 0 || done === comparisons.length) console.log(`  ... ${done}/${comparisons.length}`)
  })

  console.log(`done. regenerated pd_file/rd_file on ${comparisons.length} pd_rd_comparisons records`)
}

main().catch(e => { console.error(e); process.exit(1) })
