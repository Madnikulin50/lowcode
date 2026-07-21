<template>
  <div class="d-flex align-items-center">
    <span
      v-if="labels?.off"
      class="mb-0 d-inline-block text-primary"
      :class="offClass"
    >
      {{ labels.off }}
    </span>

    <div
      :class="['form-check', switchable ? 'form-switch' : '', 'mb-0']"
    >
      <input
        class="form-check-input"
        type="checkbox"
        :checked="modelValue"
        :role="switchable ? 'switch' : undefined"
        v-bind="$attrs"
        @change="onChange"
      />
      <label class="form-check-label">
        <slot />
      </label>
    </div>

    <span
      v-if="labels?.on"
      class="mb-0 d-inline-block text-primary"
      :class="onClass"
    >
      {{ labels.on }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  modelValue?: boolean
  labels?: Record<string, string>
  invert?: boolean
  switch?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  labels: () => ({}),
  invert: false,
  switch: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const switchable = computed(() => props.switch)

const offClass = computed(() => {
  const value = props.invert ? !props.modelValue : props.modelValue
  return {
    'text-muted': !!value,
    'fw-bold': !value,
    'me-2': !!props.labels?.off,
  }
})

const onClass = computed(() => {
  const value = props.invert ? !props.modelValue : props.modelValue
  return {
    'text-muted': !value,
    'fw-bold': !!value,
    'ms-1': !!props.labels?.on,
  }
})

function onChange (e: Event) {
  const checked = (e.target as HTMLInputElement).checked
  emit('update:modelValue', checked ? !props.invert : props.invert)
}
</script>
