<template>
  <div v-if="step.join">
    <div class="row g-0">
      <div class="col">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('datasources.name-required') }}</label>
          <input
            v-model="step.join.name"
            class="form-control form-control-sm"
            :disabled="!creating"
            :class="{ 'is-invalid': nameState === false }"
            :placeholder="$t('datasources.datasource-name')"
          >
        </div>
      </div>
    </div>

    <hr>

    <div class="row g-0">
      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('datasources.primary.source') }}</label>
          <select
            v-model="step.join.localSource"
            class="form-select form-control"
            @change="onSourceChange('local')"
          >
            <option value="">{{ $t('label.none') }}</option>
            <option
              v-for="opt in supportedSources"
              :key="opt.value"
              :value="opt.value"
            >{{ opt.text }}</option>
          </select>
        </div>
      </div>

      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('datasources.secondary.source') }}</label>
          <select
            v-model="step.join.foreignSource"
            class="form-select form-control"
            @change="onSourceChange('foreign')"
          >
            <option value="">{{ $t('label.none') }}</option>
            <option
              v-for="opt in supportedSources"
              :key="opt.value"
              :value="opt.value"
            >{{ opt.text }}</option>
          </select>
        </div>
      </div>
    </div>

    <div class="row g-0">
      <div class="col-12 col-lg-6">
        <div
          v-if="step.join.localSource"
          class="mb-3"
        >
          <label class="form-label text-primary">{{ $t('datasources.primary.column') }}</label>
          <column-selector
            v-model="step.join.localColumn"
            :columns="localColumns"
            style="min-width: 100% !important;"
          />
        </div>
      </div>

      <div class="col-12 col-lg-6">
        <div
          v-if="step.join.foreignSource"
          class="mb-3"
        >
          <label class="form-label text-primary">{{ $t('datasources.secondary.column') }}</label>
          <column-selector
            v-model="step.join.foreignColumn"
            :columns="foreignColumns"
            style="min-width: 100% !important;"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="js">
import { ref, computed, watch } from 'vue'
import { useDataSourceBase } from './base.vue'
import ColumnSelector from 'corteza-webapp-compose/src/components/Common/ColumnSelector.vue'

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

const localColumns = ref([])
const foreignColumns = ref([])

const supportedSources = computed(() => {
  const options = []
  props.datasources.forEach(({ step }, index) => {
    Object.entries(step).forEach(([kind, { name }]) => {
      if (index !== props.index) {
        options.push({ value: name || `${index}`, text: name || `${index}` })
      }
    })
  })
  return options
})

watch(() => props.step.join.name, {
  immediate: true,
  handler (newStep, oldStep) {
    if (!oldStep && newStep) {
      getSourceColumns(['local', 'foreign'])
    }
  },
})

watch(() => props.step.join.localSource, {
  handler () {
    getSourceColumns(['local'])
  },
})

watch(() => props.step.join.foreignSource, {
  handler () {
    getSourceColumns(['foreign'])
  },
})

function onSourceChange (source) {
  props.step.join[`${source}Column`] = ''
}

async function getSourceColumns (sources = []) {
  sources.forEach(source => {
    localColumns.value = []
    foreignColumns.value = []

    const sourceType = props.step.join[`${source}Source`]

    if (sourceType) {
      const steps = props.datasources.filter(({ step }, index) => index !== props.index && !step.link).map(({ step }) => step)
      const describe = [sourceType]

      if (steps.length && describe.length) {
        window.__systemAPI.reportDescribe({ steps, describe })
          .then((frames = []) => {
            const { columns = [] } = frames.find(({ source }) => describe.includes(source)) || {}
            if (source === 'local') {
              localColumns.value = columns
            } else {
              foreignColumns.value = columns
            }
          }).catch((e) => {
            toastErrorHandler(t('notification.datasource.describe-failed'))(e)
          })
      }
    }
  })
}

function toastErrorHandler (msg) {
  return (err) => {
    console.error(msg, err)
  }
}
</script>
