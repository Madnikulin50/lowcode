<script setup>
// Findings list: severity toggles, check / act filters, text search,
// grouped by where the finding lives (registry, КС-2, each act).
import { ref, computed, watch } from 'vue'
import { SEV_LABEL, SOURCE_LABEL } from './labels.js'

const props = defineProps({
  findings: Array, acts: Array, checkNames: Object, selectedId: String, focusActId: String,
})
const emit = defineEmits(['select', 'visible'])

const sev = ref({ high: true, medium: true, low: false, info: false })
const check = ref('')
const actFilter = ref(props.focusActId || '')
const q = ref('')

const actName = computed(() => Object.fromEntries((props.acts || []).map(a => [a.id, a])))

const filtered = computed(() => props.findings.filter(f =>
  sev.value[f.severity] !== false &&
  (!check.value || String(f.check) === check.value) &&
  (!actFilter.value || f.actID === actFilter.value) &&
  (!q.value || (f.description + ' ' + (f.actual || '') + ' ' + (f.position || '')).toLowerCase().includes(q.value.toLowerCase())),
))

const groups = computed(() => {
  const out = []
  const by = new Map()
  for (const f of filtered.value) {
    let key, title, sub
    if (f.actID) {
      const a = actName.value[f.actID]
      key = 'act:' + f.actID
      title = a ? a.number || a.fileName : 'Акт'
      sub = a ? [a.date, a.positions].filter(Boolean).join(' · ') : ''
    } else if (f.check === 8) {
      key = 'ks2'; title = 'Реестр ↔ КС-2'; sub = 'сверка объёмов по позициям'
    } else {
      key = 'reg'; title = 'Реестр ИД'; sub = 'номера, даты, файлы актов'
    }
    if (!by.has(key)) { by.set(key, { key, title, sub, items: [] }); out.push(by.get(key)) }
    by.get(key).items.push(f)
  }
  return out
})

watch(groups, g => emit('visible', g.flatMap(x => x.items)), { immediate: true })

const sevCount = computed(() => {
  const c = { high: 0, medium: 0, low: 0, info: 0 }
  for (const f of props.findings) if (c[f.severity] != null) c[f.severity]++
  return c
})

function where (f) {
  if (f.actID) return f.page ? `стр. ${f.page}` + (f.box ? '' : ' ⓘ') : ''
  if (f.check === 8) return f.position
  if (f.registryRow) return `стр. реестра ${f.registryRow}`
  return ''
}
</script>

<template>
  <aside class="ic-panel">
    <div class="ic-filters">
      <div class="ic-sev">
        <button v-for="s in ['high', 'medium', 'low', 'info']" :key="s" v-show="sevCount[s] || s !== 'info'"
          class="chip toggle" :class="[s, { off: !sev[s] }]" @click="sev[s] = !sev[s]">
          <span class="dot"></span>{{ SEV_LABEL[s] }} {{ sevCount[s] }}
        </button>
      </div>
      <div class="ic-row">
        <select v-model="check" aria-label="Проверка">
          <option value="">Все проверки</option>
          <option v-for="(name, n) in checkNames" :key="n" :value="String(n)">{{ n }}. {{ name }}</option>
        </select>
        <select v-model="actFilter" aria-label="Акт">
          <option value="">Все акты и реестр</option>
          <option v-for="a in acts" :key="a.id" :value="a.id">{{ a.number || a.fileName }}</option>
        </select>
      </div>
      <input v-model="q" type="search" placeholder="Поиск по тексту замечания…">
    </div>

    <div class="ic-list">
      <div v-if="!filtered.length" class="ic-empty">
        Нет замечаний по выбранным фильтрам.<br>
        <small v-if="!sev.low">Низкая критичность скрыта — это сомнения распознавания.</small>
      </div>
      <section v-for="g in groups" :key="g.key" class="ic-group">
        <div class="ic-group-head"><b>{{ g.title }}</b><span>{{ g.sub }}</span><em>{{ g.items.length }}</em></div>
        <button
          v-for="f in g.items" :key="f.id" :data-id="f.id"
          class="ic-finding" :class="[f.severity, { 'is-active': f.id === selectedId }]"
          @click="emit('select', f)"
        >
          <span class="ic-f-top">
            <span class="sev-dot" :title="SEV_LABEL[f.severity]"></span>
            <span class="ic-check" :title="checkNames[f.check]">{{ f.check }}</span>
            <span class="ic-where">{{ where(f) }}</span>
            <span class="ic-src">{{ SOURCE_LABEL[f.source] || '' }}</span>
          </span>
          <span class="ic-desc">{{ f.description }}</span>
          <span v-if="f.expected || f.actual" class="ic-ea">
            <span v-if="f.expected"><i>ожидалось</i> {{ f.expected }}</span>
            <span v-if="f.actual"><i>в документе</i> {{ f.actual }}</span>
          </span>
        </button>
      </section>
    </div>
  </aside>
</template>
