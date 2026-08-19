<template>
  <Wrap v-bind="$props">
    <div class="p-3">
      <div v-if="result" class="mb-3 p-2 border rounded" :class="result.success ? 'border-success bg-success-subtle' : 'border-danger bg-danger-subtle'">
        <div v-if="result.success" class="text-success">
          <font-awesome-icon :icon="['fas', 'check-circle']" class="me-1" />
          {{ successText }}
        </div>
        <div v-else class="text-danger">
          <font-awesome-icon :icon="['fas', 'exclamation-circle']" class="me-1" />
          {{ result.error || 'Ошибка выполнения' }}
        </div>
        <pre v-if="result.output && typeof result.output === 'object' && !reorderSummary && !riskSummary" class="mt-2 mb-0 small">{{ formatOutput(result.output) }}</pre>
      </div>

      <button
        class="btn"
        :class="btnClass"
        :disabled="running"
        @click="runChain"
      >
        <span v-if="running" class="spinner-border spinner-border-sm me-1" role="status" />
        <font-awesome-icon :icon="icon" class="me-1" />
        {{ label }}
      </button>
    </div>
  </Wrap>
</template>

<script setup>
import { ref, computed, inject } from 'vue'
import { NoID } from 'corteza-lib/js/dist'
import Wrap from './Wrap/index.js'
import { scanIDsFromTrigger } from './cmdbAgentSync.js'

const props = defineProps({
  blockIndex: { type: Number, default: -1 },
  namespace: { type: Object, required: true },
  page: { type: Object, required: true },
  blocks: { type: Array, default: () => [] },
  block: { type: Object, required: true },
  module: { type: Object, default: undefined },
  record: { type: Object, default: undefined },
})

const $ComposeAPI = inject('$ComposeAPI')

const running = ref(false)
const result = ref(null)

const chainID = computed(() => props.block.options?.chainID || '')
const label = computed(() => props.block.options?.label || 'Run Rule Chain')
const icon = computed(() => ['fas', props.block.options?.icon || 'play'])
const btnClass = computed(() => `btn-${props.block.options?.variant || 'primary'} ${(props.block.options?.size) ? 'btn-' + props.block.options.size : ''}`)

const reorderSummary = computed(() => {
  const out = result.value?.output
  if (!out || typeof out !== 'object') return null
  if (out.orderCount == null && out.OrderCount == null) return null
  return {
    orderCount: Number(out.orderCount ?? out.OrderCount ?? 0),
    lineCount: Number(out.lineCount ?? out.LineCount ?? 0),
    totalQty: Number(out.totalQty ?? out.TotalQty ?? 0),
    totalSum: Number(out.totalSum ?? out.TotalSum ?? 0),
  }
})

const riskSummary = computed(() => {
  const out = result.value?.output
  if (!out || typeof out !== 'object') return null
  if (out.level == null && out.residualScore == null && out.score == null) return null
  return {
    level: out.level || '',
    residual: out.residualScore ?? out.residual,
    score: out.score,
    name: out.name || '',
  }
})

const successText = computed(() => {
  const s = reorderSummary.value
  if (s) {
    const qty = Number.isFinite(s.totalQty) ? s.totalQty.toLocaleString('ru-RU') : s.totalQty
    const sum = Number.isFinite(s.totalSum) ? s.totalSum.toLocaleString('ru-RU', { maximumFractionDigits: 2 }) : s.totalSum
    return `Создано заказов: ${s.orderCount}, строк: ${s.lineCount}, кол-во: ${qty}, сумма: ${sum}`
  }
  const r = riskSummary.value
  if (r) {
    const who = r.name ? `${r.name}: ` : ''
    const res = r.residual != null ? Number(r.residual).toLocaleString('ru-RU', { maximumFractionDigits: 1 }) : '—'
    const inn = r.score != null ? Number(r.score).toLocaleString('ru-RU', { maximumFractionDigits: 1 }) : '—'
    return `${who}риск ${r.level || 'n/a'} (residual ${res}, inherent ${inn})`
  }
  if (typeof result.value?.output === 'string' && result.value.output) {
    return result.value.output
  }
  return 'Цепочка выполнена успешно'
})

function formatOutput (output) {
  if (output == null) return ''
  if (typeof output === 'string') return output
  try {
    return JSON.stringify(output, null, 2)
  } catch {
    return String(output)
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
    recordID: record.recordID,
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

function triggerContext () {
  const ctx = { ...(props.block.options?.context || {}) }
  const rec = recordPayload(props.record)
  const values = rec?.values || {}
  if (!ctx.cidr || ctx.cidr === 'auto') {
    if (values.cidr) ctx.cidr = unwrapScalar(values.cidr)
    else if (values.target) ctx.cidr = unwrapScalar(values.target)
  } else {
    ctx.cidr = unwrapScalar(ctx.cidr)
  }
  return ctx
}

async function runChain () {
  if (!chainID.value) return
  running.value = true
  result.value = null
  try {
    const { data } = await $ComposeAPI.api().request({
      method: 'post',
      url: $ComposeAPI.baseURL + '/pageblock/trigger',
      data: {
        chainID: chainID.value,
        pageID: props.page?.pageID,
        moduleID: props.module?.moduleID,
        namespaceID: props.namespace?.namespaceID,
        recordID: props.record?.recordID && props.record.recordID !== NoID ? props.record.recordID : undefined,
        record: recordPayload(props.record),
        context: triggerContext(),
      },
    })
    result.value = data?.response || data || { success: true }
    if (result.value.success) {
      window.dispatchEvent(new CustomEvent('refetch-records', { detail: { stayOnPage: true } }))
    }
    const ids = scanIDsFromTrigger(result.value)
    const isScan = chainID.value === 'cmdb-trigger-scan' || ids.scanID
    if (result.value.success && isScan && ids.scanID) {
      result.value = { success: true, output: 'Сканирование запущено. Устройства появятся в списке по мере ingest на сервере.' }
    } else if (result.value.success && isScan && !ids.scanID) {
      result.value = { success: false, error: 'Агент не вернул scanID. Проверьте CMDB agent на :8085 и что цепочка POST /api/scan проходит.' }
    }
  } catch (err) {
    result.value = { success: false, error: err.message || 'Ошибка запроса' }
  } finally {
    running.value = false
  }
}
</script>
