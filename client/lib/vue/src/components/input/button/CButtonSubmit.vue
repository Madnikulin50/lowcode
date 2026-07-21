<template>
  <button
    data-test-id="button-submit"
    type="submit"
    :class="[`btn btn-${variant}`, btnSize, block ? 'w-100' : '', buttonClass]"
    :disabled="disabled || processing || success"
    :title="title"
    @click.prevent="$emit('submit')"
  >
    <span
      v-if="processing"
      data-test-id="spinner"
      class="spinner-border spinner-border-sm align-middle"
    />

    <font-awesome-icon
      v-else-if="success"
      data-test-id="icon-success"
      :icon="['fas', 'check']"
    />

    <span
      v-else-if="text"
      data-test-id="button-text"
    >
      {{ text }}
    </span>
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  processing?: boolean
  success?: boolean
  disabled?: boolean
  title?: string
  buttonClass?: string
  text?: string
  loadingText?: string
  size?: string
  block?: boolean
  variant?: string
}

const props = withDefaults(defineProps<Props>(), {
  processing: false,
  success: false,
  disabled: false,
  title: '',
  buttonClass: '',
  text: '',
  loadingText: '',
  size: 'md',
  block: false,
  variant: 'primary',
})

defineEmits<{
  submit: []
}>()

const btnSize = computed(() => props.size === 'sm' ? 'btn-sm' : props.size === 'lg' ? 'btn-lg' : '')
</script>
