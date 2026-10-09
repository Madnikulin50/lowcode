<template>
  <div class="d-flex">
    <button
      class="btn btn-outline-secondary btn-lg flex-fill"
      data-bs-toggle="modal"
      data-bs-target="#import"
    >
      {{ $t('import.label') }}
    </button>

    <div class="modal fade" id="import" tabindex="-1" role="dialog">
      <div class="modal-dialog modal-lg" role="document">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ $t('import.json') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">{{ $t('import.reassign-run-as') }}</label>
              <input class="form-control" type="file" @change="fileUpload" />
            </div>
          </div>
          <div class="modal-footer">
            <button
              class="btn btn-primary btn-lg d-flex justify-content-center align-items-center"
              :disabled="!workflows.length || processing"
              @click="handleImport"
            >
              <div v-if="processing" class="spinner-border spinner-border-sm" role="status">
                <span class="visually-hidden">Loading...</span>
              </div>
              <span v-else>{{ $t('import.label') }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { useToast } from 'corteza-lib/vue/dist'

const { t } = useI18n()
const toast = useToast()

defineProps({
  disabled: { type: Boolean, default: false },
})

const emit = defineEmits(['import'])

const workflows = ref([])
const processing = ref(false)

function fileUpload(e = {}) {
  const { files = [] } = (e.type === 'drop' ? e.dataTransfer : e.target) || {}

  if (files[0]) {
    processing.value = true
    const reader = new FileReader()

    reader.readAsText(files[0])

    reader.onload = (evt) => {
      try {
        const { workflows: wf = [] } = JSON.parse(evt.target.result)
        workflows.value = wf
      } catch (err) {
        toast.error(t('notification.general.warning'), err.message || t('notification.failed-load-file'))
      } finally {
        processing.value = false
      }
    }

    reader.onerror = () => {
      toast.error(t('notification.failed-load-file'))
      processing.value = false
    }
  }
}

function handleImport() {
  emit('import', workflows.value)
}
</script>
