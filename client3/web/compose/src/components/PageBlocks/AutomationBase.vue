<template>
  <Wrap
    v-bind="$props"
  >
    <div
      v-if="processing"
      class="d-flex align-items-center justify-content-center h-100"
    >
      <span class="spinner-border spinner-border-sm" />
    </div>

    <automation-buttons
      v-else
      class="d-flex flex-wrap p-3"
      button-class="flex-fill"
      :buttons="options.buttons"
      :automation-scripts="automationScripts"
      v-bind="$props"
    />
  </Wrap>
</template>

<script setup>
import { ref, computed, onBeforeUnmount, inject } from 'vue'
import { usePageBlockBase } from './usePageBlockBase'
import Wrap from './Wrap/index.js'
import AutomationButtons from './Shared/AutomationButtons'

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
  extraEventArgs: { type: Object, default: () => ({}) },
})

const $ComposeAPI = inject('$ComposeAPI')
const $UIHooks = inject('$UIHooks')

const emit = defineEmits(['errors'])

const { options, setBaseDefaultValues } = usePageBlockBase(props, emit)

const processing = ref(false)
const automationScripts = ref([])
const abortableRequests = ref([])

const hasUIHooks = computed(() => $UIHooks.set && !!$UIHooks.set.length)

onBeforeUnmount(() => {
  abortRequests()
  setBaseDefaultValues()
})

if (!hasUIHooks.value) {
  fetchAutomationLists()
}

function fetchAutomationLists () {
  processing.value = true
  const { response, cancel } = $ComposeAPI
    .automationListCancellable({ eventTypes: ['onManual'], excludeInvalid: true })
  abortableRequests.value.push(cancel)
  return response()
    .then(({ set = [] }) => {
      automationScripts.value = set
    })
    .finally(() => {
      setTimeout(() => { processing.value = false }, 300)
    })
}

function abortRequests () {
  abortableRequests.value.forEach((cancel) => cancel())
}
</script>
