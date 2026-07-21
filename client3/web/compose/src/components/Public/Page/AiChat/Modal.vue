<template>
  <div
    v-if="showModal"
    class="modal d-block"
    tabindex="-1"
    role="dialog"
    @click.self="onHidden"
  >
    <div class="modal-dialog modal-xl" :class="dialogClass" role="document">
      <div class="modal-content" :class="contentClass">
        <Chat
          v-if="showModal"
          :start-prompt="startPrompt"
          :files="attachedFiles"
          :page="page"
          :module="module"
          :namespace="namespace"
          magnified
        />
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

function startChatModal(data) {
  const { prompt = '', files = [] } = data
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
.position-initial {
  position: initial;
}

.modal-max-width {
  max-width: 80vw;
}
</style>
