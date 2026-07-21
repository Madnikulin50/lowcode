<template>
  <div class="progress bg-light position-relative">
    <div
      class="progress-bar"
      :class="[`bg-${progressVariant}`, { 'progress-bar-striped': striped, 'progress-bar-animated': animated }]"
      role="progressbar"
      :style="{ width: `${progressPercent}%` }"
      :aria-valuenow="progressValue < 0 ? 0 : progressValue"
      aria-valuemin="0"
      :aria-valuemax="maxValue"
    >
      <strong
        :class="textVariant"
        class="d-flex align-items-center justify-content-center position-absolute mb-0 w-100"
        :style="textStyle"
      >
        {{ progressLabel }}
      </strong>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  value: number
  min?: number
  max?: number
  labeled?: boolean
  relative?: boolean
  progress?: boolean
  striped?: boolean
  animated?: boolean
  variant?: string
  thresholds?: Array<{ value: number; variant: string }>
  textStyle?: string
}>(), {
  min: 0,
  max: 100,
  labeled: true,
  relative: true,
  variant: 'success',
  thresholds: () => [],
  textStyle: '',
})

const maxValue = computed(() => Math.abs(props.max - props.min))

const progressValue = computed(() => {
  if (props.value < props.min && props.max > props.min) {
    return props.value - props.min
  } else if (props.value > props.min && props.max < props.min) {
    return props.min - props.value
  }
  return Math.abs(props.value - props.min)
})

const progressPercent = computed(() => {
  const pv = Math.max(0, progressValue.value)
  if (maxValue.value === 0) return 0
  return Math.round((pv / maxValue.value) * 10000) / 100
})

const progressLabel = computed(() => {
  if (!props.labeled) return undefined
  let value: string | number = props.value
  if (props.relative) {
    value = `${progressPercent.value}%`
  }
  if (props.progress) {
    value = `${value} / ${props.relative ? '100' : props.max}${props.relative ? '%' : ''}`
  }
  return value
})

const sortedVariants = computed(() => {
  return [...props.thresholds].filter(t => t.value >= 0).sort((a, b) => b.value - a.value)
})

const progressVariant = computed(() => {
  const value = progressPercent.value
  let pv = props.variant
  if (sortedVariants.value.length) {
    const found = sortedVariants.value.find(t => value >= t.value)
    if (found) pv = found.variant
  }
  return pv
})

const textVariant = computed(() => {
  return ['dark', 'primary'].includes(progressVariant.value) ? 'text-white' : 'text-dark'
})
</script>
