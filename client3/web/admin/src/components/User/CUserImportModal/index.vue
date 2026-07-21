<template>
  <div class="d-flex">
    <button
      data-test-id="button-import"
      class="btn btn-outline-secondary btn-lg flex-fill"
      @click="showModal = true"
    >
      {{ t('import.buttonLabel') }}
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
              <h5 class="modal-title">{{ t('import.title') }}</h5>
              <button type="button" class="btn-close" @click="showModal = false"></button>
            </div>
            <div class="modal-body p-0">
              <keep-alive>
                <component
                  :is="stepComponent"
                  v-bind="$attrs"
                  @imported="onImported"
                  @reset="onReset"
                  @close="onClose"
                />
              </keep-alive>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import FileUpload from './FileUpload'

const { t } = useI18n()

const emit = defineEmits(['imported', 'reset'])

const step = ref(0)
const showModal = ref(false)

const components = [FileUpload]

const stepComponent = computed(() => components[step.value])

function onModalHide () {
  step.value = 0
  showModal.value = false
}

function onImported (e) {
  emit('imported')
  onReset()
  onClose()
}

function onReset () {
  step.value = 0
  emit('reset')
}

function onClose () {
  showModal.value = false
}
</script>
