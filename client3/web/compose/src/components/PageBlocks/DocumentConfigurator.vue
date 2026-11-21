<template>
  <div class="tab-pane">
    <div class="mb-3">
      <label class="form-label text-primary">{{ $t('document.pick') }}</label>
      <select
        v-model="options.documentID"
        class="form-select"
      >
        <option value="">
          {{ $t('document.empty') }}
        </option>
        <option
          v-for="doc in documents"
          :key="doc.documentID"
          :value="doc.documentID"
        >
          {{ doc.title }}
        </option>
      </select>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { inject, onMounted, ref } from 'vue'
import { usePageBlockBase } from './usePageBlockBase'

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

const emit = defineEmits(['errors'])
const { options } = usePageBlockBase(props, emit)
const $ComposeAPI = inject('$ComposeAPI', null) || window.__composeAPI
const documents = ref([])

onMounted(async () => {
  if (!props.namespace?.namespaceID) return
  try {
    const { set } = await $ComposeAPI.documentList({ namespaceID: props.namespace.namespaceID })
    documents.value = set || []
  } catch {
    documents.value = []
  }
})
</script>
