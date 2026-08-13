<template>
  <div class="d-flex">
    <button
      data-test-id="button-import"
      class="btn btn-outline-secondary btn-lg flex-fill"
      @click="showModal=true"
    >
      {{ $t('import.buttonLabel') }}
    </button>

    <div
      v-if="showModal"
      class="modal fade show d-block"
      tabindex="-1"
      @click.self="onModalHide"
    >
      <div class="modal-dialog modal-lg modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ $t('import.title') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" @click="onModalHide"></button>
          </div>
          <div class="modal-body p-0">
            <div
              v-if="importing"
              class="p-5 h-100 d-flex align-items-center justify-content-center"
            >
              <span class="spinner-border spinner-border-sm" />
            </div>

            <component
              :is="stepComponent"
              v-else
              v-bind="$props"
              :session="session"
              @fileUploaded="onFileUploaded"
              @configured="onConfigured"
              @back="onBack"
              @reset="onReset"
              @close="onClose"
            />
          </div>
        </div>
      </div>
    </div>
    <div
      v-if="showModal"
      class="modal-backdrop fade show"
      @click="onModalHide"
    />
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'namespace' } })
import { ref, computed, inject } from 'vue'
import FileUpload from './FileUpload'
import ImportConfiguration from './ImportConfiguration.vue'

const emit = defineEmits(['imported', 'failed', 'reset'])

const $ComposeAPI = inject('$ComposeAPI')

const step = ref(0)
const showModal = ref(false)
const session = ref({})
const components = [FileUpload, ImportConfiguration]
const importing = ref(false)

const stepComponent = computed(() => components[step.value])

function onModalHide () {
  step.value = 0
  session.value = {}
  showModal.value = false
}

function onBack () {
  step.value = Math.max(0, step.value - 1)
}

function onFileUploaded (e) {
  session.value = e
  step.value = 1
}

async function onConfigured (e) {
  importing.value = true
  try {
    const out = await $ComposeAPI.namespaceImportRun({
      sessionID: session.value.sessionID,
      connectionID: e.connectionID,
      importData: e.importData,
      name: e.name,
      slug: e.slug,
    })
    emit('imported', out)
  } catch (err) {
    emit('failed', err)
  }
  onReset()
  onClose()
  importing.value = false
}

function onReset () {
  step.value = 0
  session.value = {}
  emit('reset')
}

function onClose () {
  showModal.value = false
}
</script>
