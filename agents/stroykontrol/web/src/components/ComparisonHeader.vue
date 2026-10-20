<script setup>
import { computed } from 'vue'
import { STATUS_LABEL, STATUS_CLASS, STATUS_ICON } from '../constants.js'
import Icon from './Icon.vue'
import { names, nm } from '../constants.js'

const props = defineProps({ comparison: Object, recordId: String, busy: Boolean, qs: String })
const emit = defineEmits(['run'])

const running = computed(() => !!props.comparison.running || props.comparison.status === 'processing')
const statusCls = computed(() => STATUS_CLASS[props.comparison.status] || '')
const title = computed(() => props.comparison.title || ('Сравнение #' + props.recordId))
const pct = computed(() => Math.max(0, Math.min(100, Number(props.comparison.similarityPercent) || 0)))
const barColor = computed(() => pct.value >= 80 ? 'var(--ok)' : pct.value >= 50 ? 'var(--warn)' : 'var(--bad)')
const refDocs = computed(() => props.comparison.documents.filter(d => d.side === 'ref'))
const targetDocs = computed(() => props.comparison.documents.filter(d => d.side === 'target'))
const reportHref = (format) => `api/report/export?${props.qs}&format=${format}`
const vol = (d) => [d.volumeCode, d.volumeTitle].filter(Boolean).join(' — ')
</script>

<template>
  <div class="head">
    <h1>{{ title }}</h1>
    <span class="badge" :class="statusCls">
      <Icon :svg="STATUS_ICON[comparison.status] || '•'" :cls="running ? 'spin' : ''" />
      {{ STATUS_LABEL[comparison.status] || comparison.status || '—' }}
      <template v-if="comparison.running">· {{ comparison.running }}</template>
    </span>
    <button class="run" :disabled="busy || running" :title="nm('Сравнить все загруженные файлы ПД с файлами РД')" @click="emit('run')">Запустить сравнение</button>
    <span v-if="qs" class="report-links" title="Сводка и все расхождения по томам и листам (справка, не список замечаний)">
      Отчёт: <a :href="reportHref('xlsx')" download>Excel</a> · <a :href="reportHref('docx')" download>Word</a>
    </span>
  </div>

  <div v-if="comparison.setsLabel" class="iteration sets-label">
    <b>Сравниваются версии комплектов:</b> <span class="chip ok">{{ comparison.setsLabel }}</span>
  </div>

  <div v-if="comparison.iteration" class="iteration">
    <b>Итерация {{ comparison.iteration.no }}</b>
    <span class="muted">— повторная проверка после исправлений (предыдущая: {{ comparison.iteration.parentTitle || comparison.iteration.parentID }})</span>
    <span class="chip ok">устранено: {{ comparison.iteration.resolved }}</span>
    <span class="chip high">не устранено: {{ comparison.iteration.persisting }}</span>
    <span class="chip medium">новых: {{ comparison.iteration.new }}</span>
  </div>

  <div v-if="refDocs.length || targetDocs.length" class="pack">
    <div class="pack-side">
      <span class="pack-h">{{ names.ref }} (эталон)</span>
      <div v-for="d in refDocs" :key="d.id" class="pack-doc"><b>{{ d.name || d.id }}</b><span v-if="vol(d)"> · {{ vol(d) }}</span><span v-if="d.pages"> · {{ d.pages }} стр.</span></div>
      <div v-if="!refDocs.length" class="pack-none">файлы не загружены</div>
    </div>
    <div class="pack-side">
      <span class="pack-h">{{ names.target }}</span>
      <div v-for="d in targetDocs" :key="d.id" class="pack-doc"><b>{{ d.name || d.id }}</b><span v-if="vol(d)"> · {{ vol(d) }}</span><span v-if="d.pages"> · {{ d.pages }} стр.</span></div>
      <div v-if="!targetDocs.length" class="pack-none">файлы не загружены</div>
    </div>
  </div>

  <div v-if="comparison.similarityPercent" class="simbar-row">
    <span class="simlabel">Схожесть</span>
    <div class="simbar"><div :style="{ width: pct + '%', background: barColor }"></div></div>
    <span class="simlabel">{{ pct }}%</span>
  </div>

  <div v-if="comparison.comment" class="comment">{{ comparison.comment }}</div>
</template>
