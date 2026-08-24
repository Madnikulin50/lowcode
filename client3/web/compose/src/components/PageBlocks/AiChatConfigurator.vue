<template>
  <div class="p-3 overflow-auto">
    <div class="mb-3">
      <label class="form-label fw-semibold">{{ t('ai.chat.startPrompt.label') }}</label>
      <textarea
        v-model="startPrompt"
        class="form-control"
        rows="4"
        :placeholder="t('ai.chat.startPrompt.placeholder')"
      />
      <div class="form-text">{{ t('ai.chat.startPrompt.description') }}</div>
    </div>
    <div class="mb-3">
      <label class="form-label fw-semibold">{{ t('ai.chat.model.label') }}</label>
      <select
        v-model="selectedModel"
        class="form-select"
        :disabled="!modelOptions.length"
      >
        <option value="">{{ t('ai.chat.model.placeholder') }}</option>
        <option
          v-for="m in modelOptions"
          :key="m"
          :value="m"
        >
          {{ modelLabel(m) }}
        </option>
      </select>
      <div class="form-text">{{ t('ai.chat.model.description') }}</div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed, onMounted, inject } from 'vue'
import { useNsI18n } from 'corteza-lib/vue/dist'
import { parseModelsPayload, modelLabel } from '../Public/Page/AiChat/chatTools.js'

const t = useNsI18n()
const $ComposeAPI = inject('$ComposeAPI', window.__composeAPI)

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

defineEmits(['errors'])

const modelOptions = ref([])

function blockOptions () {
  if (!props.block.options || typeof props.block.options !== 'object') {
    props.block.options = {}
  }
  return props.block.options
}

const startPrompt = computed({
  get: () => {
    const o = props.block?.options || {}
    return o.startPrompt || o.prompt || ''
  },
  set: (val) => {
    const o = blockOptions()
    const text = val == null ? '' : String(val)
    o.startPrompt = text
    o.prompt = text
  },
})

const selectedModel = computed({
  get: () => props.block?.options?.model || '',
  set: (val) => {
    blockOptions().model = val == null ? '' : String(val)
  },
})

onMounted(() => {
  const o = blockOptions()
  if (!o.startPrompt && o.prompt) o.startPrompt = String(o.prompt)
  if (!o.model) o.model = ''
  if (!$ComposeAPI?.pageAiModels) return
  $ComposeAPI.pageAiModels().then((payload = {}) => {
    const names = parseModelsPayload(payload).names
    const current = o.model
    modelOptions.value = current && !names.includes(current) ? [current, ...names] : names
  }).catch(() => {})
})
</script>
