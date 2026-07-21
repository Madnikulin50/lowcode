<template>
  <report-edit
    :report="editReport"
    :modules="modules"
    :supported-metrics="1"
    :uses-dimensions-field="false"
    un-skippable
    @update:report="v => editReport = v"
  >
    <template #dimension-options="{ dimension }">
      <div class="mb-3">
        <label class="form-label text-primary">
          {{ $t('edit.dimension.gaugeSteps') }}
        </label>
        <div
          v-for="(step, i) in dimension.meta.steps"
          :key="i"
          class="input-group input-group-sm mb-1"
        >
          <input
            v-model="step.label"
            class="form-control form-control-sm w-50"
            :placeholder="$t('label.title')"
          />
          <span class="input-group-text">
            <chart-translator
              :field="step.label"
              :chart="chart"
              :disabled="isNew"
              :highlight-key="`dimensions.${dimension.dimensionID}.meta.steps.${step.stepID}.label`"
              @update:field="v => step.label = v"
            />
          </span>

          <input
            v-model="step.value"
            type="number"
            class="form-control form-control-sm text-right w-25"
            :placeholder="$t('value')"
          />

          <span class="input-group-text">
            <c-input-confirm
              show-icon
              @confirmed="dimension.meta.steps.splice(i, 1)"
            />
          </span>
        </div>

        <button
          type="button"
          class="btn btn-link p-0"
          @click="dimension.meta.steps.push({ label: undefined, color: undefined, value: undefined })"
        >
          + {{ $t('label.add') }}
        </button>
      </div>
    </template>

    <template #metric-options="{ metric }">
      <div class="row">
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
          </div>
        </div>
      </div>

      <div class="row">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.metric.angle.start') }}
            </label>
            <input
              v-model="metric.startAngle"
              type="number"
              class="form-control form-control-sm"
            />
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.metric.angle.end') }}
            </label>
            <input
              v-model="metric.endAngle"
              type="number"
              class="form-control form-control-sm"
            />
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
  </report-edit>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose, NoID } from 'corteza-lib/js/dist'
import ReportEdit from './ReportEdit.vue'
import ChartTranslator from 'corteza-webapp-compose/src/components/Chart/ChartTranslator.vue'

const { t } = useI18n()

const props = defineProps({
  report: { type: Object, required: false, default: undefined },
  modules: { type: Array, required: true },
  chart: { type: compose.Chart, required: true },
})

const emit = defineEmits(['update:report'])

const editReport = computed({
  get: () => props.report,
  set: (v) => emit('update.report', v),
})

const isNew = computed(() => props.chart.chartID === NoID)

const checkboxLabel = ref({
  on: t('label.yes'),
  off: t('label.no'),
})

const formatOptions = ref([
  { value: 'custom', text: t('edit.formatting.presetFormats.options.custom') },
  { value: 'accounting', text: t('edit.formatting.presetFormats.options.accounting') },
])
</script>
