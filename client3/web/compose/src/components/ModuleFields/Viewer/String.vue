<template>
  <div class="rt-content">
    <p
      v-if="formatted"
      :style="{ 'white-space': field.options.useRichTextEditor ? 'pre-line' : undefined }"
      :class="[{ multiline: field.isMulti || field.options.multiLine }, ...classes]"
      v-html="formatted"
    />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useViewerBase } from './useViewerBase'
import { compose } from 'corteza-lib/js/dist'

const props = defineProps({
  namespace: { type: compose.Namespace, required: true },
  field: { type: compose.ModuleField, required: true },
  record: { type: compose.Record, required: true },
  valueOnly: { type: Boolean, required: false },
  extraOptions: { type: Object, default: () => ({}) },
  includeStyles: { type: Boolean, default: false },
  disableClick: { type: Boolean, default: false },
})

const { formatted, classes } = useViewerBase(props)
</script>
