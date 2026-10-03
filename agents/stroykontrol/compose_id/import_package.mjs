/**
 * Создаёт «Пакет проверки» из папки комплекта ИД и загружает файлы:
 *   <папка>/Реестр ИД/*.xlsx|pdf, <папка>/КС/*.xlsx|pdf, <папка>/АОСР/**\/*.pdf
 *
 *   node import_package.mjs "<папка>" [--only=2.14.1.9] [--title="…"] [--run]
 *
 * --only   — загрузить только акты, в пути которых есть подстрока (для
 *            быстрой проверки на подмножестве, CPU-распознавание медленное);
 * --run    — сразу поставить пакет в очередь агента (IDCHECK_AGENT_URL).
 */
import { readFileSync, readdirSync, statSync, existsSync } from 'node:fs'
import { basename, dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { mintToken, detectBase, apiFactory } from '../compose/helpers.mjs'
import { uploadAttachment } from '../compose/filegen.mjs'

const HERE = dirname(fileURLToPath(import.meta.url))
const AGENT_URL = (process.env.IDCHECK_AGENT_URL || 'http://localhost:8092').replace(/\/$/, '')

const args = process.argv.slice(2)
const dir = args.find(a => !a.startsWith('--'))
const opt = name => (args.find(a => a.startsWith(`--${name}=`)) || '').split('=').slice(1).join('=')
if (!dir) {
  console.error('usage: node import_package.mjs <folder> [--only=<substr>] [--title=<title>] [--run]')
  process.exit(1)
}

function walk (d) {
  const out = []
  for (const name of readdirSync(d)) {
    if (name.startsWith('.~lock')) continue
    const p = join(d, name)
    if (statSync(p).isDirectory()) out.push(...walk(p))
    else out.push(p)
  }
  return out.sort()
}

const MIME = { '.pdf': 'application/pdf', '.xlsx': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' }
const fileOf = p => {
  const ext = p.slice(p.lastIndexOf('.')).toLowerCase()
  return { name: basename(p), mime: MIME[ext] || 'application/octet-stream', buf: readFileSync(p) }
}

async function main () {
  const applied = JSON.parse(readFileSync(join(HERE, 'applied_id.json'), 'utf8'))
  const nsID = applied.namespaceID
  const modID = applied.modules.packages
  const token = await mintToken()
  const base = await detectBase(token)
  const api = apiFactory(base, token)

  const all = walk(dir)
  const pick = (sub, ext) => all.find(p => p.includes(`/${sub}/`) && p.toLowerCase().endsWith(ext))
  const only = opt('only')
  const acts = all.filter(p => p.includes('/АОСР/') && p.toLowerCase().endsWith('.pdf') && (!only || p.includes(only)))
  const regX = pick('Реестр ИД', '.xlsx')
  const regP = pick('Реестр ИД', '.pdf')
  const ksX = pick('КС', '.xlsx')
  const ksP = pick('КС', '.pdf')
  const ksNum = (basename(ksX || '').match(/№\s*([^_ ]+_?[^_ ]*)/) || [])[1] || ''

  const title = opt('title') || `ИД ${basename(regX || dir).replace(/\.xlsx$/, '')}${only ? ` (${only})` : ''}`
  const rec = await api('POST', `/namespace/${nsID}/module/${modID}/record/`, {
    values: [
      { name: 'title', value: title },
      { name: 'ks2_number', value: ksNum.replace(/_/g, '/') },
      { name: 'status', value: 'new' },
    ],
  })
  const recordID = rec.recordID
  console.log('package', recordID, title)

  const values = [
    { name: 'title', value: title },
    { name: 'ks2_number', value: ksNum.replace(/_/g, '/') },
    { name: 'status', value: 'new' },
  ]
  const up = async (fieldName, path) => {
    if (!path || !existsSync(path)) return
    const id = await uploadAttachment(base, token, nsID, modID, recordID, fieldName, fileOf(path))
    values.push({ name: fieldName, value: id })
    console.log(' ', fieldName, basename(path))
  }
  await up('registry_xlsx', regX)
  await up('registry_pdf', regP)
  await up('ks2_xlsx', ksX)
  await up('ks2_pdf', ksP)
  for (const p of acts) await up('act_files', p)

  const cur = await api('GET', `/namespace/${nsID}/module/${modID}/record/${recordID}`)
  await api('POST', `/namespace/${nsID}/module/${modID}/record/${recordID}`, { values, updatedAt: cur.updatedAt })
  console.log(`uploaded ${acts.length} acts`)

  if (args.includes('--run')) {
    const res = await fetch(`${AGENT_URL}/api/idcheck/run?packageID=${recordID}&namespaceID=${nsID}`, { method: 'POST' })
    console.log('queued:', res.status, await res.text())
  }
  console.log(`open: /compose/ns/stroykontrol-id/pages/${applied.pages.package}/record/${recordID}`)
}

main().catch(e => { console.error(e); process.exit(1) })
