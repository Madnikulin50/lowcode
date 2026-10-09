// DOCX has no server-side page-image concept (no LibreOffice on the server
// host), so it's rendered client-side: docx-preview lays it out into
// paginated HTML (one <section class="docx"> per page), then html2canvas
// rasterizes each page into an image — same page-image abstraction the PDF
// side already gets from the server, so the rest of the viewer (side-by-side
// / overlay / diff) doesn't need to know which kind it's looking at. Ported
// verbatim from the original vanilla-JS index.html.
import { renderAsync } from 'docx-preview'
import html2canvas from 'html2canvas'

// Recursively pulls one text "line" per paragraph/heading/list-item/table
// row out of a docx-preview page section. Plain sec.textContent would flatten
// every paragraph into one blob (browsers don't insert separators between
// block elements), which is fine for the old flowing word-diff but useless
// for a line-based git-style diff — this is what gives DOCX pages the same
// line boundaries a PDF page already gets for free from `pdftotext -layout`.
export function linesFromElement (root) {
  const lines = []
  const LEAF = /^(P|LI|H[1-6])$/
  ;(function walk (node) {
    for (const child of node.children || []) {
      if (child.tagName === 'TABLE') {
        for (const tr of child.querySelectorAll('tr')) {
          lines.push(Array.from(tr.children).map(td => (td.textContent || '').trim()).join(' | '))
        }
      } else if (LEAF.test(child.tagName)) {
        lines.push((child.textContent || '').trim())
      } else {
        walk(child)
      }
    }
  })(root)
  return lines
}

export async function renderDocxPages (qs, side) {
  const res = await fetch(`api/file?${qs}&side=${side}`)
  if (!res.ok) throw new Error('не удалось получить файл (' + res.status + ')')
  const blob = await res.blob()

  // A far-offscreen fixed position (the old approach) still grows the
  // document's scrollable area in most browsers, dragging the real viewport
  // sideways. Clip through a 0x0 overflow:hidden ancestor instead — the
  // inner host keeps its real width for layout/html2canvas, it's just not
  // visible or scrollable-to.
  const clip = document.createElement('div')
  clip.style.cssText = 'position:fixed; top:0; left:0; width:0; height:0; overflow:hidden;'
  document.body.appendChild(clip)
  const host = document.createElement('div')
  host.style.cssText = 'width:794px; background:#fff;'
  clip.appendChild(host)
  const shadow = host.attachShadow({ mode: 'open' })
  const styleHost = document.createElement('div')
  const body = document.createElement('div')
  shadow.appendChild(styleHost)
  shadow.appendChild(body)

  try {
    await renderAsync(blob, body, styleHost, { inWrapper: true, breakPages: true, ignoreLastRenderedPageBreak: true })
    const sections = body.querySelectorAll('.docx-wrapper > section.docx')
    const images = []
    const texts = []
    for (const sec of sections) {
      const canvas = await html2canvas(sec, { scale: 2, backgroundColor: '#ffffff' })
      images.push(canvas.toDataURL('image/jpeg', 0.92))
      texts.push(linesFromElement(sec).join('\n'))
    }
    return { images, texts }
  } finally {
    clip.remove()
  }
}
