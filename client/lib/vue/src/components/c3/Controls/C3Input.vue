<template>
  <div class="mb-3">
    <label class="form-label">{{ label }}</label>
    <input
      type="text"
      class="form-control"
      :value="value"
      @input="emit('update', ($event.target as HTMLInputElement).value)"
    >
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  label: string
  value?: string
}>()

const emit = defineEmits<{
  (e: 'update', value: string): void
}>()
</script>
