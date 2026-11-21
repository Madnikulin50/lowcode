<template>
  <div :class="['position-relative', attrs.class]">
    <c-ace-editor
      v-model="editorValue"
      auto-complete
      init-expressions
      :auto-complete-suggestions="autoCompleteSuggestions"
      v-bind="omitClass(attrs)"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, useAttrs } from 'vue'
import CAceEditor from './CAceEditor.vue'

defineOptions({ inheritAttrs: false })

const props = withDefaults(defineProps<{
  modelValue?: string
  lang?: string
  minHeight?: string
  showLineNumbers?: boolean
  fontSize?: string
  border?: boolean
  showPopout?: boolean
  readOnly?: boolean
  highlightActiveLine?: boolean
  showPrintMargin?: boolean
  suggestionParams?: any[]
  fontFamily?: string
  placeholder?: string
  resizable?: boolean
}>(), {
  modelValue: '',
  lang: 'text',
  minHeight: '2.35rem',
  showLineNumbers: false,
  fontSize: '14px',
  border: true,
  showPopout: false,
  readOnly: false,
  highlightActiveLine: false,
  showPrintMargin: false,
  suggestionParams: () => [],
  fontFamily: '',
  placeholder: '',
  resizable: true,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const attrs = useAttrs()

function omitClass (a: Record<string, unknown>) {
  const { class: _, ...rest } = a
  return rest
}

const editorValue = computed({
  get: () => props.modelValue,
  set: (value = '') => emit('update:modelValue', value),
})

const autoCompleteSuggestions = computed(() => {
  const params = Array.isArray(props.suggestionParams) ? props.suggestionParams : []
  return getRecordBasedSuggestions(params)
})

function getRecordBasedSuggestions(params: any[] = []): Record<string, any[]> {
  const result: Record<string, any[]> = {}

  function addSuggestion(key: string, caption: string, value: string) {
    if (!result[key]) result[key] = []
    result[key].push({ caption, value })
  }

  function processProperties(prefix: string, properties: any[], interpolate: boolean) {
    (properties || []).forEach((prop: any) => {
      if (typeof prop === 'string') {
        const value = prefix + '.' + prop + (interpolate ? '}' : '')
        addSuggestion(prefix, prop, value)
      } else {
        const nestedPrefix = prefix + '.' + prop.value + '.'
        addSuggestion(prefix, prop.value, nestedPrefix)

        if (prop.properties) {
          (prop.properties || []).forEach((nestedProp: any) => {
            const nestedValue = nestedPrefix + (typeof nestedProp === 'string' ? nestedProp : nestedProp) + (interpolate ? '}' : '')
            addSuggestion(prefix + '.' + prop.value, typeof nestedProp === 'string' ? nestedProp : nestedProp, nestedValue)
          })
        }
      }
    })
  }

  (params || []).forEach((p: any) => {
    if (typeof p === 'string') {
      addSuggestion('', '', p)
    } else {
      const { interpolate = false, properties = [], value, root = true } = p
      const prefix = interpolate ? '${' : ''
      const suffix = interpolate && !properties.length ? '}' : ''
      const prefixAsValue = prefix + value + suffix + (properties.length > 0 ? '.' : '')

      if (root) {
        addSuggestion('', '', prefixAsValue)
      }

      if (interpolate) {
        addSuggestion('$', prefixAsValue.slice(1), prefixAsValue)
        addSuggestion('${', prefixAsValue.slice(2), prefixAsValue)
      }

      if (properties.length) {
        processProperties(prefix + value, properties, interpolate)
      }
    }
  })

  return result
}
</script>
