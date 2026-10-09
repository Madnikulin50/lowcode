<template>
  <Wrap v-bind="$props" :scrollable-body="false" @refreshBlock="refresh">
    <div
      v-if="!hasRecord"
      class="p-3 text-secondary"
    >
      {{ $t('recordGraph.saveFirst') }}
    </div>

    <div
      v-else
      ref="container"
      class="record-graph position-relative h-100"
      :class="{ 'bg-body': isFullscreen }"
      style="min-height: 320px;"
    >
      <e-charts
        ref="chartRef"
        :key="renderKey"
        :option="option"
        autoresize
        class="w-100 h-100"
        @click="onClick"
        @dblclick="onDoubleClick"
        @legendselectchanged="onLegend"
      />

      <div
        v-if="options.showToolbar"
        class="record-graph-toolbar position-absolute top-0 end-0 m-2 d-flex flex-wrap justify-content-end align-items-center gap-1"
      >
        <input
          v-model="query"
          type="search"
          class="form-control form-control-sm"
          style="width: 9rem;"
          :placeholder="$t('recordGraph.toolbar.search')"
          :aria-label="$t('recordGraph.toolbar.search')"
        >

        <select
          :value="layoutKind"
          class="form-select form-select-sm w-auto"
          :title="$t('recordGraph.toolbar.layout')"
          @change="setLayout($event.target.value)"
        >
          <option
            v-for="l in layouts"
            :key="l"
            :value="l"
          >
            {{ $t(`recordGraph.layout.${l}`) }}
          </option>
        </select>

        <div class="btn-group btn-group-sm">
          <button
            type="button"
            class="btn btn-outline-secondary"
            :title="$t('recordGraph.toolbar.fit')"
            @click="fit"
          >
            <font-awesome-icon :icon="['fas', 'crosshairs']" />
          </button>
          <button
            type="button"
            class="btn btn-outline-secondary"
            :title="$t('recordGraph.toolbar.relayout')"
            @click="relayout"
          >
            <font-awesome-icon :icon="['fas', 'sync']" />
          </button>
          <button
            type="button"
            class="btn btn-outline-secondary"
            :class="{ active: frozen }"
            :title="frozen ? $t('recordGraph.toolbar.unfreeze') : $t('recordGraph.toolbar.freeze')"
            :aria-pressed="frozen"
            @click="toggleFreeze"
          >
            <font-awesome-icon :icon="['fas', frozen ? 'lock' : 'lock-open']" />
          </button>
          <button
            type="button"
            class="btn btn-outline-secondary"
            :title="$t('recordGraph.toolbar.collapse')"
            :disabled="loading"
            @click="collapseAll"
          >
            <font-awesome-icon :icon="['fas', 'compress-alt']" />
          </button>
          <button
            type="button"
            class="btn btn-outline-secondary"
            :title="$t('recordGraph.toolbar.png')"
            @click="exportPng"
          >
            <font-awesome-icon :icon="['fas', 'download']" />
          </button>
          <button
            type="button"
            class="btn btn-outline-secondary"
            :title="isFullscreen ? $t('recordGraph.toolbar.exitFullscreen') : $t('recordGraph.toolbar.fullscreen')"
            @click="toggleFullscreen"
          >
            <font-awesome-icon :icon="['fas', isFullscreen ? 'compress' : 'expand']" />
          </button>
        </div>
      </div>

      <div
        v-if="selectedNode"
        class="record-graph-selected card shadow-sm position-absolute start-0 m-2"
        style="max-width: 20rem; bottom: 2.4rem;"
      >
        <div class="card-body p-2">
          <div class="d-flex align-items-start gap-2">
            <div class="flex-fill overflow-hidden">
              <div class="fw-semibold text-break">
                {{ selectedNode.full || selectedNode.label }}
              </div>
              <div class="small text-secondary">
                {{ moduleNameOf(selectedNode.moduleID) }}
              </div>
              <div
                v-if="selectedNode.status && selectedNode.status.text"
                class="small"
              >
                <span :style="{ color: selectedNode.status.color }">●</span> {{ selectedNode.status.text }}
              </div>
            </div>
            <button
              type="button"
              class="btn-close btn-sm"
              :aria-label="$t('recordGraph.close')"
              @click="selected = ''"
            />
          </div>

          <div
            v-for="d in selectedNode.details"
            :key="d.label"
            class="small text-break"
          >
            <span class="text-secondary">{{ d.label }}:</span> {{ d.value }}
          </div>

          <div class="d-flex flex-wrap gap-1 mt-2">
            <button
              v-if="canOpen(selectedNode)"
              type="button"
              class="btn btn-sm btn-primary"
              @click="openRecord(selectedNode)"
            >
              {{ $t('recordGraph.open') }}
            </button>
            <button
              v-if="!selectedNode.expanded"
              type="button"
              class="btn btn-sm btn-outline-secondary"
              :disabled="loading"
              @click="expandSelected"
            >
              {{ $t('recordGraph.expand') }}
            </button>
          </div>
        </div>
      </div>

      <div
        v-if="loading"
        class="position-absolute top-0 start-0 m-2 d-flex align-items-center text-secondary small"
      >
        <span class="spinner-border spinner-border-sm me-2" />{{ $t('recordGraph.loading') }}
      </div>

      <div
        v-else-if="error"
        class="position-absolute top-0 start-0 m-2 text-danger small"
      >
        {{ error }}
      </div>

      <div
        v-else-if="alone"
        class="position-absolute top-0 start-0 m-2 text-secondary small"
      >
        {{ $t('recordGraph.nothing') }}
      </div>

      <div
        v-if="dropped"
        class="position-absolute bottom-0 end-0 m-2 text-secondary small"
      >
        {{ $t('recordGraph.truncated', { n: dropped }) }}
      </div>
    </div>
  </Wrap>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, reactive, shallowRef, computed, watch, onMounted, onBeforeUnmount, inject } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { NoID } from 'corteza-lib/js/dist'
import { debounce } from 'lodash'
import { usePageBlockBase } from './../usePageBlockBase'
import { useStore } from '../../../store'
import {
  createGraph, addRecord, addEdge, parentRefs, childQueries, moreID, toEchartsOption, categoryList, searchMatches,
  applyPositions, seedNear, forgetPositions, inheritPositions, shouldCluster, clusterNode, LAYOUTS,
} from 'corteza-webapp-compose/src/lib/record-graph'
import Wrap from './../Wrap/index.js'

const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  blockIndex: { type: Number, default: -1 },
  namespace: { type: Object, required: true },
  page: { type: Object, required: true },
  blocks: { type: Array, default: () => [] },
  block: { type: Object, required: true },
  module: { type: Object, required: false, default: undefined },
  record: { type: Object, required: false, default: undefined },
  mode: { type: String, required: false, default: '' },
  editable: { type: Boolean, required: false, default: false },
  resizing: { type: Boolean, required: false, default: false },
  magnified: { type: Boolean, required: false, default: false },
  unsavedBlocks: { type: Set, default: () => new Set() },
  loadingRecord: { type: Boolean, required: false, default: false },
  errors: { type: Object, required: false, default: () => ({}) },
})

const emit = defineEmits(['errors'])

const $ComposeAPI = inject('$ComposeAPI')
const store = useStore()
const router = useRouter()
const route = useRoute()
const { options } = usePageBlockBase(props, emit)

const layouts = LAYOUTS
const graph = shallowRef(createGraph())
const version = ref(0) // the graph is changed in place: this tells the option to be made again
const loading = ref(false)
const selected = ref('') // id of the node the panel shows
const error = ref('')
const themeKey = ref(0)
const renderKey = ref(0) // a new key makes a new chart: the view is back to the start
const layoutKind = ref(options.value.layout || 'force')
const frozen = ref(false)
const query = ref('')
const hiddenModules = reactive(new Set()) // modules switched off in the legend
const isFullscreen = ref(false)
const container = ref(null)
const chartRef = ref(null)
const remembered = new Set() // nodes the user opened, to open them again after a refresh
const pending = new Set() // requests in flight, to cancel them with the block
let generation = 0

const hasRecord = computed(() => !!props.record && props.record.recordID && props.record.recordID !== NoID)
const alone = computed(() => { version.value; return graph.value.nodes.size <= 1 })
const selectedNode = computed(() => { version.value; const n = graph.value.nodes.get(selected.value); return n && !n.more && !n.cluster ? n : undefined })
const dropped = computed(() => { version.value; return graph.value.dropped })

const exclude = computed(() => (options.value.excludeModules || []).map(String))
// how the records of a module are named, and which field colours its nodes
const templates = computed(() => Object.fromEntries((options.value.labels || []).filter(l => l.template).map(l => [String(l.moduleID), l.template])))
const statusFields = computed(() => Object.fromEntries((options.value.labels || []).filter(l => l.statusField).map(l => [String(l.moduleID), l.statusField])))
const locale = typeof navigator !== 'undefined' ? navigator.language : undefined
const moduleOf = id => store.module.getByID(String(id))
const moduleNameOf = id => { const m = moduleOf(id); return (m && (m.name || m.handle)) || id }
const touch = () => { version.value++ }

function themeVariables () {
  const style = getComputedStyle(document.documentElement)
  const vars = ['white', 'black', 'primary', 'secondary', 'success', 'info', 'warning', 'danger', 'light', 'dark', 'font-regular'].reduce((acc, name) => {
    acc[name] = style.getPropertyValue(`--${name}`).trim()
    return acc
  }, {})
  // what text is drawn over: its outline takes the colour of the page
  vars.background = getComputedStyle(document.body).backgroundColor
  return vars
}
const themeVars = computed(() => { themeKey.value; return themeVariables() })

const recordOptions = () => ({
  maxNodes: options.value.maxNodes,
  templates: templates.value,
  statusFields: statusFields.value,
  themeVars: themeVars.value,
  locale,
})

// ---- reading records

async function list (moduleID, params) {
  const { response, cancel } = $ComposeAPI.recordListCancellable({
    namespaceID: props.namespace.namespaceID,
    moduleID,
    ...params,
  })
  pending.add(cancel)
  try {
    return await response()
  } finally {
    pending.delete(cancel)
  }
}

// a module that cannot be read (no right to search it, gone) is left out of the
// graph; the others still draw
async function safeList (moduleID, params) {
  try {
    return await list(moduleID, params)
  } catch (e) {
    return { set: [], filter: {} }
  }
}

const chunks = (ids, n = 50) => Array.from({ length: Math.ceil(ids.length / n) }, (_, i) => ids.slice(i * n, i * n + n))

async function loadParents (g, node) {
  const added = []
  const mod = moduleOf(node.moduleID)

  const wanted = new Map() // moduleID -> Map(recordID -> [field, ...])
  for (const { field, moduleID, recordID } of parentRefs(node.record, mod, { exclude: exclude.value })) {
    if (!wanted.has(moduleID)) wanted.set(moduleID, new Map())
    const byRecord = wanted.get(moduleID)
    byRecord.set(recordID, [...(byRecord.get(recordID) || []), field])
  }

  for (const [moduleID, byRecord] of wanted) {
    const parentModule = moduleOf(moduleID)
    if (!parentModule) continue

    for (const ids of chunks([...byRecord.keys()])) {
      const { set = [] } = await safeList(moduleID, { query: ids.map(id => `recordID = ${id}`).join(' OR '), limit: ids.length, deleted: 0 })
      for (const r of set) {
        const fields = byRecord.get(String(r.recordID)) || []
        const parent = addRecord(g, r, parentModule, { ...recordOptions(), preferredField: fields[0]?.options?.labelField })
        if (!parent) continue
        seedNear(node, parent)
        for (const f of fields) addEdge(g, node.id, parent.id, f.name, f.label || f.name)
        added.push(parent)
      }
    }
  }

  return added
}

// The records of one list as children of a node. Many of them are one cluster
// node, opened by a click; a full page leaves a "more" node. `from` is where the
// new nodes start (the node, or the cluster / "more" node they replace).
function takeChildren (g, node, plan, result, added, { noCluster = false, from = node } = {}) {
  const { set = [], filter = {} } = result

  if (!noCluster && shouldCluster(set.length, options.value.clusterFrom)) {
    const cl = clusterNode(node, plan, result, moduleNameOf(plan.moduleID))
    g.nodes.set(cl.id, cl)
    seedNear(from, cl)
    g.edges.set(`${cl.id}>${node.id}`, { id: `${cl.id}>${node.id}`, source: cl.id, target: node.id, field: plan.field.name, label: plan.field.label || plan.field.name })
    return
  }

  for (const r of set) {
    const child = addRecord(g, r, plan.module, recordOptions())
    if (!child) continue
    seedNear(from, child)
    addEdge(g, child.id, node.id, plan.field.name, plan.field.label || plan.field.name)
    added.push(child)
  }

  if (set.length >= plan.limit && filter.nextPage) {
    const id = moreID(node.id, plan.moduleID, plan.field.name)
    const more = {
      id, more: true, moduleID: plan.moduleID, recordID: '', root: false, expanded: false,
      label: `+ ${$t('recordGraph.more')}`, plan, parent: node.id, cursor: filter.nextPage,
    }
    g.nodes.set(id, more)
    seedNear(from, more)
    g.edges.set(`${id}>${node.id}`, { id: `${id}>${node.id}`, source: id, target: node.id, field: plan.field.name, label: '' })
  }
}

async function loadChildren (g, node) {
  const added = []
  const plans = childQueries(store.module.set, node.moduleID, node.recordID, { exclude: exclude.value, limit: options.value.perModuleLimit })

  for (const plan of plans) {
    const result = await safeList(plan.moduleID, { query: plan.query, limit: plan.limit, sort: 'createdAt DESC', deleted: 0 })
    takeChildren(g, node, plan, result, added)
  }

  return added
}

async function loadMore (g, moreNode) {
  g.nodes.delete(moreNode.id)
  g.edges.delete(`${moreNode.id}>${moreNode.parent}`)

  const parent = g.nodes.get(moreNode.parent)
  const { plan } = moreNode
  const result = await safeList(plan.moduleID, { query: plan.query, limit: plan.limit, sort: 'createdAt DESC', pageCursor: moreNode.cursor, deleted: 0 })
  takeChildren(g, parent, plan, result, [], { noCluster: true, from: moreNode })
}

// a cluster is opened with the records it already holds: no new request
function openCluster (g, cl) {
  g.nodes.delete(cl.id)
  g.edges.delete(`${cl.id}>${cl.parent}`)

  const parent = g.nodes.get(cl.parent)
  takeChildren(g, parent, cl.plan, { set: cl.members, filter: cl.filter }, [], { noCluster: true, from: cl })
}

// opens a node: its parents and its children join the graph
async function expand (g, node) {
  if (node.expanded || node.more || node.cluster) return []

  node.expanded = true
  remembered.add(node.id)

  const added = []
  if (options.value.showParents) added.push(...await loadParents(g, node))
  if (options.value.showChildren) added.push(...await loadChildren(g, node))
  return added
}

// ---- building

function rootNode (g) {
  const r = props.record
  const record = { recordID: String(r.recordID), moduleID: String(r.moduleID), values: r.values || {} }
  return addRecord(g, record, moduleOf(record.moduleID) || props.module, { ...recordOptions(), root: true })
}

async function build ({ depth = options.value.depth } = {}) {
  if (!hasRecord.value || props.loadingRecord) return

  const mine = ++generation
  loading.value = true
  error.value = ''
  captureLayout()

  try {
    const g = createGraph()
    const root = rootNode(g)

    // `depth` levels around the record
    let level = [root]
    for (let d = 0; d < depth && level.length; d++) {
      const next = []
      for (const node of level) next.push(...await expand(g, node))
      level = next
      if (mine !== generation) return
    }

    // what the user had opened before the refresh
    for (let pass = 0; pass < 5; pass++) {
      const again = [...remembered].map(id => g.nodes.get(id)).filter(n => n && !n.expanded)
      if (!again.length) break
      for (const node of again) await expand(g, node)
      if (mine !== generation) return
    }

    // the nodes that were there stay where they were
    inheritPositions(graph.value, g)
    graph.value = g
    touch()
  } catch (e) {
    if (mine === generation) error.value = (e && e.message) || String(e)
  } finally {
    if (mine === generation) loading.value = false
  }
}

async function refresh () {
  remembered.clear()
  await build()
}

const rebuild = debounce(() => build(), 300)

// ---- where the nodes are

// What the force layout has made of the nodes, kept, so that opening a node
// does not move the ones already there.
function captureLayout () {
  if (layoutKind.value !== 'force') return

  const chart = chartRef.value && chartRef.value.chart
  if (!chart) return

  try {
    const data = chart.getModel().getSeriesByIndex(0).getData()
    const found = new Map()
    data.each(i => {
      const raw = data.getRawDataItem(i)
      const at = data.getItemLayout(i)
      if (raw && raw.id && at) found.set(raw.id, [at[0], at[1]])
    })
    applyPositions(graph.value, found)
  } catch (e) {
    // the chart is not ready: there is nothing to keep
  }
}

// ---- toolbar

function fit () {
  captureLayout()
  renderKey.value++
}

function relayout () {
  forgetPositions(graph.value)
  frozen.value = false
  touch()
  renderKey.value++
}

function toggleFreeze () {
  captureLayout()
  frozen.value = !frozen.value
  touch()
}

function setLayout (kind) {
  if (!LAYOUTS.includes(kind)) return
  captureLayout()
  layoutKind.value = kind
  if (kind === 'force') forgetPositions(graph.value)
  touch()
  renderKey.value++
}

async function collapseAll () {
  remembered.clear()
  selected.value = ''
  await build({ depth: 0 })
}

function exportPng () {
  const chart = chartRef.value && chartRef.value.chart
  if (!chart) return

  const url = chart.getDataURL({ type: 'png', pixelRatio: 2, backgroundColor: themeVars.value.background || '#ffffff' })
  const a = document.createElement('a')
  a.href = url
  a.download = `${moduleNameOf(props.record.moduleID)}-${props.record.recordID}.png`
  document.body.appendChild(a)
  a.click()
  a.remove()
}

function toggleFullscreen () {
  if (document.fullscreenElement) document.exitFullscreen()
  else if (container.value && container.value.requestFullscreen) container.value.requestFullscreen()
}

const onFullscreenChange = () => { isFullscreen.value = document.fullscreenElement === container.value }

// the legend switches modules off; the names carry counts that change, so what
// is off is kept by module
function onLegend (params) {
  hiddenModules.clear()
  for (const c of categoryList(graph.value, moduleNames.value)) {
    if (params.selected && params.selected[c.name] === false) hiddenModules.add(c.moduleID)
  }
}

// ---- clicks

const pageIDOf = moduleID => (store.page.set || []).find(p => String(p.moduleID) === String(moduleID))?.pageID

const inModal = computed(() => !!route.query.recordPageID || !!route.query.magnifiedBlockID)

// a record can be opened when the block allows it, it is not the page's own
// record, and its module has a record page
const canOpen = node => !!node && !node.more && !node.cluster && !node.root && options.value.displayOption !== 'doNothing' && !!pageIDOf(node.moduleID)

// as the other blocks do (see RecordOrganizerBase): the same tab, a new tab or a modal
function openRecord (node) {
  if (!canOpen(node)) return

  const pageID = pageIDOf(node.moduleID)
  const target = { name: 'page.record', params: { pageID, recordID: node.recordID }, query: null }

  if (options.value.displayOption === 'modal' || inModal.value) {
    window.dispatchEvent(new CustomEvent('show-record-modal', { detail: { recordID: node.recordID, recordPageID: pageID } }))
  } else if (options.value.displayOption === 'newTab') {
    window.open(router.resolve(target).href)
  } else {
    router.push(target)
  }
}

async function working (fn) {
  loading.value = true
  try {
    await fn()
    touch()
  } finally {
    loading.value = false
  }
}

const expandNode = node => working(() => expand(graph.value, node))
const expandSelected = () => selectedNode.value && !loading.value ? expandNode(selectedNode.value) : undefined

// a click selects a node and opens its relations; a node that is already open
// is only selected - the panel offers to open the record. "More" and cluster
// nodes are opened by the click.
async function onClick (params) {
  if (!params || params.dataType !== 'node' || loading.value) return

  const g = graph.value
  const node = g.nodes.get(params.data.id)
  if (!node) return

  captureLayout()

  if (node.more) return working(() => loadMore(g, node))
  if (node.cluster) return working(() => openCluster(g, node))

  selected.value = node.id
  touch()
  if (!node.expanded) await expandNode(node)
}

function onDoubleClick (params) {
  if (!params || params.dataType !== 'node') return
  openRecord(graph.value.nodes.get(params.data.id))
}

// ---- drawing

const moduleNames = computed(() => {
  version.value
  const names = {}
  for (const n of graph.value.nodes.values()) names[n.moduleID] = moduleNameOf(n.moduleID)
  return names
})

const highlightIDs = computed(() => { version.value; return searchMatches(graph.value, query.value, moduleNames.value) })

const option = computed(() => {
  version.value
  const root = [...graph.value.nodes.values()].find(n => n.root)

  return toEchartsOption(graph.value, {
    themeVars: themeVars.value,
    moduleNames: moduleNames.value,
    showLegend: options.value.showLegend,
    selectedID: selected.value,
    highlightIDs: highlightIDs.value,
    hiddenModules,
    layout: layoutKind.value,
    frozen: frozen.value,
    rootID: root ? root.id : '',
    hints: {
      expand: $t('recordGraph.hint.expand'),
      open: options.value.displayOption === 'doNothing' ? '' : $t('recordGraph.hint.open'),
      more: $t('recordGraph.hint.more'),
      cluster: $t('recordGraph.hint.cluster'),
    },
  })
})

let themeObserver
onMounted(() => {
  window.addEventListener('module-records-updated', rebuild)
  window.addEventListener('refetch-records', rebuild)
  document.addEventListener('fullscreenchange', onFullscreenChange)
  themeObserver = new MutationObserver(() => themeKey.value++)
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['data-bs-theme', 'class'] })
})

onBeforeUnmount(() => {
  generation++
  window.removeEventListener('module-records-updated', rebuild)
  window.removeEventListener('refetch-records', rebuild)
  document.removeEventListener('fullscreenchange', onFullscreenChange)
  rebuild.cancel()
  if (themeObserver) themeObserver.disconnect()
  pending.forEach(cancel => { try { cancel() } catch (e) { /* already done */ } })
})

watch(() => [props.record && props.record.recordID, props.loadingRecord], refresh, { immediate: true })
watch(() => [
  options.value.depth, options.value.maxNodes, options.value.perModuleLimit, options.value.showParents, options.value.showChildren,
  options.value.clusterFrom, (options.value.excludeModules || []).join(','), JSON.stringify(options.value.labels || []),
], refresh)
watch(() => options.value.layout, kind => { if (kind) setLayout(kind) })
</script>
