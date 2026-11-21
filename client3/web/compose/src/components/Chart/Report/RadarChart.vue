<template>
  <report-edit
    :report="editReport"
    :modules="modules"
    @update:report="v => editReport = v"
  >
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
              class="form-select form-control form-select-sm"
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

    <template #dimension-options-options="{ dimension }">
      <div class="form-check">
        <input
          v-model="dimension.fixTooltips"
          class="form-check-input"
          type="checkbox"
          :id="`fixtooltips-${dimension.dimensionID}`"
        />
        <label
          class="form-check-label"
          :for="`fixtooltips-${dimension.dimensionID}`"
        >
          {{ $t('edit.metric.fixTooltips') }}
        </label>
      </div>
    </template>

    <template #dimension-options="{ dimension }">
      <div class="row">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.metric.radar.shape.label') }}
            </label>
            <select
              v-model="dimension.shape"
              class="form-select form-control form-select-sm"
            >
              <option
                v-for="opt in shapeOptions"
                :key="opt.value"
                :value="opt.value"
              >
                {{ opt.text }}
              </option>
            </select>
          </div>
        </div>
      </div>
    </template>
  </report-edit>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose } from 'corteza-lib/js/dist'
import ReportEdit from './ReportEdit.vue'
import ChartTranslator from 'corteza-webapp-compose/src/components/Chart/ChartTranslator.vue'

const { t } = useI18n()

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

const formatOptions = ref([
  { value: 'custom', text: t('edit.formatting.presetFormats.options.custom') },
  { value: 'accounting', text: t('edit.formatting.presetFormats.options.accounting') },
])

const shapeOptions = ref([
  { value: 'polygon', text: t('edit.metric.radar.shape.polygon') },
  { value: 'circle', text: t('edit.metric.radar.shape.circle') },
])
</script>

<style scoped lang="scss">
.cursor-pointer {
  cursor: pointer;
}
</style>
