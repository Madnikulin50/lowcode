<template>
  <div v-if="displayElement">
    <div class="mb-3">
      <label class="text-primary form-label">{{ t('label.name') }}</label>
      <input v-model="displayElement.name" class="form-control" />
    </div>
    <button
      class="btn btn-primary w-100 mb-2"
      data-bs-toggle="collapse"
      data-bs-target="#datasources-collapse"
      :disabled="!usesDatasources"
    >
      {{ t('builder.datasources.label') }}
    </button>
    <div v-if="usesDatasources" id="datasources-collapse" class="collapse show">
      <div class="mb-3">
        <label class="text-primary form-label">{{ t('builder.datasources.label') }}</label>
        <select v-model="options.source" class="form-select" @change="setConfigurableSources">
          <option value="">{{ t('label.none') }}</option>
          <option v-for="opt in sources" :key="opt.value" :value="opt.value">{{ opt.text }}</option>
        </select>
      </div>
      <div v-if="currentConfigurableDatasourceName">
        <div v-if="hasMultipleConfigurableDatasources" class="mb-3">
          <label class="text-primary form-label">{{ t('builder.joined-datasource-handling') }}</label>
          <select v-model="currentConfigurableDatasourceName" class="form-select" @change="configurableDatasourceChanged">
            <option v-for="ds in options.datasources" :key="ds.name" :value="ds.name">{{ ds.name }}</option>
          </select>
        </div>
        <div v-if="currentConfigurableDatasourceIndex >= 0">
          <div v-if="columns.length" class="mb-3">
            <label class="text-primary form-label">{{ t('builder.prefilter') }}</label>
            <Prefilter :filter="options.datasources[currentConfigurableDatasourceIndex].filter" :columns="columns[currentConfigurableDatasourceIndex]" />
          </div>
          <div v-if="displayElement.kind === 'Table'" class="mb-3">
            <label class="text-primary form-label">{{ t('builder.limit') }}</label>
            <input v-model.number="pagingLimit" type="number" class="form-control" />
          </div>
        </div>
      </div>
    </div>
    <button
      class="btn btn-primary w-100 mb-2"
      data-bs-toggle="collapse"
      data-bs-target="#display-collapse"
      :disabled="!showDisplayElementConfigurator"
    >
      {{ t('builder.element') }}
    </button>
    <div id="display-collapse" class="collapse show">
      <component
        :is="displayElementConfigurator"
        :display-element-options="options"
        :columns="columns"
        :datasource="currentDatasource"
        @update:display-element-options="options = $event"
      />
    </div>
  </div>
</template>
<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useToast } from 'corteza-lib/vue/dist'
import getDisplayElementConfigurator from './loader.js'
import Prefilter from '../../../../Common/Prefilter.vue'

const props = defineProps({
  displayElement: { type: Object, required: true },
  block: { type: Object, required: true },
  datasources: { type: Array, required: true },
})
const emit = defineEmits(['update:displayElement'])

const { t } = useI18n()
const toast = useToast()
const toastErrorHandler = toast.toastErrorHandler

const columns = ref([])
const currentConfigurableDatasourceName = ref(undefined)
const currentConfigurableDatasourceIndex = ref(undefined)

const usesDatasources = computed(() => !['Text'].includes(props.displayElement.kind))
const displayElementConfigurator = computed(() => getDisplayElementConfigurator(props.displayElement.kind))
const showDisplayElementConfigurator = computed(() => usesDatasources.value ? !!currentDatasource.value : true)

const sources = computed(() => {
  const options = [{ value: '', text: t('label.none') }]
  props.datasources.forEach(({ step }, index) => {
    Object.values(step).forEach(({ name }) => { options.push({ value: name || `${index}`, text: name || `${index}` }) })
  })
  return options
})

const currentDatasource = computed(() => {
  if (options.value?.source) {
    return props.datasources.find(({ step: { load = {}, link = {}, join = {}, aggregate = {} } }) =>
      [load.name, link.name, join.name, aggregate.name].includes(options.value.source)
    )
  }
  return undefined
})

const hasMultipleConfigurableDatasources = computed(() =>
  currentDatasource.value?.step?.link && options.value?.datasources?.length > 1
)

const options = computed({
  get: () => props.displayElement?.options,
  set: (opts = {}) => { emit('update.displayElement', { ...props.displayElement, options: opts }) },
})

const pagingLimit = computed({
  get: () => {
    if (currentConfigurableDatasourceIndex.value >= 0) {
      const { paging = {} } = options.value?.datasources?.[currentConfigurableDatasourceIndex.value] || {}
      return paging.limit || 0
    }
    return 0
  },
  set: (limit = 0) => {
    if (currentConfigurableDatasourceIndex.value >= 0) {
      if (!options.value.datasources[currentConfigurableDatasourceIndex.value].paging) {
        options.value.datasources[currentConfigurableDatasourceIndex.value].paging = {}
      }
      options.value.datasources[currentConfigurableDatasourceIndex.value].paging.limit = limit || 0
    }
  },
})

watch(() => options.value?.source, (source) => { describeReport(source) }, { immediate: true })
watch(() => props.displayElement.elementID, () => {
  currentConfigurableDatasourceIndex.value = props.datasources.length ? 0 : -1
  if (usesDatasources.value) currentConfigurableDatasourceName.value = (options.value?.datasources?.[0] || {}).name
}, { immediate: true })

function describeReport(source) {
  columns.value = []
  if (source) {
    const steps = props.datasources.map(({ step }) => step)
    window.__systemAPI.reportDescribe({ steps, describe: [options.value.source] })
      .then((frames = []) => {
        columns.value = frames.filter(({ source: s }) => s === options.value.source).map(({ columns: cols = [] }) => cols) || []
      }).catch((e) => { toastErrorHandler(t('notification.datasource.describe-failed'))(e) })
  }
}

function setConfigurableSources(source) {
  if (!options.value) return
  options.value.datasources = []
  currentConfigurableDatasourceName.value = undefined
  let configurableDatasources = []
  if (source) {
    configurableDatasources = [source]
    const { link } = currentDatasource.value?.step || {}
    if (link) configurableDatasources = [link.localSource, link.foreignSource]
  }
  currentConfigurableDatasourceName.value = configurableDatasources[0]
  options.value.datasources = configurableDatasources.map(s => ({ name: s, sort: '', filter: {} }))
}

function configurableDatasourceChanged(source) {
  if (source) currentConfigurableDatasourceIndex.value = options.value?.datasources?.findIndex(({ name }) => source === name) ?? -1
}
</script>