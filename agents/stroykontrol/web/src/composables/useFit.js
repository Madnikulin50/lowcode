// Sizes/positions the "top" (РД) image against "bottom" (ПД) so their
// content lines up even when the two page images have different pixel
// dimensions (different export path/DPI — DOCX vs PDF vs DXF client-side
// rendering, or genuinely a different page size). Without this, both images
// independently fill the stage width via CSS and same-looking content
// quietly drifts out of alignment. 'contain'/'cover' scale uniformly (no
// distortion, only differ in whether they under- or over-fill ПД's box —
// 'cover' can overflow it, left to the stage's own scroll rather than a
// hard clip); 'stretch' scales X/Y independently to force an exact match.
// manualDX/DY/Scale is a fine-tune (typed in, or set by the align-* buttons)
// layered on top of whichever auto-fit is active. Ported from the original
// vanilla-JS computeFitBox/applyFit/alignTo/scheduleApplyFit.
import { ref, watch, onMounted } from 'vue'

// Shared by applyFit (needs the scale factors to build the CSS transform)
// and the align-* buttons (need to know how much smaller/larger top's box
// ends up than bottom's, to compute the dx/dy that lines a given edge or
// the center up) — bottom's own rendered box plus top's rendered box size
// after the active auto-fit mode's scale, both in real CSS px, before any
// manual offset/scale. Returns null while either image isn't laid out yet.
export function computeFitBox (bottom, top, fitMode) {
  if (!bottom.complete || !top.complete || !bottom.naturalWidth || !top.naturalWidth) return null
  const bRect = bottom.getBoundingClientRect()
  if (!bRect.width || !bRect.height) return null
  const bw = bottom.naturalWidth, bh = bottom.naturalHeight
  const tw = top.naturalWidth, th = top.naturalHeight
  const pxPerUnit = bRect.width / bw
  let sx, sy
  if (fitMode === 'stretch') {
    sx = bw / tw
    sy = bh / th
  } else {
    const s = fitMode === 'cover' ? Math.max(bw / tw, bh / th) : Math.min(bw / tw, bh / th)
    sx = sy = s
  }
  return { bw: bRect.width, bh: bRect.height, finalW: tw * pxPerUnit * sx, finalH: th * pxPerUnit * sy }
}

export function applyFit (bottom, top, viewer) {
  top.style.transform = 'none'
  const box = computeFitBox(bottom, top, viewer.fitMode)
  if (!box) return
  const tRect = top.getBoundingClientRect()
  if (!tRect.width || !tRect.height) return
  const scaleX = box.finalW / tRect.width
  const scaleY = box.finalH / tRect.height
  const manualS = viewer.manualScale / 100
  top.style.transformOrigin = 'top left'
  top.style.transform = `translate(${viewer.manualDX}px, ${viewer.manualDY}px) scale(${scaleX * manualS}, ${scaleY * manualS})`
}

// Quick-align buttons: sets manualDX/DY to whatever offset makes the given
// edge (or center) of РД's *current* auto-fit box (contain/cover/stretch,
// already scaled by manualScale) coincide with ПД's — a one-click substitute
// for dragging the manual X/Y inputs to the right number by eye. 'left'/'top'
// is just 0 (also the default, before any manual offset); the other four are
// genuinely only reachable through this or hand-typed input.
export function alignTo (bottom, top, h, v, viewer) {
  const box = computeFitBox(bottom, top, viewer.fitMode)
  if (!box) return
  const s = viewer.manualScale / 100
  const scaledW = box.finalW * s
  const scaledH = box.finalH * s
  if (h === 'left') viewer.manualDX = 0
  else if (h === 'right') viewer.manualDX = box.bw - scaledW
  else if (h === 'center') viewer.manualDX = (box.bw - scaledW) / 2
  if (v === 'top') viewer.manualDY = 0
  else if (v === 'bottom') viewer.manualDY = box.bh - scaledH
  else if (v === 'center') viewer.manualDY = (box.bh - scaledH) / 2
}

// applyFit needs real layout (getBoundingClientRect) from both images, which
// only exists once each has actually loaded AND the stage is attached to the
// document. Re-running opportunistically on each image's own load event,
// plus once via requestAnimationFrame for the case both are already
// browser-cached and .complete immediately, covers every ordering;
// applyFit's own completeness guard makes redundant calls harmless no-ops.
export function scheduleApplyFit (bottom, top, viewer) {
  const run = () => applyFit(bottom, top, viewer)
  if (bottom.complete) run(); else bottom.onload = run
  if (top.complete) run(); else top.onload = run
  requestAnimationFrame(run)
}

// Shared by OverlayStage/DiffStage: both render a bottom/top <img> pair and
// need the same "recompute fit whenever the page, fit mode or manual
// offset/scale changes" wiring — the original got this for free from a full
// rerender() on every such change; here it's an explicit watcher instead.
export function useFitStage (props) {
  const bottomRef = ref(null)
  const topRef = ref(null)

  function reapply () {
    if (bottomRef.value && topRef.value) scheduleApplyFit(bottomRef.value, topRef.value, props.viewer)
  }

  function onAlign (h, v) {
    if (!bottomRef.value || !topRef.value) return
    alignTo(bottomRef.value, topRef.value, h, v, props.viewer)
    reapply()
  }

  watch(() => [props.pdUrl, props.rdUrl, props.viewer.fitMode, props.viewer.manualDX, props.viewer.manualDY, props.viewer.manualScale], reapply)
  onMounted(reapply)

  return { bottomRef, topRef, onAlign }
}
