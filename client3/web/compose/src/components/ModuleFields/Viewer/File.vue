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

const attachmentSet = computed(() => {
  return props.field.isMulti ? value.value : [value.value]
})
</script>
