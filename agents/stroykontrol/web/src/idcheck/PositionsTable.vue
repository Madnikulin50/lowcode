<script setup>
// Check 8 in place: per position, Σ min(col 12, col 13) vs КС-2.
import { computed, watch, nextTick } from 'vue'
import { POS_STATUS } from './labels.js'

const props = defineProps({ rows: Array, marks: Object, selected: Object })
const emit = defineEmits(['select'])
const bad = computed(() => props.rows.filter(r => r.status !== 'ok').length)
const isSel = r => props.selected && props.selected.check === 8 && props.selected.position === r.position
watch(() => props.selected, () => nextTick(() => {
  const el = document.querySelector('.ic-pos .is-selected')
  if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center' })
}))
</script>

<template>
  <div class="ic-pos">
    <div class="ic-reg-bar">
      <span v-if="!rows.length" class="muted">КС-2 не загружена или сверка не выполнялась.</span>
      <span v-else-if="!bad" class="ok-text">✓ Все {{ rows.length }} позиций совпадают с КС-2.</span>
      <span v-else class="bad-text">Расхождений: {{ bad }} из {{ rows.length }} позиций.</span>
    </div>
    <div class="ic-table-scroll">
      <table class="ic-table">
        <thead><tr><th>Позиция</th><th>Наименование</th><th>Ед.</th><th>Σ min(12,13)</th><th>Реестр, кол. 11</th><th>КС-2</th><th>Разница</th><th>Статус</th></tr></thead>
        <tbody>
          <tr v-for="r in rows" :key="r.position" :class="[r.status === 'ok' ? '' : 'high marked', { 'is-selected': isSel(r) }]"
            @click="marks[r.position] && emit('select', marks[r.position][0])">
            <td>{{ r.position }}</td>
            <td class="name">{{ r.name }}</td>
            <td>{{ r.unit }}</td>
            <td class="num">{{ r.registrySum }}</td>
            <td class="num">{{ r.registryKS2 }}</td>
            <td class="num">{{ r.ks2 }}</td>
            <td class="num">{{ r.delta && Number(r.delta) !== 0 ? r.delta : '' }}</td>
            <td><span class="pill" :class="r.status">{{ POS_STATUS[r.status] || r.status }}</span></td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
