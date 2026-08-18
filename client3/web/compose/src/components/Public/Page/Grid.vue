<template>
  <grid
    :blocks="blocks"
    :editable="false"
    :loading-record="loadingRecord"
    class="h-100"
  >
    <template #default="{ blockIndex, block, resizing }">
      <component
        :is="componentFor(block)"
        :namespace="namespace"
        :module="module"
        :page="page"
        :record="record"
        :blocks="blocks"
        :block="block"
        :block-index="blockIndex"
        :loading-record="loadingRecord"
        :mode="mode"
        :errors="errors"
        :editable="editable"
        :resizing="resizing"
        :magnified="magnified"
        :unsaved-blocks="unsavedBlocks"
        class="p-2"
      />
    </template>
  </grid>
</template>

<script setup>
defineOptions({ inheritAttrs: false })
import Grid from '../../Common/Grid.vue'
import { GetComponent } from '../../PageBlocks/index.js'
import { compose } from 'corteza-lib/js/dist'

const props = defineProps({
  page: { type: compose.Page, required: true },
  blocks: { type: Array, required: true },
  loadingRecord: { type: Boolean, default: false },
  namespace: { type: Object, required: false, default: undefined },
  module: { type: Object, required: false, default: undefined },
  record: { type: Object, required: false, default: undefined },
  mode: { type: String, default: '' },
  errors: { type: Object, default: undefined },
  editable: { type: Boolean, default: false },
  magnified: { type: Boolean, default: false },
  unsavedBlocks: { type: Set, default: undefined },
})

function componentFor(block) {
  return GetComponent({ block, mode: props.mode })
}
</script>
