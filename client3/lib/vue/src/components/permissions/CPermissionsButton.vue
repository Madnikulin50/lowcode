<template>
  <button
    data-test-id="button-permissions"
    :title="tooltip"
    type="button"
    :class="`btn btn-${buttonVariant} btn-${size}`"
    @click="onClick"
  >
    <slot>
      <font-awesome-icon
        v-if="showButtonIcon"
        :icon="['fas', 'lock']"
      />
      <span
        v-if="buttonLabel"
        class="permissions-button-label"
      >
        {{ buttonLabel }}
      </span>
    </slot>
  </button>
</template>

<script setup lang="ts">
import { modalOpenEventName } from './def.ts'
import { library } from '@fortawesome/fontawesome-svg-core'
import { faLock } from '@fortawesome/free-solid-svg-icons'

library.add(faLock)

const props = withDefaults(defineProps<{
  size?: string
  buttonVariant?: string
  resource: string
  title?: string
  buttonLabel?: string
  modalOpenEvent?: string
  target?: string
  showButtonIcon?: boolean
  allSpecific?: boolean
  tooltip?: string
}>(), {
  size: 'md',
  buttonVariant: 'light',
  modalOpenEvent: modalOpenEventName,
  showButtonIcon: true,
  allSpecific: false,
  tooltip: '',
  title: undefined,
  buttonLabel: undefined,
  target: undefined,
})

function onClick() {
  window.dispatchEvent(new CustomEvent(props.modalOpenEvent!, {
    detail: {
      target: props.target,
      resource: props.resource,
      title: props.title,
      allSpecific: props.allSpecific,
    },
  }))
}
</script>
