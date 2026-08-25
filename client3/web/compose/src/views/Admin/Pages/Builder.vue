<template>
  <div
    v-if="page"
    id="page-builder"
    ref="pageBuilder"
    class="flex-grow-1 overflow-auto d-flex p-3 w-100 bg-light"
    tabindex="1"
  >
    <Teleport to="#topbar-title">
      {{ title }}
    </Teleport>

    <Teleport to="#topbar-tools">
      <div
        class="input-group"
        style="max-width: 300px;"
      >
        <c-input-select
          v-model="scenarios.selected"
          :options="scenarioOptions"
          :get-option-key="getScenarioOptionKey"
          :placeholder="$t('scenarios.pick-scenario')"
          :disabled="processing"
          size="sm"
          @input="refreshReport()"
        />
        <button
          data-bs-toggle="tooltip"
          :title="$t('scenarios.tooltip.configure-scenarios')"
          class="btn btn-extra-light"
          :disabled="!page?.canUpdatePage"
          @click="openScenarioConfigurator"
        >
          <font-awesome-icon
            :icon="['fas', 'cog']"
            class="text-primary"
          />
        </button>
      </div>

      <c-input-select
        v-if="page && layout && layouts.length > 1"
        ref="layoutSelect"
        :value="layout.pageLayoutID"
        :options="layouts"
        :reduce="l => l.pageLayoutID"
        size="sm"
        style="min-width: 250px; max-width: 300px;"
        class="me-2"
        @input="setLayout"
      />

      <router-link
        v-if="page && isRecordPage"
        variant="primary"
        :disabled="!moduleEditor"
        :to="moduleEditor"
        class="btn btn-primary btn-sm d-flex align-items-center me-2"
        style="margin-right:2px;"
      >
        {{ $t('navigation.editModule') }}
        <font-awesome-icon
          :icon="['far', 'edit']"
          class="ms-2"
        />
      </router-link>

      <div
        v-if="page && page.canUpdatePage"
        class="btn-group text-nowrap"
      >
        <router-link
          variant="primary"
          :to="pageViewer"
          class="btn btn-primary d-flex align-items-center"
        >
          {{ $t('navigation.viewPage') }}
          <font-awesome-icon
            :icon="['far', 'eye']"
            class="ms-2"
          />
        </router-link>

        <router-link
          data-bs-toggle="tooltip"
          :title="$t('tooltip.edit.page')"
          variant="primary"
          :to="pageEditor"
          class="btn btn-primary d-flex align-items-center"
          style="margin-left:2px;"
        >
          <font-awesome-icon
            :icon="['far', 'edit']"
          />
        </router-link>

        <page-translator
          v-model:page="trPage"
          v-model:page-layout="layout"
          button-variant="primary"
          style="margin-left:2px;"
        />
      </div>
    </Teleport>

    <div
      v-if="processingLayout"
      class="d-flex align-items-center justify-content-center w-100"
    >
      <span class="spinner-border" />
    </div>

    <grid
      v-else-if="layout"
      v-model:blocks="blocks"
      editable
      @item-updated="onBlockUpdated"
    >
      <template
        v-slot="{ blockIndex, block, resizing }"
      >
        <div
          v-if="block"
          :data-test-id="`block-${block.kind}`"
          class="h-100"
        >
          <div
            class="toolbox border-0 p-2 m-0 text-white text-center"
            data-test-id="block-toolbox"
          >
            <div
              v-if="unsavedBlocks.has(block.blockID !== '0' ? block.blockID : block.meta.tempID)"
              title="$t('label.unsavedChanges')"
              class="btn border-0"
            >
              <font-awesome-icon
                :icon="['fas', 'exclamation-triangle']"
                class="text-warning"
              />
            </div>

            <div class="btn-group">
              <button
                title="$t('tooltip.edit.block')"
                data-test-id="button-edit"
                class="btn btn-outline-light border-0"
                @click="editBlock(blockIndex)"
              >
                <font-awesome-icon
                  :icon="['far', 'edit']"
                />
              </button>

              <button
                title="$t('tooltip.clone.block')"
                class="btn btn-outline-light border-0"
                @click="cloneBlock(blockIndex)"
              >
                <font-awesome-icon
                  :icon="['far', 'clone']"
                />
              </button>

              <button
                title="$t('tooltip.copy.block')"
                class="btn btn-outline-light border-0"
                @click="copyBlock(blockIndex)"
              >
                <font-awesome-icon
                  :icon="['far', 'copy']"
                />
              </button>
            </div>

            <c-input-confirm
              :tooltip="$t('tooltip.delete.block')"
              show-icon
              link
              size="md"
              class="ms-1"
              @confirmed="deleteBlock(blockIndex)"
            />
          </div>

          <page-block
            v-bind="{
              ...$attrs,
              ...$props
            }"
            :page="page"
            :blocks="usedBlocks"
            :block-index="blockIndex"
            :block="block"
            :module="module"
            :record="record"
            :resizing="resizing"
            :unsaved-blocks="unsavedBlocks"
            editable
            class="p-2"
            @edit-block="editBlock"
            @clone-block="cloneTabbedBlock"
            @copy-block="copyBlock"
            @delete-tab="deleteTab"
          />
        </div>
      </template>
    </grid>

    <div
      id="createBlockSelector"
      ref="modalCreateBlockSelectorEl"
      class="modal fade"
      tabindex="-1"
    >
      <div class="modal-dialog modal-xl modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">
              {{ $t('build.selectBlockTitle') }}
            </h5>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
            />
          </div>
          <div class="modal-body">
            <new-block-selector
              :record-page="!!module"
              :existing-blocks="selectableExistingBlocks"
              @select="addBlock"
            />
          </div>
          <div
            v-if="!module"
            class="modal-footer text-muted small"
          >
            {{ $t('block.selectBlockFootnote') }}
          </div>
        </div>
      </div>
    </div>

    <div
      ref="modalCreatorEl"
      class="modal fade"
      tabindex="-1"
    >
      <div class="modal-dialog modal-xl modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header p-3 pb-0 border-bottom-0">
            <div class="d-flex gap-1 align-items-center">
              <h5 class="modal-title mb-3">
                {{ $t('block.general.title') }}
              </h5>
              <font-awesome-icon
                v-if="isEditorBlockReferenced"
                title="$t('referencedBlock')"
                :icon="['fas', 'exclamation-circle']"
                class="text-warning"
              />
            </div>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
            />
          </div>
           <div class="modal-body px-3 pb-3 pt-0 border-top-0 position-static">
            <page-blocks-configurator
              v-if="showCreator"
              :namespace="namespace"
              :module="module"
              :page="page"
              :blocks="usedBlocks"
              v-model:block="editor.block"
              :record="record"
            />
          </div>
          <div class="modal-footer">
            <button
              class="btn btn-outline-secondary"
              data-bs-dismiss="modal"
            >
              {{ $t('block.general.label.cancel') }}
            </button>
            <button
              class="btn btn-primary"
              :disabled="blockEditorOkDisabled"
              @click="updateBlocks()"
            >
              {{ $t('build.addBlock') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div
      ref="modalEditorEl"
      class="modal fade"
      tabindex="-1"
    >
      <div class="modal-dialog modal-xl modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header p-3 pb-0 border-bottom-0">
            <div class="d-flex gap-1 align-items-center">
              <h5 class="modal-title mb-3">
                {{ $t('changeBlock') }}
              </h5>
              <font-awesome-icon
                v-if="isEditorBlockReferenced"
                title="$t('referencedBlock')"
                :icon="['fas', 'exclamation-circle']"
                class="text-warning"
              />
            </div>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
            />
          </div>
           <div class="modal-body px-3 pb-3 pt-0 border-top-0 position-static">
            <page-blocks-configurator
              v-if="showEditor"
              :namespace="namespace"
              :module="module"
              :page="page"
              :blocks="usedBlocks"
              v-model:block="editor.block"
              :block-index="editor.index"
              :record="record"
            />
          </div>
          <div class="modal-footer d-flex justify-content-between">
            <c-input-confirm
              size="md"
              size-confirm="md"
              variant="danger"
              :text="$t('label.delete')"
              :tooltip="$t('label.delete')"
              @confirmed="deleteBlock(editor.index)"
            />

            <div>
              <button
                class="btn btn-outline-secondary me-2"
                data-bs-dismiss="modal"
              >
                {{ $t('label.cancel') }}
              </button>

              <button
                class="btn btn-primary"
                :title="$t('label.saveAndClose')"
                :disabled="blockEditorOkDisabled"
                @click="updateBlocks()"
              >
                {{ $t('label.saveAndClose') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div
      ref="modalScenariosEl"
      class="modal fade"
      tabindex="-1"
    >
      <div class="modal-dialog modal-xl modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">
              {{ $t('scenarios.scenarios.label') }}
            </h5>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
            />
          </div>
          <div class="modal-body py-3">
            <common-configurator
              v-if="page"
              :items="currentScenarios"
              :current-index="scenarios.currentIndex"
              draggable
              @select="setCurrentScenario"
              @add="addScenario()"
              @delete="deleteCurrentScenario()"
            >
              <template #label="{ item: { label } }">
                <span class="d-inline-block text-truncate">
                  {{ label }}
                </span>
              </template>
              <template #configurator>
                <scenario-configurator
                  v-if="currentScenario"
                  :current-index="scenarios.currentIndex"
                  v-model:scenario="currentScenario"
                />
              </template>
            </common-configurator>
          </div>
          <div class="modal-footer">
            <button
              class="btn btn-outline-secondary"
              data-bs-dismiss="modal"
            >
              {{ $t('label.cancel') }}
            </button>
            <button
              class="btn btn-primary"
              data-bs-dismiss="modal"
            >
              {{ $t('scenarios.scenarios.save') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <Teleport to="#admin-toolbar">
      <editor-toolbar
        :hide-save="!page?.canUpdatePage"
        :processing="processing"
        :processing-save="processingSave"
        :processing-save-and-close="processingSaveAndClose"
        :processing-delete="processingDelete"
        :processing-clone="processingClone"
        hide-clone
        @save="handleSaveLayout()"
        @delete="handleDeleteLayout()"
        @saveAndClose="handleSaveLayout({ closeOnSuccess: true })"
        @back="router.push(previousPage || { name: 'admin.pages' })"
      >
        <button
          v-if="page?.canUpdatePage"
          data-test-id="button-add-block"
          class="btn btn-outline-secondary btn-lg"
          @click="showBlockSelector"
        >
          + {{ $t('build.addBlock') }}
        </button>

        <template #saveAsCopy>
          <div
            v-if="page?.canUpdatePage"
            class="dropdown"
          >
            <button
              data-test-id="dropdown-saveAsCopy"
              class="btn btn-outline-secondary btn-lg dropdown-toggle"
              type="button"
              data-bs-toggle="dropdown"
              :disabled="processing"
              aria-expanded="false"
            >
              {{ $t('label.saveAsCopy') }}
            </button>
            <ul class="dropdown-menu m-0">
              <li>
                <button
                  data-test-id="dropdown-item-saveAsCopy-ref"
                  class="dropdown-item"
                  @click="handleCloneLayout({ ref: true })"
                >
                  {{ $t('build.saveAsCopy.ref') }}
                </button>
              </li>
              <li>
                <button
                  data-test-id="dropdown-item-saveAsCopy-noRef"
                  class="dropdown-item"
                  @click="handleCloneLayout({ ref: false })"
                >
                  {{ $t('build.saveAsCopy.noRef') }}
                </button>
              </li>
            </ul>
          </div>
        </template>
      </editor-toolbar>
    </Teleport>

    <record-modal
      :namespace="namespace"
    />

    <magnification-modal
      :namespace="namespace"
    />
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'page' } })
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useStore } from '../../../store'
import { useRouter, useRoute } from 'vue-router'

import { useUiStore } from '../../../store/ui'

const uiStore = useUiStore()
const { setPageHandle, setLayoutHandle } = uiStore
import NewBlockSelector from 'corteza-webapp-compose/src/components/Admin/Page/Builder/Selector'
import PageTranslator from 'corteza-webapp-compose/src/components/Admin/Page/PageTranslator'
import Grid from 'corteza-webapp-compose/src/components/Common/Grid'
import PageBlock from 'corteza-webapp-compose/src/components/PageBlocks'
import EditorToolbar from 'corteza-webapp-compose/src/components/Admin/EditorToolbar'
import { compose, NoID } from 'corteza-lib/js/dist'
import CommonConfigurator from 'corteza-webapp-compose/src/components/Common/Configurator'
import PageBlocksConfigurator from 'corteza-webapp-compose/src/components/PageBlocks/Configurator'
import RecordModal from 'corteza-webapp-compose/src/components/Public/Record/Modal'
import MagnificationModal from 'corteza-webapp-compose/src/components/Public/Page/Block/Modal'
import { fetchID } from 'corteza-webapp-compose/src/lib/block'
import { normalizeXYWH } from 'corteza-webapp-compose/src/lib/block-layout'
import { handle, useNsI18n } from 'corteza-lib/vue/dist'
import ScenarioConfigurator from 'corteza-webapp-compose/src/components/Public/Page/Scenarios'
import { Modal } from 'bootstrap'

const t = useNsI18n()
const store = useStore()
const router = useRouter()
const route = useRoute()

const props = defineProps({
  namespace: { type: compose.Namespace, required: true },
  pageID: { type: String, required: true },
})

const title = ref('')
const processing = ref(false)
const processingSave = ref(false)
const processingSaveAndClose = ref(false)
const processingClone = ref(false)
const processingDelete = ref(false)
const processingLayout = ref(false)
const page = ref(undefined)
const layout = ref(undefined)
const layouts = ref([])
const blocks = ref([])
const editor = ref(undefined)
const unsavedBlocks = ref(new Set())
const scenarios = ref({
  showConfigurator: false,
  currentIndex: undefined,
  selected: undefined,
})

const modalCreateBlockSelectorEl = ref(null)
const modalCreatorEl = ref(null)
const modalEditorEl = ref(null)
const modalScenariosEl = ref(null)


function getBlockSelectorModal() { return modalCreateBlockSelectorEl.value ? Modal.getOrCreateInstance(modalCreateBlockSelectorEl.value) : null }
function getBlockCreatorModal() { return modalCreatorEl.value ? Modal.getOrCreateInstance(modalCreatorEl.value) : null }
function getBlockEditorModal() { return modalEditorEl.value ? Modal.getOrCreateInstance(modalEditorEl.value) : null }
function getScenariosModal() { return modalScenariosEl.value ? Modal.getOrCreateInstance(modalScenariosEl.value) : null }

const pagesStore = computed(() => store.getters['page/set'])
const getModuleByID = computed(() => store.getters['module/getByID'])
const previousPage = computed(() => store.getters['ui/previousPage'])

const trPage = computed({
  get () { return page.value || new compose.Page() },
  set (v) {
    page.value = v
    store.dispatch('page/updateSet', v)
  },
})

const showEditor = computed(() => editor.value && editor.value.index !== undefined)
const showCreator = computed(() => editor.value && editor.value.index === undefined)

const module = computed(() => {
  if (!page.value || page.value.moduleID === NoID) return undefined
  return store.getters['module/getByID'](page.value.moduleID)
})

const isRecordPage = computed(() => !!module.value)

const moduleEditor = computed(() => {
  if (!module.value) return undefined
  return { name: 'admin.modules.edit', params: { moduleID: module.value.moduleID }, query: null }
})

const record = computed(() => {
  if (module.value) return new compose.Record({}, module.value)
  return null
})

const pageViewer = computed(() => {
  const name = module.value ? 'page.record.create' : 'page'
  return { name, params: { pageID: props.pageID } }
})

const pageEditor = computed(() => ({ name: 'admin.pages.edit', params: { pageID: props.pageID } }))

const hasChildren = computed(() => page.value ? pagesStore.value.some(({ selfID }) => selfID === page.value.pageID) : false)

const hideDelete = computed(() => hasChildren.value || !page.value.canDeletePage || !!page.value.deletedAt)

const selectableExistingBlocks = computed(() => {
  return page.value.blocks.filter(({ blockID }) => !usedBlocks.value.some(b => b.blockID === blockID))
})

const usedBlocks = computed(() => {
  const tabbedIDs = new Set()
  blocks.value.forEach(b => {
    if (b.kind !== 'Tabs') return
    const { tabs = [] } = b.options
    tabs.forEach(tab => {
      if (blocks.value.some(({ blockID }) => blockID === tab.blockID)) return
      const { blockID } = page.value.blocks.find(({ blockID }) => blockID === tab.blockID) || {}
      if (blockID) tabbedIDs.add(blockID)
    })
  })
  return [
    ...blocks.value.filter(({ blockID }) => !tabbedIDs.has(blockID)),
    ...page.value.blocks.filter(({ blockID }) => tabbedIDs.has(blockID)),
  ]
})

const otherLayoutBlockIDs = computed(() => {
  const set = new Set()
  return layouts.value.reduce((acc, { blocks: lyBlocks, pageLayoutID }) => {
    if (pageLayoutID === layout.value.pageLayoutID) return acc
    lyBlocks.forEach(({ blockID }) => acc.add(blockID))
    return acc
  }, set)
})

const isEditorBlockReferenced = computed(() => {
  const { block } = editor.value || {}
  if (!block || block.blockID === NoID) return false
  return otherLayoutBlockIDs.value.has(editor.value.block.blockID)
})

const blockEditorOkDisabled = computed(() => {
  if (!editor.value) return true
  const { block } = editor.value
  if (!block) return true
  const { customCSSClass, customID } = block.meta || {}
  return [handle.handleState(customID), handle.classState(customCSSClass)].includes(false)
})

const currentScenarios = computed({
  get () { return page.value ? (page.value.meta.scenarios || []) : [] },
  set (val) { if (page.value) page.value.meta.scenarios = val },
})

const currentScenario = computed({
  get () { return scenarios.value.currentIndex !== undefined ? currentScenarios.value[scenarios.value.currentIndex] : {} },
  set (val) {
    if (scenarios.value.currentIndex !== undefined) {
      currentScenarios.value[scenarios.value.currentIndex] = val
    }
  },
})

const currentSelectedScenario = computed(() => {
  return scenarios.value.selected ? currentScenarios.value.find(({ label }) => label === scenarios.value.selected) : undefined
})

const scenarioOptions = computed(() => currentScenarios.value.map(({ label }) => label))

watch(() => props.pageID, (pageID) => {
  processingLayout.value = true
  unsavedBlocks.value.clear()
  layouts.value = []
  layout.value = undefined
  const { namespaceID } = props.namespace

  store.dispatch('page/findByID', { namespaceID, pageID, force: true }).then(p => {
    const { title: ttl = '', handle } = p
    title.value = ttl || handle
    document.title = t('label.app-name.page.builder', { label: title.value, interpolation: { escapeValue: false } })
    page.value = p.clone()
    return fetchPageLayouts().then(() => { setLayout() })
  }).catch(() => { processingLayout.value = false })
}, { immediate: true })

watch(showCreator, (val) => {
  const m = getBlockCreatorModal()
  if (!m) return
  val ? m.show() : m.hide()
})

watch(showEditor, (val) => {
  const m = getBlockEditorModal()
  if (!m) return
  val ? m.show() : m.hide()
})

watch(() => scenarios.value.showConfigurator, (val) => {
  const m = getScenariosModal()
  if (!m) return
  val ? m.show() : m.hide()
})

watch(() => page.value ? page.value.handle : undefined, (handle, oldHandle) => {
  if (handle !== oldHandle) setPageHandle(handle)
}, { immediate: true })

watch(() => layout.value ? layout.value.handle : undefined, (handle, oldHandle) => {
  if (handle !== oldHandle) setLayoutHandle(handle)
}, { immediate: true })

onMounted(() => {
  window.addEventListener('paste', pasteBlock)
})

onBeforeUnmount(() => {
  getBlockSelectorModal()?.hide()
  getBlockCreatorModal()?.hide()
  getBlockEditorModal()?.hide()
  getScenariosModal()?.hide()
  if (!['page', 'page.record', 'page.record.create', 'page.record.edit'].includes(route.name)) {
    setPageHandle('')
    setLayoutHandle('')
  }
  destroyEvents()
  setDefaultValues()
  getBlockSelectorModal()?.dispose()
  getBlockCreatorModal()?.dispose()
  getBlockEditorModal()?.dispose()
  getScenariosModal()?.dispose()
})

watch([modalCreateBlockSelectorEl, modalCreatorEl, modalEditorEl, modalScenariosEl], ([m1, m2, m3, m4]) => {
  if (m1 && m2 && m3 && m4) {
    m2.addEventListener('hidden.bs.modal', () => { if (editor.value && editor.value.index === undefined) editor.value = undefined })
    m3.addEventListener('hidden.bs.modal', () => { if (editor.value && editor.value.index !== undefined) editor.value = undefined })
    m4.addEventListener('hidden.bs.modal', () => { scenarios.value.showConfigurator = false })
  }
})

function toastSuccess (msg) {}
function toastErrorHandler (msg) { return (e) => {} }
function toastWarning (msg) {}

function openScenarioConfigurator () {
  scenarios.value.showConfigurator = true
  if (currentScenarios.value.length) setCurrentScenario(0)
}

function setCurrentScenario (index = -1) {
  scenarios.value.currentIndex = currentScenarios.value.length && index >= 0 ? index : undefined
}

function addScenario () {
  if (!currentScenarios.value) currentScenarios.value = []
  currentScenarios.value.push({ label: 'Scenario Name', filters: {} })
  setCurrentScenario(currentScenarios.value.length - 1)
}

function deleteCurrentScenario () {
  currentScenarios.value.splice(scenarios.value.currentIndex, 1)
  scenarios.value.currentIndex = currentScenarios.value.length ? 0 : undefined
  setCurrentScenario(scenarios.value.currentIndex)
}

function getScenarioOptionKey (scenario) { return scenario }

function refreshReport () {}

function showBlockSelector () {
  getBlockSelectorModal()?.show()
}

function addBlock (block, index = undefined) {
  getBlockSelectorModal()?.hide()
  calculateNewBlockPosition(block)
  editor.value = { index, block: compose.PageBlockMaker(block) }
}

function editBlock (index = undefined) {
  nextTick(() => {
    editor.value = { index, block: compose.PageBlockMaker(blocks.value[index]) }
  })
}

function deleteBlock (index) {
  if (blocks.value[index].meta.hidden) {
    blocks.value.forEach((b) => {
      if (b.kind !== 'Tabs' || !b.options.tabs.some(({ blockID }) => blockID === fetchID(blocks.value[index]))) return
      b.options.tabs = b.options.tabs.filter(({ blockID }) => blockID !== fetchID(blocks.value[index]))
    })
  }
  const block = blocks.value[index]
  blocks.value.splice(index, 1)
  if (block.blockID !== NoID) unsavedBlocks.value.add(block.blockID)
  else unsavedBlocks.value.delete(block.meta.tempID)
  if (block.kind === 'Tabs') showUntabbedHiddenBlocks()
  if (editor.value) editor.value = undefined
}

function deleteTab ({ blockIndex, tabIndex }) {
  const { blockID } = blocks.value[blockIndex] || {}
  if (!blockID) return
  unsavedBlocks.value.add(blockID)
  blocks.value[blockIndex].options.tabs.splice(tabIndex, 1)
  showUntabbedHiddenBlocks()
}

function showUntabbedHiddenBlocks () {
  const tabbedBlocks = new Set()
  blocks.value.forEach(b => {
    if (b.kind !== 'Tabs') return
    b.options.tabs.forEach(({ blockID }) => tabbedBlocks.add(blockID))
  })
  blocks.value.forEach((b, index) => {
    if (!b.meta.hidden || tabbedBlocks.has(fetchID(b))) return
    blocks.value[index].meta.hidden = false
    calculateNewBlockPosition(blocks.value[index])
  })
  tabbedBlocks.clear()
}

function onBlockUpdated (index) {
  unsavedBlocks.value.add(fetchID(blocks.value[index]))
}

function updateBlocks (block) {
  if (!editor.value) return
  block = compose.PageBlockMaker(block || editor.value.block)
  const creatingTabbedBlock = editor.value.block.kind !== block.kind

  if (creatingTabbedBlock) {
    // $root emit - would need provide/inject
  }

  if (editor.value.index !== undefined && !creatingTabbedBlock) {
    const oldBlock = blocks.value[editor.value.index]
    if (oldBlock.meta.hidden === true && editor.value.block.meta.hidden === false) {
      // untabBlock(editor.value.block)
      calculateNewBlockPosition(block)
    }
    blocks.value.splice(editor.value.index, 1, block)
    unsavedBlocks.value.add(fetchID(block))
  } else {
    blocks.value.push(block)
    unsavedBlocks.value.add(fetchID(block))
    scrollToBottom()
  }

  if (block.kind === 'Tabs') {
    block.options.tabs.forEach((tab) => {
      if (!tab.blockID) return
      let tabbedBlock = blocks.value.find(b => fetchID(b) === tab.blockID)
      if (!tabbedBlock) {
        tabbedBlock = page.value.blocks.find(({ blockID }) => blockID === tab.blockID)
        if (tabbedBlock) blocks.value.push(tabbedBlock)
      }
      if (tabbedBlock) tabbedBlock.meta.hidden = true
    })
    showUntabbedHiddenBlocks()
  }

  if (editor.value.block.kind === block.kind) editor.value = undefined
}

function cloneBlock (index) {
  appendBlock(blocks.value[index].clone(), t('notification.page.cloneSuccess'))
}

function cloneTabbedBlock ({ tabbedBlockIndex, tabBlockIndex, title }) {
  const block = blocks.value[tabbedBlockIndex].clone()
  block.meta.hidden = true
  blocks.value[tabBlockIndex].options.tabs.push({ blockID: fetchID(block), title })
  blocks.value.push(block)
  unsavedBlocks.value.add(fetchID(block))
}

function appendBlock (block, msg) {
  calculateNewBlockPosition(block)
  editor.value = { index: undefined, block }
  updateBlocks()
  if (!editor.value) {
    msg && toastSuccess(msg)
    return true
  } else {
    msg && toastErrorHandler(t('notification.page.duplicateFailed'))
    return false
  }
}

function calculateNewBlockPosition (block) {
  if (blocks.value.length) {
    const maxY = blocks.value.filter(({ meta }) => !meta.hidden).map((b) => b.xywh[1]).reduce((acc, val) => acc > val ? acc : val, 0)
    block.xywh = [0, maxY + 2, 20, 15]
  }
}

async function fetchPageLayouts () {
  const { namespaceID } = props.namespace
  return store.dispatch('pageLayout/findByPageID', { namespaceID, pageID: props.pageID }).then(ly => {
    layouts.value = ly.map(l => {
      l = new compose.PageLayout(l)
      l.label = l.meta.title || l.handle || l.pageLayoutID
      return l
    })
  })
}

function checkRequiredRecordFields () {
  if (!module.value) return true
  const req = new Set(module.value.fields.filter(({ isRequired = false }) => isRequired).map(({ name }) => name))
  for (const b of usedBlocks.value) {
    if (b.kind !== 'Record') continue
    if (!b.options || !b.options.fields.length) return true
    for (const f of b.options.fields) req.delete(f.name)
  }
  return !req.size
}

async function handleSaveLayout ({ closeOnSuccess = false, previewOnSuccess = false, alert = true } = {}) {
  const { namespaceID } = props.namespace
  if (module.value && !checkRequiredRecordFields()) {
    toastErrorHandler(t('notification.page.saveFailedRequired'))()
    return
  }

  const hasInvalidRecordList = usedBlocks.value.some(b => {
    if (b.kind === 'RecordList' && b.options.editable) {
      const recordListModule = getModuleByID.value(b.options.moduleID)
      if (!recordListModule) return false
      const req = new Set(recordListModule.fields.filter(({ isRequired = false }) => isRequired).map(({ name }) => name))
      for (const f of b.options.editFields) req.delete(f.name)
      return req.size > 0
    }
    return false
  })

  if (hasInvalidRecordList) {
    toastErrorHandler(t('notification.page.saveFailedRequired'))()
    return
  }

  processing.value = true
  if (closeOnSuccess) processingSaveAndClose.value = true
  else processingSave.value = true

  return Promise.all([
    store.dispatch('page/findByID', { ...page.value, force: true }),
    store.dispatch('pageLayout/findByID', { ...layout.value }),
  ]).then(([p, ly]) => {
    const blocksData = [
      ...p.blocks.filter(({ blockID }) => {
        return !blocks.value.some(b => b.blockID === blockID) && layouts.value.some(({ pageLayoutID, blocks: lyBlocks }) => pageLayoutID !== ly.pageLayoutID && lyBlocks.some(b => b.blockID === blockID))
      }),
      ...blocks.value,
    ]
    const savePayload = { ...p, namespaceID, blocks: blocksData }
    if (page.value.config) savePayload.config = page.value.config
    if (page.value.meta.scenarios) {
      savePayload.meta = { ...(p.meta || {}), scenarios: page.value.meta.scenarios }
    }
    return store.dispatch('page/update', savePayload)
      .then((freshPage) => {
        page.value.blocks = freshPage.blocks
        const newBlocks = blocks.value.map(({ blockID, meta, xywh }) => {
          if (blockID === NoID) blockID = (freshPage.blocks.find(b => b.meta.tempID === meta.tempID) || {}).blockID
          return { blockID, xywh: normalizeXYWH(xywh), meta }
        })
        return store.dispatch('pageLayout/update', { ...ly, blocks: newBlocks })
      })
  }).then(async () => {
    unsavedBlocks.value.clear()
    if (closeOnSuccess) { router.push(previousPage.value || { name: 'admin.pages' }); return }
    if (alert) toastSuccess(t('notification.page.page-layout.save.success'))
    await fetchPageLayouts()
    setLayout(layout.value ? layout.value.pageLayoutID : undefined, false)
  }).finally(() => {
    processing.value = false
    if (closeOnSuccess) processingSaveAndClose.value = false
    else processingSave.value = false
  }).catch(toastErrorHandler(t('notification.page.page-layout.save.failed')))
}

async function handleCloneLayout ({ ref = false }) {
  processing.value = true
  processingLayout.value = true
  processingClone.value = true

  const ly = {
    ...layout.value.clone(),
    handle: '',
    weight: layouts.value.length + 1,
  }
  ly.meta.title = `${t('copyOf')}${ly.meta.title}`

  if (!ref) {
    const oldBlockIDs = {}
    ly.blocks = []
    blocks.value = blocks.value.toSorted((a, b) => {
      if (a.kind === 'Tabs' && b.kind !== 'Tabs') return 1
      else if (a.kind !== 'Tabs' && b.kind === 'Tabs') return -1
      return 0
    }).map(b => {
      const oldBlockID = b.blockID
      if (b.kind === 'Tabs') {
        b.options.tabs = b.options.tabs.map(tab => { tab.blockID = oldBlockIDs[tab.blockID]; return tab })
      }
      b = b.clone()
      oldBlockIDs[oldBlockID] = b.meta.tempID
      return b
    })
  }

  store.dispatch('pageLayout/create', ly).then(newLayout => {
    layout.value = newLayout
    layouts.value.push({ ...newLayout, label: newLayout.meta.title || newLayout.handle || newLayout.pageLayoutID })
    return handleSaveLayout({ alert: false })
  }).then(() => {
    toastSuccess(t('notification.page.page-layout.clone.success'))
  }).finally(() => {
    processing.value = false
    processingLayout.value = false
    processingClone.value = false
  }).catch(toastErrorHandler(t('notification.page.page-layout.clone.failed')))
}

function handleDeleteLayout () {
  processing.value = true
  processingDelete.value = true
  store.dispatch('pageLayout/delete', { ...layout.value }).then(() => {
    return fetchPageLayouts()
  }).then(() => {
    setLayout()
    toastSuccess(t('notification.page.page-layout.delete.success'))
  }).finally(() => {
    processing.value = false
    processingDelete.value = false
  }).catch(toastErrorHandler(t('notification.page.page-layout.delete.failed')))
}

function isValid (block) {
  if (typeof block.validate === 'function') return block.validate().length === 0
  return true
}

async function copyBlock (index) {
  const block = JSON.stringify(blocks.value[index].clone())
  if (block.kind === 'Tabs') {
    const { tabs = [] } = block.options
    block.options.tabs = tabs.map(b => {
      const { tempID } = (blocks.value.find(({ blockID }) => blockID === b.blockID) || {}).meta || {}
      b.blockID = tempID
      return b
    })
  }
  navigator.clipboard.writeText(block).then(() => {
    toastSuccess(t('notification.page.copySuccess'))
    document.getElementById('page-builder').focus()
  }, (err) => {
    toastErrorHandler(t('notification.page.copyFailed', { reason: err }))
  })
}

function pasteBlock (event) {
  if (document.querySelector('#page-builder') === document.activeElement) {
    event.preventDefault()
    const paste = (event.clipboardData || window.clipboardData).getData('text')
    try {
      const block = compose.PageBlockMaker(JSON.parse(paste))
      if (isValid(block)) appendBlock(block, t('notification.page.pasteSuccess'))
    } catch { toastWarning(t('notification.page.invalidBlock')) }
  }
}

async function setLayout (layoutID, processingFlag = true) {
  const oldLayoutID = route.query.layoutID

  if (layoutID && oldLayoutID !== layoutID) {
    try { await router.replace({ ...route, query: { ...route.query, layoutID } }) }
    catch { return }
  }

  if (processingFlag) processingLayout.value = true
  layoutID = layoutID || route.query.layoutID

  if (layoutID) layout.value = layouts.value.find(({ pageLayoutID }) => pageLayoutID === layoutID)
  layout.value = layout.value || layouts.value[0]
  if (!layout.value) {
    toastWarning(t('notification.page.page-layout.notFound.edit'))
    router.push(pageEditor.value)
    return
  }

  if (route.query.layoutID !== layout.value.pageLayoutID) {
    router.replace({ ...route, query: { ...route.query, layoutID: layout.value.pageLayoutID } })
  }

  unsavedBlocks.value.clear()
  const tempBlocks = []
  const { blocks: lyBlocks = [] } = layout.value || {}
  lyBlocks.forEach(({ blockID, xywh, meta = {} }) => {
    let block = page.value.blocks.find(b => b.blockID === blockID)
    if (block) {
      block.xywh = normalizeXYWH(xywh)
      block.meta.hidden = !!meta.hidden
      tempBlocks.push(block)
      if (block.kind === 'Tabs') {
        const { tabs = [] } = block.options
        tabs.forEach(tab => {
          if (lyBlocks.some(b => b.blockID === tab.blockID)) return
          block = page.value.blocks.find(b => b.blockID === tab.blockID)
          if (block) tempBlocks.push(block)
        })
      }
    }
  })
  blocks.value = tempBlocks

  setTimeout(() => {
    processingLayout.value = false
    if (!blocks.value.length) {
  getBlockSelectorModal()?.show()
    }
  }, 400)
}

function scrollToBottom () {
  const el = document.getElementById('page-builder')
  nextTick(() => {
    el.scrollTo({ top: el.scrollHeight, behavior: 'smooth' })
  })
}

function setDefaultValues () {
  title.value = ''
  processing.value = false
  processingSaveAndClose.value = false
  processingSave.value = false
  processingClone.value = false
  processingLayout.value = false
  page.value = undefined
  layout.value = undefined
  layouts.value = []
  blocks.value = []
  editor.value = undefined
  unsavedBlocks.value.clear()
}

function destroyEvents () {
  window.removeEventListener('paste', pasteBlock)
}
</script>

<style lang="scss">
#page-builder .vue-grid-layout {
  background-color: var(--white, #ffffff);
  border-radius: 8px;
  box-shadow: inset 0 0 0 1px var(--extra-light, #e9ecef);
  padding: 4px;
}

#page-builder .grid-item {
  transition: box-shadow 0.15s ease;

  &:hover {
    box-shadow: 0 0 0 2px var(--primary);
    z-index: 10;
  }
}

div.toolbox {
  position: absolute;
  background-color: var(--secondary);
  bottom: 0;
  left: 0;
  z-index: 1001;
  border-top-right-radius: 10px;
  opacity: 0.85;
  pointer-events: auto;
  transition: opacity 0.15s ease;

  &:hover {
    opacity: 1;
  }
}

[dir="rtl"] {
  div.toolbox {
    left: 0;
    right: auto;
  }
}
</style>
