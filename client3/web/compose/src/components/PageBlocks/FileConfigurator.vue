<template>
  <div class="tab-pane">
    <div class="mb-3">
      <label class="form-label text-primary">{{ $t('kind.file.view.modeLabel') }}</label>
      <small class="form-text">{{ $t('kind.file.view.modeFootnote') }}</small>
      <div class="btn-group btn-group-sm" data-bs-toggle="buttons">
        <label
          v-for="mode in modes"
          :key="mode.value"
          class="btn btn-outline-secondary"
          :class="{ active: options.mode === mode.value }"
        >
          <input
            v-model="options.mode"
            type="radio"
            :value="mode.value"
            class="btn-check"
          />
          {{ mode.text }}
        </label>
      </div>
    </div>

    <div class="mb-3">
      <div
        v-if="enablePreviewStyling"
        class="form-check"
      >
        <input
          v-model="options.hideFileName"
          type="checkbox"
          class="form-check-input"
        />
        <label class="form-check-label">{{ $t('kind.file.view.showName') }}</label>
      </div>

      <div class="form-check">
        <input
          v-model="options.clickToView"
          type="checkbox"
          class="form-check-input"
        />
        <label class="form-check-label">{{ $t('kind.file.view.clickToView') }}</label>
      </div>

      <div class="form-check">
        <input
          v-model="options.enableDownload"
          type="checkbox"
          class="form-check-input"
        />
        <label class="form-check-label">{{ $t('kind.file.view.enableDownload') }}</label>
      </div>
    </div>

    <div class="d-flex gap-1">
      <c-uploader
        ref="uploader"
        :endpoint="endpoint"
        :max-filesize="$s('compose.Page.Attachments.MaxSize', 100)"
        :accepted-files="$s('compose.Page.Attachments.Mimetypes', ['*/*'])"
        :auth-token="authToken"
        class="flex-grow-1"
        @upload="appendAttachment"
      />

      <c-webcam
        :labels="{
          tooltip: $t('webcam.tooltip'),
          modalTitle: $t('webcam.title'),
          cancelButtonLabel: $t('webcam.buttons.cancel'),
          confirmButtonLabel: $t('webcam.buttons.confirm'),
          captureButtonLabel: $t('webcam.buttons.capture'),
          cameraErrorMessage: $t('webcam.errors.camera')
        }"
        @upload="uploadWebcamImage"
      >
        <font-awesome-icon
          class="text-primary"
          :icon="['fas', 'camera']"
        />
      </c-webcam>
    </div>

    <list-loader
      kind="page"
      enable-delete
      :namespace="namespace"
      v-model:set="options.attachments"
      mode="list"
      class="mt-2"
    />

    <template v-if="enablePreviewStyling">
      <hr />

      <h5 class="mb-2">
        {{ $t('kind.file.view.previewStyle') }}
      </h5>

      <small>{{ $t('kind.file.view.description' ) }}</small>

      <div class="row mb-2 mt-2">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('kind.file.view.height') }}</label>
            <div class="input-group">
              <input v-model="options.height" class="form-control" />
            </div>
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('kind.file.view.width') }}</label>
            <div class="input-group">
              <input v-model="options.width" class="form-control" />
            </div>
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('kind.file.view.maxHeight') }}</label>
            <div class="input-group">
              <input v-model="options.maxHeight" class="form-control" />
            </div>
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('kind.file.view.maxWidth') }}</label>
            <div class="input-group">
              <input v-model="options.maxWidth" class="form-control" />
            </div>
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('kind.file.view.borderRadius') }}</label>
            <div class="input-group">
              <input v-model="options.borderRadius" class="form-control" />
            </div>
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('kind.file.view.margin') }}</label>
            <div class="input-group">
              <input v-model="options.margin" class="form-control" />
            </div>
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('kind.file.view.background') }}</label>
            <c-input-color-picker
              v-model="options.backgroundColor"
              :translations="{
                modalTitle: $t('kind.file.view.colorPicker'),
                light: $t('themes.labels.light'),
                dark: $t('themes.labels.dark'),
                cancelBtnLabel: $t('label.cancel'),
                saveBtnLabel: $t('label.saveAndClose')
              }"
              :theme-settings="themeSettings"
            />
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'field' } })
import { ref, computed, inject } from 'vue'
import { usePageBlockBase } from './usePageBlockBase'
import ListLoader from 'corteza-webapp-compose/src/components/Public/Page/Attachment/ListLoader'
import { components } from 'corteza-lib/vue/dist'
import CUploader from 'corteza-lib/vue/src/components/uploader/CUploader.vue'
const { CInputColorPicker } = components

const props = defineProps({
  blockIndex: { type: Number, default: -1 },
  namespace: { type: Object, required: true },
  page: { type: Object, required: true },
  blocks: { type: Array, default: () => [] },
  block: { type: Object, required: true },
  module: { type: Object, required: false, default: undefined },
  record: { type: Object, required: false, default: undefined },
  mode: { type: String, required: false, default: '' },
  editable: { type: Boolean, required: false, default: false },
  resizing: { type: Boolean, required: false, default: false },
  magnified: { type: Boolean, required: false, default: false },
  unsavedBlocks: { type: Set, default: () => new Set() },
  loadingRecord: { type: Boolean, required: false, default: false },
  errors: { type: Object, required: false, default: () => ({}) },
})

const emit = defineEmits(['errors'])
const $ComposeAPI = inject('$ComposeAPI')
const $Settings = inject('$Settings')
const $auth = inject('$auth')
const authToken = computed(() => $auth?.accessToken || '')

const { options, themeSettings } = usePageBlockBase(props, emit)

const uploader = ref(null)

const endpoint = computed(() => {
  const { pageID } = props.page
  return $ComposeAPI.baseURL + $ComposeAPI.pageUploadEndpoint({
    namespaceID: props.namespace.namespaceID,
    pageID,
  })
})

const modes = computed(() => [
  { value: 'list', text: 'List' },
  { value: 'gallery', text: 'Gallery' },
])

const enablePreviewStyling = computed(() => {
  const { mode } = options.value
  return mode === 'gallery'
})

function unwrapUpload (payload) {
  if (!payload || typeof payload !== 'object') return {}
  if (payload.attachmentID) return payload
  if (payload.response && typeof payload.response === 'object') return payload.response
  return payload
}

function appendAttachment (payload = {}) {
  const { attachmentID } = unwrapUpload(payload)
  if (!attachmentID) return
  options.value.attachments.push(attachmentID)
}

function uploadWebcamImage (file) {
  uploader.value?.addFile?.(file)
}
</script>
