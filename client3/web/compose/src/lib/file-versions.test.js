import { describe, it, expect } from 'vitest'
import { versionLabel, replaceValue, canMakeCurrent } from './file-versions'

describe('file-versions', () => {
  it('labels only versioned files', () => {
    expect(versionLabel({ version: 3 })).toBe('v3')
    expect(versionLabel({ version: 0 })).toBe('')
    expect(versionLabel(undefined)).toBe('')
  })

  it('replaces a single value', () => {
    expect(replaceValue('10', '10', '11')).toBe('11')
    expect(replaceValue('10', '99', '11')).toBe('10')
  })

  it('replaces in place in a multi value, others untouched', () => {
    expect(replaceValue(['1', '2', '3'], '2', '9')).toEqual(['1', '9', '3'])
    expect(replaceValue(['1', '2'], '5', '9')).toEqual(['1', '2'])
  })

  it('compares ids as strings', () => {
    expect(canMakeCurrent(['5', '6'], 5)).toBe(false)
    expect(canMakeCurrent(['5'], '7')).toBe(true)
    expect(canMakeCurrent(undefined, '7')).toBe(true)
  })
})
