<template>
  <Wrap v-bind="$props">
    <div
      v-if="loading"
      class="p-3 text-muted"
    >
      <span class="spinner-border spinner-border-sm" />
    </div>
    <div
      v-else-if="error"
      class="alert alert-warning m-3"
    >
      {{ error }}
    </div>
    <div
      v-else-if="!document"
      class="p-3 text-muted"
    >
      {{ $t('document.empty') }}
    </div>
    <div
      v-else-if="document.kind === 'markdown'"
      class="p-3 overflow-auto h-100"
    >
      <markdown-view :body="document.body" />
    </div>
    <pdf-view
      v-else-if="document.kind === 'pdf' && document.attachmentID && document.attachmentID !== '0'"
      :namespaceID="namespaceID"
      :documentID="document.documentID"
      :title="document.title"
    />
    <div
      v-else
      class="p-3 text-muted"
    >
      {{ $t('document.noFile') }}
    </div>
  </Wrap>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { computed, inject, ref, watch } from 'vue'
import Wrap from './Wrap/index.js'
import MarkdownView from '../Docs/MarkdownView.vue'
import PdfView from '../Docs/PdfView.vue'

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

const $ComposeAPI = inject('$ComposeAPI', null) || window.__composeAPI
const document = ref(null)
const loading = ref(false)
const error = ref('')
const namespaceID = computed(() => props.namespace?.namespaceID)
const documentID = computed(() => props.block?.options?.documentID || '')

watch([namespaceID, documentID], load, { immediate: true })

async function load () {
  document.value = null
  error.value = ''
  if (!namespaceID.value || !documentID.value) return
  loading.value = true
  try {
    const res = await $ComposeAPI.documentRead({
      namespaceID: namespaceID.value,
      documentID: documentID.value,
    })
    document.value = res.document
  } catch (e) {
    error.value = e?.message || String(e)
  } finally {
    loading.value = false
  }
}
</script>
