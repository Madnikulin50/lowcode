<template>
  <Wrap
    v-bind="$props"
    class="fixed-corner-container"
    @refreshBlock="refresh"
  >
    <div
      v-if="isProcessing"
      class="d-flex align-items-center justify-content-center h-100"
    >
      <span class="spinner-border spinner-border-sm" />
    </div>

    <label
      v-else-if="error"
      class="text-primary p-3"
    >
      {{ error }}
    </label>

    <template v-else>
      <button
        v-if="!block.options?.hideBrainButton"
        title="Ask about metrics"
        :disabled="editable"
        class="btn btn-outline-light text-secondary border-0 fixed-corner-button btn-sm"
        @click="promptAiChat"
      >
        <font-awesome-icon :icon="['fas', 'brain']" />
      </button>

      <!-- Record-style layout -->
      <div
        v-if="options.likeRecordList !== false"
        class="rb mb-metrics px-3 pt-3"
        :class="densityClass"
      >
        <!-- Title metrics -->
        <div v-if="headerTitles.length" class="rb-header mb-3 pe-4">
          <div
            v-for="item in headerTitles"
            :key="`title-${item.index}`"
            class="mb-2"
            :class="{ pointer: item.metric.drillDown?.enabled }"
            @click="drillDown(item.metric, item.index)"
          >
            <metric-item
              v-for="(v, vi) in item.values"
              :key="vi"
              :metric="item.metric"
              :options="options"
              :hover="!!item.metric.drillDown?.enabled"
              :value="v"
            />
          </div>
        </div>

        <!-- Badges -->
        <div v-if="headerBadges.length" class="rb-badges d-flex flex-wrap gap-2 mb-3">
          <div
            v-for="item in headerBadges"
            :key="`badge-${item.index}`"
            :class="{ pointer: item.metric.drillDown?.enabled }"
            @click="drillDown(item.metric, item.index)"
          >
            <metric-item
              v-for="(v, vi) in item.values"
              :key="vi"
              :metric="item.metric"
              :options="options"
              :hover="!!item.metric.drillDown?.enabled"
              :value="v"
            />
          </div>
        </div>

        <!-- Meta strip — same itemsPerRow grid as body metrics -->
        <div
          v-if="metaItems.length"
          class="rb-meta mb-3 metric-grid"
          :class="gridModifierClass"
          :style="gridStyle"
          :data-items-per-row="resolvedItemsPerRow"
        >
          <div
            v-for="item in metaItems"
            :key="`meta-${item.index}`"
            class="field-container"
            :class="{ pointer: item.metric.drillDown?.enabled }"
            @click="drillDown(item.metric, item.index)"
          >
            <metric-item
              v-for="(v, vi) in item.values"
              :key="vi"
              :metric="item.metric"
              :options="options"
              :hover="!!item.metric.drillDown?.enabled"
              :value="v"
            />
          </div>
        </div>

        <!-- Heroes + default/balloon body share one grid so itemsPerRow applies -->
        <div
          class="metric-grid"
          :class="gridModifierClass"
          :style="gridStyle"
          :data-items-per-row="resolvedItemsPerRow"
        >
          <div
            v-for="item in heroItems"
            :key="`hero-${item.index}`"
            class="field-container"
            :class="{ pointer: item.metric.drillDown?.enabled }"
            @click="drillDown(item.metric, item.index)"
          >
            <metric-item
              v-for="(v, vi) in item.values"
              :key="vi"
              :metric="item.metric"
              :options="options"
              :hover="!!item.metric.drillDown?.enabled"
              :value="v"
            />
          </div>

          <template v-for="(section, sIdx) in displaySections" :key="`section-${sIdx}`">
            <h6
              v-if="section.title"
              class="rb-section-title text-muted text-uppercase metric-span-full"
            >
              {{ section.title }}
            </h6>
            <div
              v-for="item in section.items"
              :key="`body-${item.index}`"
              class="field-container"
              :class="{ pointer: item.metric.drillDown?.enabled }"
              :style="spanStyle(item.metric)"
              @click="drillDown(item.metric, item.index)"
            >
              <metric-item
                v-for="(v, vi) in item.values"
                :key="vi"
                :metric="withTopKMeta(item)"
                :options="options"
                :hover="!!item.metric.drillDown?.enabled"
                :value="v"
                :bar-ratio="barRatioFor(item, section.items)"
              />
            </div>
          </template>
        </div>
      </div>

      <!-- Legacy centered SVG layout -->
      <div
        v-else
        class="fixed-corner-container metric-grid"
        :class="gridModifierClass"
        :style="gridStyle"
        :data-items-per-row="resolvedItemsPerRow"
      >
        <div
          v-for="(m, mi) in options.metrics"
          :key="mi"
          class="d-flex align-items-center justify-content-center overflow-hidden"
          :class="{ 'h-100': m.valueStyle?.notFitVertical !== true }"
          :style="spanStyle(m)"
        >
          <div
            v-for="(v, i) in formatResponse(m, mi)"
            :key="i"
            class="py-1 px-2"
            :class="{
              pointer: m.drillDown?.enabled,
              'w-100': m.valueStyle?.notFitHorizontal !== true,
              'h-100': m.valueStyle?.notFitVertical !== true,
            }"
            @click="drillDown(m, mi)"
          >
            <metric-item
              :metric="withTopKMeta({ metric: m, index: mi, values: formatResponse(m, mi), role: metricRole(m) })"
              :options="options"
              :hover="!!m.drillDown?.enabled"
              :value="v"
              :bar-ratio="barRatioFor({ metric: m, index: mi, values: formatResponse(m, mi), role: metricRole(m) })"
            />
          </div>
        </div>
      </div>
    </template>
  </Wrap>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed, watch, onMounted, onBeforeUnmount, inject } from 'vue'
import { debounce } from 'lodash'
import { usePageBlockBase } from './usePageBlockBase'
import Wrap from './Wrap/index.js'
import MetricItem from './Metric/Item'
import numeral from 'numeral'
import moment from 'moment'
import { NoID, compose } from 'corteza-lib/js/dist'
import { evalPrefilterOrSkip, evaluatePrefilter, isFieldInFilter } from 'corteza-webapp-compose/src/lib/record-filter'
import { useStore } from '../../store'

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
  previewMode: { type: Boolean, required: false, default: false },
})

const emit = defineEmits(['errors'])
const $auth = inject('$auth')
const $ComposeAPI = inject('$ComposeAPI')
const store = useStore()

const { options, isProcessing, processing, browserLocale, refreshBlock, setBaseDefaultValues } = usePageBlockBase(props, emit)

// ${variables.x} in a metric's filter/transform expression, sourced from
// the page's session-only variable values (see PageBlocks/Variables).
const pageVariables = computed(() => store.pageVariables.getValuesForPage(props.page.pageID))

const densityClass = computed(() =>
  options.value.density === 'compact' ? 'rb-density-compact' : 'rb-density-comfortable',
)

const resolvedItemsPerRow = computed(() => {
  const opts = options.value || props.block?.options || {}
  const s = opts.itemsPerRow == null ? '' : String(opts.itemsPerRow)
  if (s === '1' || s === '2' || s === '4' || s === 'auto') return s
  const legacy = opts.recordFieldLayoutOption
  if (legacy === 'wrap') return '2'
  if (legacy === 'noWrap') return 'auto'
  return '1'
})

const gridModifierClass = computed(() => `metric-grid-${resolvedItemsPerRow.value}`)

const gridStyle = computed(() => {
  const per = resolvedItemsPerRow.value
  const compact = options.value.density === 'compact'
  const colGap = compact ? '0.75rem' : '1rem'
  const n = per === '2' ? 2 : per === '4' ? 4 : 1
  const style = {
    display: 'flex',
    flexWrap: 'wrap',
    alignItems: 'stretch',
    columnGap: colGap,
    rowGap: compact ? '0.5rem' : '0.75rem',
    width: '100%',
    minWidth: 0,
    '--metric-gap': colGap,
    '--metric-n': String(n),
  }
  if (per === 'auto') {
    style['--metric-basis'] = '10rem'
  } else if (n <= 1) {
    style['--metric-basis'] = '100%'
  } else {
    // Max N per row; leftover items in the last row grow to fill the block.
    style['--metric-basis'] = 'calc((100% - (var(--metric-n) - 1) * var(--metric-gap)) / var(--metric-n))'
  }
  return style
})

function spanStyle (metric) {
  if (isBalloonRole(metric) && metric.balloonFullWidth === true) {
    return { flex: '1 1 100%', minWidth: '100%', maxWidth: '100%' }
  }
  return {}
}

function metricRole (m) {
  const r = m?.role || 'default'
  return r === 'topK' ? 'balloon' : r
}

function isBalloonRole (m) {
  return metricRole(m) === 'balloon'
}

function numericAbs (item) {
  const raw = item?.values?.[0]?.value
  const n = typeof raw === 'number' ? raw : Number(String(raw).replace(/[^\d.-]/g, ''))
  return Number.isFinite(n) ? Math.abs(n) : 0
}

function barRatioFor (item, peers) {
  if (!isBalloonRole(item.metric)) return 1
  const list = (peers || preparedMetrics.value).filter(i => isBalloonRole(i.metric))
  const max = Math.max(0, ...list.map(numericAbs))
  if (!max) return 1
  return Math.min(1, numericAbs(item) / max)
}

function withTopKMeta (item) {
  const m = item.metric || {}
  if (!isBalloonRole(m)) return m
  const list = preparedMetrics.value.filter(i => isBalloonRole(i.metric))
  const balloonIndex = Math.max(0, list.findIndex(i => i.index === item.index))
  return { ...m, role: 'balloon', _balloonIndex: balloonIndex, _topKIndex: balloonIndex }
}

function isMetricEmpty (values) {
  if (!values?.length) return true
  return values.every(({ value }) => value === undefined || value === null || value === '' || (typeof value === 'number' && Number.isNaN(value)))
}

const preparedMetrics = computed(() => {
  return (options.value.metrics || []).map((metric, index) => {
    const values = formatResponse(metric, index)
    return { metric, index, values, role: metricRole(metric), empty: isMetricEmpty(values) }
  }).filter(item => {
    if (!item.metric.moduleID && !item.values.length) return false
    if (options.value.hideEmptyMetrics && item.empty) return false
    return true
  })
})

const headerTitles = computed(() => preparedMetrics.value.filter(i => i.role === 'title'))
const headerBadges = computed(() => preparedMetrics.value.filter(i => i.role === 'badge'))
const metaItems = computed(() => preparedMetrics.value.filter(i => i.role === 'meta'))
const heroItems = computed(() => preparedMetrics.value.filter(i => i.role === 'hero'))

const bodyItems = computed(() => {
  const special = new Set(['title', 'badge', 'meta', 'hero'])
  return preparedMetrics.value.filter(i => !special.has(i.role))
})

const displaySections = computed(() => {
  const body = bodyItems.value
  const sections = (options.value.sections || []).filter(s => s && (s.title || (s.metrics && s.metrics.length)))
  if (!sections.length) {
    return [{ title: '', items: body }]
  }

  const used = new Set()
  const result = []
  for (const section of sections) {
    const idxs = new Set((section.metrics || []).map(Number))
    const items = body.filter(i => idxs.has(i.index))
    items.forEach(i => used.add(i.index))
    if (items.length || section.title) {
      result.push({ title: section.title || '', items })
    }
  }
  const rest = body.filter(i => !used.has(i.index))
  if (rest.length) result.push({ title: '', items: rest })
  return result.length ? result : [{ title: '', items: body }]
})

const error = ref(undefined)
const reports = ref([])
const abortableRequests = ref([])

watch(() => [props.record?.recordID, props.loadingRecord], () => { if (!props.loadingRecord) refresh() }, { immediate: true })
watch(() => options.value, debounce(() => { if (!props.loadingRecord) refresh() }, 300), { deep: true })

onMounted(() => { createEvents() })
onBeforeUnmount(() => { abortRequests(); destroyEvents(); setBaseDefaultValues() })

onMounted(() => { refreshBlock(refresh) })

function createEvents () {
  window.addEventListener('metric.update', refresh)
  window.addEventListener('drill-down-chart', drillDown)
  window.addEventListener('module-records-updated', refreshOnRelatedRecordsUpdate)
  window.addEventListener('record-field-change', refetchOnPrefilterValueChange)
  window.addEventListener('page-variable-change', refetchOnPageVariableChange)
  window.addEventListener('refetch-records', refresh)
}

function refetchOnPrefilterValueChange ({ fieldName }) {
  const { metrics } = options.value
  if (metrics.some(({ filter }) => isFieldInFilter(fieldName, filter))) refresh()
}

function refetchOnPageVariableChange ({ detail: { pageID, fieldName } } = {}) {
  if (pageID !== props.page.pageID) return
  const needle = `variables.${fieldName}`
  const { metrics } = options.value
  if (metrics.some(({ filter, transformFx }) => isFieldInFilter(needle, filter) || isFieldInFilter(needle, transformFx))) refresh()
}

function formatResponse (m, i) {
  const vals = reports.value[i]
  if (!vals) return []
  return vals.map(({ label, value }) => {
    if (m.numberFormat) {
      const n = typeof value === 'number' ? value : Number(value)
      if (Number.isFinite(n)) value = numeral(n).format(m.numberFormat)
    }
    if (m.dateFormat) label = moment(label).format(m.dateFormat)
    return { label, value }
  })
}

async function refresh () {
  error.value = undefined
  processing.value = true
  try {
    const rtr = []
    const namespaceID = props.namespace.namespaceID
    const reporter = r => {
      const { response, cancel } = $ComposeAPI.recordReportCancellable({ ...r, namespaceID })
      abortableRequests.value.push(cancel)
      return response()
    }
    for (const m of options.value.metrics) {
      if (m.moduleID) {
        const auxM = { ...m }
        if (auxM.filter) {
          const { skip, filter } = evalPrefilterOrSkip(auxM.filter, {
            record: props.record, user: $auth.user || {}, recordID: (props.record || {}).recordID || NoID,
            ownerID: (props.record || {}).ownedBy || NoID, userID: ($auth.user || {}).userID || NoID,
            loadingRecord: !!props.loadingRecord, variables: pageVariables.value,
          })
          if (skip) {
            rtr.push([])
            continue
          }
          auxM.filter = filter
        }
        if (auxM.transformFx) {
          auxM.transformFx = evaluatePrefilter(auxM.transformFx, {
            record: props.record, user: $auth.user || {}, recordID: (props.record || {}).recordID || NoID,
            ownerID: (props.record || {}).ownedBy || NoID, userID: ($auth.user || {}).userID || NoID,
            variables: pageVariables.value,
          })
        }
        const vals = await props.block.fetch({ m: auxM }, reporter)
        rtr.push(vals)
      } else {
        rtr.push([])
      }
    }
    reports.value = rtr
    setTimeout(() => { processing.value = false }, 300)
  } catch (e) {
    error.value = e.message || 'Error'
    setTimeout(() => { processing.value = false }, 300)
  }
}

function drillDown ({ label: name = '', filter, moduleID, drillDown }, metricIndex) {
  if (!drillDown?.enabled) return
  if (drillDown.blockID) {
    const { pageID = NoID } = props.page
    const { recordID = NoID } = props.record || {}
    const recordListUniqueID = [pageID, recordID, drillDown.blockID, false].map(v => v || NoID).join('-')
    window.dispatchEvent(new CustomEvent(`drill-down-recordList:${recordListUniqueID}`, { detail: filter }))
  } else {
    const metricID = `${props.block.blockID}-${name.replace(/\s+/g, '-').toLowerCase()}-${moduleID}-${metricIndex}`
    const { title } = props.block
    const { fields = [] } = options.value.metrics[metricIndex].drillDown.recordListOptions || {}
    const block = new compose.PageBlockRecordList({
      title: name || title || 'Metric drill-down',
      blockID: `drillDown-${metricID}`,
      options: { moduleID, fields, prefilter: filter, presort: '', hideRecordReminderButton: true, hideRecordViewButton: false, hideConfigureFieldsButton: false, hideImportButton: true, enableRecordPageNavigation: true, selectable: true, allowExport: true, perPage: 14, showTotalCount: true, recordDisplayOption: 'modal' },
    })
    window.dispatchEvent(new CustomEvent('magnify-page-block', { detail: { block } }))
  }
}

function abortRequests () {
  abortableRequests.value.forEach((cancel) => cancel())
}

function refreshOnRelatedRecordsUpdate ({ moduleID } = {}) {
  const { metrics = [] } = options.value || {}
  const hasMatchingModule = metrics.some((m) => m.moduleID === moduleID)
  if (hasMatchingModule) refresh()
}

function promptAiChat () {
  const block = props.block
  const page = props.page
  const namespace = props.namespace
  const locale = browserLocale()
  let prompt = block.prompt || page.config.prompt || namespace.prompt || ''
  if (prompt.length === 0) {
    switch (locale) {
      case 'en-US': prompt = 'What is this? '; break
      case 'ru-RU': prompt = 'Что это за показатели? Зачем о чем они говорят? '; break
    }
  }
  prompt += '\r\n '
  prompt += page.title + '\r\n' + block.title + '\r\n'
  for (const mi in options.value.metrics) {
    const m = options.value.metrics[mi]
    if (m.moduleID) {
      const vals = formatResponse(m, mi).map(item => {
        if (item.label !== undefined) return item.label + ': ' + m.prefix + item.value + m.suffix
        return m.prefix + item.value + m.suffix
      })
      prompt += m.label + ' = ' + vals.join('; ') + '\r\n'
    }
  }
  window.dispatchEvent(new CustomEvent('show-chat-modal', { detail: { namespace: page.namespaceID, module: page.moduleID, prompt } }))
}

function destroyEvents () {
  window.removeEventListener('metric.update', refresh)
  window.removeEventListener('drill-down-chart', drillDown)
  window.removeEventListener('module-records-updated', refreshOnRelatedRecordsUpdate)
  window.removeEventListener('record-field-change', refetchOnPrefilterValueChange)
  window.removeEventListener('page-variable-change', refetchOnPageVariableChange)
  window.removeEventListener('refetch-records', refresh)
}
</script>

<style scoped lang="scss">
.fixed-corner-button {
  position: absolute;
  top: 20px;
  right: 0;
  transform: translateY(-50%);
  z-index: 10;
}

.field-col > * { margin-left: 1rem; margin-right: 1rem; }

.metric-grid {
  display: flex;
  flex-wrap: wrap;
  align-items: stretch;
  width: 100%;
  min-width: 0;
}

.metric-grid > .field-container,
.metric-grid > *:not(.metric-span-full) {
  flex: 1 1 var(--metric-basis, 100%);
  min-width: min(100%, var(--metric-basis, 100%));
  max-width: 100%;
  box-sizing: border-box;
}

.metric-grid > .metric-span-full {
  flex: 1 1 100%;
  min-width: 100%;
  max-width: 100%;
}

.metric-grid :deep(.mb-metric) {
  width: 100%;
}

.rb-meta {
  padding: 0.65rem 0;
  border-top: 1px solid var(--bs-border-color-translucent, rgba(0, 0, 0, 0.08));
  border-bottom: 1px solid var(--bs-border-color-translucent, rgba(0, 0, 0, 0.08));
}

.rb-section-title {
  font-size: 0.7rem;
  letter-spacing: 0.06em;
  font-weight: 600;
  margin-bottom: 0.75rem;
}

.rb-density-compact .rb-section-title { margin-bottom: 0.5rem; }
.rb-density-compact .rb-meta { padding: 0.4rem 0; }
.rb-density-compact :deep(.rb-title) { font-size: 1.25rem; }
.rb-density-compact :deep(.mb-metric-hero-value) { font-size: 1.5rem; }
.rb-density-compact :deep(.mb-balloon) { padding: 0.35rem 0.9rem; }

.pointer { cursor: pointer; }
</style>

<style lang="scss">
/* Unscoped so Bootstrap width:100% on flex/grid children cannot collapse the row */
.mb-metrics .metric-grid,
.fixed-corner-container.metric-grid {
  display: flex !important;
  flex-wrap: wrap !important;
  align-items: stretch !important;
  width: 100% !important;
  min-width: 0 !important;
}
.mb-metrics .metric-grid > .field-container,
.mb-metrics .metric-grid > *:not(.metric-span-full),
.fixed-corner-container.metric-grid > *:not(.metric-span-full) {
  flex: 1 1 var(--metric-basis, 100%) !important;
  min-width: min(100%, var(--metric-basis, 100%)) !important;
  max-width: 100% !important;
  box-sizing: border-box;
}
.mb-metrics .metric-grid > .metric-span-full,
.fixed-corner-container.metric-grid > .metric-span-full {
  flex: 1 1 100% !important;
  min-width: 100% !important;
  max-width: 100% !important;
}
.mb-metrics .metric-grid .mb-metric {
  width: 100%;
}
</style>
