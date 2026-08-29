<template>
  <div class="h-100">
    <Wrap
      v-bind="$props"
      :headerClass="editable ? '' : 'd-none'"
      :scrollable-body="false"
    >
    <div class="rulechain-body h-100 d-flex flex-column justify-content-center">
      <div class="d-flex flex-wrap align-items-center gap-2">
        <button
          class="btn flex-shrink-0"
          :class="btnClass"
          :disabled="running"
          @click="runChain"
        >
          <font-awesome-icon :icon="icon" class="me-1" />
          {{ label }}
        </button>

        <div
          v-if="running"
          class="rulechain-status d-flex align-items-center text-muted"
        >
          <span class="spinner-border spinner-border-sm me-2" role="status" />
          <span>{{ statusText }}</span>
        </div>
      </div>
      <div
        v-if="!running && result"
        class="rulechain-status mt-2"
        :class="resultOk ? 'is-success' : 'is-danger'"
        :title="statusText"
      >
        <font-awesome-icon
          :icon="['fas', resultOk ? 'check-circle' : 'exclamation-circle']"
          class="me-1"
        />
        <span class="rulechain-status-text">{{ statusText }}</span>
      </div>
    </div>
  </Wrap>

  <Teleport to="body">
    <div
      v-if="balloon"
      class="rulechain-balloon"
      :class="balloon.ok ? 'is-success' : 'is-danger'"
      role="status"
    >
      <div class="rulechain-balloon-header">
        <strong>{{ balloon.title }}</strong>
        <button
          type="button"
          class="btn-close btn-close-white"
          aria-label="Close"
          @click="dismissBalloon"
        />
      </div>
      <div class="rulechain-balloon-body">{{ balloon.text }}</div>
    </div>
  </Teleport>
  </div>
</template>

<script setup>
defineOptions({ inheritAttrs: false, i18nOptions: { namespaces: 'block' } })
import { ref, computed, inject, onMounted, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { NoID } from 'corteza-lib/js/dist'
import bus from '../../lib/bus'
import Wrap from './Wrap/index.js'
import { scanIDsFromTrigger, pullScanResultsIntoCompose } from './cmdbAgentSync.js'

const props = defineProps({
  blockIndex: { type: Number, default: -1 },
  namespace: { type: Object, required: true },
  page: { type: Object, required: true },
  blocks: { type: Array, default: () => [] },
  block: { type: Object, required: true },
  module: { type: Object, default: undefined },
  record: { type: Object, default: undefined },
  mode: { type: String, default: '' },
  editable: { type: Boolean, default: false },
  resizing: { type: Boolean, default: false },
  magnified: { type: Boolean, default: false },
  unsavedBlocks: { type: Set, default: () => new Set() },
  loadingRecord: { type: Boolean, default: false },
  errors: { type: Object, default: () => ({}) },
})

const { t, locale } = useI18n({ useScope: 'global' })
const route = useRoute()
const $ComposeAPI = inject('$ComposeAPI')

const running = ref(false)
const result = ref(null)
const chainName = ref('')
const balloon = ref(null)
let pollAbort = false
let balloonTimer = null

onBeforeUnmount(() => {
  pollAbort = true
  clearTimeout(balloonTimer)
})

const chainID = computed(() => props.block.options?.chainID || '')
const label = computed(() => props.block.options?.label || 'Run Rule Chain')
const icon = computed(() => ['fas', props.block.options?.icon || 'play'])
const btnClass = computed(() => `btn-${props.block.options?.variant || 'primary'} ${(props.block.options?.size) ? 'btn-' + props.block.options.size : ''}`)
const reloadOnSuccess = computed(() => !!props.block.options?.reloadOnSuccess)

async function loadChainName () {
  const id = chainID.value
  if (!id) {
    chainName.value = ''
    return
  }
  try {
    const { data } = await $ComposeAPI.api().request({
      method: 'get',
      url: $ComposeAPI.baseURL + '/rulechain/',
    })
    const chains = data?.response?.chains || data?.chains || []
    const found = chains.find(c => c.id === id || c.ID === id)
    chainName.value = found?.name || found?.Name || ''
  } catch {
    chainName.value = ''
  }
}

onMounted(loadChainName)
watch(chainID, loadChainName)

function locFallback (key, ru, en) {
  const v = t(key)
  if (v && !String(v).includes(key.split('.').pop())) return v
  const loc = String(locale.value || '').split('-')[0]
  return loc === 'en' ? en : ru
}

const runningLabel = computed(() => locFallback('ruleChain.running', 'Выполняется…', 'Running…'))
const successLabel = computed(() => locFallback('ruleChain.success', 'Цепочка выполнена успешно', 'Rule chain completed successfully'))
const errorLabel = computed(() => locFallback('ruleChain.error', 'Ошибка выполнения', 'Execution failed'))
const resultTitle = computed(() => locFallback('ruleChain.resultTitle', 'Результат', 'Result'))

function reloadPageIfNeeded () {
  if (pollAbort || !reloadOnSuccess.value) return
  window.location.reload()
}

const SECRET_KEY = /^(authToken|token|password|secret|authorization)$/i
const CONTEXT_KEY = /^(namespaceID|recordID|pageID|moduleID|userID|blockID|chainID)$/i

const FIELD_LABELS = {
  wbs: 'WBS',
  updated: 'Обновлено',
  project: 'Проект',
  projectID: 'Проект',
  spi: 'SPI',
  SPI: 'SPI',
  cpi: 'CPI',
  CPI: 'CPI',
  eac: 'EAC',
  EAC: 'EAC',
  pv: 'PV (плановый объём)',
  PV: 'PV (плановый объём)',
  ev: 'EV (освоенный объём)',
  EV: 'EV (освоенный объём)',
  ac: 'AC (факт. затраты)',
  AC: 'AC (факт. затраты)',
  bac: 'BAC (бюджет)',
  BAC: 'BAC (бюджет)',
  critical: 'На критическом пути',
  alerts: 'Алертов',
  created: 'Создано рисков',
  items: 'Элементы',
  statusCode: 'HTTP',
}

function asObject (v) {
  if (v == null) return null
  if (typeof v === 'string') {
    const s = v.trim()
    if ((s.startsWith('{') && s.endsWith('}')) || (s.startsWith('[') && s.endsWith(']'))) {
      try { v = JSON.parse(s) } catch { return null }
    } else {
      return null
    }
  }
  if (typeof v !== 'object') return null
  return v
}

function unwrapEnvelope (body) {
  body = asObject(body)
  if (!body || Array.isArray(body)) return body
  for (const key of ['response', 'Response', 'body', 'Body', 'result', 'data']) {
    const inner = body[key]
    if (inner && typeof inner === 'object' && inner !== body) {
      const deeper = unwrapEnvelope(inner)
      if (deeper) return deeper
    }
  }
  return body
}

function nodeList (res) {
  return res?.nodes || res?.Nodes || []
}

function candidates (res) {
  const list = []
  const seen = new Set()
  const add = (v) => {
    const obj = unwrapEnvelope(v) || asObject(v)
    if (!obj || seen.has(obj)) return
    seen.add(obj)
    list.push(obj)
  }
  add(res?.output)
  add(res?.Output)
  add(res?.result)
  add(res)
  for (const n of nodeList(res)) {
    const out = n?.output || n?.Output
    add(out)
    add(out?.body)
    add(out?.Body)
    add(n)
  }
  const extra = []
  for (const obj of list) {
    extra.push(
      obj.body, obj.Body, obj.response, obj.Response,
      obj.project, obj.Project, obj.http, obj.result, obj.data,
    )
  }
  extra.forEach(add)
  return list
}

function extractPayload (res) {
  for (const obj of candidates(res)) {
    if (formatEvm(obj) || formatCpm(obj) || formatAlerts(obj) || summarizeReorder(obj) || summarizeRisk(obj)) {
      return obj
    }
  }
  const bodies = []
  for (const n of nodeList(res)) {
    const out = n?.output || n?.Output
    if (out && (out.body != null || out.Body != null)) bodies.push(out.body ?? out.Body)
  }
  if (bodies.length) return bodies[bodies.length - 1]
  const out = res?.output ?? res?.Output
  if (typeof out === 'string' && out) return out
  if (out && typeof out === 'object') return out
  return null
}

function fmtNum (n, digits = 2) {
  const x = Number(n)
  if (!Number.isFinite(x)) return String(n)
  return x.toLocaleString('ru-RU', { maximumFractionDigits: digits })
}

function summarizeReorder (out) {
  if (!out || typeof out !== 'object') return null
  if (out.orderCount == null && out.OrderCount == null) return null
  return {
    orderCount: Number(out.orderCount ?? out.OrderCount ?? 0),
    lineCount: Number(out.lineCount ?? out.LineCount ?? 0),
    totalQty: Number(out.totalQty ?? out.TotalQty ?? 0),
    totalSum: Number(out.totalSum ?? out.TotalSum ?? 0),
  }
}

function summarizeRisk (out) {
  if (!out || typeof out !== 'object') return null
  if (out.level == null && out.residualScore == null && out.score == null) return null
  if (out.wbs != null || out.WBS != null || out.project || out.Project) return null
  return {
    level: out.level || '',
    residual: out.residualScore ?? out.residual,
    score: out.score,
    name: out.name || '',
  }
}

function formatEvm (body) {
  body = unwrapEnvelope(body)
  if (!body || Array.isArray(body)) return null
  const project = unwrapEnvelope(body.project || body.Project) || {}
  const wbs = body.wbs ?? body.WBS
  const updated = body.updated ?? body.Updated
  const spi = project.spi ?? project.SPI ?? body.spi ?? body.SPI
  const cpi = project.cpi ?? project.CPI ?? body.cpi ?? body.CPI
  const eac = project.eac ?? project.EAC ?? body.eac ?? body.EAC
  const pv = project.pv ?? project.PV ?? body.pv ?? body.PV
  const ev = project.ev ?? project.EV ?? body.ev ?? body.EV
  const ac = project.ac ?? project.AC ?? body.ac ?? body.AC
  const bac = project.bac ?? project.BAC ?? body.bac ?? body.BAC
  if (wbs == null && updated == null && spi == null && cpi == null) return null
  const lines = []
  if (updated != null) lines.push(`Обновлено позиций WBS: ${updated}`)
  else if (wbs != null) lines.push(`Позиций WBS: ${wbs}`)
  const saved = body.saved ?? body.Saved
  if (saved != null) lines.push(`Записей проекта обновлено: ${saved}`)
  const saveErr = body.saveError || body.SaveError
  if (saveErr) lines.push(`Ошибка записи проекта: ${saveErr}`)
  if (spi != null) lines.push(`SPI: ${fmtNum(spi, 3)}`)
  if (cpi != null) lines.push(`CPI: ${fmtNum(cpi, 3)}`)
  if (eac != null) lines.push(`EAC: ${fmtNum(eac)}`)
  const rest = []
  if (pv != null) rest.push(`PV ${fmtNum(pv)}`)
  if (ev != null) rest.push(`EV ${fmtNum(ev)}`)
  if (ac != null) rest.push(`AC ${fmtNum(ac)}`)
  if (bac != null) rest.push(`BAC ${fmtNum(bac)}`)
  if (rest.length) lines.push(rest.join(' · '))
  return lines.join('\n')
}

function formatCpm (body) {
  body = unwrapEnvelope(body)
  if (!body || Array.isArray(body)) return null
  const critical = body.critical ?? body.Critical
  const wbs = body.wbs ?? body.WBS
  if (critical == null) return null
  if (body.project || body.Project) return null
  return wbs != null
    ? `На критическом пути: ${critical} из ${wbs} позиций WBS`
    : `На критическом пути: ${critical}`
}

function formatAlerts (body) {
  body = unwrapEnvelope(body)
  if (!body || Array.isArray(body)) return null
  const created = body.created ?? body.Created
  const alerts = body.alerts ?? body.Alerts
  if (created == null && typeof alerts !== 'number' && !Array.isArray(alerts)) return null
  if (body.project || body.Project) return null
  const n = typeof alerts === 'number' ? alerts : Array.isArray(alerts) ? alerts.length : created
  const lines = [`Найдено алертов: ${n}`]
  if (created != null) lines.push(`Создано рисков: ${created}`)
  const items = body.items || body.Items
  if (Array.isArray(items) && items.length) {
    for (const it of items.slice(0, 5)) {
      const title = it.title || it.Title || it.kind || it.Kind
      if (title) lines.push(`• ${title}`)
    }
    if (items.length > 5) lines.push(`… и ещё ${items.length - 5}`)
  }
  return lines.join('\n')
}

function formatHuman (obj, depth = 0) {
  obj = unwrapEnvelope(obj)
  if (obj == null) return ''
  if (typeof obj !== 'object') return String(obj)
  if (Array.isArray(obj)) {
    return obj.map(item => formatHuman(item, depth + 1)).filter(Boolean).join('\n')
  }
  const skip = ['record', 'values', 'statusCode', 'status', 'truncated', 'nodes', 'success', 'error', 'chainID', 'blockID']
  const lines = []
  for (const [k, v] of Object.entries(obj)) {
    if (SECRET_KEY.test(k) || CONTEXT_KEY.test(k) || skip.includes(k)) continue
    if (v == null || v === '') continue
    const label = FIELD_LABELS[k] || FIELD_LABELS[k.toLowerCase()] || k
    if (typeof v === 'object') {
      const nested = formatEvm(v) || formatCpm(v) || formatAlerts(v) || formatHuman(v, depth + 1)
      if (!nested) continue
      if (['project', 'Project', 'http', 'response'].includes(k) || label === 'Проект') {
        lines.push(nested)
      } else {
        lines.push(`${label}:`)
        lines.push(nested.split('\n').map(l => '  ' + l).join('\n'))
      }
    } else if (typeof v === 'number') {
      lines.push(`${label}: ${fmtNum(v, Number.isInteger(v) ? 0 : 3)}`)
    } else {
      lines.push(`${label}: ${v}`)
    }
    if (lines.length > 24) break
  }
  return lines.join('\n')
}

function formatSuccess (res) {
  const sources = candidates(res)
  for (const src of sources) {
    const s = summarizeReorder(src)
    if (s) {
      const qty = Number.isFinite(s.totalQty) ? s.totalQty.toLocaleString('ru-RU') : s.totalQty
      const sum = Number.isFinite(s.totalSum) ? s.totalSum.toLocaleString('ru-RU', { maximumFractionDigits: 2 }) : s.totalSum
      return `Создано заказов: ${s.orderCount}\nСтрок: ${s.lineCount}\nКоличество: ${qty}\nСумма: ${sum}`
    }
    const r = summarizeRisk(src)
    if (r) {
      const who = r.name ? `${r.name}\n` : ''
      const residual = r.residual != null ? Number(r.residual).toLocaleString('ru-RU', { maximumFractionDigits: 1 }) : '—'
      const inn = r.score != null ? Number(r.score).toLocaleString('ru-RU', { maximumFractionDigits: 1 }) : '—'
      return `${who}Уровень риска: ${r.level || 'n/a'}\nResidual: ${residual}\nInherent: ${inn}`
    }
    const evm = formatEvm(src)
    if (evm) return evm
    const cpm = formatCpm(src)
    if (cpm) return cpm
    const alerts = formatAlerts(src)
    if (alerts) return alerts
  }
  const payload = extractPayload(res)
  if (typeof payload === 'string' && payload && !payload.trim().startsWith('{')) return payload
  const human = formatHuman(payload)
  if (human) return human
  return successLabel.value
}

const successText = computed(() => formatSuccess(result.value))
const resultOk = computed(() => {
  const res = result.value
  if (!res) return false
  return res.success !== false && !res.error && !res.Error
})

const statusText = computed(() => {
  if (running.value) {
    const out = result.value?.output
    if (typeof out === 'string' && out) return out
    return runningLabel.value
  }
  if (!result.value) return ''
  if (resultOk.value) return successText.value
  return result.value.error || result.value.Error || errorLabel.value
})

function resChainName (res) {
  if (!res || typeof res !== 'object') return ''
  return res.chainName || res.ChainName || ''
}

function dismissBalloon () {
  clearTimeout(balloonTimer)
  balloon.value = null
}

function notifyResult (res) {
  if (!res || pollAbort) return
  if (res.success && reloadOnSuccess.value) return
  const title = resChainName(res) || chainName.value || label.value || resultTitle.value
  const ok = res.success !== false && !res.error && !res.Error
  balloon.value = {
    ok,
    title,
    text: ok ? formatSuccess(res) : (res.error || res.Error || errorLabel.value),
  }
  clearTimeout(balloonTimer)
  balloonTimer = setTimeout(() => { balloon.value = null }, 12000)
}

function unwrapTrigger (data) {
  let res = data?.response || data || {}
  if (res.response && (res.response.nodes || res.response.output || res.response.success != null)) {
    res = res.response
  }
  const payload = extractPayload(res)
  return {
    ...res,
    success: res.success !== false && !res.error && !res.Error,
    error: res.error || res.Error || '',
    output: payload != null ? payload : res.output,
    nodes: res.nodes || res.Nodes || [],
  }
}

function recordPayload (record) {
  if (!record) return undefined
  const values = {}
  const src = record.values
  if (src && typeof src === 'object' && !Array.isArray(src)) {
    for (const [k, v] of Object.entries(src)) {
      if (typeof v === 'function') continue
      values[k] = v
    }
  } else if (Array.isArray(src)) {
    for (const row of src) {
      if (row?.name) values[row.name] = row.value
    }
  }
  return {
    recordID: liveRecordID() || record.recordID,
    moduleID: record.moduleID,
    namespaceID: record.namespaceID,
    values,
  }
}

function unwrapScalar (v) {
  if (Array.isArray(v)) {
    return v.length ? unwrapScalar(v[0]) : ''
  }
  if (v && typeof v === 'object' && 'value' in v) {
    return unwrapScalar(v.value)
  }
  return v
}

function asRecordID (id) {
  if (id == null || id === '' || id === NoID || id === '0' || id === 0) return ''
  return String(id)
}

function liveRecordID () {
  // View/modal often mount RuleChain before props.record is assigned.
  // The URL always has the document id (params on the record page, query in a modal).
  return asRecordID(props.record?.recordID)
    || asRecordID(route.query?.recordID)
    || asRecordID(route.params?.recordID)
}

function interpolatePlaceholders (v, recID) {
  if (typeof v === 'string') {
    if (recID) {
      return v.split('${recordID}').join(recID).split('{{recordID}}').join(recID)
    }
    if (v.includes('${recordID}') || v.includes('{{recordID}}')) return ''
    return v
  }
  if (Array.isArray(v)) return v.map(x => interpolatePlaceholders(x, recID))
  if (v && typeof v === 'object') {
    const out = {}
    for (const [k, val] of Object.entries(v)) out[k] = interpolatePlaceholders(val, recID)
    return out
  }
  return v
}

function triggerContext () {
  const recID = liveRecordID()
  const ctx = interpolatePlaceholders({ ...(props.block.options?.context || {}) }, recID)
  const rec = recordPayload(props.record)
  const values = rec?.values || {}
  if (!ctx.cidr || ctx.cidr === 'auto') {
    if (values.cidr) ctx.cidr = unwrapScalar(values.cidr)
    else if (values.target) ctx.cidr = unwrapScalar(values.target)
  } else {
    ctx.cidr = unwrapScalar(ctx.cidr)
  }
  // Prefer the record's project field (document / WBS / RFC). Using the
  // current recordID as projectID made submit-approval look up a project
  // in the documents module → API error 200: not found.
  if (!ctx.projectID || ctx.projectID === 'auto') {
    if (values.project) ctx.projectID = String(unwrapScalar(values.project) || '')
    else if (recID) ctx.projectID = recID
  }
  if (recID) {
    ctx.documentID = recID
    ctx.recordID = recID
  }
  return ctx
}

async function runChain () {
  if (!chainID.value) return
  const recID = liveRecordID()
  if (!recID) {
    result.value = {
      success: false,
      error: locFallback(
        'ruleChain.needRecord',
        'Сначала откройте или сохраните запись',
        'Open or save the record first',
      ),
    }
    notifyResult(result.value)
    return
  }
  running.value = true
  result.value = null
  pollAbort = false
  try {
    const { data } = await $ComposeAPI.api().request({
      method: 'post',
      url: $ComposeAPI.baseURL + '/pageblock/trigger',
      data: {
        chainID: chainID.value,
        pageID: props.page?.pageID,
        moduleID: props.module?.moduleID,
        namespaceID: props.namespace?.namespaceID,
        recordID: recID,
        userID: window.__auth?.user?.userID,
        record: recordPayload(props.record),
        context: triggerContext(),
      },
    })
    result.value = unwrapTrigger(data)
    if (result.value.success) {
      const detail = { stayOnPage: true }
      window.dispatchEvent(new CustomEvent('refetch-records', { detail }))
      bus.$emit('refetch-records', detail)
    }
    const ids = scanIDsFromTrigger(result.value)
    const isScan = chainID.value === 'cmdb-trigger-scan' || ids.scanID
    if (result.value.success && isScan && ids.scanID && !pollAbort) {
      const agentUrl = props.block.options?.context?.agentUrl || 'http://localhost:8085/api'
      result.value = { success: true, output: 'Сканирование запущено, загрузка из CMDB API…' }
      const pulled = await pullScanResultsIntoCompose({
        $ComposeAPI,
        namespaceID: props.namespace?.namespaceID,
        agentUrl,
        scanID: ids.scanID,
        composeScanRecordID: ids.composeScanRecordID,
        onProgress (s) {
          const pct = s.progress != null ? Math.round(s.progress) : 0
          const found = s.found != null ? s.found : 0
          const tgt = s.target ? ` · ${s.target}` : ''
          result.value = { success: true, output: `Сканирование ${pct}% · найдено ${found}${tgt}` }
        },
      })
      const st = pulled.status?.status || pulled.status?.Status
      const why = pulled.status?.error || pulled.status?.message || pulled.status?.Error
      if (st === 'error' || st === 'failed') {
        result.value = { success: false, error: why || 'Скан завершился с ошибкой' }
      } else if (!pulled.found) {
        result.value = { success: true, output: why ? `Найдено 0 устройств. ${why}` : 'Найдено 0 устройств' }
      } else {
        result.value = { success: true, output: `Загружено устройств: ${pulled.found}` }
      }
    } else if (result.value.success && isScan && !ids.scanID) {
      result.value = { success: false, error: 'Агент не вернул scanID. Проверьте CMDB agent на :8085 и что цепочка POST /api/scan проходит.' }
    }
    notifyResult(result.value)
    if (result.value.success) {
      reloadPageIfNeeded()
    }
  } catch (err) {
    result.value = { success: false, error: err.message || errorLabel.value }
    notifyResult(result.value)
  } finally {
    running.value = false
  }
}
</script>

<style scoped>
.rulechain-body {
  padding: 0.5rem 0.75rem;
  min-height: 0;
}
.rulechain-status {
  flex: 1 1 10rem;
  min-width: 8rem;
  padding: 0.3rem 0.7rem;
  border-radius: 0.5rem;
  font-size: 0.8125rem;
  line-height: 1.3;
  background: var(--bs-secondary-bg, #f8f9fa);
}
.rulechain-status-text {
  white-space: pre-wrap;
  word-break: break-word;
}
.rulechain-status.is-success {
  color: var(--bs-success-text-emphasis, #0a3622);
  background: var(--bs-success-bg-subtle, #d1e7dd);
}
.rulechain-status.is-danger {
  color: var(--bs-danger-text-emphasis, #58151c);
  background: var(--bs-danger-bg-subtle, #f8d7da);
}
.rulechain-balloon {
  position: fixed;
  top: 4.5rem;
  right: 1rem;
  z-index: 20000;
  width: min(28rem, calc(100vw - 2rem));
  border-radius: 0.5rem;
  box-shadow: 0 0.5rem 1.5rem rgba(0, 0, 0, 0.25);
  overflow: hidden;
  color: #fff;
}
.rulechain-balloon.is-success {
  background: #198754;
}
.rulechain-balloon.is-danger {
  background: #dc3545;
}
.rulechain-balloon-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.6rem 0.85rem 0.25rem;
  font-size: 0.95rem;
}
.rulechain-balloon-body {
  padding: 0.25rem 0.85rem 0.85rem;
  font-size: 0.9rem;
  line-height: 1.4;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
