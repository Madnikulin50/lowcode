<script setup>
import { errText } from '../utils.js'
// Просмотр одного документа (вложение Compose, документ Compose, файл комплекта, акт ИД) с ассистентом:
// постраничный просмотр и панель вопросов по открытому листу. Открывается из Compose (?view=doc).
import { reactive, ref, computed, onMounted, nextTick } from 'vue'
import AssistantPanel from '../components/AssistantPanel.vue'

const params = new URLSearchParams(location.search)
const source = params.get('source') || 'attachment'
const id = params.get('id') || ''
const namespaceID = params.get('namespaceID') || ''
const recordID = params.get('recordID') || '' // set when the file is opened from a comparison
const state = reactive({ loading: true, error: '', info: null, page: Math.max(1, Number(params.get('page')) || 1), zoom: 100 })
const imgFailed = ref(false)
// a region to show on the opened page (?box=x0,y0,x1,y1 in page fractions): the fragment a remark names or an AI quotation
const boxPage = state.page
const hlBox = (() => {
  const a = (params.get('box') || '').replace(/[[\]]/g, '').split(',').map(Number)
  return a.length === 4 && a.every(n => Number.isFinite(n)) ? a : null
})()
const boxStyle = computed(() => {
  if (!hlBox || state.page !== boxPage) return null
  const [x0, y0, x1, y1] = hlBox
  const pad = 0.006
  return {
    left: Math.max(0, x0 - pad) * 100 + '%', top: Math.max(0, y0 - pad) * 100 + '%',
    width: (Math.min(1, x1 + pad) - Math.max(0, x0 - pad)) * 100 + '%', height: (Math.min(1, y1 + pad) - Math.max(0, y0 - pad)) * 100 + '%',
  }
})

const qs = () => `source=${encodeURIComponent(source)}&id=${encodeURIComponent(id)}&namespaceID=${encodeURIComponent(namespaceID)}` + (recordID ? `&recordID=${encodeURIComponent(recordID)}` : '')
const backHref = recordID ? `?recordID=${encodeURIComponent(recordID)}&namespaceID=${encodeURIComponent(namespaceID)}` : ''
const pageSrc = computed(() => `api/assist/page?${qs()}&n=${state.page}`)
const count = computed(() => state.info?.pageCount || 0)

onMounted(async () => {
  try {
    const r = await fetch(`api/assist/doc?${qs()}`)
    const data = await r.json().catch(() => ({}))
    if (!r.ok || data.error) throw new Error(errText(data.error, r.status) || r.statusText)
    state.info = data
    state.page = Math.min(state.page, Math.max(1, data.pageCount))
  } catch (e) {
    state.error = e.message
  } finally {
    state.loading = false
  }
})

const go = (n) => { state.page = Math.min(Math.max(1, n), Math.max(1, count.value)); imgFailed.value = false }
const getContext = () => (state.info ? { source, id, recordID, page: state.page } : null)
const chips = computed(() => (state.info ? [`${state.info.name}, стр. ${state.page} из ${count.value}`] : []))
function gotoSheet (j) { if (j.page) go(j.page); nextTick(() => document.querySelector('.doc-stage')?.scrollIntoView({ behavior: 'smooth' })) }
</script>

<template>
  <div class="wrap doc">
    <template v-if="state.loading">Загрузка…</template>
    <div v-else-if="state.error" class="err">{{ state.error }}</div>
    <template v-else>
      <div class="head">
        <a v-if="backHref" class="back" :href="backHref" title="Вернуться к сравнению">← к сравнению</a>
        <h1>{{ state.info.name }}</h1>
        <span v-if="state.info.title" class="muted">{{ state.info.title }}</span>
      </div>
      <div class="doc-bar">
        <button class="small" :disabled="state.page <= 1" @click="go(state.page - 1)">‹</button>
        <span>стр. <input class="pg" type="number" min="1" :max="count" :value="state.page" @change="e => go(Number(e.target.value))"> / {{ count }}</span>
        <button class="small" :disabled="state.page >= count" @click="go(state.page + 1)">›</button>
        <span class="sep"></span>
        <button class="small" @click="state.zoom = Math.max(40, state.zoom - 20)">−</button>
        <span>{{ state.zoom }}%</span>
        <button class="small" @click="state.zoom = Math.min(300, state.zoom + 20)">+</button>
      </div>
      <div class="doc-layout">
        <div class="doc-stage">
          <div v-if="!imgFailed" class="doc-page" :style="{ width: state.zoom + '%' }">
            <img :src="pageSrc" alt="страница" @error="imgFailed = true">
            <div v-if="boxStyle" class="doc-hl" :style="boxStyle" title="Место из замечания"></div>
          </div>
          <div v-else class="empty">Для этого типа файла нет изображения страницы. Ассистент всё равно может отвечать по тексту, если он есть.</div>
        </div>
        <AssistantPanel
          class="doc-assistant" :namespace-id="namespaceID" :get-context="getContext" :chips="chips" start-open
          title="Ассистент по документу" @goto="gotoSheet"
        />
      </div>
    </template>
  </div>
</template>

<style scoped>
.doc-bar { display: flex; align-items: center; gap: 8px; margin: 6px 0; font-size: 12px; }
.doc-bar .sep { width: 1px; background: var(--border); align-self: stretch; margin: 0 4px; }
.doc-bar input.pg { width: 56px; border: 1px solid var(--border); background: var(--bg); color: var(--fg); border-radius: var(--radius); padding: 1px 4px; }
.doc-layout { display: grid; grid-template-columns: minmax(0, 1.3fr) minmax(320px, 1fr); gap: 12px; align-items: start; }
.doc-stage { overflow: auto; max-height: 85vh; border: 1px solid var(--border); border-radius: var(--radius-lg); background: var(--panel); text-align: center; }
.doc-page { position: relative; margin: 0 auto; }
.doc-stage img { max-width: none; width: 100%; display: block; }
.doc-hl { position: absolute; border: 2px solid #e8590c; background: rgba(232, 89, 12, .12); border-radius: 3px; pointer-events: none; }
.doc-assistant { margin-top: 0; }
button.small { border: 1px solid var(--border); background: var(--panel); color: var(--fg); border-radius: var(--radius); padding: 2px 10px; font-size: 12px; cursor: pointer; }
button.small:disabled { opacity: .5; cursor: default; }
.muted { color: var(--muted); }
.back { font-size: 12px; color: var(--accent); text-decoration: none; }
@media (max-width: 900px) { .doc-layout { grid-template-columns: 1fr; } }
</style>
