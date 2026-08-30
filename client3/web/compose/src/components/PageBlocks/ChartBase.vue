<template>
  <Wrap
    v-bind="$props"
    body-class="position-relative"
    @refreshBlock="refresh"
  >
    <chart-component
      v-if="chart"
      :key="key"
      :chart="chart"
      :record="record"
      :reporter="reporter"
      @drill-down="drillDown"
    />

    <button
      v-if="chart && !block.options?.hideBrainButton"
      class="btn btn-outline-light chart-brain-button position-absolute d-flex d-print-none border-0 px-1 text-secondary"
      title="Ask about metrics"
      @click="promptAiChat"
    >
      <font-awesome-icon :icon="['fas', 'brain']" />
    </button>
    <block-help-button
      v-if="chart"
      :block="block"
      variant="chart"
      :offset="!block.options?.hideBrainButton"
      :title="chart?.name"
      :description="chartDescription"
      :help="chartHelpBody"
    />

    <template v-if="options.liveFilterEnabled">
      <button
        class="btn btn-outline-light chart-filter-button position-absolute d-flex d-print-none border-0 px-1"
        :class="[
          hasLiveFilter ? 'text-primary' : 'text-secondary',
          hasSaveChartEnabled && 'save-chart-enabled',
          hasDataTableEnabled && 'table-enabled'
        ]"
        @click="showFilterModal"
      >
        <font-awesome-icon :icon="['fas', 'filter']" />
      </button>

      <div
        v-if="liveFilterModal.show"
        class="modal fade show d-block"
        tabindex="-1"
      >
        <div class="modal-dialog modal-dialog-centered modal-lg">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">{{ $t('chart.filter.modal.title') }}</h5>
              <button type="button" class="btn-close" data-bs-dismiss="modal" @click="liveFilterModal.show = false"></button>
            </div>
            <div class="modal-body">
              <div
                v-if="originalFilter"
                class="mb-3"
              >
                <label class="form-label text-primary">{{ $t('chart.filter.modal.originalFilter.label') }}</label>
                <textarea
                  :value="originalFilter"
                  class="form-control"
                  readonly
                />
              </div>

              <div class="mb-3">
                <label class="form-label text-primary">{{ $t('chart.filter.modal.liveFilter.label') }}</label>
                <c-input-select
                  v-model="liveFilterModal.value"
                  :options="predefinedFilters"
                  label="text"
                  :reduce="filter => filter.value"
                  :placeholder="$t('chart.filter.modal.liveFilter.placeholder')"
                />
              </div>

              <div v-if="originalFilter && liveFilterModal.value">
                <hr class="my-3" />

                <div class="mb-3">
                  <label class="form-label text-primary">{{ $t('chart.filter.modal.filterPreview.label') }}</label>
                  <textarea
                    :value="liveFilterPreview"
                    class="form-control"
                    readonly
                  />
                </div>

                <div class="mb-3">
                  <label class="form-label text-primary pb-0 pr-2">{{ $t('chart.filter.modal.options.label') }}</label>
                  <div class="btn-group" data-bs-toggle="buttons">
                    <label
                      v-for="(option, ci) in filterOptions"
                      :key="ci"
                      class="btn btn-outline-secondary"
                      :class="{ active: liveFilterModal.option === option.value }"
                    >
                      <input
                        v-model="liveFilterModal.option"
                        type="radio"
                        :value="option.value"
                        class="btn-check"
                      />
                      {{ $t(`chart.filter.modal.options.${option.label}`) }}
                    </label>
                  </div>
                </div>
              </div>
            </div>
            <div class="modal-footer">
              <div class="d-flex justify-content-between align-items-center w-100">
                <button
                  type="button"
                  class="btn btn-outline-secondary"
                  @click="resetLiveFilter()"
                >
                  {{ $t('label.reset') }}
                </button>

                <div class="d-flex gap-1">
                  <button
                    class="btn btn-outline-secondary"
                    type="button"
                    @click="cancelLiveFilter()"
                  >
                    {{ $t('label.cancel') }}
                  </button>
                  <button
                    class="btn btn-primary"
                    @click="updateLiveFilter"
                  >
                    {{ $t('label.saveAndClose') }}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div
        v-if="liveFilterModal.show"
        class="modal-backdrop fade show"
        @click="liveFilterModal.show = false"
      />
    </template>
  </Wrap>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount, inject } from 'vue'
import { useStore } from '../../store'
import { usePageBlockBase } from './usePageBlockBase'
import Wrap from './Wrap/index.js'
import BlockHelpButton from './Shared/BlockHelpButton.vue'
import ChartComponent from '../Chart'
import { NoID, compose } from 'corteza-lib/js/dist'
import { debounce } from 'lodash'
import { evalPrefilterOrSkip, isFieldInFilter } from 'corteza-webapp-compose/src/lib/record-filter'
import { hydrateChartDocs } from '../../help/appDocs'

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
const store = useStore()
const $auth = inject('$auth')
const $ComposeAPI = inject('$ComposeAPI')

const { key, options, inModal, browserLocale, refreshBlock, setBaseDefaultValues } = usePageBlockBase(props, emit)

// ${variables.x} in a chart's filter, sourced from the page's session-only
// variable values (see PageBlocks/Variables).
const pageVariables = computed(() => store.pageVariables.getValuesForPage(props.page.pageID))

const chart = ref(null)

const chartDescription = computed(() => String(chart.value?.config?.description || '').trim())
const chartHelpBody = computed(() => String(chart.value?.config?.help || '').trim())
const chartData = ref(null)
const originalFilter = ref(undefined)
const filter = ref(undefined)
const drillDownFilter = ref(undefined)
const liveFilterModal = ref({ show: false, value: undefined, option: 'AND' })
const predefinedFilters = ref([])
const selectedFilter = ref(undefined)
const customDate = ref({ start: undefined, end: undefined })
const liveFilterValue = ref(undefined)
const liveFilterOption = ref('AND')
const filterOptions = ref([
  { label: 'and', value: 'AND' },
  { label: 'or', value: 'OR' },
  { label: 'overwrite', value: '' },
])

const isDrillDownEnabled = computed(() => {
  if (!options.value) return false
  return options.value.drillDown && options.value.drillDown.enabled
})

const liveFilterPreview = computed(() => getFilter(liveFilterModal.value.value, liveFilterModal.value.option))

const hasSaveChartEnabled = computed(() => {
  const { config = {} } = chart.value || {}
  if (!config.toolbox) return false
  return config.toolbox.saveAsImage
})

const hasDataTableEnabled = computed(() => {
  const { config = {} } = chart.value || {}
  if (!config.toolbox) return false
  return config.toolbox.showDataTable
})

const hasLiveFilter = computed(() => !!liveFilterValue.value)

watch(() => options.value, debounce(() => { if (!props.loadingRecord) refresh() }, 300), { deep: true })
watch(() => [props.record?.recordID, props.loadingRecord], () => { if (!props.loadingRecord) refresh() })

onMounted(() => {
  fetchChart()
  refreshBlock(refresh)
  createEvents()
})

onBeforeUnmount(() => {
  destroyEvents()
  setDefaultValues()
})

function createEvents () {
  window.addEventListener('drill-down-chart', drillDown)
  window.addEventListener('module-records-updated', refreshOnRelatedRecordsUpdate)
  window.addEventListener('record-field-change', refetchOnPrefilterValueChange)
  window.addEventListener('page-variable-change', refetchOnPageVariableChange)
  window.addEventListener('refetch-records', refresh)
}

function refetchOnPrefilterValueChange ({ fieldName }) {
  const { filter: f } = filter.value || {}
  if (isFieldInFilter(fieldName, f)) refresh()
}

function refetchOnPageVariableChange ({ detail: { pageID, fieldName } } = {}) {
  if (pageID !== props.page.pageID) return
  const { filter: f } = filter.value || {}
  if (isFieldInFilter(`variables.${fieldName}`, f)) refresh()
}

function refreshOnRelatedRecordsUpdate ({ moduleID } = {}) {
  if (filter.value?.moduleID === moduleID) refresh()
}

async function fetchChart (params = {}) {
  const { chartID } = options.value
  if (!chartID) return
  const { namespaceID } = props.namespace
  return store.chart.findByID({ chartID, namespaceID, ...params }).then((c) => {
    chart.value = hydrateChartDocs(props.namespace, c) || c
    if (isDrillDownEnabled.value) {
      const { moduleID, dimensions = [] } = c.config.reports[0] || {}
      store.module.findByID({ namespace: props.namespace, moduleID }).then(chartModule => {
        if (!chartModule) return
        const { field } = dimensions[0] || {}
        const { name, label } = chartModule.fields.find(({ name }) => name === field) || {}
        filter.value = filter.value || {}
        filter.value.field = { name, label }
      })
    }
  })
}

function reporter (r = {}) {
  if (!originalFilter.value) {
    originalFilter.value = r.filter
    filter.value = r
  }
  let f = getFilter()
  if (f) {
    const { skip, filter: evaluated } = evalPrefilterOrSkip(f, {
      record: props.record,
      user: $auth.user || {},
      recordID: (props.record || {}).recordID || NoID,
      ownerID: (props.record || {}).ownedBy || NoID,
      userID: ($auth.user || {}).userID || NoID,
      loadingRecord: !!props.loadingRecord,
      variables: pageVariables.value,
    })
    if (skip) return Promise.resolve([])
    f = evaluated
    filter.value = filter.value || {}
    filter.value.filter = f
  }
  const { namespaceID } = props.namespace
  return $ComposeAPI.recordReport({ namespaceID, ...r, filter: f }).then(data => {
    chartData.value = data
    return data
  })
}

function getFilter (liveFilter = liveFilterValue.value, option = liveFilterOption.value) {
  if (liveFilter) return originalFilter.value && option ? `(${originalFilter.value}) ${option} (${liveFilter})` : liveFilter
  return originalFilter.value
}

function updateLiveFilter () {
  liveFilterValue.value = liveFilterModal.value.value
  liveFilterOption.value = liveFilterModal.value.option
  liveFilterModal.value.show = false
  updateChart()
}

function resetLiveFilter () {
  liveFilterModal.value.value = undefined
  liveFilterModal.value.option = 'AND'
}

function cancelLiveFilter () {
  liveFilterModal.value.show = false
}

function refresh () {
  fetchChart({ force: true }).then(() => { updateChart() })
}

function updateChart () {
  if (chart.value) { chart.value.config.noAnimation = true; key.value++ }
}

function drillDown ({ trueName, value }) {
  const { chartID, drillDown: dd } = options.value
  if (!dd.enabled) return
  const report = chart.value.config.reports[0] || {}
  const { yAxis = {} } = report
  let drillDownValue = trueName
  if (!trueName) drillDownValue = yAxis.horizontal ? value[1] : value[0]
  let { moduleID, dimensions, filter: f, field } = filter.value
  const { name, label } = field || {}
  const dimensionFilter = dimensions ? `(${dimensions} = '${drillDownValue}')` : ''
  if (dd.blockID) {
    const { pageID = NoID } = props.page
    const { recordID = NoID } = props.record || {}
    const recordListUniqueID = [pageID, recordID, dd.blockID, false].map(v => v || NoID).join('-')
    window.dispatchEvent(new CustomEvent(`drill-down-recordList:${recordListUniqueID}`, { detail: { prefilter: dimensionFilter, name: name || label || dimensions, value: drillDownValue } }))
  } else {
    f = f ? `(${f})` : ''
    const prefilter = [dimensionFilter, f].filter(f => f).join(' AND ')
    const { title: t } = props.block
    const { fields = [] } = dd.recordListOptions || {}
    const block = new compose.PageBlockRecordList({
      title: t ? `${t} - "${drillDownValue}"` : drillDownValue,
      blockID: `drillDown-${chartID}`,
      options: { moduleID, fields, prefilter, presort: '', hideRecordReminderButton: true, hideRecordViewButton: false, hideConfigureFieldsButton: false, hideImportButton: true, enableRecordPageNavigation: true, selectable: true, allowExport: true, perPage: 14, showTotalCount: true, recordDisplayOption: 'modal' },
    })
    window.dispatchEvent(new CustomEvent('magnify-page-block', { detail: { block } }))
  }
}

function showFilterModal () {
  liveFilterModal.value.value = liveFilterValue.value
  liveFilterModal.value.option = liveFilterOption.value
  liveFilterModal.value.show = true
}

function promptAiChat () {
  const page = props.page
  const block = props.block
  const namespace = props.namespace
  const locale = browserLocale()
  let prompt = block.prompt || page.config.prompt || namespace.prompt || ''
  if (prompt.length === 0) {
    switch (locale) {
      case 'en-US': prompt = 'What does this chart show? '; break
      case 'ru-RU': prompt = 'Что показывает этот график? О чём он говорит? '; break
    }
    prompt += '\r\n '
  }
  prompt += '\r\n*' + page.title + '*\r\n*' + block.title + '*\r\n'
  if (chart.value) {
    const report = chart.value.config.reports[0] || {}
    if (report.filter) prompt += 'Фильтр: ' + report.filter + '\r\n'
  }
  const files = []
  if (chartData.value && chartData.value.length > 0) {
    const lastRows = chartData.value.slice(-25)
    let headers = Object.keys(lastRows[0])
    const csv = toCSV(lastRows, headers.map(h => ({ key: h, label: h })))
    files.push({ name: 'chart_data.csv', content: csv, type: 'text/csv' })
  }
  window.dispatchEvent(new CustomEvent('show-chat-modal', { detail: { namespace: page.namespaceID, module: page.moduleID, prompt, files } }))
}

function toCSV (rows, headers) {
  const esc = (v) => { const s = v === null || v === undefined ? '' : String(v); return '"' + s.replace(/"/g, '""') + '"' }
  const head = headers.map(h => esc(h.label)).join(',')
  const body = rows.map(r => headers.map(h => esc(r[h.key])).join(',')).join('\r\n')
  return head + '\r\n' + body + '\r\n'
}

function setDefaultValues () {
  chart.value = null
  chartData.value = null
  filter.value = undefined
  drillDownFilter.value = undefined
}

function destroyEvents () {
  window.removeEventListener('drill-down-chart', drillDown)
  window.removeEventListener('module-records-updated', refreshOnRelatedRecordsUpdate)
  window.removeEventListener('record-field-change', refetchOnPrefilterValueChange)
  window.removeEventListener('page-variable-change', refetchOnPageVariableChange)
  window.removeEventListener('refetch-records', refresh)
}
</script>

<style lang="scss" scoped>
.chart-filter-button {
  right: 0.5rem;
  top: 2rem;
  z-index: 1;
  &.save-chart-enabled { right: 2.2rem; }
  &.table-enabled { right: 3.9rem; }
  &.save-chart-enabled.table-enabled { right: 3.9rem; }
}

.chart-brain-button {
  right: 0.5rem;
  top: 0.7rem;
  z-index: 1;
}
</style>
