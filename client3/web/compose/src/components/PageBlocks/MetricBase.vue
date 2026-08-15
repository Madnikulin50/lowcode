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

        <!-- Meta strip -->
        <div v-if="metaItems.length" class="rb-meta mb-3">
          <div
            v-for="item in metaItems"
            :key="`meta-${item.index}`"
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

        <!-- Hero metrics -->
        <div
          v-for="item in heroItems"
          :key="`hero-${item.index}`"
          class="mb-3"
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

        <!-- Sections / default body -->
        <template v-for="(section, sIdx) in displaySections" :key="`section-${sIdx}`">
          <div v-if="section.items.length" class="rb-section" :class="{ 'mb-3': sIdx < displaySections.length - 1 }">
            <h6 v-if="section.title" class="rb-section-title text-muted text-uppercase">
              {{ section.title }}
            </h6>
            <div :class="sectionLayoutClass">
              <div
                v-for="item in section.items"
                :key="`body-${item.index}`"
                class="field-container"
                :class="[
                  columnWrapClass,
                  bodyColClass(item.metric),
                  options.density === 'compact' ? 'mb-2' : 'mb-3',
                  { pointer: item.metric.drillDown?.enabled },
                ]"
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
            </div>
          </div>
        </template>
      </div>

      <!-- Legacy centered SVG layout -->
      <div
        v-else
        class="fixed-corner-container"
        :class="fieldLayoutClass"
      >
        <div
          v-for="(m, mi) in options.metrics"
          :key="mi"
          class="d-flex align-items-center justify-content-center overflow-hidden"
          :class="{
            'h-100': m.valueStyle?.notFitVertical !== true,
            'px-3': false,
          }"
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
import { evaluatePrefilter, isFieldInFilter } from 'corteza-webapp-compose/src/lib/record-filter'

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

const { options, isProcessing, processing, browserLocale, refreshBlock, setBaseDefaultValues } = usePageBlockBase(props, emit)

const fieldLayoutClass = computed(() => {
  const classes = { default: 'd-flex flex-column', noWrap: 'd-flex gap-2', wrap: 'row g-0' }
  return classes[options.value.recordFieldLayoutOption]
})

const densityClass = computed(() =>
  options.value.density === 'compact' ? 'rb-density-compact' : 'rb-density-comfortable',
)

const sectionLayoutClass = computed(() => {
  const classes = { default: 'd-flex flex-column', noWrap: 'd-flex gap-2 flex-wrap', wrap: 'row g-2' }
  return classes[options.value.recordFieldLayoutOption] || classes.default
})

const columnWrapClass = computed(() => {
  if (options.value.recordFieldLayoutOption === 'noWrap') return 'field-col'
  return ''
})

function bodyColClass (metric) {
  if (isBalloonRole(metric)) return 'col-12'
  if (options.value.recordFieldLayoutOption === 'wrap') return 'col-md-6'
  return ''
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

watch(() => props.record?.recordID, () => { refresh() }, { immediate: true })
watch(() => options.value, debounce(() => { refresh() }, 300), { deep: true })

onMounted(() => { createEvents() })
onBeforeUnmount(() => { abortRequests(); destroyEvents(); setBaseDefaultValues() })

onMounted(() => { refreshBlock(refresh) })

function createEvents () {
  window.addEventListener('metric.update', refresh)
  window.addEventListener('drill-down-chart', drillDown)
  window.addEventListener('module-records-updated', refreshOnRelatedRecordsUpdate)
  window.addEventListener('record-field-change', refetchOnPrefilterValueChange)
  window.addEventListener('refetch-records', refresh)
}

function refetchOnPrefilterValueChange ({ fieldName }) {
  const { metrics } = options.value
  if (metrics.some(({ filter }) => isFieldInFilter(fieldName, filter))) refresh()
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
        if (auxM.filter && !props.record && (auxM.filter.includes('${record') || auxM.filter.includes('${ownerID}'))) {
          rtr.push([])
          continue
        }
        if (auxM.filter) {
          auxM.filter = evaluatePrefilter(auxM.filter, {
            record: props.record, user: $auth.user || {}, recordID: (props.record || {}).recordID || NoID,
            ownerID: (props.record || {}).ownedBy || NoID, userID: ($auth.user || {}).userID || NoID,
          })
        }
        if (auxM.transformFx) {
          auxM.transformFx = evaluatePrefilter(auxM.transformFx, {
            record: props.record, user: $auth.user || {}, recordID: (props.record || {}).recordID || NoID,
            ownerID: (props.record || {}).ownedBy || NoID, userID: ($auth.user || {}).userID || NoID,
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
      options: { moduleID, fields, prefilter: filter, presort: 'createdAt DESC', hideRecordReminderButton: true, hideRecordViewButton: false, hideConfigureFieldsButton: false, hideImportButton: true, enableRecordPageNavigation: true, selectable: true, allowExport: true, perPage: 14, showTotalCount: true, recordDisplayOption: 'modal' },
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

.rb-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem 1.5rem;
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
.rb-density-compact .rb-meta { gap: 0.5rem 1rem; padding: 0.4rem 0; }
.rb-density-compact :deep(.rb-title) { font-size: 1.25rem; }
.rb-density-compact :deep(.mb-metric-hero-value) { font-size: 1.5rem; }
.rb-density-compact :deep(.mb-balloon) { padding: 0.35rem 0.9rem; }

.field-container:has(.mb-balloon) {
  margin-bottom: 0.5rem !important;
}

.pointer { cursor: pointer; }
</style>
