<template>
  <div>
    <div>{{ formatted }}</div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
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

const { t } = useI18n({ useScope: 'global', messages: {} })
const { value } = useViewerBase(props)

const formatted = computed(() => {
  const { trueLabel, falseLabel } = props.field.options
  if (value.value === '1') return trueLabel || t('label.yes')
  if (value.value === '0') return falseLabel || t('label.no')
  return ''
})
</script>
