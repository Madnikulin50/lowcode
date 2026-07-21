<template>
  <div class="d-flex overflow-auto p-2 w-100">
    <Teleport v-if="!fetchingReport" to="#topbar-title">{{ pageTitle }}</Teleport>

    <Teleport to="#topbar-tools">
      <c-input-select
        v-if="scenarioOptions.length"
        v-model="scenariosState.selected"
        :options="scenarioOptions"
        :get-option-key="getOptionKey"
        :placeholder="t('pick-scenario')"
        :disabled="processing || fetchingReport"
        size="sm"
        style="max-width: 300px; min-width: 150px;"
        @input="refreshReport()"
      />

      <div v-if="canUpdate" class="btn-group btn-group-sm">
        <button
          class="btn btn-primary d-flex align-items-center justify-content-center"
          @click="$router.push(reportBuilder)"
        >
          {{ t('report.builder') }}
          <font-awesome-icon class="ms-2" :icon="['fas', 'tools']" />
        </button>
        <button
          class="btn btn-primary d-flex align-items-center justify-content-center"
          style="margin-left:2px"
          :title="t('tooltip.edit.report')"
          @click="$router.push(reportEditor)"
        >
          <font-awesome-icon :icon="['far', 'edit']" />
        </button>
      </div>
    </Teleport>

    <div v-if="fetchingReport" class="d-flex align-items-center justify-content-center w-100 h-100">
      <div class="spinner-border" />
    </div>

    <grid
      v-if="report && canRead && showReport && !fetchingReport"
      :blocks="report.blocks"
    >
      <template #default="{ block, index }">
        <block
          :index="index"
          :block="block"
          :scenario="currentSelectedScenario"
          :report-i-d="reportID"
        />
      </template>
    </grid>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useToast } from 'corteza-lib/vue/dist'
import { system } from 'corteza-lib/js/dist'
import Grid from '../../components/Report/Grid.vue'
import Block from '../../components/Report/Blocks/index.vue'

const { t } = useI18n()
const route = useRoute()
const toast = useToast()
const toastErrorHandler = toast.toastErrorHandler

const processing = ref(false)
const showReport = ref(true)
const fetchingReport = ref(false)
const report = ref(undefined)
const dataframes = ref([])

const scenariosState = ref({ selected: undefined })

const reportID = computed(() => route.params.reportID)

const pageTitle = computed(() => {
  const title = report.value ? (report.value.meta.name || report.value.handle) : ''
  return title || t('report.view')
})

const canRead = computed(() => report.value ? report.value.canReadReport : false)
const canUpdate = computed(() => report.value ? report.value.canUpdateReport : false)

const reportBuilder = computed(() => report.value ? { name: 'report.builder', params: { reportID: report.value.reportID } } : undefined)
const reportEditor = computed(() => report.value ? { name: 'report.edit', params: { reportID: report.value.reportID } } : undefined)

const reportScenarios = computed(() => report.value ? report.value.scenarios : [])
const scenarioOptions = computed(() => report.value ? reportScenarios.value.map(({ label }) => label) : [])

const currentSelectedScenario = computed(() => {
  return scenariosState.value.selected ? reportScenarios.value.find(({ label }) => label === scenariosState.value.selected) : undefined
})

function refreshReport() {
  showReport.value = false
  setTimeout(() => { showReport.value = true }, 50)
}

function getOptionKey(scenario) { return scenario }

watch(reportID, (id) => {
  scenariosState.value.selected = undefined
  report.value = undefined

  if (id) {
    fetchingReport.value = true
    processing.value = true
    window.__systemAPI.reportRead({ reportID: id })
      .then(r => {
        report.value = new system.Report(r)
        report.value.blocks = report.value.blocks.map(({ xywh, ...p }, i) => {
          const [x, y, w, h] = xywh
          return { ...p, x, y, w, h, i }
        })
      })
      .catch(toastErrorHandler(t('notification.report.loadFailed')))
      .finally(() => {
        setTimeout(() => {
          fetchingReport.value = false
          processing.value = false
        }, 400)
      })
  }
}, { immediate: true })
</script>
