<template>
  <div>
    <span
      v-for="(v, index) of resolvedValue"
      :key="index"
      :class="{ 'd-block': field.options.multiDelimiter === '\n' }"
    >
      <span
        :class="{ 'badge rounded-pill': field.options.displayType === 'badge', 'mt-1': field.options.multiDelimiter === '\n' && index !== 0 }"
        :style="v.style"
      >
        {{ v.text }}
      </span>
      {{ index !== resolvedValue.length - 1 ? field.options.multiDelimiter : '' }}
    </span>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useViewerBase } from './useViewerBase'
import { compose } from 'corteza-lib/js/dist'
import { badgeGradient } from 'corteza-webapp-compose/src/lib/color.js'

const props = defineProps({
  namespace: { type: compose.Namespace, required: true },
  field: { type: compose.ModuleField, required: true },
  record: { type: compose.Record, required: true },
  valueOnly: { type: Boolean, required: false },
  extraOptions: { type: Object, default: () => ({}) },
  includeStyles: { type: Boolean, default: false },
  disableClick: { type: Boolean, default: false },
})

const { getColor } = useViewerBase(props)

function getFieldValue () {
  let v
  if (props.field.isSystem) v = props.record[props.field.name]
  v = props.record ? props.record.values[props.field.name] : undefined
  return v
}

const resolvedValue = computed(() => {
  const v = getFieldValue()
  if (props.field.isMulti) {
    if (!Array.isArray(v)) return []
    return v.map(val => resolveValue(val) || val).filter(val => val && val.text)
  }
  return [resolveValue(v) || v].filter(val => val && val.text)
})

function resolveValue (v) {
  const opts = props.field.options?.options || []
  const opt = opts.find(({ value }) => value === v) || { text: v, style: {} }
  return { text: opt.text || v, style: getOptionStyle(opt) }
}

function getOptionStyle (opt) {
  const style = {}
  if (props.field.options.displayType === 'badge') {
    const oStyle = opt.style || {}
    style.fontSize = '0.9rem'
    const fg = getColor(oStyle.textColor) || 'var(--dark)'
    const bg = getColor(oStyle.backgroundColor) || 'var(--extra-light)'
    style.color = fg
    const gradient = props.field.options.badgeGradient ? badgeGradient(bg) : undefined
    if (gradient) {
      style.background = gradient
    } else {
      style.backgroundColor = bg
    }
  }
  return style
}
</script>

<style lang="scss" scoped>
.badge {
  font-family: var(--font-medium);
}
</style>
