import { describe, it, expect } from 'vitest'
import { findReferencingFields, relationKey, hasRelation, childrenQuery, usableRelations, relationListBlock, interpretCount, countLabel, COUNT_PAGE_SIZE } from './related-records'

const field = (name, moduleID, extra = {}) => ({ name, kind: 'Record', options: { moduleID }, isMulti: false, ...extra })
const modules = [
  { moduleID: '1', name: 'Объекты', fields: [field('parent', '1'), { name: 'title', kind: 'String', options: {} }] },
  { moduleID: '2', name: 'Позиции ВОИСР', fields: [field('object', '1')] },
  { moduleID: '3', name: 'Файлы ИД', fields: [field('object', '1'), field('related_object', '1', { isMulti: true }), field('other', '9')] },
  { moduleID: '4', name: 'Внешний', config: { type: 'datasource' }, fields: [field('object', '1')] },
  { moduleID: '5', name: 'Без ссылок', fields: [{ name: 'object', kind: 'String', options: { moduleID: '1' } }] },
]

describe('findReferencingFields', () => {
  it('finds the Record fields that refer to the module, in any module', () => {
    const got = findReferencingFields(modules, '1').map(c => `${c.module.name}/${c.field.name}`)
    expect(got).toEqual(['Объекты/parent', 'Позиции ВОИСР/object', 'Файлы ИД/object', 'Файлы ИД/related_object'])
  })

  it('includes the module itself, for trees of records', () => {
    expect(findReferencingFields(modules, '1').some(c => c.module.moduleID === '1')).toBe(true)
  })

  it('compares ids as text, and ignores fields of other kinds, other targets and outside modules', () => {
    expect(findReferencingFields(modules, 1).length).toBe(4)
    expect(findReferencingFields(modules, '9').map(c => c.field.name)).toEqual(['other'])
    expect(findReferencingFields(modules, '')).toEqual([])
    expect(findReferencingFields(undefined, '1')).toEqual([])
  })
})

describe('relations', () => {
  it('keys a relation by module and field', () => {
    expect(relationKey({ moduleID: '3', refField: 'object' })).toBe('3:object')
    expect(hasRelation([{ moduleID: '3', refField: 'object' }], '3', 'object')).toBe(true)
    expect(hasRelation([{ moduleID: '3', refField: 'object' }], '3', 'related_object')).toBe(false)
  })

  it('builds the query for a single and a multi value field', () => {
    expect(childrenQuery({ name: 'object', isMulti: false }, '77')).toBe("(object = '77')")
    expect(childrenQuery({ name: 'rel', isMulti: true }, '77')).toBe("('77' IN rel)")
  })

  it('leaves out relations whose module or field is gone', () => {
    const find = id => modules.find(m => m.moduleID === id)
    const got = usableRelations([
      { moduleID: '2', refField: 'object' },
      { moduleID: '2', refField: 'deleted_field' },
      { moduleID: '5', refField: 'object' }, // no longer a Record field
      { moduleID: '404', refField: 'object' },
    ], find)
    expect(got.map(g => g.relation.moduleID)).toEqual(['2'])
    expect(got[0].field.name).toBe('object')
  })

  it('renders a relation as a record list that is filtered by the page record', () => {
    const block = relationListBlock({ moduleID: '2', refField: 'object', fields: [{ name: 'x' }], perPage: 25, presort: 'createdAt DESC', hideAddButton: true })
    expect(block.kind).toBe('RecordList')
    expect(block.options).toMatchObject({ moduleID: '2', refField: 'object', perPage: 25, hideAddButton: true, hideHeader: true })
    expect(block.options.fields).toEqual([{ name: 'x' }])
  })
})

describe('counting children', () => {
  const rows = n => Array.from({ length: n }, (_, i) => ({ recordID: String(i) }))

  it('takes the total when the server gives one', () => {
    expect(interpretCount(rows(3), { total: 3 })).toEqual({ count: 3, more: false })
    expect(interpretCount([], { total: 0 })).toEqual({ count: 0, more: false })
  })

  // the server answers -1 for a query on a Record field
  it('counts the page when the total is unknown (-1)', () => {
    expect(interpretCount(rows(1), { total: -1 })).toEqual({ count: 1, more: false })
    expect(interpretCount(rows(37), { total: -1 })).toEqual({ count: 37, more: false })
    expect(interpretCount([], {})).toEqual({ count: 0, more: false })
  })

  it('says "more" when the page is full or has a next page', () => {
    expect(interpretCount(rows(COUNT_PAGE_SIZE), { total: -1 })).toEqual({ count: COUNT_PAGE_SIZE, more: true })
    expect(interpretCount(rows(10), { total: -1, nextPage: 'abc' })).toEqual({ count: 10, more: true })
  })

  it('writes the count for the badge', () => {
    expect(countLabel({ count: 5, more: false })).toBe('5')
    expect(countLabel({ count: 100, more: true })).toBe('100+')
    expect(countLabel({ count: 0, more: false })).toBe('0')
  })
})
