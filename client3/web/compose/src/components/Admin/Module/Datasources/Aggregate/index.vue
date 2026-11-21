<template>
  <div v-if="step.aggregate">
    <div class="mb-3">
      <label class="form-label text-primary">{{ $t('datasources.name-required') }}</label>
      <input
        v-model="step.aggregate.name"
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
        v-model="step.aggregate.source"
        class="form-select form-control"
        @change="reset"
      >
        <option :value="undefined">{{ $t('label.none') }}</option>
        <option
          v-for="opt in supportedSources"
          :key="opt.value"
          :value="opt.value"
        >{{ opt.text }}</option>
      </select>
    </div>

    <div v-if="step.aggregate.source">
      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('datasources.group-by') }}</label>
        <group-by
          v-model:group-by="step.aggregate.keys"
        />
      </div>

      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('datasources.aggregate') }}</label>
        <aggregate
          v-model:aggregate="step.aggregate.columns"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="js">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useDataSourceBase } from '../base.vue'
import GroupBy from './GroupBy'
import Aggregate from './Aggregate'

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

async function getSourceColumns () {
  const steps = props.datasources.filter(({ step }) => step.load).map(({ step }) => step)
  steps.push(props.step)
  const describe = [props.step.aggregate.name]

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
  props.step.aggregate.filter = {}
  props.step.aggregate.sort = ''
  props.step.aggregate.keys = []
  props.step.aggregate.columns = []
}

function toastErrorHandler (msg) {
  return (err) => {
    console.error(msg, err)
  }
}
</script>
