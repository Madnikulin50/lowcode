<template>
  <Wrap @refreshBlock="refresh">
    <div v-if="isProcessing" class="d-flex align-items-center justify-content-center h-100">
      <span class="spinner-border" />
    </div>
    <div v-else class="d-flex h-100" :class="{ 'p-2': block.style.wrap.kind === 'card' }">
      <c-progress
        :value="value"
        :min="min"
        :max="max"
        :labeled="options.display.showValue"
        :relative="options.display.showRelative"
        :progress="options.display.showProgress"
        :striped="options.display.striped"
        :animated="options.display.animated"
        :variant="options.display.variant"
        :thresholds="options.display.thresholds"
        class="flex-fill h-100"
      />
    </div>
  </Wrap>
</template>

<script setup>
import { ref, watch, onBeforeUnmount, onMounted } from 'vue'
import { composables } from 'corteza-lib/vue/dist'
import { NoID } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'
import { evaluatePrefilter, isFieldInFilter } from 'corteza-webapp-compose/src/lib/record-filter'
import { usePageBlockBase } from './usePageBlockBase'
import Wrap from './Wrap/index.js'

const { CProgress } = components
const { toastErrorHandler } = composables.useToast()
const $auth = window.__auth
const $ComposeAPI = window.__composeAPI

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

const { processing, isProcessing, options, refreshBlock } = usePageBlockBase(props, {})

const value = ref(undefined)
const min = ref(undefined)
const max = ref(undefined)

watch(() => props.record?.recordID, () => { refresh() }, { immediate: true })
watch(options, () => { refresh() }, { deep: true })

onMounted(() => {
  refreshBlock(refresh)
  window.addEventListener('module-records-updated', refreshOnRelatedRecordsUpdate)
  window.addEventListener('record-field-change', refetchOnPrefilterValueChange)
  window.addEventListener('refetch-records', refresh)
})

onBeforeUnmount(() => {
  window.removeEventListener('module-records-updated', refreshOnRelatedRecordsUpdate)
  window.removeEventListener('record-field-change', refetchOnPrefilterValueChange)
  window.removeEventListener('refetch-records', refresh)
})

function refetchOnPrefilterValueChange({ detail: { fieldName } }) {
  if (isFieldInFilter(fieldName, options.value.value.filter)) refresh()
}

async function refresh() {
  processing.value = true
  const { namespaceID } = props.namespace || {}
  const additionalOptions = {
    value: { filter: evaluatePrefilter(options.value.value.filter, {
      record: props.record, user: $auth?.user || {},
      recordID: (props.record || {}).recordID || NoID,
      ownerID: (props.record || {}).ownedBy || NoID,
      userID: ($auth?.user || {}).userID || NoID,
    })},
    minValue: { filter: evaluatePrefilter(options.value.minValue.filter, {
      record: props.record, user: $auth?.user || {},
      recordID: (props.record || {}).recordID || NoID,
      ownerID: (props.record || {}).ownedBy || NoID,
      userID: ($auth?.user || {}).userID || NoID,
    })},
    maxValue: { filter: evaluatePrefilter(options.value.maxValue.filter, {
      record: props.record, user: $auth?.user || {},
      recordID: (props.record || {}).recordID || NoID,
      ownerID: (props.record || {}).ownedBy || NoID,
      userID: ($auth?.user || {}).userID || NoID,
    })},
  }
  return props.block.fetch(additionalOptions, $ComposeAPI, namespaceID)
    .then(({ value: v, min: m = 0, max: mx = 100 }) => { min.value = m; max.value = mx; value.value = v })
    .catch(toastErrorHandler('Progress fetch failed'))
    .finally(() => { setTimeout(() => { processing.value = false }, 300) })
}

function refreshOnRelatedRecordsUpdate({ detail: { moduleID } } = {}) {
  if (options.value.value?.moduleID === moduleID) refresh()
}
</script>
