<template>
  <div
    v-if="namespace"
    class="py-3 d-flex flex-column flex-grow-1"
    style="min-height: 0"
  >
    <Teleport to="#topbar-title">
      {{ isEdit ? $t('rulechain.edit.title') : $t('rulechain.edit.createTitle') }}
    </Teleport>

    <div
      v-if="loading"
      class="d-flex align-items-center justify-content-center h-100"
    >
      <span class="spinner-border" />
    </div>

    <div
      v-else
      class="d-flex flex-column flex-grow-1"
      style="min-height: 0"
      @submit.prevent="handleSave()"
    >
      <div class="container-fluid flex-grow-1 d-flex flex-column" style="min-height: 0">
        <div class="row flex-grow-1" style="min-height: 0">
          <div class="col d-flex flex-column" style="min-height: 0">
            <div class="card shadow-sm d-flex flex-column flex-grow-1" style="min-height: 0">
              <div class="card-header d-flex py-3 align-items-center border-bottom gap-2">
                <button
                  class="btn btn-outline-primary"
                  :disabled="processing"
                  @click="openTestModal"
                >
                  <font-awesome-icon :icon="['fas', 'play']" class="me-1" />
                  {{ $t('rulechain.test.title') }}
                </button>
              </div>

              <div class="overflow-auto p-3" style="flex: 1 1 0%; min-height: 0;">
                <div class="row pb-3">
                  <div class="col-12 col-lg-6">
                    <h5>
                      {{ $t('rulechain.edit.generalSettings') }}
                    </h5>

                    <div class="mb-3">
                      <label class="form-label text-primary">
                        {{ $t('rulechain.edit.name.label') }} *
                      </label>
                      <input
                        v-model="form.name"
                        type="text"
                        class="form-control"
                        :placeholder="$t('rulechain.edit.name.placeholder')"
                        required
                      />
                    </div>

                    <div class="mb-3">
                      <label class="form-label text-primary">
                        {{ $t('rulechain.edit.description.label') }}
                      </label>
                      <textarea
                        v-model="form.description"
                        class="form-control"
                        rows="3"
                        :placeholder="$t('rulechain.edit.description.placeholder')"
                      />
                    </div>

                    <div class="mb-3">
                      <label class="form-label text-primary">
                        {{ $t('rulechain.edit.entryNode.label') }}
                      </label>
                      <select
                        v-model="form.entryNode"
                        class="form-select"
                        :disabled="!form.nodes.length"
                      >
                        <option value="">
                          {{ $t('rulechain.edit.entryNode.placeholder') }}
                        </option>
                        <option
                          v-for="node in form.nodes"
                          :key="node.id"
                          :value="node.id"
                        >
                          {{ node.label || node.id }}
                        </option>
                      </select>
                    </div>
                  </div>
                </div>

                <hr />

                <div class="d-flex justify-content-between align-items-center mb-2">
                  <h5 class="mb-0">
                    {{ $t('rulechain.edit.nodes.label') }} ({{ form.nodes.length }})
                  </h5>
                  <button
                    class="btn btn-sm btn-outline-primary"
                    type="button"
                    @click="addNode"
                  >
                    <font-awesome-icon :icon="['fas', 'plus']" class="me-1" />
                    {{ $t('rulechain.edit.nodes.add') }}
                  </button>
                </div>

                <div
                  v-for="(node, ix) in form.nodes"
                  :key="node.id"
                  class="card mb-2"
                >
                  <div class="card-body py-2">
                    <div class="row g-2">
                      <div class="col-12 col-md-3">
                        <input
                          v-model="node.label"
                          type="text"
                          class="form-control form-control-sm"
                          :placeholder="$t('rulechain.edit.nodes.placeholder')"
                        />
                      </div>
                      <div class="col-12 col-md-3">
                        <select
                          v-model="node.type"
                          class="form-select form-select-sm"
                        >
                          <option value="">
                            {{ $t('rulechain.edit.nodes.type.placeholder') }}
                          </option>
                          <option
                            v-for="nt in nodeTypes"
                            :key="nt.type"
                            :value="nt.type"
                          >
                            {{ nt.label }} ({{ nt.type }})
                          </option>
                        </select>
                      </div>
                      <div class="col-12 col-md-4">
                        <textarea
                          v-model="node.configText"
                          rows="2"
                          class="form-control form-control-sm font-monospace"
                          :placeholder="$t('rulechain.edit.nodes.config.placeholder')"
                          spellcheck="false"
                        />
                      </div>
                      <div class="col-12 col-md-2 d-flex align-items-center gap-2">
                        <span class="text-muted small text-nowrap">
                          {{ node.id }}
                        </span>
                        <button
                          class="btn btn-sm btn-outline-danger ms-auto"
                          type="button"
                          :title="$t('rulechain.edit.nodes.delete')"
                          @click="removeNode(ix)"
                        >
                          <font-awesome-icon :icon="['fas', 'trash']" />
                        </button>
                      </div>
                      <div
                        v-if="nodeTypesByType[node.type]"
                        class="col-12"
                      >
                        <small class="text-muted">
                          {{ nodeTypesByType[node.type].description }}
                        </small>
                        <small
                          v-if="hasConfigSchema(node)"
                          class="d-block text-muted"
                        >
                          <code>
                            {{ formatConfigSchema(node) }}
                          </code>
                        </small>
                      </div>
                    </div>
                  </div>
                </div>

                <hr />

                <div class="d-flex justify-content-between align-items-center mb-2">
                  <h5 class="mb-0">
                    {{ $t('rulechain.edit.edges.label') }} ({{ form.edges.length }})
                  </h5>
                  <button
                    class="btn btn-sm btn-outline-primary"
                    type="button"
                    @click="addEdge"
                  >
                    <font-awesome-icon :icon="['fas', 'plus']" class="me-1" />
                    {{ $t('rulechain.edit.edges.add') }}
                  </button>
                </div>

                <div
                  v-for="(edge, ix) in form.edges"
                  :key="ix"
                  class="card mb-2"
                >
                  <div class="card-body py-2">
                    <div class="row g-2">
                      <div class="col-12 col-md-3">
                        <select
                          v-model="edge.from"
                          class="form-select form-select-sm"
                        >
                          <option value="">
                            {{ $t('rulechain.edit.edges.from.placeholder') }}
                          </option>
                          <option
                            v-for="node in form.nodes"
                            :key="node.id"
                            :value="node.id"
                          >
                            {{ node.label || node.id }}
                          </option>
                        </select>
                      </div>
                      <div class="col-12 col-md-3">
                        <select
                          v-model="edge.to"
                          class="form-select form-select-sm"
                        >
                          <option value="">
                            {{ $t('rulechain.edit.edges.to.placeholder') }}
                          </option>
                          <option
                            v-for="node in form.nodes"
                            :key="node.id"
                            :value="node.id"
                          >
                            {{ node.label || node.id }}
                          </option>
                        </select>
                      </div>
                      <div class="col-12 col-md-2">
                        <input
                          v-model="edge.label"
                          type="text"
                          class="form-control form-control-sm"
                          :placeholder="$t('rulechain.edit.edges.placeholder')"
                        />
                      </div>
                      <div class="col-12 col-md-3">
                        <input
                          v-model="edge.condition"
                          type="text"
                          class="form-control form-control-sm font-monospace"
                          :placeholder="$t('rulechain.edit.edges.condition.placeholder')"
                          spellcheck="false"
                        />
                      </div>
                      <div class="col-12 col-md-1 d-flex align-items-center">
                        <button
                          class="btn btn-sm btn-outline-danger"
                          type="button"
                          :title="$t('rulechain.edit.edges.delete')"
                          @click="form.edges.splice(ix, 1)"
                        >
                          <font-awesome-icon :icon="['fas', 'trash']" />
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <Teleport to="#admin-toolbar">
      <editor-toolbar
        :processing="processing"
        :processing-save="processingSave"
        :processing-save-and-close="processingSaveAndClose"
        :processing-delete="processingDelete"
        :hide-delete="!isEdit"
        :hide-clone="true"
        :disable-save="!form.name"
        @delete="handleDelete()"
        @save="handleSave()"
        @saveAndClose="handleSave({ closeOnSuccess: true })"
        @back="router.push({ name: 'admin.rulechains' })"
      />
    </Teleport>

    <div
      ref="testModalEl"
      class="modal fade"
      tabindex="-1"
      aria-hidden="true"
    >
      <div class="modal-dialog modal-lg modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">
              {{ $t('rulechain.test.modal.title') }}
            </h5>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
              aria-label="Close"
            />
          </div>
          <div class="modal-body">
            <label class="form-label text-primary">
              {{ $t('rulechain.test.input.label') }}
            </label>
            <textarea
              v-model="testInput"
              rows="6"
              class="form-control font-monospace mb-3"
              spellcheck="false"
              :placeholder="$t('rulechain.test.input.placeholder')"
            />

            <label class="form-label text-primary">
              {{ $t('rulechain.test.output.label') }}
            </label>
            <pre
              class="bg-light border rounded p-2 mb-0"
              style="max-height: 40vh; overflow: auto; white-space: pre-wrap;"
            >{{ testResult || '—' }}</pre>
          </div>
          <div class="modal-footer">
            <button
              type="button"
              class="btn btn-primary"
              :disabled="testRunning"
              @click="runTest"
            >
              <span
                v-if="testRunning"
                class="spinner-border spinner-border-sm me-1"
              />
              <font-awesome-icon
                v-else
                :icon="['fas', 'play']"
                class="me-1"
              />
              {{ $t('rulechain.test.run') }}
            </button>
            <button
              type="button"
              class="btn btn-secondary"
              data-bs-dismiss="modal"
            >
              {{ $t('label.close', 'Close') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, reactive, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Modal } from 'bootstrap'
import { NoID } from 'corteza-lib/js/dist'
import { composables } from 'corteza-lib/vue/dist'
import EditorToolbar from 'corteza-webapp-compose/src/components/Admin/EditorToolbar'

const { useToast } = composables
const { t } = useI18n()
const router = useRouter()

const $ComposeAPI = window.__composeAPI

const props = defineProps({
  namespace: {
    type: Object,
    required: false,
    default: undefined,
  },
  chainID: {
    type: String,
    required: false,
    default: NoID,
  },
})

const isEdit = computed(() => props.chainID && props.chainID !== NoID)

const loading = ref(false)
const processing = ref(false)
const processingSave = ref(false)
const processingSaveAndClose = ref(false)
const processingDelete = ref(false)

const form = reactive({
  name: '',
  description: '',
  entryNode: '',
  nodes: [],
  edges: [],
})

const nodeTypes = ref([])
const nodeTypesByType = computed(() => {
  const map = {}
  for (const nt of nodeTypes.value) map[nt.type] = nt
  return map
})

const testModalEl = ref(null)
const testModal = ref(undefined)
const testInput = ref('')
const testResult = ref('')
const testRunning = ref(false)

let nodeCounter = 0

const { toastSuccess, toastErrorHandler } = useToast()

watch(() => props.chainID, () => { fetchChain() }, { immediate: true })

onMounted(() => {
  document.title = t('label.app-name.rulechains', { label: props.namespace?.name, interpolation: { escapeValue: false } })
  loadNodeTypes()
  testModal.value = new Modal(testModalEl.value)
  testModalEl.value.addEventListener('hidden.bs.modal', () => {
    testResult.value = ''
    testRunning.value = false
  })
})

onBeforeUnmount(() => {
  testModal.value?.dispose()
})

function loadNodeTypes () {
  $ComposeAPI.ruleChainNodeTypes()
    .then(({ nodes }) => { nodeTypes.value = nodes || [] })
    .catch(() => {})
}

function fetchChain () {
  if (!isEdit.value) {
    form.name = ''
    form.description = ''
    form.entryNode = ''
    form.nodes = []
    form.edges = []
    return
  }
  loading.value = true
  $ComposeAPI.ruleChainRead({ chainID: props.chainID })
    .then(({ chain }) => {
      form.name = chain.name || ''
      form.description = chain.description || ''
      form.entryNode = chain.entryNode || ''
      form.nodes = (chain.nodes || []).map((n) => ({
        id: n.id,
        type: n.type || '',
        label: n.label || '',
        configText: formatConfig(n.config),
      }))
      form.edges = (chain.edges || []).map((e) => ({
        from: e.from || '',
        to: e.to || '',
        label: e.label || '',
        condition: e.condition || '',
      }))
    })
    .catch(toastErrorHandler(t('rulechain.notification.loadFailed')))
    .finally(() => { loading.value = false })
}

function formatConfig (config) {
  if (!config) return '{}'
  if (typeof config === 'string') return config
  try {
    return JSON.stringify(config, null, 2)
  } catch (e) {
    return '{}'
  }
}

function hasConfigSchema (node) {
  const schema = nodeTypesByType.value[node.type]?.configSchema
  return schema && Object.keys(schema).length
}

function formatConfigSchema (node) {
  const schema = nodeTypesByType.value[node.type]?.configSchema
  return JSON.stringify(schema, null, 2)
}

function addNode () {
  nodeCounter++
  form.nodes.push({
    id: `n${Date.now().toString(36)}${nodeCounter}`,
    type: '',
    label: '',
    configText: '{}',
  })
}

function removeNode (ix) {
  const removed = form.nodes[ix]?.id
  form.nodes.splice(ix, 1)
  if (removed) {
    form.edges = form.edges.filter((e) => e.from !== removed && e.to !== removed)
    if (form.entryNode === removed) form.entryNode = ''
  }
}

function addEdge () {
  form.edges.push({ from: '', to: '', label: '', condition: '' })
}

function parseConfig (text) {
  const str = (text || '').trim()
  if (!str) return {}
  return JSON.parse(str)
}

function buildPayload () {
  const nodes = form.nodes.map((n) => {
    let config = {}
    try {
      config = parseConfig(n.configText)
    } catch (e) {
      config = { __invalid: n.configText }
    }
    return { id: n.id, type: n.type, label: n.label, config }
  })
  const edges = form.edges.map((e) => ({ from: e.from, to: e.to, label: e.label, condition: e.condition }))
  return {
    name: form.name,
    description: form.description,
    entryNode: form.entryNode,
    nodes,
    edges,
  }
}

function validateConfig () {
  for (const n of form.nodes) {
    if (!n.configText) continue
    try {
      parseConfig(n.configText)
    } catch (e) {
      toastErrorHandler(t('rulechain.edit.nodes.config.invalid') + ` (${n.label || n.id})`)(e)
      return false
    }
  }
  return true
}

function handleSave ({ closeOnSuccess = false } = {}) {
  if (!form.name) return
  if (!validateConfig()) return

  processing.value = true
  if (closeOnSuccess) processingSaveAndClose.value = true
  else processingSave.value = true

  const payload = buildPayload()
  const onSuccess = () => {
    if (closeOnSuccess) toastSuccess(t('rulechain.notification.updated'))
    else toastSuccess(t(isEdit.value ? 'rulechain.notification.updated' : 'rulechain.notification.created'))
    if (closeOnSuccess) {
      router.push({ name: 'admin.rulechains' })
    }
  }

  const onError = toastErrorHandler(t(isEdit.value ? 'rulechain.notification.updateFailed' : 'rulechain.notification.createFailed'))

  const action = isEdit.value
    ? $ComposeAPI.ruleChainUpdate({ chainID: props.chainID, ...payload })
    : $ComposeAPI.ruleChainCreate(payload)

  action.then(onSuccess).catch(onError).finally(() => {
    processing.value = false
    processingSave.value = false
    processingSaveAndClose.value = false
  })
}

function handleDelete () {
  processing.value = true
  processingDelete.value = true
  $ComposeAPI.ruleChainDelete({ chainID: props.chainID })
    .then(() => {
      toastSuccess(t('rulechain.notification.deleted'))
      router.push({ name: 'admin.rulechains' })
    })
    .catch(toastErrorHandler(t('rulechain.notification.deleteFailed')))
    .finally(() => {
      processing.value = false
      processingDelete.value = false
    })
}

function openTestModal () {
  if (!isEdit.value) {
    handleSave()
    if (!form.name || !validateConfig()) return
  }
  testInput.value = '{\n\n}'
  testResult.value = ''
  testModal.value?.show()
}

async function runTest () {
  let chainID = props.chainID
  if (!isEdit.value) {
    try {
      const { chain } = await $ComposeAPI.ruleChainCreate(buildPayload())
      chainID = chain.id
      toastSuccess(t('rulechain.notification.created'))
    } catch (e) {
      toastErrorHandler(t('rulechain.notification.createFailed'))(e)
      return
    }
  }

  let input = {}
  try {
    input = JSON.parse(testInput.value || '{}')
  } catch (e) {
    toastErrorHandler(t('rulechain.test.input.invalid', 'Input must be valid JSON'))(e)
    return
  }

  testRunning.value = true
  testResult.value = ''
  $ComposeAPI.ruleChainTest({ chainID, input })
    .then((result) => {
      testResult.value = JSON.stringify(result, null, 2)
    })
    .catch((e) => {
      testResult.value = JSON.stringify({ error: e?.response?.data?.error?.message || String(e.message || e) }, null, 2)
    })
    .finally(() => { testRunning.value = false })
}
</script>
