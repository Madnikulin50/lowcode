<script setup>
// Remarks of the expertise (from the attached Excel) with the designer's
// answer and what the comparison found for each: the difference is still
// there, it was fixed in this iteration, or nothing related was found.
import { computed, reactive } from 'vue'
import { VERDICT_LABEL, EXPERT_STATUS_LABEL } from '../constants.js'

const props = defineProps({ list: Array, discrepancies: Array, busy: Boolean })
const emit = defineEmits(['import', 'status', 'jump'])

const open = reactive(new Set())
const filter = reactive({ verdict: '' })

const counts = computed(() => {
  const c = { still_differs: 0, resolved: 0, no_match: 0, claim: 0 }
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
function toggle (id) { if (open.has(id)) open.delete(id); else open.add(id) }
</script>

<template>
  <div class="expertise">
    <h2>
      Замечания экспертизы ({{ list.length }})
      <button class="small" :disabled="busy" title="Прочитать приложенный Excel и заново сопоставить замечания с расхождениями" @click="emit('import')">Загрузить и сопоставить</button>
    </h2>
    <div v-if="!list.length" class="empty">
      Замечания не загружены: приложите Excel с замечаниями (и, при необходимости, файл с ответами проектировщика)
      в поле «Замечания экспертизы» карточки сравнения и нажмите «Загрузить и сопоставить».
    </div>
    <template v-else>
      <div class="disc-summary">
        <span class="chip high"><span class="dot"></span>сохраняется: {{ counts.still_differs }}</span>
        <span class="chip ok"><span class="dot"></span>устранено: {{ counts.resolved }}</span>
        <span class="chip"><span class="dot"></span>нет связи: {{ counts.no_match }}</span>
        <span v-if="counts.claim" class="chip medium"><span class="dot"></span>ответ заявляет правку, а различие осталось: {{ counts.claim }}</span>
      </div>
      <div class="disc-filters">
        <select v-model="filter.verdict">
          <option value="">все замечания</option>
          <option v-for="(label, k) in VERDICT_LABEL" :key="k" :value="k">{{ label }}</option>
        </select>
      </div>
      <div v-for="e in rows" :key="e.id" class="exp" :class="'v-' + e.verdict">
        <div class="exp-head">
          <b class="exp-num">№ {{ e.number }}</b>
          <span class="tag" :class="'ev-' + e.verdict">{{ VERDICT_LABEL[e.verdict] || e.verdict }}</span>
          <span v-if="e.claimMismatch" class="tag ev-claim" title="Ответ проектировщика говорит, что правка внесена, но связанное расхождение в документах сохраняется">ответ ≠ документы</span>
          <span v-if="e.location" class="tag">{{ e.location }}</span>
          <span class="grow"></span>
          <select :value="e.status" :disabled="busy" @change="emit('status', e.id, $event.target.value)">
            <option v-for="(label, k) in EXPERT_STATUS_LABEL" :key="k" :value="k">{{ label }}</option>
          </select>
        </div>
        <div class="exp-text">{{ e.text }}</div>
        <div v-if="e.answer" class="exp-answer"><b>Ответ проектировщика:</b> {{ e.answer }}</div>
        <div v-if="e.note" class="exp-note">{{ e.note }}</div>
        <div v-if="e.findings.length" class="exp-links">
          <span class="ev-toggle" @click="toggle(e.id)">{{ open.has(e.id) ? '▾' : '▸' }} связанные находки ({{ e.findings.length }})</span>
          <template v-if="open.has(e.id)">
            <div v-for="id in e.findings" :key="id" class="exp-link" @click="byId.get(id) && emit('jump', byId.get(id))">{{ label(id) }}</div>
          </template>
        </div>
      </div>
    </template>
  </div>
</template>
