export function hexToRgb (hex) {
  const m = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex)
  if (!m) return undefined
  return m.slice(1).map(h => parseInt(h, 16))
}

export function rgbToHex ([r, g, b]) {
  return '#' + [r, g, b].map(v => Math.max(0, Math.min(255, v)).toString(16).padStart(2, '0')).join('')
}

export function lightenHex (hex, amount = 0.25) {
  const rgb = hexToRgb(hex)
  if (!rgb) return undefined
  return rgbToHex(rgb.map(v => Math.round(v + (255 - v) * amount)))
}

export function darkenHex (hex, amount = 0.25) {
  const rgb = hexToRgb(hex)
  if (!rgb) return undefined
  return rgbToHex(rgb.map(v => Math.round(v * (1 - amount))))
}

export function badgeGradient (bg, direction = '180deg') {
  if (!bg || bg[0] !== '#') return undefined
  const light = lightenHex(bg, 0.35)
  const dark = darkenHex(bg, 0.15)
  if (!light || !dark) return undefined
  return `linear-gradient(${direction}, ${light} 0%, ${bg} 55%, ${dark} 100%)`
}
