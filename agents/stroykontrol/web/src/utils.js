// A box is stored as the JSON text "[x0,y0,x1,y1]" in page fractions.
export function parseBox (s) {
  if (!s) return null
  try {
    const a = JSON.parse(s)
    return Array.isArray(a) && a.length === 4 && a.every(Number.isFinite) ? a : null
  } catch (e) {
    return null
  }
}

export function parseEvidence (s) {
  if (!s) return null
  try {
    return JSON.parse(s)
  } catch (e) {
    return null
  }
}
