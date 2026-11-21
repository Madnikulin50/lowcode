<script setup>
// Root of the comparison viewer: the package header, the page viewer for the
// chosen ПД/РД documents, the findings list and the remark cards.
import { computed, ref, watch, nextTick } from 'vue'
import { names } from './constants.js'
import { useComparison } from './composables/useComparison.js'
import { useViewerState } from './composables/useViewerState.js'
import { parseBox } from './utils.js'
import ComparisonHeader from './components/ComparisonHeader.vue'
import PairingPanel from './components/PairingPanel.vue'
import SheetChangesPanel from './components/SheetChangesPanel.vue'
import AssistantPanel from './components/AssistantPanel.vue'
import Viewer from './components/Viewer.vue'
import DiscrepancyList from './components/DiscrepancyList.vue'
import RemarkCards from './components/RemarkCards.vue'
import ExpertisePanel from './components/ExpertisePanel.vue'
import IssuePanel from './components/IssuePanel.vue'

const { state: cmp, qs, recordID, namespaceID, documents, loadPair, run, setStatus, setStatusMany, registerIssue, createRemark, saveRemark, suggestRemark } = useComparison()
const { state: viewer, clampPages, expandFold } = useViewerState()

const pdDocs = computed(() => documents('ref'))
const rdDocs = computed(() => documents('target'))
const pdInfo = computed(() => cmp.docs[viewer.pdDoc])
const rdInfo = computed(() => cmp.docs[viewer.rdDoc])
const canView = computed(() => !!(pdInfo.value?.supported && rdInfo.value?.supported && (pdInfo.value.pageCount || rdInfo.value.pageCount)))
const unsupportedReason = computed(() => {
  if (!pdDocs.value.length || !rdDocs.value.length) return 'Загрузите хотя бы один файл ПД и один файл РД.'
  if (!pdInfo.value || !rdInfo.value) return 'Загрузка страниц…'
  const bad = []
  for (const [label, info, doc] of [[names.ref, pdInfo.value, viewer.pdDoc], [names.target, rdInfo.value, viewer.rdDoc]]) {
    if (info.error) bad.push(`${label} — ${docName(doc)}: ${info.error}`)
    else if (!info.supported) bad.push(`${label} — ${docName(doc)}: этот тип файла (${info.kind}) нельзя показать постранично.`)
    else if (!info.pageCount) bad.push(`${label} — ${docName(doc)}: в файле нет страниц для показа.`)
  }
  return bad.length ? bad.join('\n') : 'Один из выбранных файлов не удалось отобразить постранично, показан только список расхождений ниже.'
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
// A finding with no page on one side: that side says so instead of showing the page that was open before.
function missingNote (d, side) {
  const here = side === 'ref' ? names.ref : names.target
  const other = side === 'ref' ? names.target : names.ref
  if (side === 'target' && (d.type === 'missing_page' || d.kind === 'not_released')) return `Для этого замечания в ${here} нет страницы: лист ${other} остался без пары (возможно, ещё не выпущен).`
  if (side === 'ref' && (d.type === 'extra_page' || d.kind === 'scope_diff')) return `Для этого замечания в ${here} нет страницы: лист есть только в ${other}.`
  return `У этого замечания нет страницы в ${here}: оно относится только к ${other}.`
}
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
  viewer.noPd = pp || legacy ? '' : missingNote(d, 'ref')
  viewer.noRd = rp || legacy ? '' : missingNote(d, 'target')
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

// What the assistant is asked about: the sheets on screen and, if one is selected, the finding.
const assistScopes = [
  { key: 'both', label: 'Оба открытых листа' },
  { key: 'ref', label: 'Только эталон' },
  { key: 'target', label: 'Только проверяемый' },
  { key: 'finding', label: 'Только выбранная находка' },
]
const docName = id => (cmp.comparison?.documents || []).find(d => d.id === id)?.name || id
const assistChips = computed(() => {
  const out = []
  if (viewer.pdDoc) out.push(`${names.ref}: ${docName(viewer.pdDoc)}, стр. ${viewer.pdPage}`)
  if (viewer.rdDoc) out.push(`${names.target}: ${docName(viewer.rdDoc)}, стр. ${viewer.rdPage}`)
  return out
})
const activeLabel = computed(() => {
  const d = (cmp.comparison?.discrepancies || []).find(x => x.id === viewer.activeId)
  const t = d ? (d.description || '') : ''
  return t.length > 60 ? t.slice(0, 60) + '…' : t
})
const assistant = ref(null)
function openAssistant () {
  assistant.value?.openAndFocus()
  nextTick(() => document.querySelector('.assistant')?.scrollIntoView({ behavior: 'smooth', block: 'start' }))
}
function assistContext (scope) {
  return {
    source: 'comparison', recordID, scope, discrepancyID: viewer.activeId || '',
    ref: { doc: viewer.pdDoc, page: viewer.pdPage }, target: { doc: viewer.rdDoc, page: viewer.rdPage },
  }
}

// The dialog sends the user to a sheet it named; a note it wrote can go into the remark card.
function gotoSheet (j) {
  if (j.side === 'ref') { viewer.pdDoc = j.doc; viewer.pdPage = j.page; viewer.noPd = '' } else { viewer.rdDoc = j.doc; viewer.rdPage = j.page; viewer.noRd = '' }
  viewer.hlPd = null
  viewer.hlRd = null
  nextTick(() => {
    const target = document.querySelector('.stage, .text-diff')
    if (target) target.scrollIntoView({ behavior: 'smooth', block: 'center' })
  })
}
async function textToRemark ({ id, text }) {
  const card = await createRemark([id])
  if (card) await saveRemark({ ...card, draftText: text })
}

// Open on the most serious finding instead of two unrelated first files of a package.
let focused = false
watch(() => cmp.comparison?.discrepancies, (list) => {
  if (focused || !list?.length) return
  focused = true
  // a link from a remark (?findingID=): open that finding's sheets and mark its region
  const wanted = new URLSearchParams(location.search).get('findingID')
  const linked = wanted && list.find(d => d.id === wanted)
  if (linked) { jump(linked); return }
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
      <ComparisonHeader :comparison="cmp.comparison" :record-id="recordID" :qs="qs" :busy="cmp.busy" @run="run" />

      <Viewer
        v-if="canView"
        :pd-docs="pdDocs" :rd-docs="rdDocs" :pd-info="pdInfo" :rd-info="rdInfo"
        :qs="qs" :viewer="viewer" :expand-fold="expandFold" @assistant="openAssistant"
      />
      <div v-else class="unsupported">
        <div><b>Страницы не показаны.</b> Визуальное постраничное сравнение доступно для PDF, DWG, DOCX, XLSX и DXF.</div>
        <div class="unsupported-why">{{ unsupportedReason }}</div>
        <div>Список расхождений ниже работает; файл можно скачать или открыть в «↗ документ».</div>
      </div>

      <PairingPanel :qs="qs" :record-id="recordID" :namespace-id="namespaceID" :running="cmp.comparison.running" :status="cmp.comparison.status" @apply="run" />
      <SheetChangesPanel :qs="qs" :running="cmp.comparison.running" :status="cmp.comparison.status" @jump="jump" />

      <AssistantPanel
        ref="assistant"
        :namespace-id="namespaceID" :get-context="assistContext" :chips="assistChips" :scopes="assistScopes" default-scope="both"
        :active-id="activeId" :active-label="activeLabel" title="Ассистент по документу"
        @goto="gotoSheet" @to-remark="textToRemark"
      />

      <DiscrepancyList
        :list="cmp.comparison.discrepancies" :documents="cmp.comparison.documents" :busy="cmp.busy" :active-id="activeId" :error="cmp.actionError" :pd-page="viewer.pdPage" :rd-page="viewer.rdPage"
        @jump="jump" @status="setStatus" @status-many="setStatusMany" @remark="createRemark"
      />
      <IssuePanel
        :issue="cmp.issue" :result="cmp.issueResult" :busy="cmp.busy" :qs="qs" :namespace-id="namespaceID" :discrepancies="cmp.comparison.discrepancies"
        @register="registerIssue" @confirm-many="(ids) => setStatusMany(ids, 'confirmed')"
      />
      <ExpertisePanel
        :list="cmp.comparison.expertise" :discrepancies="cmp.comparison.discrepancies" :object-id="cmp.comparison.objectID" :namespace-id="namespaceID"
        @jump="jump"
      />
      <RemarkCards :remarks="cmp.comparison.remarks" :busy="cmp.busy" @save="saveRemark" @suggest="suggestRemark" />
    </template>
  </div>
</template>
