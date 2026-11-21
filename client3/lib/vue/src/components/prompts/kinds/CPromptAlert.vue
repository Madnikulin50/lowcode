<template>
  <div class="d-flex flex-column gap-1">
    <p
      v-if="!!message"
      class="text-break"
      v-html="message"
    />
    <button
      :class="['btn', `btn-${pVal('buttonVariant', 'primary')}`]"
      :disabled="loading"
      class="ms-auto"
      @click="emit('submit', { confirmed: pRaw(undefined, true, 'Boolean') })"
    >
      {{ pVal('buttonLabel', 'Ok') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { pVal as _pVal, pType as _pType } from '../utils'

const props = withDefaults(defineProps<{
  loading?: boolean
  payload?: Record<string, any>
}>(), {
  loading: false,
  payload: () => ({}),
})

const emit = defineEmits<{
  (e: 'submit', value: Record<string, any>): void
}>()

const message = computed(() => _pVal(props.payload, 'message', ''))
const label = computed(() => _pVal(props.payload, 'label', ''))

function pVal(k: string, def?: any) {
  return _pVal(props.payload, k, def)
}

function pType(k: string, def?: any) {
  return _pType(props.payload, k, def)
}

function pRaw(k?: string, defValue?: any, defType?: string) {
  if (k && props.payload && props.payload[k] && props.payload[k] !== undefined) {
    return props.payload[k]
  }
  return { '@type': defType, '@value': defValue }
}
</script>
