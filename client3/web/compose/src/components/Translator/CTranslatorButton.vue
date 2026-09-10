<template>
  <button
    class="btn d-flex align-items-center justify-content-center"
    :class="[`btn-${buttonVariant}`, sizeClass, buttonClass]"
    :disabled="disabled"
    @click="onClick"
  >
    <slot>
      <font-awesome-icon :icon="['fas', 'language']" />
    </slot>
  </button>
</template>

<script setup>
import { computed } from 'vue'
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

// `size` used to be bound as a literal (meaningless) `size="md"` HTML
// attribute on the <button>, so it never actually resized anything — inside
// a .btn-group-sm this button kept its default line-height-driven height
// instead of matching its icon-sized siblings (which are all d-flex, sizing
// tightly to their icon rather than a text line box), and `align-items:
// stretch` on the group then dragged those siblings back up to match it.
const sizeClass = computed(() => props.size === 'sm' ? 'btn-sm' : props.size === 'lg' ? 'btn-lg' : '')

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
