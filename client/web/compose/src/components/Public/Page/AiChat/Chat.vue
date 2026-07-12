<template>
  <div class="chat-container">
    <div class="chat-header">
      <span class="chat-header-title">LowCoooode AI-assistant</span>
      <div class="export-dropdown">
        <button class="export-trigger" @click="exportOpen = !exportOpen" title="Export">
          <font-awesome-icon :icon="['fas', 'download']" size="xs" /> Export
        </button>
        <div v-if="exportOpen" class="export-menu" @click="exportOpen = false">
          <button class="export-menu-item" @click="exportMarkdown">Markdown (.md)</button>
          <button class="export-menu-item" @click="exportPdf">PDF</button>
          <button class="export-menu-item" @click="exportDocx">DOCX</button>
        </div>
      </div>
    </div>
    <div
      ref="messagesContainer"
      class="messages"
    >
      <div
        v-for="(msg, idx) in messages"
        :key="idx"
        :class="['message', msg.role]"
      >
        <template v-if="msg.content">
          <div class="avatar">
            <font-awesome-icon :icon="['fas', msg.role === 'user' ? 'user' : 'brain']" />
          </div>
          <div class="message-body">
            <div
              v-show="!msg.collapsed"
              class="content expanded-content"
            >
              <div
                class="content-text"
                v-html="formatMessage(msg.content)"
              />
              <button
                v-if="hasManyLines(msg.content) && msg.active !== true"
                class="collapse-btn"
                @click="toggleCollapse(idx)"
                title="Collapse"
              >
                <font-awesome-icon :icon="['fas', 'chevron-up']" size="xs" />
              </button>
            </div>
            <div
              v-show="msg.collapsed"
              class="content collapsed-content"
              @click="toggleCollapse(idx)"
            >
              <div
                class="content-text"
                v-html="formatMessage(preview(msg.content))"
              />
              <button
                v-if="hasManyLines(msg.content) && msg.active !== true"
                class="collapse-btn"
                @click.stop="toggleCollapse(idx)"
                title="Expand"
              >
                <font-awesome-icon :icon="['fas', 'chevron-down']" size="xs" />
              </button>
            </div>
          </div>
        </template>
      </div>
      <div
        v-if="loading"
        class="message assistant"
      >
        <div class="avatar">
          <font-awesome-icon :icon="['fas', 'brain']" />
        </div>
        <div class="content typing-indicator">
          <span /><span /><span />
        </div>
      </div>
    </div>

    <div class="input-area">
      <div
        v-if="files.length"
        class="file-chips"
      >
        <div
          v-for="(f, i) in files"
          :key="i"
          class="file-chip"
        >
          <span class="file-chip-name">{{ f.name }}</span>
          <button
            class="file-chip-download"
            @click="downloadFile(f)"
            :title="$t('general:label.download')"
          >
            <font-awesome-icon :icon="['fas', 'download']" size="xs" />
          </button>
        </div>
      </div>
      <textarea
        v-model="inputText"
        :placeholder="$t('aiChat.sendMessage.placeholder')"
        rows="2"
        @keydown.enter.prevent="sendMessage"
      />
      <button
        :disabled="!inputText.trim() || loading"
        @click="sendMessage"
      >
        {{ $t('aiChat.sendMessage.button') }}
      </button>
    </div>
  </div>
</template>

<script>
import { nextTick } from 'vue'
import markdownIt from 'markdown-it'
import html2pdf from 'html2pdf.js'
import { Document, Packer, Paragraph, TextRun } from 'docx'

export default {
  i18nOptions: {
    namespaces: 'page',
  },

  props: {
    startPrompt: {
      type: String,
      required: true,
    },
    page: {
      type: String,
      required: false,
      default: '',
    },
    module: {
      type: String,
      required: false,
      default: '',
    },
    namespace: {
      type: String,
      required: false,
      default: '',
    },
    files: {
      type: Array,
      required: false,
      default: () => [],
    },
  },

  data () {
    return {
      messages: [
        {
          role: 'assistant',
          content: this.$t('aiChat.greeting'),
          collapsed: true,
        },
      ],
      inputText: '',
      loading: false,
      messagesContainer: null,
      chatID: `chat-${Date.now()}-${Math.random().toString(36).substring(2, 10)}`,
      exportOpen: false,
    }
  },
  mounted () {
    document.addEventListener('click', this.closeExport)
    this.$nextTick(function () {
      if (!this.startPrompt) {
        return
      }
      this.inputText = this.startPrompt.trim()
      this.sendMessage()
    })
  },
  beforeDestroy () {
    document.removeEventListener('click', this.closeExport)
  },
  methods: {
    async sendMessage () {
      const text = this.inputText.trim()
      if (!text || this.loading) return

      const msgIdxAsk = this.messages.length
      this.messages.push({ role: 'user', content: text, active: false, collapsed: false })
      this.inputText = ''
      this.scrollToBottom()

      const msgIdxReasoning = this.messages.length
      this.messages.push({ role: 'assistant', reason: true, active: true, collapsed: false, content: '' })
      const msgIdxAnswer = this.messages.length
      this.messages.push({ role: 'assistant', reason: false, active: true, collapsed: false, content: '' })
      this.loading = true
      this.scrollToBottom()

      try {
        // send completed exchanges as context (skip greeting and current exchange)
        const history = this.messages.slice(1, -2).map(m => ({
          role: m.role,
          content: m.content,
        }))

        await this.$ComposeAPI.pageAiPromptStream({
          prompt: text,
          messages: history,
          files: this.files,
          chatID: this.chatID,
          namespaceID: this.namespace,
          pageID: this.page,
          moduleID: this.module,
        }, ({ token, reason }) => {
          this.messages[msgIdxAnswer].content += token
          if (token !== '') {
            this.messages[msgIdxReasoning].active = false
            this.messages[msgIdxReasoning].collapsed = true
          }
          if (reason !== '') {
            if (this.messages[msgIdxReasoning].content === '') {
              this.messages[msgIdxReasoning].content = '<b>Размышляю...</b> '
            }
            this.messages[msgIdxReasoning].content += reason.replace(/\n/g, '<br>')
          }
          this.scrollToBottom()
        })
      } catch (e) {
        this.messages[msgIdxAnswer].content = e.message || 'Error'
      } finally {
        this.loading = false
        this.messages[msgIdxAnswer].active = false
        this.messages[msgIdxAsk].active = false
        this.messages[msgIdxReasoning].active = false

        this.scrollToBottom()
      }
    },
    hasManyLines (text) {
      if (!text) return false
      const breaks = (text.match(/\n/g) || []).length + (text.match(/<br\s*\/?>/gi) || []).length
      return breaks >= 4
    },
    toggleCollapse (idx) {
      this.messages[idx].collapsed = !this.messages[idx].collapsed
    },
    downloadFile (f) {
      const blob = new Blob([f.content], { type: f.type || 'text/csv' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = f.name
      a.click()
      URL.revokeObjectURL(url)
    },
    preview (text) {
      const plain = text.replace(/<[^>]*>/g, '')
      return plain.length > 120 ? plain.slice(0, 120) + '…' : plain
    },
    scrollToBottom: () => {
      if (this === null || this === undefined) {
        return
      }
      const container = this.messagesContainer
      nextTick(() => {
        if (container !== null) {
          container.scrollTop = container.scrollHeight
        }
      })
    },

    formatMessage: (text) => {
      const needMarkdown = (text) => {
        const symbolRegex = /[`*#]/
        return symbolRegex.test(text)
      }
      if (needMarkdown(text)) {
        const md = markdownIt({
          html: true, // Enable HTML tags in source
          linkify: true, // Autoconvert URL-like text to links
        })

        return md.render(text)
      }
      return text.replace(/\n/g, '<br>')
    },

    closeExport (e) {
      if (!this.$el.contains(e.target)) return
      if (!e.target.closest('.export-dropdown')) {
        this.exportOpen = false
      }
    },

    stripHtml (text) {
      const div = document.createElement('div')
      div.innerHTML = text
      return div.textContent || div.innerText || ''
    },

    downloadBlob (blob, filename) {
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = filename
      a.click()
      URL.revokeObjectURL(url)
    },

    exportMarkdown () {
      const lines = ['# Chat Export', '', '---', '']
      for (const m of this.messages) {
        if (!m.content) continue
        const role = m.role === 'user' ? '**User**' : '**Assistant**'
        const plain = this.stripHtml(m.content)
        lines.push(`${role}:`)
        lines.push('')
        lines.push(plain)
        lines.push('')
        lines.push('---')
        lines.push('')
      }
      const blob = new Blob([lines.join('\n')], { type: 'text/markdown;charset=utf-8' })
      this.downloadBlob(blob, 'chat-export.md')
    },

    async exportPdf () {
      const el = this.$refs.messagesContainer.cloneNode(true)
      el.querySelectorAll('.expanded-content').forEach(e => { e.style.display = 'block' })
      el.querySelectorAll('.collapsed-content').forEach(e => { e.style.display = 'none' })
      el.classList.remove('messages')
      el.style.position = 'absolute'
      el.style.left = '-9999px'
      el.style.width = '800px'
      el.style.background = '#fff'
      el.style.padding = '20px'
      document.body.appendChild(el)
      try {
        await html2pdf().set({ margin: 10 }).from(el).save('chat-export.pdf')
      } finally {
        document.body.removeChild(el)
      }
    },

    async exportDocx () {
      const children = []
      for (const m of this.messages) {
        if (!m.content) continue
        const plain = this.stripHtml(m.content)
        const label = m.role === 'user' ? 'User' : 'Assistant'
        children.push(
          new Paragraph({
            children: [
              new TextRun({ text: label + ': ', bold: true, size: 24 }),
              new TextRun({ text: plain, size: 22 }),
            ],
            spacing: { after: 200 },
          }),
        )
      }
      const doc = new Document({
        title: 'Chat Export',
        sections: [{ children }],
      })
      const blob = await Packer.toBlob(doc)
      this.downloadBlob(blob, 'chat-export.docx')
    },
  },
}

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
}

.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  background: #fff;
  border-bottom: 1px solid #e0e0e0;
}

.chat-header-title {
  font-size: 14px;
  font-weight: 600;
  color: #555;
}

.export-dropdown {
  position: relative;
}

.export-trigger {
  border: 1px solid #ddd;
  background: #fff;
  border-radius: 4px;
  padding: 3px 10px;
  font-size: 12px;
  cursor: pointer;
  color: #666;
  line-height: 1.6;
  transition: background 0.15s, border-color 0.15s;
  display: flex;
  align-items: center;
  gap: 4px;
}

.export-trigger:hover {
  background: #f5f5f5;
  border-color: #bbb;
  color: #333;
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
  padding: 20px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
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

.message.assistant .avatar {
}

.message-body {
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
  display: inline;
}

.message.user .content .collapse-btn {
  color: rgba(255,255,255,0.7);
}

.message.user .content .collapse-btn:hover {
  color: rgba(255,255,255,0.9);
  background: rgba(255,255,255,0.12);
}

h1 {
  color: red;
  font-size: 1.5rem;
}
.content h2 {
  font-size: 1.25rem;
}

.message.user .content {
  background: var(--primary);
  color: white;
}

/* Индикатор печати */
.typing-indicator {
  display: flex;
  gap: 4px;
  padding: 8px 0;
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

/* Ввод */
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
  gap: 12px;
  padding: 16px;
  background: white;
  border-top: 1px solid #e0e0e0;
}

.input-area textarea {
  flex: 1;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 8px;
  resize: none;
  font-family: inherit;
  font-size: 14px;
}

.input-area button {
  padding: 0 24px;
  background: var(--primary);
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  transition: background 0.2s;
}
.typing-indicator {
  min-width: 100px;
}

.input-area button:hover:not(:disabled) {
  background: var(--primary);
}

.input-area button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
