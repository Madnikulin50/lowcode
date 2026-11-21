/**
 * Snapshots one real pd_rd_comparisons record (its title/status/similarity/
 * comment, its pd_rd_discrepancies, and its pd_file/rd_file attachments)
 * out of a live Compose instance into the on-disk layout FixtureStore
 * expects (see ../fixtures/README.md) — so it can be viewed by
 * stroykontrol-web afterwards with `--fixtures`, no Compose/token needed.
 *
 * Same auth/API-discovery path as seed.mjs: mintToken() (or TOKEN env) +
 * detectBase().
 *
 *   node export_fixture.mjs <namespaceID> <recordID> [caseID]
 *
 * caseID defaults to recordID and becomes the fixtures/<caseID>/ dirname.
 *
 *   COMPOSE_DSN=postgres://postgres:Zse45rdx@127.0.0.1:5432/test11?sslmode=disable \
 *   node export_fixture.mjs 512312736219004929 512315423893749761
 */
import { mkdirSync, writeFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

import { mintToken, detectBase, apiFactory } from './helpers.mjs'

const HERE = dirname(fileURLToPath(import.meta.url))
const FIXTURES_DIR = join(HERE, '..', 'fixtures')

const COMPARISON_MODULE = 'pd_rd_comparisons'
const DISCREPANCY_MODULE = 'pd_rd_discrepancies'

function val (rec, name) {
  const v = (rec.values || []).find(v => v.name === name)
  return v && v.value != null ? String(v.value) : ''
}

async function resolveModuleID (api, nsID, handle) {
  const out = await api('GET', `/namespace/${nsID}/module/?handle=${encodeURIComponent(handle)}&limit=1`)
  const id = out.set?.[0]?.moduleID
  if (!id) throw new Error(`module ${handle} not found in namespace ${nsID}`)
  return id
}

// downloadAttachment mirrors ComposeClient.downloadOriginal in compose_client.go:
// attachment metadata's .url comes back freshly signed for our own identity,
// plain Bearer auth against /original alone isn't accepted for record attachments.
async function downloadAttachment (base, token, nsID, attachmentID) {
  const meta = await fetch(`${base}/namespace/${nsID}/attachment/record/${attachmentID}`, {
    headers: { Authorization: 'Bearer ' + token, Accept: 'application/json' },
    signal: AbortSignal.timeout(30000),
  }).then(r => r.json())
  const att = meta.response !== undefined ? meta.response : meta
  if (!att?.url) throw new Error(`attachment ${attachmentID} has no url (${JSON.stringify(meta.error || att)})`)
  const url = att.url.startsWith('/') ? base + att.url : att.url
  const res = await fetch(url, { headers: { Authorization: 'Bearer ' + token }, signal: AbortSignal.timeout(30000) })
  if (!res.ok) throw new Error(`download ${attachmentID} → HTTP ${res.status}`)
  const buf = Buffer.from(await res.arrayBuffer())
  return { name: att.name || attachmentID, contentType: res.headers.get('content-type') || '', buf }
}

// extFor guesses a file extension from the attachment's own name first
// (most reliable), falling back to sniffing the content — same content
// FixtureStore/IsPDF/IsDOCX would sniff anyway, so a wrong extension here
// wouldn't actually break the viewer, but a right one reads better on disk.
function extFor (name, contentType, buf) {
  const m = /\.([a-zA-Z0-9]+)$/.exec(name || '')
  if (m) return '.' + m[1].toLowerCase()
  if (/pdf/i.test(contentType) || (buf.length > 4 && buf.slice(0, 4).toString() === '%PDF')) return '.pdf'
  if (/wordprocessingml/i.test(contentType)) return '.docx'
  return '.bin'
}

async function main () {
  const [nsID, recordID, caseIDArg] = process.argv.slice(2)
  if (!nsID || !recordID) {
    console.error('usage: node export_fixture.mjs <namespaceID> <recordID> [caseID]')
    process.exit(1)
  }
  const caseID = caseIDArg || recordID
  const outDir = join(FIXTURES_DIR, caseID)

  const token = await mintToken()
  const base = await detectBase(token)
  const api = apiFactory(base, token)
  console.log('API', base, 'ns', nsID, '→', outDir)

  const comparisonModID = await resolveModuleID(api, nsID, COMPARISON_MODULE)
  const rec = await api('GET', `/namespace/${nsID}/module/${comparisonModID}/record/${recordID}`)

  mkdirSync(outDir, { recursive: true })

  const meta = {
    title: val(rec, 'title'),
    status: val(rec, 'status'),
    similarityPercent: val(rec, 'similarity_percent'),
    comment: val(rec, 'comment'),
  }
  writeFileSync(join(outDir, 'meta.json'), JSON.stringify(meta, null, 2) + '\n')
  console.log('wrote meta.json', meta)

  let discModID
  try {
    discModID = await resolveModuleID(api, nsID, DISCREPANCY_MODULE)
  } catch (e) {
    console.warn('discrepancies module not found, skipping:', e.message)
  }
  if (discModID) {
    const q = `?query=${encodeURIComponent(`comparison = '${recordID}'`)}&limit=500`
    const discs = await api('GET', `/namespace/${nsID}/module/${discModID}/record/${q}`)
    const out = (discs.set || []).map(d => ({
      type: val(d, 'discrepancy_type'),
      severity: val(d, 'severity'),
      description: val(d, 'description'),
      pageNumber: val(d, 'page_number'),
    }))
    writeFileSync(join(outDir, 'discrepancies.json'), JSON.stringify(out, null, 2) + '\n')
    console.log(`wrote discrepancies.json (${out.length} rows)`)
  }

  for (const [side, field] of [['pd', 'pd_file'], ['rd', 'rd_file']]) {
    const attachmentID = val(rec, field)
    if (!attachmentID) {
      console.log(`no ${field} on this record, skipping ${side}`)
      continue
    }
    const { name, contentType, buf } = await downloadAttachment(base, token, nsID, attachmentID)
    const ext = extFor(name, contentType, buf)
    writeFileSync(join(outDir, side + ext), buf)
    console.log(`wrote ${side}${ext} (${buf.length} bytes, from ${name})`)
  }

  console.log(`done. run: go run . --fixtures=fixtures  then open /?recordID=${caseID}&namespaceID=x`)
}

main().catch(e => { console.error(e); process.exit(1) })
