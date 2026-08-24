<template>
  <Wrap v-bind="$props" @refreshBlock="refresh">
    <div v-if="isProcessing" class="d-flex align-items-center justify-content-center h-100">
      <span class="spinner-border" />
    </div>
    <div
      v-else-if="unknown"
      class="d-flex flex-column align-items-center justify-content-center h-100 text-secondary px-3 text-center"
      :title="t('progress.countUnavailableHint')"
    >
      <span class="fs-4">{{ t('metric.emptyPlaceholder') || '—' }}</span>
      <small>{{ t('progress.countUnavailable') }}</small>
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
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, watch, onBeforeUnmount, onMounted } from 'vue'
import { composables, components, useNsI18n } from 'corteza-lib/vue/dist'
import { NoID, isUnknownTotal } from 'corteza-lib/js/dist'
import { evalPrefilterOrSkip, isFieldInFilter } from 'corteza-webapp-compose/src/lib/record-filter'
import { usePageBlockBase } from './usePageBlockBase'
import Wrap from './Wrap/index.js'

const { CProgress } = components
const { toastErrorHandler } = composables.useToast()
const t = useNsI18n()
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
const unknown = ref(false)

watch(() => [props.record?.recordID, props.loadingRecord], () => { if (!props.loadingRecord) refresh() }, { immediate: true })
watch(options, () => { if (!props.loadingRecord) refresh() }, { deep: true })

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
  const pfCtx = {
    record: props.record, user: $auth?.user || {},
    recordID: (props.record || {}).recordID || NoID,
    ownerID: (props.record || {}).ownedBy || NoID,
    userID: ($auth?.user || {}).userID || NoID,
    loadingRecord: !!props.loadingRecord,
  }
  function evalFilter (f) {
    const { skip, filter } = evalPrefilterOrSkip(f, pfCtx)
    return skip ? 'false' : filter
  }
  const additionalOptions = {
    value: { filter: evalFilter(options.value.value.filter) },
    minValue: { filter: evalFilter(options.value.minValue.filter) },
    maxValue: { filter: evalFilter(options.value.maxValue.filter) },
  }
  return props.block.fetch(additionalOptions, $ComposeAPI, namespaceID)
    .then(({ value: v, min: m = 0, max: mx = 100, unknown: unk }) => {
      unknown.value = !!(unk || isUnknownTotal(v))
      min.value = isUnknownTotal(m) ? 0 : m
      max.value = isUnknownTotal(mx) ? 100 : mx
      value.value = unknown.value ? 0 : v
    })
    .catch(toastErrorHandler('Progress fetch failed'))
    .finally(() => { setTimeout(() => { processing.value = false }, 300) })
}

function refreshOnRelatedRecordsUpdate({ detail: { moduleID } } = {}) {
  if (options.value.value?.moduleID === moduleID) refresh()
}
</script>
