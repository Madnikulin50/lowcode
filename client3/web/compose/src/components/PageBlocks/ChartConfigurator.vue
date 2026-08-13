<template>
  <div class="tab-pane">
    <div class="mb-3">
      <label class="form-label text-primary">{{ $t('chart.display') }}</label>
      <div class="input-group d-flex w-100">
        <c-input-select
          v-model="block.options.chartID"
          :options="charts"
          :get-option-key="getOptionKey"
          :placeholder="$t('chart.pick')"
          :reduce="option => option.chartID"
          label="name"
          :selectable="c => !c.deletedAt"
          class="flex-grow-1"
          @input="chartSelected"
        />
        <button
          title="Chart"
          :disabled="selectedChart && (!selectedChart.canUpdateChart && !selectedChart.canDeleteChart)"
          class="btn btn-outline-secondary d-flex align-items-center"
          @click="$router.push({ name: chartExternalLink, params: { chartID: (selectedChart || {}).chartID }, query: null })"
        >
          <font-awesome-icon :icon="['fas', 'external-link-alt']" />
        </button>
      </div>
    </div>

    <div
      v-if="selectedChart"
      class="row"
    >
      <div class="col-12 col-lg-6">
        <template v-if="isDrillDownAvailable">
          <div class="mb-3">
            <label class="d-flex align-items-center text-primary form-label">
              {{ $t('chart.drillDown.label') }}
              <c-input-checkbox
                v-model="options.drillDown.enabled"
                switch
                class="ms-1"
              />
            </label>
            <small class="form-text">{{ $t('chart.drillDown.description') }}</small>
            <div class="input-group">
              <c-input-select
                v-model="options.drillDown.blockID"
                :options="drillDownOptions"
                :get-option-key="getOptionKey"
                :disabled="!options.drillDown.enabled"
                :get-option-label="o => o.title || o.kind"
                :reduce="option => option.blockID"
                :clearable="true"
                :placeholder="$t('chart.drillDown.openInModal')"
                class="flex-grow-1"
              />
              <column-picker
                ref="columnPicker"
                :module="selectedChartModule"
                :fields="selectedDrilldownFields"
                :disabled="!!options.drillDown.blockID || !options.drillDown.enabled"
                variant="extra-light"
                size="md"
                @updateFields="onUpdateFields"
              >
                <font-awesome-icon :icon="['fas', 'wrench']" />
              </column-picker>
            </div>
          </div>
        </template>
      </div>

      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="d-flex align-items-center text-primary form-label">
            {{ $t('chart.enableLiveFilter') }}
          </label>
          <c-input-checkbox
            v-model="options.liveFilterEnabled"
            switch
            :labels="checkboxLabel"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed, inject } from 'vue'
import { usePageBlockBase } from './usePageBlockBase'
import { useStore } from '../../store'
import { NoID } from 'corteza-lib/js/dist'
import ColumnPicker from 'corteza-webapp-compose/src/components/Admin/Module/Records/ColumnPicker'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

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

const { options } = usePageBlockBase(props, emit)

const checkboxLabel = ref({ on: t('label.yes'), off: t('label.no') })

const charts = computed(() => store.chart.set)
const getModuleByID = computed(() => store.module.getByID)

const selectedChart = computed(() => {
  if (!options.value.chartID || options.value.chartID === NoID) return
  return charts.value.find(({ chartID }) => chartID === options.value.chartID)
})

const chartExternalLink = computed(() => !selectedChart.value ? 'admin.charts' : 'admin.charts.edit')

const selectedChartModuleID = computed(() => {
  if (!selectedChart.value) return
  const { moduleID } = (selectedChart.value.config.reports[0] || {})
  return moduleID
})

const selectedChartModule = computed(() => {
  if (!selectedChartModuleID.value) return
  return getModuleByID.value(selectedChartModuleID.value)
})

const selectedDrilldownFields = computed(() => {
  if (!selectedChart.value) return []
  return options.value.drillDown.recordListOptions.fields
})

const isDrillDownAvailable = computed(() => {
  if (!selectedChart.value) return
  const { metrics = [] } = (selectedChart.value.config.reports[0] || {})
  return !metrics.some(({ type }) => type === 'gauge' || type === 'radar')
})

const drillDownOptions = computed(() => {
  return props.blocks.filter(({ blockID, kind, options: o = {} }) =>
    kind === 'RecordList' && blockID !== NoID && o.moduleID === selectedChartModuleID.value)
})

function chartSelected () {
  props.block.resetDrillDown()
}

function getOptionKey ({ chartID }) {
  return chartID
}

function onUpdateFields (fields) {
  options.value.drillDown.recordListOptions.fields = fields.map(({ fieldID }) => fieldID)
}
</script>
