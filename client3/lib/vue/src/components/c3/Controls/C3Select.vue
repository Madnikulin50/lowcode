<template>
  <div class="mb-3">
    <label class="form-label">{{ label }}</label>
    <select
      class="form-select"
      :value="value"
      @change="emit('update', ($event.target as HTMLSelectElement).value)"
    >
      <option
        v-for="opt in options"
        :key="opt.value"
        :value="opt.value"
        :disabled="opt.disabled"
      >
        {{ opt.text }}
      </option>
    </select>
  </div>
</template>

<script setup lang="ts">
interface SelectOption {
  value: string
  text: string
  disabled?: boolean
}

const props = defineProps<{
  label: string
  value?: string
  options?: SelectOption[]
}>()

const emit = defineEmits<{
  (e: 'update', value: string): void
}>()
</script>
