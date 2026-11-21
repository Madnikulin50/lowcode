<template>
  <div v-if="step.load">
    <div class="mb-3">
      <label class="text-primary form-label">{{ t('datasources.name-required') }}</label>
      <input
        v-model="step.load.name"
        class="form-control"
        :class="{ 'is-invalid': nameState === false }"
        :disabled="!creating"
        :placeholder="t('datasources.datasource-name')"
      />
    </div>
    <hr />
    <div class="mb-3">
      <label class="text-primary form-label">{{ t('datasources.source') }}</label>
      <select v-model="step.load.source" class="form-select" @change="reset">
        <option v-for="opt in supportedSources" :key="opt.value" :value="opt.value">{{ opt.text }}</option>
      </select>
    </div>
    <component
      :is="sourceTypeComponent(step.load.source)"
      v-if="step.load.source"
      :definition="step.load.definition"
      @update:definition="step.load.definition = $event"
    />
    <div class="mb-3">
      <label class="text-primary form-label">{{ t('datasources.prefilter') }}</label>
      <Prefilter :filter="step.load.filter" :columns="columns" />
    </div>
  </div>
</template>
<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useToast } from 'corteza-lib/vue/dist'
import Prefilter from '../../../Common/Prefilter.vue'
import sourceTypeComponent from './loader.js'

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
const columns = ref([])

const nameState = computed(() => {
  const name = props.step.load?.name
  if (!name) return false
  const isDuplicate = props.datasources.some(({ step }, index) =>
    index !== props.index && step[Object.keys(step)]?.name === name
  )
  return !isDuplicate ? null : false
})

const supportedSources = [
  { text: t('datasources.compose-records'), value: 'composeRecords', definition: [{ label: 'namespace', key: 'namespace' }, { label: 'module', key: 'module' }] },
]

watch(() => props.step.load?.definition, ({ moduleID, namespaceID } = {}) => {
  if (moduleID && namespaceID) getSourceColumns()
}, { immediate: true, deep: true })

async function getSourceColumns() {
  const steps = [props.step]
  const describe = [props.step.load.name]
  if (steps.length && describe.length) {
    window.__systemAPI.reportDescribe({ steps, describe })
      .then((frames = [], warnings = []) => {
        const { columns: cols = [] } = frames.find(({ source }) => describe.includes(source)) || {}
        columns.value = cols
        warnings.forEach((warning) => { toastErrorHandler(t('notification.datasource.describe-failed'))(warning) })
      }).catch((e) => { toastErrorHandler(t('notification.datasource.describe-failed'))(e) })
  }
}

function reset() {
  props.step.load.filter = {}
  props.step.load.sort = ''
}
</script>