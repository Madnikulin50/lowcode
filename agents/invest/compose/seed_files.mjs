/**
 * Generate demo DOCX/XLSX and attach them to invest `documents` records.
 *
 * Idempotent: skips records that already have a file unless SEED_FILES_FORCE=1.
 *
 *   node seed_files.mjs
 */
import { crc32 } from 'node:zlib'
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

import { mintToken, detectBase, apiFactory, setOf } from './helpers.mjs'

const HERE = dirname(fileURLToPath(import.meta.url))

const XLSX_TYPES = new Set(['ESTIMATE', 'KS2', 'KS3'])

function recID (row) {
  return String(row.recordID || row.ID || '')
}

function recVal (row, name) {
  const hits = (row.values || []).filter(v => v.name === name && v.value)
  if (!hits.length) return ''
  return hits.map(v => String(v.value)).join('\n')
}

function xml (s) {
  return String(s ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

function zipStore (files) {
  const locals = []
  const centrals = []
  let offset = 0
  for (const [name, body] of Object.entries(files)) {
    const data = Buffer.isBuffer(body) ? body : Buffer.from(body, 'utf8')
    const nameBuf = Buffer.from(name, 'utf8')
    const crc = crc32(data) >>> 0
    const local = Buffer.alloc(30)
    local.writeUInt32LE(0x04034b50, 0)
    local.writeUInt16LE(20, 4)
    local.writeUInt16LE(0, 8) // STORE
    local.writeUInt32LE(crc, 14)
    local.writeUInt32LE(data.length, 18)
    local.writeUInt32LE(data.length, 22)
    local.writeUInt16LE(nameBuf.length, 26)
    const file = Buffer.concat([local, nameBuf, data])
    locals.push(file)

    const central = Buffer.alloc(46)
    central.writeUInt32LE(0x02014b50, 0)
    central.writeUInt16LE(20, 4)
    central.writeUInt16LE(20, 6)
    central.writeUInt32LE(crc, 16)
    central.writeUInt32LE(data.length, 20)
    central.writeUInt32LE(data.length, 24)
    central.writeUInt16LE(nameBuf.length, 28)
    central.writeUInt32LE(offset, 42)
    centrals.push(Buffer.concat([central, nameBuf]))
    offset += file.length
  }
  const dir = Buffer.concat(centrals)
  const eocd = Buffer.alloc(22)
  eocd.writeUInt32LE(0x06054b50, 0)
  eocd.writeUInt16LE(centrals.length, 8)
  eocd.writeUInt16LE(centrals.length, 10)
  eocd.writeUInt32LE(dir.length, 12)
  eocd.writeUInt32LE(offset, 16)
  return Buffer.concat([...locals, dir, eocd])
}

function buildDocx (paragraphs) {
  const body = paragraphs.map(p => {
    const t = xml(p)
    return `<w:p><w:r><w:t xml:space="preserve">${t}</w:t></w:r></w:p>`
  }).join('')
  const document = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>${body}<w:sectPr/></w:body>
</w:document>`
  return zipStore({
    '[Content_Types].xml': `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`,
    '_rels/.rels': `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`,
    'word/_rels/document.xml.rels': `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"/>`,
    'word/document.xml': document,
  })
}

function colLetter (i) {
  let n = i
  let s = ''
  while (n >= 0) {
    s = String.fromCharCode(65 + (n % 26)) + s
    n = Math.floor(n / 26) - 1
  }
  return s
}

function buildXlsx (sheetName, rows) {
  const strings = []
  const indexOf = (t) => {
    const s = String(t ?? '')
    let i = strings.indexOf(s)
    if (i < 0) {
      i = strings.length
      strings.push(s)
    }
    return i
  }
  const sheetRows = rows.map((row, ri) => {
    const cells = row.map((val, ci) => {
      const r = `${colLetter(ci)}${ri + 1}`
      if (typeof val === 'number' && Number.isFinite(val)) {
        return `<c r="${r}"><v>${val}</v></c>`
      }
      return `<c r="${r}" t="s"><v>${indexOf(val)}</v></c>`
    }).join('')
    return `<row r="${ri + 1}">${cells}</row>`
  }).join('')
  const sst = strings.map(s => `<si><t xml:space="preserve">${xml(s)}</t></si>`).join('')
  const name = xml(sheetName || 'Лист1').slice(0, 31)
  return zipStore({
    '[Content_Types].xml': `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
  <Override PartName="/xl/sharedStrings.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sharedStrings+xml"/>
</Types>`,
    '_rels/.rels': `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`,
    'xl/workbook.xml': `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets><sheet name="${name}" sheetId="1" r:id="rId1"/></sheets>
</workbook>`,
    'xl/_rels/workbook.xml.rels': `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/sharedStrings" Target="sharedStrings.xml"/>
</Relationships>`,
    'xl/sharedStrings.xml': `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="${strings.length}" uniqueCount="${strings.length}">${sst}</sst>`,
    'xl/worksheets/sheet1.xml': `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>${sheetRows}</sheetData></worksheet>`,
  })
}

function money (n) {
  return Number(n).toLocaleString('ru-RU')
}

function docxParagraphs (ctx) {
  const { type, title, number, project, notes, due } = ctx
  const commonHead = [
    title,
    `Документ № ${number}`,
    `Проект: ${project || '—'}`,
    `Дата: ${due || '01.09.2026'}`,
    '',
  ]
  switch (type) {
    case 'TEO':
      return [...commonHead,
        '1. Цель и границы',
        `Технико-экономическое обоснование «${title}». Заказчик — АО РосИнвест, разработчик — ООО ПроектГидро.`,
        'Мощность объекта 40–200 МВт. Горизонт расчёта — 20 лет. Валюта — рубль РФ.',
        '2. Капитальные затраты (CAPEX)',
        `Ориентировочный CAPEX ${money(2_400_000_000)} руб. без НДС, в т.ч. оборудование 62%, СМР 28%, прочее 10%.`,
        'Резерв непредвиденных — 8%. НДС 20% сверху.',
        '3. Эффекты',
        'NPV при ставке 12% — 410 млн руб., IRR — 14,6%, простой срок окупаемости — 9,5 лет.',
        '4. Риски',
        'Сетка присоединения, цена оборудования, сроки землеотвода. Меры: запрос ТУ, опцион поставщика, параллельные площадки.',
        notes ? `Примечание: ${notes}` : 'Статус: демо-редакция для ИИ-summary.',
        'Стороны обязуются не разглашать финансовую модель до инвесткомитета.',
      ]
    case 'CONTRACT':
      return [...commonHead,
        'ДОГОВОР ПОДРЯДА (демо)',
        'Заказчик: АО РосИнвест, ИНН 7700000002, в лице генерального директора.',
        'Подрядчик: ООО Генподрядчик-1, ИНН 7700000001.',
        `Предмет: выполнение работ по проекту «${project}» в соответствии с ПСД и сметой.`,
        `Цена договора: ${money(1_850_000_000)} руб. с НДС 20%, твердая, кроме согласованных RFC.`,
        'Срок: с 15.01.2026 по 30.06.2028. Гарантия — 24 месяца с даты акта ввода.',
        'Порядок расчётов: аванс 15%, ежемесячные КС-2/КС-3, удержание 5% до гарантийного срока.',
        'Ответственность: пеня 0,1% в день, не более 10% цены. Подсудность — арбитражный суд г. Москвы.',
        'Эскалация цены — только через change_request (RFC) с пересчётом EAC.',
        notes ? `Особые условия: ${notes}` : 'Особые условия: человеко-часы PMO не входят в цену.',
      ]
    case 'PSD':
      return [...commonHead,
        'Проектная документация. Стадия «П».',
        'Раздел 1. Пояснительная записка. Раздел 2. Схема планировочной организации.',
        'Раздел 5. Сведения об инженерном оборудовании. Раздел 12. Смета (отдельным файлом).',
        'Главный инженер проекта: Иванов П.С. Шифр: П-2026-01.',
        'Климат: I климатический район. Сейсмика 6 баллов. Грунты — суглинки тугопластичные.',
        'Основные решения: монолитный ж/б каркас, фундаменты на сваях 18 м, КНС и ОРУ 110 кВ.',
        'Экспертиза: промышленная безопасность + Главгосэкспертиза (демо).',
        notes ? `Замечание: ${notes}` : 'Комплект готов к направлению на экспертизу.',
      ]
    case 'SURVEY':
      return [...commonHead,
        'Программа / отчёт инженерных изысканий.',
        'Состав: инженерно-геодезические, геологические, гидрометеорологические, экологические.',
        'Трасса / площадка: коридор 500 м, скважины 152 шт., глубина до 25 м.',
        'Выявлено: участки ООПТ, высокий УГВ, торф до 1,8 м на ПК 24–31.',
        'Рекомендации: обход заказника на 12 км, замена фундамента на сваи в зоне торфа.',
        'Исполнитель: ООО ГеоИзыскания, ИНН 7700000004. Лицензия (демо) ГИ-2024-088.',
        notes ? `Комментарий: ${notes}` : 'Полевые работы выполнены на 55%.',
      ]
    case 'ORDER':
      return [...commonHead,
        'ПРИКАЗ',
        `О введении / приостановке работ по проекту «${project}».`,
        '1. Утвердить график и ответственных: PMO — куратор, ГИП — технический контроль.',
        '2. Запретить оплату КС-2 без исполнительной документации.',
        '3. Контроль исполнения — дирекция капитального строительства.',
        'Основание: решение инвесткомитета, протокол (демо) № 14-ИК.',
        notes ? notes : 'Приказ вступает в силу с даты подписания УКЭП.',
      ]
    case 'AS_BUILT':
      return [...commonHead,
        'Исполнительная документация.',
        'Состав: акты скрытых работ, исполнительные схемы, паспорта изделий, журналы работ.',
        'Отклонения от проекта: отметка верха плиты −12 мм (в допуске СП 70.13330).',
        'Применённые материалы: бетон B25 ГОСТ 26633, арматура А500С ГОСТ 34028.',
        'Подписи: производитель работ, технадзор заказчика, авторский надзор.',
        notes ? notes : 'Комплект для закрытия КС-2.',
      ]
    case 'PNR':
      return [...commonHead,
        'Акт пусконаладочных работ.',
        'Объект: ОРУ / котлоагрегат / узел учёта (демо).',
        'Испытания: сопротивление изоляции, проверка защит, пробный пуск 72 часа.',
        'Результат: соответствует паспортным данным. Замечания устранены.',
        'Допуск: комплект документов для Ростехнадзора прилагается.',
        notes ? notes : 'Акт является основанием для ввода в эксплуатацию.',
      ]
    case 'PERMIT':
      return [...commonHead,
        'Разрешение / согласование (демо-форма).',
        `Объект: ${project}. Заявитель: АО РосИнвест.`,
        'Орган: Ростехнадзор / администрация / сетевая компания — по типу объекта.',
        'Срок действия: 24 месяца. Условия: соблюдение границ землеотвода и ТУ присоединения.',
        notes ? notes : 'Черновик заявления, подпись руководителя не проставлена.',
      ]
    case 'RD':
      return [...commonHead,
        'Рабочая документация. Марка КМ / КЖ / ТМ.',
        'Лист 1 — общие данные. Лист 2 — план на отм. 0.000. Лист 3 — разрезы 1-1, 2-2.',
        'Спецификация: балки 35Б1 — 48 т, анкерные болты М36 — 220 шт.',
        'Штамп: ГИП Иванов, разработал Петров, проверил Сидоров, дата 15.08.2026.',
        notes ? notes : 'Комплект для производства работ 1-й очереди.',
      ]
    default:
      return [...commonHead, notes || 'Демо-документ инвестиционного проекта.', 'Текст подготовлен для проверки извлечения и summary.']
  }
}

function xlsxRows (ctx) {
  const { type, title, number, project, notes } = ctx
  if (type === 'KS2') {
    const lines = [
      ['Акт о приёмке выполненных работ КС-2', '', '', '', '', ''],
      ['Документ', number, 'Проект', project, 'Заказчик', 'АО РосИнвест'],
      ['Подрядчик', 'ООО Генподрядчик-1', 'Договор', number.replace(/^KS2/, 'ГП'), 'Период', 'март 2026'],
      [],
      ['№', 'Наименование работ', 'Ед.', 'Кол-во', 'Цена, руб.', 'Сумма, руб.'],
      [1, 'Фундамент монолитный ж/б', 'м³', 420, 18500, 7770000],
      [2, 'Арматурные каркасы А500С', 'т', 38.4, 62000, 2380800],
      [3, 'Обратная засыпка', 'м³', 960, 850, 816000],
      [4, 'Гидроизоляция фундаментов', 'м²', 540, 2100, 1134000],
      ['', 'Итого без НДС', '', '', '', 12104800],
      ['', 'НДС 20%', '', '', '', 2420960],
      ['', 'Всего с НДС', '', '', '', 14525760],
      [],
      ['Примечание', notes || 'Объёмы по исполнительной схеме. Демо-файл для summary.'],
    ]
    return { sheet: 'КС-2', rows: lines }
  }
  if (type === 'KS3') {
    return {
      sheet: 'КС-3',
      rows: [
        ['Справка о стоимости выполненных работ КС-3', '', '', ''],
        ['Документ', number, 'Проект', project],
        ['Начало работ', '15.01.2026', 'Отчётный месяц', 'март 2026'],
        [],
        ['Показатель', 'С начала строительства', 'С начала года', 'За месяц'],
        ['Выполнено без НДС', 186400000, 41200000, 12104800],
        ['НДС 20%', 37280000, 8240000, 2420960],
        ['Выполнено с НДС', 223680000, 49440000, 14525760],
        ['Аванс зачтён', 18600000, 0, 0],
        ['К оплате', 205080000, 49440000, 14525760],
        [],
        ['Примечание', notes || 'Демо КС-3. Оплата — после согласования КС-2.'],
      ],
    }
  }
  // ESTIMATE
  return {
    sheet: 'Смета',
    rows: [
      ['Сводный сметный расчёт', title, '', ''],
      ['Номер', number, 'Проект', project],
      ['Базис', 'ТЕР-2022, индексы 1 кв. 2026', 'НДС', '20%'],
      [],
      ['Глава', 'Наименование', 'СМР', 'Оборудование', 'Прочее', 'Итого'],
      [1, 'Подготовка территории', 12000000, 0, 1800000, 13800000],
      [2, 'Основные объекты', 410000000, 980000000, 22000000, 1412000000],
      [4, 'Энергетическое хозяйство', 64000000, 185000000, 9000000, 258000000],
      [8, 'Временные здания и сооружения', 18500000, 0, 0, 18500000],
      [9, 'Прочие работы и затраты', 24000000, 0, 11000000, 35000000],
      [10, 'Содержание дирекции (2%)', 0, 0, 34756000, 34756000],
      [12, 'Резерв непредвиденных (8%)', 0, 0, 141800000, 141800000],
      ['', 'Итого без НДС', 528500000, 1165000000, 220056000, 1913556000],
      ['', 'НДС 20%', '', '', '', 382711200],
      ['', 'Всего с НДС', '', '', '', 2296267200],
      [],
      ['Вывод', notes || 'Резерв 8% — на пересмотр. Демо-смета для извлечения текста.'],
    ],
  }
}

function safeName (number, ext) {
  const base = String(number || 'doc').replace(/[\\/:*?"<>|]+/g, '-').trim() || 'doc'
  return `${base}.${ext}`
}

function fileFor (ctx) {
  if (XLSX_TYPES.has(ctx.type)) {
    const { sheet, rows } = xlsxRows(ctx)
    return {
      name: safeName(ctx.number, 'xlsx'),
      mime: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
      buf: buildXlsx(sheet, rows),
    }
  }
  return {
    name: safeName(ctx.number, 'docx'),
    mime: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    buf: buildDocx(docxParagraphs(ctx)),
  }
}

async function uploadAttachment (base, token, nsID, moduleID, recordID, fieldName, file) {
  const form = new FormData()
  form.append('upload', new Blob([file.buf], { type: file.mime }), file.name)
  form.append('recordID', String(recordID))
  form.append('fieldName', fieldName)
  const res = await fetch(`${base}/namespace/${nsID}/module/${moduleID}/record/attachment`, {
    method: 'POST',
    headers: { Authorization: 'Bearer ' + token },
    body: form,
    signal: AbortSignal.timeout(30000),
  })
  const text = await res.text()
  let data
  try { data = JSON.parse(text) } catch { data = { raw: text } }
  if (!res.ok || data.error) {
    const err = data.error?.message || data.error || text.slice(0, 400)
    throw new Error(`upload ${file.name} → ${res.status}: ${typeof err === 'string' ? err : JSON.stringify(err)}`)
  }
  const body = data.response !== undefined ? data.response : data
  return String(body.attachmentID || body.attachment?.attachmentID || '')
}

async function patchFileField (api, nsID, moduleID, recordID, attachmentID) {
  const rec = await api('GET', `/namespace/${nsID}/module/${moduleID}/record/${recordID}`)
  const current = {}
  for (const v of rec.values || []) {
    if (!v.name) continue
    if (v.name === 'file') continue
    current[v.name] = v.value == null ? '' : String(v.value)
  }
  const payload = Object.entries(current)
    .filter(([, v]) => v !== '' && v != null)
    .map(([name, value]) => ({ name, value }))
  payload.push({ name: 'file', value: String(attachmentID) })
  await api('POST', `/namespace/${nsID}/module/${moduleID}/record/${recordID}`, {
    values: payload,
    updatedAt: rec.updatedAt,
  })
}

export async function seedDocumentFiles (api, nsID, modules, { token, base, force = false } = {}) {
  if (!token) token = await mintToken()
  if (!base) base = await detectBase(token)
  if (!api) api = apiFactory(base, token)

  const typeRows = setOf(await api('GET', `/namespace/${nsID}/module/${modules.document_types}/record/?limit=500`))
  const typeByID = new Map()
  for (const r of typeRows) typeByID.set(recID(r), recVal(r, 'code') || recVal(r, 'name'))

  const projectRows = setOf(await api('GET', `/namespace/${nsID}/module/${modules.projects}/record/?limit=500`))
  const projectByID = new Map()
  for (const r of projectRows) projectByID.set(recID(r), recVal(r, 'name') || recVal(r, 'code'))

  const docs = setOf(await api('GET', `/namespace/${nsID}/module/${modules.documents}/record/?limit=500`))
  let uploaded = 0
  let skipped = 0
  for (const doc of docs) {
    const id = recID(doc)
    const number = recVal(doc, 'number') || id
    if (!force && recVal(doc, 'file')) {
      skipped++
      continue
    }
    const type = typeByID.get(recVal(doc, 'doc_type')) || 'TEO'
    const ctx = {
      type,
      title: recVal(doc, 'title') || number,
      number,
      project: projectByID.get(recVal(doc, 'project')) || '',
      notes: recVal(doc, 'notes'),
      due: recVal(doc, 'due_date'),
    }
    const file = fileFor(ctx)
    const attID = await uploadAttachment(base, token, nsID, modules.documents, id, 'file', file)
    if (!attID) throw new Error(`no attachmentID for ${number}`)
    await patchFileField(api, nsID, modules.documents, id, attID)
    uploaded++
    console.log('  file', number, file.name, attID)
  }
  console.log(`seeded document files (${uploaded} uploaded, ${skipped} already had file)`)
  return { uploaded, skipped }
}

async function main () {
  const applied = JSON.parse(readFileSync(join(HERE, 'applied.json'), 'utf8'))
  const token = await mintToken()
  const base = await detectBase(token)
  const api = apiFactory(base, token)
  const force = process.env.SEED_FILES_FORCE === '1' || process.argv.includes('--force')
  console.log('API', base, 'ns', applied.namespaceID, force ? 'FORCE' : '')
  await seedDocumentFiles(api, applied.namespaceID, applied.modules, { token, base, force })
}

const isDirect = process.argv[1] && fileURLToPath(import.meta.url) === process.argv[1]
if (isDirect) {
  main().catch(err => {
    console.error(err)
    process.exit(1)
  })
}
