<template>
  <div class="mb-3">
    <label class="form-label">{{ label }}</label>
    <textarea
      class="form-control"
      :value="value"
      @input="emit('update', ($event.target as HTMLTextAreaElement).value)"
    />
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
