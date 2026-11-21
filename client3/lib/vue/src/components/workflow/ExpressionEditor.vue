<template>
  <div class="position-relative">
    <c-ace-editor
      :model-value="expressionValue"
      :lang="lang"
      :min-height="minHeight"
      :show-line-numbers="showLineNumbers"
      :font-size="fontSize"
      :show-popout="showPopout"
      :auto-complete="autoComplete"
      :border="border"
      :auto-complete-suggestions="expressionAutoCompleteValues"
      resizable
      @update:model-value="emitValue"
    />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { components } from 'corteza-lib/vue/dist'
import { EXPRESSION_EDITOR_AUTO_COMPLETE_VALUES } from './lib/editor-auto-complete.js'

const { CAceEditor } = components

const props = defineProps({
  value: { type: String, default: '' },
  modelValue: { type: String, default: '' },
  lang: { type: String, default: 'text' },
  minHeight: { type: String, default: '6rem' },
  showLineNumbers: { type: Boolean, default: false },
  fontSize: { type: String, default: '14px' },
  border: { type: Boolean, default: true },
  showPopout: { type: Boolean, default: true },
  autoComplete: { type: Boolean, default: true },
})

const emit = defineEmits(['update:value', 'update:modelValue'])

const expressionAutoCompleteValues = EXPRESSION_EDITOR_AUTO_COMPLETE_VALUES

const expressionValue = computed({
  get: () => props.modelValue || props.value,
  set: (val) => {
    emitValue(val || '')
  },
})

function emitValue(value = '') {
  emit('update.value', value)
  emit('update.modelValue', value)
}
</script>

<style lang="scss" scoped>
.popout {
  z-index: 7;
  bottom: 0;
  right: 0;
}
</style>
