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
                <button
                  class="btn btn-outline-secondary"
                  :disabled="!isEdit"
                  :title="!isEdit ? $t('rulechain.runs.selectHint') : ''"
                  @click="openRunsModal"
                >
                  <font-awesome-icon :icon="['fas', 'list']" class="me-1" />
                  {{ $t('rulechain.runs.title') }}
                </button>
                <div class="d-flex ms-auto gap-2">
                  <button
                    v-if="!isGraphMode"
                    class="btn btn-sm btn-outline-secondary"
                    @click="switchToGraph"
                  >
                    <font-awesome-icon :icon="['fas', 'project-diagram']" class="me-1" />
                    Visual
                  </button>
                  <button
                    v-if="isGraphMode"
                    class="btn btn-sm btn-outline-secondary"
                    @click="switchToForm"
                  >
                    <font-awesome-icon :icon="['fas', 'list']" class="me-1" />
                    Form
                  </button>
                </div>
              </div>

              <div
                v-if="!isGraphMode"
                class="overflow-auto p-3"
                style="flex: 1 1 0%; min-height: 0;"
              >
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
                      <div class="col-12">
                        <RuleChainNodeConfigEditor
                          v-model="node.configText"
                          :node-type="node.type"
                          :fields="nodeTypesByType[node.type]?.configFields"
                          :node-schema="nodeTypesByType[node.type]"
                          :description="nodeTypesByType[node.type]?.description"
                          :json-rows="8"
                        />
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

              <div
                v-if="isGraphMode"
                class="d-flex flex-grow-1"
                style="min-height: 0"
              >
                <div class="d-flex flex-column flex-grow-1 p-2" style="min-height: 0">
                  <div class="d-flex justify-content-between align-items-center mb-2 px-1">
                    <div>
                      <span class="small text-muted">
                        {{ $t('rulechain.edit.generalSettings') }}:
                      </span>
                      <input
                        v-model="form.name"
                        class="form-control form-control-sm d-inline-block ms-2"
                        style="width: 200px"
                        type="text"
                        :placeholder="$t('rulechain.edit.name.placeholder')"
                      >
                    </div>
                    <div class="d-flex gap-2">
                      <button
                        class="btn btn-sm btn-outline-primary"
                        @click="addGraphNode"
                      >
                        + Node
                      </button>
                    </div>
                  </div>
                  <RuleChainGraph
                    ref="graphRef"
                    v-model:nodes="form.nodes"
                    v-model:edges="form.edges"
                    v-model:entry-node="form.entryNode"
                    :available-node-types="nodeTypes"
                    class="flex-grow-1"
                    style="min-height: 0"
                  />
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

    <div
      ref="runsModalEl"
      class="modal fade"
      tabindex="-1"
      aria-hidden="true"
    >
      <div class="modal-dialog modal-xl modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">
              {{ $t('rulechain.runs.title') }}
            </h5>
            <button
              type="button"
              class="btn btn-sm btn-outline-secondary ms-auto me-2"
              :disabled="runsLoading"
              @click="loadRuns"
            >
              <span
                v-if="runsLoading"
                class="spinner-border spinner-border-sm"
              />
              <template v-else>
                {{ $t('rulechain.runs.refresh') }}
              </template>
            </button>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
              aria-label="Close"
            />
          </div>
          <div
            class="modal-body d-flex gap-3"
            style="min-height: 60vh;"
          >
            <div
              class="border-end pe-3 overflow-auto"
              style="flex: 1 1 40%; min-width: 0;"
            >
              <div
                v-if="runsLoading && !runs.length"
                class="d-flex align-items-center justify-content-center py-4"
              >
                <span class="spinner-border spinner-border-sm" />
              </div>
              <div
                v-else-if="!runs.length"
                class="text-muted small py-3"
              >
                {{ $t('rulechain.runs.empty') }}
              </div>
              <table
                v-else
                class="table table-sm table-hover mb-0"
              >
                <thead>
                  <tr>
                    <th style="width: 2rem" />
                    <th>{{ $t('rulechain.runs.columns.trigger') }}</th>
                    <th>{{ $t('rulechain.runs.columns.startedAt') }}</th>
                    <th class="text-end">
                      {{ $t('rulechain.runs.columns.duration') }}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="run in runs"
                    :key="run.runID"
                    role="button"
                    :class="{ 'table-active': selectedRun && selectedRun.runID === run.runID }"
                    @click="selectRun(run)"
                  >
                    <td>
                      <font-awesome-icon
                        :icon="['fas', run.success ? 'check-circle' : 'exclamation-circle']"
                        :class="run.success ? 'text-success' : 'text-danger'"
                      />
                    </td>
                    <td>
                      <span class="badge bg-secondary">{{ run.triggerType || '—' }}</span>
                    </td>
                    <td class="small text-nowrap">
                      {{ formatRunDate(run.startedAt) }}
                    </td>
                    <td class="small text-nowrap text-end">
                      {{ run.durationMs != null ? `${run.durationMs} ms` : '—' }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div
              class="overflow-auto"
              style="flex: 1 1 60%; min-width: 0;"
            >
              <div
                v-if="!selectedRun"
                class="text-muted small py-3"
              >
                {{ $t('rulechain.runs.selectHint') }}
              </div>
              <div
                v-else-if="runDetailLoading"
                class="d-flex align-items-center justify-content-center py-4"
              >
                <span class="spinner-border spinner-border-sm" />
              </div>
              <div v-else-if="selectedRunDetail">
                <div class="mb-2 d-flex align-items-center gap-2">
                  <span
                    class="badge"
                    :class="selectedRunDetail.success ? 'bg-success' : 'bg-danger'"
                  >
                    {{ selectedRunDetail.success ? $t('rulechain.test.modal.success') : $t('rulechain.test.modal.failed') }}
                  </span>
                  <span class="text-muted small">
                    {{ formatRunDate(selectedRunDetail.startedAt) }}
                  </span>
                </div>

                <div
                  v-if="selectedRunDetail.error"
                  class="alert alert-danger py-2 small mb-3"
                >
                  {{ selectedRunDetail.error }}
                </div>

                <h6 class="small text-muted text-uppercase mb-2">
                  {{ $t('rulechain.runs.trace') }}
                </h6>

                <div
                  v-if="!(selectedRunDetail.nodes && selectedRunDetail.nodes.length)"
                  class="text-muted small mb-3"
                >
                  {{ $t('rulechain.runs.noTrace') }}
                </div>

                <div
                  v-for="(node, ix) in selectedRunDetail.nodes"
                  :key="ix"
                  class="card mb-2"
                >
                  <div class="card-body py-2">
                    <div class="d-flex align-items-center gap-2 mb-1">
                      <font-awesome-icon
                        :icon="['fas', node.error ? 'exclamation-circle' : 'check-circle']"
                        :class="node.error ? 'text-danger' : 'text-success'"
                      />
                      <strong class="small">{{ node.nodeID }}</strong>
                      <span class="badge bg-light text-dark border">{{ node.type }}</span>
                    </div>
                    <div
                      v-if="node.error"
                      class="text-danger small mb-1"
                    >
                      {{ node.error }}
                    </div>
                    <pre
                      v-if="node.output"
                      class="bg-light border rounded p-2 mb-0 small"
                      style="max-height: 20vh; overflow: auto; white-space: pre-wrap;"
                    >{{ JSON.stringify(node.output, null, 2) }}</pre>
                  </div>
                </div>

                <h6 class="small text-muted text-uppercase mt-3 mb-2">
                  {{ $t('rulechain.runs.output') }}
                </h6>
                <pre
                  class="bg-light border rounded p-2 mb-0 small"
                  style="max-height: 20vh; overflow: auto; white-space: pre-wrap;"
                >{{ JSON.stringify(selectedRunDetail.output, null, 2) }}</pre>
              </div>
            </div>
          </div>
          <div class="modal-footer">
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
import { ref, computed, reactive, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Modal } from 'bootstrap'
import { NoID } from 'corteza-lib/js/dist'
import { composables } from 'corteza-lib/vue/dist'
import EditorToolbar from 'corteza-webapp-compose/src/components/Admin/EditorToolbar'
import RuleChainGraph from 'corteza-webapp-compose/src/components/Admin/RuleChains/RuleChainGraph'
import RuleChainNodeConfigEditor from 'corteza-webapp-compose/src/components/Admin/RuleChains/RuleChainNodeConfigEditor'

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

const isGraphMode = ref(false)
const graphRef = ref(null)

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

const runsModalEl = ref(null)
const runsModal = ref(undefined)
const runs = ref([])
const runsLoading = ref(false)
const selectedRun = ref(null)
const selectedRunDetail = ref(null)
const runDetailLoading = ref(false)

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

  runsModal.value = new Modal(runsModalEl.value)
  runsModalEl.value.addEventListener('hidden.bs.modal', () => {
    runs.value = []
    selectedRun.value = null
    selectedRunDetail.value = null
  })
})

onBeforeUnmount(() => {
  testModal.value?.dispose()
  runsModal.value?.dispose()
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
    id: isEdit.value ? props.chainID : undefined,
    namespaceID: props.namespace?.namespaceID,
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

function switchToGraph () {
  isGraphMode.value = true
}

function switchToForm () {
  isGraphMode.value = false
}

function addGraphNode () {
  graphRef.value?.addNode()
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

function openRunsModal () {
  if (!isEdit.value) return
  runsModal.value?.show()
  loadRuns()
}

function loadRuns () {
  runsLoading.value = true
  $ComposeAPI.ruleChainRuns({ chainID: props.chainID, limit: 20 })
    .then(({ runs: list }) => {
      runs.value = list || []
      // keep the current selection in sync with the freshly loaded list,
      // or default to the most recent run on first load
      if (runs.value.length) {
        const stillThere = selectedRun.value && runs.value.find((r) => r.runID === selectedRun.value.runID)
        selectRun(stillThere || runs.value[0])
      } else {
        selectedRun.value = null
        selectedRunDetail.value = null
      }
    })
    .catch(toastErrorHandler(t('rulechain.runs.loadFailed')))
    .finally(() => { runsLoading.value = false })
}

function selectRun (run) {
  selectedRun.value = run
  selectedRunDetail.value = null
  runDetailLoading.value = true
  $ComposeAPI.ruleChainRunGet({ chainID: props.chainID, runID: run.runID })
    .then(({ run: detail }) => {
      selectedRunDetail.value = {
        ...detail,
        nodes: detail.nodes || [],
        output: detail.output || {},
      }
    })
    .catch(toastErrorHandler(t('rulechain.runs.detailFailed')))
    .finally(() => { runDetailLoading.value = false })
}

function formatRunDate (v) {
  if (!v) return '—'
  try {
    return new Date(v).toLocaleString()
  } catch (e) {
    return v
  }
}
</script>
