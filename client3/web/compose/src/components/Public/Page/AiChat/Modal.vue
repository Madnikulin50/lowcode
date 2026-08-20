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
          class="chat-tools-badge"
          :class="toolsBadgeClass"
          :title="toolsTitle"
          role="img"
          :aria-label="toolsTitle"
        >
          <font-awesome-icon :icon="['fas', 'tools']" />
        </span>
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
        :show-tools-badge="false"
        :model-tools="modelTools"
        @tools-state="onToolsState"
        @export-menu="exportOpen = false"
      />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'page' } })
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useNsI18n } from 'corteza-lib/vue/dist'
import Chat from './Chat.vue'
import { parseModelsPayload, modelToolsEnabled, modelLabel, pickChatModel, readStoredModel, writeStoredModel } from './chatTools.js'
import { usePageStore } from '../../../../store/page'
import { useModuleStore } from '../../../../store/module'
import { useNamespaceStore } from '../../../../store/namespace'

const $t = useNsI18n()

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
const modelTools = ref({})
const selectedModel = ref('')
const liveToolsEnabled = ref(null)
const toolsActive = ref(false)
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

const catalogTools = computed(() => modelToolsEnabled(selectedModel.value, modelTools.value))
const toolsEnabled = computed(() => {
  if (liveToolsEnabled.value !== null) return liveToolsEnabled.value
  if (catalogTools.value !== null) return catalogTools.value
  return false
})
const toolsTitle = computed(() => {
  if (toolsActive.value) return $t('aiChat.tools.invoked')
  return toolsEnabled.value ? $t('aiChat.tools.enabled') : $t('aiChat.tools.disabled')
})
const toolsBadgeClass = computed(() => ({
  on: toolsEnabled.value && !toolsActive.value,
  off: !toolsEnabled.value && !toolsActive.value,
  active: toolsActive.value,
}))

function onToolsState ({ enabled, active } = {}) {
  if (enabled === true || enabled === false) {
    liveToolsEnabled.value = enabled
  } else {
    liveToolsEnabled.value = null
  }
  toolsActive.value = !!active
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
    const parsed = parseModelsPayload(payload)
    const models = parsed.names
    const serverDefault = parsed.defaultModel || ''
    modelOptions.value = models
    modelTools.value = parsed.tools
    if (!models.length) {
      selectedModel.value = ''
      return
    }
    const saved = readStoredModel('aiChat.model')
    selectedModel.value = pickChatModel(models, saved, serverDefault)
    writeStoredModel(selectedModel.value, 'aiChat.model')
    warmUp()
  }).catch(() => {})
}

watch(selectedModel, (model, prev) => {
  if (model !== prev && model) {
    liveToolsEnabled.value = null
    toolsActive.value = false
    warmUp()
    writeStoredModel(model, 'aiChat.model')
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
  const saved = readStoredModel('aiChat.model')
  if (saved) selectedModel.value = saved
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

.chat-tools-badge {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  color: #8a93a0;
  background: #f3f5f8;
  flex-shrink: 0;
}

.chat-tools-badge.on {
  color: #1f7a4d;
  background: #e8f6ee;
}

.chat-tools-badge.off::after {
  content: '';
  position: absolute;
  width: 16px;
  height: 2px;
  background: currentColor;
  transform: rotate(-45deg);
  opacity: 0.85;
}

.chat-tools-badge.active {
  color: #1f4b7a;
  background: #e8eef6;
  animation: chat-tools-pulse 1.2s ease-in-out infinite;
}

@keyframes chat-tools-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.45; }
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
