<template>
  <!-- Balloon (баллон) — pill bar, color by value thresholds -->
  <div
    v-if="isBalloon"
    class="mb-metric mb-metric--balloon"
    :class="{ 'mb-metric--drill': hover, 'mb-metric--empty': isEmpty }"
  >
    <div
      class="mb-balloon fw-bolder rounded-pill d-flex align-items-center"
      :style="balloonStyle"
    >
      <div class="mb-balloon-label flex-grow-1 me-2 text-truncate">
        {{ metric.label || '—' }}
      </div>
      <div class="mb-balloon-value flex-shrink-0 d-inline-flex align-items-center gap-1">
        <font-awesome-icon
          v-if="thresholdIcon"
          :icon="thresholdIcon"
          class="mb-threshold-icon"
        />
        <template v-if="isEmpty">
          <span v-if="showEmptyPlaceholder">{{ emptyText }}</span>
        </template>
        <template v-else>
          <template v-if="metric.prefix">{{ metric.prefix }}</template>{{ displayValue }}<template v-if="metric.suffix">{{ metric.suffix }}</template>
        </template>
      </div>
    </div>
  </div>

  <!-- Record-style layout -->
  <div
    v-else-if="recordStyle"
    class="mb-metric"
    :class="[
      `mb-metric--${role}`,
      { 'mb-metric--drill': hover, 'mb-metric--empty': isEmpty },
    ]"
  >
    <!-- Title role: large value as header -->
    <template v-if="role === 'title'">
      <div class="rb-title d-inline-flex align-items-center gap-2" :style="valueStyle">
        <font-awesome-icon
          v-if="thresholdIcon"
          :icon="thresholdIcon"
          class="mb-threshold-icon"
        />
        <span>
          <template v-if="isEmpty">
            <span v-if="showEmptyPlaceholder" class="rb-empty">{{ emptyText }}</span>
          </template>
          <template v-else>
            <template v-if="metric.prefix">{{ metric.prefix }}</template>{{ displayValue }}<template v-if="metric.suffix">{{ metric.suffix }}</template>
          </template>
        </span>
      </div>
      <div v-if="metric.label" class="rb-label mt-1">{{ metric.label }}</div>
    </template>

    <!-- Badge role -->
    <template v-else-if="role === 'badge'">
      <div class="rb-badge" :style="badgeStyle">
        <span v-if="metric.label" class="rb-badge-label">{{ metric.label }}</span>
        <span class="rb-badge-value d-inline-flex align-items-center gap-1">
          <font-awesome-icon
            v-if="thresholdIcon"
            :icon="thresholdIcon"
            class="mb-threshold-icon"
          />
          <template v-if="isEmpty">
            <span v-if="showEmptyPlaceholder" class="rb-empty">{{ emptyText }}</span>
          </template>
          <template v-else>
            <template v-if="metric.prefix">{{ metric.prefix }}</template>{{ displayValue }}<template v-if="metric.suffix">{{ metric.suffix }}</template>
          </template>
        </span>
      </div>
    </template>

    <!-- Meta role -->
    <template v-else-if="role === 'meta'">
      <div class="rb-meta-item">
        <span v-if="metric.label" class="rb-meta-label">{{ metric.label }}</span>
        <span class="rb-meta-value d-inline-flex align-items-center gap-1" :style="valueStyle">
          <font-awesome-icon
            v-if="thresholdIcon"
            :icon="thresholdIcon"
            class="mb-threshold-icon"
          />
          <template v-if="isEmpty">
            <span v-if="showEmptyPlaceholder" class="rb-empty">{{ emptyText }}</span>
          </template>
          <template v-else>
            <template v-if="metric.prefix">{{ metric.prefix }}</template>{{ displayValue }}<template v-if="metric.suffix">{{ metric.suffix }}</template>
          </template>
        </span>
      </div>
    </template>

    <!-- Hero role: large centered number without SVG -->
    <template v-else-if="role === 'hero'">
      <div class="mb-metric-hero text-center" :style="valueStyle">
        <div class="mb-metric-hero-value d-inline-flex align-items-center gap-2">
          <font-awesome-icon
            v-if="thresholdIcon"
            :icon="thresholdIcon"
            class="mb-threshold-icon"
          />
          <span>
            <template v-if="isEmpty">
              <span v-if="showEmptyPlaceholder" class="rb-empty">{{ emptyText }}</span>
            </template>
            <template v-else>
              <template v-if="metric.prefix">{{ metric.prefix }}</template>{{ displayValue }}<template v-if="metric.suffix">{{ metric.suffix }}</template>
            </template>
          </span>
        </div>
        <div v-if="metric.label" class="rb-label mt-1">{{ metric.label }}</div>
      </div>
    </template>

    <!-- Default: label → value like Record field -->
    <template v-else>
      <div v-if="options.horizontalFieldLayoutEnabled" class="row g-2">
        <div class="col-md-6 col-xl-5">
          <label class="form-label d-flex align-items-center mb-0 rb-label">
            <span class="d-flex" style="margin-top: 0.1rem;">{{ metric.label }}</span>
          </label>
        </div>
        <div class="col-md-6 col-xl-7 d-flex align-items-center">
          <span class="value w-100 d-inline-flex align-items-center gap-1" :style="valueStyle">
            <font-awesome-icon
              v-if="thresholdIcon"
              :icon="thresholdIcon"
              class="mb-threshold-icon"
            />
            <template v-if="isEmpty">
              <span v-if="showEmptyPlaceholder" class="rb-empty">{{ emptyText }}</span>
            </template>
            <template v-else>
              <template v-if="metric.prefix">{{ metric.prefix }}</template>{{ displayValue }}<template v-if="metric.suffix">{{ metric.suffix }}</template>
            </template>
          </span>
        </div>
      </div>
      <template v-else>
        <label class="form-label d-flex align-items-center mb-0 rb-label">
          <span class="d-flex" style="margin-top: 0.1rem;">{{ metric.label }}</span>
        </label>
        <span class="value d-inline-flex align-items-center gap-1" :style="valueStyle">
          <font-awesome-icon
            v-if="thresholdIcon"
            :icon="thresholdIcon"
            class="mb-threshold-icon"
          />
          <template v-if="isEmpty">
            <span v-if="showEmptyPlaceholder" class="rb-empty">{{ emptyText }}</span>
          </template>
          <template v-else>
            <template v-if="metric.prefix">{{ metric.prefix }}</template>{{ displayValue }}<template v-if="metric.suffix">{{ metric.suffix }}</template>
          </template>
        </span>
      </template>
    </template>
  </div>

  <!-- Legacy SVG hero mode (likeRecordList off) -->
  <div
    v-else
    :style="genStyle(metric.valueStyle)"
    class="text-center"
    :class="{ 'h-100': metric.valueStyle?.notFitVertical !== true, 'mb-metric--drill': hover }"
  >
    <div
      class="d-none d-print-flex w-100 align-items-center justify-content-center overflow-hidden metric-value"
      :style="genStyle(metric.valueStyle)"
    >
      <template v-if="isEmpty">
        <span v-if="showEmptyPlaceholder" class="rb-empty">{{ emptyText }}</span>
      </template>
      <template v-else>
        <template v-if="metric.prefix">{{ metric.prefix }}</template>
        {{ displayValue }}
        <template v-if="metric.suffix">{{ metric.suffix }}</template>
      </template>
    </div>

    <template v-if="metric.valueStyle?.notFitVertical || metric.valueStyle?.notFitHorizontal">
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
        <span>
          <template v-if="isEmpty">
            <span v-if="showEmptyPlaceholder" class="rb-empty">{{ emptyText }}</span>
          </template>
          <template v-else>
            <template v-if="metric.prefix">{{ metric.prefix }}</template>
            {{ displayValue }}
            <template v-if="metric.suffix">{{ metric.suffix }}</template>
          </template>
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
        >
          <template v-if="metric.showLabel">{{ metric.label }}:&nbsp;</template>
          <template v-if="isEmpty">{{ showEmptyPlaceholder ? emptyText : '' }}</template>
          <template v-else>
            <template v-if="metric.prefix">{{ metric.prefix }}</template>
            {{ displayValue }}
            <template v-if="metric.suffix">{{ metric.suffix }}</template>
          </template>
        </text>
      </svg>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, watch, onBeforeUnmount, nextTick, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { fmt } from 'corteza-lib/js/dist'

const { t } = useI18n({ useScope: 'global' })

const props = defineProps({
  metric: { type: Object, required: false, default: () => ({}) },
  options: { type: Object, required: false, default: () => ({}) },
  value: { type: Object, required: false, default: () => ({}) },
  hover: { type: Boolean, required: false, default: false },
  /** 0..1 relative bar length among sibling topK metrics */
  barRatio: { type: Number, required: false, default: 1 },
})

const $Settings = inject('$Settings')

const vvb = ref(['0', '0', '0', '0'])
const metricItem = ref(null)

const getVB = computed(() => vvb.value.join(' '))

const recordStyle = computed(() => props.options.likeRecordList !== false)

const role = computed(() => {
  const r = props.metric.role || 'default'
  return r === 'topK' ? 'balloon' : r
})

const isBalloon = computed(() => role.value === 'balloon')

const showEmptyPlaceholder = computed(() => props.options.showEmptyPlaceholder !== false)

const emptyText = computed(() => t('metric.emptyPlaceholder') || t('record.emptyPlaceholder') || '—')

const isEmpty = computed(() => {
  const v = props.value?.value
  return v === undefined || v === null || v === '' || (typeof v === 'number' && Number.isNaN(v))
})

const displayValue = computed(() => {
  const v = props.value?.value
  if (v === undefined || v === null || v === '') return ''
  const n = typeof v === 'number' ? v : Number(v)
  if (Number.isFinite(n)) return fmt.number(n)
  return String(v)
})

const valueStyle = computed(() => genStyle(props.metric.valueStyle || {}))

const badgeStyle = computed(() => {
  const s = genStyle(props.metric.valueStyle || {})
  // Soft badge: keep threshold color on text/border, light bg
  return {
    ...s,
    backgroundColor: s.backgroundColor && s.backgroundColor !== 'transparent' && s.backgroundColor !== '#FFFFFF00'
      ? s.backgroundColor
      : undefined,
  }
})

const BALLOON_PALETTE = ['#f64e60', '#3699ff', '#43b682', '#ffa800', '#8950fc', '#1bc5bd']

const THRESHOLD_ICONS = {
  'arrow-up': ['fas', 'arrow-up'],
  'arrow-down': ['fas', 'arrow-down'],
  'arrow-right': ['fas', 'arrow-right'],
  alert: ['fas', 'exclamation-triangle'],
  'alert-circle': ['fas', 'exclamation-circle'],
}

/** Highest matching threshold for current value (value ≥ threshold). */
function matchThreshold (rawValue, thresholds = []) {
  if (!thresholds?.length) return undefined
  const n = typeof rawValue === 'number' ? rawValue : Number(rawValue)
  if (!Number.isFinite(n)) return undefined
  const sorted = [...thresholds].sort((a, b) => Number(b.value) - Number(a.value))
  return sorted.find(t => n >= Number(t.value))
}

const matchedThreshold = computed(() =>
  matchThreshold(props.value?.value, props.metric.valueStyle?.colorThresholds || []),
)

const thresholdIcon = computed(() => {
  const key = matchedThreshold.value?.icon
  if (!key) return null
  return THRESHOLD_ICONS[key] || null
})

/** Pick threshold color for current numeric value (highest matching threshold). */
function balloonColorFromThresholds (rawValue, thresholds = []) {
  return matchThreshold(rawValue, thresholds)?.variant
}

const balloonStyle = computed(() => {
  const s = genStyle(props.metric.valueStyle || {})
  const thresholds = props.metric.valueStyle?.colorThresholds || []
  const fromThreshold = balloonColorFromThresholds(props.value?.value, thresholds)

  let bg = fromThreshold || s.backgroundColor
  const transparent = !bg || bg === 'transparent' || bg === '#FFFFFF00' || bg === '#ffffff00'
  if (transparent) {
    const idx = Number(props.metric._balloonIndex ?? props.metric._topKIndex) || 0
    bg = BALLOON_PALETTE[idx % BALLOON_PALETTE.length]
  }
  bg = getColor(bg) || bg

  const ratio = Math.min(1, Math.max(0, Number(props.barRatio) || 0))
  const fullWidth = props.metric.balloonFullWidth === true
  const widthPct = fullWidth ? 100 : (42 + 58 * ratio)
  return {
    backgroundColor: bg,
    color: '#fff',
    fontSize: s.fontSize || '0.9em',
    width: `${widthPct}%`,
    minWidth: fullWidth ? '100%' : '10rem',
    maxWidth: '100%',
    '--balloon-ratio': String(ratio),
  }
})

watch(() => [props.metric, props.options, props.value], () => { update() }, { immediate: true })

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
  if (!value) return value
  if (value[0] === '#') return value
  const themes = themeSettings()
    .filter((theme) => theme.id !== 'general')
    .map((theme) => ({ id: theme.id, values: JSON.parse(theme.values) }))
  return themes[0]?.values?.[value] || value
}

function genStyle (s = {}, forLabel = false) {
  const d = {
    fill: forLabel ? (s.labelColor || s.color) : s.color,
    backgroundColor: s.backgroundColor,
    fontSize: s.fontSize ? s.fontSize + 'px' : undefined,
    color: forLabel ? (s.labelColor || s.color) : s.color,
  }
  if (s.colorThresholds && forLabel === false) {
    const val = props.value?.value
    const { variant } = [...s.colorThresholds].sort((a, b) => b.value - a.value).find(t => val >= t.value) || {}
    if (variant !== undefined) { d.color = variant; d.fill = variant }
  }
  for (const v of Object.keys(d)) { if (d[v] === undefined) delete d[v] }
  if (d.color) d.color = getColor(d.color)
  if (d.backgroundColor) d.backgroundColor = getColor(d.backgroundColor)
  if (d.fill) d.fill = getColor(d.fill)
  return d
}

function setDefaultValues () {
  vvb.value = ['0', '0', '0', '0']
}
</script>

<style scoped lang="scss">
.rb-label {
  color: var(--bs-secondary-color, #6c757d);
  font-size: 0.8rem;
  font-weight: 500;
}

.rb-title {
  font-size: 1.5rem;
  font-weight: 600;
  line-height: 1.3;
  color: var(--bs-body-color, inherit);
}

.rb-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.2rem 0.65rem;
  border-radius: 0.375rem;
  background: var(--bs-tertiary-bg, #f8f9fa);
  border: 1px solid var(--bs-border-color, #dee2e6);
  font-size: 0.8125rem;
}

.rb-badge-label {
  color: var(--bs-secondary-color, #6c757d);
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.02em;
}

.rb-meta-item {
  display: inline-flex;
  align-items: baseline;
  gap: 0.4rem;
  min-width: 0;
}

.rb-meta-label {
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  color: var(--bs-secondary-color, #6c757d);
  white-space: nowrap;
}

.rb-meta-value {
  font-size: 0.875rem;
  min-width: 0;
}

.rb-empty {
  color: var(--bs-secondary-color, #adb5bd);
}

.mb-metric-hero-value {
  font-size: 2rem;
  font-weight: 600;
  line-height: 1.2;
}

.mb-balloon {
  padding: 0.45rem 1.1rem;
  line-height: 1.2;
  box-sizing: border-box;
}

.mb-metric--balloon {
  width: 100%;
}

.mb-balloon-label,
.mb-balloon-value {
  color: inherit;
  font-weight: 600;
  font-size: inherit;
}

.mb-threshold-icon {
  flex-shrink: 0;
  opacity: 0.95;
}

.mb-metric--balloon.mb-metric--drill {
  border-radius: 800px;
  padding-right: 0;

  &:hover {
    filter: brightness(1.06);
    background-color: transparent;
  }
}

.mb-metric--drill {
  cursor: pointer;
  border-radius: 0.25rem;
  transition: background-color 0.15s ease;

  &:hover {
    background-color: var(--bs-tertiary-bg, rgba(0, 0, 0, 0.04));
  }

  &.mb-metric--default,
  &.mb-metric--meta {
    box-shadow: inset -3px 0 0 var(--bs-primary, #0d6efd);
    padding-right: 0.5rem;
  }
}
</style>
