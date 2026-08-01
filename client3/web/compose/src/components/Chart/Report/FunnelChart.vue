<template>
  <report-edit
    :report="editReport"
    :modules="modules"
    :supported-metrics="1"
    @update:report="v => editReport = v"
  >
    <template #dimension-options="{ index, dimension, field }">
      <div class="mb-3">
        <label
          v-if="showPicker(field)"
          class="form-label text-primary"
        >
          {{ $t('edit.dimension.options.label') }}
        </label>
        <c-item-picker
          v-if="showPicker(field)"
          :value="getOptions(dimension)"
          :options="field.options.options"
          :labels="{
            searchPlaceholder: $t('edit.dimension.optionsPicker.searchPlaceholder'),
            availableItems: $t('edit.dimension.optionsPicker.availableItems'),
            selectAllItems: $t('edit.dimension.optionsPicker.selectAllItems'),
            selectedItems: $t('edit.dimension.optionsPicker.selectedItems'),
            unselectAllItems: $t('edit.dimension.optionsPicker.unselectAllItems'),
            noItemsFound: $t('edit.dimension.optionsPicker.noItemsFound'),
          }"
          class="d-flex flex-column"
          style="height: 100% !important;"
          @update:value="setOptions(index, field, $event)"
        />
      </div>
    </template>

    <template #metric-options="{ metric }">
      <div class="row">
        <div class="col-12 col-lg-6 offset-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">
              {{ $t('edit.metric.options.label') }}
            </label>
            <div class="form-check">
              <input
                v-model="metric.cumulative"
                class="form-check-input"
                type="checkbox"
                :id="`cumulative-${metric.metricID}`"
              />
              <label
                class="form-check-label"
                :for="`cumulative-${metric.metricID}`"
              >
                {{ $t('edit.metric.cumulative') }}
              </label>
            </div>

            <div class="form-check">
              <input
                v-model="metric.relativeValue"
                class="form-check-input"
                type="checkbox"
                :id="`relative-${metric.metricID}`"
              />
              <label
                class="form-check-label"
                :for="`relative-${metric.metricID}`"
              >
                {{ $t('edit.metric.relative') }}
              </label>
            </div>

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
  </report-edit>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'
import ReportEdit from './ReportEdit.vue'
const { CItemPicker } = components

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

const formatOptions = ref([
  { value: 'custom', text: t('edit.formatting.presetFormats.options.custom') },
  { value: 'accounting', text: t('edit.formatting.presetFormats.options.accounting') },
])

function showPicker (field) {
  return field && field.kind === 'Select' && field.options.options
}

function getOptions ({ meta = {} }) {
  const { fields = [] } = meta
  return fields.map(({ value }) => value)
}

function setOptions (index, field, fields) {
  editReport.value.dimensions[index].meta.fields = fields.map(f => {
    const { options = [] } = field.options || {}
    return options.find(({ value }) => value === f)
  })
}
</script>

<style scoped lang="scss">
.cursor-pointer {
  cursor: pointer;
}
</style>
