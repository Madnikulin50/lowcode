<script setup>
// One act's scan with the findings drawn in place: a page strip (kind of
// each page + how many findings sit on it) and the page itself with a box
// per located finding. Clicking a box selects the finding in the list.
import { computed, ref, watch, nextTick } from 'vue'
import { SEV_LABEL, PAGE_KIND } from './labels.js'

const props = defineProps({
  acts: Array, act: Object, actId: String, page: Number, findings: Array, selected: Object, ns: String,
})
const emit = defineEmits(['update:actId', 'update:page', 'select'])

const zoom = ref(100)
const pageCount = computed(() => props.act?.pages?.length || 0)
const src = n => `api/idcheck/page?namespaceID=${encodeURIComponent(props.ns)}&actID=${encodeURIComponent(props.actId)}&n=${n}`

const onPage = computed(() => props.findings.filter(f => (f.page || 1) === props.page))
// Findings on the same spot (e.g. ИНН by checksum and by ЕГРЮЛ) share one
// box labelled "1,2"; clicking it cycles through them.
const boxed = computed(() => {
  const groups = []
  let n = 0
  for (const f of onPage.value.filter(f => f.box)) {
    n++
    const g = groups.find(g => g.box.every((v, i) => Math.abs(v - f.box[i]) < 0.006))
    if (g) { g.items.push(f); g.labels.push(n) } else groups.push({ box: f.box, items: [f], labels: [n] })
  }
  for (const g of groups) {
    g.severity = ['high', 'medium', 'low'].find(s => g.items.some(f => f.severity === s)) || 'low'
    g.id = g.items[0].id
  }
  return groups
})
const isSelGroup = g => props.selected && g.items.some(f => f.id === props.selected.id)
function clickGroup (g) {
  const i = g.items.findIndex(f => props.selected && f.id === props.selected.id)
  emit('select', g.items[(i + 1) % g.items.length])
}
const perPage = computed(() => {
  const m = {}
  for (const f of props.findings) {
    const p = f.page || 1
    m[p] ||= { high: 0, medium: 0, low: 0 }
    if (m[p][f.severity] != null) m[p][f.severity]++
  }
  return m
})
const selectedHere = computed(() => props.selected && props.selected.actID === props.actId && (props.selected.page || 1) === props.page)
const unlocated = computed(() => selectedHere.value && !props.selected.box)

const actIdx = computed(() => props.acts.findIndex(a => a.id === props.actId))
function goAct (d) {
  const a = props.acts[actIdx.value + d]
  if (a) { emit('update:actId', a.id); emit('update:page', 1) }
}
function boxStyle (f) {
  const [x, y, w, h] = f.box
  const pad = 0.004
  return {
    left: (x - pad) * 100 + '%', top: (y - pad) * 100 + '%',
    width: (w + 2 * pad) * 100 + '%', height: (h + 2 * pad) * 100 + '%',
  }
}
const loading = ref(true)
watch(() => [props.actId, props.page], () => { loading.value = true })

// keep the selected box in view after the page image loads
function onLoad () {
  loading.value = false
  nextTick(() => {
    const el = document.querySelector('.ic-page .is-selected')
    if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center', inline: 'center' })
  })
}
</script>

<template>
  <div class="ic-act">
    <div class="ic-act-bar">
      <button :disabled="actIdx <= 0" title="Предыдущий акт" @click="goAct(-1)">‹</button>
      <select :value="actId" aria-label="Акт" @change="e => { emit('update:actId', e.target.value); emit('update:page', 1) }">
        <option v-for="a in acts" :key="a.id" :value="a.id">
          {{ a.high ? '● ' : a.medium ? '◐ ' : '○ ' }}{{ a.number || a.fileName }} {{ a.date ? 'от ' + a.date : '' }}
        </option>
      </select>
      <button :disabled="actIdx >= acts.length - 1" title="Следующий акт" @click="goAct(1)">›</button>
      <span v-if="act" class="ic-act-meta">
        <span v-if="act.high" class="chip high"><span class="dot"></span>{{ act.high }}</span>
        <span v-if="act.medium" class="chip medium"><span class="dot"></span>{{ act.medium }}</span>
        <span v-if="act.low" class="chip low"><span class="dot"></span>{{ act.low }}</span>
        <span class="muted">{{ act.fileName }}</span>
      </span>
      <span class="zoom">
        <button :disabled="zoom <= 50" @click="zoom -= 25">−</button>
        <span class="zlevel">{{ zoom }}%</span>
        <button :disabled="zoom >= 250" @click="zoom += 25">+</button>
      </span>
    </div>

    <div v-if="!pageCount" class="ic-empty">Скан акта ещё не обработан.</div>
    <div v-else class="ic-act-body">
      <ol class="ic-strip" aria-label="Страницы">
        <li v-for="p in act.pages" :key="p.n">
          <button :class="{ active: p.n === page }" @click="emit('update:page', p.n)">
            <span class="n">{{ p.n }}</span>
            <span class="k">{{ PAGE_KIND[p.kind] || p.kind }}</span>
            <span v-if="perPage[p.n]" class="marks">
              <i v-if="perPage[p.n].high" class="high">{{ perPage[p.n].high }}</i>
              <i v-if="perPage[p.n].medium" class="medium">{{ perPage[p.n].medium }}</i>
              <i v-if="perPage[p.n].low" class="low">{{ perPage[p.n].low }}</i>
            </span>
          </button>
        </li>
      </ol>

      <div class="ic-page-wrap">
        <div v-if="unlocated" class="ic-banner">
          Точное место на странице не определено (значение не найдено в тексте скана) — замечание относится к стр. {{ page }}.
        </div>
        <div class="ic-page-scroll">
          <div class="ic-page" :class="{ 'page-flag': unlocated }" :style="{ width: zoom + '%' }">
            <img :src="src(page)" :alt="'Страница ' + page" @load="onLoad">
            <div v-if="loading" class="ic-loading">Загрузка страницы…</div>
            <button
              v-for="g in boxed" :key="g.id"
              class="ic-box" :class="[g.severity, { 'is-selected': isSelGroup(g) }]"
              :style="boxStyle(g)" :title="g.items.map(f => SEV_LABEL[f.severity] + ': ' + f.description).join('\n')"
              @click="clickGroup(g)"
            ><span class="ic-box-n">{{ g.labels.join(',') }}</span></button>
          </div>
        </div>
        <div v-if="selectedHere" class="ic-callout" :class="selected.severity">
          <b>{{ SEV_LABEL[selected.severity] }} · проверка {{ selected.check }}</b>
          <p>{{ selected.description }}</p>
          <p v-if="selected.expected || selected.actual" class="ic-ea">
            <span v-if="selected.expected"><i>ожидалось</i> {{ selected.expected }}</span>
            <span v-if="selected.actual"><i>в документе</i> {{ selected.actual }}</span>
          </p>
        </div>
      </div>
    </div>
  </div>
</template>
