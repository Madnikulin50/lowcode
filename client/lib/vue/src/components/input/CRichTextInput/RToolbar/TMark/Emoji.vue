<template>
  <span class="position-relative">
    <button
      :id="popoverId"
      class="btn btn-link text-dark fw-bold text-decoration-none"
      @click="togglePopover"
    >
      <font-awesome-icon :icon="['far', 'face-smile']" />
    </button>

    <div
      v-if="visible"
      class="emoji-picker-dropdown"
    >
      <c-emoji-picker
        ref="picker"
        :emojis="allEmojis"
        :labels="labels.emojiPicker || {}"
        :show-quick-reactions="false"
        @select="onSelect"
      />
    </div>
  </span>
</template>

<script setup lang="ts">
import { ref, computed, nextTick } from 'vue'
import CEmojiPicker from '~corteza-vue/components/CEmojiPicker.vue'

let popoverCounter = 0

const props = defineProps<{
  editor: any
  format: any
  labels: Record<string, any>
}>()

const popoverId = `emoji-picker-${++popoverCounter}`
const visible = ref(false)
const picker = ref<any>(null)

const allEmojis = computed(() => {
  return props.editor.storage?.emoji?.emojis || []
})

function togglePopover() {
  visible.value = !visible.value
}

function onShown() {
  nextTick(() => {
    if (picker.value) {
      picker.value.reset()
    }
  })
}

function onSelect(emoji: any) {
  if (emoji && emoji.name) {
    props.editor.chain().focus().insertContent({
      type: 'emoji',
      attrs: { name: emoji.name },
    }).run()
  }
  visible.value = false
  window.dispatchEvent(new CustomEvent('bv::hide::popover', { detail: { id: popoverId } }))
}
</script>

<style lang="scss">
.emoji-picker-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  z-index: 1060;
  background: var(--white, #fff);
  max-width: 300px;
  border: 1px solid var(--light, #f8f9fa);
  border-radius: 0.25rem;
  box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15);
}
</style>
