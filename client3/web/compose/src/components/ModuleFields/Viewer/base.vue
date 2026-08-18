<template>
  <div>
    <div :class="classes">{{ formatted }}</div>
  </div>
</template>

<script setup>
import { computed, inject } from 'vue'
import { useRoute } from 'vue-router'
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

const route = useRoute()
const $settings = inject('$Settings')

const value = computed(() => {
  if (props.field.isSystem) return props.record[props.field.name]
  return props.record ? props.record.values[props.field.name] : undefined
})

const formatted = computed(() => {
  if (props.field.isMulti) return value.value ? value.value.join(props.field.options.multiDelimiter) : ''
  return value.value
})

const classes = computed(() => {
  const cls = []
  const { fieldID } = props.field
  const { textStyles = {} } = props.extraOptions
  if (props.field.isMulti) cls.push('multiline')
  else if (props.includeStyles) {
    if (!textStyles.wrappedFields || !textStyles.wrappedFields.includes(fieldID)) cls.push('text-nowrap')
  }
  return cls
})

const options = computed(() => props.field.options)
const inModal = computed(() => {
  const { recordPageID, magnifiedBlockID } = route.query
  return !!recordPageID || !!magnifiedBlockID
})

const themeSettings = computed(() => {
  return $settings?.get ? $settings.get('ui.studio.themes', []) : []
})

function getColor (value) {
  if (!value) return undefined
  if (value[0] === '#') return value
  const themes = themeSettings.value
    .filter(theme => theme.id !== 'general')
    .map(theme => ({ id: theme.id, values: typeof theme.values === 'string' ? JSON.parse(theme.values) : theme.values }))
  return themes[0]?.values?.[value] || value
}
</script>

<style>
</style>
