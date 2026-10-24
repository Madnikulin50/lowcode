<script setup>
// Page viewer for the chosen ПД and РД documents.
import { computed } from 'vue'
import { ZOOM_LEVELS, MODE_META, names, nm } from '../constants.js'
import { pageSrc } from '../composables/useComparison.js'
import ModeTabs from './ModeTabs.vue'
import DocPager from './DocPager.vue'
import ZoomControls from './ZoomControls.vue'
import SideBySideStage from './SideBySideStage.vue'
import OverlayStage from './OverlayStage.vue'
import DiffStage from './DiffStage.vue'
import TextDiffView from './TextDiffView.vue'

const props = defineProps({
  pdDocs: Array,
  rdDocs: Array,
  pdInfo: Object,
  rdInfo: Object,
  qs: String,
  viewer: Object,
  expandFold: Function,
})

const hintKey = computed(() => (props.viewer.mode === 'diff' && props.viewer.diffColorized) ? 'diffColorized' : props.viewer.mode)
const zoomPct = computed(() => ZOOM_LEVELS[props.viewer.zoomIdx])
const pdCount = computed(() => props.pdInfo?.pageCount || 0)
const rdCount = computed(() => props.rdInfo?.pageCount || 0)
const pdUrl = computed(() => props.viewer.pdPage <= pdCount.value ? pageSrc(props.qs, props.pdInfo, props.viewer.pdDoc, props.viewer.pdPage) : '')
const rdUrl = computed(() => props.viewer.rdPage <= rdCount.value ? pageSrc(props.qs, props.rdInfo, props.viewer.rdDoc, props.viewer.rdPage) : '')
const nameOf = (docs, id, fallback) => docs.find(d => d.id === id)?.name || fallback
const pdLabel = computed(() => names.ref + ' — ' + nameOf(props.pdDocs, props.viewer.pdDoc, '') + ` (стр. ${props.viewer.pdPage})`)
const rdLabel = computed(() => names.target + ' — ' + nameOf(props.rdDocs, props.viewer.rdDoc, '') + ` (стр. ${props.viewer.rdPage})`)
const foldBase = computed(() => `${props.viewer.pdDoc}:${props.viewer.pdPage}|${props.viewer.rdDoc}:${props.viewer.rdPage}`)
const emit = defineEmits(['assistant'])

// Optional linked paging: turning a page of one document moves the other by
// the same number of pages (the offset between the two stays as it was).
function turn (side, page) {
  const v = props.viewer
  const other = side === 'pd' ? 'rd' : 'pd'
  const delta = page - v[side + 'Page']
  v[side + 'Page'] = page
  if (!v.syncPages || !delta) return
  const count = other === 'pd' ? pdCount.value : rdCount.value
  const next = v[other + 'Page'] + delta
  v[other + 'Page'] = count > 0 ? Math.min(Math.max(next, 1), count) : Math.max(next, 1)
}
// a link to the viewer of one document: the same file, the page on screen, back to this comparison
const docHref = (doc, page) => `?view=doc&source=comparison_doc&${props.qs}&id=${encodeURIComponent(doc || '')}&page=${page || 1}`
</script>

<template>
  <div>
    <ModeTabs :viewer="viewer" @assistant="emit('assistant')" />
    <div class="mode-hint">{{ nm(MODE_META[hintKey].hint) }}</div>

    <div class="toolbar-row">
      <div class="doc-pagers">
        <DocPager :open-href="docHref(viewer.pdDoc, viewer.pdPage)" :label="names.ref" :docs="pdDocs" v-model:doc-id="viewer.pdDoc" :page="viewer.pdPage" @update:page="turn('pd', $event)" :page-count="pdCount" :sheet-no="pdInfo?.sheetNo" />
        <DocPager :open-href="docHref(viewer.rdDoc, viewer.rdPage)" :label="names.target" :docs="rdDocs" v-model:doc-id="viewer.rdDoc" :page="viewer.rdPage" @update:page="turn('rd', $event)" :page-count="rdCount" :sheet-no="rdInfo?.sheetNo" />
      </div>
      <label class="sync-pages" title="Листать страницы обоих документов вместе: при смене страницы одного второй сдвигается на столько же">
        <input type="checkbox" v-model="viewer.syncPages" /> синхронно
      </label>
      <ZoomControls v-if="viewer.mode !== 'text'" :viewer="viewer" />
    </div>

    <SideBySideStage
      v-if="viewer.mode === 'side'"
      :pd-url="pdUrl" :rd-url="rdUrl" :pd-label="pdLabel" :rd-label="rdLabel" :zoom-pct="zoomPct"
      :hl-pd="viewer.hlPd" :hl-rd="viewer.hlRd"
    />
    <OverlayStage v-else-if="viewer.mode === 'overlay'" :pd-url="pdUrl" :rd-url="rdUrl" :zoom-pct="zoomPct" :viewer="viewer" />
    <DiffStage v-else-if="viewer.mode === 'diff'" :pd-url="pdUrl" :rd-url="rdUrl" :zoom-pct="zoomPct" :viewer="viewer" />
    <TextDiffView
      v-else-if="viewer.mode === 'text'"
      :pd-info="pdInfo" :rd-info="rdInfo" :pd-page="viewer.pdPage" :rd-page="viewer.rdPage"
      :fold-base="foldBase" :fold-state="viewer.foldState" :expand-fold="expandFold"
    />
  </div>
</template>
