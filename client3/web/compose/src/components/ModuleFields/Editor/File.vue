<template>
  <div class="mb-3" :class="formGroupStyleClasses">
    <div v-if="!valueOnly" :class="labelColClass">
      <div class="d-flex align-items-center text-primary p-0">
        <span :title="label" class="d-inline-block mw-100">{{ label }}</span>
        <c-hint :tooltip="hint" />
        <slot name="tools" />
      </div>
      <div class="small text-muted" :class="{ 'mb-1': description }">{{ description }}</div>
    </div>
    <div :class="contentColClass">
      <div class="d-flex gap-1">
        <div
          class="uploader flex-grow-1"
          @dragenter.prevent="isDragOver = true"
          @dragover.prevent="isDragOver = true"
          @dragleave.prevent="isDragOver = false"
          @drop.prevent="onDrop"
        >
          <div
            class="drop-container w-100 h-100 position-relative bg-light rounded"
            :class="{ 'bg-extra-light': isDragOver }"
            @click="openFileDialog"
          >
            <span
              v-if="uploading"
              class="d-flex align-items-center h-100 w-100 justify-content-center position-relative py-2"
            >
              {{ t('label.uploading') }}
            </span>
            <span
              v-else-if="uploadError"
              class="d-flex align-items-center h-100 w-100 justify-content-center text-danger p-2"
            >
              {{ uploadError }}
            </span>
            <span
              v-else
              class="d-flex align-items-center h-100 w-100 p-2 droparea justify-content-center text-muted"
            >
              {{ t('label.dropFiles') }}
            </span>
          </div>
          <input
            ref="fileInput"
            type="file"
            :accept="acceptAttr"
            class="d-none"
            @change="onFileSelected"
          >
        </div>
        <c-webcam
          v-if="field.options?.enableWebcam"
          :labels="webcamLabels"
          @upload="uploadFile"
        >
          <font-awesome-icon class="text-primary" :icon="['fas', 'camera']" />
        </c-webcam>
      </div>

      <div v-if="files.length" class="mt-3">
        <div
          v-for="item in files"
          :key="item.attachmentID"
          class="d-flex align-items-center justify-content-between gap-2 mb-1 rounded px-2 py-1"
          style="background: var(--extra-light);"
        >
          <span class="text-break">{{ item.name || item.attachmentID }}</span>
          <button
            type="button"
            class="btn btn-sm btn-link text-danger p-0"
            :title="t('label.delete')"
            @click="removeFile(item.attachmentID)"
          >
            <font-awesome-icon :icon="['fas', 'times']" />
          </button>
        </div>
      </div>
      <FieldErrors :errors="errors" />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'general' } })
import { computed, ref, inject, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useEditorBase } from './base'
import { NoID } from 'corteza-lib/js/dist'
import FieldErrors from '../errors'

const props = defineProps({
  namespace: { type: Object, required: true },
  field: { type: Object, required: true },
  record: { type: Object, required: true },
  errors: { type: Object, required: true },
  valueOnly: { type: Boolean, default: false },
  horizontal: { type: Boolean, default: false },
  extraOptions: { type: Object, default: () => ({}) },
})

const emit = defineEmits(['change', 'update:preventPopoverClose'])

const { t } = useI18n({ useScope: 'global', messages: {} })
const { value, formGroupStyleClasses, labelColClass, contentColClass, label, hint, description } = useEditorBase(props, emit)

const $ComposeAPI = inject('$ComposeAPI', typeof window !== 'undefined' ? window.__composeAPI : undefined)
const $settings = inject('$Settings')
const $auth = inject('$auth', typeof window !== 'undefined' ? window.__auth : undefined)

const fileInput = ref(null)
const isDragOver = ref(false)
const uploading = ref(false)
const uploadError = ref('')
const files = ref([])

const IMAGE_MIMES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/svg+xml', 'image/bmp']
const DOC_MIMES = [
  'application/pdf',
  'application/msword',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  'application/vnd.ms-excel',
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  'application/vnd.ms-powerpoint',
  'application/vnd.openxmlformats-officedocument.presentationml.presentation',
  'application/vnd.oasis.opendocument.text',
  'application/rtf',
  'text/plain',
  '.doc', '.docx', '.xls', '.xlsx', '.ppt', '.pptx', '.odt', '.rtf', '.txt', '.zip',
]

const EXT_MIME = {
  jpg: 'image/jpeg', jpeg: 'image/jpeg', png: 'image/png', gif: 'image/gif',
  webp: 'image/webp', svg: 'image/svg+xml', bmp: 'image/bmp',
  pdf: 'application/pdf',
  doc: 'application/msword',
  docx: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  xls: 'application/vnd.ms-excel',
  xlsx: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  ppt: 'application/vnd.ms-powerpoint',
  pptx: 'application/vnd.openxmlformats-officedocument.presentationml.presentation',
  odt: 'application/vnd.oasis.opendocument.text',
  rtf: 'application/rtf',
  txt: 'text/plain',
  zip: 'application/zip',
}

function authToken () {
  return $auth?.accessToken
    || (typeof $auth?.accessTokenFn === 'function' ? $auth.accessTokenFn() : '')
    || (typeof window !== 'undefined' ? window.__auth?.accessToken : '')
    || ''
}

function sameOriginEndpoint (endpoint) {
  if (!endpoint || endpoint.startsWith('/')) return endpoint
  if (typeof window === 'undefined') return endpoint
  try {
    const abs = new URL(endpoint, window.location.href)
    if (abs.origin === window.location.origin) return endpoint
    const api = window.CortezaAPI
    if (!api) return endpoint
    const apiOrigin = new URL(String(api), window.location.href).origin
    if (abs.origin === apiOrigin) {
      return abs.pathname + abs.search
    }
  } catch {
    return endpoint
  }
  return endpoint
}

const fileUploadEndpoint = computed(() => {
  const { moduleID, recordID } = props.record
  const { namespaceID } = props.namespace
  const api = $ComposeAPI || {}
  if (!api.baseURL || !api.recordUploadEndpoint) return ''
  return api.baseURL + api.recordUploadEndpoint({
    namespaceID,
    moduleID,
    recordID,
    fieldName: props.field.name,
  })
})

const mimetypes = computed(() => {
  const a = (props.field.options?.mimetypes || '').trim()
  if (a) return a.split(',').map(p => p.trim()).filter(Boolean)
  const allowImg = props.field.options?.allowImages !== false
  const allowDoc = props.field.options?.allowDocuments !== false
  if (allowImg && allowDoc) return ['*/*']
  if (allowImg) return IMAGE_MIMES
  if (allowDoc) return [...DOC_MIMES]
  const fromSettings = $settings?.get ? $settings.get('compose.Record.Attachments.Mimetypes', ['*/*']) : ['*/*']
  return Array.isArray(fromSettings) && fromSettings.length ? fromSettings : ['*/*']
})

const acceptAttr = computed(() => {
  const list = mimetypes.value
  if (!list.length || list.some(t => t === '*/*' || t === '*')) return ''
  return list.join(',')
})

const maxSize = computed(() => {
  const fieldSize = Number(props.field.options?.maxSize || 0)
  if (fieldSize > 0) return fieldSize
  const fromSettings = $settings?.get ? $settings.get('compose.Record.Attachments.MaxSize', 100) : 100
  return Number(fromSettings) || 100
})

const webcamLabels = computed(() => ({
  tooltip: t('webcam.tooltip'),
  modalTitle: t('webcam.title'),
  cancelButtonLabel: t('webcam.buttons.cancel'),
  confirmButtonLabel: t('webcam.buttons.confirm'),
  captureButtonLabel: t('webcam.buttons.capture'),
  cameraErrorMessage: t('webcam.errors.camera'),
}))

function asAttachmentIDs (raw) {
  const list = Array.isArray(raw) ? raw : (raw != null && raw !== '' ? [raw] : [])
  return list
    .map(item => (item && typeof item === 'object' ? item.attachmentID : item))
    .filter(id => (typeof id === 'string' || typeof id === 'number') && String(id) !== '' && String(id) !== '0')
    .map(id => String(id))
}

function writeFieldValue (ids) {
  const next = props.field.isMulti ? ids : (ids[0] || undefined)
  // Direct assignment: Record.setValue is a no-op when fieldIndex missed the
  // name (Vue 3 proxy), and the class instance itself is often not reactive.
  const rec = props.record
  if (props.field.isSystem) {
    rec[props.field.name] = next
  } else if (rec.values) {
    rec.values[props.field.name] = next
  }
  value.value = next
}

function persistFiles () {
  writeFieldValue(files.value.map(f => String(f.attachmentID)))
}

function validateFileType (name, types, mime = '') {
  if (!types?.length) return true
  if (types.some(t => t === '*/*' || t === '*')) return true
  const ext = name.includes('.') ? name.split('.').pop().toLowerCase() : ''
  const fileMime = (mime || EXT_MIME[ext] || '').toLowerCase()
  return types.some(raw => {
    const tt = String(raw || '').trim().toLowerCase()
    if (!tt) return false
    if (tt.startsWith('.')) return ext === tt.slice(1)
    if (tt.startsWith('*.')) return ext === tt.slice(2)
    if (tt.includes('/')) {
      const [cat, sub] = tt.split('/')
      if (cat === '*' || sub === '*') return cat === '*' || fileMime.startsWith(cat + '/')
      return fileMime === tt || EXT_MIME[ext] === tt
    }
    return ext === tt.replace(/^\./, '')
  })
}

function openFileDialog () {
  if (uploading.value) return
  fileInput.value?.click()
}

function onDrop (e) {
  isDragOver.value = false
  const dropped = e.dataTransfer?.files
  if (dropped?.length) uploadFile(dropped[0])
}

function onFileSelected (e) {
  const picked = e.target.files
  if (picked?.length) uploadFile(picked[0])
  e.target.value = ''
}

function unwrapAttachment (payload) {
  if (payload == null || typeof payload !== 'object') return null
  if (payload.attachmentID) return payload
  if (payload.response?.attachmentID) return payload.response
  return null
}

async function uploadFile (file) {
  if (!file) return
  uploadError.value = ''

  if (!validateFileType(file.name, mimetypes.value, file.type)) {
    uploadError.value = t('label.fileTypeNotAllowed')
    return
  }
  if (maxSize.value > 0 && file.size > maxSize.value * 1024 * 1024) {
    uploadError.value = t('label.fileTooLarge') || `File exceeds ${maxSize.value}MB limit`
    return
  }

  const endpoint = sameOriginEndpoint(fileUploadEndpoint.value)
  if (!endpoint) {
    uploadError.value = 'Upload endpoint is not configured'
    return
  }

  uploading.value = true
  try {
    const fd = new FormData()
    fd.append('upload', file)
    fd.append('fieldName', props.field.name)
    if (props.record && props.record.recordID !== NoID) {
      fd.append('recordID', String(props.record.recordID))
    }

    const headers = {}
    const token = authToken()
    if (token) headers.Authorization = 'Bearer ' + token

    const res = await fetch(endpoint, {
      method: 'POST',
      headers,
      credentials: 'include',
      body: fd,
    })

    let json = null
    try {
      json = await res.json()
    } catch {
      throw new Error('Upload failed: invalid server response')
    }

    if (!res.ok || json?.error) {
      throw new Error(json?.error?.message || json?.error || json?.message || res.statusText || 'Upload failed')
    }

    const att = unwrapAttachment(json)
    if (!att?.attachmentID) {
      throw new Error('Upload failed: missing attachment id')
    }

    const id = String(att.attachmentID)
    const row = {
      attachmentID: id,
      name: att.name || file.name || id,
      url: att.url,
      previewUrl: att.previewUrl,
      meta: att.meta,
    }
    if (props.field.isMulti) {
      files.value = [row, ...files.value.filter(f => f.attachmentID !== id)]
    } else {
      files.value = [row]
    }
    persistFiles()
  } catch (err) {
    uploadError.value = err?.message || String(err)
    console.error('[File editor]', err)
  } finally {
    uploading.value = false
  }
}

function removeFile (id) {
  files.value = files.value.filter(f => f.attachmentID !== String(id))
  persistFiles()
}

async function hydrateFromRecord () {
  const ids = asAttachmentIDs(value.value)
  const extras = files.value.filter(f => !ids.includes(String(f.attachmentID)))
  const api = $ComposeAPI || (typeof window !== 'undefined' ? window.__composeAPI : undefined)
  const namespaceID = props.namespace?.namespaceID
  if (!ids.length) {
    files.value = extras
    return
  }
  if (!api?.attachmentRead || !namespaceID) {
    files.value = [...extras, ...ids.map(id => ({ attachmentID: id, name: id }))]
    return
  }
  const loaded = await Promise.all(ids.map(id =>
    api.attachmentRead({ kind: 'record', attachmentID: id, namespaceID })
      .then(a => ({
        attachmentID: String(a.attachmentID || id),
        name: a.name || id,
        url: a.url,
        previewUrl: a.previewUrl,
        meta: a.meta,
      }))
      .catch(() => ({ attachmentID: id, name: id }))
  ))
  const stillExtras = files.value.filter(f => !ids.includes(String(f.attachmentID)))
  files.value = [...stillExtras, ...loaded]
}

watch(() => String(props.record?.recordID || ''), (id, prev) => {
  if (prev && prev !== id) {
    files.value = []
  }
  hydrateFromRecord()
}, { immediate: true })
</script>

<style scoped>
.drop-container:hover {
  background-color: var(--extra-light) !important;
}
.droparea {
  cursor: pointer;
  min-height: 2.5rem;
}
</style>
