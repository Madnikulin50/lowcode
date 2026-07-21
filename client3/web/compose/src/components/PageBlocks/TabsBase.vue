<template>
  <Wrap v-bind="$props" :scrollable-body="false" card-class="tabs-base-block-container" header-class="border-0 border-white">
    <div v-if="!tabbedBlocks.length && editable" class="d-flex h-100 align-items-center justify-content-center">
      <p class="mb-0">{{ $t('noTabs') }}</p>
    </div>

    <BTabs
      v-else
      card
      :pills="styleOpts.appearance === 'pills'"
      :small="styleOpts.appearance === 'small'"
      :vertical="styleOpts.orientation === 'vertical'"
      :end="styleOpts.position === 'end'"
      :justified="styleOpts.justify === 'justify'"
      :fill="styleOpts.justify === 'justify'"
      :align="styleOpts.alignment"
      no-fade
      lazy
      :content-class="contentClass"
      class="h-100"
      :class="{ 'd-flex flex-column': styleOpts.orientation !== 'vertical' }"
      @activate-tab="onTabActivated"
    >
      <BTab
        v-for="(tab, index) in tabbedBlocks"
        :key="`${getTabTitle(tab, index)}-${index}`"
        class="h-100 overflow-hidden"
        :title-link-class="getTitleLinkClass(index)"
        no-body
      >
        <template #title>
          <span>
            {{ getTabTitle(tab, index) }}
            <font-awesome-icon v-if="hasTabErrors(index)" :icon="['fas', 'exclamation-triangle']" :class="errorTriangleClass(index)" class="ms-1" />
          </span>

          <div v-if="tab.block && editable" class="d-inline ms-3">
            <div v-if="unsavedBlocks.has(tab.block.blockID !== '0' ? tab.block.blockID : tab.block.meta.tempID)" class="btn btn-sm border-0 p-0 px-1" data-bs-toggle="tooltip" :title="$t('label.unsavedChanges')">
              <font-awesome-icon :icon="['fas', 'exclamation-triangle']" class="text-warning" />
            </div>

            <div class="btn-group btn-group-sm">
              <button class="btn btn-outline-light text-primary border-0 toolbox-button p-0 px-1" data-bs-toggle="tooltip" :title="$t('tooltip.edit')" @click="editTabbedBlock(tab)">
                <font-awesome-icon :icon="['far', 'edit']" />
              </button>
              <button class="btn btn-outline-light text-primary border-0 toolbox-button p-0 px-1" data-bs-toggle="tooltip" :title="$t('tooltip.clone')" @click="cloneTabbedBlock(tab)">
                <font-awesome-icon :icon="['far', 'clone']" />
              </button>
              <button class="btn btn-outline-light text-primary border-0 toolbox-button p-0 px-1" data-bs-toggle="tooltip" :title="$t('tooltip.copy')" @click="copyTabbedBlock(tab)">
                <font-awesome-icon :icon="['far', 'copy']" />
              </button>
            </div>

            <c-input-confirm :tooltip="$t('tooltip.delete')" show-icon button-class="p-0 px-1" class="ms-1" @confirmed="deleteTab(index)" />
          </div>
        </template>

        <PageBlockTab v-if="tab.block && shouldRenderTab(tab, index)" v-bind="{ ...$props, page, block: tab.block, blockIndex: index }" :record="record" :module="module" :magnified="magnified" header-class="border-0 border-white" @errors="setTabErrors(index, $event)" />

        <div v-else-if="!tab.block" class="d-flex h-100 align-items-center justify-content-center">
          <p class="mb-0">{{ $t('noBlock') }}</p>
        </div>
      </BTab>
    </BTabs>
  </Wrap>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePageBlockBase } from './usePageBlockBase'
import { compose, NoID } from 'corteza-lib/js/dist'
import { fetchID } from 'corteza-webapp-compose/src/lib/block'
import { evaluatePrefilter } from 'corteza-webapp-compose/src/lib/record-filter'
import Wrap from './Wrap/index.js'
import PageBlockTab from './index.js'
import { BTabs, BTab } from 'bootstrap-vue-next'

const { t: $t } = useI18n({ useScope: 'global', messages: {}, keyPrefix: 'block:tabs' })

const props = defineProps({
  blockIndex: { type: Number, default: -1 },
  namespace: { type: Object, required: true },
  page: { type: Object, required: true },
  blocks: { type: Array, default: () => [] },
  block: { type: Object, required: true },
  module: { type: Object, required: false, default: undefined },
  record: { type: Object, required: false, default: undefined },
  mode: { type: String, required: false, default: '' },
  editable: { type: Boolean, required: false, default: false },
  resizing: { type: Boolean, required: false, default: false },
  magnified: { type: Boolean, required: false, default: false },
  unsavedBlocks: { type: Set, default: () => new Set() },
  loadingRecord: { type: Boolean, required: false, default: false },
  errors: { type: Object, required: false, default: () => ({}) },
})

const emit = defineEmits(['edit-block', 'clone-block', 'copy-block', 'delete-tab', 'errors'])

const { options, refreshBlock } = usePageBlockBase(props, emit)

const activeTab = ref(0)
const tabErrors = ref([])
const visitedTabs = ref({ 0: true })

const $auth = typeof window !== 'undefined' ? window.__auth : null

const tabbedBlocks = computed(() => {
  return (props.block.options.tabs || []).reduce((acc, { blockID, title, lazy = true }) => {
    const unparsedBlock = blockID ? props.blocks.find(b => fetchID(b) === blockID) : undefined
    if (!unparsedBlock) {
      if (!blockID && title) acc.push({ title })
      return acc
    }
    if (unparsedBlock.meta?.invisible && !props.editable) return acc
    let block = JSON.parse(JSON.stringify(unparsedBlock))
    block.style.wrap.kind = 'Plain'
    block.style.border.enabled = false
    block = compose.PageBlockMaker(block)
    acc.push({ block, title, lazy })
    return acc
  }, [])
})

const styleOpts = computed(() => props.block.options.style || {})

const contentClass = computed(() => {
  return `overflow-hidden mh-100 ${styleOpts.value.orientation === 'vertical' ? '' : 'flex-fill'}`
})

function getTitleLinkClass(index) {
  const { justify, alignment, orientation } = styleOpts.value
  return `${orientation === 'horizontal' ? 'text-truncate' : ''} text-${alignment} ${justify !== 'none' ? 'flex-fill' : ''}`
}

function getTabTitle({ title = '', block = {} }, tabIndex) {
  const { title: blockTitle, kind } = block
  const interpolatedTitle = evaluatePrefilter(blockTitle, {
    record: props.record,
    user: $auth?.user || {},
    recordID: (props.record || {}).recordID || NoID,
    ownerID: (props.record || {}).ownedBy || NoID,
    userID: ($auth?.user || {}).userID || NoID,
  })
  title = evaluatePrefilter(title, {
    record: props.record,
    user: $auth?.user || {},
    recordID: (props.record || {}).recordID || NoID,
    ownerID: (props.record || {}).ownedBy || NoID,
    userID: ($auth?.user || {}).userID || NoID,
  })
  return title || interpolatedTitle || kind || `${$t('tab')} ${tabIndex + 1}`
}

function onTabActivated({ newTabIndex }) {
  activeTab.value = newTabIndex
  visitedTabs.value[newTabIndex] = true
}

function shouldRenderTab({ lazy = true }, index) {
  if (!lazy) return true
  return !!visitedTabs.value[index]
}

function editTabbedBlock(tab) {
  const blockIndex = props.blocks.findIndex(b => fetchID(b) === fetchID(tab.block))
  if (blockIndex > -1) emit('edit-block', blockIndex)
}

function cloneTabbedBlock(tab) {
  const tabbedBlockIndex = props.blocks.findIndex(b => fetchID(b) === fetchID(tab.block))
  if (tabbedBlockIndex > -1) emit('clone-block', { tabbedBlockIndex, tabBlockIndex: props.blockIndex, title: tab.title })
}

function copyTabbedBlock(tab) {
  const tabbedBlockIndex = props.blocks.findIndex(b => fetchID(b) === fetchID(tab.block))
  if (tabbedBlockIndex > -1) emit('copy-block', tabbedBlockIndex)
}

function deleteTab(tabIndex) {
  emit('delete-tab', { tabIndex, blockIndex: props.blockIndex })
}

function setTabErrors(index, { errors, id }) {
  if (!tabErrors.value[index]) tabErrors.value[index] = {}
  if (!errors) {
    tabErrors.value[index][id] = undefined
  } else {
    const errorKind = errors.set.some(error => error.meta?.isWarning) ? 'warning' : 'error'
    tabErrors.value[index][id] = errorKind
  }
}

function hasTabErrors(index) {
  if (props.mode === 'base' || !tabErrors.value[index]) return false
  return Object.values(tabErrors.value[index]).filter(error => !!error).length > 0
}

function errorTriangleClass(index) {
  const errorKinds = Object.values(tabErrors.value[index]).filter(error => !!error)
  if (errorKinds.length > 0) return errorKinds.includes('error') ? 'text-danger' : 'text-secondary'
  return undefined
}

onMounted(() => { refreshBlock(() => {}) })
</script>
<style lang="scss">
.tabs-base-block-container {
  .nav-pills .active .toolbox-button { color: var(--white) !important; }
}
</style>
