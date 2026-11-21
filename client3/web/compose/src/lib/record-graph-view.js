// How the graph of the RecordGraph block looks: layouts, shapes, colours,
// legend, search, and the echarts option that draws it. Plain data in and out,
// so it can be tested without a page. The graph itself is built by record-graph.js.

export const LAYOUTS = ['force', 'radial', 'layers']

// a shape for each module besides its colour, so modules can be told apart by
// people who do not see the colours
export const SYMBOLS = ['circle', 'roundRect', 'diamond', 'triangle']

const fallbackPalette = ['#0d6efd', '#198754', '#fd7e14', '#6f42c1', '#d63384', '#20c997', '#ffc107', '#6c757d', '#0dcaf0', '#dc3545']

export function palette (themeVars = {}) {
  const fromTheme = ['primary', 'success', 'warning', 'danger', 'secondary', 'dark'].map(k => themeVars[k]).filter(Boolean)
  return [...fromTheme, ...fallbackPalette.filter(c => !fromTheme.includes(c))]
}

const escape = s => String(s).replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]))

// ---- legend

// The modules on the graph with the number of records of each, in the order they
// first appear. A cluster stands for all the records it holds.
export function categoryList (graph, moduleNames = {}) {
  const found = new Map()
  for (const n of graph.nodes.values()) {
    const entry = found.get(n.moduleID) || { moduleID: n.moduleID, count: 0, plus: false }
    if (n.cluster) {
      entry.count += n.count
      entry.plus = entry.plus || !!n.countMore
    } else if (!n.more) {
      entry.count++
    }
    found.set(n.moduleID, entry)
  }

  return [...found.values()].map(c => {
    const base = moduleNames[c.moduleID] || c.moduleID
    return { ...c, name: c.count ? `${base} (${c.count}${c.plus ? '+' : ''})` : base }
  })
}

// ---- selection and search

// A node and the nodes next to it
export function neighbours (graph, id) {
  const out = new Set([id])
  for (const e of graph.edges.values()) {
    if (e.source === id) out.add(e.target)
    if (e.target === id) out.add(e.source)
  }
  return out
}

// The nodes whose name (or module) contains the text, ignoring case
export function searchMatches (graph, query, moduleNames = {}) {
  const q = String(query || '').trim().toLowerCase()
  const out = new Set()
  if (!q) return out

  for (const n of graph.nodes.values()) {
    const hay = `${n.full || ''} ${n.label || ''} ${moduleNames[n.moduleID] || ''}`.toLowerCase()
    if (hay.includes(q)) out.add(n.id)
  }
  return out
}

// ---- status colour

// The colour of a Select value as the module defines it ("danger", "#ff0000"),
// as a colour a canvas can draw; '' when the value has none.
export function statusColor (record, module, fieldName, themeVars = {}) {
  if (!fieldName) return ''

  const field = ((module && module.fields) || []).find(f => f.name === fieldName && f.kind === 'Select')
  const raw = record && record.values ? record.values[fieldName] : undefined
  const value = Array.isArray(raw) ? raw[0] : raw
  if (!field || value === undefined || value === null || value === '') return ''

  const hit = ((field.options && field.options.options) || []).find(o => String(o.value) === String(value))
  const style = hit && hit.style ? hit.style : {}
  const color = style.backgroundColor || style.textColor || ''

  if (themeVars[color]) return themeVars[color]
  return /^(#|rgb|hsl)/i.test(color) ? color : ''
}

// ---- layouts

// Where the nodes go in a layout that is not the force one. Returns
// Map(nodeID -> [x, y]).
//
// radial: the record in the middle, rings by distance, the modules kept together
// layers: the records the root refers to above it, the ones that refer to it below
export function computeLayout (graph, rootID, kind) {
  const out = new Map()
  if (!graph.nodes.has(rootID)) return out

  const order = new Map()
  for (const n of graph.nodes.values()) if (!order.has(n.moduleID)) order.set(n.moduleID, order.size)
  const byModule = (a, b) => (order.get(a.moduleID) - order.get(b.moduleID)) || String(a.label).localeCompare(String(b.label))

  // breadth first from the root, over the edges in both directions; in a layer
  // layout the level is signed: up along an edge from a record to what it refers to
  const level = new Map([[rootID, 0]])
  const depth = new Map([[rootID, 0]])
  const queue = [rootID]
  while (queue.length) {
    const cur = queue.shift()
    for (const e of graph.edges.values()) {
      let next
      let step
      if (e.source === cur) { next = e.target; step = -1 } else if (e.target === cur) { next = e.source; step = 1 } else continue
      if (depth.has(next)) continue
      depth.set(next, depth.get(cur) + 1)
      level.set(next, level.get(cur) + step)
      queue.push(next)
    }
  }

  // anything not joined to the root goes one step beyond the last
  const farthest = Math.max(...depth.values())
  for (const n of graph.nodes.values()) {
    if (!depth.has(n.id)) { depth.set(n.id, farthest + 1); level.set(n.id, farthest + 1) }
  }

  const groups = new Map()
  const place = kind === 'layers' ? level : depth
  for (const n of graph.nodes.values()) {
    const key = place.get(n.id)
    if (!groups.has(key)) groups.set(key, [])
    groups.get(key).push(n)
  }

  for (const [key, nodes] of groups) {
    nodes.sort(byModule)

    if (kind === 'layers') {
      const gap = 96
      nodes.forEach((n, i) => out.set(n.id, [(i - (nodes.length - 1) / 2) * gap, key * 150]))
    } else if (key === 0) {
      nodes.forEach(n => out.set(n.id, [0, 0]))
    } else {
      // a ring wide enough that the nodes on it do not touch
      const radius = Math.max(key * 130, (nodes.length * 64) / (2 * Math.PI))
      nodes.forEach((n, i) => {
        const angle = (2 * Math.PI * (i + 0.5)) / nodes.length - Math.PI / 2 + key * 0.35
        out.set(n.id, [Math.cos(angle) * radius, Math.sin(angle) * radius])
      })
    }
  }

  return out
}

// ---- size

const degreeOf = graph => {
  const d = new Map()
  for (const e of graph.edges.values()) {
    d.set(e.source, (d.get(e.source) || 0) + 1)
    d.set(e.target, (d.get(e.target) || 0) + 1)
  }
  return d
}

// Bigger for what is more connected, within bounds that keep labels readable
export function nodeSize (node, degree = 0) {
  if (node.root) return 50
  if (node.more) return 22
  if (node.cluster) return 42
  return 26 + Math.min(degree, 8) * 2
}

// ---- the option

// The echarts option that draws the graph.
//   moduleNames   moduleID -> name
//   selectedID    node the panel shows
//   highlightIDs  Set of nodes a search found (the others dim)
//   hiddenModules Set of moduleIDs switched off in the legend
//   layout        'force' | 'radial' | 'layers'
//   frozen        keep every placed node where it is
export function toEchartsOption (graph, {
  themeVars = {}, moduleNames = {}, showLegend = true, hints = {}, selectedID = '',
  highlightIDs = new Set(), hiddenModules = new Set(), layout = 'force', frozen = false, rootID = '',
} = {}) {
  const colors = palette(themeVars)
  const cats = categoryList(graph, moduleNames)
  const categoryIndex = new Map(cats.map((c, i) => [c.moduleID, i]))
  const text = themeVars['font-regular'] || undefined
  const background = themeVars.background || themeVars.white || '#ffffff'
  const primary = themeVars.primary || '#0d6efd'
  const dark = themeVars.dark || '#212529'
  const secondary = themeVars.secondary || '#6c757d'

  const categories = cats.map((c, i) => ({
    name: c.name,
    symbol: SYMBOLS[i % SYMBOLS.length],
    itemStyle: { color: colors[i % colors.length] },
  }))

  const degree = degreeOf(graph)
  const searching = highlightIDs.size > 0
  const near = selectedID && graph.nodes.has(selectedID) ? neighbours(graph, selectedID) : null
  const active = id => (searching ? highlightIDs.has(id) : near ? near.has(id) : true)

  const positions = layout === 'force' ? null : computeLayout(graph, rootID || [...graph.nodes.values()].find(n => n.root)?.id, layout)

  const nodes = [...graph.nodes.values()].map(n => {
    const on = active(n.id)
    const ring = n.status && n.status.color

    let itemStyle
    if (n.id === selectedID) {
      itemStyle = { borderColor: primary, borderWidth: 5, shadowBlur: 12, shadowColor: primary }
    } else if (searching && highlightIDs.has(n.id)) {
      itemStyle = { borderColor: primary, borderWidth: 4, shadowBlur: 10, shadowColor: primary }
    } else if (ring) {
      itemStyle = { borderColor: ring, borderWidth: 4, ...(n.expanded || n.root ? { shadowBlur: 8, shadowColor: dark } : {}) }
    } else if (n.expanded || n.root) {
      itemStyle = { borderColor: dark, borderWidth: 3 }
    } else if (n.more || n.cluster) {
      itemStyle = { borderType: 'dashed', borderWidth: 2, borderColor: secondary }
    }
    if (!on) itemStyle = { ...(itemStyle || {}), opacity: 0.22 }

    const pos = positions ? positions.get(n.id) : n.pos
    const cat = categoryIndex.get(n.moduleID)

    return {
      id: n.id,
      name: n.label,
      category: cat,
      symbol: SYMBOLS[(cat || 0) % SYMBOLS.length],
      symbolSize: nodeSize(n, degree.get(n.id) || 0),
      itemStyle,
      // placed nodes stay put; a new one is let to find its place
      x: pos ? pos[0] : undefined,
      y: pos ? pos[1] : undefined,
      fixed: positions ? true : (frozen ? !!n.pos : !!n.settled),
      // labels that overlap: the ones with the higher z2 stay
      z2: n.id === selectedID ? 10 : n.root ? 9 : highlightIDs.has(n.id) ? 8 : n.expanded ? 4 : 1,
      label: { show: true, color: text, opacity: on ? 1 : 0.25, textBorderColor: background, textBorderWidth: 3 },
      hint: n.more ? hints.more : n.cluster ? hints.cluster : n.expanded ? hints.open : hints.expand,
      moduleName: moduleNames[n.moduleID] || n.moduleID,
      full: n.full || n.label,
      details: n.details || [],
      status: n.status,
    }
  })

  const edges = [...graph.edges.values()].map(e => {
    const touching = selectedID && (e.source === selectedID || e.target === selectedID)
    const on = searching ? highlightIDs.has(e.source) && highlightIDs.has(e.target) : near ? touching : true
    return {
      source: e.source,
      target: e.target,
      value: e.label,
      // the name of the field is only shown on the edges of the selected node,
      // and on the one under the pointer (see emphasis)
      label: { show: !!touching && !!e.label, formatter: e.label, fontSize: 10, color: text, textBorderColor: background, textBorderWidth: 2 },
      lineStyle: on ? undefined : { opacity: 0.08 },
    }
  })

  const selected = {}
  for (const c of cats) if (hiddenModules.has(c.moduleID)) selected[c.name] = false

  return {
    animationDurationUpdate: 300,
    tooltip: {
      confine: true,
      formatter: ({ dataType, data }) => {
        if (dataType !== 'node') return escape(data.value || '')
        const rows = (data.details || []).map(d => `${escape(d.label)}: ${escape(d.value)}`)
        const status = data.status && data.status.text ? `<b style="color:${escape(data.status.color)}">●</b> ${escape(data.status.text)}` : ''
        return [
          `<b>${escape(data.full || data.name)}</b>`,
          `<span style="opacity:.7">${escape(data.moduleName)}</span>`,
          status,
          ...rows,
          data.hint ? `<i>${escape(data.hint)}</i>` : '',
        ].filter(Boolean).join('<br>')
      },
    },
    legend: showLegend ? [{ data: cats.map(c => c.name), selected, textStyle: { color: text }, type: 'scroll', bottom: 0 }] : undefined,
    series: [{
      type: 'graph',
      layout: layout === 'force' ? 'force' : 'none',
      roam: true,
      draggable: true,
      categories,
      data: nodes,
      links: edges,
      edgeSymbol: ['none', 'arrow'],
      edgeSymbolSize: 8,
      labelLayout: { hideOverlap: true },
      edgeLabel: { show: false },
      lineStyle: { color: 'source', opacity: 0.7, curveness: 0.1 },
      force: { repulsion: 260, edgeLength: [70, 140], gravity: 0.08, layoutAnimation: !frozen },
      emphasis: { focus: 'adjacency', lineStyle: { width: 3 }, edgeLabel: { show: true } },
      blur: { itemStyle: { opacity: 0.25 }, lineStyle: { opacity: 0.1 }, label: { opacity: 0.25 } },
    }],
  }
}
