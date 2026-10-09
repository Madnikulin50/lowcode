<template>
  <div :class="hideTrigger ? '' : 'd-flex'" :style="hideTrigger ? { display: 'contents' } : undefined">
    <button
      v-if="!hideTrigger"
      class="btn btn-outline-secondary flex-fill"
      @click="toggleModal"
    >
      {{ $t('label.export') }}
    </button>

    <div
      v-if="showExportModal"
      class="modal d-block"
      tabindex="-1"
      role="dialog"
      @click.self="toggleModal"
    >
      <div class="modal-dialog modal-lg modal-dialog-scrollable" role="document">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ $t('label.export') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" @click="toggleModal"></button>
          </div>
          <div class="modal-body">
            <template v-if="showExportModal">
              <div class="mb-3">
                <label class="form-label text-primary">{{ $t('recordList.export.selectFields') }}</label>
                <small class="text-muted d-block mb-2">{{ $t('recordList.export.limitations') }}</small>
                <FieldPicker
                  v-if="module"
                  :module="module"
                  :system-fields="systemFields"
                  :disabled-types="disabledTypes"
                  v-model:fields="selectedFields"
                  style="height: 45vh;"
                />
              </div>

              <div class="mb-3">
                <div class="btn-group-vertical" role="group" data-bs-toggle="buttons">
                  <label
                    v-for="opt in rangeTypeOptions"
                    :key="opt.value"
                    class="btn btn-outline-primary"
                    :class="{ active: rangeType === opt.value }"
                  >
                    <input
                      type="radio"
                      :value="opt.value"
                      v-model="rangeType"
                      @change="getTotalCount()"
                    />
                    {{ opt.text }}
                  </label>
                </div>
              </div>

              <template v-if="rangeType === 'range'">
                <div class="row">
                  <div class="col-lg-6">
                    <div class="mb-3">
                      <label class="form-label text-primary">{{ $t('recordList.export.rangeBy') }}</label>
                      <select
                        v-model="rangeBy"
                        class="form-select form-control"
                        @change="getTotalCount()"
                      >
                        <option
                          v-for="opt in rangeByOptions"
                          :key="opt.value"
                          :value="opt.value"
                        >
                          {{ opt.text }}
                        </option>
                      </select>
                    </div>
                  </div>
                  <div class="col-lg-6">
                    <div class="mb-3">
                      <label class="form-label text-primary">{{ $t('recordList.export.dateRange') }}</label>
                      <select
                        v-model="range"
                        class="form-select form-control"
                        @change="getTotalCount()"
                      >
                        <option
                          v-for="opt in dateRangeOptions"
                          :key="opt.value"
                          :value="opt.value"
                        >
                          {{ opt.text }}
                        </option>
                      </select>
                    </div>
                  </div>
                </div>
              </template>

              <div v-if="rangeType === 'range'" class="row">
                <div class="col-lg-6">
                  <div class="mb-3">
                    <label class="form-label text-primary">{{ $t('recordList.export.filter.from') }}</label>
                    <c-input-date-time
                      v-model="start"
                      no-time
                      only-past
                      :labels="dateLabels"
                    />
                  </div>
                </div>
                <div class="col-lg-6">
                  <div class="mb-3">
                    <label class="form-label text-primary">{{ $t('recordList.export.filter.to') }}</label>
                    <c-input-date-time
                      v-model="end"
                      no-time
                      only-past
                      :labels="dateLabels"
                    />
                  </div>
                </div>
              </div>

              <template v-if="rangeType !== 'selection'">
                <div class="mb-3">
                  <label class="form-label text-primary">{{ $t('recordList.export.filter.label') }}</label>
                  <textarea
                    v-model="exportFilter"
                    class="form-control"
                    :placeholder="$t('recordList.export.filter.placeholder')"
                  />
                </div>
              </template>

              <div class="mb-3">
                <div class="form-check mb-2">
                  <input
                    id="timezone-check"
                    v-model="forTimezone"
                    class="form-check-input"
                    type="checkbox"
                  />
                  <label class="form-check-label" for="timezone-check">
                    {{ $t('recordList.export.specifyTimezone') }}
                  </label>
                </div>
                <c-input-select
                  v-if="forTimezone"
                  v-model="exportTimezone"
                  :options="timezones"
                  :get-option-key="getOptionKey"
                  :placeholder="$t('recordList.export.timezonePlaceholder')"
                />
              </div>

              <div v-if="isMultiFieldSelected" class="mb-3">
                <label class="form-label text-primary">{{ $t('recordList.export.multiValueDelimiter.label') }}</label>
                <select
                  v-model="multiValueDelimiter"
                  class="form-select form-control"
                >
                  <option
                    v-for="opt in multiValueDelimiterOptions"
                    :key="opt.value"
                    :value="opt.value"
                  >
                    {{ opt.text }}
                  </option>
                </select>
              </div>

              <div class="mb-3">
                <small class="text-muted d-block mb-1">{{ $t('recordList.export.resolveRefsNote') }}</small>
                <div class="form-check mb-2">
                  <input
                    id="resolve-refs-check"
                    v-model="resolveRefs"
                    class="form-check-input"
                    type="checkbox"
                  />
                  <label class="form-check-label" for="resolve-refs-check">
                    {{ $t('recordList.export.resolveRefs') }}
                  </label>
                </div>
              </div>
            </template>
          </div>
          <div class="modal-footer d-flex align-items-center justify-content-between">
            <div>
              <span v-if="processingCount" class="spinner-border spinner-border-sm" />
              <span v-else>
                {{ $t('recordList.export.recordCount', { count: getExportableCount || 0 }) }}
              </span>
            </div>
            <div>
              <c-input-processing
                v-if="allowJSON"
                :processing="processing"
                :disabled="exportDisabled"
                variant="outline-secondary"
                @click="doExport('json')"
              >
                {{ $t('recordList.export.json') }}
              </c-input-processing>
              <c-input-processing
                v-if="allowCSV"
                :processing="processing"
                :disabled="exportDisabled"
                variant="outline-secondary"
                class="ms-2"
                @click="doExport('csv')"
              >
                {{ $t('recordList.export.csv') }}
              </c-input-processing>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'
import moment from 'moment'
import { throttle } from 'lodash'
import tz from 'compact-timezone-list'
import FieldPicker from 'corteza-webapp-compose/src/components/Common/FieldPicker'
import { getFieldFilter } from 'corteza-webapp-compose/src/lib/record-filter'

const { CInputDateTime } = components
const { t: $t } = useI18n({ useScope: 'global' })
const $ComposeAPI = window.__composeAPI

const fmtDate = (d) => d.format('YYYY-MM-DD')

const props = defineProps({
  allowJSON: { type: Boolean, default: true },
  allowCSV: { type: Boolean, default: true },
  module: { type: compose.Module, required: true },
  preselectedFields: { type: Array, default: () => [] },
  filter: { type: String, required: false, default: '' },
  selection: { type: Array, required: false, default: () => [] },
  selectedAllRecords: { type: Boolean, default: false },
  filterRangeType: { type: String, default: 'all' },
  filterRangeBy: { type: String, default: 'createdAt' },
  startDate: { type: String, default: null },
  endDate: { type: String, default: null },
  systemFields: { type: Array, default: () => ['ownedBy', 'createdAt', 'createdBy', 'updatedAt', 'updatedBy'] },
  disabledTypes: { type: Array, default: () => ['File'] },
  processing: { type: Boolean, default: false },
  hideTrigger: { type: Boolean, default: false },
})

const emit = defineEmits(['export'])

const showExportModal = ref(false)

const fields = ref([])
const resolveRefs = ref(false)
const forTimezone = ref(false)
const exportTimezone = ref(undefined)
const exportConfig = ref({
  rangeType: 'all',
  query: '',
  filter: '',
  multiValueDelimiter: ';',
  rangeBy: null,
  date: {
    range: 'lastMonth',
    start: fmtDate(moment()),
    end: fmtDate(moment()),
  },
})
const rangeRecordCount = ref(0)
const processingCount = ref(false)

const dateLabels = {
  clear: $t('label.clear'),
  none: $t('label.none'),
  now: $t('label.now'),
  today: $t('label.today'),
}

const timezones = computed(() => {
  return tz.map(({ label, tzCode, offset }) => ({ label, tzCode, offset }))
})

const rangeTypeOptions = computed(() => {
  const options = [
    { value: 'all', text: $t('recordList.export.all') },
    { value: 'range', text: $t('recordList.export.inRange') },
  ]
  if (hasSelection.value) {
    options.push({ value: 'selection', text: $t('recordList.export.selection') })
  }
  return options
})

const dateRangeValid = computed(() => {
  if (start.value < end.value) return false
  return true
})

const rangeByOptions = computed(() => [
  { value: 'createdAt', text: $t('recordList.export.filter.createdAt') },
  { value: 'updatedAt', text: $t('recordList.export.filter.updatedAt') },
])

const multiValueDelimiterOptions = computed(() => [
  { value: ';', text: $t('recordList.export.multiValueDelimiter.semicolon.label') },
  { value: ',', text: $t('recordList.export.multiValueDelimiter.comma.label') },
  { value: '|', text: $t('recordList.export.multiValueDelimiter.pipe.label') },
  { value: '[;]', text: $t('recordList.export.multiValueDelimiter.semicolonArray.label') },
  { value: '[,]', text: $t('recordList.export.multiValueDelimiter.commaArray.label') },
  { value: '[|]', text: $t('recordList.export.multiValueDelimiter.pipeArray.label') },
])

const dateRangeOptions = computed(() => [
  { value: 'lastMonth', text: $t('recordList.export.filter.lastMonth') },
  { value: 'thisMonth', text: $t('recordList.export.filter.thisMonth') },
  { value: 'lastWeek', text: $t('recordList.export.filter.lastWeek') },
  { value: 'thisWeek', text: $t('recordList.export.filter.thisWeek') },
  { value: 'today', text: $t('recordList.export.filter.today') },
  { value: 'custom', text: $t('recordList.export.filter.custom') },
])

const hasSelection = computed(() => !!props.selection.length)

const getExportableCount = computed(() => {
  if (exportConfig.value.rangeType === 'selection') {
    return props.selectedAllRecords ? rangeRecordCount.value : props.selection.length
  }
  return rangeRecordCount.value
})

const exportDisabled = computed(() => {
  return !dateRangeValid.value || fields.value.length === 0 || !getExportableCount.value
})

const selectedFields = computed({
  get: () => fields.value,
  set: (val) => { fields.value = val },
})

const rangeBy = computed({
  get: () => exportConfig.value.rangeBy,
  set: (val) => { exportConfig.value.rangeBy = val },
})

const multiValueDelimiter = computed({
  get: () => exportConfig.value.multiValueDelimiter,
  set: (val) => { exportConfig.value.multiValueDelimiter = val },
})

const exportFilter = computed({
  get: () => exportConfig.value.filter,
  set: (v) => {
    exportConfig.value.filter = v
    getTotalCount()
  },
})

const rangeType = computed({
  get: () => exportConfig.value.rangeType,
  set: (val) => { exportConfig.value.rangeType = val },
})

const range = computed({
  get: () => exportConfig.value.date.range,
  set: (val) => {
    exportConfig.value.date.range = val
    if (val !== 'custom') {
      const m = moment()
      exportConfig.value.date.start = calcStart(m, val)
      exportConfig.value.date.end = calcEnd(m, val)
    }
  },
})

const start = computed({
  get: () => exportConfig.value.date.start,
  set: (val) => {
    exportConfig.value.date.start = val
    exportConfig.value.date.range = 'custom'
    getTotalCount()
  },
})

const end = computed({
  get: () => exportConfig.value.date.end,
  set: (val) => {
    exportConfig.value.date.end = val
    exportConfig.value.date.range = 'custom'
    getTotalCount()
  },
})

const currentRange = computed(() => {
  if (exportConfig.value.rangeType === 'range') {
    return { start: start.value, end: end.value, rangeBy: rangeBy.value }
  }
  return undefined
})

const isMultiFieldSelected = computed(() => {
  return fields.value.some(f => f.isMulti)
})

watch(() => props.filter, (filter) => {
  exportConfig.value.filter = filter
}, { immediate: true, deep: true })

watch(() => props.preselectedFields, (value) => {
  if (!fields.value.length) {
    fields.value = value.filter(f => props.disabledTypes.indexOf(f.kind) < 0)
  }
}, { immediate: true })

watch(() => props.filterRangeType, (value) => {
  exportConfig.value.rangeType = value
}, { immediate: true })

watch(() => props.filterRangeBy, (value) => {
  exportConfig.value.rangeBy = value
}, { immediate: true })

watch(() => props.startDate, (value) => {
  start.value = value
})

watch(() => props.endDate, (value) => {
  end.value = value
})

watch(hasSelection, (h) => {
  if (h) {
    exportConfig.value.rangeType = 'selection'
  }
}, { immediate: true })

function toggleModal() {
  showExportModal.value = !showExportModal.value
  if (showExportModal.value) {
    getTotalCount()
  }
}

function open() {
  showExportModal.value = true
  getTotalCount()
}

defineExpose({ open })

function calcStart(m, range) {
  if (range === 'lastMonth') return fmtDate(m.subtract(1, 'months').startOf('month'))
  else if (range === 'thisMonth') return fmtDate(m.startOf('month'))
  else if (range === 'lastWeek') return fmtDate(m.subtract(1, 'week').startOf('week'))
  else if (range === 'thisWeek') return fmtDate(m.startOf('week'))
  else if (range === 'today') return fmtDate(m.startOf('day'))
  else throw new Error($t('recordList.export.datePresetUndefined'))
}

function calcEnd(m, range) {
  if (range === 'lastMonth') return fmtDate(m.subtract(1, 'months').endOf('month'))
  else if (range === 'thisMonth') return fmtDate(m.endOf('month'))
  else if (range === 'lastWeek') return fmtDate(m.subtract(1, 'week').endOf('week'))
  else if (range === 'thisWeek') return fmtDate(m.endOf('week'))
  else if (range === 'today') return fmtDate(m.endOf('day'))
  else throw new Error($t('recordList.export.datePresetUndefined'))
}

function makeFilter({ filter, rangeType, rangeBy, date }) {
  if (rangeType === 'all') return filter
  if (rangeType === 'selection') return props.selectedAllRecords ? filter : props.selection.map(r => `recordID='${r}'`).join(' OR ')

  let dateRangeQuery = ''
  if (date.start && date.end) {
    date = { ...date }
    if (date.start === date.end) {
      date.start = moment(date.start, 'YYYY-MM-DD HH:mm').utc().format()
      date.end = moment(date.end, 'YYYY-MM-DD HH:mm').add(1, 'days').utc().format()
    }
    dateRangeQuery = getFieldFilter(rangeBy, 'DateTime', date, 'BETWEEN')
  } else if (date.start) {
    dateRangeQuery = getFieldFilter(rangeBy, 'DateTime', date.start, '>=')
  } else if (date.end) {
    dateRangeQuery = getFieldFilter(rangeBy, 'DateTime', date.end, '<=')
  }
  return filter && dateRangeQuery ? `(${filter}) AND ${dateRangeQuery}` : dateRangeQuery
}

function doExport(kind) {
  emit('export', {
    ext: kind,
    fields: encodeURIComponent(fields.value.map(({ name }) => name)),
    filter: encodeURIComponent(makeFilter(exportConfig.value)),
    filterRaw: encodeURIComponent(exportConfig.value),
    multiValueDelimiter: encodeURIComponent(exportConfig.value.multiValueDelimiter),
    timezone: encodeURIComponent(forTimezone.value ? exportTimezone.value : undefined),
    resolveRefs: encodeURIComponent(resolveRefs.value),
  })
}

const getTotalCount = throttle(function () {
  const { moduleID, namespaceID } = props.module || {}
  if (moduleID && namespaceID) {
    const query = makeFilter(exportConfig.value)
    processingCount.value = true
    $ComposeAPI.recordList({ namespaceID, moduleID, query, limit: 1, incTotal: true })
      .then(({ filter = {} }) => {
        rangeRecordCount.value = filter.total || 0
      })
      .catch(() => {
        rangeRecordCount.value = 0
      })
      .finally(() => {
        processingCount.value = false
      })
  }
}, 500)

function getOptionKey({ tzCode }) {
  return tzCode
}
</script>
