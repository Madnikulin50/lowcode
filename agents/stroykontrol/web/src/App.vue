<script setup>
// Root of the comparison viewer: the package header, the page viewer for the
// chosen ПД/РД documents, the findings list and the remark cards.
import { computed, watch, nextTick } from 'vue'
import { useComparison } from './composables/useComparison.js'
import { useViewerState } from './composables/useViewerState.js'
import { parseBox } from './utils.js'
import ComparisonHeader from './components/ComparisonHeader.vue'
import Viewer from './components/Viewer.vue'
import DiscrepancyList from './components/DiscrepancyList.vue'
import RemarkCards from './components/RemarkCards.vue'
import ExpertisePanel from './components/ExpertisePanel.vue'

const { state: cmp, qs, recordID, documents, loadPair, run, setStatus, createRemark, saveRemark, suggestRemark, importExpertise, setExpertiseStatus } = useComparison()
const { state: viewer, clampPages, expandFold } = useViewerState()

const pdDocs = computed(() => documents('ref'))
const rdDocs = computed(() => documents('target'))
const pdInfo = computed(() => cmp.docs[viewer.pdDoc])
const rdInfo = computed(() => cmp.docs[viewer.rdDoc])
const canView = computed(() => !!(pdInfo.value?.supported && rdInfo.value?.supported && (pdInfo.value.pageCount || rdInfo.value.pageCount)))
const unsupportedReason = computed(() => {
  if (!pdDocs.value.length || !rdDocs.value.length) return 'Загрузите хотя бы один файл ПД и один файл РД.'
  if (!pdInfo.value || !rdInfo.value) return 'Загрузка страниц…'
  return 'Один из выбранных файлов не удалось отобразить постранично, показан только список расхождений ниже.'
})

// default selection: the first file of each stage
watch(() => cmp.comparison?.documents, () => {
  if (!pdDocs.value.some(d => d.id === viewer.pdDoc)) viewer.pdDoc = pdDocs.value[0]?.id || ''
  if (!rdDocs.value.some(d => d.id === viewer.rdDoc)) viewer.rdDoc = rdDocs.value[0]?.id || ''
}, { immediate: true, deep: true })

watch(() => [viewer.pdDoc, viewer.rdDoc], async ([pd, rd]) => {
  if (pd || rd) await loadPair(pd, rd)
}, { immediate: true })

watch(() => [pdInfo.value?.pageCount, rdInfo.value?.pageCount], ([a, b]) => clampPages(a || 0, b || 0))

// A click on a finding: open its pages in both documents and mark the region.
const activeId = computed(() => viewer.activeId)
function jump (d, scroll = true) {
  viewer.activeId = d.id || ''
  if (d.docRef) viewer.pdDoc = d.docRef
  if (d.docTarget) viewer.rdDoc = d.docTarget
  const pp = Number(d.pageRef) || 0
  const rp = Number(d.pageTarget) || 0
  const legacy = Number(d.pageNumber) || 0
  if (pp) viewer.pdPage = pp
  if (rp) viewer.rdPage = rp
  if (!pp && !rp && legacy) { viewer.pdPage = legacy; viewer.rdPage = legacy }
  viewer.hlPd = parseBox(d.bboxRef)
  viewer.hlRd = parseBox(d.bboxTarget)
  // regions are drawn on the side-by-side pages only
  if ((viewer.hlPd || viewer.hlRd) && (viewer.mode === 'overlay' || viewer.mode === 'diff')) viewer.mode = 'side'
  if (!scroll) return
  nextTick(() => {
    const target = document.querySelector('.stage, .text-diff')
    if (target) target.scrollIntoView({ behavior: 'smooth', block: 'center' })
  })
}

// Open on the most serious finding instead of two unrelated first files of a package.
let focused = false
watch(() => cmp.comparison?.discrepancies, (list) => {
  if (focused || !list?.length) return
  focused = true
  const first = list.slice().sort((a, b) => ({ high: 0, medium: 1, low: 2 }[a.severity] ?? 9) - ({ high: 0, medium: 1, low: 2 }[b.severity] ?? 9))
    .find(d => d.docRef && d.docTarget)
  if (first) jump(first, false)
}, { immediate: true })
</script>

<template>
  <div class="wrap">
    <template v-if="cmp.loading">Загрузка…</template>
    <div v-else-if="cmp.error" class="err">{{ cmp.error }}</div>
    <template v-else>
      <ComparisonHeader :comparison="cmp.comparison" :record-id="recordID" :busy="cmp.busy" @run="run" />

      <Viewer
        v-if="canView"
        :pd-docs="pdDocs" :rd-docs="rdDocs" :pd-info="pdInfo" :rd-info="rdInfo"
        :qs="qs" :viewer="viewer" :expand-fold="expandFold"
      />
      <div v-else class="unsupported">
        <div>Визуальное постраничное сравнение доступно для PDF, DOCX, XLSX и DXF.</div>
        <div>{{ unsupportedReason }}</div>
      </div>

      <DiscrepancyList
        :list="cmp.comparison.discrepancies" :documents="cmp.comparison.documents" :busy="cmp.busy" :active-id="activeId" :error="cmp.actionError"
        @jump="jump" @status="setStatus" @remark="createRemark"
      />
      <ExpertisePanel
        :list="cmp.comparison.expertise" :discrepancies="cmp.comparison.discrepancies" :busy="cmp.busy"
        @import="importExpertise" @status="setExpertiseStatus" @jump="jump"
      />
      <RemarkCards :remarks="cmp.comparison.remarks" :busy="cmp.busy" @save="saveRemark" @suggest="suggestRemark" />
    </template>
  </div>
</template>
