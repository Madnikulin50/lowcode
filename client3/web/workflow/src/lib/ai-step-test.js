// Seeds the "test step" dialog from a step's configured arguments.
//
// Constants can be sent as they are. Arguments given as expressions cannot be
// evaluated in the browser (they refer to variables of a running workflow),
// so they are listed for the author to fill with sample values.

const JSON_TYPES = ['KV', 'KVV', 'Vars']

function constant (arg) {
  const raw = arg.value

  if (arg.type === 'Boolean') return raw === true || raw === 'true'

  if (arg.type === 'Integer' || arg.type === 'UnsignedInteger' || arg.type === 'Float') {
    const n = Number(raw)
    return Number.isNaN(n) ? raw : n
  }

  if (JSON_TYPES.includes(arg.type) && typeof raw === 'string') {
    try {
      return JSON.parse(raw)
    } catch (e) {
      return raw
    }
  }

  return raw
}

export function seedTestArgs (args = []) {
  const values = {}
  const expressions = []

  for (const arg of args) {
    if (!arg || !arg.target) continue

    if (arg.valueType === 'expr') {
      if (arg.expr) {
        expressions.push(arg.target)
        values[arg.target] = ''
      }
      continue
    }

    if (arg.value === null || arg.value === undefined || arg.value === '') continue
    values[arg.target] = constant(arg)
  }

  return { values, expressions }
}
