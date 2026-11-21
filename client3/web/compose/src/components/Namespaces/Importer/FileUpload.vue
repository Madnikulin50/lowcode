<template>
  <div class="card">
    <div class="card-body">
      <c-uploader
        :endpoint="fileUploadEndpoint"
        :accepted-files="['application/zip']"
        :labels="{
          uploading: $t('label.uploading'),
          placeholder: $t('import.uploadFilePlaceholder'),
          fileTypeNotAllowed: $t('label.fileTypeNotAllowed'),
        }"
        @upload="onUploaded"
      />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'namespace' } })
import { ref, computed, inject } from 'vue'
import { components } from 'corteza-lib/vue/dist'
const { CUploader } = components

const emit = defineEmits(['fileUploaded'])

const $ComposeAPI = inject('$ComposeAPI')

const session = ref(null)
const sessionFile = ref(null)

const fileUploadEndpoint = computed(() => $ComposeAPI.baseURL + $ComposeAPI.namespaceImportInitEndpoint({}))

const canContinue = computed(() => !!session.value)

function onUploaded (e, f) {
  session.value = e
  sessionFile.value = f
  fileUploaded()
}

function fileUploaded () {
  emit('fileUploaded', { ...(session.value || {}) })
}
</script>
