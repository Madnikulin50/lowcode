<template>
  <div>
    <div class="px-3">
      <h5 class="mb-3">
        {{ $t('edit.module.title') }}
      </h5>

      <div class="row">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.module.label') }}
            </label>
            <c-input-select
              v-model="moduleID"
              :options="modules"
              label="name"
              :reduce="module => module.moduleID"
              :get-option-key="option => option.moduleID"
              :placeholder="$t('edit.module.placeholder')"
            />
          </div>
        </div>

        <div
          v-if="!!module"
          class="col-12 col-lg-6"
        >
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.filter.preset') }}
            </label>
            <c-input-select
              v-model="report.filter"
              :options="predefinedFilters"
              label="text"
              :reduce="filter => filter.value"
              :placeholder="$t('edit.filter.noFilter')"
            />
          </div>
        </div>

        <div
          v-if="!!module"
          class="col-12 mt-1"
        >
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.filter.label') }}
            </label>
            <textarea
              v-model="report.filter"
              class="form-control"
              :placeholder="$t('edit.filter.placeholder')"
            ></textarea>

            <i18next
              path="edit.filter.footnote"
              tag="small"
              class="text-muted"
            >
              <code>${record.values.fieldName}</code>
              <code>${recordID}</code>
              <code>${ownerID}</code>
              <span><code>${userID}</code>, <code>${user.name}</code></span>
            </i18next>
          </div>
        </div>
      </div>
    </div>
    <hr v-if="module">

    <div
      v-if="!!module"
      class="px-3"
    >
      <fieldset
        v-for="(d, i) in dimensions"
        :key="i"
      >
        <h5 class="mb-3">
          {{ $t('edit.dimension.label') }}
          <small
            v-if="dimensions.length > 1"
            class="text-muted"
          >
            {{ i + 1 }}
          </small>
          <button
            v-if="i === dimensions.length - 1 && canAddDimension"
            type="button"
            class="btn btn-link text-decoration-none p-0 ms-2 align-baseline"
            @click.prevent="addDimension"
          >
            + {{ $t('edit.dimension.add') }}
          </button>
        </h5>

        <template v-if="usesDimensionsField">
          <div class="row">
            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">
                  {{ $t('edit.dimension.fieldLabel') }}
                </label>
                <c-input-select
                  v-model="d.field"
                  :options="dimensionFields"
                  label="text"
                  :reduce="field => field.value"
                  :placeholder="$t('edit.dimension.fieldPlaceholder')"
                  @input="value => onDimFieldChange(value, d)"
                />
              </div>
            </div>

            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">
                  {{ $t('edit.dimension.function.label') }}
                </label>
                <c-input-select
                  v-model="d.modifier"
                  :disabled="!d.field || !isTemporalField(d.field)"
                  :options="dimensionModifiers"
                  label="text"
                  :reduce="modifier => modifier.value"
                  :placeholder="$t('edit.dimension.function.placeholder')"
                  @input="onDimModifierChange($event, d)"
                />
              </div>
            </div>
          </div>

          <template v-if="!unSkippable">
            <div class="row">
              <div class="col-12 col-lg-6">
                <div class="mb-3">
                  <label class="form-label text-primary">
                    {{ $t('edit.dimension.defaultValueLabel') }}
                  </label>
                  <input
                    v-model="d.default"
                    :type="defaultValueInputType(d)"
                    class="form-control form-control-sm"
                  />
                  <div class="form-text">
                    {{ $t('edit.dimension.defaultValueFootnote') }}
                  </div>
                </div>
              </div>

              <div class="col-12 col-lg-6">
                <div class="mb-3">
                  <label class="form-label text-primary">
                    {{ $t('edit.dimension.options.label') }}
                  </label>
                  <div class="form-check">
                    <input
                      v-model="d.skipMissing"
                      class="form-check-input"
                      type="checkbox"
                      :id="`skipMissing-${i}`"
                    />
                    <label
                      class="form-check-label"
                      :for="`skipMissing-${i}`"
                    >
                      {{ $t('edit.dimension.skipMissingValues') }}
                    </label>
                  </div>

                  <slot
                    name="dimension-options-options"
                    :dimension="d"
                    :is-temporal="isTemporalField(d.field)"
                  />
                </div>
              </div>
            </div>
          </template>
        </template>

        <slot
          name="dimension-options"
          :index="i"
          :dimension="d"
          :field="getField(d)"
        />
      </fieldset>
    </div>
    <hr v-if="!!module">

    <div
      v-if="!!module"
      class="px-3"
    >
      <h5 class="d-flex align-items-center mb-3">
        {{ $t('edit.metric.title') }}
        <button
          v-if="canAddMetric"
          type="button"
          class="btn btn-link text-decoration-none"
          @click.prevent="addMetric"
        >
          + {{ $t('edit.metric.add') }}
        </button>
      </h5>

      <draggable
            item-key="id"
        class="metrics mb-3"
        :list="metrics"
        handle=".grab"
        :group="`metrics_${moduleID}`"
      >
        <template #item="{ element, index }">
          <div
            :key="index"
            class="metric rounded border border-light p-3 mb-3"
          >
            <h5
              v-if="metrics.length > 1"
              class="d-flex align-items-center mb-3"
            >
              {{ $t('edit.metric.label') }} {{ index + 1 }}

              <div class="d-flex align-items-center ms-auto">
                <c-input-confirm
                  show-icon
                  class="me-2"
                  @confirmed="removeMetric(index)"
                />

                <button
                  type="button"
                  class="btn btn-link btn-sm ms-auto px-0"
                >
                  <font-awesome-icon
                    :icon="['fas', 'bars']"
                    class="grab text-secondary"
                  />
                </button>
              </div>
            </h5>

            <div class="row">
              <div class="col-12 col-lg-6">
                <div class="mb-3">
                  <label class="form-label text-primary">
                    {{ $t('edit.metric.fieldLabel') }}
                  </label>
                  <c-input-select
                    v-model="element.field"
                    :options="metricFields"
                    :get-option-key="option => option.text"
                    label="text"
                    :reduce="option => option.value"
                    @input="value => onMetricFieldChange(value, element)"
                  />
                </div>
              </div>

              <div class="col-12 col-lg-6">
                <div class="mb-3">
                  <label class="form-label text-primary">
                    {{ $t('edit.metric.function.label') }}
                  </label>
                  <c-input-select
                    v-model="element.aggregate"
                    :disabled="!element.field || element.field === 'count'"
                    :options="metricAggregates"
                    label="text"
                    :reduce="option => option.value"
                    :get-option-key="option => option.text"
                    :placeholder="$t('edit.metric.function.placeholder')"
                    @input="value => onMetricFieldChange(value, element)"
                  />
                </div>
              </div>
            </div>

            <slot
              name="metric-options"
              :metric="element"
              :report="editReport"
            />
          </div>
        </template>
      </draggable>
    </div>

    <hr v-if="!!module && hasAxis">

    <template v-if="hasAxis">
      <slot
        name="y-axis"
        :report="editReport"
      />
    </template>

    <hr v-if="hasLegend">

    <div
      v-if="hasLegend"
      class="px-3"
    >
      <h5 class="mb-3">
        {{ $t('edit.additionalConfig.legend.label') }}
      </h5>

      <div class="row">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.additionalConfig.legend.orientation.label') }}
            </label>
            <select
              v-model="report.legend.orientation"
              class="form-select form-control form-select-sm"
            >
              <option
                v-for="opt in orientations"
                :key="opt.value"
                :value="opt.value"
              >
                {{ opt.text }}
              </option>
            </select>
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.additionalConfig.legend.show') }}
            </label>
            <c-input-checkbox
              v-model="legendShown"
              switch
              :labels="checkboxLabel"
            />
          </div>
        </div>
      </div>

      <div class="row">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.additionalConfig.legend.align.label') }}
            </label>
            <select
              v-model="report.legend.align"
              class="form-select form-control form-select-sm"
              :disabled="!report.legend.position.isDefault"
            >
              <option
                v-for="opt in alignments"
                :key="opt.value"
                :value="opt.value"
              >
                {{ opt.text }}
              </option>
            </select>
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.additionalConfig.legend.options.label') }}
            </label>
            <div class="form-check">
              <input
                v-model="report.legend.isScrollable"
                class="form-check-input"
                type="checkbox"
                :disabled="report.legend.orientation !== 'horizontal'"
                :id="`legendScrollable`"
              />
              <label
                class="form-check-label"
                for="legendScrollable"
              >
                {{ $t('edit.additionalConfig.legend.scrollable') }}
              </label>
            </div>

            <div class="form-check">
              <input
                v-model="report.legend.position.isDefault"
                class="form-check-input"
                type="checkbox"
                :id="`legendPositionCustom`"
              />
              <label
                class="form-check-label"
                for="legendPositionCustom"
              >
                {{ $t('edit.additionalConfig.legend.position.customize') }}
              </label>
            </div>
          </div>
        </div>
      </div>

      <div
        v-if="!report.legend.position.isDefault"
        class="row"
      >
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.additionalConfig.legend.position.top') }}
            </label>
            <input
              v-model="report.legend.position.top"
              class="form-control form-control-sm"
            />
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.additionalConfig.legend.position.right') }}
            </label>
            <input
              v-model="report.legend.position.right"
              class="form-control form-control-sm"
            />
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.additionalConfig.legend.position.bottom') }}
            </label>
            <input
              v-model="report.legend.position.bottom"
              class="form-control form-control-sm"
            />
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.additionalConfig.legend.position.left') }}
            </label>
            <input
              v-model="report.legend.position.left"
              class="form-control form-control-sm"
            />
          </div>
        </div>

        <div class="col-12">
          <small class="text-muted">
            {{ $t('edit.additionalConfig.legend.valueRange') }}
          </small>
        </div>
      </div>
    </div>

    <slot
      name="additional-config"
      :report="editReport"
      :metrics="metrics"
      :has-axis="hasAxis"
    />
  </div>
</template>

<script setup>
import { computed, ref, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import draggable from 'vuedraggable'
import { compose } from 'corteza-lib/js/dist'

const { t } = useI18n()

const aggregateFunctions = [
  { value: 'SUM', text: 'sum' },
  { value: 'MAX', text: 'max' },
  { value: 'MIN', text: 'min' },
  { value: 'AVG', text: 'avg' },
  { value: 'STD', text: 'std' },
  { value: 'uniqueCount', text: 'uniqueCount' },
]

const props = defineProps({
  report: { type: Object, required: false, default: undefined },
  chart: { type: Object, default: () => new compose.Chart() },
  modules: { type: Array, required: true },
  supportedMetrics: { type: Number, default: -1 },
  dimensionFieldKind: { type: Array, default: () => ['DateTime', 'Select', 'Number', 'Bool', 'String', 'Record', 'User'] },
  usesDimensionsField: { type: Boolean, default: true },
  unSkippable: { type: Boolean, default: false },
})

const emit = defineEmits(['update:report'])

const checkboxLabel = ref({
  on: t('label.yes'),
  off: t('label.no'),
})

const formatOptions = ref([
  { value: 'custom', text: t('edit.formatting.presetFormats.options.custom') },
  { value: 'accounting', text: t('edit.formatting.presetFormats.options.accounting') },
])

const metricAggregates = ref(aggregateFunctions.map(af => ({ ...af, text: t(`edit.metric.function.${af.text}`) })))
const dimensionModifiers = ref(compose.chartUtil.dimensionFunctions.map(df => ({ ...df, text: t(`edit.dimension.function.${df.text}`) })))
const predefinedFilters = ref(compose.chartUtil.predefinedFilters.map(pf => ({ ...pf, text: t(`edit.filter.${pf.text}`) })))

const alignments = ref([
  { value: 'left', text: t('edit.additionalConfig.legend.align.left') },
  { value: 'center', text: t('edit.additionalConfig.legend.align.center') },
  { value: 'right', text: t('edit.additionalConfig.legend.align.right') },
])

const orientations = ref([
  { value: 'horizontal', text: t('edit.additionalConfig.legend.orientation.horizontal') },
  { value: 'vertical', text: t('edit.additionalConfig.legend.orientation.vertical') },
])

const editReport = computed({
  get: () => props.report,
  set: (v) => emit('update.report', v),
})

const legendShown = computed({
  get: () => !props.report?.legend?.isHidden,
  set: (v) => {
    props.report.legend.isHidden = !v
    emit('update.report', { ...props.report, legend: props.report.legend })
  },
})

const module = computed(() => props.modules.find(m => m.moduleID === moduleID.value))

const moduleID = computed({
  get: () => props.report?.moduleID,
  set: (v) => {
    props.report.moduleID = v
    emit('update.report', { ...props.report, moduleID: v })
  },
})

const metrics = computed({
  get: () => props.report?.metrics,
  set: (v) => {
    props.report.metrics = v
    emit('update.report', { ...props.report, metrics: v })
  },
})

const dimensions = computed({
  get: () => props.report?.dimensions,
  set: (v) => {
    emit('update.report', { ...props.report, dimensions: v })
  },
})

const hasLegend = computed(() => !metrics.value?.some(({ type }) => ['gauge'].includes(type)))

const hasAxis = computed(() => metrics.value?.some(({ type }) => ['bar', 'line', 'scatter', 'waterfall', 'boxplot', 'candlestick', 'heatmap', 'parallel'].includes(type)))

// Charts that need more than one dimension (source/target style charts)
const multiDimensionCharts = ['sankey', 'graph', 'heatmap', 'sunburst']

const canAddDimension = computed(() => {
  if (!moduleID.value) return false
  const t = metrics.value?.[0]?.type
  return multiDimensionCharts.includes(t) && (dimensions.value?.length || 0) < 2
})

const canAddMetric = computed(() => (props.supportedMetrics < 0 || (metrics.value?.length || 0) < props.supportedMetrics) && moduleID.value)

const metricFields = computed(() => {
  if (!module.value) return []
  return [
    { value: 'count', text: t('label.count') },
    ...module.value.fields.filter(f => f.kind === 'Number')
      .sort((a, b) => (a.label || a.name).localeCompare((b.label || b.name)))
      .map(({ label, name }) => ({ value: name, text: label || name })),
  ]
})

const dimensionFields = computed(() => {
  if (!module.value) return []
  return [
    ...[...module.value.fields].sort((a, b) => (a.label || a.name).localeCompare((b.label || b.name))),
    ...module.value.systemFields().map(sf => {
      sf.label = t(`system.${sf.name}`)
      return sf
    }),
  ].filter(({ kind, options = {} }) => {
    return props.dimensionFieldKind.includes(kind) && !(options.useRichTextEditor || options.multiLine)
  }).map(({ name, label, kind }) => {
    return { value: name, text: `${label || name} (${kind})`, kind }
  })
})

function defaultValueInputType ({ field }) {
  return (module.value?.fields?.filter?.(f => f.name === field)?.[0] || {}).kind === 'DateTime' ? 'date' : 'text'
}

function getField ({ field }) {
  if (!field || !module.value) return undefined
  return module.value.fields.find(({ name }) => name === field)
}

function addMetric () {
  metrics.value = [...(metrics.value || []), props.chart.defMetric()]
}

function addDimension () {
  dimensions.value = [...(dimensions.value || []), props.chart.defDimension()]
}

function onDimFieldChange (f, d) {
  if (!isTemporalField(f)) {
    d.modifier = dimensionModifiers.value[0]?.value
    d.timeLabels = false
  }
  d.meta.fields = []
}

function onDimModifierChange (modifier, d) {
  if (['WEEK', 'QUARTER'].includes(modifier)) {
    d.timeLabels = false
  }
}

function onMetricFieldChange (field, m) {
  if (field === 'count') {
    m.aggregate = undefined
  } else if (field) {
    const moduleField = module.value?.fields?.find?.(f => f.name === field)
    if (moduleField) {
      const { presetFormat, format, prefix, suffix } = moduleField.options
      m.formatting = { presetFormat, format, prefix, suffix }
    }
    if (!m.aggregate) {
      m.aggregate = metricAggregates.value[0]?.value
    }
  }
}

function removeMetric (i) {
  const newMetrics = [...(metrics.value || [])]
  newMetrics.splice(i, 1)
  metrics.value = newMetrics
}

function isTemporalField (name) {
  return dimensionFields.value.some(f => f.value === name && f.kind === 'DateTime')
}

onBeforeUnmount(() => {
  metricAggregates.value = []
  dimensionModifiers.value = []
  predefinedFilters.value = []
  alignments.value = []
  orientations.value = []
})
</script>

<style lang="scss" scoped>
.metrics {
  .metric {
    background-color: var(--body-bg);
  }
}
</style>
