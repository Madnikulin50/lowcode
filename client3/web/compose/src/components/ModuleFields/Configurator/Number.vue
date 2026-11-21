<template>
  <div>
    <div class="row">
      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary pr-2">{{ t('kind.number.displayType.label') }}</label>
          <div class="btn-group" data-bs-toggle="buttons">
            <label
              v-for="opt in displayOptions"
              :key="opt.value"
              class="btn btn-outline-primary btn-sm"
              :class="{ active: f.options.display === opt.value }"
            >
              <input
                type="radio"
                class="btn-check"
                :value="opt.value"
                :checked="f.options.display === opt.value"
                autocomplete="off"
                @change="f.options.display = opt.value"
              />
              {{ opt.text }}
            </label>
          </div>
        </div>
      </div>
    </div>

    <div class="row">
      <div class="col-12 col-lg-6">
        <div class="mb-3" :title="hasData ? t('not-configurable') : t('kind.number.precisionTooltip')">
          <label class="form-label mb-2 text-primary">{{ t('kind.number.precisionLabel') }} {{ f.options.precision }}</label>
          <input
            v-model="f.options.precision"
            type="range"
            class="form-range"
            min="0"
            max="6"
            :disabled="hasData"
          />
        </div>
      </div>
      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ t('kind.number.step.label') }}</label>
          <div class="form-text">{{ t('kind.number.step.description') }}</div>
          <input v-model="f.options.step" type="number" class="form-control form-control-sm" />
        </div>
      </div>
    </div>

    <hr />

    <div class="row">
      <template v-if="f.options.display === 'number' || f.options.display === 'colorGrade' || f.options.display === 'trafficLight'">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('kind.number.prefixLabel') }}</label>
            <input v-model="f.options.prefix" type="text" class="form-control form-control-sm" placeholder="USD/mo" />
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('kind.number.suffixLabel') }}</label>
            <input v-model="f.options.suffix" type="text" class="form-control form-control-sm" placeholder="$" />
          </div>
        </div>
      </template>

      <template v-if="f.options.display === 'number' || f.options.display === 'colorGrade' || f.options.display === 'trafficLight'">
        <div class="col-12 col-lg-6">
          <div class="mb-3" style="white-space: pre-line;">
            <label class="form-label text-primary">{{ t('kind.number.presetFormats.label') }}</label>
            <div class="form-text">{{ formattedOptionsDescription }}</div>
            <select v-model="f.options.presetFormat" class="form-select form-control form-select-sm">
              <option v-for="opt in formatOptions" :key="opt.value" :value="opt.value">{{ opt.text }}</option>
            </select>
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('kind.number.formatLabel') }}</label>
            <input
              v-model="f.options.format"
              type="text"
              class="form-control form-control-sm"
              :disabled="f.options.presetFormat !== 'custom'"
              placeholder="0.00"
            />
          </div>
        </div>
        <div class="col-12">
          <div v-if="f.options.display === 'number' || f.options.display === 'colorGrade' || f.options.display === 'trafficLight'" class="mb-3">
            <label class="form-label text-primary">{{ t('kind.number.examplesLabel') }}</label>
            <table class="table table-sm w-100">
              <thead>
                <tr>
                  <th id="example-input">{{ t('kind.number.exampleInput') }}</th>
                  <th id="example-format">{{ t('kind.number.exampleFormat') }}</th>
                  <th id="example-result">{{ t('kind.number.exampleResult') }}</th>
                </tr>
              </thead>
              <tr><td>1000.234</td><td>0,0.00</td><td>1,000.23</td></tr>
              <tr><td>1000.234</td><td>0,0</td><td>1,000</td></tr>
              <tr><td>0.974878234</td><td>0.000%</td><td>97.488%</td></tr>
              <tr><td>100</td><td>0o</td><td>100th</td></tr>
              <tr><td>238</td><td>00:00:00</td><td>0:03:58</td></tr>
            </table>
          </div>
        </div>
      </template>

      <template v-if="f.options.display === 'progress' || f.options.display === 'colorGrade' || f.options.display === 'trafficLight'">
        <div v-if="f.options.display === 'progress'" class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('kind.number.progress.minimumValue') }}</label>
            <input v-model="f.options.min" type="number" class="form-control form-control-sm" />
          </div>
        </div>
        <div v-if="f.options.display === 'progress'" class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('kind.number.progress.maximumValue') }}</label>
            <input v-model="f.options.max" type="number" class="form-control form-control-sm" />
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('kind.number.progress.variants.default') }}</label>
            <select v-model="f.options.variant" class="form-select form-control form-select-sm">
              <option v-for="opt in variants" :key="opt.value" :value="opt.value">{{ opt.text }}</option>
            </select>
          </div>
        </div>
        <div v-if="f.options.display === 'progress'" class="col-12 col-sm-3">
          <div class="mb-0">
            <div class="form-check mb-2">
              <input id="showValue" v-model="f.options.showValue" type="checkbox" class="form-check-input" />
              <label class="form-check-label" for="showValue">{{ t('kind.number.progress.show.value') }}</label>
            </div>
            <div class="form-check">
              <input id="animated" v-model="f.options.animated" type="checkbox" class="form-check-input" />
              <label class="form-check-label" for="animated">{{ t('kind.number.progress.animated') }}</label>
            </div>
          </div>
        </div>
        <div v-if="f.options.display === 'progress'" class="col-12 col-sm-3">
          <div v-if="f.options.showValue" class="mb-0">
            <div class="form-check mb-2">
              <input id="showRelative" v-model="f.options.showRelative" type="checkbox" class="form-check-input" />
              <label class="form-check-label" for="showRelative">{{ t('kind.number.progress.show.relative') }}</label>
            </div>
            <div class="form-check">
              <input id="showProgress" v-model="f.options.showProgress" type="checkbox" class="form-check-input" />
              <label class="form-check-label" for="showProgress">{{ t('kind.number.progress.show.progress') }}</label>
            </div>
          </div>
        </div>
        <div class="col-12">
          <div class="mb-3">

              <div class="d-flex align-items-center text-primary">
                {{ t('kind.number.progress.thresholds.label') }}
                <button class="btn btn-link btn-sm text-decoration-none ms-1" @click="addThreshold">{{ t('label.add-with-plus') }}</button>
              </div>
              <small class="text-muted">{{ t('kind.number.progress.thresholds.description') }}</small>

            <div v-for="(t, i) in field.options.thresholds" :key="i" class="row align-items-center" :class="{ 'mt-2': i }">
              <div class="col">
                <div class="input-group input-group-sm">
                  <input v-model="t.value" type="number" class="form-control form-control-sm" placeholder="Threshold" />
                  <span v-if="f.options.display === 'progress'" class="input-group-text">%</span>
                </div>
              </div>
              <div class="col d-flex align-items-center justify-content-center">
                <select v-model="t.variant" class="form-select form-control form-select-sm">
                  <option v-for="opt in variants" :key="opt.value" :value="opt.value">{{ opt.text }}</option>
                </select>
                <font-awesome-icon :icon="['fas', 'times']" class="pointer text-danger ms-3" @click="removeThreshold(i)" />
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>

    <hr />

    <div class="mb-3 w-100">
      <label class="form-label text-primary">{{ t('kind.number.liveExample') }}</label>
      <div class="row align-items-center">
        <div class="col-12 col-lg-6">
          <input v-model="liveExample" type="number" class="form-control form-control-sm" :step="f.options.step" />
        </div>
        <div class="col-12 col-lg-6">
          <FieldViewer value-only v-bind="mock" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'field' } })
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useConfiguratorBase } from './base'
import FieldViewer from '../Viewer'
import { compose, validator } from 'corteza-lib/js/dist'

const props = defineProps({
  namespace: { type: Object, required: true },
  module: { type: Object, required: true },
  field: { type: Object, required: true },
  hasRecords: { type: Boolean, default: false },
})

const emit = defineEmits(['update:field'])

const { t } = useI18n({ useScope: 'global', messages: {} })
const { f, isNew, hasData } = useConfiguratorBase(props, emit)

const liveExample = ref(undefined)

const displayOptions = computed(() => [
  { text: t('kind.number.displayType.number'), value: 'number' },
  { text: t('kind.number.displayType.progress'), value: 'progress' },
  { text: t('kind.number.displayType.colorGrade'), value: 'colorGrade' },
  { text: t('kind.number.displayType.trafficLight'), value: 'trafficLight' },
])

const variants = computed(() => [
  { text: t('kind.number.progress.variants.primary'), value: 'primary' },
  { text: t('kind.number.progress.variants.secondary'), value: 'secondary' },
  { text: t('kind.number.progress.variants.success'), value: 'success' },
  { text: t('kind.number.progress.variants.warning'), value: 'warning' },
  { text: t('kind.number.progress.variants.danger'), value: 'danger' },
  { text: t('kind.number.progress.variants.info'), value: 'info' },
  { text: t('kind.number.progress.variants.light'), value: 'light' },
  { text: t('kind.number.progress.variants.dark'), value: 'dark' },
])

const formatOptions = computed(() => [
  { value: 'custom', text: t('kind.number.presetFormats.options.custom') },
  { value: 'accounting', text: t('kind.number.presetFormats.options.accounting') },
])

const mock = ref({
  namespace: undefined,
  module: undefined,
  field: undefined,
  record: undefined,
  errors: new validator.Validated(),
})

const formattedOptionsDescription = computed(() => {
  return t(`kind.number.presetFormats.description.${f.value.options.presetFormat}`)
})

watch(() => f.value.options.display, (display) => {
  liveExample.value = display === 'number' ? 1234.56789 : 33.45679
})

watch(() => f.value.options, (options) => {
  if (mock.value.field) {
    mock.value.field.options = options
    mock.value.record.values.mockField = Number(liveExample.value).toFixed(f.value.options.precision)
  }
}, { deep: true })

watch(liveExample, (value) => {
  if (mock.value.field) {
    value = Number(value).toFixed(f.value.options.precision)
    mock.value.record.values.mockField = value
  }
})

mock.value.namespace = props.namespace
mock.value.field = compose.ModuleFieldMaker(f.value)
mock.value.field.isMulti = false
mock.value.field.apply({ name: 'mockField' })
mock.value.module = new compose.Module({ fields: [mock.value.field] }, props.namespace)
mock.value.record = new compose.Record(mock.value.module, { })
liveExample.value = f.value.options.display === 'number' ? 1234.56789 : 33.45679

onBeforeUnmount(() => {
  setDefaultValues()
})

function addThreshold () {
  f.value.options.thresholds.push({ value: 0, variant: 'success' })
}

function removeThreshold (index) {
  if (index > -1) {
    f.value.options.thresholds.splice(index, 1)
  }
}

function setDefaultValues () {
  liveExample.value = undefined
  mock.value = {}
}
</script>
