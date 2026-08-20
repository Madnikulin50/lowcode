<template>
  <div>
    <div
      v-if="processing"
      class="d-flex align-items-center justify-content-center h-100"
    >
      <span class="spinner-border" />
    </div>

    <div
      v-else-if="mode === 'list'"
    >
      <draggable
        item-key="attachmentID"
        v-model="attachments"
        :disabled="!enableOrder"
        handle=".handle"
      >
        <template #item="{ element, index }">
          <div
            :key="element.attachmentID"
            class="row g-0 list-item flex-nowrap mb-1 rounded"
          >
            <div class="col-auto">
              <font-awesome-icon
                v-if="enableOrder"
                :icon="['fas', 'bars']"
                class="handle text-secondary my-1 me-3"
                style="padding-top: 0.05rem;"
              />
            </div>

            <div class="col">
              <div class="d-flex flex-column flex-wrap align-items-start">
                <div
                  class="d-flex align-items-start gap-1"
                  style="word-break: break-all;"
                >
                  <div style="margin-top: 0.1rem;">
                    <AttachmentLink :attachment="element">
                      {{ element.name }}
                    </AttachmentLink>
                  </div>

                  <div class="d-flex align-items-center gap">
                    <a
                      v-if="element.download"
                      :href="element.download"
                      class="btn btn-outline-extra-light btn-sm download-button border-0"
                      @click.stop
                    >
                      <font-awesome-icon
                        :icon="['fas', 'download']"
                        class="text-secondary"
                      />
                    </a>

                    <c-input-confirm
                      v-if="enableDelete"
                      show-icon
                      class="delete-button"
                      @confirmed="deleteAttachment(index)"
                    />
                  </div>
                </div>

                <i18next
                  path="general.label.attachmentFileInfo"
                  tag="small"
                  class="d-block text-muted"
                >
                  <span>{{ size(element) }}</span>

                  <span>{{ uploadedAt(element) }}</span>
                </i18next>
              </div>
            </div>
          </div>
        </template>
      </draggable>
    </div>

    <div
      v-else
      class="d-flex align-items-start justify-content-around gap-3 flex-wrap h-100"
    >
      <div
        v-for="a in attachments"
        :key="a.attachmentID"
        class="item-preview"
      >
        <c-preview-inline
          v-if="canPreviewCheck(a)"
          :src="inlineUrl(a)"
          :title="a.name"
          :meta="a.meta"
          :name="a.name"
          :alt="a.name"
          :preview-style="{ width: 'unset', ...inlineCustomStyles(a) }"
          :labels="previewLabels"
          @openPreview="openLightbox({ ...a, ...$event })"
        />

        <div
          class="d-flex align-items-start justify-content-center"
          :style="{ width: `calc(${inlineCustomStyles(a).width})` }"
        >
          <div
            v-if="!hideFileName"
            class="text-wrap filename-container text-center"
            :style="{ marginTop: '0.1rem' }"
          >
            <AttachmentLink :attachment="a" />
          </div>
        </div>
        <a
          v-if="a.download"
          :href="a.download"
          class="btn btn-extra-light btn-sm preview-download-button border-0"
          @click.stop
        >
          <font-awesome-icon
            :icon="['fas', 'download']"
            class="text-secondary"
          />
        </a>
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'preview' } })
import { ref, watch, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import numeral from 'numeral'
import moment from 'moment'
import { compose, shared, NoID } from 'corteza-lib/js/dist'
import AttachmentLink from './Link.vue'
import draggable from 'vuedraggable'
import { url, components } from 'corteza-lib/vue/dist'

const { CPreviewInline, canPreview, getExtensionIconType } = components
const { t: $t } = useI18n({ useScope: 'global' })
const $ComposeAPI = window.__composeAPI

const props = defineProps({
  enableDelete: { type: Boolean },
  enableOrder: { type: Boolean, default: false },
  namespace: { type: compose.Namespace, required: true },
  kind: { type: String, required: true },
  mode: { type: String, required: true },
  set: { type: Array, required: true },
  hideFileName: { type: Boolean, default: false },
  previewOptions: { type: Object, default: () => ({}) },
})

const emit = defineEmits(['update:set'])

const processing = ref(false)
const attachments = ref([])

const inlineUrl = (a) => a.url

const previewLabels = {
  loading: $t('pdf.loading'),
  firstPagePreview: $t('pdf.firstPagePreview'),
  pageLoadFailed: $t('pdf.pageLoadFailed'),
  pageLoading: $t('pdf.pageLoading'),
}

const canPreviewCheck = (a) => {
  const meta = a.meta || {}
  const type = (meta.preview || meta.original || {}).mimetype
  const src = inlineUrl(a)
  return canPreview({ type, src, name: a.name })
}

const baseURL = url.Make({ url: window.CortezaAPI + '/compose' })

function isAttachmentObject (a) {
  return !!a && typeof a === 'object' && !Array.isArray(a)
}

watch(() => props.set, (set) => {
  const list = Array.isArray(set) ? set : []
  const att = list.map(a => {
    if (isAttachmentObject(a)) {
      return new shared.Attachment(a, baseURL)
    }
    return null
  })

  const namespaceID = props.namespace.namespaceID
  processing.value = true

  Promise.all(list.map((attachmentID, index) => {
    if (typeof attachmentID === 'string' || typeof attachmentID === 'number') {
      return $ComposeAPI.attachmentRead({ kind: props.kind, attachmentID: String(attachmentID), namespaceID }).then(a => {
        att.splice(index, 1, new shared.Attachment(a, baseURL))
      })
    }
    return Promise.resolve()
  }))
    .then(() => {
      const { clickToView = true, enableDownload = true } = props.previewOptions
      attachments.value = att
        .filter(a => isAttachmentObject(a) && a.attachmentID && a.attachmentID !== NoID)
        .map(a => ({
          ...a,
          download: enableDownload ? a.download : undefined,
          clickToView,
        }))
    })
    .catch(err => {
      console.error('Failed to load attachments', err)
    })
    .finally(() => {
      processing.value = false
    })
}, { immediate: true, deep: true })

function size(a) {
  const bytes = a?.meta?.original?.size
  return bytes == null ? '' : numeral(bytes).format('0b')
}

function uploadedAt(a) {
  const ts = a?.updatedAt || a?.createdAt
  return ts ? moment(ts).fromNow() : ''
}

function openLightbox(e) {
  if (ext(e) === 'pdf') {
    window.open(e.url, '_blank')
  } else {
    window.dispatchEvent(new CustomEvent('showAttachmentsModal', {
      detail: e,
    }))
  }
}

function deleteAttachment(index) {
  attachments.value.splice(index, 1)
  emit('update:set', attachments.value.map(a => a.attachmentID))
}

function ext(a) {
  const { meta } = a
  const { original = {} } = meta || {}
  const { ext: fileExt } = original || {}
  return getExtensionIconType(fileExt)
}

function inlineCustomStyles(a) {
  const {
    borderRadius,
    backgroundColor,
  } = props.previewOptions
  let { width, height, maxWidth, maxHeight, margin } = props.previewOptions

  maxWidth = maxWidth || '100%'
  maxHeight = maxHeight || '100%'
  margin = margin || 'auto'

  if (ext(a) !== 'image') {
    width = width || '200px'
    height = height || 'auto'
  }

  return {
    width,
    height,
    maxWidth,
    maxHeight,
    borderRadius,
    backgroundColor,
    margin,
  }
}

function setDefaultValues() {
  processing.value = false
  attachments.value = []
}

onBeforeUnmount(() => {
  setDefaultValues()
})
</script>

<style lang="scss" scoped>
.handle {
  cursor: grab;
}

.list-item {
  .download-button {
    opacity: 0;
    transition: opacity 0.2s;
  }

  &:hover .download-button {
    opacity: 1;
  }

  .delete-button {
    opacity: 0;
    transition: opacity 0.2s;
  }

  &:hover .delete-button {
    opacity: 1;
  }

  &:hover {
    background-color: var(--light);
  }
}

.item-preview {
  position: relative;
  .preview-download-button {
    position: absolute;
    top: 0;
    right: 0;
    opacity: 0;
    transition: opacity 0.2s;
  }

  &:hover .preview-download-button {
    opacity: 1;
  }

  .filename-container {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    text-overflow: ellipsis;
    word-break: break-word;
    max-width: 100%;

    &:hover {
      -webkit-line-clamp: unset;
      line-clamp: unset;
      overflow: visible;
    }
  }
}
</style>
