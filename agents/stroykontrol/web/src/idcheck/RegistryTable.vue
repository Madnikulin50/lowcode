<script setup>
// The registry as a table with the rows that have findings marked in place.
// By default only rows near a finding are shown (± context), so the
// relevant row is visible without scrolling through hundreds of rows.
import { computed, ref, watch, nextTick } from 'vue'
import { SEV_LABEL, SEV_ORDER } from './labels.js'

const props = defineProps({ rows: Array, marks: Object, selected: Object })
const emit = defineEmits(['select'])
const all = ref(false)
const CONTEXT = 3

const worst = r => (props.marks[r.row] || []).slice().sort((a, b) => SEV_ORDER[a.severity] - SEV_ORDER[b.severity])[0]
const shown = computed(() => {
  if (all.value) return props.rows
  const near = new Set()
  const idx = props.rows.map(r => r.row)
  props.rows.forEach((r, i) => {
    if (props.marks[r.row]) for (let k = Math.max(0, i - CONTEXT); k <= Math.min(idx.length - 1, i + CONTEXT); k++) near.add(k)
  })
  const out = []
  let last = -2
  props.rows.forEach((r, i) => {
    if (!near.has(i)) return
    if (i !== last + 1 && out.length) out.push({ gap: true, key: 'gap' + i })
    out.push(r)
    last = i
  })
  return out
})
const isSel = r => props.selected && !props.selected.actID && props.selected.registryRow === r.row
watch(() => props.selected, () => nextTick(() => {
  const el = document.querySelector('.ic-reg .is-selected')
  if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center' })
}))
</script>

<template>
  <div class="ic-reg">
    <div class="ic-reg-bar">
      <label><input v-model="all" type="checkbox"> показать весь реестр ({{ rows.length }} строк)</label>
      <span class="muted">Строки с замечаниями подсвечены; кол. 11 — КС-2, 12 — проект, 13 — факт по АОСР.</span>
    </div>
    <div class="ic-table-scroll">
      <table class="ic-table">
        <thead><tr><th>Стр.</th><th>Позиция</th><th>Наименование</th><th>№ документа</th><th>Дата</th><th>Ед.</th><th>11</th><th>12</th><th>13</th></tr></thead>
        <tbody>
          <template v-for="r in shown" :key="r.key || r.row">
            <tr v-if="r.gap" class="gap"><td colspan="9">⋯</td></tr>
            <tr v-else :class="[r.kind, worst(r)?.severity, { marked: marks[r.row], 'is-selected': isSel(r) }]"
              @click="marks[r.row] && emit('select', marks[r.row][0])">
              <td>{{ r.row }}</td>
              <td>{{ r.position }}</td>
              <td class="name">
                {{ r.name }}
                <div v-for="f in marks[r.row] || []" :key="f.id" class="ic-inline" :class="f.severity">
                  ● {{ SEV_LABEL[f.severity] }}: {{ f.description }}
                </div>
              </td>
              <td class="mono">{{ r.docNumber }}</td>
              <td>{{ r.docDate }}</td>
              <td>{{ r.unit }}</td>
              <td class="num">{{ r.qtyKS2 }}</td>
              <td class="num">{{ r.qtyProject }}</td>
              <td class="num">{{ r.qtyFact }}</td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>
  </div>
</template>
