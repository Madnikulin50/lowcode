<template>
  <div>
    <p
      class="mb-0"
      v-html="message"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
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

onMounted(() => {
  emit('submit', { keep: true })
})
</script>
