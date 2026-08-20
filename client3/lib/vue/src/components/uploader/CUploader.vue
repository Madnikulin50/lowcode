<template>
  <div>
    <div
      class="uploader"
      @dragenter.prevent="onDragEnter"
      @dragover.prevent="onDragOver"
      @dragleave.prevent="onDragLeave"
      @drop.prevent="onDrop"
    >
      <div
        class="drop-container w-100 h-100 position-relative bg-light rounded"
        :class="{ 'bg-extra-light': isDragOver }"
        @click="openFileDialog"
      >
        <template v-if="processing">
          <div
            class="bg-primary h-100 progress-bar position-absolute"
            :style="progressBarStyle"
          />
          <span class="d-flex align-items-center h-100 w-100 uploading justify-content-center position-relative py-2">
            {{ uploadingLabel }}
          </span>
        </template>

        <div
          v-else
          data-test-id="drop-area"
          class="d-flex align-items-center h-100 w-100 p-2 droparea justify-content-center"
        >
          <span
            v-if="error"
            class="text-danger"
          >
            {{ error }}
          </span>
          <span
            v-else-if="activeLabel"
          >
            {{ activeLabel }}
          </span>
          <span
            v-else
            class="text-muted"
          >
            {{ placeholderLabel }}
          </span>
        </div>
      </div>
      <input
        ref="fileInput"
        type="file"
        :accept="acceptedFilesString"
        :disabled="disabled"
        class="d-none"
        @change="onFileSelected"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, inject } from 'vue'
import numeral from 'numeral'

const props = withDefaults(defineProps<{
  endpoint: string
  disabled?: boolean
  acceptedFiles?: string[]
  maxFilesize?: number
  labels?: Record<string, string>
  formData?: Record<string, any>
  paramName?: string
  maxFiles?: number
  showUploadedFileName?: boolean
  authToken?: string
  onUploaded?: (response: any, file: File) => void
}>(), {
  disabled: false,
  acceptedFiles: () => [],
  maxFilesize: 100,
  labels: () => ({}),
  formData: () => ({}),
  paramName: 'upload',
  maxFiles: 1000,
  showUploadedFileName: false,
  authToken: '',
  onUploaded: undefined,
})

const emit = defineEmits<{
  upload: [response: any, file: File]
}>()

const $auth = inject<any>('$auth', typeof window !== 'undefined' ? (window as any).__auth : undefined)

const fileInput = ref<HTMLInputElement>()
const isDragOver = ref(false)
const active = ref<File | null>(null)
const processing = ref<{ file: File; progress: number; bytesSent: number } | null>(null)
const error = ref<string | null>(null)

const resolvedToken = computed(() => {
  return props.authToken
    || $auth?.accessToken
    || (typeof $auth?.accessTokenFn === 'function' ? $auth.accessTokenFn() : '')
    || (typeof window !== 'undefined' ? (window as any).__auth?.accessToken : '')
    || ''
})

function sameOriginEndpoint (endpoint: string): string {
  if (!endpoint || endpoint.startsWith('/')) return endpoint
  if (typeof window === 'undefined') return endpoint
  try {
    const abs = new URL(endpoint, window.location.href)
    if (abs.origin === window.location.origin) return endpoint
    const api = (window as any).CortezaAPI
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

function asAcceptList (types: unknown): string[] {
  if (!types) return []
  if (Array.isArray(types)) return types.map(String).map(s => s.trim()).filter(Boolean)
  if (typeof types === 'string') return types.split(',').map(s => s.trim()).filter(Boolean)
  return []
}

const acceptList = computed(() => asAcceptList(props.acceptedFiles))
const acceptedFilesString = computed(() => acceptList.value.join(','))

const progressBarStyle = computed(() => ({
  width: (processing.value?.progress || 0) + '%',
}))

const uploadingLabel = computed(() => {
  const base = props.labels.uploading || 'Uploading files'
  const file = processing.value?.file
  return file ? `${base} ${file.name} (${size(file)})` : base
})

const activeLabel = computed(() => {
  if (!props.showUploadedFileName || !active.value) return null
  return `${active.value.name} (${size(active.value)})`
})

const placeholderLabel = computed(() => {
  return props.labels.placeholder || 'Click or drop files here to upload'
})

function size(a: File) {
  return numeral(a.size).format('0b')
}

function openFileDialog() {
  if (props.disabled) return
  fileInput.value?.click()
}

function onDragEnter() {
  if (props.disabled) return
  isDragOver.value = true
}

function onDragOver() {
  if (props.disabled) return
  isDragOver.value = true
}

function onDragLeave() {
  isDragOver.value = false
}

function onDrop(e: DragEvent) {
  isDragOver.value = false
  if (props.disabled) return
  const files = e.dataTransfer?.files
  if (files?.length) {
    handleFile(files[0])
  }
}

function onFileSelected(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files?.length) {
    handleFile(target.files[0])
  }
  target.value = ''
}

function handleFile(file: File) {
  error.value = null

  if (!validateFileType(file.name, acceptList.value, file.type)) {
    const errorMsg = props.labels.fileTypeNotAllowed || 'File type not allowed'
    onError(null, errorMsg)
    return
  }

  if (props.maxFilesize > 0 && file.size > props.maxFilesize * 1024 * 1024) {
    const errorMsg = props.labels.fileTooLarge || `File exceeds ${props.maxFilesize}MB limit`
    onError(null, errorMsg)
    return
  }

  uploadFile(file)
}

const EXT_MIME: Record<string, string> = {
  jpg: 'image/jpeg', jpeg: 'image/jpeg', png: 'image/png', gif: 'image/gif',
  webp: 'image/webp', svg: 'image/svg+xml', bmp: 'image/bmp', ico: 'image/x-icon',
  pdf: 'application/pdf',
  doc: 'application/msword',
  docx: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  xls: 'application/vnd.ms-excel',
  xlsx: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  ppt: 'application/vnd.ms-powerpoint',
  pptx: 'application/vnd.openxmlformats-officedocument.presentationml.presentation',
  odt: 'application/vnd.oasis.opendocument.text',
  ods: 'application/vnd.oasis.opendocument.spreadsheet',
  rtf: 'application/rtf',
  txt: 'text/plain',
  csv: 'text/csv',
  json: 'application/json',
  zip: 'application/zip',
  xml: 'application/xml',
}

function validateFileType(name: string, types: string[], mime = '') {
  if (!types?.length) return true
  if (types.some(t => t === '*/*' || t === '*')) return true
  const ext = name.includes('.') ? name.split('.').pop()?.toLowerCase() || '' : ''
  const fileMime = (mime || EXT_MIME[ext] || '').toLowerCase()
  return types.some((raw: string) => {
    const t = String(raw || '').trim().toLowerCase()
    if (!t) return false
    if (t.startsWith('.')) return ext === t.slice(1)
    if (t.startsWith('*.')) return ext === t.slice(2)
    if (t.includes('/')) {
      const [cat, sub] = t.split('/')
      if (cat === '*' || sub === '*') {
        return cat === '*' || fileMime.startsWith(cat + '/')
      }
      return fileMime === t || EXT_MIME[ext] === t
    }
    return ext === t.replace(/^\./, '')
  })
}

function uploadFile(file: File) {
  processing.value = { file, progress: 0, bytesSent: 0 }

  const xhr = new XMLHttpRequest()

  xhr.upload.addEventListener('progress', (e) => {
    if (e.lengthComputable) {
      processing.value = { file, progress: Math.round((e.loaded / e.total) * 100), bytesSent: e.loaded }
    }
  })

  xhr.addEventListener('load', () => {
    if (xhr.status >= 200 && xhr.status < 300) {
      let response
      try {
        response = JSON.parse(xhr.responseText)
      } catch {
        onError(null, 'Upload failed: invalid server response')
        return
      }
      if (response?.error) {
        onError(null, response.error?.message || response.error || 'Upload failed')
        return
      }
      // Corteza wraps payloads as { response: { attachmentID, ... } }
      if (response && typeof response === 'object' && response.response) {
        response = response.response
      }
      if (!response || typeof response !== 'object' || !(response as any).attachmentID) {
        onError(null, 'Upload failed: missing attachment id')
        return
      }
      active.value = file
      processing.value = null
      error.value = null
      emit('upload', response, file)
      props.onUploaded?.(response, file)
    } else {
      let message = 'Upload failed'
      try {
        const err = JSON.parse(xhr.responseText)
        message = err.error?.message || err.message || message
      } catch {
        message = xhr.statusText || message
      }
      onError(null, message)
    }
  })

  xhr.addEventListener('error', () => {
    onError(null, 'Network error')
  })

  xhr.open('POST', sameOriginEndpoint(props.endpoint))
  if (resolvedToken.value) {
    xhr.setRequestHeader('Authorization', 'Bearer ' + resolvedToken.value)
  }

  const formData = new FormData()
  formData.append(props.paramName, file)
  for (const [k, v] of Object.entries(props.formData || {})) {
    if (v == null || v === '') continue
    formData.append(k, String(v))
  }

  xhr.withCredentials = true
  xhr.send(formData)
}

function onError(_e: any, message: string) {
  active.value = null
  error.value = message
  processing.value = null
  console.error('[CUploader]', message)
}

defineExpose({ handleFile, openFileDialog })
</script>

<style lang="scss" scoped>
.drop-container {
  &:hover {
    background-color: var(--extra-light) !important;
  }
}

.droparea {
  cursor: pointer;
}

.progress-bar {
  width: 0;
  opacity: 0.3;
}

.uploading {
  background-size: 100% 100%;
  background-position: right bottom;
  cursor: wait;
}
</style>

<style lang="scss">
.uploader {
  .dz-preview {
    display: none !important;
  }
}
</style>
