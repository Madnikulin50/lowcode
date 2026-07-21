<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom">
      <h4 class="m-0">
        {{ t("title") }}
      </h4>
    </div>

    <div class="card-body">
      <div class="row">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="d-flex align-items-center form-label text-primary">
              {{ t('image.uploader.label') }}

              <c-input-confirm
                v-if="uploadedFile('auth.ui.background-image-src')"
                show-icon
                class="ms-auto"
                @confirmed="emit('resetAttachment', 'auth.ui.background-image-src')"
              />
            </label>

            <c-uploader-with-preview
              :value="uploadedFile('auth.ui.background-image-src')"
              endpoint="/settings/auth.ui.background-image-src"
              :disabled="!canManage"
              @upload="emit('onUpload')"
              @clear="emit('resetAttachment', 'auth.ui.background-image-src')"
            />
          </div>
        </div>

        <div class="col-12">
          <div class="mb-3">
            <label class="d-flex align-items-center form-label text-primary">
              {{ t('image.editor.label') }}
            </label>
            <c-ace-editor
              v-model="settings['auth.ui.styles']"
              data-test-id="auth-bg-image-styling-editor"
              name="editor/css"
              lang="css"
              min-height="500px"
              show-line-numbers
              auto-complete
              :auto-complete-suggestions="customCssAutocompleteVal"
              resizable
              class="flex-fill w-100"
            />
          </div>
        </div>
      </div>
    </div>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-button-submit
        v-if="canManage"
        :processing="processing"
        :success="success"
        :text="t('admin.general.label.submit')"
        class="ms-auto"
        @submit="emit('submit', settings['auth.ui.styles'])"
      />
    </div>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'
import { CUSTOM_CSS_AUTO_COMPLETE_VALUES } from 'corteza-webapp-admin/src/lib/cssAutoComplete'
import CUploaderWithPreview from 'corteza-webapp-admin/src/components/CUploaderWithPreview'

const { CAceEditor } = components

const { t } = useI18n()

const props = defineProps({
  settings: {
    type: Object,
    required: true,
  },
  canManage: {
    type: Boolean,
    required: true,
  },
  processing: {
    type: Boolean,
    default: false,
  },
  success: {
    type: Boolean,
    default: false,
  },
})

defineEmits(['submit', 'resetAttachment', 'onUpload'])

const customCssAutocompleteVal = CUSTOM_CSS_AUTO_COMPLETE_VALUES

function uploadedFile (name) {
  const localAttachment = /^attachment:(\d+)/

  switch (true) {
    case props.settings[name] && localAttachment.test(props.settings[name]):
      const [, attachmentID] = localAttachment.exec(props.settings[name])

      return (
        window.__systemAPI.baseURL +
        window.__systemAPI.attachmentOriginalEndpoint({
          attachmentID,
          kind: 'settings',
          name,
        })
      )
  }

  return undefined
}
</script>
