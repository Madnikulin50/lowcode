<template>
  <div>
  <div class="card shadow-sm">
    <div class="card-header border-bottom d-flex align-items-center justify-content-between flex-wrap gap-2">
      <h4 class="ae-section-title">{{ t('toolkits.title') }}</h4>
      <button
        type="button"
        class="btn btn-outline-primary btn-sm"
        :disabled="!canManage || loading"
        @click="openEditor()"
      >
        {{ t('toolkits.add') }}
      </button>
    </div>

    <div class="card-body">
      <p class="text-muted small">{{ t('toolkits.help') }}</p>
      <div v-if="loadError" class="alert alert-warning py-2">{{ loadError }}</div>

      <div class="table-responsive">
        <table class="table table-sm table-hover align-middle mb-0">
          <thead>
            <tr>
              <th>{{ t('toolkits.handle') }}</th>
              <th>{{ t('toolkits.url') }}</th>
              <th>{{ t('toolkits.source') }}</th>
              <th style="width: 4rem">{{ t('toolkits.enabled') }}</th>
              <th style="width: 9rem" />
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.handle">
              <td class="font-monospace">{{ row.handle }}</td>
              <td class="text-break small">{{ row.url }}</td>
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
                  :disabled="!canManage"
                  @click="openEditor(row)"
                >
                  {{ t('toolkits.edit') }}
                </button>
                <button
                  v-if="row.customized && row.seeded"
                  type="button"
                  class="btn btn-outline-secondary btn-sm"
                  :disabled="!canManage"
                  @click="resetRow(row)"
                >
                  {{ t('toolkits.reset') }}
                </button>
                <button
                  v-else-if="!row.seeded"
                  type="button"
                  class="btn btn-outline-danger btn-sm"
                  :disabled="!canManage"
                  @click="resetRow(row)"
                >
                  {{ t('toolkits.remove') }}
                </button>
              </td>
            </tr>
            <tr v-if="!rows.length">
              <td colspan="5" class="text-muted">{{ t('toolkits.empty') }}</td>
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
          <div v-if="modal.error" class="alert alert-danger py-2">{{ modal.error }}</div>
          <div class="row g-3">
            <div class="col-12 col-md-6">
              <label class="form-label">{{ t('toolkits.handle') }}</label>
              <input
                v-model="modal.data.handle"
                class="form-control"
                :disabled="modal.existing"
              >
              <div class="form-text">{{ t('toolkits.handleHelp') }}</div>
            </div>
            <div class="col-12 col-md-6 d-flex align-items-end">
              <div class="form-check form-switch mb-2">
                <input
                  id="tk-enabled"
                  v-model="modal.data.enabled"
                  class="form-check-input"
                  type="checkbox"
                >
                <label class="form-check-label" for="tk-enabled">{{ t('toolkits.enabled') }}</label>
              </div>
            </div>
            <div class="col-12">
              <label class="form-label">{{ t('toolkits.url') }}</label>
              <input
                v-model="modal.data.url"
                class="form-control font-monospace"
                placeholder="http://localhost:8085/api"
              >
              <div class="form-text">{{ t('toolkits.urlHelp') }}</div>
            </div>
            <div class="col-12">
              <label class="form-label">{{ t('toolkits.token') }}</label>
              <input
                v-model="modal.data.token"
                class="form-control"
                type="password"
                autocomplete="off"
              >
              <div class="form-text">{{ t('toolkits.tokenHelp') }}</div>
            </div>
            <div class="col-12">
              <button
                type="button"
                class="btn btn-outline-secondary btn-sm"
                :disabled="probing || !modal.data.url"
                @click="probe"
              >
                <span v-if="probing" class="spinner-border spinner-border-sm me-1" />
                {{ t('toolkits.test') }}
              </button>
              <div v-if="probeError" class="text-danger small mt-2">{{ probeError }}</div>
              <div v-else-if="probeTools.length" class="mt-2">
                <span
                  v-for="name in probeTools"
                  :key="name"
                  class="badge text-bg-light text-muted me-1"
                >{{ name }}</span>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-outline-secondary" @click="closeModal">
            {{ t('toolkits.cancel') }}
          </button>
          <button
            type="button"
            class="btn btn-primary"
            :disabled="!canManage || saving || !modal.data.handle || !modal.data.url"
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

import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { cloneDeep } from 'lodash'

const props = defineProps({
  canManage: { type: Boolean, default: false },
})

const { t } = useI18n()

const seed = ref([])
const overlay = ref([])
const loading = ref(false)
const saving = ref(false)
const loadError = ref('')
const probing = ref(false)
const probeError = ref('')
const probeTools = ref([])

const modal = reactive({
  open: false,
  existing: false,
  title: '',
  error: '',
  data: emptyToolkit(),
})

const rows = computed(() => mergeRows(seed.value, overlay.value))

onMounted(() => { reload() })

function emptyToolkit () {
  return {
    handle: '',
    url: '',
    enabled: true,
    token: '',
  }
}

function normalizeToolkit (e) {
  const enabled = e?.enabled
  return {
    handle: (e?.handle || '').trim(),
    url: String(e?.url || '').trim().replace(/\/+$/, ''),
    enabled: enabled !== false && enabled !== 'false',
    token: e?.token || '',
  }
}

function isReserved (handle) {
  return String(handle || '').toLowerCase().startsWith('compose.')
}

function mergeRows (seedList, overlayList) {
  const map = new Map()
  for (const s of seedList) {
    const n = normalizeToolkit(s)
    if (!n.handle || isReserved(n.handle)) continue
    map.set(n.handle, { ...n, seeded: true, customized: false, source: 'env' })
  }
  for (const o of overlayList) {
    const n = normalizeToolkit(o)
    if (!n.handle || isReserved(n.handle)) continue
    const prev = map.get(n.handle)
    map.set(n.handle, {
      ...(prev || emptyToolkit()),
      ...n,
      token: n.token || prev?.token || '',
      seeded: !!prev?.seeded,
      customized: true,
      source: 'settings',
    })
  }
  return [...map.values()]
}

function sourceBadge (row) {
  if (row.source === 'settings') return 'text-bg-primary'
  return 'text-bg-secondary'
}

function sourceLabel (row) {
  if (row.source === 'settings') return t('toolkits.sourceSettings')
  return t('toolkits.sourceEnv')
}

function unwrapPayload (res) {
  if (res?.connectors || res?.toolkits) return res
  if (res?.response) return res.response
  return res || {}
}

async function loadSeed () {
  const api = window.__ComposeAPI
  if (!api) return
  let res
  if (typeof api.pageAiAgents === 'function') {
    res = unwrapPayload(await api.pageAiAgents())
  } else {
    const raw = await api.api().request({ method: 'get', url: '/chat/agents' })
    res = unwrapPayload(raw?.data?.response || raw?.data)
  }
  seed.value = Array.isArray(res.connectors) ? res.connectors : []
}

async function loadOverlay () {
  const list = await window.__SystemAPI.settingsList({ prefix: 'ai.toolkits' })
  const row = Array.isArray(list) ? list.find(s => s.name === 'ai.toolkits') : null
  overlay.value = Array.isArray(row?.value) ? cloneDeep(row.value).map(normalizeToolkit) : []
}

async function reload () {
  loading.value = true
  loadError.value = ''
  try {
    await Promise.all([loadSeed(), loadOverlay()])
  } catch (e) {
    loadError.value = e?.message || String(e)
  } finally {
    loading.value = false
  }
}

function openEditor (row) {
  probeError.value = ''
  probeTools.value = []
  modal.error = ''
  if (row) {
    modal.existing = true
    modal.title = row.handle
    modal.data = normalizeToolkit(row)
  } else {
    modal.existing = false
    modal.title = t('toolkits.add')
    modal.data = emptyToolkit()
  }
  modal.open = true
}

function closeModal () {
  modal.open = false
}

function validURL (raw) {
  try {
    const u = new URL(String(raw || '').trim())
    return u.protocol === 'http:' || u.protocol === 'https:'
  } catch {
    return false
  }
}

async function probe () {
  probing.value = true
  probeError.value = ''
  probeTools.value = []
  try {
    const api = window.__ComposeAPI
    if (!api) throw new Error('Compose API is not available')
    const raw = await api.api().request({
      method: 'post',
      url: '/chat/toolkits/probe',
      data: { url: modal.data.url, token: modal.data.token },
    })
    const res = unwrapPayload(raw?.data?.response || raw?.data || raw)
    if (res?.error) throw new Error(res.error)
    if (!modal.data.handle && res.handle) {
      modal.data.handle = res.handle
    }
    probeTools.value = Array.isArray(res.tools) ? res.tools : []
    if (!probeTools.value.length) {
      probeError.value = t('toolkits.testEmpty')
    }
  } catch (e) {
    probeError.value = e?.message || String(e)
  } finally {
    probing.value = false
  }
}

async function persistOverlay (next) {
  saving.value = true
  loadError.value = ''
  try {
    await window.__SystemAPI.settingsUpdate({
      values: [{ name: 'ai.toolkits', value: next }],
    })
    overlay.value = next
    closeModal()
    window.dispatchEvent(new CustomEvent('ai-catalog-changed'))
  } catch (e) {
    loadError.value = e?.message || String(e)
    modal.error = e?.message || String(e)
  } finally {
    saving.value = false
  }
}

function saveModal () {
  const spec = normalizeToolkit(modal.data)
  modal.error = ''
  if (!spec.handle) return
  if (isReserved(spec.handle)) {
    modal.error = t('toolkits.reserved')
    return
  }
  if (!validURL(spec.url)) {
    modal.error = t('toolkits.urlInvalid')
    return
  }
  const next = overlay.value.filter(a => a.handle !== spec.handle)
  next.push(spec)
  persistOverlay(next)
}

function resetRow (row) {
  persistOverlay(overlay.value.filter(a => a.handle !== row.handle))
}
</script>
