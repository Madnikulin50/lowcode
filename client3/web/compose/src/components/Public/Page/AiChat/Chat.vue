<template>
  <div class="chat-container">
    <div ref="messagesContainer" class="messages">
      <div class="export-dropdown position-absolute" style="top: 8px; right: 12px; z-index: 10;">
        <button class="btn btn-outline-light d-print-none text-secondary px-2 py-1 border-0" @click="exportOpen = !exportOpen" title="Export">
          <font-awesome-icon :icon="['fas', 'download']" size="xs" />
        </button>
        <div v-if="exportOpen" class="export-menu" @click="exportOpen = false">
          <button class="export-menu-item" @click="exportMarkdown">Markdown (.md)</button>
          <button class="export-menu-item" @click="exportPdf">PDF</button>
          <button class="export-menu-item" @click="exportDocx">DOCX</button>
        </div>
      </div>
      <div v-for="(msg, idx) in messages" :key="idx" :class="['message', msg.role, { active: msg.active }]">
        <template v-if="msg.content">
          <div class="avatar">
            <font-awesome-icon :icon="['fas', msg.role === 'user' ? 'user' : 'brain']" />
          </div>
          <div class="message-body">
            <div v-show="!msg.collapsed" class="content expanded-content">
              <div class="content-text" v-html="formatMessage(msg.content)" />
              <button
                v-if="hasManyLines(msg.content) && msg.active !== true"
                class="collapse-btn"
                @click="toggleCollapse(idx)"
                title="Collapse"
              >
                <font-awesome-icon :icon="['fas', 'chevron-up']" size="xs" />
              </button>
            </div>
            <div v-show="msg.collapsed" class="content collapsed-content" @click="toggleCollapse(idx)">
              <div class="content-text" v-html="formatMessage(preview(msg.content))" />
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
      <div v-if="loading" class="message assistant">
        <div class="avatar">
          <font-awesome-icon :icon="['fas', 'brain']" />
        </div>
        <div class="content typing-indicator">
          <span /><span /><span />
        </div>
      </div>
    </div>

    <div class="input-area">
      <div v-if="files.length" class="file-chips">
        <div v-for="(f, i) in files" :key="i" class="file-chip">
          <span class="file-chip-name">{{ f.name }}</span>
          <button class="file-chip-download" @click="downloadFile(f)" title="Download">
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
      <button :disabled="!inputText.trim() || loading" @click="sendMessage">
        {{ $t('aiChat.sendMessage.button') }}
      </button>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'page' } })
import { ref, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import markdownIt from 'markdown-it'
import html2pdf from 'html2pdf.js'
import { Document, Packer, Paragraph, TextRun, ExternalHyperlink, HeadingLevel, AlignmentType, NumberFormat, WidthType, BorderStyle, ShadingType, Table, TableRow, TableCell } from 'docx'

const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  startPrompt: { type: String, required: true },
  page: { type: String, required: false, default: '' },
  module: { type: String, required: false, default: '' },
  namespace: { type: String, required: false, default: '' },
  magnified: { type: Boolean, default: false },
  files: { type: Array, required: false, default: () => [] },
  model: { type: String, required: false, default: '' },
})

const $ComposeAPI = window.__composeAPI

const messages = ref([
  { role: 'assistant', content: $t('aiChat.greeting'), collapsed: true },
])
const inputText = ref('')
const loading = ref(false)
const messagesContainer = ref(null)
const chatID = `chat-${Date.now()}-${Math.random().toString(36).substring(2, 10)}`
const exportOpen = ref(false)
const abortController = ref(null)

function hasManyLines(text) {
  if (!text) return false
  const breaks = (text.match(/\n/g) || []).length + (text.match(/<br\s*\/?>/gi) || []).length
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

function preview(text) {
  const hidden = stripXmlContent(text).replace(/\r\n/g, '\n').replace(/\r/g, '\n')
  const noCode = stripCodeBlocks(hidden)
  const plain = noCode.replace(/<[^>]*>/g, '')
  return plain.length > 120 ? plain.slice(0, 120) + '\u2026' : plain
}

function scrollToBottom() {
  const container = messagesContainer.value
  nextTick(() => {
    if (container !== null) {
      container.scrollTop = container.scrollHeight
    }
  })
}

function stripXmlContent(text) {
  let normalized = text.replace(/&lt;/g, '<').replace(/&gt;/g, '>')
  let prev
  do {
    prev = normalized
    normalized = normalized.replace(/<([a-zA-Z_][a-zA-Z0-9_]*)[^>]*>[\s\S]*?<\/\1>/g, '<$1>...</$1>')
  } while (normalized !== prev)
  return normalized
}

function stripCodeBlocks(text) {
  return text.replace(/```[\s\S]*?(```|$)/g, '').replace(/\n{3,}/g, '\n\n')
}

function formatMessage(text) {
  text = text.replace(/\r\n/g, '\n').replace(/\r/g, '\n')
  text = stripCodeBlocks(text)
  const needMarkdown = (text) => {
    const symbolRegex = /[`*#]/
    return symbolRegex.test(text)
  }
  if (needMarkdown(text)) {
    const md = markdownIt({ html: true, linkify: true, breaks: true })
    return md.render(text)
  }
  return text.replace(/\n/g, '<br>')
}

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
  normalized = stripXmlContent(normalized)
  const md = markdownIt({ html: false, linkify: true, breaks: true })
  return md.render(normalized)
}

function buildExportDocument() {
  const items = exportableMessages()
  const when = new Date().toLocaleString()
  const model = props.model || '—'
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
  const lines = ['# Chat Export', '', `> ${new Date().toLocaleString()} · model: ${props.model || '—'}`, '', '---', '']
  for (const m of exportableMessages()) {
    const role = m.role === 'user' ? '**User**' : '**Assistant**'
    lines.push(`${role}`)
    lines.push('')
    lines.push(stripXmlContent(String(m.content)))
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

  // Keep in viewport (offscreen left:-9999px yields empty html2canvas captures).
  wrap.style.cssText = [
    'position:fixed',
    'left:0',
    'top:0',
    'width:794px',
    'z-index:-1',
    'opacity:0.01',
    'pointer-events:none',
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
          text: `${new Date().toLocaleString()} · Model: ${stripHtml(props.model) || '—'}`,
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
  content = String(content || '').replace(/\r\n/g, '\n').replace(/\r/g, '\n')
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

async function sendMessage() {
  const text = inputText.value.trim()
  if (!text || loading.value) return

  const msgIdxAsk = messages.value.length
  messages.value.push({ role: 'user', content: text, active: false, collapsed: false })
  inputText.value = ''
  scrollToBottom()

  const msgIdxReasoning = messages.value.length
  messages.value.push({ role: 'assistant', reason: true, active: true, collapsed: false, content: '' })
  const msgIdxAnswer = messages.value.length
  messages.value.push({ role: 'assistant', reason: false, active: true, collapsed: false, content: '' })
  loading.value = true
  scrollToBottom()

  abortController.value = new AbortController()

  try {
    const history = messages.value.slice(1, -2).slice(-8).map(m => ({
      role: m.role,
      content: m.content,
    }))

    await $ComposeAPI.pageAiPromptStream({
      prompt: text,
      messages: history,
      files: props.files,
      chatID,
      namespaceID: props.namespace,
      pageID: props.page,
      moduleID: props.module,
      model: props.model,
      signal: abortController.value.signal,
    }, ({ token, reason }) => {
      messages.value[msgIdxAnswer].content += token
      if (token !== '') {
        messages.value[msgIdxReasoning].active = false
        messages.value[msgIdxReasoning].collapsed = true
      }
      if (reason !== '') {
        if (messages.value[msgIdxReasoning].content === '') {
          messages.value[msgIdxReasoning].content = '<b>\u0420\u0430\u0437\u043c\u044b\u0448\u043b\u044f\u044e...</b> '
        }
        messages.value[msgIdxReasoning].content += reason.replace(/\n/g, '<br>')
      }
      scrollToBottom()
    })
  } catch (e) {
    if (e?.name !== 'AbortError') {
      messages.value[msgIdxAnswer].content = e.message || 'Error'
    }
  } finally {
    loading.value = false
    messages.value[msgIdxAnswer].active = false
    messages.value[msgIdxAsk].active = false
    messages.value[msgIdxReasoning].active = false
    abortController.value = null
    scrollToBottom()
  }
}

function handleDocumentClick(e) {
  closeExport(e)
}

onMounted(() => {
  document.addEventListener('click', handleDocumentClick)
  nextTick(() => {
    if (!props.startPrompt) return
    inputText.value = props.startPrompt.trim()
    sendMessage()
  })
})

onBeforeUnmount(() => {
  if (abortController.value) {
    abortController.value.abort()
    abortController.value = null
  }
  document.removeEventListener('click', handleDocumentClick)
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
  flex-shrink: 0;
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
