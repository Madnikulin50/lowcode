<template>
  <div
    v-if="options.likeRecordList"
    ref="fieldContainer"
  >
    <div
      :key="`${metric.label}`"
      :style="fieldWidth"
      class="mb-0 field-container"
    >
      <div v-if="options.horizontalFieldLayoutEnabled" class="row">
        <label class="col-md-6 col-xl-5 d-flex align-items-center text-primary mb-0 form-label">
          <span
            class="d-flex metric-label"
            style="margin-top: 0.1rem;"
            :style="genStyle(metric.valueStyle, true)"
          >
            {{ metric.label }}
          </span>
        </label>
        <div class="col-md-6 col-xl-7 d-flex align-items-center">
          <span
            :style="genStyle(metric.valueStyle)"
            :class="{ 'metric-hover-value': hover }"
          >
            <template v-if="metric.prefix">{{ metric.prefix }}</template>
            {{ displayValue }}
            <template v-if="metric.suffix">{{ metric.suffix }}</template>
          </span>
        </div>
      </div>
      <template v-else>
        <label class="d-flex align-items-center text-primary mb-0 form-label">
          <span
            class="d-flex metric-label"
            style="margin-top: 0.1rem;"
            :style="genStyle(metric.valueStyle, true)"
          >
            {{ metric.label }}
          </span>
        </label>
        <span
          :style="genStyle(metric.valueStyle)"
          :class="{ 'metric-hover-value': hover }"
        >
          <template v-if="metric.prefix">{{ metric.prefix }}</template>
          {{ displayValue }}
          <template v-if="metric.suffix">{{ metric.suffix }}</template>
        </span>
      </template>
    </div>
  </div>
  <div
    v-else
    :style="genStyle(metric.valueStyle)"
    class="text-center"
    :class="{'h-100': metric.valueStyle.notFitVertical !== true}"
  >
    <div
      class="d-none d-print-flex w-100 align-items-center justify-content-center overflow-hidden metric-value"
      :style="genStyle(metric.valueStyle)"
      :class="{ 'metric-hover-value': hover }"
    >
      <template v-if="metric.prefix">{{ metric.prefix }}</template>
      {{ displayValue }}
      <template v-if="metric.suffix">{{ metric.suffix }}</template>
    </div>

    <template v-if="metric.valueStyle.notFitVertical || metric.valueStyle.notFitHorizontal">
      <div
        class="d-print-flex align-items-center justify-content-center overflow-hidden"
        :style="genStyle(metric.valueStyle)"
      >
        <span
          v-if="metric.showLabel"
          :style="genStyle(metric.valueStyle, true)"
        >
          {{ metric.label }}:&nbsp;
        </span>
        <span :class="{ 'metric-hover-value': hover }">
          <template v-if="metric.prefix">{{ metric.prefix }}</template>
          {{ displayValue }}
          <template v-if="metric.suffix">{{ metric.suffix }}</template>
        </span>
      </div>
    </template>
    <template v-else>
      <svg
        :viewBox="getVB"
        class="h-100 w-100 d-flex d-print-none"
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
          :class="{ 'metric-hover-value': hover }"
        >
          <template v-if="metric.showLabel">{{ metric.label }}:&nbsp;</template>
          <template v-if="metric.prefix">{{ metric.prefix }}</template>
          {{ displayValue }}
          <template v-if="metric.suffix">{{ metric.suffix }}</template>
        </text>
      </svg>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, watch, onBeforeUnmount, nextTick, inject } from 'vue'
import { fmt } from 'corteza-lib/js/dist'

const props = defineProps({
  metric: { type: Object, required: false, default: () => ({}) },
  options: { type: Object, required: false, default: () => ({}) },
  value: { type: Object, required: false, default: () => ({}) },
  hover: { type: Boolean, required: false, default: false },
})

const $Settings = inject('$Settings')

const vvb = ref(['0', '0', '0', '0'])
const fieldContainer = ref(null)
const metricItem = ref(null)

const getVB = computed(() => vvb.value.join(' '))

const displayValue = computed(() => fmt.number(props.value.value))

const fieldWidth = computed(() => {
  if (props.options.recordFieldLayoutOption !== 'noWrap') return {}
  return { 'min-width': '13rem' }
})

watch(() => [props.metric, props.options, props.value], () => { update() }, { immediate: true })

watch(() => props.options.recordFieldLayoutOption, (newVal) => {
  if (newVal === 'wrap') {
    if (fieldContainer.value) {
      initializeResizeObserver(fieldContainer.value, newVal)
    }
  } else if (resizeObserver) {
    resizeObserver.unobserve(fieldContainer.value)
    columnWrapClass = ''
  }
})

onBeforeUnmount(() => {
  setDefaultValues()
})

function update () {
  nextTick(() => {
    if (!metricItem.value) return
    const { width, height } = metricItem.value.getBBox()
    const tmp = [...vvb.value]
    tmp[2] = parseInt(Math.ceil(width))
    tmp[3] = parseInt(Math.ceil(height))
    vvb.value = tmp
  })
}

function themeSettings () {
  return $Settings.get('ui.studio.themes', [])
}

function getColor (value) {
  if (value[0] === '#') return value
  const themes = themeSettings()
    .filter((theme) => theme.id !== 'general')
    .map((theme) => ({ id: theme.id, values: JSON.parse(theme.values) }))
  return themes[0].values[value] || value
}

function genStyle (s = {}, forLabel = false) {
  const d = {
    fill: forLabel ? (s.labelColor || s.color) : s.color,
    backgroundColor: s.backgroundColor,
    fontSize: s.fontSize ? s.fontSize + 'px' : undefined,
    color: forLabel ? (s.labelColor || s.color) : s.color,
  }
  if (s.colorThresholds && forLabel === false) {
    const val = props.value.value
    const { variant } = [...s.colorThresholds].sort((a, b) => b.value - a.value).find(t => val >= t.value) || {}
    if (variant !== undefined) { d.color = variant; d.fill = variant }
  }
  for (const v of Object.keys(d)) { if (d[v] === undefined) delete d[v] }
  d.color = getColor(d.color)
  d.backgroundColor = getColor(d.backgroundColor)
  d.fill = getColor(d.fill)
  return d
}

function setDefaultValues () {
  vvb.value = ['0', '0', '0', '0']
}
</script>

<style lang="scss">
.metric-hover-value {
  &:hover {
    transform: translateY(-5px);
    box-shadow: 0 10px 20px rgba(0, 0, 0, 0.15);
  }
}

.metric-label {
  overflow: hidden;
}
</style>
