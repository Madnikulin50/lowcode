const PALETTES = [
  ['#1d4ed8', '#38bdf8'],
  ['#0f766e', '#2dd4bf'],
  ['#6d28d9', '#a78bfa'],
  ['#b45309', '#fbbf24'],
  ['#be123c', '#fb7185'],
  ['#166534', '#4ade80'],
  ['#1e3a8a', '#60a5fa'],
  ['#9f1239', '#f472b6'],
  ['#155e75', '#22d3ee'],
  ['#7c2d12', '#fb923c'],
]

const SLUG_ICONS = {
  cmdb: 'network-wired',
  crm: 'handshake',
  loop: 'project-diagram',
  service: 'cog',
  services: 'cog',
  inventory: 'boxes-stacked',
  stock: 'boxes-stacked',
  hr: 'people-group',
  people: 'people-group',
  shop: 'shop',
  store: 'shop',
  finance: 'wallet',
  billing: 'file-invoice',
  wiki: 'book',
  docs: 'book',
  knowledge: 'book',
  admin: 'cog',
  system: 'server',
  iot: 'plug',
  security: 'lock',
  vuln: 'bug',
  network: 'network-wired',
  device: 'server',
  analytics: 'chart-pie',
  report: 'chart-bar',
  sales: 'bullseye',
  support: 'comments',
}

const FALLBACK_ICONS = [
  'cube',
  'sitemap',
  'globe',
  'folder',
  'star',
  'building',
  'briefcase',
  'database',
  'sitemap',
  'compass',
]

function hashString (s) {
  let h = 0
  for (let i = 0; i < s.length; i++) {
    h = ((h << 5) - h) + s.charCodeAt(i)
    h |= 0
  }
  return Math.abs(h)
}

export function namespaceInitials (ns) {
  const base = String(ns?.name || ns?.slug || '').trim()
  if (!base) return '?'
  const words = base.split(/\s+/).filter(Boolean)
  const letters = words.map(w => Array.from(w)[0]).filter(Boolean)
  const pick = letters.length >= 2 ? letters.slice(0, 2) : Array.from(base).slice(0, 2)
  return pick.join('').toUpperCase()
}

export function namespaceIconName (ns) {
  const slug = String(ns?.slug || '').toLowerCase()
  if (SLUG_ICONS[slug]) return SLUG_ICONS[slug]
  for (const [key, icon] of Object.entries(SLUG_ICONS)) {
    if (slug.includes(key)) return icon
  }
  const key = slug || String(ns?.name || '')
  return FALLBACK_ICONS[hashString(key) % FALLBACK_ICONS.length]
}

export function namespacePalette (ns) {
  const key = String(ns?.slug || ns?.name || ns?.namespaceID || '')
  return PALETTES[hashString(key) % PALETTES.length]
}

export function namespaceImageSrc (ns) {
  const meta = ns?.meta || {}
  if (meta.icon) return meta.icon
  if (meta.logoEnabled && meta.logo) return meta.logo
  return ''
}
