<template>
  <div>
    <c-uploader
      :labels="{
        placeholder: t('label.dropFiles'),
        uploading: t('label.uploading'),
        fileTypeNotAllowed: t('label.fileTypeNotAllowed'),
      }"
      :endpoint="userImportEndpoint"
      :accepted-files="['application/zip']"
      @upload="onUploaded"
    />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'

const { CUploader } = components

const { t } = useI18n()

const emit = defineEmits(['imported'])

const userImportEndpoint = computed(() => {
  return window.__systemAPI.baseURL + window.__systemAPI.userImportEndpoint({})
})

function onUploaded () {
  emit('imported', {})
}
</script>
