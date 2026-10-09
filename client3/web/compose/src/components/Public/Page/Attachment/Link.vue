<template>
  <a
    v-if="canPreview && attachment.clickToView"
    :href="attachment.url"
    @click.prevent.stop="openLightbox"
  >
    <slot>
      {{ attachment.name }}
    </slot>
  </a>

  <a v-else>
    <slot>
      {{ attachment.name }}
    </slot>
  </a>
</template>

<script setup>
import { computed } from 'vue'
import { components } from 'corteza-lib/vue/dist'

const { canPreview: canPreviewCheck } = components

const props = defineProps({
  attachment: { type: Object, required: true },
})

const canPreview = computed(() => {
  const meta = props.attachment.meta || {}
  const type = (meta.original || meta.preview || {}).mimetype
  const src = props.attachment.url
  return canPreviewCheck({ type, src, name: props.attachment.name, meta })
})

function openLightbox() {
  window.dispatchEvent(new CustomEvent('showAttachmentsModal', {
    detail: { ...props.attachment },
  }))
}
</script>
