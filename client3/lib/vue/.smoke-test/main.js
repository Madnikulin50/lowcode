import { createApp, h } from 'vue'
import CRichTextInput from '../src/components/input/CRichTextInput/index.vue'

const app = createApp({
  render () {
    return h('div', { style: 'padding:2rem' }, [
      h('div', { id: 'status', style: 'margin-bottom:1rem;font-family:monospace' }, 'mounting...'),
      // Mirrors formats.js: StarterKit + TaskList/TaskItem + TableKit + Placeholder + Emoji + Mention,
      // the exact combination that produced the "plugin$" key collision.
      h(CRichTextInput, { modelValue: '<p>Hello <strong>world</strong></p><table><tr><td>a</td></tr></table>' }),
    ])
  },
  mounted () {
    const el = document.getElementById('status')
    if (el && el.textContent === 'mounting...') el.textContent = 'MOUNTED_OK'
  },
})

app.config.globalProperties.$SystemAPI = {}

window.addEventListener('error', (e) => {
  const el = document.getElementById('status')
  if (el) el.textContent = 'MOUNT_ERROR: ' + (e.error?.message || e.message)
})

app.mount('#app')
