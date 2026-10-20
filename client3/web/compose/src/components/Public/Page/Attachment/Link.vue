<template>
  <span class="attachment-link">
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

    <a
      v-if="assistantDoc"
      href="#"
      class="ms-1 text-decoration-none"
      title="Спросить ИИ-ассистента об этом файле"
      @click.prevent.stop="openAssistant(assistantDoc)"
    >✦</a>
  </span>
</template>

<script setup>
import { computed } from 'vue'
import { components } from 'corteza-lib/vue/dist'
import { openAssistant, parseAttachmentURL } from 'corteza-webapp-compose/src/composables/useAssistantContext'

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

// the assistant reads PDF and pictures
const assistantDoc = computed(() => {
  const meta = props.attachment.meta || {}
  const type = (meta.original || meta.preview || {}).mimetype || ''
  if (!(type === 'application/pdf' || type.startsWith('image/'))) return null
  const p = parseAttachmentURL(props.attachment.url)
  return p ? { source: 'attachment', id: props.attachment.attachmentID || p.id, namespaceID: p.namespaceID, name: props.attachment.name } : null
})

function openLightbox() {
  window.dispatchEvent(new CustomEvent('showAttachmentsModal', {
    detail: { ...props.attachment },
  }))
}
</script>
