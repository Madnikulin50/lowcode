<script setup>
// Лента уведомлений объекта (ТЗ 3.7.5): назначения, смены статуса, приближение и нарушение сроков.
// События находит агент; здесь — прочитать, отфильтровать по получателю и перейти к замечанию.
import { reactive, ref, computed, onMounted } from 'vue'
import { errText } from '../utils.js'
import { fmtDate } from './labels.js'

const props = defineProps({ qs: String, namespaceId: String, assignees: { type: Array, default: () => [] } })
const emit = defineEmits(['open'])

const EVENT = { assigned: 'назначение', status: 'статус', due_soon: 'срок скоро', overdue: 'просрочка' }
const DELIVERY = { sent: 'письмо отправлено', pending: 'письмо в очереди', failed: 'письмо не ушло', none: '' }

const state = reactive({ items: [], unread: 0, mail: false, fixture: false, error: '', busy: false })
const open = ref(false)
const who = ref('')
const onlyUnread = ref(false)

async function api (url, body) {
  const res = await fetch(url, body ? { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) } : undefined)
  const data = await res.json().catch(() => ({}))
  if (!res.ok || data.error) throw new Error(errText(data.error, res.status) || res.statusText)
  return data
}

async function load () {
  try {
    const d = await api(`api/notifications?${props.qs}&recipient=${encodeURIComponent(who.value)}`)
    Object.assign(state, { items: d.items || [], unread: d.unread || 0, mail: !!d.mail, fixture: !!d.fixture, error: '' })
  } catch (e) {
    state.error = e.message
  }
}
onMounted(load)
defineExpose({ load })

const recipients = computed(() => [...new Set([...props.assignees, ...state.items.map(n => n.recipient)].filter(Boolean))].sort())
const rows = computed(() => onlyUnread.value ? state.items.filter(n => !n.read) : state.items)

async function markRead (ids) {
  if (!ids.length) return
  try {
    await api('api/notifications/read', { namespaceID: props.namespaceId, ids })
    for (const n of state.items) if (ids.includes(n.id)) n.read = true
    state.unread = state.items.filter(n => !n.read).length
  } catch (e) {
    state.error = e.message
  }
}

function go (n) {
  if (!n.read) markRead([n.id])
  if (n.remark) emit('open', n.remark)
}

async function scan () {
  state.busy = true
  try {
    await api(`api/notifications/scan?${props.qs}`, {})
    await load()
  } catch (e) {
    state.error = e.message
  } finally {
    state.busy = false
  }
}
</script>

<template>
  <button class="small bell" :class="{ hot: state.unread }" title="Уведомления: назначения, смены статуса, сроки" @click="open = !open">
    🔔 Уведомления<span v-if="state.unread" class="cnt">{{ state.unread }}</span>
  </button>
  <Teleport to="body">
    <div v-if="open" class="nt-panel">
      <div class="nt-head">
        <b>Уведомления</b>
        <select v-model="who" @change="load"><option value="">все получатели</option><option v-for="r in recipients" :key="r" :value="r">{{ r }}</option></select>
        <label><input v-model="onlyUnread" type="checkbox"> непрочитанные</label>
        <span class="grow"></span>
        <button class="small" @click="open = false">✕</button>
      </div>
      <div class="nt-tools">
        <button class="small" :disabled="!state.unread" @click="markRead(state.items.filter(n => !n.read).map(n => n.id))">Прочитать все</button>
        <button v-if="!state.fixture" class="small" :disabled="state.busy" title="Найти новые события сейчас, не дожидаясь очередного прохода" @click="scan">Проверить сейчас</button>
        <span class="muted">{{ state.fixture ? 'тестовый режим: письма не отправляются' : state.mail ? 'письма отправляются по e-mail' : 'e-mail не настроен — только лента' }}</span>
      </div>
      <div v-if="state.error" class="err">{{ state.error }}</div>
      <div v-if="!rows.length" class="empty">Уведомлений нет.</div>
      <div v-for="n in rows" :key="n.id" class="nt-item" :class="{ unread: !n.read }" @click="go(n)">
        <div class="nt-top">
          <span :class="['tag', 'ev-' + n.event]">{{ EVENT[n.event] || n.event }}</span>
          <b>{{ n.title }}</b>
        </div>
        <div class="nt-text">{{ n.text }}</div>
        <div class="muted">{{ n.recipient }} · {{ fmtDate(n.created) }}<template v-if="DELIVERY[n.delivery]"> · <span :class="{ bad: n.delivery === 'failed' }" :title="n.error">{{ DELIVERY[n.delivery] }}</span></template></div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.bell .cnt { display: inline-block; margin-left: 6px; min-width: 18px; padding: 0 5px; border-radius: 9px; background: var(--bad); color: #fff; font-size: 11px; text-align: center; }
.nt-panel { position: fixed; top: 0; right: 0; bottom: 0; width: min(440px, 100vw); background: var(--bg); border-left: 1px solid var(--border); box-shadow: -4px 0 16px rgba(0,0,0,.12); overflow-y: auto; z-index: 50; padding: 12px; font-size: 12.5px; }
.nt-head, .nt-tools { display: flex; flex-wrap: wrap; gap: 8px; align-items: center; margin-bottom: 8px; }
.nt-head select { border: 1px solid var(--border); background: var(--bg); color: var(--fg); border-radius: 6px; padding: 2px 6px; font-size: 12px; max-width: 200px; }
.grow { flex: 1; }
.muted { color: var(--muted); font-size: 11.5px; }
.bad { color: var(--bad); }
.nt-item { border: 1px solid var(--border); border-radius: 8px; padding: 6px 8px; margin-bottom: 6px; cursor: pointer; }
.nt-item.unread { border-left: 3px solid var(--accent); background: var(--panel); }
.nt-top { display: flex; gap: 6px; align-items: baseline; }
.nt-text { white-space: pre-line; margin: 3px 0; }
.tag { display: inline-block; padding: 0 7px; border-radius: 8px; font-size: 11px; border: 1px solid var(--border); white-space: nowrap; }
.tag.ev-overdue { border-color: var(--bad); color: var(--bad); }
.tag.ev-due_soon { border-color: var(--warn); color: var(--warn); }
.empty { color: var(--muted); padding: 8px 0; }
</style>
