<template>
  <div v-if="step.aggregate">
    <div class="mb-3">
      <label class="text-primary form-label">{{ t('datasources.name-required') }}</label>
      <input
        v-model="step.aggregate.name"
        class="form-control"
        :class="{ 'is-invalid': nameState === false }"
        :disabled="!creating"
        :placeholder="t('datasources.datasource-name')"
      />
    </div>
    <hr />
    <div class="mb-3">
      <label class="text-primary form-label">{{ t('datasources.source') }}</label>
      <select v-model="step.aggregate.source" class="form-select" @change="reset">
        <option :value="undefined">{{ t('label.none') }}</option>
        <option v-for="opt in supportedSources" :key="opt.value" :value="opt.value">{{ opt.text }}</option>
      </select>
    </div>
    <div v-if="step.aggregate.source">
      <div class="mb-3">
        <label class="text-primary form-label">{{ t('datasources.group-by') }}</label>
        <GroupBy :group-by="step.aggregate.keys" @update:group-by="step.aggregate.keys = $event" />
      </div>
      <div class="mb-3">
        <label class="text-primary form-label">{{ t('datasources.aggregate') }}</label>
        <AggregateCmp :aggregate="step.aggregate.columns" @update:aggregate="step.aggregate.columns = $event" />
      </div>
    </div>
  </div>
</template>
<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useToast } from 'corteza-lib/vue/dist'
import GroupBy from './GroupBy.vue'
import AggregateCmp from './Aggregate.vue'

const props = defineProps({
  index: { type: Number, required: true },
  step: { type: Object, required: true },
  datasources: { type: Array, default: () => [] },
  creating: { type: Boolean, default: true },
})

const { t } = useI18n()
const toast = useToast()
const toastErrorHandler = toast.toastErrorHandler

const columns = ref([])

const nameState = computed(() => {
  const name = props.step.aggregate?.name
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
      options.push({ value: name || `${index}`, text: name || `${index}` })
    })
  })
  return options
})

async function getSourceColumns() {
  const steps = props.datasources.filter(({ step }) => step.load).map(({ step }) => step)
  steps.push(props.step)
  const describe = [props.step.aggregate.name]
  if (steps.length && describe.length) {
    window.__systemAPI.reportDescribe({ steps, describe })
      .then((frames = []) => {
        const { columns: cols = [] } = frames.find(({ source }) => describe.includes(source)) || {}
        columns.value = cols
      }).catch((e) => { toastErrorHandler(t('notification.datasource.describe-failed'))(e) })
  }
}

function reset() {
  props.step.aggregate.filter = {}
  props.step.aggregate.sort = ''
  props.step.aggregate.keys = []
  props.step.aggregate.columns = []
}
</script>