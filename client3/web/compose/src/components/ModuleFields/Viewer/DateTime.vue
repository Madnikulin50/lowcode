<template>
  <div>
    <div>{{ formatted }}</div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useViewerBase } from './useViewerBase'
import { compose } from 'corteza-lib/js/dist'

const props = defineProps({
  namespace: { type: compose.Namespace, required: true },
  field: { type: compose.ModuleField, required: true },
  record: { type: Object, required: true },
  valueOnly: { type: Boolean, required: false },
  extraOptions: { type: Object, default: () => ({}) },
  includeStyles: { type: Boolean, default: false },
  disableClick: { type: Boolean, default: false },
})

const { value } = useViewerBase(props)

const formatted = computed(() => {
  if (props.field.isMulti && value.value) {
    return value.value.map(v => props.field.formatValue(v)).join(props.field.options.multiDelimiter)
  }
  return value.value ? props.field.formatValue(value.value) : null
})
</script>
