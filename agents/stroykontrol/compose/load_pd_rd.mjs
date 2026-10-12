#!/usr/bin/env node
/**
 * Loads the sample ПД/РД documents ("Пример ПД и РД") into the stroykontrol
 * namespace, grouped by project (объект) and by version:
 *
 *   объект  ──  сравнение (pd_rd_comparisons)  ──  комплект (cmp_documents: сторона + версия + файл)
 *
 * Two kinds of comparison are made:
 *   · ПД ↔ РД  — стадия П is the reference (side `ref`), the working docs the target;
 *   · РД: версия → версия — the older edition is the reference, the newer the target
 *     (before/after corrections), so the agent shows what the revision changed.
 *
 * Idempotent: objects and comparisons are found by code/title, uploaded files are
 * remembered in load_pd_rd.state.json (path → cmp_documents record ID). Nothing is
 * compared here — start «Сравнить документы» on the comparison card.
 *
 *   node load_pd_rd.mjs ["/path/to/Пример ПД и РД"]      (DRY_RUN=1 — only print the plan)
 */
import { existsSync, readFileSync, statSync, writeFileSync } from 'node:fs'
import { readFile } from 'node:fs/promises'
import { basename, dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { createRecord, setOf, mintToken, detectBase, apiFactory } from './helpers.mjs'
import { patchFields } from './filegen.mjs'

const HERE = dirname(fileURLToPath(import.meta.url))
const SRC = process.argv[2] || '/home/maxim/Загрузки/Пример ПД и РД'
const STATE_FILE = join(HERE, 'load_pd_rd.state.json')
const DRY = !!process.env.DRY_RUN

const NETWORK = 'Сеть автомобильных дорог общего пользования регионального значения Московской области «Солнцево – Бутово – Видное – Каширское шоссе – Молоково – Лыткарино – Томилино – Красково – Железнодорожный» на платной основе'

const PD = 'Стадия П'

const PROJECTS = [
  {
    code: 'ИССО 3.10',
    name: 'ИССО 3.10. Эстакада на пересечении с Павелецким направлением МЖД, ф.а.д. М-4 «Дон» на ПК 79+87,32',
    address: NETWORK + ', этап 3',
    notes: 'РД: раздел 4 «Искусственные сооружения», подраздел 1 «Опоры», часть 1 «Фундаменты», книга 13 «Опоры №17п…№19п. Свайные основания» (03102022/РД-ИССО3.10-ОП-КЖ1.13). ПД: раздел ПД №3, том 3.5.1.2 (ТКР5-1.2).',
    comparisons: [
      {
        title: 'ИССО 3.10 · КЖ1.13 · ПД ↔ РД (версия с ответами на замечания)',
        comment: 'Эталон — стадия П (раздел ПД №3, том 3.5.1.2 и лист ведомости свай). Проверяется РД КЖ1.13 в редакции с ответами на замечания №10351-РП от 15.11.23.',
        docs: [
          { side: 'ref', version: PD, file: 'РД и ПД/Раздел ПД №3. Том 3.5.1.2_ТКР5-1.2.pdf' },
          { side: 'ref', version: PD + ' (лист)', file: 'РД и ПД/ПД ИССО3.10 Том 3.5.1.2_ТКР5-1.2.pdf' },
          { side: 'target', version: 'С ответами на замечания №10351-РП', file: 'РД и ПД/03102022--РД-ИССО3.10-ОП-КЖ1.13.pdf' },
        ],
      },
      {
        title: 'ИССО 3.10 · КЖ1.13 · РД: исходная → с ответами на замечания',
        comment: 'Что изменилось в РД после ответов на замечания экспертизы №10351-РП от 15.11.23 (эталон — исходная редакция 2023 г.).',
        docs: [
          { side: 'ref', version: 'Исходная (2023)', file: 'РД и ПД/03102022_РД_ИССО3_10_ОП_КЖ1_13.pdf' },
          { side: 'target', version: 'С ответами на замечания №10351-РП', file: 'РД и ПД/03102022--РД-ИССО3.10-ОП-КЖ1.13.pdf' },
        ],
      },
    ],
  },
  {
    code: 'ИССО 10.1',
    name: 'ИССО 10.1. Мост через р. Москва',
    address: NETWORK + ', этап 1',
    notes: 'РД: раздел 4 «Искусственные сооружения», подраздел 1 «Опоры», часть 6 «Сопряжения с насыпями подходов», книга 1 «Опоры №1.1, №1.2. Сопряжение с насыпью» (03102022/РД-ИССО10.1-ОП-КЖ6-1).',
    comparisons: [
      {
        title: 'ИССО 10.1 · КЖ6-1 · РД: исходная → Изм. 2',
        comment: 'Исходная редакция (ноябрь 2022) и Изм. 2 (22.10.2024). Изм. 1 в примере не представлено.',
        docs: [
          { side: 'ref', version: 'Исходная (2022)', file: 'РД и РД Изм.1/03102022_РД-ИССО10.1-ОП-КЖ6-1.pdf' },
          { side: 'target', version: 'Изм. 2', file: 'РД и РД Изм.1/03102022_РД-ИССО10.1-ОП-КЖ6-1 изм. 2.pdf' },
        ],
      },
    ],
  },
  {
    code: 'ПИК1',
    name: 'ПИК1. Раздел 3.13. Переустройство коммуникаций',
    address: NETWORK + ', этап 3, часть 1 ПК0+00 – ПК122+00',
    notes: 'РД: раздел 3.13 «Переустройство коммуникаций». Книги: НК2 (подраздел 3 «Сети водоотведения», ПК80+00 – ПК90+55) и ЭС21 (подраздел 1 «Электрические сети», ПК33+94 – ПК47+20), шифр 03102022/РД-3-ПИК1-….',
    comparisons: [
      {
        title: 'ПИК1 · НК2 (сети водоотведения, кн. 2) · РД: исходная → Изм. 2',
        comment: 'Исходная редакция (2023) и комплект с Изм. 1 (25.07.24) и Изм. 2 (09.12.25).',
        docs: [
          { side: 'ref', version: 'Исходная (2023)', file: 'РД и РД Изм.1/03102022РД-3-ПИК1-НК2.pdf' },
          { side: 'target', version: 'Изм. 2', file: 'РД и РД Изм.1/Комплект РД 03102022_РД-3-ПИК1-НК2 изм.2.pdf' },
        ],
      },
      {
        title: 'ПИК1 · ЭС21 (электрические сети, кн. 21) · РД: исходная → Изм. 1',
        comment: 'Исходная редакция (2024) и комплект с Изм. 1.',
        docs: [
          { side: 'ref', version: 'Исходная (2024)', file: 'РД и РД Изм.1/03102022РД-3-ПИК1-ЭС21.pdf' },
          { side: 'target', version: 'Изм. 1', file: 'РД и РД Изм.1/Комплект РД 03102022_РД-3-ПИК1-ЭС21 изм.1.pdf' },
        ],
      },
    ],
  },
]

const mb = (n) => (n / 1048576).toFixed(1) + ' МБ'
const esc = (s) => String(s).replace(/\\/g, '\\\\').replace(/'/g, "\\'")

function loadState () {
  return existsSync(STATE_FILE) ? JSON.parse(readFileSync(STATE_FILE, 'utf8')) : { files: {} }
}

async function main () {
  // plan + sanity check of the source files
  let total = 0
  for (const p of PROJECTS) {
    console.log(`\n■ ${p.code} — ${p.name}`)
    for (const c of p.comparisons) {
      console.log(`  ▸ ${c.title}`)
      for (const d of c.docs) {
        const path = join(SRC, d.file)
        if (!existsSync(path)) throw new Error('нет файла: ' + path)
        const size = statSync(path).size
        total += size
        console.log(`      ${d.side === 'ref' ? 'эталон ' : 'проверка'} [${d.version}] ${basename(d.file)} (${mb(size)})`)
      }
    }
  }
  console.log(`\nК загрузке: ${mb(total)} (файл, нужный двум сравнениям, загружается в каждое отдельно)`)
  if (DRY) return

  const token = await mintToken()
  const base = await detectBase(token)
  const api = apiFactory(base, token)
  const applied = JSON.parse(readFileSync(join(HERE, 'applied.json'), 'utf8'))
  const nsID = applied.namespaceID
  const m = applied.modules
  const state = loadState()
  const save = () => writeFileSync(STATE_FILE, JSON.stringify(state, null, 2))

  const find = async (moduleID, query) => setOf(await api('GET',
    `/namespace/${nsID}/module/${moduleID}/record/?query=${encodeURIComponent(query)}&limit=5`))[0]

  async function upload (moduleID, recordID, fieldName, path) {
    const buf = await readFile(path)
    const form = new FormData()
    form.append('upload', new Blob([buf], { type: 'application/pdf' }), basename(path))
    form.append('recordID', String(recordID))
    form.append('fieldName', fieldName)
    const res = await fetch(`${base}/namespace/${nsID}/module/${moduleID}/record/attachment`, {
      method: 'POST',
      headers: { Authorization: 'Bearer ' + token },
      body: form,
      signal: AbortSignal.timeout(30 * 60 * 1000),
    })
    const text = await res.text()
    let data
    try { data = JSON.parse(text) } catch { data = { raw: text } }
    if (!res.ok || data.error) {
      throw new Error(`upload ${basename(path)} → ${res.status}: ${JSON.stringify(data.error || text.slice(0, 300))}`)
    }
    const body = data.response !== undefined ? data.response : data
    return String(body.attachmentID || body.attachment?.attachmentID || '')
  }

  for (const p of PROJECTS) {
    let obj = await find(m.objects, `code = '${esc(p.code)}'`)
    if (!obj) {
      obj = await createRecord(api, nsID, m.objects, { name: p.name, code: p.code, address: p.address, status: 'active', notes: p.notes })
      console.log('объект создан:', p.code, obj.recordID)
    } else {
      console.log('объект есть:', p.code, obj.recordID)
    }

    for (const c of p.comparisons) {
      let cmp = await find(m.pd_rd_comparisons, `title = '${esc(c.title)}'`)
      if (!cmp) {
        cmp = await createRecord(api, nsID, m.pd_rd_comparisons, { title: c.title, object: obj.recordID, status: 'new', comment: c.comment })
        console.log('  сравнение создано:', c.title, cmp.recordID)
      } else {
        console.log('  сравнение есть:', c.title, cmp.recordID)
      }

      for (const d of c.docs) {
        const key = `${cmp.recordID}|${d.side}|${d.file}`
        if (state.files[key]) {
          console.log('    пропуск (загружен):', basename(d.file))
          continue
        }
        const path = join(SRC, d.file)
        const rec = await createRecord(api, nsID, m.cmp_documents, { comparison: cmp.recordID, side: d.side, doc_version: d.version, detected_by: 'auto' })
        const t0 = Date.now()
        const att = await upload(m.cmp_documents, rec.recordID, 'file', path)
        if (!att) throw new Error('сервер не вернул attachmentID для ' + path)
        await patchFields(api, nsID, m.cmp_documents, rec.recordID, { file: att })
        state.files[key] = rec.recordID
        save()
        console.log(`    загружено: ${basename(d.file)} (${mb(statSync(path).size)}, ${((Date.now() - t0) / 1000).toFixed(0)} с) → ${rec.recordID}`)
      }
    }
  }
  console.log('\nготово')
}

main().catch(e => { console.error(e); process.exit(1) })
