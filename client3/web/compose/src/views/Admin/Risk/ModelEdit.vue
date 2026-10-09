<template>
  <div
    v-if="namespace"
    class="container-fluid d-flex flex-column py-3"
  >
    <Teleport to="#topbar-title">
      {{ isEdit ? 'Редактирование модели риска' : 'Новая модель риска' }}
    </Teleport>

    <div
      v-if="loading"
      class="d-flex justify-content-center py-5"
    >
      <span class="spinner-border" />
    </div>

    <div
      v-else
      class="row g-3"
    >
      <div class="col-lg-7 d-flex flex-column gap-3">
        <!-- basics -->
        <div class="card shadow-sm">
          <div class="card-header"><strong>Основное</strong></div>
          <div class="card-body row g-3">
            <div class="col-md-6">
              <label class="form-label">Название</label>
              <input
                v-model="model.name"
                class="form-control"
              >
            </div>
            <div class="col-md-3">
              <label class="form-label">Стратегия</label>
              <select
                v-model="model.strategy"
                class="form-select"
              >
                <option value="weighted">
                  weighted
                </option>
                <option value="matrix">
                  matrix
                </option>
                <option value="bayes">
                  bayes
                </option>
                <option value="blend">
                  blend
                </option>
              </select>
            </div>
            <div class="col-md-3">
              <label class="form-label">Статус</label>
              <select
                v-model="model.status"
                class="form-select"
              >
                <option value="draft">
                  draft
                </option>
                <option value="published">
                  published
                </option>
                <option value="archived">
                  archived
                </option>
              </select>
            </div>
            <div class="col-12">
              <label class="form-label">Описание</label>
              <textarea
                v-model="model.description"
                class="form-control"
                rows="2"
              />
            </div>
          </div>
        </div>

        <!-- AI factor suggestion -->
        <div class="card shadow-sm">
          <div class="card-header"><strong>Подсказать факторы (черновик)</strong></div>
          <div class="card-body">
            <div class="input-group mb-2">
              <input
                v-model="suggestText"
                class="form-control"
                placeholder="Например: риск поставщика — задержки поставок, срыв договора"
              >
              <button
                class="btn btn-outline-primary"
                :disabled="suggesting"
                @click="suggest"
              >
                Предложить
              </button>
            </div>
            <p class="text-secondary small mb-2">
              Эвристический черновик (без вызова LLM) — просмотрите и создайте нужные в библиотеке.
            </p>
            <ul
              v-if="suggested.length"
              class="list-group"
            >
              <li
                v-for="(s, i) in suggested"
                :key="i"
                class="list-group-item d-flex justify-content-between align-items-center"
              >
                <span><code>{{ s.handle }}</code> — {{ s.label }} ({{ s.scale }})</span>
                <button
                  class="btn btn-sm btn-outline-success"
                  @click="createSuggested(s)"
                >
                  Создать в библиотеке
                </button>
              </li>
            </ul>
          </div>
        </div>

        <!-- nodes -->
        <div class="card shadow-sm">
          <div class="card-header d-flex justify-content-between align-items-center">
            <strong>Узлы</strong>
            <button
              class="btn btn-sm btn-outline-primary"
              @click="addNode"
            >
              + узел
            </button>
          </div>
          <div class="card-body p-0">
            <table class="table table-sm mb-0 align-middle">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Фактор</th>
                  <th>Роль</th>
                  <th v-if="model.strategy === 'bayes'">
                    CPT
                  </th>
                  <th />
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(n, i) in model.nodes"
                  :key="i"
                >
                  <td style="width:15%"><input
                    v-model="n.id"
                    class="form-control form-control-sm"
                  ></td>
                  <td>
                    <select
                      v-model="n.factorID"
                      class="form-select form-select-sm"
                    >
                      <option value="0">
                        — (нет, для output у weighted/matrix)
                      </option>
                      <option
                        v-for="f in factors"
                        :key="f.factorID"
                        :value="f.factorID"
                      >
                        {{ f.label }} ({{ f.handle }})
                      </option>
                    </select>
                  </td>
                  <td style="width:20%">
                    <select
                      v-model="n.role"
                      class="form-select form-select-sm"
                    >
                      <option value="evidence">
                        evidence
                      </option>
                      <option value="intermediate">
                        intermediate
                      </option>
                      <option value="output">
                        output
                      </option>
                    </select>
                  </td>
                  <td v-if="model.strategy === 'bayes'">
                    <button
                      v-if="n.role !== 'evidence'"
                      class="btn btn-sm btn-outline-secondary"
                      @click="openCPT(n)"
                    >
                      {{ n.cpt ? 'Изменить CPT' : 'Задать CPT' }}
                    </button>
                  </td>
                  <td class="text-end">
                    <button
                      class="btn btn-sm btn-outline-danger"
                      @click="model.nodes.splice(i, 1)"
                    >
                      &times;
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- edges (weighted / bayes) -->
        <div
          v-if="model.strategy === 'weighted' || model.strategy === 'bayes'"
          class="card shadow-sm"
        >
          <div class="card-header d-flex justify-content-between align-items-center">
            <strong>Связи</strong>
            <button
              class="btn btn-sm btn-outline-primary"
              @click="addEdge"
            >
              + связь
            </button>
          </div>
          <div class="card-body p-0">
            <table class="table table-sm mb-0 align-middle">
              <thead>
                <tr>
                  <th>Откуда</th>
                  <th>Куда</th>
                  <th v-if="model.strategy === 'weighted'">
                    Вес
                  </th>
                  <th />
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(e, i) in model.edges"
                  :key="i"
                >
                  <td>
                    <select
                      v-model="e.from"
                      class="form-select form-select-sm"
                    >
                      <option
                        v-for="n in model.nodes"
                        :key="n.id"
                        :value="n.id"
                      >
                        {{ n.id }}
                      </option>
                    </select>
                  </td>
                  <td>
                    <select
                      v-model="e.to"
                      class="form-select form-select-sm"
                    >
                      <option
                        v-for="n in model.nodes"
                        :key="n.id"
                        :value="n.id"
                      >
                        {{ n.id }}
                      </option>
                    </select>
                  </td>
                  <td v-if="model.strategy === 'weighted'">
                    <input
                      v-model.number="e.weight"
                      type="number"
                      step="0.01"
                      class="form-control form-control-sm"
                    >
                  </td>
                  <td class="text-end">
                    <button
                      class="btn btn-sm btn-outline-danger"
                      @click="model.edges.splice(i, 1)"
                    >
                      &times;
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- strategy config -->
        <div class="card shadow-sm">
          <div class="card-header"><strong>Конфигурация стратегии</strong></div>
          <div class="card-body">
            <template v-if="model.strategy === 'weighted'">
              <div class="form-check mb-2">
                <input
                  id="wNorm"
                  v-model="weightedConfig.normalize"
                  class="form-check-input"
                  type="checkbox"
                >
                <label
                  class="form-check-label"
                  for="wNorm"
                >Нормализовать к шкале</label>
              </div>
              <label class="form-label">Максимум шкалы</label>
              <input
                v-model.number="weightedConfig.scaleMax"
                type="number"
                class="form-control"
                style="max-width:160px"
              >
            </template>

            <template v-else-if="model.strategy === 'matrix'">
              <div class="row g-3">
                <div class="col-md-4">
                  <label class="form-label">Узел «правдоподобие»</label>
                  <select
                    v-model="matrixConfig.likelihoodNode"
                    class="form-select"
                  >
                    <option
                      v-for="n in model.nodes"
                      :key="n.id"
                      :value="n.id"
                    >
                      {{ n.id }}
                    </option>
                  </select>
                </div>
                <div class="col-md-4">
                  <label class="form-label">Узел «влияние»</label>
                  <select
                    v-model="matrixConfig.impactNode"
                    class="form-select"
                  >
                    <option
                      v-for="n in model.nodes"
                      :key="n.id"
                      :value="n.id"
                    >
                      {{ n.id }}
                    </option>
                  </select>
                </div>
                <div class="col-md-2">
                  <label class="form-label">Размер шкалы</label>
                  <input
                    v-model.number="matrixConfig.scaleSize"
                    type="number"
                    class="form-control"
                  >
                </div>
                <div class="col-md-2">
                  <label class="form-label">Формула</label>
                  <select
                    v-model="matrixConfig.formula"
                    class="form-select"
                  >
                    <option value="product">
                      product
                    </option>
                    <option value="sum">
                      sum
                    </option>
                  </select>
                </div>
              </div>
            </template>

            <template v-else-if="model.strategy === 'bayes'">
              <div class="mb-3">
                <label class="form-label">Выходной узел</label>
                <select
                  v-model="bayesConfig.outputNode"
                  class="form-select"
                  style="max-width:300px"
                >
                  <option
                    v-for="n in model.nodes"
                    :key="n.id"
                    :value="n.id"
                  >
                    {{ n.id }}
                  </option>
                </select>
              </div>
              <label class="form-label">Score по состоянию выходного узла</label>
              <table class="table table-sm" style="max-width:400px">
                <tbody>
                  <tr
                    v-for="st in outputStates"
                    :key="st"
                  >
                    <td>{{ st }}</td>
                    <td>
                      <input
                        v-model.number="bayesConfig.scoreByState[st]"
                        type="number"
                        class="form-control form-control-sm"
                      >
                    </td>
                  </tr>
                </tbody>
              </table>
            </template>

            <template v-else>
              <p class="text-secondary small">
                Для «blend» узлы, связи и конфигурация компонентов редактируются как JSON —
                каждый компонент переиспользует Evaluate/Explain своей собственной стратегии
                над этим же графом (см. архитектурную записку).
              </p>
              <label class="form-label">nodes (JSON)</label>
              <textarea
                v-model="blendNodesJSON"
                class="form-control mb-2 font-monospace"
                rows="6"
              />
              <label class="form-label">edges (JSON)</label>
              <textarea
                v-model="blendEdgesJSON"
                class="form-control mb-2 font-monospace"
                rows="4"
              />
              <label class="form-label">config (JSON — { "components": [...] })</label>
              <textarea
                v-model="blendConfigJSON"
                class="form-control font-monospace"
                rows="6"
              />
            </template>
          </div>
        </div>

        <div class="d-flex justify-content-end gap-2">
          <router-link
            :to="{ name: 'admin.risk.models' }"
            class="btn btn-outline-secondary"
          >
            Отмена
          </router-link>
          <button
            class="btn btn-primary"
            @click="save"
          >
            Сохранить
          </button>
        </div>
      </div>

      <!-- simulator -->
      <div class="col-lg-5">
        <div class="card shadow-sm sticky-top" style="top:1rem">
          <div class="card-header"><strong>Симулятор</strong></div>
          <div class="card-body">
            <div
              v-for="n in evidenceNodes"
              :key="n.id"
              class="mb-2"
            >
              <label class="form-label mb-1">{{ factorLabel(n.factorID) }}</label>
              <select
                v-if="factorByID(n.factorID)?.scale === 'states'"
                v-model="evidence[factorByID(n.factorID).handle]"
                class="form-select form-select-sm"
              >
                <option
                  v-for="s in factorByID(n.factorID).states"
                  :key="s.code"
                  :value="s.code"
                >
                  {{ s.label }}
                </option>
              </select>
              <input
                v-else
                v-model.number="evidence[factorByID(n.factorID)?.handle]"
                type="number"
                class="form-control form-control-sm"
              >
            </div>

            <div class="mb-3">
              <label class="form-label mb-1">Эффективность контролей (0–1, опционально)</label>
              <input
                v-model.number="controlEffectiveness"
                type="number"
                min="0"
                max="1"
                step="0.05"
                class="form-control form-control-sm"
              >
            </div>

            <button
              class="btn btn-primary w-100 mb-3"
              :disabled="running"
              @click="runSimulation"
            >
              Рассчитать
            </button>

            <div v-if="result">
              <div class="d-flex gap-2 mb-2">
                <span class="badge bg-secondary">inherent {{ round1(result.inherentScore) }}</span>
                <span class="badge bg-secondary">residual {{ round1(result.residualScore) }}</span>
                <span
                  class="badge"
                  :class="levelBadgeClass(result.level)"
                >{{ result.level }}</span>
              </div>
              <p class="small">
                {{ result.explanation }}
              </p>

              <div v-if="result.attribution">
                <div
                  v-for="c in result.attribution.contributions"
                  :key="c.nodeID"
                  class="mb-1"
                >
                  <div class="d-flex justify-content-between small">
                    <span>{{ factorLabelByHandle(c.factorHandle) }}</span>
                    <span>{{ c.value >= 0 ? '+' : '' }}{{ round1(c.value) }}</span>
                  </div>
                  <div class="progress" style="height:6px">
                    <div
                      class="progress-bar"
                      :class="c.value >= 0 ? 'bg-primary' : 'bg-success'"
                      :style="{ width: barWidth(c.value) + '%' }"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- CPT modal -->
    <div
      v-if="cptModal"
      class="modal d-block"
      style="background: rgba(0,0,0,.5)"
    >
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <strong>CPT для узла «{{ cptModal.node.id }}»</strong>
            <button
              class="btn-close"
              @click="cptModal = null"
            />
          </div>
          <div class="modal-body">
            <p
              v-if="cptModal.parents.length === 0"
              class="text-secondary"
            >
              У узла нет родителей — добавьте связи в него в разделе «Связи» перед заданием CPT.
            </p>
            <div
              v-else
              class="table-responsive"
            >
              <table class="table table-sm table-bordered">
                <thead>
                  <tr>
                    <th
                      v-for="p in cptModal.parents"
                      :key="p"
                    >
                      {{ p }}
                    </th>
                    <th
                      v-for="s in cptModal.nodeStates"
                      :key="s"
                    >
                      P({{ s }})
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="(row, ri) in cptModal.rows"
                    :key="ri"
                  >
                    <td
                      v-for="(val, pi) in cptModal.rowLabels[ri]"
                      :key="pi"
                    >
                      {{ val }}
                    </td>
                    <td
                      v-for="(_, ci) in cptModal.nodeStates"
                      :key="ci"
                    >
                      <input
                        v-model.number="cptModal.rows[ri][ci]"
                        type="number"
                        step="0.01"
                        min="0"
                        max="1"
                        class="form-control form-control-sm"
                      >
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            <p class="text-secondary small">
              Каждая строка должна суммироваться к 1.
            </p>
          </div>
          <div class="modal-footer">
            <button
              class="btn btn-outline-secondary"
              @click="cptModal = null"
            >
              Отмена
            </button>
            <button
              class="btn btn-primary"
              @click="saveCPT"
            >
              Сохранить CPT
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { composables } from 'corteza-lib/vue/dist'

const { useToast } = composables
const { toastSuccess, toastErrorHandler } = useToast()
const route = useRoute()
const router = useRouter()
const $ComposeAPI = window.__composeAPI

const props = defineProps({
  namespace: {
    type: Object,
    required: false,
    default: undefined,
  },
})

const isEdit = computed(() => !!route.params.modelID)
const loading = ref(true)
const factors = ref([])

const model = reactive({
  modelID: '',
  namespaceID: '',
  name: '',
  description: '',
  strategy: 'weighted',
  status: 'draft',
  version: 1,
  nodes: [],
  edges: [],
})

const weightedConfig = reactive({ normalize: true, scaleMax: 100 })
const matrixConfig = reactive({ likelihoodNode: '', impactNode: '', scaleSize: 5, formula: 'product' })
const bayesConfig = reactive({ outputNode: '', scoreByState: {} })
const blendNodesJSON = ref('[]')
const blendEdgesJSON = ref('[]')
const blendConfigJSON = ref('{\n  "components": []\n}')

function factorByID (id) {
  return factors.value.find((f) => f.factorID === String(id))
}
function factorLabel (id) {
  const f = factorByID(id)
  return f ? `${f.label} (${f.handle})` : id
}
function factorLabelByHandle (handle) {
  const f = factors.value.find((x) => x.handle === handle)
  return f?.label || handle
}

const evidenceNodes = computed(() => model.nodes.filter((n) => n.role === 'evidence' && n.factorID !== '0'))
const outputStates = computed(() => {
  const out = model.nodes.find((n) => n.id === bayesConfig.outputNode)
  const fd = out ? factorByID(out.factorID) : null
  return (fd?.states || []).map((s) => s.code)
})

function loadFactors () {
  return $ComposeAPI.riskFactorList({ namespaceID: props.namespace?.namespaceID })
    .then(({ factors: set }) => { factors.value = set || [] })
}

function loadModel () {
  if (!isEdit.value) {
    loading.value = false
    return Promise.resolve()
  }
  return $ComposeAPI.riskModelRead({ modelID: route.params.modelID })
    .then(({ model: m }) => {
      Object.assign(model, m, { nodes: m.nodes || [], edges: m.edges || [] })
      const cfg = m.config || {}
      if (m.strategy === 'weighted') Object.assign(weightedConfig, { normalize: cfg.normalize !== false, scaleMax: cfg.scaleMax || 100 })
      if (m.strategy === 'matrix') Object.assign(matrixConfig, cfg)
      if (m.strategy === 'bayes') Object.assign(bayesConfig, { outputNode: cfg.outputNode || '', scoreByState: cfg.scoreByState || {} })
      if (m.strategy === 'blend') {
        blendNodesJSON.value = JSON.stringify(m.nodes || [], null, 2)
        blendEdgesJSON.value = JSON.stringify(m.edges || [], null, 2)
        blendConfigJSON.value = JSON.stringify(cfg, null, 2)
      }
    })
    .catch(toastErrorHandler('Не удалось загрузить модель'))
    .finally(() => { loading.value = false })
}

onMounted(() => {
  loading.value = true
  Promise.all([loadFactors(), loadModel()]).finally(() => { loading.value = false })
})

function addNode () {
  model.nodes.push({ id: `n${model.nodes.length + 1}`, factorID: '0', role: 'evidence', x: 0, y: 0 })
}
function addEdge () {
  model.edges.push({ from: model.nodes[0]?.id || '', to: model.nodes[1]?.id || '', weight: 1 })
}

// --- CPT modal ---
const cptModal = ref(null)
function openCPT (node) {
  const parentEdges = model.edges.filter((e) => e.to === node.id)
  const parents = parentEdges.map((e) => e.from)
  const parentStatesArr = parents.map((pid) => {
    const pn = model.nodes.find((n) => n.id === pid)
    return (factorByID(pn?.factorID)?.states || []).map((s) => s.code)
  })
  const nodeStates = (factorByID(node.factorID)?.states || []).map((s) => s.code)

  // cartesian product, existing cpt.rows reused if shape matches
  const combos = cartesian(parentStatesArr)
  const existing = node.cpt
  const rows = combos.map((combo, i) => {
    if (existing && existing.rows[i] && existing.rows[i].length === nodeStates.length) return [...existing.rows[i]]
    return nodeStates.map(() => +(1 / nodeStates.length).toFixed(2))
  })

  cptModal.value = { node, parents, parentStatesList: parentStatesArr, nodeStates, rows, rowLabels: combos }
}
function cartesian (arrays) {
  return arrays.reduce((acc, arr) => acc.flatMap((a) => arr.map((v) => [...a, v])), [[]])
}
function saveCPT () {
  cptModal.value.node.cpt = {
    parentStates: cptModal.value.parentStatesList,
    nodeStates: cptModal.value.nodeStates,
    rows: cptModal.value.rows,
  }
  cptModal.value = null
}

// --- suggest factors ---
const suggestText = ref('')
const suggesting = ref(false)
const suggested = ref([])
function suggest () {
  if (!suggestText.value) return
  suggesting.value = true
  $ComposeAPI.riskSuggestFactors({ description: suggestText.value })
    .then(({ factors: set }) => { suggested.value = set || [] })
    .catch(toastErrorHandler('Не удалось получить подсказку'))
    .finally(() => { suggesting.value = false })
}
function createSuggested (s) {
  $ComposeAPI.riskFactorCreate({ ...s, namespaceID: props.namespace?.namespaceID })
    .then(() => {
      toastSuccess(`Фактор «${s.handle}» создан`)
      return loadFactors()
    })
    .catch(toastErrorHandler('Не удалось создать фактор'))
}

// --- save model ---
function buildConfig () {
  if (model.strategy === 'weighted') return weightedConfig
  if (model.strategy === 'matrix') return matrixConfig
  if (model.strategy === 'bayes') return bayesConfig
  try {
    return JSON.parse(blendConfigJSON.value)
  } catch {
    return {}
  }
}

function save () {
  const payload = {
    namespaceID: props.namespace?.namespaceID,
    name: model.name,
    description: model.description,
    strategy: model.strategy,
    status: model.status,
    version: model.version,
    config: buildConfig(),
    nodes: model.strategy === 'blend' ? safeParse(blendNodesJSON.value, []) : model.nodes,
    edges: model.strategy === 'blend' ? safeParse(blendEdgesJSON.value, []) : model.edges,
  }
  const isNew = !isEdit.value
  const call = isNew ? $ComposeAPI.riskModelCreate(payload) : $ComposeAPI.riskModelUpdate({ ...payload, modelID: route.params.modelID })
  call
    .then(({ model: saved }) => {
      toastSuccess(isNew ? 'Модель создана' : 'Модель сохранена')
      router.push({ name: 'admin.risk.models.edit', params: { modelID: saved.modelID } })
    })
    .catch(toastErrorHandler('Не удалось сохранить модель'))
}
function safeParse (s, fallback) {
  try {
    return JSON.parse(s)
  } catch {
    return fallback
  }
}

// --- simulator ---
const evidence = reactive({})
const controlEffectiveness = ref(null)
const running = ref(false)
const result = ref(null)

function runSimulation () {
  running.value = true
  const ev = { ...evidence }
  if (controlEffectiveness.value !== null && controlEffectiveness.value !== '') {
    ev.controlEffectiveness = controlEffectiveness.value
  }
  $ComposeAPI.riskModelExplain({
    modelID: route.params.modelID,
    evidence: ev,
    controlFactorHandle: 'controlEffectiveness',
  })
    .then(({ assessment }) => { result.value = assessment })
    .catch(toastErrorHandler('Расчёт не удался — сохраните модель и проверьте узлы/факторы'))
    .finally(() => { running.value = false })
}

function round1 (v) {
  return Math.round((v || 0) * 10) / 10
}
function levelBadgeClass (level) {
  return { low: 'bg-success', medium: 'bg-warning text-dark', high: 'bg-orange text-white', critical: 'bg-danger' }[level] || 'bg-secondary'
}
function barWidth (v) {
  return Math.min(100, Math.abs(v) * 2)
}
</script>
