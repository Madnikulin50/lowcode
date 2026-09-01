<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom d-flex align-items-center justify-content-between flex-wrap gap-2">
      <h4 class="m-0">{{ t('title') }}</h4>
      <div class="d-flex gap-2">
        <button
          type="button"
          class="btn btn-outline-secondary btn-sm"
          :disabled="discovering || !canManage"
          @click="syncFromOllama"
        >
          <span v-if="discovering" class="spinner-border spinner-border-sm me-1" />
          {{ t('sync') }}
        </button>
      </div>
    </div>

    <form @submit.prevent="onSave">
      <div class="card-body">
        <div v-if="discoverError" class="alert alert-warning py-2">{{ discoverError }}</div>

        <div class="form-check form-switch mb-4">
          <input
            id="ai-enabled"
            v-model="local['ai.enabled']"
            class="form-check-input"
            type="checkbox"
            :disabled="!canManage"
          >
          <label class="form-check-label" for="ai-enabled">{{ t('enabled') }}</label>
          <div class="form-text">{{ t('enabledHelp') }}</div>
        </div>

        <div class="mb-4">
          <label class="form-label text-primary">{{ t('ollamaURL') }}</label>
          <input
            v-model="local['ai.ollama-url']"
            class="form-control"
            type="text"
            :placeholder="effectiveOllamaURL || t('ollamaURLPlaceholder')"
            :disabled="!canManage"
          >
          <div class="form-text">
            {{ t('ollamaURLHelp') }}
            <span v-if="effectiveOllamaURL" class="d-block mt-1">
              {{ t('ollamaURLEffective', { url: effectiveOllamaURL, source: effectiveOllamaFrom }) }}
            </span>
          </div>
        </div>

        <h5 class="mb-3">{{ t('roles.title') }}</h5>
        <p class="text-muted small">{{ t('roles.help') }}</p>
        <div class="row g-3 mb-4">
          <div
            v-for="role in roleFields"
            :key="role.key"
            class="col-12 col-md-6"
          >
            <label class="form-label">{{ t(role.labelKey) }}</label>
            <select
              v-model="local[role.key]"
              class="form-select"
              :disabled="!canManage || discovering"
            >
              <option value="">{{ t('roles.fallback') }}</option>
              <option
                v-for="m in roleModelOptions"
                :key="m.name"
                :value="m.name"
              >
                {{ m.label || m.name }}
              </option>
            </select>
            <div class="form-text">{{ t(role.helpKey) }}</div>
          </div>
        </div>

        <div class="d-flex align-items-center justify-content-between mb-2">
          <h5 class="m-0">{{ t('catalog.title') }}</h5>
          <button
            type="button"
            class="btn btn-outline-primary btn-sm"
            :disabled="!canManage"
            @click="addManualModel"
          >
            {{ t('catalog.add') }}
          </button>
        </div>
        <p class="text-muted small">{{ t('catalog.help') }}</p>

        <div v-if="discovering && !catalog.length" class="text-muted mb-3">
          {{ t('catalog.loading') }}
        </div>
        <div v-else-if="!catalog.length" class="text-muted mb-3">
          {{ t('catalog.empty') }}
        </div>

        <div class="table-responsive mb-3">
          <table v-if="catalog.length" class="table table-sm align-middle">
            <thead>
              <tr>
                <th>{{ t('catalog.columns.name') }}</th>
                <th style="width: 5rem">{{ t('catalog.columns.enabled') }}</th>
                <th style="width: 5rem">{{ t('catalog.columns.tools') }}</th>
                <th>{{ t('catalog.columns.usedAs') }}</th>
                <th style="width: 8rem" />
              </tr>
            </thead>
            <tbody>
              <tr v-for="(row, idx) in catalog" :key="row.name + '-' + idx">
                <td>
                  <input
                    v-model="row.name"
                    class="form-control form-control-sm"
                    :disabled="!canManage"
                    @change="onCatalogNameChange"
                  >
                  <input
                    v-model="row.note"
                    class="form-control form-control-sm mt-1"
                    :placeholder="t('catalog.notePlaceholder')"
                    :disabled="!canManage"
                  >
                </td>
                <td>
                  <div class="form-check form-switch m-0">
                    <input
                      v-model="row.enabled"
                      class="form-check-input"
                      type="checkbox"
                      :disabled="!canManage"
                    >
                  </div>
                </td>
                <td>
                  <span
                    class="badge"
                    :class="toolsBadgeClass(row.name)"
                  >{{ toolsLabel(row.name) }}</span>
                </td>
                <td>
                  <span
                    v-for="r in rolesFor(row.name)"
                    :key="r"
                    class="badge text-bg-secondary me-1"
                  >{{ roleBadge(r) }}</span>
                  <span
                    v-if="row.enabled && !rolesFor(row.name).length"
                    class="text-muted small"
                  >{{ t('catalog.pickerOnly') }}</span>
                  <span
                    v-if="!row.enabled"
                    class="text-muted small"
                  >—</span>
                </td>
                <td class="text-end text-nowrap">
                  <button
                    type="button"
                    class="btn btn-outline-secondary btn-sm me-1"
                    :disabled="warming === row.name"
                    @click="warmup(row.name)"
                  >
                    {{ t('catalog.warmup') }}
                  </button>
                  <button
                    type="button"
                    class="btn btn-outline-danger btn-sm"
                    :disabled="!canManage"
                    @click="removeModel(idx)"
                  >
                    {{ t('catalog.remove') }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div class="card-footer text-end">
        <button
          type="submit"
          class="btn btn-primary"
          :disabled="!canManage || processing"
        >
          <span v-if="processing" class="spinner-border spinner-border-sm me-1" />
          <span v-if="success">{{ t('saved') }}</span>
          <span v-else>{{ t('save') }}</span>
        </button>
      </div>
    </form>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'system.ai', keyPrefix: 'editor' } })

import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { cloneDeep } from 'lodash'

const props = defineProps({
  settings: { type: Object, required: true },
  canManage: { type: Boolean, default: false },
  processing: { type: Boolean, default: false },
  success: { type: Boolean, default: false },
})

const emit = defineEmits(['submit'])
const { t } = useI18n()

const local = reactive({})
const catalog = ref([])
const discovering = ref(false)
const discoverError = ref('')
const warming = ref('')
const toolsMap = ref({})
const effectiveOllamaURL = ref('')
const effectiveOllamaFrom = ref('')
const ollamaModels = ref([]) // names discovered from Ollama

const roleFields = [
  { key: 'ai.roles.compose-chat', labelKey: 'roles.composeChat', helpKey: 'roles.composeChatHelp', id: 'compose.chat' },
  { key: 'ai.roles.mcp-agent', labelKey: 'roles.mcpAgent', helpKey: 'roles.mcpAgentHelp', id: 'mcp.agent' },
  { key: 'ai.roles.automation-chat', labelKey: 'roles.automationChat', helpKey: 'roles.automationChatHelp', id: 'automation.chat' },
  { key: 'ai.roles.rulesgo-ai', labelKey: 'roles.rulesgoAi', helpKey: 'roles.rulesgoAiHelp', id: 'rulesgo.ai' },
]

// Combobox options: enabled catalog entries, or all Ollama models when catalog is empty.
const roleModelOptions = computed(() => {
  const fromCatalog = catalog.value.filter(m => m.enabled && m.name)
  if (fromCatalog.length) {
    return fromCatalog
  }
  return ollamaModels.value.map(name => ({ name, label: name, enabled: true }))
})

watch(() => props.settings, (s) => {
  Object.keys(local).forEach(k => delete local[k])
  Object.assign(local, cloneDeep(s || {}))
  catalog.value = Array.isArray(local['ai.catalog'])
    ? cloneDeep(local['ai.catalog']).map(normalizeEntry)
    : []
}, { immediate: true, deep: true })

onMounted(() => {
  syncFromOllama({ fillEmptyCatalogOnly: true })
})

function normalizeEntry (e) {
  return {
    name: e?.name || '',
    enabled: e?.enabled !== false,
    label: e?.label || '',
    note: e?.note || '',
  }
}

function rolesFor (name) {
  const out = []
  for (const r of roleFields) {
    if (local[r.key] === name) out.push(r.id)
  }
  return out
}

function roleBadge (id) {
  const map = {
    'compose.chat': t('roles.composeChat'),
    'mcp.agent': t('roles.mcpAgent'),
    'automation.chat': t('roles.automationChat'),
    'rulesgo.ai': t('roles.rulesgoAi'),
  }
  return map[id] || id
}

function toolsLabel (name) {
  if (!(name in toolsMap.value)) return '—'
  return toolsMap.value[name] ? t('catalog.toolsYes') : t('catalog.toolsNo')
}

function toolsBadgeClass (name) {
  if (!(name in toolsMap.value)) return 'text-bg-light text-muted'
  return toolsMap.value[name] ? 'text-bg-success' : 'text-bg-secondary'
}

function onCatalogNameChange () {}

function addManualModel () {
  catalog.value.push({ name: '', enabled: true, label: '', note: '' })
}

function removeModel (idx) {
  const removed = catalog.value[idx]?.name
  catalog.value.splice(idx, 1)
  for (const r of roleFields) {
    if (local[r.key] === removed) local[r.key] = ''
  }
}

function unwrapPayload (res) {
  // stdResolve already unwraps { response }, but raw axios may not
  if (res?.models) return res
  if (res?.response?.models) return res.response
  return res || {}
}

async function syncFromOllama (opts = {}) {
  const { fillEmptyCatalogOnly = false } = opts
  discovering.value = true
  discoverError.value = ''
  try {
    const api = window.__ComposeAPI
    if (!api) {
      throw new Error('Compose API is not available')
    }
    let res
    if (typeof api.pageAiDiscoverModels === 'function') {
      res = unwrapPayload(await api.pageAiDiscoverModels())
    } else {
      const raw = await api.api().request({ method: 'get', url: '/chat/models/discover' })
      res = unwrapPayload(raw?.data?.response || raw?.data)
    }
    const list = res?.models || []
    if (res?.ollamaURL) {
      effectiveOllamaURL.value = res.ollamaURL
      effectiveOllamaFrom.value = res.ollamaFrom || ''
    }

    const names = []
    const tools = { ...toolsMap.value }
    for (const m of list) {
      const name = typeof m === 'string' ? m : (m?.name || '')
      if (!name) continue
      names.push(name)
      if (typeof m === 'object' && 'tools' in m) {
        tools[name] = !!m.tools
      }
    }
    ollamaModels.value = names
    toolsMap.value = tools

    const shouldFillCatalog = !fillEmptyCatalogOnly || catalog.value.length === 0
    if (shouldFillCatalog) {
      const byName = new Map(catalog.value.filter(e => e.name).map(e => [e.name, e]))
      const next = []
      for (const name of names) {
        if (byName.has(name)) {
          next.push(byName.get(name))
          byName.delete(name)
        } else {
          next.push({ name, enabled: true, label: '', note: '' })
        }
      }
      for (const leftover of byName.values()) {
        next.push(leftover)
      }
      catalog.value = next
    }

    if (!names.length) {
      discoverError.value = t('catalog.discoverEmpty')
    }
  } catch (e) {
    discoverError.value = e?.message || String(e)
  } finally {
    discovering.value = false
  }
}

async function warmup (name) {
  if (!name) return
  warming.value = name
  try {
    await window.__ComposeAPI.pageAiWarmUp({ model: name })
  } catch (e) {
    discoverError.value = e?.message || String(e)
  } finally {
    warming.value = ''
  }
}

function onSave () {
  const cleaned = catalog.value
    .map(normalizeEntry)
    .filter(e => e.name.trim())
  const seen = new Set()
  const unique = []
  for (const e of cleaned) {
    if (seen.has(e.name)) continue
    seen.add(e.name)
    unique.push(e)
  }
  local['ai.catalog'] = unique
  delete local['ai.agents']
  delete local['ai.toolkits']
  emit('submit', cloneDeep(local))
}
</script>
