<template>
  <div>
    <div class="row">
      <div class="col">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('report.label') }}</label>
          <c-input-select
            v-model="options.reportID"
            :options="reports"
            :get-option-label="getReportLabel"
            default-value="0"
            :reduce="o => o.reportID"
            @input="handleReportChange"
          />
        </div>
      </div>
      <div v-if="selectedReport?.scenarios?.length > 1" class="col">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('report.scenario.label') }}</label>
          <c-input-select
            v-model="options.scenarioID"
            :options="selectedReport.scenarios"
            default-value="0"
            :reduce="o => o.scenarioID"
          />
        </div>
      </div>
    </div>
    <div v-if="selectedReport" class="mb-3">
      <label class="form-label text-primary">{{ $t('report.element.label') }}</label>
      <c-input-select
        v-model="options.elementID"
        :options="allElements"
        :reduce="o => o.elementID"
        default-value="0"
      >
        <template #selected-option="option">
          {{ option.label }} <small class="text-muted">({{ option.blockLabel }})</small>
        </template>
        <template #option="option">
          {{ option.label }} <small class="text-muted">({{ option.blockLabel }})</small>
        </template>
      </c-input-select>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onBeforeUnmount, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { composables } from 'corteza-lib/vue/dist'
import { NoID } from 'corteza-lib/js/dist'

const { t: $t } = useI18n({ useScope: 'global' })
const { toastErrorHandler } = composables.useToast()
const $SystemAPI = window.__systemAPI

const props = defineProps({
  block: { type: Object, required: true },
})

const reports = ref([])
const options = computed(() => props.block.options)

const selectedReport = computed(() => {
  const { reportID = NoID } = options.value
  if (reportID !== NoID) return reports.value.find(r => r.reportID === reportID)
  return undefined
})

const allElements = computed(() => {
  if (!selectedReport.value) return []
  const elements = []
  selectedReport.value.blocks.forEach(block => {
    const blockLabel = block.title || `${$t('label.block')} ${block.key}`
    if (block.elements && Array.isArray(block.elements)) {
      block.elements.forEach(element => {
        elements.push({ elementID: element.elementID, blockLabel, label: element.name || element.kind })
      })
    }
  })
  return elements
})

onMounted(() => { fetchReports() })
onBeforeUnmount(() => { reports.value = [] })

function fetchReports() {
  $SystemAPI.reportList()
    .then(({ set = [] }) => { reports.value = set })
    .catch(toastErrorHandler($t('notification.report.listFetchFailed')))
}

function getReportLabel({ kind, meta = {} } = {}) {
  return meta.name || kind
}

function handleReportChange() {
  if (options.value.elementID) {
    options.value.elementID = NoID
    options.value.scenarioID = NoID
  }
}
</script>
