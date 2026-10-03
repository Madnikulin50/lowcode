<script setup>
// Ported from the original vanilla-JS buildDiscrepancies().
import { computed } from 'vue'
import { SEV_ORDER, SEV_LABEL, TYPE_LABEL } from '../constants.js'
import { splitSource } from '../composables/useTextDiff.js'

const props = defineProps({ list: Array, pageCount: Number })
const emit = defineEmits(['jump'])

const counts = computed(() => {
  const c = { high: 0, medium: 0, low: 0 }
  for (const d of props.list) if (c[d.severity] != null) c[d.severity]++
  return c
})

const rows = computed(() => {
  const sorted = props.list.slice().sort((a, b) => (SEV_ORDER[a.severity] ?? 9) - (SEV_ORDER[b.severity] ?? 9))
  return sorted.map(d => {
    const { source, text } = splitSource(d.description)
    const typeLabel = d.type && TYPE_LABEL[d.type] ? TYPE_LABEL[d.type] : null
    const hasText = !!text.trim()
    // Some discrepancies (demo/seed data, or a check that only flags a page
    // without generating wording) carry no description at all — rendering
    // an empty row reads as broken. Fall back to naming the type instead of
    // showing nothing, and say plainly that there's no further detail.
    const primaryText = hasText ? text : (typeLabel ? `Расхождение: ${typeLabel}` : 'Расхождение без описания')

    // A page number is only trustworthy if it's actually inside the
    // rendered preview — a check can flag a page the viewer never loaded
    // (fewer real pages than expected, a fetch that came back short, …).
    // Rendering it exactly like a working "стр. N" link in that case reads
    // as a broken reference rather than what it is: informational only.
    const n = Number(d.pageNumber)
    const hasPage = !!d.pageNumber
    const pageAvailable = props.pageCount > 0 && n >= 1 && n <= props.pageCount

    let note = ''
    if (!hasText) {
      note = 'Автоматическая проверка не добавила текстовое описание.'
      if (hasPage) note += pageAvailable ? ' Откройте страницу, чтобы сравнить визуально.' : ' Указанной страницы нет в загруженном предпросмотре — визуальное сравнение недоступно.'
    }

    return {
      severity: d.severity,
      source,
      typeLabel,
      hasText,
      primaryText,
      note,
      n,
      hasPage,
      pageAvailable,
      pageText: 'стр. ' + d.pageNumber + (pageAvailable ? '' : ' ⚠'),
      pageTitle: pageAvailable ? 'Открыть страницу ' + d.pageNumber : `Страницы ${d.pageNumber} нет в загруженном предпросмотре (доступно страниц: ${props.pageCount || 0})`,
    }
  })
})
</script>

<template>
  <div class="discrepancies">
    <h2>Расхождения ({{ list.length }})</h2>

    <div v-if="!list.length" class="empty">Расхождений не найдено.</div>
    <template v-else>
      <div class="disc-summary">
        <template v-for="sev in ['high', 'medium', 'low']" :key="sev">
          <span v-if="counts[sev]" class="chip" :class="sev">
            <span class="dot"></span>{{ SEV_LABEL[sev] }}: {{ counts[sev] }}
          </span>
        </template>
      </div>

      <div
        v-for="(row, idx) in rows" :key="idx"
        class="disc" :class="{ clickable: row.pageAvailable }"
        :title="row.pageAvailable ? 'Открыть страницу ' + row.n : null"
        @click="row.pageAvailable && emit('jump', row.n)"
      >
        <span class="sev" :class="row.severity || ''">● {{ SEV_LABEL[row.severity] || row.severity || '' }}</span>
        <div class="body">
          <div class="tags">
            <span v-if="row.source" class="tag" :class="row.source === 'ai' ? 'src-ai' : ''">{{ row.source === 'ai' ? '🤖 ИИ' : '⚙ детерминированно' }}</span>
            <span v-if="row.typeLabel" class="tag">{{ row.typeLabel }}</span>
            <span v-if="row.hasPage" class="tag tag-page" :class="{ 'tag-page-unavailable': !row.pageAvailable }" :title="row.pageTitle">{{ row.pageText }}</span>
          </div>
          <span :class="row.hasText ? '' : 'disc-fallback'">{{ row.primaryText }}</span>
          <div v-if="!row.hasText" class="disc-note">{{ row.note }}</div>
        </div>
      </div>
    </template>
  </div>
</template>
