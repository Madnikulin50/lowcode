<template>
  <button
    :class="['btn', `btn-${options?.variant || 'primary'}`, options?.size ? `btn-${options.size}` : '']"
    :disabled="options?.disabled"
    @click="emit('action')"
  >
    <span v-html="label" />
  </button>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  options?: Record<string, any>
  label?: string
}>(), {
  options: () => ({}),
  label: undefined,
})

const emit = defineEmits<{
  (e: 'action'): void
}>()
</script>
