/**
 * Fills id_files.extracted_text from the attached file (docx / pdf / text).
 * Run apply.mjs first so the field exists. Re-running only touches records
 * whose text is empty, unless FORCE=1.
 *
 *   node backfill_id_files_text.mjs
 */
import { readFileSync, writeFileSync, mkdtempSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { execFileSync } from 'node:child_process'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { setOf, mintToken, detectBase, apiFactory } from './helpers.mjs'
import { patchFields } from './filegen.mjs'

const HERE = dirname(fileURLToPath(import.meta.url))
const MAX_CHARS = 64000

function val (rec, name) {
  const f = (rec.values || []).find(v => v.name === name)
  return f ? f.value : undefined
}

function decodeXml (s) {
  return s.replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&quot;/g, '"').replace(/&apos;/g, "'").replace(/&amp;/g, '&')
}

export function extractText (buf, name) {
  const lower = name.toLowerCase()
  let text = ''
  if (lower.endsWith('.docx')) {
    const dir = mkdtempSync(join(tmpdir(), 'idf-'))
    let xml
    try {
      writeFileSync(join(dir, 'f.docx'), buf)
      xml = execFileSync('unzip', ['-p', join(dir, 'f.docx'), 'word/document.xml'], { maxBuffer: 1 << 28 }).toString('utf8')
    } finally { rmSync(dir, { recursive: true, force: true }) }
    text = decodeXml(xml.replace(/<\/w:p>/g, '\n').replace(/<w:tab\/>/g, '\t').replace(/<[^>]+>/g, ''))
  } else if (lower.endsWith('.pdf')) {
    text = execFileSync('pdftotext', ['-layout', '-', '-'], { input: buf, maxBuffer: 1 << 28 }).toString('utf8')
  } else if (/\.(txt|csv|md|dxf)$/.test(lower)) {
    text = buf.toString('utf8')
  }
  text = text.replace(/\r/g, '').replace(/[ \t]+\n/g, '\n').replace(/\n{3,}/g, '\n\n').trim()
  return text.slice(0, MAX_CHARS)
}

async function main () {
  const applied = JSON.parse(readFileSync(join(HERE, 'applied.json'), 'utf8'))
  const token = await mintToken()
  const base = await detectBase(token)
  const api = apiFactory(base, token)
  const nsID = applied.namespaceID
  const mod = applied.modules.id_files
  const records = setOf(await api('GET', `/namespace/${nsID}/module/${mod}/record/?limit=1000`))
  let done = 0, skipped = 0, failed = 0
  for (const rec of records) {
    const id = String(rec.recordID)
    const attID = val(rec, 'file')
    if (!attID) { skipped++; continue }
    if (val(rec, 'extracted_text') && !process.env.FORCE) { skipped++; continue }
    try {
      const meta = await api('GET', `/namespace/${nsID}/attachment/record/${attID}`)
      const name = meta.name || 'file'
      // the original is served only with the signed URL from the attachment meta
      const res = await fetch(base + meta.url, { headers: { Authorization: 'Bearer ' + token } })
      if (!res.ok) throw new Error(`download ${res.status}`)
      const text = extractText(Buffer.from(await res.arrayBuffer()), name)
      if (!text) { console.log(`- ${id}: no text from ${name}`); skipped++; continue }
      await patchFields(api, nsID, mod, id, { extracted_text: text })
      console.log(`+ ${id}: ${name}, ${text.length} chars`)
      done++
    } catch (e) {
      console.log(`! ${id}: ${e.message}`)
      failed++
    }
  }
  console.log(`done=${done} skipped=${skipped} failed=${failed}`)
}

if (process.argv[1] === fileURLToPath(import.meta.url)) main().catch(e => { console.error(e); process.exit(1) })
