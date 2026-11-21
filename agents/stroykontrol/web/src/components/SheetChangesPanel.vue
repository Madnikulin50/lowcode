<script setup>
// «Изменения листов» (ТЗ 3.3.7, 3.3.8): при сравнении двух редакций тома — что с каждым листом стало
// на самом деле (по расхождениям сравнения) и что заявлено в таблице регистрации изменений и в графах
// изменений основной надписи. Строка с расхождением между ними — повод для находки «регистрация изменений».
import { reactive, ref, computed, watch, onMounted } from 'vue'
import { errText } from '../utils.js'

const props = defineProps({ qs: String, running: String, status: String })
const emit = defineEmits(['jump'])

const STATE = { same: 'без изменений', changed: 'изменён', replaced: 'заменён', new: 'новый', annulled: 'аннулирован', absent: 'нет в томе' }
const state = reactive({ error: '', list: [] })
const open = ref(true)
const onlyIssues = ref(false)

async function load () {
  try {
    const res = await fetch(`api/pairing?${props.qs}`)
    const data = await res.json().catch(() => ({}))
    if (!res.ok || data.error) throw new Error(errText(data.error, res.status) || res.statusText)
    state.list = data.state?.sheetChanges || []
    state.error = ''
  } catch (e) {
    state.error = e.message
  }
}
onMounted(load)
watch(() => [props.running, props.status], load)

const issues = computed(() => state.list.filter(s => !s.ok).length)
const rows = computed(() => onlyIssues.value ? state.list.filter(s => !s.ok) : state.list)
const change = computed(() => state.list[0]?.change || 0)
const label = (s) => STATE[s] || '—'
const go = (s) => emit('jump', { docRef: s.refDoc, docTarget: s.targetDoc, pageRef: s.pageRef, pageTarget: s.pageTarget })
</script>

<template>
  <div v-if="state.list.length || state.error" class="sheets">
    <h2 class="sheets-h" @click="open = !open">
      <span class="caret">{{ open ? '▾' : '▸' }}</span> Изменения листов
      <span class="muted">изм. {{ change }} · листов: {{ state.list.length }}<template v-if="issues"> · расхождений с регистрацией: <b class="bad">{{ issues }}</b></template><template v-else> · всё заявлено верно</template></span>
    </h2>
    <template v-if="open">
      <div v-if="state.error" class="err">{{ state.error }}</div>
      <label v-if="issues" class="only"><input v-model="onlyIssues" type="checkbox"> только с расхождениями</label>
      <table>
        <thead><tr><th>Лист</th><th>Фактически</th><th>По таблице регистрации</th><th>По основной надписи</th><th>Разл.</th><th>Проверка</th></tr></thead>
        <tbody>
          <tr v-for="s in rows" :key="s.volume + s.sheet" :class="{ issue: !s.ok }" @click="go(s)">
            <td>{{ s.sheet }}</td>
            <td><span :class="['tag', 'st-' + s.fact]">{{ label(s.fact) }}</span></td>
            <td>{{ label(s.declared) }}<span v-if="s.permitted === false" class="bad" title="Лист не назван в разрешении на внесение изменений"> · нет в разрешении</span></td>
            <td>{{ label(s.stamp) }}</td>
            <td>{{ s.diffs || '' }}</td>
            <td>{{ s.ok ? '✓' : s.note }}</td>
          </tr>
        </tbody>
      </table>
      <div class="muted">«Фактически» — по результату сравнения: лист с расхождениями изменён, отсутствовавший — новый, исчезнувший — аннулирован. Нажмите строку, чтобы открыть лист.</div>
    </template>
  </div>
</template>

<style scoped>
.sheets { margin-top: 16px; }
.sheets-h { font-size: 13px; margin: 0 0 6px; cursor: pointer; display: flex; flex-wrap: wrap; gap: 4px 12px; align-items: baseline; }
.muted { font-size: 11.5px; font-weight: 400; color: var(--muted); }
.caret { width: 12px; display: inline-block; }
.only { font-size: 12px; display: inline-flex; gap: 4px; align-items: center; margin-bottom: 4px; }
table { width: 100%; border-collapse: collapse; font-size: 12px; }
th, td { text-align: left; padding: 3px 6px; border-bottom: 1px solid var(--border); vertical-align: top; }
th { font-weight: 600; color: var(--muted); }
tbody tr { cursor: pointer; }
tbody tr:hover { background: var(--panel); }
tr.issue td:last-child { color: var(--bad, #c0392b); }
.bad { color: var(--bad, #c0392b); }
.tag { display: inline-block; padding: 0 7px; border-radius: var(--radius-lg); font-size: 11px; border: 1px solid var(--border); background: var(--panel); white-space: nowrap; }
.tag.st-changed, .tag.st-new { border-color: var(--warn); color: var(--warn); }
.tag.st-annulled { border-color: var(--muted); color: var(--muted); text-decoration: line-through; }
</style>
