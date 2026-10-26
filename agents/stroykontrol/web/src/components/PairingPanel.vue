<script setup>
import { errText } from '../utils.js'
// «Сопоставление»: какие тома и фрагменты с чем сравнивались, и ручная правка — когда автоподбор
// ошибся (том ПД с несколькими сооружениями против тома РД по одному из них). Указанные пары
// сравниваются как заданы, остальное подбирается автоматически; решение остаётся за специалистом.
import { reactive, ref, computed, watch, onMounted } from 'vue'
import { names, nm } from '../constants.js'

const props = defineProps({ qs: String, recordId: String, namespaceId: String, running: String, status: String })
const emit = defineEmits(['apply'])

const state = reactive({ loading: true, error: '', busy: false, data: null })
const draft = ref([]) // manual pairs being edited
const exDraft = ref([]) // areas left out of the comparison, being edited
const open = ref(false)
const slot = reactive({ ref: { doc: '', from: 1, to: 0 }, targets: [{ doc: '', from: 1, to: 0 }] })

async function api (url, body) {
  const res = await fetch(url, body ? { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) } : undefined)
  const data = await res.json().catch(() => ({}))
  if (!res.ok || data.error) throw new Error(errText(data.error, res.status) || res.statusText)
  return data
}

async function load () {
  try {
    state.data = await api(`api/pairing?${props.qs}`)
    draft.value = JSON.parse(JSON.stringify(state.data.pairing || []))
    exDraft.value = JSON.parse(JSON.stringify(state.data.excluded || []))
    state.error = ''
  } catch (e) {
    state.error = e.message
  } finally {
    state.loading = false
  }
}
onMounted(load)
// the comparison finished (or started) — look at the new pairs
watch(() => [props.running, props.status], load)

const docs = computed(() => (state.data?.documents || []))
const byID = computed(() => Object.fromEntries(docs.value.map(d => [d.id, d])))
const docName = (id) => byID.value[id]?.name || id
const refDocs = computed(() => docs.value.filter(d => d.side === 'ref'))
const tgtDocs = computed(() => docs.value.filter(d => d.side === 'target'))
const range = (v) => `стр. ${v.from}–${v.to}`
const volLabel = (v) => `${docName(v.doc)}${v.title ? ' — ' + v.title : ''}${(v.title || '').includes('стр.') ? '' : ' (' + range(v) + ')'}${v.change ? ' · изм. ' + v.change : ''}`
const sliceLabel = (s) => `${docName(s.doc)}, стр. ${s.from || 1}–${s.to || 'конец'}`

const pairs = computed(() => state.data?.state?.pairs || [])
const refOnly = computed(() => state.data?.state?.refOnly || [])
const targetOnly = computed(() => state.data?.state?.targetOnly || [])
const volumes = computed(() => state.data?.state?.volumes || {})
const dirty = computed(() => JSON.stringify(draft.value) !== JSON.stringify(state.data?.pairing || []))

// Excluded areas: a rectangle of a page (percent of its width/height) whose text is not compared and whose
// drawing differences are not reported. The files themselves are never changed.
const exForm = reactive({ side: 'both', doc: '', page: 0, x0: 0, y0: 0, x1: 100, y1: 10, label: '' })
const exDirty = computed(() => JSON.stringify(exDraft.value) !== JSON.stringify(state.data?.excluded || []))
const exValid = computed(() => exForm.x1 > exForm.x0 && exForm.y1 > exForm.y0 && exForm.x0 >= 0 && exForm.y0 >= 0 && exForm.x1 <= 100 && exForm.y1 <= 100)
const frac = (v) => (Number(v) / 100).toFixed(4).replace(/0+$/, '').replace(/\.$/, '')
function addExcluded () {
  exDraft.value.push({
    side: exForm.side, doc: exForm.doc, page: Number(exForm.page) || 0,
    box: `[${frac(exForm.x0)},${frac(exForm.y0)},${frac(exForm.x1)},${frac(exForm.y1)}]`, label: exForm.label.trim(),
  })
}
const removeExcluded = (i) => exDraft.value.splice(i, 1)
const SIDE_NAME = computed(() => ({ both: 'обе стороны', ref: names.ref, target: names.target }))
const pct = (b) => (JSON.parse(b) || []).map(x => Math.round(x * 100)).join('/')
function exLabel (r) {
  const where = r.doc ? docName(r.doc) : 'любой файл'
  return `${SIDE_NAME.value[r.side || 'both']} · ${where} · ${r.page ? 'стр. ' + r.page : 'все страницы'} · область ${pct(r.box)} %${r.label ? ' — ' + r.label : ''}`
}
async function applyExcluded () {
  state.busy = true
  state.error = ''
  try {
    await api('api/exclusions', { recordID: props.recordId, namespaceID: props.namespaceId, excluded: exDraft.value, run: false })
    await load()
    emit('apply')
  } catch (e) {
    state.error = e.message
  } finally {
    state.busy = false
  }
}

// take a volume found in a file as a slot of the form
function useVolume (v, side) {
  if (side === 'ref') Object.assign(slot.ref, { doc: v.doc, from: v.from, to: v.to })
  else {
    const t = slot.targets.find(x => !x.doc) || slot.targets[slot.targets.length - 1]
    Object.assign(t, { doc: v.doc, from: v.from, to: v.to })
  }
}
const addTarget = () => slot.targets.push({ doc: '', from: 1, to: 0 })
const canAdd = computed(() => slot.ref.doc && slot.targets.some(t => t.doc))
function addPair () {
  draft.value.push({
    ref: { doc: slot.ref.doc, from: Number(slot.ref.from) || 1, to: Number(slot.ref.to) || 0 },
    targets: slot.targets.filter(t => t.doc).map(t => ({ doc: t.doc, from: Number(t.from) || 1, to: Number(t.to) || 0 })),
  })
  Object.assign(slot.ref, { doc: '', from: 1, to: 0 })
  slot.targets.splice(0, slot.targets.length, { doc: '', from: 1, to: 0 })
}
const removePair = (i) => draft.value.splice(i, 1)

async function apply () {
  state.busy = true
  state.error = ''
  try {
    await api('api/pairing', { recordID: props.recordId, namespaceID: props.namespaceId, pairing: draft.value, run: false })
    await load()
    emit('apply') // the viewer starts the comparison again with the new pairing
  } catch (e) {
    state.error = e.message
  } finally {
    state.busy = false
  }
}
</script>

<template>
  <div class="pairing">
    <h2 @click="open = !open" class="pairing-h">
      <span class="caret">{{ open ? '▾' : '▸' }}</span> Сопоставление томов и фрагментов
      <span class="muted">
        {{ pairs.length }} пар<template v-if="pairs.some(p => p.manual)"> · вручную: {{ pairs.filter(p => p.manual).length }}</template>
        <template v-if="refOnly.length || targetOnly.length"> · без пары: {{ refOnly.length + targetOnly.length }}</template>
      </span>
    </h2>

    <template v-if="open">
      <div v-if="state.loading">Загрузка…</div>
      <div v-else-if="state.error" class="err">{{ state.error }}</div>
      <template v-else>
        <div v-if="!state.data.state" class="empty">Сравнение ещё не запускалось — после первого запуска здесь появятся тома и пары.</div>

        <h3>Как сопоставлено</h3>
        <div v-for="(p, i) in pairs" :key="'p' + i" class="pr-row">
          <span :class="['tag', p.manual ? 'manual' : 'auto']">{{ p.manual ? 'вручную' : 'авто' }}</span>
          <span v-if="p.fragment" class="tag frag" :title="nm('Проверяемые тома покрывают только часть эталонного тома')">фрагмент</span>
          <b>{{ names.ref }}:</b> {{ volLabel(p.ref) }}
          <span class="arrow">↔</span>
          <b>{{ names.target }}:</b> <template v-for="(t, j) in p.targets" :key="j">{{ j ? '; ' : '' }}{{ volLabel(t) }}</template>
        </div>
        <div v-if="!pairs.length && state.data.state" class="empty">Ни один том не сопоставлен.</div>
        <div v-if="refOnly.length || targetOnly.length" class="pr-rest">
          <div v-for="(v, i) in refOnly" :key="'ro' + i"><span class="tag">{{ names.ref }} без пары</span> {{ volLabel(v) }}</div>
          <div v-for="(v, i) in targetOnly" :key="'to' + i"><span class="tag">{{ names.target }} без пары</span> {{ volLabel(v) }}</div>
          <div class="muted">Это не ошибка: том мог быть ещё не выпущен или лежит вне загруженного объёма. Если пара есть — свяжите вручную ниже.</div>
        </div>

        <h3>Тома в файлах <span class="muted">(нажмите, чтобы подставить в форму)</span></h3>
        <div v-for="d in docs" :key="d.id" class="pr-doc">
          <b>{{ d.side === 'ref' ? names.ref : names.target }}</b> {{ d.name }}
          <template v-for="(v, i) in (volumes[d.id] || [])" :key="i">
            <button class="small vol" @click="useVolume(v, d.side)">{{ v.title || v.code || 'том' }} · {{ range(v) }}</button>
          </template>
        </div>

        <h3>Ручные пары</h3>
        <div v-for="(p, i) in draft" :key="'d' + i" class="pr-row">
          <b>{{ names.ref }}:</b> {{ sliceLabel(p.ref) }} <span class="arrow">↔</span>
          <b>{{ names.target }}:</b> <template v-for="(t, j) in p.targets" :key="j">{{ j ? '; ' : '' }}{{ sliceLabel(t) }}</template>
          <button class="small" @click="removePair(i)">убрать</button>
        </div>
        <div v-if="!draft.length" class="empty">Ручных пар нет: всё подбирается автоматически.</div>

        <div class="pr-form">
          <div class="pr-slot">
            <b>{{ names.ref }} (эталон)</b>
            <select v-model="slot.ref.doc"><option value="">— файл —</option><option v-for="d in refDocs" :key="d.id" :value="d.id">{{ d.name }}</option></select>
            стр. <input v-model.number="slot.ref.from" type="number" min="1"> — <input v-model.number="slot.ref.to" type="number" min="0" placeholder="конец">
          </div>
          <div v-for="(t, i) in slot.targets" :key="i" class="pr-slot">
            <b>{{ names.target }}</b>
            <select v-model="t.doc"><option value="">— файл —</option><option v-for="d in tgtDocs" :key="d.id" :value="d.id">{{ d.name }}</option></select>
            стр. <input v-model.number="t.from" type="number" min="1"> — <input v-model.number="t.to" type="number" min="0" placeholder="конец">
          </div>
          <div class="pr-actions">
            <button class="small" @click="addTarget">+ ещё том {{ names.target }}</button>
            <button class="small" :disabled="!canAdd" @click="addPair">Добавить пару</button>
            <span class="grow"></span>
            <button class="small primary" :disabled="state.busy || !dirty || !!running" @click="apply">Применить и пересчитать</button>
          </div>
        </div>
        <div v-if="state.error" class="err">{{ state.error }}</div>
        <div class="muted">«конец» — до последней страницы файла. Пересчёт заново сопоставит тома и сохранит ваши решения по находкам, которые остались.</div>

        <h3>Исключаемые области <span class="muted">(штамп согласования, облако изменений — то, что сравнивать не нужно)</span></h3>
        <div v-for="(r, i) in exDraft" :key="'x' + i" class="pr-row">
          {{ exLabel(r) }} <button class="small" @click="removeExcluded(i)">убрать</button>
        </div>
        <div v-if="!exDraft.length" class="empty">Областей нет: сравнивается всё, кроме основной надписи.</div>
        <div class="pr-form">
          <div class="pr-slot">
            <select v-model="exForm.side"><option value="both">обе стороны</option><option value="ref">{{ names.ref }}</option><option value="target">{{ names.target }}</option></select>
            <select v-model="exForm.doc"><option value="">любой файл</option><option v-for="d in docs" :key="d.id" :value="d.id">{{ d.name }}</option></select>
            стр. <input v-model.number="exForm.page" type="number" min="0" title="0 — все страницы"> <span class="muted">(0 — все)</span>
          </div>
          <div class="pr-slot">
            область, % страницы: слева <input v-model.number="exForm.x0" type="number" min="0" max="100"> сверху <input v-model.number="exForm.y0" type="number" min="0" max="100">
            справа <input v-model.number="exForm.x1" type="number" min="0" max="100"> снизу <input v-model.number="exForm.y1" type="number" min="0" max="100">
            <input v-model="exForm.label" type="text" placeholder="пояснение" class="wide">
          </div>
          <div class="pr-actions">
            <button class="small" :disabled="!exValid" @click="addExcluded">Добавить область</button>
            <span class="grow"></span>
            <button class="small primary" :disabled="state.busy || !exDirty || !!running" @click="applyExcluded">Применить и пересчитать</button>
          </div>
        </div>
        <div class="muted">Строки текста внутри области в сравнение не попадают, расхождения чертежей, чей центр в области, не фиксируются; файлы не меняются. Основная надпись исключена всегда.</div>
      </template>
    </template>
  </div>
</template>

<style scoped>
.pairing { margin-top: 16px; }
.pairing-h { font-size: 13px; margin: 0 0 6px; cursor: pointer; display: flex; flex-wrap: wrap; gap: 4px 12px; align-items: baseline; }
.pairing-h .muted, .muted { font-size: 11.5px; font-weight: 400; color: var(--muted); }
.pairing h3 { font-size: 12px; margin: 10px 0 4px; }
.caret { width: 12px; display: inline-block; }
.pr-row, .pr-rest > div, .pr-doc { font-size: 12px; padding: 3px 0; border-bottom: 1px solid var(--border); }
.arrow { color: var(--muted); margin: 0 4px; }
.tag { display: inline-block; padding: 0 7px; border-radius: var(--radius-lg); font-size: 11px; border: 1px solid var(--border); background: var(--panel); margin-right: 4px; }
.tag.manual { background: var(--accent); border-color: var(--accent); color: #fff; }
.tag.frag { color: var(--warn); border-color: var(--warn); }
button.small { border: 1px solid var(--border); background: var(--panel); color: var(--fg); border-radius: var(--radius); padding: 2px 10px; font-size: 11.5px; cursor: pointer; margin: 2px 4px 2px 0; }
button.small:disabled { opacity: .5; cursor: default; }
button.small.primary { background: var(--accent); border-color: var(--accent); color: #fff; }
button.vol { background: var(--bg); }
.pr-form { border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 8px; background: var(--panel); margin: 6px 0; }
.pr-slot { display: flex; flex-wrap: wrap; align-items: center; gap: 6px; padding: 3px 0; font-size: 12px; }
.pr-slot b { min-width: 110px; }
.pr-slot select, .pr-slot input { border: 1px solid var(--border); background: var(--bg); color: var(--fg); border-radius: var(--radius); padding: 2px 6px; font-size: 12px; }
.pr-slot input { width: 64px; }
.pr-slot input.wide { width: 220px; }
.pr-actions { display: flex; flex-wrap: wrap; align-items: center; gap: 4px; margin-top: 4px; }
.pr-actions .grow { flex: 1; }
</style>
