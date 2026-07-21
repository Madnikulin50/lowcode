<template>
  <div>
    <ul class="nav nav-tabs" role="tablist">
      <li class="nav-item" role="presentation">
        <button
          class="nav-link"
          :class="{ active: activeTab === 0 }"
          @click="activeTab = 0"
        >
          {{ $t('metric.edit.tabTitle') }}
        </button>
      </li>
    </ul>

    <div class="tab-content">
      <div class="tab-pane active">
        <div class="row g-0">
          <div class="col-12">
            <div
              v-for="(m, i) in metrics"
              :key="i"
              class="mb-2"
            >
              <button
                class="btn btn-outline-secondary me-1"
                @click="editMetric(m)"
              >
                {{ $t('label.edit') }}
              </button>
              <button
                class="btn btn-outline-danger me-2"
                @click="removeMetric(i)"
              >
                {{ $t('label.remove') }}
              </button>
              <span class="btn">
                {{ m.label || $t('metric.defaultMetricLabel') }}
              </span>
            </div>

            <button
              class="btn btn-link px-1"
              @click="addMetric"
            >
              + {{ $t('label.add') }}
            </button>
          </div>
        </div>

        <hr />

        <div class="row mt-3">
          <div
            v-if="edit"
            class="col-12 col-lg-7"
          >
            <div class="card mb-5">
              <div class="card-body">
                <fieldset>
                  <div class="mb-3">
                    <label class="form-label text-primary">{{ $t('metric.edit.labelLabel') }}</label>
                    <input
                      v-model="edit.label"
                      class="form-control mb-1"
                      :placeholder="$t('metric.edit.labelPlaceholder')"
                    />
                  </div>
                </fieldset>

                <template v-if="likeRecordList !== true">
                  <div class="mb-3">
                    <div class="form-check">
                      <input
                        v-model="edit.showLabel"
                        type="checkbox"
                        class="form-check-input"
                      />
                      <label class="form-check-label">{{ $t('metric.edit.showLabel') }}</label>
                    </div>
                  </div>

                  <div class="mb-3">
                    <div class="form-check">
                      <input
                        v-model="edit.showLabelTooltip"
                        type="checkbox"
                        class="form-check-input"
                      />
                      <label class="form-check-label">{{ $t('metric.edit.showLabelTooltip') }}</label>
                    </div>
                  </div>
                </template>

                <fieldset>
                  <h5>{{ $t('metric.edit.dimensionLabel') }}</h5>

                  <div class="mb-3">
                    <label class="form-label text-primary">{{ $t('metric.edit.moduleLabel') }}</label>
                    <c-input-select
                      v-model="edit.moduleID"
                      :options="modules"
                      label="name"
                      class="mt-1"
                      :reduce="o => o.moduleID"
                      :placeholder="$t('metric.edit.modulePlaceholder')"
                    />
                  </div>

                  <div class="mb-3">
                    <label class="form-label text-primary">{{ $t('metric.edit.filterLabel') }}</label>
                    <c-input-expression
                      v-model="edit.filter"
                      auto-complete
                      placeholder="(A > B) OR (A < C)"
                      class="mb-1"
                      min-height="3.448rem"
                      :suggestion-params="recordAutoCompleteParams"
                    />
                    <i18next
                      path="metric.edit.filterFootnote"
                      tag="small"
                      class="d-block text-muted"
                    >
                      <code>${record.values.fieldName}</code>
                      <code>${recordID}</code>
                      <code>${ownerID}</code>
                      <span><code>${userID}</code>, <code>${user.name}</code></span>
                    </i18next>
                  </div>
                </fieldset>

                <fieldset v-if="selectedMetricModule">
                  <h5>{{ $t('metric.edit.metricLabel') }}</h5>

                  <div class="mb-3">
                    <label class="form-label text-primary">{{ $t('metric.edit.metricFieldLabel') }}</label>
                    <c-input-select
                      v-model="edit.metricField"
                      :placeholder="$t('metric.edit.metricFieldSelect')"
                      :options="metricFields"
                      :get-option-key="getOptionMetricFieldKey"
                      :get-option-label="getOptionMetricFieldLabel"
                      :reduce="f => f.name"
                      @input="onMetricFieldChange"
                    />
                  </div>

                  <div
                    v-if="edit.metricField !== 'number_expression'"
                    class="mb-3"
                  >
                    <label class="form-label text-primary">{{ $t('metric.edit.metricAggregateLabel') }}</label>
                    <c-input-select
                      v-model="edit.operation"
                      :disabled="edit.metricField === 'count' || edit.metricField === 'number_expression'"
                      :placeholder="$t('metric.edit.metricSelectAggregate')"
                      :options="aggregationOperations"
                      :get-option-key="getOptionAggregationOperationKey"
                      :reduce="a => a.operation"
                    />
                  </div>

                  <div
                    v-if="edit.metricField === 'number_expression'"
                    class="mb-3"
                  >
                    <label class="form-label text-primary">{{ $t('metric.edit.expressionFieldLabel') }}</label>
                    <c-input-expression
                      v-model="edit.expression"
                      auto-complete
                      placeholder="$t('metric.edit.expressionPlaceholder')"
                      class="mb-1"
                      min-height="3.448rem"
                      :suggestion-params="recordAutoCompleteParams"
                    />
                  </div>

                  <div class="mb-3">
                    <label class="form-label text-primary">{{ $t('metric.edit.transformFunctionLabel') }}</label>
                    <c-input-expression
                      v-model="edit.transformFx"
                      auto-complete
                      placeholder="v"
                      class="mb-1"
                      min-height="3.448rem"
                      :suggestion-params="recordAutoCompleteParams"
                    />
                    <small>{{ $t('metric.edit.transformFunctionDescription') }}</small>
                    <i18next
                      path="metric.edit.transformFootnote"
                      tag="small"
                      class="d-block text-muted"
                    >
                      <code>${record.values.fieldName}</code>
                      <code>${recordID}</code>
                      <code>${ownerID}</code>
                      <span><code>${userID}</code>, <code>${user.name}</code></span>
                    </i18next>
                  </div>

                  <div class="mb-3">
                    <label class="form-label text-primary">{{ $t('metric.edit.numberFormat') }}</label>
                    <input
                      v-model="edit.numberFormat"
                      class="form-control mb-1"
                      placeholder="0.00"
                    />
                  </div>

                  <div class="mb-3">
                    <label class="form-label text-primary">{{ $t('metric.edit.prefixLabel') }}</label>
                    <input
                      v-model="edit.prefix"
                      class="form-control mb-1"
                      placeholder="$"
                    />
                  </div>

                  <div class="mb-3">
                    <label class="form-label text-primary">{{ $t('metric.edit.suffixLabel') }}</label>
                    <input
                      v-model="edit.suffix"
                      class="form-control mb-1"
                      placeholder="USD/mo"
                    />
                  </div>

                  <div class="mb-3">
                    <label class="d-flex align-items-center text-primary form-label">
                      {{ $t('metric.drillDown.label') }}
                      <input
                        v-model="edit.drillDown.enabled"
                        type="checkbox"
                        class="form-check-input ms-1 mb-1"
                      />
                    </label>
                    <small class="form-text">{{ $t('metric.drillDown.description') }}</small>
                    <div class="input-group">
                      <c-input-select
                        v-model="edit.drillDown.blockID"
                        :options="drillDownOptions"
                        :disabled="!edit.drillDown.enabled"
                        :get-option-label="o => o.title || o.kind"
                        :reduce="option => option.blockID"
                        :clearable="true"
                        :placeholder="$t('metric.drillDown.openInModal')"
                        append-to-body
                        class="flex-grow-1"
                      />
                      <column-picker
                        :module="selectedMetricModule"
                        :disabled="!!edit.drillDown.blockID || !edit.drillDown.enabled"
                        :fields="selectedDrilldownFields"
                        variant="extra-light"
                        size="md"
                        @updateFields="onUpdateFields"
                      >
                        <font-awesome-icon :icon="['fas', 'wrench']" />
                      </column-picker>
                    </div>
                  </div>
                </fieldset>
              </div>
            </div>

            <div class="mb-3">
              <div class="form-check">
                <input
                  v-model="likeRecordList"
                  type="checkbox"
                  class="form-check-input"
                />
                <label class="form-check-label">{{ $t('metric.likeRecordList') }}</label>
              </div>
            </div>

            <template v-if="likeRecordList">
              <div class="row g-0">
                <div class="col-12 col-lg-6">
                  <div class="mb-3">
                    <label class="form-label text-primary">{{ $t('record.horizontalFormLayout') }}</label>
                    <c-input-checkbox
                      v-model="options.horizontalFieldLayoutEnabled"
                      switch
                      :disabled="options.recordFieldLayoutOption === 'noWrap'"
                      :labels="checkboxLabel"
                    />
                  </div>
                </div>

                <div class="col-12 col-lg-6">
                  <div class="mb-3">
                    <label class="form-label text-primary">{{ $t('record.fieldsLayoutMode.label') }}</label>
                    <c-input-select
                      v-model="options.recordFieldLayoutOption"
                      :options="recordFieldLayoutOptions"
                      :reduce="option => option.value"
                      :get-option-key="option => option.label"
                      @input="handleRecordFieldLayout"
                    />
                  </div>
                </div>
              </div>
            </template>

            <m-style
              class="mt-2"
              :options="edit.valueStyle"
            >
              <template #title>
                <h5>{{ $t('metric.editStyle.valueLabel') }}</h5>
              </template>
            </m-style>
          </div>

          <div class="col-12 col-lg-5">
            <div
              v-if="metrics.length"
              class="d-flex flex-column position-sticky pt-2"
              style="top: 0;"
            >
              <button
                title="Refresh data"
                class="btn btn-outline-light d-flex align-items-center text-primary ms-auto border-0 px-2 mt-2 me-2 btn-lg"
                @click.prevent="refreshMetric"
              >
                <font-awesome-icon :icon="['fas', 'sync']" />
              </button>

              <div
                class="mt-2"
                style="height: 400px;"
              >
                <metric-base
                  v-bind="$props"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { usePageBlockBase } from '../usePageBlockBase'
import { useStore } from '../../../store'
import MStyle from './MStyle'
import MetricBase from '../MetricBase'
import ColumnPicker from 'corteza-webapp-compose/src/components/Admin/Module/Records/ColumnPicker'
import { compose, NoID } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
const { CInputExpression } = components

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

const { options, setBaseDefaultValues } = usePageBlockBase(props, emit)
const activeTab = ref(0)
const edit = ref(undefined)

const checkboxLabel = ref({ on: t('label.yes'), off: t('label.no') })

const dimensionModifiers = computed(() =>
  compose.chartUtil.dimensionFunctions.map(df => ({ ...df, text: t(`chart.edit.dimension.function.${df.text}`) }))
)

const aggregationOperations = computed(() => [
  { label: t('metric.edit.operationSum'), operation: 'sum' },
  { label: t('metric.edit.operationMax'), operation: 'max' },
  { label: t('metric.edit.operationMin'), operation: 'min' },
  { label: t('metric.edit.operationAvg'), operation: 'avg' },
  { label: t('metric.edit.operationUniqueCount'), operation: 'uniqueCount' },
])

const modules = computed(() => store.module.set)
const getModuleByID = computed(() => store.module.getByID)

const fields = computed(() => {
  if (!edit.value || !edit.value.moduleID) return []
  return getModuleByID.value(edit.value.moduleID).fields
})

const selectedDrilldownFields = computed(() => {
  if (!edit.value || !edit.value.drillDown?.recordListOptions?.fields) return []
  return edit.value.drillDown.recordListOptions.fields
})

const metricFields = computed(() => {
  return [
    { name: 'count', label: 'Count' },
    { name: 'number_expression', label: 'Number Expression' },
    ...fields.value.filter(f => f.kind === 'Number').sort((a, b) => a.label.localeCompare(b.label)),
  ]
})

const likeRecordList = computed({
  get: () => options.value.likeRecordList || false,
  set: (check) => { options.value.likeRecordList = check },
})

const metrics = computed({
  get: () => options.value.metrics,
  set: (m) => { options.value.metrics = m },
})

const drillDownOptions = computed(() =>
  props.page.blocks.filter(({ blockID, kind, options: o = {} }) =>
    kind === 'RecordList' && blockID !== NoID && o.moduleID === edit.value?.moduleID)
)

const selectedMetricModule = computed(() => {
  if (!edit.value?.moduleID) return undefined
  return getModuleByID.value(edit.value.moduleID)
})

const recordFieldLayoutOptions = computed(() => [
  { value: 'default', label: t('record.fieldsLayoutMode.default') },
  { value: 'noWrap', label: t('record.fieldsLayoutMode.noWrap') },
  { value: 'wrap', label: t('record.fieldsLayoutMode.wrap') },
])

watch(() => edit.value?.dimensionField, (df) => {
  if (!isTemporalField(df)) {
    edit.value.bucketSize = undefined
    edit.value.dateFormat = undefined
  } else {
    edit.value.dateFormat = edit.value.dateFormat || 'YYYY-MM-DD'
  }
})

if (!metrics.value.length) { addMetric() }
edit.value = metrics.value[0]

onBeforeUnmount(() => { setDefaultValues() })

function addMetric () {
  const m = props.block.makeMetric()
  metrics.value.push(m)
  editMetric(m)
}

function editMetric (m) { edit.value = m }
function removeMetric (i) { metrics.value.splice(i, 1); edit.value = undefined }
function isTemporalField (name) { return !!fields.value.find(f => f.name === name && f.kind === 'DateTime') }
function getOptionMetricFieldKey ({ name }) { return name }
function getOptionMetricFieldLabel ({ name, label }) { return label || name }
function getOptionAggregationOperationKey ({ operation }) { return operation }

function onMetricFieldChange (field) {
  if (field === 'count' || field === 'number_expression') {
    edit.value.operation = undefined
  } else if (!edit.value.operation) {
    edit.value.operation = aggregationOperations.value[0].operation
  }
}

function onUpdateFields (fields) { edit.value.drillDown.recordListOptions.fields = fields }

function handleRecordFieldLayout (v) {
  if (v !== 'noWrap') return
  options.value.horizontalFieldLayoutEnabled = false
}

function refreshMetric () {
  window.dispatchEvent(new CustomEvent('metric.update'))
}
</script>
