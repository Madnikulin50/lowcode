<template>
  <report-edit
    :report="editReport"
    :modules="modules"
    @update:report="v => editReport = v"
  >
    <template #dimension-options-options="{ dimension, isTemporal }">
      <div class="form-check">
        <input
          v-if="isTemporal && !['WEEK', 'QUARTER'].includes(dimension.modifier)"
          v-model="dimension.timeLabels"
          class="form-check-input"
          type="checkbox"
          :id="`timeLabels-${dimension.dimensionID}`"
        />
        <label
          v-if="isTemporal && !['WEEK', 'QUARTER'].includes(dimension.modifier)"
          class="form-check-label"
          :for="`timeLabels-${dimension.dimensionID}`"
        >
          {{ $t('edit.dimension.timeLabels') }}
        </label>
      </div>
    </template>

    <template #dimension-options="{ dimension }">
      <div class="row">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.dimension.rotate.label') }}
            </label>
            <input
              v-model="dimension.rotateLabel"
              type="number"
              class="form-control form-control-sm"
            />
            <div class="form-text">
              {{ $t('edit.dimension.rotate.description') }}
            </div>
          </div>
        </div>
      </div>
    </template>

    <template #y-axis="{ report }">
      <div class="px-3">
        <h5 class="mb-3">
          {{ $t('edit.yAxis.label') }}
        </h5>

        <div class="row">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('edit.yAxis.labelLabel') }}
              </label>
              <div class="input-group input-group-sm">
                <input
                  v-model="report.yAxis.label"
                  class="form-control form-control-sm"
                />
                <span class="input-group-text">
                  <chart-translator
                    :field="report.yAxis.label"
                    :chart="chart"
                    :disabled="isNew"
                    highlight-key="yAxis.label"
                    @update:field="v => report.yAxis.label = v"
                  />
                </span>
              </div>
            </div>
          </div>

          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('edit.yAxis.labelPosition.label') }}
              </label>
              <select
                v-model="report.yAxis.labelPosition"
                class="form-select form-select-sm"
              >
                <option
                  v-for="opt in axisLabelPositions"
                  :key="opt.value"
                  :value="opt.value"
                >
                  {{ opt.text }}
                </option>
              </select>
            </div>
          </div>
        </div>

        <div class="row">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('edit.yAxis.minLabel') }}
              </label>
              <input
                v-model="report.yAxis.min"
                type="number"
                class="form-control form-control-sm"
                :placeholder="$t('edit.yAxis.minPlaceholder')"
              />
            </div>
          </div>

          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('edit.yAxis.maxLabel') }}
              </label>
              <input
                v-model="report.yAxis.max"
                type="number"
                class="form-control form-control-sm"
                :placeholder="$t('edit.yAxis.maxPlaceholder')"
              />
            </div>
          </div>
        </div>

        <div class="row mb-2">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('edit.yAxis.rotate.label') }}
              </label>
              <input
                v-model="report.yAxis.rotateLabel"
                type="number"
                class="form-control form-control-sm"
              />
              <div class="form-text">
                {{ $t('edit.yAxis.rotate.description') }}
              </div>
            </div>
          </div>

          <div class="col-12 col-lg-6">
            <div class="mb-3 mb-0">
              <label class="form-label text-primary">
                {{ $t('edit.yAxis.options.label') }}
              </label>
              <div class="form-check">
                <input
                  v-model="report.yAxis.axisType"
                  class="form-check-input"
                  type="checkbox"
                  true-value="logarithmic"
                  false-value="linear"
                  :id="`axisType-${report.moduleID}`"
                />
                <label
                  class="form-check-label"
                  :for="`axisType-${report.moduleID}`"
                >
                  {{ $t('edit.yAxis.logarithmicScale') }}
                </label>
              </div>

              <div class="form-check">
                <input
                  v-model="report.yAxis.axisPosition"
                  class="form-check-input"
                  type="checkbox"
                  true-value="right"
                  false-value="left"
                  :id="`axisPosition-${report.moduleID}`"
                />
                <label
                  class="form-check-label"
                  :for="`axisPosition-${report.moduleID}`"
                >
                  {{ $t('edit.yAxis.axisOnRight') }}
                </label>
              </div>

              <div class="form-check">
                <input
                  v-model="report.yAxis.beginAtZero"
                  class="form-check-input"
                  type="checkbox"
                  :id="`beginAtZero-${report.moduleID}`"
                />
                <label
                  class="form-check-label"
                  :for="`beginAtZero-${report.moduleID}`"
                >
                  {{ $t('edit.yAxis.axisScaleFromZero') }}
                </label>
              </div>

              <div class="form-check">
                <input
                  v-model="report.yAxis.horizontal"
                  class="form-check-input"
                  type="checkbox"
                  :id="`horizontal-${report.moduleID}`"
                />
                <label
                  class="form-check-label"
                  :for="`horizontal-${report.moduleID}`"
                >
                  {{ $t('edit.yAxis.horizontal.label') }}
                </label>
              </div>
            </div>
          </div>
        </div>

        <hr>

        <div class="row">
          <div class="col-12 col-md-6">
            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('edit.formatting.prefix.label') }}
              </label>
              <input
                v-model="report.yAxis.formatting.prefix"
                class="form-control form-control-sm"
                :placeholder="$t('edit.formatting.prefix.placeholder')"
              />
            </div>
          </div>

          <div class="col-12 col-md-6">
            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('edit.formatting.suffix.label') }}
              </label>
              <input
                v-model="report.yAxis.formatting.suffix"
                class="form-control form-control-sm"
                :placeholder="$t('edit.formatting.suffix.placeholder')"
              />
            </div>
          </div>

          <div class="col-12 col-md-6">
            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('edit.formatting.presetFormats.label') }}
              </label>
              <select
                v-model="report.yAxis.formatting.presetFormat"
                class="form-select form-select-sm"
              >
                <option
                  v-for="opt in formatOptions"
                  :key="opt.value"
                  :value="opt.value"
                >
                  {{ opt.text }}
                </option>
              </select>
              <div
                v-if="report.yAxis.formatting.presetFormat"
                class="form-text"
                style="white-space: pre-line;"
              >
                {{ $t(`edit.formatting.presetFormats.description.${report.yAxis.formatting.presetFormat}`) }}
              </div>
            </div>
          </div>

          <div class="col-12 col-md-6">
            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('edit.formatting.format.label') }}
              </label>
              <input
                v-model="report.yAxis.formatting.format"
                class="form-control form-control-sm"
                :disabled="report.yAxis.formatting.presetFormat !== 'custom'"
                :placeholder="$t('edit.formatting.format.placeholder')"
              />
            </div>
          </div>
        </div>
      </div>
    </template>

    <template #metric-options="{ metric }">
      <div class="row">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.metric.labelLabel') }}
            </label>
            <div class="input-group input-group-sm">
              <input
                v-model="metric.label"
                class="form-control form-control-sm"
              />
              <span class="input-group-text">
                <chart-translator
                  :field="metric.label"
                  :chart="chart"
                  :disabled="isNew"
                  :highlight-key="`metrics.${metric.metricID}.label`"
                  @update:field="v => metric.label = v"
                />
              </span>
            </div>
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.metric.output.label') }}
            </label>
            <c-input-select
              v-model="metric.type"
              :options="chartTypes"
              label="text"
              :reduce="option => option.value"
              :get-option-key="option => option.text"
              :placeholder="$t('edit.metric.output.placeholder')"
              @input="value => chartTypeChanged(metric)"
            />
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.metric.fx.label') }}
            </label>
            <textarea
              v-model="metric.fx"
              class="form-control"
              placeholder="n"
            ></textarea>
            <div class="form-text">
              {{ $t('edit.metric.fx.description') }}
            </div>
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.metric.options.label') }}
            </label>
            <div class="form-check">
              <input
                v-model="metric.fixTooltips"
                class="form-check-input"
                type="checkbox"
                :id="`fixtooltips-${metric.metricID}`"
              />
              <label
                class="form-check-label"
                :for="`fixtooltips-${metric.metricID}`"
              >
                {{ $t('edit.metric.fixTooltips') }}
              </label>
            </div>

            <div class="form-check">
              <input
                v-if="hasRelativeDisplay(metric)"
                v-model="metric.relativeValue"
                class="form-check-input"
                type="checkbox"
                :id="`relative-${metric.metricID}`"
              />
              <label
                v-if="hasRelativeDisplay(metric)"
                class="form-check-label"
                :for="`relative-${metric.metricID}`"
              >
                {{ $t('edit.metric.relative') }}
              </label>
            </div>

            <div class="form-check">
              <input
                v-if="metric.type === 'pie'"
                v-model="metric.rose"
                class="form-check-input"
                type="checkbox"
                :id="`rose-${metric.metricID}`"
              />
              <label
                v-if="metric.type === 'pie'"
                class="form-check-label"
                :for="`rose-${metric.metricID}`"
              >
                {{ $t('edit.metric.rose') }}
              </label>
            </div>

            <div class="form-check">
              <input
                v-if="metric.type === 'line'"
                v-model="metric.fill"
                class="form-check-input"
                type="checkbox"
                :id="`fill-${metric.metricID}`"
              />
              <label
                v-if="metric.type === 'line'"
                class="form-check-label"
                :for="`fill-${metric.metricID}`"
              >
                {{ $t('edit.metric.fillArea') }}
              </label>
            </div>
          </div>

          <div
            v-if="metric.type === 'line'"
            class="mb-3"
          >
            <label class="form-label text-primary">
              {{ $t('edit.metric.lineStyle.label') }}
            </label>
            <div class="btn-group" data-bs-toggle="buttons">
              <label
                v-for="opt in lineStyleOptions"
                :key="opt.value"
                class="btn btn-outline-primary btn-sm"
                :class="{ active: getLineStyle(metric) === opt.value }"
              >
                <input
                  type="radio"
                  class="btn-check"
                  name="lineStyle"
                  autocomplete="off"
                  :value="opt.value"
                  :checked="getLineStyle(metric) === opt.value"
                  @change="setLineStyle($event, metric)"
                />
                {{ opt.text }}
              </label>
            </div>
          </div>
        </div>

        <div
          v-if="!hasRelativeDisplay(metric)"
          class="col-12 col-lg-6"
        >
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.metric.stack.label') }}
            </label>
            <input
              v-model="metric.stack"
              class="form-control form-control-sm"
            />
            <div class="form-text">
              {{ $t('edit.metric.stack.description') }}
            </div>
          </div>
        </div>

        <div
          v-if="metric.type === 'scatter'"
          class="col-12 col-lg-6"
        >
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.metric.symbol.label') }}
            </label>
            <select
              v-model="metric.symbol"
              class="form-select form-select-sm"
            >
              <option
                v-for="opt in scatterSymbolOptions"
                :key="opt.value"
                :value="opt.value"
              >
                {{ opt.text }}
              </option>
            </select>
          </div>
        </div>
      </div>

      <hr>

      <div class="row">
        <div class="col-12 col-md-6">
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.formatting.prefix.label') }}
            </label>
            <input
              v-model="metric.formatting.prefix"
              class="form-control form-control-sm"
              :placeholder="$t('edit.formatting.prefix.placeholder')"
            />
          </div>
        </div>

        <div class="col-12 col-md-6">
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.formatting.suffix.label') }}
            </label>
            <input
              v-model="metric.formatting.suffix"
              class="form-control form-control-sm"
              :placeholder="$t('edit.formatting.suffix.placeholder')"
            />
          </div>
        </div>

        <div class="col-12 col-md-6">
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.formatting.presetFormats.label') }}
            </label>
            <select
              v-model="metric.formatting.presetFormat"
              class="form-select form-select-sm"
            >
              <option
                v-for="opt in formatOptions"
                :key="opt.value"
                :value="opt.value"
              >
                {{ opt.text }}
              </option>
            </select>
            <div
              v-if="metric.formatting.presetFormat"
              class="form-text"
              style="white-space: pre-line;"
            >
              {{ $t(`edit.formatting.presetFormats.description.${metric.formatting.presetFormat}`) }}
            </div>
          </div>
        </div>

        <div class="col-12 col-md-6">
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.formatting.format.label') }}
            </label>
            <input
              v-model="metric.formatting.format"
              class="form-control form-control-sm"
              :disabled="metric.formatting.presetFormat !== 'custom'"
              :placeholder="$t('edit.formatting.format.placeholder')"
            />
          </div>
        </div>
      </div>
    </template>

    <template #additional-config="{ hasAxis, report }">
      <hr>
      <div class="px-3">
        <h5 class="d-flex mb-3">
          {{ $t('edit.additionalConfig.tooltip.label') }}
          <c-hint
            :tooltip="$t('edit.additionalConfig.tooltip.formatting.disclaimer')"
            icon-class="text-warning"
          />
        </h5>

        <div class="row">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('edit.additionalConfig.tooltip.formatting.label') }}
              </label>
              <input
                v-model="report.tooltip.formatting"
                class="form-control form-control-sm"
                :placeholder="$t('edit.additionalConfig.tooltip.formatting.placeholder')"
              />
              <div class="form-text">
                {{ $t('edit.additionalConfig.tooltip.formatting.description') }}
              </div>
            </div>
          </div>

          <div
            v-if="!hasAxis"
            class="col-12 col-lg-6"
          >
            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('edit.additionalConfig.tooltip.labelNextToChart') }}
              </label>
              <c-input-checkbox
                :value="!!report.tooltip.labelsNextToPartition"
                switch
                :labels="checkboxLabel"
                @input="report.tooltip.labelsNextToPartition = $event"
              />
            </div>
          </div>
        </div>
      </div>

      <hr>

      <div class="px-3 mb-2">
        <h5 class="mb-3">
          {{ $t('edit.additionalConfig.offset.label') }}
        </h5>

        <div class="row">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('edit.additionalConfig.offset.default') }}
              </label>
              <c-input-checkbox
                v-model="report.offset.isDefault"
                switch
                :labels="checkboxLabel"
                class="mb-3"
              />
            </div>
          </div>
        </div>

        <div
          v-if="!report.offset.isDefault"
          class="row"
        >
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('edit.additionalConfig.offset.position.top') }}
              </label>
              <input
                v-model="report.offset.top"
                class="form-control form-control-sm"
              />
            </div>
          </div>

          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('edit.additionalConfig.offset.position.right') }}
              </label>
              <input
                v-model="report.offset.right"
                class="form-control form-control-sm"
              />
            </div>
          </div>

          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('edit.additionalConfig.offset.position.bottom') }}
              </label>
              <input
                v-model="report.offset.bottom"
                class="form-control form-control-sm"
              />
            </div>
          </div>

          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('edit.additionalConfig.offset.position.left') }}
              </label>
              <input
                v-model="report.offset.left"
                class="form-control form-control-sm"
              />
            </div>
          </div>

          <div class="col-12">
            <small class="text-muted">
              {{ $t('edit.additionalConfig.offset.valueRange') }}
            </small>
          </div>
        </div>
      </div>

      <template v-if="hasAxis">
        <hr>

        <div class="px-3">
          <h5 class="mb-3">
            {{ $t('edit.additionalConfig.anomaly.label') }}
          </h5>

          <div class="row">
            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">
                  {{ $t('edit.additionalConfig.anomaly.enable') }}
                </label>
                <c-input-checkbox
                  v-model="report.anomaly.enabled"
                  switch
                  :labels="checkboxLabel"
                />
              </div>
            </div>
          </div>

          <template v-if="report.anomaly.enabled">
            <div class="row">
              <div class="col-12 col-lg-6">
                <div class="mb-3">
                  <label class="form-label text-primary">
                    {{ $t('edit.additionalConfig.anomaly.method') }}
                  </label>
                  <select
                    v-model="report.anomaly.method"
                    class="form-select form-select-sm"
                  >
                    <option
                      v-for="opt in anomalyMethods"
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
                    {{ $t('edit.additionalConfig.anomaly.threshold') }}
                  </label>
                  <input
                    v-model="report.anomaly.threshold"
                    type="number"
                    step="0.1"
                    class="form-control form-control-sm"
                  />
                </div>
              </div>
            </div>

            <div
              v-if="report.anomaly.method === 'fixed'"
              class="row"
            >
              <div class="col-12 col-lg-6">
                <div class="mb-3">
                  <label class="form-label text-primary">
                    {{ $t('edit.additionalConfig.anomaly.min') }}
                  </label>
                  <input
                    v-model="report.anomaly.min"
                    type="number"
                    class="form-control form-control-sm"
                  />
                </div>
              </div>

              <div class="col-12 col-lg-6">
                <div class="mb-3">
                  <label class="form-label text-primary">
                    {{ $t('edit.additionalConfig.anomaly.max') }}
                  </label>
                  <input
                    v-model="report.anomaly.max"
                    type="number"
                    class="form-control form-control-sm"
                  />
                </div>
              </div>
            </div>

            <div class="row">
              <div class="col-12 col-lg-6">
                <div class="mb-3">
                  <label class="form-label text-primary">
                    {{ $t('edit.additionalConfig.anomaly.color') }}
                  </label>
                  <input
                    v-model="report.anomaly.color"
                    type="color"
                    class="form-control form-control-sm color-picker"
                  />
                </div>
              </div>
            </div>
          </template>
        </div>
      </template>
    </template>
  </report-edit>
</template>

<script setup>
import { computed, ref, watch, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose } from 'corteza-lib/js/dist'
import ReportEdit from './ReportEdit.vue'
import ChartTranslator from 'corteza-webapp-compose/src/components/Chart/ChartTranslator.vue'

const { t } = useI18n()

const ignoredCharts = ['funnel', 'gauge', 'radar']

const props = defineProps({
  report: { type: Object, required: false, default: undefined },
  modules: { type: Array, required: true },
  chart: { type: Object, default: () => ({}) },
})

const emit = defineEmits(['update:report'])

const editReport = computed({
  get: () => props.report,
  set: (v) => emit('update.report', v),
})

const isNew = computed(() => props.chart?.chartID === compose.NoID)

const checkboxLabel = ref({
  on: t('label.yes'),
  off: t('label.no'),
})

const formatOptions = ref([
  { value: 'custom', text: t('edit.formatting.presetFormats.options.custom') },
  { value: 'accounting', text: t('edit.formatting.presetFormats.options.accounting') },
])

const chartTypes = ref(
  Object.values(compose.chartUtil.ChartType)
    .filter(v => !ignoredCharts.includes(v))
    .map(value => ({ value, text: t(`edit.metric.output.${value}`) }))
)

const legendPositions = ref([
  { value: 'top', text: t('edit.metric.legend.top') },
  { value: 'left', text: t('edit.metric.legend.left') },
  { value: 'bottom', text: t('edit.metric.legend.bottom') },
  { value: 'right', text: t('edit.metric.legend.right') },
])

const axisLabelPositions = ref([
  { value: 'end', text: t('edit.yAxis.labelPosition.top') },
  { value: 'center', text: t('edit.yAxis.labelPosition.center') },
  { value: 'start', text: t('edit.yAxis.labelPosition.bottom') },
])

const tensionSteps = ref([
  { text: t('edit.metric.lineTension.straight'), value: 0.0 },
  { text: t('edit.metric.lineTension.slight'), value: 0.2 },
  { text: t('edit.metric.lineTension.medium'), value: 0.4 },
  { text: t('edit.metric.lineTension.curvy'), value: 0.6 },
])

const lineStyleOptions = ref([
  { value: '', text: t('edit.metric.lineStyle.default') },
  { value: 'smooth', text: t('edit.metric.lineStyle.smooth') },
  { value: 'step', text: t('edit.metric.lineStyle.step') },
])

const scatterSymbolOptions = ref([
  { value: 'circle', text: t('edit.metric.symbol.circle') },
  { value: 'triangle', text: t('edit.metric.symbol.triangle') },
  { value: 'diamond', text: t('edit.metric.symbol.diamond') },
  { value: 'pin', text: t('edit.metric.symbol.pin') },
  { value: 'arrow', text: t('edit.metric.symbol.arrow') },
  { value: 'rect', text: t('edit.metric.symbol.rect') },
  { value: 'roundRect', text: t('edit.metric.symbol.roundRect') },
])

const anomalyMethods = ref([
  { value: 'zscore', text: t('edit.additionalConfig.anomaly.methodZscore') },
  { value: 'iqr', text: t('edit.additionalConfig.anomaly.methodIqr') },
  { value: 'fixed', text: t('edit.additionalConfig.anomaly.methodFixed') },
  { value: 'pct_change', text: t('edit.additionalConfig.anomaly.methodPctChange') },
])

watch(() => props.report, (r) => {
  if (r && !r.anomaly) {
    r.anomaly = {
      enabled: false,
      method: 'zscore',
      threshold: 2,
      min: undefined,
      max: undefined,
      color: '',
    }
  }
}, { immediate: true })

onBeforeUnmount(() => {
  chartTypes.value = []
  legendPositions.value = []
  axisLabelPositions.value = []
  tensionSteps.value = []
  lineStyleOptions.value = []
  scatterSymbolOptions.value = []
})

function hasRelativeDisplay (metric) {
  return compose.chartUtil.hasRelativeDisplay(metric)
}

function getLineStyle (metric) {
  if (metric.smooth) return 'smooth'
  else if (metric.step) return 'step'
  return ''
}

function setLineStyle (style, metric) {
  metric.smooth = style === 'smooth'
  metric.step = style === 'step'
}

function chartTypeChanged (metric) {
  metric.relativeValue = false
}
</script>

<style lang="scss" scoped>
.color-picker {
  max-width: 50px;
}
</style>
