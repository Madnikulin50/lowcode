// The assistant of the «Стройконтроль» agent for any document that is open in Compose.
//
// Components that show a document (the attachment lightbox, the PDF view of Docs, the file links
// of record lists) register it here; the launcher (AssistantLauncher.vue) shows a floating button
// while at least one document is registered and opens the agent's document viewer with the
// assistant on the most recently registered one.
import { reactive, computed } from 'vue'

const state = reactive({
  stack: [], // [{ key, source, id, namespaceID, name }], the last one is the document on screen
  open: false,
  pinned: null, // a document the person asked about explicitly (from a file link)
})

const AGENT_PATH = '/gateway/agents/stroykontrol/'

export function assistantURL (doc, page) {
  const base = (window.CortezaAPI || '').replace(/\/$/, '')
  const q = new URLSearchParams({ view: 'doc', source: doc.source, id: doc.id, namespaceID: doc.namespaceID })
  if (page) q.set('page', String(page))
  return `${base}${AGENT_PATH}?${q.toString()}`
}

// an attachment URL looks like …/namespace/<ns>/attachment/<kind>/<id>/original/<name>
export function parseAttachmentURL (url) {
  const m = String(url || '').match(/\/namespace\/(\d+)\/attachment\/[^/]+\/(\d+)\//)
  return m ? { namespaceID: m[1], id: m[2] } : null
}

export function registerAssistantDocument (doc) {
  if (!doc || !doc.id || !doc.namespaceID) return () => {}
  const key = `${doc.source}:${doc.id}:${Math.random().toString(36).slice(2)}`
  state.stack.push({ ...doc, key })
  return () => {
    const i = state.stack.findIndex(d => d.key === key)
    if (i >= 0) state.stack.splice(i, 1)
  }
}

// opens the assistant on a document right away (the ✦ next to a file name)
export function openAssistant (doc) {
  if (!doc || !doc.id || !doc.namespaceID) return
  state.pinned = { ...doc }
  state.open = true
}

export function closeAssistant () {
  state.open = false
  state.pinned = null
}

export function useAssistantContext () {
  const current = computed(() => state.pinned || state.stack[state.stack.length - 1] || null)
  return { state, current }
}
