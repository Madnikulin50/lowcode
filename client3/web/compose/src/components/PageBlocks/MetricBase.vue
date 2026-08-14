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
      <div
        class="fixed-corner-container"
        :class="fieldLayoutClass"
      >
        <div
          v-for="(m, mi) in options.metrics"
          :key="mi"
          class="d-flex align-items-center justify-content-center overflow-hidden"
          :class="{
            'h-100': options.likeRecordList !== true && m.valueStyle.notFitVertical !== true,
            'px-3': options.likeRecordList === true,
            'pt-3': options.likeRecordList === true && mi === 0
          }"
        >
        <div
          v-for="(v, i) in formatResponse(m, mi)"
          :key="i"
          class="py-1"
          :class="{
            'px-2': options.likeRecordList !== true,
            'pointer': m.drillDown.enabled,
            'w-100': options.likeRecordList === true || m.valueStyle.notFitHorizontal !== true,
            'h-100': (options.likeRecordList !== true && m.valueStyle.notFitVertical !== true)
          }"
          @click="drillDown(m, mi)"
        >
          <metric-item
            :metric="m"
            :options="options"
            :hover="m.drillDown.enabled"
            :theme-settings="themeSettings"
            :value="v"
          />
        </div>
        </div>
      </div>
    </template>
  </Wrap>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed, watch, onMounted, onBeforeUnmount, inject, nextTick } from 'vue'
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

const { options, isProcessing, processing, browserLocale, themeSettings, refreshBlock, setBaseDefaultValues } = usePageBlockBase(props, emit)

const fieldLayoutClass = computed(() => {
  const classes = { default: 'd-flex flex-column', noWrap: 'd-flex gap-2', wrap: 'row g-0' }
  return classes[options.value.recordFieldLayoutOption]
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
  if (!drillDown.enabled) return
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
</style>
