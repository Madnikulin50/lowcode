<template>
  <button
    data-test-id="button-submit"
    type="submit"
    :class="[`btn btn-${variant}`, btnSize, block ? 'w-100' : '', buttonClass]"
    :disabled="disabled || processing || success"
    :title="title"
    :aria-label="icon && !text ? title : undefined"
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

    <template v-else>
      <font-awesome-icon
        v-if="icon"
        data-test-id="button-icon"
        :icon="icon"
        :class="text ? iconClass : ''"
      />

      <span
        v-if="text"
        data-test-id="button-text"
        :class="textClass"
      >
        {{ text }}
      </span>
    </template>
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
  // Optional icon (e.g. ['far', 'save']) shown next to the text — or alone
  // when `textClass` visually hides the text (e.g. `d-none d-sm-inline` for
  // an icon-only button on narrow screens). Falls back to `title` as the
  // accessible name in that case, same pattern as CInputConfirm's showIcon.
  icon?: unknown[]
  iconClass?: string
  textClass?: string
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
  icon: undefined,
  iconClass: 'me-2',
  textClass: '',
})

defineEmits<{
  submit: []
}>()

const btnSize = computed(() => props.size === 'sm' ? 'btn-sm' : props.size === 'lg' ? 'btn-lg' : '')
</script>
