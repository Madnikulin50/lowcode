<template>
  <div
    v-if="visible"
    ref="anchor"
    class="block-help-wrap d-print-none"
    :class="wrapClass"
    @mouseenter="onEnter"
    @mouseleave="onLeave"
    @focusin="onEnter"
    @focusout="onLeave"
  >
    <button
      type="button"
      :class="buttonClass"
      :aria-label="help.labels.title"
      @click.stop.prevent="openPanel"
    >
      <font-awesome-icon :icon="['far', 'question-circle']" />
    </button>
    <c-help-panel
      v-model:open="open"
      v-bind="help.panelProps"
    />
  </div>

  <Teleport to="body">
    <div
      v-if="tipVisible && hoverText"
      class="c-help-hover-tip"
      :style="tipStyle"
      role="tooltip"
    >
      {{ hoverText }}
    </div>
  </Teleport>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useHelp } from '../../../composables/useHelp'

const props = defineProps({
  block: { type: Object, required: true },
  offset: { type: Boolean, default: false },
  title: { type: String, default: '' },
  description: { type: String, default: '' },
  help: { type: String, default: '' },
  variant: {
    type: String,
    default: 'float',
    validator: (v) => ['float', 'metric', 'chart'].includes(v),
  },
})

const open = ref(false)
const tipVisible = ref(false)
const anchor = ref(null)
const tipStyle = ref({})

const helpText = computed(() => String(props.help || props.block?.options?.help || '').trim())

const visible = computed(() => {
  if (!helpText.value) return false
  // Explicit help (chart/metric docs) always shows; hideHelpButton is for block.options.help
  if (props.help) return true
  return !props.block?.options?.hideHelpButton
})

const hoverText = computed(() => String(
  props.description || props.block?.description || '',
).trim())

const wrapClass = computed(() => ({
  'block-help-float': props.variant === 'float',
  'block-help-float-offset': props.variant === 'float' && props.offset,
  'block-help-metric': props.variant === 'metric',
  'block-help-metric-offset': props.variant === 'metric' && props.offset,
  'block-help-chart': props.variant === 'chart',
  'block-help-chart-offset': props.variant === 'chart' && props.offset,
}))

const buttonClass = computed(() => {
  if (props.variant === 'metric') {
    return 'btn btn-outline-light text-secondary border-0 btn-sm'
  }
  if (props.variant === 'chart') {
    return 'btn btn-outline-light d-flex d-print-none border-0 px-1 text-secondary'
  }
  return 'block-help-button'
})

const help = useHelp('', computed(() => ({
  title: props.title || props.block?.title || '',
  description: hoverText.value,
  help: helpText.value,
})), { includeProduct: false })

function positionTip () {
  const el = anchor.value
  if (!el) return
  const r = el.getBoundingClientRect()
  const maxWidth = 320
  const left = Math.min(Math.max(8, r.left), window.innerWidth - maxWidth - 8)
  tipStyle.value = {
    top: `${Math.round(r.bottom + 8)}px`,
    left: `${Math.round(left)}px`,
    maxWidth: `${maxWidth}px`,
  }
}

function onEnter () {
  if (!hoverText.value || open.value) return
  positionTip()
  tipVisible.value = true
}

function onLeave () {
  tipVisible.value = false
}

function openPanel () {
  tipVisible.value = false
  open.value = true
}
</script>

<style scoped>
.block-help-float {
  position: absolute;
  z-index: 1041;
  top: 0.5rem;
  right: 0.5rem;
}
.block-help-float-offset {
  right: 3rem;
}
.block-help-button {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  border-radius: 50%;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.15);
  border: 1px solid var(--bs-border-color, #dee2e6);
  color: var(--secondary);
  font-size: 16px;
  cursor: pointer;
  padding: 0;
}
.block-help-button:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.block-help-metric {
  position: absolute;
  top: 20px;
  right: 0;
  transform: translateY(-50%);
  z-index: 10;
}
.block-help-metric-offset {
  right: 1.75rem;
}

.block-help-chart {
  position: absolute;
  right: 0.5rem;
  top: 0.7rem;
  z-index: 10;
}
.block-help-chart-offset {
  right: 2.2rem;
}
</style>
