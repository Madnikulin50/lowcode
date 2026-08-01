<template>
  <div
    v-if="showModal"
    class="modal d-block"
    tabindex="-1"
    role="dialog"
    @click.self="onHidden"
  >
    <div class="modal-dialog" :class="dialogClass" role="document">
      <div class="modal-content d-flex flex-column" :class="contentClass">
        <div class="modal-header py-2 px-3 border-bottom">
          <h5 class="modal-title">LowCoooode AI-assistant</h5>
          <div class="d-flex gap-1">
            <button
              class="btn btn-outline-secondary border-0 btn-sm"
              :title="fullscreen ? 'Collapse' : 'Expand'"
              @click="fullscreen = !fullscreen"
            >
              <font-awesome-icon :icon="fullscreen ? ['fas', 'compress'] : ['fas', 'expand']" />
            </button>
            <button
              class="btn btn-outline-secondary border-0 btn-sm"
              title="Close"
              @click="onHidden"
            >
              <font-awesome-icon :icon="['fas', 'times']" size="lg" />
            </button>
          </div>
        </div>
        <div class="modal-body p-0 d-flex flex-column" style="min-height: 0">
          <Chat
            v-if="showModal"
            :start-prompt="startPrompt"
            :files="attachedFiles"
            :page="page"
            :module="module"
            :namespace="namespace"
            :magnified="fullscreen"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import Chat from './Chat.vue'

const props = defineProps({
  page: { type: String, required: false, default: '' },
  module: { type: String, required: false, default: '' },
  namespace: { type: String, required: false, default: '' },
})

const showModal = ref(false)
const startPrompt = ref('')
const attachedFiles = ref([])
const fullscreen = ref(false)

const dialogClass = computed(() => fullscreen.value ? 'modal-fullscreen' : '')
const contentClass = computed(() => fullscreen.value ? 'h-100' : '')

function startChatModal(data) {
  const { prompt = '', files = [] } = data.detail
  startPrompt.value = prompt
  attachedFiles.value = files
  showModal.value = true
}

function onHidden() {
  showModal.value = false
}

function setDefaultValues() {
  showModal.value = false
  startPrompt.value = ''
  attachedFiles.value = []
}

onMounted(() => {
  window.addEventListener('show-chat-modal', startChatModal)
})

onBeforeUnmount(() => {
  window.removeEventListener('show-chat-modal', startChatModal)
  setDefaultValues()
})
</script>

<style lang="scss">
.modal-dialog:not(.modal-fullscreen) {
  max-width: 70vw;
}
.position-initial {
  position: initial;
}
</style>
