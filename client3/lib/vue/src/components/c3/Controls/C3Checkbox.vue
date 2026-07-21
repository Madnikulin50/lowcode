<template>
  <div class="form-check">
    <input
      type="checkbox"
      class="form-check-input"
      :checked="value"
      @change="emit('update', ($event.target as HTMLInputElement).checked)"
    >
    <label class="form-check-label">{{ label }}</label>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  label: string
  value?: boolean
}>()

const emit = defineEmits<{
  (e: 'update', value: boolean): void
}>()
</script>
