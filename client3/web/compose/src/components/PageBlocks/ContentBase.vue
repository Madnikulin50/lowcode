<template>
  <Wrap
    v-bind="$props"
  >
    <div
      :style="{ 'white-space': 'pre-wrap' }"
      class="rt-content p-3"
      v-html="contentBody"
    />
  </Wrap>
</template>

<script setup>
import { computed, inject } from 'vue'
import { NoID } from 'corteza-lib/js/dist'
import { evaluatePrefilter } from 'corteza-webapp-compose/src/lib/record-filter'
import { useI18n } from 'vue-i18n'

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

const contentBody = computed(() => {
  try {
    const { body = '' } = props.block.options
    return evaluatePrefilter(body, {
      record: props.record,
      user: $auth.user || {},
      recordID: (props.record || {}).recordID || NoID,
      ownerID: (props.record || {}).ownedBy || NoID,
      userID: ($auth.user || {}).userID || NoID,
    })
  } catch (e) {
    return e
  }
})
</script>
