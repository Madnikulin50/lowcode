<template>
  <div class="chat-container">
    <div
      ref="messagesContainer"
      class="messages"
    >
      <div
        v-for="(msg, idx) in messages"
        :key="idx"
        :class="['message', msg.role]"
      >
        <div class="avatar">
          {{ msg.role === 'user' ? '👤' : '🧠' }}
        </div>
        <div
          class="content"
          v-html="formatMessage(msg.content)"
        />
      </div>
      <div
        v-if="loading"
        class="message assistant"
      >
        <div class="avatar">
          🧠
        </div>
        <div class="content typing-indicator">
          <span /><span /><span />
        </div>
      </div>
    </div>

    <div class="input-area">
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
  },

  data () {
    return {
      messages: [
        {
          role: 'assistant',
          content: this.$t('aiChat.greeting'),
        },
      ],
      inputText: '',
      loading: false,
      messagesContainer: null,
    }
  },
  mounted () {
    this.$nextTick(function () {
      if (!this.startPrompt) {
        return
      }
      this.inputText = this.startPrompt.trim()
      this.sendMessage()
    })
  },
  methods: {
    async sendMessage () {
      const text = this.inputText.trim()
      if (!text || this.loading) return

      // Добавляем сообщение пользователя
      this.messages.push({ role: 'user', content: text })
      this.inputText = ''
      this.scrollToBottom()

      this.loading = true
      return this.$ComposeAPI
        .pageAiPrompt({
          prompt: text,
          namespaceID: this.namespace,
          pageID: this.page,
          moduleID: this.module,
        })
        .then(set => {
          this.messages.push({ role: 'assistant', content: set.response })
          this.scrollToBottom()
        })
        .finally(() => {
          this.loading = false
          this.scrollToBottom()
        })
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

.content {
  padding: 12px 16px;
  background: var(--body-bg);
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  line-height: 1.5;
  word-wrap: break-word;

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
