<template>
  <div class="rt-content">
    <p class="text-muted font-monospace">{{ display }}</p>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'field' } })
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
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

const { t } = useI18n({ useScope: 'global' })
const { value } = useViewerBase(props)

// API отдаёт маску; реальное значение доступно только через exposeSecrets.
const display = computed(() => {
  if (value.value === undefined || value.value === null || value.value === '') {
    return t('kind.secret.empty')
  }
  return '••••••••'
})
</script>
