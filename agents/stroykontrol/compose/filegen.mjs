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

// ---------------------------------------------------------------------------
// Synthetic DXF (ASCII, R12-compatible) — a real чертёж, not a text stub.
// Lets seed data exercise the actual CAD path: server/pkg/rag's parseDXF
// (server/pkg/rag/parser_cad.go) reads TEXT/MTEXT/LAYER group codes straight
// out of this, same as it would a real exported drawing, so the
// extract_attachment_text rule-chain node (chains.mjs's extractNode) and the
// 'drawing' discrepancy_type get genuine content to work with instead of
// being unreachable (see the comment this replaces in seed.mjs).
// ---------------------------------------------------------------------------

const DXF_LAYERS = [
  ['0', 7], ['РАМКА', 7], ['ОСИ', 1], ['РАЗМЕРЫ', 3], ['ТЕКСТ', 7], ['ШТАМП', 7],
  ['СТЕНЫ', 7], ['ДВЕРИ', 5], ['ОКНА', 4],
]

const DXF_SHEET_SECTIONS = [
  ['АР', 'Архитектурно-строительные решения'],
  ['КЖ', 'Конструкции железобетонные'],
  ['КМ', 'Конструкции металлические'],
  ['ОВ', 'Отопление, вентиляция и кондиционирование'],
  ['ВК', 'Водоснабжение и канализация'],
  ['ЭОМ', 'Электроснабжение и электроосвещение'],
]

const DXF_ROOM_NAMES = [
  'Жилая комната', 'Кухня', 'Прихожая', 'Санузел', 'Кабинет',
  'Коридор', 'Кладовая', 'Гостиная', 'Спальня', 'Гардеробная',
]

function dxfLine (layer, x1, y1, x2, y2) {
  return [
    [0, 'LINE'], [8, layer],
    [10, x1.toFixed(2)], [20, y1.toFixed(2)], [30, '0'],
    [11, x2.toFixed(2)], [21, y2.toFixed(2)], [31, '0'],
  ]
}

function dxfSafeText (str) {
  return String(str ?? '').replace(/[\r\n]+/g, ' ').slice(0, 250)
}

function dxfText (layer, x, y, height, str) {
  return [
    [0, 'TEXT'], [8, layer],
    [10, x.toFixed(2)], [20, y.toFixed(2)], [30, '0'],
    [40, height.toFixed(2)],
    [1, dxfSafeText(str)],
  ]
}

// Closed outline as a single LWPOLYLINE (repeat the first point as the last
// vertex — simplest way to close it without needing the polyline flag).
function dxfPoly (layer, points) {
  const pairs = [[0, 'LWPOLYLINE'], [8, layer]]
  for (const [x, y] of points) pairs.push([10, x.toFixed(2)], [20, y.toFixed(2)])
  return pairs
}

function dxfArc (layer, cx, cy, r, startDeg, endDeg) {
  return [
    [0, 'ARC'], [8, layer],
    [10, cx.toFixed(2)], [20, cy.toFixed(2)], [30, '0'],
    [40, r.toFixed(2)],
    [50, startDeg.toFixed(2)], [51, endDeg.toFixed(2)],
  ]
}

// A wall segment with a door opening cut into it: two LINE stubs either
// side of the gap (on `wallLayer`, same as the wall they interrupt), plus
// a quarter-circle door-swing ARC hinged at the gap's start and a short
// leaf LINE (both on the 'ДВЕРИ' layer) — the same graphic convention a
// real архитектурные решения floor plan uses for a door in a wall. `frac`
// places the gap along the segment (0=start..1=end); doorLen defaults to
// a plausible door width relative to the pseudo-mm scale the rest of the
// sheet uses (see dxfEntries' span/span2).
function dxfDoorOnSegment (wallLayer, x1, y1, x2, y2, frac, doorLen) {
  const dx = x2 - x1
  const dy = y2 - y1
  const len = Math.hypot(dx, dy)
  const dl = doorLen || 90
  if (len < 2 || dl >= len) return [dxfLine(wallLayer, x1, y1, x2, y2)]
  const ux = dx / len
  const uy = dy / len
  const gs = Math.max(0, Math.min(1, frac)) * (len - dl)
  const ge = gs + dl
  const gx1 = x1 + ux * gs
  const gy1 = y1 + uy * gs
  const gx2 = x1 + ux * ge
  const gy2 = y1 + uy * ge
  const entities = []
  if (gs > 1) entities.push(dxfLine(wallLayer, x1, y1, gx1, gy1))
  if (len - ge > 1) entities.push(dxfLine(wallLayer, gx2, gy2, x2, y2))
  const nx = -uy
  const ny = ux
  const startDeg = Math.atan2(uy, ux) * 180 / Math.PI
  entities.push(dxfArc('ДВЕРИ', gx1, gy1, dl, startDeg, startDeg + 90))
  entities.push(dxfLine('ДВЕРИ', gx1, gy1, gx1 + nx * dl, gy1 + ny * dl))
  return entities
}

/**
 * Actual чертёж content — outer walls, a couple of interior partition
 * walls with door openings, window ticks on the street-facing wall, and a
 * name+area label per room — instead of just axes/dimension numbers with
 * no real geometry behind them. `drift` (already 0 for ПД, see dxfEntries)
 * nudges one interior wall's position so a differing pair gets a genuine
 * *visual* mismatch (a moved wall, a resized room), not just a changed
 * number floating with nothing to point at.
 */
function dxfFloorPlan (sheetIndex, x0, y0, x1, y1, drift) {
  const entities = []
  const WALL = 'СТЕНЫ'

  entities.push(dxfPoly(WALL, [[x0, y0], [x1, y0], [x1, y1], [x0, y1], [x0, y0]]))

  // main entrance, centered on the south (bottom) wall
  const entranceFrac = 0.5 - 45 / (x1 - x0)
  entities.push(...dxfDoorOnSegment(WALL, x0, y0, x1, y0, entranceFrac))

  // windows along the north (top, street-facing) wall
  const winCount = 2 + (sheetIndex % 2)
  for (let i = 1; i <= winCount; i++) {
    const wx = x0 + (x1 - x0) * (i / (winCount + 1))
    entities.push(dxfLine('ОКНА', wx - 35, y1 - 6, wx + 35, y1 - 6))
    entities.push(dxfLine('ОКНА', wx - 35, y1 + 6, wx + 35, y1 + 6))
  }

  const midY = (y0 + y1) / 2
  entities.push(dxfLine(WALL, x0, midY, x1, midY))

  const rowsCols = [2 + (sheetIndex % 2), 2 + ((sheetIndex + 1) % 2)]
  const rowRanges = [[y0, midY], [midY, y1]]
  let roomN = 0
  rowRanges.forEach(([ry0, ry1], rowIdx) => {
    const cols = rowsCols[rowIdx]
    const xs = [x0]
    for (let c = 1; c < cols; c++) {
      let x = x0 + (x1 - x0) * (c / cols)
      // Only ever move the first interior wall of the top row — everything
      // else stays byte-for-byte identical between ПД and РД so the diff
      // isolates this one change instead of scattering noise everywhere.
      if (rowIdx === 0 && c === 1) x += drift / 8
      xs.push(x)
    }
    xs.push(x1)
    for (let c = 1; c < cols; c++) {
      entities.push(...dxfDoorOnSegment(WALL, xs[c], ry0, xs[c], ry1, 0.15))
    }
    for (let c = 0; c < cols; c++) {
      const cx = (xs[c] + xs[c + 1]) / 2
      const cy = (ry0 + ry1) / 2
      const widthMm = (xs[c + 1] - xs[c]) * 3
      const heightMm = (ry1 - ry0) * 3
      const areaM2 = ((widthMm / 1000) * (heightMm / 1000)).toFixed(1)
      const name = DXF_ROOM_NAMES[(roomN + sheetIndex) % DXF_ROOM_NAMES.length]
      entities.push(dxfText('ТЕКСТ', cx - 130, cy + 15, 32, name))
      entities.push(dxfText('ТЕКСТ', cx - 130, cy - 30, 28, areaM2 + ' м²'))
      roomN++
    }
  })

  return entities
}

function buildDxf (entries, layers) {
  const pairs = []
  const push = (arr) => arr.forEach(p => pairs.push(p))

  // No HEADER section: extractDXFText (server/pkg/rag/parser_cad.go) isn't
  // SECTION-aware, so a $ACADVER value under group code 1 would otherwise
  // leak into the extracted text/title as a bogus first line. A minimal DXF
  // doesn't need HEADER to be valid or parseable by our own reader.
  push([[0, 'SECTION'], [2, 'TABLES'], [0, 'TABLE'], [2, 'LAYER'], [70, String(layers.length)]])
  for (const [name, color] of layers) {
    push([[0, 'LAYER'], [2, name], [70, '0'], [62, String(color)], [6, 'CONTINUOUS']])
  }
  push([[0, 'ENDTAB'], [0, 'ENDSEC']])

  push([[0, 'SECTION'], [2, 'BLOCKS'], [0, 'ENDSEC']])

  push([[0, 'SECTION'], [2, 'ENTITIES']])
  for (const e of entries) push(e)
  push([[0, 'ENDSEC']])

  push([[0, 'EOF']])

  return pairs.map(([c, v]) => `${c}\n${v}`).join('\n') + '\n'
}

/**
 * Entity list for one sheet (ПД or РД variant). `meta.drift` (mm) is added
 * to the РД span so a differing pair produces an actual numeric mismatch
 * for the deterministic diff / AI verdict to catch — same role as
 * pdRdParagraphs' totalPages nudge, but on drawing content instead of text.
 */
export function dxfEntries (kind, obj, meta = {}) {
  const isPd = kind === 'pd'
  const sheetIndex = meta.sheetIndex || 0
  const [sectionCode, sectionName] = DXF_SHEET_SECTIONS[sheetIndex % DXF_SHEET_SECTIONS.length]
  const sheetNo = meta.sheetNo || 1
  const sheetsTotal = meta.sheetsTotal || sheetNo
  const drift = isPd ? 0 : (meta.drift || 0)
  const span = (meta.baseSpan || 6000) + drift
  const span2 = 3200 + (isPd ? 0 : Math.round((meta.drift || 0) / 2))

  const W = 2970
  const H = 2100
  const entries = []

  // outer + inner frame (рамка)
  entries.push(dxfLine('РАМКА', 0, 0, W, 0))
  entries.push(dxfLine('РАМКА', W, 0, W, H))
  entries.push(dxfLine('РАМКА', W, H, 0, H))
  entries.push(dxfLine('РАМКА', 0, H, 0, 0))
  entries.push(dxfLine('РАМКА', 20, 20, W - 20, 20))
  entries.push(dxfLine('РАМКА', W - 20, 20, W - 20, H - 20))
  entries.push(dxfLine('РАМКА', W - 20, H - 20, 20, H - 20))
  entries.push(dxfLine('РАМКА', 20, H - 20, 20, 20))

  // Structural grid (разбивочные оси) doubles as the building footprint the
  // floor plan below is drawn inside — fixed extents (not scaled by span)
  // so the axis lines always land inside the frame regardless of drift.
  const FX0 = 300
  const FX1 = 2100
  const FY0 = 220
  const FY1 = 1420
  ;['А', 'Б', 'В', 'Г'].forEach((letter, i) => {
    const y = FY0 + i * (FY1 - FY0) / 3
    entries.push(dxfLine('ОСИ', 60, y, W - 60, y))
    entries.push(dxfText('ОСИ', 40, y, 60, letter))
  })
  ;['1', '2', '3', '4'].forEach((num, i) => {
    const x = FX0 + i * (FX1 - FX0) / 3
    entries.push(dxfLine('ОСИ', x, 60, x, H - 60))
    entries.push(dxfText('ОСИ', x, 40, 60, num))
  })

  // dimension text — the values that can drift between ПД and РД
  entries.push(dxfText('РАЗМЕРЫ', FX0, 150, 50, String(span)))
  entries.push(dxfText('РАЗМЕРЫ', 100, FY0, 50, String(span2)))

  // caption above the plan
  entries.push(dxfText('ТЕКСТ', FX0, FY1 + 90, 70, obj.name || 'Объект'))
  entries.push(dxfText('ТЕКСТ', FX0, FY1 + 30, 42, `${sectionCode} — ${sectionName}. План. М 1:100`))

  entries.push(...dxfFloorPlan(sheetIndex, FX0, FY0, FX1, FY1, drift))

  // штамп (title block), bottom-right corner
  const sx = W - 700
  let sy = 260
  const stampLines = [
    obj.name || 'Объект',
    `Шифр: ${obj.code || '—'}`,
    `${sectionCode} — ${sectionName}`,
    `Стадия: ${isPd ? 'П' : 'Р'}`,
    `Лист ${sheetNo} из ${sheetsTotal}`,
    'М 1:100',
    isPd ? 'Проектная документация' : 'Рабочая документация',
  ]
  for (const line of stampLines) {
    entries.push(dxfText('ШТАМП', sx, sy, 40, line))
    sy += 60
  }

  return entries
}

export function dxfFile (baseName, kind, obj, meta = {}) {
  const content = buildDxf(dxfEntries(kind, obj, meta), DXF_LAYERS)
  return {
    name: `${safeName(baseName)}.dxf`,
    mime: 'image/vnd.dxf',
    // utf8, not ascii: Cyrillic title-block/label text is multi-byte —
    // 'ascii' silently truncates each char to its low byte and corrupts it.
    buf: Buffer.from(content, 'utf8'),
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
