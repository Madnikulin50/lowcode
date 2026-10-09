<template>
  <c-input-select
    data-test-id="select-sens-lvl"
    :value="_value"
    :disabled="_disabled"
    :options="sensitivityLevels"
    :get-option-label="getLabel"
    :get-option-key="getOptionKey"
    :placeholder="placeholder"
    :reduce="l => l.sensitivityLevelID"
    append-to-body
    @input="onInput"
  />
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { NoID } from 'corteza-lib/js/dist'
import CInputSelect from '../input/CInputSelect.vue'

const props = withDefaults(defineProps<{
  value?: string
  options: Array<{ sensitivityLevelID: string; level: number; handle: string; meta?: Record<string, string> }>
  placeholder?: string
  maxLevel?: string
  disabled?: boolean
}>(), {
  value: '',
  placeholder: '',
  disabled: false,
})

const emit = defineEmits<{
  (e: 'input', value: string): void
}>()

const _value = computed(() => props.value === NoID ? undefined : props.value)

const _disabled = computed(() => props.disabled || props.maxLevel === NoID)

const sensitivityLevels = computed(() => {
  if (props.maxLevel === NoID) return []
  if (props.maxLevel) {
    const maxLevelConnection = props.options.find(({ sensitivityLevelID }) => sensitivityLevelID === props.maxLevel)
    if (maxLevelConnection) {
      return props.options.filter(({ level }) => level <= maxLevelConnection.level)
    }
  }
  return props.options
})

watch(sensitivityLevels, () => {
  const isValueCompatible = sensitivityLevels.value.some(({ sensitivityLevelID }) => sensitivityLevelID === props.value)
  if (!isValueCompatible) {
    emit('input', NoID)
  }
}, { immediate: true })

function getLabel({ handle, meta = {} }: { handle: string; meta?: Record<string, string> }): string {
  return meta.name || handle
}

function onInput(sensitivityLevelID: string | undefined): void {
  emit('input', sensitivityLevelID || NoID)
}

function getOptionKey({ sensitivityLevelID }: { sensitivityLevelID: string }): string {
  return sensitivityLevelID
}
</script>

<style>

</style>
