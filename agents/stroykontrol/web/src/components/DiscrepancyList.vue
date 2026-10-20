<script setup>
// Findings of the comparison: filter, review (confirm / reject), pick
// several and turn them into a remark card. A click on a row opens the
// pages and highlights the region in the viewer.
import { computed, reactive } from 'vue'
import { SEV_ORDER, SEV_LABEL, TYPE_LABEL, KIND_LABEL, REVIEW_LABEL, ITER_LABEL, names } from '../constants.js'
import { splitSource } from '../composables/useTextDiff.js'
import { parseEvidence } from '../utils.js'

const props = defineProps({ list: Array, documents: Array, busy: Boolean, activeId: String, error: String })
const emit = defineEmits(['jump', 'status', 'status-many', 'remark'])

const INFO_KINDS = ['not_released', 'scope_diff']
const filter = reactive({ severity: '', status: '', kind: '', iter: '', hideInfo: false })
const selected = reactive(new Set())
const open = reactive(new Set())

const docName = (id) => props.documents.find(d => d.id === id)?.name || ''

const counts = computed(() => {
  const c = { high: 0, medium: 0, low: 0 }
  for (const d of props.list) if (c[d.severity] != null) c[d.severity]++
  return c
})
const reviewCounts = computed(() => {
  const c = { new: 0, confirmed: 0, rejected: 0 }
  for (const d of props.list) c[d.status || 'new'] = (c[d.status || 'new'] || 0) + 1
  return c
})
const hasIterations = computed(() => props.list.some(d => d.iterState))
const kindsPresent = computed(() => [...new Set(props.list.map(d => d.kind).filter(Boolean))])

const rows = computed(() => {
  const sorted = props.list.slice().sort((a, b) => (SEV_ORDER[a.severity] ?? 9) - (SEV_ORDER[b.severity] ?? 9))
  return sorted
    .filter(d => !filter.severity || d.severity === filter.severity)
    .filter(d => !filter.status || (d.status || 'new') === filter.status)
    .filter(d => !filter.kind || d.kind === filter.kind)
    .filter(d => !filter.iter || d.iterState === filter.iter)
    .filter(d => !(filter.hideInfo && INFO_KINDS.includes(d.kind)))
    .map(d => {
      const { source, text } = splitSource(d.description)
      const typeLabel = d.type && TYPE_LABEL[d.type] ? TYPE_LABEL[d.type] : null
      const hasText = !!text.trim()
      const primaryText = hasText ? text : (typeLabel ? `Расхождение: ${typeLabel}` : 'Расхождение без описания')
      const legacyPage = d.pageNumber && !d.pageRef && !d.pageTarget ? d.pageNumber : ''
      const ev = parseEvidence(d.evidence)
      return {
        d, source, text: primaryText, hasText, typeLabel,
        kindLabel: d.kind ? KIND_LABEL[d.kind] : '',
        info: INFO_KINDS.includes(d.kind),
        status: d.status || 'new',
        iterState: d.iterState || '',
        pd: d.pageRef ? `${names.ref} стр. ${d.pageRef}` : '',
        rd: d.pageTarget ? `${names.target} стр. ${d.pageTarget}` : '',
        pdTitle: docName(d.docRef),
        rdTitle: docName(d.docTarget),
        legacyPage: legacyPage ? `стр. ${legacyPage}` : '',
        clickable: !!(d.pageRef || d.pageTarget || d.pageNumber),
        evidence: ev && (ev.ref || ev.target || ev.rule) ? ev : null,
      }
    })
})

// everything above "low" that nobody has decided on yet
const confirmable = computed(() => props.list.filter(d => (d.status || 'new') === 'new' && d.severity !== 'low' && !INFO_KINDS.includes(d.kind) && d.id).map(d => d.id))
function confirmAll () { emit('status-many', confirmable.value, 'confirmed') }
function toggle (set, id) { if (set.has(id)) set.delete(id); else set.add(id) }
function makeCard () {
  emit('remark', [...selected])
  selected.clear()
}
</script>

<template>
  <div class="discrepancies">
    <h2>Расхождения ({{ list.length }})</h2>
    <div v-if="error" class="err action-err">Не удалось выполнить действие: {{ error }}</div>

    <div v-if="!list.length" class="empty">Расхождений не найдено.</div>
    <template v-else>
      <div class="disc-summary">
        <template v-for="sev in ['high', 'medium', 'low']" :key="sev">
          <span v-if="counts[sev]" class="chip" :class="sev"><span class="dot"></span>{{ SEV_LABEL[sev] }}: {{ counts[sev] }}</span>
        </template>
        <span class="chip">подтверждено: {{ reviewCounts.confirmed }}</span>
        <span class="chip">отклонено: {{ reviewCounts.rejected }}</span>
      </div>

      <div class="disc-filters">
        <select v-model="filter.severity">
          <option value="">любая критичность</option>
          <option v-for="s in ['high', 'medium', 'low']" :key="s" :value="s">{{ SEV_LABEL[s] }}</option>
        </select>
        <select v-model="filter.status">
          <option value="">любое решение</option>
          <option v-for="(label, k) in REVIEW_LABEL" :key="k" :value="k">{{ label }}</option>
        </select>
        <select v-if="hasIterations" v-model="filter.iter">
          <option value="">все итерации</option>
          <option v-for="(label, k) in ITER_LABEL" :key="k" :value="k">{{ label }}</option>
        </select>
        <select v-if="kindsPresent.length" v-model="filter.kind">
          <option value="">любой вид</option>
          <option v-for="k in kindsPresent" :key="k" :value="k">{{ KIND_LABEL[k] || k }}</option>
        </select>
        <label class="chk"><input type="checkbox" v-model="filter.hideInfo" /> скрыть «не выпущено / вне объёма»</label>
        <span class="grow"></span>
        <button class="small" :disabled="busy || !confirmable.length" title="Подтвердить все нерешённые расхождения выше «низкой» критичности: по подтверждённым формируются замечания" @click="confirmAll">Подтвердить все выше «низкой» ({{ confirmable.length }})</button>
        <button class="primary" :disabled="busy || !selected.size" @click="makeCard">Создать карточку замечания ({{ selected.size }})</button>
      </div>

      <div v-if="!rows.length" class="empty">Нет расхождений под выбранный фильтр.</div>
      <div
        v-for="row in rows" :key="row.d.id || row.text"
        class="disc" :class="{ clickable: row.clickable, info: row.info, resolved: row.iterState === 'resolved', active: row.d.id && row.d.id === activeId, rejected: row.status === 'rejected' }"
        @click="row.clickable && emit('jump', row.d)"
      >
        <input v-if="row.d.id" type="checkbox" class="pick" :checked="selected.has(row.d.id)" @click.stop @change="toggle(selected, row.d.id)" />
        <span class="sev" :class="row.d.severity || ''">● {{ SEV_LABEL[row.d.severity] || row.d.severity || '' }}</span>
        <div class="body">
          <div class="tags">
            <span v-if="row.source" class="tag" :class="row.source === 'ai' ? 'src-ai' : ''">{{ row.source === 'ai' ? '🤖 ИИ' : '⚙ детерминированно' }}</span>
            <span v-if="row.kindLabel" class="tag">{{ row.kindLabel }}</span>
            <span v-else-if="row.typeLabel" class="tag">{{ row.typeLabel }}</span>
            <span v-if="row.pd" class="tag tag-page" :title="row.pdTitle">{{ row.pd }}</span>
            <span v-if="row.rd" class="tag tag-page" :title="row.rdTitle">{{ row.rd }}</span>
            <span v-if="row.legacyPage" class="tag tag-page">{{ row.legacyPage }}</span>
            <span v-if="row.iterState" class="tag" :class="'it-' + row.iterState">{{ ITER_LABEL[row.iterState] }}</span>
            <span v-if="row.status !== 'new'" class="tag" :class="'st-' + row.status">{{ REVIEW_LABEL[row.status] }}</span>
          </div>
          <span :class="row.hasText ? '' : 'disc-fallback'">{{ row.text }}</span>
          <div v-if="row.evidence" class="ev-toggle" @click.stop="toggle(open, row.d.id)">{{ open.has(row.d.id) ? '▾' : '▸' }} на основании чего</div>
          <div v-if="row.evidence && open.has(row.d.id)" class="evidence" @click.stop>
            <div v-if="row.evidence.ref"><b>{{ names.ref }}:</b> {{ row.evidence.ref }}</div>
            <div v-if="row.evidence.target"><b>{{ names.target }}:</b> {{ row.evidence.target }}</div>
            <div v-if="row.evidence.volumeRef"><b>Том «{{ names.ref }}»:</b> {{ row.evidence.volumeRef }}</div>
            <div v-if="row.evidence.volumeTarget"><b>Том(а) «{{ names.target }}»:</b> {{ row.evidence.volumeTarget }}</div>
            <div v-if="row.evidence.rule" class="rule">правило: {{ row.evidence.rule }}</div>
          </div>
        </div>
        <div v-if="row.d.id" class="review" @click.stop>
          <button :class="{ on: row.status === 'confirmed' }" :disabled="busy" title="Подтвердить расхождение" @click="emit('status', row.d.id, row.status === 'confirmed' ? 'new' : 'confirmed')">✓</button>
          <button :class="{ on: row.status === 'rejected' }" :disabled="busy" title="Отклонить (ложное срабатывание)" @click="emit('status', row.d.id, row.status === 'rejected' ? 'new' : 'rejected')">✗</button>
        </div>
      </div>
    </template>
  </div>
</template>
