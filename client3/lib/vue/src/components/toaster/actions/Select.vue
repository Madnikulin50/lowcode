<template>
  <select
    class="form-select form-control w-auto"
    v-bind="options"
    :value="null"
    @change="emit('action', ($event.target as HTMLSelectElement).value)"
  >
    <option
      :value="null"
      disabled
      v-html="label"
    />
    <option
      v-for="(opt, i) in options?.items || []"
      :key="i"
      v-bind="opt"
      :value="opt.value"
      v-html="opt.label"
    />
  </select>
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
  (e: 'action', value?: any): void
}>()
</script>
