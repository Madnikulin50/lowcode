<template>
  <div class="modal fade" id="workflow-filter" tabindex="-1" role="dialog">
    <div class="modal-dialog modal-lg" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">{{ $t('filter.title') }}</h5>
          <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
        </div>
        <div class="modal-body p-3">
          <namespace-module-selector
            ref="selector"
            :namespace-labels="localNamespaceLabels"
            :module-labels="localModuleLabels"
            @change="handleChange"
          />
        </div>
        <div class="modal-footer">
          <div class="d-flex gap-1 w-100">
            <button class="btn btn-outline-secondary" @click="handleReset">
              {{ $t('reset') }}
            </button>
            <div class="d-flex ms-auto gap-1">
              <button class="btn btn-outline-secondary" data-bs-dismiss="modal">
                {{ $t('cancel') }}
              </button>
              <button class="btn btn-primary" @click="handleApply">
                {{ $t('filter.apply') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import NamespaceModuleSelector from './NamespaceModuleSelector.vue'

const { t } = useI18n()

const props = defineProps({
  namespaceLabels: { type: Array, default: () => [] },
  moduleLabels: { type: Array, default: () => [] },
})

const emit = defineEmits(['apply'])

const localNamespaceLabels = ref([])
const localModuleLabels = ref([])
const selector = ref(null)

function handleChange({ namespaceLabels, moduleLabels }) {
  localNamespaceLabels.value = namespaceLabels
  localModuleLabels.value = moduleLabels
}

function handleApply() {
  emit('apply', {
    namespaceLabels: localNamespaceLabels.value,
    moduleLabels: localModuleLabels.value,
  })
  const modal = document.getElementById('workflow-filter')
  if (modal) {
    const bsModal = bootstrap.Modal.getInstance(modal)
    if (bsModal) bsModal.hide()
  }
}

function handleReset() {
  localNamespaceLabels.value = []
  localModuleLabels.value = []
  if (selector.value) {
    selector.value.reset()
  }
}

// Initialize local state from props when modal shows
watch(() => props.namespaceLabels, (val) => {
  localNamespaceLabels.value = [...val]
}, { immediate: true })

watch(() => props.moduleLabels, (val) => {
  localModuleLabels.value = [...val]
}, { immediate: true })
</script>
