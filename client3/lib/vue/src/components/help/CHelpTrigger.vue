<template>
  <div
    v-if="hintOnly"
    ref="anchor"
    class="d-flex align-items-center ms-1 c-help-trigger"
    :aria-label="hoverText"
    @mouseenter="onEnter"
    @mouseleave="onLeave"
    @focusin="onEnter"
    @focusout="onLeave"
  >
    <font-awesome-icon
      :icon="icon"
      :class="iconClass"
    />
  </div>
  <div
    v-else-if="hasContent"
    ref="anchor"
    class="d-inline-flex align-items-center c-help-trigger"
    @mouseenter="onEnter"
    @mouseleave="onLeave"
    @focusin="onEnter"
    @focusout="onLeave"
  >
    <button
      type="button"
      class="btn btn-link p-0 border-0 d-flex align-items-center text-decoration-none"
      :class="buttonClass"
      :aria-label="labels.title"
      @click.stop.prevent="openPanel"
    >
      <font-awesome-icon
        :icon="icon"
        :class="iconClass"
      />
      <span
        v-if="label"
        class="ms-1 small"
      >{{ label }}</span>
    </button>
    <c-help-panel
      v-model:open="open"
      :title="title"
      :description="description"
      :body-html="bodyHtml"
      :product-hint="productHint"
      :product-html="productHtml"
      :labels="labels"
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

<script setup lang="ts">
import { computed, ref } from 'vue'
import CHelpPanel from './CHelpPanel.vue'

const props = defineProps({
  icon: {
    type: Array,
    default: () => ['far', 'question-circle'],
  },
  iconClass: {
    type: String,
    default: 'text-primary',
  },
  buttonClass: {
    type: String,
    default: '',
  },
  label: {
    type: String,
    default: '',
  },
  title: { type: String, default: '' },
  hint: { type: String, default: '' },
  description: { type: String, default: '' },
  bodyHtml: { type: String, default: '' },
  productHint: { type: String, default: '' },
  productHtml: { type: String, default: '' },
  labels: {
    type: Object,
    default: () => ({
      title: 'Help',
      app: 'About this',
      product: 'How to use',
      close: 'Close',
    }),
  },
})

const open = ref(false)
const tipVisible = ref(false)
const anchor = ref<HTMLElement | null>(null)
const tipStyle = ref<Record<string, string>>({})

const hasPanel = computed(() => !!(
  props.description ||
  props.bodyHtml ||
  props.productHtml
))
const hasContent = computed(() => !!(props.hint || hasPanel.value || props.productHint))
const hintOnly = computed(() => !!(props.hint && !hasPanel.value && !props.productHtml))
const hoverText = computed(() => (props.hint || props.description || '').trim())

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

<style>
.c-help-hover-tip {
  position: fixed;
  z-index: 1090;
  padding: 0.5rem 0.75rem;
  background: rgba(33, 37, 41, 0.95);
  color: #fff;
  border-radius: 0.375rem;
  font-size: 0.8125rem;
  line-height: 1.4;
  white-space: pre-wrap;
  box-shadow: 0 0.25rem 0.75rem rgba(0, 0, 0, 0.2);
  pointer-events: none;
}
</style>
