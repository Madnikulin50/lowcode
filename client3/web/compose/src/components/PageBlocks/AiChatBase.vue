<script setup>
import Wrap from './Wrap/index.js'
import { computed } from 'vue'
import { usePageBlockBase } from './usePageBlockBase'
import Chat from '../Public/Page/AiChat/Chat.vue'

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

const emit = defineEmits(['errors'])

const { options } = usePageBlockBase(props, emit)

const startPrompt = computed(() => {
  const o = props.block?.options || options.value || {}
  return o.startPrompt || o.prompt || props.block?.prompt || props.page?.config?.prompt || props.namespace?.meta?.prompt || ''
})

const preferredModel = computed(() => options.value.model || '')

const modelStorageKey = computed(() => {
  const id = props.block?.blockID || props.block?.meta?.tempID || props.blockIndex
  return `aiChat.model.block.${id || '0'}`
})
</script>

<template>
  <Wrap v-bind="$props">
    <div class="ai-chat-block h-100 d-flex flex-column">
      <Chat
        :start-prompt="startPrompt"
        :page="page?.pageID || ''"
        :module="module?.moduleID || ''"
        :namespace="namespace?.namespaceID || ''"
        :magnified="magnified"
        :framed="false"
        :show-model-switcher="true"
        :preferred-model="preferredModel"
        :model-storage-key="modelStorageKey"
      />
    </div>
  </Wrap>
</template>

<style scoped>
.ai-chat-block {
  min-height: 300px;
}
</style>
