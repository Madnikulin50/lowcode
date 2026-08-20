<template>
  <div class="chat-container" :class="{ frameless: !framed }">
    <div
      v-if="showModelSwitcher || showToolsBadge"
      class="chat-meta-bar"
    >
      <select
        v-if="showModelSwitcher"
        v-model="selectedModel"
        class="form-select form-select-sm chat-model-select"
        :title="$t('aiChat.model.label')"
        :disabled="!modelOptions.length || loading"
        @change="onModelPicked"
      >
        <option v-for="m in modelOptions" :key="m" :value="m">{{ modelLabel(m) }}</option>
      </select>
      <span
        v-if="showModelSwitcher && preloadWarming"
        class="d-flex align-items-center gap-1 text-secondary small text-nowrap"
        :title="$t('aiChat.warmup.inProgress')"
      >
        <span class="spinner-border spinner-border-sm" role="status" aria-hidden="true" />
        <span>{{ $t('aiChat.warmup.short') }}</span>
      </span>
      <span
        v-if="showToolsBadge"
        class="chat-tools-badge"
        :class="[toolsBadgeClass, { 'ms-auto': showModelSwitcher }]"
        :title="toolsTitle"
        role="img"
        :aria-label="toolsTitle"
      >
        <font-awesome-icon :icon="['fas', 'tools']" />
      </span>
    </div>
    <div
      v-if="isEmpty"
      class="empty-state"
    >
      <div class="empty-greeting">{{ $t('aiChat.greeting') }}</div>
      <div class="suggestion-row">
        <button
          v-for="s in suggestions"
          :key="s.key"
          type="button"
          class="suggestion-chip"
          @click="useSuggestion(s.text)"
        >{{ s.text }}</button>
      </div>
    </div>
    <div
      v-else
      ref="messagesContainer"
      class="messages"
      @scroll="onScroll"
    >
      <div
        v-for="(msg, idx) in messages"
        :key="idx"
        :class="['message', msg.role, { active: msg.active }]"
      >
        <div class="avatar">
          <font-awesome-icon :icon="['fas', msg.role === 'user' ? 'user' : 'brain']" />
        </div>
        <div class="message-body">
          <div
            v-if="msg.role === 'assistant' && (msg.usedTools || (msg.active && toolsActive))"
            class="msg-tools-flag"
            :class="{ active: msg.active && toolsActive }"
            :title="$t('aiChat.tools.invoked')"
          >
            <font-awesome-icon :icon="['fas', 'tools']" size="xs" />
            <span>{{ $t('aiChat.tools.invoked') }}</span>
          </div>
          <div
            v-if="msg.role === 'assistant' && (msg.reasoning || (msg.active && streamStatus === 'thinking'))"
            class="reasoning"
          >
            <button
              type="button"
              class="reasoning-toggle"
              @click="msg.reasoningOpen = !msg.reasoningOpen"
            >
              <font-awesome-icon :icon="['fas', msg.reasoningOpen ? 'chevron-up' : 'chevron-down']" size="xs" />
              {{ $t('aiChat.reasoning.label') }}
            </button>
            <div
              v-show="msg.reasoningOpen && msg.reasoning"
              class="reasoning-body"
            >{{ msg.reasoning }}</div>
          </div>
          <div v-show="!msg.collapsed" class="content expanded-content">
            <div
              v-if="promptXml(msg.content)"
              class="prompt-xml"
            >
              <button
                type="button"
                class="prompt-xml-toggle"
                @click="msg.xmlOpen = !msg.xmlOpen"
              >
                <font-awesome-icon :icon="['fas', msg.xmlOpen ? 'chevron-up' : 'chevron-down']" size="xs" />
                {{ $t('aiChat.promptXml.label') }}
              </button>
              <pre
                v-show="msg.xmlOpen"
                class="prompt-xml-body"
              >{{ promptXml(msg.content) }}</pre>
            </div>
            <div class="content-text">
              <template v-if="!msg.content && msg.active">
                <div v-if="warmingUp" class="warmup-indicator">
                  <span class="spinner-border spinner-border-sm text-secondary" role="status" aria-hidden="true" />
                  <span>{{ statusLabel }}</span>
                </div>
                <div v-else class="typing-indicator">
                  <span /><span /><span />
                </div>
              </template>
              <template v-for="(part, pi) in messageParts(msg.content)" :key="pi">
                <div v-if="part.kind === 'html'" class="chat-md" v-html="part.html" />
                <div
                  v-else-if="part.kind === 'chart'"
                  class="chat-chart"
                >
                  <e-charts
                    :option="part.option"
                    autoresize
                    class="position-absolute w-100 h-100 overflow-hidden"
                  />
                </div>
                <div
                  v-else-if="part.kind === 'compose-chart'"
                  class="chat-chart"
                >
                  <chart-component
                    v-if="composeChartRecord(part.spec)"
                    :key="part.spec.chartID"
                    class="h-100 w-100"
                    :chart="composeChartRecord(part.spec)"
                    :reporter="composeChartReporter(part.spec)"
                  />
                  <div v-else-if="composeChartError(part.spec)" class="chat-chart-error">
                    {{ composeChartError(part.spec) }}
                  </div>
                  <div v-else class="chat-chart-loading">
                    <span class="spinner-border spinner-border-sm" />
                  </div>
                </div>
                <div v-else-if="part.kind === 'chart-error'" class="chat-chart-error">
                  {{ $t('aiChat.chart.error') }}
                </div>
                <div v-else-if="part.kind === 'confirm'" class="confirm-card">
                  <div class="confirm-title">{{ $t('aiChat.confirm.title') }}</div>
                  <ul class="confirm-tools">
                    <li v-for="(tool, ti) in (part.spec?.tools || [])" :key="ti">
                      {{ tool.summary || tool.name }}
                    </li>
                  </ul>
                  <div v-if="canConfirm(idx, msg)" class="confirm-actions">
                    <button type="button" class="btn btn-sm btn-primary" @click="resolveConfirm(true)">
                      {{ $t('aiChat.confirm.execute') }}
                    </button>
                    <button type="button" class="btn btn-sm btn-outline-secondary" @click="resolveConfirm(false)">
                      {{ $t('aiChat.confirm.cancel') }}
                    </button>
                  </div>
                </div>
              </template>
            </div>
            <button
              v-if="hasManyLines(msg.content) && !msg.active"
              class="collapse-btn"
              @click="toggleCollapse(idx)"
              :title="$t('aiChat.collapse')"
            >
              <font-awesome-icon :icon="['fas', 'chevron-up']" size="xs" />
            </button>
          </div>
          <div v-show="msg.collapsed" class="content collapsed-content" @click="toggleCollapse(idx)">
            <div class="content-text" v-html="formatMessage(preview(msg.content))" />
            <button
              class="collapse-btn"
              @click.stop="toggleCollapse(idx)"
              :title="$t('aiChat.expand')"
            >
              <font-awesome-icon :icon="['fas', 'chevron-down']" size="xs" />
            </button>
          </div>
          <div v-if="!msg.active && (msg.content || msg.role === 'user')" class="msg-actions">
            <button type="button" class="msg-action" :title="$t('aiChat.copy')" @click="copyMessage(msg)">
              <font-awesome-icon :icon="copiedIdx === idx ? ['fas', 'check'] : ['fas', 'copy']" size="xs" />
            </button>
            <button
              v-if="msg.role === 'user'"
              type="button"
              class="msg-action"
              :title="$t('aiChat.edit')"
              @click="editMessage(idx)"
            >
              <font-awesome-icon :icon="['fas', 'pen']" size="xs" />
            </button>
            <button
              v-if="msg.role === 'assistant' && idx === lastAssistantIdx"
              type="button"
              class="msg-action"
              :title="$t('aiChat.retry')"
              @click="retryLast"
            >
              <font-awesome-icon :icon="['fas', 'sync']" size="xs" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="input-area">
      <div v-if="attachedFiles.length" class="file-chips">
        <div v-for="(f, i) in attachedFiles" :key="i" class="file-chip">
          <span class="file-chip-name">{{ f.name }}</span>
          <button type="button" class="file-chip-download" @click="downloadFile(f)" :title="$t('aiChat.export.label')">
            <font-awesome-icon :icon="['fas', 'download']" size="xs" />
          </button>
          <button type="button" class="file-chip-download" @click="removeFile(i)" :title="$t('aiChat.attach.remove')">
            <font-awesome-icon :icon="['fas', 'times']" size="xs" />
          </button>
        </div>
      </div>
      <div class="composer">
        <input
          ref="fileInput"
          type="file"
          class="d-none"
          multiple
          accept=".csv,.txt,.json,.md,.tsv,.xml"
          @change="onAttach"
        >
        <button
          type="button"
          class="composer-icon"
          :title="$t('aiChat.attach.label')"
          @click="fileInput?.click()"
        >
          <font-awesome-icon :icon="['fas', 'paperclip']" />
        </button>
        <textarea
          ref="inputEl"
          v-model="inputText"
          :placeholder="$t('aiChat.sendMessage.placeholder')"
          rows="2"
          @input="autoGrow"
          @keydown="onComposerKeydown"
        />
        <button
          v-if="loading"
          type="button"
          class="composer-send stop"
          @click="stopGeneration"
        >
          <font-awesome-icon :icon="['fas', 'stop']" />
          {{ $t('aiChat.sendMessage.stop') }}
        </button>
        <button
          v-else
          type="button"
          class="composer-send"
          :disabled="!inputText.trim()"
          @click="sendMessage()"
        >
          {{ $t('aiChat.sendMessage.button') }}
        </button>
      </div>
      <div class="composer-hint">
        <span v-if="loading">{{ statusLabel }}</span>
        <span v-else>{{ $t('aiChat.sendMessage.hint') }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'page' } })
import { ref, reactive, computed, watch, nextTick, onMounted, onBeforeUnmount, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import markdownIt from 'markdown-it'
import html2pdf from 'html2pdf.js'
import { Document, Packer, Paragraph, TextRun, ExternalHyperlink, HeadingLevel, AlignmentType, NumberFormat, WidthType, BorderStyle, ShadingType, Table, TableRow, TableCell } from 'docx'
import ECharts from 'vue-echarts'
import { splitChartParts, replaceChartFences } from './chatChart.js'
import { parseModelsPayload, modelToolsEnabled, modelLabel, pickChatModel, readStoredModel, writeStoredModel } from './chatTools.js'
import { useStore } from '../../../../store'
import ChartComponent from '../../../Chart/index.vue'

const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  startPrompt: { type: String, required: false, default: '' },
  page: { type: String, required: false, default: '' },
  module: { type: String, required: false, default: '' },
  namespace: { type: String, required: false, default: '' },
  magnified: { type: Boolean, default: false },
  files: { type: Array, required: false, default: () => [] },
  model: { type: String, required: false, default: '' },
  preferredModel: { type: String, required: false, default: '' },
  modelStorageKey: { type: String, default: 'aiChat.model' },
  active: { type: Boolean, default: true },
  framed: { type: Boolean, default: true },
  showModelSwitcher: { type: Boolean, default: false },
  showToolsBadge: { type: Boolean, default: true },
  modelTools: { type: Object, default: null },
})

const emit = defineEmits(['tools-state'])

const store = useStore()
const $ComposeAPI = inject('$ComposeAPI', window.__composeAPI)

function makeChatID () {
  return `chat-${Date.now()}-${Math.random().toString(36).substring(2, 10)}`
}

const messages = ref([])
const inputText = ref('')
const loading = ref(false)
const warmingUp = ref(false)
const streamStatus = ref('')
const messagesContainer = ref(null)
const inputEl = ref(null)
const fileInput = ref(null)
const chatID = ref(makeChatID())
const exportOpen = ref(false)
const abortController = ref(null)
let pendingAutoSend = null
const attachedFiles = ref([...(props.files || [])])
const stickToBottom = ref(true)
const copiedIdx = ref(-1)
let persistTimer = null
let copiedTimer = null
const localModelTools = ref({})
const defaultModel = ref('')
const modelOptions = ref([])
const selectedModel = ref('')
const preloadWarming = ref(false)
const sessionTools = ref(null)
const toolsActive = ref(false)
let warmUpSeq = 0

const resolvedModel = computed(() => props.model || selectedModel.value || defaultModel.value)
const toolsLookup = computed(() => props.modelTools || localModelTools.value)
const catalogTools = computed(() => modelToolsEnabled(resolvedModel.value, toolsLookup.value))
const toolsEnabled = computed(() => {
  if (sessionTools.value !== null) return sessionTools.value
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

function emitToolsState () {
  emit('tools-state', {
    enabled: sessionTools.value,
    active: toolsActive.value,
  })
}

function warmUpSelected () {
  if (!props.showModelSwitcher || !selectedModel.value || !$ComposeAPI?.pageAiWarmUp) return
  const seq = ++warmUpSeq
  preloadWarming.value = true
  $ComposeAPI.pageAiWarmUp({ model: selectedModel.value }).catch(() => {}).finally(() => {
    if (seq === warmUpSeq) preloadWarming.value = false
  })
}

function applyCatalogSelection (names, serverDefault) {
  if (!props.showModelSwitcher || !names.length) return
  const saved = readStoredModel(props.modelStorageKey)
  selectedModel.value = pickChatModel(names, saved, props.preferredModel || serverDefault)
  warmUpSelected()
}

function onModelPicked () {
  if (!props.showModelSwitcher || !selectedModel.value) return
  writeStoredModel(selectedModel.value, props.modelStorageKey)
  warmUpSelected()
}

function loadModelTools () {
  if (!$ComposeAPI?.pageAiModels) return
  if (!props.showModelSwitcher && props.modelTools) return
  $ComposeAPI.pageAiModels().then((payload = {}) => {
    const parsed = parseModelsPayload(payload)
    if (!props.modelTools) localModelTools.value = parsed.tools
    if (parsed.defaultModel) defaultModel.value = parsed.defaultModel
    modelOptions.value = parsed.names
    applyCatalogSelection(parsed.names, parsed.defaultModel)
    emitToolsState()
  }).catch(() => {})
}

const isEmpty = computed(() => !messages.value.some(m => m.role === 'user' || (m.role === 'assistant' && m.content)))
const lastAssistantIdx = computed(() => {
  for (let i = messages.value.length - 1; i >= 0; i--) {
    if (messages.value[i].role === 'assistant') return i
  }
  return -1
})
const statusLabel = computed(() => {
  if (warmingUp.value || streamStatus.value === 'warming') return $t('aiChat.status.warming')
  if (streamStatus.value === 'using-tools' || toolsActive.value) return $t('aiChat.status.usingTools')
  if (streamStatus.value === 'writing') return $t('aiChat.status.writing')
  if (loading.value) return $t('aiChat.status.thinking')
  return ''
})
const suggestions = computed(() => [
  { key: 'capabilities', text: $t('aiChat.empty.suggestions.capabilities') },
  { key: 'modules', text: $t('aiChat.empty.suggestions.modules') },
  { key: 'page', text: $t('aiChat.empty.suggestions.page') },
])

function hasManyLines (text) {
  const body = splitPromptXml(text).body
  if (!body) return false
  const breaks = (body.match(/\n/g) || []).length
  return breaks >= 4
}

function toggleCollapse(idx) {
  messages.value[idx].collapsed = !messages.value[idx].collapsed
}

function downloadFile(f) {
  const blob = new Blob([f.content], { type: f.type || 'text/csv' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = f.name
  a.click()
  URL.revokeObjectURL(url)
}

function removeFile(i) {
  attachedFiles.value.splice(i, 1)
}

function sessionKey() {
  return `aiChat.session.${props.namespace || '0'}.${props.page || '0'}`
}

function persistSession() {
  try {
    const usable = messages.value.filter(m => m.role && (m.content || m.reasoning))
    if (!usable.length) {
      localStorage.removeItem(sessionKey())
      return
    }
    const payload = {
      chatID: chatID.value,
      messages: usable.slice(-40).map(m => ({
        role: m.role,
        content: String(m.content || '').slice(0, 50000),
        reasoning: String(m.reasoning || '').slice(0, 20000),
        collapsed: !!m.collapsed,
        usedTools: !!m.usedTools,
      })),
    }
    localStorage.setItem(sessionKey(), JSON.stringify(payload))
  } catch (e) {}
}

function restoreSession() {
  try {
    const raw = localStorage.getItem(sessionKey())
    if (!raw) return false
    const payload = JSON.parse(raw)
    if (!payload || !Array.isArray(payload.messages) || !payload.messages.length) return false
    chatID.value = payload.chatID || makeChatID()
    messages.value = payload.messages.map(m => ({
      role: m.role,
      content: m.content || '',
      reasoning: m.reasoning || '',
      reasoningOpen: false,
      collapsed: !!m.collapsed,
      usedTools: !!m.usedTools,
      active: false,
    }))
    return true
  } catch (e) {
    return false
  }
}

function schedulePersist() {
  clearTimeout(persistTimer)
  persistTimer = setTimeout(persistSession, 400)
}

function focusInput() {
  nextTick(() => {
    const el = inputEl.value
    if (!el) return
    el.focus()
    autoGrow()
  })
}

function autoGrow() {
  const el = inputEl.value
  if (!el) return
  el.style.height = 'auto'
  el.style.height = Math.min(el.scrollHeight, 160) + 'px'
}

function onComposerKeydown(e) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    sendMessage()
  }
}

function onScroll() {
  const el = messagesContainer.value
  if (!el) return
  stickToBottom.value = el.scrollHeight - el.scrollTop - el.clientHeight < 96
}

function useSuggestion(text) {
  inputText.value = text
  sendMessage()
}

function canConfirm(idx, msg) {
  return !loading.value && !msg.confirmDone && idx === messages.value.length - 1
}

function resolveConfirm(ok) {
  const last = messages.value[messages.value.length - 1]
  if (last) last.confirmDone = true
  sendMessage(ok ? $t('aiChat.confirm.execute') : $t('aiChat.confirm.cancel'), {
    apiPrompt: ok ? 'да' : 'отмена',
  })
}

function copyMessage(msg) {
  const idx = messages.value.indexOf(msg)
  const text = stripHtml(preview(msg.content || '')).replace(/\u2026$/, '')
  const plain = stripXmlContent(replaceChartFences(String(msg.content || ''))).replace(/<[^>]*>/g, '')
  navigator.clipboard?.writeText(plain || text).catch(() => {})
  copiedIdx.value = idx
  clearTimeout(copiedTimer)
  copiedTimer = setTimeout(() => { copiedIdx.value = -1 }, 1500)
}

function editMessage(idx) {
  const msg = messages.value[idx]
  if (!msg || msg.role !== 'user') return
  inputText.value = msg.content || ''
  messages.value = messages.value.slice(0, idx)
  schedulePersist()
  focusInput()
}

function retryLast() {
  const lastUser = [...messages.value].reverse().find(m => m.role === 'user')
  if (!lastUser) return
  const idx = messages.value.lastIndexOf(lastUser)
  const text = lastUser.content
  messages.value = messages.value.slice(0, idx)
  sendMessage(text)
}

function newChat() {
  if (abortController.value) {
    abortController.value.abort()
    abortController.value = null
  }
  loading.value = false
  warmingUp.value = false
  streamStatus.value = ''
  messages.value = []
  attachedFiles.value = []
  inputText.value = ''
  chatID.value = makeChatID()
  persistSession()
  focusInput()
}

function stopGeneration() {
  if (abortController.value) {
    abortController.value.abort()
    abortController.value = null
  }
}

function onAttach(e) {
  const list = Array.from(e.target.files || [])
  e.target.value = ''
  for (const file of list) {
    if (file.size > 1024 * 1024) continue
    const reader = new FileReader()
    reader.onload = () => {
      attachedFiles.value.push({
        name: file.name,
        content: String(reader.result || ''),
        type: file.type || 'text/plain',
      })
    }
    reader.readAsText(file)
  }
}

function preview(text) {
  const hidden = stripXmlContent(text).replace(/\r\n/g, '\n').replace(/\r/g, '\n')
  const noCharts = replaceChartFences(hidden)
  const noCode = stripCodeBlocks(noCharts)
  const plain = noCode.replace(/<[^>]*>/g, '')
  return plain.length > 120 ? plain.slice(0, 120) + '\u2026' : plain
}

function scrollToBottom() {
  const container = messagesContainer.value
  nextTick(() => {
    if (container !== null && stickToBottom.value) {
      container.scrollTop = container.scrollHeight
    }
  })
}

const HTML_TAGS = new Set([
  'a', 'abbr', 'b', 'blockquote', 'br', 'code', 'del', 'div', 'em',
  'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'hr', 'i', 'img', 'li', 'ol',
  'p', 'pre', 's', 'span', 'strong', 'sub', 'sup', 'table', 'tbody',
  'td', 'th', 'thead', 'tr', 'ul', 'u', 'mark', 'kbd', 'figure', 'figcaption',
])

function splitPromptXml (text) {
  let rest = String(text || '').replace(/&lt;/g, '<').replace(/&gt;/g, '>')
  const blocks = []
  const re = /<([A-Za-z_][\w:.-]*)(?:\s[^>]*)?>[\s\S]*?<\/\1>/g
  rest = rest.replace(re, (full, tag) => {
    if (HTML_TAGS.has(String(tag).toLowerCase()) || String(tag).toLowerCase() === 'tool') {
      return full
    }
    blocks.push(full.trim())
    return '\n'
  })
  return {
    body: rest.replace(/[ \t]+\n/g, '\n').replace(/\n{3,}/g, '\n\n').trim(),
    xml: blocks.join('\n\n'),
  }
}

function promptXml (text) {
  return splitPromptXml(text).xml
}

function stripXmlContent (text) {
  return splitPromptXml(text).body
}

function stripToolXml(text) {
  return String(text || '').replace(/<tool\b[^>]*>[\s\S]*?<\/tool>/gi, '').replace(/\n{3,}/g, '\n\n')
}

// Skip ```chart / ```echarts fences so they can be rendered as figures.
function stripCodeBlocks(text) {
  return String(text || '')
    .replace(/```(?!(?:\s*(?:compose-chart|chart|echarts|json|chat-confirm))\b)[\s\S]*?(```|$)/gi, '')
    .replace(/\n{3,}/g, '\n\n')
}

const mdRenderer = markdownIt({ html: false, linkify: true, breaks: true })

function formatMarkdown (text) {
  const src = String(text || '').replace(/\r\n/g, '\n').replace(/\r/g, '\n').trim()
  if (!src) return ''
  return mdRenderer.render(src)
}

function formatMessage (text) {
  return formatMarkdown(splitPromptXml(text).body)
}

function normalizeMessageText (text) {
  return stripToolXml(String(text || '').replace(/\r\n/g, '\n').replace(/\r/g, '\n'))
}

function messageParts (text) {
  const { body } = splitPromptXml(normalizeMessageText(text))
  return splitChartParts(body).map(part => {
    if (part.kind === 'chart') {
      return { kind: 'chart', option: part.option }
    }
    if (part.kind === 'compose-chart') {
      return { kind: 'compose-chart', spec: part.spec }
    }
    if (part.kind === 'chart-error') {
      return { kind: 'chart-error' }
    }
    if (part.kind === 'confirm') {
      return { kind: 'confirm', spec: part.spec || { tools: [] } }
    }
    return { kind: 'html', html: formatMarkdown(part.text) }
  })
}

const composeCharts = reactive({})
const composeChartInflight = new Set()
const composeChartReporters = new Map()

function composeChartKey(spec) {
  return String(spec?.chartID || '')
}

function composeChartRecord(spec) {
  const st = composeCharts[composeChartKey(spec)]
  return st && st.chart ? st.chart : null
}

function composeChartError(spec) {
  const st = composeCharts[composeChartKey(spec)]
  return st && st.error ? String(st.error) : ''
}

function namespaceIDOf(spec) {
  if (spec?.namespaceID != null && String(spec.namespaceID).trim()) {
    return String(spec.namespaceID).trim()
  }
  const ns = props.namespace
  if (ns && typeof ns === 'object') return String(ns.namespaceID || ns.ID || '')
  return String(ns || '')
}

function composeChartReporter(spec) {
  const chartID = composeChartKey(spec)
  const namespaceID = namespaceIDOf(spec)
  const cacheKey = `${namespaceID}:${chartID}`
  if (composeChartReporters.has(cacheKey)) return composeChartReporters.get(cacheKey)
  const reporter = (r = {}) => {
    const f = String(r.filter || '')
    if (f.includes('${record') || f.includes('${ownerID}')) {
      return Promise.resolve([])
    }
    return $ComposeAPI.recordReport({ namespaceID, ...r })
  }
  composeChartReporters.set(cacheKey, reporter)
  return reporter
}

async function loadComposeChart(spec) {
  const chartID = composeChartKey(spec)
  if (!chartID) return
  if (composeChartInflight.has(chartID) || composeCharts[chartID]?.chart || composeCharts[chartID]?.error) {
    return
  }
  composeChartInflight.add(chartID)
  composeCharts[chartID] = { loading: true, chart: null, error: null }
  const namespaceID = namespaceIDOf(spec)
  try {
    if (!namespaceID) {
      throw new Error('Missing namespaceID')
    }
    const chart = await store.chart.findByID({ namespaceID, chartID })
    if (!chart) {
      throw new Error(`Chart ${chartID} not found`)
    }
    const ns = store.namespace.getByID(namespaceID) || { namespaceID }
    const reports = chart?.config?.reports || []
    for (const report of reports) {
      const moduleID = report?.moduleID != null ? String(report.moduleID) : ''
      if (!moduleID) continue
      const existing = store.module.getByID(moduleID) || store.module.getByID(report.moduleID)
      if (!existing) {
        await store.module.findByID({ namespace: ns, moduleID })
      }
    }
    composeCharts[chartID] = { loading: false, chart, error: null }
  } catch (e) {
    composeCharts[chartID] = {
      loading: false,
      chart: null,
      error: e?.message || $t('aiChat.chart.error') || 'Could not render this chart',
    }
  } finally {
    composeChartInflight.delete(chartID)
  }
}

watch(
  () => messages.value.map(m => m.content).join('\n\x1e'),
  () => {
    for (const msg of messages.value) {
      if (!msg?.content) continue
      for (const part of splitChartParts(normalizeMessageText(msg.content))) {
        if (part.kind === 'compose-chart' && part.spec) {
          loadComposeChart(part.spec)
        }
      }
    }
  },
  { flush: 'post', immediate: true },
)

function closeExport(e) {
  if (!e.target.closest('.export-dropdown')) {
    exportOpen.value = false
  }
}

function stripHtml(text) {
  const div = document.createElement('div')
  div.innerHTML = text
  return div.textContent || div.innerText || ''
}

function downloadBlob(blob, filename) {
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

function exportStamp() {
  const d = new Date()
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}_${pad(d.getHours())}-${pad(d.getMinutes())}`
}

function exportableMessages() {
  return messages.value.filter(m => m.content && String(m.content).trim())
}

function formatMessageForExport(text) {
  let normalized = String(text || '').replace(/\r\n/g, '\n').replace(/\r/g, '\n')
  normalized = replaceChartFences(stripXmlContent(normalized))
  const md = markdownIt({ html: false, linkify: true, breaks: true })
  return md.render(normalized)
}

function buildExportDocument() {
  const items = exportableMessages()
  const when = new Date().toLocaleString()
  const model = resolvedModel.value || '—'
  const bubbles = items.map((m, i) => {
    const isUser = m.role === 'user'
    const roleLabel = isUser ? 'User' : 'Assistant'
    const body = formatMessageForExport(m.content)
    return `
      <article class="bubble ${isUser ? 'user' : 'assistant'}">
        <header class="bubble-meta">
          <span class="role">${roleLabel}</span>
          <span class="idx">#${i + 1}</span>
        </header>
        <div class="bubble-body">${body}</div>
      </article>`
  }).join('\n')

  return `
    <div class="export-root">
      <header class="export-header">
        <div class="brand">AI Chat</div>
        <h1>Chat export</h1>
        <div class="meta">
          <span>${when}</span>
          <span>·</span>
          <span>Model: ${stripHtml(model)}</span>
          <span>·</span>
          <span>${items.length} messages</span>
        </div>
      </header>
      <main class="export-body">
        ${bubbles || '<p class="empty">No messages to export.</p>'}
      </main>
      <footer class="export-footer">Generated from Compose AI Chat</footer>
    </div>`
}

const exportPdfStyles = `
  .export-root {
    font-family: "Segoe UI", system-ui, -apple-system, sans-serif;
    color: #1a1a1a;
    background: #fff;
    padding: 8px 4px 24px;
    line-height: 1.55;
  }
  .export-header {
    border-bottom: 2px solid #1f4b7a;
    padding-bottom: 14px;
    margin-bottom: 22px;
  }
  .export-header .brand {
    font-size: 11px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: #1f4b7a;
    font-weight: 700;
    margin-bottom: 4px;
  }
  .export-header h1 {
    margin: 0 0 8px;
    font-size: 22px;
    font-weight: 700;
    color: #12263a;
  }
  .export-header .meta {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    font-size: 12px;
    color: #667788;
  }
  .bubble {
    border: 1px solid #e3e8ef;
    border-radius: 10px;
    padding: 12px 14px;
    margin-bottom: 14px;
    page-break-inside: avoid;
    background: #fafbfc;
  }
  .bubble.user {
    border-left: 4px solid #1f4b7a;
    background: #f3f7fb;
  }
  .bubble.assistant {
    border-left: 4px solid #2f9e6b;
    background: #f6fbf8;
  }
  .bubble-meta {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.06em;
  }
  .bubble.user .role { color: #1f4b7a; font-weight: 700; }
  .bubble.assistant .role { color: #2f9e6b; font-weight: 700; }
  .bubble-meta .idx { color: #99a3ad; }
  .bubble-body { font-size: 13px; color: #243041; }
  .bubble-body p { margin: 0 0 0.6em; }
  .bubble-body p:last-child { margin-bottom: 0; }
  .bubble-body h1, .bubble-body h2, .bubble-body h3 {
    margin: 0.8em 0 0.4em;
    color: #12263a;
    line-height: 1.25;
  }
  .bubble-body h1 { font-size: 1.25em; }
  .bubble-body h2 { font-size: 1.12em; }
  .bubble-body h3 { font-size: 1.05em; }
  .bubble-body ul, .bubble-body ol { margin: 0.4em 0 0.6em 1.2em; padding: 0; }
  .bubble-body li { margin: 0.15em 0; }
  .bubble-body code {
    font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
    font-size: 0.92em;
    background: #eef2f6;
    padding: 0.1em 0.35em;
    border-radius: 4px;
  }
  .bubble-body pre {
    background: #12263a;
    color: #e8eef5;
    padding: 10px 12px;
    border-radius: 8px;
    overflow-x: auto;
    font-size: 11px;
    line-height: 1.45;
  }
  .bubble-body pre code { background: transparent; color: inherit; padding: 0; }
  .bubble-body table {
    border-collapse: collapse;
    width: 100%;
    margin: 0.5em 0;
    font-size: 12px;
  }
  .bubble-body th, .bubble-body td {
    border: 1px solid #d5dde6;
    padding: 6px 8px;
    text-align: left;
  }
  .bubble-body th { background: #eef2f6; }
  .bubble-body blockquote {
    margin: 0.5em 0;
    padding: 6px 12px;
    border-left: 3px solid #c5d0dc;
    color: #4a5560;
    background: #f7f9fb;
  }
  .empty { color: #8899aa; font-style: italic; }
  .export-footer {
    margin-top: 18px;
    padding-top: 10px;
    border-top: 1px solid #e3e8ef;
    font-size: 10px;
    color: #99a3ad;
    text-align: center;
  }
`

function exportMarkdown() {
  const lines = ['# Chat Export', '', `> ${new Date().toLocaleString()} · model: ${resolvedModel.value || '—'}`, '', '---', '']
  for (const m of exportableMessages()) {
    const role = m.role === 'user' ? '**User**' : '**Assistant**'
    lines.push(`${role}`)
    lines.push('')
    lines.push(replaceChartFences(stripXmlContent(String(m.content))))
    lines.push('')
    lines.push('---')
    lines.push('')
  }
  const blob = new Blob([lines.join('\n')], { type: 'text/markdown;charset=utf-8' })
  downloadBlob(blob, `chat-export-${exportStamp()}.md`)
}

async function exportPdf() {
  const wrap = document.createElement('div')
  wrap.innerHTML = buildExportDocument()
  const style = document.createElement('style')
  style.textContent = exportPdfStyles
  wrap.prepend(style)

  // Note: html2pdf clones the source into its own hidden overlay container and
  // captures the clone. position:fixed / opacity / z-index on the source leak
  // into the clone and make html2canvas produce an empty canvas → blank PDF.
  // A plain static block in flow (width matching html2canvas windowWidth) works.
  wrap.style.cssText = [
    'width:794px',
    'box-sizing:border-box',
    'background:#fff',
  ].join(';')

  document.body.appendChild(wrap)
  await nextTick()
  try {
    await html2pdf().set({
      margin: [10, 10, 12, 10],
      filename: `chat-export-${exportStamp()}.pdf`,
      image: { type: 'jpeg', quality: 0.98 },
      html2canvas: {
        scale: 2,
        useCORS: true,
        logging: false,
        backgroundColor: '#ffffff',
        windowWidth: 794,
      },
      jsPDF: { unit: 'mm', format: 'a4', orientation: 'portrait' },
      pagebreak: { mode: ['avoid-all', 'css', 'legacy'] },
    }).from(wrap).save()
  } finally {
    document.body.removeChild(wrap)
  }
}

const DOCX_NUMBERING = [
  {
    reference: 'chat-bullets',
    levels: [0, 1, 2, 3, 4].map(level => ({
      level,
      format: NumberFormat.BULLET,
      text: ['•', '◦', '▪', '•', '◦'][level] || '•',
      alignment: AlignmentType.LEFT,
      style: { paragraph: { indent: { left: 720 + level * 360, hanging: 360 } } },
    })),
  },
  {
    reference: 'chat-ordered',
    levels: [0, 1, 2, 3, 4].map(level => ({
      level,
      format: NumberFormat.DECIMAL,
      text: `%${level + 1}.`,
      alignment: AlignmentType.LEFT,
      style: { paragraph: { indent: { left: 720 + level * 360, hanging: 360 } } },
    })),
  },
]

async function exportDocx() {
  const children = [
    new Paragraph({
      children: [new TextRun({ text: 'Chat export', bold: true, size: 44, color: '12263A' })],
      alignment: AlignmentType.CENTER,
      spacing: { after: 80 },
    }),
    new Paragraph({
      children: [
        new TextRun({
          text: `${new Date().toLocaleString()} · Model: ${stripHtml(resolvedModel.value) || '—'}`,
          size: 20,
          color: '667788',
        }),
      ],
      alignment: AlignmentType.CENTER,
      spacing: { after: 320 },
    }),
  ]

  const items = exportableMessages()
  items.forEach((m, i) => {
    const isUser = m.role === 'user'
    const label = isUser ? 'User' : 'Assistant'
    const color = isUser ? '1F4B7A' : '2F9E6B'
    const fill = isUser ? 'F3F7FB' : 'F6FBF8'

    children.push(
      new Paragraph({
        children: [
          new TextRun({ text: `${label}`, bold: true, size: 26, color, allCaps: true }),
          new TextRun({ text: `  #${i + 1}`, size: 18, color: '99A3AD' }),
        ],
        shading: { type: ShadingType.CLEAR, fill, color: 'auto' },
        border: {
          left: { style: BorderStyle.SINGLE, size: 24, color },
          top: { style: BorderStyle.SINGLE, size: 2, color },
          bottom: { style: BorderStyle.SINGLE, size: 2, color },
          right: { style: BorderStyle.SINGLE, size: 2, color },
        },
        spacing: { before: 240, after: 100 },
      }),
    )

    const parts = renderMessageToDocx(String(m.content))
    if (parts.length) {
      children.push(...parts)
    } else {
      children.push(new Paragraph({ children: [new TextRun(' ')], spacing: { after: 60 } }))
    }
  })

  const doc = new Document({
    title: 'Chat Export',
    creator: 'Compose AI Chat',
    styles: {
      default: {
        document: { run: { font: 'Calibri', size: 22 } },
      },
    },
    numbering: {
      config: DOCX_NUMBERING,
    },
    sections: [{
      properties: {
        page: {
          margin: { top: 1080, right: 1080, bottom: 1080, left: 1080 },
        },
      },
      children,
    }],
  })
  const blob = await Packer.toBlob(doc)
  downloadBlob(blob, `chat-export-${exportStamp()}.docx`)
}

function renderMessageToDocx(content) {
  const HTML = markdownToHtml(content)
  const parsed = new DOMParser().parseFromString(HTML, 'text/html')
  const out = []
  walkDocxNodes(parsed.body, out, 0)
  return out
}

function markdownToHtml(content) {
  content = replaceChartFences(stripXmlContent(String(content || '').replace(/\r\n/g, '\n').replace(/\r/g, '\n')))
  const md = markdownIt({ html: true, linkify: true, breaks: true })
  return md.render(content)
}

function walkDocxNodes(node, out, listDepth) {
  if (!node || node.nodeType !== Node.ELEMENT_NODE) return
  for (const el of node.childNodes) {
    if (el.nodeType !== Node.ELEMENT_NODE) continue
    const tag = el.tagName.toLowerCase()
    switch (tag) {
      case 'h1': case 'h2': case 'h3': case 'h4': case 'h5': case 'h6': {
        const level = parseInt(tag[1], 10)
        out.push(new Paragraph({
          children: collectDocxRuns(el),
          heading: HeadingLevel[`HEADING_${Math.min(level, 6)}`],
          spacing: { before: 160, after: 80 },
        }))
        break
      }
      case 'p': {
        const runs = collectDocxRuns(el)
        if (runs.length) out.push(new Paragraph({ children: runs, spacing: { after: 120 } }))
        break
      }
      case 'ul': case 'ol': {
        walkDocxList(el, out, listDepth)
        break
      }
      case 'pre': {
        const code = el.textContent.replace(/\n+$/, '')
        const lines = code.split('\n')
        const runs = []
        lines.forEach((line, i) => {
          if (i > 0) runs.push(new TextRun({ break: 1 }))
          runs.push(new TextRun({ text: line, font: 'Consolas', size: 19, color: '1F2937' }))
        })
        out.push(new Paragraph({
          children: runs,
          shading: { type: ShadingType.CLEAR, fill: 'F3F4F6', color: 'auto' },
          border: { left: { style: BorderStyle.SINGLE, size: 18, color: '9CA3AF' } },
          spacing: { before: 100, after: 140 },
        }))
        break
      }
      case 'blockquote': {
        out.push(new Paragraph({
          children: collectDocxRuns(el),
          indent: { left: 480 },
          border: { left: { style: BorderStyle.SINGLE, size: 18, color: 'ABB2B9' } },
          shading: { type: ShadingType.CLEAR, fill: 'F8F9F9', color: 'auto' },
          spacing: { after: 120 },
        }))
        break
      }
      case 'table': {
        out.push(docxTable(el))
        break
      }
      case 'hr': {
        out.push(new Paragraph({
          children: [],
          border: { bottom: { style: BorderStyle.SINGLE, size: 8, color: 'ABB2B9' } },
          spacing: { before: 100, after: 100 },
        }))
        break
      }
      case 'div': case 'section': case 'article': case 'main': {
        walkDocxNodes(el, out, listDepth)
        break
      }
      default: {
        const runs = collectDocxRuns(el)
        if (runs.length) out.push(new Paragraph({ children: runs, spacing: { after: 120 } }))
      }
    }
  }
}

function walkDocxList(listEl, out, depth) {
  const isOrdered = listEl.tagName.toLowerCase() === 'ol'
  const reference = isOrdered ? 'chat-ordered' : 'chat-bullets'
  for (const li of listEl.children) {
    if (li.tagName.toLowerCase() !== 'li') continue

    const runs = []
    const nested = []
    for (const c of li.childNodes) {
      if (c.nodeType === Node.ELEMENT_NODE && ['UL', 'OL'].includes(c.tagName.toUpperCase())) {
        nested.push(c)
      } else {
        runs.push(...collectDocxRuns(c))
      }
    }

    if (runs.length) {
      out.push(new Paragraph({
        children: runs,
        numbering: { reference, level: depth },
        spacing: { after: 60 },
      }))
    }

    for (const n of nested) {
      walkDocxList(n, out, depth + 1)
    }
  }
}

function docxTable(tableEl) {
  const rows = []
  tableEl.querySelectorAll('tr').forEach((tr) => {
    const cells = []
    tr.querySelectorAll(':scope > th, :scope > td').forEach((cell) => {
      const isHead = cell.tagName.toLowerCase() === 'th'
      const runs = collectDocxRuns(cell)
      cells.push(new TableCell({
        children: [new Paragraph({ children: runs, spacing: { after: 40 } })],
        shading: isHead ? { type: ShadingType.CLEAR, fill: 'D6EAF8', color: 'auto' } : undefined,
        margins: { top: 80, bottom: 80, left: 120, right: 120 },
      }))
    })
    if (cells.length) rows.push(new TableRow({ children: cells }))
  })
  return new Table({
    rows,
    width: { size: 100, type: WidthType.PERCENTAGE },
    borders: {
      top: { style: BorderStyle.SINGLE, size: 4, color: 'ABB2B9' },
      bottom: { style: BorderStyle.SINGLE, size: 4, color: 'ABB2B9' },
      left: { style: BorderStyle.SINGLE, size: 4, color: 'ABB2B9' },
      right: { style: BorderStyle.SINGLE, size: 4, color: 'ABB2B9' },
      insideHorizontal: { style: BorderStyle.SINGLE, size: 4, color: 'D5D8DC' },
      insideVertical: { style: BorderStyle.SINGLE, size: 4, color: 'D5D8DC' },
    },
  })
}

function appendDocxRuns(node, runs, fmt) {
  if (node.nodeType === Node.TEXT_NODE) {
    const text = node.nodeValue.replace(/\u00a0/g, ' ')
    if (text.trim()) runs.push(new TextRun({ text, ...fmt }))
    return
  }
  if (node.nodeType !== Node.ELEMENT_NODE) return
  const tag = node.tagName.toLowerCase()
  switch (tag) {
    case 'br':
      runs.push(new TextRun({ break: 1 }))
      break
    case 'strong': case 'b':
    case 'h1': case 'h2': case 'h3': case 'h4': case 'h5': case 'h6':
      for (const c of node.childNodes) appendDocxRuns(c, runs, { ...fmt, bold: true })
      break
    case 'em': case 'i':
      for (const c of node.childNodes) appendDocxRuns(c, runs, { ...fmt, italics: true })
      break
    case 'u': case 'ins':
      for (const c of node.childNodes) appendDocxRuns(c, runs, { ...fmt, underline: {} })
      break
    case 'del': case 's': case 'strike':
      for (const c of node.childNodes) appendDocxRuns(c, runs, { ...fmt, strike: true })
      break
    case 'code':
      for (const c of node.childNodes) appendDocxRuns(c, runs, { ...fmt, font: 'Consolas', size: 19, color: 'C0392B' })
      break
    case 'a': {
      const href = node.getAttribute('href') || ''
      const linkRuns = []
      for (const c of node.childNodes) appendDocxRuns(c, linkRuns, { ...fmt, color: '2471A3', underline: {} })
      if (href && linkRuns.length) {
        runs.push(new ExternalHyperlink({ link: href, children: linkRuns }))
      } else {
        runs.push(...linkRuns)
      }
      break
    }
    case 'sub':
      for (const c of node.childNodes) appendDocxRuns(c, runs, { ...fmt, subScript: true })
      break
    case 'sup':
      for (const c of node.childNodes) appendDocxRuns(c, runs, { ...fmt, superScript: true })
      break
    default:
      for (const c of node.childNodes) appendDocxRuns(c, runs, { ...fmt })
  }
}

function collectDocxRuns(node) {
  const runs = []
  appendDocxRuns(node, runs, {})
  return runs
}

async function sendMessage(overrideText, opts = {}) {
  const displayText = String(overrideText != null ? overrideText : inputText.value).trim()
  if (!displayText || loading.value) return
  const apiPrompt = String(opts.apiPrompt || displayText).trim()

  const msgIdxAsk = messages.value.length
  messages.value.push({ role: 'user', content: displayText, active: false, collapsed: false })
  if (!opts.keepInput) inputText.value = ''
  stickToBottom.value = true
  scrollToBottom()
  autoGrow()

  const msgIdxAnswer = messages.value.length
  messages.value.push({
    role: 'assistant',
    content: '',
    reasoning: '',
    reasoningOpen: true,
    active: true,
    collapsed: false,
    usedTools: false,
  })
  loading.value = true
  warmingUp.value = false
  toolsActive.value = false
  streamStatus.value = 'thinking'
  emitToolsState()
  scrollToBottom()

  abortController.value = new AbortController()

  try {
    const history = messages.value.slice(0, -2).filter(m => m.content).slice(-8).map(m => ({
      role: m.role,
      content: m.content,
    }))

    await $ComposeAPI.pageAiPromptStream({
      prompt: apiPrompt,
      messages: history,
      files: attachedFiles.value,
      chatID: chatID.value,
      namespaceID: props.namespace,
      pageID: props.page,
      moduleID: props.module,
      model: resolvedModel.value,
      signal: abortController.value.signal,
    }, ({ token, reason, status }) => {
      if (status === 'warming') {
        warmingUp.value = true
        streamStatus.value = 'warming'
        scrollToBottom()
        return
      }
      if (status === 'ready') {
        warmingUp.value = false
        streamStatus.value = 'thinking'
        scrollToBottom()
        return
      }
      if (status === 'tools-enabled') {
        sessionTools.value = true
        emitToolsState()
        return
      }
      if (status === 'tools-disabled') {
        sessionTools.value = false
        emitToolsState()
        return
      }
      if (status === 'using-tools') {
        toolsActive.value = true
        messages.value[msgIdxAnswer].usedTools = true
        streamStatus.value = 'using-tools'
        emitToolsState()
        scrollToBottom()
        return
      }
      if (token) {
        streamStatus.value = 'writing'
        messages.value[msgIdxAnswer].content += token
        const stub = 'Модель не сгенерировала ответ.'
        const reasoningText = messages.value[msgIdxAnswer].reasoning
        if (messages.value[msgIdxAnswer].content.trim() === stub && reasoningText) {
          messages.value[msgIdxAnswer].content = reasoningText
        }
        if (reasoningText) {
          messages.value[msgIdxAnswer].reasoningOpen = false
        }
      }
      if (reason) {
        messages.value[msgIdxAnswer].reasoning += reason
      }
      scrollToBottom()
    })
  } catch (e) {
    const aborted = e?.name === 'AbortError' || /abort/i.test(String(e?.message || ''))
    if (aborted) {
      if (!messages.value[msgIdxAnswer].content && !messages.value[msgIdxAnswer].reasoning) {
        messages.value[msgIdxAnswer].content = $t('aiChat.stopped')
      }
    } else {
      const msg = e.message || 'Error'
      if (messages.value[msgIdxAnswer].content) {
        messages.value[msgIdxAnswer].content += '\n\n' + msg
      } else if (messages.value[msgIdxAnswer].reasoning) {
        messages.value[msgIdxAnswer].content = messages.value[msgIdxAnswer].reasoning
      } else {
        messages.value[msgIdxAnswer].content = msg
      }
    }
  } finally {
    loading.value = false
    warmingUp.value = false
    toolsActive.value = false
    streamStatus.value = ''
    messages.value[msgIdxAnswer].active = false
    messages.value[msgIdxAsk].active = false
    abortController.value = null
    emitToolsState()
    schedulePersist()
    scrollToBottom()
    const queued = pendingAutoSend
    pendingAutoSend = null
    if (queued) {
      sendMessage(queued)
    }
  }
}

function handleDocumentClick(e) {
  closeExport(e)
}

watch(() => props.files, (files) => {
  attachedFiles.value = [...(files || [])]
})

watch(() => props.startPrompt, (prompt) => {
  if (!prompt || !String(prompt).trim()) return
  inputText.value = String(prompt).trim()
  focusInput()
})

watch(messages, schedulePersist, { deep: true })

watch(resolvedModel, () => {
  sessionTools.value = null
  emitToolsState()
})

watch(() => props.preferredModel, () => {
  if (!props.showModelSwitcher || !modelOptions.value.length) return
  if (readStoredModel(props.modelStorageKey)) return
  applyCatalogSelection(modelOptions.value, defaultModel.value)
})

watch(() => props.active, (active) => {
  if (active) focusInput()
})

onMounted(() => {
  document.addEventListener('click', handleDocumentClick)
  loadModelTools()
  restoreSession()
  nextTick(() => {
    const incoming = String(props.startPrompt || '').trim()
    if (props.framed && incoming && !messages.value.some(m => m.role === 'user')) {
      inputText.value = incoming
      sendMessage()
    } else if (incoming) {
      inputText.value = incoming
    }
    focusInput()
    stickToBottom.value = true
    scrollToBottom()
  })
})

onBeforeUnmount(() => {
  if (abortController.value) {
    abortController.value.abort()
    abortController.value = null
  }
  clearTimeout(persistTimer)
  clearTimeout(copiedTimer)
  document.removeEventListener('click', handleDocumentClick)
})

function applyIncomingPrompt (prompt, files) {
  if (Array.isArray(files)) attachedFiles.value = [...files]
  const text = String(prompt || '').trim()
  if (!text) {
    focusInput()
    return
  }
  if (loading.value) {
    pendingAutoSend = text
    stopGeneration()
    return
  }
  sendMessage(text)
}

defineExpose({
  exportMarkdown,
  exportPdf,
  exportDocx,
  newChat,
  focusInput,
  applyIncomingPrompt,
  toolsEnabled,
  toolsActive,
  toolsTitle,
})
</script>

<style scoped>
.chat-container {
  display: flex;
  flex-direction: column;
  min-height: 300px;
  height: 100%;
  width: 100%;
  margin: 0 auto;
  border: 1px solid #e0e0e0;
  border-radius: 12px;
  overflow: hidden;
  background: #f9f9f9;
  flex: 1;
}

.chat-container.frameless {
  border: none;
  border-radius: 0;
}

.chat-meta-bar {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  flex-shrink: 0;
  padding: 6px 12px 0;
  min-width: 0;
}

.chat-model-select {
  width: auto;
  max-width: 180px;
  flex-shrink: 0;
  margin-right: auto;
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

.msg-tools-flag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
  font-size: 12px;
  color: #1f7a4d;
}

.msg-tools-flag.active {
  color: #1f4b7a;
  animation: chat-tools-pulse 1.2s ease-in-out infinite;
}

.export-dropdown {
  position: relative;
}

.export-menu {
  position: absolute;
  right: 0;
  top: 100%;
  margin-top: 4px;
  min-width: 160px;
  background: #fff;
  border: 1px solid #ddd;
  border-radius: 6px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
  z-index: 100;
  overflow: hidden;
}

.export-menu-item {
  display: block;
  width: 100%;
  border: none;
  background: transparent;
  padding: 8px 14px;
  font-size: 13px;
  text-align: left;
  cursor: pointer;
  color: #333;
  line-height: 1.4;
}

.export-menu-item:hover {
  background: #f0f0f0;
}

.messages {
  flex: 1;
  min-height: 0;
  padding: 20px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
  position: relative;
}

.message {
  display: flex;
  gap: 12px;
  max-width: 80%;
  animation: fadeIn 0.3s ease;
}

.message.user {
  align-self: flex-end;
  flex-direction: row-reverse;
}

.message.assistant {
  align-self: flex-start;
}

.avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  flex-shrink: 0;
}

.message-body {
  min-width: 0;
  flex: 1;
}

.collapse-btn {
  border: none;
  background: transparent;
  cursor: pointer;
  color: #999;
  padding: 2px 6px;
  border-radius: 4px;
  line-height: 1;
  display: block;
  margin-left: auto;
  margin-top: 4px;
}

.collapse-btn:hover {
  background: rgba(0,0,0,0.06);
  color: #666;
}

.collapsed-content {
  cursor: pointer;
  opacity: 0.7;
}

.collapsed-content:hover {
  opacity: 1;
}

.content {
  padding: 12px 16px;
  background: var(--body-bg);
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  line-height: 1.5;
  word-wrap: break-word;
}

.content-text {
  display: block;
}

.chat-md :deep(p) {
  margin: 0 0 0.55em;
}

.chat-md :deep(p:last-child) {
  margin-bottom: 0;
}

.chat-md :deep(ul),
.chat-md :deep(ol) {
  margin: 0.4em 0;
  padding-left: 1.3rem;
}

.chat-md :deep(li) {
  margin: 0.15em 0;
}

.chat-md :deep(pre) {
  background: rgba(0, 0, 0, 0.06);
  border-radius: 8px;
  padding: 8px 10px;
  overflow-x: auto;
  font-size: 0.85em;
  margin: 0.5em 0;
}

.chat-md :deep(code) {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 0.9em;
}

.chat-md :deep(pre code) {
  background: none;
  padding: 0;
}

.chat-md :deep(:not(pre) > code) {
  background: rgba(0, 0, 0, 0.06);
  padding: 0.1em 0.35em;
  border-radius: 4px;
}

.chat-md :deep(blockquote) {
  margin: 0.4em 0;
  padding-left: 0.8em;
  border-left: 3px solid rgba(0, 0, 0, 0.15);
  opacity: 0.92;
}

.chat-md :deep(h1),
.chat-md :deep(h2),
.chat-md :deep(h3),
.chat-md :deep(h4) {
  font-size: 1.05em;
  font-weight: 650;
  margin: 0.6em 0 0.35em;
}

.chat-md :deep(h1:first-child),
.chat-md :deep(h2:first-child),
.chat-md :deep(h3:first-child) {
  margin-top: 0;
}

.chat-md :deep(table) {
  border-collapse: collapse;
  font-size: 0.9em;
  margin: 0.5em 0;
}

.chat-md :deep(th),
.chat-md :deep(td) {
  border: 1px solid rgba(0, 0, 0, 0.12);
  padding: 4px 8px;
}

.chat-md :deep(a) {
  color: inherit;
  text-decoration: underline;
}

.prompt-xml {
  margin-bottom: 8px;
}

.prompt-xml-toggle {
  border: none;
  background: transparent;
  color: #8899aa;
  font-size: 12px;
  padding: 0 2px 4px;
  cursor: pointer;
}

.prompt-xml-body {
  font-size: 11px;
  line-height: 1.45;
  color: #5a6570;
  white-space: pre-wrap;
  word-break: break-word;
  background: #f3f5f8;
  border-radius: 8px;
  padding: 8px 10px;
  margin: 0 0 6px;
  max-height: 220px;
  overflow: auto;
}

.message.user .prompt-xml-toggle {
  color: rgba(255, 255, 255, 0.75);
}

.message.user .prompt-xml-body {
  background: rgba(255, 255, 255, 0.14);
  color: rgba(255, 255, 255, 0.92);
}

.message.user .chat-md :deep(pre),
.message.user .chat-md :deep(:not(pre) > code) {
  background: rgba(255, 255, 255, 0.16);
}

.message.user .chat-md :deep(th),
.message.user .chat-md :deep(td) {
  border-color: rgba(255, 255, 255, 0.28);
}

.chat-chart {
  position: relative;
  height: 320px;
  min-height: 280px;
  width: 100%;
  margin: 8px 0;
}

.chat-chart-error {
  font-size: 12px;
  color: #8899aa;
  font-style: italic;
  margin: 6px 0;
}

.chat-chart-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #8899aa;
}

.message.user .content .collapse-btn {
  color: rgba(255,255,255,0.7);
}

.message.user .content .collapse-btn:hover {
  color: rgba(255,255,255,0.9);
  background: rgba(255,255,255,0.12);
}

.content h1 {
  font-size: 1.35rem;
}
.content h2 {
  font-size: 1.25rem;
}

.message.user .content {
  background: var(--primary);
  color: white;
}

.typing-indicator {
  display: flex;
  gap: 4px;
  padding: 8px 0;
}

.warmup-indicator {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 8px 0;
  color: #6c757d;
  font-size: 0.9rem;
}

.typing-indicator span {
  width: 8px;
  height: 8px;
  background: #999;
  border-radius: 50%;
  animation: bounce 1.4s infinite;
}

.typing-indicator span:nth-child(2) { animation-delay: 0.2s; }
.typing-indicator span:nth-child(3) { animation-delay: 0.4s; }

@keyframes bounce {
  0%, 60%, 100% { transform: translateY(0); }
  30% { transform: translateY(-10px); }
}

.file-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 6px 0;
}

.file-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--extra-light);
  border: 1px solid #ddd;
  border-radius: 6px;
  padding: 4px 10px;
  font-size: 13px;
}

.file-chip-name {
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-chip-download {
  border: none;
  background: transparent;
  cursor: pointer;
  color: #999;
  padding: 0;
  width: 24px;
  height: 24px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.file-chip-download:hover {
  color: #666;
}

.input-area {
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  gap: 8px;
  padding: 12px 16px 10px;
  background: white;
  border-top: 1px solid #e0e0e0;
}

.composer {
  display: flex;
  align-items: flex-end;
  gap: 8px;
}

.composer textarea {
  flex: 1;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 8px;
  resize: none;
  font-family: inherit;
  font-size: 14px;
  max-height: 160px;
  line-height: 1.4;
}

.composer-icon {
  width: 36px;
  height: 36px;
  border: none;
  background: transparent;
  color: #8899aa;
  cursor: pointer;
  border-radius: 8px;
  flex-shrink: 0;
}

.composer-icon:hover {
  background: #f0f0f0;
  color: #445;
}

.composer-send {
  padding: 0 16px;
  min-height: 36px;
  background: var(--primary);
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  flex-shrink: 0;
}

.composer-send.stop {
  background: #6c757d;
}

.composer-send:hover:not(:disabled) {
  filter: brightness(0.95);
}

.composer-send:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.composer-hint {
  font-size: 11px;
  color: #99a3ad;
  min-height: 16px;
}

.empty-state {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 24px 16px;
  gap: 16px;
}

.empty-greeting {
  font-size: 1.05rem;
  color: #445;
  text-align: center;
}

.suggestion-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: center;
}

.suggestion-chip {
  border: 1px solid #d5dde6;
  background: #fff;
  border-radius: 999px;
  padding: 6px 12px;
  font-size: 13px;
  color: #334;
  cursor: pointer;
}

.suggestion-chip:hover {
  background: #f3f7fb;
  border-color: #c5d0dc;
}

.reasoning {
  margin-bottom: 6px;
}

.reasoning-toggle {
  border: none;
  background: transparent;
  color: #8899aa;
  font-size: 12px;
  padding: 0 2px 4px;
  cursor: pointer;
}

.reasoning-body {
  font-size: 12px;
  color: #667788;
  white-space: pre-wrap;
  background: #f3f5f8;
  border-radius: 8px;
  padding: 8px 10px;
  margin-bottom: 6px;
  max-height: 160px;
  overflow: auto;
}

.msg-actions {
  display: flex;
  gap: 2px;
  opacity: 0;
  margin-top: 2px;
}

.message:hover .msg-actions,
.message:focus-within .msg-actions {
  opacity: 1;
}

.msg-action {
  border: none;
  background: transparent;
  color: #99a3ad;
  width: 24px;
  height: 24px;
  border-radius: 4px;
  cursor: pointer;
}

.msg-action:hover {
  background: rgba(0,0,0,0.06);
  color: #445;
}

.confirm-card {
  border: 1px solid #d6eaf8;
  background: #f3f7fb;
  border-radius: 8px;
  padding: 10px 12px;
  margin: 8px 0 4px;
}

.confirm-title {
  font-weight: 600;
  font-size: 13px;
  margin-bottom: 6px;
}

.confirm-tools {
  margin: 0 0 8px 1.1em;
  padding: 0;
  font-size: 13px;
}

.confirm-actions {
  display: flex;
  gap: 8px;
}

.typing-indicator {
  min-width: 100px;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
