// Bootstraps the comparison view: fetches /api/comparison + both sides'
// pages, then polls while the agent's own analysis chain is still running.
// Ported from the original vanilla-JS main()/schedulePoll()/loadSide() —
// state that used to live in module-level `let`s and trigger a manual
// rerender() now lives in a single `reactive()` object Vue tracks itself.
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

const emptyPages = () => ({ supported: false, kind: 'other', pageCount: 0, images: null, texts: null })

async function getJSON (url) {
  const res = await fetch(url)
  const data = await res.json().catch(() => ({}))
  if (!res.ok || data.error) throw new Error(data.error || res.statusText)
  return data
}

// Resolves one page's image source regardless of whether it came from the
// server (PDF, fetched on demand) or was rendered client-side (DOCX/DXF/XLSX,
// already a data: URL in memory).
export function pageSrc (qs, info, side, n) {
  if (!info) return ''
  if (info.kind === 'pdf') return `api/page?${qs}&side=${side}&n=${n}`
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
    pdPages: emptyPages(),
    rdPages: emptyPages(),
  })

  let pollTimer = null

  async function loadSide (side, hasFile) {
    if (!hasFile) return emptyPages()
    let info
    try {
      info = await getJSON(`api/pages?${qs}&side=${side}`)
    } catch (e) {
      return emptyPages()
    }
    const render = clientRenderers[info.kind]
    if (render) {
      try {
        const { images, texts, sheets } = await render(qs, side)
        info.images = images
        info.texts = texts
        info.sheets = sheets
        info.pageCount = images.length
        if (!images.length) { info.supported = false; info.kind = 'other' }
      } catch (e) {
        console.error(info.kind + ' render failed for side=' + side, e)
        info.supported = false
        info.kind = 'other'
      }
    } else if (info.kind === 'pdf') {
      // Page-image fetching for PDF stays lazy (/api/page, per view), but
      // text is cheap and cached server-side, so just fetch it eagerly here
      // — the "text" view mode needs it and there's no per-page round trip
      // to avoid.
      try {
        const t = await getJSON(`api/text?${qs}&side=${side}`)
        info.texts = t.pages || []
      } catch (e) {
        console.error('pdftotext failed for side=' + side, e)
        info.texts = []
      }
    }
    return info
  }

  // The agent's own analysis chain takes minutes on CPU-only inference —
  // while it's running, poll the lightweight /api/comparison endpoint (no
  // page images) rather than making the user reload by hand. Once the
  // status leaves new/processing, the pages/attachments may have changed
  // too, so a full reload picks those up cleanly instead of reconciling
  // state in place.
  function schedulePoll () {
    clearTimeout(pollTimer)
    if (!state.comparison) return
    if (state.comparison.status !== 'new' && state.comparison.status !== 'processing') return
    pollTimer = setTimeout(async () => {
      let next
      try {
        next = await getJSON(`api/comparison?${qs}`)
      } catch (e) {
        schedulePoll()
        return
      }
      if (next.status !== state.comparison.status) {
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
      state.comparison = await getJSON(`api/comparison?${qs}`)
    } catch (e) {
      state.loading = false
      state.error = 'Не удалось загрузить сравнение: ' + e.message
      return
    }
    const [pd, rd] = await Promise.allSettled([
      loadSide('pd', state.comparison.hasPdFile),
      loadSide('rd', state.comparison.hasRdFile),
    ])
    const pdInfo = pd.status === 'fulfilled' ? pd.value : emptyPages()
    const rdInfo = rd.status === 'fulfilled' ? rd.value : emptyPages()
    if (pdInfo.kind === 'xlsx' && rdInfo.kind === 'xlsx') alignXlsxSides(pdInfo, rdInfo)
    state.pdPages = pdInfo
    state.rdPages = rdInfo
    state.loading = false
    schedulePoll()
  }

  onMounted(start)
  onUnmounted(() => clearTimeout(pollTimer))

  return { state, qs, recordID }
}
