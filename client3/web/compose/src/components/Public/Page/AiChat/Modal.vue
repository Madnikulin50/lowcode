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
          <select
            v-model="selectedModel"
            class="form-select form-select-sm ms-3"
            style="width: auto; max-width: 200px;"
            :title="$t('aiChat.model.label')"
          >
            <option v-for="m in modelOptions" :key="m" :value="m">{{ m }}</option>
          </select>
          <div class="d-flex gap-1 ms-auto">
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
            :model="selectedModel"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'chat' } })
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
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
const modelOptions = ref([])
const selectedModel = ref('')

const $ComposeAPI = window.__composeAPI

function warmUp () {
  if (!selectedModel.value) return
  $ComposeAPI.pageAiWarmUp({ model: selectedModel.value }).catch(() => {})
}

function loadModels () {
  $ComposeAPI.pageAiModels().then((payload = {}) => {
    const models = payload.models || []
    const serverDefault = payload.default || ''
    modelOptions.value = models
    if (!models.length) {
      selectedModel.value = ''
      return
    }
    let saved = ''
    try { saved = localStorage.getItem('aiChat.model') || '' } catch (e) {}
    if (saved && models.includes(saved)) {
      selectedModel.value = saved
    } else if (serverDefault && models.includes(serverDefault)) {
      selectedModel.value = serverDefault
    } else {
      selectedModel.value = models[0]
    }
    try {
      localStorage.setItem('aiChat.model', selectedModel.value)
    } catch (e) {}
    warmUp()
  }).catch(() => {})
}

watch(selectedModel, (model, prev) => {
  if (model !== prev && model) {
    warmUp()
    try {
      localStorage.setItem('aiChat.model', model)
    } catch (e) {}
  }
})

const dialogClass = computed(() => fullscreen.value ? 'modal-fullscreen' : '')
const contentClass = computed(() => fullscreen.value ? 'h-100' : '')

function startChatModal(data) {
  const { prompt = '', files = [] } = data.detail
  startPrompt.value = prompt
  attachedFiles.value = files
  showModal.value = true
  warmUp()
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
  try {
    const saved = localStorage.getItem('aiChat.model')
    if (saved) selectedModel.value = saved
  } catch (e) {}
  loadModels()
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
