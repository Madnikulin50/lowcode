// Line diff of two texts (longest common subsequence), for comparing prompt
// and skill versions. Returns [{ type: 'same'|'add'|'del', text }] in reading
// order. Texts are small (a prompt is a few KB), so the O(n*m) table is fine;
// beyond MAX_LINES the texts are shown as fully replaced instead of risking a
// huge table.

const MAX_LINES = 2000

export function lineDiff (before = '', after = '') {
  const a = before === '' ? [] : String(before).split('\n')
  const b = after === '' ? [] : String(after).split('\n')

  if (a.length > MAX_LINES || b.length > MAX_LINES) {
    return [...a.map(text => ({ type: 'del', text })), ...b.map(text => ({ type: 'add', text }))]
  }

  const lcs = Array.from({ length: a.length + 1 }, () => new Array(b.length + 1).fill(0))
  for (let i = a.length - 1; i >= 0; i--) {
    for (let j = b.length - 1; j >= 0; j--) {
      lcs[i][j] = a[i] === b[j] ? lcs[i + 1][j + 1] + 1 : Math.max(lcs[i + 1][j], lcs[i][j + 1])
    }
  }

  const out = []
  let i = 0
  let j = 0
  while (i < a.length && j < b.length) {
    if (a[i] === b[j]) {
      out.push({ type: 'same', text: a[i] }); i++; j++
    } else if (lcs[i + 1][j] >= lcs[i][j + 1]) {
      out.push({ type: 'del', text: a[i++] })
    } else {
      out.push({ type: 'add', text: b[j++] })
    }
  }
  while (i < a.length) out.push({ type: 'del', text: a[i++] })
  while (j < b.length) out.push({ type: 'add', text: b[j++] })

  return out
}

export function hasChanges (diff) {
  return diff.some(d => d.type !== 'same')
}
