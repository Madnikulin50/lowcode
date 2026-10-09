/**
 * Demo data for the "Стройконтроль ПТО" namespace: ≥100 объектов строительства,
 * каждый — со своим комплектом документов (реестр ИД + загруженные файлы +
 * сравнение ПД/РД + сметы/ВОИСР), включая реальные вложенные .docx-файлы.
 *
 * Idempotent: если объектов уже >= TARGET_OBJECTS, ничего не досоздаёт
 * (кроме случая SEED_FORCE=1 — тогда досеивает поверх, без удаления старого).
 *
 *   COMPOSE_DSN=postgres://postgres:Zse45rdx@127.0.0.1:5432/test11?sslmode=disable \
 *   node seed.mjs
 */
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

import { createRecord, setOf, mintToken, detectBase, apiFactory } from './helpers.mjs'
import { docFile, dxfFile, uploadAttachment, patchFields, pdRdParagraphs } from './filegen.mjs'

const HERE = dirname(fileURLToPath(import.meta.url))
const TARGET_OBJECTS = Number(process.env.SEED_OBJECTS || 120)
const CONCURRENCY = Number(process.env.SEED_CONCURRENCY || 6)
const FORCE = process.env.SEED_FORCE === '1'

function recID (row) {
  return String(row.recordID || row.ID || '')
}

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

function randInt (min, max) { return Math.floor(Math.random() * (max - min + 1)) + min }
function pick (arr) { return arr[randInt(0, arr.length - 1)] }
function chance (p) { return Math.random() < p }
function shuffle (arr) {
  const a = arr.slice()
  for (let i = a.length - 1; i > 0; i--) {
    const j = randInt(0, i)
    ;[a[i], a[j]] = [a[j], a[i]]
  }
  return a
}
function round2 (n) { return Math.round(n * 100) / 100 }
function isoDate (daysAgo) {
  const d = new Date(Date.now() - daysAgo * 86400000)
  return d.toISOString()
}

const OBJECT_TYPES = [
  'Жилой дом', 'Детский сад', 'Школа', 'Поликлиника', 'Спортивный комплекс',
  'Торгово-развлекательный центр', 'Многоуровневый паркинг', 'Административно-деловой центр',
  'Складской комплекс', 'Котельная', 'Мост', 'Автомобильная дорога', 'Насосная станция',
  'Очистные сооружения',
]

const CITIES = [
  'Москва', 'Санкт-Петербург', 'Казань', 'Новосибирск', 'Екатеринбург',
  'Нижний Новгород', 'Краснодар', 'Воронеж', 'Самара', 'Уфа', 'Красноярск', 'Пермь', 'Тюмень', 'Ростов-на-Дону',
]

const STREETS = [
  'Ленина', 'Гагарина', 'Мира', 'Строителей', 'Молодёжная', 'Садовая', 'Центральная',
  'Промышленная', 'Заречная', 'Северная', 'Южная', 'Октябрьская', 'Пушкина', 'Кирова', 'Полевая', 'Лесная',
]

function objectStatus () {
  const r = Math.random()
  if (r < 0.7) return 'active'
  if (r < 0.85) return 'suspended'
  return 'closed'
}

const DOC_CATALOG = [
  ['Разрешение на строительство', 'other'],
  ['Общий журнал работ', 'journal'],
  ['Журнал сварочных работ', 'journal'],
  ['Журнал бетонных работ', 'journal'],
  ['Акт освидетельствования скрытых работ', 'act'],
  ['Акт освидетельствования ответственных конструкций', 'act'],
  ['Акт входного контроля материалов', 'act'],
  ['Паспорт качества бетона', 'certificate'],
  ['Сертификат соответствия на металлопрокат', 'certificate'],
  ['Сертификат соответствия на кабельную продукцию', 'certificate'],
  ['Исполнительная схема фундаментов', 'scheme'],
  ['Исполнительная схема инженерных сетей', 'scheme'],
  ['Протокол испытаний сварных соединений', 'act'],
  ['Акт приёмки геодезической разбивочной основы', 'act'],
]

const WORK_CATALOG = [
  ['Устройство фундаментной плиты', 'м3', 150, 600, 12000, 22000],
  ['Устройство свайного основания', 'м.п.', 80, 400, 3500, 7000],
  ['Кладка стен из кирпича', 'м3', 200, 900, 6500, 12000],
  ['Монтаж сборных ж/б конструкций', 'м3', 100, 500, 9000, 16000],
  ['Монтаж металлоконструкций каркаса', 'т', 20, 120, 55000, 85000],
  ['Устройство кровли', 'м2', 300, 2000, 1200, 3200],
  ['Штукатурные работы фасада', 'м2', 400, 2500, 900, 2200],
  ['Устройство стяжки полов', 'м2', 500, 3000, 700, 1600],
  ['Монтаж внутренней электропроводки', 'м.п.', 1000, 6000, 180, 450],
  ['Прокладка сетей водоснабжения', 'м.п.', 200, 1200, 2200, 5000],
  ['Прокладка сетей канализации', 'м.п.', 150, 900, 2500, 5500],
  ['Устройство навесного вентилируемого фасада', 'м2', 300, 1800, 3500, 7500],
  ['Благоустройство территории', 'м2', 500, 4000, 900, 2500],
  ['Устройство асфальтобетонного покрытия', 'м2', 300, 2500, 1100, 2600],
]

async function put (api, nsID, moduleID, values) {
  const compact = Object.fromEntries(Object.entries(values).filter(([, v]) => v !== '' && v != null))
  const rec = await createRecord(api, nsID, moduleID, compact)
  return recID(rec)
}

function genObject (i) {
  const type = pick(OBJECT_TYPES)
  const city = pick(CITIES)
  const street = pick(STREETS)
  const house = randInt(1, 180)
  const code = `OBJ-${String(i).padStart(4, '0')}`
  return {
    code,
    name: `${type} №${i}`,
    address: `г. ${city}, ул. ${street}, д. ${house}`,
    status: objectStatus(),
  }
}

async function seedObject (ctx, i) {
  const { api, base, token, nsID, m } = ctx
  const o = genObject(i)
  const objectID = await put(api, nsID, m.objects, {
    name: o.name, address: o.address, code: o.code, status: o.status,
  })

  // ---- Реестр ИД ----
  const registryDocs = shuffle(DOC_CATALOG).slice(0, randInt(5, 8))
  const registryIDs = []
  for (const [docName, docType] of registryDocs) {
    const registryID = await put(api, nsID, m.id_registry, {
      object: objectID,
      doc_name: docName,
      doc_number: `ИД-${o.code}-${randInt(100, 999)}`,
      doc_version: pick(['v1', 'v1.1', 'v2']),
      doc_type: docType,
      is_required: chance(0.8) ? 'true' : 'false',
    })
    registryIDs.push({ registryID, docName })
  }

  // ---- Файлы ИД (сопоставление с реестром) ----
  let uploadedCount = 0
  let mismatchedCount = 0
  for (const { registryID, docName } of registryIDs) {
    if (!chance(0.82)) continue // часть документов пока не загружена
    const r = Math.random()
    let matchStatus = 'matched'
    let recognizedName = docName
    if (r > 0.92) { matchStatus = 'version_mismatch' } else if (r > 0.85) { matchStatus = 'number_mismatch' } else if (r > 0.75) { matchStatus = 'name_mismatch'; recognizedName = docName + ' (скан)' }
    if (matchStatus !== 'matched') mismatchedCount++
    else uploadedCount++

    const fileID = await put(api, nsID, m.id_files, {
      object: objectID,
      recognized_name: recognizedName,
      recognized_number: `ИД-${o.code}-${randInt(100, 999)}`,
      recognized_version: pick(['v1', 'v1.1', 'v2']),
      matched_registry: registryID,
      match_status: matchStatus,
      uploaded_at: isoDate(randInt(1, 180)),
    })
    const file = docFile(`${o.code}-${fileID}`, [
      docName, `Объект: ${o.name}`, `Адрес: ${o.address}`, '',
      'Демо-документ для проверки комплектности исполнительной документации.',
    ])
    const attID = await uploadAttachment(base, token, nsID, m.id_files, fileID, 'file', file)
    if (attID) await patchFields(api, nsID, m.id_files, fileID, { file: attID })
  }
  // немного "лишних" файлов, не найденных в реестре
  for (let k = 0; k < randInt(0, 2); k++) {
    const [docName] = pick(DOC_CATALOG)
    await put(api, nsID, m.id_files, {
      object: objectID,
      recognized_name: docName + ' (без номера)',
      match_status: 'not_in_registry',
      uploaded_at: isoDate(randInt(1, 180)),
    })
    mismatchedCount++
  }

  await put(api, nsID, m.id_check_runs, {
    object: objectID,
    run_date: isoDate(randInt(0, 30)),
    total_registry_docs: registryIDs.length,
    uploaded_docs: uploadedCount,
    missing_docs: Math.max(registryIDs.length - uploadedCount - mismatchedCount, 0),
    mismatched_docs: mismatchedCount,
    overall_status: (mismatchedCount === 0 && uploadedCount === registryIDs.length) ? 'success' : 'fail',
  })

  // ---- Сравнение ПД/РД ----
  const comparisonCount = chance(0.25) ? 2 : 1
  for (let c = 0; c < comparisonCount; c++) {
    const status = pick(['done', 'done', 'done', 'processing', 'new', 'failed'])
    const totalPagesPd = randInt(15, 120)
    const totalPagesRd = totalPagesPd + randInt(-3, 3)
    let similarity = null
    let matching = null
    let differing = null
    if (status === 'done') {
      similarity = round2(randInt(850, 1000) / 10)
      differing = randInt(0, 6)
      matching = Math.max(totalPagesPd - differing, 0)
    } else if (status === 'failed') {
      similarity = round2(randInt(0, 400) / 10)
      differing = randInt(5, 15)
      matching = Math.max(totalPagesPd - differing, 0)
    }
    const title = `Сравнение ПД/РД — ${o.name}${comparisonCount > 1 ? ` (этап ${c + 1})` : ''}`
    const comparisonID = await put(api, nsID, m.pd_rd_comparisons, {
      title,
      object: objectID,
      status,
      similarity_percent: similarity,
      total_pages_pd: totalPagesPd,
      total_pages_rd: totalPagesRd,
      matching_pages: matching,
      differing_pages: differing,
    })
    // ~30% of pairs are real чертежи (synthetic ASCII DXF, see dxfFile() in
    // filegen.mjs) instead of docx placeholders, so extract_attachment_text
    // and the 'drawing' discrepancy_type below have actual drawing content
    // to work with rather than being permanently unreachable.
    const isDrawingPair = chance(0.3)
    const dxfDrift = isDrawingPair && chance(0.5) ? randInt(50, 400) : 0
    let pdFile, rdFile
    if (isDrawingPair) {
      const dxfMeta = { sheetIndex: randInt(0, 5), sheetNo: c + 1, sheetsTotal: comparisonCount, drift: dxfDrift }
      pdFile = dxfFile(`${o.code}-PD-${c}`, 'pd', o, dxfMeta)
      rdFile = dxfFile(`${o.code}-RD-${c}`, 'rd', o, dxfMeta)
    } else {
      pdFile = docFile(`${o.code}-PD-${c}`, pdRdParagraphs('pd', o, { totalPages: totalPagesPd }))
      rdFile = docFile(`${o.code}-RD-${c}`, pdRdParagraphs('rd', o, { totalPages: totalPagesRd }))
    }
    const [pdAtt, rdAtt] = await Promise.all([
      uploadAttachment(base, token, nsID, m.pd_rd_comparisons, comparisonID, 'pd_file', pdFile),
      uploadAttachment(base, token, nsID, m.pd_rd_comparisons, comparisonID, 'rd_file', rdFile),
    ])
    await patchFields(api, nsID, m.pd_rd_comparisons, comparisonID, { pd_file: pdAtt, rd_file: rdAtt })

    if (differing) {
      for (let d = 0; d < Math.min(differing, 4); d++) {
        await put(api, nsID, m.pd_rd_discrepancies, {
          comparison: comparisonID,
          // pd_file/rd_file (docx or dxf, see above) always render as a
          // single page/sheet in the viewer, regardless of the fictional
          // total_pages_pd metadata above. A random page_number up to that
          // fake page count reads as a broken/nonexistent page reference
          // once you actually open the comparison. Page 1 is the only page
          // that ever really exists, so that's what demo data should point at.
          page_number: 1,
          // 'drawing' only makes sense when the pair is actually a dxf
          // (isDrawingPair) — otherwise there's no drawing content for a
          // "чертёж" finding to plausibly refer to.
          discrepancy_type: pick(isDrawingPair
            ? ['content', 'numbering', 'missing_page', 'extra_page', 'drawing']
            : ['content', 'numbering', 'missing_page', 'extra_page']),
          severity: pick(['low', 'low', 'medium', 'high']),
        })
      }
    }
  }

  // ---- Сметы и ВОИСР ----
  const smetaNumber = `С-${o.code}`
  const smetaItems = []
  const workPick = shuffle(WORK_CATALOG).slice(0, randInt(4, 7))
  for (let idx = 0; idx < workPick.length; idx++) {
    const [workName, unit, volMin, volMax, priceMin, priceMax] = workPick[idx]
    const volume = randInt(volMin, volMax)
    const unitPrice = randInt(priceMin, priceMax)
    const totalAmount = round2(volume * unitPrice)
    const smetaID = await put(api, nsID, m.smeta_items, {
      object: objectID,
      smeta_number: smetaNumber,
      item_number: String(idx + 1),
      work_name: workName,
      unit,
      volume,
      unit_price: unitPrice,
      total_amount: totalAmount,
    })
    smetaItems.push({ smetaID, workName, unit, volume, unitPrice, totalAmount })
  }

  const voisrPick = shuffle(smetaItems).slice(0, Math.min(randInt(2, 4), smetaItems.length))
  for (let idx = 0; idx < voisrPick.length; idx++) {
    const s = voisrPick[idx]
    const jitter = chance(0.6) ? 1 : (1 + (chance(0.5) ? 1 : -1) * randInt(3, 20) / 100)
    const voisrVolume = round2(s.volume * jitter)
    const voisrAmount = round2(voisrVolume * s.unitPrice)
    const voisrID = await put(api, nsID, m.voisr_items, {
      object: objectID,
      position_number: String(idx + 1),
      smeta_ref: s.smetaID,
      work_name: s.workName,
      volume: voisrVolume,
      amount: voisrAmount,
    })
    const notFound = chance(0.05)
    const volumeDiff = notFound ? null : round2(voisrVolume - s.volume)
    const amountDiff = notFound ? null : round2(voisrAmount - s.totalAmount)
    const amountDiffPct = notFound || !s.totalAmount ? null : round2((amountDiff / s.totalAmount) * 100)
    const status = notFound ? 'not_found' : (Math.abs(amountDiffPct) < 1 ? 'match' : 'mismatch')
    await put(api, nsID, m.voisr_check_results, {
      voisr_item: voisrID,
      smeta_volume: notFound ? null : s.volume,
      smeta_amount: notFound ? null : s.totalAmount,
      volume_diff: volumeDiff,
      amount_diff: amountDiff,
      amount_diff_percent: amountDiffPct,
      status,
    })
  }

  return objectID
}

async function main () {
  const applied = JSON.parse(readFileSync(join(HERE, 'applied.json'), 'utf8'))
  const token = await mintToken()
  const base = await detectBase(token)
  const api = apiFactory(base, token)
  const nsID = applied.namespaceID
  const m = applied.modules
  console.log('API', base, 'ns', nsID)

  const existing = setOf(await api('GET', `/namespace/${nsID}/module/${m.objects}/record/?limit=1000`))
  console.log('existing objects:', existing.length)
  if (existing.length >= TARGET_OBJECTS && !FORCE) {
    console.log(`already have ${existing.length} >= ${TARGET_OBJECTS}; nothing to do (set SEED_FORCE=1 to add more)`)
    return
  }
  const startIndex = existing.length + 1
  const toCreate = TARGET_OBJECTS - existing.length
  console.log(`creating ${toCreate} objects (${startIndex}..${startIndex + toCreate - 1}) with concurrency ${CONCURRENCY}`)

  const ctx = { api, base, token, nsID, m }
  let done = 0
  await mapPool(Array.from({ length: toCreate }, (_, k) => startIndex + k), CONCURRENCY, async (i) => {
    await seedObject(ctx, i)
    done++
    if (done % 10 === 0) console.log(`  ... ${done}/${toCreate} objects`)
  })

  console.log(`done. seeded ${toCreate} objects with documents in namespace ${nsID}`)
}

main().catch(e => { console.error(e); process.exit(1) })
