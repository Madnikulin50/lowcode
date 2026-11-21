<template>
  <div :class="classes">
    <span
      v-for="(v, index) of formattedValue"
      :key="index"
      :class="{ 'd-block': field.options.multiDelimiter === '\n' }"
    >
      <span v-if="field.options.outputPlain || disableClick">
        {{ fixUrl(v) }}{{ index !== formattedValue.length - 1 ? field.options.multiDelimiter : '' }}
      </span>
      <a
        v-else
        :href="fixUrl(v)"
        target="_blank"
        rel="noopener noreferrer"
        @click.stop
      >
        {{ fixUrl(v) }}{{ index !== formattedValue.length - 1 ? field.options.multiDelimiter : '' }}
      </a>
    </span>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useViewerBase } from './useViewerBase'
import { compose } from 'corteza-lib/js/dist'
import { trimUrlFragment, trimUrlQuery, trimUrlPath, onlySecureUrl } from '../url'

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

function fixUrl (value) {
  if (props.field.options.trimFragment) value = trimUrlFragment(value)
  if (props.field.options.trimQuery) value = trimUrlQuery(value)
  if (props.field.options.trimPath) value = trimUrlPath(value)
  if (props.field.options.onlySecure) value = onlySecureUrl(value)
  return value
}
</script>
