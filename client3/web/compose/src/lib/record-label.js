// Names of records for places where a record is shown as one short line - the
// nodes of the RecordGraph block. A module can have a template
// ("{{position_number}} · {{work_name}}"); without one the name is guessed
// from the fields. Everything here is plain data in and out.

export const MAX_LABEL = 48
export const MAX_DETAIL = 80

const hidden = ['Record', 'File', 'User', 'Geometry']
const systemDates = ['createdAt', 'updatedAt', 'deletedAt']

const isEmpty = v => v === undefined || v === null || v === '' || (Array.isArray(v) && v.length === 0)

export function clip (text, max = MAX_LABEL) {
  const t = String(text).replace(/\s+/g, ' ').trim()
  return t.length > max ? `${t.slice(0, max - 1)}…` : t
}

export function fieldOf (module, name) {
  return ((module && module.fields) || []).find(f => f.name === name)
}

// the raw value of a field of a record, system fields included
export function rawValue (record, name) {
  if (!record) return undefined
  if (name === 'recordID') return record.recordID
  if (systemDates.includes(name)) return record[name]
  return record.values ? record.values[name] : undefined
}

function formatDate (v, field, locale) {
  const d = new Date(v)
  if (Number.isNaN(d.getTime())) return String(v)

  const o = (field && field.options) || {}
  if (o.onlyDate) return d.toLocaleDateString(locale)
  if (o.onlyTime) return d.toLocaleTimeString(locale, { hour: '2-digit', minute: '2-digit' })
  return d.toLocaleString(locale, { dateStyle: 'short', timeStyle: 'short' })
}

function formatOne (field, v, locale) {
  if (isEmpty(v)) return ''
  const kind = field ? field.kind : 'String'
  const o = (field && field.options) || {}

  switch (kind) {
    case 'Select': {
      const hit = (o.options || []).find(x => String(x.value) === String(v))
      return hit ? String(hit.text) : String(v)
    }
    case 'DateTime':
      return formatDate(v, field, locale)
    case 'Number': {
      const n = Number(v)
      if (!Number.isFinite(n)) return String(v)
      const digits = Number.isInteger(o.precision) && o.precision >= 0 ? o.precision : 6
      return `${o.prefix || ''}${n.toLocaleString(locale, { maximumFractionDigits: digits })}${o.suffix || ''}`
    }
    case 'Bool': {
      const on = v === true || v === 'true' || v === '1' || v === 1
      return on ? (o.trueLabel || '✓') : (o.falseLabel || '✗')
    }
    case 'Record':
    case 'File':
    case 'User':
    case 'Geometry':
      return '' // only ids: nothing a person can read
    default:
      return typeof v === 'object' ? '' : String(v).trim()
  }
}

// A value as a person reads it: a Select by its text, a date in the user's
// format, a number with separators. Several values (a multi-value field) are
// joined, the first three.
export function formatValue (field, value, { locale } = {}) {
  const sys = field ? field : (systemDates.includes(value) ? { kind: 'DateTime' } : undefined)
  const values = Array.isArray(value) ? value : [value]
  const parts = values.map(v => formatOne(sys, v, locale)).filter(Boolean)
  if (parts.length <= 3) return parts.join(', ')
  return `${parts.slice(0, 3).join(', ')} +${parts.length - 3}`
}

export function valueText (record, module, name, opts) {
  const field = fieldOf(module, name) || (systemDates.includes(name) ? { name, kind: 'DateTime' } : undefined)
  return formatValue(field, rawValue(record, name), opts)
}

// Joins that a missing value leaves dangling: "1.1 ·  " "· name" "a ·  · b"
const separators = '·,;:|/\\-–—'
const dangling = new RegExp(`(?:\\s*[${separators.replace(/[\\\]-]/g, '\\$&')}]\\s*){2,}`, 'g')
const edges = new RegExp(`^[\\s${separators.replace(/[\\\]-]/g, '\\$&')}]+|[\\s${separators.replace(/[\\\]-]/g, '\\$&')}]+$`, 'g')

function tidy (text) {
  return text
    .replace(dangling, match => ` ${match.trim().charAt(0)} `)
    .replace(edges, '')
    .replace(/\s{2,}/g, ' ')
    .trim()
}

// "{{position_number}} · {{work_name}}" -> "1.1 · Plastering". A field without a
// value takes its separator with it. Returns '' when nothing was filled.
export function renderTemplate (template, record, module, opts) {
  if (!template || !String(template).trim()) return ''

  let filled = 0
  const text = String(template).replace(/\{\{\s*([\w.]+)\s*\}\}/g, (_, name) => {
    const v = valueText(record, module, name, opts)
    if (v) filled++
    return v
  })

  return filled ? tidy(text) : ''
}

const mainName = [/^(name|title|label|subject|caption)$/i, /(^|_)(name|title)$/i]
const numberName = /(^|_)(number|num|no|code)$/i
const longName = /(text|description|notes?|comment|content|body|summary|remark)/i

const isShortString = f => f.kind === 'String' && !f.isMulti && !longName.test(f.name)

// Without a template: the best guess of what a record is called. A name, with
// its number when it has one; else a short text; else the Select/date values;
// else the id.
export function autoLabel (record, module, opts) {
  const fields = (module && module.fields) || []
  const text = f => valueText(record, module, f.name, opts)
  const strings = fields.filter(isShortString).filter(f => text(f))

  let main
  for (const re of mainName) {
    main = strings.find(f => re.test(f.name))
    if (main) break
  }

  const number = strings.find(f => f !== main && numberName.test(f.name) && text(f).length <= 16)

  if (main && number) return `${text(number)} · ${text(main)}`
  if (main) return text(main)
  if (number) return text(number)
  // any other text field: only if it reads as a name, not as a paragraph
  const plain = strings.find(f => text(f).length <= MAX_DETAIL)
  if (plain) return text(plain)

  const picks = [
    ...fields.filter(f => f.kind === 'Select' && text(f)).slice(0, 2),
    ...fields.filter(f => f.kind === 'DateTime' && text(f)).slice(0, 1),
  ]
  if (picks.length) return picks.map(text).join(' · ')

  return `#${record ? record.recordID : ''}`
}

// The name of a record: the module's template, else the label field of the
// Record field it was reached by, else a guess. full is not shortened.
export function labelFor (record, module, { template = '', preferredField = '', locale } = {}) {
  const opts = { locale }

  let full = renderTemplate(template, record, module, opts)
  if (!full && preferredField) full = valueText(record, module, preferredField, opts)
  if (!full) full = autoLabel(record, module, opts)

  return { label: clip(full), full }
}

// The fields shown in a tooltip: filled ones a person can read, in the order
// of the module, up to max. What the name already says is not repeated.
export function details (record, module, { max = 4, locale, name = '' } = {}) {
  const out = []
  for (const f of (module && module.fields) || []) {
    if (out.length >= max) break
    if (hidden.includes(f.kind)) continue

    const v = formatValue(f, rawValue(record, f.name), { locale })
    if (!v || (name && name.includes(v))) continue

    out.push({ label: f.label || f.name, value: clip(v, MAX_DETAIL) })
  }
  return out
}
