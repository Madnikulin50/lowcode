<template>
  <div>
    <div class="mb-3">
      <h5 class="text-primary mb-2">{{ t('chart.configurator.general') }}</h5>
      <div class="row align-items-stretch">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="text-primary form-label">{{ t('chart.configurator.type') }}</label>
            <select v-model="displayElementOptions.type" class="form-select" @change="typeChanged">
              <option v-for="ct in chartTypes" :key="ct.value" :value="ct.value">{{ ct.text }}</option>
            </select>
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="text-primary form-label">{{ t('chart.configurator.chart-title') }}</label>
            <input v-model="localOptions.title" class="form-control" />
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="text-primary form-label">{{ t('chart.configurator.color-scheme') }}</label>
            <c-input-select
              v-model="localOptions.colorScheme"
              :options="colorSchemes"
              :get-option-key="getOptionKey"
              :reduce="cs => cs.value"
              :placeholder="t('label.default')"
            >
              <template #option="option">
                <p class="mb-1">{{ option.label }}</p>
                <div v-for="(color, idx) in option.colors" :key="`${option.value}-${idx}`" :style="`background:${color}`" class="d-inline-block color-box me-1 mb-1" />
              </template>
            </c-input-select>
            <template v-if="currentColorScheme">
              <div v-for="(color, idx) in currentColorScheme.colors" :key="`${currentColorScheme.value}-${idx}`" :style="`background:${color}`" class="d-inline-block color-box me-1" />
            </template>
          </div>
        </div>
        <div class="col-12 col-lg-6 d-flex flex-column justify-content-center">
          <div class="form-check form-switch mt-3 pt-2">
            <input v-model="localOptions.noAnimation" :true-value="undefined" :false-value="true" class="form-check-input" type="checkbox" id="noAnim" />
            <label class="form-check-label" for="noAnim">{{ t('chart.configurator.animation.enabled') }}</label>
          </div>
        </div>
      </div>
      <hr />
    </div>
    <div v-if="localOptions.source" class="mb-3">
      <h5 class="text-primary mb-2">{{ t('chart.configurator.data') }}</h5>
      <div class="mb-3">
        <label class="text-primary form-label">{{ t('chart.configurator.label-column') }}</label>
        <ColumnSelector v-model="localOptions.labelColumn" :columns="labelColumns" style="min-width:100% !important" />
      </div>
      <div class="mb-3">
        <label class="text-primary form-label">{{ t('chart.configurator.data-columns') }}</label>
        <ColumnPicker :all-items="dataColumns" :selected-items="localOptions.dataColumns" class="d-flex flex-column" @update:selected-items="updateDataColumns" />
      </div>
      <div v-if="['bar', 'line'].includes(localOptions.type)">
        <hr />
        <h5 class="text-primary mb-2">{{ t('chart.configurator.metrics.label') }}</h5>
        <div v-if="localOptions.dataColumns?.length" class="mb-3">
          <c-form-table-wrapper v-for="(col, idx) in metricColumns" :key="idx" hide-add-button class="mb-3">
            <h6 class="mb-3">{{ col.label }}</h6>
            <div class="row">
              <div class="col-12 col-md-6">
                <div class="mb-3">
                  <label class="text-primary form-label">{{ t('chart.configurator.metrics.label-field.label') }}</label>
                  <input v-model="col.label" class="form-control" />
                </div>
              </div>
              <div class="col-12 col-md-6">
                <div class="mb-3">
                  <label class="text-primary form-label">{{ t('chart.configurator.metrics.stack.label') }}</label>
                  <input v-model="col.stack" class="form-control" />
                </div>
              </div>
            </div>
          </c-form-table-wrapper>
        </div>
      </div>
      <div v-if="['bar', 'line'].includes(localOptions.type)">
        <hr />
        <div class="mb-3">
          <h5 class="text-primary mb-2">{{ t('chart.configurator.x-axis.name') }}</h5>
          <div class="row">
            <div class="col">
              <div class="mb-3">
                <label class="text-primary form-label">{{ t('chart.configurator.x-axis.label') }}</label>
                <input v-model="localOptions.xAxis.label" class="form-control" />
              </div>
            </div>
            <div class="col">
              <div class="mb-3">
                <label class="text-primary form-label">{{ t('chart.configurator.x-axis.type') }}</label>
                <select v-model="localOptions.xAxis.type" class="form-select">
                  <option value="">{{ t('chart.configurator.default') }}</option>
                  <option v-for="at in AxisTypes" :key="at.value" :value="at.value">{{ at.text }}</option>
                </select>
              </div>
            </div>
          </div>
          <div class="row">
            <div class="col">
              <div class="mb-3">
                <label class="text-primary form-label">{{ t('chart.configurator.default-value') }}</label>
                <input v-model="localOptions.xAxis.defaultValue" class="form-control" :disabled="localOptions.xAxis.skipMissing" :type="localOptions.xAxis.type === 'time' ? 'date' : 'text'" />
              </div>
              <div class="form-check mb-3">
                <input v-model="localOptions.xAxis.skipMissing" class="form-check-input" type="checkbox" id="skipMissing" />
                <label class="form-check-label" for="skipMissing">{{ t('chart.configurator.skip-missing-values') }}</label>
              </div>
            </div>
            <div class="col">
              <div class="mb-3">
                <label class="text-primary form-label">{{ t('chart.configurator.x-axis.labelRotation.label') }}</label>
                <input v-model.number="localOptions.xAxis.labelRotation" type="number" class="form-control" />
              </div>
            </div>
          </div>
        </div>
        <hr />
        <div>
          <h5 class="text-primary mb-2">{{ t('chart.configurator.y-axis.name') }}</h5>
          <div class="row">
            <div class="col">
              <div class="mb-3">
                <label class="text-primary form-label">{{ t('chart.configurator.y-axis.label') }}</label>
                <input v-model="localOptions.yAxis.label" class="form-control" />
              </div>
            </div>
            <div class="col">
              <div class="mb-3">
                <label class="text-primary form-label">{{ t('chart.configurator.y-axis.labelPosition.label') }}</label>
                <select v-model="localOptions.yAxis.labelPosition" class="form-select">
                  <option v-for="alp in axisLabelPositions" :key="alp.value" :value="alp.value">{{ alp.text }}</option>
                </select>
              </div>
            </div>
            <div class="col">
              <div class="mb-3">
                <label class="text-primary form-label">{{ t('chart.configurator.y-axis.labelRotation.label') }}</label>
                <input v-model.number="localOptions.yAxis.labelRotation" type="number" class="form-control" />
              </div>
            </div>
          </div>
          <div class="row">
            <div class="col">
              <div class="mb-3">
                <label class="text-primary form-label">{{ t('chart.configurator.value.min') }}</label>
                <input v-model.number="localOptions.yAxis.min" type="number" class="form-control" />
              </div>
            </div>
            <div class="col">
              <div class="mb-3">
                <label class="text-primary form-label">{{ t('chart.configurator.value.max') }}</label>
                <input v-model.number="localOptions.yAxis.max" type="number" class="form-control" />
              </div>
            </div>
            <div class="col">
              <div class="mb-3">
                <label class="text-primary form-label">{{ t('chart.configurator.step-size') }}</label>
                <input v-model="localOptions.yAxis.stepSize" class="form-control" />
              </div>
            </div>
          </div>
          <div class="row">
            <div class="col">
              <div class="form-check">
                <input v-model="localOptions.yAxis.type" true-value="logarithmic" false-value="linear" class="form-check-input" type="checkbox" id="logScale" />
                <label class="form-check-label" for="logScale">{{ t('chart.configurator.logarithmic-scale') }}</label>
              </div>
              <div class="form-check">
                <input v-model="localOptions.yAxis.beginAtZero" class="form-check-input" type="checkbox" id="beginAtZero" />
                <label class="form-check-label" for="beginAtZero">{{ t('chart.configurator.begin-axis-at-zero') }}</label>
              </div>
              <div class="form-check">
                <input v-model="localOptions.yAxis.position" true-value="right" false-value="left" class="form-check-input" type="checkbox" id="axisRight" />
                <label class="form-check-label" for="axisRight">{{ t('chart.configurator.place-axis-on-right-side') }}</label>
              </div>
            </div>
          </div>
        </div>
      </div>
      <hr />
      <div>
        <h5 class="text-primary mb-2">{{ t('chart.configurator.legend.name') }}</h5>
        <div class="form-check mb-3">
          <input v-model="localOptions.legend.hide" class="form-check-input" type="checkbox" id="hideLegend" />
          <label class="form-check-label" for="hideLegend">{{ t('chart.configurator.legend.hide') }}</label>
        </div>
      </div>
      <hr />
      <div class="mb-2">
        <h5 class="text-primary mb-2">{{ t('chart.configurator.offset.name') }}</h5>
        <div class="form-check mb-3">
          <input v-model="localOptions.offset.default" class="form-check-input" type="checkbox" id="defaultOffset" />
          <label class="form-check-label" for="defaultOffset">{{ t('chart.configurator.offset.default') }}</label>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSettings } from 'corteza-lib/vue/dist'
import { reporter, shared } from 'corteza-lib/js/dist'
import ColumnSelector from '../../../../Common/ColumnSelector.vue'
import ColumnPicker from '../../../../Common/ColumnPicker.vue'
const { colorschemes } = shared

const props = defineProps({
  displayElementOptions: { type: Object, default: () => ({}) },
  columns: { type: Array, required: true },
  datasource: { type: Object, default: undefined },
})
const emit = defineEmits(['update:displayElementOptions'])

const { t } = useI18n()
const { settings } = useSettings()

const colorSchemes = ref([])
const allowedLabelKinds = ['Date', 'DateTime', 'Select', 'Number', 'Bool', 'String', 'Record', 'User']
const axisLabelPositions = [
  { value: 'end', text: t('chart.configurator.y-axis.labelPosition.top') },
  { value: 'center', text: t('chart.configurator.y-axis.labelPosition.center') },
  { value: 'start', text: t('chart.configurator.y-axis.labelPosition.bottom') },
]
const orientations = [
  { value: 'horizontal', text: t('chart.configurator.legend.orientation.horizontal') },
  { value: 'vertical', text: t('chart.configurator.legend.orientation.vertical') },
]
const alignments = [
  { value: 'left', text: t('chart.configurator.legend.align.left') },
  { value: 'center', text: t('chart.configurator.legend.align.center') },
  { value: 'right', text: t('chart.configurator.legend.align.right') },
]

const localOptions = computed({
  get: () => props.displayElementOptions || {},
  set: (val) => emit('update.displayElementOptions', val),
})

const mergedColumns = computed(() => [].concat(...props.columns))

const chartTypes = computed(() => {
  const types = [
    { value: 'bar', text: t('chart.configurator.types.bar') },
    { value: 'line', text: t('chart.configurator.types.line') },
    { value: 'pie', text: t('chart.configurator.types.pie') },
    { value: 'doughnut', text: t('chart.configurator.types.doughnut') },
  ]
  if (props.datasource?.step?.aggregate) types.push({ value: 'funnel', text: t('chart.configurator.types.funnel') })
  return types
})

const AxisTypes = [
  { value: 'time', text: t('chart.configurator.time.label') },
  { value: 'category', text: 'Category' },
]

const timeUnits = [
  { value: 'day', text: t('chart.configurator.time.unit.types.date') },
  { value: 'week', text: t('chart.configurator.time.unit.types.week') },
  { value: 'month', text: t('chart.configurator.time.unit.types.month') },
  { value: 'quarter', text: t('chart.configurator.time.unit.types.quarter') },
  { value: 'year', text: t('chart.configurator.time.unit.types.year') },
]

const labelColumns = computed(() => {
  const cols = props.columns.length ? props.columns[0] : []
  return cols.filter(({ kind }) => allowedLabelKinds.includes(kind))
})

const dataColumns = computed(() =>
  [...mergedColumns.value.filter(({ kind }) => kind === 'Number')].sort((a, b) => a.label.localeCompare(b.label))
)

const currentColorScheme = computed(() => colorSchemes.value.find(({ value }) => value === localOptions.value.colorScheme))

const metricColumns = computed(() => localOptions.value.dataColumns)

onMounted(() => {
  const capitalize = w => `${w[0].toUpperCase()}${w.slice(1)}`
  const splicer = sc => {
    const rr = (/(\D+)(\d+)$/gi).exec(sc)
    return { label: rr[1], count: rr[2] }
  }
  const rr = []
  for (const g in colorschemes) {
    for (const sc in colorschemes[g]) {
      const gn = splicer(sc)
      rr.push({ label: `${capitalize(g)}: ${capitalize(gn.label)}`, colors: [...colorschemes[g][sc]], value: `${g}.${sc}` })
    }
  }
  colorSchemes.value = rr
})

function updateDataColumns(columns) {
  const result = columns.map(c => dataColumns.value.find(({ name }) => name === c))
  result.forEach((c, idx) => {
    const initialCol = localOptions.value.dataColumns?.find(({ name }) => name === c?.name)
    if (initialCol) result[idx] = initialCol
  })
  localOptions.value.dataColumns = result
}

function typeChanged(type) {
  localOptions.value = reporter.ChartOptionsMaker({ ...localOptions.value, type })
}

function getOptionKey({ value }) { return value }
</script>
<style lang="scss" scoped>
.color-box { width: 28px; height: 12px; }
</style>