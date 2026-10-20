<script setup>
import { errText } from '../utils.js'
// Workbench of the expertise remarks of one object: registry with filters, KPIs,
// quick work on a remark (status, assignee, due date), history, import journal.
import { ref, reactive, computed, onMounted } from 'vue'
import RemarkDetail from './RemarkDetail.vue'
import { STATUS_LABEL, STATUS_ORDER, CRIT_LABEL, VERDICT_LABEL, AI_LABEL, fmtDate, isOverdue } from './labels.js'

const params = new URLSearchParams(location.search)
const objectID = params.get('objectID') || ''
const namespaceID = params.get('namespaceID') || ''
const qs = `objectID=${encodeURIComponent(objectID)}&namespaceID=${encodeURIComponent(namespaceID)}`

const state = reactive({ loading: true, error: '', board: null, busy: false, actionError: '', notice: '' })
const selectedID = ref('')
const history = ref(null)
const showJournal = ref(false)
const filter = reactive({ q: '', status: '', crit: '', expert: '', discipline: '', section: '', assignee: '', iter: '', verdict: '', claim: false, overdue: false, archived: false })

async function api (url, body) {
  const res = await fetch(url, body ? { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) } : undefined)
  const data = await res.json().catch(() => ({}))
  if (!res.ok || data.error) throw new Error(errText(data.error, res.status) || res.statusText)
  return data
}

async function load () {
  try {
    const b = await api(`api/remarks/board?${qs}`)
    b.namespaceID = namespaceID
    state.board = b
  } catch (e) {
    state.error = 'Не удалось загрузить замечания: ' + e.message
  } finally {
    state.loading = false
  }
}
onMounted(load)

async function act (fn) {
  state.busy = true
  state.actionError = ''
  try { return await fn() } catch (e) { state.actionError = e.message; return null } finally { state.busy = false }
}

const importNow = (force) => act(async () => {
  const s = await api(`api/expertise/import?${qs}${force ? '&force=1' : ''}`, {})
  state.notice = s.files === 0 && !s.checked
    ? 'Новых выгрузок нет: все приложенные файлы уже загружены.'
    : `Выгрузок: ${s.files}. Новых замечаний: ${s.new}, изменено: ${s.changed}, смен статуса: ${s.statusChanges}, возвращено: ${s.reopened}, исчезло: ${s.archived}. Проверено по документам: ${s.checked}.`
  await load()
})
// The model needs minutes: the reading runs in the background and the board tells when it is over
// (X-AI-Status while it runs, X-AI-Error if it failed).
async function waitAI (total) {
  const started = Date.now()
  for (;;) {
    await new Promise(r => setTimeout(r, 3000))
    const res = await fetch(`api/remarks/board?${qs}`)
    const status = res.headers.get('X-AI-Status')
    if (!status) {
      const err = res.headers.get('X-AI-Error')
      await load()
      state.notice = err ? decodeURIComponent(err) : `ИИ-разбор готов (замечаний: ${total}).`
      return
    }
    state.notice = `ИИ-разбор: ${status} из ${total} · ${Math.round((Date.now() - started) / 1000)} с…`
    if (Date.now() - started > 15 * 60 * 1000) { state.notice = 'ИИ-разбор идёт дольше 15 минут — обновите страницу позже.'; return }
  }
}
const aiOne = (id) => act(async () => {
  await api('api/remarks/ai', { objectID, namespaceID, id })
  state.notice = 'ИИ-разбор запущен…'
  await waitAI(1)
})
const aiMany = () => act(async () => {
  const ids = rows.value.filter(r => r.documents.length).slice(0, 25).map(r => r.id)
  if (!ids.length) { state.notice = 'Среди отобранных замечаний нет привязанных к загруженным томам.'; return }
  if (!window.confirm(`Отдать модели ${ids.length} замечаний из списка (не более 25 за раз, по одному)? Это может занять несколько минут.`)) return
  const s = await api('api/remarks/ai', { objectID, namespaceID, ids })
  state.notice = `ИИ-разбор запущен для ${s.queued} замечаний…`
  await waitAI(s.queued)
})
const checkNow = () => act(async () => {
  const s = await api(`api/expertise/check?${qs}`, {})
  state.notice = `Проверка по документам: обновлено замечаний — ${s.updated}.`
  await load()
})

const remarks = computed(() => state.board ? state.board.remarks : [])
const uniq = (key) => [...new Set(remarks.value.map(r => r[key]).filter(Boolean))].sort((a, b) => String(a).localeCompare(String(b), 'ru'))
const experts = computed(() => uniq('expert'))
const disciplines = computed(() => uniq('discipline'))
const sections = computed(() => uniq('section'))
const assignees = computed(() => uniq('assignee'))

const live = computed(() => remarks.value.filter(r => filter.archived || !r.archived))
const counts = computed(() => {
  const c = { total: live.value.length, critOpen: 0, overdue: 0, claim: 0, noDocs: 0 }
  for (const s of STATUS_ORDER) c[s] = 0
  for (const r of live.value) {
    if (c[r.status] != null) c[r.status]++
    if (r.criticality === 'critical' && r.status !== 'resolved') c.critOpen++
    if (isOverdue(r)) c.overdue++
    if (r.claimMismatch) c.claim++
    if (r.verdict === 'no_documents') c.noDocs++
  }
  return c
})

const rows = computed(() => {
  const q = filter.q.trim().toLowerCase()
  return live.value.filter(r =>
    (!filter.status || r.status === filter.status) &&
    (!filter.crit || r.criticality === filter.crit) &&
    (!filter.expert || r.expert === filter.expert) &&
    (!filter.discipline || r.discipline === filter.discipline) &&
    (!filter.section || r.section === filter.section) &&
    (!filter.assignee || (filter.assignee === '—' ? !r.assignee : r.assignee === filter.assignee)) &&
    (!filter.iter || r.iterations >= Number(filter.iter)) &&
    (!filter.verdict || r.verdict === filter.verdict) &&
    (!filter.claim || r.claimMismatch) &&
    (!filter.overdue || isOverdue(r)) &&
    (!q || (r.text + ' ' + r.answer + ' ' + r.number + ' ' + r.refs).toLowerCase().includes(q)),
  )
})

const selected = computed(() => remarks.value.find(r => r.id === selectedID.value))
async function select (r) {
  selectedID.value = r.id
  history.value = null
  history.value = await act(() => api(`api/remarks/history?${qs}&remarkID=${encodeURIComponent(r.id)}`)) || []
}

async function update (id, patch) {
  const saved = await act(() => api('api/remarks/update', { objectID, namespaceID, id, ...patch }))
  if (!saved) return
  const i = state.board.remarks.findIndex(r => r.id === id)
  if (i >= 0) state.board.remarks[i] = saved
  if (patch.status) history.value = await act(() => api(`api/remarks/history?${qs}&remarkID=${encodeURIComponent(id)}`)) || history.value
}

function resetFilters () {
  Object.assign(filter, { q: '', status: '', crit: '', expert: '', discipline: '', section: '', assignee: '', iter: '', verdict: '', claim: false, overdue: false, archived: false })
}
const clip = (s, n = 140) => (s || '').length > n ? s.slice(0, n) + '…' : (s || '')
</script>

<template>
  <div class="wrap rm">
    <template v-if="state.loading">Загрузка…</template>
    <div v-else-if="state.error" class="err">{{ state.error }}</div>
    <template v-else>
      <div class="head">
        <h1>Замечания экспертизы — {{ state.board.objectName }}</h1>
        <span class="grow"></span>
        <button class="small" :disabled="state.busy" title="Прочитать новые выгрузки, приложенные к объекту" @click="importNow(false)">Загрузить выгрузку</button>
        <button class="small" :disabled="state.busy" title="Заново связать замечания с загруженными томами и проверить ответы" @click="checkNow">Проверить по документам</button>
        <button class="small" :disabled="state.busy" title="Модель прочитает отобранные замечания и фрагменты их томов (не более 25 за раз)" @click="aiMany">ИИ-разбор отобранных</button>
        <button class="small" @click="showJournal = !showJournal">Журнал выгрузок ({{ state.board.imports.length }})</button>
      </div>
      <div v-if="state.actionError" class="err action-err">{{ state.actionError }}</div>
      <div v-if="state.notice" class="comment">{{ state.notice }}</div>

      <div v-if="showJournal" class="rm-journal">
        <div v-if="!state.board.imports.length" class="empty">Выгрузок ещё не загружали: приложите Excel к объекту и нажмите «Загрузить выгрузку».</div>
        <details v-for="j in state.board.imports" :key="j.id" class="rm-j">
          <summary>
            <b>{{ fmtDate(j.at) }}</b> · {{ j.file }} — замечаний {{ j.total }}, новых {{ j.new }}, изменено {{ j.changed }}, смен статуса {{ j.statusChanges }}, возвращено {{ j.reopened }}, исчезло {{ j.archived }}
          </summary>
          <pre class="rm-j-d">{{ j.details || 'без изменений' }}</pre>
        </details>
        <button class="small" :disabled="state.busy" @click="importNow(true)">Загрузить последнюю выгрузку заново</button>
      </div>

      <div class="rm-kpis">
        <button class="kpi" :class="{ on: !filter.status && !filter.overdue }" @click="resetFilters"><b>{{ counts.total }}</b><span>всего</span></button>
        <button v-for="s in STATUS_ORDER" :key="s" class="kpi" :class="['st-' + s, { on: filter.status === s }]" @click="filter.status = filter.status === s ? '' : s">
          <b>{{ counts[s] }}</b><span>{{ STATUS_LABEL[s] }}</span>
        </button>
        <button class="kpi bad" :class="{ on: filter.crit === 'critical' && filter.status === '' }" @click="filter.crit = filter.crit === 'critical' ? '' : 'critical'">
          <b>{{ counts.critOpen }}</b><span>критичных не устранено</span>
        </button>
        <button class="kpi warn" :class="{ on: filter.overdue }" @click="filter.overdue = !filter.overdue"><b>{{ counts.overdue }}</b><span>просрочено</span></button>
        <button class="kpi warn" :class="{ on: filter.claim }" @click="filter.claim = !filter.claim"><b>{{ counts.claim }}</b><span>ответ ≠ документы</span></button>
      </div>

      <div class="disc-filters rm-filters">
        <input v-model="filter.q" class="search" placeholder="Поиск по тексту, ответу, номеру" />
        <select v-model="filter.crit"><option value="">любая критичность</option><option v-for="(l, k) in CRIT_LABEL" :key="k" :value="k">{{ l }}</option></select>
        <select v-model="filter.expert"><option value="">любой эксперт</option><option v-for="e in experts" :key="e" :value="e">{{ e }}</option></select>
        <select v-model="filter.discipline"><option value="">любое НД</option><option v-for="e in disciplines" :key="e" :value="e">{{ e }}</option></select>
        <select v-model="filter.section"><option value="">любой раздел</option><option v-for="e in sections" :key="e" :value="e">{{ clip(e, 60) }}</option></select>
        <select v-model="filter.assignee"><option value="">любой ответственный</option><option value="—">не назначен</option><option v-for="e in assignees" :key="e" :value="e">{{ e }}</option></select>
        <select v-model="filter.iter"><option value="">любая итерация</option><option value="2">итерация ≥ 2</option><option value="3">итерация ≥ 3</option></select>
        <select v-model="filter.verdict"><option value="">любая проверка</option><option v-for="(l, k) in VERDICT_LABEL" :key="k" :value="k">{{ l }}</option></select>
        <label class="chk"><input v-model="filter.archived" type="checkbox" /> и пропавшие из выгрузки</label>
        <button class="small" @click="resetFilters">Сбросить</button>
        <span class="muted">показано {{ rows.length }} из {{ live.length }}</span>
      </div>

      <div class="rm-layout" :class="{ split: selected }">
        <div class="rm-table-wrap">
          <table class="rm-table">
            <thead><tr><th>№</th><th></th><th>Статус</th><th>НД</th><th>Эксперт</th><th>Ит.</th><th>Проверка</th><th>Ответственный</th><th>Срок</th><th>Замечание</th></tr></thead>
            <tbody>
              <tr v-for="r in rows" :key="r.id" :class="{ active: r.id === selectedID, archived: r.archived }" @click="select(r)">
                <td class="num">{{ r.number || '—' }}</td>
                <td><span v-if="r.criticality === 'critical'" class="dot-crit" title="Критичное">●</span></td>
                <td><span class="st-tag" :class="'st-' + r.status">{{ STATUS_LABEL[r.status] || r.status }}</span></td>
                <td>{{ r.discipline }}</td>
                <td class="nowrap">{{ r.expert }}</td>
                <td>{{ r.iterations }}</td>
                <td><span v-if="r.verdict" class="tag" :class="'ev-' + r.verdict" :title="r.matchNote">{{ VERDICT_LABEL[r.verdict] }}</span><span v-if="r.claimMismatch" class="tag ev-claim">≠</span><span v-if="r.aiVerdict" class="tag" :class="'ai-' + r.aiVerdict" :title="AI_LABEL[r.aiVerdict]">ИИ</span></td>
                <td class="nowrap">{{ r.assignee }}</td>
                <td class="nowrap" :class="{ overdue: isOverdue(r) }">{{ fmtDate(r.dueDate) }}</td>
                <td class="txt">{{ clip(r.text) }}</td>
              </tr>
              <tr v-if="!rows.length"><td colspan="10" class="empty">Нет замечаний под выбранный фильтр.</td></tr>
            </tbody>
          </table>
        </div>
        <RemarkDetail
          v-if="selected" :remark="selected" :board="state.board" :history="history" :busy="state.busy" :assignees="assignees"
          @update="update" @close="selectedID = ''" @ai="aiOne"
        />
      </div>
    </template>
  </div>
</template>
