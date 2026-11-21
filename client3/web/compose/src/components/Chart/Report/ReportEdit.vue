<template>
  <div>
    <section
      id="section-data-source"
      class="chart-editor-section"
    >
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
    </section>

    <section
      v-if="!!module"
      id="section-dimensions"
      class="chart-editor-section"
    >
      <h5 class="d-flex align-items-center mb-3">
        {{ $t('edit.dimension.label') }}
        <button
          v-if="canAddDimension"
          type="button"
          class="btn btn-link text-decoration-none"
          @click.prevent="addDimension"
        >
          + {{ $t('edit.dimension.add') }}
        </button>
      </h5>

      <div
        v-for="(d, i) in dimensions"
        :key="i"
        class="chart-editor-chip"
        :class="{ 'is-open': !!openDimensions[i] }"
      >
        <div
          class="chart-editor-chip-head"
          @click="toggleDimension(i)"
        >
          <span class="chart-editor-chip-tag chart-editor-chip-tag--dimension">
            {{ $t('edit.dimension.label') }}
          </span>
          <span class="chart-editor-chip-title">
            {{ dimensionSummary(d).title }}
          </span>
          <span
            v-if="dimensionSummary(d).subtitle"
            class="chart-editor-chip-sub"
          >
            · {{ dimensionSummary(d).subtitle }}
          </span>
          <small
            v-if="dimensions.length > 1"
            class="text-muted ms-1"
          >
            {{ i + 1 }}
          </small>
          <font-awesome-icon
            :icon="['fas', 'chevron-down']"
            class="chart-editor-chip-chevron ms-auto"
          />
        </div>

        <div
          v-show="openDimensions[i]"
          class="chart-editor-chip-body"
        >
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
        </div>
      </div>
    </section>

    <section
      v-if="!!module"
      id="section-metrics"
      class="chart-editor-section"
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
            class="chart-editor-chip"
            :class="{ 'is-open': !!openMetrics[index] }"
          >
            <div
              class="chart-editor-chip-head"
              @click="toggleMetric(index)"
            >
              <button
                v-if="metrics.length > 1"
                type="button"
                class="btn btn-link btn-sm p-0 chart-editor-chip-grab"
                @click.stop
              >
                <font-awesome-icon
                  :icon="['fas', 'bars']"
                  class="grab text-secondary"
                />
              </button>

              <span class="chart-editor-chip-tag chart-editor-chip-tag--metric">
                {{ $t('edit.metric.label') }}
              </span>
              <span class="chart-editor-chip-title">
                {{ metricSummary(element).title }}
              </span>
              <span
                v-if="metricSummary(element).subtitle"
                class="chart-editor-chip-sub"
              >
                · {{ metricSummary(element).subtitle }}
              </span>
              <small
                v-if="metrics.length > 1"
                class="text-muted ms-1"
              >
                {{ index + 1 }}
              </small>

              <c-input-confirm
                v-if="metrics.length > 1"
                show-icon
                class="ms-auto chart-editor-chip-delete"
                @click.stop
                @confirmed="removeMetric(index)"
              />

              <font-awesome-icon
                :icon="['fas', 'chevron-down']"
                class="chart-editor-chip-chevron"
                :class="{ 'ms-auto': metrics.length <= 1 }"
              />
            </div>

            <div
              v-show="openMetrics[index]"
              class="chart-editor-chip-body"
            >
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
          </div>
        </template>
      </draggable>
    </section>

    <section
      v-if="!!module && (hasAxis || hasLegend)"
      id="section-axes-legend"
      class="chart-editor-section"
    >
    <template v-if="hasAxis">
      <slot
        name="y-axis"
        :report="editReport"
      />
    </template>

    <div
      v-if="hasLegend"
      :class="{ 'mt-4': hasAxis }"
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
    </section>

    <slot
      name="additional-config"
      :report="editReport"
      :metrics="metrics"
      :has-axis="hasAxis"
    />
  </div>
</template>

<script setup>
import { computed, ref, watch, onBeforeUnmount } from 'vue'
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
  set: (v) => emit('update:report', v),
})

const legendShown = computed({
  get: () => !props.report?.legend?.isHidden,
  set: (v) => {
    props.report.legend.isHidden = !v
    emit('update:report', { ...props.report, legend: props.report.legend })
  },
})

const module = computed(() => props.modules.find(m => m.moduleID === moduleID.value))

const moduleID = computed({
  get: () => props.report?.moduleID,
  set: (v) => {
    props.report.moduleID = v
    emit('update:report', { ...props.report, moduleID: v })
  },
})

const metrics = computed({
  get: () => props.report?.metrics,
  set: (v) => {
    props.report.metrics = v
    emit('update:report', { ...props.report, metrics: v })
  },
})

const dimensions = computed({
  get: () => props.report?.dimensions,
  set: (v) => {
    emit('update:report', { ...props.report, dimensions: v })
  },
})

// Dimension/metric cards are collapsible — only the first of each starts
// expanded, the rest collapse to a one-line summary (field + function).
// Re-seeded whenever the report itself is swapped for another one (e.g.
// funnel charts editing a different report), not on every field edit.
const openDimensions = ref((props.report?.dimensions || []).map((_, i) => i === 0))
const openMetrics = ref((props.report?.metrics || []).map((_, i) => i === 0))

watch(() => props.report, () => {
  openDimensions.value = (dimensions.value || []).map((_, i) => i === 0)
  openMetrics.value = (metrics.value || []).map((_, i) => i === 0)
})

function toggleDimension (i) { openDimensions.value[i] = !openDimensions.value[i] }

function toggleMetric (i) { openMetrics.value[i] = !openMetrics.value[i] }

function dimensionSummary (d) {
  if (!d?.field) return { title: t('edit.dimension.fieldPlaceholder') }
  const field = dimensionFields.value.find(f => f.value === d.field)
  const modifier = dimensionModifiers.value.find(m => m.value === d.modifier)
  return {
    title: field ? field.text : d.field,
    subtitle: modifier?.value ? modifier.text : '',
  }
}

function metricSummary (m) {
  if (!m?.field) return { title: t('edit.metric.fieldPlaceholder') }
  const field = metricFields.value.find(f => f.value === m.field)
  const aggregate = metricAggregates.value.find(a => a.value === m.aggregate)
  return {
    title: field ? field.text : m.field,
    subtitle: aggregate ? aggregate.text : '',
  }
}

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
  openMetrics.value.push(true)
}

function addDimension () {
  dimensions.value = [...(dimensions.value || []), props.chart.defDimension()]
  openDimensions.value.push(true)
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
  openMetrics.value.splice(i, 1)
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
