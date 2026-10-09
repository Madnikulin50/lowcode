// XLSX gets client-side rasterization, same as DOCX/DXF: there's no
// spreadsheet renderer on the server host, so the workbook is unzipped and
// parsed right here (JSZip + DOMParser) and each sheet is drawn as a plain
// grid onto a <canvas>. Long sheets are split into pages of ROWS_PER_PAGE
// rows so the pager / overlay / diff modes behave like they do for PDF pages.
// No number formats, merged cells, styles or charts — cell values are shown
// raw, which is exactly what the backend's extractor (parseXLSX in
// server/pkg/rag/parser_xlsx.go) fed into the det/AI comparison.
import JSZip from 'jszip'

const ROWS_PER_PAGE = 50
const MAX_COLS = 40
const ROW_H = 22
const GUTTER_W = 48
const TITLE_H = 30
const MIN_COL_W = 48
const MAX_COL_W = 280
const FONT = '13px sans-serif'

const NS_REL = 'http://schemas.openxmlformats.org/officeDocument/2006/relationships'

function parseXml (src) {
  return new DOMParser().parseFromString(src, 'application/xml')
}

async function readXml (zip, path) {
  const f = zip.file(path) || zip.file(Object.keys(zip.files).find(n => n.toLowerCase() === path.toLowerCase()) || '')
  return f ? parseXml(await f.async('string')) : null
}

function textOf (el) {
  // Shared/inline strings may be split into rich-text runs; phonetic
  // hints (<rPh>) are not part of the visible value.
  return Array.from(el.getElementsByTagName('t'))
    .filter(t => !t.closest('rPh'))
    .map(t => t.textContent)
    .join('')
}

function colIndex (ref) {
  const m = /^([A-Z]+)/i.exec(ref || '')
  if (!m) return -1
  let n = 0
  for (const ch of m[1].toUpperCase()) n = n * 26 + (ch.charCodeAt(0) - 64)
  return n - 1
}

function colName (i) {
  let s = ''
  for (i++; i > 0; i = Math.floor((i - 1) / 26)) s = String.fromCharCode(65 + ((i - 1) % 26)) + s
  return s
}

// Workbook (tab) order + names; falls back to archive order like the backend.
async function sheetRefs (zip) {
  const wb = await readXml(zip, 'xl/workbook.xml')
  const rels = await readXml(zip, 'xl/_rels/workbook.xml.rels')
  const out = []
  if (wb && rels) {
    const targets = {}
    for (const r of rels.getElementsByTagName('Relationship')) {
      let t = (r.getAttribute('Target') || '').replace(/^\//, '')
      if (!t.startsWith('xl/')) t = 'xl/' + t
      targets[r.getAttribute('Id')] = t
    }
    for (const sh of wb.getElementsByTagName('sheet')) {
      const t = targets[sh.getAttributeNS(NS_REL, 'id') || sh.getAttribute('r:id')]
      if (t && /worksheets\//i.test(t)) out.push({ name: sh.getAttribute('name') || '', path: t })
    }
  }
  if (out.length) return out
  return Object.keys(zip.files)
    .filter(n => /^xl\/worksheets\/sheet[^/]*\.xml$/i.test(n))
    .map(path => ({ name: '', path }))
}

// Rows as [{ num, cells: [{ col, value }] }], only rows with a value.
function readRows (doc, shared) {
  const rows = []
  for (const row of doc.getElementsByTagName('row')) {
    const num = Number(row.getAttribute('r')) || rows.length + 1
    const cells = []
    let next = 0
    for (const c of row.getElementsByTagName('c')) {
      const ci = colIndex(c.getAttribute('r'))
      const col = ci >= 0 ? ci : next
      next = col + 1
      const type = c.getAttribute('t')
      const v = c.getElementsByTagName('v')[0]
      let value = ''
      if (type === 'inlineStr') {
        const is = c.getElementsByTagName('is')[0]
        value = is ? textOf(is) : ''
      } else if (v) {
        value = v.textContent
        if (type === 's') value = shared[Number(value)] ?? ''
      }
      value = String(value).trim()
      if (value) cells.push({ col, value })
    }
    if (cells.length) rows.push({ num, cells })
  }
  return rows
}

function drawPage (title, rows, colCount, widths) {
  const canvas = document.createElement('canvas')
  const scale = 1.5
  const w = GUTTER_W + widths.slice(0, colCount).reduce((a, b) => a + b, 0)
  const h = TITLE_H + ROW_H * (rows.length + 1)
  canvas.width = Math.ceil(w * scale)
  canvas.height = Math.ceil(h * scale)
  const ctx = canvas.getContext('2d')
  ctx.scale(scale, scale)
  ctx.fillStyle = '#ffffff'
  ctx.fillRect(0, 0, w, h)
  ctx.textBaseline = 'middle'

  ctx.fillStyle = '#212529'
  ctx.font = 'bold 14px sans-serif'
  ctx.fillText(title, 8, TITLE_H / 2)

  const top = TITLE_H
  ctx.fillStyle = '#f1f3f5'
  ctx.fillRect(0, top, w, ROW_H)
  ctx.fillRect(0, top, GUTTER_W, ROW_H * (rows.length + 1))

  ctx.font = FONT
  ctx.fillStyle = '#6c757d'
  let x = GUTTER_W
  for (let c = 0; c < colCount; c++) {
    ctx.fillText(colName(c), x + 6, top + ROW_H / 2)
    x += widths[c]
  }
  rows.forEach((row, i) => {
    ctx.fillStyle = '#6c757d'
    ctx.fillText(String(row.num), 6, top + ROW_H * (i + 1.5))
    ctx.fillStyle = '#212529'
    for (const cell of row.cells) {
      if (cell.col >= colCount) continue
      const cx = GUTTER_W + widths.slice(0, cell.col).reduce((a, b) => a + b, 0)
      const cy = top + ROW_H * (i + 1)
      ctx.save()
      ctx.beginPath()
      ctx.rect(cx, cy, widths[cell.col], ROW_H)
      ctx.clip()
      ctx.fillText(cell.value, cx + 6, cy + ROW_H / 2)
      ctx.restore()
    }
  })

  ctx.strokeStyle = '#dee2e6'
  ctx.lineWidth = 1
  ctx.beginPath()
  for (let r = 0; r <= rows.length + 1; r++) {
    const y = top + r * ROW_H + 0.5
    ctx.moveTo(0, y)
    ctx.lineTo(w, y)
  }
  x = GUTTER_W
  ctx.moveTo(x + 0.5, top)
  ctx.lineTo(x + 0.5, h)
  for (let c = 0; c < colCount; c++) {
    x += widths[c]
    ctx.moveTo(x + 0.5, top)
    ctx.lineTo(x + 0.5, h)
  }
  ctx.stroke()
  return canvas
}

async function parseXlsx (qs, side) {
  const res = await fetch(`api/file?${qs}&side=${side}`)
  if (!res.ok) throw new Error('не удалось получить файл (' + res.status + ')')
  const zip = await JSZip.loadAsync(await res.arrayBuffer())

  const sst = await readXml(zip, 'xl/sharedStrings.xml')
  const shared = sst ? Array.from(sst.getElementsByTagName('si')).map(textOf) : []

  const measure = document.createElement('canvas').getContext('2d')
  measure.font = FONT

  const sheets = []
  for (const ref of await sheetRefs(zip)) {
    const doc = await readXml(zip, ref.path)
    if (!doc) continue
    const rows = readRows(doc, shared)
    if (!rows.length) continue

    const colCount = Math.min(MAX_COLS, 1 + Math.max(...rows.flatMap(r => r.cells.map(c => c.col))))
    const widths = new Array(colCount).fill(MIN_COL_W)
    for (const row of rows) {
      for (const cell of row.cells) {
        if (cell.col < colCount) {
          widths[cell.col] = Math.min(MAX_COL_W, Math.max(widths[cell.col], measure.measureText(cell.value).width + 12))
        }
      }
    }
    sheets.push({ name: ref.name, label: ref.name || ref.path.replace(/^.*\//, '').replace(/\.xml$/i, ''), rows, colCount, widths })
  }
  return sheets
}

function drawSheets (sheets) {
  const images = []
  const texts = []
  for (const sheet of sheets) {
    const pageCount = Math.ceil(sheet.rows.length / ROWS_PER_PAGE)
    for (let p = 0; p < pageCount; p++) {
      const chunk = sheet.rows.slice(p * ROWS_PER_PAGE, (p + 1) * ROWS_PER_PAGE)
      const title = 'Лист: ' + sheet.label + (pageCount > 1 ? ` (${p + 1}/${pageCount})` : '')
      images.push(drawPage(title, chunk, sheet.colCount, sheet.widths).toDataURL('image/png'))
      // Same line shape as the backend extractor: non-empty cells joined
      // by tabs, plus the sheet header line on the sheet's first page.
      const lines = chunk.map(r => r.cells.map(c => c.value).join('\t'))
      if (p === 0 && sheet.name) lines.unshift('Лист: ' + sheet.name)
      texts.push(lines.join('\n'))
    }
  }
  return { images, texts }
}

export async function renderXlsxPages (qs, side) {
  const sheets = await parseXlsx(qs, side)
  return { ...drawSheets(sheets), sheets }
}

// Column widths come from each file's own content, so the same sheet in PD
// and RD would be drawn with different geometry (one longer cell widens a
// whole column) and the overlay/auto-diff modes would light up everything.
// Give same-named sheets (or same-position ones, when unnamed) one shared
// set of widths and redraw both sides. Mutates and returns both infos.
export function alignXlsxSides (pd, rd) {
  if (!pd?.sheets || !rd?.sheets) return
  const key = (s, i) => s.name || '#' + i
  const rdByKey = new Map(rd.sheets.map((s, i) => [key(s, i), s]))
  pd.sheets.forEach((a, i) => {
    const b = rdByKey.get(key(a, i))
    if (!b) return
    const n = Math.max(a.colCount, b.colCount)
    const widths = Array.from({ length: n }, (_, c) => Math.max(a.widths[c] || MIN_COL_W, b.widths[c] || MIN_COL_W))
    a.colCount = b.colCount = n
    a.widths = b.widths = widths
  })
  for (const info of [pd, rd]) {
    const { images, texts } = drawSheets(info.sheets)
    info.images = images
    info.texts = texts
    info.pageCount = images.length
  }
}
