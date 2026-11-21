<template>
  <div v-if="step.link">
    <div class="row">
      <div class="col">
        <div class="mb-3">
          <label class="text-primary form-label">{{ t('datasources.name-required') }}</label>
          <input
            v-model="step.link.name"
            class="form-control"
            :class="{ 'is-invalid': linkNameState === false }"
            :disabled="!creating"
            :placeholder="t('datasources.datasource-name')"
          />
        </div>
      </div>
    </div>
    <hr />
    <div class="row">
      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="text-primary form-label">{{ t('datasources.primary.source') }}</label>
          <select v-model="step.link.localSource" class="form-select" @change="onSourceChange('local')">
            <option value="">{{ t('label.none') }}</option>
            <option v-for="opt in supportedSources" :key="opt.value" :value="opt.value">{{ opt.text }}</option>
          </select>
        </div>
      </div>
      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="text-primary form-label">{{ t('datasources.secondary.source') }}</label>
          <select v-model="step.link.foreignSource" class="form-select" @change="onSourceChange('foreign')">
            <option value="">{{ t('label.none') }}</option>
            <option v-for="opt in supportedSources" :key="opt.value" :value="opt.value">{{ opt.text }}</option>
          </select>
        </div>
      </div>
    </div>
    <div class="row">
      <div class="col-12 col-lg-6">
        <div v-if="step.link.localSource" class="mb-3">
          <label class="text-primary form-label">{{ t('datasources.primary.column') }}</label>
          <ColumnSelector
            :value="step.link.localColumn"
            :columns="localColumns"
            style="min-width:100% !important"
            @input="step.link.localColumn = $event"
          />
        </div>
      </div>
      <div class="col-12 col-lg-6">
        <div v-if="step.link.foreignSource" class="mb-3">
          <label class="text-primary form-label">{{ t('datasources.secondary.column') }}</label>
          <ColumnSelector
            :value="step.link.foreignColumn"
            :columns="foreignColumns"
            style="min-width:100% !important"
            @input="step.link.foreignColumn = $event"
          />
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useToast } from 'corteza-lib/vue/dist'
import ColumnSelector from '../../Common/ColumnSelector.vue'

const props = defineProps({
  index: { type: Number, required: true },
  step: { type: Object, required: true },
  datasources: { type: Array, default: () => [] },
  creating: { type: Boolean, default: true },
})
defineEmits(['update:step'])

const { t } = useI18n()
const toast = useToast()
const toastErrorHandler = toast.toastErrorHandler

const localColumns = ref([])
const foreignColumns = ref([])

const kind = computed(() => Object.keys(props.step))
const linkNameState = computed(() => {
  const name = props.step.link?.name
  if (!name) return false
  const isDuplicate = props.datasources.some(({ step }, index) =>
    index !== props.index && step[Object.keys(step)]?.name === name
  )
  return !isDuplicate ? null : false
})

const supportedSources = computed(() => {
  const options = []
  props.datasources.forEach(({ step }, index) => {
    Object.entries(step).forEach(([kind, { name }]) => {
      if (['load', 'group'].includes(kind)) options.push({ value: name || `${index}`, text: name || `${index}` })
    })
  })
  return options
})

watch(() => props.step.link?.name, (newVal, oldVal) => {
  if (!oldVal && newVal) getSourceColumns(['local', 'foreign'])
}, { immediate: false })
watch(() => props.step.link?.localSource, () => getSourceColumns(['local']))
watch(() => props.step.link?.foreignSource, () => getSourceColumns(['foreign']))

function onSourceChange(source) { props.step.link[`${source}Column`] = '' }

async function getSourceColumns(sources = []) {
  sources.forEach(source => {
    const arrName = `${source}Columns`
    if (arrName === 'localColumns') localColumns.value = []
    else foreignColumns.value = []
    const sourceType = props.step.link[`${source}Source`]
    if (sourceType) {
      const steps = props.datasources.filter(({ step }) => step.load || step.aggregate).map(({ step }) => step)
      const describe = [sourceType]
      if (steps.length && describe.length) {
        window.__systemAPI.reportDescribe({ steps, describe })
          .then((frames = []) => {
            const { columns = [] } = frames.find(({ source }) => describe.includes(source)) || {}
            if (arrName === 'localColumns') localColumns.value = columns
            else foreignColumns.value = columns
          }).catch((e) => { toastErrorHandler(t('notification.datasource.describe-failed'))(e) })
      }
    }
  })
}
</script>