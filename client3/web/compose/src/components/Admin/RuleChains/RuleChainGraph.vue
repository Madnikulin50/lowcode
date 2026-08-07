<template>
  <div class="d-flex flex-grow-1" style="min-height: 0">
    <div
      class="flex-grow-1"
      style="min-height: 0; position: relative"
    >
      <VueFlow
        ref="vueFlowRef"
        v-model:nodes="vfNodes"
        v-model:edges="vfEdges"
        :node-types="nodeTypes"
        :default-edge-options="{ type: 'smoothstep', animated: false, style: { strokeWidth: 2 } }"
        class="rulechain-graph"
        fit-view-on-init
        :nodes-draggable="true"
        :nodes-connectable="true"
        :edges-updatable="true"
        @connect="onConnect"
        @node-click="onNodeClick"
        @edge-click="onEdgeClick"
        @nodes-change="onNodesChange"
        @edges-change="onEdgesChange"
        @pane-click="onPaneClick"
      >
        <template #node-rulechain="ruleNodeProps">
          <RuleChainGraphNode
            v-bind="ruleNodeProps"
            :entry-node="entryNode"
            @delete="removeNode(ruleNodeProps.id)"
          />
        </template>

        <Background />
      </VueFlow>
    </div>

    <div
      v-if="selectedNode || selectedEdgeId"
      class="border-start bg-white p-3 d-flex flex-column"
      style="width: 320px; min-height: 0; overflow-y: auto"
    >
      <template v-if="selectedEdgeId && !selectedNode">
        <div class="d-flex justify-content-between align-items-center mb-3">
          <h6 class="mb-0">{{ $t('rulechain.edit.edges.label') }}</h6>
          <button
            class="btn btn-sm btn-outline-secondary"
            @click="selectedEdgeId = null"
          >
            ×
          </button>
        </div>
        <div class="mb-2">
          <label class="form-label small fw-bold text-muted">
            {{ $t('rulechain.edit.edges.from.label') }}
          </label>
          <span class="d-block small font-monospace">{{ selectedEdge?.source }}</span>
        </div>
        <div class="mb-2">
          <label class="form-label small fw-bold text-muted">
            {{ $t('rulechain.edit.edges.to.label') }}
          </label>
          <span class="d-block small font-monospace">{{ selectedEdge?.target }}</span>
        </div>
        <div class="mb-3">
          <label class="form-label small fw-bold text-muted">
            Label
          </label>
          <input
            v-model="selectedEdgeLabel"
            class="form-control form-control-sm mb-2"
            type="text"
          >
          <label class="form-label small fw-bold text-muted mt-2">
            {{ $t('rulechain.edit.edges.condition.label') }}
          </label>
          <input
            v-model="selectedEdgeCondition"
            class="form-control form-control-sm font-monospace"
            type="text"
          >
        </div>
        <div class="mt-auto">
          <button
            class="btn btn-sm btn-outline-danger w-100"
            @click="removeEdge(selectedEdgeId)"
          >
            {{ $t('rulechain.edit.edges.delete') }}
          </button>
        </div>
      </template>

      <template v-if="selectedNode">
        <div class="d-flex justify-content-between align-items-center mb-3">
          <h6 class="mb-0 text-truncate">
            {{ selectedNode.data.label || selectedNode.id }}
          </h6>
          <button
            class="btn btn-sm btn-outline-secondary"
            @click="selectedNode = null; selectedEdgeId = null"
          >
            ×
          </button>
        </div>

      <div class="mb-3">
        <label class="form-label small fw-bold text-muted">
          {{ $t('rulechain.edit.nodes.type.label') }}
        </label>
        <select
          v-model="selectedNode.data.nodeType"
          class="form-select form-select-sm"
        >
          <option value="">
            --
          </option>
          <option
            v-for="nt in availableNodeTypes"
            :key="nt.type"
            :value="nt.type"
          >
            {{ nt.label || nt.type }}
          </option>
        </select>
      </div>

      <div class="mb-3">
        <label class="form-label small fw-bold text-muted">
          {{ $t('rulechain.edit.nodes.label') }}
        </label>
        <input
          v-model="selectedNode.data.label"
          class="form-control form-control-sm"
          type="text"
        >
      </div>

      <div
        v-if="selectedNodeSchema"
        class="mb-3"
      >
        <label class="form-label small fw-bold text-muted">
          Config
        </label>
        <p class="small text-muted mb-1">
          {{ selectedNodeSchema.description }}
        </p>
        <textarea
          v-model="selectedNode.data.configText"
          class="form-control form-control-sm font-monospace"
          rows="6"
          spellcheck="false"
        />
        <div
          v-if="selectedNodeSchema.configSchema"
          class="mt-1"
        >
          <pre class="bg-light rounded p-2 small mb-0" style="max-height: 120px; overflow: auto; font-size: 0.7rem;">{{ JSON.stringify(selectedNodeSchema.configSchema, null, 2) }}</pre>
        </div>
      </div>

      <div class="mt-auto">
        <div class="d-flex gap-2 mb-3">
          <button
            class="btn btn-sm btn-outline-danger flex-grow-1"
            @click="removeNode(selectedNode.id); selectedNode = null"
          >
            {{ $t('rulechain.edit.nodes.delete') }}
          </button>
        </div>

        <div
          v-if="selectedEdgeId"
          class="mb-3 p-2 bg-light rounded"
        >
          <label class="form-label small fw-bold text-muted d-flex justify-content-between">
            {{ $t('rulechain.edit.edges.label') }}
            <button
              class="btn btn-sm btn-outline-danger border-0 py-0 px-1"
              @click="removeEdge(selectedEdgeId)"
            >
              ×
            </button>
          </label>
          <input
            v-model="selectedEdgeLabel"
            class="form-control form-control-sm mb-2"
            type="text"
            placeholder="Label"
          >
          <input
            v-model="selectedEdgeCondition"
            class="form-control form-control-sm font-monospace"
            type="text"
            placeholder="Condition"
          >
        </div>
      </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { VueFlow } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import RuleChainGraphNode from './RuleChainGraphNode.vue'

const { t } = useI18n()

const props = defineProps({
  nodes: { type: Array, default: () => [] },
  edges: { type: Array, default: () => [] },
  entryNode: { type: String, default: '' },
  availableNodeTypes: { type: Array, default: () => [] },
})

const emit = defineEmits(['update:nodes', 'update:edges', 'update:entryNode'])

const vueFlowRef = ref(null)
const selectedNode = ref(null)
const selectedEdgeId = ref(null)

const nodeTypes = { 'rulechain': RuleChainGraphNode }

const vfNodes = ref([])
const vfEdges = ref([])

const selectedEdge = computed(() => vfEdges.value.find((e) => e.id === selectedEdgeId.value))

const selectedEdgeLabel = computed({
  get: () => selectedEdge.value?.label || '',
  set: (v) => {
    const idx = vfEdges.value.findIndex((e) => e.id === selectedEdgeId.value)
    if (idx >= 0) {
      vfEdges.value[idx] = { ...vfEdges.value[idx], label: v }
    }
  },
})

const selectedEdgeCondition = computed({
  get: () => selectedEdge.value?.data?.condition || '',
  set: (v) => {
    const idx = vfEdges.value.findIndex((e) => e.id === selectedEdgeId.value)
    if (idx >= 0) {
      vfEdges.value[idx] = { ...vfEdges.value[idx], data: { ...vfEdges.value[idx].data, condition: v } }
    }
  },
})

const selectedNodeSchema = computed(() => {
  if (!selectedNode.value?.data?.nodeType) return null
  return props.availableNodeTypes.find((nt) => nt.type === selectedNode.value.data.nodeType) || null
})

function rcNodesToVf (rcNodes) {
  return (rcNodes || []).map((n, i) => ({
    id: n.id,
    type: 'rulechain',
    position: { x: n.x || (100 + (i % 3) * 280), y: n.y || (50 + Math.floor(i / 3) * 120) },
    data: {
      label: n.label || '',
      nodeType: n.type || '',
      configText: n.configText || '{}',
      entry: false,
    },
    draggable: true,
    connectable: true,
  }))
}

function rcEdgesToVf (rcEdges) {
  return (rcEdges || []).map((e, i) => ({
    id: e.from + '-' + e.to + (i > 0 ? '-' + i : ''),
    source: e.from,
    target: e.to,
    type: 'smoothstep',
    label: e.label || '',
    style: { stroke: 'var(--bs-secondary, #6c757d)', strokeWidth: 2 },
    data: { condition: e.condition || '' },
  }))
}

function syncFromProps () {
  vfNodes.value = rcNodesToVf(props.nodes)
  vfEdges.value = rcEdgesToVf(props.edges)
  if (props.entryNode) {
    const en = vfNodes.value.find((n) => n.id === props.entryNode)
    if (en) en.data.entry = true
  }
}

function emitNodes () {
  const result = vfNodes.value.map((n) => ({
    id: n.id,
    type: n.data.nodeType || '',
    label: n.data.label || '',
    configText: n.data.configText || '{}',
    x: n.position.x,
    y: n.position.y,
  }))
  emit('update:nodes', result)
}

function emitEdges () {
  const result = vfEdges.value.map((e) => ({
    from: e.source,
    to: e.target,
    label: e.label || '',
    condition: (e.data?.condition) || '',
  }))
  emit('update:edges', result)
}

function emitEntryNode () {
  const en = vfNodes.value.find((n) => n.data.entry)
  emit('update:entryNode', en?.id || '')
}

onMounted(() => syncFromProps())

function onConnect ({ source, target }) {
  if (source === target) return
  const exists = vfEdges.value.some((e) => e.source === source && e.target === target)
  if (exists) return

  vfEdges.value.push({
    id: source + '-' + target + '-' + Date.now(),
    source,
    target,
    type: 'smoothstep',
    label: '',
    style: { stroke: 'var(--bs-secondary, #6c757d)', strokeWidth: 2 },
    data: { condition: '' },
  })
  emitEdges()
}

function onNodesChange () {
  emitNodes()
}

function onEdgesChange () {
  emitEdges()
}

function onNodeClick ({ node }) {
  selectedNode.value = vfNodes.value.find((n) => n.id === node.id)
  selectedEdgeId.value = null
}

function onEdgeClick ({ edge }) {
  selectedNode.value = null
  selectedEdgeId.value = edge.id
}

function onPaneClick () {
  selectedNode.value = null
  selectedEdgeId.value = null
}

function removeNode (id) {
  vfNodes.value = vfNodes.value.filter((n) => n.id !== id)
  vfEdges.value = vfEdges.value.filter((e) => e.source !== id && e.target !== id)
  selectedNode.value = null
  emitNodes()
  emitEdges()
  emitEntryNode()
}

function removeEdge (edgeId) {
  vfEdges.value = vfEdges.value.filter((e) => e.id !== edgeId)
  selectedEdgeId.value = null
  emitEdges()
}

function addNode () {
  const id = 'n' + Date.now().toString(36) + '_' + Math.random().toString(36).slice(2, 6)
  const cols = Math.max(1, Math.floor((vueFlowRef.value?.$el?.clientWidth || 800) / 280))
  const i = vfNodes.value.length
  vfNodes.value.push({
    id,
    type: 'rulechain',
    position: { x: 50 + (i % cols) * 280, y: 50 + Math.floor(i / cols) * 120 },
    data: { label: '', nodeType: '', configText: '{}', entry: false },
    draggable: true,
    connectable: true,
  })
  emitNodes()
  nextTick(() => {
    selectedNode.value = vfNodes.value.find((n) => n.id === id)
  })
}

defineExpose({ addNode })
</script>

<style>
@import '@vue-flow/core/dist/style.css';
@import '@vue-flow/core/dist/theme-default.css';
</style>

<style lang="scss" scoped>
.rulechain-graph {
  height: 100%;
  border-radius: 0.5rem;
  border: 1px solid var(--bs-border-color, #dee2e6);
}
</style>
