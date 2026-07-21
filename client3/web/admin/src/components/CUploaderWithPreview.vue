<template>
  <form
    @submit.prevent="$emit('upload')"
  >
    <c-uploader
      v-if="!props.disabled"
      :endpoint="uploadEndpoint"
      :labels="{
        placeholder: $t('label.dropFiles'),
        uploading: $t('label.uploading'),
        fileTypeNotAllowed: $t('label.fileTypeNotAllowed'),
      }"
      :accepted-files="['image/*']"
      @upload="$emit('upload', $event)"
    />

    <div
      v-if="props.value"
      class="d-flex justify-content-center w-100 mt-2"
    >
      <img
        :src="props.value"
        class="mw-100 h-auto"
      >
    </div>
  </form>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'
const { CUploader } = components

const { t } = useI18n()

const props = defineProps({
  value: {
    type: String,
    default: () => undefined,
  },
  disabled: {
    type: Boolean,
    default: () => false,
  },
  labels: {
    type: Object,
    default: () => ({}),
  },
  endpoint: {
    type: String,
    required: true,
  },
  acceptedFiles: {
    type: Array,
    default: () => [],
  },
  maxFilesize: {
    type: Number,
    default: 100,
  },
})

defineEmits(['upload'])

const uploadEndpoint = computed(() => {
  return window.__systemAPI.baseURL + props.endpoint
})
</script>
