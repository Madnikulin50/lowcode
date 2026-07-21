<template>
  <div v-if="scenario">
    <div class="mb-3">
      <label class="text-primary form-label">{{ t('scenarios.label') }}</label>
      <input v-model="scenario.label" class="form-control" :placeholder="t('scenarios.scenario-name')" />
    </div>
    <div class="mb-3">
      <label class="text-primary form-label">{{ t('scenarios.datasource') }}</label>
      <select v-model="currentDatasourceName" class="form-select">
        <option value="">{{ t('label.none') }}</option>
        <option v-for="opt in datasourceOptions" :key="opt.value" :value="opt.value">{{ opt.text }}</option>
      </select>
    </div>
    <div v-if="currentDatasourceName && scenario.filters?.[currentDatasourceName]" class="mb-3">
      <label class="text-primary form-label">{{ t('scenarios.prefilter') }}</label>
      <Prefilter :filter="scenario.filters[currentDatasourceName]" :columns="columns" />
    </div>
  </div>
</template>
<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useToast } from 'corteza-lib/vue/dist'
import Prefilter from '../../Common/Prefilter.vue'

const props = defineProps({
  currentIndex: { type: Number, default: -1 },
  scenario: { type: Object, required: true },
  datasources: { type: Array, default: () => [] },
})

const { t } = useI18n()
const toast = useToast()
const toastErrorHandler = toast.toastErrorHandler

const currentDatasourceName = ref('')
const columns = ref([])

const datasourceOptions = computed(() => {
  const options = [{ value: '', text: t('label.none') }]
  props.datasources.forEach(({ step }, index) => {
    Object.entries(step).forEach(([kind, { name }]) => {
      if (['load'].includes(kind)) options.push({ value: name || `${index}`, text: name || `${index}` })
    })
  })
  return options
})

watch(() => props.currentIndex, () => {
  const { filters = {} } = props.scenario
  const definedFilters = Object.keys(filters)
  currentDatasourceName.value = definedFilters.length ? definedFilters[0] : ''
}, { immediate: true })

watch(currentDatasourceName, (name) => {
  if (name && !props.scenario.filters[name]) props.scenario.filters[name] = {}
  getSourceColumns()
}, { immediate: true })

async function getSourceColumns() {
  columns.value = []
  if (currentDatasourceName.value) {
    const steps = props.datasources.map(({ step }) => step)
    window.__systemAPI.reportDescribe({ steps, describe: [currentDatasourceName.value] })
      .then((frames = []) => { columns.value = (frames[0] || {}).columns || [] })
      .catch((e) => { toastErrorHandler(t('notification.datasource.describe-failed'))(e) })
  }
}
</script>