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
      <CUploader
        ref="uploaderRef"
        :endpoint="fileUploadEndpoint"
        :accepted-files="mimetypes"
        :max-filesize="maxSize"
        :form-data="uploaderFormData"
        :labels="uploadLabels"
        class="flex-grow-1"
        @upload="appendAttachment"
      />
      <c-webcam
        v-if="field.options.enableWebcam"
        :labels="webcamLabels"
        @upload="uploadWebcamImage"
      >
        <font-awesome-icon class="text-primary" :icon="['fas', 'camera']" />
      </c-webcam>
    </div>

    <ListLoader
      v-if="attachmentSet.length > 0"
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
import { computed, ref, inject } from 'vue'
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

const $ComposeAPI = inject('$ComposeAPI')
const $settings = inject('$Settings')

const uploaderRef = ref(null)

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

const mimetypes = computed(() => {
  const a = (props.field.options.mimetypes || '').trim()
  if (!a) {
    return $settings?.get ? $settings.get('compose.Record.Attachments.Mimetypes', ['*/*']) : ['*/*']
  }
  return a.split(',').map(p => p.trim())
})

const maxSize = computed(() => {
  return props.field.options.maxSize || ($settings?.get ? $settings.get('compose.Record.Attachments.MaxSize', 100) : 100)
})

const attachmentSet = computed({
  get () {
    return props.field.isMulti ? (value.value || []) : [(value.value || [])].filter(id => !!id)
  },
  set (v) {
    if (props.field.isMulti) {
      value.value = v
    } else {
      value.value = (Array.isArray(v) && v.length > 0) ? v[0] : undefined
    }
  },
})

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

function appendAttachment ({ attachmentID } = {}) {
  if (props.field.isMulti) {
    value.value = [attachmentID, ...(value.value || [])]
  } else {
    value.value = attachmentID
  }
}

function uploadWebcamImage (file) {
  if (uploaderRef.value) {
    uploaderRef.value.$refs.dropzone.addFile(file)
  }
}
</script>
