<template>
  <div class="d-flex">
    <button
      class="btn btn-lg btn-outline-secondary flex-fill"
      @click="showModal=true"
    >
      {{ $t('label.import') }}
    </button>

    <div
      v-if="showModal"
      class="modal d-block"
      tabindex="-1"
      role="dialog"
      @click.self="onModalHide"
    >
      <div class="modal-dialog modal-lg modal-dialog-scrollable" role="document">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ $t('recordList.import.to', { modulename: module.name }) }}</h5>
            <button type="button" class="btn-close" @click="showModal=false"></button>
          </div>
          <div class="modal-body p-0">
            <component
              :is="stepComponent"
              :session="session"
              :namespace="namespace"
              :module="module"
              @fileUploaded="onFileUploaded"
              @fieldsMatched="onFieldsMatched"
              @importSuccessful="onImportSuccessful"
              @importFailed="onImportFailed"
              @back="onBack"
              @reset="onReset"
              @close="onClose"
            >
              <label
                v-if="progress.failed"
                slot="uploadLabel"
                class="text-danger"
              >
                {{ $t('recordList.import.failed', progress) }}
              </label>
            </component>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import FileUpload from './FileUpload.vue'
import FieldMatch from './FieldMatch.vue'
import Progress from './Progress.vue'
import ErrorReport from './ErrorReport.vue'

const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  namespace: { type: Object, required: true, default: () => ({}) },
  module: { type: Object, required: true, default: () => ({}) },
})

const emit = defineEmits(['importSuccessful', 'reset'])

const step = ref(0)
const showModal = ref(false)
const session = ref({})
const components = [FileUpload, FieldMatch, Progress, ErrorReport]

const stepComponent = computed(() => components[step.value])
const progress = computed(() => session.value.progress || {})

function onModalHide() {
  step.value = 0
  session.value = {}
  showModal.value = false
}

function onBack() {
  step.value = Math.max(0, step.value - 1)
}

function onFileUploaded(e) {
  session.value = e
  step.value = 1
}

function onFieldsMatched(e) {
  session.value.fields = e
  step.value = 2
  window.__composeAPI.recordImportRun(session.value)
}

function onImportSuccessful() {
  emit('importSuccessful')
}

function onImportFailed(e) {
  session.value.progress = e
  step.value = 3
}

function onReset() {
  step.value = 0
  session.value = {}
  emit('reset')
}

function onClose() {
  showModal.value = false
}
</script>
