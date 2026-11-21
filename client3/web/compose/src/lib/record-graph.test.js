import { describe, it, expect } from 'vitest'
import {
  createGraph, nodeID, normalizeRecord, reachableModules, recordRefs, recordLabel, addRecord, addEdge, parentRefs, childQueries, relatedModules, toEchartsOption, palette,
} from './record-graph'

const ref = (name, moduleID, extra = {}) => ({ name, kind: 'Record', options: { moduleID }, isMulti: false, ...extra })
const str = name => ({ name, kind: 'String', isMulti: false, options: {} })

const objects = { moduleID: '1', name: 'Объекты', fields: [str('title')] }
const voisr = { moduleID: '2', name: 'Позиции ВОИСР', fields: [str('name'), ref('object', '1'), ref('smeta_ref', '3')] }
const smeta = { moduleID: '3', name: 'Позиции смет', fields: [str('code'), ref('object', '1')] }
const files = { moduleID: '4', name: 'Файлы ИД', fields: [ref('object', '1'), ref('related', '1', { isMulti: true })] }
const modules = [objects, voisr, smeta, files]
const rec = (moduleID, recordID, values = {}) => ({ moduleID, recordID, values })

describe('record references', () => {
  it('reads one id, a list of ids, and nothing', () => {
    expect(recordRefs({ values: { object: '11' } }, 'object')).toEqual(['11'])
    expect(recordRefs({ values: { related: ['1', '2', '0', ''] } }, 'related')).toEqual(['1', '2'])
    expect(recordRefs({ values: {} }, 'object')).toEqual([])
    expect(recordRefs({ values: { object: '0' } }, 'object')).toEqual([])
    expect(recordRefs(undefined, 'object')).toEqual([])
  })

  it('finds the parents of a record by its Record fields', () => {
    const r = { recordID: '20', moduleID: '2', values: { name: 'A', object: '11', smeta_ref: '31' } }
    expect(parentRefs(r, voisr).map(p => `${p.field.name}->${p.moduleID}:${p.recordID}`)).toEqual(['object->1:11', 'smeta_ref->3:31'])
    expect(parentRefs(r, voisr, { exclude: ['3'] }).map(p => p.field.name)).toEqual(['object'])
  })

  it('finds the lists that hold the children of a record', () => {
    const q = childQueries(modules, '1', '11', { limit: 5 })
    expect(q.map(x => `${x.module.name}/${x.field.name}`)).toEqual(['Позиции ВОИСР/object', 'Позиции смет/object', 'Файлы ИД/object', 'Файлы ИД/related'])
    expect(q[0]).toMatchObject({ moduleID: '2', limit: 5, query: "(object = '11')" })
    expect(q[3].query).toBe("('11' IN related)")
    expect(childQueries(modules, '1', '11', { exclude: ['4'] }).map(x => x.moduleID)).toEqual(['2', '3'])
  })

  it('lists the modules around a module', () => {
    expect(relatedModules(modules, '1').map(m => m.moduleID)).toEqual(['4', '3', '2'].sort((a, b) => modules.find(m => m.moduleID === a).name.localeCompare(modules.find(m => m.moduleID === b).name)))
    expect(relatedModules(modules, '2').map(m => m.moduleID).sort()).toEqual(['1', '3'])
  })
})

describe('labels', () => {
  it('prefers the label field, then the first filled text field, then the id', () => {
    const r = { recordID: '20', values: { name: 'Стена', code: 'К-1' } }
    expect(recordLabel(r, voisr, 'code')).toBe('К-1')
    expect(recordLabel(r, voisr)).toBe('Стена')
    expect(recordLabel({ recordID: '21', values: {} }, voisr)).toBe('#21')
    expect(recordLabel({ recordID: '22', values: { name: 'x'.repeat(80) } }, voisr).length).toBe(48)
  })
})

describe('the graph', () => {
  it('does not make a node twice, and keeps the root when full', () => {
    const g = createGraph()
    const root = addRecord(g, rec('1', '11', { title: 'Объект' }), objects, { root: true, maxNodes: 2 })
    expect(addRecord(g, rec('1', '11'), objects, { maxNodes: 2 })).toBe(root)
    expect(addRecord(g, rec('2', '20', { name: 'A' }), voisr, { maxNodes: 2 }).id).toBe(nodeID('2', '20'))
    expect(addRecord(g, rec('2', '21'), voisr, { maxNodes: 2 })).toBeUndefined()
    expect(g.nodes.size).toBe(2)
    expect(g.dropped).toBe(1)
  })

  it('joins nodes with edges once, and only when both exist', () => {
    const g = createGraph()
    const a = addRecord(g, rec('2', '20'), voisr, { root: true })
    const b = addRecord(g, rec('1', '11'), objects)
    expect(addEdge(g, a.id, b.id, 'object', 'Объект')).toMatchObject({ source: '2:20', target: '1:11', label: 'Объект' })
    addEdge(g, a.id, b.id, 'object')
    expect(g.edges.size).toBe(1)
    expect(addEdge(g, a.id, '9:9', 'x')).toBeUndefined()
  })

  it('does not refuse a ring of records', () => {
    const g = createGraph()
    const a = addRecord(g, rec('1', '1'), objects, { root: true })
    const b = addRecord(g, rec('1', '2'), objects)
    addEdge(g, a.id, b.id, 'parent')
    addEdge(g, b.id, a.id, 'parent')
    expect(g.edges.size).toBe(2)
  })
})

describe('echarts option', () => {
  it('draws categories per module, the root big, expanded nodes outlined', () => {
    const g = createGraph()
    const a = addRecord(g, rec('1', '11', { title: 'Объект' }), objects, { root: true })
    const b = addRecord(g, rec('2', '20', { name: 'Стена' }), voisr)
    addEdge(g, b.id, a.id, 'object', 'Объект')
    b.expanded = true

    const o = toEchartsOption(g, { themeVars: { primary: '#123456' }, moduleNames: { 1: 'Объекты', 2: 'Позиции' }, hints: { open: 'open', expand: 'expand' } })
    const s = o.series[0]
    expect(s.type).toBe('graph')
    expect(s.layout).toBe('force')
    expect(s.categories.map(c => c.name)).toEqual(['Объекты (1)', 'Позиции (1)'])
    expect(s.categories[0].itemStyle.color).toBe('#123456')
    expect(s.data.find(n => n.id === '1:11').symbolSize).toBeGreaterThan(s.data.find(n => n.id === '2:20').symbolSize)
    expect(s.data.find(n => n.id === '2:20').itemStyle.borderWidth).toBe(3)
    expect(s.data.find(n => n.id === '2:20').hint).toBe('open')
    expect(s.links).toHaveLength(1)
    expect(s.links[0]).toMatchObject({ source: '2:20', target: '1:11' })
    expect(o.legend[0].data).toEqual(['Объекты (1)', 'Позиции (1)'])
  })

  it('marks the selected node', () => {
    const g = createGraph()
    addRecord(g, rec('1', '11', { title: 'Объект' }), objects, { root: true })
    addRecord(g, rec('2', '20', { name: 'Стена' }), voisr)
    const data = toEchartsOption(g, { selectedID: '2:20', themeVars: { primary: '#123456' } }).series[0].data
    expect(data.find(n => n.id === '2:20').itemStyle.borderColor).toBe('#123456')
    expect(data.find(n => n.id === '2:20').itemStyle.borderWidth).toBeGreaterThan(data.find(n => n.id === '1:11').itemStyle.borderWidth)
    expect(toEchartsOption(g).series[0].data.find(n => n.id === '2:20').itemStyle).toBeUndefined()
  })

  it('turns the legend off, and escapes the tooltip', () => {
    const g = createGraph()
    addRecord(g, rec('1', '11', { title: '<b>x</b>' }), objects, { root: true })
    const o = toEchartsOption(g, { showLegend: false })
    expect(o.legend).toBeUndefined()
    const html = o.tooltip.formatter({ dataType: 'node', data: o.series[0].data[0] })
    expect(html).not.toContain('<b>x</b>')
    expect(html).toContain('&lt;b&gt;x&lt;/b&gt;')
  })

  it('has colours without a theme', () => {
    expect(palette({}).length).toBeGreaterThanOrEqual(10)
    expect(palette({ primary: '#111' })[0]).toBe('#111')
  })
})

describe('reachableModules', () => {
  const checks = { moduleID: '5', name: 'Проверки', fields: [ref('voisr_item', '2')] } // refers to voisr, two steps from objects

  it('walks the references both ways, nearest first', () => {
    const all = [...modules, checks]
    expect(reachableModules(all, '1', 1).map(m => m.moduleID).sort()).toEqual(['1', '2', '3', '4'])
    expect(reachableModules(all, '1', 2).map(m => m.moduleID).sort()).toEqual(['1', '2', '3', '4', '5'])
    expect(reachableModules(all, '1', 0).map(m => m.moduleID)).toEqual(['1'])
    expect(reachableModules(all, '1')[0].moduleID).toBe('1')
  })

  it('has nothing for an unknown module', () => {
    expect(reachableModules(modules, '99')).toEqual([])
  })
})

// the API lists values as an array of { name, value, ref }, not as an object
describe('records as the API lists them', () => {
  const raw = (moduleID, recordID, values) => ({ moduleID, recordID, values: values.map(([name, value]) => ({ name, value })) })

  it('turns the array of values into an object, multi-value fields into lists', () => {
    const r = normalizeRecord(raw('4', '40', [['object', '11'], ['related', '1'], ['related', '2']]), files)
    expect(r.values).toEqual({ object: '11', related: ['1', '2'] })
    expect(normalizeRecord({ recordID: '1', moduleID: '1', values: { title: 'x' } }, objects).values).toEqual({ title: 'x' })
    expect(normalizeRecord({ recordID: '1', moduleID: '1' }, objects).values).toEqual({})
    expect(normalizeRecord(raw('1', '1', [['title', 'gone']]).values.map(v => v) && { recordID: '1', moduleID: '1', values: [{ name: 'title', value: 'x', deletedAt: '2026-01-01' }] }, objects).values).toEqual({})
  })

  it('names a node, and finds the parents, of a record in that form', () => {
    const g = createGraph()
    const node = addRecord(g, raw('2', '20', [['name', 'Штукатурка'], ['object', '11'], ['smeta_ref', '31']]), voisr)
    expect(node.label).toBe('Штукатурка')
    expect(parentRefs(node.record, voisr).map(p => `${p.field.name}->${p.moduleID}:${p.recordID}`)).toEqual(['object->1:11', 'smeta_ref->3:31'])
  })

  it('uses the template with such a record', () => {
    const g = createGraph()
    const node = addRecord(g, raw('2', '20', [['name', 'Штукатурка'], ['object', '11']]), voisr, { templates: { 2: 'Позиция «{{name}}»' } })
    expect(node.label).toBe('Позиция «Штукатурка»')
    expect(node.full).toBe('Позиция «Штукатурка»')
  })
})
