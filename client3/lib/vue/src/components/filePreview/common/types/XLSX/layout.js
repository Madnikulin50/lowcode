// Reads the parts of an XLSX layout that SheetJS CE does not expose (cell borders,
// default row height / column width). SpreadsheetML is regular enough that a few
// regexes are sufficient and keep this usable outside the browser (no DOMParser).

const DEFAULT_ROW_HEIGHT_PT = 15
const DEFAULT_COL_WIDTH_PX = 64

const BORDER_STYLES = {
  thin: '1px solid',
  medium: '2px solid',
  thick: '3px solid',
  double: '3px double',
  hair: '1px dotted',
  dotted: '1px dotted',
  dashed: '1px dashed',
  dashDot: '1px dashed',
  dashDotDot: '1px dashed',
  slantDashDot: '2px dashed',
  mediumDashed: '2px dashed',
  mediumDashDot: '2px dashed',
  mediumDashDotDot: '2px dashed',
}

const EDGES = { left: 'borderLeft', right: 'borderRight', top: 'borderTop', bottom: 'borderBottom', start: 'borderLeft', end: 'borderRight' }

export function ptToPx (pt) {
  return Math.round(pt * 96 / 72)
}

export function colWidthToPx (col) {
  if (!col) return undefined
  if (col.wpx != null) return Math.round(col.wpx)
  if (col.wch != null) return Math.round(col.wch * 7 + 5)
  if (col.width != null) return Math.round(col.width * 7 + 5)
  return undefined
}

export { DEFAULT_ROW_HEIGHT_PT, DEFAULT_COL_WIDTH_PX }

function unescapeXml (s) {
  return String(s)
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&apos;/g, "'")
    .replace(/&amp;/g, '&')
}

function attrs (tagAttrs = '') {
  const out = {}
  const re = /([\w:.-]+)\s*=\s*("([^"]*)"|'([^']*)')/g
  let m
  while ((m = re.exec(tagAttrs))) {
    const value = unescapeXml(m[3] ?? m[4] ?? '')
    out[m[1]] = value
    // Also index by local name so prefixes (r:id, x:style…) do not matter.
    const local = m[1].includes(':') ? m[1].split(':').pop() : null
    if (local && !(local in out)) out[local] = value
  }
  return out
}

// Matches <prefix:name …/> or <prefix:name …>…</prefix:name>; yields [attrs, inner].
function elements (xml, name) {
  const re = new RegExp(`<(?:[\\w-]+:)?${name}\\b([^>]*?)(?:/>|>([\\s\\S]*?)</(?:[\\w-]+:)?${name}>)`, 'g')
  const out = []
  let m
  while ((m = re.exec(xml))) out.push([attrs(m[1]), m[2] || ''])
  return out
}

function firstBlock (xml, name) {
  const [el] = elements(xml, name)
  return el ? el[1] : ''
}

function colorToCss (colorAttrs) {
  const rgb = colorAttrs?.rgb
  if (rgb && /^[0-9a-f]{6,8}$/i.test(rgb)) return `#${rgb.slice(-6)}`
  return '#000'
}

export function borderToCss (edge) {
  if (!edge || !edge.style || edge.style === 'none') return undefined
  const base = BORDER_STYLES[edge.style] || '1px solid'
  return `${base} ${colorToCss(edge.color)}`
}

export function parseBorders (stylesXml) {
  const borders = elements(firstBlock(stylesXml, 'borders'), 'border').map(([, inner]) => {
    const css = {}
    for (const edge of Object.keys(EDGES)) {
      const [el] = elements(inner, edge)
      if (!el) continue
      const [edgeAttrs, edgeInner] = el
      const [color] = elements(edgeInner, 'color')
      const value = borderToCss({ style: edgeAttrs.style, color: color?.[0] })
      if (value && !css[EDGES[edge]]) css[EDGES[edge]] = value
    }
    return Object.keys(css).length ? css : null
  })

  // cellXfs: style index (the `s` attribute on a cell) → border
  const xfs = elements(firstBlock(stylesXml, 'cellXfs'), 'xf')
    .map(([a]) => borders[Number(a.borderId || 0)] || null)

  return xfs
}

// "B3" → { r: 2, c: 1 }
export function decodeCell (ref) {
  const m = /^\$?([A-Z]+)\$?(\d+)$/i.exec(ref || '')
  if (!m) return null
  let c = 0
  for (const ch of m[1].toUpperCase()) c = c * 26 + (ch.charCodeAt(0) - 64)
  return { r: Number(m[2]) - 1, c: c - 1 }
}

export function parseSheet (sheetXml, xfBorders) {
  const borders = new Map()
  const cellRe = /<(?:[\w-]+:)?c\b([^>]*)>/g
  let m
  while ((m = cellRe.exec(sheetXml))) {
    const a = attrs(m[1])
    if (a.s == null) continue
    const border = xfBorders[Number(a.s)]
    const pos = border && decodeCell(a.r)
    if (pos) borders.set(`${pos.r}:${pos.c}`, border)
  }

  const [fmt] = elements(sheetXml, 'sheetFormatPr')
  const defaultRowHeightPt = Number(fmt?.[0].defaultRowHeight) || DEFAULT_ROW_HEIGHT_PT
  const defaultColWidth = Number(fmt?.[0].defaultColWidth)
  const defaultColWidthPx = defaultColWidth ? Math.round(defaultColWidth * 7 + 5) : DEFAULT_COL_WIDTH_PX

  return { borders, defaultRowHeightPt, defaultColWidthPx }
}

function resolveTarget (target) {
  if (target.startsWith('/')) return target.slice(1)
  const parts = ['xl']
  for (const p of target.split('/')) {
    if (p === '..') parts.pop()
    else if (p && p !== '.') parts.push(p)
  }
  return parts.join('/')
}

/**
 * Returns { [sheetName]: { borders: Map<'r:c', style>, defaultRowHeightPt, defaultColWidthPx } }.
 * Never throws: anything unreadable (xls, ods, broken parts) yields an empty result.
 */
export async function readLayout (buffer) {
  try {
    const { default: JSZip } = await import('jszip')
    const zip = await JSZip.loadAsync(buffer)
    const read = async (path) => (await zip.file(path)?.async('string')) || ''

    const [workbookXml, relsXml, stylesXml] = await Promise.all([
      read('xl/workbook.xml'),
      read('xl/_rels/workbook.xml.rels'),
      read('xl/styles.xml'),
    ])
    if (!workbookXml) return {}

    const targets = Object.fromEntries(elements(relsXml, 'Relationship').map(([a]) => [a.Id, a.Target]))
    const xfBorders = stylesXml ? parseBorders(stylesXml) : []

    const result = {}
    await Promise.all(elements(workbookXml, 'sheet').map(async ([a]) => {
      const target = targets[a['r:id'] || a.id]
      if (!a.name || !target) return
      result[a.name] = parseSheet(await read(resolveTarget(target)), xfBorders)
    }))
    return result
  } catch (err) {
    return {}
  }
}
