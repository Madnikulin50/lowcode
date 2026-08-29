// Branding JSON uses `black`, not `dark`. Select badges still store/default
// to the Bootstrap leftover name "dark", which is not a valid CSS color.
const THEME_COLOR_ALIASES = {
  dark: 'black',
}

/**
 * Resolve a stored theme key or hex into a CSS color.
 * Theme keys like "dark" / "extra-light" become var(--…) so they follow
 * light/dark mode. "dark" aliases to --black (the actual branding text color).
 */
export function resolveThemeColor (value, themeValues) {
  if (value == null) return undefined
  const v = String(value).trim()
  if (!v) return undefined
  if (v[0] === '#' || v.startsWith('rgb') || v.startsWith('hsl') || v.startsWith('var(')) return v

  const alias = THEME_COLOR_ALIASES[v] || v
  if (themeValues) {
    if (themeValues[v]) return themeValues[v]
    if (themeValues[alias]) return themeValues[alias]
  }

  if (v === 'dark' || alias === 'black') {
    return 'var(--dark, var(--black, #212529))'
  }
  return `var(--${v})`
}

export function firstThemeValues (themes) {
  if (!Array.isArray(themes) || !themes.length) return undefined
  const list = themes.filter((t) => t && t.id !== 'general')
  const raw = (list[0] || themes[0] || {}).values
  if (!raw) return undefined
  if (typeof raw === 'string') {
    try { return JSON.parse(raw) } catch { return undefined }
  }
  return raw
}

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
