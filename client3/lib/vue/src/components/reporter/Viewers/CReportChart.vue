<template>
  <div class="d-flex h-100 position-relative">
    <CChart
      v-if="chart"
      :chart="chart"
      class="flex-fill p-1"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { CChart } from '../../chart/index.ts'

const props = defineProps<{
  displayElement: any
  labels?: Record<string, any>
}>()

const dataframes = computed(() => props.displayElement?.dataframes || [])
const options = computed(() => props.displayElement?.options || undefined)

const chart = ref<any>(undefined)

watch(dataframes, () => {
  nextTick(() => renderChart())
}, { deep: true, immediate: true })

watch(options, () => {
  chart.value = undefined
  nextTick(() => renderChart())
}, { deep: true })

function renderChart() {
  const meta = {
    themeVariables: getThemeVariables(),
  }
  if (options.value?.getChartConfiguration) {
    chart.value = options.value.getChartConfiguration(dataframes.value, meta)
  }
}

function getThemeVariables() {
  const getCssVariable = (variableName: string) => {
    return getComputedStyle(document.documentElement).getPropertyValue(variableName).trim()
  }
  return ['white', 'black', 'primary', 'secondary', 'success', 'warning', 'danger', 'light', 'extra-light', 'dark', 'font-regular'].reduce((acc: Record<string, string>, variable) => {
    acc[variable] = getCssVariable(`--${variable}`)
    return acc
  }, {})
}
</script>
