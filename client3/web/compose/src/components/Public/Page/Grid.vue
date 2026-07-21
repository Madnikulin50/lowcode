<template>
  <grid
    :blocks="blocks"
    :editable="false"
  >
    <template #default="{ blockIndex, block }">
      <component
        :is="componentFor(block)"
        v-bind="$attrs"
        :page="page"
        :blocks="blocks"
        :block="block"
        :block-index="blockIndex"
        :loading-record="loadingRecord"
        class="p-2"
      />
    </template>
  </grid>
</template>

<script setup>
import Grid from '../../Common/Grid.vue'
import { GetComponent } from '../../PageBlocks/index.js'
import { compose } from 'corteza-lib/js/dist'

const props = defineProps({
  page: { type: compose.Page, required: true },
  blocks: { type: Array, required: true },
  loadingRecord: { type: Boolean, default: false },
})

function componentFor(block) {
  return GetComponent({ block })
}
</script>
