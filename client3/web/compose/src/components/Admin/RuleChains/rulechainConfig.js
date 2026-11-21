export function parseConfigText (text) {
  const str = (text || '').trim()
  if (!str) return { ok: true, value: {} }
  try {
    let value = JSON.parse(str)
    if (typeof value === 'string') {
      try { value = JSON.parse(value) } catch (e) { /* keep string */ }
    }
    if (value === null || typeof value !== 'object' || Array.isArray(value)) {
      return { ok: true, value: {} }
    }
    return { ok: true, value }
  } catch (e) {
    return { ok: false, error: e.message, value: null }
  }
}

export function stringifyConfig (obj) {
  try {
    return JSON.stringify(obj || {}, null, 2)
  } catch (e) {
    return '{}'
  }
}

export function isFieldVisible (field, scope = {}) {
  const vis = field?.visibleIf
  if (!vis || typeof vis !== 'object') return true
  return Object.entries(vis).every(([key, vals]) => {
    const allowed = Array.isArray(vals) ? vals : [vals]
    return allowed.includes(scope?.[key])
  })
}

export function isEmptyValue (value) {
  if (value === undefined || value === null || value === '') return true
  if (Array.isArray(value) && value.length === 0) return true
  if (typeof value === 'object' && !Array.isArray(value) && Object.keys(value).length === 0) return true
  return false
}

export function readFieldValue (field, scope = {}) {
  if (Object.prototype.hasOwnProperty.call(scope, field.key)) {
    return scope[field.key]
  }
  if (field.default !== undefined && field.widget !== 'keymap' && field.widget !== 'stringlist' && field.widget !== 'objectlist' && field.widget !== 'json') {
    return field.default
  }
  return undefined
}

export function scopeWithDefaults (config, fields = []) {
  const src = config && typeof config === 'object' && !Array.isArray(config) ? { ...config } : {}
  for (const field of fields) {
    if (src[field.key] === undefined && field.default !== undefined) {
      src[field.key] = field.default
    }
  }
  return src
}

export function applyTypeDefaults (current, fields = []) {
  const src = current && typeof current === 'object' && !Array.isArray(current) ? current : {}
  const next = {}
  for (const field of fields) {
    if (Object.prototype.hasOwnProperty.call(src, field.key)) {
      next[field.key] = src[field.key]
    } else if (field.default !== undefined) {
      next[field.key] = field.default
    }
  }
  return next
}

export const FIELD_WIDGET_PRESETS = {
  body: { widget: 'textarea', rows: 8, template: true },
  prompt: { widget: 'textarea', rows: 6, template: true },
  payload: { widget: 'textarea', rows: 4 },
  code: { widget: 'code', lang: 'javascript', rows: 14 },
  fields: { widget: 'keymap' },
  headers: { widget: 'keymap' },
}

export function applyFieldPreset (field, nodeType = '') {
  if (!field || !field.key) return field
  let out = { ...field }
  const widgetPreset = FIELD_WIDGET_PRESETS[field.key]
  if (widgetPreset) out = { ...out, ...widgetPreset }
  const listPreset = OBJECT_LIST_PRESETS[field.key]
  if (listPreset) {
    out.widget = 'objectlist'
    out.itemFields = Array.isArray(field.itemFields) && field.itemFields.length ? field.itemFields : listPreset
  }
  if (out.key === 'code') {
    out.widget = 'code'
    out.lang = nodeType === 'gonec' ? 'golang' : (out.lang || 'javascript')
  }
  return out
}

export const OBJECT_LIST_PRESETS = {
  bands: [
    { key: 'name', widget: 'string', label: 'Name', required: true, placeholder: 'low' },
    { key: 'max', widget: 'number', label: 'Max', help: 'Inclusive upper bound' },
  ],
  factors: [
    { key: 'field', widget: 'string', label: 'Field', required: true },
    { key: 'weight', widget: 'number', label: 'Weight' },
    { key: 'max', widget: 'number', label: 'Max' },
    { key: 'invert', widget: 'bool', label: 'Invert' },
  ],
}

export function inferField (key, value) {
  const preset = applyFieldPreset({ key, label: key }, '')
  if (preset.widget && preset.widget !== 'string') {
    return { label: key, ...preset, key }
  }
  if (OBJECT_LIST_PRESETS[key]) {
    return { key, widget: 'objectlist', label: key, itemFields: OBJECT_LIST_PRESETS[key] }
  }
  if (typeof value === 'boolean') return { key, widget: 'bool', label: key }
  if (typeof value === 'number') return { key, widget: 'number', label: key }
  if (Array.isArray(value)) {
    if (value.every((v) => typeof v === 'string' || v === '')) {
      return { key, widget: 'stringlist', label: key }
    }
    if (value.length && value[0] && typeof value[0] === 'object' && !Array.isArray(value[0])) {
      return { key, widget: 'objectlist', label: key, itemFields: inferFieldsFromConfig(value[0]) }
    }
    return { key, widget: 'objectlist', label: key, itemFields: OBJECT_LIST_PRESETS[key] || [] }
  }
  if (value && typeof value === 'object') return { key, widget: 'keymap', label: key }
  if (typeof value === 'string' && (value.includes('\n') || value.length > 120)) {
    return { key, widget: 'textarea', label: key, rows: 6 }
  }
  return applyFieldPreset({ key, widget: 'string', label: key })
}

export function inferFieldsFromConfig (obj) {
  if (!obj || typeof obj !== 'object' || Array.isArray(obj)) return []
  return Object.keys(obj).map((key) => inferField(key, obj[key]))
}

export function fieldsFromNodeType (nt) {
  if (!nt || typeof nt !== 'object') return []
  if (Array.isArray(nt.configFields) && nt.configFields.length) return nt.configFields
  const schema = nt.configSchema
  if (schema && typeof schema === 'object' && !Array.isArray(schema)) {
    return Object.entries(schema).map(([key, help]) => ({
      key,
      widget: 'string',
      label: key,
      help: typeof help === 'string' ? help : '',
    }))
  }
  return []
}

export function resolveFields (schemaFields, config, nodeType = '') {
  const list = Array.isArray(schemaFields) && schemaFields.length
    ? schemaFields
    : inferFieldsFromConfig(config)
  return list.map((f) => applyFieldPreset(f, nodeType))
}

export function serializeConfig (config, fields = []) {
  const src = config && typeof config === 'object' && !Array.isArray(config) ? config : {}
  const known = new Set(fields.map((f) => f.key))
  const scoped = scopeWithDefaults(src, fields)
  const out = {}

  for (const [key, value] of Object.entries(src)) {
    if (!known.has(key)) out[key] = value
  }

  for (const field of fields) {
    if (!isFieldVisible(field, scoped)) continue
    if (Object.prototype.hasOwnProperty.call(src, field.key)) {
      const value = src[field.key]
      if (isEmptyValue(value) && !field.required) continue
      out[field.key] = value
      continue
    }
    if (field.required && field.default !== undefined) {
      out[field.key] = field.default
    }
  }

  return out
}
