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

// errText turns whatever an error field holds into text: the agent sends a string, the API gateway
// may send an object ({message, …}) — "[object Object]" tells the person nothing.
export function errText (e, status) {
  if (!e) return status ? `Ответ агента не разобран (HTTP ${status}). Возможно, шлюз оборвал долгий запрос.` : ''
  if (typeof e === 'string') return e
  if (e.message) return String(e.message)
  if (e.error) return errText(e.error)
  try { return JSON.stringify(e).slice(0, 240) } catch (x) { return String(e) }
}
