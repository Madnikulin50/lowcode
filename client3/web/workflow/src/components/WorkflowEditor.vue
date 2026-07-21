<template>
  <div
    id="editor"
    ref="editor"
    class="d-flex w-100 h-100"
  >
    <Teleport to="#topbar-title">
      {{ workflow.meta.name || workflow.handle }}
    </Teleport>

    <Teleport to="#topbar-tools">
      <button
        data-test-id="button-configure-workflow"
        class="btn btn-primary btn-sm d-flex align-items-center"
        @click="showWorkflowModal = !showWorkflowModal"
      >
        {{ t('configurator.configuration') }}
        <font-awesome-icon
          :icon="['fas', 'cog']"
          class="ms-1"
        />
      </button>
    </Teleport>

    <div class="toolbar d-flex flex-column h-100 border-end shadow-lg">
      <div
        id="toolbar"
        ref="toolbar"
        class="d-flex flex-column align-items-center mt-1 overflow-auto"
      />

      <div class="d-flex flex-grow-1 align-items-end justify-content-center py-3">
        <button
          ref="help"
          class="btn btn-outline-light d-flex align-items-center border-0 p-2"
          @click="showHelpModal = !showHelpModal"
        >
          <font-awesome-icon
            :icon="['far', 'question-circle']"
            class="h4 mb-0 text-primary"
          />
        </button>
      </div>
    </div>

    <div
      ref="tooltips"
      class="mh-100"
    />

    <div class="card w-100 h-100 border-0 shadow-sm rounded-0">
      <div class="card-body p-0">
        <div
          v-if="workflow.meta"
          class="position-absolute ps-2 pt-2"
          style="z-index: 1;"
        >
          <p
            v-if="workflow.meta.description"
            :class="{ 'mb-2': getRunAs }"
            class="mb-0 text-truncate"
            style="white-space: pre-line; max-height: 48px;"
          >
            {{ workflow.meta.description }}
          </p>

          <p
            v-if="getRunAs"
            class="mb-0 text-truncate"
          >
            <b>{{ t('editor.run-as') }}</b> <samp>{{ getRunAs }}</samp>
          </p>

          <div
            v-if="workflowLabelsDisplay.length > 0"
            class="mb-2"
          >
            <div
              v-for="group in workflowLabelsDisplay"
              :key="'group-' + group.namespaceID"
              class="d-flex align-items-center flex-wrap gap-1 mb-1"
            >
              <span
                title="Namespace"
                class="badge bg-primary"
                style="font-size: 90%;"
              >
                {{ group.namespaceName }}
              </span>

              <span
                v-for="mod in group.modules"
                :key="'mod-' + group.namespaceID + '-' + mod.id"
                title="Module"
                class="badge bg-extra-light"
                style="font-size: 90%;"
              >
                {{ mod.name }}
              </span>
            </div>
          </div>

          <div class="d-flex align-items-center mb-1">
            <h5
              v-if="workflow.deletedAt"
              class="mb-0 me-1"
            >
              <span class="badge bg-danger">{{ t('editor.deleted') }}</span>
            </h5>

            <h5
              v-if="!workflow.enabled"
              class="mb-0 me-1"
            >
              <span class="badge bg-danger">{{ t('editor.disabled') }}</span>
            </h5>

            <h5
              v-if="hasIssues"
              class="mb-0 me-1"
            >
              <span class="badge bg-danger">{{ t('editor.detected-issues') }}</span>
            </h5>

            <h5
              v-if="workflow.meta.subWorkflow"
              class="mb-0 me-1"
            >
              <span class="badge bg-info">{{ t('subworkflow') }}</span>
            </h5>

            <h5
              v-if="deferred"
              class="mb-0 me-1"
            >
              <span class="badge bg-info">{{ t('editor.deferred') }}</span>
            </h5>

            <h5
              v-if="triggersPathsChanged"
              class="mb-0 me-1"
            >
              <span class="badge bg-warning">{{ t('notification.trigger-paths-changed') }}</span>
            </h5>
          </div>
        </div>

        <div class="d-flex flex-wrap position-absolute fixed-bottom m-2 gap-2" style="z-index: 1;">
          <c-button-submit
            v-if="changeDetected && canUpdateWorkflow"
            data-test-id="button-save-workflow"
            variant="primary"
            :processing="processingSave"
            :text="t('editor.detected-changes') + `${canUpdateWorkflow ? t('editor.click-to-save') : ''}`"
            :loading-text="t('editor.saving')"
            class="rounded py-2 px-3"
            style="min-width: 20rem;"
            @submit="saveWorkflow()"
          />

          <div class="d-flex align-items-center bg-white border border-secondary py-2 px-3 ms-auto gap-1 rounded" style="z-index: 1;">
            {{ getZoomPercent }}
            <button
              class="btn btn-link ms-4 p-0"
              @click="zoom(false)"
            >
              <font-awesome-icon
                :icon="['fas', 'search-minus']"
              />
            </button>
            <button
              class="btn btn-link ms-1 p-0"
              @click="zoom()"
            >
              <font-awesome-icon
                :icon="['fas', 'search-plus']"
                class="pointer"
              />
            </button>
            <button
              class="btn btn-link ms-2 p-0 text-decoration-none"
              @click="resetZoom()"
            >
              {{ t('editor.reset') }}
            </button>
          </div>
        </div>

        <div
          id="graph"
          ref="graphRef"
          class="h-100 p-0"
        />
      </div>
    </div>

    <div
      v-if="sidebar.show"
      class="offcanvas offcanvas-end show"
      tabindex="-1"
      style="width: 600px;"
    >
      <div class="offcanvas-header bg-white border-bottom border-light p-2">
        <div class="d-flex align-items-center w-100 h5 mb-0 p-2">
          <img
            v-if="getSidebarItemIcon"
            :src="getSidebarItemIcon"
            class="me-2"
          >
          <h4 class="text-primary fw-bold mb-0">
            <b>{{ getSidebarItemType }}</b>
          </h4>

          <div class="ms-auto">
            {{ t('editor.id') }} <var>{{ getSelectedItem.node.id }}</var>
          </div>
        </div>
      </div>

      <div class="offcanvas-body bg-white">
        <transition
          name="component-fade"
          mode="out-in"
        >
          <configurator
            v-if="sidebar.showItem"
            v-model:item="sidebar.item"
            v-model:edges="edges"
            :out-edges="sidebar.outEdges"
            :is-subworkflow="!!workflow.meta.subWorkflow"
            @update-value="setValue($event)"
            @update-default-value="setValue($event, true)"
          />
        </transition>
      </div>

      <div class="offcanvas-footer bg-white border-top border-light p-1">
        <div class="d-flex m-2">
          <c-input-confirm
            size="md"
            size-confirm="md"
            variant="danger"
            :processing="processingDelete"
            :text="t('editor.delete')"
            @confirmed="sidebarDelete()"
          />

          <div class="ms-auto">
            <Teleport to="#sidebar-footer" />
          </div>
        </div>
      </div>
    </div>

    <workflow-configurator
      v-if="workflow.workflowID"
      :workflow="workflow"
      :can-create="canCreate"
      :processing-save="processingSave"
      :processing-delete="processingDelete"
      :import-processing="importProcessing"
      :show="showWorkflowModal"
      @update:show="showWorkflowModal = $event"
      @save="handleWorkflowSave"
      @import="importJSON"
      @delete="emit('delete')"
      @undelete="emit('undelete')"
    />

    <Teleport to="body">
      <div
        v-if="showHelpModal"
        class="modal fade show d-block"
        tabindex="-1"
        style="background: rgba(0,0,0,0.5);"
      >
        <div class="modal-dialog modal-lg modal-dialog-scrollable">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">{{ t('editor.help') }}</h5>
              <button type="button" class="btn-close" @click="showHelpModal = false"></button>
            </div>
            <div class="modal-body p-0">
              <help />
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <Teleport to="body">
      <div
        v-if="issuesModal.show"
        class="modal fade show d-block"
        tabindex="-1"
        style="background: rgba(0,0,0,0.5);"
      >
        <div class="modal-dialog modal-dialog-centered">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">{{ t('editor.issues') }}</h5>
              <button type="button" class="btn-close" @click="issuesModal.show = false"></button>
            </div>
            <div class="modal-body">
              <div
                v-for="(issue, index) in issuesModal.issues"
                :key="index"
              >
                <p>
                  <code>{{ issue[0].toUpperCase() + issue.slice(1) }}</code>
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <Teleport to="body">
      <div
        v-if="dryRun.show"
        class="modal fade show d-block"
        tabindex="-1"
        style="background: rgba(0,0,0,0.5);"
      >
        <div class="modal-dialog modal-lg modal-dialog-scrollable">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">{{ t('editor.initial-scope') }}</h5>
              <button type="button" class="btn-close" @click="dryRun.show = false"></button>
            </div>
            <div :class="dryRun.lookup ? 'modal-body' : 'modal-body p-1'">
              <div v-if="dryRun.lookup">
                <small>
                  {{ t('editor.input-ids-or-handles') }}<br>
                  {{ t('editor.modify-initial-scope-if-no-variables-are-loaded') }}<br>
                  {{ t('editor.auto-initialize-empty-variable') }}
                  <br><br>
                  {{ t('editor.open-webapp-on-prompt-use') }}
                </small>
                <div
                  v-for="(p, index) in Object.values(dryRun.initialScope)"
                  :key="index"
                  class="mt-4"
                >
                  <div
                    v-if="p.lookup"
                    class="mb-3"
                  >
                    <label class="form-label text-primary">{{ p.label }}</label>
                    <div class="form-text mb-2">{{ p.description }}</div>
                    <input
                      v-model="p.value"
                      class="form-control"
                    >
                  </div>
                </div>
              </div>
              <div
                v-else
                class="h-100"
              >
                <vue-json-editor
                  :value="dryRun.input"
                  :options="{ name: t('editor.initial-scope') }"
                  class="h-100"
                  @input="onDryRunEdit"
                />
              </div>
            </div>
            <div class="modal-footer">
              <button
                v-if="dryRun.lookup"
                type="button"
                class="btn btn-outline-secondary"
                @click="dryRun.lookup = true"
              >
                {{ t('editor.back') }}
              </button>
              <button
                type="button"
                class="btn btn-success"
                @click="dryRunOk"
              >
                {{ dryRun.lookup ? t('editor.load-and-configure') : t('editor.run-workflow') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>

<script setup>
import mxgraph from 'mxgraph'
import { ref, reactive, computed, watch, onMounted, onBeforeUnmount, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { encodeGraph } from '../lib/codec'
import { getStyleFromKind, getKindFromStyle } from '../lib/style'
import { encodeInput } from '../lib/dry-run'
import toolbarConfig from '../lib/toolbar'
import { getConstraintNameLabel } from '../lib/constraint'
import { camelToTitle } from '../lib/string'
import Configurator from '../components/Configurator'
import WorkflowConfigurator from '../components/Configurator/Workflow'
import Help from '../components/Help'
import VueJsonEditor from 'v-jsoneditor'
import { NoID } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'
import { useLabelsStore } from '../store'

const { t } = useI18n()
const labelsStore = useLabelsStore()
const $AutomationAPI = inject('automationAPI', {})
const $ComposeAPI = inject('composeAPI', {})
const $SystemAPI = inject('systemAPI', {})

globalThis.mxLoadResources = false
globalThis.mxForceIncludes = false
globalThis.mxResourceExtension = '.txt'
globalThis.mxLoadStylesheets = false

const {
  mxClient,
  mxGraph,
  mxEvent,
  mxUtils,
  mxCell,
  mxGeometry,
  mxUndoManager,
  mxGraphHandler,
  mxEdgeHandler,
  mxKeyHandler,
  mxDivResizer,
  mxToolbar,
  mxConstants,
  mxDragSource,
  mxRubberband,
  mxPerimeter,
  mxEdgeStyle,
  mxConnectionHandler,
  mxClipboard,
  mxPoint,
  mxRectangle,
  mxLog,
  mxImage,
  mxConstraintHandler,
  mxConnectionConstraint,
  mxCellState,
  mxEllipse,
  mxCellOverlay,
  mxCellHighlight,
} = mxgraph.call(globalThis, {})

mxClient.imageBasePath = `${document.getElementsByTagName('base')[0]?.href || import.meta.env.BASE_URL}icons`

const originPoint = -2042

const props = defineProps({
  workflowObject: {
    type: Object,
    default: () => {},
  },
  workflowTriggers: {
    type: Array,
    default: () => [],
  },
  changeDetected: {
    type: Boolean,
  },
  canCreate: {
    type: Boolean,
  },
  processingSave: {
    type: Boolean,
  },
  processingDelete: {
    type: Boolean,
  },
})

const emit = defineEmits(['save', 'delete', 'undelete'])

const editor = ref(null)
const toolbar = ref(null)
const graphRef = ref(null)
const tooltips = ref(null)

const initialized = ref(false)
const deferred = ref(false)
const triggersPathsChanged = ref(false)

const graph = ref(undefined)
const keyHandler = ref(undefined)
const undoManager = ref(undefined)

const workflow = ref({})
const triggers = ref([])
const vertices = ref({})
const edges = ref({})
const issues = ref({})

const highlights = ref([])
const runAsUser = ref(undefined)
const rendering = ref(false)

const sidebar = reactive({
  item: undefined,
  itemType: undefined,
  outEdges: 0,
  show: false,
  showItem: false,
})

const issuesModal = reactive({
  show: false,
  issues: [],
})

const dryRun = reactive({
  show: false,
  processing: false,
  lookup: false,
  cellID: undefined,
  initialScope: {},
  input: {},
  inputEdited: {},
  sessionID: undefined,
})

const selection = ref([])
const importProcessing = ref(false)
const zoomLevel = ref(1)
const currentLabel = ref(undefined)
const eventTypes = ref([])
const functionTypes = ref([])
const deferredKinds = ['delay', 'prompt']

const showWorkflowModal = ref(false)
const showHelpModal = ref(false)

const workflowLabelsDisplay = computed(() => {
  const namespaceIDs = []
  const modulesByNamespace = {}

  if (!workflow.value.labels) {
    return []
  }

  if (workflow.value.labels.ref_namespace) {
    const nsValues = Array.isArray(workflow.value.labels.ref_namespace)
      ? workflow.value.labels.ref_namespace
      : [workflow.value.labels.ref_namespace]

    nsValues.forEach(label => {
      const nsID = label.split('/')[1]
      if (nsID && !namespaceIDs.includes(nsID)) {
        namespaceIDs.push(nsID)
        labelsStore.resolveNamespace({ namespaceID: nsID, api: $ComposeAPI })
      }
    })
  }

  if (workflow.value.labels.ref_module) {
    const modValues = Array.isArray(workflow.value.labels.ref_module)
      ? workflow.value.labels.ref_module
      : [workflow.value.labels.ref_module]

    modValues.forEach(label => {
      const parts = label.split('/')
      const nsID = parts[1]
      const modID = parts[2]

      if (!nsID || !modID) return

      if (!namespaceIDs.includes(nsID)) {
        namespaceIDs.push(nsID)
        labelsStore.resolveNamespace({ namespaceID: nsID, api: $ComposeAPI })
      }

      if (!modulesByNamespace[nsID]) {
        modulesByNamespace[nsID] = []
      }

      labelsStore.resolveModule({ moduleID: modID, namespaceID: nsID, api: $ComposeAPI })
      const name = labelsStore.getModule(modID)
      modulesByNamespace[nsID].push({
        id: modID,
        name: name || modID,
      })
    })
  }

  return namespaceIDs.map(nsID => {
    const name = labelsStore.getNamespace(nsID)
    return {
      namespaceID: nsID,
      namespaceName: name || nsID,
      modules: modulesByNamespace[nsID] || [],
    }
  })
})

const getSidebarItemType = computed(() => {
  const { item = {} } = sidebar || {}
  const { style, edge } = item.node || {}

  if (edge) {
    return t('steps.path.short')
  }

  return t(`steps.${style}.short`) || style
})

const getSidebarItemIcon = computed(() => {
  const { item } = sidebar

  if (item && item.config) {
    return getIcon(getStyleFromKind(item.config).icon, currentTheme.value)
  }

  return undefined
})

const getSelectedItem = computed(() => {
  return sidebar.item ? sidebar.item : undefined
})

const getZoomPercent = computed(() => {
  return `${Math.floor(zoomLevel.value * 100).toFixed(0)}%`
})

const canUpdateWorkflow = computed(() => {
  return workflow.value.workflowID === '0' ? props.canCreate : workflow.value.canUpdateWorkflow
})

const hasIssues = computed(() => {
  return (workflow.value.issues || []).length
})

const getRunAs = computed(() => {
  if (runAsUser.value) {
    const { userID, name, username, email } = runAsUser.value
    return name || username || email || `<@${userID}>`
  }
  return undefined
})

const currentTheme = computed(() => {
  return window.__auth?.user?.meta?.theme || 'light'
})

watch(() => workflow.value.runAs, (runAs = '0') => {
  if (runAs !== '0') {
    $SystemAPI.userRead({ userID: runAs })
      .then(user => {
        runAsUser.value = user
      })
  } else {
    runAsUser.value = undefined
  }
}, { immediate: true })

watch(() => props.workflowObject, (wf) => {
  if (wf.workflowID !== workflow.value.workflowID) {
    showWorkflowModal.value = false
  }

  workflow.value = wf

  if (initialized.value) {
    renderWorkflow(workflow.value)
  }
}, { immediate: true })

watch(() => props.workflowTriggers, (triggers) => {
  triggers.value = triggers
}, { immediate: true })

onMounted(() => {
  try {
    if (!mxClient.isBrowserSupported()) {
      throw new Error(mxUtils.error(t('editor.unsupported-browser'), 200, false))
    }

    mxEvent.disableContextMenu(graphRef.value)
    graph.value = new mxGraph(graphRef.value, null, mxConstants.DIALECT_STRICTHTML)
    keyHandler.value = new mxKeyHandler(graph.value)

    setup()
    initToolbar()
    initUndoManager()
    initClipboard()
    keys()
    events()
    cellOverlay()
    styling()
    connectionHandler()
    getEventTypes()
    getFunctionTypes()

    window.addEventListener('trigger-updated', ({ detail: mxObjectId }) => {
      redrawLabel(mxObjectId)
    })

    renderWorkflow(workflow.value, true)

    if (workflow.value.workflowID && workflow.value.workflowID === '0') {
      showWorkflowModal.value = true
    }

    initialized.value = true
  } catch (e) {
    console.error(e)
  }
})

onBeforeUnmount(() => {
  graph.value?.destroy()
  keyHandler.value?.destroy()
  toolbar.value?.destroy?.()
  document.removeEventListener('keydown', keybinds)
})

function deleteSelectedCells () {
  if (sidebar.item && graph.value.isCellSelected(sidebar.item.node)) {
    sidebarClose()
  }
  graph.value.removeCells()
  clearHighlights()
}

function sidebarClose () {
  sidebar.show = false

  setTimeout(() => {
    const mxObjectId = sidebar.item.node.mxObjectId
    sidebar.showItem = false
    sidebar.item = undefined
    sidebar.itemType = undefined
    redrawLabel(mxObjectId)
  }, 300)
}

function sidebarDelete () {
  if (getSelectedItem.value) {
    graph.value.removeCells([getSelectedItem.value.node])
    sidebarClose()
  }
}

function sidebarReopen (item, itemType) {
  sidebar.outEdges = (item.node.edges || []).length

  if (!sidebar.show) {
    sidebar.item = item
    sidebar.itemType = itemType
    sidebar.show = true
    sidebar.showItem = true
    redrawLabel(item.node.mxObjectId)
  } else {
    if (sidebar.item && item.node.id === sidebar.item.node.id) {
      return
    }

    const oldMxObjectId = ((getSelectedItem.value || {}).node || {}).mxObjectId
    sidebar.showItem = false
    sidebar.item = item
    sidebar.itemType = itemType
    redrawLabel(oldMxObjectId)
    redrawLabel(item.node.mxObjectId)
    setTimeout(() => {
      sidebar.showItem = true
    }, 100)
  }
}

function setup () {
  graph.value.zoomFactor = 1.2

  graph.value.setBackgroundImage(new mxImage(getIcon('grid', currentTheme.value), 8192, 8192))
  graph.value.maximumGraphBounds = new mxRectangle(0, 0, 8192, 8192)
  graph.value.gridSize = 8

  graph.value.setPanning(true)
  graph.value.setConnectable(true)
  graph.value.setAllowDanglingEdges(false)
  graph.value.setTooltips(true)

  graph.value.container.style.overflow = 'hidden'

  graph.value.setPanning(true)
  graph.value.panningHandler.useLeftButtonForPanning = false
  graph.value.panningHandler.usePopupTrigger = true
  graph.value.panningHandler.ignoreCell = true
  graph.value.panningHandler.isForcePanningEvent = (me) => {
    const evt = me.getEvent()
    return mxEvent.isMiddleMouseButton(evt) || mxEvent.isRightMouseButton(evt)
  }

  const panningHandler = graph.value.panningHandler
  const originalMouseDown = panningHandler.mouseDown
  panningHandler.mouseDown = function (sender, me) {
    const evt = me.getEvent()
    const isPanButton = mxEvent.isMiddleMouseButton(evt) || mxEvent.isRightMouseButton(evt)
    if (isPanButton) {
      sender.container.style.cursor = 'grabbing'
    }
    if (originalMouseDown) {
      originalMouseDown.apply(this, arguments)
    }
  }

  const originalMouseUp = panningHandler.mouseUp
  panningHandler.mouseUp = function (sender, me) {
    sender.container.style.cursor = 'default'
    if (originalMouseUp) {
      originalMouseUp.apply(this, arguments)
    }
  }

  mxEvent.addListener(graph.value.container, 'mousedown', (evt) => {
    if (mxEvent.isMiddleMouseButton(evt) || mxEvent.isRightMouseButton(evt)) {
      graph.value.container.style.cursor = 'grabbing'
      graph.value.panningHandler.start(evt)
      mxEvent.consume(evt)
    }
  })
  mxEvent.disableContextMenu(graph.value.container)

  /* eslint-disable no-new */
  new mxRubberband(graph.value)
  graph.value.edgeLabelsMovable = false

  graph.value.getTooltipForCell = () => {}

  mxGraphHandler.prototype.guidesEnabled = true

  mxGraphHandler.prototype.useGuidesForEvent = (evt) => {
    return !mxEvent.isAltDown(evt.getEvent())
  }

  const mxGraphHandlerIsValidDropTarget = mxGraphHandler.prototype.isValidDropTarget
  mxGraphHandler.prototype.isValidDropTarget = function (target, me) {
    return mxGraphHandlerIsValidDropTarget.apply(this, arguments) && !target.edge
  }

  mxEdgeHandler.prototype.snapToTerminals = true

  mxGraph.prototype.minFitScale = 1
  mxGraph.prototype.maxFitScale = 1

  graph.value.isHtmlLabel = () => true
  graph.value.isWrapping = () => true

  graph.value.getLabel = cell => {
    let label = mxGraph.prototype.getLabel.apply(graph.value, arguments)

    const encodeHTML = (value = '') => {
      if (value) {
        return value.replace(/[\u00A0-\u9999<>&]/gim, i => {
          return '&#' + i.charCodeAt(0) + ';'
        })
      }
      return value
    }

    if (cell.edge) {
      if (cell.value) {
        label = `<div class="text-nowrap py-1 px-3 mb-0 rounded bg-white pointer" style="border: 2px solid #A7D0E3; border-radius: 5px; color: var(--dark);">${encodeHTML(cell.value)}</div>`
      }
    } else if (vertices.value[cell.id]) {
      const vertex = vertices.value[cell.id]
      const { kind } = vertex.config
      const { style } = vertex.node

      if (vertex && kind !== 'visual') {
        const icon = getIcon(getStyleFromKind(vertex.config).icon, currentTheme.value)
        const type = t(`steps.${style}.short`)
        const isSelected = selection.value.includes(cell.mxObjectId)
        const shadow = isSelected ? 'shadow' : 'shadow-sm'
        const issue = getIcon('issue')
        const playIcon = getIcon('play')
        const stopIcon = getIcon('stop')
        const opacity = kind === 'trigger' && !vertex.triggers.enabled ? 'opacity: 0.7;' : ''

        let test = ''
        let issuesHtml = ''
        let id = ''
        if (issues.value[cell.id]) {
          issuesHtml = `<img id="openIssues" src="${issue}" class="ms-2 pointer" style="width: 20px;"/>`
        } else {
          id = `<span class="show id-label">${cell.id}</span>`
        }

        let values = []

        if (kind === 'gateway' && cell.edges && cell.style !== 'gatewayParallel') {
          values = cell.edges
            .filter(({ source }) => cell.id === source.id)
            .map(({ id }) => edges.value[id])
            .map(({ node, config }) => `<tr><td><var>${encodeHTML(node.value)}</var></td><td><code>${encodeHTML(config.expr || '')}</code></td></tr>`)
            .join('')
        } else if (['expressions', 'function', 'prompt', 'iterator', 'exec-workflow', 'error-handler'].includes(kind)) {
          let { arguments: args = [], results = [], ref, kind } = vertex.config || {}

          if (!ref) { ref = kind }

          const { meta = {}, results: functionResults = [], parameters = [] } = functionTypes.value.find(f => f.ref === ref) || {}

          const functionLabel = meta.short

          if (functionLabel) {
            values.push(`<tr><td><b class="text-primary">${functionLabel}</b></td><td/></tr>`)
          }

          if (kind === 'expressions') {
            args = args.map(({ target, expr, type }) => {
              return `<tr><td><var>${encodeHTML(target)}</var> <samp>(${type})</samp></td><td><code>${encodeHTML(expr)}</code></td><td/></tr>`
            })
          } else {
            if (args.length) {
              values.push('<tr class="title"><td><b>Arguments</b></td><td/><td/></tr>')
            }

            args = parameters.map(({ name, types = [] }) => {
              const { type, expr, value } = args.find(({ target }) => target === name) || {}
              const exprType = type || `${types[0]}`
              const exprBadge = expr ? '<span title="Expression" class="circle-badge badge-small ms-1">e</span>' : ''

              return `<tr><td><var>${encodeHTML(name)}</var> <samp>(${exprType})</samp></td><td><code>${encodeHTML(expr || value)}</code></td><td>${exprBadge}</td></tr>`
            })
          }

          if (results.length) {
            args.push('<tr class="title border-top"><td><b>Results</b></td><td/><td/></tr>')
          }

          results = results.map(({ target = '', expr = '', value = '' }) => {
            const { types = [] } = functionResults.find(({ name }) => name === expr || name === value) || {}
            const type = types.length ? `(${types[0]})` : ''
            return `<tr><td><code>${encodeHTML(target)}</code> <samp>${type}</samp></td><td><var>${encodeHTML(expr || value)}</var></td><td/></tr>`
          })

          values = [...values, ...args, ...results].join('')
        } else if (kind === 'trigger') {
          let { resourceType = '', eventType = '', constraints = [] } = vertex.triggers || {}
          let { properties = [] } = eventTypes.value.find(et => resourceType === et.resourceType && eventType === et.eventType) || {}

          if (resourceType) {
            resourceType = resourceType.split(':')
              .map(part => part
                .split('-')
                .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
                .join(' '))
              .join(' - ')
          }

          if (eventType) {
            eventType = camelToTitle(eventType.replace('on', ''))
          }

          values.push('<tr class="title"><td><b>Configuration</b></td><td/><td/></tr>')
          values.push(`<tr><td><var>Resource</var></td><td/><td><code>${resourceType || ''}</code></td></tr>`)
          values.push(`<tr><td><var>Event</var></td><td/><td><code>${eventType || ''}</code></td></tr>`)

          if (constraints.length && eventType && eventType !== 'onManual') {
            constraints = [
              '<tr class="title"><td><b>Constraints</b></td><td/><td/></tr>',
              ...constraints.map(({ name = '', op = '', values = '' }) => {
                return `<tr><td><samp>${getConstraintNameLabel(name) || eventType}</samp></td><td><samp>${op}</samp></td><td><code>${encodeHTML(values.join(' or '))}</code></td></tr>`
              }),
            ]
          } else {
            constraints = []
          }

          if (properties.length) {
            properties = [
              '<tr class="title"><td><b>Initial scope</b></td><td/><td/></tr>',
              ...properties.map(({ name = '', type = '' }) => {
                return `<tr><td><var>${name}</var></td><td/><td><samp>${type || 'Any'}</samp></td></tr>`
              }),
            ]
          }

          values = [...values, ...constraints, ...properties].join('')
        } else if (['error', 'delay'].includes(kind)) {
          const { arguments: args = [] } = vertex.config || {}
          const { target, expr, value } = args[0] || {}

          if (target) {
            values = `<tr><td><var>${target}</var></td><td><code>${encodeHTML(expr || value)}</code></td></tr>`
          }
        } else {
          values = ''
        }

        if (values) {
          values = values
            ? '<div class="step-values rounded hide-label">' +
                '<table class="table bg-white shadow mb-0">' +
                  values +
                '</table>' +
              '</div>'
            : ''
        }

        if (workflow.value.canExecuteWorkflow && vertex.triggers && (cell.edges || []).length) {
          if (!dryRun.processing) {
            test = `<img id="testWorkflow" title="${t('configurator.tooltip.run-workflow')}" src="${playIcon}" class="hide pointer" style="width: 20px;"/>`
          } else if (dryRun.cellID === cell.id) {
            test = `<span class="spinner-border text-secondary" data-toggle="tooltip" data-placement="top" style="width: 20px; height: 20px; cursor: default;" title="Testing in progress. If your workflow includes Prompt or Delay steps, it may be waiting for them to complete">
                      <span class="sr-only">Spinning</span>
                    </span>`
            if (dryRun.sessionID) {
              test = test + `<img id="cancelWorkflow" src="${stopIcon}" class="ms-2 pointer" style="width: 20px; height: 20px;"/>`
            }
          }
        }

        label = `<div class="d-flex flex-column bg-white border rounded step position-relative ${shadow}" style="min-width: 200px; border-radius: 5px;${opacity}">` +
                  '<div class="label-container">' +
                    '<div class="d-flex flex-row align-items-center text-primary px-2 my-1 h6 mb-0" style="width: 200px; height: 36px;">' +
                      `<img src="${icon}" class="me-2"/>${type}` +
                      '<div class="d-flex h-100 ms-auto align-items-center">' +
                        test + id + issuesHtml +
                      '</div>' +
                    '</div>' +
                    `<div class="label d-flex flex-grow-1 align-items-stretch bg-white border-top ${values ? 'wide-label' : ''}" style="max-width: 200px; min-height: 36px;">` +
                      `<span class="d-inline-block hover-untruncate p-2 bg-white">${encodeHTML(cell.value || '/')}</span>` +
                    '</div>' +
                  '</div>' +
                  values +
                '</div>'
      } else {
        label = cell.value ? `<div class="rt-content text-wrap">${cell.value}</div>` : ''
      }
    }

    return label
  }

  graph.value.isCellEditable = () => false

  mxLog.setVisible = () => {}
  mxLog.DEBUG = false
  mxLog.TRACE = false

  mxGraph.prototype.expandedImage = undefined

  if (mxClient.IS_QUIRKS) {
    document.body.style.overflow = 'hidden'
    /* eslint-disable no-new */
    new mxDivResizer(graph.value.container)
  }

  if (mxClient.IS_NS) {
    mxEvent.addListener(graph.value.container, 'mousedown', () => {
      if (!graph.value.isEditing()) {
        graph.value.container.setAttribute('tabindex', '-1')
      }
    })
  }
}

function initToolbar () {
  toolbar.value = new mxToolbar(toolbar.value)
  graph.value.dropEnabled = true

  mxDragSource.prototype.getDropTarget = (graph, x, y) => {
    let cell = graph.getCellAt(x, y)
    if (!graph.isValidDropTarget(cell)) {
      cell = null
    }
    return cell
  }

  const addCell = ({ icon, width = 60, height = 60, style }) => {
    const { label, tooltip } = translateCell(style)
    let value = tooltip

    if (['break', 'continue'].includes(style)) {
      value = style === 'break' ? 'Stop iterator execution' : 'Skip current iteration'
    } else if (style.includes('gateway')) {
      value = style.split('gateway')[1]
    } else if (style === 'expressions') {
      value = 'Define and mutate scope variables'
    } else if (style === 'content') {
      value = 'Text here'
    }

    const cell = new mxCell(
      value,
      new mxGeometry(0, 0, width, height),
      style,
    )
    cell.setVertex(true)

    addToolbarItem(label, graph.value, toolbar.value, cell, icon, tooltip)
  }

  toolbarConfig.forEach(cell => {
    if (cell.kind === 'hr') {
      toolbar.value.addLine()
    } else if (cell.kind === 'nl') {
      toolbar.value.addBreak()
    } else {
      const cellStyle = getStyleFromKind(cell)
      if (cellStyle) {
        addCell({ ...cell, ...cellStyle })
      }
    }
  })
}

function initUndoManager () {
  undoManager.value = new mxUndoManager()
  const listener = (sender, evt) => {
    if (!rendering.value) {
      undoManager.value.undoableEditHappened(evt.getProperty('edit'))
    }
  }

  graph.value.getModel().addListener(mxEvent.UNDO, listener)
  graph.value.getView().addListener(mxEvent.UNDO, listener)
  graph.value.getModel().addListener(mxEvent.REDO, listener)
  graph.value.getView().addListener(mxEvent.REDO, listener)
}

function makeCellCopy ({ edge, id }) {
  const cell = edge ? edges.value[id] : vertices.value[id]
  const node = graph.value.model.cloneCell(cell.node, false)
  node.id = cell.node.id
  node.parent = cell.node.parent.id

  if (edge) {
    node.source = cell.node.source.id
    node.target = cell.node.target.id
  }

  const cellCopy = { node }

  if (cell.config) {
    cellCopy.config = JSON.parse(JSON.stringify(cell.config))
  }

  if (cell.triggers) {
    const triggers = {
      enabled: cell.triggers.enabled,
      constraints: cell.triggers.constraints,
      eventType: cell.triggers.eventType,
      resourceType: cell.triggers.resourceType,
    }
    cellCopy.triggers = JSON.parse(JSON.stringify(triggers))
  }

  return cellCopy
}

function initClipboard () {
  const absoluteGeometry = (cell) => {
    if (!cell.parent.geometry) {
      return { x: cell.geometry.x, y: cell.geometry.y }
    }
    const { x, y } = absoluteGeometry(cell.parent)
    return { x: cell.geometry.x + x, y: cell.geometry.y + y }
  }

  const copyCells = (cells, parentID) => {
    let copiedCells = {}
    let copiedEdges = []

    cells.forEach(cell => {
      const newCell = makeCellCopy(cell)

      if (cell.edge) {
        copiedEdges.push(newCell)
      } else {
        if (!copiedCells[cell.parent.id]) {
          copiedCells[cell.parent.id] = []
        }

        if (cell.parent.geometry && cell.parent.id !== parentID) {
          const { x, y } = absoluteGeometry(cell)
          newCell.node.geometry.x = x
          newCell.node.geometry.y = y
        }

        copiedCells[cell.parent.id].push(newCell)

        if (cell.children) {
          if (!copiedCells[cell.id]) {
            copiedCells[cell.id] = []
          }
          const childrenCells = copyCells(cell.children, cell.id)
          copiedCells = { ...copiedCells, ...childrenCells.cells }
          copiedEdges = [...copiedEdges, ...childrenCells.edges]
        }
      }
    })

    return { cells: copiedCells, edges: copiedEdges }
  }

  const pasteCells = (evt) => {
    if (evt.clipboardData.getData('text').includes('"cells":')) {
      const { cells = {}, edges = [] } = JSON.parse(evt.clipboardData.getData('text')) || {}

      const delta = mxClipboard.insertCount * graph.value.gridSize * 20
      const defaultParent = graph.value.getDefaultParent()
      const newCellIDs = {}
      const allCells = []

      graph.value.getModel().beginUpdate()

      Object.entries(cells).forEach(([parentID, children]) => {
        children.forEach(({ node, ...rest }) => {
          const parent = newCellIDs[parentID] ? graph.value.model.getCell(newCellIDs[parentID]) : defaultParent
          const { id, geometry, value, style } = node

          if (!newCellIDs[parentID]) {
            geometry.x += delta
            geometry.y += delta
          }

          const newVertex = graph.value.insertVertex(parent, null, value, geometry.x, geometry.y, geometry.width, geometry.height, style)
          allCells.push(newVertex)

          newCellIDs[id] = newVertex.id
          rest.config.stepID = newVertex.id

          vertices.value[newVertex.id] = { node: newVertex, ...rest }
        })
      })

      edges.forEach(({ node, ...rest }) => {
        const parent = newCellIDs[node.id] ? graph.value.model.getCell(newCellIDs[node.id]) : defaultParent
        const { id, geometry, value, style } = node

        const source = (vertices.value[newCellIDs[rest.config.parentID || node.source]] || {}).node
        const target = (vertices.value[newCellIDs[rest.config.childID || node.target]] || {}).node
        if (!source || !target) return

        node.source = source
        node.target = target

        const newEdge = graph.value.insertEdge(parent, null, value, node.source, node.target, style)
        newEdge.geometry.points = (geometry.points || []).map(({ x, y }) => new mxPoint(x, y))
        allCells.push(newEdge)

        newCellIDs[id] = newEdge.id
        rest.config.parentID = node.source.id
        rest.config.childID = node.target.id

        edges.value[newEdge.id] = { node: newEdge, ...rest }
      })

      Object.keys(vertices.value).forEach(vID => updateVertexConfig(vID))

      mxClipboard.insertCount++
      graph.value.setSelectionCells(allCells)
      graph.value.getModel().endUpdate()
    }
  }

  mxClipboard.copy = (graph, cells) => {
    const exportableCells = graph.getExportableCells(graph.model.getTopmostCells(cells || graph.getSelectionCells()))
    const copiedCells = copyCells(exportableCells)

    const editorEl = editor.value
    const tempInput = document.createElement('input')
    editorEl.appendChild(tempInput)
    tempInput.setAttribute('value', JSON.stringify(copiedCells))
    tempInput.select()
    document.execCommand('copy')
    editorEl.removeChild(tempInput)

    mxClipboard.insertCount = 1
    return copiedCells
  }

  mxClipboard.cut = (graph, cells) => {
    const copiedCells = mxClipboard.copy(graph, cells)

    const cutCells = []
    Object.entries(copiedCells.cells).forEach(([parentID, children]) => {
      children.forEach(({ node }) => {
        cutCells.push(graph.value.model.getCell(node.id))
      })
    })

    copiedCells.edges.forEach(({ node }) => {
      cutCells.push(graph.value.model.getCell(node.id))
    })

    mxClipboard.insertCount = 0
    graph.value.removeCells(cutCells)

    return cells
  }

  document.querySelector('body').addEventListener('paste', pasteCells)
}

function keybinds (event) {
  if ((event.ctrlKey || event.metaKey) && event.key === 's') {
    event.preventDefault()
    if (!document.getElementById('expression-editor')) {
      saveWorkflow()
    }
  }
}

function keys () {
  document.addEventListener('keydown', keybinds)

  keyHandler.value.getFunction = (evt) => {
    if (evt != null) {
      if (evt.ctrlKey || (mxClient.IS_MAC && evt.metaKey)) {
        if (evt.shiftKey) {
          return keyHandler.value.controlShiftKeys[evt.keyCode]
        }
        return keyHandler.value.controlKeys[evt.keyCode]
      }
      return keyHandler.value.normalKeys[evt.keyCode]
    }
    return null
  }

  keyHandler.value.controlKeys[88] = () => {
    mxClipboard.cut(graph.value, graph.value.getSelectionCells())
  }

  keyHandler.value.controlKeys[67] = () => {
    mxClipboard.copy(graph.value, graph.value.getSelectionCells())
  }

  keyHandler.value.controlKeys[65] = () => {
    graph.value.selectAll()
  }

  keyHandler.value.controlKeys[90] = () => {
    undoManager.value.undo()
    checkExistingTriggerPaths()
  }

  keyHandler.value.controlShiftKeys[90] = () => {
    undoManager.value.redo()
    checkExistingTriggerPaths()
  }

  keyHandler.value.normalKeys[8] = () => {
    deleteSelectedCells()
  }

  keyHandler.value.normalKeys[46] = () => {
    deleteSelectedCells()
  }

  keyHandler.value.controlKeys[32] = () => {
    if (graph.value.model.getChildCount(graph.value.getDefaultParent())) {
      graph.value.fit()
      graph.value.view.setTranslate(graph.value.view.translate.x + 79, graph.value.view.translate.y + 220)
      zoomLevel.value = graph.value.view.scale
    } else {
      resetZoom()
      graph.value.view.setTranslate(originPoint, originPoint)
    }
  }

  keyHandler.value.bindKey(191, (event) => {
    if (event.shiftKey && event.key === '?') {
      showHelpModal.value = true
    }
  })

  const nudge = (keyCode, evt) => {
    if (!graph.value.isSelectionEmpty()) {
      let dx = 0
      let dy = 0
      const delta = evt.shiftKey ? graph.value.gridSize : graph.value.gridSize * 5

      if (keyCode === 37) { dx = -delta }
      else if (keyCode === 38) { dy = -delta }
      else if (keyCode === 39) { dx = delta }
      else if (keyCode === 40) { dy = delta }

      graph.value.moveCells(graph.value.getSelectionCells(), dx, dy)
    }
  }

  keyHandler.value.bindKey(37, (evt) => { nudge(37, evt) })
  keyHandler.value.bindKey(38, (evt) => { nudge(38, evt) })
  keyHandler.value.bindKey(39, (evt) => { nudge(39, evt) })
  keyHandler.value.bindKey(40, (evt) => { nudge(40, evt) })
}

function checkExistingTriggerPaths () {
  triggersPathsChanged.value = [...triggers.value].some(({ stepID = '0', meta = {} }) => {
    if (stepID !== NoID) {
      let [triggerEdge] = vertices.value[meta.visual.id].node.edges || []

      if (triggerEdge) {
        triggerEdge = graph.value.model.getCell(triggerEdge.id)
        return triggerEdge.target && triggerEdge.target.id !== stepID
      } else {
        return true
      }
    }
    return false
  })
}

function events () {
  graph.value.getSelectionModel().addListener(mxEvent.CHANGE, (sender, evt) => {
    const cells = [...(evt.getProperty('added') || []), ...(evt.getProperty('removed') || [])]
    selection.value = graph.value.getSelectionCells().map(({ mxObjectId }) => mxObjectId)
    cells.forEach(({ mxObjectId }) => {
      redrawLabel(mxObjectId)
    })
  })

  graph.value.connectionHandler.addListener(mxEvent.CONNECT, (sender, evt) => {
    const node = evt.getProperty('cell')

    edges.value[node.id] = {
      node,
      config: { parentID: node.source.id, childID: node.source.id },
    }

    const source = vertices.value[node.source.id]
    const target = vertices.value[node.target.id]
    const outPaths = source.node.edges.filter(e => e.source.id === source.node.id) || []

    if (target.config.kind === 'gateway') {
      if (['join', 'fork'].includes(target.config.ref)) {
        updateVertexConfig(target.node.id)
      }
    }

    if (source.config.kind === 'gateway') {
      if (['join', 'fork'].includes(source.config.ref)) {
        updateVertexConfig(source.node.id)
      }

      if (source.config.ref === 'excl') {
        edges.value[node.id].node.value = `#${outPaths.length} - ${outPaths.length === 1 ? 'If' : 'Else (if)'}`
      } else if (source.config.ref === 'incl') {
        edges.value[node.id].node.value = 'If'
      }

      sidebar.outEdges = (source.node.edges || []).length
    } else if (source.config.kind === 'error-handler') {
      edges.value[node.id].node.value = `${outPaths.length === 1 ? 'Try' : 'Catch'}`
    } else if (source.config.kind === 'iterator') {
      edges.value[node.id].node.value = `${outPaths.length === 1 ? 'Body' : 'End'}`
    }
  })

  graph.value.addListener(mxEvent.CELL_CONNECTED, (sender, evt) => {
    if (!rendering.value) {
      const edge = evt.getProperty('edge')
      const source = vertices.value[edge.source.id]
      if (source.config.kind === 'trigger') {
        checkExistingTriggerPaths()
      }
    }
  })

  graph.value.addListener(mxEvent.CELLS_ADDED, (sender, evt) => {
    if (!rendering.value) {
      const cells = evt.getProperty('cells')
      let lastVertexID = null
      cells.forEach(cell => {
        if (cell && cell.vertex) {
          if (!rendering.value) {
            cell.defaultName = true
            addCellToVertices(cell)
            graph.value.setSelectionCells([cell])
            lastVertexID = cell.id
          }
        }
      })

      if (lastVertexID) {
        graph.value.view.validate()
        const vertex = vertices.value[lastVertexID]
        sidebarReopen(vertex, vertex.config.kind)
      }
    }
  })

  graph.value.addListener(mxEvent.CELLS_REMOVED, (sender, evt) => {
    const cells = evt.getProperty('cells') || []
    cells.forEach(cell => {
      if (cell.edge) {
        const source = vertices.value[cell.source.id]
        const target = vertices.value[cell.target.id]

        if (source.config.kind === 'gateway') {
          if (source.config.ref === 'excl') {
            source.node.edges.filter(e => e.source.id === source.node.id).forEach((edge, index) => {
              const [edgeID, ...rest] = edge.value.split(' - ')
              edges.value[edge.id].node.value = `#${index + 1} - ${rest.join(' - ')}`
              redrawLabel(edge.mxObjectId)
            })
          }

          if (['join', 'fork'].includes(target.config.ref)) {
            updateVertexConfig(source.node.id)
          }
        } else if (source.config.kind === 'iterator' || source.config.kind === 'error-handler') {
          graph.value.removeCells(source.node.edges.filter(e => e.source.id === source.node.id && e.id > cell.id))
        } else if (source.config.kind === 'trigger') {
          checkExistingTriggerPaths()
        }

        if (target.config.kind === 'gateway') {
          if (['join', 'fork'].includes(target.config.ref)) {
            updateVertexConfig(target.node.id)
          }
        }
      }
    })
  })

  graph.value.container.addEventListener('wheel', (event) => {
    if (event.ctrlKey || (mxClient.IS_MAC && event.metaKey)) {
      zoom(event.deltaY < 0)
      event.preventDefault()
      event.stopPropagation()
    } else {
      const view = graph.value.getView()
      view.setTranslate(
        view.translate.x - (event.deltaX || 0) / view.scale,
        view.translate.y - (event.deltaY || 0) / view.scale,
      )
      event.preventDefault()
    }
  }, { passive: false })

  graph.value.addMouseListener({
    mouseMove: (sender, evt) => {
      if (currentLabel.value !== null && evt.getState() === currentLabel.value) {
        return
      }

      let tmp = sender.view.getState(evt.getCell())

      if (tmp !== null && !sender.getModel().isVertex(tmp.cell)) {
        tmp = null
      }

      if (tmp !== currentLabel.value) {
        currentLabel.value = tmp
        if (currentLabel.value?.cell) {
          rendering.value = true
          sender.orderCells(false, [currentLabel.value.cell])
          rendering.value = false
        }
      }
    },

    mouseUp: (sender, evt) => {
      evt.consume()
    },

    mouseDown: (sender, evt) => {
      const event = evt.evt
      const cell = evt.state?.cell

      if (event) {
        const isAnchorPoint = event.target?.href?.baseVal?.includes('connection-point')

        if (isAnchorPoint) {
          evt.consume()
          return
        }

        if (mxEvent.isControlDown(event) || (mxClient.IS_MAC && mxEvent.isMetaDown(event))) {
          if (cell) {
            highlightConnectedPaths(cell)
          }
        } else if (cell) {
          clearHighlights()
          highlightConnectedPaths(cell)

          const item = cell.edge ? edges.value[cell.id] : vertices.value[cell.id]
          const itemType = cell.edge ? 'edge' : item.config.kind

          if (event.target.id === 'openIssues') {
            issuesModal.issues = issues.value[cell.id]
            issuesModal.show = true
          } else if (event.target.id === 'testWorkflow') {
            dryRun.cellID = cell.id
            loadTestScope()
          } else if (event.target.id === 'cancelWorkflow') {
            cancelWorkflow()
          } else {
            sidebarReopen(item, itemType)
          }
        } else if (!event.defaultPrevented) {
          graph.value.getSelectionModel().clear()
          sidebar.show = false
          if (getSelectedItem.value) {
            sidebarClose()
          }
          clearHighlights()
        }
      }

      evt.consume()
    },
  })

  graph.value.model.addListener(mxEvent.CHANGE, (sender, evt) => {
    if (!rendering.value) {
      window.dispatchEvent(new CustomEvent('change-detected'))
    }
  })
}

function styling () {
  mxConstants.VERTEX_SELECTION_COLOR = '#A7D0E3'
  mxConstants.VERTEX_SELECTION_STROKEWIDTH = 2
  mxConstants.VERTEX_SELECTION_DASHED = false
  mxConstants.EDGE_SELECTION_COLOR = '#A7D0E3'
  mxConstants.EDGE_SELECTION_STROKEWIDTH = 2
  mxConstants.DEFAULT_FONTFAMILY = 'var(--font-regular)'
  mxConstants.DEFAULT_FONTSIZE = 13

  mxConstants.HANDLE_FILLCOLOR = 'var(--primary)'
  mxConstants.HANDLE_STROKECOLOR = 'none'
  mxConstants.HANDLE_SIZE = 9
  mxConstants.CONNECT_HANDLE_FILLCOLOR = '#A7D0E3'
  mxConstants.OUTLINE_HIGHLIGHT_COLOR = '#A7D0E3'
  mxConstants.TARGET_HIGHLIGHT_COLOR = '#A7D0E3'
  mxConstants.DROP_TARGET_COLOR = '#A7D0E3'
  mxConstants.DEFAULT_VALID_COLOR = '#A7D0E3'
  mxConstants.VALID_COLOR = '#A7D0E3'
  mxGraphHandler.prototype.previewColor = '#A7D0E3'

  mxConstants.STYLE_PERIMETER = mxPerimeter.RectanglePerimeter

  mxConstants.GUIDE_COLOR = 'var(--dark)'
  mxConstants.GUIDE_STROKEWIDTH = 1

  let style = graph.value.getStylesheet().getDefaultVertexStyle()
  style[mxConstants.STYLE_SHAPE] = mxConstants.SHAPE_RECTANGLE
  style[mxConstants.STYLE_PERIMETER] = mxPerimeter.RectanglePerimeter
  style[mxConstants.STYLE_STROKECOLOR] = 'none'
  style[mxConstants.STYLE_STROKEWIDTH] = 0
  style[mxConstants.STYLE_ROUNDED] = true
  style[mxConstants.STYLE_ARCSIZE] = 5
  style[mxConstants.STYLE_RESIZABLE] = false
  style[mxConstants.STYLE_FILLCOLOR] = 'none'
  style[mxConstants.STYLE_FONTCOLOR] = 'var(--dark)'
  style[mxConstants.STYLE_FONTSIZE] = 13
  graph.value.getStylesheet().putDefaultVertexStyle(style)

  style = graph.value.getStylesheet().getDefaultEdgeStyle()
  style[mxConstants.STYLE_STROKECOLOR] = '#A7D0E3'
  style[mxConstants.STYLE_EDGE] = mxEdgeStyle.OrthConnector
  style[mxConstants.STYLE_ROUNDED] = true
  style[mxConstants.STYLE_ORTHOGONAL] = true
  style[mxConstants.STYLE_MOVABLE] = false
  style[mxConstants.STYLE_FONTCOLOR] = 'var(--dark)'
  style[mxConstants.STYLE_STROKEWIDTH] = 2
  style[mxConstants.STYLE_ENDSIZE] = 15
  style[mxConstants.STYLE_STARTSIZE] = 15
  style[mxConstants.STYLE_SOURCE_JETTY_SIZE] = 48
  style[mxConstants.STYLE_TARGET_JETTY_SIZE] = 48
  graph.value.getStylesheet().putDefaultEdgeStyle(style)

  style = {}
  style[mxConstants.STYLE_ROUNDED] = true
  style[mxConstants.STYLE_ARCSIZE] = 5
  style[mxConstants.STYLE_RESIZABLE] = true
  style[mxConstants.STYLE_SHAPE] = mxConstants.SHAPE_SWIMLANE
  style[mxConstants.STYLE_FONTSIZE] = 15
  style[mxConstants.STYLE_HORIZONTAL] = false
  style[mxConstants.STYLE_VERTICAL_LABEL_POSITION] = mxConstants.ALIGN_MIDDLE
  style[mxConstants.STYLE_VERTICAL_ALIGN] = mxConstants.ALIGN_MIDDLE
  style[mxConstants.STYLE_FILLCOLOR] = 'var(--white)'
  style[mxConstants.STYLE_STROKECOLOR] = 'var(--dark)'
  style[mxConstants.STYLE_STROKEWIDTH] = 1
  graph.value.getStylesheet().putCellStyle('swimlane', style)

  style = {}
  style[mxConstants.STYLE_RESIZABLE] = true
  style[mxConstants.STYLE_CONNECTABLE] = false
  style[mxConstants.STYLE_FILLCOLOR] = 'var(--white)'
  style[mxConstants.STYLE_STROKECOLOR] = 'var(--extra-light)'
  style[mxConstants.STYLE_STROKEWIDTH] = 1
  style[mxConstants.STYLE_VERTICAL_ALIGN] = mxConstants.ALIGN_TOP
  style[mxConstants.STYLE_ALIGN] = mxConstants.ALIGN_LEFT
  style[mxConstants.STYLE_SPACING_TOP] = 10
  style[mxConstants.STYLE_SPACING_LEFT] = 10
  style[mxConstants.STYLE_WHITE_SPACE] = 'wrap'
  style[mxConstants.STYLE_OVERFLOW] = 'hidden'
  graph.value.getStylesheet().putCellStyle('content', style)
}

function translateCell (style) {
  return {
    label: t(`steps.${style}.label`),
    tooltip: t(`steps.${style}.tooltip`),
  }
}

function cellOverlay () {
  mxCellOverlay.prototype.defaultOverlap = 1.2
}

function connectionHandler () {
  mxConstraintHandler.prototype.intersects = function (icon, point, source, existingEdge) {
    return (!source || existingEdge) || mxUtils.intersects(icon.bounds, point)
  }

  if (graph.value.connectionHandler.connectImage === null) {
    graph.value.connectionHandler.isConnectableCell = () => false
    mxEdgeHandler.prototype.isConnectableCell = cell => {
      return graph.value.connectionHandler.isConnectableCell(cell)
    }
  }

  graph.value.getAllConnectionConstraints = function (terminal, source = false) {
    if (!terminal) return null

    const { cell } = terminal

    let isConnectable = this.model.isVertex(cell) && !['swimlane', 'content'].includes(cell.style)

    if (cell.style.includes('trigger') && cell.edges) {
      isConnectable = isConnectable && !cell.edges.length
    }

    if (isConnectable) {
      let possibleConnections = [
        [0, 0], [0.25, 0], [0.5, 0], [0.75, 0], [1, 0],
        [1, 0.25], [1, 0.5], [1, 0.75], [1, 1],
        [0.75, 1], [0.5, 1], [0.25, 1], [0, 1],
        [0, 0.75], [0, 0.5], [0, 0.25],
      ]

      if (source) {
        const edges = cell.edges || []
        edges.forEach(({ source, target, style }) => {
          const points = {}
          if (style) {
            style.split(';').forEach(point => {
              const [key, value] = point.split('=')
              if (key && value) {
                points[key] = parseFloat(value)
              }
            })

            possibleConnections = possibleConnections.filter(([x, y]) => {
              if (source.id === cell.id) {
                return !(x === points.exitX && y === points.exitY)
              } else if (target.id === cell.id) {
                return !(x === points.entryX && y === points.entryY)
              }
              return true
            })
          }
        })
      } else {
        if (cell.style.includes('trigger')) {
          possibleConnections = []
        }
      }

      return possibleConnections.map(([x, y]) => new mxConnectionConstraint(new mxPoint(x, y), true))
    }

    return null
  }

  mxConnectionHandler.prototype.createEdgeState = function (me) {
    const edge = this.graph.createEdge(null, null, null, null, null)
    return new mxCellState(this.graph.view, edge, this.graph.getStylesheet().getDefaultEdgeStyle())
  }

  graph.value.resetEdgesOnMove = true
  mxGraph.prototype.resetEdges = function (cells) {
    if (cells != null) {
      this.model.beginUpdate()
      try {
        cells.forEach(cell => {
          const edges = this.model.getEdges(cell)
          if (edges != null) {
            edges.forEach(edge => { this.resetEdge(edge) })
          }
          this.resetEdges(this.model.getChildren(cell))
        })
      } finally {
        this.model.endUpdate()
      }
    }
  }

  mxConstraintHandler.prototype.pointImage = new mxImage(getIcon('connection-point'), 16, 16)

  mxConstraintHandler.prototype.createHighlightShape = function () {
    return new mxEllipse(null, '#A7D0E3', '#A7D0E3', 1)
  }
}

function addToolbarItem (title, graph, toolbar, prototype, icon, tooltip) {
  const funct = (graph, evt, cell) => {
    graph.stopEditing(false)

    const pt = graph.getPointForEvent(evt)
    const vertex = graph.getModel().cloneCell(prototype)
    vertex.geometry.x = pt.x
    vertex.geometry.y = pt.y

    graph.importCells([vertex], 0, 0, cell)
  }

  const dragElt = document.createElement('div')
  dragElt.style.border = 'dashed #A7D0E3 2px'
  dragElt.style.width = `${prototype.geometry.width}px`
  dragElt.style.height = `${prototype.geometry.height}px`

  const iconUrl = getIcon(icon, currentTheme.value)

  const img = toolbar.addMode(title, iconUrl, funct)

  const ds = mxUtils.makeDraggable(img, graph, funct, dragElt, null, null, graph.autoscroll, true)

  img.id = prototype.style.split(';')[0]
  img.setAttribute('title', tooltip || title)

  ds.createDragElement = mxDragSource.prototype.createDragElement
}

function addCellToVertices (cell) {
  const foundTrigger = triggers.value.find(({ meta }) => {
    return ((meta || {}).visual || {}).id === cell.id
  })

  const {
    kind = '',
    ref = '',
    defaultName = false,
    arguments: args,
    results = [],
    meta = {},
  } = (workflow.value.steps || []).find(({ stepID }) => {
    return stepID === cell.id
  }) || {}

  vertices.value[cell.id] = {
    node: cell,
    config: {
      stepID: cell.id,
      kind: kind || '',
      ref: ref || '',
      defaultName: defaultName || meta.visual?.defaultName || cell.defaultName || false,
      ...(rendering.value ? {} : getKindFromStyle(cell)),
    },
  }

  if (args) {
    vertices.value[cell.id].config.arguments = args
  }

  if (results) {
    vertices.value[cell.id].config.results = results
  }

  if (foundTrigger || cell.style === 'trigger') {
    vertices.value[cell.id].triggers = foundTrigger || {
      resourceType: null,
      eventType: null,
      constraints: [],
      enabled: true,
    }
  }
}

function updateVertexConfig (vID) {
  const { node, config } = vertices.value[vID]
  vertices.value[vID].config = { ...config, ...(rendering.value ? {} : getKindFromStyle(node)) }
}

function setValue (value, defaultName = false) {
  graph.value.model.setValue(sidebar.item.node, value)

  if (sidebar.itemType !== 'edge') {
    vertices.value[sidebar.item.node.id].config.defaultName = defaultName
  }
}

function zoom (up = true) {
  if (up && graph.value.view.scale < 3) {
    graph.value.zoomIn()
  } else if (!up && graph.value.view.scale > 0.1) {
    graph.value.zoomOut()
  }
  zoomLevel.value = graph.value.view.scale
}

function resetZoom () {
  graph.value.zoomTo(1)
  zoomLevel.value = graph.value.view.scale
}

function redrawLabel (id = '') {
  if (id) {
    const state = graph.value.view.states.map[id]
    if (state) {
      graph.value.cellRenderer.redrawLabel(state)
    }
  }
}

function clearHighlights () {
  if (highlights.value.length > 0) {
    highlights.value.forEach(h => { h.destroy() })
    highlights.value = []
    graph.value.clearCellOverlays()
  }
}

function highlightConnectedPaths (cell) {
  if (cell.vertex) {
    const edges = cell.edges || []
    edges.forEach(edge => {
      const state = graph.value.view.getState(edge)
      if (state) {
        const highlight = new mxCellHighlight(graph.value, 'var(--primary)', 2)
        highlight.highlight(state)
        highlights.value.push(highlight)
      }
    })
  } else if (cell.edge) {
    if (cell.source) {
      const sourceState = graph.value.view.getState(cell.source)
      if (sourceState) {
        const highlight = new mxCellHighlight(graph.value, 'var(--primary)', 2)
        highlight.highlight(sourceState)
        highlights.value.push(highlight)
      }
    }

    if (cell.target) {
      const targetState = graph.value.view.getState(cell.target)
      if (targetState) {
        const highlight = new mxCellHighlight(graph.value, 'var(--primary)', 2)
        highlight.highlight(targetState)
        highlights.value.push(highlight)
      }
    }
  }
}

async function loadTestScope () {
  if (props.changeDetected) {
    toastWarning(t('notification.save-workflow'), t('notification.failed-test'))
    return
  }

  if (hasIssues.value) {
    toastWarning(t('notification.resolve-issues'), t('notification.failed-test'))
    return
  }

  const lookupableTypes = [
    'record', 'oldRecord', 'module', 'oldModule', 'page', 'oldPage',
    'namespace', 'oldNamespace', 'user', 'oldUser', 'role', 'oldRole',
    'application', 'oldApplication',
  ]

  const { resourceType, eventType } = vertices.value[dryRun.cellID].triggers
  const et = (eventTypes.value.find(et => resourceType === et.resourceType && eventType === et.eventType) || {}).properties
  if (et) {
    let lookup = false
    if (et.length) {
      dryRun.initialScope = et.reduce((initialScope, p) => {
        let label = `${p.name}${lookupableTypes.includes(p.name) ? t('editor.id-parenthesis') : ''}`
        if (p.type === 'ComposeNamespace' || p.type === 'ComposeModule') {
          label = `${p.name} ${t('editor.handle')}`
        }

        let description = ''
        if (p.type === 'ComposeRecord') {
          description = t('editor.required-namespace-and-module')
        } else if (p.type === 'ComposeModule' || p.name === 'page' || p.name === 'oldPage') {
          description = t('editor.required-namespace')
        }

        initialScope[p.name] = {
          label,
          value: (dryRun.initialScope[p.name] || {}).value,
          lookup: lookupableTypes.includes(p.name),
          description,
        }

        lookup = lookup ? true : lookupableTypes.includes(p.name)
        return initialScope
      }, {})

      encodeInput(dryRun.initialScope, $ComposeAPI, $SystemAPI)
        .then(input => {
          dryRun.input = input
          dryRun.lookup = lookup
          dryRun.show = true
        })
        .catch(toastErrorHandler(t('notification.initial-scope-load-failed')))
    } else {
      dryRun.initialScope = {}
      testWorkflow()
    }
  } else {
    toastWarning(t('notification.event-type-not-found'), t('notification.failed-test'))
  }
}

async function dryRunOk (e) {
  if (dryRun.lookup) {
    e.preventDefault()
    encodeInput(dryRun.initialScope, $ComposeAPI, $SystemAPI)
      .then(input => {
        dryRun.input = input
        dryRun.inputEdited = input
        dryRun.lookup = false
      })
      .catch(toastErrorHandler(t('notification.initial-scope-load-failed')))
  } else {
    testWorkflow(dryRun.inputEdited)
  }
}

function onDryRunEdit (e) {
  dryRun.inputEdited = e
}

async function testWorkflow (input = {}) {
  clearHighlights()
  dryRun.processing = true
  redrawLabel(graph.value.model.getCell(dryRun.cellID).mxObjectId)

  const testParams = {
    workflowID: workflow.value.workflowID,
    stepID: vertices.value[dryRun.cellID].triggers.stepID,
    trace: workflow.value.canManageWorkflowSessions || false,
    wait: false,
    async: true,
    input,
  }

  toastInfo(t('notification.started-test'), t('notification.test-in-progress'))

  $AutomationAPI.workflowExec(testParams).then(({ sessionID, error: wfExecErr }) => {
    dryRun.sessionID = sessionID
    redrawLabel(graph.value.model.getCell(dryRun.cellID).mxObjectId)

    const pollSession = () => {
      return new Promise((resolve, reject) => {
        const checkSession = () => {
          $AutomationAPI.sessionRead({ sessionID }).then(session => {
            const { completedAt, status, stacktrace, error = false } = session

            setTimeout(() => {
              if (completedAt) {
                if (stacktrace) {
                  renderTrace(testParams.stepID, stacktrace)
                  if (status === 'completed') {
                    toastSuccess(t('notification.workflow-test-completed'), t('notification.test-completed'))
                  }
                } else {
                  toastWarning(t('notification.trace-unavailable'), t('notification.test-completed'))
                }

                if (error) {
                  reject(new Error(error))
                } else {
                  resolve()
                }
              } else {
                checkSession()
              }
            }, 1000)
          }).catch(reject)
        }

        checkSession()
      })
    }

    return pollSession()
  }).catch(toastErrorHandler(t('notification.failed-test')))
    .finally(() => {
      dryRun.lookup = true
      dryRun.processing = false
      dryRun.sessionID = undefined
      redrawLabel(graph.value.model.getCell(dryRun.cellID).mxObjectId)
    })
}

function cancelWorkflow () {
  const { sessionID, processing } = dryRun
  if (processing && sessionID) {
    dryRun.sessionID = undefined
    dryRun.processing = false
    redrawLabel(graph.value.model.getCell(dryRun.cellID).mxObjectId)

    $AutomationAPI.sessionCancel({ sessionID })
      .then(() => {
        toastInfo('Workflow test canceled', 'Stopping test')
      }).catch(e => {
        toastErrorHandler('Test cancel failed')(e)
      })
  }
}

function renderWorkflow (workflow, initial = false) {
  rendering.value = true

  if (sidebar.show) {
    sidebarClose()
  }

  clearHighlights()

  const { x = originPoint, y = originPoint } = graph.value.view.translate
  const { scale } = graph.value.view

  if (!workflow.steps) { workflow.steps = [] }
  if (!workflow.paths) { workflow.paths = [] }

  triggers.value.forEach(({ meta, ...config }) => {
    workflow.steps.push({
      stepID: meta.visual.id,
      kind: 'trigger',
      defaultName: meta.visual.defaultName || false,
      meta,
    })

    meta.visual.edges.forEach(edge => {
      workflow.paths.push(edge)
    })
  })

  issues.value = {}
  if (workflow.issues) {
    workflow.issues.forEach(({ culprit, description }) => {
      if (culprit) {
        const { step = -1, trigger = -1 } = culprit
        let stepID = ''

        if (step >= 0) {
          stepID = (workflow.steps[step] || {}).stepID
        } else if (trigger >= 0) {
          stepID = (triggers.value[trigger] || {}).meta?.visual?.id || ''
        }

        if (stepID) {
          issues.value[stepID] ? issues.value[stepID].push(description) : issues.value[stepID] = [description]
        }
      }
    })
  }

  deferred.value = false
  triggersPathsChanged.value = false

  const steps = workflow.steps || []
  const paths = workflow.paths || []
  const root = graph.value.getDefaultParent()

  vertices.value = {}
  edges.value = {}

  if (initial) {
    graph.value.view.rendering = false
  }

  graph.value.getModel().clear()

  graph.value.getModel().beginUpdate()

  try {
    steps.sort((a, b) => a.meta.visual.parent - b.meta.visual.parent)
      .forEach(({ meta = {}, ...config }) => {
        const node = (meta || {}).visual
        if (node) {
          node.parent = graph.value.model.getCell(node.parent) || root

          const { width, height, style } = getStyleFromKind(config)

          const newCell = graph.value.insertVertex(node.parent, node.id, node.value, node.xywh[0], node.xywh[1], node.xywh[2] || width, node.xywh[3] || height, style)
          addCellToVertices(newCell)

          deferred.value = deferred.value || deferredKinds.includes(config.kind)
        }
      })

    paths.forEach(({ meta, ...config }) => {
      const edge = (meta || {}).visual
      if (edge) {
        edge.parent = graph.value.model.getCell(edge.parent) || root
        edge.source = config.parentID || edge.source
        edge.target = config.childID || edge.target

        const newEdge = graph.value.insertEdge(edge.parent, edge.id, edge.value, vertices.value[edge.source].node, vertices.value[edge.target].node, edge.style)
        newEdge.geometry.points = (edge.points || []).map(({ x, y }) => new mxPoint(x, y))

        edges.value[edge.id] = { node: newEdge, config }
      }
    })

    Object.keys(vertices.value).forEach(vID => updateVertexConfig(vID))
  } finally {
    graph.value.view.scale = scale
    graph.value.view.setTranslate(x || originPoint, y || originPoint)

    if (undoManager.value && initial) {
      undoManager.value.clear()
    }

    graph.value.getModel().endUpdate()

    graph.value.getModel().nextId = graph.value.getModel().nextId + 1

    if (initial) {
      graph.value.fit()
      graph.value.view.rendering = true
      graph.value.refresh()

      graph.value.view.setTranslate(graph.value.view.translate.x + 79, graph.value.view.translate.y + 220)
      zoomLevel.value = graph.value.view.scale
    }

    rendering.value = false
  }
}

function renderTrace (firstStepID, trace = []) {
  const cells = {}
  clearHighlights()

  trace.filter(t => t).forEach(({ stepID, parentID, stepTime, error = false }, index) => {
    const cell = { index, stepID, parentID, stepTime, error }

    if (cells[stepID]) {
      cells[stepID].push(cell)
    } else {
      cells[stepID] = [cell]
    }
  })

  highlights.value = []

  const firstEdge = graph.value.model.getEdgesBetween(graph.value.model.getCell(dryRun.cellID), graph.value.model.getCell(firstStepID), true)[0]
  if (firstEdge) {
    highlights.value[highlights.value.push(new mxCellHighlight(graph.value, 'var(--success)', 2)) - 1].highlight(graph.value.view.getState(firstEdge))
  }

  Object.entries(cells).forEach(([stepID, frames]) => {
    if (stepID !== '0') {
      let error = frames[0].error
      let log = `#${frames[0].index + 1} - ${frames[0].stepTime}ms${error ? t('notification.error') + error : ''}`
      if (frames.length < 2) {
        const [cell] = frames
        if (cell && cell.index !== 0) {
          graph.value.model.getEdgesBetween(graph.value.model.getCell(cell.parentID), graph.value.model.getCell(stepID), true)
            .forEach(edge => {
              highlights.value[highlights.value.push(new mxCellHighlight(graph.value, 'var(--success)', 2)) - 1].highlight(graph.value.view.getState(edge))
            })
        }
      } else {
        const time = {
          min: frames[0].stepTime,
          max: frames[0].stepTime,
          avg: 0,
          sum: 0.0,
        }

        error = ''

        frames.forEach(({ index, parentID, stepTime, error }, i) => {
          if (i !== 0) {
            if (stepTime < time.min) { time.min = stepTime }
            if (stepTime > time.max) { time.max = stepTime }
            log = `${log}<br>#${index + 1} - ${stepTime}ms${error ? t('notification.error') + error : ''}`
          }

          time.sum += stepTime
          graph.value.model.getEdgesBetween(graph.value.model.getCell(parentID), graph.value.model.getCell(stepID), true)
            .forEach(edge => {
              highlights.value[highlights.value.push(new mxCellHighlight(graph.value, 'var(--success)', 2)) - 1].highlight(graph.value.view.getState(edge))
            })
        })

        time.avg = time.sum ? (time.sum / frames.length).toFixed(2) : time.sum
        log = `${log}<br><br>MIN: ${time.min}<br>MAX: ${time.max}<br>AVG: ${time.avg}<br>SUM: ${time.sum}`
      }

      const timeOverlay = new mxCellOverlay(new mxImage(getIcon(`clock-${error ? 'danger' : 'success'}`), 16, 16), `<span>${log}</span>`)
      graph.value.addCellOverlay(graph.value.model.getCell(stepID), timeOverlay)

      if (error) {
        highlights.value[highlights.value.push(new mxCellHighlight(graph.value, 'var(--danger)', 2)) - 1].highlight(graph.value.view.getState(graph.value.model.getCell(stepID)))
      } else {
        highlights.value[highlights.value.push(new mxCellHighlight(graph.value, 'var(--success)', 2)) - 1].highlight(graph.value.view.getState(graph.value.model.getCell(stepID)))
      }
    }
  })
}

function getJsonModel () {
  return encodeGraph(graph.value.getModel(), vertices.value, edges.value)
}

function importJSON (workflows = []) {
  try {
    importProcessing.value = true

    const [wf] = workflows

    triggers.value = wf.triggers || []

    workflow.value = {
      ...workflow.value,
      steps: wf.steps || [],
      paths: wf.paths || [],
    }

    renderWorkflow(workflow.value)

    importProcessing.value = false
    window.dispatchEvent(new CustomEvent('change-detected'))
    showWorkflowModal.value = false
    toastSuccess(t('notification.imported-workflow'))
  } catch (e) {
    toastErrorHandler(t('notification.import-failed'))(e)
  }
}

function handleWorkflowSave (wf) {
  workflow.value = wf
  saveWorkflow()
}

function saveWorkflow () {
  emit('save', { ...workflow.value, ...getJsonModel() })
}

async function getFunctionTypes () {
  return $AutomationAPI.functionList()
    .then(({ set }) => {
      functionTypes.value = [
        ...set,
        ...components.promptDefinitions,
        ...[
          {
            ref: 'error-handler',
            kind: 'error-handler',
            meta: { short: 'Handle error' },
            parameters: [],
            results: [{ name: 'error', types: ['Any'] }, { name: 'errorMessage', types: ['String'] }, { name: 'errorStepID', types: ['Integer'] }],
          },
          {
            ref: 'exec-workflow',
            kind: 'error-handler',
            meta: { short: 'Execute a workflow' },
            parameters: [{ name: 'workflow', types: ['ID', 'Handle'], required: true }, { name: 'scope', types: ['Vars'], required: false }],
            results: [],
          },
        ],
      ]
    })
    .catch(toastErrorHandler(t('notification.failed-fetch-functions')))
}

async function getEventTypes () {
  return $AutomationAPI.eventTypesList()
    .then(({ set }) => {
      eventTypes.value = set
    })
    .catch(toastErrorHandler(t('notification.event-type-fetch-failed')))
}

function getIcon (icon, mode = 'light') {
  if (!icon) return ''
  return `${mxClient.imageBasePath}/${mode === 'dark' ? 'dark/' : ''}${icon}.svg`
}

function toastWarning (title, msg) {
  console.warn(title, msg)
}

function toastInfo (title, msg) {
  console.info(title, msg)
}

function toastSuccess (title, msg) {
  console.log(title, msg)
}

function toastErrorHandler (msg) {
  return (e) => {
    console.error(msg, e)
  }
}
</script>

<style lang="scss" scoped>
#workflow-editor {
  color: var(--dark);
}

#graph {
  outline: none;
}

.toolbar {
  background-color: var(--sidebar-bg) !important;
  width: 86px;
}

.component-fade-enter-active, .component-fade-leave-active {
  transition: opacity 0.3s ease;
}

.component-fade-enter, .component-fade-leave-to {
  opacity: 0;
}

.saving::after {
  display: inline-block;
  animation: saving steps(1, end) 1s infinite;
  content: '';
}

@keyframes saving {
  0% { content: ''; }
  25% { content: '.'; }
  50% { content: '..'; }
  75% { content: '...'; }
  100% { content: ''; }
}
</style>

<style>
.hide {
  display: none;
}

.step:hover .hide {
  display: flex;
}

.show {
  display: flex;
}

.step:hover .show {
  display: none;
}

.hide-label {
  display: none;
}

.step:hover .hide-label {
  text-align: justify;
  display: flex;
}

.id-label {
  position: absolute;
  font-size: 8px;
  top: 4px;
  right: 4px;
}

.hover-untruncate {
  text-align: left;
  line-height: 18px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.step:hover .hover-untruncate {
  overflow: visible;
}

.label-container {
  overflow: hidden;
}

.step:hover .label-container {
  overflow-x: visible;
}

.step-values {
  position: absolute;
  min-width: 200px;
  top: 80px;
  border-top: 0;
}

.step-values td, th {
  text-align: left;
  padding: 8px;
  white-space: nowrap;
}

.step-values tr.title {
  background-color: var(--light) !important;
}

.step-values tr.title th {
  border-top: none;
}

#toolbar > hr {
  margin: 0.5rem 0 0.5rem 0 !important;
  align-self: stretch;
}
</style>
