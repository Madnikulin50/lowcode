#!/usr/bin/env node
/**
 * Attaches an expertise export (Excel «Выгрузка данных») to an object of the
 * stroykontrol namespace and asks the agent to import it: the registry of the
 * object's remarks, the history of iterations, the import journal and the
 * check of the answers against the loaded documents.
 *
 *   node load_expertise.mjs "/path/Выгрузка 07.11.24.xlsx" \
 *        --code "28/12-22-ГП-599" --name "Мост через р. Иртыш"   [--agent http://localhost:8092]
 *
 * The object is found by code (created if missing). Each run uploads the file
 * once more as a new export; the agent skips files it has already imported.
 */
import { readFile } from 'node:fs/promises'
import { existsSync, readFileSync } from 'node:fs'
import { basename, dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { createRecord, setOf, mintToken, detectBase, apiFactory } from './helpers.mjs'

const HERE = dirname(fileURLToPath(import.meta.url))
const args = process.argv.slice(2)
const file = args.find(a => !a.startsWith('--') && !args[args.indexOf(a) - 1]?.startsWith('--'))
const opt = (n, d = '') => { const i = args.indexOf('--' + n); return i >= 0 ? args[i + 1] : d }
const code = opt('code')
const name = opt('name', code)
const agent = (opt('agent') || process.env.STROYKONTROL_AGENT_URL || 'http://localhost:8092').replace(/\/$/, '')
if (!file || !code) {
  console.error('usage: node load_expertise.mjs <export.xlsx> --code <код объекта> [--name <наименование>] [--agent <url>]')
  process.exit(2)
}
if (!existsSync(file)) { console.error('нет файла:', file); process.exit(2) }

const token = await mintToken()
const base = await detectBase(token)
const api = apiFactory(base, token)
const applied = JSON.parse(readFileSync(join(HERE, 'applied.json'), 'utf8'))
const nsID = applied.namespaceID
const objects = applied.modules.objects

const esc = s => String(s).replace(/\\/g, '\\\\').replace(/'/g, "\\'")
let obj = setOf(await api('GET', `/namespace/${nsID}/module/${objects}/record/?query=${encodeURIComponent(`code = '${esc(code)}'`)}&limit=2`))[0]
if (!obj) {
  obj = await createRecord(api, nsID, objects, { name, code, status: 'active', notes: 'Объект экспертизы: реестр замечаний ведёт агент по выгрузкам Excel.' })
  console.log('объект создан:', obj.recordID)
} else {
  console.log('объект есть:', obj.recordID)
}

const buf = await readFile(file)
const form = new FormData()
form.append('upload', new Blob([buf], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' }), basename(file))
form.append('recordID', String(obj.recordID))
form.append('fieldName', 'remarks_file')
const res = await fetch(`${base}/namespace/${nsID}/module/${objects}/record/attachment`, { method: 'POST', headers: { Authorization: 'Bearer ' + token }, body: form, signal: AbortSignal.timeout(120000) })
const up = await res.json().catch(() => ({}))
if (!res.ok || up.error) { console.error('загрузка вложения:', JSON.stringify(up.error || up)); process.exit(1) }
const attID = String((up.response || up).attachmentID || '')
console.log('вложение:', attID)

// the field is multi-valued: keep the earlier exports and add this one
const rec = await api('GET', `/namespace/${nsID}/module/${objects}/record/${obj.recordID}`)
const values = (rec.values || []).filter(v => v.value !== '' && v.value != null).map(v => ({ name: v.name, value: String(v.value) }))
values.push({ name: 'remarks_file', value: attID })
await api('POST', `/namespace/${nsID}/module/${objects}/record/${obj.recordID}`, { values, updatedAt: rec.updatedAt })

const r = await fetch(`${agent}/api/expertise/import?objectID=${obj.recordID}&namespaceID=${nsID}`, { method: 'POST', signal: AbortSignal.timeout(600000) })
console.log('агент:', r.status, await r.text())
console.log(`рабочее место: ${agent}/?view=remarks&objectID=${obj.recordID}&namespaceID=${nsID}`)
