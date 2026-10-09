import { describe, it, expect } from 'vitest'
import { flatten, diff, parseSnapshot } from './json-diff'

describe('json-diff', () => {
  it('flattens nested objects and arrays', () => {
    expect(flatten({ a: 1, b: { c: [true, null], d: {} } })).toEqual({ a: 1, 'b.c.0': true, 'b.c.1': null, 'b.d': '{}' })
  })

  it('reports added, removed and changed paths, sorted', () => {
    const out = diff({ title: 'A', gone: 1, blocks: [{ k: 'x' }] }, { title: 'B', blocks: [{ k: 'x' }, { k: 'y' }] })
    expect(out).toEqual([
      { path: 'blocks.1.k', kind: 'added', after: 'y' },
      { path: 'gone', kind: 'removed', before: 1 },
      { path: 'title', kind: 'changed', before: 'A', after: 'B' },
    ])
  })

  it('ignores timestamps and treats equal documents as no change', () => {
    expect(diff({ x: 1, updatedAt: 'a', page: { createdAt: 'q' } }, { x: 1, updatedAt: 'b', page: { createdAt: 'r' } })).toEqual([])
  })

  it('does not choke on a broken snapshot', () => {
    expect(parseSnapshot('{oops')).toBeNull()
    expect(parseSnapshot('')).toBeNull()
    expect(diff(null, { a: 1 })).toEqual([{ path: 'a', kind: 'added', after: 1 }])
  })
})
