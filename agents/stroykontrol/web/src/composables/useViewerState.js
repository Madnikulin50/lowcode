// Replaces the original vanilla-JS module-level `let mode/page/zoomIdx/...`
// globals that every click handler mutated before calling the manual
// rerender() — here they're plain reactive fields Vue tracks itself.
import { reactive } from 'vue'
import { ZOOM_LEVELS } from '../constants.js'

export function useViewerState () {
  const state = reactive({
    mode: 'side',
    page: 1,
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

  function clampPage (pageCount) {
    if (state.page > pageCount) state.page = pageCount
    if (state.page < 1) state.page = 1
  }

  function zoomLevel () {
    return ZOOM_LEVELS[state.zoomIdx]
  }

  function expandFold (key) {
    state.foldState[key] = true
  }

  return { state, clampPage, zoomLevel, expandFold }
}
