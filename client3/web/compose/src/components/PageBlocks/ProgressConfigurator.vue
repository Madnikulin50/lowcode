<template>
  <div class="vh-100">
    <h5 class="text-primary">{{ $t('progress.value.label') }}</h5>
    <div class="row">
      <div v-if="!options.value.moduleID" class="col-12">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('progress.value.default.label') }}</label>
          <small class="text-muted d-block mb-1">{{ $t('progress.value.default.description') }}</small>
          <input v-model="options.value.default" type="number" class="form-control" />
        </div>
      </div>
      <div class="col-12">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('progress.module.label') }}</label>
          <c-input-select v-model="options.value.moduleID" label="name" :placeholder="$t('progress.module.select')" :options="modules" :get-option-key="m => m.moduleID" :reduce="m => m.moduleID" />
        </div>
      </div>
      <template v-if="options.value.moduleID">
        <div class="col-12">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('metric.edit.filterLabel') }}</label>
            <c-input-expression v-model="options.value.filter" min-height="3.688rem" placeholder="(A > B) OR (A < C)" class="mb-1" :suggestion-params="recordAutoCompleteParams.value" />
            <small class="text-muted d-block">
              <code>${record.values.fieldName}</code> <code>${recordID}</code> <code>${ownerID}</code>
              <code>${userID}</code>, <code>${user.name}</code>
            </small>
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('progress.field.label') }}</label>
            <c-input-select v-model="options.value.field" :placeholder="$t('progress.field.select')" :options="valueModuleFields" :get-option-key="f => f.name" :get-option-label="f => f.label || f.name" :reduce="f => f.name" @input="fieldChanged($event, options.value)" />
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('progress.aggregate.label') }}</label>
            <c-input-select v-model="options.value.operation" label="name" :disabled="!options.value.field || options.value.field === 'count'" :placeholder="$t('progress.aggregate.select')" :options="aggregationOperations" :get-option-key="a => a.operation" :reduce="a => a.operation" />
          </div>
        </div>
      </template>
    </div>

    <hr />

    <h5 class="text-primary">{{ $t('progress.value.min') }}</h5>
    <div class="row">
      <div v-if="!options.minValue.moduleID" class="col-12">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('progress.value.default.label') }}</label>
          <small class="text-muted d-block mb-1">{{ $t('progress.value.default.description') }}</small>
          <input v-model="options.minValue.default" type="number" class="form-control" />
        </div>
      </div>
      <div class="col-12">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('progress.module.label') }}</label>
          <c-input-select v-model="options.minValue.moduleID" label="name" :placeholder="$t('progress.module.select')" :options="modules" :get-option-key="m => m.moduleID" :reduce="m => m.moduleID" />
        </div>
      </div>
      <template v-if="options.minValue.moduleID">
        <div class="col-12">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('metric.edit.filterLabel') }}</label>
            <c-input-expression v-model="options.minValue.filter" min-height="3.688rem" placeholder="(A > B) OR (A < C)" class="mb-1" :suggestion-params="recordAutoCompleteParams.min" />
            <small class="text-muted d-block">
              <code>${record.values.fieldName}</code> <code>${recordID}</code> <code>${ownerID}</code>
              <code>${userID}</code>, <code>${user.name}</code>
            </small>
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('progress.field.label') }}</label>
            <c-input-select v-model="options.minValue.field" :placeholder="$t('progress.field.select')" :options="minValueModuleFields" :get-option-key="f => f.name" :get-option-label="f => f.label || f.name" :reduce="f => f.name" @input="fieldChanged($event, options.minValue)" />
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('progress.aggregate.label') }}</label>
            <c-input-select v-model="options.minValue.operation" label="name" :disabled="!options.minValue.field || options.minValue.field === 'count'" :placeholder="$t('progress.aggregate.select')" :options="aggregationOperations" :get-option-key="a => a.operation" :reduce="a => a.operation" />
          </div>
        </div>
      </template>
    </div>

    <hr />

    <h5 class="text-primary">{{ $t('progress.value.max') }}</h5>
    <div class="row">
      <div v-if="!options.maxValue.moduleID" class="col-12">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('progress.value.default.label') }}</label>
          <small class="text-muted d-block mb-1">{{ $t('progress.value.default.description') }}</small>
          <input v-model="options.maxValue.default" type="number" class="form-control" />
        </div>
      </div>
      <div class="col-12">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('progress.module.label') }}</label>
          <c-input-select v-model="options.maxValue.moduleID" label="name" :placeholder="$t('progress.module.select')" :options="modules" :get-option-key="m => m.moduleID" :reduce="m => m.moduleID" />
        </div>
      </div>
      <template v-if="options.maxValue.moduleID">
        <div class="col-12">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('metric.edit.filterLabel') }}</label>
            <c-input-expression v-model="options.maxValue.filter" min-height="3.688rem" placeholder="(A > B) OR (A < C)" class="mb-1" :suggestion-params="recordAutoCompleteParams.max" />
            <small class="text-muted d-block">
              <code>${record.values.fieldName}</code> <code>${recordID}</code> <code>${ownerID}</code>
              <code>${userID}</code>, <code>${user.name}</code>
            </small>
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('progress.field.label') }}</label>
            <c-input-select v-model="options.maxValue.field" :placeholder="$t('progress.field.select')" :options="maxValueModuleFields" :get-option-key="f => f.name" :get-option-label="f => f.label || f.name" :reduce="f => f.name" @input="fieldChanged($event, options.maxValue)" />
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('progress.aggregate.label') }}</label>
            <c-input-select v-model="options.maxValue.operation" label="name" :disabled="!options.maxValue.field || options.maxValue.field === 'count'" :placeholder="$t('progress.aggregate.select')" :options="aggregationOperations" :get-option-key="a => a.operation" :reduce="a => a.operation" />
          </div>
        </div>
      </template>
    </div>

    <hr />

    <h5 class="text-primary">{{ $t('progress.display-options') }}</h5>
    <div class="row align-items-center">
      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('progress.default-variant') }}</label>
          <c-input-select v-model="options.display.variant" :options="variants" label="text" :reduce="v => v.value" />
        </div>
      </div>
      <div class="col-6 col-sm-3 mb-2 mb-sm-0">
        <div class="form-check form-switch mb-2">
          <input v-model="options.display.showValue" class="form-check-input" type="checkbox" role="switch" id="showValue" />
          <label class="form-check-label" for="showValue">{{ $t('progress.show.value') }}</label>
        </div>
        <div class="form-check form-switch">
          <input v-model="options.display.animated" class="form-check-input" type="checkbox" role="switch" id="animated" />
          <label class="form-check-label" for="animated">{{ $t('progress.animated') }}</label>
        </div>
      </div>
      <div class="col-6 col-sm-3 mb-2 mb-sm-0">
        <template v-if="options.display.showValue">
          <div class="form-check form-switch mb-2">
            <input v-model="options.display.showRelative" class="form-check-input" type="checkbox" role="switch" id="showRelative" />
            <label class="form-check-label" for="showRelative">{{ $t('progress.show.relative') }}</label>
          </div>
          <div class="form-check form-switch">
            <input v-model="options.display.showProgress" class="form-check-input" type="checkbox" role="switch" id="showProgress" />
            <label class="form-check-label" for="showProgress">{{ $t('progress.show.progress') }}</label>
          </div>
        </template>
      </div>
      <div class="col-12">
        <div class="mb-3">
          <label class="form-label">
            <div class="d-flex align-items-center">
              {{ $t('progress.thresholds') }}
              <button class="btn btn-link text-decoration-none ms-1" @click="addThreshold()">{{ $t('progress.add') }}</button>
            </div>
            <small class="text-muted d-block">{{ $t('progress.threshold.variant') }}</small>
          </label>
          <div v-for="(t, i) in options.display.thresholds" :key="i" class="row align-items-center mt-2">
            <div class="col">
              <div class="input-group">
                <input v-model="t.value" :placeholder="$t('progress.threshold.label')" type="number" class="form-control" />
                <span class="input-group-text">%</span>
              </div>
            </div>
            <div class="col d-flex align-items-center justify-content-center">
              <select v-model="t.variant" class="form-select">
                <option v-for="v in variants" :key="v.value" :value="v.value">{{ v.text }}</option>
              </select>
              <font-awesome-icon :icon="['fas', 'times']" class="pointer text-danger ms-3" @click="removeThreshold(i)" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <hr />

    <h6 class="text-primary">{{ $t('progress.preview') }}</h6>
    <div class="row">
      <div class="col-12">
        <FieldViewer value-only v-bind="mock" class="mb-2" />
      </div>
      <div class="col-12 col-sm-4">
        <label class="form-label">{{ $t('progress.value.label') }}</label>
        <input v-model="mock.record.values.mockField" :placeholder="$t('progress.value.label')" size="sm" type="number" class="form-control form-control-sm" />
      </div>
      <div class="col-12 col-sm-4">
        <label class="form-label">{{ $t('progress.value.min') }}</label>
        <input v-model="mock.field.options.min" :placeholder="$t('progress.value.min')" size="sm" type="number" class="form-control form-control-sm" />
      </div>
      <div class="col-12 col-sm-4">
        <label class="form-label">{{ $t('progress.value.max') }}</label>
        <input v-model="mock.field.options.max" :placeholder="$t('progress.value.max')" size="sm" type="number" class="form-control form-control-sm" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onBeforeUnmount, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { useStore } from '../../store'
import { compose, validator } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'
import FieldViewer from '../ModuleFields/Viewer'

const { CInputExpression } = components
const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  block: { type: Object, required: true },
  module: { type: Object, required: false },
  namespace: { type: Object, required: true },
})

const store = useStore()

const options = computed(() => props.block.options)
const modules = computed(() => store.module.set)
const moduleByID = computed(() => store.module.getByID)

const aggregationOperations = ref([
  { name: $t('metric.edit.operationSum'), operation: 'sum' },
  { name: $t('metric.edit.operationMax'), operation: 'max' },
  { name: $t('metric.edit.operationMin'), operation: 'min' },
  { name: $t('metric.edit.operationAvg'), operation: 'avg' },
  { label: $t('metric.edit.operationUniqueCount'), operation: 'uniqueCount' },
])

const variants = ref([
  { text: $t('progress.variant.primary'), value: 'primary' },
  { text: $t('progress.variant.secondary'), value: 'secondary' },
  { text: $t('progress.variant.success'), value: 'success' },
  { text: $t('progress.variant.warning'), value: 'warning' },
  { text: $t('progress.variant.danger'), value: 'danger' },
  { text: $t('progress.variant.info'), value: 'info' },
  { text: $t('progress.variant.light'), value: 'light' },
  { text: $t('progress.variant.dark'), value: 'dark' },
])

const mock = ref({
  namespace: undefined,
  module: undefined,
  field: undefined,
  record: undefined,
  errors: new validator.Validated(),
})

const sharedModuleFields = computed(() => [{ name: 'count', label: $t('progress.count') }])

function returnValueModuleFields(moduleID) {
  const m = moduleByID.value(moduleID)
  if (!m) return []
  return [
    ...sharedModuleFields.value,
    ...m.fields.filter(f => f.kind === 'Number').sort((a, b) => a.label.localeCompare(b.label)),
  ]
}

const valueModuleFields = computed(() => returnValueModuleFields(options.value.value?.moduleID))
const minValueModuleFields = computed(() => returnValueModuleFields(options.value.minValue?.moduleID))
const maxValueModuleFields = computed(() => returnValueModuleFields(options.value.maxValue?.moduleID))

const valueModule = computed(() => moduleByID.value(options.value.value?.moduleID))
const minValueModule = computed(() => moduleByID.value(options.value.minValue?.moduleID))
const maxValueModule = computed(() => moduleByID.value(options.value.maxValue?.moduleID))

const recordAutoCompleteParams = computed(() => ({
  value: typeof window.__composeAPI?.processRecordAutoCompleteParams === 'function' ? window.__composeAPI.processRecordAutoCompleteParams({ module: valueModule.value, operators: true }) : {},
  min: typeof window.__composeAPI?.processRecordAutoCompleteParams === 'function' ? window.__composeAPI.processRecordAutoCompleteParams({ module: minValueModule.value, operators: true }) : {},
  max: typeof window.__composeAPI?.processRecordAutoCompleteParams === 'function' ? window.__composeAPI.processRecordAutoCompleteParams({ module: maxValueModule.value, operators: true }) : {},
}))

watch(() => options.value, ({ display = {} }) => {
  if (mock.value.field) {
    mock.value.field.options = { ...mock.value.field.options, ...display }
  }
}, { deep: true })

// Initialize mock
mock.value.namespace = props.namespace
mock.value.field = compose.ModuleFieldMaker({ kind: 'Number' })
mock.value.field.apply({ name: 'mockField' })
mock.value.field.options.display = 'progress'
mock.value.field.options = { display: 'progress', ...mock.value.field.options, ...(options.value.display || {}) }
mock.value.module = new compose.Module({ fields: [mock.value.field] }, props.namespace)
mock.value.record = new compose.Record(mock.value.module, { mockField: 15 })

onBeforeUnmount(() => { setDefaultValues() })

function addThreshold() {
  if (!options.value.display.thresholds) options.value.display.thresholds = []
  options.value.display.thresholds.push({ value: 0, variant: 'success' })
}

function removeThreshold(index) {
  if (index > -1) options.value.display.thresholds.splice(index, 1)
}

function fieldChanged(value, optionsType) {
  if (!value || value === 'count') optionsType.operation = ''
}

function setDefaultValues() {
  aggregationOperations.value = []
  variants.value = []
  mock.value = {}
}
</script>
<style lang="scss" scoped>
.preview { bottom: 0; left: 0; z-index: 2; width: 100%; box-shadow: 0 -0.25rem 1rem rgb(0 0 0 / 15%); }
</style>
