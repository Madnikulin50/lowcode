<template>
  <button
    class="btn"
    :class="[`btn-${buttonVariant}`, buttonClass]"
    :disabled="disabled"
    :size="size"
    @click="onClick"
  >
    <slot>
      <font-awesome-icon :icon="['fas', 'language']" />
    </slot>
  </button>
</template>

<script setup>
import { library } from '@fortawesome/fontawesome-svg-core'
import { faLanguage } from '@fortawesome/free-solid-svg-icons'

library.add(faLanguage)

const props = defineProps({
  buttonVariant: { type: String, default: 'extra-light' },
  buttonClass: { type: String, default: '' },
  size: { type: String, default: 'md' },
  disabled: { type: Boolean, default: false },
  resource: { type: String, required: true },
  highlightKey: { type: String, default: '' },
  tooltip: { type: String, default: '' },
  titles: { type: Object, default: () => ({}) },
  fetcher: { type: Function, default: undefined },
  updater: { type: Function, default: undefined },
  keyPrettyfier: { type: Function, default: undefined },
})

function onClick() {
  window.dispatchEvent(new CustomEvent('c-translator', {
    detail: {
      resource: props.resource,
      titles: props.titles,
      highlightKey: props.highlightKey,
      fetcher: props.fetcher,
      updater: props.updater,
      keyPrettyfier: props.keyPrettyfier,
    },
  }))
}
</script>

<style lang="scss" scoped>
.pointer {
  cursor: pointer;
}
</style>
