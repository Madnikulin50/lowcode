<script setup>
// Document + page choice for one stage. A stage can hold a whole package of
// files; the page counter belongs to the chosen file.
import { computed } from 'vue'

const props = defineProps({
  label: String,
  docs: Array, // documentRecord[] of this stage
  docId: String,
  page: Number,
  pageCount: Number,
  sheetNo: Object, // PDF page → the sheet number of its title block
  openHref: String, // the single-document viewer (with the assistant) on the file and page on screen
})
const emit = defineEmits(['update:docId', 'update:page'])

const current = computed(() => props.docs.find(d => d.id === props.docId))
const volume = computed(() => {
  const d = current.value
  return d && (d.volumeCode || d.volumeTitle) ? [d.volumeCode, d.volumeTitle].filter(Boolean).join(' — ') : ''
})
</script>

<template>
  <div class="doc-pager">
    <span class="dp-label">{{ label }}</span>
    <select v-if="docs.length > 1" :value="docId" @change="emit('update:docId', $event.target.value)">
      <option v-for="d in docs" :key="d.id" :value="d.id">{{ d.name || d.id }}</option>
    </select>
    <b v-else-if="docs.length === 1" class="dp-name">{{ docs[0].name || docs[0].id }}</b>
    <span v-else class="dp-none">нет файлов</span>
    <a v-if="openHref && docId" class="dp-open" :href="openHref" target="_blank" title="Открыть этот файл в отдельном просмотре документа (на этой странице) с ассистентом">↗ документ</a>
    <span v-if="volume" class="dp-volume" :title="volume">{{ volume }}</span>
    <span class="pager">
      <button :disabled="page <= 1" @click="emit('update:page', page - 1)">‹</button>
      <span>стр. {{ page }} / {{ pageCount || '—' }}<template v-if="sheetNo && sheetNo[page]"> · лист {{ sheetNo[page] }}</template></span>
      <button :disabled="page >= pageCount" @click="emit('update:page', page + 1)">›</button>
    </span>
  </div>
</template>
