import { describe, it, expect } from 'vitest'
import { formatValue, renderTemplate, autoLabel, labelFor, details, clip } from './record-label'

const str = (name, extra = {}) => ({ name, kind: 'String', isMulti: false, options: {}, ...extra })
const sel = (name, options) => ({ name, kind: 'Select', isMulti: false, options: { options } })
const num = (name, options = {}) => ({ name, kind: 'Number', isMulti: false, options })
const dt = (name, options = {}) => ({ name, kind: 'DateTime', isMulti: false, options })
const rec = (values, extra = {}) => ({ recordID: '77', values, ...extra })

const voisr = { fields: [str('position_number'), { name: 'object', kind: 'Record', options: { moduleID: '1' } }, str('work_name'), num('volume')] }
const files = { fields: [{ name: 'file', kind: 'File', options: {} }, str('extracted_text'), str('recognized_name'), str('recognized_number'), sel('match_status', [{ value: 'ok', text: 'Совпало' }])] }
const runs = { fields: [dt('run_date'), sel('overall_status', [{ value: 'success', text: 'Успешно' }, { value: 'fail', text: 'Не успешно' }])] }
const discrepancies = { fields: [{ name: 'comparison', kind: 'Record', options: {} }, num('page_number'), sel('discrepancy_type', [{ value: 'missing_page', text: 'Отсутствует лист' }]), sel('severity', [{ value: 'high', text: 'Высокая' }]), str('description')] }

describe('formatValue', () => {
  it('shows a Select by its text, a number with separators, a flag', () => {
    expect(formatValue(sel('s', [{ value: 'a', text: 'Alpha' }]), 'a')).toBe('Alpha')
    expect(formatValue(sel('s', [{ value: 'a', text: 'Alpha' }]), 'zzz')).toBe('zzz')
    expect(formatValue(num('n'), '1234.5', { locale: 'en-US' })).toBe('1,234.5')
    expect(formatValue(num('n', { suffix: ' m²' }), 12, { locale: 'en-US' })).toBe('12 m²')
    expect(formatValue({ kind: 'Bool', options: { trueLabel: 'Да' } }, 'true')).toBe('Да')
    expect(formatValue({ kind: 'Bool', options: {} }, 'false')).toBe('✗')
  })

  it('writes dates the way the field says', () => {
    const iso = '2026-03-12T12:30:00Z'
    expect(formatValue(dt('d', { onlyDate: true }), iso, { locale: 'en-GB' })).toMatch(/12\/03\/2026/)
    expect(formatValue(dt('d'), iso, { locale: 'en-GB' })).toMatch(/2026/)
    expect(formatValue(dt('d'), 'not a date')).toBe('not a date')
  })

  it('joins several values, and leaves out what only an id says', () => {
    expect(formatValue(str('t', { isMulti: true }), ['a', 'b'])).toBe('a, b')
    expect(formatValue(str('t', { isMulti: true }), ['a', 'b', 'c', 'd', 'e'])).toBe('a, b, c +2')
    expect(formatValue({ kind: 'Record', options: {} }, '12')).toBe('')
    expect(formatValue({ kind: 'File', options: {} }, '12')).toBe('')
    expect(formatValue(str('t'), '')).toBe('')
    expect(formatValue(str('t'), undefined)).toBe('')
  })
})

describe('renderTemplate', () => {
  it('fills the fields in', () => {
    expect(renderTemplate('{{position_number}} · {{work_name}}', rec({ position_number: '1.1', work_name: 'Штукатурка' }), voisr)).toBe('1.1 · Штукатурка')
    expect(renderTemplate('№ {{position_number}}, {{volume}} м²', rec({ position_number: '1.1', volume: '12.5' }), voisr, { locale: 'en-US' })).toBe('№ 1.1, 12.5 м²')
  })

  it('takes the separator with a missing value', () => {
    const t = '{{position_number}} · {{work_name}}'
    expect(renderTemplate(t, rec({ work_name: 'Штукатурка' }), voisr)).toBe('Штукатурка')
    expect(renderTemplate(t, rec({ position_number: '1.1' }), voisr)).toBe('1.1')
    expect(renderTemplate('{{a}} - {{position_number}} - {{work_name}}', rec({ work_name: 'X' }), voisr)).toBe('X')
    expect(renderTemplate('{{work_name}} · {{position_number}} · {{volume}}', rec({ work_name: 'X', volume: '3' }), voisr)).toBe('X · 3')
  })

  it('gives nothing when nothing is filled, or there is no template', () => {
    expect(renderTemplate('{{position_number}} · {{work_name}}', rec({}), voisr)).toBe('')
    expect(renderTemplate('', rec({ work_name: 'X' }), voisr)).toBe('')
    expect(renderTemplate(undefined, rec({ work_name: 'X' }), voisr)).toBe('')
  })

  it('reads system fields and Selects', () => {
    expect(renderTemplate('{{overall_status}} · {{createdAt}}', rec({ overall_status: 'fail' }, { createdAt: '2026-03-12T12:30:00Z' }), runs, { locale: 'en-GB' })).toMatch(/^Не успешно · .*2026/)
    expect(renderTemplate('#{{recordID}}', rec({}), runs)).toBe('#77')
  })
})

describe('autoLabel', () => {
  it('puts a number before the name', () => {
    expect(autoLabel(rec({ position_number: '1.1', work_name: 'Штукатурка' }), voisr)).toBe('1.1 · Штукатурка')
    expect(autoLabel(rec({ work_name: 'Штукатурка' }), voisr)).toBe('Штукатурка')
    expect(autoLabel(rec({ position_number: '1.1' }), voisr)).toBe('1.1')
  })

  it('does not take a long text for a name', () => {
    expect(autoLabel(rec({ extracted_text: 'OCR '.repeat(100) }), files)).toBe('#77')
    expect(autoLabel(rec({ extracted_text: 'OCR '.repeat(100), recognized_name: 'Акт скрытых работ', recognized_number: '12' }), files)).toBe('12 · Акт скрытых работ')
  })

  it('falls back on Select and date values, then the id', () => {
    expect(autoLabel(rec({ overall_status: 'success', run_date: '2026-03-12T12:30:00Z' }), runs, { locale: 'en-GB' })).toMatch(/^Успешно · .*2026/)
    expect(autoLabel(rec({ discrepancy_type: 'missing_page', severity: 'high', description: 'long '.repeat(40) }), discrepancies)).toBe('Отсутствует лист · Высокая')
    expect(autoLabel(rec({}), runs)).toBe('#77')
  })
})

describe('labelFor', () => {
  it('prefers the template, then the label field of the reference, then the guess', () => {
    const r = rec({ position_number: '1.1', work_name: 'Штукатурка' })
    expect(labelFor(r, voisr, { template: '{{work_name}} ({{position_number}})' }).label).toBe('Штукатурка (1.1)')
    expect(labelFor(r, voisr, { preferredField: 'work_name' }).label).toBe('Штукатурка')
    expect(labelFor(r, voisr, {}).label).toBe('1.1 · Штукатурка')
    // a template that fills nothing does not leave the node nameless
    expect(labelFor(r, voisr, { template: '{{nothing}}' }).label).toBe('1.1 · Штукатурка')
  })

  it('shortens the label but keeps the full text', () => {
    const r = rec({ work_name: 'x'.repeat(100) })
    const { label, full } = labelFor(r, voisr, {})
    expect(label.length).toBe(48)
    expect(full.length).toBe(100)
    expect(clip('a   b\n c')).toBe('a b c')
  })
})

describe('details', () => {
  it('lists the filled fields a person can read, not repeating the name', () => {
    const r = rec({ position_number: '1.1', work_name: 'Штукатурка', volume: '12', object: '5' })
    const d = details(r, voisr, { name: '1.1 · Штукатурка', locale: 'en-US' })
    expect(d).toEqual([{ label: 'volume', value: '12' }])
  })

  it('stops at the maximum, and shortens long values', () => {
    const r = rec({ recognized_name: 'x'.repeat(200), recognized_number: '1', extracted_text: 'y', match_status: 'ok' })
    const d = details(r, files, { max: 2 })
    expect(d).toHaveLength(2)
    expect(d[0].value.length).toBeLessThanOrEqual(80)
  })
})
