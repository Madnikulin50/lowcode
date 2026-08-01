<template>
  <div v-if="size === 'sm'" class="d-flex align-items-center gap-2">
    <div class="progress bg-light progress-sm flex-fill">
      <div
        class="progress-bar"
        :class="[
          `bg-${progressVariant}`,
          {
            'progress-bar-striped': striped,
            'progress-bar-animated': animated,
            'progress-bar-gradient': true,
            'rounded-end': progressPercent < 100,
            rounded: progressPercent >= 100,
          },
        ]"
        role="progressbar"
        :style="{ width: `${progressPercent}%` }"
        :aria-valuenow="progressValue < 0 ? 0 : progressValue"
        aria-valuemin="0"
        :aria-valuemax="maxValue"
      />
    </div>
    <span
      v-if="progressLabel"
      class="progress-label-sm flex-shrink-0"
      :style="textStyle"
    >
      {{ progressLabel }}
    </span>
  </div>
  <div
    v-else
    class="progress bg-light position-relative progress-default"
  >
    <div
      class="progress-bar"
      :class="[
        `bg-${progressVariant}`,
        {
          'progress-bar-striped': striped,
          'progress-bar-animated': animated,
        },
      ]"
      role="progressbar"
      :style="{ width: `${progressPercent}%` }"
      :aria-valuenow="progressValue < 0 ? 0 : progressValue"
      aria-valuemin="0"
      :aria-valuemax="maxValue"
    >
      <strong
        v-if="progressLabel"
        class="d-flex align-items-center justify-content-center position-absolute mb-0 w-100"
        :class="textVariant"
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
  size?: 'default' | 'sm'
}>(), {
  min: 0,
  max: 100,
  labeled: true,
  relative: true,
  variant: 'success',
  thresholds: () => [],
  textStyle: '',
  size: 'default',
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
    const suffix = props.relative ? '100%' : String(props.max)
    value = `${value} / ${suffix}`
    if (value.endsWith(' / 100%')) value = progressPercent.value + '%'
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

<style scoped>
.progress-sm {
  height: 0.625rem;
  min-width: 8rem;
  border-radius: 0.3125rem;
  overflow: hidden;
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.08);
}

.progress-sm .progress-bar {
  border-radius: 0.3125rem;
  transition: width 0.5s ease;
  box-shadow: 0 0 6px rgba(0, 0, 0, 0.12);
}

.progress-bar-gradient {
  background-image: linear-gradient(135deg, rgba(255, 255, 255, 0.18) 25%, transparent 25%, transparent 50%, rgba(255, 255, 255, 0.18) 50%, rgba(255, 255, 255, 0.18) 75%, transparent 75%, transparent);
  background-size: 1.2rem 1.2rem;
}

.progress-label-sm {
  font-size: 0.8125rem;
  font-weight: 600;
  white-space: nowrap;
  color: inherit;
}

.progress-default {
  height: 1rem;
  border-radius: 0.25rem;
}
</style>
