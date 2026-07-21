<template>
  <button
    :class="[`btn btn-${variant}`, btnSize, 'd-inline-flex align-items-center justify-content-center']"
    :disabled="disabled || processing"
  >
    <slot />
    <span
      v-if="processing"
      class="spinner-border spinner-border-sm ms-1"
    />
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  variant?: string
  size?: string
  processing?: boolean
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary',
  size: 'md',
  processing: false,
  disabled: false,
})

const btnSize = computed(() => props.size === 'sm' ? 'btn-sm' : props.size === 'lg' ? 'btn-lg' : '')
</script>
