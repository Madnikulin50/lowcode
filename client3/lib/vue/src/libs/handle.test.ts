import { expect } from 'chai'
import { isValid, handleState, classState, classNamePattern } from './handle'

// `handle` validation is shared by every admin form that lets a user pick an
// identifier; a regression here silently accepts handles the server rejects.
describe('libs/handle', () => {
  describe('isValid', () => {
    it('accepts a simple handle', () => {
      expect(isValid('myHandle')).to.equal(true)
    })

    it('accepts digits, dash, underscore and dot inside', () => {
      expect(isValid('a-b_c.d1')).to.equal(true)
    })

    it('rejects a leading digit', () => {
      expect(isValid('1abc')).to.equal(false)
    })

    it('rejects a trailing dash', () => {
      expect(isValid('abc-')).to.equal(false)
    })

    it('rejects whitespace', () => {
      expect(isValid('a b')).to.equal(false)
    })

    it('rejects a single character (needs at least two)', () => {
      expect(isValid('a')).to.equal(false)
    })
  })

  describe('handleState', () => {
    it('returns null for a valid handle', () => {
      expect(handleState('ok')).to.equal(null)
    })

    it('returns false for an invalid handle', () => {
      expect(handleState('-bad')).to.equal(false)
    })

    it('returns undefined for an empty handle', () => {
      expect(handleState('')).to.equal(undefined)
    })
  })

  describe('classNamePattern / classState', () => {
    it('accepts a single class', () => {
      expect(classNamePattern.test('btn-primary')).to.equal(true)
    })

    it('accepts multiple space separated classes', () => {
      expect(classNamePattern.test('btn btn-primary')).to.equal(true)
    })

    it('rejects a class starting with a digit', () => {
      expect(classNamePattern.test('1btn')).to.equal(false)
    })

    it('classState returns null for valid input', () => {
      expect(classState('btn btn-primary')).to.equal(null)
    })

    it('classState returns false for invalid input', () => {
      expect(classState('1btn')).to.equal(false)
    })

    it('classState returns undefined for empty input', () => {
      expect(classState('')).to.equal(undefined)
    })
  })
})
