<template>
  <div>
    <div>{{ formatted }}</div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useViewerBase } from './useViewerBase'
import { compose, NoID } from 'corteza-lib/js/dist'
import { useUserStore } from '../../../store/user'

const props = defineProps({
  namespace: { type: compose.Namespace, required: true },
  field: { type: compose.ModuleField, required: true },
  record: { type: compose.Record, required: true },
  valueOnly: { type: Boolean, required: false },
  extraOptions: { type: Object, default: () => ({}) },
  includeStyles: { type: Boolean, default: false },
  disableClick: { type: Boolean, default: false },
})

const { t } = useI18n({ useScope: 'global', messages: {} })
const { value } = useViewerBase(props)

const userStore = useUserStore()

const formatted = computed(() => {
  const f = (u) => {
    if (u === NoID) return
    if (typeof u === 'string') u = userStore.findByID(u)
    if (!u) return
    return props.field.formatter(u)
  }
  if (props.field.isMulti) {
    return ((value.value) || []).map(v => f(v) || t('kind.user.na')).join(props.field.options.multiDelimiter)
  }
  return f(value.value) || t('kind.user.na')
})
</script>
