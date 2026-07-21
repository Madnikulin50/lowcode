<template>
  <div class="d-flex flex-column gap-1">
    <p
      v-if="!!message"
      class="text-break"
      v-html="message"
    />
    <div class="d-flex flex-wrap align-items-center gap-2">
      <button
        :class="['btn', `btn-${pVal('confirmButtonVariant', 'primary')}`]"
        :disabled="loading"
        class="flex-grow-1"
        @click="emit('submit', { value: pRaw('confirmButtonValue', true, 'Boolean') })"
      >
        {{ pVal('confirmButtonLabel', 'Yes') }}
      </button>
      <button
        :disabled="loading"
        :class="['btn', `btn-${pVal('rejectButtonVariant', 'light')}`]"
        class="flex-grow-1"
        @click="emit('submit', { value: pRaw('rejectButtonValue', false, 'Boolean') })"
      >
        {{ pVal('rejectButtonLabel', 'No') }}
      </button>
    </div>
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
</script>
