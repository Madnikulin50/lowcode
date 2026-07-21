<template>
  <div v-if="scenario">
    <div class="mb-3">
      <label class="form-label text-primary">{{ $t('scenarios.label') }}</label>
      <input
        v-model="scenario.label"
        class="form-control"
        :placeholder="$t('scenarios.scenario-name')"
      />
    </div>
    <div class="mb-3">
      <label class="form-label text-primary">{{ $t('scenarios.values') }}</label>
      <Values
        v-model="values"
      />
    </div>
  </div>
</template>

<script setup>
import { computed, watch, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Values from 'corteza-webapp-compose/src/components/Common/Values'

const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  currentIndex: { type: Number, default: -1 },
  scenario: { type: Object, required: true, default: () => ({ values: { list: [] } }) },
  module: { type: Object, required: false, default: () => ({}) },
})

const columns = ref([])
const currentDatasourceName = ref('')

const values = computed({
  get: () => props.scenario.values,
  set: (v) => { props.scenario.values = v },
})

watch(() => props.currentIndex, () => {
  const { filters = {} } = props.scenario
  const definedFilters = Object.keys(filters)
  currentDatasourceName.value = definedFilters.length ? definedFilters[0] : ''
}, { immediate: true })

watch(currentDatasourceName, (name) => {
  if (name && !props.scenario.filters[name]) {
    props.scenario.filters[name] = {}
  }
  getSourceColumns()
}, { immediate: true })

async function getSourceColumns() {
  columns.value = []
}
</script>
