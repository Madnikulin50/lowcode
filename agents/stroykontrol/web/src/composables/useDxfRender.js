// DXF gets client-side rasterization: no CAD rasterizer on the server host,
// so it's parsed and drawn to a <canvas> right here in the browser, producing
// one page image (a DXF's ENTITIES section is one model-space sheet — no
// multi-sheet/layout concept, unlike PDF/DOCX). Only straight-line geometry
// (LINE/LWPOLYLINE/POLYLINE), circles/arcs and TEXT/MTEXT labels are drawn —
// covers what dxfFile() in filegen.mjs generates and the bulk of a typical 2D
// чертёж's linework; splines, hatches, dimensions-as-blocks and inserted
// blocks are silently skipped rather than attempted badly. Ported verbatim
// from the original vanilla-JS index.html.
import { ACI_COLORS, DXF_DEFAULT_STROKE } from '../constants.js'

// Same group-code walk as extractDXFText in server/pkg/rag/parser_cad.go —
// kept as a *separate* pass (not fused with the geometry parser below) so
// the "Текст" view mode shows exactly the text the backend's own extractor
// fed into the det/AI comparison, not a reformatted version of it.
export function extractDxfText (src) {
  const lines = src.replace(/\r\n/g, '\n').split('\n')
  const out = []
  let entity = ''
  let pending = ''
  const flush = () => { if (pending) { out.push(pending); pending = '' } }
  for (let i = 0; i + 1 < lines.length; i += 2) {
    const code = lines[i].trim()
    const val = lines[i + 1].trim()
    if (code === '0') { flush(); entity = val.toUpperCase(); continue }
    if (code === '1' || code === '3' || code === '4') {
      if (!val) continue
      if (entity === 'MTEXT' && pending) pending += ' ' + val
      else { flush(); pending = val }
    } else if (code === '2') {
      if (val && ['LAYER', 'TABLE', 'BLOCK', 'STYLE'].includes(entity)) out.push(entity + ': ' + val)
    } else if (code === '8') {
      if (val && ['TEXT', 'MTEXT', 'DIMENSION', 'INSERT'].includes(entity)) out.push('LAYER ' + val)
    }
  }
  flush()
  const seen = new Set()
  return out.filter(s => { const t = s.trim(); if (!t || seen.has(t)) return false; seen.add(t); return true }).join('\n')
}

// Geometry-only pass: walks the same group-code stream but only cares about
// ENTITIES-section drawables, collecting points/radius/angle/text per entity
// so drawDxfPage() below can rasterize them. Also collects the TABLES
// section's LAYER colors (name -> ACI color) along the way — entities here
// carry no explicit per-entity color (group 62), they're BYLAYER like real
// DXF exports default to, so drawDxfPage needs the layer table to know what
// color each one actually is.
export function parseDxfEntities (src) {
  const lines = src.replace(/\r\n/g, '\n').split('\n')
  const DRAWABLE = new Set(['LINE', 'LWPOLYLINE', 'POLYLINE', 'CIRCLE', 'ARC', 'TEXT', 'MTEXT'])
  const entities = []
  const layerColors = {}
  let cur = null
  let curLayerDef = null
  for (let i = 0; i + 1 < lines.length; i += 2) {
    const code = lines[i].trim()
    const val = lines[i + 1].trim()
    if (code === '0') {
      if (cur && cur.pts.length) entities.push(cur)
      if (curLayerDef && curLayerDef.name) layerColors[curLayerDef.name] = curLayerDef.color
      curLayerDef = val === 'LAYER' ? { name: null, color: null } : null
      cur = DRAWABLE.has(val) ? { type: val, pts: [], layer: '0', color: null, text: '', size: null, startAngle: 0, endAngle: 360, _x: null, _y: null, _x2: null, _y2: null } : null
      continue
    }
    if (curLayerDef) {
      if (code === '2') curLayerDef.name = val
      else if (code === '62') curLayerDef.color = parseInt(val, 10)
      continue
    }
    if (!cur) continue
    switch (code) {
      case '8': cur.layer = val; break
      case '62': cur.color = parseInt(val, 10); break
      case '10': cur._x = parseFloat(val); break
      case '20':
        cur._y = parseFloat(val)
        if (cur._x != null && !isNaN(cur._x) && !isNaN(cur._y)) { cur.pts.push([cur._x, cur._y]); cur._x = null }
        break
      // A LINE's *end* point comes as group 11/21 (start is 10/20 above) —
      // without this, every LINE entity ends up with only its start point
      // and gets silently dropped by drawDxfPage's `pts.length >= 2` check.
      case '11': cur._x2 = parseFloat(val); break
      case '21':
        cur._y2 = parseFloat(val)
        if (cur._x2 != null && !isNaN(cur._x2) && !isNaN(cur._y2)) { cur.pts.push([cur._x2, cur._y2]); cur._x2 = null }
        break
      case '40': cur.size = parseFloat(val); break
      case '50': cur.startAngle = parseFloat(val); break
      case '51': cur.endAngle = parseFloat(val); break
      case '1': case '3': case '4':
        if (val) cur.text = cur.text ? cur.text + ' ' + val : val
        break
    }
  }
  if (cur && cur.pts.length) entities.push(cur)
  if (curLayerDef && curLayerDef.name) layerColors[curLayerDef.name] = curLayerDef.color
  return { entities, layerColors }
}

export function dxfBBox (entities) {
  let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity
  const consider = (x, y) => { if (x < minX) minX = x; if (x > maxX) maxX = x; if (y < minY) minY = y; if (y > maxY) maxY = y }
  for (const e of entities) {
    for (const [x, y] of e.pts) consider(x, y)
    if ((e.type === 'CIRCLE' || e.type === 'ARC') && e.pts[0] && e.size) {
      const [cx, cy] = e.pts[0]
      consider(cx - e.size, cy - e.size); consider(cx + e.size, cy + e.size)
    }
  }
  if (!isFinite(minX)) return { minX: 0, minY: 0, maxX: 100, maxY: 100 }
  return { minX, minY, maxX, maxY }
}

export function drawDxfPage (entities, layerColors) {
  const bbox = dxfBBox(entities)
  const w = Math.max(bbox.maxX - bbox.minX, 1)
  const h = Math.max(bbox.maxY - bbox.minY, 1)
  const PAD = 40
  const CW = 1600
  const CH = Math.round(CW * (h / w)) || 1131
  const canvas = document.createElement('canvas')
  canvas.width = CW
  canvas.height = CH
  const ctx = canvas.getContext('2d')
  ctx.fillStyle = '#ffffff'
  ctx.fillRect(0, 0, CW, CH)

  const scale = Math.min((CW - PAD * 2) / w, (CH - PAD * 2) / h)
  const ox = PAD + (CW - PAD * 2 - w * scale) / 2 - bbox.minX * scale
  const oy = CH - (PAD + (CH - PAD * 2 - h * scale) / 2) + bbox.minY * scale // DXF is Y-up, canvas is Y-down
  const tx = (x) => ox + x * scale
  const ty = (y) => oy - y * scale

  for (const e of entities) {
    // BYLAYER: an entity with no explicit group-62 color (the common case —
    // real DXF exports default to it) takes its layer's color from the
    // TABLES section instead of a flat default.
    const aci = e.color != null ? e.color : (layerColors || {})[e.layer]
    const color = ACI_COLORS[aci] || DXF_DEFAULT_STROKE
    ctx.strokeStyle = color
    ctx.fillStyle = color
    ctx.lineWidth = e.layer === 'СТЕНЫ' ? 2.5 : 1
    if (e.type === 'LINE' && e.pts.length >= 2) {
      ctx.beginPath()
      ctx.moveTo(tx(e.pts[0][0]), ty(e.pts[0][1]))
      ctx.lineTo(tx(e.pts[1][0]), ty(e.pts[1][1]))
      ctx.stroke()
    } else if ((e.type === 'LWPOLYLINE' || e.type === 'POLYLINE') && e.pts.length >= 2) {
      ctx.beginPath()
      ctx.moveTo(tx(e.pts[0][0]), ty(e.pts[0][1]))
      for (let i = 1; i < e.pts.length; i++) ctx.lineTo(tx(e.pts[i][0]), ty(e.pts[i][1]))
      ctx.stroke()
    } else if (e.type === 'CIRCLE' && e.pts[0] && e.size) {
      ctx.beginPath()
      ctx.arc(tx(e.pts[0][0]), ty(e.pts[0][1]), e.size * scale, 0, Math.PI * 2)
      ctx.stroke()
    } else if (e.type === 'ARC' && e.pts[0] && e.size != null) {
      const a0 = -(e.startAngle || 0) * Math.PI / 180
      const a1 = -(e.endAngle || 0) * Math.PI / 180
      ctx.beginPath()
      ctx.arc(tx(e.pts[0][0]), ty(e.pts[0][1]), Math.max(e.size, 0.01) * scale, a0, a1, true)
      ctx.stroke()
    } else if ((e.type === 'TEXT' || e.type === 'MTEXT') && e.pts[0] && e.text) {
      const size = Math.max((e.size || 2.5) * scale, 9)
      ctx.font = size + 'px sans-serif'
      ctx.fillText(e.text, tx(e.pts[0][0]), ty(e.pts[0][1]))
    }
  }
  return canvas
}

export async function renderDxfPages (qs, side) {
  const res = await fetch(`api/file?${qs}&side=${side}`)
  if (!res.ok) throw new Error('не удалось получить файл (' + res.status + ')')
  const src = await res.text()
  const { entities, layerColors } = parseDxfEntities(src)
  if (!entities.length) return { images: [], texts: [] }
  const canvas = drawDxfPage(entities, layerColors)
  return { images: [canvas.toDataURL('image/jpeg', 0.92)], texts: [extractDxfText(src)] }
}
