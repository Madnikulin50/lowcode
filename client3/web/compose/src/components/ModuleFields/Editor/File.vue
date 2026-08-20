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
      <c-uploader
        ref="uploaderRef"
        :endpoint="fileUploadEndpoint"
        :accepted-files="mimetypes"
        :max-filesize="maxSize"
        :form-data="uploaderFormData"
        :auth-token="authToken"
        :labels="uploadLabels"
        :on-uploaded="appendAttachment"
        class="flex-grow-1"
        @upload="appendAttachment"
      />
      <c-webcam
        v-if="field.options?.enableWebcam"
        :labels="webcamLabels"
        @upload="uploadWebcamImage"
      >
        <font-awesome-icon class="text-primary" :icon="['fas', 'camera']" />
      </c-webcam>
    </div>

    <ListLoader
      v-if="attachmentSet.length > 0"
      :key="attachmentSetKey"
      kind="record"
      v-model:set="attachmentSet"
      :namespace="namespace"
      :enable-order="field.isMulti"
      enable-delete
      mode="list"
      class="mt-3"
    />
    <FieldErrors :errors="errors" />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'general' } })
import { computed, ref, inject, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useEditorBase } from './base'
import { components } from 'corteza-lib/vue/dist'
import ListLoader from 'corteza-webapp-compose/src/components/Public/Page/Attachment/ListLoader'
import { NoID } from 'corteza-lib/js/dist'
import FieldErrors from '../errors'
const { CUploader } = components

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

const uploaderRef = ref(null)
const authToken = computed(() => $auth?.accessToken || $auth?.accessTokenFn?.() || '')

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

const fileUploadEndpoint = computed(() => {
  const { moduleID, recordID } = props.record
  const { namespaceID } = props.namespace
  return ($ComposeAPI || {}).baseURL + ($ComposeAPI || {}).recordUploadEndpoint({
    namespaceID,
    moduleID,
    recordID,
    fieldName: props.field.name,
  })
})

const uploaderFormData = computed(() => {
  const fd = { fieldName: props.field.name }
  if (props.record && props.record.recordID !== NoID) {
    fd.recordID = props.record.recordID
  }
  return fd
})

function normalizeMimes (raw) {
  if (!raw) return ['*/*']
  if (Array.isArray(raw)) {
    const list = raw.map(v => String(v).trim()).filter(Boolean)
    return list.length ? list : ['*/*']
  }
  if (typeof raw === 'string') {
    const list = raw.split(',').map(p => p.trim()).filter(Boolean)
    return list.length ? list : ['*/*']
  }
  return ['*/*']
}

const mimetypes = computed(() => {
  const a = (props.field.options?.mimetypes || '').trim()
  if (a) return a.split(',').map(p => p.trim()).filter(Boolean)
  const allowImg = props.field.options?.allowImages !== false
  const allowDoc = props.field.options?.allowDocuments !== false
  if (allowImg && allowDoc) return ['*/*']
  if (allowImg) return IMAGE_MIMES
  if (allowDoc) return [...DOC_MIMES]
  return normalizeMimes($settings?.get ? $settings.get('compose.Record.Attachments.Mimetypes', ['*/*']) : ['*/*'])
})

const maxSize = computed(() => {
  const fieldSize = Number(props.field.options?.maxSize || 0)
  if (fieldSize > 0) return fieldSize
  const fromSettings = $settings?.get ? $settings.get('compose.Record.Attachments.MaxSize', 100) : 100
  return Number(fromSettings) || 100
})

function asAttachmentIDs (raw) {
  const list = Array.isArray(raw) ? raw : (raw != null && raw !== '' ? [raw] : [])
  return list
    .map(item => (item && typeof item === 'object' ? item.attachmentID : item))
    .filter(id => (typeof id === 'string' || typeof id === 'number') && String(id) !== '' && String(id) !== '0')
    .map(id => String(id))
}

function unwrapAttachment (payload) {
  if (payload == null || typeof payload !== 'object') return null
  if (payload.attachmentID) return payload
  if (payload.response?.attachmentID) return payload.response
  return null
}

function writeFieldValue (ids) {
  const rec = props.record
  const next = props.field.isMulti ? ids : (ids[0] || undefined)
  if (typeof rec.setValue === 'function') {
    rec.setValue(props.field.name, next)
  } else if (rec.values) {
    rec.values[props.field.name] = next
  }
  emit('change', next)
}

// Full objects from the upload response so the list can render without a second fetch.
const uploadedAttachments = ref([])

watch(() => props.record?.recordID, () => {
  uploadedAttachments.value = []
})

const attachmentSet = computed({
  get () {
    const ids = asAttachmentIDs(value.value)
    const extras = uploadedAttachments.value
    const extraIDs = new Set(extras.map(a => String(a.attachmentID)))
    return [...extras, ...ids.filter(id => !extraIDs.has(id))]
  },
  set (v) {
    const list = Array.isArray(v) ? v : []
    const ids = asAttachmentIDs(list)
    uploadedAttachments.value = uploadedAttachments.value.filter(a => ids.includes(String(a.attachmentID)))
    writeFieldValue(ids)
  },
})

const attachmentSetKey = computed(() => asAttachmentIDs(attachmentSet.value).join(','))

const uploadLabels = computed(() => ({
  uploading: t('label.uploading'),
  placeholder: t('label.dropFiles'),
  fileTypeNotAllowed: t('label.fileTypeNotAllowed'),
}))

const webcamLabels = computed(() => ({
  tooltip: t('webcam.tooltip'),
  modalTitle: t('webcam.title'),
  cancelButtonLabel: t('webcam.buttons.cancel'),
  confirmButtonLabel: t('webcam.buttons.confirm'),
  captureButtonLabel: t('webcam.buttons.capture'),
  cameraErrorMessage: t('webcam.errors.camera'),
}))

function appendAttachment (payload = {}) {
  const att = unwrapAttachment(payload)
  if (!att?.attachmentID) {
    console.error('Upload succeeded but response has no attachmentID', payload)
    return
  }
  const id = String(att.attachmentID)
  uploadedAttachments.value = [att, ...uploadedAttachments.value.filter(a => String(a.attachmentID) !== id)]
  const prev = asAttachmentIDs(value.value).filter(existing => existing !== id)
  writeFieldValue(props.field.isMulti ? [id, ...prev] : [id])
}

function uploadWebcamImage (file) {
  uploaderRef.value?.handleFile?.(file)
}
</script>
