// Flat comparison of two JSON documents for the version history screen.
//
// Both documents are flattened to "path -> value" (objects and arrays are
// walked, a leaf is a string, number, boolean or null); the result lists the
// paths that were added, removed or changed. Arrays of blocks keep their
// order, so a path looks like "page.blocks.2.options.title".

const MAX_LEAVES = 5000

export function flatten (value, prefix = '', out = {}) {
  if (Object.keys(out).length > MAX_LEAVES) return out

  if (value !== null && typeof value === 'object') {
    const keys = Object.keys(value)
    if (keys.length === 0) {
      if (prefix) out[prefix] = Array.isArray(value) ? '[]' : '{}'
      return out
    }
    for (const k of keys) flatten(value[k], prefix ? `${prefix}.${k}` : k, out)
    return out
  }

  // a lone null/undefined at the top level is "no document", not a value
  if (prefix || (value !== null && value !== undefined)) out[prefix] = value
  return out
}

// diff returns [{ path, kind: 'added'|'removed'|'changed', before, after }]
// sorted by path. Fields that are only bookkeeping are ignored.
export function diff (before, after, { ignore = ['updatedAt', 'createdAt', 'deletedAt'] } = {}) {
  const a = flatten(before)
  const b = flatten(after)
  const skipped = path => ignore.some(i => path === i || path.endsWith(`.${i}`))
  const out = []

  for (const path of new Set([...Object.keys(a), ...Object.keys(b)])) {
    if (skipped(path)) continue

    const inA = path in a
    const inB = path in b

    if (inA && !inB) out.push({ path, kind: 'removed', before: a[path] })
    else if (!inA && inB) out.push({ path, kind: 'added', after: b[path] })
    else if (a[path] !== b[path]) out.push({ path, kind: 'changed', before: a[path], after: b[path] })
  }

  return out.sort((x, y) => (x.path < y.path ? -1 : x.path > y.path ? 1 : 0))
}

export function parseSnapshot (snapshot) {
  try {
    return JSON.parse(snapshot || 'null')
  } catch (e) {
    return null
  }
}
