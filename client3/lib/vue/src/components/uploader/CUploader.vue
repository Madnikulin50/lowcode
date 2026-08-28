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
})

const emit = defineEmits<{
  upload: [response: any, file: File]
}>()

const $auth = inject<{ accessToken?: string } | null>('$auth', null)
const fileInput = ref<HTMLInputElement>()
const isDragOver = ref(false)
const active = ref<File | null>(null)
const processing = ref<{ file: File; progress: number; bytesSent: number } | null>(null)
const error = ref<string | null>(null)

const acceptedFilesString = computed(() => props.acceptedFiles.join(','))

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

  if (!validateFileType(file.name, props.acceptedFiles)) {
    const errorMsg = props.labels.fileTypeNotAllowed || 'File type not allowed'
    onError(null, errorMsg)
    return
  }

  if (file.size > props.maxFilesize * 1024 * 1024) {
    const errorMsg = props.labels.fileTooLarge || `File exceeds ${props.maxFilesize}MB limit`
    onError(null, errorMsg)
    return
  }

  uploadFile(file)
}

function addFile(file: File) {
  handleFile(file)
}

function validateFileType(_name: string, types: string[]) {
  if (!types.length || types.includes('*/*')) return true
  const ext = _name.split('.').pop()?.toLowerCase()
  return types.some((t: string) => {
    if (t.startsWith('.')) return ext === t.slice(1)
    if (t.includes('/')) {
      const [category] = t.split('/')
      if (category === '*') return true
      const mimeMap: Record<string, string> = {
        'jpg': 'image/jpeg',
        'jpeg': 'image/jpeg',
        'png': 'image/png',
        'gif': 'image/gif',
        'pdf': 'application/pdf',
        'csv': 'text/csv',
        'xls': 'application/vnd.ms-excel',
        'xlsx': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
      }
      return mimeMap[ext || '']?.startsWith(category) ?? false
    }
    return ext === t.toLowerCase().replace('.', '')
  })
}

function uploadFile(file: File) {
  const xhr = new XMLHttpRequest()

  xhr.upload.addEventListener('progress', (e) => {
    if (e.lengthComputable) {
      processing.value = { file, progress: Math.round((e.loaded / e.total) * 100), bytesSent: e.loaded }
    }
  })

  xhr.addEventListener('load', () => {
    let payload: any
    try {
      payload = JSON.parse(xhr.responseText)
    } catch {
      payload = xhr.responseText
    }

    if (xhr.status >= 200 && xhr.status < 300) {
      if (payload && typeof payload === 'object') {
        if (payload.error) {
          const err = payload.error
          onError(null, typeof err === 'string' ? err : (err.message || 'Upload failed'))
          return
        }
        if ('response' in payload) {
          payload = payload.response
        }
      }
      active.value = file
      processing.value = null
      error.value = null
      emit('upload', payload, file)
    } else {
      let message = 'Upload failed'
      if (payload && typeof payload === 'object') {
        const err = payload.error || payload
        message = typeof err === 'string' ? err : (err.message || message)
      } else if (xhr.statusText) {
        message = xhr.statusText
      }
      onError(null, message)
    }
  })

  xhr.addEventListener('error', () => {
    onError(null, 'Network error')
  })

  xhr.open('POST', props.endpoint)
  const token = props.authToken || $auth?.accessToken || ''
  if (token) {
    xhr.setRequestHeader('Authorization', 'Bearer ' + token)
  }

  const formData = new FormData()
  formData.append(props.paramName, file)
  for (const [k, v] of Object.entries(props.formData || {})) {
    formData.append(k, v as string)
  }

  xhr.withCredentials = true
  xhr.send(formData)
}

function onError(_e: any, message: string) {
  active.value = null
  error.value = message
  processing.value = null
}

defineExpose({
  addFile,
  openFileDialog,
})
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
