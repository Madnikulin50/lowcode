import { findReferencingFields, childrenQuery } from './related-records'
import { labelFor, details, valueText } from './record-label'
import { statusColor } from './record-graph-view'

// The graph of the RecordGraph page block: records of any modules as nodes, a
// Record field that refers from one to another as an edge (child -> parent).
// Everything here is plain data in and out, so it can be tested without a page.

export const nodeID = (moduleID, recordID) => `${moduleID}:${recordID}`
export const edgeID = (source, target, field) => `${source}>${target}#${field}`
export const moreID = (nodeKey, moduleID, field) => `more:${nodeKey}:${moduleID}:${field}`

const hasID = id => id !== undefined && id !== null && id !== '' && id !== '0' && id !== 0

export function createGraph () {
  return { nodes: new Map(), edges: new Map(), dropped: 0 }
}

// The ids a Record field of a record holds: one id, a list of them (a
// multi-value field), or none.
export function recordRefs (record, fieldName) {
  const raw = record && record.values ? record.values[fieldName] : undefined
  return (Array.isArray(raw) ? raw : [raw]).map(v => (v === undefined || v === null ? '' : String(v))).filter(hasID)
}

// A record as the graph reads it. The API lists a record's values as an array
// of { name, value, ref } (compose.Record turns that into an object); a record
// that is already an object is taken as it is. A multi-value field becomes an
// array, a single one its value.
export function normalizeRecord (raw, module) {
  if (!raw) return raw

  let values = raw.values
  if (Array.isArray(values)) {
    const multi = new Set(((module && module.fields) || []).filter(f => f.isMulti).map(f => f.name))
    const map = {}
    for (const v of values) {
      if (!v || !v.name || v.deletedAt) continue
      if (multi.has(v.name)) map[v.name] = [...(map[v.name] || []), v.value]
      else map[v.name] = v.value
    }
    values = map
  }

  return {
    recordID: raw.recordID,
    moduleID: raw.moduleID,
    values: values || {},
    createdAt: raw.createdAt,
    updatedAt: raw.updatedAt,
  }
}

// What a node is called: the module's template, else the label field of the
// Record field it was reached by, else a guess (see record-label.js).
export function recordLabel (record, module, preferredField = '', template = '') {
  return labelFor(record, module, { template, preferredField }).label
}

// Adds the records as nodes, up to maxNodes. The root is always kept. Returns
// the node of each record (or undefined for one that did not fit).
export function addRecord (graph, raw, module, { maxNodes = 150, preferredField = '', root = false, templates = {}, statusFields = {}, themeVars = {}, locale } = {}) {
  const record = normalizeRecord(raw, module)
  if (!record || !hasID(record.recordID)) return undefined

  const id = nodeID(record.moduleID, record.recordID)
  const known = graph.nodes.get(id)
  if (known) return known

  if (!root && graph.nodes.size >= maxNodes) {
    graph.dropped++
    return undefined
  }

  const moduleID = String(record.moduleID)
  const { label, full } = labelFor(record, module, { template: templates[moduleID], preferredField, locale })

  const node = {
    id,
    moduleID,
    recordID: String(record.recordID),
    label,
    full,
    details: details(record, module, { name: full, locale }),
    status: statusOf(record, module, statusFields[moduleID], themeVars, locale),
    record,
    root,
    expanded: false,
  }
  graph.nodes.set(id, node)
  return node
}

// The colour (and the words) of a record's status, if its module names a status field
function statusOf (record, module, fieldName, themeVars, locale) {
  const color = statusColor(record, module, fieldName, themeVars)
  if (!color) return undefined
  return { color, text: valueText(record, module, fieldName, { locale }) }
}

export function addEdge (graph, source, target, field, label = '') {
  if (!source || !target || !graph.nodes.has(source) || !graph.nodes.has(target)) return undefined

  const id = edgeID(source, target, field)
  if (!graph.edges.has(id)) graph.edges.set(id, { id, source, target, field, label: label || field })
  return graph.edges.get(id)
}

// The parents of a record: for each Record field of its module the ids it
// holds, grouped by the module they point into.
export function parentRefs (record, module, { exclude = [] } = {}) {
  const out = []
  for (const field of (module && module.fields) || []) {
    const target = field.kind === 'Record' && field.options ? String(field.options.moduleID || '') : ''
    if (!hasID(target) || exclude.includes(target)) continue

    for (const recordID of recordRefs(record, field.name)) {
      out.push({ field, moduleID: target, recordID })
    }
  }
  return out
}

// The lists to read to find the children of a record: one per Record field,
// in any module, that refers to the record's module.
export function childQueries (modules, moduleID, recordID, { exclude = [], limit = 20 } = {}) {
  return findReferencingFields(modules, moduleID)
    .filter(({ module }) => !exclude.includes(String(module.moduleID)))
    .map(({ module, field }) => ({
      module,
      field,
      moduleID: String(module.moduleID),
      query: childrenQuery(field, recordID),
      limit,
    }))
}

// The modules that can take part in a graph around a module: the ones that
// refer to it, the ones it refers to, and itself when it does - for the
// configurator's exclusion list.
export function relatedModules (modules, moduleID) {
  const byID = new Map(modules.map(m => [String(m.moduleID), m]))
  const out = new Map()

  for (const { module } of findReferencingFields(modules, moduleID)) out.set(String(module.moduleID), module)

  const self = byID.get(String(moduleID))
  for (const field of (self && self.fields) || []) {
    const target = field.kind === 'Record' && field.options ? String(field.options.moduleID || '') : ''
    if (byID.has(target)) out.set(target, byID.get(target))
  }

  return [...out.values()].sort((a, b) => (a.name || a.handle || '').localeCompare(b.name || b.handle || ''))
}

// Every module a graph around a module can reach within `hops` steps, along
// Record fields in either direction (the module itself first) - the modules the
// configurator offers a name template or an exclusion for.
export function reachableModules (modules, moduleID, hops = 2) {
  const byID = new Map(modules.map(m => [String(m.moduleID), m]))
  const start = String(moduleID)
  if (!byID.has(start)) return []

  const seen = new Set([start])
  let frontier = [start]
  for (let h = 0; h < hops && frontier.length; h++) {
    const next = []
    for (const id of frontier) {
      for (const m of relatedModules(modules, id)) {
        const mid = String(m.moduleID)
        if (!seen.has(mid)) { seen.add(mid); next.push(mid) }
      }
    }
    frontier = next
  }

  return [...seen].map(id => byID.get(id))
}

export { toEchartsOption, palette, categoryList, neighbours, searchMatches, statusColor, computeLayout, nodeSize, LAYOUTS, SYMBOLS } from './record-graph-view'

// ---- where the nodes are

// Remembers the positions the layout found (a Map of nodeID -> [x, y]), so that
// opening a node does not move the ones already there.
export function applyPositions (graph, positions) {
  for (const n of graph.nodes.values()) {
    const p = positions.get(n.id)
    if (p && Number.isFinite(p[0]) && Number.isFinite(p[1])) {
      n.pos = [p[0], p[1]]
      n.settled = true
    }
  }
}

// A new node starts next to the node it came from; it is not settled, so the
// layout moves it into place while the others stay.
export function seedNear (parent, child, spread = 50, random = Math.random) {
  if (!parent || !parent.pos || !child || child.pos) return
  child.pos = [parent.pos[0] + (random() - 0.5) * spread, parent.pos[1] + (random() - 0.5) * spread]
  child.settled = false
}

// A graph built anew (after a refresh) keeps the positions of the nodes the old
// one had.
export function inheritPositions (from, to) {
  for (const n of to.nodes.values()) {
    const old = from && from.nodes.get(n.id)
    if (old && old.pos) {
      n.pos = [old.pos[0], old.pos[1]]
      n.settled = !!old.settled
    }
  }
}

// Forget where the nodes are, for a layout from scratch
export function forgetPositions (graph) {
  for (const n of graph.nodes.values()) {
    delete n.pos
    n.settled = false
  }
}

// ---- clusters

// A node stands for the records of one list instead of drawing each of them,
// when there are many (a hundred positions of one estimate). Opening it shows them.
export const clusterID = (nodeKey, moduleID, field) => `cluster:${nodeKey}:${moduleID}:${field}`

export function shouldCluster (count, clusterFrom) {
  return clusterFrom > 0 && count >= clusterFrom
}

export function clusterNode (parent, plan, result, moduleName) {
  const { set = [], filter = {} } = result
  const id = clusterID(parent.id, plan.moduleID, plan.field.name)
  const countMore = !!filter.nextPage

  return {
    id,
    cluster: true,
    moduleID: plan.moduleID,
    recordID: '',
    root: false,
    expanded: false,
    count: set.length,
    countMore,
    label: `${moduleName} · ${set.length}${countMore ? '+' : ''}`,
    full: `${moduleName} · ${set.length}${countMore ? '+' : ''}`,
    details: [],
    plan,
    members: set,
    filter,
    parent: parent.id,
  }
}
