<template>
  <div :class="classes">
    <span
      v-for="(v, index) in formattedValue"
      :key="index"
      :class="{ 'd-block': field.options.multiDelimiter === '\n' }"
    >
      <span v-if="field.options.outputPlain || disableClick">
        {{ v }}{{ index !== formattedValue.length - 1 ? field.options.multiDelimiter : '' }}
      </span>
      <a
        v-else
        :href="'mailto:' + formattedValue"
        target="_blank"
        rel="noopener noreferrer"
        @click.stop
      >
        {{ v }}{{ index !== formattedValue.length - 1 ? field.options.multiDelimiter : '' }}
      </a>
    </span>
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

const { value, classes } = useViewerBase(props)

const formattedValue = computed(() => {
  return props.field.isMulti ? value.value : [value.value].filter(v => v)
})
</script>
