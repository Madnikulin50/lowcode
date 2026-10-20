<script setup>
// Remarks of the expertise that concern this comparison — linked to its
// documents or matched to its findings — with the designer's answer and what
// the check against the documents found. The registry itself (history, status,
// assignees) is in the remarks workbench of the object.
import { computed, reactive } from 'vue'
import { VERDICT_LABEL } from '../remarks/labels.js'
import { STATUS_LABEL } from '../remarks/labels.js'

const props = defineProps({ list: Array, discrepancies: Array, objectId: String, namespaceId: String })
const emit = defineEmits(['jump'])

const open = reactive(new Set())
const filter = reactive({ verdict: '' })

const counts = computed(() => {
  const c = { still_differs: 0, resolved: 0, no_match: 0, no_documents: 0, claim: 0 }
  for (const e of props.list) {
    if (c[e.verdict] != null) c[e.verdict]++
    if (e.claimMismatch) c.claim++
  }
  return c
})
const rows = computed(() => props.list.filter(e => !filter.verdict || e.verdict === filter.verdict))
const byId = computed(() => new Map(props.discrepancies.map(d => [d.id, d])))
const label = (id) => {
  const d = byId.value.get(id)
  if (!d) return id
  return (d.description || '').replace(/^\[[^\]]+\]\s*/, '').slice(0, 90)
}
const workbench = computed(() => `?view=remarks&objectID=${encodeURIComponent(props.objectId || '')}&namespaceID=${encodeURIComponent(props.namespaceId || '')}`)
function toggle (id) { if (open.has(id)) open.delete(id); else open.add(id) }
</script>

<template>
  <div v-if="list.length" class="expertise">
    <h2>
      Замечания экспертизы по этому сравнению ({{ list.length }})
      <a class="small-link" :href="workbench" target="_self">Открыть рабочее место замечаний →</a>
    </h2>
    <div class="disc-summary">
      <span class="chip high"><span class="dot"></span>сохраняется: {{ counts.still_differs }}</span>
      <span class="chip ok"><span class="dot"></span>устранено: {{ counts.resolved }}</span>
      <span class="chip"><span class="dot"></span>нет связи: {{ counts.no_match }}</span>
      <span v-if="counts.claim" class="chip medium"><span class="dot"></span>ответ заявляет правку, а различие осталось: {{ counts.claim }}</span>
    </div>
    <div class="disc-filters">
      <select v-model="filter.verdict">
        <option value="">все замечания</option>
        <option v-for="(l, k) in VERDICT_LABEL" :key="k" :value="k">{{ l }}</option>
      </select>
    </div>
    <div v-for="e in rows" :key="e.id" class="exp" :class="'v-' + e.verdict">
      <div class="exp-head">
        <b class="exp-num">№ {{ e.number }}</b>
        <span class="tag" :class="'ev-' + e.verdict">{{ VERDICT_LABEL[e.verdict] || e.verdict }}</span>
        <span v-if="e.claimMismatch" class="tag ev-claim" title="Ответ проектировщика говорит, что правка внесена, но связанное расхождение в документах сохраняется">ответ ≠ документы</span>
        <span v-if="e.section" class="tag">{{ e.section }}</span>
        <span class="grow"></span>
        <span class="tag">{{ STATUS_LABEL[e.status] || e.status }}</span>
      </div>
      <div class="exp-text">{{ e.text }}</div>
      <div v-if="e.answer" class="exp-answer"><b>Ответ проектировщика:</b> {{ e.answer }}</div>
      <div v-if="e.matchNote" class="exp-note">{{ e.matchNote }}</div>
      <div v-if="e.findings.length" class="exp-links">
        <span class="ev-toggle" @click="toggle(e.id)">{{ open.has(e.id) ? '▾' : '▸' }} связанные находки ({{ e.findings.length }})</span>
        <template v-if="open.has(e.id)">
          <div v-for="id in e.findings" :key="id" class="exp-link" @click="byId.get(id) && emit('jump', byId.get(id))">{{ label(id) }}</div>
        </template>
      </div>
    </div>
  </div>
</template>
