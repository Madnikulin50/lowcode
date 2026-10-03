<script setup>
// Ported from the original vanilla-JS main()/render() — this is the root
// that used to wipe and rebuild #app on every state change.
import { computed, watch, nextTick } from 'vue'
import { useComparison } from './composables/useComparison.js'
import { useViewerState } from './composables/useViewerState.js'
import ComparisonHeader from './components/ComparisonHeader.vue'
import Viewer from './components/Viewer.vue'
import DiscrepancyList from './components/DiscrepancyList.vue'

const { state: cmp, qs, recordID } = useComparison()
const { state: viewer, clampPage, expandFold } = useViewerState()

const bothSupported = computed(() => cmp.pdPages.supported && cmp.rdPages.supported)
const pageCount = computed(() => Math.max(cmp.pdPages.pageCount || 0, cmp.rdPages.pageCount || 0))
const unsupportedReason = computed(() => {
  const c = cmp.comparison
  if (!c) return ''
  return (c.hasPdFile && !cmp.pdPages.supported) || (c.hasRdFile && !cmp.rdPages.supported)
    ? 'Один из файлов не удалось отобразить постранично, показан только текстовый список расхождений ниже.'
    : 'Файлы ещё не загружены.'
})

watch(pageCount, (pc) => clampPage(pc), { immediate: true })

// Jumps the page viewer to a discrepancy's page and scrolls it into view —
// called from a discrepancy row click (see DiscrepancyList).
function jumpToPage (n) {
  viewer.page = n
  nextTick(() => {
    const target = document.querySelector('.stage, .text-diff')
    if (target) target.scrollIntoView({ behavior: 'smooth', block: 'center' })
  })
}
</script>

<template>
  <div class="wrap">
    <template v-if="cmp.loading">Загрузка…</template>
    <div v-else-if="cmp.error" class="err">{{ cmp.error }}</div>
    <template v-else>
      <ComparisonHeader :comparison="cmp.comparison" :record-id="recordID" />

      <Viewer
        v-if="bothSupported && pageCount > 0"
        :pd-pages="cmp.pdPages" :rd-pages="cmp.rdPages" :comparison="cmp.comparison"
        :qs="qs" :viewer="viewer" :page-count="pageCount" :expand-fold="expandFold"
      />
      <div v-else class="unsupported">
        <div>Визуальное постраничное наложение доступно для PDF, DOCX и DXF.</div>
        <div>{{ unsupportedReason }}</div>
      </div>

      <DiscrepancyList :list="cmp.comparison.discrepancies || []" :page-count="bothSupported ? pageCount : 0" @jump="jumpToPage" />
    </template>
  </div>
</template>
