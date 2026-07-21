<template>
  <Wrap
    v-bind="$props"
    @refreshBlock="refresh"
  >
    <img
      v-if="src"
      ref="iframe"
      class="h-100 w-100 border-0"
      :src="src"
      style="object-fit: contain;"
    >
  </Wrap>
</template>

<script setup>
import { computed, ref, onMounted, onBeforeUnmount } from 'vue'
import { NoID } from 'corteza-lib/js/dist'
import { evaluatePrefilter, isFieldInFilter } from 'corteza-webapp-compose/src/lib/record-filter'
import { usePageBlockBase } from './usePageBlockBase'
import Wrap from './Wrap/index.js'

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

const $auth = inject('$auth')

const emit = defineEmits(['errors'])
const iframe = ref(null)

const { refreshBlock } = usePageBlockBase(props, emit)

const src = computed(() => {
  const { srcField, src: srcUrl } = props.block.options
  const blank = 'about:blank'
  let url = srcUrl
  if (props.block.options.srcField) {
    if (props.record) url = props.record.values[srcField]
  }
  let interpolatedURL = evaluatePrefilter(url, {
    record: props.record,
    user: $auth.user || {},
    recordID: (props.record || {}).recordID || NoID,
    ownerID: (props.record || {}).ownedBy || NoID,
    userID: ($auth.user || {}).userID || NoID,
  })
  if (interpolatedURL[0] !== 'h') interpolatedURL = window.CortezaAPI + interpolatedURL
  return interpolatedURL || blank
})

onMounted(() => {
  refreshBlock(refresh)
  createEvents()
})

onBeforeUnmount(() => {
  destroyEvents()
})

function refresh () {
  if (iframe.value) iframe.value.src = src.value
}

function createEvents () {
  window.addEventListener('record-field-change', refetchOnPrefilterValueChange)
}

function refetchOnPrefilterValueChange ({ fieldName }) {
  const { src: s } = props.block.options
  if (isFieldInFilter(fieldName, s)) refresh()
}

function destroyEvents () {
  window.removeEventListener('record-field-change', refetchOnPrefilterValueChange)
}
</script>

<style scoped lang="scss">
img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
</style>
