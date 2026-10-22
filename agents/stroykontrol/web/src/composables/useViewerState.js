// Replaces the original vanilla-JS module-level `let mode/page/zoomIdx/...`
// globals that every click handler mutated before calling the manual
// rerender() — here they're plain reactive fields Vue tracks itself.
import { reactive } from 'vue'
import { ZOOM_LEVELS } from '../constants.js'

export function useViewerState () {
  const state = reactive({
    mode: 'side',
    // The comparison can hold a package of files per stage: each side shows
    // one document at its own page (ПД page N is not РД page N once volumes
    // are matched across files).
    pdDoc: '',
    rdDoc: '',
    pdPage: 1,
    rdPage: 1,
    // Boxes (page fractions) of the discrepancy being looked at, drawn on
    // the page images in the side-by-side mode.
    activeId: '',
    hlPd: null,
    hlRd: null,
    zoomIdx: ZOOM_LEVELS.indexOf(100),
    diffColorized: false, // "Авто-разница" sub-option: gray (mix-blend difference) vs red/green

    // How the РД image gets sized against ПД in "Наложение"/"Авто-разница"
    // when the two page images have different pixel dimensions (different
    // source format/DPI — DOCX vs PDF vs DXF, or just a different page
    // size) — without this they'd each independently fill the stage width
    // and silently drift out of alignment. 'contain'/'cover' scale РД
    // uniformly (no distortion), 'stretch' scales X/Y independently to
    // exactly match ПД's box. manualDX/DY/Scale are a fine-tune layered on
    // top of whichever auto-fit is picked, for scans/exports that are close
    // but not pixel-aligned even after auto-fit.
    fitMode: 'contain',
    manualDX: 0,
    manualDY: 0,
    manualScale: 100,

    // "<page>:<blockIndex>" -> true once a text-diff fold's been expanded by
    // the user. Lives here (not inside TextDiffView) so it survives
    // switching away from the "Текст" mode and back.
    foldState: {},
  })

  function clampPages (pdCount, rdCount) {
    // An unknown count (the document is still loading after a jump to another
    // file) must not reset the page that was just chosen.
    if (pdCount > 0) state.pdPage = Math.min(Math.max(state.pdPage, 1), pdCount)
    if (rdCount > 0) state.rdPage = Math.min(Math.max(state.rdPage, 1), rdCount)
  }

  function zoomLevel () {
    return ZOOM_LEVELS[state.zoomIdx]
  }

  function expandFold (key) {
    state.foldState[key] = true
  }

  return { state, clampPages, zoomLevel, expandFold }
}
