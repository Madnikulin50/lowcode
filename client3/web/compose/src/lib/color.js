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

// Default palette used when a theme does not define a named color
// (Bootstrap / sb-admin colors). Themes may omit keys such as `info`
// or `dark`; without this fallback the raw name would end up in the
// style attribute as an invalid CSS color and the browser would drop it.
export const DEFAULT_THEME_VALUES = {
  primary: '#4e73df',
  secondary: '#858796',
  success: '#1cc88a',
  info: '#36b9cc',
  warning: '#f6c23e',
  danger: '#e74a3b',
  light: '#f8f9fc',
  'extra-light': '#f8f9fc',
  dark: '#5a5c69',
  white: '#FFFFFF',
  black: '#0B344E',
  'body-bg': '#F3F5F7',
}

export function themeColor (value, themes) {
  if (!value) return undefined
  if (value[0] === '#') return value
  const parsed = (themes || [])
    .filter(theme => theme.id !== 'general')
    .map(theme => ({ id: theme.id, values: typeof theme.values === 'string' ? JSON.parse(theme.values) : theme.values }))
  return parsed[0]?.values?.[value] || DEFAULT_THEME_VALUES[value] || value
}

export function badgeGradient (bg, direction = '180deg') {
  if (!bg || bg[0] !== '#') return undefined
  const light = lightenHex(bg, 0.35)
  const dark = darkenHex(bg, 0.15)
  if (!light || !dark) return undefined
  return `linear-gradient(${direction}, ${light} 0%, ${bg} 55%, ${dark} 100%)`
}
