<template>
  <div class="d-flex">
    <button
      data-test-id="button-export"
      class="btn btn-outline-secondary btn-lg flex-fill"
      @click="showModal = true"
    >
      {{ t('export.buttonLabel') }}
    </button>

    <Teleport to="body">
      <div
        v-if="showModal"
        class="modal fade show d-block"
        tabindex="-1"
        style="background: rgba(0,0,0,0.5);"
      >
        <div class="modal-dialog modal-lg modal-dialog-centered">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">{{ t('export.title') }}</h5>
              <button type="button" class="btn-close" @click="showModal = false"></button>
            </div>
            <div class="modal-body p-0">
              <keep-alive>
                <component
                  :is="stepComponent"
                  v-if="!processing"
                  v-bind="$attrs"
                  :session="session"
                  @configured="onConfigured"
                  @close="onClose"
                />
                <div
                  v-else
                  class="p-5 h-100 d-flex align-items-center justify-content-center"
                >
                  <div class="spinner-border" />
                </div>
              </keep-alive>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'system.users' } })
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import ExportConfiguration from './ExportConfiguration.vue'

const { t } = useI18n()

const emit = defineEmits(['export', 'reset'])

const step = ref(0)
const showModal = ref(false)
const session = ref({})
const processing = ref(false)

defineProps({
  filter: { type: Object, default: () => ({}) },
  selectedUserIDs: { type: Array, default: () => [] },
})

const components = [ExportConfiguration]

const stepComponent = computed(() => components[step.value])

function onModalHide () {
  step.value = 0
  session.value = {}
  showModal.value = false
}

async function onConfigured (e) {
  processing.value = true
  emit('export', e)
  onReset()
  onClose()
  processing.value = false
}

function onReset () {
  step.value = 0
  emit('reset')
}

function onClose () {
  showModal.value = false
}
</script>
