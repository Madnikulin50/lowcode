<template>
  <dwg-preview
    v-if="show && isDwg"
    :src="(attachment || {}).document || (attachment || {}).src"
    :name="(attachment || {}).name"
    :download="(attachment || {}).download"
    :meta="(attachment || {}).meta"
    @close="attachment=undefined"
  />

  <c-preview-lightbox
    v-else-if="show"
    :src="(attachment || {}).document || (attachment || {}).src"
    :name="(attachment || {}).name"
    :alt="(attachment || {}).name"
    :mime="attachmentMime"
    :labels="previewLabels"
    :meta="(attachment || {}).meta"
    @close="attachment=undefined"
  >
    <template #header.left>
      <p class="m-0">
        {{ (attachment || {}).name }}
      </p>
    </template>

    <template #header.right>
      <a
        v-if="attachment.download"
        :href="(attachment || {}).download"
      >
        {{ $t('label.download') }}
      </a>
    </template>
  </c-preview-lightbox>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'preview' } })
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { CPreviewLightbox } from 'corteza-lib/vue/dist'
import DwgPreview from './DwgPreview.vue'
import { registerAssistantDocument, parseAttachmentURL } from 'corteza-webapp-compose/src/composables/useAssistantContext'
const { t: $t } = useI18n({ useScope: 'global' })

const attachment = ref(undefined)

const show = computed({
  get: () => !!attachment.value,
  set: (val) => { if (!val) attachment.value = undefined },
})

// while a file is open in the lightbox the assistant can be asked about it
let unregister = () => {}
watch(attachment, (a) => {
  unregister()
  unregister = () => {}
  if (!a) return
  const type = ((a.meta || {}).original || (a.meta || {}).preview || {}).mimetype || ''
  const p = parseAttachmentURL(a.src) || parseAttachmentURL(a.download)
  if (p && (type === 'application/pdf' || type.startsWith('image/'))) {
    unregister = registerAssistantDocument({ source: 'attachment', id: p.id, namespaceID: p.namespaceID, name: a.name })
  }
})

const attachmentMime = computed(() => {
  const meta = (attachment.value || {}).meta || {}
  return (meta.original || meta.preview || {}).mimetype
})

const isDwg = computed(() => {
  const name = String((attachment.value || {}).name || '')
  const src = String((attachment.value || {}).src || (attachment.value || {}).document || '')
  const mime = String(attachmentMime.value || '')
  return /\.dwg($|\?)/i.test(name) || /\.dwg($|\?)/i.test(src) || /dwg|acad/i.test(mime)
})

const previewLabels = computed(() => ({
  loading: $t('pdf.loading'),
  downloadForAll: $t('pdf.downloadForAll'),
  pageLoadFailed: $t('pdf.pageLoadFailed'),
  pageLoading: $t('pdf.pageLoading'),
  noPages: $t('pdf.noPages'),
  clickToRetry: $t('pdf.clickToRetry'),
  previewUnavailable: $t('label.previewUnavailable'),
  loadFailed: $t('general.loadFailed'),
  tooLarge: $t('general.tooLarge'),
  truncated: $t('general.truncated'),
  hintDwg: $t('hint.dwg'),
  hintArchicad: $t('hint.archicad'),
  hintBimx: $t('hint.bimx'),
}))

function onKeyUp({ key }) {
  if (key === 'Escape') {
    attachment.value = undefined
  }
}

function showAttachmentModal({ detail: { url, download, name, document: doc, meta, enableDownload } }) {
  attachment.value = {
    document: doc,
    download,
    meta,
    src: url,
    name,
    caption: name,
    enableDownload,
  }
}

function setDefaultValues() {
  attachment.value = undefined
}

function destroyEvents() {
  window.removeEventListener('keyup', onKeyUp)
  window.removeEventListener('showAttachmentsModal', showAttachmentModal)
}

onMounted(() => {
  window.addEventListener('keyup', onKeyUp)
  window.addEventListener('showAttachmentsModal', showAttachmentModal)
})

onBeforeUnmount(() => {
  unregister()
  destroyEvents()
  setDefaultValues()
})
</script>
