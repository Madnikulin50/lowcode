export function normalizeXYWH (xywh) {
  const src = Array.isArray(xywh) ? xywh : []
  const num = (v, fallback, min) => {
    const n = Number(v)
    if (!Number.isFinite(n) || n < min) return fallback
    return n
  }
  return [
    num(src[0], 0, 0),
    num(src[1], 0, 0),
    num(src[2], 20, 1),
    num(src[3], 15, 1),
  ]
}

export function xywhSignature (blocks = []) {
  return blocks.map((b, i) => {
    const hidden = b?.meta?.hidden ? 1 : 0
    const invisible = b?.meta?.invisible ? 1 : 0
    const [x, y, w, h] = normalizeXYWH(b?.xywh)
    return `${i}:${hidden}:${invisible}:${x},${y},${w},${h}`
  }).join('|')
}
