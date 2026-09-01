<template>
  <ListLoader
    kind="record"
    :set="attachmentSet"
    :namespace="namespace"
    :mode="field.options.mode"
    :hide-file-name="field.options.hideFileName"
    :preview-options="options"
    style="min-width: 15rem;"
  />
</template>

<script setup>
import { computed } from 'vue'
import { useViewerBase } from './useViewerBase'
import { compose } from 'corteza-lib/js/dist'
import ListLoader from 'corteza-webapp-compose/src/components/Public/Page/Attachment/ListLoader'

const props = defineProps({
  namespace: { type: compose.Namespace, required: true },
  field: { type: compose.ModuleField, required: true },
  record: { type: compose.Record, required: true },
  valueOnly: { type: Boolean, required: false },
  extraOptions: { type: Object, default: () => ({}) },
  includeStyles: { type: Boolean, default: false },
  disableClick: { type: Boolean, default: false },
})

const { value, options } = useViewerBase(props)

function attachmentIDs (raw) {
  if (raw == null || raw === '') return []
  const list = Array.isArray(raw) ? raw : [raw]
  return list
    .flatMap(v => Array.isArray(v) ? v : [v])
    .map(v => {
      if (v == null || v === '') return ''
      if (typeof v === 'object') return v
      return String(v)
    })
    .filter(v => v !== '')
}

const attachmentSet = computed(() => attachmentIDs(value.value))
</script>
