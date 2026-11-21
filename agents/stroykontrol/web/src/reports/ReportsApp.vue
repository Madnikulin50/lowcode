<script setup>
import { errText } from '../utils.js'
// «Отчёты» (ТЗ 3.9): выгрузки по объектам в одном месте. Реестр замечаний — с отбором по статусу,
// ответственному, разделу и периоду; по каждому сравнению — отчёт (Excel/Word) и перечень замечаний.
import { reactive, computed, onMounted } from 'vue'
import { STATUS_LABEL, STATUS_ORDER } from '../remarks/labels.js'

const namespaceID = new URLSearchParams(location.search).get('namespaceID') || ''
const state = reactive({ loading: true, error: '', objects: [], notice: '', busy: false })
// the filters of the registry export, shared by every object
const f = reactive({ status: '', assignee: '', section: '', from: '', to: '', overdue: false })

onMounted(async () => {
  try {
    const res = await fetch(`api/reports/index?namespaceID=${encodeURIComponent(namespaceID)}`)
    const data = await res.json().catch(() => ({}))
    if (!res.ok || data.error) throw new Error(errText(data.error, res.status) || res.statusText)
    state.objects = data.objects || []
  } catch (e) {
    state.error = 'Не удалось загрузить список объектов: ' + e.message
  } finally {
    state.loading = false
  }
})

const q = (obj) => {
  const p = new URLSearchParams({ objectID: obj.id, namespaceID, format: 'xlsx' })
  if (f.status) p.set('status', f.status)
  if (f.assignee.trim()) p.set('assignee', f.assignee.trim())
  if (f.section.trim()) p.set('section', f.section.trim())
  if (f.from) p.set('from', f.from)
  if (f.to) p.set('to', f.to)
  if (f.overdue) p.set('overdue', '1')
  return p
}
const registryHref = (obj) => `api/remarks/export?${q(obj)}`
const cmpQs = (c, format) => `recordID=${encodeURIComponent(c.id)}&namespaceID=${encodeURIComponent(namespaceID)}&format=${format}`
const reportHref = (c, format) => `api/report/export?${cmpQs(c, format)}`
const issueHref = (c, format) => `api/issue/export?${cmpQs(c, format)}`
const workbenchHref = (o) => `?view=remarks&objectID=${encodeURIComponent(o.id)}&namespaceID=${encodeURIComponent(namespaceID)}`
const viewerHref = (c) => `?recordID=${encodeURIComponent(c.id)}&namespaceID=${encodeURIComponent(namespaceID)}`
const periodBad = computed(() => f.from && f.to && f.from > f.to)

// A download that fails (no confirmed findings for a list of remarks, …) says why instead of showing a page of JSON.
async function download (href) {
  state.notice = ''
  state.busy = true
  try {
    const res = await fetch(href)
    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(errText(data.error, res.status) || res.statusText)
    }
    const name = decodeURIComponent((res.headers.get('Content-Disposition') || '').replace(/^.*filename\*=UTF-8''/, '')) || 'report'
    const url = URL.createObjectURL(await res.blob())
    const a = document.createElement('a')
    a.href = url
    a.download = name
    document.body.appendChild(a)
    a.click()
    a.remove()
    setTimeout(() => URL.revokeObjectURL(url), 5000)
  } catch (e) {
    state.notice = 'Не удалось выгрузить: ' + e.message
  } finally {
    state.busy = false
  }
}
</script>

<template>
  <div class="wrap rp">
    <h1>Отчёты</h1>
    <template v-if="state.loading">Загрузка…</template>
    <div v-else-if="state.error" class="err">{{ state.error }}</div>
    <template v-else>
      <div class="rp-filters">
        <b>Отбор для реестра замечаний</b>
        <select v-model="f.status"><option value="">любой статус</option><option v-for="s in STATUS_ORDER" :key="s" :value="s">{{ STATUS_LABEL[s] }}</option></select>
        <input v-model="f.assignee" placeholder="ответственный (— : не назначен)" />
        <input v-model="f.section" placeholder="раздел (точно)" />
        <label class="chk">период с <input v-model="f.from" type="date" /> по <input v-model="f.to" type="date" /></label>
        <label class="chk"><input v-model="f.overdue" type="checkbox" /> только просроченные</label>
        <span v-if="periodBad" class="err">Начало периода позже конца</span>
      </div>
      <div v-if="state.notice" class="err">{{ state.notice }}</div>
      <div class="muted">Период отбирает замечания, по которым что-то менялось в нём: выгрузка экспертизы или запись истории (назначение, срок, заметка, статус, решение).</div>

      <div v-if="!state.objects.length" class="empty">Объектов нет.</div>
      <div v-for="o in state.objects" :key="o.id" class="rp-obj">
        <div class="rp-head">
          <b>{{ o.name }}</b><span v-if="o.code" class="muted"> · {{ o.code }}</span>
          <span class="muted"> · замечаний: {{ o.remarks }}<template v-if="o.packages"> · пакетов исправления: {{ o.packages }}</template></span>
        </div>
        <div class="rp-row">
          <span class="rp-what">Реестр замечаний экспертизы</span>
          <a class="rp-btn" :class="{ off: periodBad || !o.remarks || state.busy }" href="#" @click.prevent="download(registryHref(o))" title="Реестр, история итераций, проверка устранения, динамика (Excel)">⬇ Excel</a>
          <a :href="workbenchHref(o)" target="_self">рабочее место замечаний</a>
        </div>
        <div v-for="c in o.comparisons" :key="c.id" class="rp-row">
          <span class="rp-what">Сравнение «{{ c.title }}»<template v-if="c.iteration"> · итерация {{ c.iteration }}</template></span>
          <template v-if="c.status === 'done'">
            <a class="rp-btn" :class="{ off: state.busy }" href="#" @click.prevent="download(reportHref(c, 'xlsx'))" title="Отчёт по сравнению: сводка, по томам, расхождения, оформление, параметры">⬇ Отчёт Excel</a>
            <a class="rp-btn" :class="{ off: state.busy }" href="#" @click.prevent="download(reportHref(c, 'docx'))">⬇ Отчёт Word</a>
            <a class="rp-btn" :class="{ off: state.busy }" href="#" @click.prevent="download(issueHref(c, 'xlsx'))" title="Перечень замечаний по подтверждённым находкам (в формате выгрузки экспертизы)">⬇ Перечень замечаний Excel</a>
            <a class="rp-btn" :class="{ off: state.busy }" href="#" @click.prevent="download(issueHref(c, 'docx'))">⬇ Перечень Word</a>
          </template>
          <span v-else class="muted">сравнение не завершено ({{ c.status || 'new' }})</span>
          <a :href="viewerHref(c)" target="_self">открыть сравнение</a>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.rp-filters { display: flex; flex-wrap: wrap; align-items: center; gap: 8px 12px; margin: 8px 0 4px; font-size: 12.5px; }
.rp-filters input, .rp-filters select { border: 1px solid var(--border); background: var(--bg); color: var(--fg); border-radius: var(--radius); padding: 3px 7px; font-size: 12.5px; }
.rp-obj { border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 8px 12px; margin: 12px 0; background: var(--panel); }
.rp-head { margin-bottom: 6px; }
.rp-row { display: flex; flex-wrap: wrap; align-items: center; gap: 6px 10px; padding: 4px 0; border-top: 1px solid var(--border); font-size: 12.5px; }
.rp-what { min-width: 260px; }
.rp-btn { border: 1px solid var(--border); background: var(--bg); color: var(--fg); border-radius: var(--radius); padding: 2px 10px; font-size: 12px; text-decoration: none; }
.rp-btn:hover { border-color: var(--accent); }
.rp-btn.off { opacity: .45; pointer-events: none; }
</style>
