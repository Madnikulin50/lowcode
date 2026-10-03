<script setup>
// Ported from the original vanilla-JS buildViewer().
import { computed } from 'vue'
import { ZOOM_LEVELS, MODE_META } from '../constants.js'
import { pageSrc } from '../composables/useComparison.js'
import ModeTabs from './ModeTabs.vue'
import Pager from './Pager.vue'
import ZoomControls from './ZoomControls.vue'
import SideBySideStage from './SideBySideStage.vue'
import OverlayStage from './OverlayStage.vue'
import DiffStage from './DiffStage.vue'
import TextDiffView from './TextDiffView.vue'

const props = defineProps({
  pdPages: Object,
  rdPages: Object,
  comparison: Object,
  qs: String,
  viewer: Object,
  pageCount: Number,
  expandFold: Function,
})

const hintKey = computed(() => (props.viewer.mode === 'diff' && props.viewer.diffColorized) ? 'diffColorized' : props.viewer.mode)
const zoomPct = computed(() => ZOOM_LEVELS[props.viewer.zoomIdx])
const pdUrl = computed(() => pageSrc(props.qs, props.pdPages, 'pd', props.viewer.page))
const rdUrl = computed(() => pageSrc(props.qs, props.rdPages, 'rd', props.viewer.page))
const pdLabel = computed(() => props.comparison?.pdFileName ? 'ПД — ' + props.comparison.pdFileName : 'ПД')
const rdLabel = computed(() => props.comparison?.rdFileName ? 'РД — ' + props.comparison.rdFileName : 'РД')
</script>

<template>
  <div>
    <ModeTabs :viewer="viewer" />
    <div class="mode-hint">{{ MODE_META[hintKey].hint }}</div>

    <div class="toolbar-row">
      <Pager :viewer="viewer" :page-count="pageCount" />
      <ZoomControls v-if="viewer.mode !== 'text'" :viewer="viewer" />
    </div>

    <SideBySideStage v-if="viewer.mode === 'side'" :pd-url="pdUrl" :rd-url="rdUrl" :pd-label="pdLabel" :rd-label="rdLabel" :zoom-pct="zoomPct" />
    <OverlayStage v-else-if="viewer.mode === 'overlay'" :pd-url="pdUrl" :rd-url="rdUrl" :zoom-pct="zoomPct" :viewer="viewer" />
    <DiffStage v-else-if="viewer.mode === 'diff'" :pd-url="pdUrl" :rd-url="rdUrl" :zoom-pct="zoomPct" :viewer="viewer" />
    <TextDiffView v-else-if="viewer.mode === 'text'" :pd-pages="pdPages" :rd-pages="rdPages" :page="viewer.page" :fold-state="viewer.foldState" :expand-fold="expandFold" />
  </div>
</template>
