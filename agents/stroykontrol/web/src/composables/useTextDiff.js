// Pure line/word diff engine — no DOM, no Vue. Ported verbatim from the
// original vanilla-JS index.html so page-diff output stays byte-identical.

export function diffSeq (a, b) {
  const n = a.length, m = b.length
  const dp = new Array(n + 1)
  for (let i = 0; i <= n; i++) dp[i] = new Uint32Array(m + 1)
  for (let i = n - 1; i >= 0; i--) {
    for (let j = m - 1; j >= 0; j--) {
      dp[i][j] = a[i] === b[j] ? dp[i + 1][j + 1] + 1 : Math.max(dp[i + 1][j], dp[i][j + 1])
    }
  }
  const ops = []
  let i = 0, j = 0
  while (i < n && j < m) {
    if (a[i] === b[j]) { ops.push({ type: 'equal', text: a[i] }); i++; j++ }
    else if (dp[i + 1][j] >= dp[i][j + 1]) { ops.push({ type: 'del', text: a[i] }); i++ }
    else { ops.push({ type: 'add', text: b[j] }); j++ }
  }
  while (i < n) { ops.push({ type: 'del', text: a[i] }); i++ }
  while (j < m) { ops.push({ type: 'add', text: b[j] }); j++ }
  return ops
}

// Splits page text into comparable lines: pdftotext -layout gives PDF pages
// real line breaks; linesFromElement() (useDocxRender.js) gives DOCX pages
// one line per paragraph/row. Internal whitespace runs are collapsed per
// line — the two extraction pipelines produce very different incidental
// spacing for identical wording, so a highlight should track real content
// changes, not that. Runs of 2+ blank lines collapse to one, so pdftotext's
// spacing blank lines don't inflate the diff with noise docx never had.
export function textLines (text) {
  const raw = (text || '').split(/\r?\n/).map(l => l.replace(/\s+/g, ' ').trim())
  const out = []
  for (const l of raw) {
    if (l === '' && out[out.length - 1] === '') continue
    out.push(l)
  }
  while (out.length && out[0] === '') out.shift()
  while (out.length && out[out.length - 1] === '') out.pop()
  return out
}

// Word-level ops for one changed line pair, as spans to insert inside that
// line's row — lets a one-line edit ("2.4 м" → "2.6 м") show exactly which
// words changed instead of tinting the whole line, same as GitHub/GitLab's
// intra-line diff highlighting.
export function wordDiffSpans (delText, addText) {
  const ops = diffSeq(delText.split(/\s+/).filter(Boolean), addText.split(/\s+/).filter(Boolean))
  const delSpans = [], addSpans = []
  for (const op of ops) {
    if (op.type === 'equal') { delSpans.push({ t: op.text, hi: false }); addSpans.push({ t: op.text, hi: false }) }
    else if (op.type === 'del') delSpans.push({ t: op.text, hi: true })
    else addSpans.push({ t: op.text, hi: true })
  }
  return { delSpans, addSpans }
}

// Groups line-level diff ops into alternating context/change blocks, the
// same shape a unified diff's hunks have: a run of untouched lines, then a
// run of removed lines followed by added lines, and so on.
export function buildBlocks (ops) {
  const blocks = []
  let i = 0
  while (i < ops.length) {
    if (ops[i].type === 'equal') {
      const lines = []
      while (i < ops.length && ops[i].type === 'equal') { lines.push(ops[i].text); i++ }
      blocks.push({ kind: 'ctx', lines })
    } else {
      const dels = [], adds = []
      while (i < ops.length && ops[i].type !== 'equal') {
        if (ops[i].type === 'del') dels.push(ops[i].text)
        else adds.push(ops[i].text)
        i++
      }
      blocks.push({ kind: 'change', dels, adds })
    }
  }
  return blocks
}

// A discrepancy's description carries its origin (AI vs deterministic
// checker) as a "[ИИ] "/"[детерм.] " prefix (chains.mjs, compareChain
// finalize) — there's no separate field for it. Pull it back out here so it
// renders as a real tag instead of leading text the eye has to parse every
// time.
export function splitSource (description) {
  const text = description || ''
  if (text.startsWith('[ИИ] ')) return { source: 'ai', text: text.slice(5) }
  if (text.startsWith('[детерм.] ')) return { source: 'det', text: text.slice(10) }
  return { source: null, text }
}
