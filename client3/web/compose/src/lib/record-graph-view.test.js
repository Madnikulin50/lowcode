import { describe, it, expect } from 'vitest'
import {
  createGraph, addRecord, addEdge, applyPositions, seedNear, forgetPositions, inheritPositions, shouldCluster, clusterNode, clusterID,
} from './record-graph'
import {
  categoryList, neighbours, searchMatches, statusColor, computeLayout, nodeSize, toEchartsOption, SYMBOLS,
} from './record-graph-view'

const ref = (name, moduleID) => ({ name, kind: 'Record', options: { moduleID }, isMulti: false })
const str = name => ({ name, kind: 'String', isMulti: false, options: {} })
const severity = {
  name: 'severity', kind: 'Select', isMulti: false,
  options: { options: [
    { value: 'low', text: 'Низкая', style: { backgroundColor: 'light' } },
    { value: 'high', text: 'Высокая', style: { backgroundColor: 'danger' } },
    { value: 'x', text: 'Свой', style: { backgroundColor: '#112233' } },
    { value: 'none', text: 'Без цвета', style: {} },
  ] },
}

const objects = { moduleID: '1', name: 'Объекты', fields: [str('title')] }
const items = { moduleID: '2', name: 'Позиции', fields: [str('name'), ref('object', '1'), severity] }
const checks = { moduleID: '5', name: 'Проверки', fields: [ref('item', '2')] }
const rec = (moduleID, recordID, values = {}) => ({ moduleID, recordID, values })

// root (object) <- a, b (items) <- c (check)
function sample () {
  const g = createGraph()
  const root = addRecord(g, rec('1', '11', { title: 'Объект' }), objects, { root: true })
  const a = addRecord(g, rec('2', '20', { name: 'A' }), items)
  const b = addRecord(g, rec('2', '21', { name: 'B' }), items)
  const c = addRecord(g, rec('5', '50', {}), checks)
  addEdge(g, a.id, root.id, 'object')
  addEdge(g, b.id, root.id, 'object')
  addEdge(g, c.id, a.id, 'item')
  return { g, root, a, b, c }
}

describe('status colour', () => {
  const theme = { danger: '#dc3545', light: '#f8f9fa' }

  it('takes the colour of the Select value from the module', () => {
    expect(statusColor(rec('2', '1', { severity: 'high' }), items, 'severity', theme)).toBe('#dc3545')
    expect(statusColor(rec('2', '1', { severity: 'low' }), items, 'severity', theme)).toBe('#f8f9fa')
    expect(statusColor(rec('2', '1', { severity: 'x' }), items, 'severity', theme)).toBe('#112233')
  })

  it('has none when the value has no colour, the field is not a Select, or nothing is named', () => {
    expect(statusColor(rec('2', '1', { severity: 'none' }), items, 'severity', theme)).toBe('')
    expect(statusColor(rec('2', '1', { severity: 'gone' }), items, 'severity', theme)).toBe('')
    expect(statusColor(rec('2', '1', { name: 'A' }), items, 'name', theme)).toBe('')
    expect(statusColor(rec('2', '1', {}), items, 'severity', theme)).toBe('')
    expect(statusColor(rec('2', '1', { severity: 'high' }), items, '', theme)).toBe('')
  })

  it('puts the status on the node', () => {
    const g = createGraph()
    const n = addRecord(g, rec('2', '20', { name: 'A', severity: 'high' }), items, { statusFields: { 2: 'severity' }, themeVars: theme })
    expect(n.status).toEqual({ color: '#dc3545', text: 'Высокая' })
    expect(addRecord(g, rec('2', '21', { name: 'B' }), items, { statusFields: { 2: 'severity' }, themeVars: theme }).status).toBeUndefined()
  })
})

describe('legend', () => {
  it('counts the records of each module, in the order they appear', () => {
    const { g } = sample()
    expect(categoryList(g, { 1: 'Объекты', 2: 'Позиции', 5: 'Проверки' }).map(c => c.name)).toEqual(['Объекты (1)', 'Позиции (2)', 'Проверки (1)'])
  })

  it('counts what a cluster holds, and does not count the "more" nodes', () => {
    const g = createGraph()
    const root = addRecord(g, rec('1', '11'), objects, { root: true })
    const cl = clusterNode(root, { moduleID: '2', field: { name: 'object' } }, { set: [1, 2, 3], filter: { nextPage: 'x' } }, 'Позиции')
    g.nodes.set(cl.id, cl)
    g.nodes.set('more:1', { id: 'more:1', more: true, moduleID: '2' })
    expect(categoryList(g, { 1: 'Объекты', 2: 'Позиции' }).map(c => c.name)).toEqual(['Объекты (1)', 'Позиции (3+)'])
  })

  it('switches a module off in the legend', () => {
    const { g } = sample()
    const o = toEchartsOption(g, { moduleNames: { 1: 'Объекты', 2: 'Позиции', 5: 'Проверки' }, hiddenModules: new Set(['2']) })
    expect(o.legend[0].selected).toEqual({ 'Позиции (2)': false })
  })
})

describe('selection and search', () => {
  it('finds the neighbours of a node', () => {
    const { g, a } = sample()
    expect([...neighbours(g, a.id)].sort()).toEqual(['1:11', '2:20', '5:50'])
  })

  it('finds nodes by name, ignoring case, and by module', () => {
    const { g } = sample()
    expect([...searchMatches(g, ' объ ')]).toEqual(['1:11'])
    expect([...searchMatches(g, 'a')].sort()).toEqual(['2:20'])
    expect([...searchMatches(g, 'позиции', { 2: 'Позиции' })].sort()).toEqual(['2:20', '2:21'])
    expect(searchMatches(g, '').size).toBe(0)
  })

  it('dims what is not next to the selected node, and labels its edges', () => {
    const { g, a } = sample()
    const o = toEchartsOption(g, { selectedID: a.id })
    const byID = Object.fromEntries(o.series[0].data.map(n => [n.id, n]))
    expect(byID['2:20'].itemStyle.borderWidth).toBe(5)
    expect(byID['2:21'].itemStyle.opacity).toBeLessThan(0.5) // not next to a
    expect(byID['1:11'].itemStyle.opacity).toBeUndefined() // next to a
    const links = o.series[0].links
    expect(links.find(l => l.source === '2:21').lineStyle.opacity).toBeLessThan(0.2)
    expect(links.find(l => l.source === '5:50').label.show).toBe(true)
    expect(links.find(l => l.source === '2:21').label.show).toBe(false)
  })

  it('shows the search result and dims the rest', () => {
    const { g } = sample()
    const o = toEchartsOption(g, { highlightIDs: new Set(['2:21']) })
    const byID = Object.fromEntries(o.series[0].data.map(n => [n.id, n]))
    expect(byID['2:21'].itemStyle.borderWidth).toBe(4)
    expect(byID['2:20'].itemStyle.opacity).toBeLessThan(0.5)
  })
})

describe('layouts', () => {
  it('puts the root in the middle and the rest on rings by distance', () => {
    const { g } = sample()
    const pos = computeLayout(g, '1:11', 'radial')
    expect(pos.get('1:11')).toEqual([0, 0])
    const dist = id => Math.hypot(...pos.get(id))
    expect(dist('2:20')).toBeGreaterThan(50)
    expect(dist('2:20')).toBeCloseTo(dist('2:21'), 5)
    expect(dist('5:50')).toBeGreaterThan(dist('2:20'))
  })

  it('puts what the root refers to above it and what refers to it below', () => {
    const g = createGraph()
    const root = addRecord(g, rec('2', '20', { name: 'A' }), items, { root: true })
    const parent = addRecord(g, rec('1', '11', { title: 'Объект' }), objects)
    const child = addRecord(g, rec('5', '50'), checks)
    addEdge(g, root.id, parent.id, 'object') // the root refers to the object
    addEdge(g, child.id, root.id, 'item') // the check refers to the root

    const pos = computeLayout(g, root.id, 'layers')
    expect(pos.get(parent.id)[1]).toBeLessThan(pos.get(root.id)[1])
    expect(pos.get(child.id)[1]).toBeGreaterThan(pos.get(root.id)[1])
  })

  it('keeps the nodes of a level apart, and does not lose a node that is not joined', () => {
    const { g } = sample()
    addRecord(g, rec('1', '99', { title: 'Одинокий' }), objects)
    const pos = computeLayout(g, '1:11', 'layers')
    expect(pos.size).toBe(g.nodes.size)
    expect(Math.abs(pos.get('2:20')[0] - pos.get('2:21')[0])).toBeGreaterThanOrEqual(90)
  })

  it('is empty without a root', () => {
    expect(computeLayout(createGraph(), 'x', 'radial').size).toBe(0)
  })

  it('draws with the layout it is given', () => {
    const { g, root } = sample()
    const force = toEchartsOption(g)
    expect(force.series[0].layout).toBe('force')
    const radial = toEchartsOption(g, { layout: 'radial', rootID: root.id })
    expect(radial.series[0].layout).toBe('none')
    const rootNode = radial.series[0].data.find(n => n.id === root.id)
    expect([rootNode.x, rootNode.y]).toEqual([0, 0])
    expect(radial.series[0].data.every(n => n.fixed)).toBe(true)
  })
})

describe('positions', () => {
  it('keeps the nodes that were placed, and lets a new one find its place', () => {
    const { g, root, a } = sample()
    applyPositions(g, new Map([[root.id, [0, 0]], [a.id, [100, 40]]]))
    const fresh = addRecord(g, rec('2', '22', { name: 'C' }), items)
    seedNear(a, fresh, 50, () => 0.5)
    expect(fresh.pos).toEqual([100, 40])
    expect(fresh.settled).toBe(false)

    const data = Object.fromEntries(toEchartsOption(g).series[0].data.map(n => [n.id, n]))
    expect(data[a.id]).toMatchObject({ x: 100, y: 40, fixed: true })
    expect(data[fresh.id]).toMatchObject({ x: 100, y: 40, fixed: false })
    expect(data['2:21'].x).toBeUndefined() // never placed, never seeded
  })

  it('freezes every placed node', () => {
    const { g, a } = sample()
    applyPositions(g, new Map([[a.id, [1, 2]]]))
    const fresh = addRecord(g, rec('2', '22'), items)
    seedNear(a, fresh)
    const data = Object.fromEntries(toEchartsOption(g, { frozen: true }).series[0].data.map(n => [n.id, n]))
    expect(data[fresh.id].fixed).toBe(true)
    expect(toEchartsOption(g, { frozen: true }).series[0].force.layoutAnimation).toBe(false)
  })

  it('forgets where the nodes were, and ignores positions that are not numbers', () => {
    const { g, a } = sample()
    applyPositions(g, new Map([[a.id, [NaN, 2]]]))
    expect(a.pos).toBeUndefined()
    applyPositions(g, new Map([[a.id, [3, 4]]]))
    forgetPositions(g)
    expect(a.pos).toBeUndefined()
    expect(a.settled).toBe(false)
  })
})

describe('a graph built anew', () => {
  it('keeps the positions the old one had', () => {
    const { g: old, root, a } = sample()
    applyPositions(old, new Map([[root.id, [0, 0]], [a.id, [7, 8]]]))
    const { g: fresh } = sample()
    addRecord(fresh, rec('2', '77', { name: 'new' }), items)

    inheritPositions(old, fresh)
    expect(fresh.nodes.get(a.id)).toMatchObject({ pos: [7, 8], settled: true })
    expect(fresh.nodes.get('2:77').pos).toBeUndefined()
    expect(fresh.nodes.get('2:21').pos).toBeUndefined() // was never placed in the old one
    fresh.nodes.get(a.id).pos[0] = 99
    expect(old.nodes.get(a.id).pos[0]).toBe(7) // a copy, not the same array
    expect(() => inheritPositions(undefined, fresh)).not.toThrow()
  })
})

describe('clusters', () => {
  it('stands for a long list from a threshold on, and not when switched off', () => {
    expect(shouldCluster(8, 8)).toBe(true)
    expect(shouldCluster(7, 8)).toBe(false)
    expect(shouldCluster(100, 0)).toBe(false)
  })

  it('knows what it holds', () => {
    const g = createGraph()
    const root = addRecord(g, rec('1', '11'), objects, { root: true })
    const plan = { moduleID: '2', field: { name: 'object' } }
    const cl = clusterNode(root, plan, { set: [rec('2', '1'), rec('2', '2')], filter: { nextPage: 'c' } }, 'Позиции')
    expect(cl).toMatchObject({ id: clusterID(root.id, '2', 'object'), cluster: true, count: 2, countMore: true, label: 'Позиции · 2+', parent: root.id })
    expect(cl.members).toHaveLength(2)
    expect(clusterNode(root, plan, { set: [rec('2', '1')] }, 'П').label).toBe('П · 1')
  })
})

describe('looks', () => {
  it('gives each module a shape of its own, and a bigger node to a better connected one', () => {
    const { g } = sample()
    const o = toEchartsOption(g, {})
    const syms = Object.fromEntries(o.series[0].data.map(n => [n.id, n.symbol]))
    expect(syms['1:11']).toBe(SYMBOLS[0])
    expect(syms['2:20']).toBe(SYMBOLS[1])
    expect(syms['5:50']).toBe(SYMBOLS[2])
    expect(nodeSize({}, 8)).toBeGreaterThan(nodeSize({}, 0))
    expect(nodeSize({}, 99)).toBe(nodeSize({}, 8))
    expect(nodeSize({ root: true }, 0)).toBeGreaterThan(nodeSize({}, 8))
  })

  it('hides overlapping labels, keeps edge names for the selection, and puts a halo on text', () => {
    const { g } = sample()
    const s = toEchartsOption(g, { themeVars: { background: '#101010' } }).series[0]
    expect(s.labelLayout.hideOverlap).toBe(true)
    expect(s.edgeLabel.show).toBe(false)
    expect(s.emphasis.edgeLabel.show).toBe(true)
    expect(s.data[0].label.textBorderColor).toBe('#101010')
  })

  it('rings a node in the colour of its status, over the plain outline', () => {
    const { g, a } = sample()
    a.status = { color: '#dc3545', text: 'Высокая' }
    a.expanded = true
    const node = toEchartsOption(g).series[0].data.find(n => n.id === a.id)
    expect(node.itemStyle).toMatchObject({ borderColor: '#dc3545', borderWidth: 4 })
    expect(node.itemStyle.shadowBlur).toBeGreaterThan(0)
  })

  it('shows the status in the tooltip', () => {
    const { g, a } = sample()
    a.status = { color: '#dc3545', text: 'Высокая' }
    const o = toEchartsOption(g)
    const html = o.tooltip.formatter({ dataType: 'node', data: o.series[0].data.find(n => n.id === a.id) })
    expect(html).toContain('Высокая')
  })
})
