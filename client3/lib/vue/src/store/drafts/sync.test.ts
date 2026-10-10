import { describe, it, expect } from 'vitest'
import { parseDraftResource, mergeDrafts } from './sync'

describe('parseDraftResource', () => {
  it('reads namespace, module and record', () => {
    expect(parseDraftResource('compose:record/1/22/333')).toEqual({ namespaceID: '1', moduleID: '22', recordID: '333' })
    expect(parseDraftResource('compose:record/1/22/0')?.recordID).toBe('0')
  })

  it('rejects anything else', () => {
    expect(parseDraftResource('')).toBeNull()
    expect(parseDraftResource('compose:record/1/22')).toBeNull()
    expect(parseDraftResource('compose:page/1/2/3')).toBeNull()
  })
})

describe('mergeDrafts', () => {
  const d = (changeID: string, t: number) => ({ changeID, timestamp: new Date(t) })

  it('takes server drafts the browser does not have', () => {
    const r = mergeDrafts([], [d('a', 1)])
    expect(Object.keys(r.merged)).toEqual(['a'])
    expect(r.merged.a.source).toBe('backend')
    expect(r.toUpload).toEqual([])
  })

  it('uploads local drafts the server does not have', () => {
    const r = mergeDrafts([d('a', 1)], [])
    expect(r.merged.a.source).toBe('local')
    expect(r.toUpload.map(x => x.changeID)).toEqual(['a'])
  })

  it('the newer copy wins', () => {
    const newerLocal = mergeDrafts([d('a', 5)], [d('a', 3)])
    expect(newerLocal.merged.a.source).toBe('local')
    expect(newerLocal.toUpload).toHaveLength(1)

    const newerRemote = mergeDrafts([d('a', 3)], [d('a', 5)])
    expect(newerRemote.merged.a.source).toBe('backend')
    expect(newerRemote.toUpload).toEqual([])

    const same = mergeDrafts([d('a', 3)], [d('a', 3)])
    expect(same.merged.a.source).toBe('backend')
    expect(same.toUpload).toEqual([])
  })
})
