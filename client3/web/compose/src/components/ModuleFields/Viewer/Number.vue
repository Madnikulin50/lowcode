<template>
  <div>
    <div v-if="field.options.display === 'number'" :class="classes">{{ formatted }}</div>

    <template v-if="field.options.display === 'progress'">
      <CProgress
        v-for="(v, i) in formatted"
        :key="i"
        :value="parseFloat(v)"
        :min="parseFloat(field.options.min)"
        :max="parseFloat(field.options.max)"
        :labeled="field.options.showValue"
        :relative="field.options.showRelative"
        :progress="field.options.showProgress"
        :striped="field.options.striped"
        :animated="field.options.animated"
        :variant="field.options.variant"
        :thresholds="field.options.thresholds"
        size="sm"
        :class="{ 'mt-2': i }"
      />
    </template>

    <template v-if="field.options.display === 'colorGrade'">
      <CColorGradeNumber
        :formatted="formatted"
        :value="parseFloat(value)"
        :variant="field.options.variant"
        :thresholds="field.options.thresholds"
        :class="{ 'mt-2': false }"
      />
    </template>

    <template v-if="field.options.display === 'trafficLight'">
      <CTrafficLight
        :formatted="formatted"
        :value="parseFloat(value)"
        :variant="field.options.variant"
        :thresholds="field.options.thresholds"
        :class="{ 'mt-2': false }"
        style="height: 2rem; min-width: 5rem;"
      />
    </template>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useViewerBase } from './useViewerBase'
import { compose } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'
const { CProgress, CColorGradeNumber, CTrafficLight } = components

const props = defineProps({
  namespace: { type: compose.Namespace, required: true },
  field: { type: compose.ModuleField, required: true },
  record: { type: Object, required: true },
  valueOnly: { type: Boolean, required: false },
  extraOptions: { type: Object, default: () => ({}) },
  includeStyles: { type: Boolean, default: false },
  disableClick: { type: Boolean, default: false },
})

const { value, classes } = useViewerBase(props)

const formatted = computed(() => {
  if (value.value === undefined) {
    return props.field.options.display === 'number' ? undefined : [props.field.options.min]
  }
  const v = props.field.isMulti ? value.value : [value.value]
  if (props.field.options.display !== 'progress') {
    return v.map(val => props.field.formatValue(val)).join(props.field.options.multiDelimiter)
  }
  return v.length ? v : [props.field.options.min]
})
</script>
