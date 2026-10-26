<script setup>
// Задачи (ТЗ 3.7.1): у замечания их несколько (по требованию, по тому, по согласованию), бывают и задачи без замечания.
// У задачи свои исполнитель, срок, оценка трудоёмкости и статус; «Принять» и «Вернуть» — решение другого специалиста, с причиной.
// Выполнение задач на статус замечания не влияет: когда все выполнены, панель лишь подсказывает.
import { reactive, ref, computed, watch, onMounted } from 'vue'
import { errText } from '../utils.js'
import { can } from '../platform.js'
import { fmtDate } from './labels.js'
import { TASK_STATUS_LABEL, TASK_STATUS_ORDER, nextStatuses, needsReason, MOVE_LABEL } from './tasks.js'

const props = defineProps({
  qs: String, objectId: String, namespaceId: String,
  remarkId: { type: String, default: '' }, // a panel of one remark; empty — of the whole object
  requirements: { type: Array, default: () => [] }, // the requirements of that remark, to tie a task to one
  assignees: { type: Array, default: () => [] },
  remarkNumbers: { type: Object, default: () => ({}) }, // remark id → number
})
const emit = defineEmits(['changed'])

const state = reactive({ tasks: [], counts: {}, error: '', busy: false, loading: true })
const filter = reactive({ status: '', assignee: '', open: true, overdue: false })
const form = reactive({ title: '', assignee: '', dueDate: '', estimateH: '', requirement: 0, description: '' })
const showForm = ref(false)
const moving = ref(null) // { id, to, comment }
const opened = ref('') // the task whose history is shown
const hist = ref([])

async function api (url, body) {
  const res = await fetch(url, body ? { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) } : undefined)
  const data = await res.json().catch(() => ({}))
  if (!res.ok || data.error) throw new Error(errText(data.error, res.status) || res.statusText)
  return data
}

async function load () {
  const p = new URLSearchParams({ objectID: props.objectId, namespaceID: props.namespaceId })
  if (props.remarkId) p.set('remarkID', props.remarkId)
  try {
    const d = await api(`api/tasks?${p}`)
    state.tasks = d.tasks || []
    state.counts = d.counts || {}
    state.error = ''
  } catch (e) {
    state.error = e.message
  } finally {
    state.loading = false
  }
}
onMounted(load)
watch(() => props.remarkId, () => { opened.value = ''; moving.value = null; load() })
defineExpose({ load })

const rows = computed(() => state.tasks.filter(t =>
  (!filter.status || t.status === filter.status) &&
  (!filter.open || !['done', 'cancelled'].includes(t.status)) &&
  (!filter.overdue || t.overdue) &&
  (!filter.assignee || (filter.assignee === '—' ? !t.assignee : t.assignee === filter.assignee))))
const people = computed(() => [...new Set([...props.assignees, ...state.tasks.map(t => t.assignee)].filter(Boolean))].sort((a, b) => a.localeCompare(b, 'ru')))
const hint = computed(() => props.remarkId && state.counts[props.remarkId]?.allDone)

async function act (fn) {
  state.busy = true
  state.error = ''
  try { await fn(); await load(); emit('changed') } catch (e) { state.error = e.message } finally { state.busy = false }
}

const create = () => act(async () => {
  await api('api/tasks', {
    objectID: props.objectId, namespaceID: props.namespaceId, remark: props.remarkId, requirement: Number(form.requirement) || 0,
    title: form.title, description: form.description, assignee: form.assignee, dueDate: form.dueDate, estimateH: form.estimateH,
  })
  Object.assign(form, { title: '', assignee: '', dueDate: '', estimateH: '', requirement: 0, description: '' })
  showForm.value = false
})

function startMove (t, to) {
  if (!needsReason(t.status, to)) { move(t.id, to, ''); return }
  moving.value = { id: t.id, to, comment: '' }
}
const move = (id, to, comment) => act(async () => {
  await api('api/tasks/update', { objectID: props.objectId, namespaceID: props.namespaceId, id, status: to, statusComment: comment })
  moving.value = null
})
const patch = (id, body) => act(() => api('api/tasks/update', { objectID: props.objectId, namespaceID: props.namespaceId, id, ...body }))

async function toggleHistory (t) {
  if (opened.value === t.id) { opened.value = ''; return }
  opened.value = t.id
  hist.value = []
  try { hist.value = await api(`api/tasks/history?objectID=${encodeURIComponent(props.objectId)}&namespaceID=${encodeURIComponent(props.namespaceId)}&id=${encodeURIComponent(t.id)}`) } catch (e) { state.error = e.message }
}
const reqLabel = (n) => { const q = props.requirements[n - 1]; return q ? `треб. ${n}: ${q.text.length > 50 ? q.text.slice(0, 50) + '…' : q.text}` : `треб. ${n}` }
</script>

<template>
  <div class="tk">
    <div class="tk-head">
      <b>{{ remarkId ? 'Задачи замечания' : 'Задачи объекта' }}</b>
      <span class="muted">{{ rows.length }} из {{ state.tasks.length }}</span>
      <span class="grow"></span>
      <button v-if="can('work')" class="small" :disabled="state.busy" @click="showForm = !showForm">+ задача</button>
    </div>
    <div v-if="hint" class="tk-hint">Все задачи замечания выполнены — его можно передать на контроль (статус замечания не меняется сам).</div>
    <div v-if="state.error" class="err">{{ state.error }}</div>

    <div v-if="showForm" class="tk-form">
      <input v-model="form.title" type="text" placeholder="Что нужно сделать (обязательно)" />
      <textarea v-model="form.description" rows="2" placeholder="Описание"></textarea>
      <div class="tk-row">
        <label>Исполнитель
          <input v-model="form.assignee" list="tk-people" placeholder="ФИО" :disabled="!can('assign')" :title="can('assign') ? '' : 'Назначает руководитель проекта'" />
          <datalist id="tk-people"><option v-for="p in people" :key="p" :value="p"></option></datalist>
        </label>
        <label>Срок <input v-model="form.dueDate" type="date" :disabled="!can('assign')" /></label>
        <label>Оценка, ч <input v-model="form.estimateH" type="number" min="0" step="0.5" style="width:80px" /></label>
        <label v-if="requirements.length > 1">Требование
          <select v-model.number="form.requirement"><option :value="0">— всё замечание —</option><option v-for="(q, i) in requirements" :key="i" :value="i + 1">{{ reqLabel(i + 1) }}</option></select>
        </label>
      </div>
      <button class="primary" :disabled="state.busy || !form.title.trim()" @click="create">Создать задачу</button>
    </div>

    <div class="tk-filters">
      <select v-model="filter.status"><option value="">любой статус</option><option v-for="s in TASK_STATUS_ORDER" :key="s" :value="s">{{ TASK_STATUS_LABEL[s] }}</option></select>
      <select v-model="filter.assignee"><option value="">любой исполнитель</option><option value="—">не назначен</option><option v-for="p in people" :key="p" :value="p">{{ p }}</option></select>
      <label class="chk"><input v-model="filter.open" type="checkbox" /> только открытые</label>
      <label class="chk"><input v-model="filter.overdue" type="checkbox" /> просроченные</label>
    </div>

    <div v-if="state.loading" class="muted">Загрузка…</div>
    <div v-else-if="!rows.length" class="empty">Задач нет.</div>
    <div v-for="t in rows" :key="t.id" class="tk-item" :class="{ overdue: t.overdue }">
      <div class="tk-line">
        <span class="tag" :class="'tks-' + t.status">{{ TASK_STATUS_LABEL[t.status] }}</span>
        <b class="tk-title" @click="toggleHistory(t)" title="Показать историю">{{ t.title }}</b>
        <span v-if="t.overdue" class="tag bad">просрочена</span>
        <span v-if="!remarkId && t.remark" class="muted">замечание № {{ remarkNumbers[t.remark] || '…' }}</span>
        <span v-if="t.requirement" class="muted">{{ reqLabel(t.requirement) }}</span>
      </div>
      <div class="tk-meta muted">
        <span>исполнитель: {{ t.assignee || '—' }}</span>
        <span>срок: {{ t.dueDate ? fmtDate(t.dueDate) : '—' }}</span>
        <span v-if="t.estimateH">оценка: {{ t.estimateH }} ч</span>
        <span>поставил: {{ t.createdBy || '—' }}</span>
      </div>
      <div v-if="t.description" class="tk-desc">{{ t.description }}</div>
      <div v-if="t.comment && ['done', 'returned', 'cancelled'].includes(t.status)" class="tk-comment"><b>Решение:</b> {{ t.comment }} <span class="muted">({{ t.closedBy }})</span></div>
      <div class="tk-actions">
        <button v-for="to in nextStatuses(t.status)" :key="to" class="small" :disabled="state.busy || !can(needsReason(t.status, to) ? 'confirm' : 'work')" @click="startMove(t, to)">{{ MOVE_LABEL[to] }}</button>
        <template v-if="can('assign')">
          <input type="text" :value="t.assignee" list="tk-people" placeholder="исполнитель" class="tk-inline" @change="patch(t.id, { assignee: $event.target.value })" />
          <input type="date" :value="t.dueDate" class="tk-inline" @change="patch(t.id, { dueDate: $event.target.value })" />
        </template>
      </div>
      <div v-if="moving && moving.id === t.id" class="tk-reason">
        <input v-model="moving.comment" type="text" :placeholder="'Причина: ' + MOVE_LABEL[moving.to].toLowerCase() + ' (обязательно)'" />
        <button class="primary" :disabled="state.busy || !moving.comment.trim()" @click="move(t.id, moving.to, moving.comment)">{{ MOVE_LABEL[moving.to] }}</button>
        <button class="small" @click="moving = null">Отмена</button>
      </div>
      <div v-if="opened === t.id" class="tk-hist">
        <div v-for="(h, i) in hist" :key="i"><span class="muted">{{ fmtDate(h.at) }} · {{ h.author }}</span> {{ h.source }}<template v-if="h.note"> — {{ h.note }}</template></div>
        <div v-if="!hist.length" class="muted">Загрузка…</div>
      </div>
    </div>
  </div>
</template>
