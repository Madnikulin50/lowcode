<template>
  <div
    v-show="showModal"
    class="chat-dock"
    :class="{ fullscreen }"
    role="dialog"
    aria-modal="false"
    :aria-label="$t('aiChat.title')"
  >
    <div class="chat-dock-header">
      <div class="chat-dock-title-row">
        <h5 class="chat-dock-title mb-0">{{ $t('aiChat.title') }}</h5>
        <div class="d-flex align-items-center gap-1 ms-auto">
          <button
            type="button"
            class="btn btn-outline-secondary border-0 btn-sm"
            :title="$t('aiChat.newChat.label')"
            @click="onNewChat"
          >
            <font-awesome-icon :icon="['fas', 'plus']" />
          </button>
          <div class="export-dropdown position-relative">
            <button
              type="button"
              class="btn btn-outline-secondary border-0 btn-sm"
              :title="$t('aiChat.export.label')"
              @click.stop="exportOpen = !exportOpen"
            >
              <font-awesome-icon :icon="['fas', 'download']" />
            </button>
            <div
              v-if="exportOpen"
              class="export-menu"
              @click="exportOpen = false"
            >
              <button type="button" class="export-menu-item" @click="runExport('markdown')">
                {{ $t('aiChat.export.markdown') }}
              </button>
              <button type="button" class="export-menu-item" @click="runExport('pdf')">
                {{ $t('aiChat.export.pdf') }}
              </button>
              <button type="button" class="export-menu-item" @click="runExport('docx')">
                {{ $t('aiChat.export.docx') }}
              </button>
            </div>
          </div>
          <button
            type="button"
            class="btn btn-outline-secondary border-0 btn-sm"
            :title="fullscreen ? $t('aiChat.collapse') : $t('aiChat.expand')"
            @click="fullscreen = !fullscreen"
          >
            <font-awesome-icon :icon="fullscreen ? ['fas', 'compress'] : ['fas', 'expand']" />
          </button>
          <button
            type="button"
            class="btn btn-outline-secondary border-0 btn-sm"
            :title="$t('aiChat.close')"
            @click="onHidden"
          >
            <font-awesome-icon :icon="['fas', 'times']" />
          </button>
        </div>
      </div>
      <div class="chat-dock-meta">
        <span
          v-if="contextLabel"
          class="chat-context-chip"
          :title="contextLabel"
        >{{ contextLabel }}</span>
        <select
          v-model="selectedModel"
          class="form-select form-select-sm chat-model-select"
          :title="$t('aiChat.model.label')"
        >
          <option v-for="m in modelOptions" :key="m" :value="m">{{ modelLabel(m) }}</option>
        </select>
        <span
          v-if="warmingUp"
          class="d-flex align-items-center gap-1 text-secondary small text-nowrap"
          :title="$t('aiChat.warmup.inProgress')"
        >
          <span class="spinner-border spinner-border-sm" role="status" aria-hidden="true" />
          <span class="d-none d-md-inline">{{ $t('aiChat.warmup.short') }}</span>
        </span>
      </div>
    </div>
    <div class="chat-dock-body">
      <Chat
        ref="chatRef"
        :start-prompt="startPrompt"
        :files="attachedFiles"
        :page="page"
        :module="module"
        :namespace="namespace"
        :magnified="fullscreen"
        :model="selectedModel"
        :active="showModal"
        :framed="false"
        @export-menu="exportOpen = false"
      />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'page' } })
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import Chat from './Chat.vue'
import { usePageStore } from '../../../../store/page'
import { useModuleStore } from '../../../../store/module'
import { useNamespaceStore } from '../../../../store/namespace'

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
const warmingUp = ref(false)
const exportOpen = ref(false)
const chatRef = ref(null)
let warmUpSeq = 0

const $ComposeAPI = window.__composeAPI
const pageStore = usePageStore()
const moduleStore = useModuleStore()
const namespaceStore = useNamespaceStore()

const contextLabel = computed(() => {
  const pageID = String(props.page || '')
  const moduleID = String(props.module || '')
  const nsID = String(props.namespace || '')
  const page = pageStore.getByID(pageID) || pageStore.getByID(props.page)
  const mod = moduleStore.getByID(moduleID) || moduleStore.getByID(props.module)
  const ns = namespaceStore.getByID(nsID) || namespaceStore.getByID(props.namespace)
  const parts = []
  if (page?.title) parts.push(page.title)
  else if (mod?.name) parts.push(mod.name)
  if (ns?.name && parts[0] !== ns.name) parts.push(ns.name)
  return parts.join(' · ')
})

function modelLabel (id) {
  if (!id) return ''
  const [name, tag] = String(id).split(':')
  const pretty = name.replace(/[-_]/g, ' ').replace(/\b\w/g, c => c.toUpperCase())
  return tag ? `${pretty} (${tag})` : pretty
}

function warmUp () {
  if (!selectedModel.value) return
  const seq = ++warmUpSeq
  warmingUp.value = true
  $ComposeAPI.pageAiWarmUp({ model: selectedModel.value }).catch(() => {}).finally(() => {
    if (seq === warmUpSeq) {
      warmingUp.value = false
    }
  })
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

function startChatModal (data) {
  const { prompt = '', files = [] } = data.detail || {}
  startPrompt.value = prompt
  attachedFiles.value = files || []
  showModal.value = true
  warmUp()
  requestAnimationFrame(() => {
    chatRef.value?.applyIncomingPrompt?.(prompt, files)
    chatRef.value?.focusInput?.()
  })
}

function onHidden () {
  showModal.value = false
  exportOpen.value = false
}

function onNewChat () {
  chatRef.value?.newChat?.()
}

function runExport (kind) {
  const chat = chatRef.value
  if (!chat) return
  if (kind === 'markdown') chat.exportMarkdown()
  else if (kind === 'pdf') chat.exportPdf()
  else if (kind === 'docx') chat.exportDocx()
}

function onKeydown (e) {
  if (e.key === 'Escape' && showModal.value) {
    e.preventDefault()
    onHidden()
  }
}

function onDocumentClick (e) {
  if (!e.target.closest('.export-dropdown')) {
    exportOpen.value = false
  }
}

onMounted(() => {
  try {
    const saved = localStorage.getItem('aiChat.model')
    if (saved) selectedModel.value = saved
  } catch (e) {}
  loadModels()
  window.addEventListener('show-chat-modal', startChatModal)
  document.addEventListener('keydown', onKeydown)
  document.addEventListener('click', onDocumentClick)
})

onBeforeUnmount(() => {
  window.removeEventListener('show-chat-modal', startChatModal)
  document.removeEventListener('keydown', onKeydown)
  document.removeEventListener('click', onDocumentClick)
})
</script>

<style lang="scss">
.chat-dock {
  position: fixed;
  top: var(--topbar-height, 64px);
  right: 0;
  bottom: 0;
  width: min(100vw, max(440px, 40vw));
  z-index: 1045;
  display: flex;
  flex-direction: column;
  background: #fff;
  border-left: 1px solid #e0e0e0;
}

.chat-dock.fullscreen {
  top: 0;
  width: 100%;
  left: 0;
  border-left: none;
}

.chat-dock-header {
  flex-shrink: 0;
  padding: 10px 12px 8px;
  border-bottom: 1px solid #e0e0e0;
  background: #fff;
}

.chat-dock-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.chat-dock-title {
  font-size: 0.95rem;
  font-weight: 600;
  line-height: 1.3;
}

.chat-dock-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  min-width: 0;
}

.chat-context-chip {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 12px;
  color: #667788;
  background: #f3f5f8;
  border-radius: 999px;
  padding: 2px 10px;
}

.chat-model-select {
  width: auto;
  max-width: 160px;
  flex-shrink: 0;
}

.chat-dock-body {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.chat-dock .export-menu {
  position: absolute;
  right: 0;
  top: 100%;
  margin-top: 4px;
  min-width: 160px;
  background: #fff;
  border: 1px solid #ddd;
  border-radius: 6px;
  z-index: 20;
  overflow: hidden;
}

.chat-dock .export-menu-item {
  display: block;
  width: 100%;
  border: none;
  background: transparent;
  padding: 8px 14px;
  font-size: 13px;
  text-align: left;
  cursor: pointer;
  color: #333;
}

.chat-dock .export-menu-item:hover {
  background: #f0f0f0;
}
</style>
