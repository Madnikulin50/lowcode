// Bootstraps the comparison view: fetches /api/comparison (the documents of
// the comparison, discrepancies with their positions, remark cards), loads
// page info of whichever documents are selected, and polls while the agent
// is still comparing. The comparison may hold a whole package of files per
// stage; pages are loaded per document, on demand.
import { reactive, onMounted, onUnmounted } from 'vue'
import { renderDocxPages } from './useDocxRender.js'
import { renderDxfPages } from './useDxfRender.js'
import { renderXlsxPages, alignXlsxSides } from './useXlsxRender.js'

// Kinds the browser renders itself from the raw /api/file bytes.
const clientRenderers = {
  docx: renderDocxPages,
  dxf: renderDxfPages,
  xlsx: renderXlsxPages,
}

export const emptyPages = () => ({ supported: false, kind: 'other', pageCount: 0, images: null, texts: null })

async function getJSON (url) {
  const res = await fetch(url)
  const data = await res.json().catch(() => ({}))
  if (!res.ok || data.error) throw new Error(data.error || res.statusText)
  return data
}

async function postJSON (url, body) {
  const res = await fetch(url, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
  const data = await res.json().catch(() => ({}))
  if (!res.ok || data.error) throw new Error(data.error || res.statusText)
  return data
}

export const docSel = (id) => `doc=${encodeURIComponent(id)}`

// Resolves one page's image source regardless of whether it came from the
// server (PDF, fetched on demand) or was rendered client-side (DOCX/DXF/XLSX,
// already a data: URL in memory).
export function pageSrc (qs, info, docID, n) {
  if (!info || !docID) return ''
  if (info.kind === 'pdf') return `api/page?${qs}&${docSel(docID)}&n=${n}`
  if (clientRenderers[info.kind] && info.images) return info.images[n - 1] || ''
  return ''
}

// Resolves one page's extracted text, regardless of source (server-side
// pdftotext for PDF, docx-preview's own DOM for DOCX, the DXF text pass for
// DXF).
export function pageText (info, n) {
  if (!info || !info.texts) return ''
  return info.texts[n - 1] || ''
}

export function useComparison () {
  const params = new URLSearchParams(location.search)
  const recordID = params.get('recordID') || ''
  const namespaceID = params.get('namespaceID') || ''
  const qs = `recordID=${encodeURIComponent(recordID)}&namespaceID=${encodeURIComponent(namespaceID)}`

  const state = reactive({
    loading: true,
    error: '',
    comparison: null,
    docs: {}, // documentID → page info (loaded lazily)
    busy: false, // a status/remark request is in flight
    actionError: '',
  })

  let pollTimer = null

  const documents = (side) => (state.comparison?.documents || []).filter(d => d.side === side)

  async function loadDoc (id) {
    if (!id) return emptyPages()
    if (state.docs[id]) return state.docs[id]
    let info
    try {
      info = await getJSON(`api/pages?${qs}&${docSel(id)}`)
    } catch (e) {
      info = emptyPages()
      state.docs[id] = info
      return info
    }
    const render = clientRenderers[info.kind]
    if (render) {
      try {
        const { images, texts, sheets } = await render(qs, docSel(id))
        info.images = images
        info.texts = texts
        info.sheets = sheets
        info.pageCount = images.length
        if (!images.length) { info.supported = false; info.kind = 'other' }
      } catch (e) {
        console.error(info.kind + ' render failed for doc=' + id, e)
        info.supported = false
        info.kind = 'other'
      }
    } else if (info.kind === 'pdf') {
      // Page-image fetching for PDF stays lazy (/api/page, per view), but
      // text is cheap and cached server-side, so just fetch it eagerly here
      // — the "text" view mode needs it and there's no per-page round trip
      // to avoid.
      try {
        const t = await getJSON(`api/text?${qs}&${docSel(id)}`)
        info.texts = t.pages || []
      } catch (e) {
        console.error('pdftotext failed for doc=' + id, e)
        info.texts = []
      }
    }
    state.docs[id] = info
    return info
  }

  // Loads the selected pair; two spreadsheets share column widths so the
  // overlay/diff modes line up cell for cell.
  async function loadPair (pdID, rdID) {
    const [pd, rd] = await Promise.all([loadDoc(pdID), loadDoc(rdID)])
    if (pd.kind === 'xlsx' && rd.kind === 'xlsx') alignXlsxSides(pd, rd)
  }

  async function refresh () {
    state.comparison = await getJSON(`api/comparison?${qs}`)
  }

  // The agent's comparison takes seconds to minutes — while it runs, poll
  // the lightweight /api/comparison endpoint rather than making the user
  // reload by hand. When it finishes, discrepancies/documents have changed,
  // so a full reload picks everything up cleanly.
  function schedulePoll () {
    clearTimeout(pollTimer)
    const c = state.comparison
    if (!c) return
    const active = c.status === 'new' || c.status === 'processing' || !!c.running
    if (!active) return
    pollTimer = setTimeout(async () => {
      let next
      try {
        next = await getJSON(`api/comparison?${qs}`)
      } catch (e) {
        schedulePoll()
        return
      }
      if (next.status !== c.status || (c.running && !next.running)) {
        location.reload()
        return
      }
      state.comparison = next
      schedulePoll()
    }, 4000)
  }

  async function start () {
    if (!recordID || !namespaceID) {
      state.loading = false
      state.error = 'Не переданы recordID/namespaceID.'
      return
    }
    try {
      await refresh()
    } catch (e) {
      state.loading = false
      state.error = 'Не удалось загрузить сравнение: ' + e.message
      return
    }
    state.loading = false
    schedulePoll()
  }

  async function act (fn) {
    state.busy = true
    state.actionError = ''
    try {
      return await fn()
    } catch (e) {
      state.actionError = e.message
      return null
    } finally {
      state.busy = false
    }
  }

  const base = () => ({ recordID, namespaceID })

  // Starts the agent's comparison of the whole package.
  const run = () => act(async () => {
    await postJSON(`api/cmp/run?${qs}`, {})
    state.comparison.running = 'в очереди'
    schedulePoll()
  })

  const setStatus = (id, status) => act(async () => {
    await postJSON('api/discrepancy/status', { ...base(), id, status })
    const d = state.comparison.discrepancies.find(x => x.id === id)
    if (d) d.status = status
  })

  const createRemark = (discrepancyIDs) => act(async () => {
    const card = await postJSON('api/remark', { ...base(), discrepancyIDs })
    state.comparison.remarks.push(card)
    return card
  })

  const saveRemark = (remark) => act(async () => {
    const saved = await postJSON('api/remark/save', { ...base(), remark })
    const i = state.comparison.remarks.findIndex(r => r.id === saved.id)
    if (i >= 0) state.comparison.remarks[i] = saved
    return saved
  })

  const suggestRemark = (id) => act(async () => {
    const saved = await postJSON('api/remark/suggest', { ...base(), id })
    const i = state.comparison.remarks.findIndex(r => r.id === saved.id)
    if (i >= 0) state.comparison.remarks[i] = saved
    return saved
  })

  // Reads the Excel of expertise remarks attached to the comparison and matches
  // the remarks with the findings (agent side), then reloads the list.
  const importExpertise = () => act(async () => {
    const sum = await postJSON(`api/expertise/import?${qs}`, {})
    await refresh()
    return sum
  })

  const setExpertiseStatus = (id, status) => act(async () => {
    await postJSON('api/expertise/status', { ...base(), id, status })
    const e = state.comparison.expertise.find(x => x.id === id)
    if (e) e.status = status
  })

  onMounted(start)
  onUnmounted(() => clearTimeout(pollTimer))

  return { state, qs, recordID, documents, loadPair, run, setStatus, createRemark, saveRemark, suggestRemark, importExpertise, setExpertiseStatus }
}
