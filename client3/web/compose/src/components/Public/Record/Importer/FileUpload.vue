<template>
  <div class="card">
    <div class="card-body">
      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('recordList.import.uploadFile') }}</label>
        <c-uploader
          :endpoint="endpoint"
          :accepted-files="['application/json', 'text/csv']"
          :max-filesize="100"
          show-uploaded-file-name
          class="uploader"
          @upload="onUploaded"
        />
      </div>

      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('recordList.import.onError') }}</label>
        <select v-model="onError" class="form-select w-auto">
          <option value="FAIL">{{ $t('recordList.import.onErrorFail') }}</option>
          <option value="SKIP">{{ $t('recordList.import.onErrorSkip') }}</option>
        </select>
      </div>

      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('recordList.import.multiValueDelimiter.label') }}</label>
        <select v-model="multiValueDelimiter" class="form-select w-auto">
          <option
            v-for="d of multiValueDelimiterOptions"
            :key="d.value"
            :value="d.value"
          >
            {{ d.text }}
          </option>
        </select>
      </div>
    </div>
    <div class="card-footer text-end">
      <button
        class="btn btn-primary"
        :disabled="!canContinue"
        @click="fileUploaded"
      >
        {{ $t('label.next') }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'

const { CUploader } = components
const { t: $t } = useI18n({ useScope: 'global' })
const $ComposeAPI = window.__composeAPI

const props = defineProps({
  namespace: { type: Object, required: true, default: () => ({}) },
  module: { type: Object, required: true, default: () => ({}) },
})

const emit = defineEmits(['fileUploaded'])

const session = ref(null)
const onError = ref('FAIL')
const multiValueDelimiter = ref(';')
const sessionFile = ref(null)

const endpoint = computed(() => {
  return $ComposeAPI.baseURL + $ComposeAPI.recordImportInitEndpoint({
    namespaceID: props.namespace.namespaceID,
    moduleID: props.module.moduleID,
  })
})

const multiValueDelimiterOptions = computed(() => [
  { value: ';', text: $t('recordList.import.multiValueDelimiter.semicolon.label') },
  { value: ',', text: $t('recordList.import.multiValueDelimiter.comma.label') },
  { value: '|', text: $t('recordList.import.multiValueDelimiter.pipe.label') },
  { value: '[;]', text: $t('recordList.import.multiValueDelimiter.semicolonArray.label') },
  { value: '[,]', text: $t('recordList.import.multiValueDelimiter.commaArray.label') },
  { value: '[|]', text: $t('recordList.import.multiValueDelimiter.pipeArray.label') },
])

const canContinue = computed(() => !!session.value)

function onUploaded(e, f) {
  session.value = e
  sessionFile.value = f
}

function fileUploaded() {
  emit('fileUploaded', {
    ...(session.value || {}),
    onError: onError.value,
    multiValueDelimiter: multiValueDelimiter.value,
  })
}
</script>
