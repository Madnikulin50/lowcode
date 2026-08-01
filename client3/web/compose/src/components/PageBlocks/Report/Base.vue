<template>
  <Wrap
    v-bind="$props"
    @refreshBlock="refresh"
  >
    <div
      v-if="isProcessing"
      class="d-flex align-items-center justify-content-center h-100"
    >
      <span class="spinner-border" />
    </div>

    <DisplayElement
      v-else-if="displayElement"
      :key="key"
      :display-element="displayElement"
      :labels="{
        previous: $t('recordList.pagination.prev'),
        next: $t('recordList.pagination.next'),
      }"
      @update="getDataframes"
    />
  </Wrap>
</template>

<script setup>
import { ref, watch, onBeforeUnmount, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { composables } from 'corteza-lib/vue/dist'
import { system, reporter, NoID } from 'corteza-lib/js/dist'
import axios from 'axios'
import { usePageBlockBase } from '../usePageBlockBase'
import Wrap from '../Wrap/index.js'
import DisplayElement from './DisplayElements/index.js'

const { t: $t } = useI18n({ useScope: 'global' })
const { toastErrorHandler } = composables.useToast()
const $SystemAPI = window.__systemAPI

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

const emit = defineEmits(['errors', 'refreshBlock'])

const { processing, key, isProcessing, options, refreshBlock } = usePageBlockBase(props, emit)

const report = ref(undefined)
const displayElement = ref(undefined)
const abortableRequests = ref([])

watch(() => options.value.reportID, (reportID = NoID) => {
  if (reportID !== NoID) fetchReport(reportID)
}, { deep: true, immediate: true })

watch(() => props.record?.recordID, () => { refresh() }, { immediate: true })

onMounted(() => {
  refreshBlock(refresh)
  window.addEventListener('refetch-records', refresh)
})

onBeforeUnmount(() => {
  window.removeEventListener('refetch-records', refresh)
  abortableRequests.value.forEach(c => c())
})

function fetchReport(reportID) {
  processing.value = true
  const { response, cancel } = $SystemAPI.reportReadCancellable({ reportID })
  abortableRequests.value.push(cancel)
  return response()
    .then(r => {
      report.value = new system.Report(r)
      return getDataframes()
    })
    .catch(e => {
      if (!axios.isCancel(e)) {
        toastErrorHandler($t('notification.report.fetchFailed'))(e)
      }
    })
    .finally(() => { processing.value = false })
}

function getDataframes(definition = {}) {
  const { elementID } = options.value
  if (elementID) {
    const block = report.value.blocks.find(({ elements }) => elements.some(e => e.elementID === elementID))
    let element = (block?.elements || []).find(e => e.elementID === elementID)
    if (element && element.kind !== 'Text') {
      element = reporter.DisplayElementMaker(element)
      const scenarioDefinition = getScenarioDefinition(element)
      Object.entries(definition).forEach(([key, value]) => {
        definition[key] = { ...value, ...scenarioDefinition[key] }
      })
      const { dataframes: frames = [] } = element.reportDefinitions({ ...definition, ...scenarioDefinition })
      if (frames.length) {
        return $SystemAPI.reportRun({ frames, reportID: options.value.reportID })
          .then(({ frames: dataframes = [] }) => {
            displayElement.value = { ...element, dataframes }
          })
          .catch(e => toastErrorHandler($t('notification.report.runFailed'))(e))
      }
    } else {
      displayElement.value = element
    }
  }
}

function getScenarioDefinition(element) {
  const scenarioDefinition = {}
  const { scenarioID } = options.value
  const scenario = report.value?.scenarios?.find(({ label }) => scenarioID === label)
  if (scenario?.filters) {
    Object.keys(scenario.filters).forEach(k => {
      scenarioDefinition[k] = { ref: k, filter: { ...scenario.filters[k] } }
    })
  }
  return scenarioDefinition
}

function refresh() {
  fetchReport(options.value.reportID).then(() => { key.value++ })
}
</script>
