// run with: node --test src/lib/line-diff.test.mjs
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { lineDiff, hasChanges } from './line-diff.js'

test('identical texts have no changes', () => {
  assert.equal(hasChanges(lineDiff('a\nb', 'a\nb')), false)
})

test('shows added and removed lines in reading order', () => {
  assert.deepEqual(lineDiff('a\nb\nc', 'a\nx\nc\nd'), [
    { type: 'same', text: 'a' },
    { type: 'del', text: 'b' },
    { type: 'add', text: 'x' },
    { type: 'same', text: 'c' },
    { type: 'add', text: 'd' },
  ])
})

test('empty sides', () => {
  assert.deepEqual(lineDiff('', 'a'), [{ type: 'add', text: 'a' }])
  assert.deepEqual(lineDiff('a', ''), [{ type: 'del', text: 'a' }])
  assert.deepEqual(lineDiff('', ''), [])
})
