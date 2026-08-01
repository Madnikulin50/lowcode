<template>
  <div v-if="step.load">
    <div class="mb-3">
      <label class="form-label text-primary">{{ $t('datasources.name-required') }}</label>
      <input
        v-model="step.load.name"
        class="form-control form-control-sm"
        :class="{ 'is-invalid': nameState === false }"
        :disabled="!creating"
        :placeholder="$t('datasources.datasource-name')"
      >
    </div>

    <hr>

    <div class="mb-3">
      <label class="form-label text-primary">{{ $t('datasources.source') }}</label>
      <select
        v-model="step.load.source"
        class="form-select form-control"
        @change="reset"
      >
        <option
          v-for="opt in supportedSources"
          :key="opt.value"
          :value="opt.value"
        >{{ opt.text }}</option>
      </select>
    </div>

    <component
      :is="sourceTypeComponent(step.load.source)"
      v-if="step.load.source"
      v-model:definition="stepDefinition"
    />

    <div
      v-if="columns.length"
      class="mb-3"
    >
      <label class="form-label text-primary">{{ $t('datasources.prefilter') }}</label>
      <prefilter
        v-model:filter="step.load.filter"
        :columns="columns"
      />
    </div>
  </div>
</template>

<script setup lang="js">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useDataSourceBase } from '../base.vue'
import loader from './loader'
import Prefilter from 'corteza-webapp-compose/src/components/Common/Prefilter'

const { t } = useI18n()

const props = defineProps({
  index: {
    type: Number,
    required: true,
  },
  step: {
    type: Object,
    required: true,
    default: () => ({}),
  },
  datasources: {
    type: Array,
    required: false,
    default: () => [],
  },
  creating: {
    type: Boolean,
    default: true,
  },
})

const { kind, nameState } = useDataSourceBase(props)

const columns = ref([])

const supportedSources = ref([
  {
    text: t('datasources.compose-records'),
    value: 'composeRecords',
    definition: [{ label: 'namespace', key: 'namespace' }, { label: 'module', key: 'module' }],
  },
  {
    text: t('datasources.url'),
    value: 'url',
    definition: [{ label: 'url', key: 'url' }],
  },
  {
    text: t('datasources.csv'),
    value: 'url',
    definition: [{ label: 'csv', key: 'csv' }],
  },
])

const stepDefinition = computed({
  get () {
    return props.step.load ? props.step.load.definition : undefined
  },
  set (definition) {
    props.step.load.definition = definition
  },
})

watch(stepDefinition, ({ moduleID, namespaceID }) => {
  if (moduleID && namespaceID) {
    getSourceColumns()
  }
}, { immediate: true, deep: true })

function sourceTypeComponent (k) {
  return loader(k)
}

async function getSourceColumns () {
  const steps = [props.step]
  const describe = [props.step.load.name]

  if (steps.length && describe.length) {
    window.__systemAPI.reportDescribe({ steps, describe })
      .then((frames = []) => {
        const { columns: cols = [] } = frames.find(({ source }) => describe.includes(source)) || {}
        columns.value = cols
      }).catch((e) => {
        toastErrorHandler(t('notification.datasource.describe-failed'))(e)
      })
  }
}

function reset () {
  props.step.load.filter = {}
  props.step.load.sort = ''
}

function toastErrorHandler (msg) {
  return (err) => {
    console.error(msg, err)
  }
}
</script>
