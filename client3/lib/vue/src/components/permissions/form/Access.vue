<template>
  <div
    class="btn-group access rounded"
    data-bs-toggle="buttons"
    role="group"
  >
    <template
      v-for="opt in options"
      :key="opt.value"
    >
      <input
        :id="`access-${opt.value}`"
        type="radio"
        class="btn-check"
        :name="'access-radio'"
        :value="opt.value"
        :checked="selected === opt.value"
        :disabled="!enabled"
        @change="onChange(opt.value)"
      >
      <label
        :for="`access-${opt.value}`"
        class="btn"
        :class="selected === opt.value ? `btn-${variant}` : 'btn-outline-primary'"
      >
        {{ opt.text }}
      </label>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  enabled?: boolean
  access?: string
  current?: string
}>()

const emit = defineEmits<{
  (e: 'update', value: string): void
  (e: 'update:access', value: string): void
}>()

const options = computed(() => {
  return ['allow', 'inherit', 'deny'].map(value => ({
    value,
    text: 'permissions:ui.access.' + value,
  }))
})

const isChanged = computed(() => selected.value !== props.current)

const variant = computed(() => isChanged.value ? 'outline-warning' : 'outline-primary')

const selected = computed({
  get: () => props.access,
  set: (sel: string) => {
    if (props.access !== sel) {
      emit('update', sel)
    }
    emit('update:access', sel)
  },
})

function onChange(value: string) {
  selected.value = value
}
</script>

<style lang="scss">
.access {
  .btn {
    background-color: var(--light);
    border: none;
  }

  .btn:nth-child(2), .btn:nth-child(3) {
    margin-left: 0.2rem !important;
  }
}
</style>
