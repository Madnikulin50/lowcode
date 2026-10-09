<template>
  <e-charts
    ref="chartRef"
    :option="chart"
    :theme="theme"
    autoresize
    class="position-absolute w-100 h-100 overflow-hidden"
    v-bind="$attrs"
  />
</template>

<script setup lang="ts">
import { computed, ref, onBeforeUnmount, useAttrs } from 'vue'
import { shared } from 'corteza-lib/js/dist'

defineOptions({
  inheritAttrs: false,
})

const props = defineProps<{
  chart: shared.Chart
}>()

const attrs = useAttrs()

const chartRef = ref<{ dispose: () => void } | null>(null)

const theme = computed(() => {
  const { darkMode } = props.chart || {}
  return darkMode ? 'dark' : 'light'
})

onBeforeUnmount(() => {
  if (chartRef.value) {
    chartRef.value.dispose()
  }
})
</script>
