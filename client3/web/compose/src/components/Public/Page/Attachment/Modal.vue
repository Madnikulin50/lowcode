<template>
  <c-preview-lightbox
    v-if="show"
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
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'

const { CPreviewLightbox } = components
const { t: $t } = useI18n({ useScope: 'global' })

const attachment = ref(undefined)

const show = computed({
  get: () => !!attachment.value,
  set: (val) => { if (!val) attachment.value = undefined },
})

const attachmentMime = computed(() => {
  const meta = (attachment.value || {}).meta || {}
  return (meta.original || meta.preview || {}).mimetype
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
  destroyEvents()
  setDefaultValues()
})
</script>
