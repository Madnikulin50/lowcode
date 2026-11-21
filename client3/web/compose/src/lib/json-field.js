const VARIANT_NAMES = ['primary', 'secondary', 'success', 'danger', 'warning', 'info', 'dark', 'light']

export function parseJSONValue (raw) {
  if (raw == null || raw === '') return { ok: true, value: null, empty: true }
  if (typeof raw === 'object') return { ok: true, value: raw, empty: isEmptyJSON(raw) }
  const s = String(raw).trim()
  if (!s) return { ok: true, value: null, empty: true }
  try {
    const value = JSON.parse(s)
    return { ok: true, value, empty: isEmptyJSON(value) }
  } catch (err) {
    return { ok: false, value: null, empty: false, error: err, raw: s }
  }
}

export function isEmptyJSON (value) {
  if (value == null) return true
  if (Array.isArray(value)) return value.length === 0
  if (typeof value === 'object') return Object.keys(value).length === 0
  if (typeof value === 'string') return value.trim() === ''
  return false
}

export function isJSONDisplay (field) {
  const mode = field?.options?.displayType
  return mode === 'json' || mode === 'ports'
}

export function jsonOptions (field) {
  const o = field?.options || {}
  const legacyPorts = o.displayType === 'ports'
  return {
    layout: o.jsonLayout || (legacyPorts ? 'chips' : 'chips'),
    template: o.jsonTemplate || (legacyPorts ? '{{port}}/{{proto}} {{service}}' : ''),
    fields: splitFields(o.jsonFields) || (legacyPorts ? ['port', 'proto', 'service'] : []),
    variantField: o.jsonVariantField || (legacyPorts ? 'port' : ''),
    variants: Array.isArray(o.jsonVariants) ? o.jsonVariants : (legacyPorts ? defaultPortVariants() : []),
  }
}

function defaultPortVariants () {
  return [
    { value: '21', variant: 'danger' },
    { value: '23', variant: 'danger' },
    { value: '22', variant: 'success' },
    { value: '80', variant: 'info' },
    { value: '443', variant: 'success' },
    { value: '3389', variant: 'warning' },
    { value: '445', variant: 'warning' },
  ]
}

export function splitFields (raw) {
  if (Array.isArray(raw)) return raw.map(s => String(s).trim()).filter(Boolean)
  if (typeof raw !== 'string' || !raw.trim()) return []
  return raw.split(/[,;\s]+/).map(s => s.trim()).filter(Boolean)
}

export function asItems (value) {
  if (value == null) return []
  if (Array.isArray(value)) return value
  return [value]
}

function interpolate (template, item) {
  if (!template) return ''
  const obj = (item && typeof item === 'object' && !Array.isArray(item)) ? item : { value: item }
  return template.replace(/\{\{\s*([\w.]+)\s*\}\}/g, (_, key) => {
    const val = key.split('.').reduce((acc, part) => (acc == null ? acc : acc[part]), obj)
    if (val == null || typeof val === 'object') return ''
    return String(val)
  }).replace(/\s+/g, ' ').trim()
}

export function detectFields (items) {
  const keys = []
  const seen = new Set()
  for (const item of items) {
    if (!item || typeof item !== 'object' || Array.isArray(item)) continue
    for (const key of Object.keys(item)) {
      if (seen.has(key)) continue
      seen.add(key)
      keys.push(key)
    }
  }
  return keys
}

export function formatItem (item, opts = {}) {
  if (item == null) return ''
  if (typeof item !== 'object') return String(item)
  if (Array.isArray(item)) return item.map(v => formatItem(v, opts)).join(', ')
  if (opts.template) {
    const text = interpolate(opts.template, item)
    if (text) return text
  }
  const fields = (opts.fields && opts.fields.length) ? opts.fields : Object.keys(item)
  const parts = fields.map(key => {
    const val = item[key]
    if (val == null || val === '' || typeof val === 'object') return ''
    return String(val)
  }).filter(Boolean)
  return parts.join(' · ')
}

export function itemTooltip (item) {
  if (item == null) return ''
  if (typeof item !== 'object') return String(item)
  try {
    return JSON.stringify(item, null, 2)
  } catch {
    return String(item)
  }
}

export function itemVariant (item, opts = {}) {
  const rules = opts.variants || []
  if (!rules.length) return 'secondary'
  const field = opts.variantField
  const haystack = []
  if (field && item && typeof item === 'object' && !Array.isArray(item)) {
    haystack.push(item[field])
  } else if (item && typeof item === 'object' && !Array.isArray(item)) {
    haystack.push(...Object.values(item))
  } else {
    haystack.push(item)
  }
  const texts = haystack.filter(v => v != null && typeof v !== 'object').map(v => String(v).toLowerCase())
  for (const rule of rules) {
    const match = String(rule?.value || '').trim().toLowerCase()
    const variant = VARIANT_NAMES.includes(rule?.variant) ? rule.variant : 'secondary'
    if (!match) continue
    if (texts.some(t => t === match || t.includes(match))) return variant
  }
  return 'secondary'
}

export function tableRows (value, opts = {}) {
  const items = asItems(value).filter(item => item != null && item !== '')
  const fields = (opts.fields && opts.fields.length) ? opts.fields : detectFields(items)
  const columns = fields.length ? fields : ['value']
  const rows = items.map(item => {
    if (item && typeof item === 'object' && !Array.isArray(item)) {
      const cells = {}
      for (const col of columns) cells[col] = formatCell(item[col])
      return { cells, raw: item }
    }
    return { cells: { [columns[0]]: formatCell(item) }, raw: item }
  })
  return { columns, rows }
}

function formatCell (val) {
  if (val == null) return ''
  if (typeof val === 'object') {
    try { return JSON.stringify(val) } catch { return String(val) }
  }
  return String(val)
}

export function stringifyJSON (value) {
  if (value == null) return ''
  if (typeof value === 'string') return value
  try {
    return JSON.stringify(value)
  } catch {
    return ''
  }
}

export function prettyJSON (value) {
  if (value == null) return ''
  try {
    return JSON.stringify(value, null, 2)
  } catch {
    return String(value)
  }
}

export function blankRow (fields) {
  const row = {}
  for (const key of fields) row[key] = ''
  if (!fields.length) row.value = ''
  return row
}

export const JSON_VARIANTS = VARIANT_NAMES
