<template>
  <div
    v-if="options"
    :style="style"
    class="d-flex align-items-center justify-content-center overflow-hidden h-100 px-2 rounded"
  >
    <svg
      :viewBox="viewbox"
      class="h-100 w-100 align-items-end d-flex"
      width="100%"
      height="100%"
    >
      <text
        ref="metricItem"
        y="50%"
        x="50%"
        text-anchor="middle"
        dominant-baseline="central"
        text-rendering="geometricPrecision"
      >
        {{ displayedMetric }}
      </text>
    </svg>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import numeral from 'numeral'

const props = defineProps<{
  displayElement: any
  labels?: Record<string, any>
}>()

const metricItem = ref<SVGTextElement>()

const dataframes = computed(() => props.displayElement?.dataframes || [])
const options = computed(() => props.displayElement?.options || undefined)

const vvb = ref(['0', '0', '0', '0'])

const viewbox = computed(() => vvb.value.join(' '))

const style = computed(() => ({
  fill: options.value?.color || '#0B344E',
  backgroundColor: options.value?.backgroundColor || '#FFFFFF00',
}))

const value = computed(() => {
  if (dataframes.value.length) {
    const { rows = [], columns = [] } = dataframes.value[0]
    if (columns.length) {
      const columnIndex = columns.findIndex(({ name }: any) => name === options.value?.valueColumn)
      if (rows.length) {
        return rows[0] ? rows[0][columnIndex] || '' : ''
      }
    }
  }
  return ''
})

const displayedMetric = computed(() => {
  const { prefix = '', suffix = '', format = '' } = options.value || {}
  if (value.value) {
    const v = format ? numeral(value.value).format(format) : value.value
    return `${prefix}${v}${suffix}`
  }
  return ''
})

watch(value, () => {
  update()
}, { immediate: true })

function update() {
  nextTick(() => {
    if (metricItem.value) {
      const { width, height } = metricItem.value.getBBox()
      const tmp = [...vvb.value]
      tmp[2] = String(Math.ceil(width))
      tmp[3] = String(Math.ceil(height))
      vvb.value = tmp
    }
  })
}
</script>
