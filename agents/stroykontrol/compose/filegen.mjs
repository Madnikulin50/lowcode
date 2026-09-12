/**
 * Minimal in-memory DOCX builder (no deps) — same STORE-only zip writer as
 * agents/invest/compose/seed_files.mjs, trimmed to what stroykontrol seeding needs.
 */
import { crc32 } from 'node:zlib'

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
    locals.push(Buffer.concat([local, nameBuf, data]))

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
    offset += locals[locals.length - 1].length
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

export function buildDocx (paragraphs) {
  const body = paragraphs.map(p => `<w:p><w:r><w:t xml:space="preserve">${xml(p)}</w:t></w:r></w:p>`).join('')
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

function safeName (name) {
  return String(name || 'doc').replace(/[\\/:*?"<>|]+/g, '-').trim() || 'doc'
}

const PD_SECTIONS = [
  'Пояснительная записка',
  'Схема планировочной организации земельного участка',
  'Архитектурные решения',
  'Конструктивные и объёмно-планировочные решения',
  'Сведения об инженерном оборудовании, сетях и мероприятиях',
  'Проект организации строительства',
  'Проект организации работ по сносу или демонтажу объектов',
  'Перечень мероприятий по охране окружающей среды',
  'Мероприятия по обеспечению пожарной безопасности',
  'Мероприятия по обеспечению доступа инвалидов',
  'Смета на строительство',
  'Иная документация в случаях, предусмотренных законодательством',
]

const RD_SECTIONS = [
  'Общие данные',
  'Архитектурно-строительные решения (АР)',
  'Конструкции железобетонные (КЖ)',
  'Конструкции металлические (КМ)',
  'Отопление, вентиляция и кондиционирование (ОВ)',
  'Водоснабжение и канализация (ВК)',
  'Электроснабжение и электроосвещение (ЭОМ)',
  'Слаботочные системы (СС)',
  'Автоматизация и диспетчеризация (АТХ)',
  'Наружные сети и сооружения (НС)',
  'Организация строительства (ПОС/ППР)',
]

/**
 * Paragraph text for a demo ПД/РД docx: title block + table of contents +
 * a sheet-by-sheet listing, padded to at least MIN_LINES paragraphs so the
 * generated file reads as a real multi-page set rather than a stub.
 */
const MIN_LINES = 30

export function pdRdParagraphs (kind, obj, meta = {}) {
  const isPd = kind === 'pd'
  const sections = isPd ? PD_SECTIONS : RD_SECTIONS
  const label = isPd ? 'Проектная документация' : 'Рабочая документация'
  const totalPages = Number(meta.totalPages) || sections.length
  const lines = [
    `${label} — ${obj.name}`,
    `Объект: ${obj.name}`,
    `Адрес: ${obj.address || '—'}`,
    `Шифр объекта: ${obj.code || '—'}`,
    `Комплект: ${isPd ? 'ПД' : 'РД'}, стадия «${isPd ? 'П' : 'Р'}»`,
    `Всего листов: ${totalPages}`,
    `Дата формирования: ${new Date().toLocaleDateString('ru-RU')}`,
    '',
    'Содержание комплекта:',
    ...sections.map((s, i) => `Раздел ${i + 1}. ${s}`),
    '',
    'Ведомость листов основного комплекта:',
  ]
  for (let i = 1; i <= totalPages && lines.length < MIN_LINES + totalPages; i++) {
    lines.push(`Лист ${i} — ${sections[(i - 1) % sections.length]}`)
  }
  while (lines.length < MIN_LINES) {
    lines.push(`Лист ${lines.length} — ${sections[lines.length % sections.length]}`)
  }
  lines.push('', `Демо-файл ${isPd ? 'ПД' : 'РД'}.`)
  return lines
}

export function docFile (baseName, paragraphs) {
  return {
    name: `${safeName(baseName)}.docx`,
    mime: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    buf: buildDocx(paragraphs),
  }
}

export async function uploadAttachment (base, token, nsID, moduleID, recordID, fieldName, file) {
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

export async function patchFields (api, nsID, moduleID, recordID, patch) {
  const rec = await api('GET', `/namespace/${nsID}/module/${moduleID}/record/${recordID}`)
  const current = {}
  for (const v of rec.values || []) {
    if (v.name) current[v.name] = v.value == null ? '' : String(v.value)
  }
  Object.assign(current, patch)
  const payload = Object.entries(current)
    .filter(([, v]) => v !== '' && v != null)
    .map(([name, value]) => ({ name, value }))
  await api('POST', `/namespace/${nsID}/module/${moduleID}/record/${recordID}`, {
    values: payload,
    updatedAt: rec.updatedAt,
  })
}
