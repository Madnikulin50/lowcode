<template>
  <div class="card shadow-sm" data-test-id="card-user-profile-avatar">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('system.users.editor.avatar.title') }}</h4>
    </div>

    <div class="card-body">
  <form
        enctype="multipart/form-data"
        @submit.prevent="$emit('submit', user)"
      >
        <div class="p-3">
          <img
            :src="uploadedAvatar('avatar')"
            style="height: 4rem; width: 4rem;"
            class="rounded-circle mb-4"
          >
  
          <div class="d-flex align-items-center">
            <c-uploader-with-preview
               :endpoint="__systemAPI.userProfileAvatarEndpoint({ userID: user.userID })"
              @upload="$emit('onUpload')"
              @clear="$emit('resetAttachment', 'avatar')"
            />
  
            <c-input-confirm
              v-if="uploadedAvatar('avatar') && isKindAvatar"
              :processing="processingAvatar"
              :text="$t('label.delete')"
              size="lg"
              size-confirm="lg"
              variant="danger"
              class="ms-2 h-100"
              @confirmed="$emit('resetAttachment', 'avatar')"
            />
          </div>
  
          <div class="row mt-3">
            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">{{ $t('system.users.editor.avatar.initial.color') }}</label>
                <c-input-color-picker
                  v-model="user.meta.avatarColor"
                  data-test-id="input-text-color"
                  :translations="{
                    modalTitle: $t('system.users.editor.avatar.colorPicker'),
                    light: $t('ui.settings.editor.corteza-studio.tabs.light'),
                    dark: $t('ui.settings.editor.corteza-studio.tabs.dark'),
                    cancelBtnLabel: $t('label.cancel'),
                    saveBtnLabel: $t('label.saveAndClose')
                  }"
                  :theme-settings="themeSettings"
                />
              </div>
            </div>
  
            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">{{ $t('system.users.editor.avatar.initial.backgroundColor') }}</label>
                <c-input-color-picker
                  v-model="user.meta.avatarBgColor"
                  data-test-id="input-background-color"
                  :translations="{
                    modalTitle: $t('system.users.editor.avatar.colorPicker'),
                    light: $t('ui.settings.editor.corteza-studio.tabs.light'),
                    dark: $t('ui.settings.editor.corteza-studio.tabs.dark'),
                    cancelBtnLabel: $t('label.cancel'),
                    saveBtnLabel: $t('label.saveAndClose')
                  }"
                  :theme-settings="themeSettings"
                />
              </div>
            </div>
          </div>
        </div>
      </form>
  </div>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-button-submit
        :processing="processing"
        :success="success"
        :text="$t('label.submit')"
        class="ms-auto"
        @submit="$emit('submit', user)"
      />
    </div>
  </div>
</template>

<script setup>
import { computed, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'
import CUploaderWithPreview from 'corteza-webapp-admin/src/components/CUploaderWithPreview'
const { CInputColorPicker } = components

const { t } = useI18n()
const $Settings = inject('$Settings', {})
const __systemAPI = window.__systemAPI

const props = defineProps({
  user: { type: Object, required: true },
  processing: { type: Boolean },
  processingAvatar: { type: Boolean },
  success: { type: Boolean },
})

defineEmits(['submit', 'onUpload', 'resetAttachment'])

const isKindAvatar = computed(() => props.user.meta.avatarKind === 'avatar')
const themeSettings = computed(() => $Settings.get('ui.studio.themes', []))

function uploadedAvatar(name) {
  const attachmentID = props.user.meta.avatarID

  if (attachmentID !== '0') {
    return (
      __systemAPI.baseURL +
        __systemAPI.attachmentOriginalEndpoint({
          attachmentID,
          kind: 'avatar',
          name,
        })
    )
  }
}
</script>
