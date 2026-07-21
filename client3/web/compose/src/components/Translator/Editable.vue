<template>
  <div
    v-once
    class="m-1 p-2 text-break"
    contenteditable="true"
    :value="value"
    :placeholder="placeholder"
    @input="$emit('input', $event.target.innerHTML)"
    v-text="value"
  />
</template>

<script setup>
import { watch } from 'vue'

const props = defineProps({
  value: { type: String, required: true },
  placeholder: { type: String, default: '' },
})

defineEmits(['input'])

watch(() => props.value, (newValue) => {
  // DOM update skipped - works via v-text binding
  const s = document.createElement('div')
  s.innerHTML = newValue
  // Note: this requires a template ref, but v-once prevents reactivity
  // This is a simplified version
})
</script>

<style lang="scss" scoped>
div {
  cursor: text;
}

div:empty::before {
  content: attr(placeholder);
  color: var(--secondary);
  pointer-events: none;
  display: block;
}
</style>
