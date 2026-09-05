<template>
  <div>
  <div class="card shadow-sm">
    <div class="card-header border-bottom d-flex align-items-center justify-content-between flex-wrap gap-2">
      <h4 class="m-0">{{ t('agents.title') }}</h4>
      <button
        type="button"
        class="btn btn-outline-primary btn-sm"
        :disabled="!canManage || loading"
        @click="openEditor()"
      >
        {{ t('agents.add') }}
      </button>
    </div>

    <div class="card-body">
      <p class="text-muted small">{{ t('agents.help') }}</p>
      <div v-if="loadError" class="alert alert-warning py-2">{{ loadError }}</div>

      <div class="table-responsive">
        <table class="table table-sm table-hover align-middle mb-0">
          <thead>
            <tr>
              <th>{{ t('agents.handle') }}</th>
              <th>{{ t('agents.description') }}</th>
              <th>{{ t('agents.toolkits') }}</th>
              <th>{{ t('agents.source') }}</th>
              <th style="width: 4rem">{{ t('agents.enabled') }}</th>
              <th style="width: 9rem" />
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.handle">
              <td class="font-monospace">{{ row.handle }}</td>
              <td>{{ row.description }}</td>
              <td>
                <span
                  v-for="kit in row.toolkits"
                  :key="kit"
                  class="badge text-bg-light text-muted me-1"
                >{{ kit }}</span>
              </td>
              <td>
                <span class="badge" :class="sourceBadge(row)">{{ sourceLabel(row) }}</span>
              </td>
              <td>
                <span v-if="row.enabled" class="text-success">●</span>
                <span v-else class="text-muted">○</span>
              </td>
              <td class="text-end text-nowrap">
                <button
                  type="button"
                  class="btn btn-outline-secondary btn-sm me-1"
                  :disabled="!canManage || row.locked"
                  @click="openEditor(row)"
                >
                  {{ t('agents.edit') }}
                </button>
                <button
                  v-if="row.customized && row.builtin && !row.locked"
                  type="button"
                  class="btn btn-outline-secondary btn-sm"
                  :disabled="!canManage"
                  @click="resetRow(row)"
                >
                  {{ t('agents.reset') }}
                </button>
                <button
                  v-else-if="!row.builtin && !row.locked"
                  type="button"
                  class="btn btn-outline-danger btn-sm"
                  :disabled="!canManage"
                  @click="resetRow(row)"
                >
                  {{ t('agents.remove') }}
                </button>
              </td>
            </tr>
            <tr v-if="!rows.length">
              <td colspan="6" class="text-muted">{{ t('agents.empty') }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>

  <div
    class="modal"
    tabindex="-1"
    :class="{ show: modal.open, 'd-block': modal.open }"
    :style="modal.open ? { backgroundColor: 'rgba(0,0,0,0.5)' } : {}"
  >
    <div class="modal-dialog modal-lg modal-dialog-scrollable">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">{{ modal.title }}</h5>
          <button type="button" class="btn-close" @click="closeModal" />
        </div>
        <div class="modal-body">
          <div class="row g-3">
            <div class="col-12 col-md-6">
              <label class="form-label">{{ t('agents.handle') }}</label>
              <input
                v-model="modal.data.handle"
                class="form-control"
                :disabled="modal.existing"
              >
            </div>
            <div class="col-12 col-md-6">
              <label class="form-label">{{ t('agents.model') }}</label>
              <select v-model="modal.data.model" class="form-select">
                <option value="mcp.agent">{{ t('roles.mcpAgent') }}</option>
                <option value="compose.chat">{{ t('roles.composeChat') }}</option>
                <option value="automation.chat">{{ t('roles.automationChat') }}</option>
                <option value="rulesgo.ai">{{ t('roles.rulesgoAi') }}</option>
              </select>
            </div>
            <div class="col-6 col-md-3">
              <label class="form-label">{{ t('agents.maxSteps') }}</label>
              <input
                v-model.number="modal.data.maxSteps"
                class="form-control"
                type="number"
                min="1"
                max="32"
              >
            </div>
            <div class="col-6 col-md-3 d-flex align-items-end">
              <div class="form-check form-switch mb-2">
                <input
                  id="ag-enabled"
                  v-model="modal.data.enabled"
                  class="form-check-input"
                  type="checkbox"
                >
                <label class="form-check-label" for="ag-enabled">{{ t('agents.enabled') }}</label>
              </div>
            </div>
            <div class="col-12">
              <div class="form-check">
                <input
                  id="ag-confirm"
                  v-model="modal.data.confirm"
                  class="form-check-input"
                  type="checkbox"
                >
                <label class="form-check-label" for="ag-confirm">{{ t('agents.confirm') }}</label>
              </div>
              <div class="form-text">{{ t('agents.confirmHelp') }}</div>
            </div>
            <div class="col-12">
              <label class="form-label">{{ t('agents.description') }}</label>
              <input v-model="modal.data.description" class="form-control">
            </div>
            <div class="col-12">
              <label class="form-label">{{ t('agents.toolkits') }}</label>
              <div class="d-flex flex-wrap gap-3">
                <div
                  v-for="kit in kitOptions"
                  :key="kit"
                  class="form-check"
                >
                  <input
                    :id="'kit-' + kit"
                    class="form-check-input"
                    type="checkbox"
                    :checked="modal.data.toolkits.includes(kit)"
                    @change="toggleKit(modal.data, kit)"
                  >
                  <label class="form-check-label" :for="'kit-' + kit">{{ kit }}</label>
                </div>
              </div>
            </div>
            <div class="col-12">
              <label class="form-label">{{ t('agents.prompt') }}</label>
              <textarea
                v-model="modal.data.prompt"
                class="form-control font-monospace"
                rows="8"
              />
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-outline-secondary" @click="closeModal">
            {{ t('agents.cancel') }}
          </button>
          <button
            type="button"
            class="btn btn-primary"
            :disabled="!canManage || saving || !modal.data.handle"
            @click="saveModal"
          >
            <span v-if="saving" class="spinner-border spinner-border-sm me-1" />
            {{ t('save') }}
          </button>
        </div>
      </div>
    </div>
  </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'system.ai', keyPrefix: 'editor' } })

import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { cloneDeep } from 'lodash'

const props = defineProps({
  canManage: { type: Boolean, default: false },
})

const { t } = useI18n()

const FALLBACK_KITS = [
  'compose.records',
  'compose.schema',
  'compose.mail',
  'compose.visualize',
  'cmdb',
  'backup',
  'invest',
]

const builtins = ref([])
const files = ref([])
const overlay = ref([])
const kitOptions = ref([...FALLBACK_KITS])
const loading = ref(false)
const saving = ref(false)
const loadError = ref('')

const modal = reactive({
  open: false,
  existing: false,
  title: '',
  data: emptyAgent(),
})

const rows = computed(() => mergeRows(builtins.value, overlay.value, files.value))

onMounted(() => {
  reload()
  window.addEventListener('ai-catalog-changed', reload)
})
onBeforeUnmount(() => {
  window.removeEventListener('ai-catalog-changed', reload)
})

function emptyAgent () {
  return {
    handle: '',
    enabled: true,
    description: '',
    prompt: '',
    model: 'mcp.agent',
    toolkits: ['compose.records'],
    maxSteps: 8,
    confirm: false,
  }
}

function kitName (k) {
  if (typeof k === 'string') return k
  return (k && (k.name || k.handle)) || ''
}

function normalizeAgent (e) {
  const enabled = e?.enabled
  return {
    handle: (e?.handle || '').trim(),
    enabled: enabled !== false && enabled !== 'false',
    description: e?.description || '',
    prompt: e?.prompt || '',
    model: e?.model || 'mcp.agent',
    toolkits: Array.isArray(e?.toolkits) ? e.toolkits.map(kitName).filter(Boolean) : [],
    maxSteps: Number(e?.maxSteps) > 0 ? Number(e.maxSteps) : 8,
    confirm: !!e?.confirm,
  }
}

function mergeRows (builtinList, overlayList, fileList) {
  const map = new Map()
  for (const b of builtinList) {
    const n = normalizeAgent(b)
    if (!n.handle) continue
    map.set(n.handle, { ...n, builtin: true, customized: false, locked: false, source: 'builtin' })
  }
  for (const o of overlayList) {
    const n = normalizeAgent(o)
    if (!n.handle) continue
    const prev = map.get(n.handle)
    map.set(n.handle, {
      ...(prev || emptyAgent()),
      ...n,
      builtin: !!prev?.builtin,
      customized: true,
      locked: false,
      source: 'settings',
    })
  }
  for (const f of fileList) {
    const n = normalizeAgent(f)
    if (!n.handle) continue
    const prev = map.get(n.handle)
    map.set(n.handle, {
      ...(prev || emptyAgent()),
      ...n,
      builtin: !!prev?.builtin,
      customized: true,
      locked: true,
      source: 'file',
    })
  }
  return [...map.values()]
}

function sourceBadge (row) {
  if (row.source === 'file') return 'text-bg-warning'
  if (row.source === 'settings') return 'text-bg-primary'
  return 'text-bg-secondary'
}

function sourceLabel (row) {
  if (row.source === 'file') return t('agents.sourceFile')
  if (row.source === 'settings') return t('agents.sourceSettings')
  return t('agents.sourceBuiltin')
}

function unwrapPayload (res) {
  if (res?.builtins) return res
  if (res?.response?.builtins) return res.response
  return res || {}
}

async function loadCatalog () {
  const api = window.__ComposeAPI
  if (!api) return
  let res
  if (typeof api.pageAiAgents === 'function') {
    res = unwrapPayload(await api.pageAiAgents())
  } else {
    const raw = await api.api().request({ method: 'get', url: '/chat/agents' })
    res = unwrapPayload(raw?.data?.response || raw?.data)
  }
  builtins.value = Array.isArray(res.builtins) ? res.builtins : []
  files.value = Array.isArray(res.files) ? res.files : []
  const names = new Set(FALLBACK_KITS)
  for (const k of (Array.isArray(res.toolkits) ? res.toolkits : [])) {
    const n = kitName(k)
    if (n) names.add(n)
  }
  for (const k of (Array.isArray(res.kits) ? res.kits : [])) {
    const n = kitName(k)
    if (n) names.add(n)
  }
  for (const c of (Array.isArray(res.connectors) ? res.connectors : [])) {
    if (c?.handle) names.add(c.handle)
  }
  kitOptions.value = [...names]
}

async function loadOverlay () {
  const list = await window.__SystemAPI.settingsList({ prefix: 'ai.agents' })
  const row = Array.isArray(list) ? list.find(s => s.name === 'ai.agents') : null
  overlay.value = Array.isArray(row?.value) ? cloneDeep(row.value).map(normalizeAgent) : []
}

async function reload () {
  loading.value = true
  loadError.value = ''
  try {
    await Promise.all([loadCatalog(), loadOverlay()])
  } catch (e) {
    loadError.value = e?.message || String(e)
  } finally {
    loading.value = false
  }
}

function openEditor (row) {
  if (row?.locked) return
  if (row) {
    modal.existing = true
    modal.title = row.handle
    modal.data = normalizeAgent(row)
  } else {
    modal.existing = false
    modal.title = t('agents.add')
    modal.data = emptyAgent()
  }
  modal.open = true
}

function closeModal () {
  modal.open = false
}

function toggleKit (row, kit) {
  const i = row.toolkits.indexOf(kit)
  if (i >= 0) {
    row.toolkits.splice(i, 1)
  } else {
    row.toolkits.push(kit)
  }
}

async function persistOverlay (next) {
  saving.value = true
  try {
    await window.__SystemAPI.settingsUpdate({
      values: [{ name: 'ai.agents', value: next }],
    })
    overlay.value = next
    closeModal()
  } catch (e) {
    loadError.value = e?.message || String(e)
  } finally {
    saving.value = false
  }
}

function saveModal () {
  const spec = normalizeAgent(modal.data)
  if (!spec.handle) return
  const next = overlay.value.filter(a => a.handle !== spec.handle)
  next.push(spec)
  persistOverlay(next)
}

function resetRow (row) {
  persistOverlay(overlay.value.filter(a => a.handle !== row.handle))
}
</script>
