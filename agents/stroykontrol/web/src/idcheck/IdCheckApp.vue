<script setup>
// In-place findings viewer for the «Проверка ИД (АОСР)» namespace: the
// findings list on the left, and on the right the place each finding refers
// to — the scanned act page with the spot boxed, the registry row, or the
// registry ↔ КС-2 position. Embedded via the IFrame block on the package
// card (?packageID=) and the act card (?actID=).
import { reactive, ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import FindingsPanel from './FindingsPanel.vue'
import ActStage from './ActStage.vue'
import RegistryTable from './RegistryTable.vue'
import PositionsTable from './PositionsTable.vue'
import { SEV_LABEL, STATUS_LABEL } from './labels.js'

const params = new URLSearchParams(location.search)
const ns = params.get('namespaceID') || ''
const qs = new URLSearchParams({ namespaceID: ns })
if (params.get('packageID')) qs.set('packageID', params.get('packageID'))
if (params.get('actID')) qs.set('actID', params.get('actID'))
if (params.get('findingID') && !params.get('packageID') && !params.get('actID')) qs.set('findingID', params.get('findingID'))

const state = reactive({
  loading: true,
  error: '',
  view: null,
  mode: 'act', // act | registry | positions
  actID: '',
  page: 1,
  selectedID: '',
})

let timer = null
async function load (quiet = false) {
  if (!quiet) state.loading = true
  try {
    const res = await fetch(`api/idcheck/view?${qs}`)
    const data = await res.json().catch(() => ({}))
    if (!res.ok || data.error) throw new Error(data.error || res.statusText)
    state.view = data
    state.error = ''
    if (!state.actID) {
      state.actID = data.focusActID || worstAct(data)?.id || data.acts?.[0]?.id || ''
      if (!data.acts?.length) state.mode = 'registry'
      // deep link: ?findingID= opens straight at that finding's place
      const target = params.get('findingID') && data.findings?.find(f => f.id === params.get('findingID'))
      if (target) nextTick(() => select(target))
      else if (params.get('mode')) state.mode = params.get('mode')
    }
  } catch (e) {
    state.error = e.message
  } finally {
    state.loading = false
  }
  clearTimeout(timer)
  if (state.view?.status === 'processing') timer = setTimeout(() => load(true), 10000)
}
onMounted(load)
onUnmounted(() => clearTimeout(timer))

function worstAct (v) {
  return (v.acts || []).slice().sort((a, b) => (b.high * 100 + b.medium) - (a.high * 100 + a.medium))[0]
}

const findings = computed(() => state.view?.findings || [])
const act = computed(() => (state.view?.acts || []).find(a => a.id === state.actID) || null)
const actFindings = computed(() => findings.value.filter(f => f.actID === state.actID))
const selected = computed(() => findings.value.find(f => f.id === state.selectedID) || null)
const registryMarks = computed(() => {
  const m = {}
  for (const f of findings.value) if (f.registryRow && !f.actID) (m[f.registryRow] ||= []).push(f)
  return m
})
const positionMarks = computed(() => {
  const m = {}
  for (const f of findings.value) if (f.check === 8 && f.position) (m[f.position] ||= []).push(f)
  return m
})
const counts = computed(() => {
  const c = { high: 0, medium: 0, low: 0 }
  for (const f of findings.value) if (c[f.severity] != null) c[f.severity]++
  return c
})

// Where a finding lives: its act page, the registry row, or the КС-2 row.
function placeOf (f) {
  if (f.actID) return 'act'
  if (f.check === 8) return 'positions'
  if (f.registryRow) return 'registry'
  return ''
}

function select (f, { fromStage = false } = {}) {
  state.selectedID = f.id
  const place = placeOf(f)
  if (place) state.mode = place
  if (place === 'act') {
    state.actID = f.actID
    state.page = f.page || state.page || 1
  }
  if (!fromStage) {
    nextTick(() => {
      const el = document.querySelector('.ic-stage .is-selected')
      if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center', inline: 'center' })
    })
  } else {
    nextTick(() => {
      const el = document.querySelector(`.ic-finding[data-id="${f.id}"]`)
      if (el) el.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
    })
  }
}

function step (dir) {
  const list = visibleOrder.value
  if (!list.length) return
  const i = list.findIndex(f => f.id === state.selectedID)
  const next = list[(i + dir + list.length) % list.length]
  select(next)
}
const visibleOrder = ref([])
function onVisible (list) { visibleOrder.value = list }

function onKey (e) {
  if (e.target.closest('input, select, textarea')) return
  if (e.key === 'j' || e.key === 'ArrowDown') { step(1); e.preventDefault() }
  if (e.key === 'k' || e.key === 'ArrowUp') { step(-1); e.preventDefault() }
}
onMounted(() => window.addEventListener('keydown', onKey))
onUnmounted(() => window.removeEventListener('keydown', onKey))

watch(() => state.actID, () => {
  if (selected.value?.actID !== state.actID) state.page = 1
})
</script>

<template>
  <div class="ic">
    <div v-if="state.loading && !state.view" class="ic-msg">Загрузка…</div>
    <div v-else-if="state.error && !state.view" class="ic-msg err">{{ state.error }}</div>
    <template v-else-if="state.view">
      <header class="ic-head">
        <h1>{{ state.view.title }}</h1>
        <span class="badge" :class="{ ok: state.view.status === 'done', warn: state.view.status === 'processing', bad: state.view.status === 'failed' }">
          {{ STATUS_LABEL[state.view.status] || state.view.status || 'новый' }}
        </span>
        <span v-for="sev in ['high', 'medium', 'low']" :key="sev" class="chip" :class="sev">
          <span class="dot"></span>{{ SEV_LABEL[sev] }}: {{ counts[sev] }}
        </span>
        <span class="ic-hint">↑/↓ или j/k — следующее замечание</span>
      </header>
      <div v-if="state.view.status === 'processing'" class="ic-progress">⏳ {{ state.view.progress }}</div>

      <div class="ic-body">
        <FindingsPanel
          :findings="findings" :acts="state.view.acts" :check-names="state.view.checkNames"
          :selected-id="state.selectedID" :focus-act-id="state.view.focusActID"
          @select="select" @visible="onVisible"
        />

        <section class="ic-stage">
          <nav class="ic-tabs">
            <button :class="{ active: state.mode === 'act' }" :disabled="!state.view.acts?.length" @click="state.mode = 'act'">
              Скан акта <small v-if="state.view.acts?.length">{{ state.view.acts.length }}</small>
            </button>
            <button :class="{ active: state.mode === 'registry' }" @click="state.mode = 'registry'">
              Реестр <small v-if="Object.keys(registryMarks).length">{{ Object.keys(registryMarks).length }} ⚠</small>
            </button>
            <button :class="{ active: state.mode === 'positions' }" @click="state.mode = 'positions'">
              Реестр ↔ КС-2 <small v-if="Object.keys(positionMarks).length">{{ Object.keys(positionMarks).length }} ⚠</small>
            </button>
          </nav>

          <ActStage
            v-if="state.mode === 'act'"
            v-model:act-id="state.actID" v-model:page="state.page"
            :acts="state.view.acts" :act="act" :findings="actFindings" :selected="selected" :ns="ns"
            @select="f => select(f, { fromStage: true })"
          />
          <RegistryTable
            v-else-if="state.mode === 'registry'"
            :rows="state.view.registry" :marks="registryMarks" :selected="selected"
            @select="f => select(f, { fromStage: true })"
          />
          <PositionsTable
            v-else
            :rows="state.view.positions" :marks="positionMarks" :selected="selected"
            @select="f => select(f, { fromStage: true })"
          />
        </section>
      </div>
    </template>
  </div>
</template>

<style src="./idcheck.css"></style>
