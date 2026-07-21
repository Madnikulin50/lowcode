<template>
  <div class="d-flex overflow-auto p-2 w-100">
    <Teleport v-if="!fetchingReport" to="#topbar-title">{{ pageTitle }}</Teleport>

    <Teleport to="#topbar-tools">
      <div class="input-group" style="max-width: 300px;">
        <c-input-select
          v-model="scenarios.selected"
          :options="scenarioOptions"
          :get-option-key="getOptionKey"
          :placeholder="t('builder.pick-scenario')"
          :disabled="processing || fetchingReport"
          size="sm"
          @input="refreshReport()"
        />
        <button
          class="btn btn-extra-light"
          :title="t('builder.tooltip.configure-scenarios')"
          :disabled="!canUpdate"
          size="sm"
          @click="openScenarioConfigurator"
        >
          <font-awesome-icon :icon="['fas', 'cog']" class="text-primary" />
        </button>
      </div>

      <button
        :disabled="!canUpdate"
        class="btn btn-extra-light btn-sm"
        @click="openDatasourceConfigurator"
      >
        {{ t('builder.datasources.label') }}
      </button>
      <div class="btn-group btn-group-sm">
        <button
          class="btn btn-primary d-flex align-items-center justify-content-center"
          :disabled="!canRead"
          @click="$router.push(reportViewer)"
        >
          {{ t('builder.report.view') }}
          <font-awesome-icon class="ms-2" :icon="['far', 'eye']" />
        </button>
        <button
          class="btn btn-primary d-flex align-items-center justify-content-center"
          style="margin-left:2px"
          :title="t('builder.tooltip.edit.report')"
          :disabled="!canUpdate"
          @click="$router.push(reportEditor)"
        >
          <font-awesome-icon :icon="['far', 'edit']" />
        </button>
      </div>
    </Teleport>

    <div v-if="fetchingReport" class="d-flex align-items-center justify-content-center w-100 h-100">
      <div class="spinner-border" />
    </div>

    <grid
      v-if="report && canRead && showReport && !fetchingReport"
      :blocks="reportBlocks"
      editable
      @update:blocks="reportBlocks = $event"
      @item-updated="onBlockUpdated"
    >
      <template #default="{ block, index }">
        <div class="h-100">
          <div class="toolbox border-0 p-2 m-0 text-light text-center">
            <div
              v-if="unsavedBlocks.has(index)"
              class="btn border-0"
              :title="t('builder.tooltip.unsavedChanges')"
            >
              <font-awesome-icon :icon="['fas', 'exclamation-triangle']" class="text-warning" />
            </div>
            <div class="btn-group">
              <button
                class="btn btn-outline-light border-0"
                :title="t('builder.tooltip.add.displayElement')"
                @click="openDisplayElementSelector(index)"
              >
                <font-awesome-icon :icon="['fas', 'plus']" />
              </button>
              <button
                class="btn btn-outline-light border-0"
                :title="t('builder.tooltip.edit.block')"
                @click="editBlock(index)"
              >
                <font-awesome-icon :icon="['far', 'edit']" />
              </button>
            </div>
            <c-input-confirm
              :tooltip="t('builder.tooltip.delete.block')"
              show-icon
              size="md"
              class="ms-1"
              @confirmed="deleteBlock(index)"
            />
          </div>
          <block
            v-if="block"
            :index="index"
            :block="block"
            :scenario="currentSelectedScenario"
            :report-i-d="reportID"
            @item-updated="onBlockUpdated"
          />
        </div>
      </template>
    </grid>

    <Teleport to="body">
      <div v-if="showEditor" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5);">
        <div class="modal-dialog modal-xl modal-dialog-scrollable">
          <div class="modal-content">
            <div class="modal-header border-bottom-0">
              <h5 class="modal-title">{{ t('builder.block.configuration') }}</h5>
              <button type="button" class="btn-close" @click="hideEditorModal" />
            </div>
            <div class="modal-body p-0 border-top-0">
              <div v-if="currentBlock">
                <ul class="nav nav-tabs" role="tablist">
                  <li class="nav-item" role="presentation">
                    <button class="nav-link active" data-bs-toggle="tab" data-bs-target="#general-tab" type="button">
                      {{ t('builder.general') }}
                    </button>
                  </li>
                  <li class="nav-item" role="presentation">
                    <button
                      class="nav-link"
                      :class="{ active: !!currentBlock.elements.length }"
                      data-bs-toggle="tab"
                      data-bs-target="#elements-tab"
                      type="button"
                    >
                      {{ t('builder.elements') }}
                    </button>
                  </li>
                </ul>
                <div class="tab-content">
                  <div class="tab-pane active" id="general-tab">
                    <div class="mb-3">
                      <label class="text-primary form-label">{{ t('builder.title') }}</label>
                      <input v-model="currentBlock.title" class="form-control" type="text" :placeholder="t('builder.block.title')" />
                    </div>
                    <div class="mb-3">
                      <label class="text-primary form-label">{{ t('builder.description') }}</label>
                      <textarea v-model="currentBlock.description" class="form-control" :placeholder="t('builder.block.description')" />
                    </div>
                    <div class="mb-3">
                      <label class="text-primary form-label">{{ t('builder.layout') }}</label>
                      <div class="btn-group" role="group">
                        <input type="radio" class="btn-check" id="layout-h" value="horizontal" v-model="currentBlock.layout" />
                        <label class="btn btn-outline-primary" for="layout-h">{{ t('builder.layout-options.horizontal') }}</label>
                        <input type="radio" class="btn-check" id="layout-v" value="vertical" v-model="currentBlock.layout" />
                        <label class="btn btn-outline-primary" for="layout-v">{{ t('builder.layout-options.vertical') }}</label>
                      </div>
                    </div>
                  </div>
                  <div class="tab-pane" id="elements-tab">
                    <configurator
                      :items="currentDisplayElements"
                      :current-index="displayElements.currentIndex"
                      draggable
                      @select="setCurrentDisplayElement"
                      @add="openDisplayElementSelector(editor.currentIndex)"
                      @delete="deleteCurrentDisplayElement"
                    >
                      <template #label="{ item: { kind, name } }">
                        {{ name || kind }}
                        <font-awesome-icon :icon="['fas', 'bars']" class="text-secondary grab" />
                      </template>
                      <template #configurator>
                        <display-element-configurator
                          v-if="currentDisplayElement"
                          :display-element="currentDisplayElement"
                          :block="currentBlock"
                          :datasources="reportDatasources"
                          class="pe-2"
                          @update:display-element="currentDisplayElement = $event"
                        />
                      </template>
                    </configurator>
                  </div>
                </div>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-outline-secondary" @click="hideEditorModal">{{ t('builder.cancel-button') }}</button>
              <button type="button" class="btn btn-primary" @click="updateEditorBlock()">{{ t('builder.save-button') }}</button>
            </div>
          </div>
        </div>
      </div>

      <div v-if="datasources.showConfigurator" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5);">
        <div class="modal-dialog modal-xl modal-dialog-scrollable">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">{{ t('builder.datasources.label') }}</h5>
              <button type="button" class="btn-close" @click="hideDatasourceConfigurator" />
            </div>
            <div class="modal-body py-3">
              <configurator
                v-if="report"
                :items="datasources.tempItems"
                :current-index="datasources.currentIndex"
                draggable
                @select="setCurrentDatasource"
                @add="openDatasourceSelector()"
                @delete="deleteCurrentDataSource"
              >
                <template #label="{ item: { step } }">
                  <span class="d-inline-block text-truncate">{{ datasourceLabel(step, datasources.currentIndex) }}</span>
                </template>
                <template #configurator>
                  <component
                    :is="getDatasourceComponent(datasources.tempItems[datasources.currentIndex])"
                    v-if="currentDatasourceStep"
                    :index="datasources.currentIndex"
                    :datasources="datasources.tempItems"
                    :step="currentDatasourceStep"
                    :creating="datasources.tempItems[datasources.currentIndex].meta.creating"
                    @update:step="currentDatasourceStep = $event"
                  />
                </template>
              </configurator>
            </div>
            <div class="modal-footer">
              <c-button-submit
                data-test-id="button-save"
                :disabled="datasourceSaveDisabled"
                :processing="datasources.processing"
                :text="t('label.saveAndClose')"
                @submit="saveDatasources"
              />
            </div>
          </div>
        </div>
      </div>

      <div v-if="displayElements.showSelector" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5);">
        <div class="modal-dialog modal-lg modal-dialog-scrollable">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">{{ t('builder.add.display-element') }}</h5>
              <button type="button" class="btn-close" @click="displayElements.showSelector = false" />
            </div>
            <div class="modal-body px-0 py-3">
              <selector
                :items="displayElements.types"
                @select="addDisplayElement"
              />
            </div>
          </div>
        </div>
      </div>

      <div v-if="datasources.showSelector" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5);">
        <div class="modal-dialog modal-lg modal-dialog-scrollable">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">{{ t('builder.add.datasource') }}</h5>
              <button type="button" class="btn-close" @click="datasources.showSelector = false" />
            </div>
            <div class="modal-body px-0 py-3">
              <selector
                :items="datasources.types"
                display-mode="text"
                @select="addDatasource"
              />
            </div>
          </div>
        </div>
      </div>

      <div v-if="scenarios.showConfigurator" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5);">
        <div class="modal-dialog modal-xl modal-dialog-scrollable">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">{{ t('builder.scenarios.label') }}</h5>
              <button type="button" class="btn-close" @click="scenarios.showConfigurator = false" />
            </div>
            <div class="modal-body py-3">
              <configurator
                v-if="report"
                :items="reportScenarios"
                :current-index="scenarios.currentIndex"
                draggable
                @select="setCurrentScenario"
                @add="addScenario()"
                @delete="deleteCurrentScenario()"
              >
                <template #label="{ item: { label } }">
                  <span class="d-inline-block text-truncate">{{ label }}</span>
                </template>
                <template #configurator>
                  <scenario-configurator
                    v-if="currentScenario"
                    :current-index="scenarios.currentIndex"
                    :datasources="reportDatasources"
                    :scenario="currentScenario"
                    @update:scenario="currentScenario = $event"
                  />
                </template>
              </configurator>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-outline-secondary" @click="scenarios.showConfigurator = false">{{ t('label.cancel') }}</button>
              <button type="button" class="btn btn-primary" @click="scenarios.showConfigurator = false">{{ t('builder.scenarios.save') }}</button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <Teleport to="#report-toolbar">
      <editor-toolbar
        :back-link="{ name: 'report.list' }"
        :delete-disabled="!canDelete"
        :save-disabled="!canUpdate"
        :processing="processing"
        :processing-save="processingSave"
        :processing-delete="processingDelete"
        :processing-clone="processingClone"
        @clone="handleReportCloning"
        @delete="handleDelete"
        @save="handleReportSave"
      >
        <button
          class="btn btn-outline-secondary btn-lg"
          :disabled="processing"
          @click="createBlock"
        >
          <font-awesome-icon :icon="['fas', 'plus']" size="sm" />
          {{ t('label.add') }}
        </button>
      </editor-toolbar>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, watch, inject } from 'vue'
import { useRouter, useRoute, onBeforeRouteUpdate, onBeforeRouteLeave } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useToast } from 'corteza-lib/vue/dist'
import { cloneDeep } from 'lodash'
import { system, reporter } from 'corteza-lib/js/dist'
import { useReportHelpers } from '../../mixins/report'
import Grid from '../../components/Report/Grid.vue'
import Block from '../../components/Report/Blocks/index.vue'
import datasources from '../../components/Report/Datasources/loader'
import Configurator from '../../components/Common/Configurator.vue'
import Selector from '../../components/Common/Selector.vue'
import EditorToolbar from '../../components/EditorToolbar.vue'
import DisplayElementConfigurator from '../../components/Report/Blocks/DisplayElements/Configurators/index.vue'
import ScenarioConfigurator from '../../components/Report/Scenarios/index.vue'
import * as displayElementThumbnails from '../../assets/DisplayElements'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const toast = useToast()
const toastSuccess = toast.success
const toastErrorHandler = toast.toastErrorHandler
const can = inject('can', () => false)
const { fetchReport, handleSave: doSave, handleDelete: doDelete, handleClone: doClone } = useReportHelpers()

const processing = ref(false)
const processingSave = ref(false)
const processingDelete = ref(false)
const processingClone = ref(false)
const fetchingReport = ref(false)
const showReport = ref(true)
const report = ref(undefined)
const unsavedBlocks = ref(new Set())
const dataframes = ref([])

const blocks = ref({
  showConfigurator: false,
  currentIndex: undefined,
  items: [],
})

const displayElements = ref({
  showSelector: false,
  currentIndex: undefined,
  types: [
    { label: t('builder.display-elements.types.text'), kind: 'Text', value: displayElementThumbnails.Text },
    { label: t('builder.display-elements.types.metric'), kind: 'Metric', value: displayElementThumbnails.Metric },
    { label: t('builder.display-elements.types.table'), kind: 'Table', value: displayElementThumbnails.Table },
    { label: t('builder.display-elements.types.chart'), kind: 'Chart', value: displayElementThumbnails.Chart },
  ],
})

const datasourcesState = ref({
  showSelector: false,
  showConfigurator: false,
  processing: false,
  currentIndex: undefined,
  tempItems: [],
  types: [
    { label: t('builder.datasource.types.load.label'), kind: 'Load', value: t('builder.datasource.types.load.data-from-resource') },
    { label: t('builder.datasource.types.link.label'), kind: 'Link', value: t('builder.datasource.types.link.load-datasources') },
    { label: t('builder.datasource.types.join.label'), kind: 'Join', value: t('builder.datasource.types.join.load-datasources') },
    { label: t('builder.datasource.types.aggregate.label'), kind: 'Aggregate', value: t('builder.datasource.types.aggregate.load-datasource') },
  ],
})

const scenariosState = ref({
  showConfigurator: false,
  currentIndex: undefined,
  selected: undefined,
})

const editor = ref(undefined)

const reportID = computed(() => route.params.reportID)

const pageTitle = computed(() => {
  const title = report.value ? (report.value.meta.name || report.value.handle) : ''
  return title ? `${t('builder.report.builder')} - "${title}"` : t('builder.report.builder')
})

const canRead = computed(() => report.value ? report.value.canReadReport : false)
const canDelete = computed(() => report.value ? report.value.canDeleteReport : false)
const canUpdate = computed(() => report.value ? report.value.canUpdateReport : false)

const currentDisplayElements = computed(() => editor.value?.block?.elements || [])

const reportDatasources = computed({
  get: () => report.value ? report.value.sources : [],
  set: (sources) => { if (report.value) report.value.sources = sources },
})

const currentDatasourceStep = computed({
  get: () => datasourcesState.value.currentIndex !== undefined ? datasourcesState.value.tempItems[datasourcesState.value.currentIndex]?.step : undefined,
  set: (step) => {
    if (datasourcesState.value.currentIndex !== undefined) {
      datasourcesState.value.tempItems[datasourcesState.value.currentIndex].step = step
    }
  },
})

const reportBlocks = computed({
  get: () => blocks.value.items || [],
  set: (val) => { blocks.value.items = val },
})

const currentBlock = computed({
  get: () => editor.value ? editor.value.block : undefined,
  set: (block) => { if (editor.value && editor.value.currentIndex !== undefined) editor.value.block = block },
})

const currentDisplayElement = computed({
  get: () => displayElements.value.currentIndex !== undefined && currentBlock.value ? currentBlock.value.elements[displayElements.value.currentIndex] : undefined,
  set: (element) => {
    if (displayElements.value.currentIndex !== undefined && currentBlock.value) {
      const els = [...currentBlock.value.elements]
      els[displayElements.value.currentIndex] = element
      currentBlock.value.elements = els
    }
  },
})

const reportScenarios = computed({
  get: () => report.value ? report.value.scenarios : [],
  set: (scenarios) => { if (report.value) report.value.scenarios = scenarios },
})

const currentScenario = computed({
  get: () => scenariosState.value.currentIndex !== undefined ? reportScenarios.value[scenariosState.value.currentIndex] : undefined,
  set: (scenario) => {
    if (scenariosState.value.currentIndex !== undefined) {
      const s = [...reportScenarios.value]
      s[scenariosState.value.currentIndex] = scenario
      reportScenarios.value = s
    }
  },
})

const currentSelectedScenario = computed(() => {
  return scenariosState.value.selected ? reportScenarios.value.find(({ label }) => label === scenariosState.value.selected) : undefined
})

const scenarioOptions = computed(() => reportScenarios.value.map(({ label }) => label))

const reportViewer = computed(() => report.value ? { name: 'report.view', params: { reportID: report.value.reportID } } : undefined)
const reportEditor = computed(() => report.value ? { name: 'report.edit', params: { reportID: report.value.reportID } } : undefined)

const datasourceSaveDisabled = computed(() => {
  const uniqueDatasources = new Set()
  const hasDuplicates = datasourcesState.value.tempItems.some(({ step }) => {
    const name = step?.[Object.keys(step)[0]]?.name
    return !name || uniqueDatasources.size === uniqueDatasources.add(name).size
  })
  return datasourcesState.value.processing || hasDuplicates
})

const showEditor = computed(() => editor.value && editor.value.currentIndex !== undefined)

function refreshReport() {
  showReport.value = false
  setTimeout(() => { showReport.value = true }, 50)
}

function reindexBlocks(newBlocks) {
  const b = newBlocks || reportBlocks.value || []
  reportBlocks.value = b.map((block, i) => ({ ...block, i }))
}

function getDatasourceComponent({ step }) {
  if (!step) return undefined
  for (const s in step) {
    const ds = datasources(s)
    if (ds) return ds
  }
  return undefined
}

function datasourceLabel(step, currentIndex) {
  for (const v of Object.values(step)) {
    if (v && v.name) return v.name
  }
  return `${t('datasources.source')} ${currentIndex}`
}

function openDatasourceSelector() {
  datasourcesState.value.showSelector = true
  datasourcesState.value.currentIndex = datasourcesState.value.tempItems.length ? 0 : undefined
}

function openDatasourceConfigurator() {
  datasourcesState.value.showConfigurator = true
  datasourcesState.value.tempItems = cloneDeep(reportDatasources.value).map(ds => {
    ds.meta.creating = false
    return ds
  })
  datasourcesState.value.currentIndex = datasourcesState.value.tempItems.length ? 0 : undefined
}

function hideDatasourceConfigurator() {
  datasourcesState.value.showConfigurator = false
  datasourcesState.value.tempItems = []
  datasourcesState.value.currentIndex = undefined
}

function setCurrentDatasource(index) {
  datasourcesState.value.currentIndex = index
}

function deleteCurrentDataSource() {
  datasourcesState.value.tempItems.splice(datasourcesState.value.currentIndex, 1)
  datasourcesState.value.currentIndex = datasourcesState.value.tempItems.length ? 0 : undefined
}

function addDatasource(kind = '') {
  if (kind) {
    let step
    switch (kind) {
      case 'Aggregate':
        step = reporter.StepFactory({ aggregate: { name: 'Aggregate', keys: [], columns: [], filter: {}, sort: '' } })
        break
      case 'Link':
        step = reporter.StepFactory({ link: { name: 'Link', foreignColumn: '', foreignSource: '', localColumn: '', localSource: '' } })
        break
      case 'Join':
        step = reporter.StepFactory({ join: { name: 'Join', foreignColumn: '', foreignSource: '', localColumn: '', localSource: '' } })
        break
      default:
        step = reporter.StepFactory({ load: { name: 'Load', source: 'composeRecords', definition: {}, filter: {}, sort: '' } })
    }
    datasourcesState.value.tempItems.push({ step, meta: {} })
  }
  datasourcesState.value.currentIndex = datasourcesState.value.tempItems.length - 1
  datasourcesState.value.showSelector = false
  datasourcesState.value.showConfigurator = true
}

async function saveDatasources() {
  datasourcesState.value.processing = true
  try {
    const sources = datasourcesState.value.tempItems
    const { reportID: rid } = report.value
    const r = await window.__systemAPI.reportRead({ reportID: rid })
    const updated = await window.__systemAPI.reportUpdate(new system.Report({ ...r, sources }))
    updated.scenarios = report.value.scenarios
    report.value = new system.Report(updated)
    refreshReport()
    hideDatasourceConfigurator()
    toastSuccess(t('notification.report.datasources.updated'))
  } catch (e) {
    toastErrorHandler(t('notification.report.datasources.updateFailed'))(e)
  } finally {
    datasourcesState.value.processing = false
  }
}

async function handleReportSave() {
  processingSave.value = true
  try {
    report.value.blocks = reportBlocks.value.map(({ moved, x, y, w, h, i, ...p }) => {
      return { ...p, key: `${i}`, xywh: [x, y, w, h] }
    })
    const r = await doSave(report.value, false, toastSuccess, toastErrorHandler, t, router)
    if (r) report.value = r
    mapBlocks()
    refreshReport()
    unsavedBlocks.value.clear()
  } finally {
    processingSave.value = false
  }
}

async function handleReportCloning() {
  processingClone.value = true
  try {
    const r = await doClone(report.value, toastSuccess, toastErrorHandler, t)
    if (r) {
      window.dispatchEvent(new CustomEvent('refetch.reports'))
      router.push({ name: 'report.builder', params: { reportID: r.reportID } })
    }
  } finally {
    processingClone.value = false
  }
}

function mapBlocks() {
  reportBlocks.value = report.value.blocks.map(({ xywh, ...p }, i) => {
    const [x, y, w, h] = xywh
    return { ...p, x, y, w, h, i }
  })
}

function createBlock() {
  let newBlock = { ...new reporter.Block() }
  const [x, y, w, h] = newBlock.xywh
  newBlock = { ...newBlock, x, y, w, h }
  reindexBlocks([...reportBlocks.value, newBlock])
}

function updateEditorBlock(block) {
  const b = block || editor.value?.block
  if (!b) return
  const { currentIndex } = editor.value
  const items = [...reportBlocks.value]
  items[currentIndex] = b
  reportBlocks.value = items
  editor.value = undefined
  onBlockUpdated(currentIndex)
  refreshReport()
}

function editBlock(index) {
  const { x, y, w, h, i } = reportBlocks.value[index]
  const block = new reporter.Block(reportBlocks.value[index])
  block.x = x
  block.y = y
  block.w = w
  block.h = h
  block.i = i
  editor.value = { currentIndex: index, block }
  setCurrentDisplayElement(editor.value.block.elements.length ? 0 : undefined)
}

function deleteBlock(index) {
  reindexBlocks(reportBlocks.value.filter((_, i) => index !== i))
  unsavedBlocks.value.add(index)
}

function hideEditorModal() {
  editor.value = undefined
  displayElements.value.currentIndex = undefined
}

function openDisplayElementSelector(index) {
  blocks.value.currentIndex = index
  displayElements.value.showSelector = true
}

function setCurrentDisplayElement(index) {
  displayElements.value.currentIndex = index
}

function deleteCurrentDisplayElement() {
  if (currentBlock.value) {
    const els = [...currentBlock.value.elements]
    els.splice(displayElements.value.currentIndex, 1)
    currentBlock.value.elements = els
    displayElements.value.currentIndex = els.length ? 0 : undefined
  }
}

function addDisplayElement(kind) {
  const newDisplayElement = reporter.DisplayElementMaker({ kind })
  const items = [...reportBlocks.value]
  items[blocks.value.currentIndex].elements.push(newDisplayElement)
  reportBlocks.value = items
  displayElements.value.showSelector = false
  editBlock(blocks.value.currentIndex)
  setCurrentDisplayElement(currentBlock.value.elements.length - 1)
  unsavedBlocks.value.add(blocks.value.currentIndex)
}

function openScenarioConfigurator() {
  scenariosState.value.showConfigurator = true
  if (reportScenarios.value.length) {
    setCurrentScenario(0)
  }
}

function setCurrentScenario(index = -1) {
  scenariosState.value.currentIndex = reportScenarios.value.length && index >= 0 ? index : undefined
}

function addScenario() {
  const s = reportScenarios.value ? [...reportScenarios.value] : []
  s.push({ label: 'Scenario Name', filters: {} })
  reportScenarios.value = s
  setCurrentScenario(s.length - 1)
}

function deleteCurrentScenario() {
  const s = [...reportScenarios.value]
  s.splice(scenariosState.value.currentIndex, 1)
  reportScenarios.value = s
  scenariosState.value.currentIndex = s.length ? 0 : undefined
  setCurrentScenario(scenariosState.value.currentIndex)
}

function getOptionKey(scenario) { return scenario }

function checkUnsavedBlocks(next) {
  if (report.value.deletedAt) return next(true)
  next(!unsavedBlocks.value.size || window.confirm(t('builder.unsaved-changes')))
}

function onBlockUpdated(index) {
  unsavedBlocks.value.add(index)
}

onBeforeRouteUpdate((to, from, next) => checkUnsavedBlocks(next))
onBeforeRouteLeave((to, from, next) => checkUnsavedBlocks(next))

watch(reportID, (id) => {
  unsavedBlocks.value.clear()
  scenariosState.value.selected = undefined
  blocks.value.items = []
  report.value = undefined
  if (id) {
    processing.value = true
    fetchingReport.value = true
    fetchReport(id, toastErrorHandler, t).then(r => {
      report.value = r
      mapBlocks()
    }).catch(() => {
      toastErrorHandler(t('notification.report.loadFailed'))
    }).finally(() => {
      setTimeout(() => {
        fetchingReport.value = false
        processing.value = false
      }, 400)
    })
  }
}, { immediate: true })
</script>

<style lang="scss">
div.toolbox {
  position: absolute;
  background-color: var(--secondary);
  bottom: 0;
  left: 0;
  z-index: 1001;
  border-top-right-radius: 10px;
  opacity: 0.5;
  pointer-events: none;

  &:hover {
    opacity: 1;
  }

  & * {
    pointer-events: auto;
  }
}

[dir="rtl"] {
  div.toolbox {
    left: 0;
    right: auto;
  }
}
</style>
