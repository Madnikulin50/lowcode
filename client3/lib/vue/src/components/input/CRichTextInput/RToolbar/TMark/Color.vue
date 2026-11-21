<template>
  <div>
    <button
      :id="`color-popover-${format.type}`"
      class="btn btn-link text-dark fw-bold text-decoration-none mb-1"
      @click.stop.prevent="showPicker"
    >
      <span
        :style="{
          backgroundColor: background ? selectedColor : 'transparent',
          'border-bottom': background ? 'none' : `2px solid ${selectedColor}`,
        }"
      >
        A
      </span>
    </button>

    <c-input-color-picker
      ref="picker"
      class="d-none"
      :value="selectedColor"
      :default-value="getDefaultColor()"
      :show-text="false"
      :width="'0px'"
      :height="'0px'"
      @input="applyFromPicker"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import CInputColorPicker from '../../../CInputColorPicker.vue'

const props = defineProps<{
  editor: any
  format: any
  isActive?: any
  getMarkAttrs?: (...args: any[]) => any
  currentValue?: string
  background?: boolean
}>()

const emit = defineEmits<{
  (e: 'click', payload: { type: string; attrs: Record<string, any> }): void
}>()

const picker = ref<any>(null)

const selectedColor = ref(getDefaultColor())

function getComputedColor(cssVar: string) {
  try {
    const computedStyle = getComputedStyle(document.documentElement)
    return computedStyle.getPropertyValue(cssVar).trim()
  } catch (error) {
    return null
  }
}

function showPicker() {
  if (picker.value && picker.value.openMenu) {
    picker.value.openMenu()
  }
}

function applyFromPicker(val: string) {
  if (!val) return
  selectedColor.value = val
  onClick(props.format.type, { color: val })
}

function getDefaultColor() {
  return props.background ? getComputedColor('--white') : getComputedColor('--dark')
}

function onClick(type: string, attrs: Record<string, any>) {
  emit('click', { type, attrs })
}

function activeClasses(attrs?: Record<string, any>) {
  const isActive = props.editor.isActive(props.format.type, attrs)
  if (isActive) {
    return ['text-primary']
  }
  return undefined
}
</script>
